package main

import (
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// 定义监控指标
var (
	// HTTP 请求总数
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP请求总数",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTP 请求耗时
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "HTTP请求耗时",
		},
		[]string{"method", "endpoint"},
	)

	// 业务指标：用户登录次数
	userLogins = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "user_logins_total",
			Help: "用户登录总数",
		},
	)
)

// 监控中间件
func prometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		// 记录监控指标
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			status,
		).Inc()

		httpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)
	}
}

func main() {
	// 创建 Gin 应用
	app := gin.Default()

	// 添加监控中间件
	app.Use(prometheusMiddleware())

	// 暴露 Prometheus 指标端点
	app.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 示例业务接口
	app.GET("/api/users", func(c *gin.Context) {
		// 模拟一些处理时间
		time.Sleep(100 * time.Millisecond)
		c.JSON(200, gin.H{
			"message": "用户列表",
			"data":    []string{"用户1", "用户2"},
		})
	})

	app.POST("/api/login", func(c *gin.Context) {
		// 模拟登录逻辑
		time.Sleep(200 * time.Millisecond)

		// 记录业务指标
		userLogins.Inc()

		c.JSON(200, gin.H{
			"message": "登录成功",
			"token":   "example-token",
		})
	})

	// 模拟一个慢接口
	app.GET("/api/slow", func(c *gin.Context) {
		time.Sleep(2 * time.Second) // 模拟慢查询
		c.JSON(200, gin.H{"message": "慢接口响应"})
	})

	// 模拟错误接口
	app.GET("/api/error", func(c *gin.Context) {
		c.JSON(500, gin.H{"error": "服务器内部错误"})
	})

	log.Println("🚀 服务启动在 :8801")
	log.Println("📊 监控指标: http://localhost:8801/metrics")
	log.Println("🔍 测试接口:")
	log.Println("   GET  /api/users  - 用户列表")
	log.Println("   POST /api/login  - 用户登录")
	log.Println("   GET  /api/slow   - 慢接口(2秒)")
	log.Println("   GET  /api/error  - 错误接口")

	// 启动服务
	if err := app.Run(":8801"); err != nil {
		log.Fatal("服务启动失败:", err)
	}
}

/*
使用方法：

1. 安装依赖：
   go mod init monitoring-example
   go get github.com/gin-gonic/gin
   go get github.com/prometheus/client_golang/prometheus
   go get github.com/prometheus/client_golang/prometheus/promhttp

2. 运行示例：
   go run monitoring_example.go

3. 测试接口：
   curl http://localhost:8801/api/users
   curl -X POST http://localhost:8801/api/login
   curl http://localhost:8801/api/slow
   curl http://localhost:8801/api/error

4. 查看监控指标：
   curl http://localhost:8801/metrics

你会看到类似这样的监控数据：
   # HELP http_requests_total HTTP请求总数
   # TYPE http_requests_total counter
   http_requests_total{endpoint="/api/users",method="GET",status="200"} 1
   http_requests_total{endpoint="/api/login",method="POST",status="200"} 1

   # HELP http_request_duration_seconds HTTP请求耗时
   # TYPE http_request_duration_seconds histogram
   http_request_duration_seconds_bucket{endpoint="/api/users",method="GET",le="0.1"} 0
   http_request_duration_seconds_bucket{endpoint="/api/users",method="GET",le="0.25"} 1

   # HELP user_logins_total 用户登录总数
   # TYPE user_logins_total counter
   user_logins_total 1
*/
