// MongoDB集合和索引设置脚本
// 使用方法: mongo < scripts/setup_mongodb_collections.js

// 切换到notification_log_db数据库
use notification_log_db;

print("开始设置 notification_log_db 数据库...");

// 1. 创建推送记录集合和索引
print("设置 push_records 集合...");
db.createCollection("push_records");
db.push_records.createIndex({"message_id": 1}, {unique: true});
db.push_records.createIndex({"push_time": -1});
db.push_records.createIndex({"message_type": 1});
db.push_records.createIndex({"target": 1});
db.push_records.createIndex({"success": 1});
db.push_records.createIndex({"sender_id": 1});
db.push_records.createIndex({"status": 1});
print("push_records 集合设置完成");

// 2. 创建通知日志集合和索引
print("设置 notification_logs 集合...");
db.createCollection("notification_logs");
db.notification_logs.createIndex({"message_id": 1});
db.notification_logs.createIndex({"timestamp": -1});
db.notification_logs.createIndex({"event_type": 1});
db.notification_logs.createIndex({"user_id": 1});
print("notification_logs 集合设置完成");

// 3. 创建管理员用户接收记录集合和索引（新增）
print("设置 admin_user_receive_records 集合...");
db.createCollection("admin_user_receive_records");
db.admin_user_receive_records.createIndex({"message_id": 1});
db.admin_user_receive_records.createIndex({"user_id": 1});
db.admin_user_receive_records.createIndex({"created_at": -1});
db.admin_user_receive_records.createIndex({"message_id": 1, "user_id": 1}, {unique: true});
db.admin_user_receive_records.createIndex({"is_received": 1});
db.admin_user_receive_records.createIndex({"is_read": 1});
db.admin_user_receive_records.createIndex({"is_confirmed": 1});
db.admin_user_receive_records.createIndex({"delivery_status": 1});
db.admin_user_receive_records.createIndex({"push_channel": 1});
db.admin_user_receive_records.createIndex({"username": 1});
print("admin_user_receive_records 集合设置完成");

// 4. 创建管理员用户在线状态集合和索引（新增）
print("设置 admin_user_online_status 集合...");
db.createCollection("admin_user_online_status");
db.admin_user_online_status.createIndex({"user_id": 1}, {unique: true});
db.admin_user_online_status.createIndex({"is_online": 1});
db.admin_user_online_status.createIndex({"last_seen": -1});
db.admin_user_online_status.createIndex({"username": 1});
print("admin_user_online_status 集合设置完成");

// 5. 验证集合创建
print("验证集合创建情况...");
var collections = db.getCollectionNames();
print("当前数据库中的集合:");
collections.forEach(function(collection) {
    print("  - " + collection);
});

// 6. 显示索引信息
print("\n各集合的索引信息:");

print("push_records 索引:");
db.push_records.getIndexes().forEach(function(index) {
    print("  - " + JSON.stringify(index.key));
});

print("notification_logs 索引:");
db.notification_logs.getIndexes().forEach(function(index) {
    print("  - " + JSON.stringify(index.key));
});

print("admin_user_receive_records 索引:");
db.admin_user_receive_records.getIndexes().forEach(function(index) {
    print("  - " + JSON.stringify(index.key));
});

print("admin_user_online_status 索引:");
db.admin_user_online_status.getIndexes().forEach(function(index) {
    print("  - " + JSON.stringify(index.key));
});

print("MongoDB集合和索引设置完成!"); 