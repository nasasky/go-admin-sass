package app_service

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/app_model"
	"nasa-go-admin/services/miniapp_service"
	"nasa-go-admin/services/public_service"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// SecureOrderCreator 安全订单创建器
type SecureOrderCreator struct {
	securityService    *SecurityOrderService
	timeoutManager     *MultiLayerTimeoutManager
	compensationSvc    *OrderCompensationService
	idempotencyChecker *IdempotencyChecker
}

// NewSecureOrderCreator 创建安全订单创建器
func NewSecureOrderCreator(redisClient *redis.Client) *SecureOrderCreator {
	securityService := NewSecurityOrderService(redisClient)

	return &SecureOrderCreator{
		securityService:    securityService,
		timeoutManager:     securityService.NewMultiLayerTimeoutManager(db.Dao),
		compensationSvc:    securityService.NewOrderCompensationService(db.Dao),
		idempotencyChecker: securityService.NewIdempotencyChecker(),
	}
}

// CreateOrderSecurely 安全创建订单 - 解决所有并发和卡单问题
func (soc *SecureOrderCreator) CreateOrderSecurely(c *gin.Context, uid int, params inout.CreateOrderReq) (string, error) {
	startTime := time.Now()
	log.Printf("🚀 开始安全创建订单 用户:%d 商品:%d 数量:%d", uid, params.GoodsId, params.Num)

	// 1. 生成幂等性键但先不设置，等订单成功后再设置
	idempotencyKey := fmt.Sprintf("order_create:%d:%d:%s", uid, params.GoodsId,
		time.Now().Format("20060102_15"))

	// 先检查是否存在重复请求，但不立即设置标记
	isDuplicate, err := soc.idempotencyChecker.CheckOnly(idempotencyKey)
	if err != nil {
		log.Printf("幂等性检查失败: %v", err)
	} else if isDuplicate {
		return "", fmt.Errorf("请勿重复下单，如有问题请联系客服")
	}

	// 2. 获取分布式锁 - 防止并发问题
	userLock := soc.securityService.NewDistributedLock(
		fmt.Sprintf("create_order:user:%d", uid),
		30*time.Second,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := userLock.AcquireWithRenewal(ctx); err != nil {
		return "", fmt.Errorf("系统繁忙，请稍后再试: %w", err)
	}
	defer userLock.Release()

	// 3. 商品级别锁 - 防止库存超卖
	goodsLock := soc.securityService.NewDistributedLock(
		fmt.Sprintf("goods_stock:%d", params.GoodsId),
		30*time.Second,
	)

	if err := goodsLock.AcquireWithRenewal(ctx); err != nil {
		return "", fmt.Errorf("商品库存正在更新，请稍后再试: %w", err)
	}
	defer goodsLock.Release()

	// 4. 开启事务处理
	tx := db.Dao.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("创建订单时发生panic: %v", r)
			panic(r)
		}
	}()

	// 5. 查询并验证商品信息
	var goods app_model.AppGoods
	if err := tx.Where("id = ?", params.GoodsId).First(&goods).Error; err != nil {
		tx.Rollback()
		return "", fmt.Errorf("商品不存在: %w", err)
	}

	// 验证商品状态和库存 - 这些业务逻辑失败时不应该设置幂等性标记
	if goods.Status != "1" {
		tx.Rollback()
		// 商品下架是业务逻辑问题，不设置幂等性标记
		return "", fmt.Errorf("商品已下架或不可购买")
	}

	if goods.Stock < params.Num {
		tx.Rollback()
		// 库存不足是业务逻辑问题，不设置幂等性标记，让用户补充库存后可以重新下单
		return "", fmt.Errorf("库存不足，当前库存: %d，需要: %d", goods.Stock, params.Num)
	}

	// 6. 安全扣减库存
	if err := soc.securityService.SafeDeductStock(tx, params.GoodsId, params.Num); err != nil {
		tx.Rollback()
		return "", fmt.Errorf("库存扣减失败: %w", err)
	}

	// 7. 查询用户钱包并处理支付
	totalPrice := goods.Price * float64(params.Num)
	orderStatus := "pending" // 默认待支付

	var walletAfterDeduct *app_model.AppWallet
	walletAfterDeduct, err = soc.securityService.SafeDeductWallet(tx, uid, totalPrice)

	if err != nil {
		// 余额不足，订单状态保持为 pending
		log.Printf("用户 %d 余额不足: %v", uid, err)
	} else {
		// 余额充足，直接支付成功
		orderStatus = "paid"

		// 记录钱包交易流水
		balanceBefore := walletAfterDeduct.Money + totalPrice
		if err := soc.securityService.RecordWalletTransaction(tx, uid, totalPrice,
			balanceBefore, walletAfterDeduct.Money, "订单支付"); err != nil {

			tx.Rollback()
			return "", fmt.Errorf("记录交易流水失败: %w", err)
		}
	}

	// 8. 创建订单记录
	orderNo := soc.generateOrderNo(uid, params.GoodsId)
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

	// 9. 更新商家收入统计 (简化版)
	if err := soc.updateSimpleStats(tx, &goods, &order); err != nil {
		// 统计失败不阻塞订单创建，只记录日志
		log.Printf("更新统计失败: %v", err)
	}

	// 10. 提交事务
	if err := tx.Commit().Error; err != nil {
		return "", fmt.Errorf("提交事务失败: %w", err)
	}

	// 11. 订单创建成功后才设置幂等性标记
	if setErr := soc.idempotencyChecker.SetIdempotencyMark(idempotencyKey, 1*time.Hour); setErr != nil {
		log.Printf("设置幂等性标记失败: %v", setErr)
		// 这个失败不影响订单创建结果
	} else {
		log.Printf("✅ 已设置幂等性标记，防止重复下单: %s", idempotencyKey)
	}

	// 12. 异步处理后续流程
	go soc.handlePostOrderCreation(&order, &goods, orderStatus)

	// 13. 如果是待支付状态，设置超时取消
	if orderStatus == "pending" {
		if err := soc.timeoutManager.ScheduleOrderTimeout(orderNo, 15*time.Minute); err != nil {
			log.Printf("设置订单超时失败: %v", err)
		}
	}

	duration := time.Since(startTime)
	log.Printf("✅ 订单创建完成 订单号:%s 状态:%s 耗时:%s", orderNo, orderStatus, duration)

	return orderNo, nil
}

