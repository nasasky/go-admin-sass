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
	// 导入Redis v9
	// 适用于 v9 版本
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

	// 获取正确的超时队列中所有元素 - 修复：使用正确的队列名称
	result, err := redisClient.ZRangeWithScores(ctx, "order_timeouts", 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("获取超时队列状态失败: %w", err)
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

	// 处理当前页的数据
	items := make([]map[string]interface{}, 0)

	for i, z := range result {
		// 分页过滤
		if i < startIndex || i >= endIndex {
			continue
		}

		orderNo := z.Member.(string)
		expireTimestamp := int64(z.Score)
		expireTime := time.Unix(expireTimestamp, 0)

		// 是否已过期但尚未处理
		isOverdue := now.After(expireTime)
		secondsLeft := int64(time.Until(expireTime).Seconds())

		if isOverdue {
			overdue++
		} else {
			pending++
		}

		// 从数据库获取订单详细信息（如果存在）
		var orderInfo map[string]interface{}
		var orderData struct {
			UserID    int       `json:"user_id"`
			GoodsID   int       `json:"goods_id"`
			Amount    float64   `json:"amount"`
			Status    string    `json:"status"`
			CreatedAt time.Time `json:"created_at"`
		}

		err := db.Dao.Table("app_order").
			Select("user_id, goods_id, amount, status, create_time as created_at").
			Where("no = ?", orderNo).
			First(&orderData).Error

		if err == nil {
			// 找到了订单信息
			orderInfo = map[string]interface{}{
				"order_no":     orderNo,
				"user_id":      orderData.UserID,
				"goods_id":     orderData.GoodsID,
				"amount":       orderData.Amount,
				"status":       orderData.Status,
				"created_at":   orderData.CreatedAt.Format("2006-01-02 15:04:05"),
				"expire_time":  expireTime.Format("2006-01-02 15:04:05"),
				"seconds_left": secondsLeft,
				"minutes_left": int(secondsLeft / 60),
				"overdue":      isOverdue,
				"has_order":    true,
			}
		} else {
			// 队列中有记录但订单不存在（可能是孤立记录）
			orderInfo = map[string]interface{}{
				"order_no":     orderNo,
				"expire_time":  expireTime.Format("2006-01-02 15:04:05"),
				"seconds_left": secondsLeft,
				"minutes_left": int(secondsLeft / 60),
				"overdue":      isOverdue,
				"has_order":    false,
				"note":         "队列记录存在但订单不存在",
			}
		}

		items = append(items, orderInfo)
	}

	// 完整统计数据
	stats := map[string]interface{}{
		"total":          total,
		"overdue_orders": overdue,
		"pending_orders": pending,
		"current_time":   now.Format("2006-01-02 15:04:05"),
		"queue_name":     "order_timeouts", // 添加队列名称信息
	}

	return map[string]interface{}{
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
		"stats":       stats,
		"items":       items,
		"message":     fmt.Sprintf("共找到 %d 个队列记录，其中 %d 个已过期，%d 个等待处理", total, overdue, pending),
	}, nil
}
