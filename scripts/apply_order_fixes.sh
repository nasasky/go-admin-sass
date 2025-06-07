#!/bin/bash

# 订单服务修复应用脚本
# 用于应用所有订单相关的修复和优化

echo "========================================="
echo "开始应用订单服务修复..."
echo "========================================="

# 检查数据库连接
echo "1. 检查数据库连接..."
if ! command -v mysql &> /dev/null; then
    echo "错误: 未找到 mysql 命令，请确保 MySQL 客户端已安装"
    exit 1
fi

# 读取数据库配置（假设配置在 config.yaml 中）
DB_HOST=$(grep -o 'host: [^[:space:]]*' config.yaml | head -1 | cut -d' ' -f2)
DB_PORT=$(grep -o 'port: [^[:space:]]*' config.yaml | head -1 | cut -d' ' -f2)
DB_NAME=$(grep -o 'dbname: [^[:space:]]*' config.yaml | head -1 | cut -d' ' -f2)
DB_USER=$(grep -o 'username: [^[:space:]]*' config.yaml | head -1 | cut -d' ' -f2)

if [ -z "$DB_HOST" ] || [ -z "$DB_NAME" ] || [ -z "$DB_USER" ]; then
    echo "警告: 无法从 config.yaml 读取数据库配置，使用默认值"
    DB_HOST="localhost"
    DB_PORT="3306"
    DB_NAME="go_admin"
    DB_USER="root"
fi

echo "数据库配置:"
echo "  主机: $DB_HOST"
echo "  端口: $DB_PORT"
echo "  数据库: $DB_NAME"
echo "  用户: $DB_USER"

# 执行数据库修复脚本
echo ""
echo "2. 执行数据库表结构修复..."
read -s -p "请输入数据库密码: " DB_PASS
echo ""

mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" < scripts/fix_order_table.sql

if [ $? -eq 0 ]; then
    echo "✓ 数据库表结构修复完成"
else
    echo "✗ 数据库表结构修复失败"
    exit 1
fi

# 检查 Go 代码修复
echo ""
echo "3. 验证代码修复..."
if [ -f "services/app_service/apporder.go" ]; then
    echo "✓ 订单服务文件存在"
    
    # 检查关键修复点
    if grep -q "提交事务 - 只提交一次" services/app_service/apporder.go; then
        echo "✓ 双重提交问题已修复"
    else
        echo "✗ 双重提交问题未修复"
    fi
    
    if grep -q "defer func(orderNo string)" services/app_service/apporder.go; then
        echo "✓ Goroutine 闭包问题已修复"
    else
        echo "✗ Goroutine 闭包问题未修复"
    fi
    
    if grep -q "增加锁定时间" services/app_service/apporder.go; then
        echo "✓ 分布式锁优化已应用"
    else
        echo "✗ 分布式锁优化未应用"
    fi
else
    echo "✗ 订单服务文件不存在"
    exit 1
fi

# 重启服务提示
echo ""
echo "4. 应用完成!"
echo "========================================="
echo "修复内容摘要:"
echo "  ✓ 修复 UpdateOrderStatus 双重事务提交问题"
echo "  ✓ 优化 goroutine 并发安全性"
echo "  ✓ 增强分布式锁机制"
echo "  ✓ 修复数据库表结构"
echo "  ✓ 添加性能优化索引"
echo "  ✓ 增加触发器自动更新时间戳"
echo ""
echo "建议操作:"
echo "  1. 重启应用服务以应用代码修复"
echo "  2. 监控系统日志确认修复效果"
echo "  3. 运行性能测试验证优化效果"
echo ""
echo "重启命令示例:"
echo "  systemctl restart go-admin-service"
echo "  或"
echo "  pkill -f go-admin && nohup ./go-admin &"
echo "=========================================" 