package app_service

import (
	"context"
	"fmt"
	"log"
	"nasa-go-admin/db"
	"nasa-go-admin/model/app_model"
	"sync"
	"time"

	"gorm.io/gorm"
)

// OrderStatus 订单状态常量
type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"   // 待支付
	StatusPaid      OrderStatus = "paid"      // 已支付
	StatusShipped   OrderStatus = "shipped"   // 已发货
	StatusDelivered OrderStatus = "delivered" // 已送达
	StatusCompleted OrderStatus = "completed" // 已完成
	StatusCancelled OrderStatus = "cancelled" // 已取消
	StatusRefunded  OrderStatus = "refunded"  // 已退款
)

// OrderStatusTransition 订单状态转换规则
type OrderStatusTransition struct {
	From        OrderStatus `json:"from"`
	To          OrderStatus `json:"to"`
	AllowedBy   []string    `json:"allowed_by"` // 允许执行转换的角色
	Description string      `json:"description"`
	Reversible  bool        `json:"reversible"` // 是否可逆转
}

// OrderStatusManager 订单状态管理器
type OrderStatusManager struct {
	transitions map[OrderStatus][]OrderStatus
	rules       []OrderStatusTransition
	mutex       sync.RWMutex
}

// NewOrderStatusManager 创建订单状态管理器
func NewOrderStatusManager() *OrderStatusManager {
	manager := &OrderStatusManager{
		transitions: make(map[OrderStatus][]OrderStatus),
		rules:       make([]OrderStatusTransition, 0),
	}

	// 初始化状态转换规则
	manager.initializeTransitionRules()

	return manager
}

// initializeTransitionRules 初始化状态转换规则
func (osm *OrderStatusManager) initializeTransitionRules() {
	rules := []OrderStatusTransition{
		// 从待支付状态的转换
		{StatusPending, StatusPaid, []string{"user", "system"}, "用户支付或系统自动支付", false},
		{StatusPending, StatusCancelled, []string{"user", "system", "admin"}, "用户取消、系统超时取消或管理员取消", false},

		// 从已支付状态的转换
		{StatusPaid, StatusShipped, []string{"merchant", "admin"}, "商家发货", false},
		{StatusPaid, StatusRefunded, []string{"admin", "system"}, "管理员或系统退款", false},
		{StatusPaid, StatusCancelled, []string{"admin"}, "管理员强制取消（异常情况）", false},

		// 从已发货状态的转换
		{StatusShipped, StatusDelivered, []string{"system", "courier", "admin"}, "快递员确认送达或系统自动确认", false},
		{StatusShipped, StatusRefunded, []string{"admin"}, "退货退款", false},

		// 从已送达状态的转换
		{StatusDelivered, StatusCompleted, []string{"user", "system"}, "用户确认收货或系统自动确认", false},
		{StatusDelivered, StatusRefunded, []string{"admin", "user"}, "确认收货后申请退款", false},

		// 从已完成状态的转换
		{StatusCompleted, StatusRefunded, []string{"admin"}, "特殊情况下的退款", false},
	}

	osm.rules = rules

	// 构建转换映射表
	for _, rule := range rules {
		if osm.transitions[rule.From] == nil {
			osm.transitions[rule.From] = make([]OrderStatus, 0)
		}
		osm.transitions[rule.From] = append(osm.transitions[rule.From], rule.To)
	}

	log.Printf("订单状态管理器初始化完成，共加载 %d 条转换规则", len(rules))
}

