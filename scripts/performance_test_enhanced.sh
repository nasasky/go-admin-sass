#!/bin/bash

# 增强版性能测试脚本
# 测试订单服务的稳定性、并发安全性和性能优化效果

echo "========================================"
echo "订单服务增强版性能测试"
echo "========================================"

# 配置参数
BASE_URL="http://localhost:8080"
CONCURRENT_USERS=50
TEST_DURATION=300  # 5分钟
RAMP_UP_TIME=30    # 30秒内逐步增加到最大并发
OUTPUT_DIR="performance_test_results_$(date +%Y%m%d_%H%M%S)"

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

echo "测试配置:"
echo "  基础URL: $BASE_URL"
echo "  并发用户数: $CONCURRENT_USERS"
echo "  测试持续时间: ${TEST_DURATION}秒"
echo "  结果目录: $OUTPUT_DIR"
echo ""

# 检查服务是否可用
echo "1. 检查服务可用性..."
if ! curl -s "$BASE_URL/health" > /dev/null; then
    echo "❌ 服务不可用，请确保服务已启动"
    exit 1
fi
echo "✅ 服务可用"

# 获取基准指标
echo ""
echo "2. 获取基准指标..."
curl -s "$BASE_URL/admin/monitor/metrics" | jq '.' > "$OUTPUT_DIR/baseline_metrics.json"
echo "✅ 基准指标已保存"

# 内存和CPU监控函数
monitor_system() {
    echo "timestamp,cpu_percent,memory_mb,goroutines" > "$OUTPUT_DIR/system_monitor.csv"
    
    while [ -f "/tmp/performance_test_running" ]; do
        timestamp=$(date '+%Y-%m-%d %H:%M:%S')
        cpu_percent=$(ps -p $(pgrep go-admin) -o %cpu= | tr -d ' ' || echo "0")
        memory_mb=$(ps -p $(pgrep go-admin) -o rss= | awk '{print $1/1024}' || echo "0")
        
        # 获取goroutine数量
        goroutines=$(curl -s "$BASE_URL/admin/monitor/health" | jq '.data.goroutines' 2>/dev/null || echo "0")
        
        echo "$timestamp,$cpu_percent,$memory_mb,$goroutines" >> "$OUTPUT_DIR/system_monitor.csv"
        sleep 5
    done
}

# 错误监控函数
monitor_errors() {
    echo "timestamp,database_errors,redis_errors,transaction_rollbacks" > "$OUTPUT_DIR/error_monitor.csv"
    
    prev_db_errors=0
    prev_redis_errors=0
    prev_rollbacks=0
    
    while [ -f "/tmp/performance_test_running" ]; do
        timestamp=$(date '+%Y-%m-%d %H:%M:%S')
        
        metrics=$(curl -s "$BASE_URL/admin/monitor/metrics" 2>/dev/null)
        if [ $? -eq 0 ]; then
            db_errors=$(echo "$metrics" | jq '.data.database_errors // 0')
            redis_errors=$(echo "$metrics" | jq '.data.redis_errors // 0')
            rollbacks=$(echo "$metrics" | jq '.data.transaction_rollbacks // 0')
            
            # 计算增量
            db_delta=$((db_errors - prev_db_errors))
            redis_delta=$((redis_errors - prev_redis_errors))
            rollback_delta=$((rollbacks - prev_rollbacks))
            
            echo "$timestamp,$db_delta,$redis_delta,$rollback_delta" >> "$OUTPUT_DIR/error_monitor.csv"
            
            prev_db_errors=$db_errors
            prev_redis_errors=$redis_errors
            prev_rollbacks=$rollbacks
        fi
        
        sleep 10
    done
}

# 响应时间监控函数
monitor_response_times() {
    echo "timestamp,avg_response_time,max_response_time,active_connections" > "$OUTPUT_DIR/response_monitor.csv"
    
    while [ -f "/tmp/performance_test_running" ]; do
        timestamp=$(date '+%Y-%m-%d %H:%M:%S')
        
        metrics=$(curl -s "$BASE_URL/admin/monitor/metrics" 2>/dev/null)
        if [ $? -eq 0 ]; then
            avg_time=$(echo "$metrics" | jq '.data.average_response_time_ms // 0')
            max_time=$(echo "$metrics" | jq '.data.max_response_time_ms // 0')
            active_conn=$(echo "$metrics" | jq '.data.active_connections // 0')
            
            echo "$timestamp,$avg_time,$max_time,$active_conn" >> "$OUTPUT_DIR/response_monitor.csv"
        fi
        
        sleep 5
    done
}

# 启动监控
echo ""
echo "3. 启动系统监控..."
touch /tmp/performance_test_running

