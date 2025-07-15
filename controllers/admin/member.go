package admin

import (
	"nasa-go-admin/inout"
	"nasa-go-admin/services/admin_service"

	"github.com/gin-gonic/gin"
)

var memberService = &admin_service.MemberService{}
var memberStatsService = &admin_service.MemberStatsService{}

func GetMemberList(c *gin.Context) {
	var params inout.GetMemberListReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	data, err := memberService.GetMemberList(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	// 将响应数据存储在上下文中
	Resp.Succ(c, data)
}

// ExportMemberList 导出会员列表
func ExportMemberList(c *gin.Context) {
	var params inout.ExportMemberListReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// 调用服务层导出功能
	fileData, filename, err := memberService.ExportMemberList(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// 设置响应头
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Transfer-Encoding", "binary")

	// 返回文件数据
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileData)
}

// GetMemberStats 获取会员统计数据
func GetMemberStats(c *gin.Context) {
	var params inout.GetMemberStatsReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, "参数错误: "+err.Error())
		return
	}

	// 调用统计服务
	data, err := memberStatsService.GetMemberStats(c, params)
	if err != nil {
		Resp.Err(c, 20001, "获取会员统计数据失败: "+err.Error())
		return
	}

	Resp.Succ(c, data)
}
