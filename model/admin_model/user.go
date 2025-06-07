package admin_model

import "time"

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
	ID             int       `json:"id"`
	Token          string    `json:"token"`
	Username       string    `json:"username"`
	Password       string    `json:"password"`
	PasswordBcrypt string    `json:"-" gorm:"column:password_bcrypt"` // 新的bcrypt密码字段
	Phone          string    `json:"phone"`
	RoleId         int       `json:"role_id" `
	UserType       int       `json:"user_type"`
	CreateTime     time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime     time.Time `json:"update_time" gorm:"column:update_time"`
	ParentId       int       `json:"parent_id"`
}

type InsertUser struct {
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	Phone      string    `json:"phone"`
	RoleId     int       `json:"role_id" `
	UserType   int       `json:"user_type"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

func (AdminUser) TableName() string {
	return "user"
}
func (InsertUser) TableName() string {
	return "user"

}

func (AdminUserReq) TableName() string {
	return "user"
}
