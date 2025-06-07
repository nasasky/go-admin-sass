package app

import (
	"nasa-go-admin/inout"
	"nasa-go-admin/redis" // 导入自定义的 redis 包
	"nasa-go-admin/services/app_service"
	"nasa-go-admin/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

var orderService = app_service.NewOrderService(redis.GetClient())

// CreateOrder

func CreateOrder(c *gin.Context) {
	var params inout.CreateOrderReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	uid := c.GetInt("uid")
	data, err := orderService.CreateOrder(c, uid, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

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
