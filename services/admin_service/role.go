package admin_service

import (
	"fmt"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RoleService struct{}

// AddRole 创建角色
func (s *RoleService) AddRole(c *gin.Context, params admin_model.AddRole, pessimism []int) error {
	userId := c.GetInt("uid")

	err := db.Dao.Transaction(func(tx *gorm.DB) error {
		// 检查是否存在相同的 role_name
		var existingRole admin_model.Role
		if err := tx.Where("user_id = ? AND role_name = ?", userId, params.RoleName).First(&existingRole).Error; err == nil {
			return fmt.Errorf("角色名 '%s' 已存在", params.RoleName)
		} else if err != gorm.ErrRecordNotFound {
			return err
		}

		// 创建角色
		if err := tx.Create(&params).Error; err != nil {
			return err
		}

		// 获取创建后的角色 ID
		roleId := params.Id

		// 创建角色权限
		if len(pessimism) > 0 {
			var rolePermissionList []admin_model.RolePermissionsPermission
			for _, permissionId := range pessimism {
				rolePermissionList = append(rolePermissionList, admin_model.RolePermissionsPermission{
					RoleId:       roleId,
					PermissionId: permissionId,
				})
			}
			if err := tx.Create(&rolePermissionList).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

// UpdateRole 更新角色 - 只修改名称和描述
func (s *RoleService) UpdateRole(c *gin.Context, params admin_model.UpdateRole, pessimism []int) error {
	userId := c.GetInt("uid")
	userType := c.GetInt("type")

	// 使用事务处理查询和更新操作
	err := db.Dao.Transaction(func(tx *gorm.DB) error {
		var role admin_model.Role
		// 根据 Id 查询对应的角色
		if err := tx.Where("id = ?", params.Id).First(&role).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("角色不存在")
			}
			return err
		}

		// 权限检查：如果不是超管，只能更新自己创建的角色
		if userType != 1 && role.UserId != userId {
			return fmt.Errorf("无权限更新此角色")
		}

		// 检查是否存在相同的 role_name（排除当前角色）
		var count int64
		if err := tx.Model(&admin_model.Role{}).Where("user_id = ? AND role_name = ? AND id != ?", userId, params.RoleName, params.Id).Count(&count).Error; err != nil {
			return err
		}

		if count > 0 {
			return fmt.Errorf("角色名 '%s' 已存在", params.RoleName)
		}

		// 只更新角色名称和描述，其他字段保持不变
		updateData := map[string]interface{}{
			"role_name":   params.RoleName,
			"role_desc":   params.RoleDesc,
			"update_time": time.Now(),
		}
		if err := tx.Model(&role).Updates(updateData).Error; err != nil {
			return err
		}

		// 不修改权限信息，保持原有权限不变
		// 注释掉权限相关的更新代码
		/*
			// 删除旧的角色权限
			if err := tx.Where("roleId = ?", params.Id).Delete(&admin_model.RolePermissionsPermission{}).Error; err != nil {
				return err
			}
			// 创建新的角色权限
			if len(pessimism) > 0 {
				var rolePermissionList []admin_model.RolePermissionsPermission
				for _, permissionId := range pessimism {
					rolePermissionList = append(rolePermissionList, admin_model.RolePermissionsPermission{
						RoleId:       params.Id,
						PermissionId: permissionId,
					})
				}
				if err := tx.Create(&rolePermissionList).Error; err != nil {
					return err
				}
			}
		*/

		return nil
	})

	return err
}

// GetRoleList 获取角色列表 - 性能优化版
func (s *RoleService) GetRoleList(c *gin.Context, params inout.GetRoleListReq) (interface{}, error) {
	var data []admin_model.Role
	var total int64

	// 获取当前用户信息
	currentUserId := c.GetInt("uid")

	// 设置默认分页参数
	params.Page = max(params.Page, 1)
	params.PageSize = max(params.PageSize, 10)

	// 构建查询 - 性能优化：只选择必要字段
	baseQuery := db.Dao.Model(&admin_model.Role{}).
		Select("id, role_name, role_desc, user_id, user_type, enable, sort, create_time, update_time").
		Scopes(
			applyRoleNameFilter(params.RoleName),
			applyUserPermissionFilter(currentUserId, 0), // 传入0表示不是超管
		)

	// 计算总数 - 使用子查询优化
	countQuery := db.Dao.Model(&admin_model.Role{}).
		Select("COUNT(*)").
		Scopes(
			applyRoleNameFilter(params.RoleName),
			applyUserPermissionFilter(currentUserId, 0), // 传入0表示不是超管
		)

	// 并行执行计数和数据查询
	type queryResult struct {
		data  []admin_model.Role
		total int64
		err   error
	}

	resultChan := make(chan queryResult, 2)

	// 异步执行计数查询
	go func() {
		var count int64
		err := countQuery.Count(&count).Error
		resultChan <- queryResult{total: count, err: err}
	}()

	// 异步执行数据查询
	go func() {
		var roles []admin_model.Role
		offset := (params.Page - 1) * params.PageSize
		err := baseQuery.Order("id DESC").Offset(offset).Limit(params.PageSize).Find(&roles).Error
		resultChan <- queryResult{data: roles, err: err}
	}()

	// 等待两个查询完成
	var countResult, dataResult queryResult
	for i := 0; i < 2; i++ {
		result := <-resultChan
		if result.err != nil {
			return nil, result.err
		}
		if result.data != nil {
			dataResult = result
		} else {
			countResult = result
		}
	}

	data = dataResult.data
	total = countResult.total

	// 格式化数据（包含创建人信息）
	formattedData, err := s.formatRoleDataWithCreatorInfo(data)
	if err != nil {
		return nil, err
	}

	// 构建响应
	responseData := inout.GetRoleListResp{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Items:    formattedData,
	}

	return responseData, nil
}

// applyRoleNameFilter 根据角色名称过滤
func applyRoleNameFilter(roleName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if roleName != "" {
			return db.Where("role_name like ?", "%"+roleName+"%")
		}
		return db
	}
}

// applyUserPermissionFilter 应用用户权限过滤
func applyUserPermissionFilter(userId, userType int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// 所有用户都只能查看自己创建的角色
		return db.Where("user_id = ?", userId)
	}
}

// formatRoleData 格式化角色数据
func formatRoleData(data []admin_model.Role) []inout.RoleListItem {
	formattedData := make([]inout.RoleListItem, len(data))
	for i, item := range data {
		createTime, _ := time.Parse(time.RFC3339, item.CreateTime)
		updateTime, _ := time.Parse(time.RFC3339, item.UpdateTime)
		formattedData[i] = inout.RoleListItem{
			Id:         item.Id,
			RoleName:   item.RoleName,
			RoleDesc:   item.RoleDesc,
			Enable:     item.Enable,
			CreateTime: createTime.Format("2006-01-02 15:04:05"),
			UpdateTime: updateTime.Format("2006-01-02 15:04:05"),
		}
	}
	return formattedData
}

// SetRolePermission 设置角色权限
func (s *RoleService) SetRolePermission(c *gin.Context, params inout.SetRolePermissionReq) error {
	userId := c.GetInt("uid")

	err := db.Dao.Transaction(func(tx *gorm.DB) error {
		// 先查询角色是否存在，并检查权限
		var role admin_model.Role
		if err := tx.Where("id = ?", params.Id).First(&role).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("角色不存在")
			}
			return err
		}

		// 权限检查：所有用户都只能设置自己创建的角色权限
		if role.UserId != userId {
			return fmt.Errorf("无权限设置此角色权限")
		}

		// 删除旧的角色权限
		if err := tx.Where("roleId = ?", params.Id).Delete(&admin_model.RolePermissionsPermission{}).Error; err != nil {
			return err
		}

		// 创建新的角色权限
		var rolePermissionList []admin_model.RolePermissionsPermission
		fmt.Println("Raw Permission:", params.Permission)

		for _, permissionIdStr := range params.Permission {
			// 检查是否是逗号分隔的字符串
			if strings.Contains(permissionIdStr, ",") {
				// 按逗号分割字符串
				permissionIdParts := strings.Split(permissionIdStr, ",")
				for _, part := range permissionIdParts {
					permissionId, err := strconv.Atoi(strings.TrimSpace(part)) // 去掉可能的空格
					if err != nil {
						return fmt.Errorf("无效的权限ID: %s: %w", part, err)
					}
					rolePermissionList = append(rolePermissionList, admin_model.RolePermissionsPermission{
						RoleId:       params.Id,
						PermissionId: permissionId,
					})
				}
			} else {
				// 原有逻辑，处理单个字符串
				permissionId, err := strconv.Atoi(permissionIdStr)
				if err != nil {
					return fmt.Errorf("无效的权限ID: %s: %w", permissionIdStr, err)
				}
				rolePermissionList = append(rolePermissionList, admin_model.RolePermissionsPermission{
					RoleId:       params.Id,
					PermissionId: permissionId,
				})
			}
		}

		// 插入角色权限
		if len(rolePermissionList) > 0 {
			if err := tx.Create(&rolePermissionList).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

// DeleteRole 删除角色

func (s *RoleService) DeleteRole(c *gin.Context, id int) error {
	userId := c.GetInt("uid")

	err := db.Dao.Transaction(func(tx *gorm.DB) error {
		// 先查询角色是否存在，并检查权限
		var role admin_model.Role
		if err := tx.Where("id = ?", id).First(&role).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("角色不存在")
			}
			return err
		}

		// 权限检查：所有用户都只能删除自己创建的角色
		if role.UserId != userId {
			return fmt.Errorf("无权限删除此角色")
		}

		// 删除角色
		if err := tx.Where("id = ?", id).Delete(&admin_model.Role{}).Error; err != nil {
			return err
		}

		// 删除角色权限
		if err := tx.Where("roleId = ?", id).Delete(&admin_model.RolePermissionsPermission{}).Error; err != nil {
			return err
		}

		return nil
	})

	return err
}

// GetAllRoleList
func (s *RoleService) GetAllRoleList(c *gin.Context) (interface{}, error) {
	// 获取当前用户信息
	currentUserId := c.GetInt("uid")

	var roles []admin_model.Role
	query := db.Dao.Model(&admin_model.Role{}).Scopes(
		applyUserPermissionFilter(currentUserId, 0), // 传入0表示不是超管
	)

	err := query.Find(&roles).Error
	if err != nil {
		return nil, err
	}

	// 格式化数据（包含创建人信息）
	formattedData, err := s.formatRoleDataWithCreatorInfo(roles)
	if err != nil {
		return nil, err
	}

	return formattedData, nil
}

// GetRoleDetail 获取角色详情
func (s *RoleService) GetRoleDetail(c *gin.Context, id int) (interface{}, error) {
	userId := c.GetInt("uid")

	// 使用一个事务查询所有需要的数据
	var roleDetail inout.RoleDetail

	err := db.Dao.Transaction(func(tx *gorm.DB) error {
		// 1. 查询角色基本信息
		var role admin_model.Role
		if err := tx.Where("id = ?", id).First(&role).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("角色(ID=%d)不存在", id)
			}
			return fmt.Errorf("查询角色信息失败: %w", err)
		}

		// 权限检查：所有用户都只能查看自己创建的角色
		if role.UserId != userId {
			return fmt.Errorf("无权限查看此角色详情")
		}

		// 2. 直接查询权限ID列表 (优化查询)
		var permissionIds []int
		if err := tx.Model(&admin_model.RolePermissionsPermission{}).
			Where("roleId = ?", id).
			Pluck("permissionId", &permissionIds).Error; err != nil {
			return fmt.Errorf("查询角色权限失败: %w", err)
		}

		roleDetail.SelectIds = permissionIds
		return nil
	})

	if err != nil {
		return nil, err
	}

	return roleDetail, nil
}

