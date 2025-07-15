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
	// WebSocket监控端点
	noAuthGroup.GET("/ws/stats", public.WebSocketStats)
	noAuthGroup.GET("/ws/health", public.WebSocketHealth)
	// 登录相关接口
	noAuthGroup.POST("/login", admin.Login)
	noAuthGroup.POST("/tenants/login", admin.TenantsLogin)
	noAuthGroup.GET("/captcha", admin.GetCaptcha) // 添加验证码接口

	// 在 InitAdmin 函数中的 noAuthGroup 部分添加
	noAuthGroup.GET("/wechat/verify", public.WechatVerify)

	// 飞书消息推送接口 - 不需要验证Token
	noAuthGroup.POST("/feishu/send", middleware.ValidationMiddleware(&inout.FeishuSendReq{}), admin.SendFeishuMessage)

	// 系统字典相关接口
	noAuthGroup.GET("/system/dict/type", middleware.ValidationMiddleware(&inout.ListpageReq{}), admin.GetDictType)
	noAuthGroup.GET("/system/dict/detail", middleware.ValidationMiddleware(&inout.DicteReq{}), admin.GetDictDetail)
	//获取所有字典类型和字典值
	noAuthGroup.GET("/system/dict/all", admin.GetAllDictType)

	// 系统信息公开接口
	noAuthGroup.GET("/system/info/first", admin.GetFirstSystemInfo)

	// 内容处理接口（无需验证）
	noAuthGroup.POST("/content/process", admin.ProcessContent)

	// 需要验证 Token 的路由组
	authGroup := r.Group("/api/admin")
	authGroup.Use(middleware.SecureAdminJWTAuth()) // 应用安全的管理员JWT中间件（支持Token黑名单）
	authGroup.Use(middleware.RequestLogger("request_admin_log"))
	authGroup.Use(middleware.UserInfoMiddleware())
	authGroup.Use(middleware.RevokeTokenMiddleware()) // 添加Token撤销中间件

	// 注册资讯news路由
	RegisterNewsRoutes(authGroup)

	// ========== 房间包厢管理接口 ==========
	{
		// 房间管理
		authGroup.POST("/rooms", admin.CreateRoom)
		authGroup.PUT("/rooms", admin.UpdateRoom)
		authGroup.GET("/rooms", admin.GetAdminRoomList)
		authGroup.GET("/rooms/:id", admin.GetAdminRoomDetail)
		authGroup.PUT("/rooms/status", admin.UpdateRoomStatus)
		authGroup.DELETE("/rooms/:id", admin.DeleteRoom)

		// 预订管理
		authGroup.GET("/bookings", admin.GetAdminBookingList)
		authGroup.PUT("/bookings/status", admin.UpdateBookingStatus)

		// 订单状态管理
		authGroup.GET("/bookings/status-info", admin.GetBookingStatusInfo)
		authGroup.POST("/bookings/manual-start", admin.ManualStartBooking)
		authGroup.POST("/bookings/manual-end", admin.ManualEndBooking)

		// 订单状态日志管理
		authGroup.GET("/bookings/logs", admin.GetBookingLogList)
		authGroup.GET("/bookings/logs/statistics", admin.GetBookingLogStatistics)

		// 统计信息
		authGroup.GET("/rooms/statistics", admin.GetRoomStatisticsAdmin)

		// ========== 房间套餐管理接口 ==========
		// 套餐管理
		roomPackageController := admin.NewRoomPackageController()
		authGroup.POST("/rooms/packages", roomPackageController.CreatePackage)
		authGroup.PUT("/rooms/packages", roomPackageController.UpdatePackage)
		authGroup.GET("/rooms/packages", roomPackageController.GetPackageList)
		authGroup.DELETE("/rooms/packages/:id", roomPackageController.DeletePackage)

		// 套餐规则管理
		authGroup.POST("/rooms/package-rules", roomPackageController.CreatePackageRule)
		authGroup.PUT("/rooms/package-rules", roomPackageController.UpdatePackageRule)
		authGroup.GET("/rooms/package-rules", roomPackageController.GetPackageRuleList)
		authGroup.DELETE("/rooms/package-rules/:id", roomPackageController.DeletePackageRule)

		// 特殊日期管理
		authGroup.GET("/rooms/special-dates", roomPackageController.GetSpecialDateList)
		authGroup.POST("/rooms/special-dates", roomPackageController.CreateSpecialDate)
		authGroup.DELETE("/rooms/special-dates/:id", roomPackageController.DeleteSpecialDate)

		// 价格计算接口
		authGroup.POST("/rooms/calculate-price", roomPackageController.CalculatePrice)
	}
	{
		//退出登录
		authGroup.POST("/auth/logout", admin.Logout)

		//发送系统消息通知
		authGroup.POST("/system/notice", admin.PostnoticeInfo)

		// 推送记录管理
		authGroup.GET("/notification/records", admin.GetPushRecordList)
		authGroup.GET("/notification/records/:id", admin.GetPushRecordDetail)
		authGroup.DELETE("/notification/records/:id", admin.DeletePushRecord)
		authGroup.GET("/notification/stats", admin.GetPushRecordStats)
		authGroup.POST("/notification/records/:id/resend", admin.ResendNotification)

		// 管理员用户接收记录管理（新增）
		authGroup.GET("/notification/admin-receive-records", admin.GetAdminUserReceiveRecords)
		authGroup.GET("/notification/messages/:messageID/receive-status", admin.GetMessageReceiveStatus)
		authGroup.POST("/notification/mark-read", admin.MarkMessageAsRead)
		authGroup.POST("/notification/mark-confirmed", admin.MarkMessageAsConfirmed)
		authGroup.POST("/notification/batch-mark-read", admin.BatchMarkAsRead)
		authGroup.GET("/notification/online-users", admin.GetOnlineAdminUsers)
		authGroup.GET("/notification/admin-receive-stats", admin.GetAdminUserReceiveStats)
		authGroup.GET("/notification/user-summary", admin.GetUserNotificationSummary)

		// 离线消息管理（新增）
		authGroup.GET("/notification/offline-messages", admin.GetOfflineMessages)
		authGroup.DELETE("/notification/offline-messages", admin.ClearOfflineMessages)
		// 管理员离线消息管理（新增）
		authGroup.GET("/notification/all-offline-messages", admin.GetAllUsersOfflineMessages)
		authGroup.DELETE("/notification/all-offline-messages", admin.ClearAllUsersOfflineMessages)

		authGroup.POST("/tenants/add", middleware.ValidationMiddleware(&inout.AddTenantsReq{}), admin.TenantsRegister) //添加租户
		//编辑租户
		authGroup.PUT("/tenants/update", middleware.ValidationMiddleware(&inout.UpdateTenantsReq{}), admin.UpdateTenants)

		//添加字典类型
		authGroup.POST("/system/dict/type/add", middleware.ValidationMiddleware(&inout.AddDictTypeReq{}), admin.AddDictType)

		//添加字典值
		authGroup.POST("/system/dict/value/add", middleware.ValidationMiddleware(&inout.AddDictValueReq{}), admin.AddDictValue)

		//获取用户信息
		authGroup.GET("/tenants/info", admin.GetUserInfo)
		//修改用户信息
		authGroup.PUT("/user/profile", admin.UpdateUserProfile)
		//修改用户密码
		authGroup.PUT("/user/password", admin.UpdateUserPassword)
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

		// 系统信息管理
		authGroup.POST("/system/info/add", admin.AddSystemInfo)
		authGroup.PUT("/system/info/update", admin.UpdateSystemInfo)
		authGroup.GET("/system/info", admin.GetSystemInfo)
		authGroup.GET("/system/info/list", admin.GetSystemInfoList)

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
		//手动刷新收益统计数据
		authGroup.POST("/order/revenue/refresh", admin.RefreshRevenueStats)

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
		//导出系统会员列表
		authGroup.GET("/member/export", admin.ExportMemberList)
		//获取会员统计数据（折线图）
		authGroup.GET("/member/stats", admin.GetMemberStats)

		//获取飞书应用配置
		authGroup.GET("/feishu/get", admin.GetFeishuGroupListdata)

		// 推送飞书机器人消息 - 已移至noAuthGroup
		// authGroup.POST("/feishu/send", middleware.ValidationMiddleware(&inout.FeishuSendReq{}), admin.SendFeishuMessage)

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

		//创建宠物平台banner轮播图
		authGroup.POST("/banner/add", middleware.ValidationMiddleware(&inout.AddBannerReq{}), admin.AddBanner)
		//获取轮播图列表
		authGroup.GET("/banner/list", admin.GetBannerList)
		//获取轮播图详情
		authGroup.GET("/banner/detail/:id", admin.GetBannerDetail)
		//修改轮播图
		authGroup.PUT("/banner/update", middleware.ValidationMiddleware(&inout.UpdateBannerReq{}), admin.UpdateBanner)
		//删除轮播图
		authGroup.DELETE("/banner/delete/:id", admin.DeleteBanner)

		// 验证码开关管理
		authGroup.GET("/captcha/status", admin.GetCaptchaStatus)
		authGroup.PUT("/captcha/status", admin.UpdateCaptchaStatus)

		// 系统访问日志接口（需要验证token）
		authGroup.GET("/system/log", admin.GetSystemLog)
		authGroup.DELETE("/system/log", admin.ClearSystemLog)

		// 用户端操作日志接口（需要验证token）
		authGroup.GET("/system/user/log", admin.GetUserLog)
		authGroup.DELETE("/system/user/log", admin.ClearUserLog)

	}
}
