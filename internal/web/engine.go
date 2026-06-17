package web

import (
	"os"
	"uy_micro/config"
	"uy_micro/internal/web/middleware"

	"github.com/gin-gonic/gin"
)

func Init(cfg *config.WebConfig) (*gin.Engine, error) {
	if !cfg.Enable {
		return nil, nil
	}
	gin.SetMode(cfg.Mode)
	engine := gin.New()
	// 自定义日志中间件：过滤 /health，其余正常打印
	engine.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/ping" {
			c.Next()
			return
		}
		gin.LoggerWithWriter(os.Stdout)(c)
	})
	engine.Use(gin.Logger(), gin.Recovery())

	// 自动内置健康检测，无需用户手动编写
	middleware.RegisterHealthRoute(engine)
	return engine, nil
}
