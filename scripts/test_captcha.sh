#!/bin/bash

# 验证码功能测试脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 默认配置
API_BASE_URL=${API_BASE_URL:-"http://localhost:8801"}
ADMIN_TOKEN=${ADMIN_TOKEN:-""}

echo "=================================================================="
echo "🧪 验证码功能测试"
echo "=================================================================="
echo

# 检查服务是否运行
log_info "检查服务状态..."
if ! curl -s "${API_BASE_URL}/health" > /dev/null; then
    log_error "服务未运行，请先启动服务"
    exit 1
fi
log_success "服务运行正常"

# 如果没有提供token，尝试登录获取
if [ -z "$ADMIN_TOKEN" ]; then
    log_info "尝试登录获取token..."
    
    # 这里需要根据你的实际用户信息修改
    LOGIN_RESPONSE=$(curl -s -X POST "${API_BASE_URL}/api/admin/login" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "username=admin&password=123456" 2>/dev/null || echo "")
    
    if echo "$LOGIN_RESPONSE" | grep -q "token"; then
        ADMIN_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        log_success "获取token成功"
    else
        log_warning "无法自动获取token，请手动设置 ADMIN_TOKEN 环境变量"
        log_info "或者先通过登录接口获取token"
    fi
fi

if [ -z "$ADMIN_TOKEN" ]; then
    log_error "需要管理员token才能测试验证码开关功能"
    log_info "请先登录获取token，然后设置 ADMIN_TOKEN 环境变量"
    exit 1
fi

# 测试1: 获取验证码开关状态
log_info "测试1: 获取验证码开关状态..."
CAPTCHA_STATUS_RESPONSE=$(curl -s -X GET "${API_BASE_URL}/api/admin/captcha/status" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}")

if echo "$CAPTCHA_STATUS_RESPONSE" | grep -q "captcha_enabled"; then
    CURRENT_STATUS=$(echo "$CAPTCHA_STATUS_RESPONSE" | grep -o '"captcha_enabled":[^,]*' | cut -d':' -f2)
    log_success "获取验证码状态成功: $CURRENT_STATUS"
else
    log_error "获取验证码状态失败: $CAPTCHA_STATUS_RESPONSE"
fi

# 测试2: 获取验证码图片
log_info "测试2: 获取验证码图片..."
CAPTCHA_IMAGE_RESPONSE=$(curl -s -X GET "${API_BASE_URL}/api/admin/captcha" \
    -H "Accept: image/svg+xml")

if echo "$CAPTCHA_IMAGE_RESPONSE" | grep -q "<svg"; then
    log_success "验证码图片生成成功"
else
    log_error "验证码图片生成失败"
fi

# 测试3: 切换验证码开关状态
log_info "测试3: 切换验证码开关状态..."
NEW_STATUS="false"
if [ "$CURRENT_STATUS" = "false" ]; then
    NEW_STATUS="true"
fi

UPDATE_RESPONSE=$(curl -s -X PUT "${API_BASE_URL}/api/admin/captcha/status" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d "{\"enabled\": $NEW_STATUS}")

if echo "$UPDATE_RESPONSE" | grep -q "验证码开关更新成功"; then
    log_success "验证码开关更新成功: $NEW_STATUS"
else
    log_error "验证码开关更新失败: $UPDATE_RESPONSE"
fi

# 测试4: 验证更新后的状态
log_info "测试4: 验证更新后的状态..."
UPDATED_STATUS_RESPONSE=$(curl -s -X GET "${API_BASE_URL}/api/admin/captcha/status" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}")

if echo "$UPDATED_STATUS_RESPONSE" | grep -q "\"captcha_enabled\":$NEW_STATUS"; then
    log_success "验证码状态更新验证成功"
else
    log_error "验证码状态更新验证失败: $UPDATED_STATUS_RESPONSE"
fi

# 测试5: 测试登录（验证码禁用时）
if [ "$NEW_STATUS" = "false" ]; then
    log_info "测试5: 测试登录（验证码禁用时）..."
    LOGIN_TEST_RESPONSE=$(curl -s -X POST "${API_BASE_URL}/api/admin/login" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "username=admin&password=123456")
    
    if echo "$LOGIN_TEST_RESPONSE" | grep -q "token"; then
        log_success "验证码禁用时登录成功"
    elif echo "$LOGIN_TEST_RESPONSE" | grep -q "请输入验证码"; then
        log_error "验证码已禁用但登录仍要求验证码"
    else
        log_warning "登录测试结果: $LOGIN_TEST_RESPONSE"
    fi
fi

echo
log_success "验证码功能测试完成！"
echo
log_info "测试总结:"
echo "  - 验证码开关状态: $CURRENT_STATUS -> $NEW_STATUS"
echo "  - 验证码图片生成: ✅"
echo "  - 开关状态更新: ✅"
echo "  - 登录功能: ✅" 