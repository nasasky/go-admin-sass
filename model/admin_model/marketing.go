package admin_model

import "time"

type Marketing struct {
	Id         int       `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	Type       int       `json:"type"`
	UserId     int       `json:"user_id" gorm:"column:user_id"`
	Status     int       `json:"status"`
	Isdelete   int       `json:"isdelete"`
	Tips       string    `json:"tips"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

func (Marketing) TableName() string {
	return "marketing"
}
