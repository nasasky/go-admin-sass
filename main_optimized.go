package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	oldConfig "nasa-go-admin/config"
	"nasa-go-admin/controllers/health"
	"nasa-go-admin/middleware"
	"nasa-go-admin/pkg/config"
	"nasa-go-admin/pkg/database"
	"nasa-go-admin/pkg/response"
	"nasa-go-admin/redis"
	"nasa-go-admin/services/public_service"

	"github.com/gin-gonic/gin"
)

func mainOptimized() {
	// 初始化配置
	if err := config.InitConfig(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	cfg := config.GetConfig()
	log.Printf("Starting nasa-go-admin in %s mode on port %s", cfg.Server.Mode, cfg.Server.Port)

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// 初始化Redis - 使用原有的配置结构
	redisConfig := oldConfig.RedisConfig{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}
	if err := redis.InitRedis(redisConfig); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	// 设置时区
	if loc, err := time.LoadLocation("Asia/Shanghai"); err != nil {
		log.Printf("Warning: failed to load timezone: %v", err)
	} else {
		time.Local = loc
	}

	// 初始化WebSocket服务
	wsService := public_service.GetWebSocketService()
	wsService.InitHub()

	// 创建Gin应用
	app := gin.New()

	// 添加全局中间件
	setupMiddleware(app)

	// 设置路由
	setupRoutes(app)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      app,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 启动服务器
	go func() {
		log.Printf("Server starting on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// 优雅关闭
	gracefulShutdown(server)
}

// setupMiddleware 设置中间件
func setupMiddleware(app *gin.Engine) {
	// 基础中间件
	app.Use(middleware.Recovery())
	app.Use(middleware.ErrorHandler())
	app.Use(middleware.RequestID())
	app.Use(middleware.SecureHeaders())

	// CORS中间件
	app.Use(middleware.Cors())

	// 根据配置启用限流
	cfg := config.GetConfig()
	if cfg.Security.EnableRateLimit {
		app.Use(middleware.RateLimitMiddleware(cfg.Security.RateLimit, 60))
	}
}

// setupRoutes 设置路由
func setupRoutes(app *gin.Engine) {
	// 健康检查路由
	healthController := health.NewHealthController()
	healthGroup := app.Group("/health")
	{
		healthGroup.GET("/", healthController.CheckHealth)
		healthGroup.GET("/live", healthController.CheckLiveness)
		healthGroup.GET("/ready", healthController.CheckReadiness)
		healthGroup.GET("/info", healthController.GetSystemInfo)
		healthGroup.GET("/metrics", healthController.GetMetrics)
	}

	// API路由组
	apiGroup := app.Group("/api")
	{
		// 公开路由（无需认证）
		setupPublicRoutes(apiGroup)

		// 管理员路由
		setupAdminRoutes(apiGroup)

		// 应用路由
		setupAppRoutes(apiGroup)
	}

	// 404处理
	app.NoRoute(func(c *gin.Context) {
		response.Error(c, response.NOT_FOUND, "接口不存在")
	})
}

// setupPublicRoutes 设置公开路由
func setupPublicRoutes(group *gin.RouterGroup) {
	// 这里可以添加不需要认证的路由
	group.GET("/ping", func(c *gin.Context) {
		response.Success(c, gin.H{"message": "pong"})
	})
}

// setupAdminRoutes 设置管理员路由
func setupAdminRoutes(group *gin.RouterGroup) {
	adminGroup := group.Group("/admin")

	// 无需认证的管理员路由
	noAuthGroup := adminGroup.Group("/")
	noAuthGroup.Use(middleware.RequestLogger("request_admin_log"))
	{
		// 登录等无需认证的接口
		// 这里集成原有的路由
		_ = noAuthGroup // 暂时标记为已使用，避免编译错误
	}

	// 需要认证的管理员路由
	authGroup := adminGroup.Group("/")
	authGroup.Use(middleware.AdminJWTAuth())
	authGroup.Use(middleware.RequestLogger("request_admin_log"))
	authGroup.Use(middleware.UserInfoMiddleware())
	{
		// 需要认证的管理员接口
		// 这里集成原有的路由
		_ = authGroup // 暂时标记为已使用，避免编译错误
	}
}

// setupAppRoutes 设置应用路由
func setupAppRoutes(group *gin.RouterGroup) {
	appGroup := group.Group("/app")

	// 有日志记录的路由
	logGroup := appGroup.Group("/")
	logGroup.Use(middleware.RequestLogger("request_app_log"))
	{
		// 登录等接口
		_ = logGroup // 暂时标记为已使用，避免编译错误
	}

	// 需要认证的应用路由
	authGroup := logGroup.Group("/")
	authGroup.Use(middleware.AppJWTAuth())
	{
		// 需要认证的应用接口
		_ = authGroup // 暂时标记为已使用，避免编译错误
	}
}

// gracefulShutdown 优雅关闭
func gracefulShutdown(server *http.Server) {
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// 设置关闭超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭服务器
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return
	}

	log.Println("Server exited")
}
