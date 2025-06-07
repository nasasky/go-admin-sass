package admin_service

import (
	"context"
	"fmt"
	"math"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/redis" // 导入自定义的 redis 包
	"time"

	"github.com/gin-gonic/gin"
	edisv9 "github.com/redis/go-redis/v9" // 适用于 v9 版本
)

type SettingService struct {
}

// GetSetting
func (s *SettingService) GetSetting(c *gin.Context, params inout.GetArticleListReq) (interface{}, error) {

	var data []admin_model.SettingList
	var total int64

	// 设置默认分页参数
	params.Page = max(params.Page, 1)
	params.PageSize = max(params.PageSize, 10)

	// 构建查询
	query := db.Dao.Model(&admin_model.SettingList{})

	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize

	// 执行查询
	err := query.Count(&total).Offset(offset).Limit(params.PageSize).Find(&data).Error
	if err != nil {
		return nil, err
	}

	// 格式化数据
	formattedData := formatSettingData(data)

	response := admin_model.Setting{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Items:    formattedData,
	}
	return response, nil

}

// GetSettingDetail
func (s *SettingService) GetSettingDetail(c *gin.Context, id int) (interface{}, error) {
	var data admin_model.SettingList
	err := db.Dao.Model(&admin_model.SettingList{}).Where("id = ?", id).First(&data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

//UpdateSetting

func (s *SettingService) UpdateSetting(c *gin.Context, params inout.UpdateSettingReq) (interface{}, error) {
	var data admin_model.SettingList
	err := db.Dao.Model(&admin_model.SettingList{}).Where("id = ?", params.Id).First(&data).Error
	if err != nil {
		return nil, err
	}

	data.Name = params.Name
	data.Appid = params.Appid
	data.Secret = params.Secret
	data.UpdateTime = time.Now()
	data.Tips = params.Tips
	data.Type = params.Type

	err = db.Dao.Save(&data).Error
	if err != nil {
		return nil, err
	}

	return data.Id, nil
}

// AddSetting
func (s *SettingService) AddSetting(c *gin.Context, params inout.SettingReq) (interface{}, error) {
	var uid = c.GetInt("uid")
	fmt.Println(params)
	data := admin_model.SettingList{
		Name:       params.Name,
		Appid:      params.Appid,
		Secret:     params.Secret,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		UserId:     uid,
		Tips:       params.Tips,
		Type:       params.Type,
	}
	err := db.Dao.Create(&data).Error
	if err != nil {
		return nil, err
	}
	return data.Id, nil
}

func formatSettingData(data []admin_model.SettingList) []admin_model.SettingList {
	var resp []admin_model.SettingList
	for _, item := range data {

		resp = append(resp, admin_model.SettingList{
			Id:         item.Id,
			Name:       item.Name,
			Appid:      item.Appid,
			Secret:     item.Secret,
			CreateTime: item.CreateTime,
			UpdateTime: item.UpdateTime,
			UserId:     item.UserId,
			Tips:       item.Tips,
			Type:       item.Type,
		})
	}
	return resp
}

// 删除系统参数配置
func (s *SettingService) DeleteSetting(c *gin.Context, id int) error {
	err := db.Dao.Where("id = ?", id).Delete(&admin_model.SettingList{}).Error
	if err != nil {
		return err
	}
	return nil

}

// GetQueueLogList 获取待取消订单队列的状态
func (s *SettingService) GetQueueLogList(page, pageSize int) (map[string]interface{}, error) {
	// 使用全局Redis客户端
	redisClient := redis.GetClient()
	if redisClient == nil {
		return nil, fmt.Errorf("Redis客户端未初始化")
	}

	ctx := context.Background()

	// 设置默认分页参数
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 获取队列中所有元素
	result, err := redisClient.ZRangeWithScores(ctx, "pending_order_cancellations", 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("获取队列状态失败: %w", err)
	}

	// 统计数据
	now := time.Now()
	total := len(result)
	overdue := 0
	pending := 0

	// 计算分页
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize

	if endIndex > total {
		endIndex = total
	}

	// 获取对应页的数据
	var pagedResult []edisv9.Z
	if startIndex < total {
		pagedResult = result[startIndex:endIndex]
	} else {
		pagedResult = []edisv9.Z{}
	}

	// 先计算总的过期和待处理数量
	for _, z := range result {
		expireTimestamp := int64(z.Score)
		expireTime := time.Unix(expireTimestamp, 0)

		if now.After(expireTime) {
			overdue++
		} else {
			pending++
		}
	}

	// 完整统计数据
	stats := map[string]interface{}{
		"total":          total,
		"overdue_orders": overdue,
		"pending_orders": pending,
		"current_time":   now.Format("2006-01-02 15:04:05"),
	}

	// 处理当前页的数据
	items := make([]map[string]interface{}, 0, len(pagedResult))

	for _, z := range pagedResult {
		orderNo := z.Member.(string)
		expireTimestamp := int64(z.Score)
		expireTime := time.Unix(expireTimestamp, 0)

		// 是否已过期但尚未处理
		isOverdue := now.After(expireTime)

		orderInfo := map[string]interface{}{
			"order_no":     orderNo,
			"expire_time":  expireTime.Format("2006-01-02 15:04:05"),
			"seconds_left": int64(time.Until(expireTime).Seconds()),
			"overdue":      isOverdue,
		}

		items = append(items, orderInfo)
	}

	return map[string]interface{}{
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
		"stats":       stats, // 将统计信息放入stats字段
		"items":       items, // 将订单列表改为items字段
	}, nil
}
