package filter

import (
	"net/http"
	"strings"
	"github.com/Cleamy/uy_micro/config"
	"github.com/Cleamy/uy_micro/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(cfg *config.GatewayAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 匹配免登录路径直接放行
		path := c.Request.URL.Path
		skip := false
		for _, p := range cfg.ExcludePaths {
			if strings.HasPrefix(path, p) {
				skip = true
				break
			}
		}
		if skip {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "missing authorization token"})
			c.Abort()
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwt.ParseToken(token, cfg.JwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "invalid token"})
			c.Abort()
			return
		}
		// 用户ID存入上下文供下游使用
		c.Set("uid", claims.UserID)
		c.Next()
	}
}
