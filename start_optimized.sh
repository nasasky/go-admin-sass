#!/bin/bash

# NASA Go Admin 优化启动脚本
# 包含性能优化和监控功能

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# 检查系统资源
check_system_resources() {
    log_info "检查系统资源..."
    
    # 检查CPU核心数
    CPU_CORES=$(nproc)
    log_info "CPU核心数: $CPU_CORES"
    
    # 检查内存
    TOTAL_MEM=$(free -h | awk 'NR==2{printf "%.1f", $2}')
    AVAILABLE_MEM=$(free -h | awk 'NR==2{printf "%.1f", $7}')
    log_info "总内存: ${TOTAL_MEM}GB, 可用内存: ${AVAILABLE_MEM}GB"
    
    # 检查磁盘空间
    DISK_USAGE=$(df -h / | awk 'NR==2{print $5}')
    log_info "磁盘使用率: $DISK_USAGE"
    
    # 资源预警
    if [ "$CPU_CORES" -lt 2 ]; then
        log_warn "CPU核心数较少，建议使用至少2核心的服务器"
    fi
    
    AVAILABLE_MEM_GB=$(free | awk 'NR==2{printf "%.1f", $7/1024/1024}')
    if (( $(echo "$AVAILABLE_MEM_GB < 1.0" | bc -l) )); then
        log_warn "可用内存不足1GB，可能影响性能"
    fi
}

# 设置环境变量
setup_environment() {
    log_info "设置环境变量..."
    
    # 设置Go运行时参数
    export GOGC=100                    # GC触发百分比
    export GOMEMLIMIT=512MiB          # 内存限制
    export GOMAXPROCS=$CPU_CORES      # 最大处理器数
    
    # 设置应用参数
    export GIN_MODE=release            # 生产模式
    export DB_MAX_OPEN_CONNS=$((CPU_CORES * 10))  # 数据库最大连接数
    export DB_MAX_IDLE_CONNS=$((CPU_CORES * 2))   # 数据库最大空闲连接数
    
    log_info "环境变量设置完成"
    log_debug "GOMAXPROCS=$GOMAXPROCS"
    log_debug "DB_MAX_OPEN_CONNS=$DB_MAX_OPEN_CONNS"
    log_debug "DB_MAX_IDLE_CONNS=$DB_MAX_IDLE_CONNS"
}

# 检查依赖服务
check_dependencies() {
    log_info "检查依赖服务..."
    
    # 检查MySQL
    if ! nc -z localhost 3306; then
        log_error "MySQL服务未启动或无法连接"
        exit 1
    fi
    log_info "MySQL连接正常"
    
    # 检查Redis
    if ! nc -z localhost 6379; then
        log_error "Redis服务未启动或无法连接"
        exit 1
    fi
    log_info "Redis连接正常"
    
    # 检查MongoDB (如果配置了的话)
    if nc -z localhost 27017; then
        log_info "MongoDB连接正常"
    else
        log_warn "MongoDB未启动，监控功能可能受影响"
    fi
}

# 优化系统参数
optimize_system() {
    log_info "优化系统参数..."
    
    # 增加文件描述符限制
    ulimit -n 65536
    log_debug "文件描述符限制设置为: $(ulimit -n)"
    
    # 设置TCP参数 (需要root权限)
    if [ "$EUID" -eq 0 ]; then
        echo 'net.core.somaxconn = 65536' >> /etc/sysctl.conf
        echo 'net.ipv4.tcp_max_syn_backlog = 65536' >> /etc/sysctl.conf
        sysctl -p
        log_info "TCP参数优化完成"
    else
        log_warn "非root用户，跳过TCP参数优化"
    fi
}

# 创建必要目录
create_directories() {
    log_info "创建必要目录..."
    
    # 日志目录
    mkdir -p logs/{access,error,gorm,monitoring}
    mkdir -p data/uploads
    mkdir -p tmp/cache
    
    # 设置权限
    chmod 755 logs data tmp
    
    log_info "目录创建完成"
}

# 编译应用
build_application() {
    log_info "编译应用..."
    
    # 设置编译参数
    export CGO_ENABLED=1
    export GOOS=linux
    
    # 编译优化参数
    BUILD_FLAGS="-ldflags='-s -w -X main.Version=$(git describe --tags --always)'"
    
    if go build $BUILD_FLAGS -o nasa-go-admin main.go; then
        log_info "应用编译成功"
    else
        log_error "应用编译失败"
        exit 1
    fi
}

# 启动应用
start_application() {
    log_info "启动应用..."
    
    # 设置启动参数
    APP_ARGS=""
    
    # 后台启动应用
    nohup ./nasa-go-admin $APP_ARGS > logs/access/app.log 2> logs/error/app.log &
    APP_PID=$!
    
    # 保存PID
    echo $APP_PID > nasa-go-admin.pid
    
    log_info "应用已启动，PID: $APP_PID"
    
    # 等待应用启动
    sleep 3
    
    # 检查应用是否正常运行
    if kill -0 $APP_PID 2>/dev/null; then
        log_info "应用运行正常"
    else
        log_error "应用启动失败"
        exit 1
    fi
}

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    # 等待应用完全启动
    sleep 5
    
    # 检查HTTP端点
    if curl -f http://localhost:8801/health >/dev/null 2>&1; then
        log_info "健康检查通过"
        return 0
    else
        log_error "健康检查失败"
        return 1
    fi
}

# 启动监控
start_monitoring() {
    log_info "启动监控系统..."
    
    # 启动系统监控
    nohup ./scripts/system_monitor.sh > logs/monitoring/system.log 2>&1 &
    echo $! > system_monitor.pid
    
    log_info "监控系统已启动"
}

# 显示状态信息
show_status() {
    log_info "应用状态信息:"
    echo "=================================="
    echo "应用PID: $(cat nasa-go-admin.pid 2>/dev/null || echo '未找到')"
    echo "监听端口: 8801"
    echo "访问地址: http://localhost:8801"
    echo "监控地址: http://localhost:8801/metrics"
    echo "日志目录: ./logs/"
    echo "=================================="
}

# 主函数
main() {
    log_info "开始启动NASA Go Admin (优化版本)..."
    
    # 执行各个步骤
    check_system_resources
    setup_environment
    check_dependencies
    optimize_system
    create_directories
    build_application
    start_application
    
    # 健康检查
    if health_check; then
        start_monitoring
        show_status
        log_info "应用启动完成！"
    else
        log_error "应用启动失败，请检查日志"
        exit 1
    fi
}

# 信号处理
trap 'log_warn "收到中断信号，正在清理..."; pkill -f nasa-go-admin; exit 0' INT TERM

# 检查参数
case "${1:-}" in
    "start")
        main
        ;;
    "stop")
        log_info "停止应用..."
        if [ -f nasa-go-admin.pid ]; then
            kill $(cat nasa-go-admin.pid)
            rm -f nasa-go-admin.pid
            log_info "应用已停止"
        else
            log_warn "未找到PID文件"
        fi
        ;;
    "restart")
        $0 stop
        sleep 2
        $0 start
        ;;
    "status")
        if [ -f nasa-go-admin.pid ] && kill -0 $(cat nasa-go-admin.pid) 2>/dev/null; then
            log_info "应用正在运行"
            show_status
        else
            log_warn "应用未运行"
        fi
        ;;
    *)
        echo "使用方法: $0 {start|stop|restart|status}"
        echo ""
        echo "start   - 启动应用"
        echo "stop    - 停止应用"
        echo "restart - 重启应用"
        echo "status  - 查看状态"
        exit 1
        ;;
esac 