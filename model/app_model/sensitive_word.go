package app_model

import "time"

// SensitiveWord 敏感词配置
type SensitiveWord struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	Word      string    `json:"word" gorm:"size:100;not null;uniqueIndex"` // 敏感词
	Level     int       `json:"level" gorm:"type:tinyint;default:1"`       // 敏感级别：1一般 2中等 3严重
	IsEnabled bool      `json:"is_enabled" gorm:"default:true"`            // 是否启用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (SensitiveWord) TableName() string {
	return "sensitive_words"
}
