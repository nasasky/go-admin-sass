-- 修复订单表结构问题
-- 1. 检查并添加 update_time 字段
SET @update_time_exists = (
    SELECT COUNT(*) 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'app_order' 
    AND COLUMN_NAME = 'update_time'
);

SET @sql = IF(@update_time_exists = 0, 
    'ALTER TABLE app_order ADD COLUMN update_time DATETIME NULL DEFAULT NULL COMMENT "更新时间"',
    'SELECT "update_time字段已存在" AS message'
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 2. 检查 status 字段类型，确保它能够存储所需的状态值
SET @status_type = (
    SELECT DATA_TYPE 
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() 
    AND TABLE_NAME = 'app_order' 
    AND COLUMN_NAME = 'status'
);

-- 如果 status 不是 VARCHAR 类型，修改它
SET @sql2 = IF(@status_type != 'varchar', 
    'ALTER TABLE app_order MODIFY COLUMN status VARCHAR(20) NOT NULL DEFAULT "pending" COMMENT "订单状态: pending=待支付, paid=已支付, cancelled=已取消, refunded=已退款"',
    'SELECT "status字段类型正确" AS message'
);

PREPARE stmt2 FROM @sql2;
EXECUTE stmt2;
DEALLOCATE PREPARE stmt2;

-- 3. 为现有记录设置 update_time
UPDATE app_order 
SET update_time = create_time 
WHERE update_time IS NULL;

-- 4. 添加索引以提升查询性能
CREATE INDEX IF NOT EXISTS idx_order_status_createtime ON app_order (status, create_time);
CREATE INDEX IF NOT EXISTS idx_order_no ON app_order (no);
CREATE INDEX IF NOT EXISTS idx_order_userid_status ON app_order (user_id, status);

-- 5. 添加触发器自动更新 update_time
DELIMITER $$

DROP TRIGGER IF EXISTS order_update_trigger$$

CREATE TRIGGER order_update_trigger 
BEFORE UPDATE ON app_order
FOR EACH ROW
BEGIN
    SET NEW.update_time = NOW();
END$$

DELIMITER ;

SELECT "订单表结构修复完成" AS result; 