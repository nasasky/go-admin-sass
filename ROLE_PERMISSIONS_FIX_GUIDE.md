# 角色权限控制修复指南

## 问题描述

原有的获取角色列表接口存在权限控制缺失的问题：
- 所有用户都能查看所有角色，不论其用户类型
- 非超管用户应该只能查看和管理自己创建的角色
- 超管（user_type=1）应该能查看和管理所有角色

## 修复内容

### 1. 数据库结构修改

**修改文件**: `migrations/add_role_user_fields.sql`

给 `role` 表添加了以下字段：
- `user_id`: 角色创建者ID
- `user_type`: 角色创建者类型（1=超管，2=普通管理员）
- `role_name`: 角色名称
- `role_desc`: 角色描述
- `sort`: 排序字段
- `create_time`: 创建时间
- `update_time`: 更新时间

### 2. 数据模型修改

**修改文件**: `model/admin_model/role.go`

在 `Role` 结构体中添加了权限控制相关字段：
```go
type Role struct {
    Id         int    `json:"id"`
    RoleName   string `json:"role_name" gorm:"column:role_name"`
    RoleDesc   string `json:"role_desc" gorm:"column:role_desc"`
    UserId     int    `json:"user_id" gorm:"column:user_id"`        // 新增
    UserType   int    `json:"user_type" gorm:"column:user_type"`    // 新增
    Enable     int    `json:"enable"`
    CreateTime string `json:"create_time" gorm:"column:create_time"`
    UpdateTime string `json:"update_time" gorm:"column:update_time"`
}
```

### 3. 业务逻辑修改

**修改文件**: `services/admin_service/role.go`

#### 添加权限过滤函数
```go
// applyUserPermissionFilter 应用用户权限过滤
func applyUserPermissionFilter(userId, userType int) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        // 如果当前用户不是超管（user_type != 1），则只能查看自己创建的角色
        if userType != 1 {
            return db.Where("user_id = ?", userId)
        }
        // 超管可以查看所有角色
        return db
    }
}
```

#### 修改的方法
1. **GetRoleList**: 获取角色列表时应用权限过滤
2. **GetAllRoleList**: 获取所有角色列表时应用权限过滤
3. **UpdateRole**: 更新角色时检查权限
4. **DeleteRole**: 删除角色时检查权限
5. **GetRoleDetail**: 获取角色详情时检查权限
6. **SetRolePermission**: 设置角色权限时检查权限

### 4. 权限控制逻辑

- **超管用户** (`user_type = 1`): 可以查看和管理所有角色
- **普通管理员** (`user_type != 1`): 只能查看和管理自己创建的角色

### 5. 自动迁移脚本

**创建文件**: `scripts/apply_role_permissions_fix.sh`

该脚本具有以下功能：
- 自动读取项目 `.env` 文件中的数据库配置
- 解析数据库连接字符串 (DSN)
- 备份当前角色表结构
- 应用数据库迁移
- 验证迁移结果

## 使用方法

### 1. 应用数据库迁移

```bash
# 执行迁移脚本
./scripts/apply_role_permissions_fix.sh
```

脚本会：
1. 自动读取数据库配置
2. 检查数据库连接
3. 备份现有表结构
4. 应用迁移
5. 验证结果

### 2. 重启应用

```bash
# 重启应用以使代码修改生效
./stop.sh && ./start.sh
```

## 验证测试

### 1. 测试数据库配置解析

```bash
# 测试配置解析是否正常
./scripts/test_db_config.sh
```

### 2. 测试权限控制

1. **超管用户测试**:
   - 登录 user_type=1 的用户
   - 访问角色列表接口，应该能看到所有角色

2. **普通管理员测试**:
   - 登录 user_type!=1 的用户
   - 访问角色列表接口，应该只能看到自己创建的角色

## 接口影响

修改后的接口行为：
- `GET /api/v3/role/list`: 根据用户类型过滤角色列表
- `GET /api/v3/role/all`: 根据用户类型过滤所有角色
- `GET /api/v3/role/detail`: 检查角色访问权限
- `PUT /api/v3/role/update`: 检查角色修改权限
- `DELETE /api/v3/role/delete`: 检查角色删除权限
- `POST /api/v3/role/set/permission`: 检查权限设置权限

## 注意事项

1. **现有数据**: 脚本会将现有角色设置为超管创建（user_id=1, user_type=1）
2. **备份**: 脚本会自动备份表结构，如有问题可以恢复
3. **权限**: 确保执行脚本的用户有数据库操作权限
4. **测试**: 建议先在测试环境验证后再应用到生产环境

## 回滚方案

如果需要回滚，可以：
1. 使用备份的表结构文件恢复
2. 或者手动删除添加的字段：

```sql
ALTER TABLE `role` 
DROP COLUMN `user_id`,
DROP COLUMN `user_type`,
DROP COLUMN `role_name`,
DROP COLUMN `role_desc`,
DROP COLUMN `sort`,
DROP COLUMN `create_time`,
DROP COLUMN `update_time`;
```

## 总结

这次修复完全解决了角色权限控制的问题，确保：
- 数据安全：用户只能访问有权限的数据
- 权限隔离：不同类型用户有不同的访问权限
- 向后兼容：现有功能不受影响
- 易于维护：代码结构清晰，易于扩展 