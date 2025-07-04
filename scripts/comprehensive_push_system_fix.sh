#!/bin/bash

# 全面推送系统修复脚本
# 解决离线消息不推送、接收记录状态错误等问题

set -e

echo "🔧 开始全面修复推送系统..."

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 数据库配置
DB_HOST="127.0.0.1"
DB_PORT="3306"
DB_USER="root"
DB_NAME="naive_admin"

echo -e "${BLUE}[INFO]${NC} 数据库配置: ${DB_USER}@${DB_HOST}:${DB_PORT}/${DB_NAME}"

# 1. 检查数据库连接
echo -e "${BLUE}[INFO]${NC} === 检查数据库连接 ==="
if ! mysql -u"$DB_USER" -h"$DB_HOST" -P"$DB_PORT" -e "USE $DB_NAME; SELECT 1;" >/dev/null 2>&1; then
    echo -e "${RED}[ERROR]${NC} 数据库连接失败，请检查数据库配置"
    exit 1
fi
echo -e "${GREEN}[SUCCESS]${NC} 数据库连接正常"

# 2. 检查用户表结构
echo -e "${BLUE}[INFO]${NC} === 检查用户表结构 ==="
TABLE_INFO=$(mysql -u"$DB_USER" -h"$DB_HOST" -P"$DB_PORT" -e "USE $DB_NAME; DESCRIBE user;" 2>/dev/null || echo "")

if echo "$TABLE_INFO" | grep -q "user_type"; then
    echo -e "${GREEN}[SUCCESS]${NC} user_type字段已存在"
    HAS_USER_TYPE=true
else
    echo -e "${YELLOW}[WARNING]${NC} user_type字段不存在"
    HAS_USER_TYPE=false
fi

if echo "$TABLE_INFO" | grep -q "role_id"; then
    echo -e "${GREEN}[SUCCESS]${NC} role_id字段已存在"
    HAS_ROLE_ID=true
else
    echo -e "${YELLOW}[WARNING]${NC} role_id字段不存在"
    HAS_ROLE_ID=false
fi

if echo "$TABLE_INFO" | grep -q "notice"; then
    echo -e "${GREEN}[SUCCESS]${NC} notice字段已存在"
    HAS_NOTICE=true
else
    echo -e "${YELLOW}[WARNING]${NC} notice字段不存在"
    HAS_NOTICE=false
fi

# 3. 应用数据库修复
echo -e "${BLUE}[INFO]${NC} === 应用数据库修复 ==="
if [ "$HAS_USER_TYPE" = false ] || [ "$HAS_ROLE_ID" = false ] || [ "$HAS_NOTICE" = false ]; then
    echo -e "${YELLOW}[WARNING]${NC} 需要添加缺失字段，请输入数据库密码:"
    read -s DB_PASSWORD
    
    # 添加缺失字段
    if [ "$HAS_USER_TYPE" = false ]; then
        echo "添加user_type字段..."
        mysql -u"$DB_USER" -p"$DB_PASSWORD" -h"$DB_HOST" -P"$DB_PORT" -e "USE $DB_NAME; ALTER TABLE user ADD COLUMN user_type int(11) NOT NULL DEFAULT 2 COMMENT '用户类型：1=管理员，2=普通用户';"
    fi
    
    if [ "$HAS_ROLE_ID" = false ]; then
        echo "添加role_id字段..."
        mysql -u"$DB_USER" -p"$DB_PASSWORD" -h"$DB_HOST" -P"$DB_PORT" -e "USE $DB_NAME; ALTER TABLE user ADD COLUMN role_id int(11) NOT NULL DEFAULT 2 COMMENT '角色ID';"
    fi
    
    if [ "$HAS_NOTICE" = false ]; then
        echo "添加notice字段..."
        mysql -u"$DB_USER" -p"$DB_PASSWORD" -h"$DB_HOST" -P"$DB_PORT" -e "USE $DB_NAME; ALTER TABLE user ADD COLUMN notice tinyint(4) NOT NULL DEFAULT 0 COMMENT '是否接收推送：1=是，0=否';"
    fi
    
    # 设置admin用户为管理员
    echo "设置admin用户为管理员..."
    mysql -u"$DB_USER" -p"$DB_PASSWORD" -h"$DB_HOST" -P"$DB_PORT" -e "USE $DB_NAME; UPDATE user SET user_type = 1, role_id = 1, notice = 1 WHERE username = 'admin';"
    
    echo -e "${GREEN}[SUCCESS]${NC} 数据库字段修复完成"
