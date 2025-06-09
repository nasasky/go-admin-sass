package app_service

import (
	"context"
	"fmt"
	"log"
	"nasa-go-admin/db"
	"nasa-go-admin/model/app_model"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// OrderMonitoringService 订单监控服务
type OrderMonitoringService struct {
	db          *gorm.DB
	redisClient *redis.Client
	alerter     AlertService
	isRunning   bool
	stopCh      chan struct{}
}

// AlertService 告警服务接口
type AlertService interface {
	SendAlert(title, message string) error
	SendUrgentAlert(title, message string) error
}

// SimpleAlertService 简单告警服务实现
type SimpleAlertService struct{}

func (s *SimpleAlertService) SendAlert(title, message string) error {
	log.Printf("🚨 [ALERT] %s: %s", title, message)
	return nil
}

func (s *SimpleAlertService) SendUrgentAlert(title, message string) error {
	log.Printf("🚨🚨 [URGENT ALERT] %s: %s", title, message)
	return nil
}

// NewOrderMonitoringService 创建订单监控服务
func NewOrderMonitoringService(db *gorm.DB, redisClient *redis.Client) *OrderMonitoringService {
	return &OrderMonitoringService{
		db:          db,
		redisClient: redisClient,
		alerter:     &SimpleAlertService{},
		stopCh:      make(chan struct{}),
	}
}

// StartMonitoring 开始监控
func (oms *OrderMonitoringService) StartMonitoring() {
	if oms.isRunning {
		log.Printf("订单监控服务已在运行中")
		return
	}

	oms.isRunning = true
	log.Printf("🔍 启动订单监控服务...")

	// 立即执行一次检查
	oms.runAllChecks()

	// 定时检查
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			oms.runAllChecks()
		case <-oms.stopCh:
			log.Printf("订单监控服务已停止")
			return
		}
	}
}

// StopMonitoring 停止监控
func (oms *OrderMonitoringService) StopMonitoring() {
	if !oms.isRunning {
		return
	}

	close(oms.stopCh)
	oms.isRunning = false
	log.Printf("正在停止订单监控服务...")
}

// runAllChecks 运行所有检查
func (oms *OrderMonitoringService) runAllChecks() {
	// 检查挂起订单
	if err := oms.checkPendingOrders(); err != nil {
		log.Printf("检查挂起订单失败: %v", err)
	}

	// 检查异常支付
	if err := oms.checkAbnormalPayments(); err != nil {
		log.Printf("检查异常支付失败: %v", err)
	}

	// 检查库存异常
	if err := oms.checkStockAnomalies(); err != nil {
		log.Printf("检查库存异常失败: %v", err)
	}

	// 检查系统性能
	if err := oms.checkSystemPerformance(); err != nil {
		log.Printf("检查系统性能失败: %v", err)
	}

	// 检查数据一致性
	if err := oms.checkDataConsistency(); err != nil {
		log.Printf("检查数据一致性失败: %v", err)
	}
}

// checkPendingOrders 检查长时间挂起的订单
func (oms *OrderMonitoringService) checkPendingOrders() error {
	// 定义不同级别的阈值
	warningThreshold := time.Now().Add(-30 * time.Minute) // 30分钟
	urgentThreshold := time.Now().Add(-2 * time.Hour)     // 2小时

	// 检查30分钟以上的挂起订单
	var longPendingOrders []app_model.AppOrder
	err := oms.db.Where("status = ? AND create_time < ?", "pending", warningThreshold).
		Find(&longPendingOrders).Error

	if err != nil {
		return fmt.Errorf("查询挂起订单失败: %w", err)
	}

	if len(longPendingOrders) > 0 {
		// 检查超级长时间挂起的订单
		var urgentOrders []app_model.AppOrder
		err := oms.db.Where("status = ? AND create_time < ?", "pending", urgentThreshold).
			Find(&urgentOrders).Error

		if err == nil && len(urgentOrders) > 0 {
			oms.alerter.SendUrgentAlert("严重订单积压告警",
				fmt.Sprintf("发现 %d 个超过2小时的挂起订单，需要立即处理", len(urgentOrders)))
		}

		if len(longPendingOrders) > 10 {
			oms.alerter.SendAlert("订单积压告警",
				fmt.Sprintf("发现 %d 个长时间挂起的订单", len(longPendingOrders)))
		}

		log.Printf("📊 监控统计: 挂起订单总数 %d，其中超过2小时的 %d 个",
			len(longPendingOrders), len(urgentOrders))
	}

	return nil
}

