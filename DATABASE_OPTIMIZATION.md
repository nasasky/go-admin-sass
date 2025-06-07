# 数据库优化建议

## 索引优化建议

### 1. 用户表 (app_user) 索引

```sql
-- 手机号索引（登录查询）
CREATE INDEX idx_app_user_phone ON app_user(phone);

-- OpenID索引（微信登录）
CREATE INDEX idx_app_user_openid ON app_user(openid);

-- 创建时间索引（按时间排序）
CREATE INDEX idx_app_user_create_time ON app_user(create_time);

-- 复合索引：状态+创建时间
CREATE INDEX idx_app_user_status_create_time ON app_user(enable, create_time);
```

### 2. 商品表 (goods_list) 索引

```sql
-- 商品名称索引（搜索功能）
CREATE INDEX idx_goods_name ON goods_list(goods_name);

-- 状态索引
CREATE INDEX idx_goods_status ON goods_list(status);

-- 分类索引
CREATE INDEX idx_goods_category_id ON goods_list(category_id);

-- 租户ID索引
CREATE INDEX idx_goods_tenants_id ON goods_list(tenants_id);

-- 软删除索引
CREATE INDEX idx_goods_isdelete ON goods_list(isdelete);

-- 复合索引：租户+状态+删除标记
CREATE INDEX idx_goods_tenants_status_delete ON goods_list(tenants_id, status, isdelete);

-- 复合索引：分类+状态+删除标记+创建时间
CREATE INDEX idx_goods_category_status_delete_time ON goods_list(category_id, status, isdelete, create_time DESC);

-- 库存索引（库存查询）
CREATE INDEX idx_goods_stock ON goods_list(stock);

-- 价格索引（价格排序）
CREATE INDEX idx_goods_price ON goods_list(price);
```

### 3. 订单表 (order) 索引

```sql
-- 用户ID索引（查询用户订单）
CREATE INDEX idx_order_user_id ON `order`(user_id);

-- 商品ID索引（查询商品相关订单）
CREATE INDEX idx_order_goods_id ON `order`(goods_id);

-- 订单号索引（订单查询）
CREATE UNIQUE INDEX idx_order_no ON `order`(no);

-- 状态索引
CREATE INDEX idx_order_status ON `order`(status);

-- 租户ID索引
CREATE INDEX idx_order_tenants_id ON `order`(tenants_id);

-- 复合索引：用户+状态+创建时间
CREATE INDEX idx_order_user_status_time ON `order`(user_id, status, create_time DESC);

-- 复合索引：租户+状态+创建时间
CREATE INDEX idx_order_tenants_status_time ON `order`(tenants_id, status, create_time DESC);

-- 复合索引：商品+状态
CREATE INDEX idx_order_goods_status ON `order`(goods_id, status);

-- 创建时间索引
CREATE INDEX idx_order_create_time ON `order`(create_time DESC);
```

### 4. 权限表 (permission_user) 索引

```sql
-- 父级ID索引（权限树查询）
CREATE INDEX idx_permission_parent_id ON permission_user(parent_id);

-- 排序索引
CREATE INDEX idx_permission_sort ON permission_user(sort DESC);

-- 复合索引：父级+排序
CREATE INDEX idx_permission_parent_sort ON permission_user(parent_id, sort DESC);

-- 权限规则索引
CREATE INDEX idx_permission_rule ON permission_user(rule);

-- 权限类型索引
CREATE INDEX idx_permission_type ON permission_user(type);
```

### 5. 角色权限关联表 (role_permissions_permission) 索引

```sql
-- 角色ID索引
CREATE INDEX idx_role_perm_role_id ON role_permissions_permission(roleId);

-- 权限ID索引
CREATE INDEX idx_role_perm_permission_id ON role_permissions_permission(permissionId);

-- 复合唯一索引
CREATE UNIQUE INDEX idx_role_perm_unique ON role_permissions_permission(roleId, permissionId);
```

### 6. 管理员用户表 (admin_user) 索引

```sql
-- 用户名索引（登录查询）
CREATE INDEX idx_admin_user_username ON admin_user(username);

-- 手机号索引
CREATE INDEX idx_admin_user_phone ON admin_user(phone);

-- 角色ID索引
CREATE INDEX idx_admin_user_role_id ON admin_user(role_id);

-- 用户类型索引
CREATE INDEX idx_admin_user_type ON admin_user(user_type);

-- 父级ID索引
CREATE INDEX idx_admin_user_parent_id ON admin_user(parent_id);

-- 复合索引：用户名+密码（登录查询）
CREATE INDEX idx_admin_user_login ON admin_user(username, password);

-- 复合索引：手机号+密码
CREATE INDEX idx_admin_user_phone_login ON admin_user(phone, password);
```

### 7. 钱包表 (app_wallet) 索引

```sql
-- 用户ID索引
CREATE UNIQUE INDEX idx_wallet_user_id ON app_wallet(user_id);

-- 余额索引（余额查询）
CREATE INDEX idx_wallet_money ON app_wallet(money);
```

## 查询优化建议

### 1. 分页查询优化

