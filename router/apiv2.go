package router

import (
	"github.com/gin-gonic/gin"
	"nasa-go-admin/controllers/app"
	"nasa-go-admin/inout"
	"nasa-go-admin/middleware"
)

// app接口
func InitApp(r *gin.Engine) {
	r.Use(middleware.Cors())
	// 使用通用请求日志中间件
	r.Use(middleware.RequestLogger("request_app_log"))
	apiGroup := r.Group("/api")
	{
		apiGroup.POST("/app/register", middleware.ValidationMiddleware(&inout.AddUserAppReq{}), app.Register)
		//登录
		apiGroup.POST("/app/login", middleware.ValidationMiddleware(&inout.LoginAppReq{}), app.Login)
	}
}
