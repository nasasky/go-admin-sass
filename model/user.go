package model

import (
	"time"
)

type User struct {
	ID             int       `json:"id"`
	Username       string    `json:"username"`
	Password       string    `json:"password"`
	PasswordBcrypt string    `json:"-" gorm:"column:password_bcrypt"` // 新的bcrypt密码字段
	Enable         string    `json:"enable"`                          // 改为string类型，匹配数据库中的"active"值
	CreateTime     time.Time `json:"createTime" gorm:"column:createTime"`
	UpdateTime     time.Time `json:"updateTime" gorm:"column:updateTime"`
}

func (User) TableName() string {
	return "user"
}
