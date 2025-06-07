package public_service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"nasa-go-admin/db"
	"nasa-go-admin/middleware"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/pkg/websocket"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// NotificationType 通知类型
type NotificationType string

const (
	// 订单相关通知
	OrderCreated   NotificationType = "order_created"
	OrderPaid      NotificationType = "order_paid"
	OrderCancelled NotificationType = "order_cancelled"
	OrderRefunded  NotificationType = "order_refunded"
	OrderShipped   NotificationType = "order_shipped"
	OrderDelivered NotificationType = "order_delivered"

	// 用户相关通知
	UserRegistered NotificationType = "user_registered"
	UserLogin      NotificationType = "user_login"
	UserUpdated    NotificationType = "user_updated"

	// 系统通知
	SystemNotice   NotificationType = "system_notice"
	SystemMaintain NotificationType = "system_maintain"
	SystemUpgrade  NotificationType = "system_upgrade"

	// 其他业务通知
	PaymentSuccess  NotificationType = "payment_success"
	PaymentFailed   NotificationType = "payment_failed"
	MessageReceived NotificationType = "message_received"
	CommentReceived NotificationType = "comment_received"
)

// NotificationPriority 通知优先级
type NotificationPriority int

const (
	PriorityLow    NotificationPriority = 0
	PriorityNormal NotificationPriority = 1
	PriorityHigh   NotificationPriority = 2
	PriorityUrgent NotificationPriority = 3
)

// NotificationTarget 通知目标
type NotificationTarget int

const (
	TargetUser   NotificationTarget = 0 // 发送给特定用户
	TargetAdmin  NotificationTarget = 1 // 发送给管理员
	TargetAll    NotificationTarget = 2 // 广播给所有人
	TargetGroup  NotificationTarget = 3 // 发送给特定组
	TargetCustom NotificationTarget = 4 // 自定义发送目标
)

// NotificationMessage 通知消息
type NotificationMessage struct {
	Type        NotificationType     `json:"type"`
	Content     string               `json:"content"`
	Data        interface{}          `json:"data,omitempty"`
	Time        string               `json:"time"`
	Priority    NotificationPriority `json:"-"`
	Target      NotificationTarget   `json:"-"`
	TargetIDs   []int                `json:"-"`
	ExcludeIDs  []int                `json:"-"`
	NeedConfirm bool                 `json:"-"`
	MessageID   string               `json:"message_id,omitempty"`
}

// SendTask 发送任务
type SendTask struct {
	Message  *NotificationMessage
	UserIDs  []int
	Response chan<- error
	Attempt  int
}

// WebSocketService 提供WebSocket通信服务
type WebSocketService struct {
	hub               *websocket.Hub
	once              sync.Once
	adminSubscribers  sync.Map
	adminCache        []int
	adminCacheTime    time.Time
	workersStarted    bool
	workersMutex      sync.Mutex
	messageQueue      chan *SendTask
	workerCount       int
	outboundRate      int64 // 每秒发送消息数统计
	inboundRate       int64 // 每秒接收消息数统计
	activeConnections int64 // 活跃连接数
	failedMessages    int64 // 失败消息数
	ctxWorkers        context.Context
	cancelWorkers     context.CancelFunc
}

var (
	wsService     *WebSocketService
	wsServiceOnce sync.Once
)

// GetWebSocketService 获取单例WebSocketService
func GetWebSocketService() *WebSocketService {
	wsServiceOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		wsService = &WebSocketService{
			messageQueue:  make(chan *SendTask, 10000), // 可容纳10000条消息的缓冲队列
			workerCount:   runtime.NumCPU() * 2,        // 工作线程数为CPU核心数的2倍
			ctxWorkers:    ctx,
			cancelWorkers: cancel,
		}
	})
	return wsService
}

// InitHub 初始化WebSocket Hub
func (s *WebSocketService) InitHub() *websocket.Hub {
	s.once.Do(func() {
		s.hub = websocket.NewHub()
		go s.hub.Run()
		s.startWorkers()           // 启动工作线程池
		go s.startStatsCollector() // 启动统计收集器
		log.Println("WebSocket Hub已初始化并启动")
	})
	return s.hub
}

