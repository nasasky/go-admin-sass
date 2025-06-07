package monitoring

import (
	"runtime"
	"sync"
	"time"
)

// Metrics 应用指标结构
type Metrics struct {
	mu sync.RWMutex

	// HTTP 指标
	HTTPRequestsTotal   map[string]int64         `json:"http_requests_total"`
	HTTPRequestDuration map[string]time.Duration `json:"http_request_duration"`
	HTTPResponseSizes   map[string]int64         `json:"http_response_sizes"`

	// 数据库指标
	DBConnectionsActive int64         `json:"db_connections_active"`
	DBConnectionsMax    int64         `json:"db_connections_max"`
	DBQueryDuration     time.Duration `json:"db_query_duration"`
	DBQueryCount        int64         `json:"db_query_count"`

	// Redis 指标
	RedisConnectionsActive int64   `json:"redis_connections_active"`
	RedisHitRate           float64 `json:"redis_hit_rate"`
	RedisCommandsTotal     int64   `json:"redis_commands_total"`

	// 系统指标
	MemoryUsage    int64         `json:"memory_usage"`
	GoroutineCount int64         `json:"goroutine_count"`
	GCDuration     time.Duration `json:"gc_duration"`
}

var (
	globalMetrics *Metrics
	metricsOnce   sync.Once
)

// GetMetrics 获取全局指标实例
func GetMetrics() *Metrics {
	metricsOnce.Do(func() {
		globalMetrics = &Metrics{
			HTTPRequestsTotal:   make(map[string]int64),
			HTTPRequestDuration: make(map[string]time.Duration),
			HTTPResponseSizes:   make(map[string]int64),
		}

		// 启动系统指标收集
		go globalMetrics.collectSystemMetrics()
	})
	return globalMetrics
}

// IncrementHTTPRequest 增加HTTP请求计数
func (m *Metrics) IncrementHTTPRequest(method, path string, duration time.Duration, size int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := method + " " + path
	m.HTTPRequestsTotal[key]++
	m.HTTPRequestDuration[key] = duration
	m.HTTPResponseSizes[key] = size
}

// IncrementDBQuery 增加数据库查询计数
func (m *Metrics) IncrementDBQuery(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DBQueryCount++
	m.DBQueryDuration = duration
}

// UpdateRedisMetrics 更新Redis指标
func (m *Metrics) UpdateRedisMetrics(connections int64, hitRate float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RedisConnectionsActive = connections
	m.RedisHitRate = hitRate
	m.RedisCommandsTotal++
}

// collectSystemMetrics 收集系统指标
func (m *Metrics) collectSystemMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		m.mu.Lock()
		m.MemoryUsage = int64(memStats.Alloc)
		m.GoroutineCount = int64(runtime.NumGoroutine())
		m.GCDuration = time.Duration(memStats.PauseTotalNs)
		m.mu.Unlock()
	}
}

// GetSnapshot 获取指标快照
func (m *Metrics) GetSnapshot() *Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshot := &Metrics{
		HTTPRequestsTotal:      make(map[string]int64),
		HTTPRequestDuration:    make(map[string]time.Duration),
		HTTPResponseSizes:      make(map[string]int64),
		DBConnectionsActive:    m.DBConnectionsActive,
		DBConnectionsMax:       m.DBConnectionsMax,
		DBQueryDuration:        m.DBQueryDuration,
		DBQueryCount:           m.DBQueryCount,
		RedisConnectionsActive: m.RedisConnectionsActive,
		RedisHitRate:           m.RedisHitRate,
		RedisCommandsTotal:     m.RedisCommandsTotal,
		MemoryUsage:            m.MemoryUsage,
		GoroutineCount:         m.GoroutineCount,
		GCDuration:             m.GCDuration,
	}

	// 复制map数据
	for k, v := range m.HTTPRequestsTotal {
		snapshot.HTTPRequestsTotal[k] = v
	}
	for k, v := range m.HTTPRequestDuration {
		snapshot.HTTPRequestDuration[k] = v
	}
	for k, v := range m.HTTPResponseSizes {
		snapshot.HTTPResponseSizes[k] = v
	}

	return snapshot
}

// Reset 重置指标
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.HTTPRequestsTotal = make(map[string]int64)
	m.HTTPRequestDuration = make(map[string]time.Duration)
	m.HTTPResponseSizes = make(map[string]int64)
	m.DBQueryCount = 0
	m.RedisCommandsTotal = 0
}
