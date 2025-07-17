package admin_service

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/pkg/jwt"
	"nasa-go-admin/pkg/monitoring"
	"nasa-go-admin/redis"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"time"

	"nasa-go-admin/pkg/security"

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
func (s *TenantsService) Login(c *gin.Context, username, password string) (map[string]interface{}, error) {
	// 添加性能监控
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if duration > time.Second {
			log.Printf("Login slow operation detected: %v", duration)
		}
	}()

	// 1. 获取并验证请求参数
	var params inout.LoginAdminReq
	if err := c.ShouldBind(&params); err != nil {
		return nil, fmt.Errorf("参数错误: %v", err)
	}

	// 2. 检查验证码开关配置
	captchaEnabled := s.IsCaptchaEnabled()

	// 3. 验证码校验（仅在启用时）
	if captchaEnabled {
		// 使用Redis管道操作
		pipe := redis.GetClient().Pipeline()
		captchaKey := "latest_captcha"
		captchaCmd := pipe.Get(context.Background(), captchaKey)
		pipe.Del(context.Background(), captchaKey)
		_, err := pipe.Exec(context.Background())
		if err != nil {
			return nil, fmt.Errorf("请先获取验证码")
		}

		captchaJSON := captchaCmd.Val()
		var captchaData map[string]interface{}
		if err := json.Unmarshal([]byte(captchaJSON), &captchaData); err != nil {
			return nil, fmt.Errorf("验证码数据错误")
		}

		if params.Captcha == "" {
			return nil, fmt.Errorf("请输入验证码")
		}
		if params.Captcha != captchaData["code"].(string) {
			return nil, fmt.Errorf("验证码错误")
		}

		expireTime := int64(captchaData["expire"].(float64))
		if time.Now().Unix() > expireTime {
			return nil, fmt.Errorf("验证码已过期，请重新获取")
		}
	}

	// 3. 使用缓存查询用户
	userCacheKey := fmt.Sprintf("user:login:%s", username)
	var user admin_model.AdminUser

	// 尝试从缓存获取用户信息
	var userFromCache admin_model.AdminUser
	userJSON, err := redis.GetClient().Get(context.Background(), userCacheKey).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(userJSON), &userFromCache); err == nil {
			// 缓存命中，但需要从数据库获取完整的用户信息（包括密码）
			err = db.Dao.Where("id = ?", userFromCache.ID).First(&user).Error
			if err != nil {
				return nil, fmt.Errorf("用户不存在")
			}
		} else {
			// 缓存数据损坏，从数据库查询
			err = db.Dao.Where("username = ?", username).First(&user).Error
			if err == gorm.ErrRecordNotFound {
				// 如果用户名未找到，尝试通过手机号查询
				err = db.Dao.Where("phone = ?", username).First(&user).Error
				if err != nil {
					return nil, fmt.Errorf("用户不存在")
				}
			} else if err != nil {
				return nil, err
			}
		}
	} else {
		// 缓存未命中，从数据库查询
		err = db.Dao.Where("username = ?", username).First(&user).Error
		if err == gorm.ErrRecordNotFound {
			// 如果用户名未找到，尝试通过手机号查询
			err = db.Dao.Where("phone = ?", username).First(&user).Error
			if err != nil {
				return nil, fmt.Errorf("用户不存在")
			}
		} else if err != nil {
			return nil, err
		}
	}

	// 将用户信息缓存，有效期1小时
	// 注意：不要缓存密码字段，避免密码修改后缓存不一致的问题
	cacheUser := admin_model.AdminUser{
		ID:         user.ID,
		Username:   user.Username,
		Phone:      user.Phone,
		RoleId:     user.RoleId,
		UserType:   user.UserType,
		CreateTime: user.CreateTime,
		UpdateTime: user.UpdateTime,
		ParentId:   user.ParentId,
		// 不缓存 Password 和 PasswordBcrypt 字段
	}

	if userJSON, err := json.Marshal(cacheUser); err == nil {
		redis.GetClient().Set(context.Background(), userCacheKey, userJSON, time.Hour)
	}
	// 4. 验证密码
	var passwordValid bool
	if user.PasswordBcrypt != "" {
		passwordValid = security.CheckPasswordHash(password, user.PasswordBcrypt)
	} else {
		hashedPassword := fmt.Sprintf("%x", md5.Sum([]byte(password)))
		passwordValid = (user.Password == hashedPassword)

		// 异步升级密码
		if passwordValid {
			go func(userId int, pwd string) {
				if newHash, err := security.HashPassword(pwd); err == nil {
					db.Dao.Model(&admin_model.AdminUser{}).Where("id = ?", userId).Update("password_bcrypt", newHash)
				}
			}(user.ID, password)
		}
	}

	if !passwordValid {
		return nil, fmt.Errorf("密码错误")
	}

	// 5. 生成Token并缓存用户信息
	// 使用安全的 JWT 管理器生成 Token（支持黑名单）
	jwtManager := jwt.NewSecureJWTManager()
	token, err := jwtManager.GenerateToken(user.ID, user.RoleId, user.UserType)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}
	user.Token = token

	// 6. 使用Redis管道操作优化缓存性能
	expiration := time.Hour * 24
	pipe := redis.GetClient().Pipeline()

	// 存储Token和用户信息
	tokenKey := fmt.Sprintf("token:%d", user.ID)
	pipe.Set(context.Background(), tokenKey, token, expiration)

	userInfo := map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"phone":    user.Phone,
		"roleId":   user.RoleId,
		"userType": user.UserType,
		"token":    token,
		"parentId": user.ParentId,
	}

	// Store user info using Redis HMSET
	if err := redis.StoreUserInfo(strconv.Itoa(user.ID), userInfo, expiration); err != nil {
		log.Printf("Failed to store user info in Redis: %v", err)
		// Continue execution, don't let cache failure block login
	}

	// 执行管道操作
	if _, err := pipe.Exec(context.Background()); err != nil {
		log.Printf("Redis缓存Token失败: %v", err)
		// 继续执行，不要因为缓存失败影响登录
	}

	// 7. 获取权限列表
	var permissions []string
	permissionsCacheKey := fmt.Sprintf("permissions:%d", user.RoleId)

	// 尝试从缓存获取权限列表
	if permissionsJSON, err := redis.GetClient().Get(context.Background(), permissionsCacheKey).Result(); err == nil {
		if err := json.Unmarshal([]byte(permissionsJSON), &permissions); err == nil {
			log.Printf("Successfully retrieved permissions from cache for role %d", user.RoleId)
			goto ReturnResponse
		} else {
			log.Printf("Failed to unmarshal cached permissions for role %d: %v", user.RoleId, err)
		}
	} else {
		log.Printf("No cached permissions found for role %d: %v", user.RoleId, err)
	}

	// 如果缓存未命中或解析失败，从数据库获取权限列表
	permissions = getPermissionListByRoleId(user.RoleId)

	// 异步缓存权限列表
	go func(roleId int, perms []string) {
		if permissionsJSON, err := json.Marshal(perms); err == nil {
			if err := redis.GetClient().Set(context.Background(), fmt.Sprintf("permissions:%d", roleId), permissionsJSON, time.Hour).Err(); err != nil {
				log.Printf("Failed to cache permissions for role %d: %v", roleId, err)
			} else {
				log.Printf("Successfully cached permissions for role %d", roleId)
			}
		} else {
			log.Printf("Failed to marshal permissions for role %d: %v", roleId, err)
		}
	}(user.RoleId, permissions)

	// 8. 异步记录登录指标
	go func() {
		monitoring.RecordUserLogin()
		monitoring.SaveBusinessMetric("user_login", user.Username)
	}()

