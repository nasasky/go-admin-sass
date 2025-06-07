#!/bin/bash

# NASA Go Admin 停止脚本

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

log_info "停止 NASA Go Admin..."

# 检查PID文件是否存在
if [ ! -f "$PID_FILE" ]; then
    log_warn "未找到PID文件 ($PID_FILE)"
    
    # 尝试通过进程名查找
    PIDS=$(pgrep -f "$APP_NAME" 2>/dev/null || true)
    if [ -n "$PIDS" ]; then
        log_warn "找到运行中的进程，尝试停止..."
        echo "$PIDS" | while read -r pid; do
            log_info "停止进程 PID: $pid"
            kill "$pid" 2>/dev/null || true
        done
        sleep 2
        
        # 强制终止仍在运行的进程
        REMAINING_PIDS=$(pgrep -f "$APP_NAME" 2>/dev/null || true)
        if [ -n "$REMAINING_PIDS" ]; then
            log_warn "强制终止残留进程..."
            echo "$REMAINING_PIDS" | while read -r pid; do
                kill -9 "$pid" 2>/dev/null || true
            done
        fi
        log_info "应用已停止"
    else
        log_info "应用未运行"
    fi
    exit 0
fi

# 读取PID
APP_PID=$(cat "$PID_FILE")

# 检查进程是否还在运行
if ! kill -0 "$APP_PID" 2>/dev/null; then
    log_warn "进程 (PID: $APP_PID) 已不存在"
    rm -f "$PID_FILE"
    log_info "清理PID文件"
    exit 0
fi

# 优雅停止进程
log_info "优雅停止进程 (PID: $APP_PID)..."
kill "$APP_PID"

# 等待进程退出
TIMEOUT=10
COUNTER=0
while kill -0 "$APP_PID" 2>/dev/null && [ $COUNTER -lt $TIMEOUT ]; do
    log_info "等待进程退出... ($COUNTER/$TIMEOUT)"
    sleep 1
    COUNTER=$((COUNTER + 1))
done

# 检查是否成功停止
if kill -0 "$APP_PID" 2>/dev/null; then
    log_warn "进程未在超时时间内退出，强制终止..."
    kill -9 "$APP_PID"
    sleep 1
    
    if kill -0 "$APP_PID" 2>/dev/null; then
        log_error "无法停止进程 (PID: $APP_PID)"
        exit 1
    else
        log_info "进程已被强制终止"
    fi
else
    log_info "进程已优雅退出"
fi

# 清理PID文件
rm -f "$PID_FILE"
log_info "应用已停止" 