package app_service

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"nasa-go-admin/inout"
	"nasa-go-admin/utils"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var (
	lockIdentifiers = make(map[string]string)
	lockMutex       = &sync.Mutex{}
	workerStarted   = false
	workerMutex     = &sync.Mutex{}
)

// 旧版本OrderService - 已废弃，使用SecureOrderCreator替代
// 这个结构体保留仅为了兼容性，不应该用于新功能
type OrderService struct {
	redisClient      *redis.Client
	dbPoolManager    *DatabasePoolManager
	queryOptimizer   *QueryOptimizer
	metricsCollector *MetricsCollector
}

// NewOrderService 创建并返回 OrderService 实例 - 已废弃
// 建议使用 NewSecureOrderCreator 替代
func NewOrderService(redisClient *redis.Client) *OrderService {
	log.Printf("⚠️  警告: OrderService已废弃，请使用SecureOrderCreator")
	return &OrderService{
		redisClient: redisClient,
	}
}

// 以下代码是遗留代码，建议使用SecureOrderCreator中的对应方法
// 保留这些方法仅为了向后兼容，但不推荐使用

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

// CreateOrder - 已废弃，使用SecureOrderCreator.CreateOrderSecurely替代
func (s *OrderService) CreateOrder(c *gin.Context, uid int, params inout.CreateOrderReq) (string, error) {
	log.Printf("⚠️  警告: OrderService.CreateOrder已废弃，请使用SecureOrderCreator.CreateOrderSecurely")
	return "", fmt.Errorf("此方法已废弃，请使用新的安全订单创建器")
}

// 以下是仍然需要保留的辅助方法，等待完全迁移到SecureOrderCreator

// 生成订单号的辅助函数
func generateOrderNo(uid, goodsId int) string {
	timestamp := time.Now().Format("20060102150405")
	random := rand.Intn(1000)
	return fmt.Sprintf("%s%d%d%03d", timestamp, uid, goodsId, random)
}

// 分布式锁相关的方法（待迁移）
func (s *OrderService) acquireLock(key string, expiration time.Duration) (bool, error) {
	if s.redisClient == nil {
		log.Printf("ERROR: Redis client is nil when acquiring lock: %s", key)
		return true, nil
	}

	// 生成唯一标识符
	uuid := fmt.Sprintf("%d-%s", time.Now().UnixNano(), utils.RandomString(16))

	// 尝试获取锁
	result, err := s.redisClient.SetNX(context.Background(), key, uuid, expiration).Result()
	if err != nil {
		log.Printf("获取锁失败: %s, 错误: %v", key, err)
		return false, err
	}

	if result {
		s.saveLockIdentifier(key, uuid)
		log.Printf("成功获取锁: %s", key)
		return true, nil
	}

	log.Printf("锁已被其他进程持有: %s", key)
	return false, nil
}

// 释放锁
func (s *OrderService) releaseLock(key string) error {
	if s.redisClient == nil {
		return nil
	}

	uuid := s.getLockIdentifier(key)
	if uuid == "" {
		log.Printf("未找到锁标识符: %s", key)
		return nil
	}

	// 使用 Lua 脚本确保只释放自己持有的锁
	luaScript := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := s.redisClient.Eval(context.Background(), luaScript, []string{key}, uuid).Result()
	if err != nil {
		log.Printf("释放锁失败: %s, 错误: %v", key, err)
		return err
	}

	s.deleteLockIdentifier(key)

	if result.(int64) == 1 {
		log.Printf("成功释放锁: %s", key)
	} else {
		log.Printf("锁已被其他进程释放或已过期: %s", key)
	}

	return nil
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
