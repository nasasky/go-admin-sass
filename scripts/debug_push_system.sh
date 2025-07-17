#!/bin/bash

# 调试推送系统问题

echo "🔍 调试推送系统问题..."

# 配置
API_BASE="http://localhost:8080/api/admin"
LOG_FILE="push_system_debug.log"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

# 清理日志文件
> "$LOG_FILE"

# 检查数据库中的管理员用户
check_admin_users() {
    log_info "=== 检查数据库中的管理员用户 ==="
    
    # 检查user表结构
    log_info "检查user表结构..."
    if command -v mysql &> /dev/null; then
        mysql -u root -p -e "DESCRIBE user;" 2>/dev/null | grep -E "(notice|user_type)" || log_warning "未找到notice或user_type字段"
    fi
    
    # 检查管理员用户数量
    log_info "检查管理员用户数量..."
    if command -v mysql &> /dev/null; then
        echo "user_type=1的用户:" && mysql -u root -p -e "SELECT COUNT(*) FROM user WHERE user_type = 1;" 2>/dev/null
        echo "notice=1的用户:" && mysql -u root -p -e "SELECT COUNT(*) FROM user WHERE notice = 1;" 2>/dev/null
    fi
}

# 检查Redis中的离线消息
check_offline_messages() {
    log_info "=== 检查Redis中的离线消息 ==="
    
    if command -v redis-cli &> /dev/null; then
        # 查找所有离线消息key
        offline_keys=$(redis-cli keys "offline_msg:*" 2>/dev/null)
        if [ -n "$offline_keys" ]; then
            log_info "找到离线消息keys:"
            echo "$offline_keys" | while read -r key; do
                count=$(redis-cli llen "$key" 2>/dev/null)
                log_info "  $key: $count 条消息"
            done
        else
            log_warning "未找到离线消息"
        fi
    else
        log_warning "redis-cli未安装，跳过Redis检查"
    fi
}

# 检查MongoDB中的接收记录
check_mongodb_records() {
    log_info "=== 检查MongoDB中的接收记录 ==="
    
    if command -v mongo &> /dev/null; then
        # 检查接收记录数量
        record_count=$(mongo --quiet --eval "db.admin_user_receive_records.count()" notification_log_db 2>/dev/null)
        log_info "接收记录总数: $record_count"
        
        # 检查最近的记录
        log_info "最近的5条接收记录:"
        mongo --quiet --eval "db.admin_user_receive_records.find().sort({created_at: -1}).limit(5).pretty()" notification_log_db 2>/dev/null
        
        # 检查在线状态记录
        online_count=$(mongo --quiet --eval "db.admin_user_online_status.count({is_online: true})" notification_log_db 2>/dev/null)
        log_info "在线用户数: $online_count"
        
        log_info "在线状态记录:"
        mongo --quiet --eval "db.admin_user_online_status.find({is_online: true}).pretty()" notification_log_db 2>/dev/null
    else
        log_warning "mongo客户端未安装，跳过MongoDB检查"
    fi
}

# 检查WebSocket连接状态
check_websocket_status() {
    log_info "=== 检查WebSocket连接状态 ==="
    
    # 检查WebSocket端点
    if curl -s -I "$API_BASE/ws" | grep -q "101"; then
        log_success "WebSocket端点正常"
    else
        log_error "WebSocket端点异常"
    fi
    
    # 检查WebSocket统计
    stats_response=$(curl -s "$API_BASE/ws/stats" 2>/dev/null)
    if [ -n "$stats_response" ]; then
        log_info "WebSocket统计: $stats_response"
    else
        log_warning "无法获取WebSocket统计"
    fi
}

# 测试推送功能
test_push_function() {
    log_info "=== 测试推送功能 ==="
    
    # 这里需要实际的token，暂时跳过
    log_warning "需要有效的管理员token才能测试推送功能"
    log_info "建议手动测试以下步骤:"
    log_info "1. 登录管理端获取token"
    log_info "2. 发送测试消息: POST $API_BASE/system/notice"
    log_info "3. 检查接收记录: GET $API_BASE/notification/admin-receive-records"
    log_info "4. 检查在线用户: GET $API_BASE/notification/online-users"
}

