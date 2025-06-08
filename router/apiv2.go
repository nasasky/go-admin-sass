package router

import (
	"nasa-go-admin/controllers/app"
	"nasa-go-admin/inout"
	"nasa-go-admin/middleware"

	"github.com/gin-gonic/gin"
)

// app接口
func InitApp(r *gin.Engine) {
	r.Use(middleware.Cors())
	appGroup := r.Group("/api/app")

	// 不需要验证和不记录日志的公开接口
	publicGroup := appGroup.Group("/")
	{
		//商品列表
		publicGroup.GET("/goods/list", app.GetGoodsList)
		//商品详情
		publicGroup.GET("/goods/detail", app.GetGoodsDetail)
		// 订单系统健康检查（公开接口，用于监控）
		publicGroup.GET("/order/health", app.GetOrderHealthStatus)
	}

	// 使用通用请求日志中间件的组
	logGroup := appGroup.Group("/")
	logGroup.Use(middleware.RequestLogger("request_app_log"))
	{
		//微信小程序登录
		logGroup.POST("/wx/login", app.WxLogin)
		logGroup.POST("/register", middleware.ValidationMiddleware(&inout.AddUserAppReq{}), app.Register)
		//登录
		logGroup.POST("/login", middleware.ValidationMiddleware(&inout.LoginAppReq{}), app.Login)

		// ========== 房间查看相关接口（无需登录，但记录日志） ==========
		// 房间列表
		logGroup.GET("/rooms", app.GetRoomList)
		// 房间详情
		logGroup.GET("/rooms/:id", app.GetRoomDetail)
		// 检查房间可用性
		logGroup.POST("/rooms/check-availability", app.CheckRoomAvailability)
		// 房间统计信息
		logGroup.GET("/rooms/statistics", app.GetRoomStatistics)
		// 获取房间可用套餐
		logGroup.GET("/rooms/packages", app.GetRoomPackages)

		// 需要JWT验证的接口组
		authGroup := logGroup.Group("/")
		authGroup.Use(middleware.AppJWTAuth())
		{
			//用户信息
			authGroup.GET("/user/info", app.GetUserInfo)
			//修改用户信息
			authGroup.POST("/user/update", app.UpdateUserInfo)
			//刷新token
			authGroup.GET("/refresh", app.Refresh)
			//用户钱包
			authGroup.GET("/user/wallet", app.GetUserWallet)
			//用户充值
			authGroup.POST("/user/recharge", middleware.ValidationMiddleware(&inout.RechargeReq{}), app.Recharge)
			//创建订单（现在使用安全订单创建器）
			authGroup.POST("/order/create", middleware.ValidationMiddleware(&inout.CreateOrderReq{}), app.CreateOrder)
			//我的订单列表
			authGroup.GET("/order/list", middleware.ValidationMiddleware(&inout.MyOrderReq{}), app.GetMyOrderList)
			//订单详情
			authGroup.GET("/order/detail", app.GetOrderDetail)
			//申请退款
			authGroup.POST("/order/refund", app.Refund)

			// ========== 房间预订相关接口（需要登录） ==========
			// 预订管理
			authGroup.POST("/bookings", app.CreateBooking)
			authGroup.GET("/bookings", app.GetMyBookingList)
			authGroup.GET("/bookings/:id", app.GetBookingDetail)
			authGroup.POST("/bookings/cancel", app.CancelBooking)
			// 预订价格预览
			authGroup.POST("/bookings/price-preview", app.BookingPricePreview)
		}
	}
}
