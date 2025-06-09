package app_model

import "time"

type AppWallet struct {
	UserId     int       `json:"user_id" gorm:"column:user_id"`
	Money      float64   `json:"money"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

// 充值记录
type AppRecharge struct {
	ID              int        `json:"id" gorm:"primary_key"`
	UserID          int        `json:"user_id" gorm:"column:user_id"`
	OrderID         *int       `json:"order_id" gorm:"column:order_id"`
	OrderNo         string     `json:"order_no" gorm:"column:order_no"`
	TransactionType string     `json:"transaction_type" gorm:"column:transaction_type"`
	Amount          float64    `json:"amount"`
	BalanceBefore   float64    `json:"balance_before" gorm:"column:balance_before"`
	BalanceAfter    float64    `json:"balance_after" gorm:"column:balance_after"`
	Status          string     `json:"status"`
	PaymentMethod   string     `json:"payment_method" gorm:"column:payment_method"`
	PaymentNo       string     `json:"payment_no" gorm:"column:payment_no"`
	Remark          string     `json:"remark"`
	CreateTime      time.Time  `json:"create_time" gorm:"column:create_time"`
	UpdateTime      time.Time  `json:"update_time" gorm:"column:update_time"`
	CompleteTime    *time.Time `json:"complete_time" gorm:"column:complete_time"`
}

func (AppWallet) TableName() string {
	return "app_wallet"
}

func (AppRecharge) TableName() string {
	return "app_recharge"
}
