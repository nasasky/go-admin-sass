package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"runtime"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
)

// OptimizedDBConfig 优化的数据库配置
type OptimizedDBConfig struct {
	MaxOpenConns    int           // 最大打开连接数
	MaxIdleConns    int           // 最大空闲连接数
	ConnMaxLifetime time.Duration // 连接最大生命周期
	ConnMaxIdleTime time.Duration // 空闲连接最大生命周期

	// 监控配置
	SlowQueryThreshold time.Duration // 慢查询阈值
	EnableMonitoring   bool          // 是否启用监控
	StatsInterval      time.Duration // 统计间隔
}

// DBStats 数据库统计信息
type DBStats struct {
	MaxOpenConnections int           `json:"max_open_connections"`
	OpenConnections    int           `json:"open_connections"`
	InUse              int           `json:"in_use"`
	Idle               int           `json:"idle"`
	WaitCount          int64         `json:"wait_count"`
	WaitDuration       time.Duration `json:"wait_duration"`
	MaxIdleClosed      int64         `json:"max_idle_closed"`
	MaxLifetimeClosed  int64         `json:"max_lifetime_closed"`

	// 自定义统计
	QueryCount       int64         `json:"query_count"`
	SlowQueryCount   int64         `json:"slow_query_count"`
	AvgQueryDuration time.Duration `json:"avg_query_duration"`
}

var (
	dbStats        DBStats
	queryCount     int64
	slowQueryCount int64
	totalQueryTime int64
)

// GetOptimizedConfig 根据系统资源获取优化配置
func GetOptimizedConfig() OptimizedDBConfig {
	cpuCount := runtime.NumCPU()

	return OptimizedDBConfig{
		MaxOpenConns:       cpuCount * 10,          // CPU核心数 × 10
		MaxIdleConns:       cpuCount * 2,           // CPU核心数 × 2
		ConnMaxLifetime:    time.Hour,              // 1小时
		ConnMaxIdleTime:    30 * time.Minute,       // 30分钟
		SlowQueryThreshold: 500 * time.Millisecond, // 500ms
		EnableMonitoring:   true,
		StatsInterval:      30 * time.Second,
	}
}

// OptimizeDB 优化数据库连接
func OptimizeDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取底层sql.DB失败: %w", err)
	}

	config := GetOptimizedConfig()

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	log.Printf("数据库连接池已优化: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v",
		config.MaxOpenConns, config.MaxIdleConns, config.ConnMaxLifetime)

	// 启动监控
	if config.EnableMonitoring {
		go startDBMonitoring(sqlDB, config.StatsInterval)
	}

	return nil
}

// startDBMonitoring 启动数据库监控
func startDBMonitoring(sqlDB *sql.DB, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		stats := sqlDB.Stats()

		// 更新统计信息
		dbStats.MaxOpenConnections = stats.MaxOpenConnections
		dbStats.OpenConnections = stats.OpenConnections
		dbStats.InUse = stats.InUse
		dbStats.Idle = stats.Idle
		dbStats.WaitCount = stats.WaitCount
		dbStats.WaitDuration = stats.WaitDuration
		dbStats.MaxIdleClosed = stats.MaxIdleClosed
		dbStats.MaxLifetimeClosed = stats.MaxLifetimeClosed

		// 计算自定义统计
		dbStats.QueryCount = atomic.LoadInt64(&queryCount)
		dbStats.SlowQueryCount = atomic.LoadInt64(&slowQueryCount)

		totalTime := atomic.LoadInt64(&totalQueryTime)
		if dbStats.QueryCount > 0 {
			dbStats.AvgQueryDuration = time.Duration(totalTime / dbStats.QueryCount)
		}

		// 健康检查
		checkDBHealth(stats)
	}
}

// checkDBHealth 检查数据库健康状态
func checkDBHealth(stats sql.DBStats) {
	// 修正连接使用率检查逻辑
	if stats.MaxOpenConnections > 0 {
		// 使用最大连接数作为基准计算使用率
		maxUsageRate := float64(stats.OpenConnections) / float64(stats.MaxOpenConnections)
		// 使用当前打开连接数作为基准计算活跃率
		activeRate := float64(stats.InUse) / float64(stats.OpenConnections)

		// 只有当连接池使用率超过80%时才警告
		if maxUsageRate > 0.8 {
			log.Printf("警告: 数据库连接池使用率过高 %.2f%% (%d/%d)",
				maxUsageRate*100, stats.OpenConnections, stats.MaxOpenConnections)
		}

		// 只有当活跃连接使用率过高且有实际连接在使用时才警告
		if stats.InUse > 0 && activeRate > 0.9 && stats.OpenConnections > 1 {
			log.Printf("警告: 数据库活跃连接使用率过高 %.2f%% (%d/%d活跃)",
				activeRate*100, stats.InUse, stats.OpenConnections)
		}

		// 调试信息：详细连接状态
		log.Printf("数据库连接详情 - 总连接池: %d/%d, 使用中: %d, 空闲: %d, 最大使用率: %.1f%%, 活跃率: %.1f%%",
			stats.OpenConnections, stats.MaxOpenConnections, stats.InUse, stats.Idle,
			maxUsageRate*100, activeRate*100)
	}

	// 等待时间检查
	if stats.WaitDuration > time.Second {
		log.Printf("警告: 数据库连接等待时间过长 %v", stats.WaitDuration)
	}

	// 连接池满检查
	if stats.OpenConnections >= stats.MaxOpenConnections {
		log.Printf("警告: 数据库连接池已满 %d/%d", stats.OpenConnections, stats.MaxOpenConnections)
	}

	// 连接池过小检查（生产环境建议）
	if stats.MaxOpenConnections < 20 {
		log.Printf("建议: 数据库连接池大小较小 (%d)，生产环境建议设置为50-100", stats.MaxOpenConnections)
	}
}