// checkAbnormalPayments 检查异常支付模式
func (oms *OrderMonitoringService) checkAbnormalPayments() error {
	// 检查1小时内的异常支付
	var recentPayments []struct {
		UserID       int     `json:"user_id"`
		PaymentCount int     `json:"payment_count"`
		TotalAmount  float64 `json:"total_amount"`
	}

	query := `
		SELECT user_id, COUNT(*) as payment_count, SUM(amount) as total_amount
		FROM app_recharge 
		WHERE create_time > ? AND transaction_type = 'order_payment'
		GROUP BY user_id
		HAVING COUNT(*) > 20 OR SUM(amount) > 5000
		ORDER BY payment_count DESC, total_amount DESC
		LIMIT 20
	`

	err := oms.db.Raw(query, time.Now().Add(-1*time.Hour)).Scan(&recentPayments).Error
	if err != nil {
		return fmt.Errorf("查询异常支付失败: %w", err)
	}

	for _, payment := range recentPayments {
		if payment.PaymentCount > 50 || payment.TotalAmount > 10000 {
			oms.alerter.SendUrgentAlert("严重异常支付告警",
				fmt.Sprintf("用户 %d 在1小时内支付 %d 次，总金额 %.2f，疑似异常",
					payment.UserID, payment.PaymentCount, payment.TotalAmount))
		} else {
			oms.alerter.SendAlert("异常支付告警",
				fmt.Sprintf("用户 %d 在1小时内支付 %d 次，总金额 %.2f",
					payment.UserID, payment.PaymentCount, payment.TotalAmount))
		}
	}

	return nil
}

// checkStockAnomalies 检查库存异常
func (oms *OrderMonitoringService) checkStockAnomalies() error {
	// 检查库存为负数的商品
	var negativeStockGoods []app_model.AppGoods
	err := oms.db.Where("stock < 0 AND isdelete != 1").Find(&negativeStockGoods).Error
	if err != nil {
		return fmt.Errorf("查询负库存商品失败: %w", err)
	}

	if len(negativeStockGoods) > 0 {
		for _, goods := range negativeStockGoods {
			oms.alerter.SendUrgentAlert("库存异常告警",
				fmt.Sprintf("商品 %d (%s) 库存为负数: %d",
					goods.Id, goods.GoodsName, goods.Stock))
		}
	}

	// 检查库存预警
	var lowStockGoods []app_model.AppGoods
	err = oms.db.Where("stock > 0 AND stock < 10 AND status = '1' AND isdelete != 1").
		Find(&lowStockGoods).Error
	if err != nil {
		return fmt.Errorf("查询低库存商品失败: %w", err)
	}

	if len(lowStockGoods) > 10 {
		oms.alerter.SendAlert("库存预警",
			fmt.Sprintf("发现 %d 个商品库存不足10件", len(lowStockGoods)))
	}

	return nil
}

// checkSystemPerformance 检查系统性能
func (oms *OrderMonitoringService) checkSystemPerformance() error {
	// 检查最近1小时的订单处理速度
	var orderStats struct {
		TotalOrders int     `json:"total_orders"`
		AvgDuration float64 `json:"avg_duration"`
	}

	query := `
		SELECT 
			COUNT(*) as total_orders,
			AVG(TIMESTAMPDIFF(SECOND, create_time, update_time)) as avg_duration
		FROM app_order 
		WHERE create_time > ? AND status != 'pending'
	`

	err := oms.db.Raw(query, time.Now().Add(-1*time.Hour)).Scan(&orderStats).Error
	if err != nil {
		return fmt.Errorf("查询订单性能统计失败: %w", err)
	}

	// 如果平均处理时间超过5分钟，发出告警
	if orderStats.AvgDuration > 300 {
		oms.alerter.SendAlert("系统性能告警",
			fmt.Sprintf("最近1小时订单平均处理时间 %.1f 秒，超过正常阈值", orderStats.AvgDuration))
	}

	// 检查Redis连接状态
	if oms.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := oms.redisClient.Ping(ctx).Err(); err != nil {
			oms.alerter.SendUrgentAlert("Redis连接异常",
				fmt.Sprintf("Redis ping失败: %v", err))
		}
	}

	return nil
}

