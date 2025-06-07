package app_model

import "time"

type AppProfile struct {
	ID         int       `json:"id"`
	Avatar     string    `json:"avatar"`
	Address    string    `json:"address"`
	Email      string    `json:"email"`
	NickName   string    `gorm:"column:nickName"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
	Phone      string    `json:"phone"`
	Openid     string    `json:"openid"`
	UnionID    string    `json:"unionid"`
	UserName   string    `json:"user_name"`
}

func (AppProfile) TableName() string {
	return "app_user"
}
