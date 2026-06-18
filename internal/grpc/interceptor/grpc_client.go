package interceptor

import (
	"context"
	"time"

	"github.com/Cleamy/uy_micro/global"
	"github.com/Cleamy/uy_micro/pkg/errcode"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryClientInterceptor gRPC 客户端基础拦截器
// 能力：统一调用日志 + 上下文超时自动透传
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()

		global.Logger.Debug("grpc client call start",
			zap.String("method", method),
			zap.String("target", cc.Target()))

		// 执行实际调用，context 自动透传超时、Trace 等信息
		err := invoker(ctx, method, req, reply, cc, opts...)

		// 调用结果日志，自动转换为框架统一错误码
		cost := time.Since(start)
		if err != nil {
			frameErr := errcode.FromGRPCStatus(err)
			global.Logger.Warn("grpc client call fail",
				zap.String("method", method),
				zap.String("target", cc.Target()),
				zap.Duration("cost", cost),
				zap.Int("code", frameErr.Code),
				zap.Error(err))
		} else {
			global.Logger.Debug("grpc client call success",
				zap.String("method", method),
				zap.String("target", cc.Target()),
				zap.Duration("cost", cost))
		}

		return err
	}
}

// 默认重试策略配置
const (
	defaultMaxRetryTimes = 3                      // 最大重试次数
	defaultRetryBackoff  = 100 * time.Millisecond // 初始退避间隔
)

// 可重试的 gRPC 状态码（仅网络/服务不可用类错误，重试无业务副作用）
var retryableCodes = map[codes.Code]bool{
	codes.Unavailable:       true, // 服务不可用/节点掉线
	codes.DeadlineExceeded:  true, // 调用超时
	codes.ResourceExhausted: true, // 下游限流
}

// RetryUnaryClientInterceptor gRPC 客户端自动重试拦截器
// 指数退避策略，仅对非业务错误重试，上下文取消时立即终止
func RetryUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var lastErr error
		backoff := defaultRetryBackoff

		for i := 0; i < defaultMaxRetryTimes; i++ {
			// 先检查上下文是否已取消/超时
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			lastErr = invoker(ctx, method, req, reply, cc, opts...)
			if lastErr == nil {
				return nil
			}

			// 判断错误类型，业务错误直接返回不重试
			st, ok := status.FromError(lastErr)
			if !ok || !retryableCodes[st.Code()] {
				return lastErr
			}

			// 非最后一次重试，等待退避时间
			if i < defaultMaxRetryTimes-1 {
				global.Logger.Debug("grpc client retrying",
					zap.String("method", method),
					zap.Int("retry_count", i+1),
					zap.Error(lastErr))

				timer := time.NewTimer(backoff)
				select {
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()
				case <-timer.C:
					backoff *= 2 // 指数退避，避免雪崩
				}
			}
		}

		// 重试全部耗尽，返回最后一次错误
		return lastErr
	}
}
