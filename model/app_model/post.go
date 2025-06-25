package app_model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// PostStatus 帖子状态
type PostStatus int

const (
	PostStatusPending  PostStatus = 0 // 待审核
	PostStatusApproved PostStatus = 1 // 已通过
	PostStatusRejected PostStatus = 2 // 已拒绝
)

// StringArray 是一个字符串数组类型，用于存储图片URL数组
type StringArray []string

// Value 实现 driver.Valuer 接口
func (a StringArray) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan 实现 sql.Scanner 接口
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}
	return json.Unmarshal(value.([]byte), a)
}

// UserPost 用户发布的信息
type UserPost struct {
	ID           uint        `json:"id" gorm:"primarykey"`
	UserID       uint        `json:"user_id" gorm:"not null;index"`                 // 发布用户ID
	Title        string      `json:"title" gorm:"size:200;not null"`                // 标题
	Content      string      `json:"content" gorm:"type:text;not null"`             // 内容
	Images       StringArray `json:"images" gorm:"type:json"`                       // 图片数组
	Status       PostStatus  `json:"status" gorm:"type:tinyint;default:0;not null"` // 状态：0待审核 1已通过 2已拒绝
	RejectReason string      `json:"reject_reason" gorm:"size:200"`                 // 拒绝原因
	CreatedAt    time.Time   `json:"created_at"`                                    // 创建时间
	UpdatedAt    time.Time   `json:"updated_at"`                                    // 更新时间
}

// TableName 指定表名
func (UserPost) TableName() string {
	return "user_posts"
}
