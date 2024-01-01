package template

import "meta-egg/internal/domain/helper"

var TplInternalCommonCerrorCerror string = helper.PH_META_EGG_HEADER + `
package cerror

import (
	"fmt"
	"net/http"
	"strings"

	api "%%GO-MODULE%%/api/%%PROJECT-NAME-DIR%%"
)

type CustomError struct {
	HttpStatus int             ` + "`" + `json:"http_status"` + "`" + ` // http 状态
	Code       api.ErrCode ` + "`" + `json:"code"` + "`" + `        // code 错误码
	Message    string          ` + "`" + `json:"msg"` + "`" + `         // msg 消息
	Detail     string          ` + "`" + `json:"detail"` + "`" + `      // detail 详细信息
}

// 实现 error 接口
func (e *CustomError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Detail)
	} else {
		return e.Message
	}
}

// 支持errors.Is
func (e CustomError) Is(target error) bool {
	err, ok := target.(*CustomError)
	if !ok {
		return false
	}
	return e.Code == err.Code
}

type NewCErrorWithDetail func(detail ...string) *CustomError

func NewCError(code api.ErrCode, message string) NewCErrorWithDetail {
	return func(detail ...string) *CustomError {
		err := &CustomError{
			HttpStatus: GetHttpStatusByCode(code),
			Code:       code,
			Message:    message,
		}
		if len(detail) > 0 {
			err.Detail = strings.Join(detail, " ")
		}
		return err
	}
}

func GetHttpStatusByCode(code api.ErrCode) int {
	if httpStatus, ok := Code2HttpStatus[code]; ok {
		return httpStatus
	} else {
		return http.StatusInternalServerError
	}
}

var Code2HttpStatus = map[api.ErrCode]int{
	api.ErrCode_Ok:                 http.StatusOK,
	api.ErrCode_Unknown:            http.StatusInternalServerError,
	api.ErrCode_InvalidArgument:    http.StatusBadRequest,
	api.ErrCode_NotFound:           http.StatusBadRequest,
	api.ErrCode_AlreadyExists:      http.StatusBadRequest,
	api.ErrCode_PermissionDenied:   http.StatusForbidden,
	api.ErrCode_ResourceExhausted:  http.StatusTooManyRequests,
	api.ErrCode_FailedPrecondition: http.StatusBadRequest,
	api.ErrCode_Aborted:            http.StatusBadRequest,
	api.ErrCode_OutOfRange:         http.StatusBadRequest,
	api.ErrCode_Internal:           http.StatusInternalServerError,
	api.ErrCode_Unavailable:        http.StatusServiceUnavailable,
	api.ErrCode_DataLoss:           http.StatusInternalServerError,
	api.ErrCode_Unauthenticated:    http.StatusUnauthorized,
}

var (
	Ok = NewCError(api.ErrCode_Ok, "ok")

	// Basic error (read comments in proto/error.proto)
	Unknown            = NewCError(api.ErrCode_Unknown, "unknown error")                  // 未知错误
	InvalidArgument    = NewCError(api.ErrCode_InvalidArgument, "invalid argument")       // 参数错
	NotFound           = NewCError(api.ErrCode_NotFound, "not found")                     // 实体不存在
	AlreadyExists      = NewCError(api.ErrCode_AlreadyExists, "already exists")           // 创建实体时冲突
	PermissionDenied   = NewCError(api.ErrCode_PermissionDenied, "permission denied")     // 权限不足
	ResourceExhausted  = NewCError(api.ErrCode_ResourceExhausted, "resource exhausted")   // 资源不足
	FailedPrecondition = NewCError(api.ErrCode_FailedPrecondition, "failed precondition") // 前置条件失败
	Aborted            = NewCError(api.ErrCode_Aborted, "aborted")                        // 操作被中止
	OutOfRange         = NewCError(api.ErrCode_OutOfRange, "out of range")                // 超出范围
	Internal           = NewCError(api.ErrCode_Internal, "internal error")                // 内部错误
	Unavailable        = NewCError(api.ErrCode_Unavailable, "unavailable")                // 服务不可用，请重试
	DataLoss           = NewCError(api.ErrCode_DataLoss, "data loss")                     // 数据丢失或损坏
	Unauthenticated    = NewCError(api.ErrCode_Unauthenticated, "unauthenticated")        // 未认证，客户端未提供凭据或提供的凭据无效

	// 以上是框架生成的, 自行新增的, 请将code值从1001开始
)
`
