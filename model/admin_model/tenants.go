package admin_model

import "time"

type TenantsReq struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
	Phone    string `form:"phone" binding:"required"`
}

type StaffReq struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
	Phone    string `form:"phone" binding:"required"`
}

type Tenants struct {
	ID         int       `json:"id"`
	Username   string    `json:"username"`
	Token      string    `json:"token"`
	TenantName string    `json:"tenant_name" gorm:"column:tenant_name"`
	RoleId     int       `json:"role_id" gorm:"column:role_id"`
	Password   string    `json:"password"`
	Phone      string    `json:"phone"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

type StaffUser struct {
	ID         int       `json:"id"`
	Username   string    `json:"username"`
	RoleId     int       `json:"role_id" gorm:"column:role_id"`
	Token      string    `json:"token"`
	Password   string    `json:"password"`
	Phone      string    `json:"phone"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

func (TenantsReq) TableName() string {
	return "tenants"

}

func (Tenants) TableName() string {
	return "tenants"
}

func (StaffReq) TableName() string {
	return "staff"

}

func (StaffUser) TableName() string {
	return "staff"
}
