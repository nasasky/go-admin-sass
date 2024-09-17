package router

import (
	"github.com/gin-gonic/gin"
	"nasa-go-admin/controllers/admin"
	"nasa-go-admin/inout"
	"nasa-go-admin/middleware"
)

// admin接口
func InitAdmin(r *gin.Engine) {
	r.Use(middleware.Cors())
	// 使用通用请求日志中间件
	r.Use(middleware.RequestLogger("request_admin_log"))
	apiGroup := r.Group("/api")
	{
		apiGroup.POST("/admin/tenants/add", middleware.ValidationMiddleware(&inout.AddTenantsReq{}), admin.TenantsRegister) //添加租户
		//登录
		apiGroup.POST("/admin/login", admin.Login)
		//商户登录
		apiGroup.POST("/admin/tenants/login", admin.TenantsLogin)

		apiGroup.Use(middleware.Jwt())
		//获取路由列表
		apiGroup.GET("/admin/route", admin.GetRoute)
	}
}
