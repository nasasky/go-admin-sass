-- 为用户表添加组合索引
CREATE INDEX IF NOT EXISTS idx_username_phone ON admin_user(username, phone);

-- 为权限相关表添加索引
CREATE INDEX IF NOT EXISTS idx_role_permissions ON role_permissions_permission(roleId, permissionId);
CREATE INDEX IF NOT EXISTS idx_permission_rules ON permission_user(rule, permiss_rule);

-- 为密码字段添加索引（如果经常使用旧的MD5密码验证）
CREATE INDEX IF NOT EXISTS idx_password ON admin_user(password);

-- 为bcrypt密码字段添加索引
CREATE INDEX IF NOT EXISTS idx_password_bcrypt ON admin_user(password_bcrypt);

-- 为用户状态相关字段添加索引（如果有的话）
CREATE INDEX IF NOT EXISTS idx_user_status ON admin_user(user_type, role_id); 