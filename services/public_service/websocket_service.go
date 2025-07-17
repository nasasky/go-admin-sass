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
	"nasa-go-admin/services/admin_service"
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
	Message    *NotificationMessage
	UserIDs    []int
	Response   chan<- error
	Attempt    int
	MaxRetries int           // 最大重试次数
	RetryDelay time.Duration // 重试延迟
}

// UserCache 用户信息缓存
type UserCache struct {
	cache   *sync.Map
	ttl     time.Duration
	maxSize int
}

// UserInfo 用户信息
type UserInfo struct {
	UserID    int
	Username  string
	UserType  string
	ExpiresAt time.Time
}

// IsExpired 检查是否过期
func (ui *UserInfo) IsExpired() bool {
	return time.Now().After(ui.ExpiresAt)
}

// OnlineStatusCache 在线状态缓存
type OnlineStatusCache struct {
	cache *sync.Map
	ttl   time.Duration
}

// OnlineStatus 在线状态
type OnlineStatus struct {
	UserID    int
	IsOnline  bool
	ExpiresAt time.Time
}

// IsExpired 检查是否过期
func (os *OnlineStatus) IsExpired() bool {
	return time.Now().After(os.ExpiresAt)
}

// MessageDeduplicator 消息去重器
type MessageDeduplicator struct {
	cache *sync.Map
	ttl   time.Duration
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
	offlineService    *OfflineMessageService // 离线消息服务

	// 优化：添加缓存和性能监控
	userCache           *UserCache           // 用户信息缓存
	onlineStatusCache   *OnlineStatusCache   // 在线状态缓存
	messageDeduplicator *MessageDeduplicator // 消息去重器
	startTime           time.Time            // 服务启动时间
	lastError           error                // 最后一次错误
	lastErrorTime       time.Time            // 最后一次错误时间
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
			messageQueue:   make(chan *SendTask, 50000), // 优化：增加到50000条消息缓冲队列
			workerCount:    runtime.NumCPU() * 2,        // 优化：调整为CPU核心数的2倍，避免过多线程
			ctxWorkers:     ctx,
			cancelWorkers:  cancel,
			offlineService: NewOfflineMessageService(), // 初始化离线消息服务

			// 初始化缓存
			userCache: &UserCache{
				cache:   &sync.Map{},
				ttl:     5 * time.Minute, // 5分钟缓存
				maxSize: 10000,
			},
			onlineStatusCache: &OnlineStatusCache{
				cache: &sync.Map{},
				ttl:   30 * time.Second, // 30秒缓存
			},
			messageDeduplicator: &MessageDeduplicator{
				cache: &sync.Map{},
				ttl:   1 * time.Minute, // 1分钟去重
			},
			startTime: time.Now(),
		}
	})
	return wsService
}

// InitHub 初始化WebSocket Hub
func (s *WebSocketService) InitHub() *websocket.Hub {
	s.once.Do(func() {
		s.hub = websocket.NewHub()

		// 设置用户离线回调
		s.hub.SetUserOfflineCallback(func(userID int, connectionID string) {
			s.UnregisterUserConnection(userID, connectionID)
		})

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

			successCount := 0
			failedCount := 0

			// 根据优先级和目标发送消息
			switch task.Message.Target {
			case TargetUser:
				for _, userID := range task.UserIDs {
					if s.sendMessageToUser(userID, msgBytes, task.Message) {
						successCount++
					} else {
						failedCount++
					}
				}
			case TargetAdmin:
				adminIDs := s.getAdminUserIDs()
				for _, adminID := range adminIDs {
					if s.sendMessageToUser(adminID, msgBytes, task.Message) {
						successCount++
					} else {
						failedCount++
					}
				}
			case TargetAll:
				// 广播消息，发送给所有用户（包括离线用户）
				// 获取所有管理员用户ID，包括在线和离线用户
				allUserIDs := s.getAdminUserIDs()
				for _, userID := range allUserIDs {
					if s.sendMessageToUser(userID, msgBytes, task.Message) {
						successCount++
					} else {
						failedCount++
					}
				}
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
						if s.sendMessageToUser(userID, msgBytes, task.Message) {
							successCount++
						} else {
							failedCount++
						}
					}
				}
			}

			atomic.AddInt64(&s.outboundRate, 1)

			// 记录发送结果
			log.Printf("消息发送完成: MessageID=%s, 成功=%d, 失败=%d",
				task.Message.MessageID, successCount, failedCount)

			if task.Response != nil {
				if failedCount == 0 {
					task.Response <- nil // 全部成功
				} else if successCount == 0 {
					task.Response <- fmt.Errorf("所有消息发送失败")
				} else {
					task.Response <- fmt.Errorf("部分消息发送失败: 成功=%d, 失败=%d", successCount, failedCount)
				}
			}
		}
	}
}

