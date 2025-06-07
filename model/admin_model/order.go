package admin_model

import (
	"time"

	"github.com/shopspring/decimal"
)

type OrderList struct {
	Id int `json:"id"`
	// 订单号
	No string `json:"no"`
	// 商品名称
	GoodsId int `json:"goods_id"`

	TenantsId int `json:"tenants_id"` // 商家ID
	// 添加用户关联
	Status   string `json:"status"`
	CouponId int    `json:"coupon_id"`
	// 订单金额
	Amount     decimal.Decimal `json:"amount"`
	Num        int             `json:"num"`
	UserId     int             `json:"user_id"`
	CreateTime time.Time       `json:"create_time"`
	UpdateTime time.Time       `json:"update_time"`
}

type AppUser struct {
	Id       int    `json:"id"`
	UserName string `json:"user_name"`
	Avatar   string `json:"avatar"`
	Phone    string `json:"phone"`
}

func (OrderList) TableName() string {
	return "order"
}

func (AppUser) TableName() string {
	return "app_user"
}