// ValidateTransition 验证状态转换是否合法
func (osm *OrderStatusManager) ValidateTransition(from, to OrderStatus, operator string) error {
	osm.mutex.RLock()
	defer osm.mutex.RUnlock()

	// 检查目标状态是否在允许的转换列表中
	allowedTransitions := osm.transitions[from]
	isAllowed := false
	for _, allowedTo := range allowedTransitions {
		if allowedTo == to {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return fmt.Errorf("不允许的状态转换: %s -> %s", from, to)
	}

	// 检查操作者是否有权限执行此转换
	for _, rule := range osm.rules {
		if rule.From == from && rule.To == to {
			hasPermission := false
			for _, allowedRole := range rule.AllowedBy {
				if allowedRole == operator {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				return fmt.Errorf("操作者 %s 无权限执行状态转换: %s -> %s", operator, from, to)
			}

			return nil
		}
	}

	return fmt.Errorf("未找到状态转换规则: %s -> %s", from, to)
}

// UpdateOrderStatus 安全地更新订单状态
func (osm *OrderStatusManager) UpdateOrderStatus(orderNo string, newStatus OrderStatus, operator string, reason string) error {
	// 开始事务
	tx := db.Dao.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("更新订单状态时发生panic: %v", r)
		}
	}()

	// 使用分布式锁防止并发更新
	lockKey := fmt.Sprintf("order_status_update:%s", orderNo)
	orderService := GetServiceInitializer().GetOrderService()
	if orderService != nil {
		if acquired, err := orderService.acquireLock(lockKey, 10*time.Second); err != nil || !acquired {
			return fmt.Errorf("获取订单状态更新锁失败，请稍后重试")
		}
		defer orderService.releaseLock(lockKey)
	}

	// 查询当前订单状态
	var order app_model.AppOrder
	err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("no = ?", orderNo).
		First(&order).Error

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("查询订单失败: %w", err)
	}

	currentStatus := OrderStatus(order.Status)

	// 如果状态相同，无需更新
	if currentStatus == newStatus {
		tx.Rollback()
		return nil
	}

	// 验证状态转换是否合法
	if err := osm.ValidateTransition(currentStatus, newStatus, operator); err != nil {
		tx.Rollback()
		return err
	}

	// 记录状态变更历史
	statusHistory := app_model.OrderStatusHistory{
		OrderId:    order.Id,
		OrderNo:    orderNo,
		FromStatus: string(currentStatus),
		ToStatus:   string(newStatus),
		Operator:   operator,
		Reason:     reason,
		CreateTime: time.Now(),
	}

	if err := tx.Create(&statusHistory).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("记录状态变更历史失败: %w", err)
	}

	// 更新订单状态
	err = tx.Model(&order).Updates(map[string]interface{}{
		"status":      string(newStatus),
		"update_time": time.Now(),
	}).Error

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("更新订单状态失败: %w", err)
	}

	// 根据新状态执行相应的业务逻辑
	if err := osm.handleStatusChange(tx, &order, currentStatus, newStatus, operator); err != nil {
		tx.Rollback()
		return fmt.Errorf("处理状态变更业务逻辑失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交状态更新事务失败: %w", err)
	}

	log.Printf("✅ 订单 %s 状态已更新: %s -> %s (操作者: %s)", orderNo, currentStatus, newStatus, operator)

	// 异步发送状态变更通知
	go osm.sendStatusChangeNotification(order, currentStatus, newStatus, operator)

	return nil
}

// handleStatusChange 处理状态变更的业务逻辑
func (osm *OrderStatusManager) handleStatusChange(tx *gorm.DB, order *app_model.AppOrder, from, to OrderStatus, operator string) error {
	switch to {
	case StatusPaid:
		// 支付完成后的处理
		return osm.handlePaymentCompleted(tx, order)

	case StatusCancelled:
		// 订单取消后的处理
		return osm.handleOrderCancelled(tx, order, from)

	case StatusRefunded:
		// 退款处理
		return osm.handleOrderRefunded(tx, order)

	case StatusCompleted:
		// 订单完成处理
		return osm.handleOrderCompleted(tx, order)

	default:
		// 其他状态暂无特殊处理
		return nil
	}
}

// handlePaymentCompleted 处理支付完成
func (osm *OrderStatusManager) handlePaymentCompleted(tx *gorm.DB, order *app_model.AppOrder) error {
	log.Printf("订单 %s 支付完成，执行后续处理", order.No)
	// 这里可以添加支付完成后的业务逻辑
	// 例如：库存扣减、积分奖励、优惠券使用等
	return nil
}

// handleOrderCancelled 处理订单取消
func (osm *OrderStatusManager) handleOrderCancelled(tx *gorm.DB, order *app_model.AppOrder, fromStatus OrderStatus) error {
	log.Printf("订单 %s 已取消，执行库存恢复等处理", order.No)

	// 如果从已支付状态取消，需要退款
	if fromStatus == StatusPaid {
		// 执行退款逻辑
		log.Printf("订单 %s 从已支付状态取消，需要退款", order.No)
	}

	// 恢复商品库存
	err := tx.Model(&app_model.AppGoods{}).
		Where("id = ?", order.GoodsId).
		Update("stock", gorm.Expr("stock + ?", order.Num)).Error

	if err != nil {
		return fmt.Errorf("恢复商品库存失败: %w", err)
	}

	log.Printf("✅ 已恢复商品 %d 库存 %d 件", order.GoodsId, order.Num)
	return nil
}