// getOnlineUserIDs 获取在线用户ID列表
func (s *WebSocketService) getOnlineUserIDs() []int {
	hub := s.GetHub()
	return hub.GetOnlineUserIDs()
}

// sendMessageToUser 向特定用户发送消息并处理状态更新
func (s *WebSocketService) sendMessageToUser(userID int, msgBytes []byte, message *NotificationMessage) bool {
	// 检查用户是否在线
	hub := s.GetHub()
	clients := hub.GetUserClients(userID)

	if len(clients) == 0 {
		// 用户不在线，保存为离线消息
		log.Printf("用户 %d 不在线，保存为离线消息: MessageID=%s", userID, message.MessageID)
		err := s.offlineService.SaveOfflineMessage(userID, message)
		if err != nil {
			log.Printf("保存离线消息失败: UserID=%d, MessageID=%s, Error=%v", userID, message.MessageID, err)
		}
		return false
	}

	// 用户在线，发送消息
	success := false
	for _, client := range clients {
		select {
		case client.Send <- msgBytes:
			success = true
			// 延迟标记消息为已投递，确保接收记录已创建
			go func() {
				time.Sleep(100 * time.Millisecond) // 等待100ms确保记录创建完成
				s.markMessageAsDelivered(message.MessageID, userID)
			}()
			log.Printf("消息已发送给用户 %d: MessageID=%s", userID, message.MessageID)
		default:
			// 客户端缓冲区已满，关闭连接
			close(client.Send)
			hub.RemoveClient(client)
		}
	}

	if !success {
		// 所有客户端都发送失败，保存为离线消息
		log.Printf("用户 %d 所有连接发送失败，保存为离线消息: MessageID=%s", userID, message.MessageID)
		err := s.offlineService.SaveOfflineMessage(userID, message)
		if err != nil {
			log.Printf("保存离线消息失败: UserID=%d, MessageID=%s, Error=%v", userID, message.MessageID, err)
		}
	}

	return success
}

// startStatsCollector 启动统计数据收集器
func (s *WebSocketService) startStatsCollector() {
	statsTicker := time.NewTicker(100 * time.Second)
	cleanupTicker := time.NewTicker(2 * time.Minute) // 每2分钟清理一次过期状态
	defer statsTicker.Stop()
	defer cleanupTicker.Stop()

	for {
		select {
		case <-s.ctxWorkers.Done():
			return
		case <-statsTicker.C:
			// 收集当前统计数据并可以选择将其写入日志或监控系统
			outbound := atomic.SwapInt64(&s.outboundRate, 0)
			inbound := atomic.SwapInt64(&s.inboundRate, 0)
			active := atomic.LoadInt64(&s.activeConnections)
			failed := atomic.LoadInt64(&s.failedMessages)

			// 获取更详细的统计信息
			hubStats := s.GetHub().GetStats()

			log.Printf("📊 WebSocket详细统计:")
			log.Printf("  ├─ 连接统计: 活跃连接=%d, 在线用户=%d", active, hubStats["unique_users"])
			log.Printf("  ├─ 消息统计: 入站=%d/100s, 出站=%d/100s, 失败=%d", inbound, outbound, failed)
			log.Printf("  ├─ 成功率: %.2f%%", float64(outbound)/float64(outbound+failed)*100)
			log.Printf("  └─ 队列状态: 待处理=%d", len(s.messageQueue))
		case <-cleanupTicker.C:
			// 清理过期的在线状态
			s.cleanupExpiredOnlineStatus()
		}
	}
}

