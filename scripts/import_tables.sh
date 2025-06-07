#!/bin/bash

# 数据库表导入脚本
# 用于快速导入订单系统所需的数据库表

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
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-"3306"}
DB_USER=${DB_USER:-"root"}
DB_NAME=${DB_NAME:-"naive_admin"}
SQL_FILE="./migrations/create_order_tables.sql"

echo "=================================================================="
echo "🚀 数据库表导入工具"
echo "=================================================================="
echo

# 检查SQL文件是否存在
if [ ! -f "${SQL_FILE}" ]; then
    log_error "SQL文件不存在: ${SQL_FILE}"
    log_info "请确保在项目根目录运行此脚本"
    exit 1
fi

# 检查MySQL客户端
if ! command -v mysql &> /dev/null; then
    log_error "MySQL客户端未安装，请先安装MySQL客户端"
    exit 1
fi

log_info "数据库配置:"
echo "  主机: ${DB_HOST}:${DB_PORT}"
echo "  用户: ${DB_USER}"
echo "  数据库: ${DB_NAME}"
echo

# 提示用户输入密码
echo -n "请输入MySQL密码: "
read -s DB_PASSWORD
echo

# 测试数据库连接
log_info "测试数据库连接..."
if ! mysql -h${DB_HOST} -P${DB_PORT} -u${DB_USER} -p${DB_PASSWORD} -e "SELECT 1;" &>/dev/null; then
    log_error "数据库连接失败，请检查配置和密码"
    exit 1
fi
log_success "数据库连接成功"

# 检查数据库是否存在
log_info "检查数据库 ${DB_NAME}..."
DB_EXISTS=$(mysql -h${DB_HOST} -P${DB_PORT} -u${DB_USER} -p${DB_PASSWORD} -e "SHOW DATABASES LIKE '${DB_NAME}';" | grep -c "${DB_NAME}" || true)

if [ "${DB_EXISTS}" -eq 0 ]; then
    log_warning "数据库 ${DB_NAME} 不存在"
    echo -n "是否创建数据库? (y/N): "
    read -r CREATE_DB
    if [[ "${CREATE_DB}" =~ ^[Yy]$ ]]; then
        mysql -h${DB_HOST} -P${DB_PORT} -u${DB_USER} -p${DB_PASSWORD} -e "CREATE DATABASE ${DB_NAME} CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
        log_success "数据库 ${DB_NAME} 创建成功"
    else
        log_error "数据库不存在，导入终止"
        exit 1
    fi
else
    log_success "数据库 ${DB_NAME} 已存在"
fi

# 检查表是否已存在
log_info "检查现有表..."
EXISTING_TABLES=$(mysql -h${DB_HOST} -P${DB_PORT} -u${DB_USER} -p${DB_PASSWORD} ${DB_NAME} -e "SHOW TABLES;" 2>/dev/null | grep -E "(app_order|app_recharge|app_goods|app_user)" || true)

if [ -n "${EXISTING_TABLES}" ]; then
    log_warning "检测到以下相关表已存在:"
    echo "${EXISTING_TABLES}"
    echo
    echo -n "是否继续导入? 现有表将被跳过 (y/N): "
    read -r CONTINUE_IMPORT
    if [[ ! "${CONTINUE_IMPORT}" =~ ^[Yy]$ ]]; then
        log_info "导入已取消"
        exit 0
    fi
fi

# 导入SQL文件
log_info "开始导入数据库表..."
echo "执行SQL文件: ${SQL_FILE}"

if mysql -h${DB_HOST} -P${DB_PORT} -u${DB_USER} -p${DB_PASSWORD} ${DB_NAME} < ${SQL_FILE}; then
    log_success "数据库表导入成功!"
else
    log_error "数据库表导入失败"
    exit 1
fi

# 验证导入结果
log_info "验证导入结果..."
TABLES_COUNT=$(mysql -h${DB_HOST} -P${DB_PORT} -u${DB_USER} -p${DB_PASSWORD} ${DB_NAME} -e "SELECT COUNT(*) as count FROM information_schema.tables WHERE table_schema='${DB_NAME}' AND table_name IN ('app_order', 'app_recharge', 'app_goods', 'app_user', 'app_user_wallet');" -s -N)

if [ "${TABLES_COUNT}" -ge 5 ]; then
    log_success "核心表验证成功，共 ${TABLES_COUNT} 个表"
else
    log_warning "表验证结果: ${TABLES_COUNT}/5 个表"
fi

# 检查测试数据
log_info "检查测试数据..."
GOODS_COUNT=$(mysql -h${DB_HOST} -P${DB_PORT} -u${DB_USER} -p${DB_PASSWORD} ${DB_NAME} -e "SELECT COUNT(*) FROM app_goods;" -s -N 2>/dev/null || echo "0")
USER_COUNT=$(mysql -h${DB_HOST} -P${DB_PORT} -u${DB_USER} -p${DB_PASSWORD} ${DB_NAME} -e "SELECT COUNT(*) FROM app_user;" -s -N 2>/dev/null || echo "0")

echo "  - 测试商品: ${GOODS_COUNT} 个"
echo "  - 测试用户: ${USER_COUNT} 个"

echo
echo "=================================================================="
log_success "🎉 数据库表导入完成！"
echo "=================================================================="
echo
echo "接下来可以:"
echo "1. 启动应用: ./nasa-go-admin"
echo "2. 测试订单接口: curl http://localhost:8801/api/app/order/health"
echo "3. 查看监控面板: http://localhost:8801/api/admin/monitor/dashboard"
echo
echo "测试账号:"
echo "- 用户ID: 1, 余额: 1000.00"
echo "- 用户ID: 2, 余额: 500.00"
echo "- 测试商品ID: 1, 2, 3"
echo 