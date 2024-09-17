package app_model

type AppProfile struct {
	ID         int    `json:"id"`
	Avatar     string `json:"avatar"`
	Address    string `json:"address"`
	Email      string `json:"email"`
	NickName   string `gorm:"column:nickName"`
	CreateTime string `json:"createTime"`
	UpdateTime string `json:"updateTime"`
	Phone      string `json:"phone"`
}

func (AppProfile) TableName() string {
	return "app_user"
}
