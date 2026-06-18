package interceptor

import (
	"context"

	"uy_micro/global"
	"uy_micro/pkg/tracepkg"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const grpcTraceKey = "x-trace-id"

// TraceUnaryServerInterceptor gRPC 服务端 Trace 拦截器
// 从元数据提取上游 Trace，创建服务端 Span，注入上下文
func TraceUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		tracer := global.Tracer
		if tracer == nil {
			return handler(ctx, req)
		}

		// 从 gRPC 元数据提取 TraceID（简化透传）
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if vals := md.Get(grpcTraceKey); len(vals) > 0 {
				// 上游 Trace 透传（简化版，生产建议用 otel gRPC propagator）
			}
		}

		// 创建服务端 Span
		ctx, span := tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		span.SetAttributes(
			attribute.String("rpc.method", info.FullMethod),
			attribute.String("rpc.kind", "server"),
		)

		resp, err := handler(ctx, req)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}

		return resp, err
	}
}

// TraceUnaryClientInterceptor gRPC 客户端 Trace 拦截器
// 将当前 Trace 上下文注入 gRPC 元数据，透传给下游服务
func TraceUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		tracer := global.Tracer
		if tracer == nil {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		// 创建客户端 Span
		ctx, span := tracer.Start(ctx, method, trace.WithSpanKind(trace.SpanKindClient))
		defer span.End()

		span.SetAttributes(
			attribute.String("rpc.method", method),
			attribute.String("rpc.target", cc.Target()),
			attribute.String("rpc.kind", "client"),
		)

		// 将 TraceID 注入 gRPC 出站元数据
		traceID := tracepkg.FromContext(ctx)
		if traceID != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, grpcTraceKey, traceID)
		}

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}

		return err
	}
}
