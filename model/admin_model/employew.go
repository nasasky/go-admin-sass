package admin_model

import "time"

type Employee struct {
	Id         int       `json:"id"`
	Username   string    `json:"username"`
	Phone      string    `json:"phone"`
	UserType   int       `json:"user_type"`
	RoleID     int       `json:"role_id"`
	Enable     string    `json:"enable"`
	Avatar     string    `json:"avatar"`
	Password   string    `json:"password"`
	Sex        int       `json:"sex"`
	ParentId   int       `json:"parent_id"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

func (Employee) TableName() string {
	return "user"
}
