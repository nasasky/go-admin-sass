package admin_model

import "time"

type Goods struct {
	Id         int       `json:"id"`
	UserId     int       `json:"user_id" gorm:"column:user_id"`
	GoodsName  string    `json:"goods_name" gorm:"column:goods_name"`
	Price      float64   `json:"price"`
	Content    string    `json:"content"`
	Cover      string    `json:"cover"`
	Status     string    `json:"status"`
	CategoryId int       `json:"category_id" gorm:"column:category_id"`
	TenantsId  int       `json:"tenants_id" gorm:"column:tenants_id"`
	Stock      int       `json:"stock"`
	Isdelete   int       `json:"isdelete"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

type GoodsCategory struct {
	Id          int       `json:"id"`
	UserId      int       `json:"user_id" gorm:"column:user_id"`
	Name        string    `json:"name" gorm:"column:name"`
	Status      string    `json:"status" gorm:"column:status"`
	TenantsId   int       `json:"tenants_id" gorm:"column:tenants_id"`
	Description string    `json:"description" gorm:"column:description"`
	CreateTime  time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime  time.Time `json:"update_time" gorm:"column:update_time"`
}

func (Goods) TableName() string {
	return "goods_list"
}

func (GoodsCategory) TableName() string {
	return "goods_category"
}
