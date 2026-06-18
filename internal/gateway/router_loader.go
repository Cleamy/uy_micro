package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	"github.com/Cleamy/uy_micro/config"
	"github.com/Cleamy/uy_micro/global"
	"github.com/Cleamy/uy_micro/internal/gateway/filter"
	"github.com/Cleamy/uy_micro/internal/registry"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 加载业务转发路由
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

		// 反向代理核心逻辑
		proxyHandler := func(c *gin.Context) {
			var target *url.URL
			var err error

			// 服务发现模式：复用框架负载均衡能力
			if routeCopy.ServiceName != "" && global.Consul != nil {
				httpTarget, err := registry.GetServiceHttpTarget(routeCopy.ServiceName, true)
				if err != nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": "target service unavailable"})
					c.Abort()
					return
				}
				target, err = url.Parse("http://" + httpTarget)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "parse target address failed"})
					c.Abort()
					return
				}
			} else {
				// 直连地址模式
				target, err = url.Parse(routeCopy.TargetURL)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "invalid target url config"})
					c.Abort()
					return
				}
			}

			// 带超时的转发客户端，避免下游阻塞拖垮网关
			proxy := httputil.NewSingleHostReverseProxy(target)
			proxy.Transport = &http.Transport{
				ResponseHeaderTimeout: 30 * time.Second,
				IdleConnTimeout:       90 * time.Second,
			}

			// 后端异常统一返回 JSON，替代默认 502 页面
			proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
				global.Logger.Warn("gateway proxy backend error",
					zap.String("service", routeCopy.ServiceName),
					zap.String("path", r.URL.Path),
					zap.Error(err))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadGateway)
				_ = json.NewEncoder(w).Encode(gin.H{"code": 502, "msg": "backend service unavailable"})
			}

			// 请求头透传：用户 ID、请求 ID 向下游传递
			if uid, exists := c.Get("uid"); exists {
				switch v := uid.(type) {
				case string:
					c.Request.Header.Set("X-User-Id", v)
				case uint64:
					c.Request.Header.Set("X-User-Id", strconv.FormatUint(v, 10))
				}
			}
			if requestID := c.GetString("request_id"); requestID != "" {
				c.Request.Header.Set("X-Request-ID", requestID)
			}

			// 修正前缀剥离逻辑
			if routeCopy.StripPrefix {
				c.Request.URL.Path = "/" + c.Param("filepath")
			}
			proxy.ServeHTTP(c.Writer, c.Request)
		}

		// 路由追加通配符，支持前缀剥离与子路径转发
		fullPath := routeCopy.Path + "/*filepath"
		rootGroup.Any(fullPath, append(middlewares, proxyHandler)...)
	}
}
