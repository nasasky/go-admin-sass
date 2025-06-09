package middleware

import (
	"context"
	"fmt"
	"log"
	"nasa-go-admin/redis"
	"nasa-go-admin/utils"
)

// LogWechatEvent 记录WebSocket事件
func LogWechatEvent(eventType string, userID int, connectionID string, data map[string]interface{}) {
	// 获取MongoDB集合
	collection := GetMongoCollection("wechat_log_db", "news_logs")

	// 准备事件日志 - 使用UTC时间
	timestamp := utils.GetCurrentTimeForMongo()

	// 创建事件日志
	eventLog := map[string]interface{}{
		"event_type":    eventType,
		"user_id":       userID,
		"connection_id": connectionID,
		"timestamp":     timestamp,
		"data":          data,
	}

	// 尝试获取用户名
	userIDStr := fmt.Sprintf("%v", userID)
	if userInfo, err := redis.GetUserInfo(userIDStr); err == nil {
		eventLog["username"] = userInfo["username"]
	}

	// 保存事件日志
	_, err := collection.InsertOne(context.Background(), eventLog)
	if err != nil {
		log.Printf("Failed to insert WebSocket event log: %v", err)
	}
}