// formatRoleDataWithCreatorInfo 格式化角色数据（通过user_id查询创建人信息）- 性能优化版
func (s *RoleService) formatRoleDataWithCreatorInfo(roles []admin_model.Role) ([]inout.RoleListItem, error) {
	if len(roles) == 0 {
		return []inout.RoleListItem{}, nil
	}

	// 收集所有的创建人ID - 优化：使用map去重
	userIdSet := make(map[int]struct{})
	for _, role := range roles {
		if role.UserId > 0 {
			userIdSet[role.UserId] = struct{}{}
		}
	}

	// 转换为slice
	userIds := make([]int, 0, len(userIdSet))
	for userId := range userIdSet {
		userIds = append(userIds, userId)
	}

	// 批量查询用户信息 - 性能优化：只查询必要字段
	userInfoMap := make(map[int]string) // userId -> username
	if len(userIds) > 0 {
		type UserBasic struct {
			ID       int    `gorm:"column:id"`
			Username string `gorm:"column:username"`
		}

		var users []UserBasic
		err := db.Dao.Table("user").Select("id, username").Where("id IN ?", userIds).Find(&users).Error
		if err != nil {
			return nil, fmt.Errorf("查询用户信息失败: %w", err)
		}

		// 构建用户信息映射
		for _, user := range users {
			userInfoMap[user.ID] = user.Username
		}
	}

	// 格式化数据 - 预分配切片容量
	formattedData := make([]inout.RoleListItem, len(roles))
	for i, role := range roles {
		// 时间解析优化：只在需要时解析
		var createTimeStr, updateTimeStr string
		if role.CreateTime != "" {
			if createTime, err := time.Parse(time.RFC3339, role.CreateTime); err == nil {
				createTimeStr = createTime.Format("2006-01-02 15:04:05")
			} else {
				createTimeStr = role.CreateTime // 如果解析失败，使用原始值
			}
		}

		if role.UpdateTime != "" {
			if updateTime, err := time.Parse(time.RFC3339, role.UpdateTime); err == nil {
				updateTimeStr = updateTime.Format("2006-01-02 15:04:05")
			} else {
				updateTimeStr = role.UpdateTime // 如果解析失败，使用原始值
			}
		}

		// 设置创建人类型描述 - 使用查找表优化
		creatorTypeDesc := getCreatorTypeDesc(role.UserType)

		// 获取创建人信息
		creatorName := "未知用户"
		if username, exists := userInfoMap[role.UserId]; exists {
			creatorName = username
		}

		formattedData[i] = inout.RoleListItem{
			Id:              role.Id,
			RoleName:        role.RoleName,
			RoleDesc:        role.RoleDesc,
			Enable:          role.Enable,
			Sort:            role.Sort,
			CreateTime:      createTimeStr,
			UpdateTime:      updateTimeStr,
			CreatorId:       role.UserId,
			CreatorName:     creatorName,
			CreatorType:     role.UserType,
			CreatorTypeDesc: creatorTypeDesc,
		}
	}

	return formattedData, nil
}

// getCreatorTypeDesc 获取创建人类型描述 - 使用查找表优化
func getCreatorTypeDesc(userType int) string {
	switch userType {
	case 1:
		return "超级管理员"
	case 2:
		return "普通管理员"
	default:
		return "未知"
	}
}
