package app

import (
	"nasa-go-admin/inout"
	"nasa-go-admin/services/app_service"
	"nasa-go-admin/utils"

	"github.com/gin-gonic/gin"
)

var orderRefundService = &app_service.OrderRefundService{}

// Refund 申请退款
func Refund(c *gin.Context) {
	var params inout.RefundReq
	if err := c.ShouldBind(&params); err != nil {
		utils.Err(c, utils.ErrCodeInvalidParams, err)
		return
	}
	uid := c.GetInt("uid")
	data, err := orderRefundService.Refund(c, uid, params)
	if err != nil {
		utils.Err(c, utils.ErrCodeInternalError, err)
		return
	}
	utils.Succ(c, data)

}
