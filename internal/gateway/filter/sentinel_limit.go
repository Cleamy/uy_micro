package filter

import (
	"net/http"
	"uy_micro/global"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SentinelLimitMiddleware 单路由限流中间件
func SentinelLimitMiddleware(resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		entry, err := sentinel.Entry(resource, sentinel.WithResourceType(base.ResTypeWeb))
		if err != nil {
			global.Logger.Warn("request blocked by sentinel limit", zap.String("resource", resource))
			c.JSON(http.StatusTooManyRequests, gin.H{"code": 429, "msg": "too many requests"})
			c.Abort()
			return
		}
		defer entry.Exit()
		c.Next()
	}
}
