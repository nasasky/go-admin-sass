#!/bin/bash

# NASA Go Admin 性能测试脚本
# 用于验证优化效果

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
BASE_URL="http://localhost:8801"
CONCURRENT_USERS=100
TEST_DURATION=60
RESULTS_DIR="performance_results"

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

# 检查依赖
check_dependencies() {
    log_info "检查测试依赖..."
    
    # 检查curl
    if ! command -v curl &> /dev/null; then
        log_error "curl未安装"
        exit 1
    fi
    
    # 检查ab (Apache Bench)
    if ! command -v ab &> /dev/null; then
        log_warn "ab未安装，尝试安装..."
        if command -v apt-get &> /dev/null; then
            sudo apt-get update && sudo apt-get install -y apache2-utils
        elif command -v yum &> /dev/null; then
            sudo yum install -y httpd-tools
        else
            log_error "无法自动安装ab，请手动安装apache2-utils或httpd-tools"
            exit 1
        fi
    fi
    
    # 检查wrk (如果可用)
    if command -v wrk &> /dev/null; then
        HAS_WRK=true
        log_info "检测到wrk，将使用wrk进行高级测试"
    else
        HAS_WRK=false
        log_info "未检测到wrk，将使用ab进行基础测试"
    fi
}

# 检查服务状态
check_service() {
    log_info "检查服务状态..."
    
    if curl -f "$BASE_URL/health" >/dev/null 2>&1; then
        log_info "服务运行正常"
    else
        log_error "服务未运行或无法访问"
        exit 1
    fi
}

# 创建结果目录
setup_results_dir() {
    mkdir -p "$RESULTS_DIR"
    TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
    RESULT_FILE="$RESULTS_DIR/performance_test_$TIMESTAMP.txt"
    
    log_info "测试结果将保存到: $RESULT_FILE"
}

# 基础性能测试
basic_performance_test() {
    log_info "开始基础性能测试..."
    
    echo "=== 基础性能测试 ===" >> "$RESULT_FILE"
    echo "测试时间: $(date)" >> "$RESULT_FILE"
    echo "并发用户: $CONCURRENT_USERS" >> "$RESULT_FILE"
    echo "测试时长: ${TEST_DURATION}秒" >> "$RESULT_FILE"
    echo "" >> "$RESULT_FILE"
    
    # 测试健康检查端点
    log_info "测试健康检查端点..."
    ab -n 1000 -c 10 "$BASE_URL/health" >> "$RESULT_FILE" 2>&1
    echo "" >> "$RESULT_FILE"
    
    # 测试监控端点
    log_info "测试监控端点..."
    ab -n 500 -c 5 "$BASE_URL/metrics" >> "$RESULT_FILE" 2>&1
    echo "" >> "$RESULT_FILE"
}

# 高级性能测试 (使用wrk)
advanced_performance_test() {
    if [ "$HAS_WRK" = false ]; then
        log_warn "跳过高级性能测试 (wrk未安装)"
        return
    fi
    
    log_info "开始高级性能测试..."
    
    echo "=== 高级性能测试 (wrk) ===" >> "$RESULT_FILE"
    
    # 测试不同并发级别
    for concurrency in 10 50 100 200; do
        log_info "测试并发数: $concurrency"
        echo "--- 并发数: $concurrency ---" >> "$RESULT_FILE"
        
        wrk -t4 -c$concurrency -d30s --latency "$BASE_URL/health" >> "$RESULT_FILE" 2>&1
        echo "" >> "$RESULT_FILE"
        
        # 等待服务恢复
        sleep 5
    done
}

# 压力测试
stress_test() {
    log_info "开始压力测试..."
    
    echo "=== 压力测试 ===" >> "$RESULT_FILE"
    
    # 逐步增加负载
    for requests in 1000 5000 10000; do
        log_info "测试请求数: $requests"
        echo "--- 请求数: $requests ---" >> "$RESULT_FILE"
        
        ab -n $requests -c 50 "$BASE_URL/health" >> "$RESULT_FILE" 2>&1
        echo "" >> "$RESULT_FILE"
        
        # 检查服务是否还在运行
        if ! curl -f "$BASE_URL/health" >/dev/null 2>&1; then
            log_error "服务在压力测试中停止响应"
            break
        fi
        
        sleep 10
    done
}

