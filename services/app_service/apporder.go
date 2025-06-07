package app_service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/app_model"
	"nasa-go-admin/services/miniapp_service"
	"nasa-go-admin/services/public_service"
	"nasa-go-admin/utils"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	lockIdentifiers = make(map[string]string)
	lockMutex       = &sync.Mutex{}
	workerStarted   = false
	workerMutex     = &sync.Mutex{}
)

// 订单状态常量
const (
	OrderStatusPaid     = 1
	OrderStatusShipped  = 2
	OrderStatusComplete = 3
)

// 订单相关的模板ID
const (
	OrderPaidTemplateID     = "FL4Qq5zBk5zpXs1Jkd7F8D_STgGm9PcdSqOkZnegm2g" // 替换为实际模板ID
	OrderShippedTemplateID  = "FL4Qq5zBk5zpXs1Jkd7F8D_STgGm9PcdSqOkZnegm2g" // 发货通知模板ID
	OrderCompleteTemplateID = "订单完成模板ID"                                    // 替换为实际模板ID
)

type OrderService struct {
	redisClient *redis.Client
}

// GetOrderDetail 获取订单详情
func (s *OrderService) GetOrderDetail(c *gin.Context, uid int, id int) (interface{}, error) {
	// 查询订单详情
	var order app_model.AppOrder
	err := db.Dao.Where("id = ? AND user_id = ?", id, uid).First(&order).Error
	if err != nil {
		return nil, err
	}

	// 查询商品详情
	var goods app_model.AppGoods
	err = db.Dao.Where("id = ?", order.GoodsId).First(&goods).Error
	if err != nil {
		return nil, err
	}

	// 格式化时间字段
	response := inout.OrderItem{
		Id:         order.Id,
		UserId:     order.UserId,
		GoodsId:    order.GoodsId,
		GoodsName:  goods.GoodsName,
		GoodsPrice: goods.Price,
		Num:        order.Num,
		Amount:     order.Amount,
		Status:     order.Status,
		CreateTime: order.CreateTime.Format("2006-01-02 15:04:05"),
		UpdateTime: order.UpdateTime.Format("2006-01-02 15:04:05"),
	}

	return response, nil
}

// GetMyOrderList 获取我的订单列表
func (s *OrderService) GetMyOrderList(c *gin.Context, uid int, params inout.MyOrderReq) (interface{}, error) {
	// 查询订单列表
	var orders []app_model.AppOrder
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
	err := db.Dao.Model(&app_model.AppOrder{}).Where("user_id = ?", uid).Count(&total).Offset(offset).Limit(params.PageSize).Find(&orders).Error
	if err != nil {
		return nil, err
	}

	// 获取所有商品 ID
	goodsIds := make([]int, len(orders))
	for i, order := range orders {
		goodsIds[i] = order.GoodsId
	}

	// 批量查询商品详情
	goodsMap, err := s.getGoodsDetailsBatch(goodsIds)
	if err != nil {
		return nil, err
	}

	// 格式化时间字段并查询商品详情
	formattedData := make([]inout.OrderItem, len(orders))
	for i, item := range orders {
		goods := goodsMap[item.GoodsId]

		formattedData[i] = inout.OrderItem{
			Id:         item.Id,
			UserId:     item.UserId,
			GoodsId:    item.GoodsId,
			GoodsName:  goods.GoodsName,
			GoodsPrice: goods.Price,
			Num:        item.Num,
			Amount:     item.Amount,
			Status:     item.Status,
			CreateTime: item.CreateTime.Format("2006-01-02 15:04:05"),
			UpdateTime: item.UpdateTime.Format("2006-01-02 15:04:05"),
		}
	}
	response := inout.MyOrderResp{
		Total:    total,
		List:     formattedData,
		Page:     params.Page,
		PageSize: params.PageSize,
	}

	return response, nil
}

