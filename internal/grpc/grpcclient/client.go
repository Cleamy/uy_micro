package grpcclient

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Cleamy/uy_micro/global"
	"github.com/Cleamy/uy_micro/internal/grpc/interceptor"
	"github.com/Cleamy/uy_micro/internal/registry"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultMaxPoolSize = 10 // 单服务默认最大连接数，可后续接入配置项

// connPool 单服务独立连接池
type connPool struct {
	mu          sync.Mutex
	idle        []*grpc.ClientConn // 空闲连接队列
	active      int                // 当前总连接数（空闲+使用中）
	maxSize     int                // 连接池容量上限
	serviceName string
}

// RpcClientFactory gRPC客户端工厂，按服务名维护独立连接池
type RpcClientFactory struct {
	pools sync.Map // key: 服务名, value: *connPool
}

func NewRpcClientFactory() *RpcClientFactory {
	return &RpcClientFactory{}
}

// getPool 获取或初始化对应服务的连接池（并发安全）
func (f *RpcClientFactory) getPool(serviceName string) *connPool {
	if p, ok := f.pools.Load(serviceName); ok {
		return p.(*connPool)
	}
	p := &connPool{
		maxSize:     defaultMaxPoolSize,
		serviceName: serviceName,
	}
	actual, _ := f.pools.LoadOrStore(serviceName, p)
	return actual.(*connPool)
}

// GetConn 从连接池中获取一条可用连接
// 返回的 cleanup 函数用于归还连接到连接池，请配合 defer 使用
func (f *RpcClientFactory) GetConn(serviceName string) (*grpc.ClientConn, func(), error) {
	pool := f.getPool(serviceName)

	pool.mu.Lock()
	// 1. 优先复用空闲连接
	for len(pool.idle) > 0 {
		conn := pool.idle[0]
		pool.idle = pool.idle[1:]

		// 健康检查：损坏连接直接丢弃
		if conn.GetState() == connectivity.Shutdown {
			pool.active--
			_ = conn.Close()
			continue
		}
		pool.mu.Unlock()

		// 归还函数：用完放回空闲池
		cleanup := func() {
			f.returnConn(serviceName, conn)
		}
		return conn, cleanup, nil
	}

	// 2. 无空闲连接，检查是否达到容量上限
	if pool.active >= pool.maxSize {
		pool.mu.Unlock()
		global.Logger.Warn("[GRPC Client] connection pool exhausted",
			zap.String("service", serviceName),
			zap.Int("max_size", pool.maxSize))
		return nil, nil, errors.New("connection pool max size reached")
	}

	// 3. 新建连接（注入全链路客户端拦截器）
	pool.active++
	pool.mu.Unlock()

	target, err := registry.GetServiceGrpcTargetDefault(serviceName, true)
	if err != nil {
		// 新建失败，回滚计数
		pool.mu.Lock()
		pool.active--
		pool.mu.Unlock()
		return nil, nil, err
	}

	global.Logger.Info("[GRPC Client] create new connection for pool",
		zap.String("service", serviceName),
		zap.String("target", target),
		zap.Int("current_total", pool.active))

	conn, err := grpc.Dial(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(
			interceptor.ChainUnaryClient(
				interceptor.RetryUnaryClientInterceptor(),      // 1. 自动重试
				interceptor.TraceUnaryClientInterceptor(),      // 2. 链路透传
				interceptor.PrometheusUnaryClientInterceptor(), // 3. 指标采集
				interceptor.SentinelUnaryClientInterceptor(),   // 4. 客户端熔断
				interceptor.UnaryClientInterceptor(),           // 5. 调用日志
			),
		),
	)
	if err != nil {
		pool.mu.Lock()
		pool.active--
		pool.mu.Unlock()
		global.Logger.Error("[GRPC Client] dial failed",
			zap.String("service", serviceName),
			zap.String("target", target),
			zap.Error(err))
		return nil, nil, fmt.Errorf("dial %s err: %w", target, err)
	}

	// 归还函数
	cleanup := func() {
		f.returnConn(serviceName, conn)
	}
	return conn, cleanup, nil
}

// returnConn 归还连接到连接池（内部自动健康检测）
func (f *RpcClientFactory) returnConn(serviceName string, conn *grpc.ClientConn) {
	pool := f.getPool(serviceName)
	pool.mu.Lock()
	defer pool.mu.Unlock()

	// 连接已损坏，直接丢弃关闭
	if conn.GetState() == connectivity.Shutdown {
		pool.active--
		_ = conn.Close()
		global.Logger.Warn("[GRPC Client] discard broken connection",
			zap.String("service", serviceName))
		return
	}

	// 正常连接放回空闲队列
	pool.idle = append(pool.idle, conn)
	global.Logger.Debug("[GRPC Client] connection returned to pool",
		zap.String("service", serviceName),
		zap.Int("idle_count", len(pool.idle)))
}

// CloseAll 关闭所有服务的全部连接，清空连接池，服务优雅关停时调用
func (f *RpcClientFactory) CloseAll() {
	f.pools.Range(func(key, value interface{}) bool {
		serviceName := key.(string)
		pool := value.(*connPool)

		pool.mu.Lock()
		// 关闭所有空闲连接
		for _, conn := range pool.idle {
			_ = conn.Close()
		}
		pool.idle = nil
		pool.active = 0
		pool.mu.Unlock()

		global.Logger.Info("[GRPC Client] connection pool cleared",
			zap.String("service", serviceName))
		return true
	})
}

// GetClient 内部泛型快捷方法（仅框架内部使用）
func GetClient[T any](serviceName string, newPbClient func(grpc.ClientConnInterface) T) (T, func(), error) {
	conn, clean, err := global.RpcFactory.GetConn(serviceName)
	if err != nil {
		var zero T
		return zero, nil, err
	}
	return newPbClient(conn), clean, nil
}
