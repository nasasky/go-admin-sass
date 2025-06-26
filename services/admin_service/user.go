package admin_service

import (
	"crypto/md5"
	"errors"
	"fmt"
	"log"
	"nasa-go-admin/db"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/pkg/jwt"
	"nasa-go-admin/pkg/monitoring"
	"nasa-go-admin/redis"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TenantsService struct{}

// CreateUser
func (s *TenantsService) CreateUser(username, password, phone string, usertype int, role int) (*admin_model.InsertUser, error) {
	var newUserApp admin_model.InsertUser
	if usertype == 0 {
		return nil, errors.New("roleId 不能为空")
	}
	err := db.Dao.Transaction(func(tx *gorm.DB) error {
		newUserApp = admin_model.InsertUser{
			Username:   username,
			Password:   fmt.Sprintf("%x", md5.Sum([]byte(password))),
			Phone:      phone,
			UserType:   usertype,
			RoleId:     role,
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}
		err := tx.Create(&newUserApp).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 记录用户注册指标
	monitoring.RecordUserRegistration()
	monitoring.SaveBusinessMetric("user_register", newUserApp.Username)

	return &newUserApp, nil
}

// UpdateUser
func (s *TenantsService) UpdateUser(id int, username, password, phone string, usertype int, role int) (*admin_model.InsertUser, error) {
	var newUserApp admin_model.InsertUser
	if usertype == 0 {
		return nil, errors.New("roleId 不能为空")
	}
	err := db.Dao.Transaction(func(tx *gorm.DB) error {
		newUserApp = admin_model.InsertUser{
			Username:   username,
			Password:   fmt.Sprintf("%x", md5.Sum([]byte(password))),
			Phone:      phone,
			UserType:   usertype,
			RoleId:     role,
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}
		err := tx.Where("id = ?", id).Updates(&newUserApp).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &newUserApp, nil
}

func (s *TenantsService) UserExists(username string) bool {
	var existingUser admin_model.AdminUser
	if err := db.Dao.Where("username = ?", username).First(&existingUser).Error; err == nil {
		return true
	}
	return false
}

// Login
func (s *TenantsService) LoginTenants(c *gin.Context, username, password string) (map[string]interface{}, error) {
	var users admin_model.AdminUser
	hashedPassword := fmt.Sprintf("%x", md5.Sum([]byte(password)))
	err := db.Dao.Where("(username = ? OR phone = ?) AND password = ?", username, username, hashedPassword).First(&users).Error

	// 如果没有找到记录，则返回提示用户不存在
	if err != nil {
		Resp.Err(c, 20001, "用户不存在或密码错误")
		return nil, err
	}

	// 如果密码不正确，则返回提示密码错误
	if users.Password != hashedPassword {
		Resp.Err(c, 20001, "密码错误")
		return nil, fmt.Errorf("密码错误")
	}

	token, err := jwt.GenerateAdminToken(users.ID, users.RoleId, users.UserType, time.Hour*24)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}
	users.Token = token
	expiration := time.Hour * 24 // 过期时间为 24 小时
	// 存储 Token
	err = redis.StoreToken(strconv.Itoa(users.ID), users.Token, expiration)
	if err != nil {
		// 如果存储 Token 失败，直接返回错误
		return nil, err
	}

	// 存储用户信息
	err = redis.StoreUserInfo(strconv.Itoa(users.ID), map[string]interface{}{
		"username": users.Username,
		"phone":    users.Phone,
		"roleId":   users.RoleId,
		"userType": users.UserType,
		"token":    users.Token,
		"parentId": users.ParentId,
	}, expiration)
	if err != nil {
		// 如果存储用户信息失败，直接返回错误
		return nil, err
	}

	permissions := getPermissionListByRoleId(users.RoleId)

	// 记录用户登录指标
	monitoring.RecordUserLogin()
	monitoring.SaveBusinessMetric("user_login", users.Username)

	// 过滤掉值为空的字段
	responseData := map[string]interface{}{
		"user":        users,
		"permissions": permissions,
		"token":       users.Token,
	}
	return responseData, nil
}

// Login
func (s *TenantsService) Login(c *gin.Context, username, password string) (map[string]interface{}, error) {
	var user admin_model.AdminUser
	hashedPassword := fmt.Sprintf("%x", md5.Sum([]byte(password)))
	err := db.Dao.Where("(username = ? OR phone = ?) AND password = ?", username, username, hashedPassword).First(&user).Error

	// 如果没有找到记录，则返回提示用户不存在
	if err != nil {
		Resp.Err(c, 20001, "用户不存在或密码错误")
		return nil, err
	}

	// 如果密码不正确，则返回提示密码错误
	if user.Password != hashedPassword {
		Resp.Err(c, 20001, "密码错误")
		return nil, fmt.Errorf("密码错误")
	}

	token, err := jwt.GenerateAdminToken(user.ID, user.RoleId, user.UserType, time.Hour*24)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}
	user.Token = token
	expiration := time.Hour * 24 // 过期时间为 24 小时
	err = redis.StoreToken(strconv.Itoa(user.ID), user.Token, expiration)

	if err != nil {
		return nil, err
	}
	err = redis.DeleteKey(strconv.Itoa(user.ID)) // 删除旧键
	if err != nil && err != redis.ErrNil {
		return nil, fmt.Errorf("failed to delete old key: %v", err)
	}

	err = redis.StoreUserInfo(strconv.Itoa(user.ID), map[string]interface{}{
		"username": user.Username,
		"phone":    user.Phone,
		"roleId":   user.RoleId,
		"userType": user.UserType,
		"token":    user.Token,
		"parentId": user.ParentId,
	}, expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to store user info: %v", err)
	}
	permissions := getPermissionListByRoleId(user.RoleId)

	// 记录用户登录指标
	monitoring.RecordUserLogin()
	monitoring.SaveBusinessMetric("user_login", user.Username)

	responseData := map[string]interface{}{
		"user":        user,
		"permissions": permissions,
		"token":       user.Token,
	}
	fmt.Println(user)
	c.Set("userInfo", user)
	return responseData, nil
}

// GetUserInfo
func (s *TenantsService) GetUserInfo(c *gin.Context, id int) (map[string]interface{}, error) {
	var user admin_model.AdminUser
	log.Println("User ID:", id)
	err := db.Dao.Where("id = ?", id).First(&user).Error
	if err != nil {
		Resp.Err(c, 20001, "用户不存在")
		return nil, err
	}

	token, err := jwt.GenerateAdminToken(user.ID, user.RoleId, user.UserType, time.Hour*24)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}
	user.Token = token
	expiration := time.Hour * 24 // 过期时间为 24 小时
	err = redis.StoreToken(strconv.Itoa(user.ID), user.Token, expiration)
	if err != nil {
		Resp.Err(c, 20001, "存储Token失败")
		return nil, err
	}
	log.Println("User:", user.RoleId)
	permissions := getPermissionListByRoleId(user.RoleId)

	// 过滤掉值为空的字段

	responseData := map[string]interface{}{
		"user":        user,
		"permissions": permissions,
		"token":       user.Token,
	}
	return responseData, nil
}

// 读取路由菜单
func (s *TenantsService) GetRoutes(c *gin.Context, id int) ([]admin_model.PermissionUser, error) {
	// 使用新的优化权限服务
	permissionService := NewPermissionService()

	// 获取用户权限菜单树
	userPermissions, err := permissionService.GetUserPermissions(c, id)
	fmt.Println(userPermissions)
	if err != nil {
		log.Printf("Failed to get user permissions: %v", err)
		return nil, err
	}

	// 构建权限树结构（基于原始权限列表）
	permissionTree := s.buildUserPermissionTree(userPermissions.Permissions)

	// 过滤掉值为空的字段
	filteredPermissList := filterEmptyFields(permissionTree)
	return filteredPermissList, nil
}

// 读取路由菜单
func (s *TenantsService) GetMenus(c *gin.Context, id int) ([]admin_model.PermissionUser, error) {
	// 使用新的优化权限服务
	permissionService := NewPermissionService()

	// 获取用户权限菜单树
	userPermissions, err := permissionService.GetUserPermissions(c, id)
	if err != nil {
		log.Printf("Failed to get user permissions: %v", err)
		return nil, err
	}

	// 构建权限树结构（基于原始权限列表）
	permissionTree := s.buildUserPermissionTree(userPermissions.Permissions)

	// 过滤掉值为空的字段
	filteredPermissList := filterEmptyFieldsmenu(permissionTree)
	return filteredPermissList, nil
}

// AddMenu
func (s *TenantsService) AddMenu(c *gin.Context, params admin_model.PermissionMenu) (int, error) {
	// 获取用户ID和角色ID
	userId, userExists := c.Get("uid")
	roleId, roleExists := c.Get("rid")

	if !userExists || !roleExists {
		return 0, fmt.Errorf("用户ID或角色ID不存在")
	}

	userID, ok := userId.(int)
	if !ok {
		return 0, fmt.Errorf("用户ID类型错误")
	}

	roleID, ok := roleId.(int)
	fmt.Println("id值", userID, roleID)
	if !ok {
		return 0, fmt.Errorf("角色ID类型错误")
	}

	//判断ParentId是否传过来，没有保存数据库为null

	var existingMenu admin_model.PermissionMenu

	// 检查标签、规则或键是否已存在
	err := db.Dao.Where("((rule = ? AND rule != ''))", params.Rule).First(&existingMenu).Error
	if err == nil {
		return 0, fmt.Errorf("menu with rule '%s'已存在", params.Rule)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	// 设置创建和更新时间
	params.CreateTime = time.Now()
	params.UpdateTime = time.Now()
	err = db.Dao.Create(&params).Error
	if err != nil {
		return 0, err
	}

	// 确保 params.ID 在创建后被设置
	if params.Id == 0 {
		return 0, fmt.Errorf("failed to retrieve the ID of the newly created menu")
	}

	// 在事务中插入 staffPermission
	err = db.Dao.Transaction(func(tx *gorm.DB) error {
		staffPermission := admin_model.StaffPermissions{
			TenantId:     userID,
			RoleId:       roleID,
			PermissionId: params.Id,
		}
		log.Println("Inserting staffPermission:", staffPermission)
		if err := tx.Create(&staffPermission).Error; err != nil {
			log.Println("Error inserting staffPermission:", err)
			return err
		}
		return nil
	})
	if err != nil {
		log.Println("Transaction error:", err)
		return 0, err
	}
	err = s.AddMenuToRole(c, roleID, params.Id)
	if err != nil {
		return 0, err
	}

	// 清除权限缓存
	permissionService := NewPermissionService()
	if err := permissionService.InvalidateUserPermissionCache(c, userID); err != nil {
		log.Printf("Failed to invalidate user permission cache: %v", err)
	}

	// 返回新菜单的ID
	return params.Id, nil
}

// 把菜单id插入到角色权限表
func (s *TenantsService) AddMenuToRole(c *gin.Context, roleId, menuId int) error {
	// 在事务中插入 staffPermission
	err := db.Dao.Transaction(func(tx *gorm.DB) error {
		rolePermission := admin_model.RolePermissionsPermission{
			RoleId:       roleId,
			PermissionId: menuId,
		}
		log.Println("Inserting rolePermission:", rolePermission)
		if err := tx.Create(&rolePermission).Error; err != nil {
			log.Println("Error inserting rolePermission:", err)
			return err
		}
		return nil
	})
	if err != nil {
		log.Println("Transaction error:", err)
		return err
	}
	return nil
}

// GetMenuDetail
func (s *TenantsService) GetMenuDetail(c *gin.Context, id int) (admin_model.PermissionMenu, error) {
	var menu admin_model.PermissionMenu
	err := db.Dao.Where("id = ?", id).First(&menu).Error
	if err != nil {
		return admin_model.PermissionMenu{}, err
	}
	return menu, nil
}

// UpdateMenu
func (s *TenantsService) UpdateMenu(c *gin.Context, id int, params admin_model.PermissionMenu) (int, error) {
	//修改菜单信息
	err := db.Dao.Model(&admin_model.PermissionMenu{}).Where("id = ?", id).Updates(params).Error
	if err != nil {
		return 0, err
	}

	// 清除权限缓存
	userId, _ := c.Get("uid")
	if userID, ok := userId.(int); ok {
		permissionService := NewPermissionService()
		if err := permissionService.InvalidateUserPermissionCache(c, userID); err != nil {
			log.Printf("Failed to invalidate user permission cache: %v", err)
		}
	}

	// 更新 路由id
	return params.Id, nil
}

func (s *TenantsService) DeleteMenu(c *gin.Context, ids []int) error {
	// 收集所有需要删除的菜单ID（包括子菜单）
	var allMenuIds []int

	err := db.Dao.Transaction(func(tx *gorm.DB) error {
		// 递归查找所有子菜单ID
		err := findAllChildMenuIds(tx, ids, &allMenuIds)
		if err != nil {
			return fmt.Errorf("查找子菜单失败: %v", err)
		}

		// 合并原始ID和子菜单ID
		allMenuIds = append(allMenuIds, ids...)

		// 删除所有相关菜单
		if err := tx.Where("id IN ?", allMenuIds).Delete(&admin_model.PermissionMenu{}).Error; err != nil {
			return fmt.Errorf("删除菜单失败: %v", err)
		}

		// 删除所有相关 staffPermission 记录
		if err := tx.Where("permission_id IN ?", allMenuIds).Delete(&admin_model.StaffPermissions{}).Error; err != nil {
			return fmt.Errorf("删除 staffPermission 记录失败: %v", err)
		}

		// 删除所有相关 role_permissions_permission 记录
		if err := tx.Where("permissionId IN ?", allMenuIds).Delete(&admin_model.RolePermissionsPermission{}).Error; err != nil {
			return fmt.Errorf("删除 role_permissions_permission 记录失败: %v", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// 清除权限缓存
	userId, _ := c.Get("uid")
	if userID, ok := userId.(int); ok {
		permissionService := NewPermissionService()
		if err := permissionService.InvalidateUserPermissionCache(c, userID); err != nil {
			log.Printf("Failed to invalidate user permission cache: %v", err)
		}
	}

	return nil
}

// 递归查找所有子菜单ID
func findAllChildMenuIds(tx *gorm.DB, parentIds []int, result *[]int) error {
	if len(parentIds) == 0 {
		return nil
	}

	var childIds []int
	if err := tx.Model(&admin_model.PermissionMenu{}).
		Where("parent_id IN ?", parentIds).
		Pluck("id", &childIds).Error; err != nil {
		return err
	}

	// 添加到结果中
	*result = append(*result, childIds...)

	// 递归查找下一级子菜单
	if len(childIds) > 0 {
		return findAllChildMenuIds(tx, childIds, result)
	}

	return nil
}

//	func getPermissionListByRoleId(roleId int) []admin_model.PermissionUser {
//		var permisIdList []int
//		db.Dao.Model(admin_model.RolePermissionsPermission{}).Where("roleId = ?", roleId).Pluck("permissionId", &permisIdList)
//		var permisList []admin_model.PermissionUser
//		log.Println("permisIdList:", permisIdList)
//		db.Dao.Model(admin_model.PermissionUser{}).Where("id in(?)", permisIdList).Find(&permisList)
//		log.Println("permisList:", permisList)
//		return permisList
//
// }
func getPermissionListByRoleId(roleId int) []string {
	var permisIdList []int
	if roleId != 0 {
		db.Dao.Model(admin_model.RolePermissionsPermission{}).Where("roleId = ?", roleId).Pluck("permissionId", &permisIdList)
	} else {
		// 如果 roleId 为 0，查询所有权限的 ID
		db.Dao.Model(admin_model.PermissionUser{}).Pluck("id", &permisIdList)
	}

	type Permission struct {
		Rule        string
		PermissRule string
	}
	var permissions []Permission
	if len(permisIdList) > 0 {
		db.Dao.Model(admin_model.PermissionUser{}).Where("id IN (?)", permisIdList).Select("rule, permiss_rule").Find(&permissions)
	}

	// 过滤掉空的规则
	var filteredRules []string
	for _, permission := range permissions {
		if permission.Rule != "" {
			filteredRules = append(filteredRules, permission.Rule)
		}
		if permission.PermissRule != "" {
			filteredRules = append(filteredRules, permission.PermissRule)
		}
	}

	log.Println("permisIdList:", permisIdList)
	log.Println("filteredRules:", filteredRules)
	return filteredRules
}

// 递归过滤值为空的字段
func filterEmptyFields(permissList []admin_model.PermissionUser) []admin_model.PermissionUser {
	var filteredList []admin_model.PermissionUser
	for _, perm := range permissList {
		perm.Children = filterEmptyFields(perm.Children)
		filteredPerm := filterEmptyFieldsFromStruct(perm)
		// 只过滤类型为BUTTON的菜单，保留其他所有菜单（包括主目录菜单）
		if filteredPerm.Type != "BUTTON" {
			// 如果是主目录菜单（没有子菜单），确保它被保留
			if len(perm.Children) == 0 && filteredPerm.ParentId == 0 {
				filteredList = append(filteredList, filteredPerm)
			} else {
				filteredList = append(filteredList, filteredPerm)
			}
		}
	}
	return filteredList
}
func filterEmptyFieldsmenu(permissList []admin_model.PermissionUser) []admin_model.PermissionUser {
	var filteredList []admin_model.PermissionUser
	for _, perm := range permissList {
		perm.Children = filterEmptyFieldsmenu(perm.Children)
		filteredPerm := perm
		//不过滤
		filteredList = append(filteredList, filteredPerm)
	}
	return filteredList
}

// 过滤结构体中值为空的字段
// 过滤结构体中值为空的字段
func filterEmptyFieldsFromStruct(perm admin_model.PermissionUser) admin_model.PermissionUser {
	v := reflect.ValueOf(&perm).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := v.Type().Field(i).Name
		if (fieldName == "Rule" || fieldName == "Key") && len(perm.Children) > 0 {
			continue
		} else if isEmptyValue(field) && field.CanSet() {
			field.Set(reflect.Zero(field.Type()))
		}
	}
	return perm
}

// 不过滤结构体中值为空的字段
func filterEmptyFieldsFromStructmenu(perm admin_model.PermissionUser) admin_model.PermissionUser {
	v := reflect.ValueOf(&perm).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := v.Type().Field(i).Name
		if (fieldName == "Rule" || fieldName == "Key") && len(perm.Children) > 0 {
			field.Set(reflect.Zero(field.Type()))
		}
	}
	return perm
}

// 判断值是否为空
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	}
	return false
}

// Logout 退出登录
func (s *TenantsService) Logout(c *gin.Context) error {
	// 获取用户ID
	userId, exists := c.Get("uid")
	if !exists {
		return fmt.Errorf("用户ID不存在")
	}

	// 删除 Redis 中的 token
	err := redis.DeleteToken(strconv.Itoa(userId.(int)))
	if err != nil {
		return fmt.Errorf("删除 token 失败: %v", err)
	}

	return nil

}

// buildUserPermissionTree 构建用户权限树结构
func (s *TenantsService) buildUserPermissionTree(permissions []admin_model.PermissionUser) []admin_model.PermissionUser {
	if len(permissions) == 0 {
		return []admin_model.PermissionUser{}
	}

	// 创建权限映射
	permMap := make(map[int]admin_model.PermissionUser)
	for _, perm := range permissions {
		permMap[perm.ID] = perm
	}

	// 找到根节点并构建树
	var roots []admin_model.PermissionUser
	for _, perm := range permissions {
		if perm.ParentId == 0 {
			// 根节点，构建完整的树结构
			root := perm
			root.Meta = admin_model.Meta{
				Title: perm.Title,
				Icon:  perm.Icon,
			}
			root.Children = s.buildChildrenRecursive(perm.ID, permMap)
			roots = append(roots, root)
		}
	}

	// 🔧 关键修复：对根节点也进行排序
	sort.Slice(roots, func(i, j int) bool {
		// 按sort字段降序排列（数值大的在前）
		// 如果sort字段相同，则按ID升序排列确保稳定排序
		if roots[i].Sort == roots[j].Sort {
			return roots[i].ID < roots[j].ID
		}
		return roots[i].Sort > roots[j].Sort
	})

	return roots
}

// buildChildrenRecursive 递归构建子节点
func (s *TenantsService) buildChildrenRecursive(parentID int, permMap map[int]admin_model.PermissionUser) []admin_model.PermissionUser {
	var children []admin_model.PermissionUser

	for _, perm := range permMap {
		if perm.ParentId == parentID {
			child := perm
			child.Meta = admin_model.Meta{
				Title: perm.Title,
				Icon:  perm.Icon,
			}
			// 递归构建子节点的子节点
			child.Children = s.buildChildrenRecursive(perm.ID, permMap)
			children = append(children, child)
		}
	}

	// 🔧 关键修复：按sort字段排序
	// 使用sort包对children进行排序
	sort.Slice(children, func(i, j int) bool {
		// 按sort字段降序排列（数值大的在前）
		// 如果sort字段相同，则按ID升序排列确保稳定排序
		if children[i].Sort == children[j].Sort {
			return children[i].ID < children[j].ID
		}
		return children[i].Sort > children[j].Sort
	})

	return children
}
