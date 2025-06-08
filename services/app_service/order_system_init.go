package app_service

import (
	"context"
	"fmt"
	"log"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/app_model"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// OrderSystemManager 订单系统管理器
type OrderSystemManager struct {
	secureCreator     *SecureOrderCreator
	monitoringService *OrderMonitoringService
	compensationSvc   *OrderCompensationService
	redisClient       *redis.Client
	isInitialized     bool
}

// NewOrderSystemManager 创建订单系统管理器
func NewOrderSystemManager(redisClient *redis.Client) *OrderSystemManager {
	return &OrderSystemManager{
		redisClient:   redisClient,
		isInitialized: false,
	}
}

// Initialize 初始化整个订单系统
func (osm *OrderSystemManager) Initialize() error {
	if osm.isInitialized {
		log.Printf("订单系统已经初始化，跳过重复初始化")
		return nil
	}

	log.Printf("🚀 开始初始化订单安全系统...")

	// 1. 初始化安全订单创建器
	if err := osm.initSecureOrderCreator(); err != nil {
		return err
	}

	// 2. 初始化监控服务
	if err := osm.initMonitoringService(); err != nil {
		return err
	}

	// 3. 初始化补偿服务
	if err := osm.initCompensationService(); err != nil {
		return err
	}

	// 4. 启动后台任务
	osm.startBackgroundTasks()

	osm.isInitialized = true
	log.Printf("✅ 订单安全系统初始化完成")

	return nil
}

// initSecureOrderCreator 初始化安全订单创建器
func (osm *OrderSystemManager) initSecureOrderCreator() error {
	log.Printf("初始化安全订单创建器...")

	osm.secureCreator = NewSecureOrderCreator(osm.redisClient)

	// 设置全局实例
	InitGlobalSecureOrderCreator(osm.redisClient)

	// 初始化诊断工具
	InitGlobalDiagnostics(osm.redisClient)

	log.Printf("✅ 安全订单创建器初始化完成")
	return nil
}

// initMonitoringService 初始化监控服务
func (osm *OrderSystemManager) initMonitoringService() error {
	log.Printf("初始化订单监控服务...")

	osm.monitoringService = NewOrderMonitoringService(db.Dao, osm.redisClient)

	// 设置全局实例
	InitGlobalMonitoring()

	log.Printf("✅ 订单监控服务初始化完成")
	return nil
}

// initCompensationService 初始化补偿服务
func (osm *OrderSystemManager) initCompensationService() error {
	log.Printf("初始化订单补偿服务...")

	securityService := NewSecurityOrderService(osm.redisClient)
	osm.compensationSvc = securityService.NewOrderCompensationService(db.Dao)

	log.Printf("✅ 订单补偿服务初始化完成")
	return nil
}

// startBackgroundTasks 启动后台任务
func (osm *OrderSystemManager) startBackgroundTasks() {
	log.Printf("启动订单系统后台任务...")

	// 1. 启动过期订单检查任务
	go osm.startExpiredOrderChecker()

	// 2. 启动数据一致性检查任务
	go osm.startConsistencyChecker()

	// 3. 启动Redis超时队列处理器
	go osm.startTimeoutQueueProcessor()

	log.Printf("✅ 后台任务启动完成")
}

// startExpiredOrderChecker 启动过期订单检查器
func (osm *OrderSystemManager) startExpiredOrderChecker() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("过期订单检查器发生panic: %v", r)
			// 5分钟后重启
			time.Sleep(5 * time.Minute)
			go osm.startExpiredOrderChecker()
		}
	}()

	ticker := time.NewTicker(2 * time.Minute) // 每2分钟检查一次
	defer ticker.Stop()

	log.Printf("🔍 过期订单检查器已启动")

	for range ticker.C {
		osm.checkExpiredOrders()
	}
}

// checkExpiredOrders 检查并处理过期订单
func (osm *OrderSystemManager) checkExpiredOrders() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("检查过期订单时发生panic: %v", r)
		}
	}()

	// 查找15分钟前创建的pending订单
	expireTime := time.Now().Add(-15 * time.Minute)

	var expiredOrders []struct {
		No string `json:"no"`
	}

	err := db.Dao.Model(&app_model.AppOrder{}).
		Select("no").
		Where("status = ? AND create_time < ?", "pending", expireTime).
		Limit(50). // 每次最多处理50个
		Find(&expiredOrders).Error

	if err != nil {
		log.Printf("查询过期订单失败: %v", err)
		return
	}

	if len(expiredOrders) == 0 {
		return
	}

	log.Printf("发现 %d 个过期订单，开始处理...", len(expiredOrders))

	for _, order := range expiredOrders {
		go func(orderNo string) {
			if err := osm.secureCreator.CancelExpiredOrder(orderNo); err != nil {
				log.Printf("取消过期订单失败 %s: %v", orderNo, err)
			}
		}(order.No)
	}
}