// checkDataConsistency 检查数据一致性 - 修复版本
func (oms *OrderMonitoringService) checkDataConsistency() error {
	// 检查订单和支付记录的一致性
	var inconsistentData []struct {
		Issue       string `json:"issue"`
		Count       int    `json:"count"`
		Description string `json:"description"`
	}

	// 查找有支付记录但没有订单的情况 - 修复查询逻辑
	query1 := `
		SELECT 
			'orphaned_payments' as issue,
			COUNT(*) as count,
			'支付记录无对应订单' as description
		FROM app_recharge ar 
		WHERE ar.transaction_type = 'order_payment' 
		AND ar.create_time > ? 
		AND ar.order_no IS NOT NULL
		AND ar.order_no != ''
		AND NOT EXISTS (
			SELECT 1 FROM app_order ao 
			WHERE ao.no = ar.order_no
			AND ao.user_id = ar.user_id 
			AND ABS(ao.amount - ar.amount) < 0.01
		)
	`

	var orphanedPayments struct {
		Issue       string `json:"issue"`
		Count       int    `json:"count"`
		Description string `json:"description"`
	}

	err := oms.db.Raw(query1, time.Now().Add(-1*time.Hour)).Scan(&orphanedPayments).Error
	if err != nil {
		return fmt.Errorf("检查孤立支付记录失败: %w", err)
	}

	if orphanedPayments.Count > 0 {
		inconsistentData = append(inconsistentData, orphanedPayments)
	}

	// 查找有订单但状态异常的情况 - 优化查询条件
	query2 := `
		SELECT 
			'status_mismatch' as issue,
			COUNT(*) as count,
			'订单状态异常' as description
		FROM app_order ao 
		WHERE ao.create_time > ?
		AND ao.status = 'paid'
		AND NOT EXISTS (
			SELECT 1 FROM app_recharge ar 
			WHERE (
				(ar.order_no = ao.no AND ar.user_id = ao.user_id) 
				OR 
				(ar.user_id = ao.user_id AND ABS(ar.amount - ao.amount) < 0.01
				 AND ar.create_time BETWEEN ao.create_time - INTERVAL 10 MINUTE 
			                           AND ao.create_time + INTERVAL 10 MINUTE)
			)
			AND ar.transaction_type = 'order_payment'
			AND ar.status = 'completed'
		)
	`

	var statusMismatch struct {
		Issue       string `json:"issue"`
		Count       int    `json:"count"`
		Description string `json:"description"`
	}

	err = oms.db.Raw(query2, time.Now().Add(-1*time.Hour)).Scan(&statusMismatch).Error
	if err != nil {
		return fmt.Errorf("检查状态不一致失败: %w", err)
	}

	if statusMismatch.Count > 0 {
		inconsistentData = append(inconsistentData, statusMismatch)
	}

	// 发送一致性告警 - 只有数量大于阈值才告警
	for _, data := range inconsistentData {
		if data.Count > 2 { // 设置阈值为2，减少误报
			oms.alerter.SendAlert("数据一致性告警",
				fmt.Sprintf("%s: 发现 %d 条异常记录", data.Description, data.Count))
			log.Printf("🚨 [ALERT] %s: 发现 %d 条异常记录", data.Description, data.Count)
		} else if data.Count > 0 {
			log.Printf("⚠️ [INFO] %s: 发现 %d 条记录，数量较少，可能为正常情况", data.Description, data.Count)
		}
	}

	return nil
}