// GetHub 获取WebSocket Hub
func (s *WebSocketService) GetHub() *websocket.Hub {
	if s.hub == nil {
		return s.InitHub()
	}
	return s.hub
}

// startWorkers 启动工作线程池
func (s *WebSocketService) startWorkers() {
	s.workersMutex.Lock()
	defer s.workersMutex.Unlock()

	if s.workersStarted {
		return
	}

	for i := 0; i < s.workerCount; i++ {
		go s.worker(i)
	}
	s.workersStarted = true
	log.Printf("已启动 %d 个WebSocket消息发送工作线程", s.workerCount)
}

// worker 消息发送工作线程
func (s *WebSocketService) worker(id int) {
	log.Printf("WebSocket工作线程 #%d 已启动", id)
	for {
		select {
		case <-s.ctxWorkers.Done():
			log.Printf("WebSocket工作线程 #%d 已停止", id)
			return
		case task := <-s.messageQueue:
			if task == nil {
				continue
			}

			// 将消息序列化
			msgBytes, err := json.Marshal(task.Message)
			if err != nil {
				log.Printf("消息序列化失败: %v", err)
				if task.Response != nil {
					task.Response <- err
				}
				atomic.AddInt64(&s.failedMessages, 1)
				continue
			}

			// 根据优先级和目标发送消息
			switch task.Message.Target {
			case TargetUser:
				for _, userID := range task.UserIDs {
					s.GetHub().SendToUser(userID, msgBytes)
				}
			case TargetAdmin:
				adminIDs := s.getAdminUserIDs()
				for _, adminID := range adminIDs {
					s.GetHub().SendToUser(adminID, msgBytes)
				}
			case TargetAll:
				s.GetHub().Broadcast <- msgBytes
			case TargetGroup:
				// 实现群组消息发送逻辑
			case TargetCustom:
				// 发送给指定的用户ID列表，排除ExcludeIDs中的用户
				for _, userID := range task.UserIDs {
					excluded := false
					for _, excludeID := range task.Message.ExcludeIDs {
						if userID == excludeID {
							excluded = true
							break
						}
					}
					if !excluded {
						s.GetHub().SendToUser(userID, msgBytes)
					}
				}
			}

			atomic.AddInt64(&s.outboundRate, 1)
			if task.Response != nil {
				task.Response <- nil // 发送成功，返回nil
			}
		}
	}
}

// startStatsCollector 启动统计数据收集器
func (s *WebSocketService) startStatsCollector() {
	ticker := time.NewTicker(100 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctxWorkers.Done():
			return
		case <-ticker.C: // 使用 ticker.C 而不是 ticker.Tick()
			// 收集当前统计数据并可以选择将其写入日志或监控系统
			outbound := atomic.SwapInt64(&s.outboundRate, 0)
			inbound := atomic.SwapInt64(&s.inboundRate, 0)
			active := atomic.LoadInt64(&s.activeConnections)
			failed := atomic.LoadInt64(&s.failedMessages)

			log.Printf("WebSocket统计: 活跃连接=%d, 入站消息率=%d/10s, 出站消息率=%d/10s, 失败消息=%d",
				active, inbound, outbound, failed)
		}
	}
}

