package interceptor

import (
	"context"
	"time"

	"github.com/Cleamy/uy_micro/pkg/metrics"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// PrometheusUnaryServerInterceptor gRPC 服务端指标采集
func PrometheusUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		method := info.FullMethod

		resp, err := handler(ctx, req)

		duration := time.Since(start).Seconds()
		code := status.Code(err).String()

		metrics.GrpcServerRequestTotal.WithLabelValues(method, code).Inc()
		metrics.GrpcServerRequestDuration.WithLabelValues(method).Observe(duration)

		return resp, err
	}
}

// PrometheusUnaryClientInterceptor gRPC 客户端指标采集
func PrometheusUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		target := cc.Target()

		err := invoker(ctx, method, req, reply, cc, opts...)

		duration := time.Since(start).Seconds()
		code := status.Code(err).String()

		metrics.GrpcClientRequestTotal.WithLabelValues(method, target, code).Inc()
		metrics.GrpcClientRequestDuration.WithLabelValues(method, target).Observe(duration)

		return err
	}
}