ReturnResponse:
	// 9. 返回响应数据
	responseData := map[string]interface{}{
		"user":        user,
		"permissions": permissions,
		"token":       token,
	}

	// Log the response data for debugging
	log.Printf("Login successful for user %s (ID: %d, Role: %d). Found %d permissions.",
		user.Username, user.ID, user.RoleId, len(permissions))

	return responseData, nil
}

// GetUserInfo
func (s *TenantsService) GetUserInfo(c *gin.Context, id int) (map[string]interface{}, error) {
	var user admin_model.AdminUser
	log.Println("User ID:", id)

	// Improved error handling for user lookup
	err := db.Dao.Where("id = ?", id).First(&user).Error
	if err != nil {
		log.Printf("Failed to fetch user info for ID %d: %v", id, err)
		return nil, fmt.Errorf("获取用户信息失败: %v", err)
	}

	// Generate token with error handling
	// 使用安全的 JWT 管理器生成 Token（支持黑名单）
	jwtManager := jwt.NewSecureJWTManager()
	token, err := jwtManager.GenerateToken(user.ID, user.RoleId, user.UserType)
	if err != nil {
		log.Printf("Failed to generate token for user %d: %v", id, err)
		return nil, fmt.Errorf("生成令牌失败: %v", err)
	}
	user.Token = token

	// Store token with error handling
	expiration := time.Hour * 24
	err = redis.StoreToken(strconv.Itoa(user.ID), user.Token, expiration)
	if err != nil {
		log.Printf("Failed to store token for user %d: %v", id, err)
		return nil, fmt.Errorf("存储Token失败: %v", err)
	}

	log.Printf("Fetching permissions for user %d with role %d", user.ID, user.RoleId)
	permissions := getPermissionListByRoleId(user.RoleId)

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
	// 添加性能监控
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if duration > 500*time.Millisecond {
			log.Printf("Slow permission query detected: %v", duration)
		}
	}()

	type Permission struct {
		Rule        string
		PermissRule string
	}
	var permissions []Permission

	if roleId != 0 {
		// 使用LEFT JOIN查询优化性能，并添加错误处理
		err := db.Dao.Table("role_permissions_permission").
			Select("DISTINCT permission_user.rule, permission_user.permiss_rule").
			Joins("LEFT JOIN permission_user ON role_permissions_permission.permissionId = permission_user.id").
			Where("role_permissions_permission.roleId = ? AND permission_user.rule IS NOT NULL", roleId).
			Find(&permissions).Error

		if err != nil {
			log.Printf("Error fetching permissions for role %d: %v", roleId, err)
			return []string{} // Return empty slice instead of nil
		}

		// 如果没有找到权限，尝试从staff_permissions表查询
		if len(permissions) == 0 {
			log.Printf("No permissions found in role_permissions_permission, trying staff_permissions for role %d", roleId)
			err = db.Dao.Table("staff_permissions").
				Select("DISTINCT permission_user.rule, permission_user.permiss_rule").
				Joins("LEFT JOIN permission_user ON staff_permissions.permission_id = permission_user.id").
				Where("staff_permissions.role_id = ? AND permission_user.rule IS NOT NULL", roleId).
				Find(&permissions).Error

			if err != nil {
				log.Printf("Error fetching permissions from staff_permissions for role %d: %v", roleId, err)
				return []string{} // Return empty slice instead of nil
			}
		}
	} else {
		// 如果roleId为0，直接查询所有权限
		err := db.Dao.Model(&admin_model.PermissionUser{}).
			Select("DISTINCT rule, permiss_rule").
			Where("rule IS NOT NULL").
			Find(&permissions).Error

		if err != nil {
			log.Printf("Error fetching all permissions: %v", err)
			return []string{} // Return empty slice instead of nil
		}
	}

	// 使用预分配内存的切片来存储规则
	totalRules := len(permissions) * 2 // 每个权限最多有两个规则
	filteredRules := make([]string, 0, totalRules)

	// 过滤并添加有效的规则
	for _, permission := range permissions {
		if permission.Rule != "" {
			filteredRules = append(filteredRules, permission.Rule)
		}
		if permission.PermissRule != "" {
			filteredRules = append(filteredRules, permission.PermissRule)
		}
	}

	// Log if no permissions found
	if len(filteredRules) == 0 {
		log.Printf("No permissions found for role %d", roleId)
	} else {
		log.Printf("Found %d permissions for role %d: %v", len(filteredRules), roleId, filteredRules)
	}

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

	// 撤销当前 Token（加入黑名单）
	if currentToken := c.GetHeader("Authorization"); currentToken != "" {
		if len(currentToken) > 7 && currentToken[:7] == "Bearer " {
			currentToken = currentToken[7:]
		}

		// 使用安全的 JWT 管理器撤销 Token
		jwtManager := jwt.NewSecureJWTManager()
		if err := jwtManager.RevokeToken(currentToken); err != nil {
			log.Printf("撤销当前 Token 失败: %v", err)
		}
	}

	// 删除 Redis 中的 token 和用户信息
	userIdStr := strconv.Itoa(userId.(int))
	err := redis.DeleteToken(userIdStr)
	if err != nil {
		log.Printf("删除 Redis token 失败: %v", err)
	}

	// 删除用户信息缓存
	err = redis.DeleteUserInfo(userIdStr)
	if err != nil {
		log.Printf("删除用户信息缓存失败: %v", err)
	}

	// 清除用户登录缓存
	userCacheKey := fmt.Sprintf("user:login:%d", userId.(int))
	redis.GetClient().Del(context.Background(), userCacheKey)

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

