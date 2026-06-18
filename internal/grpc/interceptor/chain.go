package interceptor

import (
	"context"

	"google.golang.org/grpc"
)

// ChainUnaryServer 串联多个服务端一元拦截器，按传入顺序从前到后执行
func ChainUnaryServer(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	if len(interceptors) == 0 {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	}
	if len(interceptors) == 1 {
		return interceptors[0]
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 构建嵌套调用链
		chainHandler := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			cur := interceptors[i]
			next := chainHandler
			chainHandler = func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return cur(currentCtx, currentReq, info, next)
			}
		}
		return chainHandler(ctx, req)
	}
}

// ChainUnaryClient 串联多个客户端一元拦截器
func ChainUnaryClient(interceptors ...grpc.UnaryClientInterceptor) grpc.UnaryClientInterceptor {
	if len(interceptors) == 0 {
		return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
	}
	if len(interceptors) == 1 {
		return interceptors[0]
	}

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		chainInvoker := invoker
		for i := len(interceptors) - 1; i >= 0; i-- {
			cur := interceptors[i]
			next := chainInvoker
			chainInvoker = func(currentCtx context.Context, currentMethod string, currentReq, currentReply interface{}, currentCC *grpc.ClientConn, currentOpts ...grpc.CallOption) error {
				return cur(currentCtx, currentMethod, currentReq, currentReply, currentCC, next, currentOpts...)
			}
		}
		return chainInvoker(ctx, method, req, reply, cc, opts...)
	}
}
