package gateway

import (
	"net/http/httputil"
	"net/url"
	"uy_micro/config"
	"uy_micro/global"
	"uy_micro/internal/gateway/filter"

	"github.com/gin-gonic/gin"
)

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
			// 服务发现模式
			if routeCopy.ServiceName != "" && global.Consul != nil {
				services, _, err := global.Consul.Health().Service(routeCopy.ServiceName, "", true, nil)
				if err != nil || len(services) == 0 {
					c.JSON(503, gin.H{"code": 503, "msg": "target service unavailable"})
					c.Abort()
					return
				}
				svc := services[0].Service
				target, _ = url.Parse("http://" + svc.Address + ":" + string(rune(svc.Port)))
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
			if routeCopy.StripPrefix {
				c.Request.URL.Path = c.Param("filepath")
			}
			proxy.ServeHTTP(c.Writer, c.Request)
		}

		// 注册路由匹配所有方法
		rootGroup.Any(routeCopy.Path, append(middlewares, proxyHandler)...)
	}
}
