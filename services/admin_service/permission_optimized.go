package admin_service

import (
	"context"
	"fmt"
	"nasa-go-admin/db"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/pkg/cache"
	"nasa-go-admin/redis"
	"time"
)

// PermissionService 优化的权限服务
type PermissionService struct {
	cache *cache.CacheManager
}

// NewPermissionService 创建权限服务
func NewPermissionService() *PermissionService {
	return &PermissionService{
		cache: cache.NewCacheManager(redis.GetClient()),
	}
}

// UserPermission 用户权限结构
type UserPermission struct {
	UserID      int                          `json:"user_id"`
	UserType    int                          `json:"user_type"`
	RoleID      int                          `json:"role_id"`
	Permissions []admin_model.PermissionUser `json:"permissions"`
	PermTree    interface{}                  `json:"perm_tree"`
	Rules       []string                     `json:"rules"`
}

// GetUserPermissions 获取用户权限（优化版本）
func (s *PermissionService) GetUserPermissions(ctx context.Context, userID int) (*UserPermission, error) {
	// 1. 尝试从缓存获取
	cacheKey := fmt.Sprintf("user:permissions:%d", userID)
	var cached UserPermission
	if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
		return &cached, nil
	}

	// 2. 从数据库查询
	result, err := s.fetchUserPermissionsFromDB(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 3. 缓存结果（15分钟）
	s.cache.Set(ctx, cacheKey, result, 15*time.Minute)

	return result, nil
}

// fetchUserPermissionsFromDB 从数据库获取用户权限
func (s *PermissionService) fetchUserPermissionsFromDB(ctx context.Context, userID int) (*UserPermission, error) {
	result := &UserPermission{UserID: userID}

	// 1. 获取用户类型和角色ID（单次查询）
	var user struct {
		UserType int `gorm:"column:user_type"`
		RoleID   int `gorm:"column:role_id"`
	}

	if err := db.Dao.WithContext(ctx).
		Select("user_type, role_id").
		Table("user").
		Where("id = ?", userID).
		First(&user).Error; err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	result.UserType = user.UserType
	result.RoleID = user.RoleID

	// 2. 根据用户类型使用不同的查询策略
	if user.UserType == 1 { // 超级管理员
		return s.getSuperAdminPermissions(ctx, result)
	}

	// 3. 普通用户：一次性获取所有权限ID
	var permissionIDs []int
	if err := db.Dao.WithContext(ctx).
		Table("role_permissions_permission").
		Where("roleId = ?", user.RoleID).
		Pluck("permissionId", &permissionIDs).Error; err != nil {
		return nil, fmt.Errorf("获取角色权限失败: %w", err)
	}

	if len(permissionIDs) == 0 {
		result.Permissions = []admin_model.PermissionUser{}
		result.PermTree = []interface{}{}
		result.Rules = []string{}
		return result, nil
	}

	// 4. 一次性获取所有权限信息
	var allPermissions []admin_model.PermissionUser
	if err := db.Dao.WithContext(ctx).
		Where("id IN ?", permissionIDs).
		Order("sort DESC").
		Find(&allPermissions).Error; err != nil {
		return nil, fmt.Errorf("获取权限详情失败: %w", err)
	}

	result.Permissions = allPermissions

	// 5. 构建权限树（内存操作，避免N+1查询）
	tree := s.buildPermissionTree(allPermissions)
	result.PermTree = tree

	// 6. 提取规则
	rules := make([]string, 0, len(allPermissions))
	for _, perm := range allPermissions {
		if perm.Rule != "" {
			rules = append(rules, perm.Rule)
		}
	}
	result.Rules = rules

	return result, nil
}

// getSuperAdminPermissions 获取超级管理员权限
func (s *PermissionService) getSuperAdminPermissions(ctx context.Context, result *UserPermission) (*UserPermission, error) {
	// 超级管理员获取所有权限
	var allPermissions []admin_model.PermissionUser
	if err := db.Dao.WithContext(ctx).
		Order("sort DESC").
		Find(&allPermissions).Error; err != nil {
		return nil, fmt.Errorf("获取所有权限失败: %w", err)
	}

	result.Permissions = allPermissions
	result.PermTree = s.buildPermissionTree(allPermissions)

	// 提取所有规则
	rules := make([]string, 0, len(allPermissions))
	for _, perm := range allPermissions {
		if perm.Rule != "" {
			rules = append(rules, perm.Rule)
		}
	}
	result.Rules = rules

	return result, nil
}

