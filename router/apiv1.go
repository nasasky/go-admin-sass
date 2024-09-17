package router

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"nasa-go-admin/api"
	"nasa-go-admin/middleware"
)

func Init(r *gin.Engine) {
	// 使用 cookie 存储会话数据
	r.Use(sessions.Sessions("mysession", cookie.NewStore([]byte("captch"))))
	r.Use(middleware.Cors())

	apiAdminGroup := r.Group("")
	{
		apiAdminGroup.POST("/auth/login", api.Auth.Login)
		apiAdminGroup.GET("/auth/captcha", api.Auth.Captcha)

		apiAdminGroup.Use(middleware.Jwt())
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
