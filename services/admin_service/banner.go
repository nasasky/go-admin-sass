package admin_service

import (
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type BannerService struct{}

func (s *BannerService) AddBanner(c *gin.Context, data admin_model.Banner) (int, error) {
	parentId, err := utils.GetParentId(c)
	if err != nil {
		return 0, err
	}

	data.TenantsId = parentId
	data.CreateTime = time.Now()
	data.UpdateTime = time.Now()

	err = db.Dao.Create(&data).Error
	if err != nil {
		return 0, err
	}
	return data.Id, nil
}

func (s *BannerService) GetBannerList(c *gin.Context, params inout.GetBannerListReq) (*inout.BannerListResponse, error) {
	var data []admin_model.Banner
	var total int64

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	offset := (params.Page - 1) * params.PageSize
	query := db.Dao.Model(&admin_model.Banner{})

	// Add tenant filter
	parentId, err := utils.GetParentId(c)
	if err != nil {
		return nil, err
	}
	query = query.Where("tenants_id = ?", parentId)

	// Add search condition if provided
	if params.Search != "" {
		query = query.Where("title LIKE ?", "%"+params.Search+"%")
	}

	err = query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	err = query.Order("sort DESC, id DESC").Offset(offset).Limit(params.PageSize).Find(&data).Error
	if err != nil {
		return nil, err
	}

	items := make([]inout.BannerItem, len(data))
	for i, item := range data {
		items[i] = inout.BannerItem{
			Id:         item.Id,
			Title:      item.Title,
			ImageUrl:   item.ImageUrl,
			LinkUrl:    item.LinkUrl,
			Sort:       item.Sort,
			Status:     item.Status,
			CreateTime: item.CreateTime,
			UpdateTime: item.UpdateTime,
		}
	}

	return &inout.BannerListResponse{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Items:    items,
	}, nil
}

func (s *BannerService) GetBannerDetail(c *gin.Context, id int) (*inout.BannerItem, error) {
	var banner admin_model.Banner

	parentId, err := utils.GetParentId(c)
	if err != nil {
		return nil, err
	}

	err = db.Dao.Where("id = ? AND tenants_id = ?", id, parentId).First(&banner).Error
	if err != nil {
		return nil, err
	}

	return &inout.BannerItem{
		Id:         banner.Id,
		Title:      banner.Title,
		ImageUrl:   banner.ImageUrl,
		LinkUrl:    banner.LinkUrl,
		Sort:       banner.Sort,
		Status:     banner.Status,
		CreateTime: banner.CreateTime,
		UpdateTime: banner.UpdateTime,
	}, nil
}

func (s *BannerService) UpdateBanner(c *gin.Context, data admin_model.Banner) error {
	parentId, err := utils.GetParentId(c)
	if err != nil {
		return err
	}

	data.UpdateTime = time.Now()
	err = db.Dao.Model(&admin_model.Banner{}).Where("id = ? AND tenants_id = ?", data.Id, parentId).Updates(&data).Error
	return err
}

func (s *BannerService) DeleteBanner(c *gin.Context, id int) error {
	parentId, err := utils.GetParentId(c)
	if err != nil {
		return err
	}

	return db.Dao.Where("id = ? AND tenants_id = ?", id, parentId).Delete(&admin_model.Banner{}).Error
}
