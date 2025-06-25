package app_model

import (
	"time"
)

type UserApp struct {
	Token      string    `json:"token"`
	ID         int       `json:"id"`
	Username   string    `json:"username"`
	Openid     string    `json:"openid"`
	UnionID    string    `json:"unionid"`
	Enable     bool      `json:"enable"`
	Phone      string    `json:"phone"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
	Avatar     string    `json:"avatar" gorm:"column:avatar"`
	Nickname   string    `json:"nickname" gorm:"column:nickname"`
	Sex        int       `json:"sex" gorm:"column:column:sex"`
}
type LoginUserApp struct {
	ID         int       `json:"id"`
	Token      string    `json:"token"`
	Openid     string    `json:"openid"`
	UnionID    string    `json:"unionid"`
	Username   string    `json:"username"`
	Password   string    `json:"passworßd"`
	Enable     bool      `json:"enable"`
	Phone      string    `json:"phone"`
	Avatar     string    `json:"avatar"`
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
func (LoginUserApp) TableName() string {
	return "app_user"

}

func (LoginUser) TableName() string {
	return "app_user"

}
