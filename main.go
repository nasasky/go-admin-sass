package main

import (
	"nasa-go-admin/config"
	"nasa-go-admin/db"
	"nasa-go-admin/middleware"
	"nasa-go-admin/mongodb"
	"nasa-go-admin/pkg/cache"
	"nasa-go-admin/redis"
	"nasa-go-admin/router"
	"nasa-go-admin/services/public_service"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
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

	// 初始化 WebSocket 服务
	wsService := public_service.GetWebSocketService()
	wsService.InitHub()

	// 初始化 MongoDB 客户端
	mongodb.InitMongoDB()

	// 初始化数据库和路由
	db.Init()

	// 初始化缓存
	cache.InitCache()

	app := gin.Default()

	// 添加全局中间件
	app.Use(middleware.Recovery())
	app.Use(middleware.Performance())

	router.Init(app)
	router.InitApp(app)
	router.InitAdmin(app)

	// 启动服务
	if err := app.Run(":8801"); err != nil {
		panic("服务启动失败: " + err.Error())
	}
}
