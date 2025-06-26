-- Create system_info table
CREATE TABLE IF NOT EXISTS `system_info` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `system_name` varchar(100) NOT NULL COMMENT '系统名称',
  `system_title` varchar(100) NOT NULL COMMENT '系统标题',
  `icp_number` varchar(50) DEFAULT NULL COMMENT '备案号',
  `copyright` varchar(200) DEFAULT NULL COMMENT '版权信息',
  `status` tinyint(1) DEFAULT 0 COMMENT '状态：0-禁用 1-启用',
  `tenants_id` int(11) NOT NULL COMMENT '租户ID',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  -- 复合索引：用于按租户和状态查询启用的系统信息
  KEY `idx_tenant_status` (`tenants_id`, `status`),
  -- 复合索引：用于按租户ID和创建时间排序查询
  KEY `idx_tenant_create_time` (`tenants_id`, `create_time`),
  -- 复合索引：用于按租户ID和更新时间排序查询
  KEY `idx_tenant_update_time` (`tenants_id`, `update_time`),
  -- 系统名称索引：用于模糊搜索
  KEY `idx_system_name` (`system_name`(20))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统信息配置表';

-- 添加索引使用说明
/*
索引使用说明：
1. idx_tenant_status：
   - 用于快速查询特定租户的启用状态系统信息
   - 适用于 WHERE tenants_id = ? AND status = 1 的查询

2. idx_tenant_create_time：
   - 用于按创建时间排序的租户系统信息列表查询
   - 适用于 WHERE tenants_id = ? ORDER BY create_time DESC 的查询

3. idx_tenant_update_time：
   - 用于按更新时间排序的租户系统信息列表查询
   - 适用于 WHERE tenants_id = ? ORDER BY update_time DESC 的查询

4. idx_system_name：
   - 用于系统名称的模糊搜索
   - 只索引前20个字符以减少索引大小
   - 适用于 WHERE system_name LIKE ? 的查询
*/ 