// cleanupExpiredOnlineStatus 清理过期的在线状态
func (s *WebSocketService) cleanupExpiredOnlineStatus() {
	recordService := admin_service.NewNotificationRecordService()

	// 获取当前Hub中的在线用户
	hub := s.GetHub()
	hubOnlineUsers := hub.GetOnlineUserIDs()
	hubUserMap := make(map[int]bool)
	for _, userID := range hubOnlineUsers {
		hubUserMap[userID] = true
	}

	// 获取数据库中的在线用户
	dbOnlineUsers, err := recordService.GetOnlineAdminUsers()
	if err != nil {
		log.Printf("清理过期在线状态失败: %v", err)
		return
	}

	// 检查数据库中的用户是否真的在Hub中在线
	for _, user := range dbOnlineUsers {
		if !hubUserMap[user.UserID] {
			// 用户不在Hub中，但在数据库中标记为在线，需要修正
			log.Printf("清理过期在线状态: 用户 %d 在数据库中标记为在线，但Hub中不存在", user.UserID)
			recordService.UpdateAdminUserOnlineStatus(user.UserID, user.Username, false, "", "", "")
		}
	}

	log.Printf("在线状态清理完成: Hub在线=%d, 数据库在线=%d", len(hubOnlineUsers), len(dbOnlineUsers))
}

