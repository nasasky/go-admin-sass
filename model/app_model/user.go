package app_model

import (
	"time"
)

type UserApp struct {
	Token      string    `json:"token"`
	ID         int       `json:"id"`
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	Enable     bool      `json:"enable"`
	Phone      string    `json:"phone"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

type LoginUser struct {
	ID         int       `json:"id"`
	Phone      string    `json:"phone"`
	Password   string    `json:"password"`
	Username   string    `json:"username"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

func (UserApp) TableName() string {
	return "app_user"
}

func (LoginUser) TableName() string {
	return "app_user"

}