// SendNotification 发送通用通知
func (s *WebSocketService) SendNotification(msg *NotificationMessage) error {
	responseCh := make(chan error, 1)

	// 根据目标类型准备接收者ID列表
	var targetIDs []int
	switch msg.Target {
	case TargetUser:
		targetIDs = msg.TargetIDs
	case TargetAdmin:
		targetIDs = s.getAdminUserIDs()
	case TargetAll:
		// 广播不需要特定用户ID
	case TargetGroup:
		// 获取群组成员ID列表
		// targetIDs = s.getGroupMemberIDs(msg.GroupID)
	case TargetCustom:
		targetIDs = msg.TargetIDs
	}

	// 设置消息时间
	if msg.Time == "" {
		msg.Time = time.Now().Format("2006-01-02 15:04:05")
	}

	// 生成消息ID
	if msg.MessageID == "" {
		msg.MessageID = generateMessageID()
	}

	task := &SendTask{
		Message:  msg,
		UserIDs:  targetIDs,
		Response: responseCh,
	}

	// 根据优先级决定是否阻塞或排队
	if msg.Priority >= PriorityUrgent {
		// 紧急消息，直接处理，不经过队列
		select {
		case s.messageQueue <- task:
			// 已放入队列
		default:
			// 队列已满，创建新的临时工作线程处理
			go func() {
				msgBytes, err := json.Marshal(msg)
				if err != nil {
					responseCh <- err
					return
				}

				switch msg.Target {
				case TargetAll:
					s.GetHub().Broadcast <- msgBytes
				default:
					for _, userID := range targetIDs {
						s.GetHub().SendToUser(userID, msgBytes)
					}
				}

				responseCh <- nil
			}()
		}
	} else {
		// 普通优先级，放入队列
		select {
		case s.messageQueue <- task:
			// 成功放入队列
		default:
			// 队列已满，返回错误
			return fmt.Errorf("通知队列已满，请稍后再试")
		}
	}

	// 等待结果
	select {
	case err := <-responseCh:
		return err
	case <-time.After(5 * time.Second):
		return fmt.Errorf("发送通知超时")
	}
}

// SendOrderNotification 发送订单通知 (向后兼容)
func (s *WebSocketService) SendOrderNotification(userID int, orderNo, status, goodsName string) error {
	// 根据状态确定消息类型
	var msgType NotificationType
	var content string
	var adminContent string

	switch status {
	case "paid":
		msgType = OrderPaid
		content = "订单支付成功"
		adminContent = "新订单支付成功通知"
	case "cancelled":
		msgType = OrderCancelled
		content = "订单已取消"
		adminContent = "订单取消通知"
	case "refunded":
		msgType = OrderRefunded
		content = "订单已退款"
		adminContent = "订单退款通知"
	default:
		msgType = OrderCreated
		content = "订单创建成功"
		adminContent = "新订单创建通知"
	}

	// 准备消息数据
	data := map[string]interface{}{
		"order_no":   orderNo,
		"status":     status,
		"goods_name": goodsName,
		"user_id":    userID,
	}

	// 创建用户消息
	userMsg := &NotificationMessage{
		Type:      msgType,
		Content:   content,
		Data:      data,
		Time:      time.Now().Format("2006-01-02 15:04:05"),
		Priority:  PriorityNormal,
		Target:    TargetUser,
		TargetIDs: []int{userID},
	}

	// 发送用户消息
	if err := s.SendNotification(userMsg); err != nil {
		log.Printf("发送用户订单通知失败: %v", err)
		return err
	}

	// 创建管理员消息
	adminMsg := &NotificationMessage{
		Type:     msgType,
		Content:  adminContent,
		Data:     data,
		Time:     time.Now().Format("2006-01-02 15:04:05"),
		Priority: PriorityNormal,
		Target:   TargetAdmin,
	}

	// 发送管理员消息
	if err := s.SendNotification(adminMsg); err != nil {
		log.Printf("发送管理员订单通知失败: %v", err)
		// 不影响用户通知，不返回错误
	}

	return nil
}

// BroadcastSystemNotice 广播系统通知 (向后兼容)
func (s *WebSocketService) BroadcastSystemNotice(content string) error {
	messageID := generateMessageID()
	msg := &NotificationMessage{
		Type:      SystemNotice,
		Content:   content,
		Time:      time.Now().Format("2006-01-02 15:04:05"),
		Priority:  PriorityNormal,
		Target:    TargetAll,
		MessageID: messageID,
	}
	// 异步记录通知发送事件
	go middleware.LogWebSocketEvent("notification_sent", 0, messageID, map[string]interface{}{
		"type":    SystemNotice,
		"content": content,
		"time":    msg.Time,
	})
	return s.SendNotification(msg)
}

