package middleware

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// PerformanceConfig 性能监控配置
type PerformanceConfig struct {
	SlowThreshold time.Duration // 慢请求阈值
	EnableLogging bool          // 是否记录日志
	SkipPaths     []string      // 跳过监控的路径
}

// DefaultPerformanceConfig 默认性能配置
func DefaultPerformanceConfig() PerformanceConfig {
	return PerformanceConfig{
		SlowThreshold: 500 * time.Millisecond,
		EnableLogging: true,
		SkipPaths:     []string{"/health", "/metrics", "/favicon.ico"},
	}
}

// Performance 性能监控中间件
func Performance(config ...PerformanceConfig) gin.HandlerFunc {
	cfg := DefaultPerformanceConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		// 检查是否跳过监控
		for _, path := range cfg.SkipPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 处理请求
		c.Next()

		// 计算耗时
		latency := time.Since(start)
		status := c.Writer.Status()

		// 记录慢请求日志
		if cfg.EnableLogging && latency > cfg.SlowThreshold {
			log.Printf("[SLOW REQUEST] %s %s - Status: %d, Latency: %v",
				method, path, status, latency)
		}

		// 在响应头中添加性能信息（开发环境）
		if gin.Mode() == gin.DebugMode {
			c.Header("X-Response-Time", latency.String())
			c.Header("X-Request-ID", fmt.Sprintf("%d", time.Now().UnixNano()))
		}
	}
}

// DatabasePerformance 数据库性能监控中间件
func DatabasePerformance() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 在上下文中设置数据库查询计数器
		c.Set("db_query_count", 0)
		c.Set("db_start_time", time.Now())

		c.Next()

		// 获取数据库查询统计
		if queryCount, exists := c.Get("db_query_count"); exists {
			if count, ok := queryCount.(int); ok && count > 10 {
				log.Printf("[DB PERFORMANCE WARNING] Path: %s, Query Count: %d",
					c.Request.URL.Path, count)
			}
		}
	}
}

// IncrementQueryCount 增加查询计数
func IncrementQueryCount(c *gin.Context) {
	if c != nil {
		if current, exists := c.Get("db_query_count"); exists {
			if count, ok := current.(int); ok {
				c.Set("db_query_count", count+1)
			}
		} else {
			c.Set("db_query_count", 1)
		}
	}
}

// RateLimit 线程安全的内存限流中间件
func RateLimit(rpm int) gin.HandlerFunc {
	// 使用sync.Map来避免竞态条件
	var requests sync.Map

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		// 获取或创建IP的请求记录
		var timestamps []time.Time
		if value, exists := requests.Load(ip); exists {
			timestamps = value.([]time.Time)
		}

		// 清理过期的请求记录
		var validTimestamps []time.Time
		cutoff := now.Add(-time.Minute)

		for _, timestamp := range timestamps {
			if timestamp.After(cutoff) {
				validTimestamps = append(validTimestamps, timestamp)
			}
		}

		// 检查是否超过限制
		if len(validTimestamps) >= rpm {
			c.AbortWithStatusJSON(429, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": 60,
			})
			return
		}

		// 记录当前请求
		validTimestamps = append(validTimestamps, now)
		requests.Store(ip, validTimestamps)

		c.Next()
	}
}
