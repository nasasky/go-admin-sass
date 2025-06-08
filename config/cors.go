package config

import (
	"os"
	"strings"
)

// CorsSettings CORS相关配置
type CorsSettings struct {
	AllowedOrigins   []string `json:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers"`
	ExposedHeaders   []string `json:"exposed_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age"`
}

// GetCorsConfig 获取CORS配置
func GetCorsConfig() CorsSettings {
	// 根据环境变量或配置确定允许的域名
	var allowedOrigins []string

	// 从环境变量获取
	if envOrigins := os.Getenv("ALLOWED_ORIGINS"); envOrigins != "" {
		allowedOrigins = strings.Split(envOrigins, ",")
		// 清理空格
		for i, origin := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(origin)
		}
	} else {
		// 默认配置 - 支持常见的开发和生产环境
		allowedOrigins = []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"http://localhost:9848",
			"https://localhost:3000",
			"https://localhost:8080",
			"https://localhost:9848",
			// 支持二级域名的通配符配置
			"*.maydayland.top",
			"*.maydayland.top",
			// 如果是开发环境，允许所有域名
		}

		// 开发环境允许所有域名
		if os.Getenv("GIN_MODE") != "release" {
			allowedOrigins = append(allowedOrigins, "*")
		}
	}

	return CorsSettings{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD",
		},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"X-CSRF-Token",
			"Authorization",
			"X-Request-ID",
			"Accept",
			"Cache-Control",
			"X-Requested-With",
			"User-Agent",
			"X-Real-IP",
			"X-Forwarded-For",
		},
		ExposedHeaders: []string{
			"Content-Length",
			"X-Request-ID",
			"X-Total-Count",
			"X-Page-Count",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24小时
	}
}

// GetAllowedOriginPatterns 获取允许的域名模式，用于更精确的匹配
func GetAllowedOriginPatterns() []string {
	return []string{
		// 本地开发
		"http://localhost:*",
		"https://localhost:*",
		"http://127.0.0.1:*",
		"https://127.0.0.1:*",

		// 二级域名模式
		"*.maydayland.top",
		"*.maydayland.top",

		// 可以添加更多的域名模式
	}
}
