package filter

import (
	"uy_micro/config"

	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CorsMiddleware(cfg *config.CorsConfig) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           time.Second * time.Duration(cfg.MaxAge),
	})
}
