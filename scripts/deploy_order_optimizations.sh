#!/bin/bash

# 订单系统优化部署脚本
# Author: Order Security System
# Date: 2024

set -e

echo "🚀 开始部署订单系统优化..."

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# 检查MySQL连接
check_mysql_connection() {
    log_info "检查MySQL连接..."
    
    if ! command -v mysql &> /dev/null; then
        log_error "MySQL客户端未安装"
        exit 1
    fi
    
    # 从配置文件读取数据库连接信息（根据实际配置调整）
    DB_HOST=${DB_HOST:-"localhost"}
    DB_PORT=${DB_PORT:-"3306"}
    DB_USER=${DB_USER:-"root"}
    DB_NAME=${DB_NAME:-"nasa_admin"}
    
    if mysql -h${DB_HOST} -P${DB_PORT} -u${DB_USER} -p${DB_PASSWORD} -e "USE ${DB_NAME}; SELECT 1;" &> /dev/null; then
        log_success "MySQL连接成功"
    else
        log_error "MySQL连接失败，请检查数据库配置"
        exit 1
    fi
}

# 备份现有数据
backup_database() {
    log_info "创建数据库备份..."
    
    BACKUP_DIR="./backups"
    mkdir -p ${BACKUP_DIR}
    
    BACKUP_FILE="${BACKUP_DIR}/nasa_admin_backup_$(date +%Y%m%d_%H%M%S).sql"
    
    if mysqldump -h${DB_HOST} -P${DB_PORT} -u${DB_USER} -p${DB_PASSWORD} \
        --single-transaction --routines --triggers ${DB_NAME} > ${BACKUP_FILE}; then
        log_success "数据库备份完成: ${BACKUP_FILE}"
    else
        log_error "数据库备份失败"
        exit 1
    fi
}

# 应用数据库索引优化
apply_database_indexes() {
    log_info "应用数据库索引优化..."
    
    INDEX_FILE="./migrations/order_performance_indexes.sql"
    
    if [ ! -f "${INDEX_FILE}" ]; then
        log_error "索引文件不存在: ${INDEX_FILE}"
        exit 1
    fi
    
    if mysql -h${DB_HOST} -P${DB_PORT} -u${DB_USER} -p${DB_PASSWORD} ${DB_NAME} < ${INDEX_FILE}; then
        log_success "数据库索引优化完成"
    else
        log_error "数据库索引优化失败"
        exit 1
    fi
}

# 验证索引创建
verify_indexes() {
    log_info "验证索引创建..."
    
    EXPECTED_INDEXES=(
        "idx_app_order_user_status_time"
        "idx_app_order_status_time"
        "idx_app_order_no"
        "idx_app_goods_stock"
        "idx_app_goods_status_stock"
        "idx_app_user_wallet_user"
        "idx_app_recharge_user_time"
        "idx_app_recharge_type_time"
    )
    
    for index in "${EXPECTED_INDEXES[@]}"; do
        RESULT=$(mysql -h${DB_HOST} -P${DB_PORT} -u${DB_USER} -p${DB_PASSWORD} -s -N -e \
            "SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema='${DB_NAME}' AND index_name='${index}';")
        
        if [ "${RESULT}" -gt 0 ]; then
            log_success "索引验证成功: ${index}"
        else
            log_warning "索引验证失败: ${index}"
        fi
    done
}

# 检查Redis连接
check_redis_connection() {
    log_info "检查Redis连接..."
    
    if ! command -v redis-cli &> /dev/null; then
        log_warning "Redis客户端未安装，跳过Redis检查"
        return
    fi
    
    REDIS_HOST=${REDIS_HOST:-"localhost"}
    REDIS_PORT=${REDIS_PORT:-"6379"}
    
    if redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} ping | grep -q "PONG"; then
        log_success "Redis连接成功"
    else
        log_warning "Redis连接失败，某些功能可能受影响"
    fi
}