// SendNotification 发送通知
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

	// 针对管理员推送，创建接收记录
	// 注意：TargetAll 也需要为管理员创建接收记录
	if msg.Target == TargetAdmin || msg.Target == TargetCustom || msg.Target == TargetAll {
		// 获取管理员用户ID列表
		adminUserIDs := s.getAdminUserIDs()
		if len(adminUserIDs) > 0 {
			go s.createAdminUserReceiveRecords(msg, adminUserIDs)
		}
	}

	task := &SendTask{
		Message:    msg,
		UserIDs:    targetIDs,
		Response:   responseCh,
		Attempt:    0,
		MaxRetries: 3,               // 最大重试3次
		RetryDelay: 2 * time.Second, // 重试间隔2秒
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

				successCount := 0
				failedCount := 0

				switch msg.Target {
				case TargetAll:
					// 广播消息，发送给所有用户（包括离线用户）
					// 获取所有管理员用户ID，包括在线和离线用户
					allUserIDs := s.getAdminUserIDs()
					for _, userID := range allUserIDs {
						if s.sendMessageToUser(userID, msgBytes, msg) {
							successCount++
						} else {
							failedCount++
						}
					}
				default:
					for _, userID := range targetIDs {
						if s.sendMessageToUser(userID, msgBytes, msg) {
							successCount++
						} else {
							failedCount++
						}
					}
				}

				if failedCount == 0 {
					responseCh <- nil
				} else if successCount == 0 {
					responseCh <- fmt.Errorf("所有消息发送失败")
				} else {
					responseCh <- fmt.Errorf("部分消息发送失败: 成功=%d, 失败=%d", successCount, failedCount)
				}
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

// createAdminUserReceiveRecords 为管理员用户创建接收记录
func (s *WebSocketService) createAdminUserReceiveRecords(msg *NotificationMessage, targetUserIDs []int) {
	if len(targetUserIDs) == 0 {
		return
	}

	recordService := admin_service.NewNotificationRecordService()

	// 获取在线用户列表
	onlineUsers, _ := recordService.GetOnlineAdminUsers()
	onlineUserMap := make(map[int]bool)
	for _, user := range onlineUsers {
		onlineUserMap[user.UserID] = true
	}

	// 为每个目标用户创建接收记录
	for _, userID := range targetUserIDs {
		// 获取用户信息
		username := s.getUsernameByID(userID)
		userRole := s.getUserRoleByID(userID)
		isOnline := onlineUserMap[userID]

		record := &admin_model.AdminUserReceiveRecord{
			MessageID:      msg.MessageID,
			UserID:         userID,
			Username:       username,
			UserRole:       userRole,
			IsOnline:       isOnline,
			IsReceived:     false,
			IsRead:         false,
			IsConfirmed:    false,
			DeviceType:     "desktop",
			Platform:       "web",
			Browser:        "unknown",
			ClientIP:       "",
			UserAgent:      "",
			ConnectionID:   "",
			PushChannel:    "websocket",
			DeliveryStatus: "pending",
			RetryCount:     0,
		}

		// 保存记录
		err := recordService.SaveAdminUserReceiveRecord(record)
		if err != nil {
			log.Printf("创建管理员用户接收记录失败: MessageID=%s, UserID=%d, Error=%v",
				msg.MessageID, userID, err)
		}
	}

	log.Printf("已为消息创建接收记录: MessageID=%s, UserCount=%d", msg.MessageID, len(targetUserIDs))
}

// markMessageAsDelivered 标记消息为已投递
func (s *WebSocketService) markMessageAsDelivered(messageID string, userID int) {
	recordService := admin_service.NewNotificationRecordService()

	updates := map[string]interface{}{
		"is_received":     true,
		"received_at":     time.Now().Format("2006-01-02 15:04:05"),
		"delivery_status": "delivered",
	}

	err := recordService.UpdateAdminUserReceiveRecord(messageID, userID, updates)
	if err != nil {
		// 如果记录不存在，尝试重试几次
		if err.Error() == "管理员用户接收记录不存在" {
			log.Printf("接收记录尚未创建，等待后重试: MessageID=%s, UserID=%d", messageID, userID)
			// 重试机制：等待更长时间后再次尝试
			go func() {
				time.Sleep(500 * time.Millisecond)
				retryErr := recordService.UpdateAdminUserReceiveRecord(messageID, userID, updates)
				if retryErr != nil {
					log.Printf("重试标记消息投递状态失败: MessageID=%s, UserID=%d, Error=%v",
						messageID, userID, retryErr)
				} else {
					log.Printf("重试成功标记消息投递状态: MessageID=%s, UserID=%d", messageID, userID)
				}
			}()
		} else {
			log.Printf("标记消息投递状态失败: MessageID=%s, UserID=%d, Error=%v",
				messageID, userID, err)
		}
	} else {
		log.Printf("成功标记消息投递状态: MessageID=%s, UserID=%d", messageID, userID)
	}
}

// getUsernameByID 根据用户ID获取用户名（简化实现）
func (s *WebSocketService) getUsernameByID(userID int) string {
	// 这里应该从数据库查询，简化处理
	var username string
	err := db.Dao.Model(&admin_model.AdminUser{}).
		Where("id = ?", userID).
		Pluck("username", &username).Error

	if err != nil {
		return "unknown"
	}
	return username
}

// getUserRoleByID 根据用户ID获取用户角色（简化实现）
func (s *WebSocketService) getUserRoleByID(userID int) string {
	// 这里应该从数据库查询用户角色，简化处理
	return "admin"
}

// RegisterUserConnection 注册用户连接（增强版）
func (s *WebSocketService) RegisterUserConnection(userID int, connectionID string, clientIP string, userAgent string) {
	// 更新用户在线状态
	recordService := admin_service.NewNotificationRecordService()
	username := s.getUsernameByID(userID)

	err := recordService.UpdateAdminUserOnlineStatus(userID, username, true, connectionID, clientIP, userAgent)
	if err != nil {
		log.Printf("更新用户在线状态失败: UserID=%d, Error=%v", userID, err)
	}

	// 更新所有相关接收记录的在线状态
	err = recordService.UpdateUserReceiveRecordsOnlineStatus(userID, true, connectionID)
	if err != nil {
		log.Printf("更新用户接收记录在线状态失败: UserID=%d, Error=%v", userID, err)
	}

	log.Printf("用户连接已注册: UserID=%d, ConnectionID=%s", userID, connectionID)
}

// UnregisterUserConnection 注销用户连接（增强版）
func (s *WebSocketService) UnregisterUserConnection(userID int, connectionID string) {
	// 更新用户在线状态
	recordService := admin_service.NewNotificationRecordService()
	username := s.getUsernameByID(userID)

	err := recordService.UpdateAdminUserOnlineStatus(userID, username, false, connectionID, "", "")
	if err != nil {
		log.Printf("更新用户离线状态失败: UserID=%d, Error=%v", userID, err)
	}

	// 更新所有相关接收记录的在线状态
	err = recordService.UpdateUserReceiveRecordsOnlineStatus(userID, false, "")
	if err != nil {
		log.Printf("更新用户接收记录在线状态失败: UserID=%d, Error=%v", userID, err)
	}

	log.Printf("用户连接已注销: UserID=%d, ConnectionID=%s", userID, connectionID)
}

// GetUserReceiveStatistics 获取用户接收统计
func (s *WebSocketService) GetUserReceiveStatistics(userID int, startDate, endDate string) (map[string]interface{}, error) {
	recordService := admin_service.NewNotificationRecordService()

	query := &admin_model.AdminUserReceiveQuery{
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
		Page:      1,
		PageSize:  1000, // 获取所有记录用于统计
	}

	result, err := recordService.GetAdminUserReceiveRecords(query)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"user_id":         userID,
		"total_messages":  result.Total,
		"received_count":  0,
		"read_count":      0,
		"confirmed_count": 0,
		"online_rate":     0.0,
		"response_time":   0.0,
	}

	return stats, nil
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

	// 从数据库查询所有用户
	var adminIDs []int
	// 查询所有用户，不限制user_type
	err := db.Dao.Model(&admin_model.AdminUser{}).
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

// GetUserInfo 获取用户信息（带缓存）
func (s *WebSocketService) GetUserInfo(userID int) (*UserInfo, error) {
	// 先从缓存获取
	if cached, ok := s.userCache.cache.Load(userID); ok {
		if userInfo, ok := cached.(*UserInfo); ok && !userInfo.IsExpired() {
			return userInfo, nil
		}
	}

	// 从数据库加载
	userInfo, err := s.loadUserInfoFromDB(userID)
	if err != nil {
		return nil, err
	}

	// 设置过期时间
	userInfo.ExpiresAt = time.Now().Add(s.userCache.ttl)

	// 存入缓存
	s.userCache.cache.Store(userID, userInfo)
	return userInfo, nil
}

// loadUserInfoFromDB 从数据库加载用户信息
func (s *WebSocketService) loadUserInfoFromDB(userID int) (*UserInfo, error) {
	// 这里应该从数据库查询用户信息
	// 暂时返回模拟数据
	return &UserInfo{
		UserID:   userID,
		Username: fmt.Sprintf("user_%d", userID),
		UserType: "admin",
	}, nil
}

// IsUserOnline 检查用户是否在线（带缓存）
func (s *WebSocketService) IsUserOnline(userID int) bool {
	// 先从缓存获取
	if status, ok := s.onlineStatusCache.cache.Load(userID); ok {
		if onlineStatus, ok := status.(*OnlineStatus); ok && !onlineStatus.IsExpired() {
			return onlineStatus.IsOnline
		}
	}

	// 从Hub检查
	hub := s.GetHub()
	isOnline := hub.IsUserOnline(userID)

	// 缓存结果
	onlineStatus := &OnlineStatus{
		UserID:    userID,
		IsOnline:  isOnline,
		ExpiresAt: time.Now().Add(s.onlineStatusCache.ttl),
	}
	s.onlineStatusCache.cache.Store(userID, onlineStatus)

	return isOnline
}

// IsMessageDuplicate 检查消息是否重复
func (s *WebSocketService) IsMessageDuplicate(messageID string, userID int) bool {
	key := fmt.Sprintf("%s:%d", messageID, userID)
	if _, exists := s.messageDeduplicator.cache.Load(key); exists {
		return true
	}

	// 设置去重标记
	s.messageDeduplicator.cache.Store(key, time.Now())

	// 异步清理过期记录
	go func() {
		time.Sleep(s.messageDeduplicator.ttl)
		s.messageDeduplicator.cache.Delete(key)
	}()

	return false
}

// SendNotificationAsync 异步发送通知（不阻塞主流程）
func (s *WebSocketService) SendNotificationAsync(msg *NotificationMessage) {
	go func() {
		err := s.SendNotification(msg)
		if err != nil {
			s.lastError = err
			s.lastErrorTime = time.Now()
			log.Printf("异步发送消息失败: %v", err)
		}
	}()
}

// GetMetrics 获取性能指标
func (s *WebSocketService) GetMetrics() map[string]interface{} {
	hub := s.GetHub()
	hubStats := hub.GetDetailedStats()

	return map[string]interface{}{
		"active_connections": atomic.LoadInt64(&s.activeConnections),
		"outbound_rate":      atomic.LoadInt64(&s.outboundRate),
		"inbound_rate":       atomic.LoadInt64(&s.inboundRate),
		"failed_messages":    atomic.LoadInt64(&s.failedMessages),
		"queue_size":         len(s.messageQueue),
		"worker_count":       s.workerCount,
		"uptime":             time.Since(s.startTime).String(),
		"last_error":         s.lastError,
		"last_error_time":    s.lastErrorTime,
		"hub_stats":          hubStats,
	}
}

// HealthCheck 健康检查
func (s *WebSocketService) HealthCheck() map[string]interface{} {
	status := "healthy"
	if s.lastError != nil && time.Since(s.lastErrorTime) < 5*time.Minute {
		status = "degraded"
	}

	return map[string]interface{}{
		"status":             status,
		"active_connections": atomic.LoadInt64(&s.activeConnections),
		"queue_size":         len(s.messageQueue),
		"worker_count":       s.workerCount,
		"last_error":         s.lastError,
		"uptime":             time.Since(s.startTime).String(),
	}
}

// 关闭WebSocket服务
func (s *WebSocketService) Close() {
	if s.cancelWorkers != nil {
		s.cancelWorkers()
	}
	log.Println("WebSocket服务已关闭")
}
