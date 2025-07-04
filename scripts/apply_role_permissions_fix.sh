#!/bin/bash

# 角色权限控制修复脚本
# 应用数据库迁移：给role表添加用户关联字段

set -e  # 遇到错误立即退出

echo "🚀 开始应用角色权限控制修复..."

# 解析数据库连接信息的函数
parse_database_config() {
    # 优先读取 .env 文件
    if [ -f ".env" ]; then
        echo "📋 读取 .env 文件配置..."
        source .env
        
        # 从环境变量获取数据库连接字符串
        if [ -n "$Mysql" ]; then
            DSN="$Mysql"
        elif [ -n "$MYSQL_DSN" ]; then
            DSN="$MYSQL_DSN"
        fi
        
        # 解析DSN字符串获取连接参数
        if [ -n "$DSN" ]; then
            echo "🔍 解析数据库连接字符串: ${DSN:0:30}..."
            
            # 解析DSN格式: user:password@tcp(host:port)/database
            if [[ "$DSN" =~ ([^:]+):([^@]+)@tcp\(([^:]+):([0-9]+)\)/([^?]+) ]]; then
                DB_USER="${BASH_REMATCH[1]}"
                DB_PASSWORD="${BASH_REMATCH[2]}"
                DB_HOST="${BASH_REMATCH[3]}"
                DB_PORT="${BASH_REMATCH[4]}"
                DB_NAME="${BASH_REMATCH[5]}"
                
                echo "✅ 成功解析数据库配置:"
                echo "   - 主机: $DB_HOST:$DB_PORT"
                echo "   - 用户: $DB_USER"
                echo "   - 数据库: $DB_NAME"
                return 0
            fi
        fi
    fi
    
    # 如果无法解析，使用默认值或手动输入
    echo "⚠️  无法从配置文件解析数据库连接信息，使用默认值或手动输入"
    DB_HOST=${DB_HOST:-"localhost"}
    DB_PORT=${DB_PORT:-"3306"}
    DB_USER=${DB_USER:-"root"}
    DB_NAME=${DB_NAME:-"nasa_go_admin"}
    
    if [ -z "$DB_PASSWORD" ]; then
        echo "请输入数据库密码:"
        read -s DB_PASSWORD
    fi
}

# 解析数据库配置
parse_database_config

# 检查数据库连接
echo "📋 检查数据库连接..."
if ! command -v mysql &> /dev/null; then
    echo "⚠️  MySQL 客户端未安装，跳过连接测试"
    echo "请确保数据库配置正确："
    echo "   - 主机: $DB_HOST:$DB_PORT"
    echo "   - 用户: $DB_USER"
    echo "   - 数据库: $DB_NAME"
    echo ""
    read -p "是否继续执行迁移？(y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ 用户取消操作"
        exit 1
    fi
elif ! mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "USE $DB_NAME; SELECT 1;" > /dev/null 2>&1; then
    echo "❌ 数据库连接失败，请检查配置："
    echo "   - 主机: $DB_HOST:$DB_PORT"
    echo "   - 用户: $DB_USER"
    echo "   - 数据库: $DB_NAME"
    echo ""
    read -p "是否强制继续执行迁移？(y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ 数据库连接失败，停止执行"
        exit 1
    fi
    echo "⚠️  强制继续执行..."
else
    echo "✅ 数据库连接成功"
fi

# 备份当前角色表结构
echo "📦 备份当前角色表结构..."
mysqldump -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" --no-data "$DB_NAME" role > "role_structure_backup_$(date +%Y%m%d_%H%M%S).sql"

# 应用迁移
echo "🔧 应用数据库迁移..."
if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < migrations/add_role_user_fields.sql; then
    echo "✅ 迁移应用成功"
else
    echo "❌ 迁移应用失败"
    exit 1
fi

# 验证迁移结果
echo "🔍 验证迁移结果..."
RESULT=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" -e "
SELECT 
    COUNT(*) as role_count,
    COUNT(CASE WHEN user_id IS NOT NULL THEN 1 END) as roles_with_user_id,
    COUNT(CASE WHEN role_name IS NOT NULL AND role_name != '' THEN 1 END) as roles_with_name
FROM role;" 2>/dev/null || echo "验证查询失败")

echo "验证结果："
echo "$RESULT"

echo ""
echo "🎉 角色权限控制修复完成！"
echo ""
echo "主要修改："
echo "  ✅ role表添加了user_id字段（创建者ID）"
echo "  ✅ role表添加了user_type字段（创建者类型）"
echo "  ✅ role表添加了role_name字段（角色名称）"
echo "  ✅ role表添加了role_desc字段（角色描述）"
echo "  ✅ role表添加了时间字段"
echo "  ✅ 创建了相关索引"
echo ""
echo "权限控制逻辑："
echo "  - 超管(user_type=1)可以查看和管理所有角色"
echo "  - 非超管用户只能管理自己创建的角色"
echo ""
echo "请重启应用以使代码修改生效。" 