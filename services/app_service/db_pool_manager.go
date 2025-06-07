package app_service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"gorm.io/gorm"
)

// DatabasePoolManager 数据库连接池管理器
type DatabasePoolManager struct {
	db                *gorm.DB
	sqlDB             *sql.DB
	config            *DatabasePoolConfig
	healthCheckTicker *time.Ticker
	ctx               context.Context
	cancel            context.CancelFunc
	mu                sync.RWMutex
	lastHealthCheck   time.Time
	isHealthy         bool
}

// DatabasePoolConfig 数据库连接池配置
type DatabasePoolConfig struct {
	MaxOpenConns        int           `json:"max_open_conns"`        // 最大打开连接数
	MaxIdleConns        int           `json:"max_idle_conns"`        // 最大空闲连接数
	ConnMaxLifetime     time.Duration `json:"conn_max_lifetime"`     // 连接最大生存时间
	ConnMaxIdleTime     time.Duration `json:"conn_max_idle_time"`    // 连接最大空闲时间
	HealthCheckInterval time.Duration `json:"health_check_interval"` // 健康检查间隔
	SlowQueryThreshold  time.Duration `json:"slow_query_threshold"`  // 慢查询阈值
}

// GetOptimalDatabasePoolConfig 获取最优数据库连接池配置
func GetOptimalDatabasePoolConfig() *DatabasePoolConfig {
	// 基于CPU核心数动态计算连接池大小
	cpuCount := runtime.NumCPU()

	config := &DatabasePoolConfig{
		MaxOpenConns:        cpuCount * 4,           // CPU核心数的4倍
		MaxIdleConns:        cpuCount * 2,           // CPU核心数的2倍
		ConnMaxLifetime:     time.Hour,              // 连接最长存活1小时
		ConnMaxIdleTime:     10 * time.Minute,       // 空闲连接10分钟后回收
		HealthCheckInterval: 30 * time.Second,       // 每30秒健康检查
		SlowQueryThreshold:  500 * time.Millisecond, // 慢查询阈值500ms
	}

	log.Printf("数据库连接池配置 - MaxOpenConns: %d, MaxIdleConns: %d, CPU核心数: %d",
		config.MaxOpenConns, config.MaxIdleConns, cpuCount)

	return config
}

// NewDatabasePoolManager 创建数据库连接池管理器
func NewDatabasePoolManager(db *gorm.DB, config *DatabasePoolConfig) (*DatabasePoolManager, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库连接为空，无法创建连接池管理器")
	}

	if config == nil {
		config = GetOptimalDatabasePoolConfig()
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取SQL DB失败: %w", err)
	}

	// 应用连接池配置
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	ctx, cancel := context.WithCancel(context.Background())

	manager := &DatabasePoolManager{
		db:                db,
		sqlDB:             sqlDB,
		config:            config,
		ctx:               ctx,
		cancel:            cancel,
		healthCheckTicker: time.NewTicker(config.HealthCheckInterval),
		lastHealthCheck:   time.Now(),
		isHealthy:         true,
	}

	// 启动健康检查
	manager.startHealthCheck()

	log.Printf("数据库连接池管理器已初始化")
	return manager, nil
}

// startHealthCheck 启动健康检查
func (dpm *DatabasePoolManager) startHealthCheck() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("数据库健康检查器发生panic: %v", r)
				// 重启健康检查
				time.Sleep(5 * time.Second)
				dpm.startHealthCheck()
			}
		}()

		log.Printf("数据库健康检查器已启动")

		for {
			select {
			case <-dpm.ctx.Done():
				log.Printf("数据库健康检查器已停止")
				return
			case <-dpm.healthCheckTicker.C:
				dpm.performHealthCheck()
			}
		}
	}()
}

// performHealthCheck 执行健康检查
func (dpm *DatabasePoolManager) performHealthCheck() {
	dpm.mu.Lock()
	defer dpm.mu.Unlock()

	start := time.Now()

	// 执行简单查询测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := dpm.sqlDB.PingContext(ctx)
	duration := time.Since(start)

	dpm.lastHealthCheck = time.Now()

	if err != nil {
		dpm.isHealthy = false
		log.Printf("❌ 数据库健康检查失败: %v, 耗时: %v", err, duration)
		orderMetrics.RecordError("database")
	} else {
		dpm.isHealthy = true
		if duration > dpm.config.SlowQueryThreshold {
			log.Printf("⚠️  数据库健康检查响应较慢: %v", duration)
		}
	}

	// 记录连接池统计信息
	stats := dpm.sqlDB.Stats()

	log.Printf("数据库连接池状态 - 打开连接: %d/%d, 使用中: %d, 空闲: %d, 等待: %d, 健康: %v",
		stats.OpenConnections, dpm.config.MaxOpenConns,
		stats.InUse, stats.Idle, stats.WaitCount, dpm.isHealthy)

	// 检查连接池异常情况
	dpm.checkPoolAlerts(stats)
}

