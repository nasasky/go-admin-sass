package main

import (
	"context"
	"fmt"
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
	"nasa-go-admin/services/app_service"
	"nasa-go-admin/services/public_service"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// 构建时注入的变量
var (
	Version            = "dev"
	BuildTime          = "unknown"
	GitCommit          = "unknown"
	GoVersion          = "unknown"
	DefaultServiceName = "nasa-go-admin"
	DefaultRouterMode  = "all"
	DefaultPort        = "8801"
)

// 全局订单系统管理器
var globalOrderSystem *app_service.OrderSystemManager

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// 处理命令行参数
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-version", "--version", "-v":
			fmt.Printf("NASA Go Admin\n")
			fmt.Printf("Version: %s\n", Version)
			fmt.Printf("Build Time: %s\n", BuildTime)
			fmt.Printf("Git Commit: %s\n", GitCommit)
			fmt.Printf("Go Version: %s\n", GoVersion)
			fmt.Printf("Default Service: %s\n", DefaultServiceName)
			fmt.Printf("Default Router Mode: %s\n", DefaultRouterMode)
			fmt.Printf("Default Port: %s\n", DefaultPort)
			return
		case "-help", "--help", "-h":
			fmt.Printf("NASA Go Admin - 微服务架构管理系统\n\n")
			fmt.Printf("Usage: %s [options]\n\n", os.Args[0])
			fmt.Printf("Options:\n")
			fmt.Printf("  -version, -v     显示版本信息\n")
			fmt.Printf("  -help, -h        显示帮助信息\n\n")
			fmt.Printf("Environment Variables:\n")
			fmt.Printf("  SERVICE_NAME     服务名称 (默认: %s)\n", DefaultServiceName)
			fmt.Printf("  ROUTER_MODE      路由模式 (默认: %s)\n", DefaultRouterMode)
			fmt.Printf("  PORT             服务端口 (默认: %s)\n", DefaultPort)
			fmt.Printf("\nAvailable Router Modes:\n")
			fmt.Printf("  all      - 所有路由 (默认)\n")
			fmt.Printf("  admin    - 管理后台路由\n")
			fmt.Printf("  app      - App端路由\n")
			fmt.Printf("  miniapp  - 小程序路由\n")
			fmt.Printf("  monitor  - 监控路由\n")
			return
		}
	}

	// 获取服务模式和端口配置，优先使用构建时的默认值
	serviceName := getEnv("SERVICE_NAME", DefaultServiceName)
	routerMode := getEnv("ROUTER_MODE", DefaultRouterMode)
	port := getEnv("PORT", DefaultPort)

	log.Printf("启动 %s (模式: %s, 端口: %s)...", serviceName, routerMode, port)

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

	// 初始化订单安全系统
	if routerMode == "app" || routerMode == "all" {
		log.Printf("🔐 初始化订单安全系统...")
		globalOrderSystem = app_service.NewOrderSystemManager(redis.GetClient())
		if err := globalOrderSystem.Initialize(); err != nil {
			log.Printf("⚠️ 订单安全系统初始化失败: %v", err)
		} else {
			log.Printf("✅ 订单安全系统初始化成功")
		}
	}

	// 根据服务类型决定是否初始化WebSocket
	if routerMode == "app" || routerMode == "all" {
		// 初始化 WebSocket 服务
		wsService := public_service.GetWebSocketService()
		wsService.InitHub()
	}

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()

	// 添加全局中间件
	app.Use(middleware.Recovery())
	app.Use(middleware.Performance())

	// 添加CORS中间件 - 解决跨域问题
	app.Use(middleware.Cors())

	// 根据服务类型设置不同的限流
	switch routerMode {
	case "admin":
		app.Use(middleware.RateLimit(500)) // 管理后台较低限制
	case "app":
		app.Use(middleware.RateLimit(2000)) // App端较高限制
	case "miniapp":
		app.Use(middleware.RateLimit(1500)) // 小程序中等限制
	default:
		app.Use(middleware.RateLimit(1000)) // 默认限制
	}

	// 添加 Prometheus 监控中间件
	app.Use(monitoring.PrometheusMiddleware())

	// 添加监控指标端点
	app.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 健康检查端点
	app.GET("/health", func(c *gin.Context) {
		healthData := gin.H{
			"service":    serviceName,
			"mode":       routerMode,
			"status":     "healthy",
			"timestamp":  time.Now(),
			"goroutines": goroutinepool.GetPool().GetStats(),
			"database":   database.GetDBStats(),
		}

		// 添加订单系统健康状态
		if globalOrderSystem != nil {
			healthData["order_system"] = globalOrderSystem.GetSystemStatus()
		}

		c.JSON(http.StatusOK, healthData)
	})

	// 根据模式初始化不同的路由
	switch routerMode {
	case "admin":
		log.Printf("初始化管理后台路由...")
		router.InitAdmin(app)
	case "app":
		log.Printf("初始化App端路由...")
		router.InitApp(app)
	case "miniapp":
		log.Printf("初始化小程序路由...")
		// 小程序路由在InitAdmin中，需要提取出来
		router.InitAdmin(app)
	case "monitor":
		log.Printf("初始化监控路由...")
		router.InitMonitoringRoutes(app)
	default:
		log.Printf("初始化所有路由...")
		router.Init(app)
		router.InitApp(app)
		router.InitAdmin(app)
		router.InitMonitoringRoutes(app)
	}

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":" + port,
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

	// 关闭订单系统
	if globalOrderSystem != nil {
		log.Printf("正在关闭订单安全系统...")
		globalOrderSystem.Shutdown()
	}

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
