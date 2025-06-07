package public

import (
	"context"
	"fmt"
	"log"
	"nasa-go-admin/middleware"
	"nasa-go-admin/services/public_service"
	"net/http"
	"runtime/debug"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	ws "nasa-go-admin/pkg/websocket"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// 全局统计信息
var (
	activeConnections int64 // 原子计数器 - 当前活跃连接数
	totalConnections  int64 // 总连接数统计
	totalMessages     int64 // 总消息数统计
)

// 连接限制配置
const (
	maxConnPerIP   = 10              // 每IP最大连接数
	connTimeout    = 5 * time.Second // 连接超时
	maxMessageSize = 4096            // 最大消息大小
	writeWait      = 3 * time.Second // 写入超时
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 生产环境应该更严格地检查Origin
		return true
	},
	// 启用压缩
	EnableCompression: true,
}

// WebSocketConnect 处理WebSocket连接请求
func WebSocketConnect(c *gin.Context) {
	// 捕获可能的panic
	defer func() {
		if r := recover(); r != nil {
			log.Printf("WebSocket连接处理panic: %v\n%s", r, debug.Stack())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		}
	}()

	// 添加详细的日志记录
	log.Printf("收到WebSocket连接请求: %s, token=%v", c.Request.URL.Path, c.Query("token") != "")

	// 连接数限制 - 全局级别
	currentConns := atomic.LoadInt64(&activeConnections)
	if currentConns > 10000 { // 全局最大连接数限制
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "服务器连接数已达上限"})
		return
	}

	// IP限流
	clientIP := c.ClientIP()
	ipConns, ipLimited := checkIPConnectionLimit(clientIP)
	if ipLimited {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": fmt.Sprintf("IP连接数超限: %d/%d", ipConns, maxConnPerIP)})
		return
	}

	// 设置连接超时
	c.Request.Context().Done()
	_, cancel := context.WithTimeout(c.Request.Context(), connTimeout)
	defer cancel()

	// 改进用户身份验证流程
	userID, err := getUserIDFromContext(c)
	if err != nil {
		log.Printf("WebSocket认证失败: %v, 请求路径: %s", err, c.Request.URL.Path)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":  "用户认证失败",
			"detail": err.Error(),
		})
		return
	}

	// 从上下文中获取连接ID (添加安全检查)
	connectionID, exists := c.Get("ws_connection_id")
	var connID string
	if exists && connectionID != nil {
		// 使用类型断言的安全形式
		if id, ok := connectionID.(string); ok {
			connID = id
		} else {
			// 类型不匹配，生成一个默认ID
			connID = fmt.Sprintf("conn-%d-%d", userID, time.Now().UnixNano())
			log.Printf("连接ID类型不匹配，使用生成的ID: %s", connID)
		}
	} else {
		// 未设置连接ID，生成一个新的
		connID = fmt.Sprintf("conn-%d-%d", userID, time.Now().UnixNano())
		log.Printf("未设置ws_connection_id，使用生成的ID: %s", connID)
	}

	// 记录成功获取用户ID
	log.Printf("WebSocket连接: 成功获取用户ID=%d, 连接ID=%s", userID, connID)

	// 获取WebSocket服务
	wsService := public_service.GetWebSocketService()
	hub := wsService.GetHub()

	// 升级HTTP连接为WebSocket连接 (使用带超时的上下文)
	upgrader.HandshakeTimeout = 3 * time.Second
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		middleware.LogWebSocketEvent("upgrade_failed", userID, connID, map[string]interface{}{
			"error": err.Error(),
		})
		log.Printf("升级WebSocket连接失败: %v", err)
		return
	}

	// 统计和监控
	atomic.AddInt64(&activeConnections, 1)
	atomic.AddInt64(&totalConnections, 1)
	incrementIPCounter(clientIP)

	// 记录连接成功事件 (异步)
	go middleware.LogWebSocketEvent("connected", userID, connID, map[string]interface{}{
		"remote_addr": conn.RemoteAddr().String(),
	})

	// 优化配置
	conn.SetReadLimit(maxMessageSize)
	err = conn.SetReadDeadline(time.Now().Add(ws.PongWait))
	if err != nil {
		return
	}
	conn.SetPongHandler(func(string) error {
		err := conn.SetReadDeadline(time.Now().Add(ws.PongWait))
		if err != nil {
			return err
		}
		return nil
	})

	// 创建客户端
	client := &ws.Client{
		Hub:          hub,
		Conn:         conn,
		Send:         make(chan []byte, 256),
		UserID:       userID,
		ConnectionID: connID,
	}

	// 将客户端注册到hub (可能是阻塞操作，考虑超时处理)
	select {
	case hub.Register <- client:
		// 注册成功
	case <-time.After(2 * time.Second):
		err := conn.Close()
		if err != nil {
			return
		}
		log.Printf("客户端注册到Hub超时: userID=%d", userID)
		atomic.AddInt64(&activeConnections, -1)
		decrementIPCounter(clientIP)
		return
	}

	// 启动goroutine处理消息读写 (使用恢复机制防止panic)
	go safeGoroutine(func() {
		client.WritePump()
	})
	go safeGoroutine(func() {
		client.ReadPump()
		// 当ReadPump退出时，记录断开连接
		middleware.LogWebSocketDisconnect(userID, connID, "client_closed")
		atomic.AddInt64(&activeConnections, -1)
		decrementIPCounter(clientIP)
	})

	// 发送欢迎消息 (使用非阻塞发送，避免死锁)
	welcomeMsg, _ := ws.NewSystemNotice("WebSocket连接已建立")
	select {
	case client.Send <- welcomeMsg:
		// 消息已入队
	default:
		// 缓冲区已满，跳过欢迎消息
	}
}

