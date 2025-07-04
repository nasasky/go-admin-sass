# 管理端系统消息推送功能全面优化计划

## 🎯 优化目标
- 提高WebSocket连接响应速度
- 优化消息推送性能
- 改善离线消息处理效率
- 增强系统稳定性和可扩展性

## 📊 当前系统分析

### 1. WebSocket服务架构
- **Hub管理**: 单线程处理注册/注销，可能存在瓶颈
- **工作线程池**: 40个线程（CPU核心数×4），可能过多
- **消息队列**: 50000条缓冲，合理
- **连接管理**: 使用互斥锁，在高并发下可能成为瓶颈

### 2. 数据库操作
- **MongoDB**: 接收记录存储，查询频繁
- **MySQL**: 用户信息查询，缓存不足
- **Redis**: 离线消息存储，性能良好

### 3. 消息流程
- 系统公告 → 创建接收记录 → 发送WebSocket消息 → 保存离线消息
- 用户上线 → 注册连接 → 发送离线消息 → 更新状态

## 🚀 优化方案

### 1. WebSocket Hub优化

#### 1.1 连接管理优化
```go
// 使用读写锁替代互斥锁
type Hub struct {
    mu sync.RWMutex  // 改为读写锁
    // ... 其他字段
}

// 优化用户客户端获取
func (h *Hub) GetUserClients(userID int) []*Client {
    h.mu.RLock()  // 使用读锁
    defer h.mu.RUnlock()
    
    clients := h.UserClients[userID]
    if len(clients) == 0 {
        return nil
    }
    
    // 返回副本避免竞态条件
    result := make([]*Client, len(clients))
    copy(result, clients)
    return result
}
```

#### 1.2 消息发送优化
```go
// 批量发送优化
func (h *Hub) SendToUsers(userIDs []int, message []byte) {
    h.mu.RLock()
    defer h.mu.RUnlock()
    
    for _, userID := range userIDs {
        clients := h.UserClients[userID]
        for _, client := range clients {
            select {
            case client.Send <- message:
            default:
                // 异步处理满缓冲区
                go h.handleFullBuffer(client)
            }
        }
    }
}
```

### 2. 工作线程池优化

#### 2.1 动态线程池
```go
// 根据负载动态调整线程数
type WebSocketService struct {
    workerCount       int
    minWorkerCount    int
    maxWorkerCount    int
    workerLoadMonitor *WorkerLoadMonitor
}

// 负载监控
type WorkerLoadMonitor struct {
    queueSize    int64
    avgProcessTime time.Duration
    activeWorkers int64
}
```

#### 2.2 任务优先级
```go
type SendTask struct {
    Priority    int           // 优先级：1-高，2-中，3-低
    Message     *NotificationMessage
    UserIDs     []int
    Response    chan<- error
    CreatedAt   time.Time
}

// 优先级队列
type PriorityQueue struct {
    highPriority   chan *SendTask
    normalPriority chan *SendTask
    lowPriority    chan *SendTask
}
```

### 3. 数据库优化

#### 3.1 MongoDB索引优化
```javascript
// 接收记录索引
db.admin_user_receive_records.createIndex({"user_id": 1, "created_at": -1})
db.admin_user_receive_records.createIndex({"message_id": 1, "user_id": 1})
db.admin_user_receive_records.createIndex({"is_online": 1, "delivery_status": 1})

// 在线状态索引
db.admin_user_online_status.createIndex({"user_id": 1})
db.admin_user_online_status.createIndex({"is_online": 1, "last_seen": -1})
```

#### 3.2 查询优化
```go
// 批量查询优化
func (s *NotificationRecordService) BatchGetUserReceiveRecords(userIDs []int) (map[int][]AdminUserReceiveRecord, error) {
    collection := mongodb.GetCollection("notification_log_db", "admin_user_receive_records")
    
    filter := bson.M{"user_id": bson.M{"$in": userIDs}}
    cursor, err := collection.Find(ctx, filter)
    if err != nil {
        return nil, err
    }
    
    // 按用户ID分组
    result := make(map[int][]AdminUserReceiveRecord)
    // ... 处理结果
    return result, nil
}
```

### 4. 缓存优化

#### 4.1 用户信息缓存
```go
type UserCache struct {
    cache    *sync.Map
    ttl      time.Duration
    maxSize  int
}

func (uc *UserCache) GetUserInfo(userID int) (*UserInfo, error) {
    if cached, ok := uc.cache.Load(userID); ok {
        if userInfo, ok := cached.(*UserInfo); ok && !userInfo.IsExpired() {
            return userInfo, nil
        }
    }
    
    // 从数据库加载
    userInfo, err := uc.loadFromDB(userID)
    if err != nil {
        return nil, err
    }
    
    uc.cache.Store(userID, userInfo)
    return userInfo, nil
}
```

#### 4.2 在线状态缓存
```go
type OnlineStatusCache struct {
    cache *sync.Map
    ttl   time.Duration
}

func (osc *OnlineStatusCache) IsUserOnline(userID int) bool {
    if status, ok := osc.cache.Load(userID); ok {
        if onlineStatus, ok := status.(*OnlineStatus); ok && !onlineStatus.IsExpired() {
            return onlineStatus.IsOnline
        }
    }
    return false
}
```

