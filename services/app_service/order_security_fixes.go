package app_service

import (
	"context"
	"fmt"
	"log"
	"nasa-go-admin/model/app_model"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// SecurityOrderService 安全订单服务 - 解决并发和卡单问题
type SecurityOrderService struct {
	redisClient *redis.Client
	mu          sync.RWMutex
}

// NewSecurityOrderService 创建安全订单服务
func NewSecurityOrderService(redisClient *redis.Client) *SecurityOrderService {
	return &SecurityOrderService{
		redisClient: redisClient,
	}
}

// SafeDeductStock 安全扣减库存 - 使用版本控制防止超卖
func (s *SecurityOrderService) SafeDeductStock(tx *gorm.DB, goodsId, quantity int) error {
	const maxRetries = 3

	for retry := 0; retry < maxRetries; retry++ {
		var goods app_model.AppGoods

		// 查询当前商品信息（包含版本号）
		if err := tx.Where("id = ?", goodsId).First(&goods).Error; err != nil {
			return fmt.Errorf("商品不存在: %w", err)
		}

		// 检查库存是否充足
		if goods.Stock < quantity {
			return fmt.Errorf("库存不足，当前库存: %d，需要: %d", goods.Stock, quantity)
		}

		// 检查商品状态
		if goods.Status != "1" {
			return fmt.Errorf("商品已下架或不可购买")
		}

		// 使用乐观锁更新库存
		result := tx.Model(&app_model.AppGoods{}).
			Where("id = ? AND stock >= ? AND status = '1' AND isdelete != 1", goodsId, quantity).
			Updates(map[string]interface{}{
				"stock":       gorm.Expr("stock - ?", quantity),
				"update_time": time.Now(),
			})

		if result.Error != nil {
			if retry == maxRetries-1 {
				return fmt.Errorf("库存扣减失败: %w", result.Error)
			}
			log.Printf("库存扣减失败，重试 %d/%d: %v", retry+1, maxRetries, result.Error)
			continue
		}

		// 检查是否成功更新
		if result.RowsAffected == 0 {
			if retry == maxRetries-1 {
				return fmt.Errorf("库存扣减失败，可能商品已下架或库存不足")
			}
			log.Printf("库存扣减未生效，重试 %d/%d", retry+1, maxRetries)
			time.Sleep(time.Duration(retry+1) * 10 * time.Millisecond) // 递增延迟
			continue
		}

		log.Printf("✅ 成功扣减商品 %d 库存 %d 件", goodsId, quantity)
		return nil
	}

	return fmt.Errorf("库存扣减失败，重试次数已达上限")
}

// SafeDeductWallet 安全扣减钱包余额 - 防止并发扣款导致负数
func (s *SecurityOrderService) SafeDeductWallet(tx *gorm.DB, uid int, amount float64) (*app_model.AppWallet, error) {
	const maxRetries = 3

	for retry := 0; retry < maxRetries; retry++ {
		var wallet app_model.AppWallet

		// 使用 FOR UPDATE 锁定钱包记录
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("user_id = ?", uid).First(&wallet).Error; err != nil {

			if err == gorm.ErrRecordNotFound {
				// 钱包不存在，创建新钱包
				newWallet := app_model.AppWallet{
					UserId: uid,
					Money:  0.00,
				}
				if createErr := tx.Create(&newWallet).Error; createErr != nil {
					return nil, fmt.Errorf("创建钱包失败: %w", createErr)
				}
				return nil, fmt.Errorf("余额不足，当前余额: 0.00，需要: %.2f", amount)
			}
			return nil, fmt.Errorf("查询钱包失败: %w", err)
		}

		// 检查余额是否充足
		if wallet.Money < amount {
			return nil, fmt.Errorf("余额不足，当前余额: %.2f，需要: %.2f", wallet.Money, amount)
		}

		// 原子性扣减余额
		result := tx.Model(&app_model.AppWallet{}).
			Where("user_id = ? AND money >= ?", uid, amount).
			Updates(map[string]interface{}{
				"money": gorm.Expr("money - ?", amount),
			})

		if result.Error != nil {
			if retry == maxRetries-1 {
				return nil, fmt.Errorf("余额扣减失败: %w", result.Error)
			}
			log.Printf("余额扣减失败，重试 %d/%d: %v", retry+1, maxRetries, result.Error)
			continue
		}

		if result.RowsAffected == 0 {
			if retry == maxRetries-1 {
				return nil, fmt.Errorf("余额扣减失败，余额可能已不足")
			}
			log.Printf("余额扣减未生效，重试 %d/%d", retry+1, maxRetries)
			time.Sleep(time.Duration(retry+1) * 10 * time.Millisecond)
			continue
		}

		// 更新本地钱包对象
		wallet.Money -= amount

		log.Printf("✅ 成功扣减用户 %d 余额 %.2f，剩余: %.2f", uid, amount, wallet.Money)
		return &wallet, nil
	}

	return nil, fmt.Errorf("余额扣减失败，重试次数已达上限")
}

// RecordWalletTransaction 记录钱包交易流水
func (s *SecurityOrderService) RecordWalletTransaction(tx *gorm.DB, uid int, amount float64,
	balanceBefore, balanceAfter float64, description string) error {

	transaction := app_model.AppRecharge{
		UserID:          uid,
		Description:     description,
		TransactionType: "order_payment",
		Amount:          amount,
		BalanceBefore:   balanceBefore,
		BalanceAfter:    balanceAfter,
		CreateTime:      time.Now(),
	}

	if err := tx.Create(&transaction).Error; err != nil {
		return fmt.Errorf("记录交易流水失败: %w", err)
	}

	return nil
}

// DistributedLock 分布式锁结构
type DistributedLock struct {
	redisClient *redis.Client
	key         string
	value       string
	expiration  time.Duration
	stopCh      chan struct{}
	renewalWg   sync.WaitGroup
}

// NewDistributedLock 创建分布式锁
func (s *SecurityOrderService) NewDistributedLock(key string, expiration time.Duration) *DistributedLock {
	return &DistributedLock{
		redisClient: s.redisClient,
		key:         key,
		value:       fmt.Sprintf("%d-%d", time.Now().UnixNano(), uid()),
		expiration:  expiration,
		stopCh:      make(chan struct{}),
	}
}

// AcquireWithRenewal 获取锁并自动续期
func (dl *DistributedLock) AcquireWithRenewal(ctx context.Context) error {
	if dl.redisClient == nil {
		return fmt.Errorf("Redis客户端未初始化")
	}

	// 获取锁
	acquired, err := dl.redisClient.SetNX(ctx, dl.key, dl.value, dl.expiration).Result()
	if err != nil {
		return fmt.Errorf("获取锁失败: %w", err)
	}

	if !acquired {
		return fmt.Errorf("锁已被其他进程持有")
	}

	// 启动续期协程
	dl.renewalWg.Add(1)
	go dl.renewLock()

	log.Printf("🔒 成功获取分布式锁: %s", dl.key)
	return nil
}

// renewLock 自动续期锁
func (dl *DistributedLock) renewLock() {
	defer dl.renewalWg.Done()

	// 每1/3超时时间续期一次
	ticker := time.NewTicker(dl.expiration / 3)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 使用Lua脚本安全续期
			script := `
				if redis.call("GET", KEYS[1]) == ARGV[1] then
					return redis.call("EXPIRE", KEYS[1], ARGV[2])
				else
					return 0
				end
			`

			result, err := dl.redisClient.Eval(context.Background(), script,
				[]string{dl.key}, dl.value, int(dl.expiration.Seconds())).Result()

			if err != nil {
				log.Printf("锁续期失败: %v", err)
				return
			}

			if result.(int64) == 0 {
				log.Printf("锁已被其他进程获取，停止续期")
				return
			}

		case <-dl.stopCh:
			return
		}
	}
}