# 内存和CPU监控
monitor_resources() {
    log_info "开始资源监控..."
    
    MONITOR_FILE="$RESULTS_DIR/resource_monitor_$TIMESTAMP.txt"
    
    # 启动资源监控
    {
        echo "=== 资源监控 ==="
        echo "时间,CPU%,内存MB,连接数"
        
        for i in {1..60}; do
            CPU=$(top -bn1 | grep "nasa-go-admin" | awk '{print $9}' | head -1)
            MEM=$(ps -o pid,vsz,comm | grep nasa-go-admin | awk '{print $2/1024}')
            CONN=$(ss -tuln | grep :8801 | wc -l)
            
            echo "$(date '+%H:%M:%S'),$CPU,$MEM,$CONN"
            sleep 1
        done
    } > "$MONITOR_FILE" &
    
    MONITOR_PID=$!
    log_info "资源监控已启动 (PID: $MONITOR_PID)"
}

# 停止资源监控
stop_monitor() {
    if [ ! -z "$MONITOR_PID" ]; then
        kill $MONITOR_PID 2>/dev/null || true
        log_info "资源监控已停止"
    fi
}

# 生成报告
generate_report() {
    log_info "生成测试报告..."
    
    REPORT_FILE="$RESULTS_DIR/performance_report_$TIMESTAMP.html"
    
    cat > "$REPORT_FILE" << EOF
<!DOCTYPE html>
<html>
<head>
    <title>NASA Go Admin 性能测试报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f0f0f0; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; padding: 15px; border: 1px solid #ddd; }
        .metric { display: inline-block; margin: 10px; padding: 10px; background: #e8f4f8; }
        pre { background: #f5f5f5; padding: 10px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="header">
        <h1>NASA Go Admin 性能测试报告</h1>
        <p>测试时间: $(date)</p>
        <p>测试配置: 并发用户 $CONCURRENT_USERS, 测试时长 ${TEST_DURATION}秒</p>
    </div>
    
    <div class="section">
        <h2>系统信息</h2>
        <div class="metric">CPU核心: $(nproc)</div>
        <div class="metric">内存: $(free -h | awk 'NR==2{print $2}')</div>
        <div class="metric">Go版本: $(go version)</div>
    </div>
    
    <div class="section">
        <h2>测试结果摘要</h2>
        <p>详细结果请查看: <a href="performance_test_$TIMESTAMP.txt">测试日志</a></p>
        <p>资源监控: <a href="resource_monitor_$TIMESTAMP.txt">监控数据</a></p>
    </div>
    
    <div class="section">
        <h2>优化建议</h2>
        <ul>
            <li>如果响应时间 > 100ms，考虑优化数据库查询</li>
            <li>如果CPU使用率 > 80%，考虑增加服务器资源</li>
            <li>如果内存使用持续增长，检查是否存在内存泄漏</li>
            <li>如果连接数接近限制，调整连接池配置</li>
        </ul>
    </div>
</body>
</html>
EOF
    
    log_info "报告已生成: $REPORT_FILE"
}

# 清理函数
cleanup() {
    stop_monitor
    log_info "测试完成，清理资源..."
}

# 主函数
main() {
    log_info "开始NASA Go Admin性能测试..."
    
    # 设置清理陷阱
    trap cleanup EXIT
    
    # 执行测试步骤
    check_dependencies
    check_service
    setup_results_dir
    
    # 启动资源监控
    monitor_resources
    
    # 执行性能测试
    basic_performance_test
    advanced_performance_test
    stress_test
    
    # 停止监控
    stop_monitor
    
    # 生成报告
    generate_report
    
    log_info "性能测试完成！"
    log_info "查看结果: $RESULT_FILE"
    log_info "查看报告: $REPORT_FILE"
}

# 检查参数
case "${1:-}" in
    "run")
        main
        ;;
    "quick")
        CONCURRENT_USERS=10
        TEST_DURATION=30
        log_info "快速测试模式"
        main
        ;;
    "stress")
        CONCURRENT_USERS=500
        TEST_DURATION=120
        log_info "压力测试模式"
        main
        ;;
    *)
        echo "使用方法: $0 {run|quick|stress}"
        echo ""
        echo "run    - 标准性能测试"
        echo "quick  - 快速测试 (低负载)"
        echo "stress - 压力测试 (高负载)"
        exit 1
        ;;
esac 