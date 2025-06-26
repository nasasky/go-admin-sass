package inout

import (
	"time"
)

type AddTenantsReq struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
	Phone    string `form:"phone" binding:"required"`
	Type     int    `form:"type" binding:"required"`
	RoleId   int    `form:"role_id" binding:"required"`
}

type UpdateTenantsReq struct {
	Id       int    `form:"id" binding:"required"`
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
	Phone    string `form:"phone" binding:"required"`
	Type     int    `form:"type" binding:"required"`
	RoleId   int    `form:"role_id" binding:"required"`
}

type LoginAdminReq struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
	Captcha  string `form:"captcha"`
}

type LoginTenantsReq struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

type AddMenuReq struct {
	ParentId  int    `form:"parent_id"`
	Label     string `form:"label"`
	Icon      string `form:"icon"`
	Rule      string `form:"rule" `
	Key       string `form:"key" `
	Path      string `form:"path" binding:"required"`
	Type      string `form:"type" binding:"required"`
	Show      int    `form:"show"`
	Sort      int    `form:"sort"`
	Title     string `form:"title"`
	Layout    string `form:"layout"`
	KeepAlive int    `form:"keepAlive"`
}

type AddArticleReq struct {
	Title   string `form:"title" binding:"required"`
	Content string `form:"content" binding:"required"`
	Type    int    `form:"type" binding:"required"`
}

type UpdateArticleReq struct {
	Id       int    `form:"id" binding:"required"`
	Title    string `form:"title" binding:"required"`
	Content  string `form:"content" binding:"required"`
	Type     int    `form:"type" binding:"required"`
	Status   int    `form:"status"`
	Tips     string `form:"tips"`
	Isdelete int    `form:"isdelete"`
}

type GetArticleListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Seach    string `form:"seach"`
}

type MarketingItem struct {
	Id         int    `json:"id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Type       int    `json:"type"`
	UserID     int    `json:"user_id"`
	Status     int    `json:"status"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
}

type MarketingListResponse struct {
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	Items    []MarketingItem `json:"items"`
}

type GetArticleDetailReq struct {
	Id int `form:"id" binding:"required"`
}

type UpdateMenuReq struct {
	Id       int    `form:"id" binding:"required"`
	ParentId int64  `form:"parent_id"`
	Label    string `form:"label" binding:"required"`
	Icon     string `form:"icon"`
	Rule     string `form:"rule" `
	Path     string `form:"path"`
	Title    string `form:"title"`
	Key      string `form:"key" `
	Type     string `form:"type" binding:"required"`
	Show     int    `form:"show"`
	Sort     int    `form:"sort"`
}

type SettingReq struct {
	Appid  string `form:"appid" binding:"required"`
	Secret string `form:"secret" binding:"required"`
	Name   string `form:"name" binding:"required"`
	Tips   string `form:"tips"`
	Type   string `form:"type" binding:"required"`
}

type UpdateSettingReq struct {
	Id     int    `form:"id" binding:"required"`
	Appid  string `form:"appid" binding:"required"`
	Secret string `form:"secret" binding:"required"`
	Name   string `form:"name" binding:"required"`
	Tips   string `form:"tips"`
	Type   string `form:"type"`
}

type FeishuSendReq struct {
	Content   string `form:"content" binding:"required"`
	MsgType   string `form:"msg_type" binding:"required"`
	ReceiveId string `form:"receive_id" binding:"required"`
}

type AddEmployeeReq struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
	Phone    string `form:"phone" binding:"required"`
	RoleId   int    `form:"role_id" binding:"required"`
	Enable   string `form:"enable"`
	Sex      int    `form:"sex"`
	Avatar   string `form:"avatar"`
	UserType int    `form:"user_type"`
}

