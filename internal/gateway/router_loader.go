package gateway

import (
	"net/http/httputil"
	"net/url"
	"uy_micro/config"
	"uy_micro/global"
	"uy_micro/internal/gateway/filter"
	"uy_micro/internal/registry"

	"github.com/gin-gonic/gin"
)

// 配置路由
func loadRoutes(engine *gin.Engine, cfg *config.GatewayConfig) {
	rootGroup := engine.Group(cfg.ContextPath)

	for _, route := range cfg.Routes {
		routeCopy := route
		middlewares := make([]gin.HandlerFunc, 0)

		// 路由级鉴权
		if routeCopy.AuthEnable && cfg.Auth.Enable {
			middlewares = append(middlewares, filter.AuthMiddleware(&cfg.Auth))
		}
		// 路由级限流
		if routeCopy.RateLimit {
			middlewares = append(middlewares, filter.SentinelLimitMiddleware(routeCopy.ID))
		}

		// 反向代理处理逻辑
		proxyHandler := func(c *gin.Context) {
			var target *url.URL
			var err error

			// 服务发现模式：复用框架负载均衡能力
			if routeCopy.ServiceName != "" && global.Consul != nil {
				httpTarget, err := registry.GetServiceHttpTarget(routeCopy.ServiceName, true)
				if err != nil {
					c.JSON(503, gin.H{"code": 503, "msg": "target service unavailable"})
					c.Abort()
					return
				}
				target, err = url.Parse("http://" + httpTarget)
				if err != nil {
					c.JSON(500, gin.H{"code": 500, "msg": "parse target address failed"})
					c.Abort()
					return
				}
			} else {
				// 直连地址模式
				target, err = url.Parse(routeCopy.TargetURL)
				if err != nil {
					c.JSON(500, gin.H{"code": 500, "msg": "invalid target url config"})
					c.Abort()
					return
				}
			}

			proxy := httputil.NewSingleHostReverseProxy(target)
			// 修正前缀剥离逻辑
			if routeCopy.StripPrefix {
				c.Request.URL.Path = "/" + c.Param("filepath")
			}
			proxy.ServeHTTP(c.Writer, c.Request)
		}

		// 路由追加通配符，支持前缀剥离
		fullPath := routeCopy.Path + "/*filepath"
		rootGroup.Any(fullPath, append(middlewares, proxyHandler)...)
	}
}
