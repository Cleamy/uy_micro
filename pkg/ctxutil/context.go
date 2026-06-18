package ctxutil

import (
	"context"

	"uy_micro/pkg/tracepkg"

	"github.com/gin-gonic/gin"
)

// 私有自定义 key 类型，彻底避免外部包 key 命名冲突
type ctxKey string

const (
	traceIDKey   ctxKey = "trace_id"
	requestIDKey ctxKey = "request_id"
	userIDKey    ctxKey = "user_id"
)

// ==================== TraceID 存取 ====================

// SetTraceID 注入 TraceID 到上下文
func SetTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID 从上下文提取 TraceID
// 优先从 OpenTelemetry 链路上下文取，兜底自定义上下文
func GetTraceID(ctx context.Context) string {
	if id := tracepkg.FromContext(ctx); id != "" {
		return id
	}
	if v, ok := ctx.Value(traceIDKey).(string); ok {
		return v
	}
	return ""
}

// GetTraceIDFromGin 从 Gin 上下文快捷提取 TraceID
func GetTraceIDFromGin(c *gin.Context) string {
	return GetTraceID(c.Request.Context())
}

// ==================== RequestID 存取 ====================

// SetRequestID 注入请求 ID 到上下文
func SetRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID 从上下文提取请求 ID
func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// GetRequestIDFromGin 从 Gin 上下文快捷提取请求 ID
func GetRequestIDFromGin(c *gin.Context) string {
	return GetRequestID(c.Request.Context())
}

// ==================== 用户身份存取 ====================

// SetUserID 注入用户 ID 到上下文（鉴权中间件调用）
func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID 从上下文提取用户 ID
func GetUserID(ctx context.Context) string {
	if v, ok := ctx.Value(userIDKey).(string); ok {
		return v
	}
	return ""
}

// GetUserIDFromGin 从 Gin 上下文快捷提取用户 ID
func GetUserIDFromGin(c *gin.Context) string {
	return GetUserID(c.Request.Context())
}