type UpdateEmployeeReq struct {
	Id       int    `json:"id" binding:"required"`
	Username string `json:"username" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	RoleId   int    `json:"role_id" binding:"required"`
	Enable   string `json:"enable"`
	UserType int    `json:"user_type"`
	Sex      int    `json:"sex"`
	Avatar   string `json:"avatar"`
}

type AddEmployeeGroupReq struct {
	Name  string `form:"name" binding:"required"`
	Rules string `form:"rules" binding:"required"`
}

type UpdateEmployeeGroupReq struct {
	Id    int    `form:"id" binding:"required"`
	Name  string `form:"name" binding:"required"`
	Rules string `form:"rules" binding:"required"`
}

type ListpageReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Search   string `form:"search"`
}
type DicteReq struct {
	Id       int    `form:"id" binding:"required"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Code     string `form:"code"`
}
type AddDictTypeReq struct {
	Type     string `form:"type" binding:"required"`
	TypeName string `form:"type_name" binding:"required"`
	TypeCode string `form:"type_code" binding:"required"`
	Remark   string `form:"remark"`
	IsLock   string `form:"is_lock"`
	DelFlag  string `form:"del_flag"`
	IsShow   string `form:"is_show"`
}
type AddDictValueReq struct {
	Id                int    `form:"id"`
	CodeName          string `form:"code_name" binding:"required"`
	Code              string `form:"code" binding:"required"`
	Alias             string `form:"alias"`
	CallbackShowStyle string `form:"callback_show_style"`
	Remark            string `form:"remark"`
	Sort              int    `form:"sort"`
	IsDefault         int    `form:"is_default"`
	IsLock            string `form:"is_lock"`
	IsShow            string `form:"is_show"`
	DelFlag           string `form:"del_flag"`
	SysDictTypeId     int    `form:"sys_dict_type_id"`
}

type GetEmployeeListResp struct {
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Items    []EmployeeItem `json:"items"`
}
type EmployeeItem struct {
	Id         int    `json:"id"`
	UserName   string `json:"username"`
	Phone      string `json:"phone"`
	RoleId     int    `json:"role_id"`
	UserType   int    `json:"user_type"`
	Enable     string `json:"enable"`
	Avatar     string `json:"avatar"`
	Sex        int    `json:"sex"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
}

type AddBannerReq struct {
	Title    string `form:"title" binding:"required"`
	ImageUrl string `form:"image_url" binding:"required"`
	LinkUrl  string `form:"link_url"`
	Sort     int    `form:"sort"`
	Status   int    `form:"status"`
}

type UpdateBannerReq struct {
	Id       int    `form:"id" binding:"required"`
	Title    string `form:"title"`
	ImageUrl string `form:"image_url"`
	LinkUrl  string `form:"link_url"`
	Sort     int    `form:"sort"`
	Status   int    `form:"status"`
}

type GetBannerListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Search   string `form:"search"`
}

type BannerItem struct {
	Id         int       `json:"id"`
	Title      string    `json:"title"`
	ImageUrl   string    `json:"image_url"`
	LinkUrl    string    `json:"link_url"`
	Sort       int       `json:"sort"`
	Status     int       `json:"status"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

type BannerListResponse struct {
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
	Items    []BannerItem `json:"items"`
}

type SystemInfoReq struct {
	SystemName  string `form:"system_name" binding:"required" json:"system_name"`   // 系统名称
	SystemTitle string `form:"system_title" binding:"required" json:"system_title"` // 系统标题
	IcpNumber   string `form:"icp_number" json:"icp_number"`                        // 备案号
	Copyright   string `form:"copyright" json:"copyright"`                          // 版权信息
	Status      int    `form:"status" json:"status"`                                // 状态：1-启用 0-禁用
}

type UpdateSystemInfoReq struct {
	Id          int    `form:"id" binding:"required" json:"id"`
	SystemName  string `form:"system_name" json:"system_name"`   // 系统名称
	SystemTitle string `form:"system_title" json:"system_title"` // 系统标题
	IcpNumber   string `form:"icp_number" json:"icp_number"`     // 备案号
	Copyright   string `form:"copyright" json:"copyright"`       // 版权信息
	Status      int    `form:"status" json:"status"`             // 状态：1-启用 0-禁用
}

type SystemInfoResponse struct {
	Id          int       `json:"id"`
	SystemName  string    `json:"system_name"`  // 系统名称
	SystemTitle string    `json:"system_title"` // 系统标题
	IcpNumber   string    `json:"icp_number"`   // 备案号
	Copyright   string    `json:"copyright"`    // 版权信息
	Status      int       `json:"status"`       // 状态：1-启用 0-禁用
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}

type GetSystemInfoListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Search   string `form:"search"`
	OrderBy  string `form:"order_by"`
}

type SystemInfoListResponse struct {
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
	Items    []SystemInfoResponse `json:"items"`
}