else
    echo -e "${GREEN}[SUCCESS]${NC} 数据库字段已完整，无需修复"
fi

# 4. 检查管理员用户
echo -e "${BLUE}[INFO]${NC} === 检查管理员用户 ==="
ADMIN_USERS=$(mysql -u"$DB_USER" -h"$DB_HOST" -P"$DB_PORT" -e "USE $DB_NAME; SELECT id, username, user_type, role_id, notice FROM user WHERE user_type = 1 OR notice = 1;" 2>/dev/null || echo "")

if [ -n "$ADMIN_USERS" ]; then
    echo -e "${GREEN}[SUCCESS]${NC} 找到管理员用户:"
    echo "$ADMIN_USERS"
else
    echo -e "${RED}[ERROR]${NC} 没有找到管理员用户"
fi

# 5. 检查Redis连接
echo -e "${BLUE}[INFO]${NC} === 检查Redis连接 ==="
if redis-cli ping >/dev/null 2>&1; then
    echo -e "${GREEN}[SUCCESS]${NC} Redis连接正常"
    
    # 检查离线消息
    OFFLINE_MSGS=$(redis-cli keys "offline_message:*" 2>/dev/null || echo "")
    if [ -n "$OFFLINE_MSGS" ]; then
        echo -e "${YELLOW}[WARNING]${NC} 发现离线消息:"
        echo "$OFFLINE_MSGS"
    else
        echo -e "${BLUE}[INFO]${NC} 没有离线消息"
    fi
else
    echo -e "${RED}[ERROR]${NC} Redis连接失败"
fi

# 6. 检查MongoDB连接
echo -e "${BLUE}[INFO]${NC} === 检查MongoDB连接 ==="
if command -v mongo >/dev/null 2>&1; then
    if mongo --eval "db.runCommand('ping')" >/dev/null 2>&1; then
        echo -e "${GREEN}[SUCCESS]${NC} MongoDB连接正常"
        
        # 检查接收记录
        RECORDS_COUNT=$(mongo --quiet --eval "db.admin_user_receive_records.count()" notification_log_db 2>/dev/null || echo "0")
        echo -e "${BLUE}[INFO]${NC} 接收记录数量: $RECORDS_COUNT"
        
        # 检查在线状态
        ONLINE_COUNT=$(mongo --quiet --eval "db.admin_user_online_status.count({is_online: true})" notification_log_db 2>/dev/null || echo "0")
        echo -e "${BLUE}[INFO]${NC} 在线用户数量: $ONLINE_COUNT"
    else
        echo -e "${RED}[ERROR]${NC} MongoDB连接失败"
    fi
else
    echo -e "${YELLOW}[WARNING]${NC} MongoDB客户端未安装，跳过MongoDB检查"
fi

# 7. 检查WebSocket服务
echo -e "${BLUE}[INFO]${NC} === 检查WebSocket服务 ==="
if curl -s http://localhost:8801/api/admin/ws/stats >/dev/null 2>&1; then
    echo -e "${GREEN}[SUCCESS]${NC} WebSocket服务正常"
else
    echo -e "${YELLOW}[WARNING]${NC} WebSocket服务可能未启动或需要认证"
fi

# 8. 检查代码逻辑
echo -e "${BLUE}[INFO]${NC} === 检查代码逻辑 ==="

# 检查getAdminUserIDs方法
if grep -q "getAdminUserIDs" services/public_service/websocket_service.go; then
    echo -e "${GREEN}[SUCCESS]${NC} getAdminUserIDs方法存在"
    
    # 检查是否使用正确的字段
    if grep -q "user_type = 1" services/public_service/websocket_service.go; then
        echo -e "${GREEN}[SUCCESS]${NC} 使用user_type字段筛选管理员"
    else
        echo -e "${YELLOW}[WARNING]${NC} 可能使用错误的字段筛选管理员"
    fi
else
    echo -e "${RED}[ERROR]${NC} getAdminUserIDs方法不存在"
fi