# 检查代码中的关键逻辑
check_code_logic() {
    log_info "=== 检查代码逻辑 ==="
    
    # 检查getAdminUserIDs方法
    if grep -q "user_type = 1" services/public_service/websocket_service.go; then
        log_success "getAdminUserIDs使用user_type字段"
    else
        log_error "getAdminUserIDs可能使用错误的字段"
    fi
    
    # 检查createAdminUserReceiveRecords方法
    if grep -q "createAdminUserReceiveRecords" services/public_service/websocket_service.go; then
        log_success "createAdminUserReceiveRecords方法存在"
    else
        log_error "createAdminUserReceiveRecords方法不存在"
    fi
    
    # 检查离线消息保存逻辑
    if grep -q "SaveOfflineMessage" services/public_service/websocket_service.go; then
        log_success "离线消息保存逻辑存在"
    else
        log_error "离线消息保存逻辑不存在"
    fi
    
    # 检查用户连接注册逻辑
    if grep -q "RegisterUserConnection" controllers/public/websocket_controller.go; then
        log_success "用户连接注册逻辑存在"
    else
        log_error "用户连接注册逻辑不存在"
    fi
}

# 生成问题诊断报告
generate_diagnosis_report() {
    log_info "=== 问题诊断报告 ==="
    
    echo "## 可能的问题原因:" | tee -a "$LOG_FILE"
    echo "1. 管理员用户筛选条件错误" | tee -a "$LOG_FILE"
    echo "   - getAdminUserIDs可能使用了错误的字段" | tee -a "$LOG_FILE"
    echo "   - 数据库中没有符合条件的管理员用户" | tee -a "$LOG_FILE"
    echo "" | tee -a "$LOG_FILE"
    
    echo "2. 用户在线状态检测问题" | tee -a "$LOG_FILE"
    echo "   - RegisterUserConnection可能没有被正确调用" | tee -a "$LOG_FILE"
    echo "   - MongoDB中的在线状态记录可能不准确" | tee -a "$LOG_FILE"
    echo "" | tee -a "$LOG_FILE"
    
    echo "3. 离线消息保存问题" | tee -a "$LOG_FILE"
    echo "   - 用户不在线时消息可能没有保存到Redis" | tee -a "$LOG_FILE"
    echo "   - 离线消息发送时可能没有更新接收状态" | tee -a "$LOG_FILE"
    echo "" | tee -a "$LOG_FILE"
    
    echo "4. 接收记录创建问题" | tee -a "$LOG_FILE"
    echo "   - createAdminUserReceiveRecords可能没有被调用" | tee -a "$LOG_FILE"
    echo "   - 接收记录保存时可能有错误" | tee -a "$LOG_FILE"
    echo "" | tee -a "$LOG_FILE"
    
    echo "## 建议的修复步骤:" | tee -a "$LOG_FILE"
    echo "1. 确认数据库中有管理员用户" | tee -a "$LOG_FILE"
    echo "2. 检查getAdminUserIDs的筛选条件" | tee -a "$LOG_FILE"
    echo "3. 验证用户连接注册逻辑" | tee -a "$LOG_FILE"
    echo "4. 测试离线消息保存和发送" | tee -a "$LOG_FILE"
    echo "5. 检查接收记录的创建和更新" | tee -a "$LOG_FILE"
}

# 主函数
main() {
    log_info "开始调试推送系统..."
    
    check_admin_users
    check_offline_messages
    check_mongodb_records
    check_websocket_status
    check_code_logic
    test_push_function
    generate_diagnosis_report
    
    log_success "调试完成！详细报告已保存到 $LOG_FILE"
}

# 运行主函数
main "$@" 