package web

import (
	"os"
	"github.com/Cleamy/uy_micro/config"
	"github.com/Cleamy/uy_micro/internal/health"
	"github.com/Cleamy/uy_micro/internal/web/middleware"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Init(cfg *config.WebConfig) (*gin.Engine, error) {
	if !cfg.Enable {
		return nil, nil
	}
	gin.SetMode(cfg.Mode)
	engine := gin.New()
	// 1. 全局panic兜底
	engine.Use(middleware.RecoveryMiddleware())
	// 2. 全链路追踪
	engine.Use(middleware.TraceMiddleware())
	// 3. prom指标采集
	engine.Use(middleware.PrometheusMiddleware())
	// 4. 参数校验拦截
	engine.Use(middleware.ValidateMiddleware())
	// 5. 限流熔断
	engine.Use(middleware.SentinelLimiter())

	// 暴露 Prometheus 指标端点
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	engine.GET("/health/live", health.LivenessCheck)
	engine.GET("/health/ready", health.ReadinessCheck)
	engine.GET("/health", health.ReadinessCheck) // 兼容旧版

	// 自定义日志中间件：过滤 /health、/ping，其余正常打印
	engine.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/health" || path == "/ping" {
			c.Next()
			return
		}
		gin.LoggerWithWriter(os.Stdout)(c)
	})

	engine.Use(middleware.RecoveryMiddleware())

	return engine, nil
}
