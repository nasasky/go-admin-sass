package inout

type RoleReq struct {
	// 角色名称
	RoleName string `form:"role_name" binding:"required"`
	// 角色描述
	RoleDesc string `form:"role_desc" binding:"required"`
}

// 新增角色
type AddRolexReq struct {
	RoleName  string `form:"role_name" binding:"required"`
	RoleDesc  string `form:"role_desc" binding:"required"`
	Enable    int    `form:"enable"`
	Sort      int    `form:"sort"`
	Pessimism string `form:"pessimism"`
}

// 编辑角色 - 只修改名称和描述
type UpdateRole struct {
	Id       int    `json:"id" binding:"required"`
	RoleName string `json:"role_name" binding:"required"`
	RoleDesc string `json:"role_desc" binding:"required"`
}

// 设置角色权限
type SetRolePermissionReq struct {
	Id         int      `form:"id" binding:"required"`
	Permission []string `form:"permission" binding:"required"`
}

// 列表
type GetRoleListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	RoleName string `form:"role_name"`
}

// 角色列表
type RoleListItem struct {
	// 角色id
	Id int `json:"id"`
	// 角色名称
	RoleName string `json:"role_name"`
	// 角色描述
	RoleDesc string `json:"role_desc"`
	// 创建时间
	CreateTime string `json:"create_time"`
	// 更新时间
	UpdateTime string `json:"update_time"`
	// 是否启用
	Enable int `json:"enable"`
	// 排序
	Sort int `json:"sort"`
	// 创建人ID
	CreatorId int `json:"creator_id"`
	// 创建人用户名
	CreatorName string `json:"creator_name"`
	// 创建人类型
	CreatorType int `json:"creator_type"`
	// 创建人类型描述
	CreatorTypeDesc string `json:"creator_type_desc"`
}

type GetRoleListResp struct {
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Items    []RoleListItem `json:"items"`
}

type DeleteRoleReq struct {
	Id int `form:"id" binding:"required"`
}