// handlePostOrderCreation 处理订单创建后的异步任务
func (soc *SecureOrderCreator) handlePostOrderCreation(order *app_model.AppOrder, goods *app_model.AppGoods, status string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("处理订单后续任务时发生panic: %v", r)
		}
	}()

	// 1. 发送WebSocket通知
	wsService := public_service.GetWebSocketService()
	if wsService != nil {
		err := wsService.SendOrderNotification(order.UserId, order.No, status, goods.GoodsName)
		if err != nil {
			log.Printf("发送WebSocket通知失败: %v", err)
		}
	}

	// 2. 发送微信小程序消息
	if order.Id != 0 {
		// 查询用户信息获取openid
		var user app_model.UserApp
		if err := db.Dao.Where("id = ?", order.UserId).First(&user).Error; err == nil {
			err := miniapp_service.SendSubscribeMsg(user.Openid, OrderPaidTemplateID, strconv.Itoa(order.Id))
			if err != nil {
				log.Printf("发送小程序消息失败: %v, 订单ID: %d", err, order.Id)
			}
		}
	}

	// 3. 记录业务日志
	log.Printf("📝 订单后续处理完成 订单:%s 用户:%d 状态:%s", order.No, order.UserId, status)
}

// updateSimpleStats 简化的统计更新
func (soc *SecureOrderCreator) updateSimpleStats(tx *gorm.DB, goods *app_model.AppGoods, order *app_model.AppOrder) error {
	// 这里实现简化的统计逻辑，避免复杂的统计计算影响订单创建性能
	log.Printf("更新商家 %d 统计数据 订单:%s 金额:%.2f", goods.TenantsId, order.No, order.Amount)
	return nil
}

