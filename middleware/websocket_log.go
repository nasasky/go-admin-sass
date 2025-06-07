package middleware

import (
	"context"
	"fmt"
	"log"
	"nasa-go-admin/redis"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

// WebSocketLogger 记录WebSocket连接事件
func WebSocketLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果是WebSocket升级请求
		if strings.Contains(c.GetHeader("Connection"), "Upgrade") &&
			strings.Contains(c.GetHeader("Upgrade"), "websocket") {

			// 获取MongoDB集合
			collection := GetMongoCollection("websocket_log_db", "connection_logs")

			// 记录连接建立信息
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			clientIP := c.ClientIP()

			// 从请求中提取信息
			path := c.Request.URL.Path
			query := c.Request.URL.RawQuery
			userAgent := c.Request.UserAgent()

			// 尝试获取用户ID
			var userID interface{} = nil
			if uid, exists := c.Get("uid"); exists {
				userID = uid
			} else {
				// 尝试从query参数中获取user_id
				if userIDStr := c.Query("user_id"); userIDStr != "" {
					userID = userIDStr
				}
				// 尝试从token参数解析
				if tokenStr := c.Query("token"); tokenStr != "" && userID == nil {
					// 使用JWT解析函数
					if parsedUID, err := ParseTokenGetUID(tokenStr); err == nil {
						userID = parsedUID
					}
				}
			}

			// 获取用户信息
			var username string = "unknown"
			if userID != nil {
				userIDStr := fmt.Sprintf("%v", userID)
				if userInfo, err := redis.GetUserInfo(userIDStr); err == nil {
					username = userInfo["username"]
				}
			}

			// 创建连接日志
			connectionLog := map[string]interface{}{
				"event_type":    "connection_attempt",
				"user_id":       userID,
				"username":      username,
				"client_ip":     clientIP,
				"path":          path,
				"query":         query,
				"user_agent":    userAgent,
				"timestamp":     timestamp,
				"connection_id": generateConnectionID(),
				"status":        "pending", // 初始状态为pending
			}

			// 将连接ID保存到上下文中，供后续使用
			connectionID := connectionLog["connection_id"].(string)
			c.Set("ws_connection_id", connectionID)

			// 保存连接尝试日志
			_, err := collection.InsertOne(context.Background(), connectionLog)
			if err != nil {
				log.Printf("Failed to insert WebSocket connection log: %v", err)
			}

			// 在请求结束后更新连接状态
			c.Next()

			// 检查响应状态，确定连接是否成功
			status := "failed"
			if c.Writer.Status() == 101 { // 101 Switching Protocols 表示WebSocket连接成功
				status = "established"
			}

			// 更新连接状态
			filter := bson.M{"connection_id": connectionID}
			update := bson.M{
				"$set": bson.M{
					"status":          status,
					"response_status": c.Writer.Status(),
					"response_time":   time.Now().Format("2006-01-02 15:04:05"),
				},
			}

			_, err = collection.UpdateOne(context.Background(), filter, update)
			if err != nil {
				log.Printf("Failed to update WebSocket connection status: %v", err)
			}
		} else {
			// 非WebSocket请求，继续处理
			c.Next()
		}
	}
}

// LogWebSocketEvent 记录WebSocket事件
func LogWebSocketEvent(eventType string, userID int, connectionID string, data map[string]interface{}) {
	// 获取MongoDB集合
	collection := GetMongoCollection("websocket_log_db", "event_logs")

	// 准备事件日志
	timestamp := time.Now().Format("2006-01-02 15:04:05")

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

// LogWebSocketDisconnect 记录WebSocket断开连接
func LogWebSocketDisconnect(userID int, connectionID string, reason string) {
	// 获取MongoDB集合
	collection := GetMongoCollection("websocket_log_db", "connection_logs")

	// 更新连接状态
	filter := bson.M{"connection_id": connectionID}
	update := bson.M{
		"$set": bson.M{
			"status":            "disconnected",
			"disconnect_time":   time.Now().Format("2006-01-02 15:04:05"),
			"disconnect_reason": reason,
		},
	}

	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("Failed to update WebSocket disconnect status: %v", err)
	}

	// 同时记录一个断开事件
	LogWebSocketEvent("disconnection", userID, connectionID, map[string]interface{}{
		"reason": reason,
	})
}

// generateConnectionID 生成唯一的连接ID
func generateConnectionID() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(),
		strings.ReplaceAll(uuid.New().String(), "-", "")[:8])
}
