package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 统一错误码定义
const (
	SUCCESS           = 200
	ERROR             = 500
	INVALID_PARAMS    = 20001
	AUTH_ERROR        = 20002
	NOT_FOUND         = 20003
	FORBIDDEN         = 20004
	TOO_MANY_REQUESTS = 20005
	INTERNAL_ERROR    = 20006
)

// 错误码消息映射
var codeMsg = map[int]string{
	SUCCESS:           "OK",
	ERROR:             "服务器内部错误",
	INVALID_PARAMS:    "请求参数错误",
	AUTH_ERROR:        "认证失败",
	NOT_FOUND:         "资源不存在",
	FORBIDDEN:         "访问被禁止",
	TOO_MANY_REQUESTS: "请求过于频繁",
	INTERNAL_ERROR:    "内部服务错误",
}

// Response 统一响应结构
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	OriginUrl string      `json:"originUrl"`
}

// 获取错误码对应的消息
func GetMsg(code int) string {
	msg, exist := codeMsg[code]
	if exist {
		return msg
	}
	return codeMsg[ERROR]
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	resp := Response{
		Code:      SUCCESS,
		Message:   GetMsg(SUCCESS),
		Data:      data,
		OriginUrl: c.Request.URL.Path,
	}
	c.Set("response", resp)
	c.JSON(http.StatusOK, resp)
}

// Error 错误响应
func Error(c *gin.Context, code int, message ...string) {
	msg := GetMsg(code)
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	
	resp := Response{
		Code:      code,
		Message:   msg,
		Error:     "error",
		OriginUrl: c.Request.URL.Path,
	}
	c.Set("response", resp)
	c.JSON(http.StatusOK, resp)
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(c *gin.Context, code int, data interface{}, message ...string) {
	msg := GetMsg(code)
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	
	resp := Response{
		Code:      code,
		Message:   msg,
		Data:      data,
		Error:     "error",
		OriginUrl: c.Request.URL.Path,
	}
	c.Set("response", resp)
	c.JSON(http.StatusOK, resp)
}

// Abort 中断请求并返回错误
func Abort(c *gin.Context, code int, message ...string) {
	Error(c, code, message...)
	c.Abort()
} 