```sql
-- 使用覆盖索引进行分页
-- 不推荐
SELECT * FROM goods_list WHERE tenants_id = 1 ORDER BY create_time DESC LIMIT 10 OFFSET 100;

-- 推荐：使用子查询先获取ID
SELECT g.* FROM goods_list g 
INNER JOIN (
    SELECT id FROM goods_list 
    WHERE tenants_id = 1 AND isdelete != 1 
    ORDER BY create_time DESC 
    LIMIT 10 OFFSET 100
) t ON g.id = t.id;
```

### 2. 避免N+1查询

```sql
-- 不推荐：循环查询
-- 在代码中循环执行
SELECT * FROM goods_list WHERE id = ?;

-- 推荐：批量查询
SELECT * FROM goods_list WHERE id IN (1,2,3,4,5);
```

### 3. 只查询需要的字段

```sql
-- 不推荐
SELECT * FROM app_user WHERE id = 1;

-- 推荐
SELECT id, username, phone, avatar FROM app_user WHERE id = 1;
```

### 4. 使用EXPLAIN分析查询

```sql
-- 分析查询执行计划
EXPLAIN SELECT * FROM `order` 
WHERE user_id = 1 AND status = 'paid' 
ORDER BY create_time DESC 
LIMIT 10;
```

## 表结构优化建议

### 1. 添加字段长度限制

```sql
-- 优化字符串字段长度
ALTER TABLE goods_list MODIFY COLUMN goods_name VARCHAR(255) NOT NULL;
ALTER TABLE goods_list MODIFY COLUMN status VARCHAR(20) NOT NULL DEFAULT 'active';
```

### 2. 添加默认值

```sql
-- 添加默认值减少NULL值
ALTER TABLE goods_list MODIFY COLUMN stock INT NOT NULL DEFAULT 0;
ALTER TABLE goods_list MODIFY COLUMN isdelete TINYINT NOT NULL DEFAULT 0;
```

### 3. 字段类型优化

```sql
-- 使用更合适的数据类型
ALTER TABLE goods_list MODIFY COLUMN price DECIMAL(10,2) NOT NULL DEFAULT 0.00;
ALTER TABLE goods_list MODIFY COLUMN isdelete TINYINT(1) NOT NULL DEFAULT 0;
```

## 连接池优化

### 1. MySQL配置优化

```ini
# my.cnf
[mysqld]
# 连接数限制
max_connections = 1000
max_user_connections = 800

# 缓冲区设置
innodb_buffer_pool_size = 2G
query_cache_size = 256M
query_cache_type = 1

# 超时设置
wait_timeout = 300
interactive_timeout = 300

# 慢查询日志
slow_query_log = 1
slow_query_log_file = /var/log/mysql/slow.log
long_query_time = 1
```

### 2. 应用层连接池配置

```go
// 已在 db/db.go 中配置
func InitDB() {
    // 设置连接池参数
    sqlDB.SetMaxOpenConns(100)    // 最大连接数
    sqlDB.SetMaxIdleConns(20)     // 最大空闲连接数
    sqlDB.SetConnMaxLifetime(time.Hour)      // 连接最大生存时间
    sqlDB.SetConnMaxIdleTime(30 * time.Minute) // 连接最大空闲时间
}
```

## 监控和分析

### 1. 慢查询监控

```sql
-- 查看慢查询
SHOW VARIABLES LIKE 'slow_query_log%';
SHOW VARIABLES LIKE 'long_query_time';

-- 分析慢查询日志
mysqldumpslow -s c -t 10 /var/log/mysql/slow.log
```

### 2. 索引使用情况

```sql
-- 查看索引使用情况
SELECT 
    OBJECT_SCHEMA,
    OBJECT_NAME,
    INDEX_NAME,
    COUNT_FETCH,
    COUNT_INSERT,
    COUNT_UPDATE,
    COUNT_DELETE
FROM performance_schema.table_io_waits_summary_by_index_usage
WHERE OBJECT_SCHEMA = 'your_database_name'
ORDER BY COUNT_FETCH DESC;
```

### 3. 表锁监控

```sql
-- 查看表锁情况
SHOW ENGINE INNODB STATUS\G
```

## 缓存策略

### 1. 查询结果缓存

- 商品列表：5分钟
- 商品详情：10分钟
- 用户信息：30分钟
- 权限信息：15分钟

### 2. 缓存失效策略

- 数据更新时主动失效相关缓存
- 定期清理过期缓存
- 使用版本号或时间戳避免缓存穿透

## 实施优先级

### 高优先级
1. 订单表索引优化
2. 商品表索引优化
3. 用户表索引优化
4. 权限查询优化

### 中优先级
1. 分页查询优化
2. 连接池配置优化
3. 慢查询监控

### 低优先级
1. 表结构细节优化
2. 高级缓存策略
3. 读写分离（如果需要）

通过以上优化，预计可以实现：
- 查询响应时间降低 60-80%
- 数据库并发处理能力提升 5-10倍
- 缓存命中率达到 85%以上
- 系统整体性能提升 3-5倍 