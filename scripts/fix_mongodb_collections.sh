#!/bin/bash

# MongoDB集合修复脚本
# 解决 "Collection admin_user_receive_records not found in database notification_log_db config" 错误

echo "🔧 开始修复MongoDB集合配置问题..."

# 1. 检查MongoDB是否运行
echo "📋 检查MongoDB服务状态..."
if ! pgrep -x "mongod" > /dev/null; then
    echo "❌ MongoDB服务未运行，请先启动MongoDB"
    echo "   启动命令: sudo systemctl start mongod"
    echo "   或者: mongod --dbpath /data/db"
    exit 1
fi

echo "✅ MongoDB服务正在运行"

# 2. 检查MongoDB客户端工具
MONGO_CLIENT=""
if command -v mongosh &> /dev/null; then
    MONGO_CLIENT="mongosh"
elif command -v mongo &> /dev/null; then
    MONGO_CLIENT="mongo"
else
    echo "❌ 未找到MongoDB客户端工具 (mongosh 或 mongo)"
    echo "   请安装MongoDB客户端工具"
    exit 1
fi

echo "✅ 使用MongoDB客户端: $MONGO_CLIENT"

# 3. 执行MongoDB集合设置脚本
echo "📦 创建MongoDB集合和索引..."
if [ -f "scripts/setup_mongodb_collections.js" ]; then
    $MONGO_CLIENT < scripts/setup_mongodb_collections.js
    if [ $? -eq 0 ]; then
        echo "✅ MongoDB集合和索引创建成功"
    else
        echo "❌ MongoDB集合创建失败"
        exit 1
    fi
else
    echo "❌ MongoDB设置脚本不存在: scripts/setup_mongodb_collections.js"
    exit 1
fi

# 4. 验证集合创建
echo "🔍 验证集合创建..."
COLLECTIONS=$($MONGO_CLIENT notification_log_db --quiet --eval "db.getCollectionNames().join(',')")
echo "当前集合: $COLLECTIONS"

if [[ $COLLECTIONS == *"admin_user_receive_records"* ]]; then
    echo "✅ admin_user_receive_records 集合已创建"
else
    echo "❌ admin_user_receive_records 集合未找到"
    exit 1
fi

if [[ $COLLECTIONS == *"admin_user_online_status"* ]]; then
    echo "✅ admin_user_online_status 集合已创建"
else
    echo "❌ admin_user_online_status 集合未找到"
    exit 1
fi

# 5. 检查配置文件
echo "📋 检查配置文件..."
if grep -q "admin_user_receive_records" config/config.yaml; then
    echo "✅ config.yaml 已包含新集合配置"
else
    echo "❌ config.yaml 缺少新集合配置，请手动添加或重新运行配置更新"
fi

# 6. 重启应用建议
echo ""
echo "🚀 修复完成! 请重启应用程序："
echo "   1. 停止当前应用: Ctrl+C"
echo "   2. 重新启动应用: go run main.go"
echo ""
echo "📊 可选：查看集合状态"
echo "   $MONGO_CLIENT notification_log_db --eval 'db.stats()'"
echo "   $MONGO_CLIENT notification_log_db --eval 'db.admin_user_receive_records.count()'"

echo "✅ MongoDB集合修复完成!" 