package app

import (
	"nasa-go-admin/inout"
	"nasa-go-admin/services/app_service"
	"strconv"

	"github.com/gin-gonic/gin"
)

var goodsService = &app_service.AppGoodsService{}

// GetGoodsList
func GetGoodsList(c *gin.Context) {
	var params inout.GetGoodsListReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	list, err := goodsService.GetGoodsList(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, list)
}

// GetGoodsDetail
func GetGoodsDetail(c *gin.Context) {
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
	detail, err := goodsService.GetDetail(c, id)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, detail)
}
