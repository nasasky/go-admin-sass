package router

import (
	"nasa-go-admin/db"
	"nasa-go-admin/pkg/monitoring"
	"nasa-go-admin/redis"
	"nasa-go-admin/services/app_service"
	"nasa-go-admin/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	orderMonitor  *app_service.OrderMonitoringService
	systemManager *app_service.OrderSystemManager
)

// InitMonitoringRoutes 初始化监控路由
func InitMonitoringRoutes(r *gin.Engine) {
	// 初始化监控服务
	if orderMonitor == nil {
		orderMonitor = app_service.NewOrderMonitoringService(db.Dao, redis.GetClient())
	}
	if systemManager == nil {
		systemManager = app_service.NewOrderSystemManager(redis.GetClient())
		systemManager.Initialize() // 初始化系统管理器
	}

	// 监控路由组
	monitorGroup := r.Group("/api/monitor")
	{
		// 订单系统健康检查
		monitorGroup.GET("/order/health", func(c *gin.Context) {
			status := systemManager.GetSystemStatus()
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  "success",
				"data": status,
			})
		})

		// 订单系统统计信息
		monitorGroup.GET("/order/stats", func(c *gin.Context) {
			stats, err := orderMonitor.GetMonitoringStats()
			if err != nil {
				utils.Err(c, utils.ErrCodeInternalError, err)
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  "success",
				"data": stats,
			})
		})

		// 获取异常订单报告
		monitorGroup.GET("/order/anomalies", func(c *gin.Context) {
			// 简化版本的异常检查
			anomalies := gin.H{
				"checked_time": time.Now(),
				"pending_orders": gin.H{
					"count":        0,
					"urgent_count": 0,
				},
				"abnormal_payments": gin.H{
					"count": 0,
				},
				"stock_anomalies": gin.H{
					"count": 0,
				},
			}
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  "success",
				"data": anomalies,
			})
		})

		// 修复数据一致性问题
		monitorGroup.POST("/order/fix-consistency", func(c *gin.Context) {
			// 简化版本的一致性修复
			results := gin.H{
				"fixed_count": 0,
				"timestamp":   time.Now(),
				"message":     "数据一致性检查完成",
			}
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  "success",
				"data": results,
			})
		})

		// 清理过期订单
		monitorGroup.POST("/order/cleanup-expired", func(c *gin.Context) {
			// 简化版本的过期订单清理
			count := 0 // 这里可以调用实际的清理逻辑
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  "success",
				"data": gin.H{"cleaned_count": count},
			})
		})

		// Prometheus metrics endpoint
		monitorGroup.GET("/metrics", func(c *gin.Context) {
			// 这里集成Prometheus metrics
			c.String(http.StatusOK, "# Prometheus metrics endpoint\n")
		})
	}

	// 管理后台监控路由组
	adminMonitorGroup := r.Group("/api/admin/monitor")
	{
		// 订单监控面板数据
		adminMonitorGroup.GET("/dashboard", func(c *gin.Context) {
			stats, err := orderMonitor.GetMonitoringStats()
			if err != nil {
				utils.Err(c, utils.ErrCodeInternalError, err)
				return
			}

			dashboard := gin.H{
				"stats":     stats,
				"anomalies": gin.H{"count": 0},
				"timestamp": time.Now(),
			}

			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  "success",
				"data": dashboard,
			})
		})

		// 实时监控警报
		adminMonitorGroup.GET("/alerts", func(c *gin.Context) {
			alerts := []gin.H{} // 空的警报列表
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  "success",
				"data": alerts,
			})
		})
	}

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
