# Sender Name 修复验证

## 问题描述
之前 `sender_name` 字段在推送记录中没有值，原因是类型断言错误。

## 修复内容
在 `controllers/admin/system.go` 的 `PostnoticeInfo` 函数中：

### 修复前：
```go
if user, ok := userInfo.(map[string]interface{}); ok {
    if name, exists := user["username"]; exists {
        senderName = name.(string)
    }
}
```

### 修复后：
```go
if user, ok := userInfo.(map[string]string); ok {
    if name, exists := user["username"]; exists {
        senderName = name
    }
}
```

## 验证步骤

### 1. 发送系统通知
```bash
POST /api/admin/system/notice
Content-Type: application/json
Authorization: Bearer <admin_token>

{
  "content": "测试推送记录修复",
  "type": "system_notice",
  "target": "all"
}
```

### 2. 查询推送记录
```bash
GET /api/admin/notification/records?page=1&page_size=10
Authorization: Bearer <admin_token>
```

### 3. 检查返回数据
应该能看到 `sender_name` 字段有正确的用户名值，而不是 "unknown"。

### 4. 直接查询MongoDB
```javascript
use notification_log_db
db.push_records.find().sort({"push_time": -1}).limit(1)
```

应该能看到类似这样的记录：
```json
{
  "_id": ObjectId("..."),
  "message_id": "msg_...",
  "content": "测试推送记录修复",
  "sender_id": 1,
  "sender_name": "admin",  // 这里应该有正确的用户名
  "push_time": "2025-01-27 10:30:00",
  "success": true,
  "status": "delivered"
}
```

## 预期结果
- `sender_name` 字段应该显示发送者的实际用户名
- 不再显示 "unknown"
- 推送记录完整保存到MongoDB 