// IsCaptchaEnabled 检查验证码是否启用
func (s *TenantsService) IsCaptchaEnabled() bool {
	// 从数据库查询验证码开关配置
	var setting admin_model.SettingList
	err := db.Dao.Where("type = ? AND name = ?", "system", "captcha_enabled").First(&setting).Error
	if err != nil {
		// 如果配置不存在，默认启用验证码
		return true
	}

	// 根据配置值判断是否启用
	return setting.Value == "1" || setting.Value == "true"
}

// UpdateCaptchaStatus 更新验证码开关状态
func (s *TenantsService) UpdateCaptchaStatus(enabled bool) error {
	value := "0"
	if enabled {
		value = "1"
	}

	// 查找现有配置
	var setting admin_model.SettingList
	err := db.Dao.Where("type = ? AND name = ?", "system", "captcha_enabled").First(&setting).Error

	if err != nil {
		// 如果配置不存在，创建新配置
		setting = admin_model.SettingList{
			Type:       "system",
			Name:       "captcha_enabled",
			Value:      value,
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}
		return db.Dao.Create(&setting).Error
	} else {
		// 更新现有配置
		setting.Value = value
		setting.UpdateTime = time.Now()
		return db.Dao.Save(&setting).Error
	}
}

// UpdateUserProfile 更新用户信息
func (s *TenantsService) UpdateUserProfile(c *gin.Context, params inout.UpdateUserProfileReq) error {
	// 先查询用户是否存在
	var user admin_model.AdminUser
	if err := db.Dao.Where("id = ?", params.Id).First(&user).Error; err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if params.Username != "" {
		updates["username"] = params.Username
	}
	if params.Phone != "" {
		updates["phone"] = params.Phone
	}
	if params.Avatar != "" {
		updates["avatar"] = params.Avatar
	}
	if params.Sex >= 0 {
		updates["sex"] = params.Sex
	}
	updates["update_time"] = time.Now()

	// 更新用户信息
	if err := db.Dao.Model(&admin_model.AdminUser{}).Where("id = ?", params.Id).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新用户信息失败: %w", err)
	}

	return nil
}

