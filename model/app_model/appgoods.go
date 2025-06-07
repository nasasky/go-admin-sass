package app_model

import "time"

type AppGoods struct {
	Id         int       `json:"id"`
	UserId     int       `json:"user_id" gorm:"column:user_id"`
	GoodsName  string    `json:"goods_name" gorm:"column:goods_name"`
	Price      float64   `json:"price"`
	Content    string    `json:"content"`
	Cover      string    `json:"cover"`
	Status     string    `json:"status"`
	TenantsId  int       `json:"tenants_id" gorm:"column:tenants_id"`
	CategoryId int       `json:"category_id" gorm:"column:category_id"`
	Stock      int       `json:"stock"`
	Isdelete   int       `json:"isdelete"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

func (AppGoods) TableName() string {
	return "goods_list"
}
