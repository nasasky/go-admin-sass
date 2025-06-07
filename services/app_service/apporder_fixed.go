package app_service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"nasa-go-admin/db"
	"nasa-go-admin/model/app_model"
	"nasa-go-admin/pkg/goroutinepool"
	"nasa-go-admin/services/public_service"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// FixedOrderService 修复后的订单服务
type FixedOrderService struct {
	redisClient *redis.Client
	mu          sync.RWMutex
}

// CheckAndCancelOrderFixed 修复后的订单检查和取消方法
func (s *FixedOrderService) CheckAndCancelOrderFixed(orderNo string) error {
	log.Printf("开始检查订单状态: %s", orderNo)

	// 获取分布式锁，防止并发操作同一订单
	lockKey := fmt.Sprintf("cancel_order:%s", orderNo)
	acquired, err := s.acquireLock(lockKey, 10*time.Second) // 增加锁定时间
	if err != nil {
		return fmt.Errorf("获取锁失败，无法取消订单 %s: %w", orderNo, err)
	}
	if !acquired {
		return fmt.Errorf("无法获取锁，订单 %s 可能正在被其他进程处理", orderNo)
	}
	defer func() {
		if releaseErr := s.releaseLock(lockKey); releaseErr != nil {
			log.Printf("释放锁失败: %v", releaseErr)
		}
	}()

	// 使用上下文控制超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 开始事务
	tx := db.Dao.WithContext(ctx).Begin(&sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})

	defer func() {
		if r := recover(); r != nil {
			log.Printf("处理订单 %s 时发生panic: %v", orderNo, r)
			if rbErr := tx.Rollback().Error; rbErr != nil {
				log.Printf("回滚事务失败: %v", rbErr)
			}
			panic(r) // 重新抛出panic
		}
	}()

	// 查询并锁定订单记录
	var checkOrder app_model.AppOrder
	err = tx.Set("gorm:query_option", "FOR UPDATE").
		Where("no = ?", orderNo).
		First(&checkOrder).Error

	if err != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			log.Printf("回滚事务失败: %v", rbErr)
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("订单 %s 不存在", orderNo)
		}
		return fmt.Errorf("查询订单 %s 失败: %w", orderNo, err)
	}

	log.Printf("订单 %s 当前状态: %s", orderNo, checkOrder.Status)

	// 只处理pending状态的订单
	if checkOrder.Status != "pending" {
		if err := tx.Commit().Error; err != nil {
			log.Printf("提交事务失败: %v", err)
			return fmt.Errorf("提交事务失败: %w", err)
		}
		log.Printf("订单 %s 状态为 %s，无需取消", orderNo, checkOrder.Status)
		return nil
	}

	// 执行订单取消操作
	if err := s.performOrderCancellation(tx, &checkOrder); err != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			log.Printf("回滚事务失败: %v", rbErr)
		}
		return fmt.Errorf("取消订单失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	log.Printf("订单 %s 已成功取消", orderNo)

	// 异步发送通知
	s.sendCancellationNotification(&checkOrder)

	return nil
}

