package admin_model

import "time"

type SystemInfo struct {
	Id          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	SystemName  string    `json:"system_name" gorm:"column:system_name"`   // 系统名称
	SystemTitle string    `json:"system_title" gorm:"column:system_title"` // 系统标题
	IcpNumber   string    `json:"icp_number" gorm:"column:icp_number"`     // 备案号
	Copyright   string    `json:"copyright" gorm:"column:copyright"`       // 版权信息
	TenantsId   int       `json:"tenants_id" gorm:"column:tenants_id"`     // 租户ID
	Status      int       `json:"status"`                                  // 状态：1-启用 0-禁用
	CreateTime  time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime  time.Time `json:"update_time" gorm:"column:update_time"`
}

func (SystemInfo) TableName() string {
	return "system_info"
}
