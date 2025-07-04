package public_service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"nasa-go-admin/pkg/websocket"
	"nasa-go-admin/redis"
	"time"
)

// OfflineMessageService 离线消息服务
type OfflineMessageService struct {
	redisClient interface{}
}

// NewOfflineMessageService 创建离线消息服务
func NewOfflineMessageService() *OfflineMessageService {
	return &OfflineMessageService{
		redisClient: redis.GetClient(),
	}
}

// SaveOfflineMessage 保存离线消息
func (oms *OfflineMessageService) SaveOfflineMessage(userID int, message *NotificationMessage) error {
	key := fmt.Sprintf("offline_msg:%d", userID)

	// 添加时间戳
	offlineMsg := struct {
		*NotificationMessage
		SavedAt string `json:"saved_at"`
	}{
		NotificationMessage: message,
		SavedAt:             time.Now().Format("2006-01-02 15:04:05"),
	}

	msgData, err := json.Marshal(offlineMsg)
	if err != nil {
		return fmt.Errorf("序列化离线消息失败: %w", err)
	}

	// 使用Redis List存储，最新消息在前面
	ctx := context.Background()
	pipe := redis.GetClient().Pipeline()

	// 添加消息到列表前面
	pipe.LPush(ctx, key, msgData)

	// 保持最多100条离线消息
	pipe.LTrim(ctx, key, 0, 99)

	// 设置7天过期时间
	pipe.Expire(ctx, key, 7*24*time.Hour)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("保存离线消息失败: %w", err)
	}

	log.Printf("💾 离线消息已保存: UserID=%d, MessageID=%s", userID, message.MessageID)
	return nil
}

// GetOfflineMessages 获取用户的离线消息
func (oms *OfflineMessageService) GetOfflineMessages(userID int) ([]*NotificationMessage, error) {
	key := fmt.Sprintf("offline_msg:%d", userID)
	ctx := context.Background()

	// 获取所有离线消息
	messages, err := redis.GetClient().LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("获取离线消息失败: %w", err)
	}

	var notifications []*NotificationMessage
	for _, msgData := range messages {
		var offlineMsg struct {
			*NotificationMessage
			SavedAt string `json:"saved_at"`
		}

		if err := json.Unmarshal([]byte(msgData), &offlineMsg); err != nil {
			log.Printf("反序列化离线消息失败: %v", err)
			continue
		}

		// 添加离线标记
		offlineMsg.NotificationMessage.Content = fmt.Sprintf("[离线消息] %s",
			offlineMsg.NotificationMessage.Content)

		notifications = append(notifications, offlineMsg.NotificationMessage)
	}

	if len(notifications) > 0 {
		log.Printf("📤 获取离线消息: UserID=%d, 数量=%d", userID, len(notifications))
	}

	return notifications, nil
}

// ClearOfflineMessages 清除用户的离线消息
func (oms *OfflineMessageService) ClearOfflineMessages(userID int) error {
	key := fmt.Sprintf("offline_msg:%d", userID)
	ctx := context.Background()

	err := redis.GetClient().Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("清除离线消息失败: %w", err)
	}

	log.Printf("🗑️ 离线消息已清除: UserID=%d", userID)
	return nil
}

// GetOfflineMessageCount 获取离线消息数量
func (oms *OfflineMessageService) GetOfflineMessageCount(userID int) (int64, error) {
	key := fmt.Sprintf("offline_msg:%d", userID)
	ctx := context.Background()

	count, err := redis.GetClient().LLen(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("获取离线消息数量失败: %w", err)
	}

	return count, nil
}

// SendOfflineMessagesToUser 用户上线时发送离线消息
func (oms *OfflineMessageService) SendOfflineMessagesToUser(userID int) error {
	// 获取离线消息
	offlineMessages, err := oms.GetOfflineMessages(userID)
	if err != nil {
		return fmt.Errorf("获取离线消息失败: %w", err)
	}

	if len(offlineMessages) == 0 {
		log.Printf("用户 %d 没有离线消息", userID)
		return nil
	}

	// 发送离线消息
	wsService := GetWebSocketService()
	successCount := 0

	for _, message := range offlineMessages {
		// 使用WebSocket消息格式发送离线消息
		wsMsg := websocket.Message{
			Type:    websocket.SystemNotice, // 使用系统通知类型
			Content: message.Content,        // 离线消息内容已经包含"[离线消息]"前缀
			Data:    message.Data,           // 保持原有数据
			Time:    message.Time,           // 保持原有时间
		}

		// 序列化为WebSocket消息格式
		msgBytes, err := json.Marshal(wsMsg)
		if err != nil {
			log.Printf("序列化离线消息失败: UserID=%d, MessageID=%s, Error=%v",
				userID, message.MessageID, err)
			continue
		}

		// 直接发送给用户
		hub := wsService.GetHub()
		clients := hub.GetUserClients(userID)

		sent := false
		for _, client := range clients {
			select {
			case client.Send <- msgBytes:
				sent = true
				// 标记离线消息为已投递
				go wsService.markMessageAsDelivered(message.MessageID, userID)
				log.Printf("离线消息已发送给用户 %d: MessageID=%s", userID, message.MessageID)
				break
			default:
				// 客户端缓冲区已满，跳过
				continue
			}
		}

		if sent {
			successCount++
		} else {
			log.Printf("离线消息发送失败: UserID=%d, MessageID=%s", userID, message.MessageID)
		}
	}

	// 如果所有消息都发送成功，清除离线消息
	if successCount == len(offlineMessages) {
		err := oms.ClearOfflineMessages(userID)
		if err != nil {
			log.Printf("清除离线消息失败: %v", err)
		}
	} else if successCount > 0 {
		// 部分成功，保留未发送的消息
		log.Printf("部分离线消息发送成功: UserID=%d, 成功=%d, 总数=%d", userID, successCount, len(offlineMessages))
	}

	log.Printf("📱 离线消息发送完成: UserID=%d, 总数=%d, 成功=%d",
		userID, len(offlineMessages), successCount)

	return nil
}

