package app_service

import (
	"log"
	"nasa-go-admin/db"
	"sync"
	"time"
)

// ServiceInitializer 服务初始化器
type ServiceInitializer struct {
	orderService  *OrderService
	statusManager *OrderStatusManager
	initialized   bool
	mu            sync.Mutex
}

var (
	globalInitializer *ServiceInitializer
	initOnce          sync.Once
)

// GetServiceInitializer 获取全局服务初始化器
func GetServiceInitializer() *ServiceInitializer {
	initOnce.Do(func() {
		globalInitializer = &ServiceInitializer{
			initialized:   false,
			statusManager: NewOrderStatusManager(),
		}
	})
	return globalInitializer
}

// InitializeOrderService 安全地初始化订单服务
func (si *ServiceInitializer) InitializeOrderService(redisClient interface{}) *OrderService {
	si.mu.Lock()
	defer si.mu.Unlock()

	if si.orderService != nil {
		return si.orderService
	}

	log.Printf("开始初始化订单服务...")

	// 创建基础订单服务（不依赖数据库连接池）
	service := &OrderService{
		redisClient:      nil, // 先设为 nil，后续设置
		dbPoolManager:    nil,
		queryOptimizer:   NewQueryOptimizer(),
		metricsCollector: NewMetricsCollector(),
	}

	// 启动指标收集器
	service.metricsCollector.Start()

	// 设置 Redis 客户端（如果提供）
	if redisClient != nil {
		// 这里需要根据实际的 Redis 客户端类型进行类型断言
		// service.redisClient = redisClient.(*redis.Client)
		log.Printf("Redis 客户端设置需要根据实际类型进行适配")
	}

	si.orderService = service

	// 异步初始化数据库相关组件
	go si.initializeDatabaseComponents(service)

	// 注意：订单取消工作器已迁移到 SecureOrderCreator 和 UnifiedOrderManager
	// 这里不再启动旧的工作器

	log.Printf("订单服务基础组件初始化完成")
	return service
}

// initializeDatabaseComponents 异步初始化数据库相关组件
func (si *ServiceInitializer) initializeDatabaseComponents(service *OrderService) {
	log.Printf("等待数据库连接建立...")

	// 等待数据库连接建立，最多等待60秒
	for i := 0; i < 60; i++ {
		if db.Dao != nil {
			log.Printf("数据库连接已建立，开始初始化连接池管理器...")

			// 初始化数据库连接池管理器
			dbPoolManager, err := NewDatabasePoolManager(db.Dao, nil)
			if err != nil {
				log.Printf("初始化数据库连接池管理器失败: %v", err)
			} else {
				service.dbPoolManager = dbPoolManager
				log.Printf("✅ 数据库连接池管理器初始化成功")
			}

			si.mu.Lock()
			si.initialized = true
			si.mu.Unlock()

			log.Printf("✅ 订单服务完全初始化完成")
			return
		}

		time.Sleep(1 * time.Second)
	}

	log.Printf("⚠️  警告: 数据库连接在60秒内未建立，订单服务将在降级模式下运行")
}

// IsFullyInitialized 检查服务是否完全初始化
func (si *ServiceInitializer) IsFullyInitialized() bool {
	si.mu.Lock()
	defer si.mu.Unlock()
	return si.initialized
}

// GetOrderService 获取订单服务实例
func (si *ServiceInitializer) GetOrderService() *OrderService {
	si.mu.Lock()
	defer si.mu.Unlock()
	return si.orderService
}

// GetOrderStatusManager 获取订单状态管理器实例
func (si *ServiceInitializer) GetOrderStatusManager() *OrderStatusManager {
	si.mu.Lock()
	defer si.mu.Unlock()
	return si.statusManager
}

// WaitForInitialization 等待服务完全初始化
func (si *ServiceInitializer) WaitForInitialization(timeout time.Duration) bool {
	start := time.Now()
	for time.Since(start) < timeout {
		if si.IsFullyInitialized() {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// GetInitializationStatus 获取初始化状态
func (si *ServiceInitializer) GetInitializationStatus() map[string]interface{} {
	si.mu.Lock()
	defer si.mu.Unlock()

	status := map[string]interface{}{
		"service_created":   si.orderService != nil,
		"fully_initialized": si.initialized,
		"timestamp":         time.Now(),
	}

	if si.orderService != nil {
		status["has_db_pool_manager"] = si.orderService.dbPoolManager != nil
		status["has_query_optimizer"] = si.orderService.queryOptimizer != nil
		status["has_metrics_collector"] = si.orderService.metricsCollector != nil
	}

	return status
}
