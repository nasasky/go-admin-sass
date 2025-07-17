-- 修复验证码配置脚本
-- 为 config_setting 表添加 value 字段

-- 检查 value 字段是否存在，如果不存在则添加
SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE TABLE_SCHEMA = DATABASE() 
     AND TABLE_NAME = 'config_setting' 
     AND COLUMN_NAME = 'value') = 0,
    'ALTER TABLE `config_setting` ADD COLUMN `value` text COMMENT "配置值" AFTER `type`',
    'SELECT "value column already exists" as message'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 插入验证码开关默认配置（默认启用）
INSERT INTO `config_setting` (`type`, `name`, `value`, `create_time`, `update_time`) 
VALUES ('system', 'captcha_enabled', '1', NOW(), NOW())
ON DUPLICATE KEY UPDATE `value` = VALUES(`value`), `update_time` = NOW();

-- 验证插入结果
SELECT * FROM `config_setting` WHERE `type` = 'system' AND `name` = 'captcha_enabled'; 