// getGoodsDetailsBatch 批量查询商品详情（优化版本）
func (s *OrderService) getGoodsDetailsBatch(goodsIds []int) (map[int]app_model.AppGoods, error) {
	if len(goodsIds) == 0 {
		return make(map[int]app_model.AppGoods), nil
	}

	// 直接使用数据库批量查询，但优化查询方式
	var goodsList []app_model.AppGoods
	err := db.Dao.Select("id, goods_name, price, content, cover, status, category_id, stock, create_time, update_time").
		Where("id IN ? AND isdelete != ?", goodsIds, 1).
		Find(&goodsList).Error

	if err != nil {
		return nil, fmt.Errorf("批量查询商品失败: %w", err)
	}

	// 将商品详情存储到映射中
	goodsMap := make(map[int]app_model.AppGoods)
	for _, goods := range goodsList {
		goodsMap[goods.Id] = goods
	}

	return goodsMap, nil
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(c *gin.Context, uid int, params inout.CreateOrderReq) (string, error) {
	// 确保订单取消工作器已启动
	s.ensureOrderCancellationWorkerStarted()

	// 1. 使用分布式锁或Redis锁防止用户短时间内重复提交
	lockKey := fmt.Sprintf("create_order:user:%d:goods:%d", uid, params.GoodsId)
	if acquired, err := s.acquireLock(lockKey, 5*time.Second); err != nil || !acquired {
		return "", fmt.Errorf("操作过于频繁，请稍后再试")
	}
	defer s.releaseLock(lockKey)

	// 2. 开始事务
	tx := db.Dao.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 3. 检查并锁定商品记录
	var goods app_model.AppGoods
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", params.GoodsId).First(&goods).Error; err != nil {
		tx.Rollback()
		return "", fmt.Errorf("商品不存在或已下架: %w", err)
	}

	// 4. 检查库存是否充足
	if goods.Stock < params.Num {
		tx.Rollback()
		return "", fmt.Errorf("商品库存不足，当前库存: %d", goods.Stock)
	}

	// 5. 计算订单总价
	totalPrice := goods.Price * float64(params.Num)

	// 6. 检查用户并锁定用户记录
	var user app_model.UserApp
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", uid).First(&user).Error; err != nil {
		tx.Rollback()
		return "", fmt.Errorf("用户不存在: %w", err)
	}

	// 7. 查询并锁定用户钱包记录
	var wallet app_model.AppWallet
	err := tx.Set("gorm:query_option", "FOR UPDATE").Where("user_id = ?", uid).First(&wallet).Error

	// 8. 处理钱包不存在的情况
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新钱包
			wallet = app_model.AppWallet{
				UserId: uid,
				Money:  0.00,
			}
			if err := tx.Create(&wallet).Error; err != nil {
				tx.Rollback()
				return "", fmt.Errorf("创建钱包失败: %w", err)
			}
		} else {
			tx.Rollback()
			return "", fmt.Errorf("查询钱包失败: %w", err)
		}
	}

	// 9. 确定订单状态并处理支付逻辑
	orderStatus := "pending" // 默认为待支付

	// 如果余额足够，则直接扣款并设为已支付
	if wallet.Money >= totalPrice {
		err = s.deductWalletAndRecordTransaction(tx, uid, totalPrice, wallet.Money)
		if err != nil {
			tx.Rollback()
			return "", fmt.Errorf("扣款失败: %w", err)
		}
		orderStatus = "paid"
	}

	// 10. 创建订单记录
	orderNo := generateOrderNo(uid, params.GoodsId)
	order := app_model.AppOrder{
		UserId:     uid,
		GoodsId:    params.GoodsId,
		Num:        params.Num,
		Amount:     totalPrice,
		TenantsId:  goods.TenantsId,
		Status:     orderStatus,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		No:         orderNo,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return "", fmt.Errorf("创建订单失败: %w", err)
	}
	wsService := public_service.GetWebSocketService()
	if orderStatus == "paid" {
		err := wsService.SendOrderNotification(uid, orderNo, "paid", goods.GoodsName)
		if err != nil {
			return "", err
		}
	} else {
		err := wsService.SendOrderNotification(uid, orderNo, "pending", goods.GoodsName)
		if err != nil {
			return "", err
		}
	}
	// 2. 推送订单创建消息
	if order.Id != 0 {
		go func() {
			// 异步发送消息，避免影响主流程
			err := miniapp_service.SendSubscribeMsg(user.Openid, OrderPaidTemplateID, strconv.Itoa(order.Id))
			if err != nil {
				log.Printf("发送订单创建消息失败: %v, 订单ID: %d", err, order.Id)
			}
		}()
	}

	// 11. 更新商品库存 - 使用原子操作
	if err := tx.Model(&app_model.AppGoods{}).
		Where("id = ? AND stock >= ?", params.GoodsId, params.Num).
		Update("stock", gorm.Expr("stock - ?", params.Num)).Error; err != nil {
		tx.Rollback()
		return "", fmt.Errorf("更新库存失败: %w", err)
	}

	// 【新增】12. 更新商家收入统计数据
	if err := s.updateMerchantRevenueStats(tx, goods.TenantsId, &order, goods.GoodsName, "create"); err != nil {
		tx.Rollback()
		log.Printf("更新商家收入统计失败: %v", err)
		return "", fmt.Errorf("更新商家收入统计失败: %w", err)
	}

	// 13. 提交事务
	if err := tx.Commit().Error; err != nil {
		return "", fmt.Errorf("提交事务失败: %w", err)
	}

	// 14. 如果是待支付状态，将订单添加到延迟取消队列
	if orderStatus == "pending" {
		// 设置合理的超时时间，如15分钟
		expireTime := time.Now().Add(15 * time.Minute)
		if err := s.scheduleOrderCancellation(orderNo, expireTime); err != nil {
			log.Printf("警告: 无法将订单 %s 添加到取消队列: %v", orderNo, err)
			// 如果Redis不可用，使用备用方案
			go func() {
				time.Sleep(15 * time.Minute)
				log.Printf("使用备用计时器检查订单: %s", orderNo)
				s.checkAndCancelOrder(orderNo)
			}()
		}
	}

	// 15. 返回订单号和状态信息
	return orderNo, nil
}

