package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	"strings"
)

// ValidationMiddleware 是一个中间件，用于绑定和验证请求数据
func ValidationMiddleware(obj interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 绑定请求数据并进行验证
		if err := c.ShouldBind(obj); err != nil {
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				out := make([]string, len(ve))
				for i, fe := range ve {
					out[i] = fe.Field() + " " + fe.Tag()
				}
				c.JSON(http.StatusBadRequest, gin.H{
					"code":      20001,
					"message":   strings.Join(out, ", "),
					"error":     "error some",
					"originUrl": c.Request.URL.Path,
				})
			} else {
				// 检查是否是 EOF 错误
				if err.Error() == "EOF" {
					c.JSON(http.StatusBadRequest, gin.H{
						"code":      20001,
						"message":   "请求体为空或格式不正确",
						"error":     "error some",
						"originUrl": c.Request.URL.Path,
					})
				} else {
					c.JSON(http.StatusBadRequest, gin.H{
						"code":      20001,
						"message":   err.Error(),
						"error":     "error some",
						"originUrl": c.Request.URL.Path,
					})
				}
			}
			c.Abort()
			return
		}
		c.Next()
	}
}
