package admin_service

import (
	"context"
	"encoding/json"
	"fmt"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/utils"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/gin-gonic/gin" // 假设项目使用了Redis

	"gorm.io/gorm"
)

type OrderService struct {
	// 可选：添加缓存客户端
	redisClient *redis.Client
}

// 使用常量替代硬编码值
const (
	// 缓存过期时间
	orderCacheTTL = 5 * time.Minute
	// 查询超时时间
	queryTimeout = 5 * time.Second
)

func (s *OrderService) GetOrderList(c *gin.Context, params inout.OrderListReq) (interface{}, error) {
	// 1. 创建带超时的上下文
	ctx, cancel := context.WithTimeout(c.Request.Context(), queryTimeout)
	defer cancel()

	userType := c.GetInt("type")

	// 2. 设置默认分页参数，使用工具函数避免负值
	params = sanitizeParams(params)

	// 3. 尝试从缓存获取结果 (高并发优化)
	cacheKey := generateCacheKey(userType, params)
	if cachedResult, found := s.getFromCache(ctx, cacheKey); found {
		return cachedResult, nil
	}

	// 4. 构建查询，减少查询次数
	result, err := s.executeOrderQuery(c, userType, params) // 传递 c 而不是 ctx
	if err != nil {
		return nil, fmt.Errorf("订单查询失败: %w", err)
	}

	// 5. 缓存结果 (高并发优化)
	s.saveToCache(ctx, cacheKey, result, orderCacheTTL)

	return result, nil
}

// 净化参数，保证参数有效性
func sanitizeParams(params inout.OrderListReq) inout.OrderListReq {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 || params.PageSize > 100 {
		params.PageSize = 10
	}
	return params
}