### 5. 消息发送优化

#### 5.1 异步消息处理
```go
// 异步发送消息，不阻塞主流程
func (s *WebSocketService) SendNotificationAsync(msg *NotificationMessage) {
    go func() {
        err := s.SendNotification(msg)
        if err != nil {
            log.Printf("异步发送消息失败: %v", err)
        }
    }()
}
```

#### 5.2 消息去重
```go
type MessageDeduplicator struct {
    cache *sync.Map
    ttl   time.Duration
}

func (md *MessageDeduplicator) IsDuplicate(messageID string, userID int) bool {
    key := fmt.Sprintf("%s:%d", messageID, userID)
    if _, exists := md.cache.Load(key); exists {
        return true
    }
    
    md.cache.Store(key, time.Now())
    return false
}
```

### 6. 离线消息优化

#### 6.1 批量处理
```go
// 批量发送离线消息
func (oms *OfflineMessageService) SendOfflineMessagesToUsers(userIDs []int) error {
    // 批量获取所有用户的离线消息
    allMessages := make(map[int][]*NotificationMessage)
    
    for _, userID := range userIDs {
        messages, err := oms.GetOfflineMessages(userID)
        if err != nil {
            continue
        }
        if len(messages) > 0 {
            allMessages[userID] = messages
        }
    }
    
    // 批量发送
    return oms.batchSendMessages(allMessages)
}
```

#### 6.2 Redis Pipeline优化
```go
func (oms *OfflineMessageService) SaveOfflineMessagesBatch(messages map[int][]*NotificationMessage) error {
    ctx := context.Background()
    pipe := redis.GetClient().Pipeline()
    
    for userID, userMessages := range messages {
        key := fmt.Sprintf("offline_msg:%d", userID)
        
        for _, message := range userMessages {
            msgData, _ := json.Marshal(message)
            pipe.LPush(ctx, key, msgData)
        }
        
        pipe.LTrim(ctx, key, 0, 99)
        pipe.Expire(ctx, key, 7*24*time.Hour)
    }
    
    _, err := pipe.Exec(ctx)
    return err
}
```

### 7. 监控和指标

#### 7.1 性能指标
```go
type WebSocketMetrics struct {
    ActiveConnections    int64
    TotalMessages       int64
    MessageLatency      time.Duration
    QueueSize           int
    WorkerUtilization   float64
    ErrorRate           float64
}

func (s *WebSocketService) GetMetrics() *WebSocketMetrics {
    return &WebSocketMetrics{
        ActiveConnections:  atomic.LoadInt64(&s.activeConnections),
        TotalMessages:     atomic.LoadInt64(&s.totalMessages),
        MessageLatency:    s.calculateAverageLatency(),
        QueueSize:         len(s.messageQueue),
        WorkerUtilization: s.calculateWorkerUtilization(),
        ErrorRate:         s.calculateErrorRate(),
    }
}
```

#### 7.2 健康检查
```go
func (s *WebSocketService) HealthCheck() map[string]interface{} {
    return map[string]interface{}{
        "status":              "healthy",
        "active_connections":  atomic.LoadInt64(&s.activeConnections),
        "queue_size":          len(s.messageQueue),
        "worker_count":        s.workerCount,
        "last_error":          s.getLastError(),
        "uptime":             time.Since(s.startTime).String(),
    }
}
```

## 📈 预期性能提升

### 1. 响应时间优化
- WebSocket连接建立：< 100ms
- 消息发送延迟：< 50ms
- 离线消息推送：< 200ms

### 2. 吞吐量提升
- 并发连接数：10,000+
- 消息处理能力：10,000 msg/s
- 离线消息处理：1,000 msg/s

### 3. 资源利用率
- CPU使用率：< 70%
- 内存使用：< 2GB
- 网络带宽：< 100Mbps

## 🔧 实施步骤

### 阶段1：基础优化（1-2天）
1. 优化Hub连接管理
2. 调整工作线程池配置
3. 添加基础监控

### 阶段2：缓存优化（2-3天）
1. 实现用户信息缓存
2. 优化在线状态缓存
3. 添加消息去重

### 阶段3：数据库优化（1-2天）
1. 添加MongoDB索引
2. 优化查询语句
3. 实现批量操作

### 阶段4：高级优化（3-5天）
1. 实现动态线程池
2. 添加优先级队列
3. 完善监控系统

## 🧪 测试验证

### 1. 压力测试
```bash
# 使用WebSocket压力测试工具
websocket-bench -c 1000 -n 10000 ws://localhost:8801/ws
```

### 2. 性能测试
```bash
# 测试消息发送性能
curl -X POST http://localhost:8801/api/admin/system/notice \
  -H "Content-Type: application/json" \
  -d '{"content":"性能测试消息","target":"all"}'
```

### 3. 监控验证
```bash
# 检查系统指标
curl http://localhost:8801/api/admin/websocket/stats
```

## 📋 检查清单

- [ ] Hub连接管理优化
- [ ] 工作线程池调整
- [ ] 数据库索引优化
- [ ] 缓存系统实现
- [ ] 消息去重机制
- [ ] 批量操作优化
- [ ] 监控系统完善
- [ ] 压力测试验证
- [ ] 性能基准测试
- [ ] 文档更新 