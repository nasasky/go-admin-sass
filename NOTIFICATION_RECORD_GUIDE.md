# 📢 推送记录管理功能指南

## 🎯 功能概述

新增的推送记录管理模块可以：

- ✅ **自动记录**：所有系统通知推送都会自动保存到MongoDB
- ✅ **详细追踪**：记录推送状态、发送者、接收者等详细信息
- ✅ **统计分析**：提供推送成功率、送达率等统计数据
- ✅ **重新发送**：支持失败推送的重新发送功能
- ✅ **日志查询**：查看每条推送的详细日志记录

## 🗄️ 数据库配置

### MongoDB配置
在 `config/config.yaml` 中已添加推送记录数据库配置：

```yaml
mongodb:
  databases:
    notification_log_db:
      uri: "mongodb://localhost:27017"
      collections:
        push_records: "push_records"        # 推送记录集合
        notification_logs: "notification_logs"  # 通知日志集合
```

## 📊 数据模型

### PushRecord (推送记录)
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
    Priority     int                    `bson:"priority" json:"priority"`         // 优先级
    NeedConfirm  bool                   `bson:"need_confirm" json:"need_confirm"` // 是否需要确认
    ExtraData    map[string]interface{} `bson:"extra_data,omitempty" json:"extra_data,omitempty"` // 扩展数据
}
```

### NotificationLog (通知日志)
```go
type NotificationLog struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    MessageID   string             `bson:"message_id" json:"message_id"`     // 关联的消息ID
    UserID      int                `bson:"user_id" json:"user_id"`           // 用户ID
    Username    string             `bson:"username" json:"username"`         // 用户名
    EventType   string             `bson:"event_type" json:"event_type"`     // 事件类型：sent, delivered, failed, read
    Status      string             `bson:"status" json:"status"`             // 状态
    Timestamp   string             `bson:"timestamp" json:"timestamp"`       // 时间戳
    CreatedAt   string             `bson:"created_at" json:"created_at"`     // 创建时间
    
    // 详细信息
    Error       string                 `bson:"error,omitempty" json:"error,omitempty"`           // 错误信息
    DeviceInfo  map[string]interface{} `bson:"device_info,omitempty" json:"device_info,omitempty"` // 设备信息
    ClientIP    string                 `bson:"client_ip,omitempty" json:"client_ip,omitempty"`   // 客户端IP
    UserAgent   string                 `bson:"user_agent,omitempty" json:"user_agent,omitempty"` // 用户代理
}
```

## 🔌 API接口

### 1. 发送系统通知 (已增强)
```bash
POST /api/admin/system/notice
Content-Type: application/json
Authorization: Bearer <admin_token>

{
  "content": "系统将于今晚23:00-23:30进行例行维护，请提前做好准备",
  "type": "system_notice",
  "target": "all",
  "user_ids": []
}
```

**响应示例：**
```json
{
  "code": 200,
  "message": "OK",
  "data": {
    "success": true,
    "message": "系统将于今晚23:00-23:30进行例行维护，请提前做好准备",
    "push_time": "2025-01-27 10:30:00",
    "target": "all",
    "message_type": "system_notice",
    "recipients_count": "all_online_users",
    "message_id": "msg_1706335800_abc123",
    "status": "delivered"
  }
}
```

### 2. 获取推送记录列表
```bash
GET /api/admin/notification/records?page=1&page_size=20&message_type=system_notice&target=all&status=delivered&sort_by=push_time&sort_order=desc
Authorization: Bearer <admin_token>
```

**查询参数：**
- `page`: 页码 (默认: 1)
- `page_size`: 每页数量 (默认: 10, 最大: 100)
- `message_type`: 消息类型过滤
- `target`: 推送目标过滤 (all/admin/custom)
- `status`: 状态过滤 (delivered/failed)
- `success`: 成功状态过滤 (true/false)
- `sender_id`: 发送者ID过滤
- `start_date`: 开始日期 (格式: 2025-01-27)
- `end_date`: 结束日期 (格式: 2025-01-27)
- `keyword`: 关键词搜索 (内容或消息ID)
- `sort_by`: 排序字段 (默认: push_time)
- `sort_order`: 排序方向 (asc/desc, 默认: desc)

**响应示例：**
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
        "message_id": "msg_1706335800_abc123",
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
        "created_at": "2025-01-27 10:30:00",
        "updated_at": "2025-01-27 10:30:00"
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

### 3. 获取推送记录详情
```bash
GET /api/admin/notification/records/{id}
Authorization: Bearer <admin_token>
```

**响应示例：**
```json
{
  "code": 200,
  "message": "OK",
  "data": {
    "record": {
      "id": "507f1f77bcf86cd799439011",
      "message_id": "msg_1706335800_abc123",
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
      "total_count": 45
    },
    "logs": [
      {
        "id": "507f1f77bcf86cd799439012",
        "message_id": "msg_1706335800_abc123",
        "user_id": 123,
        "username": "user123",
        "event_type": "delivered",
        "status": "success",
        "timestamp": "2025-01-27 10:30:05",
        "client_ip": "192.168.1.100"
      }
    ]
  }
}
```

### 4. 删除推送记录
```bash
DELETE /api/admin/notification/records/{id}
Authorization: Bearer <admin_token>
```

### 5. 获取推送统计
```bash
GET /api/admin/notification/stats?start_date=2025-01-01&end_date=2025-01-27
Authorization: Bearer <admin_token>
```

**响应示例：**
```json
{
  "code": 200,
  "message": "OK",
  "data": {
    "total_records": 156,
    "success_records": 142,
    "failed_records": 14,
    "total_recipients": 3240,
    "delivered_count": 2980,
    "failed_count": 260
  }
}
```

### 6. 重新发送通知
```bash
POST /api/admin/notification/records/{id}/resend
Authorization: Bearer <admin_token>
```

**响应示例：**
```json
{
  "code": 200,
  "message": "OK",
  "data": {
    "message": "重新发送成功",
    "status": "delivered"
  }
}
```

## 🔍 MongoDB查询示例

### 直接查询MongoDB
```javascript
// 连接到MongoDB
use notification_log_db