// RecordQuery 记录查询统计
func RecordQuery(duration time.Duration, isSlowQuery bool) {
	atomic.AddInt64(&queryCount, 1)
	atomic.AddInt64(&totalQueryTime, int64(duration))

	if isSlowQuery {
		atomic.AddInt64(&slowQueryCount, 1)
	}
}

// GetDBStats 获取数据库统计信息
func GetDBStats() DBStats {
	return dbStats
}

// QueryStatsMiddleware GORM查询统计中间件
func QueryStatsMiddleware(slowThreshold time.Duration) func(*gorm.DB) {
	return func(db *gorm.DB) {
		start := time.Now()

		// 执行查询
		db.Callback().Query().After("gorm:query").Register("stats:query", func(db *gorm.DB) {
			duration := time.Since(start)
			isSlowQuery := duration > slowThreshold

			RecordQuery(duration, isSlowQuery)

			if isSlowQuery {
				log.Printf("慢查询检测: SQL=%s, 耗时=%v", db.Statement.SQL.String(), duration)
			}
		})

		// 执行更新
		db.Callback().Update().After("gorm:update").Register("stats:update", func(db *gorm.DB) {
			duration := time.Since(start)
			isSlowQuery := duration > slowThreshold

			RecordQuery(duration, isSlowQuery)

			if isSlowQuery {
				log.Printf("慢更新检测: SQL=%s, 耗时=%v", db.Statement.SQL.String(), duration)
			}
		})

		// 执行创建
		db.Callback().Create().After("gorm:create").Register("stats:create", func(db *gorm.DB) {
			duration := time.Since(start)
			isSlowQuery := duration > slowThreshold

			RecordQuery(duration, isSlowQuery)

			if isSlowQuery {
				log.Printf("慢创建检测: SQL=%s, 耗时=%v", db.Statement.SQL.String(), duration)
			}
		})

		// 执行删除
		db.Callback().Delete().After("gorm:delete").Register("stats:delete", func(db *gorm.DB) {
			duration := time.Since(start)
			isSlowQuery := duration > slowThreshold

			RecordQuery(duration, isSlowQuery)

			if isSlowQuery {
				log.Printf("慢删除检测: SQL=%s, 耗时=%v", db.Statement.SQL.String(), duration)
			}
		})
	}
}

// DatabaseHealthCheck 数据库健康检查
func DatabaseHealthCheck(db *gorm.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("数据库连接检查失败: %w", err)
	}

	return nil
}

// GetConnectionPoolHealth 获取连接池健康状态
func GetConnectionPoolHealth(db *gorm.DB) map[string]interface{} {
	sqlDB, err := db.DB()
	if err != nil {
		return map[string]interface{}{
			"error":   err.Error(),
			"healthy": false,
		}
	}

	stats := sqlDB.Stats()

	// 计算健康评分
	healthScore := calculateHealthScore(stats)

	return map[string]interface{}{
		"healthy":      healthScore > 70,
		"health_score": healthScore,
		"stats": map[string]interface{}{
			"max_open":      stats.MaxOpenConnections,
			"open":          stats.OpenConnections,
			"in_use":        stats.InUse,
			"idle":          stats.Idle,
			"wait_count":    stats.WaitCount,
			"wait_duration": stats.WaitDuration.String(),
		},
		"recommendations": getHealthRecommendations(stats),
	}
}

// calculateHealthScore 计算健康评分 (0-100)
func calculateHealthScore(stats sql.DBStats) int {
	score := 100

	// 连接使用率评分 (期望 < 80%)
	if stats.OpenConnections > 0 {
		usageRate := float64(stats.InUse) / float64(stats.OpenConnections)
		if usageRate > 0.8 {
			score -= int((usageRate - 0.8) * 100)
		}
	}

	// 等待时间评分
	if stats.WaitDuration > time.Second {
		score -= 20
	}

	// 连接池满检查
	if stats.OpenConnections >= stats.MaxOpenConnections {
		score -= 30
	}

	if score < 0 {
		score = 0
	}

	return score
}

// getHealthRecommendations 获取健康建议
func getHealthRecommendations(stats sql.DBStats) []string {
	var recommendations []string

	if stats.OpenConnections > 0 {
		usageRate := float64(stats.InUse) / float64(stats.OpenConnections)
		if usageRate > 0.9 {
			recommendations = append(recommendations, "考虑增加MaxOpenConns以支持更高并发")
		}
	}

	if stats.WaitDuration > time.Second {
		recommendations = append(recommendations, "等待时间过长，建议检查查询性能或增加连接数")
	}

	if stats.OpenConnections >= stats.MaxOpenConnections {
		recommendations = append(recommendations, "连接池已满，建议增加MaxOpenConns或优化查询")
	}

	if stats.Idle > stats.MaxOpenConnections/2 {
		recommendations = append(recommendations, "空闲连接过多，可以考虑减少MaxIdleConns")
	}

	return recommendations
}
