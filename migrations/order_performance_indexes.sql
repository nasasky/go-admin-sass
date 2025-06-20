-- 订单性能优化索引
-- Author: Order Security System
-- Date: 2024

-- 1. 订单表相关索引 
-- 用户订单查询优化
CREATE INDEX IF NOT EXISTS idx_app_order_user_status_time ON app_order(user_id, status, create_time);

-- 订单状态查询优化
CREATE INDEX IF NOT EXISTS idx_app_order_status_time ON app_order(status, create_time);

-- 订单号查询优化（如果不存在唯一索引）
CREATE INDEX IF NOT EXISTS idx_app_order_no ON app_order(no);

-- 过期订单清理优化
CREATE INDEX IF NOT EXISTS idx_app_order_pending_expired ON app_order(status, create_time) WHERE status = 'pending';

-- 2. 商品库存相关索引
-- 商品库存查询优化
CREATE INDEX IF NOT EXISTS idx_app_goods_stock ON app_goods(stock);

-- 商品库存和状态复合索引
CREATE INDEX IF NOT EXISTS idx_app_goods_status_stock ON app_goods(status, stock);

-- 3. 用户钱包相关索引
-- 用户钱包查询优化
CREATE INDEX IF NOT EXISTS idx_app_user_wallet_user ON app_user_wallet(user_id);

-- 钱包余额查询优化
CREATE INDEX IF NOT EXISTS idx_app_user_wallet_balance ON app_user_wallet(balance);

-- 4. 交易记录相关索引
-- 用户交易记录查询优化
CREATE INDEX IF NOT EXISTS idx_app_recharge_user_time ON app_recharge(user_id, create_time);

-- 交易类型和时间复合索引
CREATE INDEX IF NOT EXISTS idx_app_recharge_type_time ON app_recharge(transaction_type, create_time);

-- 支付状态查询优化
CREATE INDEX IF NOT EXISTS idx_app_recharge_status_time ON app_recharge(status, create_time);

-- 异常支付模式检测优化
CREATE INDEX IF NOT EXISTS idx_app_recharge_user_type_time ON app_recharge(user_id, transaction_type, create_time);

-- 5. 如果使用了订单商品关联表
-- 订单商品关联查询优化
CREATE INDEX IF NOT EXISTS idx_app_order_goods_order ON app_order_goods(order_id) WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'app_order_goods');
CREATE INDEX IF NOT EXISTS idx_app_order_goods_goods ON app_order_goods(goods_id) WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'app_order_goods');

-- 6. 性能监控相关
-- 创建时间范围查询优化
CREATE INDEX IF NOT EXISTS idx_app_order_create_time ON app_order(create_time);
CREATE INDEX IF NOT EXISTS idx_app_recharge_create_time ON app_recharge(create_time);

-- 7. 用户相关优化索引
-- 用户状态查询优化
CREATE INDEX IF NOT EXISTS idx_app_user_status ON app_user(status) WHERE EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'app_user' AND column_name = 'status');

-- 用户创建时间索引
CREATE INDEX IF NOT EXISTS idx_app_user_create_time ON app_user(create_time);

-- 8. 分区表支持（如果需要）
-- 注意：分区表需要根据实际数据量和业务需求来设计
-- 示例：按月分区订单表（需要在创建表时定义）
/*
-- 如果需要分区，可以考虑以下结构：
ALTER TABLE app_order PARTITION BY RANGE (YEAR(create_time) * 100 + MONTH(create_time)) (
    PARTITION p202401 VALUES LESS THAN (202402),
    PARTITION p202402 VALUES LESS THAN (202403),
    -- 添加更多分区...
    PARTITION p_future VALUES LESS THAN MAXVALUE
);
*/

-- 9. 添加一些有用的统计视图（可选）
CREATE OR REPLACE VIEW v_order_daily_stats AS
SELECT 
    DATE(create_time) as order_date,
    status,
    COUNT(*) as order_count,
    SUM(amount) as total_amount,
    AVG(amount) as avg_amount
FROM app_order 
GROUP BY DATE(create_time), status;

CREATE OR REPLACE VIEW v_user_order_summary AS
SELECT 
    user_id,
    COUNT(*) as total_orders,
    SUM(CASE WHEN status = 'completed' THEN amount ELSE 0 END) as completed_amount,
    COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_orders,
    MAX(create_time) as last_order_time
FROM app_order 
GROUP BY user_id;

-- 10. 性能优化建议注释
/*
性能优化建议：

1. 定期维护索引：
   - 定期运行 ANALYZE TABLE 来更新表统计信息
   - 监控索引使用情况，删除未使用的索引

2. 查询优化：
   - 使用 EXPLAIN 分析查询计划
   - 避免在 WHERE 子句中使用函数
   - 合理使用 LIMIT 限制结果集大小

3. 数据归档：
   - 对于历史订单数据，考虑定期归档
   - 保持主表数据量在合理范围内

4. 缓存策略：
   - 对热点数据使用 Redis 缓存
   - 合理设置缓存过期时间

5. 监控指标：
   - 监控慢查询日志
   - 定期检查索引碎片率
   - 监控数据库连接池使用情况
*/

-- 收益统计表优化索引
-- 复合索引：租户ID + 统计日期（用于分页查询）
CREATE INDEX IF NOT EXISTS idx_revenue_tenants_date ON merchant_revenue_stats(tenants_id, stat_date DESC);

-- 复合索引：租户ID + 时间范围查询
CREATE INDEX IF NOT EXISTS idx_revenue_tenants_period ON merchant_revenue_stats(tenants_id, period_start, period_end);

-- 统计日期索引（用于搜索）
CREATE INDEX IF NOT EXISTS idx_revenue_stat_date ON merchant_revenue_stats(stat_date); 