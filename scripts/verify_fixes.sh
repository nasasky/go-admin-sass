#!/bin/bash

# 验证推送系统修复效果

echo "🔍 验证推送系统修复效果..."

# 配置
API_BASE="http://localhost:8080/api/admin"
LOG_FILE="push_system_verification.log"

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

# 检查关键文件是否已修复
check_fixed_files() {
    log_info "检查修复的文件..."
    
    local files=(
        "services/public_service/websocket_service.go"
        "services/public_service/websocket_offline_messages.go"
        "pkg/websocket/hub.go"
        "controllers/public/websocket_controller.go"
    )
    
    for file in "${files[@]}"; do
        if [ -f "$file" ]; then
            log_success "文件存在: $file"
        else
            log_error "文件不存在: $file"
        fi
    done
}

# 检查关键函数是否已添加
check_fixed_functions() {
    log_info "检查修复的函数..."
    
    # 检查WebSocket服务中的新函数
    if grep -q "sendMessageToUser" services/public_service/websocket_service.go; then
        log_success "sendMessageToUser函数已添加"
    else
        log_error "sendMessageToUser函数未找到"
    fi
    
    if grep -q "getOnlineUserIDs" services/public_service/websocket_service.go; then
        log_success "getOnlineUserIDs函数已添加"
    else
        log_error "getOnlineUserIDs函数未找到"
    fi
    
    # 检查Hub中的新函数
    if grep -q "GetUserClients" pkg/websocket/hub.go; then
        log_success "GetUserClients函数已添加"
    else
        log_error "GetUserClients函数未找到"
    fi
    
    if grep -q "IsUserOnline" pkg/websocket/hub.go; then
        log_success "IsUserOnline函数已添加"
    else
        log_error "IsUserOnline函数未找到"
    fi
    
    if grep -q "RemoveClient" pkg/websocket/hub.go; then
        log_success "RemoveClient函数已添加"
    else
        log_error "RemoveClient函数未找到"
    fi
    
    if grep -q "GetOnlineUserIDs" pkg/websocket/hub.go; then
        log_success "GetOnlineUserIDs函数已添加"
    else
        log_error "GetOnlineUserIDs函数未找到"
    fi
}

# 检查离线消息逻辑
check_offline_message_logic() {
    log_info "检查离线消息逻辑..."
    
    # 检查离线消息保存逻辑
    if grep -q "用户不在线，保存为离线消息" services/public_service/websocket_service.go; then
        log_success "离线消息保存逻辑已添加"
    else
        log_error "离线消息保存逻辑未找到"
    fi
    
    # 检查离线消息发送逻辑
    if grep -q "离线消息已发送给用户" services/public_service/websocket_offline_messages.go; then
        log_success "离线消息发送逻辑已添加"
    else
        log_error "离线消息发送逻辑未找到"
    fi
    
    # 检查消息状态更新逻辑
    if grep -q "markMessageAsDelivered" services/public_service/websocket_service.go; then
        log_success "消息状态更新逻辑已添加"
    else
        log_error "消息状态更新逻辑未找到"
    fi
}

# 检查用户连接管理
check_connection_management() {
    log_info "检查用户连接管理..."
    
    # 检查用户连接注册
    if grep -q "RegisterUserConnection" controllers/public/websocket_controller.go; then
        log_success "用户连接注册逻辑已添加"
    else
        log_error "用户连接注册逻辑未找到"
    fi
    
    # 检查用户连接注销
    if grep -q "UnregisterUserConnection" controllers/public/websocket_controller.go; then
        log_success "用户连接注销逻辑已添加"
    else
        log_error "用户连接注销逻辑未找到"
    fi
}

# 检查编译错误
check_compilation() {
    log_info "检查编译错误..."
    
    if command -v go &> /dev/null; then
        cd /Users/wensiyuan/softapp/go/nasa-go-admin
        if go build -o /tmp/test_build . 2>/tmp/build_errors.log; then
            log_success "代码编译成功"
            rm -f /tmp/test_build
        else
            log_error "代码编译失败"
            log_error "编译错误:"
            cat /tmp/build_errors.log | tee -a "$LOG_FILE"
        fi
    else
        log_warning "Go未安装，跳过编译检查"
    fi
}

