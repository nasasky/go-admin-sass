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

// UpdateRole 更新角色
func (s *RoleService) UpdateRole(c *gin.Context, params admin_model.UpdateRole, pessimism []int) error {
	userId := c.GetInt("uid")
	userType := c.GetInt("type")
	if userType == 2 {
		params.Code = "STORE"
	} else {
		params.Code = "SUPER_ADMIN"
	}
	params.UserId = userId
	params.UserType = userType
	params.UpdateTime = time.Now()

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

		// 检查是否存在相同的 role_name
		var count int64
		if err := tx.Model(&admin_model.Role{}).Where("user_id = ? AND role_name = ? AND id != ?", userId, params.RoleName, params.Id).Count(&count).Error; err != nil {
			return err
		}

		if count > 0 {
			return fmt.Errorf("角色名 '%s' 已存在", params.RoleName)
		}

		// 更新角色
		updateData := map[string]interface{}{
			"role_name":   params.RoleName,
			"role_desc":   params.RoleDesc,
			"code":        params.Code,
			"user_id":     params.UserId,
			"user_type":   params.UserType,
			"update_time": params.UpdateTime,
			"enable":      params.Enable,
			"sort":        params.Sort,
		}
		if err := tx.Model(&role).Updates(updateData).Error; err != nil {
			return err
		}

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

		return nil
	})

	return err
}

// GetRoleList 获取角色列表
func (s *RoleService) GetRoleList(c *gin.Context, params inout.GetRoleListReq) (interface{}, error) {
	var data []admin_model.Role
	var total int64

	// 设置默认分页参数
	params.Page = max(params.Page, 1)
	params.PageSize = max(params.PageSize, 10)

	// 构建查询
	query := db.Dao.Model(&admin_model.Role{}).Scopes(
		applyRoleNameFilter(params.RoleName),
	)

	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize

	// 执行查询
	err := query.Count(&total).Offset(offset).Limit(params.PageSize).Find(&data).Error
	if err != nil {
		return nil, err

	}

	// 格式化数据
	formattedData := formatRoleData(data)

	// 构建响应

	resopnseData := inout.GetRoleListResp{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Items:    formattedData,
	}
	fmt.Println(resopnseData)

	return resopnseData, nil
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

	err := db.Dao.Transaction(func(tx *gorm.DB) error {
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
	err := db.Dao.Transaction(func(tx *gorm.DB) error {
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

	var roles []admin_model.Role
	err := db.Dao.Model(&admin_model.Role{}).Find(&roles).Error
	if err != nil {
		return nil, err
	}

	// 格式化数据
	formattedData := make([]inout.RoleListItem, len(roles))
	for i, item := range roles {
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

	return formattedData, nil
}

// GetRoleDetail 获取角色详情
func (s *RoleService) GetRoleDetail(c *gin.Context, id int) (interface{}, error) {
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
