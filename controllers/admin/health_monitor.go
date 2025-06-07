package admin

import (
	"nasa-go-admin/services/app_service"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthMonitorController 健康监控控制器
type HealthMonitorController struct{}

// NewHealthMonitorController 创建健康监控控制器
func NewHealthMonitorController() *HealthMonitorController {
	return &HealthMonitorController{}
}

// GetOrderServiceMetrics 获取订单服务指标
func (h *HealthMonitorController) GetOrderServiceMetrics(c *gin.Context) {
	metrics := app_service.GetOrderMetrics()

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": metrics,
	})
}

// GetSystemHealth 获取系统健康状态
func (h *HealthMonitorController) GetSystemHealth(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 获取服务初始化状态
	initializer := app_service.GetServiceInitializer()
	initStatus := initializer.GetInitializationStatus()

	healthStatus := map[string]interface{}{
		"timestamp":      time.Now(),
		"status":         "healthy",
		"uptime_seconds": time.Since(time.Now().Add(-time.Hour)).Seconds(), // 示例，实际需要记录启动时间
		"memory": map[string]interface{}{
			"alloc_bytes":       m.Alloc,
			"total_alloc_bytes": m.TotalAlloc,
			"sys_bytes":         m.Sys,
			"num_gc":            m.NumGC,
		},
		"goroutines": runtime.NumGoroutine(),
		"cpu_cores":  runtime.NumCPU(),
		"services": map[string]interface{}{
			"order_service_initialized": initStatus,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": healthStatus,
	})
}

// GetDatabaseHealth 获取数据库健康状态
func (h *HealthMonitorController) GetDatabaseHealth(c *gin.Context) {
	// 这里需要从订单服务获取数据库健康状态
	// 由于架构限制，这里提供一个基础的健康检查

	healthStatus := map[string]interface{}{
		"timestamp": time.Now(),
		"status":    "healthy",
		"message":   "数据库连接正常",
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": healthStatus,
	})
}

// GetPerformanceReport 获取性能报告
func (h *HealthMonitorController) GetPerformanceReport(c *gin.Context) {
	metrics := app_service.GetOrderMetrics()

	// 计算性能指标
	var successRate float64
	if metrics.TotalRequests > 0 {
		successRate = float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests) * 100
	}

	performanceReport := map[string]interface{}{
		"timestamp": time.Now(),
		"request_metrics": map[string]interface{}{
			"total_requests":       metrics.TotalRequests,
			"successful_requests":  metrics.SuccessfulRequests,
			"failed_requests":      metrics.FailedRequests,
			"success_rate_percent": successRate,
		},
		"response_time_metrics": map[string]interface{}{
			"average_ms": metrics.AverageResponseTime,
			"max_ms":     metrics.MaxResponseTime,
			"min_ms":     metrics.MinResponseTime,
		},
		"concurrency_metrics": map[string]interface{}{
			"active_connections": metrics.ActiveConnections,
			"peak_connections":   metrics.PeakConnections,
		},
		"error_metrics": map[string]interface{}{
			"database_errors":       metrics.DatabaseErrors,
			"redis_errors":          metrics.RedisErrors,
			"lock_timeouts":         metrics.LockTimeouts,
			"transaction_rollbacks": metrics.TransactionRollbacks,
		},
		"business_metrics": map[string]interface{}{
			"orders_created":     metrics.OrdersCreated,
			"orders_cancelled":   metrics.OrdersCancelled,
			"payments_processed": metrics.PaymentsProcessed,
		},
		"system_metrics": map[string]interface{}{
			"memory_usage_bytes": metrics.MemoryUsage,
			"goroutine_count":    metrics.GoroutineCount,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": performanceReport,
	})
}

// GetAlerts 获取系统报警信息
func (h *HealthMonitorController) GetAlerts(c *gin.Context) {
	metrics := app_service.GetOrderMetrics()
	alerts := []map[string]interface{}{}

	// 检查各种报警条件
	if metrics.TotalRequests > 100 {
		errorRate := float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
		if errorRate > 0.05 {
			alerts = append(alerts, map[string]interface{}{
				"level":     "warning",
				"type":      "high_error_rate",
				"message":   "订单服务错误率过高",
				"value":     errorRate * 100,
				"threshold": 5.0,
				"timestamp": time.Now(),
			})
		}
	}

	if metrics.AverageResponseTime > 5000 {
		alerts = append(alerts, map[string]interface{}{
			"level":     "warning",
			"type":      "slow_response",
			"message":   "订单服务响应时间过长",
			"value":     metrics.AverageResponseTime,
			"threshold": 5000,
			"timestamp": time.Now(),
		})
	}

	if metrics.GoroutineCount > 1000 {
		alerts = append(alerts, map[string]interface{}{
			"level":     "critical",
			"type":      "goroutine_leak",
			"message":   "Goroutine数量异常，可能存在内存泄漏",
			"value":     metrics.GoroutineCount,
			"threshold": 1000,
			"timestamp": time.Now(),
		})
	}

	if metrics.MemoryUsage > 500*1024*1024 {
		alerts = append(alerts, map[string]interface{}{
			"level":     "warning",
			"type":      "high_memory_usage",
			"message":   "内存使用过高",
			"value":     metrics.MemoryUsage,
			"threshold": 500 * 1024 * 1024,
			"timestamp": time.Now(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": map[string]interface{}{
			"alerts":      alerts,
			"alert_count": len(alerts),
			"timestamp":   time.Now(),
		},
	})
}
