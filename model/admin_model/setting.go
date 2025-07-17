package admin_model

import "time"

type SettingList struct {
	Id         int       `json:"id"`
	UserId     int       `json:"user_id"`
	Name       string    `json:"name"`
	Appid      string    `json:"appid"`
	Secret     string    `json:"secret"`
	Tips       string    `json:"tips"`
	Endpoint   string    `json:"endpoint"`
	BucketName string    `json:"bucket_name"`
	BaseUrl    string    `json:"base_url"`
	Type       string    `json:"type"`
	Value      string    `json:"value"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}

type Setting struct {
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
	Items    []SettingList `json:"items"`
}

// 修改Setting结构体
type SettingType struct {
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
	Items    []DictTypeResp `json:"items"` // 修改为DictType类型
}

type SettingTypeList struct {
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
	Items    []DictDetail `json:"items"` // 修改为DictDetail类型
}

type DictTypeResp struct {
	Id         int    `json:"id"`
	TypeName   string `json:"type_name"`
	TypeCode   string `json:"type_code"`
	IsLock     string `json:"is_lock"`
	IsShow     string `json:"is_show"`
	Type       string `json:"type"`
	DelFlag    string `json:"del_flag"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
	Remark     string `json:"remark"`
}

type DictType struct {
	Id         int       `json:"id"`
	TypeName   string    `json:"type_name"`
	TypeCode   string    `json:"type_code"`
	IsLock     string    `json:"is_lock"`
	IsShow     string    `json:"is_show"`
	Type       string    `json:"type"`
	DelFlag    string    `json:"del_flag"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
	Remark     string    `json:"remark"`
}

type DictDetail struct {
	Id                int       `json:"id"`
	Alias             string    `json:"alias"`
	DelFlag           string    `json:"del_flag"`
	Code              string    `json:"code"`
	SysDictTypeId     int       `json:"sys_dict_type_id"`
	CodeName          string    `json:"code_name"`
	CallbackShowStyle string    `json:"callback_show_style"`
	Sort              int       `json:"sort"`
	IsLock            string    `json:"is_lock"`
	IsShow            string    `json:"is_show"`
	SysDictTypeCode   string    `json:"sys_dict_type_code"`
	CreateTime        time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime        time.Time `json:"update_time" gorm:"column:update_time"`
	Remark            string    `json:"remark"`
}

type AddDictDetail struct {
	Id                int       `json:"id"`
	Alias             string    `json:"alias"`
	DelFlag           string    `json:"del_flag"`
	Code              string    `json:"code"`
	SysDictTypeId     int       `json:"sys_dict_type_id"`
	CodeName          string    `json:"code_name"`
	CallbackShowStyle string    `json:"callback_show_style"`
	Sort              int       `json:"sort"`
	IsLock            string    `json:"is_lock"`
	IsShow            string    `json:"is_show"`
	CreateTime        time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime        time.Time `json:"update_time" gorm:"column:update_time"`
	Remark            string    `json:"remark"`
}

func (SettingList) TableName() string {
	return "config_setting"
}

func (DictType) TableName() string {
	return "sys_dict_type"
}
func (DictDetail) TableName() string {
	return "sys_dict"
}
func (AddDictDetail) TableName() string {
	return "sys_dict"
}