// UpdateUserPassword 更新用户密码 - 高性能优化版本
func (s *TenantsService) UpdateUserPassword(c *gin.Context, params inout.UpdateUserPasswordReq) error {
	// 1. 快速验证旧密码（单次查询 + 优先bcrypt）
	var user admin_model.AdminUser
	if err := db.Dao.Select("id, username, phone, password, password_bcrypt").Where("id = ?", params.Id).First(&user).Error; err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 2. 快速密码验证（优先bcrypt，避免双重检查）
	var oldPasswordValid bool
	if user.PasswordBcrypt != "" {
		oldPasswordValid = security.CheckPasswordHash(params.OldPassword, user.PasswordBcrypt)
	} else {
		oldPasswordHash := fmt.Sprintf("%x", md5.Sum([]byte(params.OldPassword)))
		oldPasswordValid = (user.Password == oldPasswordHash)
	}

	if !oldPasswordValid {
		return fmt.Errorf("旧密码错误")
	}

	// 3. 并行处理：同时生成MD5和bcrypt哈希
	var wg sync.WaitGroup
	var newPasswordHash string
	var newPasswordBcrypt string
	var hashErr error

	// 启动MD5哈希生成（很快，同步执行）
	newPasswordHash = fmt.Sprintf("%x", md5.Sum([]byte(params.NewPassword)))

	// 启动bcrypt哈希生成（异步，因为比较慢）
	wg.Add(1)
	go func() {
		defer wg.Done()
		hash, err := security.HashPassword(params.NewPassword)
		newPasswordBcrypt = hash
		hashErr = err
	}()

	// 4. 等待bcrypt哈希生成完成
	wg.Wait()
	if hashErr != nil {
		return fmt.Errorf("生成密码哈希失败: %w", hashErr)
	}

	// 5. 使用事务更新密码（确保原子性）
	err := db.Dao.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"password":        newPasswordHash,
			"password_bcrypt": newPasswordBcrypt,
			"update_time":     time.Now(),
		}

		if err := tx.Model(&admin_model.AdminUser{}).Where("id = ?", params.Id).Updates(updates).Error; err != nil {
			return fmt.Errorf("更新密码失败: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	// 6. 异步清理缓存和撤销Token（不阻塞响应）
	go func() {
		// 使用Redis管道批量操作
		pipe := redis.GetClient().Pipeline()

		// 批量删除缓存
		userCacheKey := fmt.Sprintf("user:login:%s", user.Username)
		phoneCacheKey := fmt.Sprintf("user:login:%s", user.Phone)
		pipe.Del(context.Background(), userCacheKey)
		pipe.Del(context.Background(), phoneCacheKey)

		// 执行管道操作
		pipe.Exec(context.Background())

		// 撤销Token（如果存在）
		if currentToken := c.GetHeader("Authorization"); currentToken != "" {
			if len(currentToken) > 7 && currentToken[:7] == "Bearer " {
				currentToken = currentToken[7:]
			}

			jwtManager := jwt.NewSecureJWTManager()
			if err := jwtManager.RevokeToken(currentToken); err != nil {
				log.Printf("撤销Token失败: %v", err)
			}
		}

		// 清理其他缓存
		redis.DeleteUserInfo(strconv.Itoa(params.Id))
		redis.DeleteToken(strconv.Itoa(params.Id))
	}()

	return nil
}
