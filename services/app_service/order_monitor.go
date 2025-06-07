package app_service

import (
	"context"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// OrderMetrics 订单服务指标
type OrderMetrics struct {
	// 请求计数器
	TotalRequests      int64 `json:"total_requests"`
	SuccessfulRequests int64 `json:"successful_requests"`
	FailedRequests     int64 `json:"failed_requests"`

	// 性能指标
	AverageResponseTime float64 `json:"average_response_time_ms"`
	MaxResponseTime     int64   `json:"max_response_time_ms"`
	MinResponseTime     int64   `json:"min_response_time_ms"`

	// 并发指标
	ActiveConnections int64 `json:"active_connections"`
	PeakConnections   int64 `json:"peak_connections"`

	// 错误指标
	DatabaseErrors       int64 `json:"database_errors"`
	RedisErrors          int64 `json:"redis_errors"`
	LockTimeouts         int64 `json:"lock_timeouts"`
	TransactionRollbacks int64 `json:"transaction_rollbacks"`

	// 业务指标
	OrdersCreated     int64 `json:"orders_created"`
	OrdersCancelled   int64 `json:"orders_cancelled"`
	PaymentsProcessed int64 `json:"payments_processed"`

	// 内存和系统指标
	MemoryUsage    uint64 `json:"memory_usage_bytes"`
	GoroutineCount int    `json:"goroutine_count"`

	LastUpdated time.Time `json:"last_updated"`
	mu          sync.RWMutex
}

// 全局指标实例
var (
	orderMetrics = &OrderMetrics{
		MinResponseTime: 9999999,
		LastUpdated:     time.Now(),
	}
)

// GetOrderMetrics 获取订单服务指标
func GetOrderMetrics() *OrderMetrics {
	orderMetrics.mu.RLock()
	defer orderMetrics.mu.RUnlock()

	// 更新系统指标
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := *orderMetrics
	metrics.MemoryUsage = m.Alloc
	metrics.GoroutineCount = runtime.NumGoroutine()
	metrics.LastUpdated = time.Now()

	return &metrics
}

// RecordRequest 记录请求
func (m *OrderMetrics) RecordRequest(success bool, responseTime time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.AddInt64(&m.TotalRequests, 1)

	if success {
		atomic.AddInt64(&m.SuccessfulRequests, 1)
	} else {
		atomic.AddInt64(&m.FailedRequests, 1)
	}

	// 更新响应时间统计
	responseTimeMs := responseTime.Milliseconds()
	if responseTimeMs > m.MaxResponseTime {
		m.MaxResponseTime = responseTimeMs
	}
	if responseTimeMs < m.MinResponseTime {
		m.MinResponseTime = responseTimeMs
	}

	// 更新平均响应时间（简单移动平均）
	if m.TotalRequests == 1 {
		m.AverageResponseTime = float64(responseTimeMs)
	} else {
		m.AverageResponseTime = (m.AverageResponseTime*float64(m.TotalRequests-1) + float64(responseTimeMs)) / float64(m.TotalRequests)
	}
}

// RecordError 记录错误
func (m *OrderMetrics) RecordError(errorType string) {
	switch errorType {
	case "database":
		atomic.AddInt64(&m.DatabaseErrors, 1)
	case "redis":
		atomic.AddInt64(&m.RedisErrors, 1)
	case "lock_timeout":
		atomic.AddInt64(&m.LockTimeouts, 1)
	case "transaction_rollback":
		atomic.AddInt64(&m.TransactionRollbacks, 1)
	}
}

// RecordBusinessEvent 记录业务事件
func (m *OrderMetrics) RecordBusinessEvent(eventType string) {
	switch eventType {
	case "order_created":
		atomic.AddInt64(&m.OrdersCreated, 1)
	case "order_cancelled":
		atomic.AddInt64(&m.OrdersCancelled, 1)
	case "payment_processed":
		atomic.AddInt64(&m.PaymentsProcessed, 1)
	}
}

// RecordActiveConnection 记录活跃连接
func (m *OrderMetrics) RecordActiveConnection(delta int64) {
	newCount := atomic.AddInt64(&m.ActiveConnections, delta)

	// 更新峰值连接数（无锁方式）
	for {
		current := atomic.LoadInt64(&m.PeakConnections)
		if newCount <= current {
			break
		}
		if atomic.CompareAndSwapInt64(&m.PeakConnections, current, newCount) {
			break
		}
	}
}

// MetricsCollector 指标收集器
type MetricsCollector struct {
	ctx    context.Context
	cancel context.CancelFunc
	ticker *time.Ticker
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector() *MetricsCollector {
	ctx, cancel := context.WithCancel(context.Background())
	return &MetricsCollector{
		ctx:    ctx,
		cancel: cancel,
		ticker: time.NewTicker(30 * time.Second), // 每30秒收集一次指标
	}
}

// Start 启动指标收集
func (mc *MetricsCollector) Start() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("指标收集器发生panic: %v", r)
				// 重启收集器
				time.Sleep(5 * time.Second)
				mc.Start()
			}
		}()

		log.Printf("订单服务指标收集器已启动")

		for {
			select {
			case <-mc.ctx.Done():
				log.Printf("指标收集器已停止")
				return
			case <-mc.ticker.C:
				mc.collectMetrics()
			}
		}
	}()
}

