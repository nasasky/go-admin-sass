package admin

import (
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/services/admin_service"
	"strconv"

	"github.com/gin-gonic/gin"
)

var bannerService = &admin_service.BannerService{}

// AddBanner 添加轮播图
func AddBanner(c *gin.Context) {
	var params inout.AddBannerReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	banner := admin_model.Banner{
		Title:    params.Title,
		ImageUrl: params.ImageUrl,
		LinkUrl:  params.LinkUrl,
		Sort:     params.Sort,
		Status:   params.Status,
	}

	id, err := bannerService.AddBanner(c, banner)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, id)
}

// GetBannerList 获取轮播图列表
func GetBannerList(c *gin.Context) {
	var params inout.GetBannerListReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	list, err := bannerService.GetBannerList(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, list)
}

// GetBannerDetail 获取轮播图详情
func GetBannerDetail(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		Resp.Err(c, 20001, "id不能为空")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	detail, err := bannerService.GetBannerDetail(c, id)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, detail)
}

// UpdateBanner 更新轮播图
func UpdateBanner(c *gin.Context) {
	var params inout.UpdateBannerReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	banner := admin_model.Banner{
		Id:       params.Id,
		Title:    params.Title,
		ImageUrl: params.ImageUrl,
		LinkUrl:  params.LinkUrl,
		Sort:     params.Sort,
		Status:   params.Status,
	}

	err := bannerService.UpdateBanner(c, banner)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, nil)
}

// DeleteBanner 删除轮播图
func DeleteBanner(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		Resp.Err(c, 20001, "id不能为空")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	err = bannerService.DeleteBanner(c, id)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, nil)
}