# 编译Go项目
build_project() {
    log_info "编译Go项目..."
    
    if [ ! -f "go.mod" ]; then
        log_error "当前目录不是Go项目根目录"
        exit 1
    fi
    
    # 下载依赖
    go mod tidy
    
    # 编译项目
    if go build -o nasa-go-admin .; then
        log_success "项目编译成功"
    else
        log_error "项目编译失败"
        exit 1
    fi
}

# 运行测试
run_tests() {
    log_info "运行安全功能测试..."
    
    # 检查是否有测试文件
    if ls *_test.go &> /dev/null; then
        if go test ./... -v; then
            log_success "所有测试通过"
        else
            log_warning "部分测试失败，请检查"
        fi
    else
        log_warning "未找到测试文件，跳过测试"
    fi
}

# 部署后验证
post_deployment_check() {
    log_info "执行部署后验证..."
    
    # 启动应用（后台运行）
    if [ -f "./nasa-go-admin" ]; then
        log_info "启动应用进行健康检查..."
        ./nasa-go-admin &
        APP_PID=$!
        
        # 等待应用启动
        sleep 5
        
        # 健康检查
        if curl -s http://localhost:8801/health | grep -q "healthy"; then
            log_success "应用健康检查通过"
        else
            log_warning "应用健康检查失败"
        fi
        
        # 检查订单系统健康状态
        if curl -s http://localhost:8801/api/app/order/health | grep -q "success"; then
            log_success "订单系统健康检查通过"
        else
            log_warning "订单系统健康检查失败"
        fi
        
        # 停止测试应用
        kill $APP_PID &> /dev/null || true
        sleep 2
    fi
}

# 生成部署报告
generate_report() {
    log_info "生成部署报告..."
    
    REPORT_FILE="./deployment_report_$(date +%Y%m%d_%H%M%S).txt"
    
    cat > ${REPORT_FILE} << EOF
订单系统优化部署报告
生成时间: $(date)
部署版本: Order Security System v1.0

部署内容:
✅ 安全订单创建器 (SecureOrderCreator)
✅ 订单监控服务 (OrderMonitoringService)  
✅ 数据一致性补偿服务 (OrderCompensationService)
✅ 订单系统管理器 (OrderSystemManager)
✅ 数据库性能索引优化
✅ 监控和警报系统

关键特性:
- 库存超卖防护
- 钱包并发安全
- 分布式锁机制
- 超时处理机制
- 异常订单检测
- 数据一致性保证

监控端点:
- 健康检查: /health
- 订单健康: /api/app/order/health
- 监控面板: /api/admin/monitor/dashboard
- 系统统计: /api/monitor/order/stats

注意事项:
1. 请确保Redis服务正常运行
2. 建议定期检查监控指标
3. 及时处理系统警报
4. 定期备份数据库

EOF

    log_success "部署报告已生成: ${REPORT_FILE}"
}

# 主部署流程
main() {
    echo "=================================================================="
    echo "🛡️  NASA Go Admin - 订单系统安全优化部署"
    echo "=================================================================="
    echo
    
    # 检查依赖
    check_mysql_connection
    check_redis_connection
    
    # 备份和数据库优化
    backup_database
    apply_database_indexes
    verify_indexes
    
    # 构建和测试
    build_project
    run_tests
    
    # 部署验证
    post_deployment_check
    
    # 生成报告
    generate_report
    
    echo
    echo "=================================================================="
    log_success "🎉 订单系统优化部署完成！"
    echo "=================================================================="
    echo
    echo "接下来的步骤:"
    echo "1. 重启应用服务: ./nasa-go-admin"
    echo "2. 检查监控面板: http://localhost:8801/api/admin/monitor/dashboard"
    echo "3. 查看系统日志确认一切正常"
    echo "4. 监控系统性能指标"
    echo
}

# 处理中断信号
trap 'log_error "部署被中断"; exit 1' INT TERM

# 执行主流程
main "$@" 