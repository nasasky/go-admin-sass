-- 收益统计表性能优化索引
-- Author: Revenue Performance System
-- Date: 2024

-- 1. 主要查询优化索引
-- 租户ID + 统计日期的复合索引（用于分页和排序）
CREATE INDEX IF NOT EXISTS idx_merchant_revenue_tenant_date ON merchant_revenue_stats(tenants_id, stat_date DESC);

-- 2. 时间范围查询优化
-- 统计日期索引（用于时间范围查询）
CREATE INDEX IF NOT EXISTS idx_merchant_revenue_date ON merchant_revenue_stats(stat_date);

-- 3. 时间戳索引（用于增量同步和更新时间过滤）
CREATE INDEX IF NOT EXISTS idx_merchant_revenue_time ON merchant_revenue_stats(create_time, update_time);

-- 4. 性能优化建议
/*
性能优化建议：

1. 查询优化：
   - 使用强制索引：FORCE INDEX(idx_merchant_revenue_tenant_date) 
   - 避免使用 SELECT *，建议查询：
     SELECT id, tenants_id, stat_date, total_orders, total_revenue, actual_revenue, paid_orders 
     FROM merchant_revenue_stats 
     WHERE tenants_id = ? 
     ORDER BY stat_date DESC 
     LIMIT ?, ?

2. 缓存策略：
   - 已实现 Redis 缓存，缓存时间 30 分钟
   - 建议按租户ID + 日期范围缓存数据
   - 在数据更新时主动清除相关缓存

3. 数据维护：
   - 定期运行：ANALYZE TABLE merchant_revenue_stats;
   - 监控索引使用情况：
     SHOW INDEX FROM merchant_revenue_stats;
   - 检查慢查询：
     SHOW VARIABLES LIKE '%slow_query%';
     SHOW VARIABLES LIKE '%long_query_time%';

4. 建议添加的字段（如果需要）：
   - 添加 deleted_at 字段实现软删除
   - 添加 remark 字段用于备注
   - 考虑添加 version 字段用于乐观锁
*/ 