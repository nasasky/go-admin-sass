package admin_service

import (
	"fmt"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/model/app_model"
	"nasa-go-admin/utils"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RevenueService struct{}

// GetRevenueList 获取收入列表
func (s *RevenueService) GetRevenueList(c *gin.Context, params inout.GetRevenueListReq) (interface{}, error) {
	var data []admin_model.Revenue
	var total int64

	// 设置默认分页参数
	params.Page = max(params.Page, 1)
	params.PageSize = max(params.PageSize, 10)

	parentId, err := utils.GetParentId(c)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return nil, nil
	}
	fmt.Println(parentId)

	// 先尝试生成统计数据（如果表为空的话）
	s.ensureRevenueDataExists(parentId)

	// 构建查询
	query := db.Dao.Model(&admin_model.Revenue{}).Scopes(
		applyTenantsIdFilter(parentId),
		applyStatDateFilter(params.Start, params.End),
		applySearchFilter(params.Search),
	).Order("stat_date DESC")
	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize
	// 执行查询
	err = query.Count(&total).Offset(offset).Limit(params.PageSize).Find(&data).Error
	if err != nil {
		return nil, err
	}
	// 格式化数据
	formattedData := []inout.RevenueRepItems{}
	if len(data) > 0 {
		formattedData = formatRevenueData(data)
	}

	// 构建响应
	response := map[string]interface{}{
		"total":    total,
		"items":    formattedData,
		"page":     params.Page,
		"pageSize": params.PageSize,
	}
	return response, nil
}

// ensureRevenueDataExists 确保收益统计数据存在，如果不存在则生成
func (s *RevenueService) ensureRevenueDataExists(tenantsId int) {
	var count int64
	db.Dao.Model(&admin_model.Revenue{}).Where("tenants_id = ?", tenantsId).Count(&count)

	// 如果统计表为空，则生成最近30天的数据
	if count == 0 {
		s.GenerateRevenueStats(tenantsId, 30)
	}
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
		TotalOrder:    stats.TotalOrders,
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
		"total_order":    stats.TotalOrders,
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

// applyTenantsIdFilter 应用租户ID过滤器
func applyTenantsIdFilter(tenantsId int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("tenants_id = ?", tenantsId)
	}

}

// applyStatDateFilter 应用统计日期过滤器
func applyStatDateFilter(start, end string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if start != "" && end != "" {
			return db.Where("stat_date BETWEEN ? AND ?", start, end)
		}
		return db
	}

}

// applySearchFilter 应用搜索过滤器
func applySearchFilter(search string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if search != "" {
			return db.Where("stat_date LIKE ?", "%"+search+"%")
		}
		return db
	}
}

// formatRevenueData 格式化收入数据
func formatRevenueData(data []admin_model.Revenue) []inout.RevenueRepItems {
	var formattedData []inout.RevenueRepItems
	for _, item := range data {
		formattedData = append(formattedData, inout.RevenueRepItems{
			Id:            item.Id,
			TenantsId:     item.TenantsId,
			StatDate:      item.StatDate,
			PeriodStart:   item.PeriodStart,
			PeriodEnd:     item.PeriodEnd,
			TotalOrder:    item.TotalOrder,
			TotalRevenue:  item.TotalRevenue,
			ActualRevenue: item.ActualRevenue,
			PaidOrders:    item.PaidOrders,
		})
	}
	return formattedData
}