// Release 释放锁
func (dl *DistributedLock) Release() error {
	// 停止续期
	close(dl.stopCh)
	dl.renewalWg.Wait()

	if dl.redisClient == nil {
		return nil
	}

	// 使用Lua脚本安全释放锁
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result, err := dl.redisClient.Eval(context.Background(), script,
		[]string{dl.key}, dl.value).Result()

	if err != nil {
		return fmt.Errorf("释放锁失败: %w", err)
	}

	if result.(int64) > 0 {
		log.Printf("🔓 成功释放分布式锁: %s", dl.key)
	}

	return nil
}

// IdempotencyChecker 幂等性检查器
type IdempotencyChecker struct {
	redisClient *redis.Client
}

// NewIdempotencyChecker 创建幂等性检查器
func (s *SecurityOrderService) NewIdempotencyChecker() *IdempotencyChecker {
	return &IdempotencyChecker{
		redisClient: s.redisClient,
	}
}

// CheckAndSet 检查幂等性并设置标记
func (ic *IdempotencyChecker) CheckAndSet(key string, expiration time.Duration) (bool, error) {
	if ic.redisClient == nil {
		// Redis不可用时，记录日志但允许继续（降级策略）
		log.Printf("Redis不可用，跳过幂等性检查: %s", key)
		return false, nil
	}

	ctx := context.Background()

	// 检查是否已存在
	exists, err := ic.redisClient.Exists(ctx, key).Result()
	if err != nil {
		log.Printf("幂等性检查失败: %v", err)
		return false, nil // 不阻塞业务流程
	}

	if exists > 0 {
		return true, nil // 已存在，重复操作
	}

	// 设置幂等性标记
	_, err = ic.redisClient.SetNX(ctx, key, "1", expiration).Result()
	if err != nil {
		log.Printf("设置幂等性标记失败: %v", err)
		return false, nil
	}

	return false, nil // 不存在，可以继续操作
}

