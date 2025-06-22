package app_service

import (
	"nasa-go-admin/db"
	"nasa-go-admin/model/app_model"

	"github.com/gin-gonic/gin"
)

type BannerService struct{}

type BannerResponse struct {
	Id       int    `json:"id"`
	Title    string `json:"title"`
	ImageUrl string `json:"image_url"`
	LinkUrl  string `json:"link_url"`
}

// GetActiveBanners 获取所有启用的轮播图
func (s *BannerService) GetActiveBanners(c *gin.Context) ([]BannerResponse, error) {
	var banners []app_model.Banner

	// 只获取状态为启用的banner，并按照排序字段降序排列
	err := db.Dao.Where("status = ?", 1).
		Order("sort DESC, id DESC").
		Find(&banners).Error

	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	response := make([]BannerResponse, 0, len(banners))
	for _, banner := range banners {
		response = append(response, BannerResponse{
			Id:       banner.Id,
			Title:    banner.Title,
			ImageUrl: banner.ImageUrl,
			LinkUrl:  banner.LinkUrl,
		})
	}

	return response, nil
}
