package model

import (
	"time"
)

type User struct {
	ID             int       `json:"id"`
	Username       string    `json:"username"`
	Password       string    `json:"password"`
	PasswordBcrypt string    `json:"-" gorm:"column:password_bcrypt"` // 新的bcrypt密码字段
	Enable         bool      `json:"enable"`
	CreateTime     time.Time `json:"createTime" gorm:"column:createTime"`
	UpdateTime     time.Time `json:"updateTime" gorm:"column:updateTime"`
}

func (User) TableName() string {
	return "user"
}
