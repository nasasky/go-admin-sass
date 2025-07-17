# 推送系统消息格式规范文档

## 📋 概述

本文档详细说明了NASA Go Admin推送系统的消息格式规范，包括API请求格式、内部消息结构、WebSocket传输格式、数据库存储格式等多个层次的定义。

## 🔄 消息流转架构

```
API请求 → NotificationMessage → WebSocket消息 → 客户端
                ↓
        MongoDB存储记录
                ↓
        接收状态追踪
```

## 📝 1. API请求格式

### 1.1 系统通知请求格式 (SystemNoticeReq)

```go
type SystemNoticeReq struct {
    Content string `json:"content" binding:"required"` // 通知内容 (必填, 最大500字符)
    Type    string `json:"type"`                       // 通知类型 (可选)
    Target  string `json:"target"`                     // 推送目标 (可选, 默认"all")
    UserIDs []int  `json:"user_ids"`                   // 目标用户ID列表 (target为"custom"时必填)
}
```

**请求示例：**
```json
{
  "content": "系统将于今晚23:00-23:30进行例行维护，请提前做好准备",
  "type": "system_notice",
  "target": "all",
  "user_ids": []
}
```

### 1.2 推送目标类型 (Target)

| 值 | 说明 | 需要UserIDs |
|---|---|---|
| `all` | 广播给所有在线用户 | 否 |
| `admin` | 只发送给管理员用户 | 否 |
| `custom` | 发送给指定用户列表 | 是 |

### 1.3 消息类型 (Type)

| 值 | 说明 | 场景 |
|---|---|---|
| `system_notice` | 系统通知 | 普通系统消息 |
| `system_maintain` | 系统维护 | 维护通知 |
| `system_upgrade` | 系统升级 | 升级通知 |
| `order_created` | 订单创建 | 新订单通知 |
| `order_paid` | 订单支付 | 支付成功通知 |
| `order_cancelled` | 订单取消 | 订单取消通知 |
| `order_refunded` | 订单退款 | 退款通知 |
| `user_login` | 用户登录 | 登录通知 |
| `payment_success` | 支付成功 | 支付通知 |
| `message_received` | 消息接收 | 消息提醒 |

## 📡 2. 内部消息格式 (NotificationMessage)

### 2.1 核心结构

```go
type NotificationMessage struct {
    Type        NotificationType     `json:"type"`               // 消息类型
    Content     string               `json:"content"`            // 消息内容
    Data        interface{}          `json:"data,omitempty"`     // 附加数据
    Time        string               `json:"time"`               // 消息时间
    Priority    NotificationPriority `json:"-"`                 // 优先级 (内部使用)
    Target      NotificationTarget   `json:"-"`                 // 目标类型 (内部使用)  
    TargetIDs   []int                `json:"-"`                 // 目标用户ID (内部使用)
    ExcludeIDs  []int                `json:"-"`                 // 排除用户ID (内部使用)
    NeedConfirm bool                 `json:"-"`                 // 需要确认 (内部使用)
    MessageID   string               `json:"message_id,omitempty"` // 消息ID
}
```

### 2.2 优先级定义 (NotificationPriority)

```go
const (
    PriorityLow    NotificationPriority = 0  // 低优先级
    PriorityNormal NotificationPriority = 1  // 普通优先级
    PriorityHigh   NotificationPriority = 2  // 高优先级
    PriorityUrgent NotificationPriority = 3  // 紧急优先级 (直接处理)
)
```

### 2.3 目标类型定义 (NotificationTarget)

```go
const (
    TargetUser   NotificationTarget = 0  // 发送给特定用户
    TargetAdmin  NotificationTarget = 1  // 发送给管理员
    TargetAll    NotificationTarget = 2  // 广播给所有人
    TargetGroup  NotificationTarget = 3  // 发送给特定组
    TargetCustom NotificationTarget = 4  // 自定义发送目标
)
```

### 2.4 消息ID格式

```go
// 格式: {timestamp_nano}-{uuid_8位}
// 示例: 1706335800123456789-abc12345
func generateMessageID() string {
    return fmt.Sprintf("%d-%s", time.Now().UnixNano(), uuid.New().String()[:8])
}
```

## 🌐 3. WebSocket消息格式

### 3.1 基础WebSocket消息结构

```go
type Message struct {
    Type    MessageType `json:"type"`    // 消息类型
    Content string      `json:"content"` // 消息内容
    Data    interface{} `json:"data"`    // 消息数据
    Time    string      `json:"time"`    // 消息时间
}
```