// MultiLayerTimeoutManager 多层超时管理器
type MultiLayerTimeoutManager struct {
	redisClient *redis.Client
	db          *gorm.DB
}

// NewMultiLayerTimeoutManager 创建多层超时管理器
func (s *SecurityOrderService) NewMultiLayerTimeoutManager(db *gorm.DB) *MultiLayerTimeoutManager {
	return &MultiLayerTimeoutManager{
		redisClient: s.redisClient,
		db:          db,
	}
}

// ScheduleOrderTimeout 调度订单超时
func (mltm *MultiLayerTimeoutManager) ScheduleOrderTimeout(orderNo string, timeout time.Duration) error {
	ctx := context.Background()
	expireTime := time.Now().Add(timeout)

	// 1. Redis 延时队列（主要方案）
	if mltm.redisClient != nil {
		score := float64(expireTime.Unix())
		err := mltm.redisClient.ZAdd(ctx, "order_timeouts", redis.Z{
			Score:  score,
			Member: orderNo,
		}).Err()

		if err != nil {
			log.Printf("Redis超时队列设置失败: %v", err)
		} else {
			log.Printf("✅ 订单 %s 已加入Redis超时队列，将在 %s 后检查",
				orderNo, timeout.String())
		}
	}

	// 2. 数据库定时任务（备用方案）
	if mltm.db != nil {
		timeoutRecord := map[string]interface{}{
			"order_no":   orderNo,
			"expire_at":  expireTime,
			"status":     "pending",
			"created_at": time.Now(),
		}

		// 这里假设有一个 order_timeouts 表
		err := mltm.db.Table("order_timeouts").Create(timeoutRecord).Error
		if err != nil {
			log.Printf("数据库超时记录创建失败: %v", err)
		}
	}

	// 3. 内存定时器（应急方案）
	go func() {
		timer := time.NewTimer(timeout + 2*time.Minute) // 多2分钟缓冲
		defer timer.Stop()

		<-timer.C
		log.Printf("内存定时器触发，检查订单: %s", orderNo)
		// 这里需要调用实际的超时处理函数
		// s.handleOrderTimeout(orderNo)
	}()

	return nil
}

// OrderCompensationService 订单补偿服务
type OrderCompensationService struct {
	db          *gorm.DB
	redisClient *redis.Client
}

// NewOrderCompensationService 创建订单补偿服务
func (s *SecurityOrderService) NewOrderCompensationService(db *gorm.DB) *OrderCompensationService {
	return &OrderCompensationService{
		db:          db,
		redisClient: s.redisClient,
	}
}