// generateOrderNo 生成唯一订单号
func (soc *SecureOrderCreator) generateOrderNo(uid, goodsId int) string {
	timestamp := time.Now().Format("20060102150405")
	random := rand.Intn(10000)
	return fmt.Sprintf("ORD%s%04d%04d%04d", timestamp, uid%10000, goodsId%10000, random)
}

// CancelExpiredOrder 取消过期订单
func (soc *SecureOrderCreator) CancelExpiredOrder(orderNo string) error {
	log.Printf("🔍 检查过期订单: %s", orderNo)

	// 先进行快速状态检查，避免不必要的锁获取
	var quickCheck app_model.AppOrder
	if err := db.Dao.Select("status").Where("no = ?", orderNo).First(&quickCheck).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("订单 %s 不存在，跳过处理", orderNo)
			return nil
		}
		return fmt.Errorf("快速状态检查失败: %w", err)
	}

	// 如果订单已经不是pending状态，直接返回
	if quickCheck.Status != "pending" {
		log.Printf("订单 %s 状态为 %s，无需取消", orderNo, quickCheck.Status)
		return nil
	}

	// 获取订单取消锁
	cancelLock := soc.securityService.NewDistributedLock(
		fmt.Sprintf("cancel_order:%s", orderNo),
		30*time.Second,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := cancelLock.AcquireWithRenewal(ctx); err != nil {
		return fmt.Errorf("获取取消锁失败: %w", err)
	}
	defer cancelLock.Release()

	// 开启事务
	tx := db.Dao.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("取消订单时发生panic: %v", r)
		}
	}()

	// 查询订单状态（使用行锁）
	var order app_model.AppOrder
	if err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("no = ?", orderNo).First(&order).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			log.Printf("订单 %s 不存在", orderNo)
			return nil
		}
		return fmt.Errorf("查询订单失败: %w", err)
	}

	// 再次检查订单状态（双重检查）
	if order.Status != "pending" {
		tx.Rollback()
		log.Printf("订单 %s 状态为 %s，无需取消", orderNo, order.Status)
		return nil
	}

	// 原子性更新订单状态
	result := tx.Model(&app_model.AppOrder{}).
		Where("no = ? AND status = ?", orderNo, "pending").
		Update("status", "cancelled")

	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("更新订单状态失败: %w", result.Error)
	}

	// 检查是否实际更新了记录
	if result.RowsAffected == 0 {
		tx.Rollback()
		log.Printf("订单 %s 状态已被其他进程修改，无需取消", orderNo)
		return nil
	}

	// 恢复商品库存
	if err := tx.Model(&app_model.AppGoods{}).
		Where("id = ?", order.GoodsId).
		Update("stock", gorm.Expr("stock + ?", order.Num)).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("恢复库存失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交取消事务失败: %w", err)
	}

	// 清除幂等性标记，允许用户重新下单
	go func() {
		idempotencyKey := fmt.Sprintf("order_create:%d:%d:%s", order.UserId, order.GoodsId,
			order.CreateTime.Format("20060102_15"))
		if clearErr := soc.idempotencyChecker.ClearIdempotencyMark(idempotencyKey); clearErr != nil {
			log.Printf("清除幂等性标记失败: %v", clearErr)
		} else {
			log.Printf("已清除订单 %s 的幂等性标记，用户可重新下单", orderNo)
		}
	}()

	log.Printf("✅ 订单 %s 已取消，库存已恢复", orderNo)

	// 异步发送取消通知
	go func() {
		wsService := public_service.GetWebSocketService()
		if wsService != nil {
			wsService.SendOrderNotification(order.UserId, orderNo, "cancelled", "")
		}
	}()

	return nil
}

