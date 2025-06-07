package admin

import (
	"nasa-go-admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// Resp 为了兼容性保留，但推荐直接使用 response 包
var Resp = &rps{}

type rps struct{}

// Succ 成功响应 - 兼容旧接口
func (rps) Succ(c *gin.Context, data interface{}) {
	response.Success(c, data)
}

// Err 错误响应 - 兼容旧接口
func (rps) Err(c *gin.Context, errCode int, message string) {
	response.Error(c, errCode, message)
}