// 修改 executeOrderQuery 函数接受 *gin.Context
func (s *OrderService) executeOrderQuery(c *gin.Context, userType int, params inout.OrderListReq) (interface{}, error) {
	// 使用 c 而不是 ctx
	parentId, err := utils.GetParentId(c)
	if err != nil {
		return nil, fmt.Errorf("获取父级ID失败: %w", err)
	}

	// 构建基本查询，只选择需要的字段 (查询优化)
	query := db.Dao.WithContext(c).Model(&admin_model.OrderList{}).
		Select("id, no, goods_id, amount, status, num, user_id, create_time, update_time, coupon_id")

	// 根据用户类型过滤
	if userType != UserTypeAdmin {
		query = query.Where("tenants_id = ?", parentId)
	}

	// 应用过滤和排序
	query = query.Scopes(
		applyOrderNoFilter(params.No),
		applyStatusFilter2(params.Status),
	).Order("create_time DESC")

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("统计总数失败: %w", err)
	}

	// 如果没有数据，直接返回空结果
	if total == 0 {
		return inout.OrderListResp{
			Items:    []inout.OrderListItem{},
			Total:    0,
			Page:     params.Page,
			PageSize: params.PageSize,
		}, nil
	}

	// 执行分页查询，使用主键索引优化 (分页优化)
	offset := (params.Page - 1) * params.PageSize

	var data []admin_model.OrderList
	if err := query.Offset(offset).Limit(params.PageSize).Find(&data).Error; err != nil {
		return nil, fmt.Errorf("查询订单列表失败: %w", err)
	}

	// 并发查询用户和商品信息 (提高性能)
	formattedData, err := s.enrichOrderData(c, data)
	if err != nil {
		return nil, err
	}

	return inout.OrderListResp{
		Items:    formattedData,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

// 并发丰富订单数据（使用优化服务）
func (s *OrderService) enrichOrderData(ctx context.Context, orders []admin_model.OrderList) ([]inout.OrderListItem, error) {
	// 提取唯一ID集合，避免重复查询 (性能优化)
	userIds := make(map[int]struct{})
	goodsIds := make(map[int]struct{})

	for _, order := range orders {
		userIds[order.UserId] = struct{}{}
		goodsIds[order.GoodsId] = struct{}{}
	}

	// 转换成切片
	uniqueUserIds := mapKeysToSlice(userIds)
	uniqueGoodsIds := mapKeysToSlice(goodsIds)

	// 使用优化服务进行并发查询
	var wg sync.WaitGroup
	wg.Add(2)

	var userMap map[int]admin_model.AppUser
	var goodsMap map[int]admin_model.Goods
	var userErr, goodsErr error

	// 并发查询用户信息
	go func() {
		defer wg.Done()
		if len(uniqueUserIds) > 0 {
			var users []admin_model.AppUser
			userErr = db.Dao.WithContext(ctx).
				Select("id, username, avatar, phone").
				Where("id IN ?", uniqueUserIds).
				Find(&users).Error

			if userErr == nil {
				userMap = make(map[int]admin_model.AppUser)
				for _, user := range users {
					userMap[user.Id] = user
				}
			}
		}
	}()

	// 并发查询商品信息（使用优化商品服务）
	go func() {
		defer wg.Done()
		if len(uniqueGoodsIds) > 0 {
			// 使用优化的商品服务批量获取
			optimizedGoodsService := NewOptimizedGoodsService()
			adminGoodsMap, err := optimizedGoodsService.BatchGetGoodsByIDs(ctx, uniqueGoodsIds)
			if err != nil {
				goodsErr = err
				return
			}

			// 直接使用返回的map
			goodsMap = adminGoodsMap
		}
	}()

	// 等待两个查询完成
	wg.Wait()

	// 检查错误
	if userErr != nil {
		return nil, fmt.Errorf("查询用户信息失败: %w", userErr)
	}
	if goodsErr != nil {
		return nil, fmt.Errorf("查询商品信息失败: %w", goodsErr)
	}

	// 组装最终数据
	result := make([]inout.OrderListItem, 0, len(orders))
	for _, order := range orders {
		item := buildOrderItem(order, userMap, goodsMap)
		result = append(result, item)
	}

	return result, nil
}

// 构建订单项
func buildOrderItem(order admin_model.OrderList, userMap map[int]admin_model.AppUser, goodsMap map[int]admin_model.Goods) inout.OrderListItem {
	amountStr := order.Amount.StringFixed(2)

	item := inout.OrderListItem{
		Id:         order.Id,
		No:         order.No,
		GoodsId:    order.GoodsId,
		Amount:     amountStr,
		Status:     order.Status,
		Num:        order.Num,
		UserId:     order.UserId,
		CreateTime: utils.FormatTime2(order.CreateTime),
		UpdateTime: utils.FormatTime2(order.UpdateTime),
		CouponId:   order.CouponId,
	}

	// 添加用户信息
	if user, exists := userMap[order.UserId]; exists {
		item.UserName = user.UserName
		item.UserAvatar = user.Avatar
		item.UserPhone = user.Phone
	}

	// 添加商品信息
	if goods, exists := goodsMap[order.GoodsId]; exists {
		item.GoodsName = goods.GoodsName
		item.GoodsPrice = goods.Price
		item.GoodsCover = goods.Cover
	}

	return item
}

// 将map的键转换为切片
func mapKeysToSlice[K comparable, V any](m map[K]V) []K {
	result := make([]K, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

// 将切片转换为map，使用提供的函数获取键
func sliceToMap[T any, K comparable](slice []T, keyFn func(T) K) map[K]T {
	result := make(map[K]T, len(slice))
	for _, item := range slice {
		result[keyFn(item)] = item
	}
	return result
}

// 过滤器优化：确保使用索引
func applyOrderNoFilter(no string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if no != "" {
			// 改为等值查询更好地利用索引
			return db.Where("no = ?", no)
			// 如果必须使用LIKE，建议索引前缀匹配:
			// return db.Where("no LIKE ?", no+"%")
		}
		return db
	}
}

// 状态过滤器
func applyStatusFilter2(status string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if status != "" {
			return db.Where("status = ?", status)
		}
		return db
	}
}

// 生成缓存键
func generateCacheKey(userType int, params inout.OrderListReq) string {
	// 创建一个唯一的缓存键，包括查询参数和用户类型
	return fmt.Sprintf("order:list:%d:%d:%d:%s:%s:%d",
		userType,
		params.Page,
		params.PageSize,
		params.Status,
		params.No,
		params.UserId)
}

// 从缓存获取结果
func (s *OrderService) getFromCache(ctx context.Context, key string) (interface{}, bool) {
	// 如果未配置 Redis 客户端，直接返回未找到
	if s.redisClient == nil {
		return nil, false
	}

	// 尝试从 Redis 获取缓存的 JSON 数据
	cachedJSON, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		// 缓存未命中或发生错误
		return nil, false
	}

	// 解析缓存的 JSON 数据
	var result inout.OrderListResp
	if err := json.Unmarshal([]byte(cachedJSON), &result); err != nil {
		// 解析错误，视为缓存未命中
		return nil, false
	}

	return result, true
}

// 保存结果到缓存
func (s *OrderService) saveToCache(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	// 如果未配置 Redis 客户端，直接返回
	if s.redisClient == nil {
		return
	}

	// 将结果序列化为 JSON
	jsonData, err := json.Marshal(value)
	if err != nil {
		// 序列化失败，忽略错误
		return
	}

	// 异步保存到缓存，不阻塞主流程
	go func() {
		// 创建新的超时上下文
		saveCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// 保存到 Redis
		s.redisClient.Set(saveCtx, key, jsonData, ttl)
	}()
}
