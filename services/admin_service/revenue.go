package admin_service

import (
	"fmt"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/utils"

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
