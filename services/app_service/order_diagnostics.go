package app_service

import (
	"context"
	"fmt"
	"log"
	"nasa-go-admin/db"
	"nasa-go-admin/model/app_model"
	"time"

	"github.com/redis/go-redis/v9"
)

// OrderDiagnostics 订单系统诊断工具
type OrderDiagnostics struct {
	redisClient *redis.Client
}

// NewOrderDiagnostics 创建订单诊断工具
func NewOrderDiagnostics(redisClient *redis.Client) *OrderDiagnostics {
	return &OrderDiagnostics{
		redisClient: redisClient,
	}
}

// DiagnosisReport 诊断报告
type DiagnosisReport struct {
	Timestamp           time.Time              `json:"timestamp"`
	PendingOrdersCount  int64                  `json:"pending_orders_count"`
	ExpiredOrdersCount  int64                  `json:"expired_orders_count"`
	ActiveLocksCount    int64                  `json:"active_locks_count"`
	TimeoutQueueSize    int64                  `json:"timeout_queue_size"`
	LockDetails         []LockInfo             `json:"lock_details"`
	ExpiredOrderDetails []ExpiredOrderInfo     `json:"expired_order_details"`
	Recommendations     []string               `json:"recommendations"`
	SystemHealth        map[string]interface{} `json:"system_health"`
}

// LockInfo 锁信息
type LockInfo struct {
	Key       string        `json:"key"`
	TTL       time.Duration `json:"ttl"`
	OrderNo   string        `json:"order_no,omitempty"`
	LockType  string        `json:"lock_type"`
	IsExpired bool          `json:"is_expired"`
}

// ExpiredOrderInfo 过期订单信息
type ExpiredOrderInfo struct {
	OrderNo    string    `json:"order_no"`
	UserId     int       `json:"user_id"`
	GoodsId    int       `json:"goods_id"`
	Amount     float64   `json:"amount"`
	CreateTime time.Time `json:"create_time"`
	ExpiredFor string    `json:"expired_for"`
	HasLock    bool      `json:"has_lock"`
}

// RunFullDiagnosis 运行完整的系统诊断
func (od *OrderDiagnostics) RunFullDiagnosis() (*DiagnosisReport, error) {
	log.Printf("🔍 开始运行订单系统诊断...")

	report := &DiagnosisReport{
		Timestamp:       time.Now(),
		LockDetails:     []LockInfo{},
		Recommendations: []string{},
	}

	// 1. 检查pending订单数量
	if err := od.checkPendingOrders(report); err != nil {
		log.Printf("检查pending订单失败: %v", err)
	}

	// 2. 检查过期订单
	if err := od.checkExpiredOrders(report); err != nil {
		log.Printf("检查过期订单失败: %v", err)
	}

	// 3. 检查活跃锁
	if err := od.checkActiveLocks(report); err != nil {
		log.Printf("检查活跃锁失败: %v", err)
	}

	// 4. 检查超时队列
	if err := od.checkTimeoutQueue(report); err != nil {
		log.Printf("检查超时队列失败: %v", err)
	}

	// 5. 生成建议
	od.generateRecommendations(report)

	// 6. 系统健康检查
	report.SystemHealth = od.getSystemHealth()

	log.Printf("✅ 订单系统诊断完成")
	return report, nil
}

// checkPendingOrders 检查pending状态的订单
func (od *OrderDiagnostics) checkPendingOrders(report *DiagnosisReport) error {
	var count int64
	err := db.Dao.Model(&app_model.AppOrder{}).
		Where("status = ?", "pending").
		Count(&count).Error

	if err != nil {
		return err
	}

	report.PendingOrdersCount = count
	log.Printf("当前pending订单数量: %d", count)
	return nil
}

