package app_service

import (
	"fmt"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/app_model"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AppGoodsService struct{}

// GetGoodsList 获取商品列表
func (s *AppGoodsService) GetGoodsList(c *gin.Context, params inout.GetGoodsListReq) (interface{}, error) {
	var data []app_model.AppGoods
	var total int64

	// 设置默认分页参数
	params.Page = max(params.Page, 1)
	params.PageSize = max(params.PageSize, 10)

	// 构建查询
	query := db.Dao.Model(&app_model.AppGoods{}).Scopes(
		applyGoodsNameFilter(params.GoodsName),
		applyStatusFilter(params.Status),
		applyCategoryFilter(params.CategoryId),
	)

	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize

	// 执行查询
	err := query.Count(&total).Offset(offset).Limit(params.PageSize).Find(&data).Error
	if err != nil {
		return nil, err
	}

	// 格式化数据
	formattedData := formatGoodsData(data)

	// 构建响应
	response := inout.GetGoodsListResp{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Items:    formattedData,
	}

	return response, nil

}

// GetDetail 获取商品详情
func (s *AppGoodsService) GetDetail(c *gin.Context, id int) (interface{}, error) {
	var data app_model.AppGoods
	err := db.Dao.Model(&app_model.AppGoods{}).Where("id = ?", id).First(&data).Error
	if err != nil {
		return nil, err
	}
	response := inout.GoodsItem{
		Id:         data.Id,
		GoodsName:  data.GoodsName,
		Content:    data.Content,
		Price:      data.Price,
		Stock:      data.Stock,
		Cover:      data.Cover,
		Status:     data.Status,
		CategoryId: data.CategoryId,
		CreateTime: formatTime(data.CreateTime),
		UpdateTime: formatTime(data.UpdateTime),
	}
	return response, nil
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// applyGoodsNameFilter 返回一个根据商品名称过滤的 Scope
func applyGoodsNameFilter(name string) func(*gorm.DB) *gorm.DB {
	fmt.Println(name)
	return func(db *gorm.DB) *gorm.DB {
		if name != "" {
			return db.Where("goods_name LIKE ?", "%"+name+"%")
		}
		return db
	}
}

// applyStatusFilter 返回一个根据状态过滤的 Scope
// applyStatusFilter 返回一个根据状态过滤的 Scope
// 支持字符串和整数类型的状态搜索
func applyStatusFilter(status interface{}) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if status == nil {
			return db
		}

		switch v := status.(type) {
		case int:
			// 整数类型，直接使用
			if v != 0 {
				return db.Where("status = ?", v)
			}
		case string:
			// 字符串类型，非空时使用
			if v != "" {
				return db.Where("status = ?", v)
			}
		}

		return db
	}
}

// applyCategoryFilter 返回一个根据类别过滤的 Scope
func applyCategoryFilter(category int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if category != 0 {
			return db.Where("category = ?", category)
		}
		return db
	}
}

// formatGoodsData 格式化商品数据
func formatGoodsData(data []app_model.AppGoods) []inout.GoodsItem {
	formattedData := make([]inout.GoodsItem, len(data))
	for i, item := range data {
		formattedData[i] = inout.GoodsItem{
			Id:         item.Id,
			GoodsName:  item.GoodsName,
			Content:    item.Content,
			Price:      item.Price,
			Stock:      item.Stock,
			Cover:      item.Cover,
			Status:     item.Status,
			CategoryId: item.CategoryId,
			CreateTime: formatTime(item.CreateTime),
			UpdateTime: formatTime(item.UpdateTime),
		}
	}
	return formattedData
}

// formatTime 格式化时间，如果时间为零值则返回空字符串
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
