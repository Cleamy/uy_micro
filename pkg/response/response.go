package response

import (
	"uy_micro/pkg/errcode"

	"github.com/gin-gonic/gin"
)

// Result 统一响应结构
type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// OK 成功响应
func OK(c *gin.Context, data interface{}) {
	c.JSON(200, Result{
		Code: errcode.CodeSuccess,
		Msg:  "success",
		Data: data,
	})
}

// Fail 错误响应，自动从 error 中解析错误码
func Fail(c *gin.Context, err error) {
	e := errcode.FromError(err)
	httpStatus := errcode.ToHTTPStatus(e.Code)
	c.JSON(httpStatus, Result{
		Code: e.Code,
		Msg:  e.Msg,
	})
}

// FailWithCode 直接按错误码返回
func FailWithCode(c *gin.Context, code int, msg string) {
	httpStatus := errcode.ToHTTPStatus(code)
	c.JSON(httpStatus, Result{
		Code: code,
		Msg:  msg,
	})
}
