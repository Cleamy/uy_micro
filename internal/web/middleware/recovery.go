package middleware

import (
	"net/http"
	"uy_micro/global"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RecoveryMiddleware 自定义panic恢复，统一返回JSON格式错误
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				global.Logger.Error("web request panic",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.Stack("stack"))
				c.JSON(http.StatusInternalServerError, gin.H{
					"code": 500,
					"msg":  "internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
