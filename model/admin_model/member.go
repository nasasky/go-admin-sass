package admin_model

import "time"

type Member struct {
	Id         int       `json:"id"`
	UserName   string    `json:"user_name" gorm:"column:username"`
	Avatar     string    `json:"avatar"`
	NickName   string    `json:"nick_name"`
	Phone      string    `json:"phone"`
	Address    string    `json:"address"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

func (Member) TableName() string {
	return "app_user"
}
