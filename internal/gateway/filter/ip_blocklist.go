package filter

import (
	"net/http"
	"github.com/Cleamy/uy_micro/global"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func IPBlacklistMiddleware(blackIP []string) gin.HandlerFunc {
	ipMap := make(map[string]bool)
	for _, ip := range blackIP {
		ipMap[ip] = true
	}
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		if ipMap[clientIP] {
			global.Logger.Warn("block black ip access", zap.String("ip", clientIP))
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "ip forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}