// startConsistencyChecker 启动数据一致性检查器
func (osm *OrderSystemManager) startConsistencyChecker() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("数据一致性检查器发生panic: %v", r)
			// 10分钟后重启
			time.Sleep(10 * time.Minute)
			go osm.startConsistencyChecker()
		}
	}()

	// 每小时检查一次数据一致性
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	log.Printf("🔍 数据一致性检查器已启动")

	for range ticker.C {
		if osm.compensationSvc != nil {
			if err := osm.compensationSvc.DetectAndFixInconsistencies(); err != nil {
				log.Printf("数据一致性检查失败: %v", err)
			}
		}
	}
}

// startTimeoutQueueProcessor 启动Redis超时队列处理器
func (osm *OrderSystemManager) startTimeoutQueueProcessor() {
	if osm.redisClient == nil {
		log.Printf("Redis客户端未配置，跳过超时队列处理器")
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Redis超时队列处理器发生panic: %v", r)
			// 1分钟后重启
			time.Sleep(1 * time.Minute)
			go osm.startTimeoutQueueProcessor()
		}
	}()

	log.Printf("🔍 Redis超时队列处理器已启动")

	for {
		osm.processTimeoutQueue()
		time.Sleep(5 * time.Second) // 每5秒检查一次
	}
}

// processTimeoutQueue 处理Redis超时队列
func (osm *OrderSystemManager) processTimeoutQueue() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("处理Redis超时队列时发生panic: %v", r)
		}
	}()

	now := time.Now().Unix()

	// 获取已过期的订单
	ctx := context.Background()
	results, err := osm.redisClient.ZRangeByScore(
		ctx,
		"order_timeouts",
		&redis.ZRangeBy{
			Min:   "0",
			Max:   fmt.Sprintf("%d", now),
			Count: 10, // 每次处理10个
		},
	).Result()

	if err != nil {
		log.Printf("获取超时订单失败: %v", err)
		return
	}

	if len(results) == 0 {
		return // 没有超时订单
	}

	log.Printf("发现 %d 个超时订单需要处理", len(results))

	for _, orderNo := range results {
		// 先尝试从队列中原子性移除订单，如果移除失败说明已被其他进程处理
		removed, err := osm.redisClient.ZRem(ctx, "order_timeouts", orderNo).Result()
		if err != nil {
			log.Printf("从超时队列移除订单失败 %s: %v", orderNo, err)
			continue
		}

		// 如果返回0，说明该订单已被其他进程移除，跳过处理
		if removed == 0 {
			log.Printf("订单 %s 已被其他进程处理，跳过", orderNo)
			continue
		}

		// 异步处理订单取消
		go func(no string) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("处理订单取消时发生panic %s: %v", no, r)
				}
			}()

			if err := osm.secureCreator.CancelExpiredOrder(no); err != nil {
				log.Printf("Redis队列取消订单失败 %s: %v", no, err)
				// 如果取消失败且是锁相关错误，重新加入队列延后处理
				errMsg := err.Error()
				if strings.Contains(errMsg, "锁已被其他进程持有") || strings.Contains(errMsg, "获取取消锁失败") {
					// 延后5分钟重新尝试
					futureTime := time.Now().Add(5 * time.Minute).Unix()
					osm.redisClient.ZAdd(ctx, "order_timeouts", redis.Z{
						Score:  float64(futureTime),
						Member: no,
					})
					log.Printf("订单 %s 取消失败，5分钟后重试", no)
				}
			} else {
				log.Printf("Redis队列成功取消超时订单: %s", no)
			}
		}(orderNo)
	}
}

