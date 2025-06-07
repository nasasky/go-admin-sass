package admin

import (
	"github.com/gin-gonic/gin"
	"nasa-go-admin/inout"
	"nasa-go-admin/services/admin_service"
)

var revenueService = &admin_service.RevenueService{}

// GetRevenueList 获取收入列表
func GetRevenueList(c *gin.Context) {
	var params inout.GetRevenueListReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// 调用公共函数获取 parent_id

	list, err := revenueService.GetRevenueList(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, list)

}