// performOrderCancellation 执行订单取消操作
func (s *FixedOrderService) performOrderCancellation(tx *gorm.DB, order *app_model.AppOrder) error {
	oldStatus := order.Status

	// 1. 更新订单状态 - 使用条件更新确保状态一致性
	result := tx.Model(&app_model.AppOrder{}).
		Where("no = ? AND status = ?", order.No, "pending").
		Updates(map[string]interface{}{
			"status":      "cancelled",
			"update_time": time.Now(), // 手动设置更新时间
		})

	if result.Error != nil {
		return fmt.Errorf("更新订单状态失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("订单状态已被其他进程修改，无法取消")
	}

	// 2. 恢复商品库存 - 使用原子操作
	result = tx.Model(&app_model.AppGoods{}).
		Where("id = ?", order.GoodsId).
		Update("stock", gorm.Expr("stock + ?", order.Num))

	if result.Error != nil {
		return fmt.Errorf("恢复商品 %d 库存失败: %w", order.GoodsId, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("商品 %d 不存在，无法恢复库存", order.GoodsId)
	}

	// 3. 获取商品信息用于统计
	var goods app_model.AppGoods
	if err := tx.Where("id = ?", order.GoodsId).First(&goods).Error; err != nil {
		return fmt.Errorf("获取商品 %d 信息失败: %w", order.GoodsId, err)
	}

	// 4. 更新商家收入统计 - 使用独立的服务方法
	statsService := NewMerchantStatsService()
	if err := statsService.UpdateStatsForCancellation(tx, &goods, order, oldStatus); err != nil {
		log.Printf("更新商家收入统计失败: %v", err)
		// 统计更新失败不应该阻止订单取消，只记录日志
	}

	return nil
}

// sendCancellationNotification 异步发送取消通知
func (s *FixedOrderService) sendCancellationNotification(order *app_model.AppOrder) {
	goroutinepool.Submit(func() error {
		wsService := public_service.GetWebSocketService()
		return wsService.SendOrderNotification(
			order.UserId,
			order.No,
			"cancelled",
			fmt.Sprintf("订单已自动取消 - %s", order.No),
		)
	})
}

// UpdateOrderStatusFixed 修复后的订单状态更新方法
func (s *FixedOrderService) UpdateOrderStatusFixed(orderNo string, newStatus string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 开始事务
	tx := db.Dao.WithContext(ctx).Begin(&sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})

	defer func() {
		if r := recover(); r != nil {
			if rbErr := tx.Rollback().Error; rbErr != nil {
				log.Printf("回滚事务失败: %v", rbErr)
			}
			panic(r)
		}
	}()

	// 查询并锁定订单
	var order app_model.AppOrder
	if err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("no = ?", orderNo).First(&order).Error; err != nil {

		if rbErr := tx.Rollback().Error; rbErr != nil {
			log.Printf("回滚事务失败: %v", rbErr)
		}
		return fmt.Errorf("查询订单失败: %w", err)
	}

	// 验证状态转换的合法性
	if !s.isValidStatusTransition(order.Status, newStatus) {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			log.Printf("回滚事务失败: %v", rbErr)
		}
		return fmt.Errorf("无效的状态转换: %s -> %s", order.Status, newStatus)
	}

	oldStatus := order.Status

	// 更新订单状态
	result := tx.Model(&order).Updates(map[string]interface{}{
		"status":      newStatus,
		"update_time": time.Now(),
	})

	if result.Error != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			log.Printf("回滚事务失败: %v", rbErr)
		}
		return fmt.Errorf("更新订单状态失败: %w", result.Error)
	}

	// 获取商品信息
	var goods app_model.AppGoods
	if err := tx.Where("id = ?", order.GoodsId).First(&goods).Error; err != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			log.Printf("回滚事务失败: %v", rbErr)
		}
		return fmt.Errorf("获取商品信息失败: %w", err)
	}

	// 更新统计数据
	statsService := NewMerchantStatsService()
	if err := statsService.UpdateStatsForStatusChange(tx, &goods, &order, oldStatus, newStatus); err != nil {
		log.Printf("更新商家收入统计失败: %v", err)
		// 统计更新失败不阻止状态更新，只记录日志
	}

	// 提交事务 - 只提交一次！
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	// 异步发送通知
	s.sendStatusUpdateNotification(&order, &goods, newStatus)

	return nil
}

// sendStatusUpdateNotification 异步发送状态更新通知
func (s *FixedOrderService) sendStatusUpdateNotification(order *app_model.AppOrder, goods *app_model.AppGoods, newStatus string) {
	goroutinepool.Submit(func() error {
		wsService := public_service.GetWebSocketService()
		return wsService.SendOrderNotification(
			order.UserId,
			order.No,
			newStatus,
			goods.GoodsName,
		)
	})
}

// isValidStatusTransition 验证状态转换是否合法
func (s *FixedOrderService) isValidStatusTransition(oldStatus, newStatus string) bool {
	validTransitions := map[string][]string{
		"pending":   {"paid", "cancelled"},
		"paid":      {"shipped", "refunded"},
		"shipped":   {"completed", "refunded"},
		"completed": {"refunded"},
		"cancelled": {}, // 已取消的订单不能再变更
		"refunded":  {}, // 已退款的订单不能再变更
	}

	allowedStates, exists := validTransitions[oldStatus]
	if !exists {
		return false
	}

	for _, allowed := range allowedStates {
		if allowed == newStatus {
			return true
		}
	}

	return false
}