// buildPermissionTree 构建权限树（内存操作，避免数据库查询）
func (s *PermissionService) buildPermissionTree(permissions []admin_model.PermissionUser) []interface{} {
	// 创建权限映射
	permMap := make(map[int]admin_model.PermissionUser)
	for _, perm := range permissions {
		permMap[perm.ID] = perm
	}

	// 找到根节点
	var roots []interface{}
	childrenMap := make(map[int][]admin_model.PermissionUser)

	// 分组：找出每个节点的子节点
	for _, perm := range permissions {
		if perm.ParentId == 0 {
			// 根节点
			node := s.buildPermissionNode(perm, permMap, childrenMap)
			roots = append(roots, node)
		} else {
			// 子节点，添加到父节点的children中
			children := childrenMap[perm.ParentId]
			children = append(children, perm)
			childrenMap[perm.ParentId] = children
		}
	}

	// 递归构建树结构
	return s.buildTreeRecursive(roots, permMap, childrenMap)
}

// buildPermissionNode 构建权限节点
func (s *PermissionService) buildPermissionNode(perm admin_model.PermissionUser, permMap map[int]admin_model.PermissionUser, childrenMap map[int][]admin_model.PermissionUser) map[string]interface{} {
	node := map[string]interface{}{
		"id":        perm.ID,
		"label":     perm.Label,
		"title":     perm.Title,
		"key":       perm.Key,
		"type":      perm.Type,
		"path":      perm.Path,
		"icon":      perm.Icon,
		"layout":    perm.Layout,
		"show":      perm.Show,
		"enable":    perm.Enable,
		"keepAlive": perm.KeepAlive,
		"order":     perm.Order,
		"parent_id": perm.ParentId,
		"rule":      perm.Rule,
		"is_link":   perm.IsLink,
		"children":  []interface{}{},
	}

	return node
}

// buildTreeRecursive 递归构建树结构
func (s *PermissionService) buildTreeRecursive(nodes []interface{}, permMap map[int]admin_model.PermissionUser, childrenMap map[int][]admin_model.PermissionUser) []interface{} {
	for i, node := range nodes {
		if nodeMap, ok := node.(map[string]interface{}); ok {
			if id, ok := nodeMap["id"].(int); ok {
				children := childrenMap[id]
				var childNodes []interface{}
				for _, child := range children {
					childNode := s.buildPermissionNode(child, permMap, childrenMap)
					childNodes = append(childNodes, childNode)
				}
				// 递归处理子节点
				if len(childNodes) > 0 {
					nodeMap["children"] = s.buildTreeRecursive(childNodes, permMap, childrenMap)
				}
				nodes[i] = nodeMap
			}
		}
	}
	return nodes
}

// InvalidateUserPermissionCache 清除用户权限缓存
func (s *PermissionService) InvalidateUserPermissionCache(ctx context.Context, userID int) error {
	cacheKey := fmt.Sprintf("user:permissions:%d", userID)
	return s.cache.Delete(ctx, cacheKey)
}

// InvalidateAllPermissionCache 清除所有权限缓存
func (s *PermissionService) InvalidateAllPermissionCache(ctx context.Context) error {
	// 可以实现通配符删除，这里简化处理
	return s.cache.Clear(ctx)
}

// BatchGetUserPermissions 批量获取多个用户权限
func (s *PermissionService) BatchGetUserPermissions(ctx context.Context, userIDs []int) (map[int]*UserPermission, error) {
	result := make(map[int]*UserPermission)
	var missedUserIDs []int

	// 1. 尝试从缓存批量获取
	for _, userID := range userIDs {
		cacheKey := fmt.Sprintf("user:permissions:%d", userID)
		var cached UserPermission
		if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
			result[userID] = &cached
		} else {
			missedUserIDs = append(missedUserIDs, userID)
		}
	}

	// 2. 对于缓存未命中的用户，批量查询数据库
	if len(missedUserIDs) > 0 {
		// 批量查询用户信息
		var users []struct {
			ID       int `gorm:"column:id"`
			UserType int `gorm:"column:user_type"`
			RoleID   int `gorm:"column:role_id"`
		}

		if err := db.Dao.WithContext(ctx).
			Select("id, user_type, role_id").
			Table("user").
			Where("id IN ?", missedUserIDs).
			Find(&users).Error; err != nil {
			return nil, fmt.Errorf("批量获取用户信息失败: %w", err)
		}

		// 为每个用户获取权限并缓存
		for _, user := range users {
			perm, err := s.fetchUserPermissionsFromDB(ctx, user.ID)
			if err != nil {
				continue // 跳过错误的用户，不影响其他用户
			}
			result[user.ID] = perm

			// 异步缓存
			go func(userID int, permission *UserPermission) {
				cacheKey := fmt.Sprintf("user:permissions:%d", userID)
				s.cache.Set(context.Background(), cacheKey, permission, 15*time.Minute)
			}(user.ID, perm)
		}
	}

	return result, nil
}
