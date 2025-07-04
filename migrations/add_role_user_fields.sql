-- 给角色表添加用户关联字段的迁移脚本
-- 支持角色管理的权限控制功能

-- 1. 添加user_id字段（创建者ID）
ALTER TABLE `role` ADD COLUMN `user_id` int(11) DEFAULT 1 COMMENT '角色创建者ID';

-- 2. 添加user_type字段（创建者类型）
ALTER TABLE `role` ADD COLUMN `user_type` int(11) DEFAULT 1 COMMENT '角色创建者类型：1=超管，2=普通管理员';

-- 3. 添加role_name字段（角色名称）
ALTER TABLE `role` ADD COLUMN `role_name` varchar(100) DEFAULT '' COMMENT '角色名称';

-- 4. 添加role_desc字段（角色描述）
ALTER TABLE `role` ADD COLUMN `role_desc` varchar(255) DEFAULT '' COMMENT '角色描述';

-- 5. 添加sort字段（排序）
ALTER TABLE `role` ADD COLUMN `sort` int(11) DEFAULT 0 COMMENT '排序';

-- 6. 添加时间字段
ALTER TABLE `role` ADD COLUMN `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间';
ALTER TABLE `role` ADD COLUMN `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间';

-- 7. 为现有角色设置默认值
UPDATE `role` SET 
    `user_id` = 1, 
    `user_type` = 1,
    `role_name` = `name`,
    `role_desc` = CONCAT('系统预设角色: ', `name`),
    `sort` = 0,
    `create_time` = NOW(),
    `update_time` = NOW()
WHERE `user_id` IS NULL OR `user_id` = 0;

-- 8. 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS `idx_role_user_id` ON `role`(`user_id`);
CREATE INDEX IF NOT EXISTS `idx_role_user_type` ON `role`(`user_type`);
CREATE INDEX IF NOT EXISTS `idx_role_name` ON `role`(`role_name`);

-- 9. 验证迁移结果
SELECT 
    id,
    code,
    name,
    role_name,
    role_desc,
    user_id,
    user_type,
    enable,
    sort,
    create_time,
    update_time,
    CASE 
        WHEN user_type = 1 THEN '超管创建'
        WHEN user_type = 2 THEN '普通管理员创建'
        ELSE '未知类型'
    END AS creator_type_desc
FROM `role` 
ORDER BY id; 