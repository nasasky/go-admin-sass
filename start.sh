#!/bin/bash

# NASA Go Admin 简单启动脚本
# 传统Go应用启动方式

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

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

APP_NAME="nasa-go-admin"
PID_FILE="${APP_NAME}.pid"

# 检查应用是否已存在
if [ -f "$APP_NAME" ]; then
    log_info "找到应用程序: $APP_NAME"
else
    log_warn "未找到应用程序，尝试构建..."
    if [ -f "build.sh" ]; then
        ./build.sh
    else
        log_info "使用 go build 构建应用..."
        go build -o $APP_NAME main.go
    fi
fi

# 检查是否已经运行
if [ -f "$PID_FILE" ]; then
    OLD_PID=$(cat $PID_FILE)
    if kill -0 $OLD_PID 2>/dev/null; then
        log_warn "应用已在运行 (PID: $OLD_PID)"
        echo "使用 './stop.sh' 停止现有进程，或者 'make stop'"
        exit 1
    else
        log_warn "发现过期的PID文件，清理中..."
        rm -f $PID_FILE
    fi
fi

# 创建必要目录
log_info "创建必要目录..."
mkdir -p logs/{access,error}
mkdir -p data/uploads
mkdir -p tmp

# 检查基本依赖服务 (可选)
log_info "检查依赖服务..."

# 检查MySQL (可选)
if command -v nc &> /dev/null; then
    if nc -z localhost 3306 2>/dev/null; then
        log_info "✓ MySQL 连接正常"
    else
        log_warn "⚠ MySQL 未运行 (localhost:3306)"
    fi
fi

# 检查Redis (可选)
if command -v nc &> /dev/null; then
    if nc -z localhost 6379 2>/dev/null; then
        log_info "✓ Redis 连接正常"
    else
        log_warn "⚠ Redis 未运行 (localhost:6379)"
    fi
fi

# 启动应用
log_info "启动 NASA Go Admin..."

# 设置环境变量
export GIN_MODE=release

# 启动应用
nohup ./$APP_NAME > logs/access/app.log 2> logs/error/app.log &
APP_PID=$!

# 保存PID
echo $APP_PID > $PID_FILE

log_info "应用已启动"
log_info "PID: $APP_PID"
log_info "访问地址: http://localhost:8801"
log_info "日志位置: logs/"

# 简单健康检查
sleep 2
if kill -0 $APP_PID 2>/dev/null; then
    log_info "✓ 应用运行正常"
    
    # 尝试HTTP健康检查
    if command -v curl &> /dev/null; then
        sleep 3
        if curl -f http://localhost:8801/health >/dev/null 2>&1; then
            log_info "✓ HTTP 健康检查通过"
        else
            log_warn "⚠ HTTP 健康检查失败，应用可能正在启动中"
        fi
    fi
else
    log_error "✗ 应用启动失败"
    rm -f $PID_FILE
    exit 1
fi

echo ""
log_info "使用以下命令管理应用:"
echo "  停止应用: ./stop.sh 或 make stop"
echo "  查看日志: tail -f logs/access/app.log"
echo "  查看错误: tail -f logs/error/app.log" 