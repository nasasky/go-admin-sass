package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"nasa-go-admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// Recovery 自定义恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// 记录panic详细信息
		err := fmt.Sprintf("panic recovered: %v", recovered)
		stack := string(debug.Stack())

		log.Printf("[PANIC RECOVERY] %s\n%s", err, stack)

		// 根据环境返回不同的错误信息
		if gin.Mode() == gin.DebugMode {
			// 开发环境返回详细错误信息
			response.ErrorWithData(c, response.INTERNAL_ERROR, gin.H{
				"panic": recovered,
				"stack": stack,
			}, "服务器内部错误")
		} else {
			// 生产环境只返回通用错误信息
			response.Error(c, response.INTERNAL_ERROR, "服务器内部错误")
		}
	})
}

// ErrorHandler 统一错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// 记录错误
			log.Printf("[ERROR] %s %s - %v", c.Request.Method, c.Request.URL.Path, err.Err)

			// 如果还没有响应，则发送错误响应
			if !c.Writer.Written() {
				switch err.Type {
				case gin.ErrorTypeBind:
					response.Error(c, response.INVALID_PARAMS, "请求参数错误: "+err.Error())
				case gin.ErrorTypePublic:
					response.Error(c, response.ERROR, err.Error())
				default:
					response.Error(c, response.INTERNAL_ERROR, "内部服务错误")
				}
			}
		}
	}
}

// RateLimitMiddleware 请求频率限制中间件（简单实现）
func RateLimitMiddleware(maxRequests int, windowSeconds int) gin.HandlerFunc {
	// 这里使用简单的内存存储，生产环境建议使用Redis
	requestCounts := make(map[string]int)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// 检查请求计数
		if count, exists := requestCounts[clientIP]; exists && count >= maxRequests {
			response.Abort(c, response.TOO_MANY_REQUESTS, "请求过于频繁，请稍后再试")
			return
		}

		// 增加计数
		requestCounts[clientIP]++

		c.Next()

		// 在实际应用中，这里应该实现基于时间窗口的清理逻辑
	}
}

// SecureHeaders 安全头中间件
func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置安全相关的HTTP头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 在生产环境中启用HSTS
		if gin.Mode() == gin.ReleaseMode {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// RequestID 为每个请求生成唯一ID
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID()
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// generateRequestID 生成请求ID（UUID实现）
func generateRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// 如果随机数生成失败，使用时间戳作为备选
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}
