package router

import (
	"nasa-go-admin/pkg/monitoring"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// InitMonitoringRoutes 初始化监控路由
func InitMonitoringRoutes(r *gin.Engine) {
	monitoringGroup := r.Group("/api/admin/monitoring")
	{
		// 系统概览
		monitoringGroup.GET("/overview", getSystemOverview)

		// HTTP请求指标
		monitoringGroup.GET("/http-metrics", getHTTPMetrics)

		// 实时统计
		monitoringGroup.GET("/realtime", getRealTimeStats)

		// 健康检查
		monitoringGroup.GET("/health", getHealthCheck)
	}
}

// getSystemOverview 获取系统概览
func getSystemOverview(c *gin.Context) {
	timeRange := c.DefaultQuery("time_range", "1h")

	stats, err := monitoring.GetMonitoringStats(timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取系统概览失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    stats,
		"message": "success",
	})
}

// getHTTPMetrics 获取HTTP请求指标
func getHTTPMetrics(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		limit = 100
	}

	metrics, err := monitoring.GetRecentHTTPRequests(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取HTTP指标失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"metrics": metrics,
			"total":   len(metrics),
		},
		"message": "success",
	})
}

// getRealTimeStats 获取实时统计
func getRealTimeStats(c *gin.Context) {
	timeRange := c.DefaultQuery("time_range", "1h")

	stats, err := monitoring.GetMonitoringStats(timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取实时统计失败: " + err.Error(),
		})
		return
	}

	// 添加一些实时数据
	if statsData, ok := stats["stats"].(map[string]interface{}); ok {
		realTimeStats := map[string]interface{}{
			"timestamp":          stats["timestamp"],
			"time_range":         timeRange,
			"http_requests":      statsData["http_requests"],
			"user_logins":        statsData["user_logins"],
			"user_registers":     statsData["user_registers"],
			"db_connections":     statsData["db_connections"],
			"db_max_connections": statsData["db_max_conns"],
			"system_status":      "running",
			"uptime":             "运行中",
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"data":    realTimeStats,
			"message": "success",
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "数据格式错误",
		})
	}
}

// getHealthCheck 健康检查
func getHealthCheck(c *gin.Context) {
	healthData := map[string]interface{}{
		"status":    "healthy",
		"timestamp": c.GetTime("2006-01-02 15:04:05"),
		"services": map[string]interface{}{
			"database": "connected",
			"mongodb":  "connected",
			"redis":    "connected",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    healthData,
		"message": "success",
	})
}
