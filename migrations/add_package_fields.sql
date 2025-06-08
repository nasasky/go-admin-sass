-- 添加套餐新字段的迁移脚本
-- 执行时间: 2024-06-08

-- 为 room_packages 表添加新字段
ALTER TABLE room_packages 
ADD COLUMN package_type VARCHAR(20) DEFAULT 'flexible' COMMENT '套餐类型(flexible/fixed_hours/daily/weekly)';

ALTER TABLE room_packages 
ADD COLUMN fixed_hours INT DEFAULT 0 COMMENT '固定时长(小时)，0表示灵活时长';

ALTER TABLE room_packages 
ADD COLUMN min_hours INT DEFAULT 1 COMMENT '最少预订小时数';

ALTER TABLE room_packages 
ADD COLUMN max_hours INT DEFAULT 24 COMMENT '最多预订小时数';

ALTER TABLE room_packages 
ADD COLUMN base_price DECIMAL(10,2) DEFAULT 0.00 COMMENT '套餐基础价格';

-- 为新字段添加索引
CREATE INDEX idx_room_packages_type ON room_packages(package_type);
CREATE INDEX idx_room_packages_active_type ON room_packages(is_active, package_type);

-- 插入一些示例套餐数据
INSERT INTO room_packages (room_id, package_name, description, package_type, fixed_hours, base_price, priority, is_active) VALUES
(1, '3小时工作套餐', '适合短期会议和小组讨论，固定3小时', 'fixed_hours', 3, 180.00, 10, 1),
(1, '周末6小时套餐', '周末特惠，固定6小时套餐', 'fixed_hours', 6, 320.00, 20, 1),
(1, '全天会议套餐', '24小时全天候会议室使用', 'daily', 24, 800.00, 30, 1),
(2, '灵活时长套餐', '按需预订，1-12小时任选', 'flexible', 0, 0.00, 5, 1);

-- 为每个套餐添加对应的规则
INSERT INTO room_package_rules (package_id, rule_name, day_type, price_type, price_value, min_hours, max_hours, is_active) VALUES
-- 3小时工作套餐规则
(1, '工作日标准价', 'weekday', 'fixed', 180.00, 3, 3, 1),
(1, '周末加价20%', 'weekend', 'multiply', 1.2, 3, 3, 1),

-- 周末6小时套餐规则
(2, '周末专享价', 'weekend', 'fixed', 320.00, 6, 6, 1),
(2, '节假日特价', 'holiday', 'multiply', 1.5, 6, 6, 1),

-- 全天会议套餐规则
(3, '工作日全天价', 'weekday', 'fixed', 800.00, 24, 24, 1),
(3, '周末全天价', 'weekend', 'fixed', 960.00, 24, 24, 1),
(3, '节假日全天价', 'holiday', 'fixed', 1200.00, 24, 24, 1),

-- 灵活时长套餐规则
(4, '工作日时段9-18点8折', 'weekday', 'multiply', 0.8, 1, 12, 1),
(4, '工作日其他时段标准价', 'weekday', 'multiply', 1.0, 1, 12, 1),
(4, '周末标准价', 'weekend', 'multiply', 1.2, 1, 12, 1);

-- 更新灵活时长套餐规则的时间段
UPDATE room_package_rules SET time_start = '09:00', time_end = '18:00' WHERE id = 4; 