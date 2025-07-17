package router

import (
	"nasa-go-admin/api"
	"nasa-go-admin/middleware"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func Init(r *gin.Engine) {
	// 创建 cookie store 并设置选项
	store := cookie.NewStore([]byte("nasa-go-admin-secret-key"))
	store.Options(sessions.Options{
		Path:     "/",                   // cookie 路径
		Domain:   "",                    // 留空以使用当前域名
		MaxAge:   3600,                  // cookie 过期时间（秒）
		Secure:   false,                 // 开发环境设置为 false，生产环境应该设置为 true
		HttpOnly: true,                  // 防止 XSS 攻击
		SameSite: http.SameSiteNoneMode, // 允许跨域
	})

	// 使用配置好的 store
	r.Use(sessions.Sessions("nasa-admin-session", store))
	r.Use(middleware.Cors())

	apiAdminGroup := r.Group("")
	{
		apiAdminGroup.POST("/auth/login", api.Auth.Login)
		apiAdminGroup.GET("/auth/captcha", api.Auth.Captcha)

		apiAdminGroup.Use(middleware.AdminJWTAuth())
		apiAdminGroup.POST("/auth/logout", api.Auth.Logout)
		apiAdminGroup.POST("/auth/password", api.Auth.Logout)

		apiAdminGroup.GET("/user", api.User.List)
		apiAdminGroup.POST("/user", api.User.Add)
		apiAdminGroup.DELETE("/user/:id", api.User.Delete)
		apiAdminGroup.PATCH("/user/password/reset/:id", api.User.Update)
		apiAdminGroup.PATCH("/user/:id", api.User.Update)
		apiAdminGroup.PATCH("/user/profile/:id", api.User.Profile)
		apiAdminGroup.GET("/user/detail", api.User.Detail)

		apiAdminGroup.GET("/role", api.Role.List)
		apiAdminGroup.POST("/role", api.Role.Add)
		apiAdminGroup.PATCH("/role/:id", api.Role.Update)
		apiAdminGroup.DELETE("/role/:id", api.Role.Delete)
		apiAdminGroup.PATCH("/role/users/add/:id", api.Role.AddUser)
		apiAdminGroup.PATCH("/role/users/remove/:id", api.Role.RemoveUser)
		apiAdminGroup.GET("/role/page", api.Role.ListPage)
		apiAdminGroup.GET("/role/permissions/tree", api.Role.PermissionsTree)

		apiAdminGroup.POST("/permission", api.Permissions.Add)
		apiAdminGroup.PATCH("/permission/:id", api.Permissions.PatchPermission)
		apiAdminGroup.DELETE("/permission/:id", api.Permissions.Delete)
		apiAdminGroup.GET("/permission/tree", api.Permissions.List)
		apiAdminGroup.GET("/permission/menu/tree", api.Permissions.List)
	}
}
