package admin_model

import "time"

type TenantsReq struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
	Phone    string `form:"phone" binding:"required"`
}

type TenantsUser struct {
	ID         int       `json:"id"`
	Username   string    `json:"username"`
	TenantName string    `json:"tenant_name"`
	RoleId     int       `json:"role_id"`
	Password   string    `json:"password"`
	Phone      string    `json:"phone"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

//user

type AdminUserReq struct {
	Username   string    `form:"username" binding:"required"`
	Password   string    `form:"password" binding:"required"`
	Phone      string    `form:"phone" binding:"required"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
	TenantId   int       `json:"tenant_id"`
}

type AdminUser struct {
	ID         int       `json:"id"`
	Token      string    `json:"token"`
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	Phone      string    `json:"phone"`
	RoleId     int       `json:"role_id"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
	TenantId   int       `json:"tenant_id"`
}

func (TenantsReq) TableName() string {
	return "tenants"

}

func (TenantsUser) TableName() string {
	return "tenants"
}

func (AdminUser) TableName() string {
	return "user"
}

func (AdminUserReq) TableName() string {
	return "user"
}