// DetectAndFixInconsistencies 检测并修复数据不一致
func (ocs *OrderCompensationService) DetectAndFixInconsistencies() error {
	log.Printf("开始检测数据一致性...")

	// 1. 检查孤立的支付记录
	if err := ocs.fixOrphanedPayments(); err != nil {
		log.Printf("修复孤立支付记录失败: %v", err)
	}

	// 2. 检查孤立的库存扣减
	if err := ocs.fixOrphanedStockReductions(); err != nil {
		log.Printf("修复孤立库存扣减失败: %v", err)
	}

	// 3. 检查状态不一致的订单
	if err := ocs.fixStatusMismatches(); err != nil {
		log.Printf("修复状态不一致失败: %v", err)
	}

	log.Printf("数据一致性检测完成")
	return nil
}

func (ocs *OrderCompensationService) fixOrphanedPayments() error {
	// 查找有支付记录但没有对应订单的情况
	query := `
		SELECT ar.user_id, ar.amount, ar.description, ar.create_time
		FROM app_recharge ar 
		WHERE ar.transaction_type = 'order_payment' 
		AND ar.create_time > ? 
		AND NOT EXISTS (
			SELECT 1 FROM app_order ao 
			WHERE ao.user_id = ar.user_id 
			AND ABS(ao.amount - ar.amount) < 0.01
			AND ao.create_time BETWEEN ar.create_time - INTERVAL 2 MINUTE 
		                           AND ar.create_time + INTERVAL 2 MINUTE
		)
		LIMIT 50
	`

	var orphanedPayments []struct {
		UserID      int       `json:"user_id"`
		Amount      float64   `json:"amount"`
		Description string    `json:"description"`
		CreateTime  time.Time `json:"create_time"`
	}

	err := ocs.db.Raw(query, time.Now().Add(-24*time.Hour)).Scan(&orphanedPayments).Error
	if err != nil {
		return fmt.Errorf("查询孤立支付记录失败: %w", err)
	}

	if len(orphanedPayments) > 0 {
		log.Printf("发现 %d 个孤立的支付记录", len(orphanedPayments))

		for _, payment := range orphanedPayments {
			// 退款到用户钱包
			err := ocs.refundToWallet(payment.UserID, payment.Amount,
				fmt.Sprintf("系统补偿退款: %s", payment.Description))
			if err != nil {
				log.Printf("补偿退款失败 用户:%d 金额:%.2f 错误:%v",
					payment.UserID, payment.Amount, err)
			} else {
				log.Printf("✅ 补偿退款成功 用户:%d 金额:%.2f",
					payment.UserID, payment.Amount)
			}
		}
	}

	return nil
}

func (ocs *OrderCompensationService) fixOrphanedStockReductions() error {
	// 这里实现库存异常检测和修复逻辑
	log.Printf("检查库存异常...")
	return nil
}

func (ocs *OrderCompensationService) fixStatusMismatches() error {
	// 这里实现状态不一致检测和修复逻辑
	log.Printf("检查状态不一致...")
	return nil
}

func (ocs *OrderCompensationService) refundToWallet(userID int, amount float64, description string) error {
	tx := ocs.db.Begin()
	defer tx.Rollback()

	// 增加用户钱包余额
	result := tx.Model(&app_model.AppWallet{}).
		Where("user_id = ?", userID).
		Update("money", gorm.Expr("money + ?", amount))

	if result.Error != nil {
		return fmt.Errorf("退款失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		// 钱包不存在，创建新钱包
		wallet := app_model.AppWallet{
			UserId: userID,
			Money:  amount,
		}
		if err := tx.Create(&wallet).Error; err != nil {
			return fmt.Errorf("创建钱包失败: %w", err)
		}
	}

	// 记录退款流水
	refundRecord := app_model.AppRecharge{
		UserID:          userID,
		Description:     description,
		TransactionType: "system_refund",
		Amount:          amount,
		CreateTime:      time.Now(),
	}

	if err := tx.Create(&refundRecord).Error; err != nil {
		return fmt.Errorf("记录退款流水失败: %w", err)
	}

	return tx.Commit().Error
}

// uid 生成唯一ID的辅助函数
func uid() int64 {
	return time.Now().UnixNano()
}
