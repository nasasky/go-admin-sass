package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CorsConfig CORS配置
type CorsConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCorsConfig 默认CORS配置
func DefaultCorsConfig() CorsConfig {
	allowedOrigins := []string{"http://localhost:3000", "http://localhost:8080"}

	// 从环境变量读取允许的域名
	if envOrigins := os.Getenv("ALLOWED_ORIGINS"); envOrigins != "" {
		allowedOrigins = strings.Split(envOrigins, ",")
	}

	return CorsConfig{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Origin", "Content-Type", "Content-Length", "Accept-Encoding",
			"X-CSRF-Token", "Authorization", "X-Request-ID",
		},
		ExposedHeaders:   []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           3600, // 减少到1小时
	}
}

func Cors(config ...CorsConfig) gin.HandlerFunc {
	cfg := DefaultCorsConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// 检查是否为允许的来源
		if isAllowedOrigin(origin, cfg.AllowedOrigins) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
		c.Writer.Header().Set("Access-Control-Expose-Headers", strings.Join(cfg.ExposedHeaders, ", "))

		if cfg.AllowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if cfg.MaxAge > 0 {
			c.Writer.Header().Set("Access-Control-Max-Age", string(rune(cfg.MaxAge)))
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// isAllowedOrigin 检查来源是否被允许
func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
		// 支持通配符子域名匹配
		if strings.HasPrefix(allowed, "*.") {
			domain := strings.TrimPrefix(allowed, "*.")
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}

	return false
}
