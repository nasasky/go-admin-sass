package admin_service

import (
	"fmt"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GoodsService struct{}

// AddGoods 添加商品
func (s *GoodsService) AddGoods(c *gin.Context, goods admin_model.Goods) (int, error) {

	//添加商品逻辑
	userId := c.GetInt("uid")
	goods.UserId = userId
	goods.CreateTime = time.Now()
	goods.UpdateTime = time.Now()
	err := db.Dao.Create(&goods).Error
	if err != nil {
		return 0, err
	}
	return goods.Id, nil

}

// GetGoodsList 获取商品列表
func (s *GoodsService) GetGoodsList(c *gin.Context, params inout.GetGoodsListReq) (interface{}, error) {
	var data []admin_model.Goods
	var total int64

	// 设置默认分页参数
	params.Page = max(params.Page, 1)
	params.PageSize = max(params.PageSize, 10)

	// 构建查询
	query := db.Dao.Model(&admin_model.Goods{}).Scopes(
		applyGoodsNameFilter(params.GoodsName),
		applyStatusFilter(params.Status),
		applyCategoryFilter(params.CategoryId),
	).Where("isdelete != ?", 1).Order("create_time DESC")

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

// UpdateGoods 更新商品
func (s *GoodsService) UpdateGoods(c *gin.Context, goods admin_model.Goods) (int, error) {
	// 更新商品逻辑
	goods.UpdateTime = time.Now()
	err := db.Dao.Model(&goods).Updates(&goods).Error
	if err != nil {
		return 0, err
	}
	return goods.Id, nil
}

// 商品详情
func (s *GoodsService) GetGoodsDetail(c *gin.Context, id int) (interface{}, error) {
	var goods admin_model.Goods
	err := db.Dao.Where("id = ?", id).First(&goods).Error
	if err != nil {
		return nil, err
	}
	return goods, nil
}

// DeleteGoods 删除商品
// DeleteGoods 删除商品
func (s *GoodsService) DeleteGoods(c *gin.Context, ids []int) error {
	// 检查商品是否存在
	var goods []admin_model.Goods
	err := db.Dao.Where("id IN ?", ids).Find(&goods).Error
	if err != nil {
		return err
	}
	if len(goods) == 0 {
		return fmt.Errorf("未找到要删除的商品")
	}

	// 软删除商品，将 isdelete 字段设置为 1
	err = db.Dao.Model(&admin_model.Goods{}).Where("id IN ?", ids).Update("isdelete", 1).Error
	if err != nil {
		return err
	}
	return nil
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
func formatGoodsData(data []admin_model.Goods) []inout.GoodsItem {
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
			UpdateTime: formatTime(item.CreateTime),
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

// AddGoodsCategory 添加商品分类
func (s *GoodsService) AddGoodsCategory(c *gin.Context, category admin_model.GoodsCategory) (int, error) {
	// 添加商品分类逻辑
	userId := c.GetInt("uid")
	category.UserId = userId
	category.CreateTime = time.Now()
	category.UpdateTime = time.Now()
	err := db.Dao.Create(&category).Error
	if err != nil {
		return 0, err
	}
	return category.Id, nil
}

// GetGoodsCategoryList 获取商品分类列表
func (s *GoodsService) GetGoodsCategoryList(c *gin.Context, params inout.GetGoodsCategoryListReq) (interface{}, error) {
	var data []admin_model.GoodsCategory
	var total int64

	// 设置默认分页参数
	params.Page = max(params.Page, 1)
	params.PageSize = max(params.PageSize, 10)

	// 构建查询
	query := db.Dao.Model(&admin_model.GoodsCategory{}).Where("isdelete != ?", 1).Order("create_time DESC")

	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize

	// 执行查询
	err := query.Count(&total).Offset(offset).Limit(params.PageSize).Find(&data).Error
	if err != nil {
		return nil, err
	}

	// 格式化数据
	formattedData := formatGoodsCategoryData(data)

	// 构建响应
	response := inout.GetGoodsCategoryListResp{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Items:    formattedData,
	}

	return response, nil
}

func formatGoodsCategoryData(data []admin_model.GoodsCategory) []inout.GoodsCategoryItem {
	formattedData := make([]inout.GoodsCategoryItem, len(data))
	for i, item := range data {
		formattedData[i] = inout.GoodsCategoryItem{
			Id:          item.Id,
			Name:        item.Name,
			Description: item.Description,
			Status:      item.Status,
			CreateTime:  formatTime(item.CreateTime),
			UpdateTime:  formatTime(item.UpdateTime),
		}
	}
	return formattedData
}

// GetGoodsCategoryDetail 获取商品分类详情
func (s *GoodsService) GetGoodsCategoryDetail(c *gin.Context, id int) (interface{}, error) {
	var category admin_model.GoodsCategory
	err := db.Dao.Where("id = ?", id).First(&category).Error
	if err != nil {
		return nil, err
	}
	return category, nil
}

// UpdateGoodsCategory 更新商品分类
func (s *GoodsService) UpdateGoodsCategory(c *gin.Context, category admin_model.GoodsCategory) (int, error) {
	// 更新商品分类逻辑
	category.UpdateTime = time.Now()
	err := db.Dao.Model(&category).Updates(&category).Error
	if err != nil {
		return 0, err
	}
	return category.Id, nil
}

// DeleteGoodsCategory 删除商品分类
func (s *GoodsService) DeleteGoodsCategory(c *gin.Context, ids []int) error {
	// 检查商品分类是否存在
	var categories []admin_model.GoodsCategory
	err := db.Dao.Where("id IN ?", ids).Find(&categories).Error
	if err != nil {
		return err
	}
	if len(categories) == 0 {
		return fmt.Errorf("未找到要删除的商品分类")
	}

	// 软删除商品分类，将 isdelete 字段设置为 1
	err = db.Dao.Model(&admin_model.GoodsCategory{}).Where("id IN ?", ids).Update("isdelete", 1).Error
	if err != nil {
		return err
	}
	return nil
}
