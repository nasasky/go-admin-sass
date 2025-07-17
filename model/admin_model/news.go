package admin_model

import "time"

type News struct {
	Id          int       `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	CoverImage  string    `json:"cover_image" gorm:"column:cover_image"`
	Sort        int       `json:"sort"`
	Status      int       `json:"status"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}

func (News) TableName() string { return "news" }
