package health

import (
	"runtime"
	"time"

	"nasa-go-admin/pkg/config"
	"nasa-go-admin/pkg/database"
	"nasa-go-admin/pkg/response"
	"nasa-go-admin/redis"

	"github.com/gin-gonic/gin"
)

// HealthController 健康检查控制器
type HealthController struct{}

// NewHealthController 创建健康检查控制器
func NewHealthController() *HealthController {
	return &HealthController{}
}

// CheckHealth 基础健康检查
func (h *HealthController) CheckHealth(c *gin.Context) {
	response.Success(c, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"service":   "nasa-go-admin",
		"version":   "1.0.0",
	})
}

// CheckLiveness 存活性检查
func (h *HealthController) CheckLiveness(c *gin.Context) {
	response.Success(c, gin.H{
		"status":    "alive",
		"timestamp": time.Now().Unix(),
	})
}

// CheckReadiness 就绪性检查
func (h *HealthController) CheckReadiness(c *gin.Context) {
	issues := make([]string, 0)

	// 检查数据库连接
	if err := database.HealthCheck(); err != nil {
		issues = append(issues, "database: "+err.Error())
	}

	// 检查Redis连接
	if !redis.IsConnected() {
		issues = append(issues, "redis: connection failed")
	}

	if len(issues) > 0 {
		response.Error(c, response.ERROR, "service not ready")
		return
	}

	response.Success(c, gin.H{
		"status":    "ready",
		"timestamp": time.Now().Unix(),
	})
}

// GetSystemInfo 获取系统信息
func (h *HealthController) GetSystemInfo(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	cfg := config.GetConfig()

	info := gin.H{
		"service": gin.H{
			"name":    "nasa-go-admin",
			"version": "1.0.0",
			"mode":    cfg.Server.Mode,
			"uptime":  time.Since(startTime).String(),
		},
		"system": gin.H{
			"go_version":    runtime.Version(),
			"num_cpu":       runtime.NumCPU(),
			"num_goroutine": runtime.NumGoroutine(),
		},
		"memory": gin.H{
			"alloc":       bToMb(m.Alloc),
			"total_alloc": bToMb(m.TotalAlloc),
			"sys":         bToMb(m.Sys),
			"num_gc":      m.NumGC,
		},
		"database": database.GetStats(),
		"redis": gin.H{
			"connected": redis.IsConnected(),
		},
		"timestamp": time.Now().Unix(),
	}

	response.Success(c, info)
}

// GetMetrics 获取应用指标（Prometheus格式）
func (h *HealthController) GetMetrics(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	dbStats := database.GetStats()

	// 简单的Prometheus格式指标
	metrics := []string{
		"# HELP nasa_go_admin_goroutines Current number of goroutines",
		"# TYPE nasa_go_admin_goroutines gauge",
		"nasa_go_admin_goroutines " + string(rune(runtime.NumGoroutine())),
		"",
		"# HELP nasa_go_admin_memory_alloc Current memory allocation",
		"# TYPE nasa_go_admin_memory_alloc gauge",
		"nasa_go_admin_memory_alloc " + string(rune(m.Alloc)),
		"",
		"# HELP nasa_go_admin_database_connections Current database connections",
		"# TYPE nasa_go_admin_database_connections gauge",
	}

	if openConns, ok := dbStats["open_connections"].(int); ok {
		metrics = append(metrics, "nasa_go_admin_database_connections "+string(rune(openConns)))
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(200, joinStrings(metrics, "\n"))
}

// startTime 应用启动时间
var startTime = time.Now()

// bToMb 字节转MB
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// joinStrings 连接字符串
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}