// checkExpiredOrders 检查过期订单
func (od *OrderDiagnostics) checkExpiredOrders(report *DiagnosisReport) error {
	expireTime := time.Now().Add(-15 * time.Minute)

	var expiredOrders []app_model.AppOrder
	err := db.Dao.Where("status = ? AND create_time < ?", "pending", expireTime).
		Limit(50).
		Find(&expiredOrders).Error

	if err != nil {
		return err
	}

	report.ExpiredOrdersCount = int64(len(expiredOrders))
	report.ExpiredOrderDetails = make([]ExpiredOrderInfo, len(expiredOrders))

	for i, order := range expiredOrders {
		expiredFor := time.Since(order.CreateTime.Add(15 * time.Minute))
		hasLock := od.checkOrderLock(order.No)

		report.ExpiredOrderDetails[i] = ExpiredOrderInfo{
			OrderNo:    order.No,
			UserId:     order.UserId,
			GoodsId:    order.GoodsId,
			Amount:     order.Amount,
			CreateTime: order.CreateTime,
			ExpiredFor: expiredFor.String(),
			HasLock:    hasLock,
		}
	}

	log.Printf("发现过期订单数量: %d", len(expiredOrders))
	return nil
}

// checkActiveLocks 检查活跃的分布式锁
func (od *OrderDiagnostics) checkActiveLocks(report *DiagnosisReport) error {
	if od.redisClient == nil {
		return nil
	}

	ctx := context.Background()

	// 查找所有订单相关的锁
	lockPatterns := []string{
		"cancel_order:*",
		"create_order:*",
		"goods_stock:*",
		"payment:*",
	}

	var allLocks []LockInfo

	for _, pattern := range lockPatterns {
		keys, err := od.redisClient.Keys(ctx, pattern).Result()
		if err != nil {
			log.Printf("查询锁键失败 %s: %v", pattern, err)
			continue
		}

		for _, key := range keys {
			ttl, err := od.redisClient.TTL(ctx, key).Result()
			if err != nil {
				continue
			}

			lockInfo := LockInfo{
				Key:       key,
				TTL:       ttl,
				IsExpired: ttl < 0,
			}

			// 解析锁类型和订单号
			if len(key) > 12 && key[:12] == "cancel_order" {
				lockInfo.LockType = "cancel"
				if len(key) > 13 {
					lockInfo.OrderNo = key[13:]
				}
			} else if len(key) > 12 && key[:12] == "create_order" {
				lockInfo.LockType = "create"
			} else if len(key) > 11 && key[:11] == "goods_stock" {
				lockInfo.LockType = "stock"
			} else if len(key) > 8 && key[:8] == "payment:" {
				lockInfo.LockType = "payment"
				if len(key) > 8 {
					lockInfo.OrderNo = key[8:]
				}
			}

			allLocks = append(allLocks, lockInfo)
		}
	}

	report.ActiveLocksCount = int64(len(allLocks))
	report.LockDetails = allLocks

	log.Printf("发现活跃锁数量: %d", len(allLocks))
	return nil
}

// checkTimeoutQueue 检查超时队列
func (od *OrderDiagnostics) checkTimeoutQueue(report *DiagnosisReport) error {
	if od.redisClient == nil {
		return nil
	}

	ctx := context.Background()
	size, err := od.redisClient.ZCard(ctx, "order_timeouts").Result()
	if err != nil {
		return err
	}

	report.TimeoutQueueSize = size
	log.Printf("超时队列大小: %d", size)
	return nil
}

// checkOrderLock 检查特定订单是否有锁
func (od *OrderDiagnostics) checkOrderLock(orderNo string) bool {
	if od.redisClient == nil {
		return false
	}

	ctx := context.Background()
	lockKey := fmt.Sprintf("cancel_order:%s", orderNo)

	exists, err := od.redisClient.Exists(ctx, lockKey).Result()
	return err == nil && exists > 0
}

