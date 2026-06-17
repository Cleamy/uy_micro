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

	// 自定义日志中间件：过滤健康接口
	engine.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/health" || path == "/ping" {
			c.Next()
			return
		}
		gin.LoggerWithWriter(os.Stdout)(c)
	})
	// 只保留崩溃恢复，移除重复的 gin.Logger()
	engine.Use(gin.Recovery())

	middleware.RegisterHealthRoute(engine)
	return engine, nil
}
