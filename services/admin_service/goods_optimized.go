package admin_service

import (
	"context"
	"fmt"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/pkg/cache"
	"nasa-go-admin/redis"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// OptimizedGoodsService 优化的商品服务
type OptimizedGoodsService struct {
	cache *cache.CacheManager
}

// NewOptimizedGoodsService 创建优化的商品服务
func NewOptimizedGoodsService() *OptimizedGoodsService {
	return &OptimizedGoodsService{
		cache: cache.NewCacheManager(redis.GetClient()),
	}
}

// GetGoodsListOptimized 获取商品列表（优化版本）
func (s *OptimizedGoodsService) GetGoodsListOptimized(c *gin.Context, params inout.GetGoodsListReq) (interface{}, error) {
	// 1. 生成缓存键
	cacheKey := s.generateGoodsCacheKey(params)

	// 2. 尝试从缓存获取
	var cached inout.GetGoodsListResp
	if err := s.cache.Get(c, cacheKey, &cached); err == nil {
		return cached, nil
	}

	// 3. 从数据库查询
	result, err := s.fetchGoodsFromDB(c, params)
	if err != nil {
		return nil, err
	}

	// 4. 缓存结果（5分钟）
	s.cache.Set(c, cacheKey, result, 5*time.Minute)

	return result, nil
}

// fetchGoodsFromDB 从数据库获取商品数据
func (s *OptimizedGoodsService) fetchGoodsFromDB(c *gin.Context, params inout.GetGoodsListReq) (*inout.GetGoodsListResp, error) {
	var data []admin_model.Goods
	var total int64

	// 设置默认分页参数
	params.Page = maxInt(params.Page, 1)
	params.PageSize = maxInt(params.PageSize, 10)

	// 构建查询，只选择需要的字段
	query := db.Dao.WithContext(c).Model(&admin_model.Goods{}).
		Select("id, goods_name, content, price, stock, cover, status, category_id, create_time, update_time").
		Where("isdelete != ?", 1).
		Scopes(
			applyGoodsNameFilterOptimized(params.GoodsName),
			applyStatusFilterOptimized(params.Status),
			applyCategoryFilterOptimized(params.CategoryId),
		).
		Order("create_time DESC")

	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize

	// 执行查询
	err := query.Count(&total).Offset(offset).Limit(params.PageSize).Find(&data).Error
	if err != nil {
		return nil, fmt.Errorf("查询商品列表失败: %w", err)
	}

	// 格式化数据
	formattedData := formatGoodsDataOptimized(data)

	// 构建响应
	result := &inout.GetGoodsListResp{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Items:    formattedData,
	}

	return result, nil
}

// BatchGetGoodsByIDs 批量获取商品信息（优化版本）
func (s *OptimizedGoodsService) BatchGetGoodsByIDs(ctx context.Context, goodsIDs []int) (map[int]admin_model.Goods, error) {
	if len(goodsIDs) == 0 {
		return make(map[int]admin_model.Goods), nil
	}

	// 去重
	uniqueIDs := make(map[int]struct{})
	for _, id := range goodsIDs {
		uniqueIDs[id] = struct{}{}
	}

	result := make(map[int]admin_model.Goods)
	var missedIDs []int

	// 1. 尝试从缓存批量获取
	for id := range uniqueIDs {
		cacheKey := fmt.Sprintf("goods:detail:%d", id)
		var goods admin_model.Goods
		if err := s.cache.Get(ctx, cacheKey, &goods); err == nil {
			result[id] = goods
		} else {
			missedIDs = append(missedIDs, id)
		}
	}

	// 2. 对于缓存未命中的商品，批量查询数据库
	if len(missedIDs) > 0 {
		var goodsList []admin_model.Goods
		err := db.Dao.WithContext(ctx).
			Select("id, goods_name, content, price, stock, cover, status, category_id, create_time, update_time").
			Where("id IN ? AND isdelete != ?", missedIDs, 1).
			Find(&goodsList).Error

		if err != nil {
			return nil, fmt.Errorf("批量查询商品失败: %w", err)
		}

		// 添加到结果中并异步缓存
		for _, goods := range goodsList {
			result[goods.Id] = goods

			// 异步缓存单个商品
			go func(g admin_model.Goods) {
				cacheKey := fmt.Sprintf("goods:detail:%d", g.Id)
				s.cache.Set(context.Background(), cacheKey, g, 10*time.Minute)
			}(goods)
		}
	}

	return result, nil
}

// GetGoodsDetailOptimized 获取商品详情（优化版本）
func (s *OptimizedGoodsService) GetGoodsDetailOptimized(ctx context.Context, id int) (*admin_model.Goods, error) {
	// 1. 尝试从缓存获取
	cacheKey := fmt.Sprintf("goods:detail:%d", id)
	var goods admin_model.Goods
	if err := s.cache.Get(ctx, cacheKey, &goods); err == nil {
		return &goods, nil
	}

	// 2. 从数据库查询
	err := db.Dao.WithContext(ctx).
		Select("id, goods_name, content, price, stock, cover, status, category_id, tenants_id, create_time, update_time").
		Where("id = ? AND isdelete != ?", id, 1).
		First(&goods).Error

	if err != nil {
		return nil, fmt.Errorf("查询商品详情失败: %w", err)
	}

	// 3. 缓存结果（10分钟）
	s.cache.Set(ctx, cacheKey, goods, 10*time.Minute)

	return &goods, nil
}

// InvalidateGoodsCache 失效商品缓存
func (s *OptimizedGoodsService) InvalidateGoodsCache(ctx context.Context, goodsID int) error {
	// 删除商品详情缓存
	detailKey := fmt.Sprintf("goods:detail:%d", goodsID)
	if err := s.cache.Delete(ctx, detailKey); err != nil {
		return fmt.Errorf("删除商品详情缓存失败: %w", err)
	}

	// 删除相关的列表缓存（通过通配符模式）
	// 注意：这里简化处理，实际项目中可能需要更精细的缓存管理
	listPattern := "goods:list:*"

	// 如果Redis支持，可以删除匹配的键
	if redisClient := redis.GetClient(); redisClient != nil {
		keys, err := redisClient.Keys(ctx, listPattern).Result()
		if err == nil && len(keys) > 0 {
			redisClient.Del(ctx, keys...)
		}
	}

	return nil
}

// generateGoodsCacheKey 生成商品列表缓存键
func (s *OptimizedGoodsService) generateGoodsCacheKey(params inout.GetGoodsListReq) string {
	return fmt.Sprintf("goods:list:%d:%d:%s:%s:%d",
		params.Page,
		params.PageSize,
		params.GoodsName,
		params.Status,
		params.CategoryId)
}

// 优化的过滤器
func applyGoodsNameFilterOptimized(name string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if name != "" {
			// 使用全文索引或者前缀匹配更高效
			return db.Where("goods_name LIKE ?", name+"%")
		}
		return db
	}
}

func applyStatusFilterOptimized(status interface{}) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if status == nil {
			return db
		}

		switch v := status.(type) {
		case int:
			if v != 0 {
				return db.Where("status = ?", v)
			}
		case string:
			if v != "" {
				return db.Where("status = ?", v)
			}
		}

		return db
	}
}

func applyCategoryFilterOptimized(category int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if category != 0 {
			return db.Where("category_id = ?", category)
		}
		return db
	}
}

// formatGoodsDataOptimized 格式化商品数据（优化版本）
func formatGoodsDataOptimized(data []admin_model.Goods) []inout.GoodsItem {
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
			CreateTime: formatTimeOptimized(item.CreateTime),
			UpdateTime: formatTimeOptimized(item.UpdateTime),
		}
	}
	return formattedData
}

// formatTimeOptimized 格式化时间（优化版本）
func formatTimeOptimized(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// maxInt 返回两个整数中的较大值
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
