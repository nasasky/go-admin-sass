package app_model

import "time"

type AppOrder struct {
	Id         int       `json:"id" gorm:"primary_key"`
	UserId     int       `json:"user_id" gorm:"column:user_id"`
	Amount     float64   `json:"amount"`
	Num        int       `json:"num"`
	No         string    `json:"no"`
	TenantsId  int       `json:"tenants_id" gorm:"column:tenants_id"`
	GoodsId    int       `json:"goods_id" gorm:"column:goods_id"`
	Status     string    `json:"status"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

type OrderRefund struct {
	Id         int       `json:"id" gorm:"primary_key"`
	UserId     int       `json:"user_id" gorm:"column:user_id"`
	Amount     float64   `json:"amount"`
	No         string    `json:"no"`
	OrderId    int       `json:"order_id" gorm:"column:order_id"`
	GoodsId    int       `json:"goods_id" gorm:"column:goods_id"`
	Status     string    `json:"status"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

// MerchantRevenueStats 商家收入统计表
type MerchantRevenueStats struct {
	Id              int       `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantsId       int       `gorm:"index:idx_merchant_period;not null" json:"tenants_id" comment:"商家ID"`
	StatDate        string    `gorm:"index:idx_stat_date;not null;type:date" json:"stat_date" comment:"统计日期"`
	StatPeriod      string    `gorm:"index:idx_merchant_period;not null;type:enum('day','week','month','year')" json:"stat_period" comment:"统计周期"`
	PeriodStart     string    `gorm:"index:idx_merchant_period;not null;type:date" json:"period_start" comment:"周期开始日期"`
	PeriodEnd       string    `gorm:"not null;type:date" json:"period_end" comment:"周期结束日期"`
	TotalOrders     int       `gorm:"not null;default:0" json:"total_orders" comment:"订单总数"`
	TotalRevenue    float64   `gorm:"not null;default:0;type:decimal(12,2)" json:"total_revenue" comment:"总收入(元)"`
	ActualRevenue   float64   `gorm:"not null;default:0;type:decimal(12,2)" json:"actual_revenue" comment:"实际收入(扣除退款后)"`
	RefundAmount    float64   `gorm:"not null;default:0;type:decimal(12,2)" json:"refund_amount" comment:"退款总额"`
	PaidOrders      int       `gorm:"not null;default:0" json:"paid_orders" comment:"已支付订单数"`
	PendingOrders   int       `gorm:"not null;default:0" json:"pending_orders" comment:"待支付订单数"`
	CancelledOrders int       `gorm:"not null;default:0" json:"cancelled_orders" comment:"已取消订单数"`
	RefundedOrders  int       `gorm:"not null;default:0" json:"refunded_orders" comment:"已退款订单数"`
	ItemsSold       int       `gorm:"not null;default:0" json:"items_sold" comment:"售出商品总数"`
	CreateTime      time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"create_time" comment:"创建时间"`
	UpdateTime      time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"update_time" comment:"更新时间"`
}

// MerchantRevenueDetails 商家收入详细分析表
type MerchantRevenueDetails struct {
	Id           int       `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantsId    int       `gorm:"index:idx_merchant_goods_date;not null" json:"tenants_id" comment:"商家ID"`
	StatDate     string    `gorm:"index:idx_stat_date;index:idx_merchant_goods_date;not null;type:date" json:"stat_date" comment:"统计日期"`
	GoodsId      int       `gorm:"index:idx_merchant_goods_date;not null" json:"goods_id" comment:"商品ID"`
	GoodsName    string    `gorm:"not null;type:varchar(255)" json:"goods_name" comment:"商品名称"`
	OrderCount   int       `gorm:"not null;default:0" json:"order_count" comment:"订单数"`
	SoldCount    int       `gorm:"not null;default:0" json:"sold_count" comment:"销售数量"`
	Revenue      float64   `gorm:"not null;default:0;type:decimal(12,2)" json:"revenue" comment:"收入金额"`
	RefundCount  int       `gorm:"not null;default:0" json:"refund_count" comment:"退款数量"`
	RefundAmount float64   `gorm:"not null;default:0;type:decimal(12,2)" json:"refund_amount" comment:"退款金额"`
	CreateTime   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"create_time" comment:"创建时间"`
	UpdateTime   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"update_time" comment:"更新时间"`
}

func (AppOrder) TableName() string {
	return "order"
}

func (OrderRefund) TableName() string {
	return "order_refud"
}

func (MerchantRevenueStats) TableName() string {
	return "merchant_revenue_stats"
}
func (MerchantRevenueDetails) TableName() string {
	return "merchant_revenue_details"
}
