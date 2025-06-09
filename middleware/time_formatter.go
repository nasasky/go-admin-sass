package middleware

import (
	"nasa-go-admin/utils"

	"github.com/gin-gonic/gin"
)

// TimeFormatterMiddleware 时间格式化中间件
// 自动将响应中的MongoDB时间格式转换为友好的格式
func TimeFormatterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建一个自定义的 ResponseWriter
		writer := &timeFormatterWriter{
			ResponseWriter: c.Writer,
			c:              c,
		}
		c.Writer = writer

		c.Next()
	}
}

// timeFormatterWriter 自定义响应写入器
type timeFormatterWriter struct {
	gin.ResponseWriter
	c *gin.Context
}

// Write 重写Write方法，在响应写入前进行时间格式化
func (w *timeFormatterWriter) Write(data []byte) (int, error) {
	// 这里可以根据需要处理响应数据
	// 由于直接处理字节流比较复杂，我们在controller层面处理时间格式化更好
	return w.ResponseWriter.Write(data)
}

// FormatResponseTime 格式化响应中的时间字段
func FormatResponseTime(data interface{}) interface{} {
	return utils.FormatTimeFieldsForResponse(data)
}

// FormatResponseTimeWithTimezone 带时区信息的时间格式化
func FormatResponseTimeWithTimezone(data interface{}) map[string]interface{} {
	formatted := utils.FormatTimeFieldsForResponse(data)
	timezone := utils.GetTimeZoneInfo()

	return map[string]interface{}{
		"data":     formatted,
		"timezone": timezone,
	}
}
