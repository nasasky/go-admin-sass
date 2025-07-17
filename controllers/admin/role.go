package admin

import (
	"fmt"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/services/admin_service"
	"nasa-go-admin/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var roleService = &admin_service.RoleService{}

func GetRoleList(c *gin.Context) {
	var params inout.GetRoleListReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	list, err := roleService.GetRoleList(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	fmt.Print(list)
	Resp.Succ(c, list)
}

//GetAllRoleList

func GetAllRoleList(c *gin.Context) {

	list, err := roleService.GetAllRoleList(c)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	fmt.Print(list)
	Resp.Succ(c, list)
}

// 角色详情
func GetRoleDetail(c *gin.Context) {
	var params inout.GetRoleDetailReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	fmt.Println(params)
	role, err := roleService.GetRoleDetail(c, params.Id)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, role)
}

func AddRole(c *gin.Context) {
	var params inout.AddRolexReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// 设置默认值
	enable := params.Enable
	if enable == 0 {
		enable = 1
	}

	sort := params.Sort
	if sort == 0 {
		sort = 0
	}
	userId := c.GetInt("uid")
	userType := c.GetInt("type")
	var Code string
	if userType == 2 {
		Code = "STORE"
	} else {
		Code = "SUPER_ADMIN"
	}

	// 解析 Pessimism 字段
	pessimism, err := utils.ParsePessimism(c, params.Pessimism)
	if err != nil {
		Resp.Err(c, 20001, "Invalid pessimism value")
		return
	}

	// 创建 role 结构体并设置默认值
	role := admin_model.AddRole{
		RoleName:   params.RoleName,
		RoleDesc:   params.RoleDesc,
		Enable:     enable,
		Sort:       sort,
		UserId:     userId,
		UserType:   userType,
		Code:       Code,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	err = roleService.AddRole(c, role, pessimism)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, nil)
}

// 更新角色 - 只修改名称和描述
func UpdateRole(c *gin.Context) {
	var params inout.UpdateRole
	if err := c.ShouldBindJSON(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// 验证必要字段
	if params.RoleName == "" {
		Resp.Err(c, 20001, "角色名称不能为空")
		return
	}

	if params.RoleDesc == "" {
		Resp.Err(c, 20001, "角色描述不能为空")
		return
	}

	// 创建 role 结构体，只包含需要更新的字段
	role := admin_model.UpdateRole{
		Id:       params.Id,
		RoleName: params.RoleName,
		RoleDesc: params.RoleDesc,
	}

	// 传递空的权限数组，因为不修改权限
	err := roleService.UpdateRole(c, role, []int{})
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, nil)
}

// SetRolePermission 设置角色权限
func SetRolePermission(c *gin.Context) {
	var params inout.SetRolePermissionReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	err := roleService.SetRolePermission(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, nil)
}

// DeleteRole 删除角色
func DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		Resp.Err(c, 20001, "Invalid ID")
		return
	}
	err = roleService.DeleteRole(c, id)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, nil)
}
