package websocket

import (
	"log"
	"nasa-go-admin/middleware"
	"sync"
	"sync/atomic"
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

	// 读写锁，保护UserClients - 优化：使用读写锁提升并发性能
	mu sync.RWMutex

	// 用户离线回调函数
	onUserOffline func(userID int, connectionID string)

	// 性能统计
	stats struct {
		totalConnections int64
		uniqueUsers      int64
		messageCount     int64
	}
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

// SetUserOfflineCallback 设置用户离线回调函数
func (h *Hub) SetUserOfflineCallback(callback func(userID int, connectionID string)) {
	h.onUserOffline = callback
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
			// 更新统计信息
			atomic.AddInt64(&h.stats.totalConnections, 1)
			if userConnCount == 1 {
				atomic.AddInt64(&h.stats.uniqueUsers, 1)
			}
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
					atomic.AddInt64(&h.stats.uniqueUsers, -1)
				}
				atomic.AddInt64(&h.stats.totalConnections, -1)
				h.mu.Unlock()

				totalConnections := len(h.Clients)
				uniqueUsers := len(h.UserClients)
				log.Printf("❌ 客户端注销: UserID=%d, ConnectionID=%s, 总连接数=%d, 在线用户数=%d",
					client.UserID, client.ConnectionID, totalConnections, uniqueUsers)

				// 如果用户没有其他连接，调用用户离线回调
				if len(h.UserClients[client.UserID]) == 0 {
					if h.onUserOffline != nil {
						h.onUserOffline(client.UserID, client.ConnectionID)
					}
					log.Printf("用户 %d 已完全离线", client.UserID)
				}
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

// GetUserClients 获取用户的客户端列表（线程安全）
func (h *Hub) GetUserClients(userID int) []*Client {
	h.mu.RLock() // 优化：使用读锁提升并发性能
	defer h.mu.RUnlock()

	clients := h.UserClients[userID]
	if len(clients) == 0 {
		return nil
	}
	// 返回副本以避免竞态条件
	result := make([]*Client, len(clients))
	copy(result, clients)
	return result
}

// IsUserOnline 检查用户是否在线
func (h *Hub) IsUserOnline(userID int) bool {
	h.mu.RLock() // 优化：使用读锁提升并发性能
	defer h.mu.RUnlock()

	clients, exists := h.UserClients[userID]
	return exists && len(clients) > 0
}

// RemoveClient 移除客户端（线程安全）
func (h *Hub) RemoveClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 从Clients中移除
	delete(h.Clients, client)

	// 从UserClients中移除
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
}

// GetOnlineUserIDs 获取在线用户ID列表
func (h *Hub) GetOnlineUserIDs() []int {
	h.mu.RLock() // 优化：使用读锁提升并发性能
	defer h.mu.RUnlock()

	var userIDs []int
	for userID := range h.UserClients {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

// SendToUsers 批量向多个用户发送消息（优化版本）
func (h *Hub) SendToUsers(userIDs []int, message []byte) {
	h.mu.RLock() // 使用读锁，避免阻塞其他读取操作
	defer h.mu.RUnlock()

	successCount := 0
	totalTargets := len(userIDs)

	for _, userID := range userIDs {
		clients := h.UserClients[userID]
		if len(clients) == 0 {
			continue
		}

		for _, client := range clients {
			select {
			case client.Send <- message:
				successCount++
			default:
				// 客户端缓冲区已满，异步处理
				go h.handleFullBuffer(client)
			}
		}
	}

	atomic.AddInt64(&h.stats.messageCount, int64(successCount))
	log.Printf("批量消息发送完成: 目标用户=%d, 成功发送=%d", totalTargets, successCount)
}

// handleFullBuffer 处理客户端缓冲区满的情况
func (h *Hub) handleFullBuffer(client *Client) {
	log.Printf("客户端缓冲区已满，关闭连接: UserID=%d, ConnectionID=%s", client.UserID, client.ConnectionID)

	// 关闭客户端连接
	close(client.Send)
	h.RemoveClient(client)
}

// GetDetailedStats 获取详细的统计信息
func (h *Hub) GetDetailedStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

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
		"atomic_total_connections": atomic.LoadInt64(&h.stats.totalConnections),
		"atomic_unique_users":      atomic.LoadInt64(&h.stats.uniqueUsers),
		"message_count":            atomic.LoadInt64(&h.stats.messageCount),
	}
}
