package bootstrap

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"uy_micro/config"
	"uy_micro/global"
	"uy_micro/internal/gateway"
	grpcclient "uy_micro/internal/grpc/grpcclient"
	grpcsvr "uy_micro/internal/grpc/grpcserver"
	"uy_micro/internal/observability"
	"uy_micro/internal/registry"
	"uy_micro/internal/storage/database"
	rediscli "uy_micro/internal/storage/redis" // 修正：正确路径 rediscli
	"uy_micro/internal/traffic"
	"uy_micro/internal/web"

	"go.uber.org/zap"
)

var PostBootHooks []func()

// Bootstrap 按顺序初始化所有组件
func Bootstrap() error {
	// 1. 加载配置
	if err := config.LoadConfig(); err != nil {
		return fmt.Errorf("load config failed: %w", err)
	}
	global.Config = config.AppCfg

	// 2. 初始化日志（最优先）
	log, err := observability.InitLogger(&config.AppCfg.Logger)
	if err != nil {
		return fmt.Errorf("init logger failed: %w", err)
	}
	global.Logger = log
	global.Logger.Info("=== uy_micro framework starting ===")

	// 3. 初始化可观测性（链路追踪）
	if config.AppCfg.Observability.Tracer.Enable {
		tracer, err := observability.InitTracer(&config.AppCfg.Observability.Tracer)
		if err != nil {
			return fmt.Errorf("init tracer failed: %w", err)
		}
		global.Tracer = tracer
		global.Logger.Info("tracer initialized")
	}

	// 4. 初始化流量治理
	if config.AppCfg.Sentinel.Enable {
		if err := traffic.InitSentinel(&config.AppCfg.Sentinel); err != nil {
			return fmt.Errorf("init sentinel failed: %w", err)
		}
		global.Logger.Info("sentinel initialized")
	}

	// 5. 初始化存储中间件
	if config.AppCfg.Database.Enable {
		db, err := database.Init(&config.AppCfg.Database)
		if err != nil {
			return fmt.Errorf("init database failed: %w", err)
		}
		global.DB = db
		global.Logger.Info("database initialized")
	}

	if config.AppCfg.Redis.Enable {
		rdb, err := rediscli.Init(&config.AppCfg.Redis) // 修正：调用 rediscli.Init
		if err != nil {
			return fmt.Errorf("init redis failed: %w", err)
		}
		global.Redis = rdb
		global.Logger.Info("redis initialized")
	}

	// 6. 初始化 Web 服务
	if config.AppCfg.Web.Enable {
		engine, err := web.Init(&config.AppCfg.Web)
		if err != nil {
			return fmt.Errorf("init web failed: %w", err)
		}
		global.Web = engine
		global.Logger.Info("web engine initialized")
	}

	// 7. 初始化 gRPC 服务
	if config.AppCfg.Grpc.Enable {
		grpcSrv, err := grpcsvr.Init(&config.AppCfg.Grpc)
		if err != nil {
			return fmt.Errorf("init grpc failed: %w", err)
		}
		global.Grpc = grpcSrv
		global.Logger.Info("grpc server initialized")
	}

	// 8. 初始化网关
	if config.AppCfg.Gateway.Enable {
		gatewayEngine, err := gateway.Init(&config.AppCfg.Gateway)
		if err != nil {
			return fmt.Errorf("init gateway failed: %w", err)
		}
		global.Gateway = gatewayEngine
		global.Logger.Info("api gateway initialized")
	}

	// 9. 服务注册发现（最后注册，确保服务已就绪）
	if config.AppCfg.Consul.Enable {
		client, err := registry.Init(&config.AppCfg.Consul)
		if err != nil {
			return fmt.Errorf("init consul failed: %w", err)
		}
		global.Consul = client
		global.Logger.Info("consul client initialized, service registered")

		// 10. 初始化 GRPC 客户端工厂（依赖 Consul 就绪）
		global.RpcFactory = grpcclient.NewRpcClientFactory()
		global.Logger.Info("grpc client factory initialized")
	}

	// 执行用户注册的启动后钩子
	for _, fn := range PostBootHooks {
		fn()
	}

	global.Logger.Info("=== all components initialized ===")
	return nil
}

// Run 启动所有服务并阻塞监听信号
func Run() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 启动 Web 服务
	var httpSrv *http.Server
	if global.Web != nil {
		httpSrv = &http.Server{
			Addr:         fmt.Sprintf(":%d", config.AppCfg.Web.Port),
			Handler:      global.Web,
			ReadTimeout:  time.Duration(config.AppCfg.Web.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(config.AppCfg.Web.WriteTimeout) * time.Second,
		}
		go func() {
			global.Logger.Info(fmt.Sprintf("http server running on :%d", config.AppCfg.Web.Port))
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				global.Logger.Fatal("http server error", zap.Error(err))
			}
		}()
	}

	// 启动网关服务
	var gatewaySrv *http.Server
	if global.Gateway != nil {
		gatewaySrv = &http.Server{
			Addr:    fmt.Sprintf(":%d", config.AppCfg.Gateway.Port),
			Handler: global.Gateway,
		}
		go func() {
			global.Logger.Info(fmt.Sprintf("api gateway running on :%d", config.AppCfg.Gateway.Port))
			if err := gatewaySrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				global.Logger.Fatal("gateway server error", zap.Error(err))
			}
		}()
	}

	// 启动 gRPC 服务
	var grpcListener net.Listener
	if global.Grpc != nil {
		var err error
		grpcListener, err = net.Listen("tcp", fmt.Sprintf(":%d", config.AppCfg.Grpc.Port))
		if err != nil {
			global.Logger.Fatal("grpc listen failed", zap.Error(err))
		}
		go func() {
			global.Logger.Info(fmt.Sprintf("grpc server running on :%d", config.AppCfg.Grpc.Port))
			if err := global.Grpc.Serve(grpcListener); err != nil {
				global.Logger.Fatal("grpc server error", zap.Error(err))
			}
		}()
	}

	<-quit
	global.Logger.Info("=== shutting down service ===")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. 先注销服务（ID 与注册时保持一致）
	if global.Consul != nil {
		svcID := fmt.Sprintf("%s-web-%d", config.AppCfg.App.Name, config.AppCfg.Web.Port)
		_ = global.Consul.Agent().ServiceDeregister(svcID)
		global.Logger.Info("consul service deregistered", zap.String("service_id", svcID))
	}

	// 2. 关闭 HTTP 与网关
	if httpSrv != nil {
		_ = httpSrv.Shutdown(ctx)
	}
	if gatewaySrv != nil {
		_ = gatewaySrv.Shutdown(ctx)
	}

	// 3. 关闭 gRPC
	if global.Grpc != nil {
		global.Grpc.GracefulStop()
	}

	// 4. 关闭数据库
	if global.DB != nil {
		sqlDB, _ := global.DB.DB()
		_ = sqlDB.Close()
	}

	// 5. 关闭 Redis
	if global.Redis != nil {
		_ = global.Redis.Close()
	}

	global.Logger.Info("=== service exited gracefully ===")
}
