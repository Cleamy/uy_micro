package grpcserver

import (
	"uy_micro/config"

	"uy_micro/internal/grpc/interceptor"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Init 初始化 gRPC 服务端
func Init(cfg *config.GrpcConfig) (*grpc.Server, error) {
	if !cfg.Enable {
		return nil, nil
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			interceptor.ChainUnaryServer(
				interceptor.TraceUnaryServerInterceptor(),      // 1. 链路追踪
				interceptor.PrometheusUnaryServerInterceptor(), // 2. 指标采集
				interceptor.SentinelUnaryServerInterceptor(),   // 3. 限流熔断
				interceptor.UnaryServerInterceptor(),           // 4. Panic兜底/日志/超时
			),
		),
	)

	// 开启 gRPC 反射（便于 grpcurl 调试）
	if cfg.Reflection {
		reflection.Register(s)
	}

	return s, nil
}
