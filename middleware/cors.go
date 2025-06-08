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
	// 更宽松的默认域名配置，支持二级域名
	allowedOrigins := []string{
		"http://localhost:3000",
		"http://localhost:8080",
		"http://localhost:9848",
		"*", // 开发环境允许所有域名
	}

	// 从环境变量读取允许的域名
	if envOrigins := os.Getenv("ALLOWED_ORIGINS"); envOrigins != "" {
		allowedOrigins = strings.Split(envOrigins, ",")
	}

	return CorsConfig{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH", "HEAD"},
		AllowedHeaders: []string{
			"Origin", "Content-Type", "Content-Length", "Accept-Encoding",
			"X-CSRF-Token", "Authorization", "X-Request-ID", "Accept",
			"Cache-Control", "X-Requested-With", "User-Agent",
		},
		ExposedHeaders:   []string{"Content-Length", "X-Request-ID", "X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           86400, // 24小时
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
		if origin != "" && isAllowedOrigin(origin, cfg.AllowedOrigins) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else if origin == "" || containsWildcard(cfg.AllowedOrigins) {
			// 如果没有Origin头或者配置了通配符，设置为通配符
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
		c.Writer.Header().Set("Access-Control-Expose-Headers", strings.Join(cfg.ExposedHeaders, ", "))

		if cfg.AllowCredentials && origin != "" {
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
		// 支持端口匹配 (例如: *.example.com:8080)
		if strings.Contains(allowed, "*.") && strings.Contains(origin, allowed[strings.Index(allowed, ".")+1:]) {
			return true
		}
	}

	return false
}

// containsWildcard 检查是否包含通配符
func containsWildcard(origins []string) bool {
	for _, origin := range origins {
		if origin == "*" {
			return true
		}
	}
	return false
}
