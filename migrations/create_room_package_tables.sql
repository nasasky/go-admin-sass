-- 房间套餐表
CREATE TABLE room_packages (
    id INT AUTO_INCREMENT PRIMARY KEY,
    room_id INT NOT NULL,
    package_name VARCHAR(100) NOT NULL COMMENT '套餐名称',
    description TEXT COMMENT '套餐描述',
    is_active TINYINT(1) DEFAULT 1 COMMENT '是否启用',
    priority INT DEFAULT 0 COMMENT '优先级，数字越大优先级越高',
    start_date DATE COMMENT '生效开始日期',
    end_date DATE COMMENT '生效结束日期',
    created_by INT COMMENT '创建人ID',
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_room_id (room_id),
    INDEX idx_is_active (is_active),
    INDEX idx_priority (priority),
    INDEX idx_date_range (start_date, end_date),
    FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='房间套餐规则';

-- 套餐定价规则表
CREATE TABLE room_package_rules (
    id INT AUTO_INCREMENT PRIMARY KEY,
    package_id INT NOT NULL,
    rule_name VARCHAR(100) NOT NULL COMMENT '规则名称',
    day_type ENUM('weekday', 'weekend', 'holiday', 'special', 'all') DEFAULT 'all' COMMENT '日期类型',
    time_start TIME COMMENT '时间段开始(HH:mm)',
    time_end TIME COMMENT '时间段结束(HH:mm)',
    price_type ENUM('fixed', 'multiply', 'add') NOT NULL COMMENT '价格类型',
    price_value DECIMAL(10,2) NOT NULL COMMENT '价格值',
    min_hours INT DEFAULT 1 COMMENT '最少预订小时数',
    max_hours INT DEFAULT 24 COMMENT '最多预订小时数',
    priority INT DEFAULT 0 COMMENT '规则优先级',
    is_active TINYINT(1) DEFAULT 1 COMMENT '是否启用',
    description TEXT COMMENT '规则描述',
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_package_id (package_id),
    INDEX idx_day_type (day_type),
    INDEX idx_time_range (time_start, time_end),
    INDEX idx_is_active (is_active),
    INDEX idx_priority (priority),
    FOREIGN KEY (package_id) REFERENCES room_packages(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='套餐定价规则';

-- 特殊日期配置表
CREATE TABLE room_special_dates (
    id INT AUTO_INCREMENT PRIMARY KEY,
    date DATE NOT NULL COMMENT '特殊日期',
    date_type ENUM('holiday', 'festival', 'special') NOT NULL COMMENT '日期类型',
    name VARCHAR(100) NOT NULL COMMENT '日期名称',
    description TEXT COMMENT '描述',
    is_active TINYINT(1) DEFAULT 1 COMMENT '是否启用',
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_date (date),
    INDEX idx_date_type (date_type),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='特殊日期配置';

-- 插入一些示例数据
INSERT INTO room_special_dates (date, date_type, name, description) VALUES
('2024-01-01', 'holiday', '元旦', '元旦节'),
('2024-02-14', 'festival', '情人节', '情人节特殊定价'),
('2024-05-01', 'holiday', '劳动节', '劳动节'),
('2024-06-01', 'festival', '儿童节', '儿童节'),
('2024-10-01', 'holiday', '国庆节', '国庆节'),
('2024-12-25', 'festival', '圣诞节', '圣诞节特殊定价');

-- 示例套餐数据 (假设房间ID为1)
INSERT INTO room_packages (room_id, package_name, description, priority, start_date, end_date, created_by) VALUES
(1, '工作日优惠套餐', '工作日时段优惠价格', 10, '2024-01-01', '2024-12-31', 1),
(1, '周末加价套餐', '周末时段加价', 20, '2024-01-01', '2024-12-31', 1),
(1, '节假日特价套餐', '节假日特殊定价', 30, '2024-01-01', '2024-12-31', 1);

-- 示例规则数据
INSERT INTO room_package_rules (package_id, rule_name, day_type, time_start, time_end, price_type, price_value, min_hours, max_hours, priority, description) VALUES
-- 工作日套餐规则
(1, '工作日白天优惠', 'weekday', '09:00', '18:00', 'multiply', 0.8, 2, 8, 10, '工作日白天时段8折优惠'),
(1, '工作日晚上标准', 'weekday', '18:00', '23:00', 'multiply', 1.0, 1, 6, 5, '工作日晚上标准价格'),

-- 周末套餐规则  
(2, '周末白天加价', 'weekend', '09:00', '18:00', 'multiply', 1.2, 2, 8, 15, '周末白天时段1.2倍价格'),
(2, '周末晚上加价', 'weekend', '18:00', '23:00', 'multiply', 1.5, 1, 6, 20, '周末晚上时段1.5倍价格'),

-- 节假日套餐规则
(3, '节假日全天加价', 'holiday', '00:00', '23:59', 'multiply', 2.0, 1, 12, 25, '节假日全天2倍价格'),
(3, '特殊节日加价', 'special', '00:00', '23:59', 'add', 100.0, 1, 12, 30, '特殊节日每小时加价100元'); 