package app_model

import "time"

// OrderStatusHistory 订单状态变更历史
type OrderStatusHistory struct {
	Id         int       `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	OrderId    int       `gorm:"not null;comment:订单ID" json:"order_id"`
	OrderNo    string    `gorm:"type:varchar(50);not null;comment:订单号" json:"order_no"`
	FromStatus string    `gorm:"type:varchar(20);not null;comment:原状态" json:"from_status"`
	ToStatus   string    `gorm:"type:varchar(20);not null;comment:新状态" json:"to_status"`
	Operator   string    `gorm:"type:varchar(50);not null;comment:操作者" json:"operator"`
	Reason     string    `gorm:"type:varchar(200);comment:变更原因" json:"reason"`
	CreateTime time.Time `gorm:"type:datetime;not null;comment:创建时间" json:"create_time"`
}

// TableName 设置表名
func (OrderStatusHistory) TableName() string {
	return "order_status_history"
}
