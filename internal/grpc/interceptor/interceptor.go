package interceptor

import (
	"context"
	"time"
	"github.com/Cleamy/uy_micro/global"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor gRPC 服务端一元拦截器
// 能力：Panic 恢复 + 调用日志 + 全局超时控制
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()
		method := info.FullMethod

		// Panic 恢复，避免单个请求拖垮整个服务
		defer func() {
			if r := recover(); r != nil {
				global.Logger.Error("grpc server panic",
					zap.String("method", method),
					zap.Any("panic", r),
					zap.Stack("stack"))
				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		// 全局默认超时 30 秒
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		// 执行业务逻辑
		resp, err = handler(ctx, req)

		// 调用日志
		cost := time.Since(start)
		if err != nil {
			global.Logger.Warn("grpc server call fail",
				zap.String("method", method),
				zap.Duration("cost", cost),
				zap.Error(err))
		} else {
			global.Logger.Debug("grpc server call success",
				zap.String("method", method),
				zap.Duration("cost", cost))
		}
		return resp, err
	}
}
