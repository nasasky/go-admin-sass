package admin_service

import (
	"context"
	"encoding/json"
	"fmt"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/model/app_model"
	"nasa-go-admin/redis"
	"nasa-go-admin/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type RevenueService struct{}

const (
	// 缓存key前缀
	revenueCacheKeyPrefix = "revenue:list:"
	// 缓存过期时间
	revenueCacheExpiration = 30 * time.Minute
)

// GetRevenueList 获取收入列表
func (s *RevenueService) GetRevenueList(c *gin.Context, params inout.GetRevenueListReq) (interface{}, error) {
	// 设置默认分页参数
	params.Page = max(params.Page, 1)
	params.PageSize = max(params.PageSize, 10)

	parentId, err := utils.GetParentId(c)
	if err != nil {
		return nil, fmt.Errorf("获取租户ID失败: %w", err)
	}

	// 构建缓存key
	cacheKey := fmt.Sprintf("%s%d:%d:%d:%s:%s:%s",
		revenueCacheKeyPrefix,
		parentId,
		params.Page,
		params.PageSize,
		params.Start,
		params.End,
		params.Search,
	)

	// 尝试从缓存获取数据
	if response, err := s.getFromCache(c, cacheKey); err == nil {
		return response, nil
	}

	// 如果缓存未命中，则从数据库查询
	var total int64
	data := make([]admin_model.Revenue, 0) // 初始化为空数组

	// 构建基础查询，只选择需要的字段
	query := db.Dao.Model(&admin_model.Revenue{}).
		Select("id, tenants_id, stat_date, total_orders, total_revenue, actual_revenue, paid_orders").
		Where("tenants_id = ?", parentId)

	// 添加日期范围过滤
	if params.Start != "" && params.End != "" {
		query = query.Where("stat_date BETWEEN ? AND ?", params.Start, params.End)
	}

	// 添加搜索条件（如果有）
	if params.Search != "" {
		query = query.Where("stat_date = ?", params.Search)
	}

	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize

	// 使用强制索引并执行查询
	err = query.Table("merchant_revenue_stats FORCE INDEX (idx_stat_date)").
		Order("stat_date DESC").
		Count(&total).
		Offset(offset).
		Limit(params.PageSize).
		Find(&data).Error

	if err != nil {
		return nil, fmt.Errorf("查询收益列表失败: %w", err)
	}

	// 格式化数据
	formattedData := formatRevenueData(data)

	// 构建响应
	response := map[string]interface{}{
		"total":    total,
		"items":    formattedData,
		"page":     params.Page,
		"pageSize": params.PageSize,
	}

	// 异步更新缓存
	go func() {
		if err := s.setCache(context.Background(), cacheKey, response); err != nil {
			fmt.Printf("缓存收益列表数据失败: %v\n", err)
		}
	}()

	return response, nil
}

// getFromCache 从缓存获取数据
func (s *RevenueService) getFromCache(ctx context.Context, key string) (map[string]interface{}, error) {
	rdb := redis.GetClient()
	if rdb == nil {
		return nil, fmt.Errorf("Redis client not initialized")
	}

	data, err := rdb.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.ErrNil {
			return nil, fmt.Errorf("cache miss")
		}
		return nil, err
	}

	var response map[string]interface{}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	// 确保items字段始终是数组
	if response["items"] == nil {
		response["items"] = make([]inout.RevenueRepItems, 0)
	}

	return response, nil
}

