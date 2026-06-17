package middleware

import (
	"uy_micro/global"

	"github.com/gin-gonic/gin"
)

// RegisterHealthRoute 自动注册内置健康接口 /ping /health
func RegisterHealthRoute(r *gin.Engine) {
	// 简易存活检测
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"msg": "pong"})
	})
	// 完整健康检测（db、redis依赖探测）
	r.GET("/health", func(c *gin.Context) {
		res := gin.H{
			"status":  "ok",
			"service": global.Config.App.Name,
			"env":     global.Config.App.Env,
		}
		// 数据库检测
		if global.DB != nil {
			sqlDB, err := global.DB.DB()
			if err == nil && sqlDB.Ping() == nil {
				res["database"] = "ok"
			} else {
				res["database"] = "error"
				res["status"] = "unhealthy"
			}
		} else {
			res["database"] = "disable"
		}
		// Redis检测
		if global.Redis != nil {
			_, err := global.Redis.Ping(c).Result()
			if err == nil {
				res["redis"] = "ok"
			} else {
				res["redis"] = "error"
				res["status"] = "unhealthy"
			}
		} else {
			res["redis"] = "disable"
		}
		c.JSON(200, res)
	})
}
