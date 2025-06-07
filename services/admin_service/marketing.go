package admin_service

import (
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"time"

	"github.com/gin-gonic/gin"
)

type MarketingService struct{}

func (s *MarketingService) AddMarketing(c *gin.Context, data admin_model.Marketing) (int, error) {
	userId := c.GetInt("uid")
	data.UserId = userId
	data.CreateTime = time.Now()
	data.UpdateTime = time.Now()
	err := db.Dao.Create(&data).Error
	if err != nil {
		return 0, err
	}
	return data.Id, nil

}

func (s *MarketingService) GetMarketingList(c *gin.Context, params inout.GetArticleListReq) (interface{}, error) {
	var data []admin_model.Marketing
	var total int64

	// 设置默认分页参数
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize

	// 使用单个数据库连接查询总数和分页数据
	err := db.Dao.Model(&admin_model.Marketing{}).Count(&total).Offset(offset).Limit(params.PageSize).Find(&data).Error
	if err != nil {
		return nil, err
	}
	// 格式化时间字段
	formattedData := make([]inout.MarketingItem, len(data))
	for i, item := range data {
		formattedData[i] = inout.MarketingItem{
			Id:         item.Id,
			Title:      item.Title,
			Content:    item.Content,
			Type:       item.Type,
			Status:     item.Status,
			CreateTime: item.CreateTime.Format("2006-01-02 15:04:05"),
			UpdateTime: item.UpdateTime.Format("2006-01-02 15:04:05"),
		}
	}

	// 封装返回结果
	response := inout.MarketingListResponse{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Items:    formattedData,
	}

	return response, nil
}

func (s *MarketingService) GetMarketingDetail(c *gin.Context, params inout.GetArticleDetailReq) (interface{}, error) {
	var data admin_model.Marketing
	err := db.Dao.Where("id = ?", params.Id).First(&data).Error
	if err != nil {
		return nil, err
	}
	// 格式化时间字段
	formattedData := inout.MarketingItem{
		Id:         data.Id,
		Title:      data.Title,
		Content:    data.Content,
		Type:       data.Type,
		Status:     data.Status,
		CreateTime: data.CreateTime.Format("2006-01-02 15:04:05"),
		UpdateTime: data.UpdateTime.Format("2006-01-02 15:04:05"),
	}
	return formattedData, nil
}

func (s *MarketingService) UpdateMarketing(c *gin.Context, params inout.UpdateArticleReq) error {
	data := admin_model.Marketing{
		Id:         params.Id,
		Title:      params.Title,
		Content:    params.Content,
		Type:       params.Type,
		Status:     params.Status,
		Tips:       params.Tips,
		Isdelete:   params.Isdelete,
		UpdateTime: time.Now(),
	}
	err := db.Dao.Model(&admin_model.Marketing{}).Where("id = ?", params.Id).Updates(&data).Error
	if err != nil {
		return err
	}
	return nil
}

// DeleteMarketing
func (s *MarketingService) DeleteMarketingId(c *gin.Context, id int) error {

	err := db.Dao.Where("id = ?", id).Delete(&admin_model.Marketing{}).Error
	if err != nil {
		return err
	}
	return err

}