// generateRecommendations 生成建议
func (od *OrderDiagnostics) generateRecommendations(report *DiagnosisReport) {
	recommendations := []string{}

	// 基于过期订单数量的建议
	if report.ExpiredOrdersCount > 10 {
		recommendations = append(recommendations,
			"过期订单数量较多，建议检查订单取消机制是否正常工作")
	}

	// 基于活跃锁数量的建议
	if report.ActiveLocksCount > 50 {
		recommendations = append(recommendations,
			"活跃锁数量较多，可能存在锁泄漏问题，建议检查锁的释放机制")
	}

	// 基于超时队列大小的建议
	if report.TimeoutQueueSize > 100 {
		recommendations = append(recommendations,
			"超时队列堆积严重，建议增加处理频率或检查处理逻辑")
	}

	// 检查过期锁
	expiredLockCount := 0
	for _, lock := range report.LockDetails {
		if lock.IsExpired {
			expiredLockCount++
		}
	}

	if expiredLockCount > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("发现%d个过期锁，建议清理", expiredLockCount))
	}

	// 检查有锁但过期的订单
	lockedExpiredOrders := 0
	for _, order := range report.ExpiredOrderDetails {
		if order.HasLock {
			lockedExpiredOrders++
		}
	}

	if lockedExpiredOrders > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("发现%d个有锁但过期的订单，可能存在死锁问题", lockedExpiredOrders))
	}

	report.Recommendations = recommendations
}

// getSystemHealth 获取系统健康状态
func (od *OrderDiagnostics) getSystemHealth() map[string]interface{} {
	health := map[string]interface{}{
		"database": "unknown",
		"redis":    "unknown",
	}

	// 检查数据库健康状态
	if db.Dao != nil {
		if sqlDB, err := db.Dao.DB(); err == nil && sqlDB.Ping() == nil {
			health["database"] = "healthy"
		} else {
			health["database"] = "unhealthy"
		}
	}

	// 检查Redis健康状态
	if od.redisClient != nil {
		if err := od.redisClient.Ping(context.Background()).Err(); err == nil {
			health["redis"] = "healthy"
		} else {
			health["redis"] = "unhealthy"
		}
	}

	return health
}

// CleanupExpiredLocks 清理过期的锁
func (od *OrderDiagnostics) CleanupExpiredLocks() error {
	if od.redisClient == nil {
		return fmt.Errorf("Redis客户端未初始化")
	}

	ctx := context.Background()
	patterns := []string{
		"cancel_order:*",
		"create_order:*",
		"goods_stock:*",
		"payment:*",
	}

	cleanedCount := 0
	for _, pattern := range patterns {
		keys, err := od.redisClient.Keys(ctx, pattern).Result()
		if err != nil {
			continue
		}

		for _, key := range keys {
			ttl, err := od.redisClient.TTL(ctx, key).Result()
			if err != nil {
				continue
			}

			// 删除过期的锁
			if ttl < 0 {
				od.redisClient.Del(ctx, key)
				cleanedCount++
			}
		}
	}

	log.Printf("清理了 %d 个过期锁", cleanedCount)
	return nil
}

// ForceUnlockOrder 强制解锁特定订单（紧急情况使用）
func (od *OrderDiagnostics) ForceUnlockOrder(orderNo string) error {
	if od.redisClient == nil {
		return fmt.Errorf("Redis客户端未初始化")
	}

	ctx := context.Background()
	lockKey := fmt.Sprintf("cancel_order:%s", orderNo)

	result, err := od.redisClient.Del(ctx, lockKey).Result()
	if err != nil {
		return fmt.Errorf("强制解锁失败: %w", err)
	}

	if result > 0 {
		log.Printf("🔓 强制解锁订单 %s 成功", orderNo)
	} else {
		log.Printf("订单 %s 没有找到对应的锁", orderNo)
	}

	return nil
}

// 全局诊断实例
var globalDiagnostics *OrderDiagnostics

// InitGlobalDiagnostics 初始化全局诊断工具
func InitGlobalDiagnostics(redisClient *redis.Client) {
	globalDiagnostics = NewOrderDiagnostics(redisClient)
	log.Printf("✅ 订单诊断工具已初始化")
}

// GetGlobalDiagnostics 获取全局诊断工具
func GetGlobalDiagnostics() *OrderDiagnostics {
	return globalDiagnostics
}
