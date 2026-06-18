package filter

import (
	"time"
	"github.com/Cleamy/uy_micro/global"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GatewayLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		c.Next()
		cost := time.Since(start)
		global.Logger.Info("gateway request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.String("ip", c.ClientIP()),
			zap.Duration("cost", cost),
		)
	}
}