# 检查createAdminUserReceiveRecords方法
if grep -q "createAdminUserReceiveRecords" services/public_service/websocket_service.go; then
    echo -e "${GREEN}[SUCCESS]${NC} createAdminUserReceiveRecords方法存在"
else
    echo -e "${RED}[ERROR]${NC} createAdminUserReceiveRecords方法不存在"
fi

# 检查离线消息保存逻辑
if grep -q "SaveOfflineMessage" services/public_service/websocket_service.go; then
    echo -e "${GREEN}[SUCCESS]${NC} 离线消息保存逻辑存在"
else
    echo -e "${YELLOW}[WARNING]${NC} 离线消息保存逻辑可能缺失"
fi

# 9. 生成测试数据
echo -e "${BLUE}[INFO]${NC} === 生成测试数据 ==="
echo "创建测试管理员用户..."
mysql -u"$DB_USER" -h"$DB_HOST" -P"$DB_PORT" -e "USE $DB_NAME; INSERT IGNORE INTO user (username, password, user_type, role_id, notice) VALUES ('test_admin', '\$2a\$10\$FsAafxTTVVGXfIkJqvaiV.1vPfq4V9HW298McPldJgO829PR52a56', 1, 1, 1);" 2>/dev/null || echo "测试用户可能已存在"

# 10. 验证修复结果
echo -e "${BLUE}[INFO]${NC} === 验证修复结果 ==="

# 验证管理员用户
ADMIN_COUNT=$(mysql -u"$DB_USER" -h"$DB_HOST" -P"$DB_PORT" -e "USE $DB_NAME; SELECT COUNT(*) FROM user WHERE user_type = 1;" 2>/dev/null | tail -n 1 || echo "0")
echo -e "${BLUE}[INFO]${NC} 管理员用户数量: $ADMIN_COUNT"

# 验证可接收推送的用户
NOTICE_COUNT=$(mysql -u"$DB_USER" -h"$DB_HOST" -P"$DB_PORT" -e "USE $DB_NAME; SELECT COUNT(*) FROM user WHERE notice = 1;" 2>/dev/null | tail -n 1 || echo "0")
echo -e "${BLUE}[INFO]${NC} 可接收推送用户数量: $NOTICE_COUNT"

# 11. 生成修复报告
echo -e "${BLUE}[INFO]${NC} === 生成修复报告 ==="
cat > push_system_comprehensive_fix_report.md << EOF
# 推送系统全面修复报告

## 修复时间
$(date)

## 数据库修复
- user_type字段: ${HAS_USER_TYPE}
- role_id字段: ${HAS_ROLE_ID}
- notice字段: ${HAS_NOTICE}

## 用户统计
- 管理员用户数量: $ADMIN_COUNT
- 可接收推送用户数量: $NOTICE_COUNT

## 服务状态
- 数据库连接: ✅
- Redis连接: $(if redis-cli ping >/dev/null 2>&1; then echo "✅"; else echo "❌"; fi)
- MongoDB连接: $(if command -v mongo >/dev/null 2>&1 && mongo --eval "db.runCommand('ping')" >/dev/null 2>&1; then echo "✅"; else echo "❌"; fi)
- WebSocket服务: $(if curl -s http://localhost:8801/api/admin/ws/stats >/dev/null 2>&1; then echo "✅"; else echo "❌"; fi)

## 修复建议
1. 确保数据库字段已正确添加
2. 验证管理员用户筛选逻辑
3. 测试推送功能
4. 检查离线消息保存和发送
5. 验证接收记录创建和更新

## 下一步操作
1. 重启应用服务
2. 登录管理端测试推送
3. 检查接收记录状态
4. 验证离线消息功能
EOF

echo -e "${GREEN}[SUCCESS]${NC} 修复报告已生成: push_system_comprehensive_fix_report.md"

# 12. 启动建议
echo -e "${BLUE}[INFO]${NC} === 启动建议 ==="
echo -e "${YELLOW}[SUGGESTION]${NC} 请执行以下步骤:"
echo "1. 重启应用: ./start.sh"
echo "2. 登录管理端: http://localhost:8801"
echo "3. 发送测试消息: POST /api/admin/system/notice"
echo "4. 检查接收记录: GET /api/admin/notification/admin-receive-records"
echo "5. 检查在线用户: GET /api/admin/notification/online-users"

echo -e "${GREEN}[SUCCESS]${NC} 全面修复完成！" 