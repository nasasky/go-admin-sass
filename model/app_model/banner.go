package app_model

import "time"

type Banner struct {
	Id         int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Title      string    `json:"title"`
	ImageUrl   string    `json:"image_url" gorm:"column:image_url"`
	LinkUrl    string    `json:"link_url" gorm:"column:link_url"`
	Sort       int       `json:"sort"`
	Status     int       `json:"status"` // 0: disabled, 1: enabled
	TenantsId  int       `json:"tenants_id" gorm:"column:tenants_id"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

func (Banner) TableName() string {
	return "pet_banners"
}
