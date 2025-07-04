package admin_service

import (
	"crypto/md5"
	"fmt"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/utils"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 用户类型常量
const (
	UserTypeAdmin   = 1 // 管理员
	UserTypeShop    = 2 // 店铺用户
	UserTypeGeneral = 3 // 普通用户
)

type EmployeeService struct{}

// AddEmployee 添加员工
func (e EmployeeService) AddEmployee(c *gin.Context, params inout.AddEmployeeReq) (int, error) {
	// 获取当前用户ID
	userId := c.GetInt("uid")
	userType := c.GetInt("type")
	//判断userType不等于1或者2的时候，提示无权限添加，如果是2的时候，插入员工的userType为3
	if userType != UserTypeAdmin && userType != UserTypeShop {
		return 0, utils.NewError("无权限添加员工")
	}
	if userType == UserTypeShop {
		params.UserType = UserTypeGeneral // 设置员工类型为普通用户

	}
	// 创建员工对象
	employee := admin_model.Employee{
		Username:   params.Username,
		Phone:      params.Phone,
		UserType:   params.UserType,
		ParentId:   userId,
		Avatar:     params.Avatar,
		Password:   fmt.Sprintf("%x", md5.Sum([]byte(params.Password))),
		Enable:     params.Enable,
		RoleID:     params.RoleId,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	// 插入员工数据
	err := db.Dao.Create(&employee).Error
	if err != nil {
		return 0, err
	}

	return employee.Id, nil
}

// UpdateEmployee 更新员工信息
// UpdateEmployee 更新员工信息
func (e EmployeeService) UpdateEmployee(c *gin.Context, params inout.UpdateEmployeeReq) (int, error) {
	// 获取当前用户ID
	userType := c.GetInt("type")

	// 判断用户类型是否有权限更新员工信息
	if userType != UserTypeAdmin && userType != UserTypeShop {
		return 0, utils.NewError("无权限更新员工信息")
	}

	// 先查询现有员工记录
	var employee admin_model.Employee
	if err := db.Dao.Where("id = ?", params.Id).First(&employee).Error; err != nil {
		return 0, fmt.Errorf("未找到员工记录: %w", err)
	}

	// 创建更新映射，只更新提供的字段
	updates := map[string]interface{}{
		"username":    params.Username,
		"phone":       params.Phone,
		"user_type":   params.UserType,
		"avatar":      params.Avatar,
		"role_id":     params.RoleId,
		"update_time": time.Now(),
		"enable":      params.Enable,
		"sex":         params.Sex,
	}

	// 只更新指定字段
	if err := db.Dao.Model(&employee).Updates(updates).Error; err != nil {
		return 0, err
	}

	return params.Id, nil
}

// DeleteEmployee 删除员工
func (e EmployeeService) DeleteEmployee(c *gin.Context, ids []int) error {
	// 获取当前用户ID
	userId := c.GetInt("uid")
	userType := c.GetInt("type")

	// 判断用户类型是否有权限删除员工
	if userType != UserTypeAdmin && userType != UserTypeShop {
		return utils.NewError("无权限删除员工")
	}

	// 删除员工数据
	var err error
	if userType == UserTypeAdmin {
		// 超管直接删除
		err = db.Dao.Where("id IN (?)", ids).Delete(&admin_model.Employee{}).Error
	} else {
		// 非超管需要判断 parent_id
		err = db.Dao.Where("id IN (?) AND parent_id = ?", ids, userId).Delete(&admin_model.Employee{}).Error
	}

	if err != nil {
		return err
	}

	return nil
}

// GetEmployeeDetail 获取员工详情
func (e EmployeeService) GetEmployeeDetail(c *gin.Context, id int) (interface{}, error) {
	// 获取当前用户ID
	//userId := c.GetInt("uid")
	userType := c.GetInt("type")

	// 判断用户类型是否有权限查看员工详情
	if userType != UserTypeAdmin && userType != UserTypeShop {
		return nil, utils.NewError("无权限查看员工详情")
	}

	// 查询员工数据
	var employee admin_model.Employee
	err := db.Dao.Where("id = ?", id).First(&employee).Error
	if err != nil {
		return nil, err
	}

	return employee, nil
}

// GetEmployeeList 获取员工列表
func (e EmployeeService) GetEmployeeList(c *gin.Context, params inout.ListpageReq) (interface{}, error) {
	// 标准化分页参数
	normalizePagination(&params)

	// 获取当前用户信息
	userType := c.GetInt("type")
	userId := c.GetInt("uid")

	user, err := getUserById(userId)
	if err != nil {
		return nil, err
	}

	// 根据用户类型和角色选择不同的查询策略
	var data []admin_model.Employee
	var total int64

	if isShopOwner(user) {
		// 店铺拥有者 - 查询下属员工
		data, total, err = queryShopEmployees(userId, params)
	} else {
		// 其他用户 - 根据权限查询
		data, total, err = queryByUserType(userType, params)
	}

	if err != nil {
		return nil, err
	}

	// 格式化并返回结果
	return buildEmployeeResponse(data, total, params), nil
}

// 标准化分页参数
func normalizePagination(params *inout.ListpageReq) {
	params.Page = max(params.Page, 1)
	params.PageSize = max(min(params.PageSize, 100), 10) // 限制最大页大小为100
}

// 根据ID获取用户信息
func getUserById(userId int) (admin_model.Employee, error) {
	var user admin_model.Employee
	err := db.Dao.Where("id = ?", userId).First(&user).Error
	return user, err
}

// 判断是否是店铺拥有者
func isShopOwner(user admin_model.Employee) bool {
	return user.UserType == UserTypeShop && user.ParentId == 0
}

// 查询店铺下的员工
func queryShopEmployees(shopOwnerId int, params inout.ListpageReq) ([]admin_model.Employee, int64, error) {
	var data []admin_model.Employee
	var total int64

	// 计算分页偏移量
	offset := (params.Page - 1) * params.PageSize

	// 获取总记录数
	if err := db.Dao.Model(&admin_model.Employee{}).Where("parent_id = ?", shopOwnerId).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询分页数据
	if err := db.Dao.Where("parent_id = ?", shopOwnerId).
		Offset(offset).
		Limit(params.PageSize).
		Find(&data).Error; err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

// 根据用户类型查询员工
func queryByUserType(userType int, params inout.ListpageReq) ([]admin_model.Employee, int64, error) {
	var data []admin_model.Employee
	var total int64

	query := db.Dao.Model(&admin_model.Employee{}).Scopes(
		employeeFilterByUserType(userType),
	)

	// 计算分页偏移量
	offset := (params.Page - 1) * params.PageSize

	// 使用单个查询获取总数和数据
	if err := query.Count(&total).Offset(offset).Limit(params.PageSize).Find(&data).Error; err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

// 根据用户类型构建过滤条件
func employeeFilterByUserType(userType int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if userType == UserTypeAdmin {
			return db // 管理员可查看所有用户
		}
		return db.Where("user_type = ?", userType) // 非管理员只能查看同类型用户
	}
}

// 构建统一的员工列表响应
func buildEmployeeResponse(data []admin_model.Employee, total int64, params inout.ListpageReq) inout.GetEmployeeListResp {
	items := formatEmployeeData(data)
	// 确保Items字段始终是数组而不是null
	if items == nil {
		items = make([]inout.EmployeeItem, 0)
	}

	return inout.GetEmployeeListResp{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Items:    items,
	}
}

// 格式化员工数据
func formatEmployeeData(data []admin_model.Employee) []inout.EmployeeItem {
	result := make([]inout.EmployeeItem, 0) // 初始化为空数组而不是nil
	for _, item := range data {
		result = append(result, inout.EmployeeItem{
			Id:         item.Id,
			UserName:   item.Username,
			Phone:      item.Phone,
			UserType:   item.UserType,
			Enable:     item.Enable,
			RoleId:     item.RoleID,
			Avatar:     item.Avatar,
			Sex:        item.Sex,
			CreateTime: utils.FormatTime2(item.CreateTime),
			UpdateTime: utils.FormatTime2(item.UpdateTime),
		})
	}
	return result
}

// 返回两个数的最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