// GetSystemStatus 获取系统状态
func (osm *OrderSystemManager) GetSystemStatus() map[string]interface{} {
	status := map[string]interface{}{
		"initialized":   osm.isInitialized,
		"timestamp":     time.Now().Format("2006-01-02 15:04:05"),
		"redis_enabled": osm.redisClient != nil,
		"components":    make(map[string]interface{}),
	}

	components := status["components"].(map[string]interface{})

	// 安全创建器状态
	components["secure_creator"] = map[string]interface{}{
		"enabled": osm.secureCreator != nil,
	}

	// 监控服务状态
	if osm.monitoringService != nil {
		components["monitoring"] = map[string]interface{}{
			"enabled": true,
			"running": osm.monitoringService.isRunning,
		}

		// 获取监控统计
		if stats, err := osm.monitoringService.GetMonitoringStats(); err == nil {
			components["monitoring_stats"] = stats
		}
	} else {
		components["monitoring"] = map[string]interface{}{
			"enabled": false,
		}
	}

	// 补偿服务状态
	components["compensation"] = map[string]interface{}{
		"enabled": osm.compensationSvc != nil,
	}

	return status
}

// PerformHealthCheck 执行健康检查
func (osm *OrderSystemManager) PerformHealthCheck() map[string]interface{} {
	health := map[string]interface{}{
		"status":     "healthy",
		"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
		"components": make(map[string]interface{}),
	}

	components := health["components"].(map[string]interface{})

	// 检查初始化状态
	if !osm.isInitialized {
		health["status"] = "unhealthy"
		components["initialization"] = map[string]interface{}{
			"status": "not_initialized",
		}
		return health
	}

	components["initialization"] = map[string]interface{}{
		"status": "healthy",
	}

	// 检查数据库连接
	if db.Dao != nil {
		sqlDB, err := db.Dao.DB()
		if err != nil || sqlDB.Ping() != nil {
			health["status"] = "unhealthy"
			components["database"] = map[string]interface{}{
				"status": "unhealthy",
			}
		} else {
			components["database"] = map[string]interface{}{
				"status": "healthy",
			}
		}
	}

	// 检查Redis连接
	if osm.redisClient != nil {
		if err := osm.redisClient.Ping(context.Background()).Err(); err != nil {
			components["redis"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			// Redis故障不算系统完全不健康
			if health["status"] == "healthy" {
				health["status"] = "degraded"
			}
		} else {
			components["redis"] = map[string]interface{}{
				"status": "healthy",
			}
		}
	} else {
		components["redis"] = map[string]interface{}{
			"status": "not_configured",
		}
	}

	// 检查监控服务
	if osm.monitoringService != nil {
		components["monitoring"] = map[string]interface{}{
			"status":  "healthy",
			"running": osm.monitoringService.isRunning,
		}
	}

	return health
}

// Shutdown 优雅关闭系统
func (osm *OrderSystemManager) Shutdown() {
	log.Printf("🛑 开始关闭订单系统...")

	// 停止监控服务
	if osm.monitoringService != nil && osm.monitoringService.isRunning {
		osm.monitoringService.StopMonitoring()
	}

	// 等待后台任务完成
	time.Sleep(2 * time.Second)

	osm.isInitialized = false
	log.Printf("✅ 订单系统已关闭")
}

// CreateOrderWithSystem 使用系统创建订单（推荐使用）
func (osm *OrderSystemManager) CreateOrderWithSystem(c *gin.Context, uid int, params inout.CreateOrderReq) (string, error) {
	if !osm.isInitialized {
		return "", fmt.Errorf("订单系统未初始化")
	}

	if osm.secureCreator == nil {
		return "", fmt.Errorf("安全订单创建器未初始化")
	}

	return osm.secureCreator.CreateOrderSecurely(c, uid, params)
}

// GetGlobalOrderSystemManager 获取全局订单系统管理器
var globalOrderSystemManager *OrderSystemManager

// InitGlobalOrderSystem 初始化全局订单系统
func InitGlobalOrderSystem(redisClient *redis.Client) error {
	globalOrderSystemManager = NewOrderSystemManager(redisClient)
	return globalOrderSystemManager.Initialize()
}

// GetGlobalOrderSystem 获取全局订单系统管理器
func GetGlobalOrderSystem() *OrderSystemManager {
	return globalOrderSystemManager
}

// GetOrderSystemHealth 获取订单系统健康状态（用于健康检查接口）
func GetOrderSystemHealth() map[string]interface{} {
	if globalOrderSystemManager == nil {
		return map[string]interface{}{
			"status":    "unhealthy",
			"error":     "order system not initialized",
			"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		}
	}

	return globalOrderSystemManager.PerformHealthCheck()
}

// GetOrderSystemStatus 获取订单系统状态（用于状态查询接口）
func GetOrderSystemStatus() map[string]interface{} {
	if globalOrderSystemManager == nil {
		return map[string]interface{}{
			"initialized": false,
			"error":       "order system not initialized",
			"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		}
	}

	return globalOrderSystemManager.GetSystemStatus()
}