// Stop 停止指标收集
func (mc *MetricsCollector) Stop() {
	if mc.cancel != nil {
		mc.cancel()
	}
	if mc.ticker != nil {
		mc.ticker.Stop()
	}
}

// collectMetrics 收集指标
func (mc *MetricsCollector) collectMetrics() {
	metrics := GetOrderMetrics()

	// 记录关键指标到日志
	log.Printf("订单服务指标 - 总请求: %d, 成功: %d, 失败: %d, 平均响应时间: %.2fms, 活跃连接: %d, 内存: %d bytes, Goroutine: %d",
		metrics.TotalRequests,
		metrics.SuccessfulRequests,
		metrics.FailedRequests,
		metrics.AverageResponseTime,
		metrics.ActiveConnections,
		metrics.MemoryUsage,
		metrics.GoroutineCount,
	)

	// 检查异常情况并报警
	mc.checkAlerts(metrics)
}

// checkAlerts 检查报警条件
func (mc *MetricsCollector) checkAlerts(metrics *OrderMetrics) {
	// 错误率过高报警
	if metrics.TotalRequests > 100 {
		errorRate := float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
		if errorRate > 0.05 { // 错误率超过5%
			log.Printf("⚠️  警告: 订单服务错误率过高 %.2f%%", errorRate*100)
		}
	}

	// 响应时间过长报警
	if metrics.AverageResponseTime > 5000 { // 平均响应时间超过5秒
		log.Printf("⚠️  警告: 订单服务响应时间过长 %.2fms", metrics.AverageResponseTime)
	}

	// Goroutine泄漏报警
	if metrics.GoroutineCount > 1000 {
		log.Printf("⚠️  警告: Goroutine数量过多 %d，可能存在泄漏", metrics.GoroutineCount)
	}

	// 内存使用过高报警
	if metrics.MemoryUsage > 500*1024*1024 { // 超过500MB
		log.Printf("⚠️  警告: 内存使用过高 %d bytes", metrics.MemoryUsage)
	}
}

// RequestTimer 请求计时器
type RequestTimer struct {
	startTime time.Time
	operation string
}

// NewRequestTimer 创建新的请求计时器
func NewRequestTimer(operation string) *RequestTimer {
	orderMetrics.RecordActiveConnection(1)
	return &RequestTimer{
		startTime: time.Now(),
		operation: operation,
	}
}

// Finish 完成计时
func (rt *RequestTimer) Finish(success bool) {
	orderMetrics.RecordActiveConnection(-1)
	duration := time.Since(rt.startTime)
	orderMetrics.RecordRequest(success, duration)

	if !success {
		log.Printf("请求失败: %s, 耗时: %v", rt.operation, duration)
	}
}

// FinishWithError 带错误类型完成计时
func (rt *RequestTimer) FinishWithError(errorType string) {
	orderMetrics.RecordActiveConnection(-1)
	duration := time.Since(rt.startTime)
	orderMetrics.RecordRequest(false, duration)
	orderMetrics.RecordError(errorType)

	log.Printf("请求失败: %s, 错误类型: %s, 耗时: %v", rt.operation, errorType, duration)
}
