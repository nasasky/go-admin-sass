package router

import (
	"nasa-go-admin/controllers/admin"
	"nasa-go-admin/controllers/miniapp"
	"nasa-go-admin/controllers/public"
	"nasa-go-admin/inout"
	"nasa-go-admin/middleware"

	"github.com/gin-gonic/gin"
)

// admin接口
func InitAdmin(r *gin.Engine) {
	r.Use(middleware.Cors())

	// 不需要验证 Token 的路由组
	noAuthGroup := r.Group("/api/admin")
	noAuthGroup.Use(middleware.Cors()) // 保留 CORS 中间件
	noAuthGroup.Use(middleware.RequestLogger("request_admin_log"))
	// 在 noAuthGroup 部分添加这两个接口
	noAuthGroup.POST("/miniapp/subscribe", miniapp.SubscribeTemplate) // 用户订阅消息模板
	noAuthGroup.POST("/miniapp/push", miniapp.PushMessage)            // 触发推送消息（可选，如果需要手动触发推送）
	// 在应用程序的路由器设置中
	noAuthGroup.Use(middleware.WebSocketLogger())
	noAuthGroup.GET("/ws", public.WebSocketConnect)
	// 登录相关接口
	noAuthGroup.POST("/login", admin.Login)
	noAuthGroup.POST("/tenants/login", admin.TenantsLogin)

	// 在 InitAdmin 函数中的 noAuthGroup 部分添加
	noAuthGroup.GET("/wechat/verify", public.WechatVerify)

	// 系统字典相关接口
	noAuthGroup.GET("/system/dict/type", middleware.ValidationMiddleware(&inout.ListpageReq{}), admin.GetDictType)
	noAuthGroup.GET("/system/dict/detail", middleware.ValidationMiddleware(&inout.DicteReq{}), admin.GetDictDetail)
	//获取所有字典类型和字典值
	noAuthGroup.GET("/system/dict/all", admin.GetAllDictType)

	// 系统访问日志接口
	noAuthGroup.GET("/system/log", admin.GetSystemLog)

	//获取用户端操作日志
	noAuthGroup.GET("/system/user/log", admin.GetUserLog)

	// 需要验证 Token 的路由组
	authGroup := r.Group("/api/admin")
	authGroup.Use(middleware.AdminJWTAuth()) // 应用统一管理员JWT中间件
	authGroup.Use(middleware.RequestLogger("request_admin_log"))
	authGroup.Use(middleware.UserInfoMiddleware())
	{
		//退出登录
		authGroup.POST("/auth/logout", admin.Logout)

		//发送系统消息通知

		authGroup.POST("/system/notice", admin.PostnoticeInfo)

		authGroup.POST("/tenants/add", middleware.ValidationMiddleware(&inout.AddTenantsReq{}), admin.TenantsRegister) //添加租户
		//编辑租户
		authGroup.PUT("/tenants/update", middleware.ValidationMiddleware(&inout.UpdateTenantsReq{}), admin.UpdateTenants)

		//添加字典类型
		authGroup.POST("/system/dict/type/add", middleware.ValidationMiddleware(&inout.AddDictTypeReq{}), admin.AddDictType)

		//添加字典值
		authGroup.POST("/system/dict/value/add", middleware.ValidationMiddleware(&inout.AddDictValueReq{}), admin.AddDictValue)

		//获取用户信息
		authGroup.GET("/tenants/info", admin.GetUserInfo)
		//获取路由列表
		authGroup.GET("/route", admin.GetRoute)
		//获取路由菜单
		authGroup.GET("/menu", admin.GetMenu)
		//新增菜单
		authGroup.POST("/menu/add", admin.AddMenu)
		//菜单或按钮详情
		authGroup.GET("/menu/detail/:id", admin.GetMenuDetail)
		//修改菜单或按钮
		authGroup.PUT("/menu/update", admin.UpdateMenu)
		//删除菜单或按钮
		authGroup.DELETE("/menu/delete", admin.DeleteMenu)

		//新增角色
		authGroup.POST("/role/add", middleware.ValidationMiddleware(&inout.AddRolexReq{}), admin.AddRole)

		//编辑角色
		authGroup.PUT("/role/update", admin.UpdateRole)

		//设置角色权限
		authGroup.POST("/role/set/permission", middleware.ValidationMiddleware(&inout.SetRolePermissionReq{}), admin.SetRolePermission)

		//获取角色列表
		authGroup.GET("/role/list", admin.GetRoleList)
		//获取所有角色列表详情
		authGroup.GET("/role/all", admin.GetAllRoleList)
		//获取角色详情
		authGroup.GET("/role/detail", admin.GetRoleDetail)

		//删除角色
		authGroup.DELETE("/role/delete/:id", admin.DeleteRole)
		//添加文章活动
		authGroup.POST("/article/add", middleware.ValidationMiddleware(&inout.AddArticleReq{}), admin.AddMarketing)
		//获取文章活动列表
		authGroup.GET("/article/list", admin.GetMarketingList)
		//获取文章活动详情
		authGroup.GET("/article/detail", admin.GetMarketingDetail)
		//修改文章活动
		authGroup.PUT("/article/update", middleware.ValidationMiddleware(&inout.UpdateArticleReq{}), admin.UpdateMarketing)
		//删除文章活动
		authGroup.DELETE("/article/delete", admin.DeleteMarketing)

		//添加商品分类
		authGroup.POST("/goods/category/add", middleware.ValidationMiddleware(&inout.AddGoodsCategoryReq{}), admin.AddGoodsCategory)

		//获取商品分类列表
		authGroup.GET("/goods/category/list", admin.GetGoodsCategoryList)

		//获取商品分类详情
		authGroup.GET("/goods/category/detail/:id", admin.GetGoodsCategoryDetail)

		//修改商品分类

		authGroup.PUT("/goods/category/update", admin.UpdateGoodsCategory)
		//删除商品分类
		authGroup.DELETE("/goods/category/delete/:id", admin.DeleteGoodsCategory)

		//添加商品
		authGroup.POST("/goods/add", middleware.ValidationMiddleware(&inout.AddGoodsReq{}), admin.AddGoods)
		//获取商品列表
		authGroup.GET("/goods/list", admin.GetGoodsList)
		//获取商品详情
		authGroup.GET("/goods/detail/:id", admin.GetGoodsDetail)
		//修改商品
		authGroup.PUT("/goods/update", admin.UpdateGoods)
		//删除商品
		authGroup.DELETE("/goods/delete", admin.DeleteGoods)

		//获取订单列表
		authGroup.GET("/order/list", admin.GetOrderList)

		//获取收益流水统计列表
		authGroup.GET("/order/revenue/list", admin.GetRevenueList)

		//添加员工
		authGroup.POST("/employee/add", admin.AddEmployee)
		//获取员工列表
		authGroup.GET("/employee/list", admin.GetEmployeeList)
		//获取员工详情
		authGroup.GET("/employee/detail/:id", admin.GetEmployeeDetail)
		//修改员工
		authGroup.PUT("/employee/update", admin.UpdateEmployee)
		//删除员工
		authGroup.DELETE("/employee/delete", admin.DeleteEmployee)

		//添加员工组
		authGroup.POST("/employee/group/add", middleware.ValidationMiddleware(&inout.AddEmployeeGroupReq{}), admin.AddEmployeeGroup)
		//获取员工组列表
		authGroup.GET("/employee/group/list", admin.GetEmployeeGroupList)
		//获取员工组详情
		authGroup.GET("/employee/group/detail", admin.GetEmployeeGroupDetail)
		//修改员工组
		authGroup.PUT("/employee/group/update", middleware.ValidationMiddleware(&inout.UpdateEmployeeGroupReq{}), admin.UpdateEmployeeGroup)
		//删除员工组
		authGroup.DELETE("/employee/group/delete", admin.DeleteEmployeeGroup)

		//获取系统会员列表
		authGroup.GET("/member/list", admin.GetMemberList)

		//获取飞书应用配置
		authGroup.GET("/feishu/get", admin.GetFeishuGroupListdata)

		//推送飞书机器人消息
		authGroup.POST("/feishu/send", middleware.ValidationMiddleware(&inout.FeishuSendReq{}), admin.SendFeishuMessage)

		//设置系统参数配置
		authGroup.POST("/system/setting", middleware.ValidationMiddleware(&inout.SettingReq{}), admin.AddSetting)
		//获取系统参数配置列表
		authGroup.GET("/system/setting", admin.GetSetting)
		//获取系统参数配置详情
		authGroup.GET("/system/setting/detail", admin.GetSettingDetail)
		//修改系统参数配置
		authGroup.PUT("/system/setting/update", admin.UpdateSetting)

		//删除系统参数配置
		authGroup.DELETE("/system/setting/delete/:id", admin.DeleteSetting)

		//获取队列消息日志记录列表
		authGroup.GET("/queue/log", admin.GetQueueLogList)

		//ai第三方接口
		authGroup.GET("/ai/chatai", admin.Chatai)

		//上传图片文件等到oss
		authGroup.POST("/upload", admin.UploadFile)

	}
}
