-- 添加用户表缺失字段的迁移脚本
-- 解决推送系统管理员用户筛选问题

-- 1. 添加user_type字段（用户类型：1=管理员，2=普通用户）
ALTER TABLE `user` ADD COLUMN `user_type` int(11) NOT NULL DEFAULT 2 COMMENT '用户类型：1=管理员，2=普通用户';

-- 2. 添加role_id字段（角色ID）
ALTER TABLE `user` ADD COLUMN `role_id` int(11) NOT NULL DEFAULT 2 COMMENT '角色ID';

-- 3. 添加notice字段（是否接收推送：1=是，0=否）
ALTER TABLE `user` ADD COLUMN `notice` tinyint(4) NOT NULL DEFAULT 0 COMMENT '是否接收推送：1=是，0=否';

-- 4. 为现有用户设置默认值
-- 将admin用户设置为管理员
UPDATE `user` SET `user_type` = 1, `role_id` = 1, `notice` = 1 WHERE `username` = 'admin';

-- 5. 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS `idx_user_type` ON `user`(`user_type`);
CREATE INDEX IF NOT EXISTS `idx_user_notice` ON `user`(`notice`);
CREATE INDEX IF NOT EXISTS `idx_user_role` ON `user`(`role_id`);

-- 6. 验证迁移结果
SELECT 
    id, 
    username, 
    user_type, 
    role_id, 
    notice,
    CASE 
        WHEN user_type = 1 THEN '管理员'
        ELSE '普通用户'
    END AS user_type_desc,
    CASE 
        WHEN notice = 1 THEN '可接收推送'
        ELSE '不接收推送'
    END AS notice_status
FROM `user` 
ORDER BY id; 