### 3.2 客户端接收的消息格式

```json
{
  "type": "system_notice",
  "content": "系统将于今晚23:00-23:30进行例行维护",
  "data": {
    "priority": 1,
    "need_confirm": false,
    "extra_info": "请提前保存工作"
  },
  "time": "2025-01-27 10:30:00",
  "message_id": "1706335800123456789-abc12345"
}
```

### 3.3 订单通知消息格式

```json
{
  "type": "order_paid",
  "content": "订单支付成功",
  "data": {
    "order_no": "ORD20250127001",
    "status": "paid",
    "goods_name": "会议室预订",
    "user_id": 123,
    "amount": "199.00"
  },
  "time": "2025-01-27 10:30:00",
  "message_id": "1706335800123456789-def67890"
}
```

### 3.4 复杂业务消息格式示例

```json
{
  "type": "system_upgrade",
  "content": "系统升级提醒",
  "data": {
    "version": "v2.5.0",
    "features": ["新增群聊功能", "性能优化"],
    "upgrade_time": "2025-04-25 01:00:00",
    "downtime_expected": "30分钟",
    "impact_services": ["订单系统", "支付系统"]
  },
  "time": "2025-01-27 10:30:00",
  "message_id": "1706335800123456789-ghi12345"
}
```

## 💾 4. 数据库存储格式

### 4.1 推送记录 (PushRecord) - MongoDB

```go
type PushRecord struct {
    ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    MessageID       string             `bson:"message_id" json:"message_id"`             // 消息唯一ID
    Content         string             `bson:"content" json:"content"`                   // 推送内容
    MessageType     string             `bson:"message_type" json:"message_type"`         // 消息类型
    Target          string             `bson:"target" json:"target"`                     // 推送目标
    TargetUserIDs   []int              `bson:"target_user_ids,omitempty" json:"target_user_ids,omitempty"` // 目标用户ID列表
    RecipientsCount string             `bson:"recipients_count" json:"recipients_count"` // 接收者数量描述
    Status          string             `bson:"status" json:"status"`                     // 推送状态：delivered, failed
    Success         bool               `bson:"success" json:"success"`                   // 是否成功
    Error           string             `bson:"error,omitempty" json:"error,omitempty"`   // 错误信息
    ErrorCode       string             `bson:"error_code,omitempty" json:"error_code,omitempty"` // 错误代码
    PushTime        string             `bson:"push_time" json:"push_time"`               // 推送时间
    CreatedAt       string             `bson:"created_at" json:"created_at"`             // 创建时间
    UpdatedAt       string             `bson:"updated_at" json:"updated_at"`             // 更新时间
    
    // 发送者信息
    SenderID   int    `bson:"sender_id" json:"sender_id"`     // 发送者ID
    SenderName string `bson:"sender_name" json:"sender_name"` // 发送者名称
    
    // 统计信息
    DeliveredCount int64 `bson:"delivered_count" json:"delivered_count"` // 实际送达数量
    FailedCount    int64 `bson:"failed_count" json:"failed_count"`       // 失败数量
    TotalCount     int64 `bson:"total_count" json:"total_count"`         // 总数量
    
    // 扩展信息
    Priority    int                    `bson:"priority" json:"priority"`         // 优先级
    NeedConfirm bool                   `bson:"need_confirm" json:"need_confirm"` // 是否需要确认
    ExtraData   map[string]interface{} `bson:"extra_data,omitempty" json:"extra_data,omitempty"` // 扩展数据
}
```

**MongoDB文档示例：**
```json
{
  "_id": ObjectId("507f1f77bcf86cd799439011"),
  "message_id": "1706335800123456789-abc12345",
  "content": "系统将于今晚23:00-23:30进行例行维护",
  "message_type": "system_notice",
  "target": "all",
  "recipients_count": "all_online_users",
  "status": "delivered",
  "success": true,
  "push_time": "2025-01-27 10:30:00",
  "sender_id": 1,
  "sender_name": "admin",
  "delivered_count": 45,
  "failed_count": 0,
  "total_count": 45,
  "priority": 1,
  "need_confirm": false,
  "created_at": "2025-01-27 10:30:00",
  "updated_at": "2025-01-27 10:30:00"
}
```

### 4.2 管理员用户接收记录 (AdminUserReceiveRecord) - MongoDB

