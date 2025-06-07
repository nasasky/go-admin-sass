package websocket

import (
	"log"
	"nasa-go-admin/middleware"
	"sync"
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
			h.mu.Unlock()
			log.Printf("新客户端注册: UserID=%d, 当前连接数: %d", client.UserID, len(h.Clients))

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

				log.Printf("客户端注销: UserID=%d, 当前连接数: %d", client.UserID, len(h.Clients))
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
