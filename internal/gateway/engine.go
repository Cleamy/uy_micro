package gateway

import (
	"uy_micro/config"
	"uy_micro/internal/gateway/filter"

	"github.com/gin-gonic/gin"
)

// 初始化网关

func Init(cfg *config.GatewayConfig) (*gin.Engine, error) {
	if !cfg.Enable {
		return nil, nil
	}
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	// 全局过滤器固定顺序：跨域 → IP黑名单 → 网关日志
	if cfg.Cors.Enable {
		engine.Use(filter.CorsMiddleware(&cfg.Cors))
	}
	if len(cfg.IPBlacklist) > 0 {
		engine.Use(filter.IPBlacklistMiddleware(cfg.IPBlacklist))
	}
	engine.Use(filter.GatewayLogMiddleware())


	// 加载业务转发路由
	loadRoutes(engine, cfg)
	return engine, nil
}
