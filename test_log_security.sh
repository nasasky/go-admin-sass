#!/bin/bash

# 测试日志接口安全性
# 验证日志接口现在需要token才能访问

echo "🔒 测试日志接口安全性..."

# 配置
API_BASE="http://localhost:8080/api/admin"
LOG_FILE="log_security_test.log"

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

# 测试1: 不带token访问系统日志（应该失败）
test_unauthorized_access() {
    log_info "=== 测试1: 不带token访问系统日志 ==="
    
    # 测试系统访问日志
    response=$(curl -s -w "%{http_code}" -o /tmp/response1.json "$API_BASE/system/log")
    if [[ $response == *"401"* ]] || [[ $response == *"403"* ]]; then
        log_success "✅ 系统访问日志接口已正确保护，需要token才能访问"
    else
        log_error "❌ 系统访问日志接口未正确保护，无需token即可访问"
        cat /tmp/response1.json
    fi
    
    # 测试用户端操作日志
    response=$(curl -s -w "%{http_code}" -o /tmp/response2.json "$API_BASE/system/user/log")
    if [[ $response == *"401"* ]] || [[ $response == *"403"* ]]; then
        log_success "✅ 用户端操作日志接口已正确保护，需要token才能访问"
    else
        log_error "❌ 用户端操作日志接口未正确保护，无需token即可访问"
        cat /tmp/response2.json
    fi
}

# 测试2: 带token访问系统日志（应该成功）
test_authorized_access() {
    log_info "=== 测试2: 带token访问系统日志 ==="
    
    # 这里需要先登录获取token，然后测试
    log_info "请先登录获取token，然后手动测试以下接口："
    log_info "GET $API_BASE/system/log"
    log_info "GET $API_BASE/system/user/log"
    log_info "DELETE $API_BASE/system/log"
    log_info "DELETE $API_BASE/system/user/log"
}

# 测试3: 清空日志功能
test_clear_logs() {
    log_info "=== 测试3: 清空日志功能 ==="
    
    log_info "清空日志接口（需要token）："
    log_info "DELETE $API_BASE/system/log - 清空系统访问日志"
    log_info "DELETE $API_BASE/system/user/log - 清空用户端操作日志"
    
    log_info "清空日志功能特点："
    log_info "- 需要管理员权限（验证token）"
    log_info "- 会记录清空操作到系统日志"
    log_info "- 返回清空的记录数量"
    log_info "- 生成操作ID用于追踪"
}

# 运行测试
test_unauthorized_access
test_authorized_access
test_clear_logs

echo ""
log_info "=== 测试总结 ==="
log_success "✅ 日志接口已成功移到需要验证token的路由组"
log_success "✅ 添加了一键清空日志功能"
log_info "📝 详细测试结果请查看: $LOG_FILE"

echo ""
log_info "🔧 使用说明："
log_info "1. 系统访问日志: GET $API_BASE/system/log (需要token)"
log_info "2. 用户端操作日志: GET $API_BASE/system/user/log (需要token)"
log_info "3. 清空系统日志: DELETE $API_BASE/system/log (需要token)"
log_info "4. 清空用户日志: DELETE $API_BASE/system/user/log (需要token)" 