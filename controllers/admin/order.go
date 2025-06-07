package admin

import (
	"github.com/gin-gonic/gin"
	"nasa-go-admin/inout"
	"nasa-go-admin/services/admin_service"
)

var OrderService = &admin_service.OrderService{}

func GetOrderList(c *gin.Context) {
	var params inout.OrderListReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	list, err := OrderService.GetOrderList(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, list)

}
