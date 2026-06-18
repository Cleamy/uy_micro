package interceptor

import (
	"context"

	"uy_micro/pkg/errcode"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"google.golang.org/grpc"
)

// SentinelUnaryServerInterceptor gRPC 服务端限流/熔断拦截器
// 资源命名规则：全方法名  例：/role.v1.RoleService/GetRoleByID
func SentinelUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		entry, blockErr := sentinel.Entry(
			info.FullMethod,
			sentinel.WithResourceType(base.ResTypeRPC),
			sentinel.WithTrafficType(base.Inbound),
		)

		if blockErr != nil {
			// 触发限流，返回 gRPC 标准状态，对齐统一错误码
			return nil, errcode.ToGRPCStatus(errcode.New(errcode.CodeTooManyRequest, "rpc service rate limited"))
		}

		defer entry.Exit()

		resp, callErr := handler(ctx, req)
		if callErr != nil {
			// 业务异常，标记到 entry 供熔断器统计异常比例
			entry.SetError(callErr)
		}
		return resp, callErr
	}
}

// SentinelUnaryClientInterceptor gRPC 客户端熔断拦截器
// 资源命名规则：全方法名，用于下游服务异常时客户端侧熔断
func SentinelUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		entry, blockErr := sentinel.Entry(
			method,
			sentinel.WithResourceType(base.ResTypeRPC),
			sentinel.WithTrafficType(base.Outbound),
		)

		if blockErr != nil {
			// 客户端熔断打开，直接返回不可用错误
			return errcode.ToGRPCStatus(errcode.New(errcode.CodeServiceUnavail, "rpc service circuit breaker open"))
		}

		defer entry.Exit()

		callErr := invoker(ctx, method, req, reply, cc, opts...)
		if callErr != nil {
			// 调用异常，标记错误供熔断器统计
			entry.SetError(callErr)
		}
		return callErr
	}
}
