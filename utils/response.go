package utils

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 定义常见的错误码和对应的错误信息
const (
	ErrCodeInvalidParams = 20001
	ErrCodeInternalError = 20002
)

var errMessages = map[int]string{
	ErrCodeInvalidParams: "Invalid parameters",
	ErrCodeInternalError: "Internal server error",
}

// Err 统一错误响应
func Err(c *gin.Context, code int, err error) {
	message, ok := errMessages[code]
	if !ok {
		message = "Unknown error"
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    code,
		"message": message,
		"error":   err.Error(),
	})
	c.Abort()
}

// Succ 统一成功响应
func Succ(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": data,
	})
}

// NewError 创建一个新的错误
func NewError(message string) error {
	return errors.New(message)
}
