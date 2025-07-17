package admin_model

import "time"

type Role struct {
	Id         int    `json:"id"`
	RoleName   string `json:"role_name" gorm:"column:role_name"`
	RoleDesc   string `json:"role_desc" gorm:"column:role_desc"`
	UserId     int    `json:"user_id" gorm:"column:user_id"`
	UserType   int    `json:"user_type" gorm:"column:user_type"`
	Enable     int    `json:"enable"`
	Sort       int    `json:"sort"`
	CreateTime string `json:"create_time" gorm:"column:create_time"`
	UpdateTime string `json:"update_time" gorm:"column:update_time"`
}

type AddRole struct {
	RoleName   string    `json:"role_name" binding:"required"`
	RoleDesc   string    `json:"role_desc" binding:"required"`
	UserId     int       `json:"user_id"`
	Enable     int       `form:"enable"`
	UserType   int       `json:"user_type"`
	Code       string    `json:"code"`
	Sort       int       `json:"sort"`
	Id         int       `json:"id"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

type UpdateRole struct {
	Id         int       `json:"id" binding:"required"`
	RoleName   string    `json:"role_name" binding:"required"`
	RoleDesc   string    `json:"role_desc" binding:"required"`
	UserId     int       `json:"user_id"`
	UserType   int       `json:"user_type"`
	Enable     int       `json:"enable"`
	Sort       int       `json:"sort"`
	Code       string    `json:"code"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

func (Role) TableName() string {
	return "role"
}

func (AddRole) TableName() string {
	return "role"
}

func (UpdateRole) TableName() string {
	return "role"
}