// batchSendOfflineMessages 批量发送离线消息（优化版本）
func (oms *OfflineMessageService) batchSendOfflineMessages(userID int, messages []*NotificationMessage) error {
	wsService := GetWebSocketService()
	hub := wsService.GetHub()
	clients := hub.GetUserClients(userID)

	if len(clients) == 0 {
		log.Printf("用户 %d 没有活跃连接，无法发送离线消息", userID)
		return fmt.Errorf("用户没有活跃连接")
	}

	successCount := 0
	failedMessages := make([]*NotificationMessage, 0)

	// 批量处理消息
	for _, message := range messages {
		// 使用WebSocket消息格式发送离线消息
		wsMsg := websocket.Message{
			Type:    websocket.SystemNotice,
			Content: message.Content,
			Data:    message.Data,
			Time:    message.Time,
		}

		// 序列化为WebSocket消息格式
		msgBytes, err := json.Marshal(wsMsg)
		if err != nil {
			log.Printf("序列化离线消息失败: UserID=%d, MessageID=%s, Error=%v",
				userID, message.MessageID, err)
			failedMessages = append(failedMessages, message)
			continue
		}

		// 尝试发送给所有客户端
		sent := false
		for _, client := range clients {
			select {
			case client.Send <- msgBytes:
				sent = true
				// 异步标记消息为已投递
				go func(msgID string) {
					time.Sleep(100 * time.Millisecond) // 等待100ms确保记录创建完成
					wsService.markMessageAsDelivered(msgID, userID)
				}(message.MessageID)
				log.Printf("离线消息已发送给用户 %d: MessageID=%s", userID, message.MessageID)
				break
			default:
				// 客户端缓冲区已满，尝试下一个客户端
				continue
			}
		}

		if sent {
			successCount++
		} else {
			failedMessages = append(failedMessages, message)
			log.Printf("离线消息发送失败: UserID=%d, MessageID=%s", userID, message.MessageID)
		}
	}

	// 处理发送结果
	if successCount == len(messages) {
		// 全部成功，清除离线消息
		err := oms.ClearOfflineMessages(userID)
		if err != nil {
			log.Printf("清除离线消息失败: %v", err)
		}
		log.Printf("所有离线消息发送成功: UserID=%d, 总数=%d", userID, successCount)
	} else if successCount > 0 {
		// 部分成功，保留失败的消息
		log.Printf("部分离线消息发送成功: UserID=%d, 成功=%d, 失败=%d", userID, successCount, len(failedMessages))

		// 重新保存失败的消息
		if len(failedMessages) > 0 {
			err := oms.saveFailedMessages(userID, failedMessages)
			if err != nil {
				log.Printf("重新保存失败消息失败: %v", err)
			}
		}
	} else {
		// 全部失败
		log.Printf("所有离线消息发送失败: UserID=%d", userID)
		return fmt.Errorf("所有离线消息发送失败")
	}

	return nil
}

// saveFailedMessages 保存失败的消息
func (oms *OfflineMessageService) saveFailedMessages(userID int, messages []*NotificationMessage) error {
	ctx := context.Background()
	pipe := redis.GetClient().Pipeline()

	key := fmt.Sprintf("offline_msg:%d", userID)

	// 先清除现有消息
	pipe.Del(ctx, key)

	// 重新添加失败的消息
	for _, message := range messages {
		msgData, err := json.Marshal(message)
		if err != nil {
			continue
		}
		pipe.LPush(ctx, key, msgData)
	}

	// 限制消息数量并设置过期时间
	pipe.LTrim(ctx, key, 0, 99)
	pipe.Expire(ctx, key, 7*24*time.Hour)

	_, err := pipe.Exec(ctx)
	return err
}

// CleanupExpiredMessages 清理过期的离线消息
func (oms *OfflineMessageService) CleanupExpiredMessages() {
	// 这个函数可以定期运行，清理过期的离线消息
	// Redis的EXPIRE会自动处理，但我们可以添加额外的清理逻辑
	log.Println("🧹 离线消息清理完成")
}