// GetMonitoringStats 获取监控统计数据
func (oms *OrderMonitoringService) GetMonitoringStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 订单统计
	var orderStats struct {
		TotalOrders     int `json:"total_orders"`
		PendingOrders   int `json:"pending_orders"`
		PaidOrders      int `json:"paid_orders"`
		CancelledOrders int `json:"cancelled_orders"`
	}

	// 最近24小时的订单统计
	err := oms.db.Model(&app_model.AppOrder{}).
		Select(`
			COUNT(*) as total_orders,
			SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending_orders,
			SUM(CASE WHEN status = 'paid' THEN 1 ELSE 0 END) as paid_orders,
			SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END) as cancelled_orders
		`).
		Where("create_time > ?", time.Now().Add(-24*time.Hour)).
		Scan(&orderStats).Error

	if err != nil {
		return nil, fmt.Errorf("获取订单统计失败: %w", err)
	}

	stats["orders"] = orderStats

	// 支付统计
	var paymentStats struct {
		TotalPayments int     `json:"total_payments"`
		TotalAmount   float64 `json:"total_amount"`
	}

	err = oms.db.Model(&app_model.AppRecharge{}).
		Select("COUNT(*) as total_payments, SUM(amount) as total_amount").
		Where("create_time > ? AND transaction_type = 'order_payment'", time.Now().Add(-24*time.Hour)).
		Scan(&paymentStats).Error

	if err != nil {
		return nil, fmt.Errorf("获取支付统计失败: %w", err)
	}

	stats["payments"] = paymentStats

	// 系统状态
	systemStatus := map[string]interface{}{
		"monitoring_running": oms.isRunning,
		"database_connected": oms.db != nil,
		"redis_connected":    oms.redisClient != nil,
		"last_check_time":    time.Now().Format("2006-01-02 15:04:05"),
	}

	// 检查Redis连接
	if oms.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		systemStatus["redis_ping_ok"] = oms.redisClient.Ping(ctx).Err() == nil
	}

	stats["system"] = systemStatus

	return stats, nil
}

// PerformHealthCheck 执行健康检查
func (oms *OrderMonitoringService) PerformHealthCheck() map[string]interface{} {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"checks":    make(map[string]interface{}),
	}

	checks := health["checks"].(map[string]interface{})

	// 检查数据库连接
	if oms.db != nil {
		sqlDB, err := oms.db.DB()
		if err != nil {
			checks["database"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			health["status"] = "unhealthy"
		} else {
			err = sqlDB.Ping()
			if err != nil {
				checks["database"] = map[string]interface{}{
					"status": "unhealthy",
					"error":  err.Error(),
				}
				health["status"] = "unhealthy"
			} else {
				checks["database"] = map[string]interface{}{
					"status": "healthy",
				}
			}
		}
	}

	// 检查Redis连接
	if oms.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := oms.redisClient.Ping(ctx).Err()
		if err != nil {
			checks["redis"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			health["status"] = "degraded" // Redis故障不算完全不健康
		} else {
			checks["redis"] = map[string]interface{}{
				"status": "healthy",
			}
		}
	} else {
		checks["redis"] = map[string]interface{}{
			"status": "not_configured",
		}
	}

	// 检查监控服务状态
	checks["monitoring"] = map[string]interface{}{
		"status":  "healthy",
		"running": oms.isRunning,
	}

	return health
}

// 全局监控服务实例
var globalMonitoringService *OrderMonitoringService

// InitGlobalMonitoring 初始化全局监控服务
func InitGlobalMonitoring() {
	globalMonitoringService = NewOrderMonitoringService(db.Dao, nil)

	// 在后台启动监控
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("监控服务发生panic: %v", r)
				// 可以在这里添加重启逻辑
			}
		}()

		globalMonitoringService.StartMonitoring()
	}()

	log.Printf("✅ 全局订单监控服务已启动")
}

// GetGlobalMonitoring 获取全局监控服务
func GetGlobalMonitoring() *OrderMonitoringService {
	return globalMonitoringService
}
