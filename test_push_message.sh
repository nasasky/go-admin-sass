#!/bin/bash

# 测试推送消息脚本
echo "🧪 开始测试推送消息..."

# 获取当前时间
CURRENT_TIME=$(date '+%Y-%m-%d %H:%M:%S')

echo "发送测试推送消息..."

# 测试API调用 - 发送给管理员
curl -X POST "http://localhost:8080/api/v3/admin/system/notice" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "content": "测试推送消息 - '$CURRENT_TIME'",
    "type": "system_notice",
    "target": "admin"
  }' \
  -w "\nHTTP Status: %{http_code}\n"

echo ""
echo "等待5秒后检查结果..."
sleep 5

# 检查MongoDB中的记录
mongosh notification_log_db --quiet --eval "
print('最新推送记录:');
db.push_records.find().sort({push_time: -1}).limit(1).forEach(function(doc) {
    print('MessageID: ' + doc.message_id);
    print('Content: ' + doc.content);
    print('Target: ' + doc.target);
    print('Success: ' + doc.success);
});

print('\\n对应的接收记录:');
var latestPush = db.push_records.findOne({}, {sort: {push_time: -1}});
if (latestPush) {
    db.admin_user_receive_records.find({message_id: latestPush.message_id}).forEach(function(doc) {
        print('UserID: ' + doc.user_id + ', Status: ' + doc.delivery_status);
    });
}
"

