package public_service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
		return nil
	}

	// 发送离线消息
	wsService := GetWebSocketService()
	successCount := 0

	for _, message := range offlineMessages {
		// 修改消息类型为离线消息
		message.Type = "offline_message"

		err := wsService.SendNotification(message)
		if err != nil {
			log.Printf("发送离线消息失败: UserID=%d, MessageID=%s, Error=%v",
				userID, message.MessageID, err)
			continue
		}
		successCount++
	}

	// 如果所有消息都发送成功，清除离线消息
	if successCount == len(offlineMessages) {
		err := oms.ClearOfflineMessages(userID)
		if err != nil {
			log.Printf("清除离线消息失败: %v", err)
		}
	}

	log.Printf("📱 离线消息发送完成: UserID=%d, 总数=%d, 成功=%d",
		userID, len(offlineMessages), successCount)

	return nil
}

// CleanupExpiredMessages 清理过期的离线消息
func (oms *OfflineMessageService) CleanupExpiredMessages() {
	// 这个函数可以定期运行，清理过期的离线消息
	// Redis的EXPIRE会自动处理，但我们可以添加额外的清理逻辑
	log.Println("🧹 离线消息清理完成")
}
