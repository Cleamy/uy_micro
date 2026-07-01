package rpcFactory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Cleamy/uy_micro/global"
	"github.com/Cleamy/uy_micro/internal/grpc/interceptor"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// RpcClientFactory gRPC 客户端连接工厂
// 按服务名缓存长连接，统一服务发现与拦截器注入
type RpcClientFactory struct {
	mu    sync.RWMutex
	conns map[string]*grpc.ClientConn // 服务名 -> 长连接缓存
}

// NewRpcClientFactory 创建连接工厂实例
func NewRpcClientFactory() *RpcClientFactory {
	return &RpcClientFactory{
		conns: make(map[string]*grpc.ClientConn),
	}
}

// GetConn 获取指定服务的 gRPC 连接（自动复用缓存）
// serviceName：Consul 中注册的服务名称
func (f *RpcClientFactory) GetConn(serviceName string) (*grpc.ClientConn, error) {
	// 读缓存：优先复用已有连接
	f.mu.RLock()
	conn, ok := f.conns[serviceName]
	f.mu.RUnlock()

	// 缓存命中且连接未关闭，直接返回
	if ok && conn.GetState() != connectivity.Shutdown {
		return conn, nil
	}

	// 缓存未命中或连接失效，加写锁新建连接
	f.mu.Lock()
	defer f.mu.Unlock()

	// 双重检查，避免并发重复创建
	if conn, ok := f.conns[serviceName]; ok && conn.GetState() != connectivity.Shutdown {
		return conn, nil
	}

	// 1. 从 Consul 拉取一个健康实例
	addr, err := f.getHealthyInstance(serviceName)
	if err != nil {
		return nil, fmt.Errorf("discover service %s failed: %w", serviceName, err)
	}

	// 2. 构建统一拨号选项（所有拦截器、通用配置集中在这里）
	dialOpts := f.buildDialOptions()

	// 3. 建立连接（兼容旧版 gRPC，带超时控制）
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	newConn, err := grpc.DialContext(ctx, addr, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("dial grpc service %s failed: %w", serviceName, err)
	}

	// 4. 存入缓存
	f.conns[serviceName] = newConn
	return newConn, nil
}

// getHealthyInstance 从 Consul 查询首个通过健康检查的服务实例
// 后续可扩展轮询、随机、权重等负载策略
func (f *RpcClientFactory) getHealthyInstance(serviceName string) (string, error) {
	if global.Consul == nil {
		return "", fmt.Errorf("consul client not initialized")
	}

	// 只查询健康检查通过的实例
	services, _, err := global.Consul.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return "", err
	}
	if len(services) == 0 {
		return "", fmt.Errorf("no healthy instance available for service: %s", serviceName)
	}

	svc := services[0].Service
	return fmt.Sprintf("%s:%d", svc.Address, svc.Port), nil
}

// buildDialOptions 统一构建客户端拨号配置
// 新增拦截器、通用配置只改这一处，全服务自动生效
func (f *RpcClientFactory) buildDialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithInsecure(), // 开发环境无 TLS，生产可替换为证书配置
		grpc.WithUnaryInterceptor(
			interceptor.ChainUnaryClient(
				interceptor.RetryUnaryClientInterceptor(),      // 1. 自动重试
				interceptor.TraceUnaryClientInterceptor(),      // 2. 链路透传
				interceptor.PrometheusUnaryClientInterceptor(), // 3. 指标采集
				interceptor.SentinelUnaryClientInterceptor(),   // 4. 客户端熔断
				interceptor.UnaryClientInterceptor(),           // 5. 调用日志
			),
		),
	}
}

// CloseAll 关闭所有缓存的连接，服务优雅退出时调用
func (f *RpcClientFactory) CloseAll() {
	f.mu.Lock()
	defer f.mu.Unlock()

	for name, conn := range f.conns {
		_ = conn.Close()
		global.Logger.Info(fmt.Sprintf("grpc client connection closed: %s", name))
	}
	// 清空缓存
	f.conns = make(map[string]*grpc.ClientConn)
}
