package health

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"nasa-go-admin/db"
	"nasa-go-admin/pkg/monitoring"
	"nasa-go-admin/redis"

	"github.com/gin-gonic/gin"
)

// EnhancedHealthController 增强版健康检查控制器
type EnhancedHealthController struct{}

// HealthStatus 健康状态
type HealthStatus struct {
	Status      string                     `json:"status"`
	Timestamp   time.Time                  `json:"timestamp"`
	Version     string                     `json:"version"`
	Environment string                     `json:"environment"`
	Uptime      time.Duration              `json:"uptime"`
	Components  map[string]ComponentStatus `json:"components"`
	Metrics     *monitoring.Metrics        `json:"metrics,omitempty"`
}

// ComponentStatus 组件状态
type ComponentStatus struct {
	Status  string        `json:"status"`
	Message string        `json:"message,omitempty"`
	Latency time.Duration `json:"latency,omitempty"`
	Error   string        `json:"error,omitempty"`
}

var (
	enhancedStartTime = time.Now()
	enhancedVersion   = "1.0.0" // 应该从构建时注入
)

// NewEnhancedHealthController 创建增强版健康检查控制器
func NewEnhancedHealthController() *EnhancedHealthController {
	return &EnhancedHealthController{}
}

// CheckComprehensive 综合健康检查
func (h *EnhancedHealthController) CheckComprehensive(c *gin.Context) {
	status := &HealthStatus{
		Timestamp:   time.Now(),
		Version:     enhancedVersion,
		Environment: gin.Mode(),
		Uptime:      time.Since(enhancedStartTime),
		Components:  make(map[string]ComponentStatus),
	}

	// 检查各个组件
	h.checkDatabase(status)
	h.checkRedis(status)
	h.checkSystem(status)

	// 添加指标信息（可选）
	if c.Query("metrics") == "true" {
		status.Metrics = monitoring.GetMetrics().GetSnapshot()
	}

	// 确定整体状态
	status.Status = h.determineOverallStatus(status.Components)

	// 根据状态返回相应的HTTP状态码
	statusCode := http.StatusOK
	if status.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if status.Status == "degraded" {
		statusCode = http.StatusPartialContent
	}

	c.JSON(statusCode, status)
}

// checkDatabase 检查数据库连接
func (h *EnhancedHealthController) checkDatabase(status *HealthStatus) {
	start := time.Now()

	if db.Dao == nil {
		status.Components["database"] = ComponentStatus{
			Status:  "unhealthy",
			Message: "Database not initialized",
			Latency: time.Since(start),
		}
		return
	}

	// 获取底层的sql.DB
	sqlDB, err := db.Dao.DB()
	if err != nil {
		status.Components["database"] = ComponentStatus{
			Status:  "unhealthy",
			Error:   err.Error(),
			Latency: time.Since(start),
		}
		return
	}

	// 检查连接
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		status.Components["database"] = ComponentStatus{
			Status:  "unhealthy",
			Error:   err.Error(),
			Latency: time.Since(start),
		}
		return
	}

	// 检查连接池状态
	stats := sqlDB.Stats()
	message := fmt.Sprintf("Connections: %d/%d (idle: %d)",
		stats.OpenConnections, stats.MaxOpenConnections, stats.Idle)

	componentStatus := "healthy"
	if float64(stats.OpenConnections)/float64(stats.MaxOpenConnections) > 0.8 {
		componentStatus = "degraded"
		message += " - High connection usage"
	}

	status.Components["database"] = ComponentStatus{
		Status:  componentStatus,
		Message: message,
		Latency: time.Since(start),
	}
}

// checkRedis 检查Redis连接
func (h *EnhancedHealthController) checkRedis(status *HealthStatus) {
	start := time.Now()

	client := redis.GetClient()
	if client == nil {
		status.Components["redis"] = ComponentStatus{
			Status:  "unhealthy",
			Message: "Redis client not initialized",
			Latency: time.Since(start),
		}
		return
	}

	// 检查Redis连接
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		status.Components["redis"] = ComponentStatus{
			Status:  "unhealthy",
			Error:   err.Error(),
			Latency: time.Since(start),
		}
		return
	}

	// 获取Redis信息
	info, err := client.Info(ctx, "memory").Result()
	if err != nil {
		status.Components["redis"] = ComponentStatus{
			Status:  "degraded",
			Message: "Connected but info unavailable",
			Error:   err.Error(),
			Latency: time.Since(start),
		}
		return
	}

	status.Components["redis"] = ComponentStatus{
		Status:  "healthy",
		Message: "Connected - " + extractRedisMemoryInfo(info),
		Latency: time.Since(start),
	}
}

// checkSystem 检查系统资源
func (h *EnhancedHealthController) checkSystem(status *HealthStatus) {
	start := time.Now()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 检查内存使用情况
	memoryUsageMB := m.Alloc / 1024 / 1024
	goroutines := runtime.NumGoroutine()

	message := fmt.Sprintf("Memory: %dMB, Goroutines: %d", memoryUsageMB, goroutines)

	componentStatus := "healthy"
	if memoryUsageMB > 1000 { // 超过1GB内存使用
		componentStatus = "degraded"
		message += " - High memory usage"
	}

	if goroutines > 1000 { // 超过1000个goroutine
		componentStatus = "degraded"
		message += " - High goroutine count"
	}

	status.Components["system"] = ComponentStatus{
		Status:  componentStatus,
		Message: message,
		Latency: time.Since(start),
	}
}

// determineOverallStatus 确定整体状态
func (h *EnhancedHealthController) determineOverallStatus(components map[string]ComponentStatus) string {
	hasUnhealthy := false
	hasDegraded := false

	for _, component := range components {
		switch component.Status {
		case "unhealthy":
			hasUnhealthy = true
		case "degraded":
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return "unhealthy"
	}
	if hasDegraded {
		return "degraded"
	}
	return "healthy"
}

// extractRedisMemoryInfo 提取Redis内存信息
func extractRedisMemoryInfo(info string) string {
	// 简单提取，实际可以更复杂的解析
	if len(info) > 50 {
		return "Memory info available"
	}
	return "Memory info unavailable"
}

// GetDetailedMetrics 获取详细指标
func (h *EnhancedHealthController) GetDetailedMetrics(c *gin.Context) {
	metrics := monitoring.GetMetrics().GetSnapshot()
	c.JSON(http.StatusOK, gin.H{
		"timestamp": time.Now(),
		"metrics":   metrics,
	})
}

// GetSystemInfo 获取系统信息
func (h *EnhancedHealthController) GetSystemInfo(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	info := gin.H{
		"version":     enhancedVersion,
		"go_version":  runtime.Version(),
		"environment": gin.Mode(),
		"uptime":      time.Since(enhancedStartTime).String(),
		"memory": gin.H{
			"alloc":       m.Alloc,
			"total_alloc": m.TotalAlloc,
			"sys":         m.Sys,
			"num_gc":      m.NumGC,
		},
		"goroutines": runtime.NumGoroutine(),
		"cpu_count":  runtime.NumCPU(),
	}

	c.JSON(http.StatusOK, info)
}