```go
type AdminUserReceiveRecord struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    MessageID string             `bson:"message_id" json:"message_id"` // 消息ID
    UserID    int                `bson:"user_id" json:"user_id"`       // 管理员用户ID
    Username  string             `bson:"username" json:"username"`     // 用户名
    UserRole  string             `bson:"user_role" json:"user_role"`   // 用户角色
    IsOnline  bool               `bson:"is_online" json:"is_online"`   // 推送时是否在线

    // 接收状态
    IsReceived  bool   `bson:"is_received" json:"is_received"`   // 是否接收到
    ReceivedAt  string `bson:"received_at" json:"received_at"`   // 接收时间
    IsRead      bool   `bson:"is_read" json:"is_read"`           // 是否已读
    ReadAt      string `bson:"read_at" json:"read_at"`           // 阅读时间
    IsConfirmed bool   `bson:"is_confirmed" json:"is_confirmed"` // 是否确认
    ConfirmedAt string `bson:"confirmed_at" json:"confirmed_at"` // 确认时间

    // 设备和环境信息
    DeviceType   string `bson:"device_type" json:"device_type"`     // 设备类型：desktop, mobile, tablet
    Platform     string `bson:"platform" json:"platform"`           // 平台：windows, mac, ios, android
    Browser      string `bson:"browser" json:"browser"`             // 浏览器
    ClientIP     string `bson:"client_ip" json:"client_ip"`         // 客户端IP
    UserAgent    string `bson:"user_agent" json:"user_agent"`       // 用户代理
    ConnectionID string `bson:"connection_id" json:"connection_id"` // WebSocket连接ID

    // 推送信息
    PushChannel    string `bson:"push_channel" json:"push_channel"`       // 推送渠道：websocket, offline
    DeliveryStatus string `bson:"delivery_status" json:"delivery_status"` // 投递状态：delivered, failed, pending
    RetryCount     int    `bson:"retry_count" json:"retry_count"`         // 重试次数
    ErrorMessage   string `bson:"error_message" json:"error_message"`     // 错误信息
    
    CreatedAt string `bson:"created_at" json:"created_at"` // 创建时间
    UpdatedAt string `bson:"updated_at" json:"updated_at"` // 更新时间
}
```

### 4.3 通知日志 (NotificationLog) - MongoDB

```go
type NotificationLog struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    MessageID string             `bson:"message_id" json:"message_id"` // 关联的消息ID
    UserID    int                `bson:"user_id" json:"user_id"`       // 用户ID
    Username  string             `bson:"username" json:"username"`     // 用户名
    EventType string             `bson:"event_type" json:"event_type"` // 事件类型：sent, delivered, failed, read, confirmed
    Status    string             `bson:"status" json:"status"`         // 状态
    Timestamp string             `bson:"timestamp" json:"timestamp"`   // 时间戳
    CreatedAt string             `bson:"created_at" json:"created_at"` // 创建时间

    // 详细信息
    Error      string                 `bson:"error,omitempty" json:"error,omitempty"`             // 错误信息
    DeviceInfo map[string]interface{} `bson:"device_info,omitempty" json:"device_info,omitempty"` // 设备信息
    ClientIP   string                 `bson:"client_ip,omitempty" json:"client_ip,omitempty"`     // 客户端IP
    UserAgent  string                 `bson:"user_agent,omitempty" json:"user_agent,omitempty"`   // 用户代理
}
```

## 📱 5. 离线消息格式

### 5.1 Redis存储格式

```json
{
  "type": "system_notice",
  "content": "[离线消息] 系统维护通知",
  "data": {
    "original_push_time": "2025-01-27 10:30:00"
  },
  "time": "2025-01-27 10:30:00",
  "message_id": "1706335800123456789-abc12345",
  "saved_at": "2025-01-27 10:30:05"
}
```

### 5.2 离线消息Key格式

```
# Redis Key格式
offline_msg:{user_id}

# 示例
offline_msg:123
```

## 📊 6. API响应格式

### 6.1 推送成功响应

```json
{
  "code": 200,
  "message": "OK",
  "data": {
    "success": true,
    "message": "系统将于今晚23:00-23:30进行例行维护",
    "push_time": "2025-01-27 10:30:00",
    "target": "all",
    "message_type": "system_notice",
    "recipients_count": "all_online_users",
    "message_id": "1706335800123456789-abc12345",
    "status": "delivered"
  }
}
```

### 6.2 推送失败响应

```json
{
  "code": 20001,
  "message": "推送失败: WebSocket服务不可用",
  "data": {
    "success": false,
    "message": "系统通知内容",
    "push_time": "2025-01-27 10:30:00",
    "target": "all",
    "message_type": "system_notice",
    "status": "failed",
    "error": "WebSocket服务不可用",
    "error_code": "PUSH_FAILED"
  }
}
```