// 查看所有推送记录
db.push_records.find().sort({"push_time": -1}).limit(10)

// 查看失败的推送记录
db.push_records.find({"success": false}).sort({"push_time": -1})

// 按消息类型统计
db.push_records.aggregate([
  {$group: {
    "_id": "$message_type",
    "count": {$sum: 1},
    "success_count": {$sum: {$cond: ["$success", 1, 0]}},
    "failed_count": {$sum: {$cond: ["$success", 0, 1]}}
  }},
  {$sort: {"count": -1}}
])

// 查看特定发送者的推送记录
db.push_records.find({"sender_id": 1}).sort({"push_time": -1})

// 查看最近24小时的推送记录
db.push_records.find({
  "push_time": {
    $gte: "2025-01-26 10:30:00",
    $lte: "2025-01-27 10:30:00"
  }
})

// 查看通知日志
db.notification_logs.find({"message_id": "msg_1706335800_abc123"})

// 按事件类型统计通知日志
db.notification_logs.aggregate([
  {$group: {
    "_id": "$event_type",
    "count": {$sum: 1}
  }},
  {$sort: {"count": -1}}
])
```

## 📈 监控和统计

### 推送成功率统计
```javascript
// 计算总体推送成功率
db.push_records.aggregate([
  {$group: {
    "_id": null,
    "total": {$sum: 1},
    "success": {$sum: {$cond: ["$success", 1, 0]}},
    "failed": {$sum: {$cond: ["$success", 0, 1]}}
  }},
  {$project: {
    "total": 1,
    "success": 1,
    "failed": 1,
    "success_rate": {$multiply: [{$divide: ["$success", "$total"]}, 100]}
  }}
])
```

### 按时间段统计
```javascript
// 按小时统计推送数量
db.push_records.aggregate([
  {$project: {
    "hour": {$substr: ["$push_time", 11, 2]},
    "success": 1
  }},
  {$group: {
    "_id": "$hour",
    "count": {$sum: 1},
    "success_count": {$sum: {$cond: ["$success", 1, 0]}}
  }},
  {$sort: {"_id": 1}}
])
```

## 🚀 使用建议

### 1. 定期清理旧记录
建议定期清理超过30天的推送记录，避免数据库过大：

```javascript
// 删除30天前的记录
db.push_records.deleteMany({
  "push_time": {
    $lt: "2024-12-27 00:00:00"
  }
})
```

### 2. 创建索引优化查询
```javascript
// 为常用查询字段创建索引
db.push_records.createIndex({"push_time": -1})
db.push_records.createIndex({"message_type": 1})
db.push_records.createIndex({"target": 1})
db.push_records.createIndex({"success": 1})
db.push_records.createIndex({"sender_id": 1})
db.push_records.createIndex({"message_id": 1})

// 为通知日志创建索引
db.notification_logs.createIndex({"message_id": 1})
db.notification_logs.createIndex({"timestamp": -1})
db.notification_logs.createIndex({"event_type": 1})
```

### 3. 监控推送质量
- 定期检查推送成功率
- 关注失败推送的原因
- 分析用户接收情况
- 优化推送内容和时间

## 🔧 故障排除

### 常见问题

1. **推送记录保存失败**
   - 检查MongoDB连接状态
   - 确认数据库和集合权限
   - 查看应用日志

2. **查询性能慢**
   - 检查是否创建了合适的索引
   - 优化查询条件
   - 考虑分页查询

3. **统计数据不准确**
   - 检查时间范围设置
   - 确认数据完整性
   - 验证统计逻辑

### 日志查看
```bash
# 查看应用日志
tail -f logs/app.log | grep "推送记录"

# 查看MongoDB日志
tail -f /var/log/mongodb/mongod.log
```

## 📝 更新日志

### v1.0.0 (2025-01-27)
- ✅ 新增推送记录自动保存功能
- ✅ 新增推送记录查询和管理API
- ✅ 新增推送统计功能
- ✅ 新增重新发送失败推送功能
- ✅ 新增通知日志记录功能
- ✅ 集成MongoDB存储
- ✅ 提供完整的REST API接口 