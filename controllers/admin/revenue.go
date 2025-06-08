package admin

import (
	"nasa-go-admin/inout"
	"nasa-go-admin/services/admin_service"
	"nasa-go-admin/utils"
	"strconv"

	"github.com/gin-gonic/gin"
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

// RefreshRevenueStats 手动刷新收益统计数据
func RefreshRevenueStats(c *gin.Context) {
	// 获取租户ID
	parentId, err := utils.GetParentId(c)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// 获取要刷新的天数参数，默认30天
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 || days > 365 {
		days = 30 // 默认30天，最大不超过365天
	}

	// 执行刷新
	err = revenueService.RefreshRevenueStats(parentId, days)
	if err != nil {
		Resp.Err(c, 20001, "刷新统计数据失败: "+err.Error())
		return
	}

	Resp.Succ(c, map[string]interface{}{
		"message": "收益统计数据刷新成功",
		"days":    days,
	})
}
