package tracepkg

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

const traceIDKey = "trace_id"

// FromContext 从上下文提取 TraceID，不存在返回空字符串
func FromContext(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return spanCtx.TraceID().String()
	}
	return ""
}

// InjectContext 将 Trace 信息注入上下文
func InjectContext(ctx context.Context, spanCtx trace.SpanContext) context.Context {
	return trace.ContextWithSpanContext(ctx, spanCtx)
}
