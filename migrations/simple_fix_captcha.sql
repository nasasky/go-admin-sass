-- 简单的验证码配置修复脚本
-- 手动添加 value 字段到 config_setting 表

-- 添加 value 字段（如果不存在会报错，可以忽略）
ALTER TABLE `config_setting` ADD COLUMN `value` text COMMENT '配置值' AFTER `type`;

-- 插入验证码开关配置
INSERT INTO `config_setting` (`type`, `name`, `value`, `create_time`, `update_time`) 
VALUES ('system', 'captcha_enabled', '1', NOW(), NOW())
ON DUPLICATE KEY UPDATE `value` = VALUES(`value`), `update_time` = NOW();

-- 查看结果
SELECT * FROM `config_setting` WHERE `type` = 'system' AND `name` = 'captcha_enabled'; 