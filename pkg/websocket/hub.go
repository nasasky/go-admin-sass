package websocket

import (
	"log"
	"nasa-go-admin/middleware"
	"sync"
	"time"
)

// Hub 维护活跃客户端的集合并广播消息
type Hub struct {
	// 所有活跃的客户端
	Clients map[*Client]bool

	// 用户ID到客户端的映射，用于定向发送消息
	UserClients map[int][]*Client

	// 广播消息的通道
	Broadcast chan []byte

	// 注册请求
	Register chan *Client

	// 注销请求
	Unregister chan *Client

	// 互斥锁，保护UserClients
	mu sync.Mutex
}

// NewHub 创建一个新的Hub实例
func NewHub() *Hub {
	return &Hub{
		Broadcast:   make(chan []byte),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Clients:     make(map[*Client]bool),
		UserClients: make(map[int][]*Client),
	}
}

// Run 启动hub的消息处理循环
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
			// 将客户端添加到用户映射
			h.mu.Lock()
			h.UserClients[client.UserID] = append(h.UserClients[client.UserID], client)
			userConnCount := len(h.UserClients[client.UserID])
			h.mu.Unlock()

			totalConnections := len(h.Clients)
			uniqueUsers := len(h.UserClients)

			log.Printf("✅ 新客户端注册: UserID=%d, ConnectionID=%s, 该用户连接数=%d, 总连接数=%d, 在线用户数=%d",
				client.UserID, client.ConnectionID, userConnCount, totalConnections, uniqueUsers)

			// 记录详细连接统计
			if totalConnections%10 == 0 { // 每10个连接记录一次详细统计
				log.Printf("📊 连接统计: 总连接=%d, 独立用户=%d, 平均每用户连接=%.2f",
					totalConnections, uniqueUsers, float64(totalConnections)/float64(uniqueUsers))
			}

			// 异步处理用户上线后的离线消息
			go func(userID int) {
				// 延迟1秒确保连接稳定
				time.Sleep(1 * time.Second)

				// 这里需要通过某种方式获取OfflineMessageService
				// 由于架构限制，我们在WebSocket连接成功后在控制器中处理
				log.Printf("📱 用户 %d 上线，准备发送离线消息", userID)
			}(client.UserID)

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)

				// 从用户映射中移除客户端
				h.mu.Lock()
				clients := h.UserClients[client.UserID]
				for i, c := range clients {
					if c == client {
						h.UserClients[client.UserID] = append(clients[:i], clients[i+1:]...)
						break
					}
				}
				if len(h.UserClients[client.UserID]) == 0 {
					delete(h.UserClients, client.UserID)
				}
				h.mu.Unlock()

				totalConnections := len(h.Clients)
				uniqueUsers := len(h.UserClients)
				log.Printf("❌ 客户端注销: UserID=%d, ConnectionID=%s, 总连接数=%d, 在线用户数=%d",
					client.UserID, client.ConnectionID, totalConnections, uniqueUsers)
			}

		case message := <-h.Broadcast:
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)

					// 从用户映射中移除客户端
					h.mu.Lock()
					clients := h.UserClients[client.UserID]
					for i, c := range clients {
						if c == client {
							h.UserClients[client.UserID] = append(clients[:i], clients[i+1:]...)
							break
						}
					}
					if len(h.UserClients[client.UserID]) == 0 {
						delete(h.UserClients, client.UserID)
					}
					h.mu.Unlock()
				}
			}
		}
	}
}

// SendToUser 向特定用户发送消息
func (h *Hub) SendToUser(userID int, message []byte) {
	h.mu.Lock()
	clients := h.UserClients[userID]
	h.mu.Unlock()

	for _, client := range clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(h.Clients, client)

			// 从用户映射中移除客户端
			h.mu.Lock()
			clients := h.UserClients[client.UserID]
			for i, c := range clients {
				if c == client {
					h.UserClients[client.UserID] = append(clients[:i], clients[i+1:]...)
					break
				}
			}
			if len(h.UserClients[client.UserID]) == 0 {
				delete(h.UserClients, client.UserID)
			}
			h.mu.Unlock()
		}
	}
	middleware.LogWebSocketEvent("message_sent", userID, "", map[string]interface{}{
		"message_length": len(message),
		"message_type":   "notification",
	})

	log.Printf("消息已发送给用户 %d, 接收客户端数量: %d", userID, len(clients))
}

// GetStats 获取Hub统计信息
func (h *Hub) GetStats() map[string]interface{} {
	h.mu.Lock()
	defer h.mu.Unlock()

	totalConnections := len(h.Clients)
	uniqueUsers := len(h.UserClients)

	// 计算每个用户的连接数分布
	var maxUserConnections int
	var totalUserConnections int
	for _, clients := range h.UserClients {
		connCount := len(clients)
		totalUserConnections += connCount
		if connCount > maxUserConnections {
			maxUserConnections = connCount
		}
	}

	avgConnectionsPerUser := 0.0
	if uniqueUsers > 0 {
		avgConnectionsPerUser = float64(totalUserConnections) / float64(uniqueUsers)
	}

	return map[string]interface{}{
		"total_connections":        totalConnections,
		"unique_users":             uniqueUsers,
		"max_user_connections":     maxUserConnections,
		"avg_connections_per_user": avgConnectionsPerUser,
	}
}
