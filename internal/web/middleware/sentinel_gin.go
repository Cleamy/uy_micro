package middleware

import (
	"net/http"

	"github.com/Cleamy/uy_micro/pkg/errcode"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/gin-gonic/gin"
)

// SentinelLimiter Gin 全局限流中间件
// 资源命名规则：METHOD:PATH  例：GET:/api/v1/user
func SentinelLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceName := c.Request.Method + ":" + c.FullPath()

		// 进入 Sentinel 流量校验
		entry, err := sentinel.Entry(
			resourceName,
			sentinel.WithResourceType(base.ResTypeWeb),
			sentinel.WithTrafficType(base.Inbound),
		)

		if err != nil {
			// 触发限流/熔断，直接返回标准错误
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": errcode.CodeTooManyRequest,
				"msg":  "too many requests, please try again later",
			})
			return
		}

		defer entry.Exit()
		c.Next()
	}
}
