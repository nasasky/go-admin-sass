package admin_model

import "time"

type Revenue struct {
	Id            int       `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantsId     int       `json:"tenants_id" gorm:"column:tenants_id"`
	StatDate      string    `json:"stat_date" gorm:"column:stat_date"`
	PeriodStart   string    `json:"period_start" gorm:"column:period_start"`
	PeriodEnd     string    `json:"period_end" gorm:"column:period_end"`
	TotalOrder    int       `json:"total_order" gorm:"column:total_order"`
	TotalRevenue  float64   `json:"total_revenue" gorm:"column:total_revenue"`
	ActualRevenue float64   `json:"actual_revenue" gorm:"column:actual_revenue"`
	PaidOrders    int       `json:"paid_orders" gorm:"column:paid_orders"`
	CreateTime    time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime    time.Time `json:"update_time" gorm:"column:update_time"`
}

func (Revenue) TableName() string {
	return "merchant_revenue_stats"
}
