-- 为预订表添加套餐关联字段
-- 执行时间: 2024-06-08

-- 为 room_bookings 表添加套餐相关字段
ALTER TABLE room_bookings 
ADD COLUMN package_id INT DEFAULT NULL COMMENT '使用的套餐ID';

ALTER TABLE room_bookings 
ADD COLUMN package_name VARCHAR(100) DEFAULT NULL COMMENT '套餐名称（冗余字段，便于查询）';

ALTER TABLE room_bookings 
ADD COLUMN original_price DECIMAL(10,2) DEFAULT 0.00 COMMENT '原始价格（按基础价格计算）';

ALTER TABLE room_bookings 
ADD COLUMN package_price DECIMAL(10,2) DEFAULT 0.00 COMMENT '套餐价格';

ALTER TABLE room_bookings 
ADD COLUMN discount_amount DECIMAL(10,2) DEFAULT 0.00 COMMENT '优惠金额';

ALTER TABLE room_bookings 
ADD COLUMN price_breakdown TEXT DEFAULT NULL COMMENT '价格明细JSON';

-- 添加外键约束
ALTER TABLE room_bookings 
ADD CONSTRAINT fk_room_bookings_package 
FOREIGN KEY (package_id) REFERENCES room_packages(id) ON DELETE SET NULL;

-- 添加索引
CREATE INDEX idx_room_bookings_package ON room_bookings(package_id);
CREATE INDEX idx_room_bookings_package_name ON room_bookings(package_name);

-- 示例数据：更新一些预订记录关联套餐
-- UPDATE room_bookings SET package_id = 1, package_name = '3小时工作套餐', package_price = 180.00 WHERE id = 1; 