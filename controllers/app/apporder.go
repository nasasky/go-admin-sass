package app

import (
	"log"
	"nasa-go-admin/inout"
	"nasa-go-admin/redis" // 导入自定义的 redis 包
	"nasa-go-admin/services/app_service"
	"nasa-go-admin/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

var orderService = app_service.NewOrderService(redis.GetClient())

// 添加安全订单创建器
var secureOrderCreator *app_service.SecureOrderCreator

// 初始化安全订单创建器
func init() {
	secureOrderCreator = app_service.NewSecureOrderCreator(redis.GetClient())
}

// CreateOrder - 使用安全订单创建器
func CreateOrder(c *gin.Context) {
	var params inout.CreateOrderReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	uid := c.GetInt("uid")

	// 使用安全订单创建器
	data, err := secureOrderCreator.CreateOrderSecurely(c, uid, params)
	if err != nil {
		// 记录详细错误日志
		log.Printf("安全订单创建失败: uid=%d, params=%+v, error=%v", uid, params, err)
		Resp.Err(c, 20001, err.Error())
		return
	}

	// 记录成功日志
	log.Printf("安全订单创建成功: uid=%d, orderID=%v", uid, data)
	Resp.Succ(c, data)
}

// GetMyOrderList 获取我的订单列表
func GetMyOrderList(c *gin.Context) {
	var params inout.MyOrderReq
	if err := c.ShouldBind(&params); err != nil {
		utils.Err(c, utils.ErrCodeInvalidParams, err)
		return
	}
	uid := c.GetInt("uid")
	data, err := orderService.GetMyOrderList(c, uid, params)
	if err != nil {
		utils.Err(c, utils.ErrCodeInternalError, err)
		return
	}
	utils.Succ(c, data)
}

// GetOrderDetail 获取订单详情
func GetOrderDetail(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		utils.Err(c, utils.ErrCodeInvalidParams, utils.NewError("id不能为空"))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Err(c, utils.ErrCodeInvalidParams, err)
		return
	}

	uid := c.GetInt("uid")
	data, err := orderService.GetOrderDetail(c, uid, id)
	if err != nil {
		utils.Err(c, utils.ErrCodeInternalError, err)
		return
	}

	utils.Succ(c, data)
}

// GetOrderHealthStatus 获取订单系统健康状态 - 新增接口
func GetOrderHealthStatus(c *gin.Context) {
	if secureOrderCreator == nil {
		utils.Err(c, utils.ErrCodeInternalError, utils.NewError("安全订单创建器未初始化"))
		return
	}

	// 简化的健康状态检查
	status := gin.H{
		"status":      "healthy",
		"service":     "secure_order_creator",
		"initialized": secureOrderCreator != nil,
	}
	utils.Succ(c, status)
}
