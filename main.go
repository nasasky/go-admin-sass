package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nasa-go-admin/config"
	"nasa-go-admin/db"
	"nasa-go-admin/middleware"
	"nasa-go-admin/mongodb"
	"nasa-go-admin/pkg/cache"
	"nasa-go-admin/pkg/database"
	"nasa-go-admin/pkg/goroutinepool"
	"nasa-go-admin/pkg/monitoring"
	"nasa-go-admin/redis"
	"nasa-go-admin/router"
	"nasa-go-admin/services/public_service"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.Printf("启动NASA Go Admin (优化版本)...")

	// 初始化 Redis 客户端
	redisConfig := config.LoadConfig()
	redis.InitRedis(redisConfig)

	// 初始化配置
	config.InitConfig()

	// 设置时区
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic("无法加载时区: " + err.Error())
	}
	time.Local = loc

	// 初始化数据库和路由
	db.Init()

	// 优化数据库连接池
	if err := database.OptimizeDB(db.Dao); err != nil {
		log.Printf("数据库优化失败: %v", err)
	}

	// 初始化 MongoDB 客户端
	mongodb.InitMongoDB()

	// 初始化缓存
	cache.InitCache()

	// 初始化 WebSocket 服务
	wsService := public_service.GetWebSocketService()
	wsService.InitHub()

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()

	// 添加全局中间件
	app.Use(middleware.Recovery())
	app.Use(middleware.Performance())
	app.Use(middleware.RateLimit(1000)) // 每分钟1000请求限制

	// 添加 Prometheus 监控中间件
	app.Use(monitoring.PrometheusMiddleware())

	// 添加监控指标端点
	app.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 健康检查端点
	app.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":     "healthy",
			"timestamp":  time.Now(),
			"goroutines": goroutinepool.GetPool().GetStats(),
			"database":   database.GetDBStats(),
		})
	})

	router.Init(app)
	router.InitApp(app)
	router.InitAdmin(app)
	router.InitMonitoringRoutes(app)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":8801",
		Handler:      app,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器
	go func() {
		log.Printf("服务器启动在端口 :8801")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("正在关闭服务器...")

	// 设置关闭超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭HTTP服务器
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("服务器强制关闭: %v", err)
	}

	// 关闭goroutine池
	goroutinepool.Stop()

	// 关闭Redis连接
	redis.CloseRedis()

	log.Printf("服务器已安全关闭")
}