// updateMerchantRevenueStats 更新商家收入统计数据
func (s *OrderService) updateMerchantRevenueStats(tx *gorm.DB, merchantId int, order *app_model.AppOrder, goodsName string, operation string, oldStatus ...string) error {
	// 获取当前日期
	orderDateStr := order.CreateTime.Format("2006-01-02")

	// 基于操作类型执行不同的更新逻辑
	switch operation {
	case "create":
		// 更新日统计数据
		var stats app_model.MerchantRevenueStats
		// 修改此处：使用字符串而不是日期对象
		result := tx.Where("tenants_id = ? AND stat_date = ? AND stat_period = ?",
			merchantId, orderDateStr, "day").First(&stats)

		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return result.Error
		}

		// 如果记录不存在，创建新的统计记录
		isNewRecord := errors.Is(result.Error, gorm.ErrRecordNotFound)
		if isNewRecord {
			stats = app_model.MerchantRevenueStats{
				TenantsId:       merchantId,
				StatDate:        orderDateStr,
				StatPeriod:      "day",
				PeriodStart:     orderDateStr,
				PeriodEnd:       orderDateStr,
				TotalOrders:     0,
				TotalRevenue:    0,
				ActualRevenue:   0,
				RefundAmount:    0,
				PaidOrders:      0,
				PendingOrders:   0,
				CancelledOrders: 0,
				RefundedOrders:  0,
				ItemsSold:       0,
				CreateTime:      time.Now(),
				UpdateTime:      time.Now(),
			}
		}

		// 更新统计数据
		stats.TotalOrders += 1
		stats.TotalRevenue += order.Amount
		stats.ItemsSold += order.Num

		// 根据订单状态更新对应字段
		switch order.Status {
		case "paid":
			stats.ActualRevenue += order.Amount
			stats.PaidOrders += 1
		case "pending":
			stats.PendingOrders += 1
		case "cancelled":
			stats.CancelledOrders += 1
		case "refunded":
			stats.RefundAmount += order.Amount
			stats.RefundedOrders += 1
		}

		stats.UpdateTime = time.Now()

		// 保存或更新统计记录
		if isNewRecord {
			if err := tx.Create(&stats).Error; err != nil {
				return fmt.Errorf("创建商家收入统计记录失败: %w", err)
			}
		} else {
			if err := tx.Save(&stats).Error; err != nil {
				return fmt.Errorf("更新商家收入统计记录失败: %w", err)
			}
		}

		// 更新商品明细统计数据
		var detail app_model.MerchantRevenueDetails
		detailResult := tx.Where("tenants_id = ? AND stat_date = ? AND goods_id = ?",
			merchantId, orderDateStr, order.GoodsId).First(&detail)

		isNewDetail := errors.Is(detailResult.Error, gorm.ErrRecordNotFound)
		if isNewDetail {
			detail = app_model.MerchantRevenueDetails{
				TenantsId:    merchantId,
				StatDate:     orderDateStr,
				GoodsId:      order.GoodsId,
				GoodsName:    goodsName,
				OrderCount:   0,
				SoldCount:    0,
				Revenue:      0,
				RefundCount:  0,
				RefundAmount: 0,
				CreateTime:   time.Now(),
				UpdateTime:   time.Now(),
			}
		}

		// 更新明细数据
		detail.OrderCount += 1
		detail.SoldCount += order.Num

		if order.Status == "paid" {
			detail.Revenue += order.Amount
		} else if order.Status == "refunded" {
			detail.RefundCount += 1
			detail.RefundAmount += order.Amount
		}

		detail.UpdateTime = time.Now()

		// 保存或更新明细记录
		if isNewDetail {
			if err := tx.Create(&detail).Error; err != nil {
				return fmt.Errorf("创建商家收入明细记录失败: %w", err)
			}
		} else {
			if err := tx.Save(&detail).Error; err != nil {
				return fmt.Errorf("更新商家收入明细记录失败: %w", err)
			}
		}

	case "update":
		// 处理订单更新的情况 - 例如从pending到paid的变更
		var stats app_model.MerchantRevenueStats
		// 修改此处：使用字符串而不是日期对象
		if err := tx.Where("tenants_id = ? AND stat_date = ? AND stat_period = ?",
			merchantId, orderDateStr, "day").First(&stats).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 如果不存在记录，则调用create操作创建新记录
				return s.updateMerchantRevenueStats(tx, merchantId, order, goodsName, "create")
			}
			return err
		}

		// 获取订单的原始状态（从参数中获取）
		var prevStatus string
		if len(oldStatus) > 0 {
			prevStatus = oldStatus[0]
		} else {
			// 如果没有提供旧状态，则默认为与当前状态相同（无变化）
			return nil
		}

		newStatus := order.Status

		// 如果状态没有变化，不需要更新统计
		if prevStatus == newStatus {
			return nil
		}

		// 减去旧状态的统计数据
		switch prevStatus {
		case "pending":
			stats.PendingOrders -= 1
		case "paid":
			stats.PaidOrders -= 1
			stats.ActualRevenue -= order.Amount
		case "cancelled":
			stats.CancelledOrders -= 1
		case "refunded":
			stats.RefundedOrders -= 1
			stats.RefundAmount -= order.Amount
		}

		// 增加新状态的统计数据
		switch newStatus {
		case "pending":
			stats.PendingOrders += 1
		case "paid":
			stats.PaidOrders += 1
			stats.ActualRevenue += order.Amount
		case "cancelled":
			stats.CancelledOrders += 1
		case "refunded":
			stats.RefundAmount += order.Amount
			stats.RefundedOrders += 1
		}

		stats.UpdateTime = time.Now()

		// 保存更新后的统计记录
		if err := tx.Save(&stats).Error; err != nil {
			return fmt.Errorf("更新商家收入统计记录失败: %w", err)
		}

		// 同样更新商品明细统计
		var detail app_model.MerchantRevenueDetails
		// 修改此处：使用字符串而不是日期对象
		if err := tx.Where("tenants_id = ? AND stat_date = ? AND goods_id = ?",
			merchantId, orderDateStr, order.GoodsId).First(&detail).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 如果不存在记录，则调用create操作创建新记录
				return s.updateMerchantRevenueStats(tx, merchantId, order, goodsName, "create")
			}
			return err
		}

		// 更新商品明细统计
		if prevStatus == "paid" && newStatus == "refunded" {
			detail.Revenue -= order.Amount
			detail.RefundCount += 1
			detail.RefundAmount += order.Amount
		} else if prevStatus == "pending" && newStatus == "paid" {
			detail.Revenue += order.Amount
		}

		detail.UpdateTime = time.Now()

		if err := tx.Save(&detail).Error; err != nil {
			return fmt.Errorf("更新商家收入明细记录失败: %w", err)
		}

	case "cancel":
		// 处理订单取消的情况 - 从pending状态变更为cancelled
		var stats app_model.MerchantRevenueStats
		// 修改此处：使用字符串而不是日期对象
		if err := tx.Where("tenants_id = ? AND stat_date = ? AND stat_period = ?",
			merchantId, orderDateStr, "day").First(&stats).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			// 如果记录不存在，无需更新
			return nil
		}

		// 更新统计数据 - 减少pending订单，增加cancelled订单
		stats.PendingOrders -= 1
		stats.CancelledOrders += 1
		stats.UpdateTime = time.Now()

		// 保存更新后的统计记录
		if err := tx.Save(&stats).Error; err != nil {
			return fmt.Errorf("更新商家收入统计记录失败: %w", err)
		}

		// 不需要更新商品明细中的收入数据，因为pending订单没有产生实际收入
		var detail app_model.MerchantRevenueDetails
		// 修改此处：使用字符串而不是日期对象
		if err := tx.Where("tenants_id = ? AND stat_date = ? AND goods_id = ?",
			merchantId, orderDateStr, order.GoodsId).First(&detail).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			// 如果记录不存在，无需更新
			return nil
		}

		// 记录订单取消
		detail.UpdateTime = time.Now()

		if err := tx.Save(&detail).Error; err != nil {
			return fmt.Errorf("更新商家收入明细记录失败: %w", err)
		}

	case "refund":
		// 处理订单退款的情况 - 从paid状态变更为refunded
		var stats app_model.MerchantRevenueStats
		// 修改此处：使用字符串而不是日期对象
		if err := tx.Where("tenants_id = ? AND stat_date = ? AND stat_period = ?",
			merchantId, orderDateStr, "day").First(&stats).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			// 如果记录不存在，无需更新
			return nil
		}

		// 更新统计数据 - 减少已支付订单，增加已退款订单，调整收入
		stats.PaidOrders -= 1
		stats.RefundedOrders += 1
		stats.ActualRevenue -= order.Amount // 减少实际收入
		stats.RefundAmount += order.Amount  // 增加退款金额
		stats.UpdateTime = time.Now()

		// 保存更新后的统计记录
		if err := tx.Save(&stats).Error; err != nil {
			return fmt.Errorf("更新商家收入统计记录失败: %w", err)
		}

		// 更新商品明细统计
		var detail app_model.MerchantRevenueDetails
		// 修改此处：使用字符串而不是日期对象
		if err := tx.Where("tenants_id = ? AND stat_date = ? AND goods_id = ?",
			merchantId, orderDateStr, order.GoodsId).First(&detail).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			// 如果记录不存在，无需更新
			return nil
		}

		// 更新商品明细的退款数据
		detail.Revenue -= order.Amount      // 减少收入
		detail.RefundCount += 1             // 增加退款计数
		detail.RefundAmount += order.Amount // 增加退款金额
		detail.UpdateTime = time.Now()

		if err := tx.Save(&detail).Error; err != nil {
			return fmt.Errorf("更新商家收入明细记录失败: %w", err)
		}
	}

	return nil
}

