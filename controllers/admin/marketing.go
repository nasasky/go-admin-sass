package admin

import (
	"fmt"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/services/admin_service"
	"strconv"

	"github.com/gin-gonic/gin"
)

var marketingService = &admin_service.MarketingService{}

func AddMarketing(c *gin.Context) {
	var params inout.AddArticleReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	data := admin_model.Marketing{
		Title:   params.Title,
		Content: params.Content,
		Type:    params.Type,
	}
	Id, err := marketingService.AddMarketing(c, data)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, Id)

}

func GetMarketingList(c *gin.Context) {
	var params inout.GetArticleListReq
	fmt.Println(params)
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	data, err := marketingService.GetMarketingList(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	// 将响应数据存储在上下文中
	Resp.Succ(c, data)
}

func GetMarketingDetail(c *gin.Context) {
	var params inout.GetArticleDetailReq
	fmt.Println(params.Id)
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	data, err := marketingService.GetMarketingDetail(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, data)
}

func UpdateMarketing(c *gin.Context) {
	var params inout.UpdateArticleReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	// data := admin_model.Marketing{
	// 	Id:       params.Id,
	// 	Title:    params.Title,
	// 	Content:  params.Content,
	// 	Type:     params.Type,
	// 	Status:   params.Status,
	// 	Tips:     params.Tips,
	// 	Isdelete: params.Isdelete,
	// }
	err := marketingService.UpdateMarketing(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, nil)
}

// DeleteMarketing 删除文章活动
func DeleteMarketing(c *gin.Context) {
	if c.Query("id") == "" {
		Resp.Err(c, 20001, "id不能为空")
		return
	}
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	err = marketingService.DeleteMarketingId(c, id)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, nil)
}
