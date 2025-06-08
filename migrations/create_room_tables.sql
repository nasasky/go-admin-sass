-- 房间包厢相关数据表创建脚本
-- 创建时间: 2024-01-01

-- 创建房间表
CREATE TABLE IF NOT EXISTS `rooms` (
    `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '房间ID',
    `room_number` varchar(20) NOT NULL COMMENT '房间号码',
    `room_name` varchar(100) NOT NULL COMMENT '房间名称',
    `room_type` varchar(20) NOT NULL COMMENT '房间类型(small/medium/large/luxury)',
    `capacity` int(11) NOT NULL COMMENT '容纳人数',
    `hourly_rate` decimal(10,2) NOT NULL COMMENT '每小时价格',
    `features` text COMMENT '房间特色设施(JSON格式)',
    `images` text COMMENT '房间图片URLs(JSON格式)',
    `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '房间状态(1:可用,2:使用中,3:维护中,4:停用)',
    `floor` int(11) DEFAULT NULL COMMENT '楼层',
    `area` decimal(8,2) DEFAULT NULL COMMENT '房间面积(平方米)',
    `description` text COMMENT '房间描述',
    `created_by` int(11) DEFAULT NULL COMMENT '创建人ID',
    `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_room_number` (`room_number`),
    KEY `idx_room_type` (`room_type`),
    KEY `idx_status` (`status`),
    KEY `idx_floor` (`floor`),
    KEY `idx_capacity` (`capacity`),
    KEY `idx_hourly_rate` (`hourly_rate`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='房间信息表';

-- 创建房间预订表
CREATE TABLE IF NOT EXISTS `room_bookings` (
    `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '预订ID',
    `room_id` int(11) NOT NULL COMMENT '房间ID',
    `user_id` int(11) NOT NULL COMMENT '用户ID',
    `booking_no` varchar(32) NOT NULL COMMENT '预订单号',
    `start_time` datetime NOT NULL COMMENT '开始时间',
    `end_time` datetime NOT NULL COMMENT '结束时间',
    `hours` int(11) NOT NULL COMMENT '预订小时数',
    `total_amount` decimal(10,2) NOT NULL COMMENT '总金额',
    `paid_amount` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '已支付金额',
    `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '预订状态(1:待支付,2:已支付,3:使用中,4:已完成,5:已取消,6:已退款)',
    `payment_id` int(11) DEFAULT NULL COMMENT '支付记录ID',
    `contact_name` varchar(50) DEFAULT NULL COMMENT '联系人姓名',
    `contact_phone` varchar(20) DEFAULT NULL COMMENT '联系人电话',
    `remarks` text COMMENT '备注信息',
    `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_booking_no` (`booking_no`),
    KEY `idx_room_id` (`room_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_status` (`status`),
    KEY `idx_start_time` (`start_time`),
    KEY `idx_end_time` (`end_time`),
    KEY `idx_create_time` (`create_time`),
    CONSTRAINT `fk_room_bookings_room_id` FOREIGN KEY (`room_id`) REFERENCES `rooms` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_room_bookings_user_id` FOREIGN KEY (`user_id`) REFERENCES `app_user` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='房间预订表';

-- 创建房间使用记录表
CREATE TABLE IF NOT EXISTS `room_usage_logs` (
    `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '记录ID',
    `room_id` int(11) NOT NULL COMMENT '房间ID',
    `booking_id` int(11) NOT NULL COMMENT '预订ID',
    `user_id` int(11) NOT NULL COMMENT '用户ID',
    `check_in_at` datetime NOT NULL COMMENT '入住时间',
    `check_out_at` datetime DEFAULT NULL COMMENT '离开时间',
    `actual_hours` decimal(8,2) DEFAULT NULL COMMENT '实际使用小时数',
    `extra_fee` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '额外费用',
    `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_room_id` (`room_id`),
    KEY `idx_booking_id` (`booking_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_check_in_at` (`check_in_at`),
    KEY `idx_check_out_at` (`check_out_at`),
    CONSTRAINT `fk_room_usage_logs_room_id` FOREIGN KEY (`room_id`) REFERENCES `rooms` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_room_usage_logs_booking_id` FOREIGN KEY (`booking_id`) REFERENCES `room_bookings` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_room_usage_logs_user_id` FOREIGN KEY (`user_id`) REFERENCES `app_user` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='房间使用记录表';

-- 插入示例数据
INSERT INTO `rooms` (`room_number`, `room_name`, `room_type`, `capacity`, `hourly_rate`, `features`, `images`, `status`, `floor`, `area`, `description`, `created_by`) VALUES
('A001', '雅致小包厢', 'small', 4, 88.00, '["KTV设备", "茶水服务", "WIFI"]', '["https://example.com/room1.jpg"]', 1, 1, 20.50, '温馨舒适的小包厢，适合小型聚会', 1),
('A002', '豪华中包厢', 'medium', 8, 168.00, '["KTV设备", "茶水服务", "WIFI", "投影设备"]', '["https://example.com/room2.jpg"]', 1, 1, 35.80, '设施齐全的中型包厢，适合朋友聚会', 1),
('B001', '商务大包厢', 'large', 12, 288.00, '["KTV设备", "茶水服务", "WIFI", "投影设备", "商务桌椅"]', '["https://example.com/room3.jpg"]', 1, 2, 50.20, '商务风格大包厢，适合商务聚会', 1),
('B002', '至尊豪华包厢', 'luxury', 20, 588.00, '["KTV设备", "茶水服务", "WIFI", "投影设备", "商务桌椅", "独立卫生间", "小厨房"]', '["https://example.com/room4.jpg"]', 1, 2, 80.00, '顶级豪华包厢，配备齐全设施', 1);

-- 创建索引优化查询性能
CREATE INDEX idx_rooms_composite ON rooms (room_type, status, capacity);
CREATE INDEX idx_bookings_time_range ON room_bookings (room_id, start_time, end_time);
CREATE INDEX idx_bookings_user_status ON room_bookings (user_id, status, create_time);
CREATE INDEX idx_usage_logs_composite ON room_usage_logs (room_id, check_in_at, check_out_at);

-- 添加数据完整性约束
-- 确保房间类型只能是指定值
ALTER TABLE rooms ADD CONSTRAINT chk_room_type CHECK (room_type IN ('small', 'medium', 'large', 'luxury'));

-- 确保房间状态只能是指定值
ALTER TABLE rooms ADD CONSTRAINT chk_room_status CHECK (status IN (1, 2, 3, 4));

-- 确保预订状态只能是指定值
ALTER TABLE room_bookings ADD CONSTRAINT chk_booking_status CHECK (status IN (1, 2, 3, 4, 5, 6));

-- 确保时间逻辑正确
ALTER TABLE room_bookings ADD CONSTRAINT chk_booking_time CHECK (end_time > start_time);

-- 确保价格为正数
ALTER TABLE rooms ADD CONSTRAINT chk_hourly_rate CHECK (hourly_rate > 0);
ALTER TABLE room_bookings ADD CONSTRAINT chk_total_amount CHECK (total_amount >= 0);
ALTER TABLE room_bookings ADD CONSTRAINT chk_paid_amount CHECK (paid_amount >= 0);

-- 创建视图用于查询统计信息
CREATE VIEW v_room_statistics AS
SELECT 
    r.room_type,
    COUNT(*) as total_rooms,
    SUM(CASE WHEN r.status = 1 THEN 1 ELSE 0 END) as available_rooms,
    SUM(CASE WHEN r.status = 2 THEN 1 ELSE 0 END) as occupied_rooms,
    SUM(CASE WHEN r.status = 3 THEN 1 ELSE 0 END) as maintenance_rooms,
    SUM(CASE WHEN r.status = 4 THEN 1 ELSE 0 END) as disabled_rooms,
    AVG(r.hourly_rate) as avg_hourly_rate,
    MIN(r.hourly_rate) as min_hourly_rate,
    MAX(r.hourly_rate) as max_hourly_rate
FROM rooms r
GROUP BY r.room_type;

-- 创建预订统计视图
CREATE VIEW v_booking_statistics AS
SELECT 
    DATE(rb.create_time) as booking_date,
    COUNT(*) as total_bookings,
    SUM(CASE WHEN rb.status = 2 THEN 1 ELSE 0 END) as paid_bookings,
    SUM(CASE WHEN rb.status = 4 THEN 1 ELSE 0 END) as completed_bookings,
    SUM(CASE WHEN rb.status = 5 THEN 1 ELSE 0 END) as cancelled_bookings,
    SUM(CASE WHEN rb.status IN (2, 4) THEN rb.total_amount ELSE 0 END) as total_revenue
FROM room_bookings rb
GROUP BY DATE(rb.create_time); 