package middleware

import (
	"uy_micro/global"
	"uy_micro/pkg/tracepkg"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const traceHeaderKey = "X-Trace-ID"

// TraceMiddleware Gin 全链路追踪中间件
// 自动从请求头提取 Trace，没有则新建，注入上下文并写入响应头
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tracer := global.Tracer
		if tracer == nil {
			c.Next()
			return
		}

		ctx := c.Request.Context()

		// 创建 Span（当前为简化版：入口新建链路；后续接入 W3C Propagator 可自动还原上游链路）
		spanName := c.Request.Method + ":" + c.FullPath()
		ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		// 设置基础属性
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.path", c.FullPath()),
			attribute.String("http.client_ip", c.ClientIP()),
		)

		// 将 Trace 注入 Gin 上下文
		c.Request = c.Request.WithContext(ctx)

		// 响应头返回 TraceID
		c.Header(traceHeaderKey, tracepkg.FromContext(ctx))

		c.Next()

		// 标记异常
		if len(c.Errors) > 0 {
			span.SetStatus(codes.Error, c.Errors.String())
		}
	}
}
