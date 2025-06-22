package app

import (
	"nasa-go-admin/services/app_service"

	"github.com/gin-gonic/gin"
)

var bannerService = &app_service.BannerService{}

// GetBannerList 获取轮播图列表
func GetBannerList(c *gin.Context) {
	banners, err := bannerService.GetActiveBanners(c)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, banners)
}
