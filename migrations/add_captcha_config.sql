-- 添加验证码开关配置
-- 如果config_setting表不存在，先创建表
CREATE TABLE IF NOT EXISTS `config_setting` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) DEFAULT NULL COMMENT '用户ID',
  `name` varchar(255) NOT NULL COMMENT '配置名称',
  `appid` varchar(255) DEFAULT NULL COMMENT '应用ID',
  `secret` varchar(255) DEFAULT NULL COMMENT '密钥',
  `tips` text COMMENT '提示信息',
  `endpoint` varchar(255) DEFAULT NULL COMMENT '端点',
  `bucket_name` varchar(255) DEFAULT NULL COMMENT '存储桶名称',
  `base_url` varchar(255) DEFAULT NULL COMMENT '基础URL',
  `type` varchar(50) NOT NULL COMMENT '配置类型',
  `value` text COMMENT '配置值',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_type_name` (`type`, `name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统配置表';

-- 如果表已存在但缺少value字段，添加该字段
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