package admin

import (
	"github.com/gin-gonic/gin"
	"nasa-go-admin/inout"
	"nasa-go-admin/services/admin_service"
)

var memberService = &admin_service.MemberService{}

func GetMemberList(c *gin.Context) {
	var params inout.ListpageReq
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
