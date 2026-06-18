package middleware

import (
	"strconv"
	"time"

	"github.com/Cleamy/uy_micro/pkg/metrics"

	"github.com/gin-gonic/gin"
)

// PrometheusMiddleware HTTP 请求全量指标采集
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		method := c.Request.Method
		path := c.FullPath()

		c.Next()

		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())

		metrics.HttpRequestTotal.WithLabelValues(method, path, statusCode).Inc()
		metrics.HttpRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}