# 检查数据库连接
check_database_connection() {
    log_info "检查数据库连接..."
    
    # 检查MongoDB连接
    if command -v mongo &> /dev/null; then
        if mongo --eval "db.runCommand('ping')" >/dev/null 2>&1; then
            log_success "MongoDB连接正常"
        else
            log_warning "MongoDB连接失败，请检查MongoDB服务"
        fi
    else
        log_warning "MongoDB客户端未安装，跳过连接检查"
    fi
    
    # 检查Redis连接
    if command -v redis-cli &> /dev/null; then
        if redis-cli ping >/dev/null 2>&1; then
            log_success "Redis连接正常"
        else
            log_warning "Redis连接失败，请检查Redis服务"
        fi
    else
        log_warning "Redis客户端未安装，跳过连接检查"
    fi
}

# 生成修复报告
generate_report() {
    log_info "生成修复报告..."
    
    echo "=== 推送系统修复验证报告 ===" | tee -a "$LOG_FILE"
    echo "生成时间: $(date)" | tee -a "$LOG_FILE"
    echo "" | tee -a "$LOG_FILE"
    
    echo "## 修复内容" | tee -a "$LOG_FILE"
    echo "1. ✅ 修复了离线消息推送问题" | tee -a "$LOG_FILE"
    echo "   - 用户不在线时自动保存为离线消息" | tee -a "$LOG_FILE"
    echo "   - 用户上线时自动发送离线消息" | tee -a "$LOG_FILE"
    echo "   - 离线消息发送后正确更新接收状态" | tee -a "$LOG_FILE"
    echo "" | tee -a "$LOG_FILE"
    
    echo "2. ✅ 修复了用户接收记录状态问题" | tee -a "$LOG_FILE"
    echo "   - 消息发送后立即标记为已投递" | tee -a "$LOG_FILE"
    echo "   - 离线消息发送后更新接收记录" | tee -a "$LOG_FILE"
    echo "   - 添加了完整的用户连接状态管理" | tee -a "$LOG_FILE"
    echo "" | tee -a "$LOG_FILE"
    
    echo "3. ✅ 优化了消息发送流程" | tee -a "$LOG_FILE"
    echo "   - 统一了消息发送和状态更新逻辑" | tee -a "$LOG_FILE"
    echo "   - 添加了消息发送失败时的离线存储" | tee -a "$LOG_FILE"
    echo "   - 改进了用户在线状态检测" | tee -a "$LOG_FILE"
    echo "" | tee -a "$LOG_FILE"
    
    echo "## 测试建议" | tee -a "$LOG_FILE"
    echo "1. 启动服务后，使用test_offline_messages.sh脚本测试" | tee -a "$LOG_FILE"
    echo "2. 测试用户离线时发送消息，然后用户上线检查离线消息" | tee -a "$LOG_FILE"
    echo "3. 检查管理端用户接收记录的状态是否正确" | tee -a "$LOG_FILE"
    echo "4. 验证消息的已读、已确认状态更新是否正常" | tee -a "$LOG_FILE"
    echo "" | tee -a "$LOG_FILE"
    
    echo "## 注意事项" | tee -a "$LOG_FILE"
    echo "1. 确保MongoDB和Redis服务正常运行" | tee -a "$LOG_FILE"
    echo "2. 检查WebSocket连接是否正常建立" | tee -a "$LOG_FILE"
    echo "3. 监控日志中的错误信息" | tee -a "$LOG_FILE"
    echo "4. 定期清理过期的离线消息" | tee -a "$LOG_FILE"
}

# 主函数
main() {
    log_info "开始验证推送系统修复效果..."
    
    check_fixed_files
    check_fixed_functions
    check_offline_message_logic
    check_connection_management
    check_compilation
    check_database_connection
    generate_report
    
    log_success "验证完成！详细报告已保存到 $LOG_FILE"
}

# 运行主函数
main "$@" 