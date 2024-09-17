package main

import (
	"github.com/gin-gonic/gin"
	"nasa-go-admin/config"
	"nasa-go-admin/db"
	"nasa-go-admin/redis"
	"nasa-go-admin/router"
	"time"
)

func main() {

	var Loc, _ = time.LoadLocation("Asia/Shanghai")
	time.Local = Loc
	// 初始化日志文件

	redisConfig := config.LoadConfig()
	// 初始化 Redis 客户端
	redis.InitRedis(redisConfig)
	app := gin.Default()
	config.Init()
	db.Init()
	router.Init(app)
	router.InitApp(app)
	router.InitAdmin(app)
	err := app.Run(":8800")
	if err != nil {
		return
	}
}