// 确保订单取消工作器已启动
func (s *OrderService) ensureOrderCancellationWorkerStarted() {
	workerMutex.Lock()
	defer workerMutex.Unlock()

	if !workerStarted {
		log.Printf("启动订单取消工作器")
		s.startOrderCancellationWorker()
		workerStarted = true
	}
}

// 将订单加入取消队列
func (s *OrderService) scheduleOrderCancellation(orderNo string, expireTime time.Time) error {
	if s.redisClient == nil {
		return fmt.Errorf("redis客户端未初始化")
	}

	// 使用Redis的有序集合作为延迟队列
	// 分数使用过期时间的Unix时间戳
	score := float64(expireTime.Unix())
	ctx := context.Background()

	err := s.redisClient.ZAdd(ctx, "pending_order_cancellations", redis.Z{
		Score:  score,
		Member: orderNo,
	}).Err()

	if err != nil {
		return fmt.Errorf("添加订单到取消队列失败: %w", err)
	}

	log.Printf("订单 %s 已加入取消队列，将在 %v 后过期", orderNo, time.Until(expireTime))
	return nil
}

// 启动订单取消工作器
func (s *OrderService) startOrderCancellationWorker() {
	go func() {
		// 确保此函数中的panic不会导致整个程序崩溃
		defer func() {
			if r := recover(); r != nil {
				log.Printf("订单取消工作器发生panic: %v", r)
				// 重新启动工作器
				time.Sleep(5 * time.Second)
				s.startOrderCancellationWorker()
			}
		}()

		log.Printf("订单取消工作器已启动")

		for {
			if s.redisClient == nil {
				log.Printf("警告: Redis客户端未初始化，订单取消工作器暂停工作")
				time.Sleep(30 * time.Second)
				continue
			}

			now := time.Now().Unix()
			ctx := context.Background()

			// 查找所有已过期的订单
			orders, err := s.redisClient.ZRangeByScore(ctx, "pending_order_cancellations", &redis.ZRangeBy{
				Min:    "0",
				Max:    fmt.Sprintf("%d", now),
				Offset: 0,
				Count:  10, // 每次处理10个，避免处理过多
			}).Result()

			if err != nil {
				log.Printf("获取待取消订单失败: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			for _, orderNo := range orders {
				// 从队列中移除此订单
				s.redisClient.ZRem(ctx, "pending_order_cancellations", orderNo)

				// 处理订单取消
				log.Printf("开始处理过期订单: %s", orderNo)
				go s.checkAndCancelOrder(orderNo) // 使用goroutine避免阻塞主循环
			}

			// 即使没有订单需要处理，也不要频繁检查
			time.Sleep(5 * time.Second)

			// 每小时执行一次数据库备份检查，确保没有漏掉的订单
			if time.Now().Minute() == 0 {
				s.checkExpiredOrdersInDatabase()
			}
		}
	}()
}

// 从数据库检查过期订单
func (s *OrderService) checkExpiredOrdersInDatabase() {
	log.Printf("执行数据库过期订单检查")
	var expiredOrders []app_model.AppOrder
	expireTime := time.Now().Add(-15 * time.Minute)

	// 查找所有已过期但仍处于pending状态的订单
	err := db.Dao.Where("status = ? AND create_time < ?", "pending", expireTime).
		Find(&expiredOrders).Error

	if err != nil {
		log.Printf("查询过期订单失败: %v", err)
		return
	}

	if len(expiredOrders) > 0 {
		log.Printf("发现 %d 个数据库中未处理的过期订单", len(expiredOrders))

		for _, order := range expiredOrders {
			log.Printf("处理数据库中的过期订单: %s", order.No)
			go s.checkAndCancelOrder(order.No)
		}
	}
}

// generateOrderNo 生成唯一的订单号
func generateOrderNo(uid, goodsId int) string {
	// 结合用户ID、商品ID、时间戳和随机数生成订单号
	timestamp := time.Now().Format("20060102150405")
	random := rand.Intn(1000)
	return fmt.Sprintf("%s%d%d%03d", timestamp, uid, goodsId, random)
}

// acquireLock 获取分布式锁
func (s *OrderService) acquireLock(key string, expiration time.Duration) (bool, error) {
	if s.redisClient == nil {
		log.Printf("ERROR: Redis client is nil when acquiring lock: %s", key)
		// 添加降级策略，在开发环境中可以允许继续
		return true, nil
	}

	// 生成唯一标识符（用于安全释放锁）
	uuid := fmt.Sprintf("%d-%s", time.Now().UnixNano(), utils.RandomString(16))

	// 使用SET NX EX命令尝试获取锁
	result, err := s.redisClient.SetNX(context.Background(), key, uuid, expiration).Result()

	if err != nil {
		log.Printf("获取锁失败 %s: %v", key, err)
		return false, fmt.Errorf("获取锁失败: %w", err)
	}

	// 将锁的唯一标识符存储在Map中，用于后续安全释放
	if result {
		s.saveLockIdentifier(key, uuid)
	}
	log.Printf("Lock acquisition result for %s: %v (error: %v)", key, result, err)
	return result, nil
}

// releaseLock 释放分布式锁
func (s *OrderService) releaseLock(key string) {
	if s.redisClient == nil {
		return
	}

	// 获取当前锁的唯一标识符
	uuid := s.getLockIdentifier(key)
	if uuid == "" {
		return // 没有找到标识符，可能锁已超时
	}

	// 使用Lua脚本确保只删除自己的锁
	luaScript := `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
	`

	ctx := context.Background()
	result, err := s.redisClient.Eval(ctx, luaScript, []string{key}, uuid).Result()
	if err != nil {
		log.Printf("释放锁 %s 时发生错误: %v", key, err)
	}

	if result.(int64) > 0 {
		log.Printf("成功释放锁: %s", key)
	}

	// 清理锁标识符
	s.deleteLockIdentifier(key)
}

// deductWalletAndRecordTransaction 扣除用户余额并记录交易
func (s *OrderService) deductWalletAndRecordTransaction(tx *gorm.DB, uid int, amount float64, allAmount float64) error {
	// 使用乐观锁保证更新的原子性
	result := tx.Model(&app_model.AppWallet{}).
		Where("user_id = ? AND money >= ?", uid, amount).
		Update("money", gorm.Expr("money - ?", amount))

	if result.Error != nil {
		return fmt.Errorf("扣款失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("扣款失败: 余额不足或钱包已被修改")
	}

	// 记录余额消耗记录
	walletTransaction := app_model.AppRecharge{
		UserID:          uid,
		Description:     "order_payment",
		TransactionType: "order_payment",
		Amount:          amount,
		BalanceBefore:   allAmount,
		BalanceAfter:    allAmount - amount,
		CreateTime:      time.Now(),
	}

	if err := tx.Create(&walletTransaction).Error; err != nil {
		return fmt.Errorf("记录交易失败: %w", err)
	}

	return nil
}

// checkAndCancelOrder 检查订单状态并取消未支付订单
func (s *OrderService) checkAndCancelOrder(orderNo string) {
	log.Printf("开始检查订单状态: %s", orderNo)

	// 获取分布式锁，防止并发操作同一订单
	lockKey := fmt.Sprintf("cancel_order:%s", orderNo)
	acquired, err := s.acquireLock(lockKey, 5*time.Second)
	if err != nil {
		log.Printf("获取锁失败，无法取消订单 %s: %v", orderNo, err)
		return
	}
	if !acquired {
		log.Printf("无法获取锁，订单 %s 可能正在被其他进程处理", orderNo)
		return
	}
	defer s.releaseLock(lockKey)

	// 开始一个新的事务
	tx := db.Dao.Begin()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("处理订单 %s 时发生panic: %v", orderNo, r)
			tx.Rollback()
		}
	}()

	// 查询订单状态，使用 FOR UPDATE 锁定记录
	var checkOrder app_model.AppOrder
	err = tx.Set("gorm:query_option", "FOR UPDATE").
		Where("no = ?", orderNo).
		First(&checkOrder).Error

	if err != nil {
		log.Printf("查询订单 %s 失败: %v", orderNo, err)
		tx.Rollback()
		return
	}

	log.Printf("订单 %s 当前状态: %s", orderNo, checkOrder.Status)

	// 如果订单状态仍然是 pending，则将其状态改为 cancelled
	if checkOrder.Status == "pending" {
		// 保存旧状态
		oldStatus := checkOrder.Status

		log.Printf("订单 %s 将被取消", orderNo)

		// 更新订单状态
		err = tx.Model(&app_model.AppOrder{}).
			Where("no = ? AND status = ?", orderNo, "pending").
			Update("status", "cancelled").Error

		if err != nil {
			log.Printf("更新订单 %s 状态失败: %v", orderNo, err)
			tx.Rollback()
			return
		}

		// 恢复商品库存
		err = tx.Model(&app_model.AppGoods{}).
			Where("id = ?", checkOrder.GoodsId).
			Update("stock", gorm.Expr("stock + ?", checkOrder.Num)).Error

		if err != nil {
			log.Printf("恢复商品 %d 库存失败: %v", checkOrder.GoodsId, err)
			tx.Rollback()
			return
		}

		// 【新增】获取商品信息以更新统计数据
		var goods app_model.AppGoods
		if err := tx.Where("id = ?", checkOrder.GoodsId).First(&goods).Error; err != nil {
			log.Printf("获取商品 %d 信息失败: %v", checkOrder.GoodsId, err)
			tx.Rollback()
			return
		}

		// 【新增】更新商家收入统计数据，传入旧状态
		if err := s.updateMerchantRevenueStats(tx, goods.TenantsId, &checkOrder, goods.GoodsName, "cancel", oldStatus); err != nil {
			log.Printf("更新商家收入统计失败: %v", err)
			tx.Rollback()
			return
		}

		log.Printf("订单 %s 已成功取消，商品 %d 库存已恢复 %d 件",
			orderNo, checkOrder.GoodsId, checkOrder.Num)
	} else {
		log.Printf("订单 %s 状态为 %s，无需取消", orderNo, checkOrder.Status)
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		log.Printf("提交事务失败: %v", err)
		tx.Rollback()
		return
	}

	log.Printf("订单 %s 处理完成", orderNo)
}

