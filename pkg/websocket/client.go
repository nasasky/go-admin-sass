package websocket

import (
	"encoding/json"
	"log"
	"nasa-go-admin/middleware"
	"nasa-go-admin/services/admin_service"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// 全局统计信息
var (
	totalMessages     int64 // 消息计数
	activeConnections int64 // 活跃连接数
	totalConnections  int64 // 总连接数
)

const (
	// 向客户端写入消息的最大时间
	writeWait = 30 * time.Second // 增加到30秒
	PongWait  = 60 * time.Second
	// 客户端读取下一个pong消息的最大时间
	pongWait = 60 * time.Second

	// 向客户端发送ping的周期，必须小于pongWait
	pingPeriod = (pongWait * 9) / 10

	// 最大消息大小
	maxMessageSize = 512
)

// Client 表示WebSocket连接的客户端
type Client struct {
	Hub          *Hub
	Conn         *websocket.Conn
	Send         chan []byte
	UserID       int // 用户ID，用于标识客户端
	ConnectionID string
}

// ReadPump 从WebSocket连接读取消息并转发到hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		// 发送正常关闭消息
		closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "连接关闭")
		err := c.Conn.WriteControl(websocket.CloseMessage, closeMsg, time.Now().Add(time.Second))
		if err != nil {
			log.Printf("发送关闭消息失败: %v", err)
		}
		err = c.Conn.Close()
		if err != nil {
			log.Printf("关闭连接失败: %v", err)
		}
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	err := c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		log.Printf("设置读取超时失败: %v", err)
		return
	}

	c.Conn.SetPongHandler(func(string) error {
		err := c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			log.Printf("更新读取超时失败: %v", err)
			return err
		}
		// 发送pong响应
		response := map[string]interface{}{
			"type": "pong",
			"code": 200,
			"data": map[string]interface{}{
				"message": "心跳响应",
				"user_id": c.UserID,
				"conn_id": c.ConnectionID,
				"status":  "connected",
			},
		}
		responseBytes, err := json.Marshal(response)
		if err == nil {
			err = c.Conn.WriteMessage(websocket.TextMessage, responseBytes)
			if err != nil {
				log.Printf("发送pong响应失败: %v", err)
			}
		}
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket读取错误: %v", err)
			}
			break
		} else {
			atomic.AddInt64(&totalMessages, 1)
			// 记录收到的消息
			middleware.LogWebSocketEvent("message_received", c.UserID, c.ConnectionID, map[string]interface{}{
				"message_length": len(message),
				"message_type":   "text",
			})
		}

		// 处理客户端消息
		c.handleClientMessage(message)
	}
}

// WritePump 将消息从hub转发到WebSocket连接
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		err := c.Conn.Close()
		if err != nil {
			log.Printf("关闭连接失败: %v", err)
		}
	}()

	for {
		select {
		case message, ok := <-c.Send:
			err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				log.Printf("设置写入超时失败: %v", err)
				return
			}
			if !ok {
				// Hub关闭了channel
				err := c.Conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, "服务关闭"))
				if err != nil {
					log.Printf("发送关闭消息失败: %v", err)
				}
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("获取写入器失败: %v", err)
				return
			}
			_, err = w.Write(message)
			if err != nil {
				log.Printf("写入消息失败: %v", err)
				return
			}

			// 添加队列中的所有消息到当前WebSocket消息
			n := len(c.Send)
			for i := 0; i < n; i++ {
				_, err = w.Write(<-c.Send)
				if err != nil {
					log.Printf("写入队列消息失败: %v", err)
					return
				}
			}

			if err := w.Close(); err != nil {
				log.Printf("关闭写入器失败: %v", err)
				return
			}
		case <-ticker.C:
			err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				log.Printf("设置ping写入超时失败: %v", err)
				return
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("发送ping消息失败: %v", err)
				return
			}
		}
	}
}

// handleClientMessage 处理客户端发送的消息
func (c *Client) handleClientMessage(message []byte) {
	// 尝试解析消息
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("解析客户端消息失败: %v", err)
		return
	}

	// 根据消息类型处理
	msgType, ok := msg["type"].(string)
	if !ok {
		log.Printf("消息类型字段缺失")
		return
	}

	switch msgType {
	case "message_received":
		// 客户端确认收到消息
		messageID, ok := msg["message_id"].(string)
		if !ok {
			log.Printf("消息确认缺少message_id")
			return
		}

		// 标记消息为已接收
		recordService := admin_service.NewNotificationRecordService()
		updates := map[string]interface{}{
			"is_received":     true,
			"received_at":     time.Now().Format("2006-01-02 15:04:05"),
			"delivery_status": "delivered",
		}

		err := recordService.UpdateAdminUserReceiveRecord(messageID, c.UserID, updates)
		if err != nil {
			log.Printf("标记消息投递状态失败: MessageID=%s, UserID=%d, Error=%v",
				messageID, c.UserID, err)
		} else {
			log.Printf("用户 %d 确认收到消息: %s，已更新接收状态", c.UserID, messageID)
		}
	case "ping":
		// 客户端ping，回复pong
		response := map[string]interface{}{
			"type": "pong",
			"code": 200,
			"data": map[string]interface{}{
				"message": "心跳响应",
				"user_id": c.UserID,
				"conn_id": c.ConnectionID,
				"status":  "connected",
			},
		}
		responseBytes, err := json.Marshal(response)
		if err == nil {
			err = c.Conn.WriteMessage(websocket.TextMessage, responseBytes)
			if err != nil {
				log.Printf("发送pong响应失败: %v", err)
			}
		}
	default:
		// 其他消息类型，广播给所有客户端
		c.Hub.Broadcast <- message
	}
}
