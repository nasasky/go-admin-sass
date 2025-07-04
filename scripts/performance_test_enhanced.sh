#!/bin/bash

# 角色管理系统性能测试脚本
# 用于验证查询优化效果

set -e

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
ENV_FILE="$PROJECT_ROOT/.env"
LOG_FILE="$PROJECT_ROOT/performance_test_results.log"

# 日志函数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log "开始性能测试..."

# 读取数据库配置
if [ ! -f "$ENV_FILE" ]; then
    log "错误: 配置文件 .env 不存在"
    exit 1
fi

DB_HOST=$(grep "^DB_HOST=" "$ENV_FILE" | cut -d'=' -f2 | tr -d '"')
DB_PORT=$(grep "^DB_PORT=" "$ENV_FILE" | cut -d'=' -f2 | tr -d '"')
DB_NAME=$(grep "^DB_NAME=" "$ENV_FILE" | cut -d'=' -f2 | tr -d '"')
DB_USER=$(grep "^DB_USER=" "$ENV_FILE" | cut -d'=' -f2 | tr -d '"')
DB_PASSWORD=$(grep "^DB_PASSWORD=" "$ENV_FILE" | cut -d'=' -f2 | tr -d '"')

# 设置默认值
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-3306}

log "数据库配置: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"

# 执行SQL查询并计时的函数
run_query_with_timing() {
    local test_name=$1
    local sql_query=$2
    local iterations=${3:-5}
    
    log "执行测试: $test_name"
    
    local total_time=0
    local fastest_time=999999
    local slowest_time=0
    
    for i in $(seq 1 $iterations); do
        local start_time=$(date +%s.%N)
        
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" \
            -e "$sql_query" > /dev/null 2>&1
        
        local end_time=$(date +%s.%N)
        local duration=$(echo "$end_time - $start_time" | bc -l)
        
        total_time=$(echo "$total_time + $duration" | bc -l)
        
        if (( $(echo "$duration < $fastest_time" | bc -l) )); then
            fastest_time=$duration
        fi
        
        if (( $(echo "$duration > $slowest_time" | bc -l) )); then
            slowest_time=$duration
        fi
        
        printf "  第 %d 次: %.3f 秒\n" $i $duration
    done
    
    local avg_time=$(echo "scale=3; $total_time / $iterations" | bc -l)
    
    log "  平均时间: ${avg_time}秒"
    log "  最快时间: ${fastest_time}秒"
    log "  最慢时间: ${slowest_time}秒"
    log ""
    
    echo $avg_time
}

# 测试函数
test_role_queries() {
    log "=== 角色查询性能测试 ==="
    
    # 1. 基本角色列表查询
    local query1="SELECT id, role_name, role_desc, user_id, user_type, enable, sort, create_time, update_time FROM role ORDER BY id DESC LIMIT 10;"
    local result1=$(run_query_with_timing "基本角色列表查询" "$query1")
    
    # 2. 带权限过滤的查询（非超管用户）
    local query2="SELECT id, role_name, role_desc, user_id, user_type, enable, sort, create_time, update_time FROM role WHERE user_id = 2 ORDER BY id DESC LIMIT 10;"
    local result2=$(run_query_with_timing "权限过滤查询（非超管）" "$query2")
    
    # 3. 角色名称搜索查询
    local query3="SELECT id, role_name, role_desc, user_id, user_type, enable, sort, create_time, update_time FROM role WHERE role_name LIKE '%管理%' ORDER BY id DESC LIMIT 10;"
    local result3=$(run_query_with_timing "角色名称搜索查询" "$query3")
    
    # 4. 复合条件查询
    local query4="SELECT id, role_name, role_desc, user_id, user_type, enable, sort, create_time, update_time FROM role WHERE enable = 1 AND user_type = 1 ORDER BY sort ASC LIMIT 10;"
    local result4=$(run_query_with_timing "复合条件查询" "$query4")
    
    # 5. 批量用户信息查询
    local query5="SELECT id, username FROM user WHERE id IN (1,2,3,4,5,6,7,8,9,10);"
    local result5=$(run_query_with_timing "批量用户信息查询" "$query5")
    
    # 6. 角色权限查询
    local query6="SELECT permissionId FROM role_permissions_permission WHERE roleId IN (1,2,3,4,5);"
    local result6=$(run_query_with_timing "角色权限查询" "$query6")
    
    log "=== 性能测试结果总结 ==="
    printf "基本角色列表查询: %.3f秒\n" $result1
    printf "权限过滤查询: %.3f秒\n" $result2
    printf "角色名称搜索: %.3f秒\n" $result3
    printf "复合条件查询: %.3f秒\n" $result4
    printf "批量用户查询: %.3f秒\n" $result5
    printf "角色权限查询: %.3f秒\n" $result6
}