// handleOrderRefunded 处理订单退款
func (osm *OrderStatusManager) handleOrderRefunded(tx *gorm.DB, order *app_model.AppOrder) error {
	log.Printf("订单 %s 已退款，执行相关处理", order.No)

	// 恢复商品库存
	err := tx.Model(&app_model.AppGoods{}).
		Where("id = ?", order.GoodsId).
		Update("stock", gorm.Expr("stock + ?", order.Num)).Error

	if err != nil {
		return fmt.Errorf("恢复商品库存失败: %w", err)
	}

	// 这里可以添加退款到用户钱包的逻辑
	log.Printf("✅ 订单 %s 退款处理完成", order.No)
	return nil
}

// handleOrderCompleted 处理订单完成
func (osm *OrderStatusManager) handleOrderCompleted(tx *gorm.DB, order *app_model.AppOrder) error {
	log.Printf("订单 %s 已完成，执行完成后处理", order.No)
	// 这里可以添加订单完成后的业务逻辑
	// 例如：积分奖励、评价提醒、推荐商品等
	return nil
}

// sendStatusChangeNotification 发送状态变更通知
func (osm *OrderStatusManager) sendStatusChangeNotification(order app_model.AppOrder, from, to OrderStatus, operator string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("发送状态变更通知时发生panic: %v", r)
		}
	}()

	// 这里可以集成各种通知方式：短信、邮件、推送等
	log.Printf("📱 发送状态变更通知 - 订单: %s, 状态: %s -> %s", order.No, from, to)

	// 示例：发送WebSocket通知
	// 实际项目中应该根据具体需求实现
}

// GetOrderStatusHistory 获取订单状态变更历史
func (osm *OrderStatusManager) GetOrderStatusHistory(orderNo string) ([]app_model.OrderStatusHistory, error) {
	var history []app_model.OrderStatusHistory
	err := db.Dao.Where("order_no = ?", orderNo).
		Order("create_time ASC").
		Find(&history).Error

	if err != nil {
		return nil, fmt.Errorf("查询订单状态历史失败: %w", err)
	}

	return history, nil
}

// GetAllowedTransitions 获取指定状态的允许转换
func (osm *OrderStatusManager) GetAllowedTransitions(currentStatus OrderStatus, operator string) []OrderStatus {
	osm.mutex.RLock()
	defer osm.mutex.RUnlock()

	allowedTransitions := make([]OrderStatus, 0)

	for _, rule := range osm.rules {
		if rule.From == currentStatus {
			// 检查操作者权限
			hasPermission := false
			for _, allowedRole := range rule.AllowedBy {
				if allowedRole == operator {
					hasPermission = true
					break
				}
			}

			if hasPermission {
				allowedTransitions = append(allowedTransitions, rule.To)
			}
		}
	}

	return allowedTransitions
}

// BatchUpdateOrderStatus 批量更新订单状态
func (osm *OrderStatusManager) BatchUpdateOrderStatus(orderNos []string, newStatus OrderStatus, operator string, reason string) error {
	if len(orderNos) == 0 {
		return fmt.Errorf("订单号列表不能为空")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	errors := make([]error, 0)
	successCount := 0

	for _, orderNo := range orderNos {
		select {
		case <-ctx.Done():
			return fmt.Errorf("批量更新超时，已处理 %d/%d 个订单", successCount, len(orderNos))
		default:
			if err := osm.UpdateOrderStatus(orderNo, newStatus, operator, reason); err != nil {
				errors = append(errors, fmt.Errorf("订单 %s: %w", orderNo, err))
				log.Printf("批量更新订单状态失败 - 订单: %s, 错误: %v", orderNo, err)
			} else {
				successCount++
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("批量更新完成，成功: %d, 失败: %d, 错误详情: %v", successCount, len(errors), errors)
	}

	log.Printf("✅ 批量更新订单状态完成，共处理 %d 个订单", successCount)
	return nil
}