// 保存锁标识符
func (s *OrderService) saveLockIdentifier(key, uuid string) {
	lockMutex.Lock()
	defer lockMutex.Unlock()
	lockIdentifiers[key] = uuid
}

// 获取锁标识符
func (s *OrderService) getLockIdentifier(key string) string {
	lockMutex.Lock()
	defer lockMutex.Unlock()
	return lockIdentifiers[key]
}

// 删除锁标识符
func (s *OrderService) deleteLockIdentifier(key string) {
	lockMutex.Lock()
	defer lockMutex.Unlock()
	delete(lockIdentifiers, key)
}

// NewOrderService 创建并返回 OrderService 实例
func NewOrderService(redisClient *redis.Client) *OrderService {
	service := &OrderService{
		redisClient: redisClient,
	}

	// 自动启动订单取消工作器
	workerMutex.Lock()
	if !workerStarted {
		service.startOrderCancellationWorker()
		workerStarted = true
	}
	workerMutex.Unlock()

	return service
}

// UpdateOrderStatus 示例：在订单状态变更的地假设我们有一个函数用于更新订单状态
func (s *OrderService) UpdateOrderStatus(orderNo string, newStatus string) error {
	tx := db.Dao.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查询订单
	var order app_model.AppOrder
	if err := tx.Where("no = ?", orderNo).First(&order).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 记录旧状态
	oldStatus := order.Status

	// 更新订单状态
	order.Status = newStatus
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 获取商品信息
	var goods app_model.AppGoods
	if err := tx.Where("id = ?", order.GoodsId).First(&goods).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新商家收入统计数据，传入旧状态
	if err := s.updateMerchantRevenueStats(tx, goods.TenantsId, &order, goods.GoodsName, "update", oldStatus); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// 发送WebSocket订单状态更新通知
	wsService := public_service.GetWebSocketService()
	err := wsService.SendOrderNotification(order.UserId, orderNo, newStatus, goods.GoodsName)
	if err != nil {
		return err
	}

	return tx.Commit().Error
}
