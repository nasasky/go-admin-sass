package websocket

import (
	"log"
	"nasa-go-admin/middleware"
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
	writeWait = 10 * time.Second
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
		err := c.Conn.Close()
		if err != nil {
			return
		}
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	err := c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		return
	}
	c.Conn.SetPongHandler(func(string) error {
		err := c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			return err
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
				"message_type":   "text", // 或其他类型
			})
		}

		// 将消息发送到Hub进行广播
		c.Hub.Broadcast <- message
	}
}

// WritePump 将消息从hub转发到WebSocket连接
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		err := c.Conn.Close()
		if err != nil {
			return
		}
	}()

	for {
		select {
		case message, ok := <-c.Send:
			err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				return
			}
			if !ok {
				// Hub关闭了channel
				err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					return
				}
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 添加队列中的所有消息到当前WebSocket消息
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				return
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