monitor_system &
MONITOR_SYSTEM_PID=$!

monitor_errors &
MONITOR_ERROR_PID=$!

monitor_response_times &
MONITOR_RESPONSE_PID=$!

echo "✅ 系统监控已启动"

# 创建测试数据
echo ""
echo "4. 准备测试数据..."
cat > "$OUTPUT_DIR/test_data.json" << EOF
{
  "goods_id": 1,
  "num": 1,
  "payment_method": "wallet"
}
EOF

# 并发测试函数
run_concurrent_test() {
    local user_id=$1
    local test_count=$2
    local output_file="$OUTPUT_DIR/user_${user_id}_results.log"
    
    echo "用户 $user_id 开始测试，预计执行 $test_count 次请求" >> "$output_file"
    
    for ((i=1; i<=test_count; i++)); do
        start_time=$(date +%s%3N)
        
        # 创建订单测试
        response=$(curl -s -w "%{http_code}|%{time_total}" \
            -X POST \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer test-token-$user_id" \
            -d @"$OUTPUT_DIR/test_data.json" \
            "$BASE_URL/app/order/create" 2>/dev/null)
        
        end_time=$(date +%s%3N)
        duration=$((end_time - start_time))
        
        http_code=$(echo "$response" | cut -d'|' -f1)
        curl_time=$(echo "$response" | cut -d'|' -f2)
        
        timestamp=$(date '+%Y-%m-%d %H:%M:%S')
        
        if [ "$http_code" = "200" ]; then
            echo "$timestamp,SUCCESS,$duration,$curl_time" >> "$output_file"
        else
            echo "$timestamp,ERROR,$duration,$curl_time,HTTP_$http_code" >> "$output_file"
        fi
        
        # 随机延迟，模拟真实用户行为
        sleep_time=$(awk "BEGIN {printf \"%.2f\", rand() * 2}")
        sleep "$sleep_time"
        
        # 检查是否应该停止
        if [ ! -f "/tmp/performance_test_running" ]; then
            break
        fi
    done
    
    echo "用户 $user_id 测试完成" >> "$output_file"
}

# 启动并发测试
echo ""
echo "5. 启动并发测试..."
echo "开始时间: $(date)"

# 计算每个用户的请求数量
requests_per_user=$((TEST_DURATION / 3))  # 每3秒一个请求

# 分批启动用户（模拟渐进式负载）
batch_size=10
for ((batch=0; batch<CONCURRENT_USERS; batch+=batch_size)); do
    echo "启动第 $((batch/batch_size + 1)) 批用户 (${batch}-$((batch+batch_size-1)))"
    
    for ((user=batch; user<batch+batch_size && user<CONCURRENT_USERS; user++)); do
        run_concurrent_test $user $requests_per_user &
        echo $! >> "$OUTPUT_DIR/test_pids.txt"
    done
    
    # 渐进式启动延迟
    if [ $batch -lt $CONCURRENT_USERS ]; then
        sleep $((RAMP_UP_TIME / (CONCURRENT_USERS / batch_size)))
    fi
done

echo "✅ 所有并发用户已启动"

# 等待测试完成
echo ""
echo "6. 等待测试完成..."
echo "测试将持续 $TEST_DURATION 秒..."

sleep $TEST_DURATION

# 停止测试
echo ""
echo "7. 停止测试..."
rm -f /tmp/performance_test_running

# 等待所有进程结束
if [ -f "$OUTPUT_DIR/test_pids.txt" ]; then
    while read pid; do
        kill $pid 2>/dev/null
    done < "$OUTPUT_DIR/test_pids.txt"
fi

# 停止监控进程
kill $MONITOR_SYSTEM_PID $MONITOR_ERROR_PID $MONITOR_RESPONSE_PID 2>/dev/null

echo "✅ 测试已停止"

# 收集最终指标
echo ""
echo "8. 收集最终指标..."
curl -s "$BASE_URL/admin/monitor/metrics" | jq '.' > "$OUTPUT_DIR/final_metrics.json"
curl -s "$BASE_URL/admin/monitor/performance" | jq '.' > "$OUTPUT_DIR/performance_report.json"
curl -s "$BASE_URL/admin/monitor/alerts" | jq '.' > "$OUTPUT_DIR/alerts.json"

echo "✅ 最终指标已收集"

# 生成测试报告
echo ""
echo "9. 生成测试报告..."

cat > "$OUTPUT_DIR/test_report.md" << EOF
# 订单服务性能测试报告

## 测试配置
- 测试时间: $(date)
- 并发用户数: $CONCURRENT_USERS
- 测试持续时间: ${TEST_DURATION}秒
- 渐进式启动时间: ${RAMP_UP_TIME}秒

## 测试结果摘要

### 请求统计
EOF

# 统计总请求数和成功率
total_requests=0
successful_requests=0
failed_requests=0

for user_file in "$OUTPUT_DIR"/user_*_results.log; do
    if [ -f "$user_file" ]; then
        user_total=$(grep -c "SUCCESS\|ERROR" "$user_file" 2>/dev/null || echo "0")
        user_success=$(grep -c "SUCCESS" "$user_file" 2>/dev/null || echo "0")
        user_failed=$(grep -c "ERROR" "$user_file" 2>/dev/null || echo "0")
        
        total_requests=$((total_requests + user_total))
        successful_requests=$((successful_requests + user_success))
        failed_requests=$((failed_requests + user_failed))
    fi
done

success_rate=0
if [ $total_requests -gt 0 ]; then
    success_rate=$(awk "BEGIN {printf \"%.2f\", ($successful_requests / $total_requests) * 100}")
fi

cat >> "$OUTPUT_DIR/test_report.md" << EOF
- 总请求数: $total_requests
- 成功请求: $successful_requests
- 失败请求: $failed_requests
- 成功率: ${success_rate}%
- 平均QPS: $(awk "BEGIN {printf \"%.2f\", $total_requests / $TEST_DURATION}")

### 系统资源使用
EOF

# 分析系统监控数据
if [ -f "$OUTPUT_DIR/system_monitor.csv" ]; then
    avg_cpu=$(awk -F',' 'NR>1 {sum+=$2; count++} END {if(count>0) printf "%.2f", sum/count; else print "0"}' "$OUTPUT_DIR/system_monitor.csv")
    max_memory=$(awk -F',' 'NR>1 {if($3>max) max=$3} END {printf "%.2f", max}' "$OUTPUT_DIR/system_monitor.csv")
    max_goroutines=$(awk -F',' 'NR>1 {if($4>max) max=$4} END {print max}' "$OUTPUT_DIR/system_monitor.csv")
    
    cat >> "$OUTPUT_DIR/test_report.md" << EOF
- 平均CPU使用率: ${avg_cpu}%
- 最大内存使用: ${max_memory}MB
- 最大Goroutine数: $max_goroutines
EOF
fi

cat >> "$OUTPUT_DIR/test_report.md" << EOF

### 错误统计
EOF

# 分析错误数据
if [ -f "$OUTPUT_DIR/error_monitor.csv" ]; then
    total_db_errors=$(awk -F',' 'NR>1 {sum+=$2} END {print sum+0}' "$OUTPUT_DIR/error_monitor.csv")
    total_redis_errors=$(awk -F',' 'NR>1 {sum+=$3} END {print sum+0}' "$OUTPUT_DIR/error_monitor.csv")
    total_rollbacks=$(awk -F',' 'NR>1 {sum+=$4} END {print sum+0}' "$OUTPUT_DIR/error_monitor.csv")
    
    cat >> "$OUTPUT_DIR/test_report.md" << EOF
- 数据库错误: $total_db_errors
- Redis错误: $total_redis_errors
- 事务回滚: $total_rollbacks
EOF
fi

cat >> "$OUTPUT_DIR/test_report.md" << EOF

## 详细数据文件
- 基准指标: baseline_metrics.json
- 最终指标: final_metrics.json
- 性能报告: performance_report.json
- 系统监控: system_monitor.csv
- 错误监控: error_monitor.csv
- 响应时间监控: response_monitor.csv
- 报警信息: alerts.json

## 建议
根据测试结果，建议关注以下方面：
1. 如果成功率低于95%，需要检查错误日志
2. 如果平均响应时间超过500ms，需要优化查询性能
3. 如果内存使用持续增长，需要检查内存泄漏
4. 如果Goroutine数量过多，需要优化并发控制
EOF

echo "✅ 测试报告已生成"

# 最终总结
echo ""
echo "========================================"
echo "性能测试完成!"
echo "========================================"
echo "测试结果摘要:"
echo "  总请求数: $total_requests"
echo "  成功率: ${success_rate}%"
echo "  平均QPS: $(awk "BEGIN {printf \"%.2f\", $total_requests / $TEST_DURATION}")"
echo ""
echo "详细报告请查看: $OUTPUT_DIR/test_report.md"
echo "监控数据目录: $OUTPUT_DIR/"
echo ""
echo "建议操作:"
echo "1. 查看测试报告: cat $OUTPUT_DIR/test_report.md"
echo "2. 检查最终指标: cat $OUTPUT_DIR/final_metrics.json"
echo "3. 查看报警信息: cat $OUTPUT_DIR/alerts.json"
echo "========================================" 