### 6.3 推送记录列表响应

```json
{
  "code": 200,
  "message": "OK",
  "data": {
    "total": 156,
    "page": 1,
    "page_size": 20,
    "items": [
      {
        "id": "507f1f77bcf86cd799439011",
        "message_id": "1706335800123456789-abc12345",
        "content": "系统维护通知",
        "message_type": "system_notice",
        "target": "all",
        "status": "delivered",
        "success": true,
        "push_time": "2025-01-27 10:30:00",
        "sender_name": "admin",
        "delivered_count": 45,
        "failed_count": 0
      }
    ],
    "stats": {
      "total_records": 156,
      "success_records": 142,
      "failed_records": 14,
      "total_recipients": 3240,
      "delivered_count": 2980,
      "failed_count": 260
    }
  }
}
```

## 🔧 7. 使用示例

### 7.1 发送系统通知

```bash
curl -X POST http://localhost:8080/api/admin/system/notice \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin_token>" \
  -d '{
    "content": "系统将于今晚23:00进行维护",
    "type": "system_maintain", 
    "target": "admin"
  }'
```

### 7.2 发送自定义用户通知

```bash
curl -X POST http://localhost:8080/api/admin/system/notice \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin_token>" \
  -d '{
    "content": "您的订单已完成",
    "type": "order_completed",
    "target": "custom",
    "user_ids": [123, 456, 789]
  }'
```

### 7.3 客户端WebSocket接收处理

```javascript
// 前端WebSocket处理示例
websocket.onmessage = function(event) {
    const message = JSON.parse(event.data);
    
    console.log('收到消息:', message);
    
    // 根据消息类型处理
    switch(message.type) {
        case 'system_notice':
            showSystemNotification(message);
            break;
        case 'order_paid':
            showOrderNotification(message);
            break;
        case 'system_maintain':
            showMaintenanceAlert(message);
            break;
    }
    
    // 发送已读确认
    if (message.message_id) {
        markMessageAsRead(message.message_id);
    }
};

function markMessageAsRead(messageId) {
    fetch('/api/admin/notification/mark-read', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + token
        },
        body: JSON.stringify({
            message_id: messageId,
            user_id: currentUserId
        })
    });
}
```

## 📈 8. 性能和限制

### 8.1 消息内容限制

- **内容长度**: 最大500字符
- **用户ID列表**: 最大1000个用户
- **附加数据**: 建议小于1KB

### 8.2 发送频率限制

- **紧急消息**: 无限制，直接处理
- **普通消息**: 通过队列处理，队列大小1000
- **批量发送**: 建议每批最多1000用户，间隔500ms

### 8.3 存储保留策略

- **推送记录**: 保留30天
- **通知日志**: 保留7天
- **离线消息**: 保留7天，最多100条/用户
- **接收记录**: 保留30天

## 🛠️ 9. 错误码说明

| 错误码 | 说明 | 解决方案 |
|-------|------|----------|
| `PUSH_FAILED` | 推送失败 | 检查WebSocket服务状态 |
| `INVALID_TARGET` | 无效的推送目标 | 检查target参数 |
| `MISSING_USER_IDS` | 缺少用户ID列表 | custom目标需要提供user_ids |
| `CONTENT_TOO_LONG` | 内容过长 | 内容不能超过500字符 |
| `RATE_LIMIT_EXCEEDED` | 频率限制 | 降低发送频率 |
| `QUEUE_FULL` | 消息队列已满 | 稍后重试 |

## 📋 10. 最佳实践

### 10.1 消息设计原则

1. **内容简洁**: 消息内容应该简洁明了
2. **类型明确**: 使用合适的消息类型
3. **数据结构**: 附加数据使用结构化格式
4. **时间格式**: 统一使用 "2006-01-02 15:04:05" 格式

### 10.2 错误处理

1. **重试机制**: 失败消息自动重试3次
2. **降级策略**: WebSocket失败时保存为离线消息
3. **监控告警**: 关键消息发送失败时及时告警

### 10.3 性能优化

1. **批量发送**: 大量用户时使用批量发送
2. **异步处理**: 使用goroutine异步保存记录
3. **缓存优化**: 管理员用户列表缓存5分钟

这个格式规范确保了推送系统的消息在各个环节都有统一的结构和处理方式，便于开发、测试和维护。 