package app_service

import (
	"log"
	"nasa-go-admin/inout"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// UnifiedOrderManager 统一订单管理器
// 整合所有订单相关功能，提供统一的接口
type UnifiedOrderManager struct {
	secureCreator *SecureOrderCreator
	redisClient   *redis.Client
}

// NewUnifiedOrderManager 创建统一订单管理器
func NewUnifiedOrderManager(redisClient *redis.Client) *UnifiedOrderManager {
	return &UnifiedOrderManager{
		secureCreator: NewSecureOrderCreator(redisClient),
		redisClient:   redisClient,
	}
}

// 订单创建相关方法

// CreateOrder 创建订单 - 统一入口
func (uom *UnifiedOrderManager) CreateOrder(c *gin.Context, uid int, params inout.CreateOrderReq) (string, error) {
	return uom.secureCreator.CreateOrderSecurely(c, uid, params)
}

// 订单查询相关方法

// GetOrderDetail 获取订单详情
func (uom *UnifiedOrderManager) GetOrderDetail(c *gin.Context, uid int, id int) (interface{}, error) {
	return uom.secureCreator.GetOrderDetail(c, uid, id)
}

// GetMyOrderList 获取我的订单列表
func (uom *UnifiedOrderManager) GetMyOrderList(c *gin.Context, uid int, params inout.MyOrderReq) (interface{}, error) {
	return uom.secureCreator.GetMyOrderList(c, uid, params)
}

// 订单状态管理相关方法

// CancelOrder 取消订单
func (uom *UnifiedOrderManager) CancelOrder(orderNo string) error {
	return uom.secureCreator.CancelExpiredOrder(orderNo)
}

// ProcessPayment 处理支付
func (uom *UnifiedOrderManager) ProcessPayment(orderNo string, amount float64) error {
	return uom.secureCreator.ProcessPayment(orderNo, amount)
}

// GetOrderStatus 获取订单状态
func (uom *UnifiedOrderManager) GetOrderStatus(orderNo string) (string, error) {
	return uom.secureCreator.GetOrderStatus(orderNo)
}

// 系统健康检查

// GetHealthStatus 获取订单系统健康状态
func (uom *UnifiedOrderManager) GetHealthStatus() map[string]interface{} {
	status := map[string]interface{}{
		"service":        "unified_order_manager",
		"secure_creator": uom.secureCreator != nil,
		"redis_client":   uom.redisClient != nil,
		"status":         "healthy",
	}

	// 可以添加更多健康检查逻辑
	return status
}

// 全局统一订单管理器实例
var globalUnifiedOrderManager *UnifiedOrderManager

// InitGlobalUnifiedOrderManager 初始化全局统一订单管理器
func InitGlobalUnifiedOrderManager(redisClient *redis.Client) {
	globalUnifiedOrderManager = NewUnifiedOrderManager(redisClient)
	// 同时初始化全局安全订单创建器
	InitGlobalSecureOrderCreator(redisClient)
	log.Printf("✅ 全局统一订单管理器已初始化")
}

// GetGlobalUnifiedOrderManager 获取全局统一订单管理器
func GetGlobalUnifiedOrderManager() *UnifiedOrderManager {
	return globalUnifiedOrderManager
}

// 向后兼容的接口

// GetLegacyOrderService 获取遗留订单服务（仅用于向后兼容）
// 推荐使用 UnifiedOrderManager 替代
func GetLegacyOrderService(redisClient *redis.Client) *OrderService {
	log.Printf("⚠️  警告: 正在使用已废弃的OrderService，建议迁移到UnifiedOrderManager")
	return NewOrderService(redisClient)
}

// 迁移帮助函数

// MigrateToUnifiedManager 迁移助手，帮助现有代码迁移到统一管理器
func MigrateToUnifiedManager() {
	log.Printf(`
	🔄 订单服务迁移指南:
	
	旧方式:
	orderService := app_service.NewOrderService(redis.GetClient())
	orderService.CreateOrder(c, uid, params)
	
	新方式:
	unifiedManager := app_service.NewUnifiedOrderManager(redis.GetClient())
	unifiedManager.CreateOrder(c, uid, params)
	
	或使用全局实例:
	app_service.InitGlobalUnifiedOrderManager(redis.GetClient())
	manager := app_service.GetGlobalUnifiedOrderManager()
	manager.CreateOrder(c, uid, params)
	`)
}
