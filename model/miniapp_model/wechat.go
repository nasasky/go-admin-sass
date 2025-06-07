package miniapp_model

import "time"

// UserSubscription 用户订阅记录
type UserSubscription struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	OpenID      string    `json:"open_id" gorm:"column:open_id;index;not null"`
	TemplateID  string    `json:"template_id" gorm:"column:template_id;not null"`
	SubscribeAt time.Time `json:"subscribe_at" gorm:"column:subscribe_at"`
	Status      int       `json:"status" gorm:"column:status;default:1"` // 1-已订阅 0-取消订阅
}

// PushHistory 推送历史记录
type PushHistory struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	OpenID     string    `json:"open_id" gorm:"column:open_id;index;not null"`
	TemplateID string    `json:"template_id" gorm:"column:template_id;not null"`
	Content    string    `json:"content" gorm:"column:content;type:text"`
	PushTime   time.Time `json:"push_time" gorm:"column:push_time"`
	Status     int       `json:"status" gorm:"column:status;default:1"` // 1-推送成功 0-推送失败
}

// TableName 设置表名
func (UserSubscription) TableName() string {
	return "wx_user_subscription"
}

// TableName 设置表名
func (PushHistory) TableName() string {
	return "wx_push_history"
}