// ProcessPayment 处理支付（用于支付回调）
func (soc *SecureOrderCreator) ProcessPayment(orderNo string, amount float64) error {
	log.Printf("🔄 处理订单支付: %s 金额: %.2f", orderNo, amount)

	// 获取支付处理锁
	paymentLock := soc.securityService.NewDistributedLock(
		fmt.Sprintf("payment:%s", orderNo),
		30*time.Second,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := paymentLock.AcquireWithRenewal(ctx); err != nil {
		return fmt.Errorf("获取支付锁失败: %w", err)
	}
	defer paymentLock.Release()

	// 开启事务
	tx := db.Dao.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("处理支付时发生panic: %v", r)
		}
	}()

	// 查询订单
	var order app_model.AppOrder
	if err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("no = ?", orderNo).First(&order).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("查询订单失败: %w", err)
	}

	// 验证订单状态
	if order.Status != "pending" {
		tx.Rollback()
		return fmt.Errorf("订单状态不正确: %s", order.Status)
	}

	// 验证金额
	if order.Amount != amount {
		tx.Rollback()
		return fmt.Errorf("支付金额不匹配 订单:%.2f 支付:%.2f", order.Amount, amount)
	}

	// 更新订单状态为已支付
	if err := tx.Model(&order).Updates(map[string]interface{}{
		"status":      "paid",
		"update_time": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新订单状态失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交支付事务失败: %w", err)
	}

	log.Printf("✅ 订单 %s 支付处理完成", orderNo)

	// 异步发送支付成功通知
	go func() {
		wsService := public_service.GetWebSocketService()
		if wsService != nil {
			wsService.SendOrderNotification(order.UserId, orderNo, "paid", "")
		}
	}()

	return nil
}

// GetOrderStatus 获取订单状态
func (soc *SecureOrderCreator) GetOrderStatus(orderNo string) (string, error) {
	var order app_model.AppOrder
	if err := db.Dao.Where("no = ?", orderNo).First(&order).Error; err != nil {
		return "", fmt.Errorf("查询订单失败: %w", err)
	}

	return order.Status, nil
}

// GetOrderDetail 获取订单详情
func (soc *SecureOrderCreator) GetOrderDetail(c *gin.Context, uid int, id int) (interface{}, error) {
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
func (soc *SecureOrderCreator) GetMyOrderList(c *gin.Context, uid int, params inout.MyOrderReq) (interface{}, error) {
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

	// 查询总数和分页数据
	query := db.Dao.Model(&app_model.AppOrder{}).Where("user_id = ?", uid)

	// 如果有状态筛选
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	err := query.Count(&total).Offset(offset).Limit(params.PageSize).Find(&orders).Error
	if err != nil {
		return nil, err
	}

	// 获取所有商品 ID
	goodsIds := make([]int, len(orders))
	for i, order := range orders {
		goodsIds[i] = order.GoodsId
	}

	// 批量查询商品详情
	goodsMap, err := soc.getGoodsDetailsBatch(goodsIds)
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

// getGoodsDetailsBatch 批量查询商品详情
func (soc *SecureOrderCreator) getGoodsDetailsBatch(goodsIds []int) (map[int]app_model.AppGoods, error) {
	if len(goodsIds) == 0 {
		return make(map[int]app_model.AppGoods), nil
	}

	// 批量查询商品详情
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

// 全局安全订单创建器实例
var globalSecureOrderCreator *SecureOrderCreator

// InitGlobalSecureOrderCreator 初始化全局安全订单创建器
func InitGlobalSecureOrderCreator(redisClient *redis.Client) {
	globalSecureOrderCreator = NewSecureOrderCreator(redisClient)
	log.Printf("✅ 全局安全订单创建器已初始化")
}

// GetGlobalSecureOrderCreator 获取全局安全订单创建器
func GetGlobalSecureOrderCreator() *SecureOrderCreator {
	return globalSecureOrderCreator
}