// BatchCheckExpiredOrders 批量检查过期订单 - 优化版本
func (s *FixedOrderService) BatchCheckExpiredOrders() error {
	log.Printf("开始批量检查过期订单")

	expireTime := time.Now().Add(-15 * time.Minute)

	// 分批查询过期订单，避免一次性加载过多数据
	const batchSize = 100
	var processedCount int

	for {
		var expiredOrders []app_model.AppOrder

		err := db.Dao.Where("status = ? AND create_time < ?", "pending", expireTime).
			Limit(batchSize).
			Find(&expiredOrders).Error

		if err != nil {
			return fmt.Errorf("查询过期订单失败: %w", err)
		}

		if len(expiredOrders) == 0 {
			break
		}

		log.Printf("发现 %d 个过期订单", len(expiredOrders))

		// 并发处理过期订单，但限制并发数
		semaphore := make(chan struct{}, 10) // 最多10个并发
		var wg sync.WaitGroup

		for _, order := range expiredOrders {
			wg.Add(1)

			goroutinepool.Submit(func() error {
				defer wg.Done()

				semaphore <- struct{}{}        // 获取信号量
				defer func() { <-semaphore }() // 释放信号量

				if err := s.CheckAndCancelOrderFixed(order.No); err != nil {
					log.Printf("处理过期订单 %s 失败: %v", order.No, err)
				}
				return nil
			})
		}

		wg.Wait()
		processedCount += len(expiredOrders)

		// 如果查询到的数量少于批次大小，说明没有更多数据了
		if len(expiredOrders) < batchSize {
			break
		}
	}

	log.Printf("批量检查完成，共处理 %d 个订单", processedCount)
	return nil
}

// acquireLock 获取分布式锁
func (s *FixedOrderService) acquireLock(key string, expiration time.Duration) (bool, error) {
	if s.redisClient == nil {
		return false, fmt.Errorf("Redis客户端未初始化")
	}

	ctx := context.Background()

	// 使用SET命令的NX选项实现分布式锁
	result, err := s.redisClient.SetNX(ctx, key, "locked", expiration).Result()
	if err != nil {
		return false, fmt.Errorf("获取锁失败: %w", err)
	}

	return result, nil
}

// releaseLock 释放分布式锁
func (s *FixedOrderService) releaseLock(key string) error {
	if s.redisClient == nil {
		return fmt.Errorf("Redis客户端未初始化")
	}

	ctx := context.Background()

	// 使用DEL命令释放锁
	_, err := s.redisClient.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("释放锁失败: %w", err)
	}

	return nil
}

// MerchantStatsService 商家统计服务
type MerchantStatsService struct{}

func NewMerchantStatsService() *MerchantStatsService {
	return &MerchantStatsService{}
}

// UpdateStatsForCancellation 为订单取消更新统计
func (m *MerchantStatsService) UpdateStatsForCancellation(tx *gorm.DB, goods *app_model.AppGoods, order *app_model.AppOrder, oldStatus string) error {
	// 这里实现统计更新逻辑
	// 由于统计表结构复杂，这里简化处理
	log.Printf("更新商家 %d 的统计数据，订单 %s 从 %s 变为 cancelled", goods.TenantsId, order.No, oldStatus)
	return nil
}

// UpdateStatsForStatusChange 为状态变更更新统计
func (m *MerchantStatsService) UpdateStatsForStatusChange(tx *gorm.DB, goods *app_model.AppGoods, order *app_model.AppOrder, oldStatus, newStatus string) error {
	// 这里实现统计更新逻辑
	log.Printf("更新商家 %d 的统计数据，订单 %s 从 %s 变为 %s", goods.TenantsId, order.No, oldStatus, newStatus)
	return nil
}

// NewFixedOrderService 创建修复后的订单服务实例
func NewFixedOrderService(redisClient *redis.Client) *FixedOrderService {
	return &FixedOrderService{
		redisClient: redisClient,
	}
}