// setCache 设置缓存
func (s *RevenueService) setCache(ctx context.Context, key string, data interface{}) error {
	rdb := redis.GetClient()
	if rdb == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	// 确保data中的items字段不为null
	if dataMap, ok := data.(map[string]interface{}); ok {
		if dataMap["items"] == nil {
			dataMap["items"] = make([]inout.RevenueRepItems, 0)
		}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return rdb.Set(ctx, key, jsonData, revenueCacheExpiration).Err()
}

// GenerateRevenueStats 生成收益统计数据
func (s *RevenueService) GenerateRevenueStats(tenantsId int, days int) error {
	// 生成最近N天的统计数据
	for i := 0; i < days; i++ {
		statDate := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		err := s.generateDayRevenueStats(tenantsId, statDate)
		if err != nil {
			fmt.Printf("生成 %s 的统计数据失败: %v\n", statDate, err)
		}
	}
	return nil
}

// generateDayRevenueStats 生成某天的收益统计数据
func (s *RevenueService) generateDayRevenueStats(tenantsId int, statDate string) error {
	// 检查是否已存在该天的统计数据
	var existingCount int64
	db.Dao.Model(&admin_model.Revenue{}).
		Where("tenants_id = ? AND stat_date = ?", tenantsId, statDate).
		Count(&existingCount)

	if existingCount > 0 {
		// 如果已存在，更新数据
		return s.updateDayRevenueStats(tenantsId, statDate)
	}

	// 查询该天的订单统计数据
	var stats struct {
		TotalOrders   int     `json:"total_orders"`
		PaidOrders    int     `json:"paid_orders"`
		TotalRevenue  float64 `json:"total_revenue"`
		ActualRevenue float64 `json:"actual_revenue"`
	}

	err := db.Dao.Model(&app_model.AppOrder{}).
		Select(`
			COUNT(*) as total_orders,
			SUM(CASE WHEN status = 'paid' THEN 1 ELSE 0 END) as paid_orders,
			SUM(CASE WHEN status = 'paid' THEN amount ELSE 0 END) as total_revenue,
			SUM(CASE WHEN status = 'paid' THEN amount ELSE 0 END) as actual_revenue
		`).
		Where("tenants_id = ? AND DATE(create_time) = ?", tenantsId, statDate).
		Scan(&stats).Error

	if err != nil {
		return fmt.Errorf("查询订单统计失败: %w", err)
	}

	// 创建统计记录
	revenue := admin_model.Revenue{
		TenantsId:     tenantsId,
		StatDate:      statDate,
		PeriodStart:   statDate,
		PeriodEnd:     statDate,
		TotalOrders:   stats.TotalOrders,
		TotalRevenue:  stats.TotalRevenue,
		ActualRevenue: stats.ActualRevenue,
		PaidOrders:    stats.PaidOrders,
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
	}

	return db.Dao.Create(&revenue).Error
}

// updateDayRevenueStats 更新某天的收益统计数据
func (s *RevenueService) updateDayRevenueStats(tenantsId int, statDate string) error {
	// 查询该天的订单统计数据
	var stats struct {
		TotalOrders   int     `json:"total_orders"`
		PaidOrders    int     `json:"paid_orders"`
		TotalRevenue  float64 `json:"total_revenue"`
		ActualRevenue float64 `json:"actual_revenue"`
	}

	err := db.Dao.Model(&app_model.AppOrder{}).
		Select(`
			COUNT(*) as total_orders,
			SUM(CASE WHEN status = 'paid' THEN 1 ELSE 0 END) as paid_orders,
			SUM(CASE WHEN status = 'paid' THEN amount ELSE 0 END) as total_revenue,
			SUM(CASE WHEN status = 'paid' THEN amount ELSE 0 END) as actual_revenue
		`).
		Where("tenants_id = ? AND DATE(create_time) = ?", tenantsId, statDate).
		Scan(&stats).Error

	if err != nil {
		return fmt.Errorf("查询订单统计失败: %w", err)
	}

	// 更新统计记录
	updates := map[string]interface{}{
		"total_orders":   stats.TotalOrders,
		"total_revenue":  stats.TotalRevenue,
		"actual_revenue": stats.ActualRevenue,
		"paid_orders":    stats.PaidOrders,
		"update_time":    time.Now(),
	}

	return db.Dao.Model(&admin_model.Revenue{}).
		Where("tenants_id = ? AND stat_date = ?", tenantsId, statDate).
		Updates(updates).Error
}

// RefreshRevenueStats 刷新收益统计数据（手动调用）
func (s *RevenueService) RefreshRevenueStats(tenantsId int, days int) error {
	return s.GenerateRevenueStats(tenantsId, days)
}

// formatRevenueData 格式化收入数据
func formatRevenueData(data []admin_model.Revenue) []inout.RevenueRepItems {
	formattedData := make([]inout.RevenueRepItems, 0) // 初始化为空数组而不是nil
	for _, item := range data {
		formattedData = append(formattedData, inout.RevenueRepItems{
			Id:            item.Id,
			TenantsId:     item.TenantsId,
			StatDate:      item.StatDate,
			PeriodStart:   item.PeriodStart,
			PeriodEnd:     item.PeriodEnd,
			TotalOrders:   item.TotalOrders,
			TotalRevenue:  item.TotalRevenue,
			ActualRevenue: item.ActualRevenue,
			PaidOrders:    item.PaidOrders,
		})
	}
	return formattedData
}
