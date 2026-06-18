package errcode

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ==================== 框架级通用错误码定义 ====================
// 码值规范：与 HTTP 语义对齐，便于跨协议映射；业务可基于此扩展自有错误码
const (
	CodeSuccess = 200 // 成功

	// 客户端类错误 4xx
	CodeInvalidParam   = 400 // 参数错误
	CodeUnauthorized   = 401 // 未授权
	CodeForbidden      = 403 // 禁止访问
	CodeNotFound       = 404 // 资源不存在
	CodeMethodNotAllow = 405 // 方法不允许
	CodeTooManyRequest = 429 // 请求过于频繁

	// 服务端类错误 5xx
	CodeInternalError  = 500 // 内部错误
	CodeNotImplemented = 501 // 未实现
	CodeServiceUnavail = 503 // 服务不可用
	CodeTimeout        = 504 // 服务超时
)

// 预定义通用错误实例
var (
	Success        = New(CodeSuccess, "success")
	InvalidParam   = New(CodeInvalidParam, "invalid parameter")
	Unauthorized   = New(CodeUnauthorized, "unauthorized")
	Forbidden      = New(CodeForbidden, "forbidden")
	NotFound       = New(CodeNotFound, "resource not found")
	InternalError  = New(CodeInternalError, "internal server error")
	ServiceUnavail = New(CodeServiceUnavail, "service unavailable")
	Timeout        = New(CodeTimeout, "request timeout")
)

// ==================== 统一错误结构体 ====================

// Error 框架统一错误类型，原生实现 error 接口
type Error struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Detail string `json:"detail,omitempty"`
	cause  error  // 原始根因错误，不对外序列化
}

// New 创建错误实例
func New(code int, msg string) *Error {
	return &Error{
		Code: code,
		Msg:  msg,
	}
}

// NewWithDetail 创建带详情的错误实例
func NewWithDetail(code int, msg, detail string) *Error {
	return &Error{
		Code:   code,
		Msg:    msg,
		Detail: detail,
	}
}

// Error 实现标准 error 接口
func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Msg, e.cause)
	}
	if e.Detail != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Msg, e.Detail)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Msg)
}

// WithMsg 替换错误信息，返回新实例
func (e *Error) WithMsg(msg string) *Error {
	newErr := *e
	newErr.Msg = msg
	return &newErr
}

// WithDetail 追加错误详情，返回新实例
func (e *Error) WithDetail(detail string) *Error {
	newErr := *e
	newErr.Detail = detail
	return &newErr
}

// Wrap 包装原始错误，保留根因堆栈
func (e *Error) Wrap(err error) *Error {
	newErr := *e
	newErr.cause = err
	return &newErr
}

// Unwrap 支持 Go 标准错误链 errors.Is / errors.As
func (e *Error) Unwrap() error {
	return e.cause
}

// Is 按错误码判断同类错误
func (e *Error) Is(target error) bool {
	var t *Error
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

// ==================== 通用工具函数 ====================

// Code 从任意 error 中提取错误码；非框架错误默认返回 500
func Code(err error) int {
	if err == nil {
		return CodeSuccess
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return CodeInternalError
}

// Is 判断错误是否属于指定错误码
func Is(err error, code int) bool {
	return Code(err) == code
}

// FromError 普通 error 转框架统一错误
// 原生框架错误原样返回，其他包装为内部错误
func FromError(err error) *Error {
	if err == nil {
		return Success
	}
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return InternalError.Wrap(err)
}

// ==================== gRPC 错误双向透传 ====================

// 自定义错误码 ↔ gRPC 标准状态码映射表
var grpcCodeMap = map[int]codes.Code{
	CodeSuccess:        codes.OK,
	CodeInvalidParam:   codes.InvalidArgument,
	CodeUnauthorized:   codes.Unauthenticated,
	CodeForbidden:      codes.PermissionDenied,
	CodeNotFound:       codes.NotFound,
	CodeTooManyRequest: codes.ResourceExhausted,
	CodeInternalError:  codes.Internal,
	CodeNotImplemented: codes.Unimplemented,
	CodeServiceUnavail: codes.Unavailable,
	CodeTimeout:        codes.DeadlineExceeded,
}

// ToGRPCStatus 框架错误 → gRPC Status，服务端返回时使用
func ToGRPCStatus(err *Error) error {
	grpcCode, ok := grpcCodeMap[err.Code]
	if !ok {
		grpcCode = codes.Unknown
	}
	return status.Error(grpcCode, err.Error())
}

// FromGRPCStatus gRPC 错误 → 框架统一错误，客户端接收时使用
func FromGRPCStatus(err error) *Error {
	if err == nil {
		return Success
	}
	st, ok := status.FromError(err)
	if !ok {
		return InternalError.Wrap(err)
	}

	code := CodeInternalError
	for c, g := range grpcCodeMap {
		if g == st.Code() {
			code = c
			break
		}
	}
	return NewWithDetail(code, st.Message(), "")
}

// ==================== HTTP 状态码映射 ====================

// ToHTTPStatus 自定义错误码 → HTTP 响应状态码
func ToHTTPStatus(code int) int {
	switch code {
	case CodeSuccess:
		return 200
	case CodeInvalidParam:
		return 400
	case CodeUnauthorized:
		return 401
	case CodeForbidden:
		return 403
	case CodeNotFound:
		return 404
	case CodeMethodNotAllow:
		return 405
	case CodeTooManyRequest:
		return 429
	case CodeInternalError, CodeNotImplemented:
		return 500
	case CodeServiceUnavail:
		return 503
	case CodeTimeout:
		return 504
	default:
		return 500
	}
}
