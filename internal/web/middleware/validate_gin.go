package middleware

import (
	"github.com/Cleamy/uy_micro/pkg/errcode"
	"github.com/Cleamy/uy_micro/pkg/response"
	"github.com/Cleamy/uy_micro/pkg/validatorx"

	"github.com/gin-gonic/gin"
)

// ValidateMiddleware 请求参数自动校验中间件
// 配合 c.ShouldBind 自动捕获校验错误，统一返回标准错误结构
func ValidateMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		// 捕获绑定参数产生的校验错误
		if len(c.Errors) == 0 {
			return
		}
		firstErr := c.Errors[0].Err
		// 翻译校验提示
		msg := validatorx.TranslateErr(firstErr)
		// 使用框架统一错误返回
		response.FailWithCode(c, errcode.CodeInvalidParam, msg)
		// 终止后续逻辑
		c.Abort()
	}
}
