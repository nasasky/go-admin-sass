-- 角色表性能优化索引
-- 为角色管理系统创建性能优化索引

-- 1. 复合索引：user_id + user_type（权限过滤优化）
CREATE INDEX IF NOT EXISTS idx_role_user_permission ON role (user_id, user_type);

-- 2. 角色名称搜索索引
CREATE INDEX IF NOT EXISTS idx_role_name ON role (role_name);

-- 3. 启用状态索引
CREATE INDEX IF NOT EXISTS idx_role_enable ON role (enable);

-- 4. 排序字段索引
CREATE INDEX IF NOT EXISTS idx_role_sort ON role (sort);

-- 5. 创建时间索引（用于排序）
CREATE INDEX IF NOT EXISTS idx_role_create_time ON role (create_time);

-- 6. 复合索引：enable + sort（启用状态的角色按排序字段查询）
CREATE INDEX IF NOT EXISTS idx_role_enable_sort ON role (enable, sort);

-- 7. 复合索引：user_type + role_name（按类型搜索角色名）
CREATE INDEX IF NOT EXISTS idx_role_type_name ON role (user_type, role_name);

-- 8. ID主键已存在，无需额外创建

-- 用户表相关索引（如果不存在）
-- 9. 用户ID索引（批量查询用户信息优化）
CREATE INDEX IF NOT EXISTS idx_user_id ON user (id);

-- 10. 用户名索引（用户信息查询优化）
CREATE INDEX IF NOT EXISTS idx_user_username ON user (username);

-- 角色权限关联表索引
-- 11. 角色ID索引
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions_permission (roleId);

-- 12. 权限ID索引
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions_permission (permissionId);

-- 13. 复合索引：roleId + permissionId（权限查询优化）
CREATE INDEX IF NOT EXISTS idx_role_permissions_composite ON role_permissions_permission (roleId, permissionId); 