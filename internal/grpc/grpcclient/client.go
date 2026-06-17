package grpcclient

import (
	"fmt"

	"uy_micro/global"
	"uy_micro/internal/registry"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// RpcClientFactory GRPC 客户端统一工厂
type RpcClientFactory struct{}

func NewRpcClientFactory() *RpcClientFactory {
	return &RpcClientFactory{}
}

// GetConn 根据服务名获取 grpc 连接与关闭回调
func (f *RpcClientFactory) GetConn(serviceName string) (*grpc.ClientConn, func(), error) {
	target, err := registry.GetServiceGrpcTargetDefault(serviceName, true)
	if err != nil {
		return nil, nil, err
	}

	global.Logger.Info("[GRPC Client] start dial remote service",
		zap.String("service", serviceName),
		zap.String("target", target))

	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		global.Logger.Error("[GRPC Client] dial failed",
			zap.String("service", serviceName),
			zap.String("target", target),
			zap.Error(err))
		return nil, nil, fmt.Errorf("dial %s err: %w", target, err)
	}

	global.Logger.Info("[GRPC Client] dial success",
		zap.String("service", serviceName),
		zap.String("target", target))

	clean := func() {
		_ = conn.Close()
		global.Logger.Info("[GRPC Client] connection closed",
			zap.String("service", serviceName),
			zap.String("target", target))
	}

	return conn, clean, nil
}

// GetClient 泛型封装：一行拿到 PB 生成的客户端，业务极致简化
func GetClient[T any](serviceName string, newPbClient func(grpc.ClientConnInterface) T) (T, func(), error) {
	factory := NewRpcClientFactory()
	conn, clean, err := factory.GetConn(serviceName)
	if err != nil {
		var zero T
		return zero, nil, err
	}
	return newPbClient(conn), clean, nil
}
