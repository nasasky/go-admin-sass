package monitoring

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus 指标定义
var (
	// HTTP 请求相关指标
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP请求总数",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP请求耗时分布",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// 数据库相关指标
	dbConnectionsInUse = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_in_use",
			Help: "当前使用中的数据库连接数",
		},
	)

	dbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "数据库查询总数",
		},
		[]string{"operation", "table"},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "数据库查询耗时分布",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 2.5, 5.0, 10.0},
		},
		[]string{"operation", "table"},
	)

	// Redis 相关指标
	redisCommandsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_commands_total",
			Help: "Redis命令执行总数",
		},
		[]string{"command", "status"},
	)

	redisCacheHitRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "redis_cache_hit_rate",
			Help: "Redis缓存命中率",
		},
		[]string{"cache_type"},
	)

	// 业务相关指标
	userRegistrations = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "user_registrations_total",
			Help: "用户注册总数",
		},
	)

	userLogins = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "user_logins_total",
			Help: "用户登录总数",
		},
	)

	activeUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users_current",
			Help: "当前活跃用户数",
		},
	)
)

// PrometheusMiddleware Gin中间件，用于收集HTTP指标
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		// 记录指标
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())

		// Prometheus 指标记录
		httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			statusCode,
		).Inc()

		httpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)

		// MongoDB 存储（异步）
		SaveHTTPMetric(c, duration)
	}
}

// 业务指标记录函数
func RecordUserRegistration() {
	userRegistrations.Inc()
}

func RecordUserLogin() {
	userLogins.Inc()
}

func UpdateActiveUsers(count float64) {
	activeUsers.Set(count)
}

func RecordDBQuery(operation, table string, duration time.Duration) {
	dbQueriesTotal.WithLabelValues(operation, table).Inc()
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

func UpdateDBConnections(inUse int) {
	dbConnectionsInUse.Set(float64(inUse))
}

func RecordRedisCommand(command, status string) {
	redisCommandsTotal.WithLabelValues(command, status).Inc()
}

func UpdateCacheHitRate(cacheType string, hitRate float64) {
	redisCacheHitRate.WithLabelValues(cacheType).Set(hitRate)
}