// checkPoolAlerts 检查连接池报警条件
func (dpm *DatabasePoolManager) checkPoolAlerts(stats sql.DBStats) {
	// 连接数接近上限报警
	utilizationRate := float64(stats.OpenConnections) / float64(dpm.config.MaxOpenConns)
	if utilizationRate > 0.8 {
		log.Printf("⚠️  警告: 数据库连接池使用率过高 %.1f%% (%d/%d)",
			utilizationRate*100, stats.OpenConnections, dpm.config.MaxOpenConns)
	}

	// 等待连接过多报警
	if stats.WaitCount > 10 {
		log.Printf("⚠️  警告: 数据库连接池等待队列过长 %d", stats.WaitCount)
	}

	// 连接超时报警
	if stats.WaitDuration > time.Second {
		log.Printf("⚠️  警告: 数据库连接等待时间过长 %v", stats.WaitDuration)
	}
}

// GetHealthStatus 获取健康状态
func (dpm *DatabasePoolManager) GetHealthStatus() map[string]interface{} {
	dpm.mu.RLock()
	defer dpm.mu.RUnlock()

	stats := dpm.sqlDB.Stats()

	return map[string]interface{}{
		"healthy":             dpm.isHealthy,
		"last_check":          dpm.lastHealthCheck,
		"open_connections":    stats.OpenConnections,
		"max_open_conns":      dpm.config.MaxOpenConns,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration_ms":    stats.WaitDuration.Milliseconds(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}
}

// ExecuteWithRetry 带重试的数据库执行
func (dpm *DatabasePoolManager) ExecuteWithRetry(operation func(*gorm.DB) error, maxRetries int) error {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		// 检查健康状态
		if !dpm.IsHealthy() && i == 0 {
			log.Printf("数据库不健康，尝试等待恢复...")
			time.Sleep(time.Second)
		}

		err := operation(dpm.db)
		if err == nil {
			return nil
		}

		lastErr = err

		// 检查是否为可重试的错误
		if !dpm.isRetryableError(err) {
			return err
		}

		if i < maxRetries {
			waitTime := time.Duration(i+1) * 100 * time.Millisecond
			log.Printf("数据库操作失败，第%d次重试，等待%v: %v", i+1, waitTime, err)
			time.Sleep(waitTime)
		}
	}

	return fmt.Errorf("数据库操作失败，已重试%d次: %w", maxRetries, lastErr)
}

// isRetryableError 判断是否为可重试的错误
func (dpm *DatabasePoolManager) isRetryableError(err error) bool {
	// 检查常见的可重试错误
	errStr := err.Error()
	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"deadlock",
		"lock wait timeout",
		"server has gone away",
	}

	for _, retryableErr := range retryableErrors {
		if contains(errStr, retryableErr) {
			return true
		}
	}

	return false
}

// contains 检查字符串是否包含子字符串（忽略大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && s[:len(substr)] == substr) ||
		(len(s) > len(substr) && s[len(s)-len(substr):] == substr) ||
		(len(s) > len(substr) && findSubstring(s, substr)))
}

// findSubstring 查找子字符串
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// IsHealthy 检查数据库是否健康
func (dpm *DatabasePoolManager) IsHealthy() bool {
	dpm.mu.RLock()
	defer dpm.mu.RUnlock()
	return dpm.isHealthy
}

// Close 关闭数据库连接池管理器
func (dpm *DatabasePoolManager) Close() {
	if dpm.cancel != nil {
		dpm.cancel()
	}
	if dpm.healthCheckTicker != nil {
		dpm.healthCheckTicker.Stop()
	}
	log.Printf("数据库连接池管理器已关闭")
}

// QueryOptimizer 查询优化器
type QueryOptimizer struct {
	slowQueryLog map[string]time.Duration
	mu           sync.RWMutex
}

// NewQueryOptimizer 创建查询优化器
func NewQueryOptimizer() *QueryOptimizer {
	return &QueryOptimizer{
		slowQueryLog: make(map[string]time.Duration),
	}
}

// LogSlowQuery 记录慢查询
func (qo *QueryOptimizer) LogSlowQuery(query string, duration time.Duration) {
	qo.mu.Lock()
	defer qo.mu.Unlock()

	// 只保留最慢的查询记录，避免内存泄漏
	if len(qo.slowQueryLog) > 1000 {
		// 清理一半的记录
		count := 0
		for key := range qo.slowQueryLog {
			delete(qo.slowQueryLog, key)
			count++
			if count >= 500 {
				break
			}
		}
	}

	// 记录或更新慢查询
	if existingDuration, exists := qo.slowQueryLog[query]; !exists || duration > existingDuration {
		qo.slowQueryLog[query] = duration
	}

	log.Printf("🐌 慢查询记录 - 耗时: %v, SQL: %.100s...", duration, query)
}

// GetSlowQueries 获取慢查询列表
func (qo *QueryOptimizer) GetSlowQueries() map[string]time.Duration {
	qo.mu.RLock()
	defer qo.mu.RUnlock()

	result := make(map[string]time.Duration)
	for query, duration := range qo.slowQueryLog {
		result[query] = duration
	}

	return result
}