// getAdminUserIDs 获取所有管理员的用户ID (带缓存)
func (s *WebSocketService) getAdminUserIDs() []int {
	// 检查缓存
	if s.adminCache != nil && time.Since(s.adminCacheTime) < 5*time.Minute {
		return s.adminCache
	}

	// 从数据库查询管理员
	var adminIDs []int
	// 查询所有管理员角色的用户
	err := db.Dao.Model(&admin_model.AdminUser{}).
		Where("notice = ?", 1).
		Pluck("id", &adminIDs).Error

	if err != nil {
		log.Printf("查询管理员ID失败: %v", err)
		return []int{}
	}

	// 更新缓存
	s.adminCache = adminIDs
	s.adminCacheTime = time.Now()

	return adminIDs
}

// SendUserNotification 发送用户相关通知
func (s *WebSocketService) SendUserNotification(userID int, noticeType NotificationType, content string, data interface{}) error {
	// 生成唯一的消息ID，用于日志追踪
	messageID := generateMessageID()
	fmt.Println("SendUserNotification")
	msg := &NotificationMessage{
		Type:      noticeType,
		Content:   content,
		Data:      data,
		Time:      time.Now().Format("2006-01-02 15:04:05"),
		Priority:  PriorityNormal,
		Target:    TargetUser,
		TargetIDs: []int{userID},
		MessageID: messageID,
	}

	// 异步记录通知发送事件
	go middleware.LogWebSocketEvent("notification_sent", userID, messageID, map[string]interface{}{
		"type":    string(noticeType),
		"content": content,
		"data":    data,
		"time":    msg.Time,
	})

	// 发送通知
	err := s.SendNotification(msg)

	// 如果发送失败，记录错误
	if err != nil {
		go middleware.LogWebSocketEvent("notification_failed", userID, messageID, map[string]interface{}{
			"error":   err.Error(),
			"type":    string(noticeType),
			"content": content,
		})
	} else {
		// 发送成功，记录成功事件
		go middleware.LogWebSocketEvent("notification_delivered", userID, messageID, map[string]interface{}{
			"delivered_at": time.Now().Format("2006-01-02 15:04:05"),
		})
	}

	return err
}

// SendGroupNotification 发送群组通知
func (s *WebSocketService) SendGroupNotification(groupID int, content string, data interface{}) error {
	// 获取群组成员ID列表
	memberIDs := s.getGroupMemberIDs(groupID)

	msg := &NotificationMessage{
		Type:      SystemNotice,
		Content:   content,
		Data:      data,
		Time:      time.Now().Format("2006-01-02 15:04:05"),
		Priority:  PriorityNormal,
		Target:    TargetCustom,
		TargetIDs: memberIDs,
	}

	return s.SendNotification(msg)
}

// getGroupMemberIDs 获取群组成员ID列表
func (s *WebSocketService) getGroupMemberIDs(groupID int) []int {
	// 实现获取群组成员ID的逻辑
	// 例如：从数据库查询群组成员
	var memberIDs []int
	// db.Dao.Model(&GroupMember{}).Where("group_id = ?", groupID).Pluck("user_id", &memberIDs)
	return memberIDs
}

// RegisterConnectionStatus 注册连接状态变更
func (s *WebSocketService) RegisterConnectionStatus(connected bool) {
	if connected {
		atomic.AddInt64(&s.activeConnections, 1)
	} else {
		atomic.AddInt64(&s.activeConnections, -1)
	}
}

// RegisterMessageReceived 注册收到消息
func (s *WebSocketService) RegisterMessageReceived() {
	atomic.AddInt64(&s.inboundRate, 1)
}

// 生成唯一的消息ID
func generateMessageID() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), uuid.New().String()[:8])
}

// 关闭WebSocket服务
func (s *WebSocketService) Close() {
	if s.cancelWorkers != nil {
		s.cancelWorkers()
	}
	log.Println("WebSocket服务已关闭")
}
