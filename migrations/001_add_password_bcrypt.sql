-- 添加 bcrypt 密码字段到用户表
-- 执行前请先备份数据库

-- 管理员用户表添加新密码字段
ALTER TABLE `user` ADD COLUMN `password_bcrypt` VARCHAR(255) NULL COMMENT 'bcrypt加密的密码' AFTER `password`;

-- 应用用户表添加新密码字段 (如果存在单独的app用户表)
-- ALTER TABLE `app_user` ADD COLUMN `password_bcrypt` VARCHAR(255) NULL COMMENT 'bcrypt加密的密码' AFTER `password`;

-- 添加索引以提高查询性能
CREATE INDEX `idx_user_username` ON `user`(`username`);
CREATE INDEX `idx_user_password_bcrypt` ON `user`(`password_bcrypt`);

-- 注意：
-- 1. 执行此迁移后，需要运行密码迁移程序
-- 2. 在所有用户密码迁移完成后，可以删除旧的password字段
-- 3. 建议分步骤执行：先添加字段，迁移数据，再删除旧字段

-- 迁移完成后可执行的清理语句（谨慎使用）：
-- ALTER TABLE `user` DROP COLUMN `password`;
-- ALTER TABLE `user` CHANGE `password_bcrypt` `password` VARCHAR(255) NOT NULL; 