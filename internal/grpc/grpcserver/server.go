package grpcserver

import (
	"uy_micro/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Init 初始化 gRPC 服务端
func Init(cfg *config.GrpcConfig) (*grpc.Server, error) {
	if !cfg.Enable {
		return nil, nil
	}

	s := grpc.NewServer()

	// 开启 gRPC 反射（便于 grpcurl 调试）
	if cfg.Reflection {
		reflection.Register(s)
	}

	return s, nil
}