# 测试索引使用情况
test_index_usage() {
    log "=== 索引使用情况测试 ==="
    
    # 检查EXPLAIN输出
    local queries=(
        "SELECT id, role_name FROM role WHERE user_id = 1 AND user_type = 1"
        "SELECT id, role_name FROM role WHERE role_name LIKE '%管理%'"  
        "SELECT id, role_name FROM role WHERE enable = 1 ORDER BY sort"
        "SELECT username FROM user WHERE id IN (1,2,3,4,5)"
        "SELECT permissionId FROM role_permissions_permission WHERE roleId = 1"
    )
    
    local descriptions=(
        "权限过滤查询"
        "角色名称搜索"
        "状态排序查询"
        "批量用户查询"
        "权限关联查询"
    )
    
    for i in "${!queries[@]}"; do
        log "EXPLAIN ${descriptions[$i]}:"
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" \
            -e "EXPLAIN ${queries[$i]};" 2>/dev/null || true
        log ""
    done
}

# 统计表大小和索引大小
show_table_stats() {
    log "=== 表和索引统计信息 ==="
    
    local tables=("role" "user" "role_permissions_permission")
    
    for table in "${tables[@]}"; do
        log "表 $table 的统计信息:"
        
        # 表大小
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" \
            -e "SELECT 
                TABLE_NAME as '表名',
                TABLE_ROWS as '行数',
                ROUND(DATA_LENGTH/1024/1024, 2) as '数据大小(MB)',
                ROUND(INDEX_LENGTH/1024/1024, 2) as '索引大小(MB)'
                FROM information_schema.TABLES 
                WHERE TABLE_SCHEMA = '$DB_NAME' AND TABLE_NAME = '$table';" 2>/dev/null || true
        
        # 索引信息
        log "表 $table 的索引:"
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" \
            -e "SHOW INDEX FROM $table;" 2>/dev/null || true
        
        log ""
    done
}

# 生成性能基准报告
generate_benchmark_report() {
    log "=== 性能基准报告 ==="
    
    local report_file="$PROJECT_ROOT/performance_benchmark_$(date +%Y%m%d_%H%M%S).md"
    
    cat > "$report_file" << EOF
# 角色管理系统性能基准报告

生成时间: $(date '+%Y-%m-%d %H:%M:%S')
数据库: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME

## 测试环境
- 数据库类型: MySQL
- 测试时间: $(date '+%Y-%m-%d %H:%M:%S')
- 测试机器: $(uname -a)

## 性能优化措施
1. 创建了权限过滤复合索引 (user_id, user_type)
2. 创建了角色名称搜索索引
3. 创建了状态和排序相关索引
4. 创建了用户信息查询索引
5. 创建了权限关联查询索引
6. 优化了查询逻辑，使用并行查询
7. 优化了数据格式化，减少内存分配

## 预期性能提升
- 角色列表查询速度提升 50-80%
- 权限过滤查询速度提升 60-90%  
- 创建人信息批量查询速度提升 40-70%
- 内存使用优化约 20-30%

## 建议
1. 定期监控查询性能
2. 根据实际使用情况调整索引
3. 考虑添加缓存层进一步提升性能
4. 定期分析慢查询日志

---
详细测试结果请查看: $LOG_FILE
EOF

    log "性能基准报告生成完成: $report_file"
}

# 主函数
main() {
    log "开始角色管理系统性能测试"
    log "测试时间: $(date '+%Y-%m-%d %H:%M:%S')"
    log ""
    
    # 检查bc命令
    if ! command -v bc &> /dev/null; then
        log "警告: bc 命令不可用，部分计算功能可能受限"
    fi
    
    # 执行各项测试
    show_table_stats
    test_index_usage  
    test_role_queries
    generate_benchmark_report
    
    log "性能测试完成!"
    log "详细结果已保存到: $LOG_FILE"
}

# 执行主函数
main "$@" 