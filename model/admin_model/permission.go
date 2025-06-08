package admin_model

import (
	"time"
)

type RolePermissionsPermission struct {
	RoleId       int `gorm:"column:roleId"`
	PermissionId int `gorm:"column:permissionId"`
}

type Permission struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	Code        string       `json:"code"`
	Type        string       `json:"type"`
	ParentId    *int         `json:"parentId" gorm:"column:parentId"`
	Path        string       `json:"path"`
	Redirect    string       `json:"redirect"`
	Icon        string       `json:"icon"`
	Component   string       `json:"component"`
	Layout      string       `json:"layout"`
	KeepAlive   int          `json:"keepAlive" gorm:"column:keepAlive"`
	Method      string       `json:"method"`
	Description string       `json:"description"`
	Show        int          `json:"show"`
	Enable      int          `json:"enable"`
	Order       int          `json:"order"`
	InLink      int          `json:"is_link"`
	Children    []Permission `json:"children" gorm:"-"`
}
type PermissionUser struct {
	ID        int              `json:"id"`
	Label     string           `json:"label"`
	Title     string           `json:"title"`
	Rule      string           `json:"rule"`
	Key       string           `json:"key"`
	Icon      string           `json:"icon"`
	Path      string           `json:"path"`
	ParentId  int              `json:"parent_id" gorm:"column:parent_id"`
	Type      string           `json:"type"`
	Layout    string           `json:"layout"`
	Show      int              `json:"show"`
	Enable    int              `json:"enable"`
	Order     int              `json:"order"`
	Sort      int              `json:"sort"`
	KeepAlive int              `json:"keepAlive" gorm:"column:keepAlive"`
	IsLink    int              `json:"is_link"`
	Children  []PermissionUser `json:"children" gorm:"-"`
	//是个对象
	Meta Meta `json:"meta" gorm:"-"`
}

type Meta struct {
	Title string `json:"title"`
	Icon  string `json:"icon"`
}

type StaffPermissions struct {
	ID           int `json:"id"`
	PermissionId int `json:"permission_id" gorm:"column:permission_id"`
	TenantId     int `json:"tenant_id" gorm:"column:tenant_id"`
	RoleId       int `json:"role_id" gorm:"column:role_id"`
}

type PermissionMenu struct {
	Id         int       `json:"id"`
	Label      string    `json:"label"`
	Icon       string    `json:"icon"`
	Rule       string    `json:"rule"`
	Key        string    `json:"key"`
	Path       string    `json:"path"`
	Layout     string    `json:"layout"`
	Show       int       `json:"show"`
	Enable     int       `json:"enable"`
	Order      int       `json:"order"`
	Title      string    `json:"title"`
	Type       string    `json:"type"`
	IsLink     int       `json:"is_link"`
	KeepAlive  int       `json:"keepAlive" gorm:"column:keepAlive"`
	ParentId   *int64    `json:"parent_id" gorm:"column:parent_id"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
	Sort       int       `json:"sort"`
}

func (Permission) TableName() string {
	return "permission"
}

func (RolePermissionsPermission) TableName() string {
	return "role_permissions_permission"
}

func (PermissionUser) TableName() string {
	return "permission_user"
}

func (StaffPermissions) TableName() string {
	return "staff_permissions"
}
func (PermissionMenu) TableName() string {
	return "permission_user"
}