// 安全启动 goroutine，包含恢复机制
func safeGoroutine(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("WebSocket goroutine panic: %v", r)
				// 这里可以添加堆栈追踪
				debug.PrintStack()
			}
		}()
		fn()
	}()
}

// IP连接限制管理
var (
	ipConnections = make(map[string]int)
	ipMutex       sync.RWMutex
)

func checkIPConnectionLimit(ip string) (int, bool) {
	ipMutex.RLock()
	count := ipConnections[ip]
	ipMutex.RUnlock()
	return count, count >= maxConnPerIP
}

func incrementIPCounter(ip string) {
	ipMutex.Lock()
	ipConnections[ip]++
	ipMutex.Unlock()
}

func decrementIPCounter(ip string) {
	ipMutex.Lock()
	if ipConnections[ip] > 0 {
		ipConnections[ip]--
	}
	ipMutex.Unlock()
}

// 从上下文中获取用户ID - 优化版本
func getUserIDFromContext(c *gin.Context) (int, error) {
	// 1. 快速路径 - 从上下文获取
	if uid, exists := c.Get("uid"); exists {
		if id, ok := uid.(int); ok {
			return id, nil
		}
	}

	// 2. 直接从查询参数获取
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		return strconv.Atoi(userIDStr)
	}

	// 3. 从token解析
	if tokenString := c.Query("token"); tokenString != "" {
		// 使用缓存层加速重复token的解析
		return middleware.ParseTokenGetUID(tokenString)
	}

	return 0, fmt.Errorf("无法获取用户ID")
}

// WebSocketStats 对外暴露的监控端点
func WebSocketStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"active_connections": atomic.LoadInt64(&activeConnections),
		"total_connections":  atomic.LoadInt64(&totalConnections),
		"total_messages":     atomic.LoadInt64(&totalMessages),
		"ip_connections":     getIPConnectionCounts(),
	})
}

func getIPConnectionCounts() map[string]int {
	ipMutex.RLock()
	defer ipMutex.RUnlock()

	result := make(map[string]int, len(ipConnections))
	for ip, count := range ipConnections {
		result[ip] = count
	}
	return result
}
