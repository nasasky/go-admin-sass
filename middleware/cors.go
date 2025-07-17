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
	var allowedOrigins []string

	// 从环境变量读取允许的域名
	if envOrigins := os.Getenv("ALLOWED_ORIGINS"); envOrigins != "" {
		allowedOrigins = strings.Split(envOrigins, ",")
		// 清理空格
		for i, origin := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(origin)
		}
	} else {
		// 默认配置 - 更宽松的局域网支持
		allowedOrigins = []string{
			// 本地开发
			"http://localhost:3000",
			"http://localhost:8080",
			"http://localhost:9848",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:8080",
			"http://127.0.0.1:9848",
			// 局域网段支持 - 常见的私有IP段
			"http://192.168.0.114:9848", // 原有配置
		}

		// 开发环境允许所有域名（包括局域网）
		if gin_mode := os.Getenv("GIN_MODE"); gin_mode != "release" {
			allowedOrigins = append(allowedOrigins, "*")
		}
	}

	return CorsConfig{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH", "HEAD"},
		AllowedHeaders: []string{
			"Origin", "Content-Type", "Content-Length", "Accept-Encoding",
			"X-CSRF-Token", "Authorization", "X-Request-ID", "Accept",
			"Cache-Control", "X-Requested-With", "User-Agent", "Cookie",
		},
		ExposedHeaders:   []string{"Content-Length", "X-Request-ID", "X-Total-Count", "Set-Cookie"},
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
			// 当指定具体的 origin 时，必须设置 Allow-Credentials
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		} else if origin == "" {
			// 如果没有 Origin 头，不设置 CORS 头
			c.Next()
			return
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
		c.Writer.Header().Set("Access-Control-Expose-Headers", strings.Join(cfg.ExposedHeaders, ", "))
		c.Writer.Header().Set("Access-Control-Max-Age", string(rune(cfg.MaxAge)))

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
		// 完全匹配
		if allowed == origin || allowed == "*" {
			return true
		}

		// 支持通配符子域名匹配
		if strings.HasPrefix(allowed, "*.") {
			domain := strings.TrimPrefix(allowed, "*.")
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}

		// 支持端口通配符匹配 (例如: http://localhost:*)
		if strings.HasSuffix(allowed, ":*") {
			baseURL := strings.TrimSuffix(allowed, ":*")
			if strings.HasPrefix(origin, baseURL+":") {
				return true
			}
		}

		// 支持IP段通配符匹配 (例如: http://192.168.*:*)
		if strings.Contains(allowed, "*") {
			// 将通配符转换为正则表达式模式
			pattern := strings.ReplaceAll(allowed, "*", ".*")
			pattern = "^" + pattern + "$"
			// 简单的模式匹配
			if matchesPattern(origin, pattern) {
				return true
			}
		}
	}

	return false
}

// matchesPattern 简单的模式匹配函数
func matchesPattern(str, pattern string) bool {
	// 将模式中的.*替换为通配符进行匹配
	pattern = strings.ReplaceAll(pattern, "^", "")
	pattern = strings.ReplaceAll(pattern, "$", "")

	// 如果模式包含.*，进行分段匹配
	if strings.Contains(pattern, ".*") {
		parts := strings.Split(pattern, ".*")
		currentStr := str

		for i, part := range parts {
			if part == "" {
				continue
			}

			if i == 0 {
				// 第一部分必须在开头
				if !strings.HasPrefix(currentStr, part) {
					return false
				}
				currentStr = currentStr[len(part):]
			} else if i == len(parts)-1 {
				// 最后一部分必须在结尾
				if !strings.HasSuffix(currentStr, part) {
					return false
				}
			} else {
				// 中间部分必须包含
				if !strings.Contains(currentStr, part) {
					return false
				}
				index := strings.Index(currentStr, part)
				currentStr = currentStr[index+len(part):]
			}
		}
		return true
	}

	return str == pattern
}
