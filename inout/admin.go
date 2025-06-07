package inout

type DictDetailReq struct {
	Id            int    `json:"id"`
	SysDictTypeId string `json:"sys_dict_type_id"`
	CodeName      string `json:"code_name"`
	Label         string `json:"label"`
	Value         string `json:"value"`
	Sort          int    `json:"sort"`
	IsDefault     string `json:"is_default"`
	IsLock        string `json:"is_lock"`
	IsShow        string `json:"is_show"`
	CreateTime    string `json:"create_time"`
	UpdateTime    string `json:"update_time"`
	Remark        string `json:"remark"`
}

type RoleDetail struct {
	SelectIds []int `json:"selectIds"`
}

type GetRoleDetailReq struct {
	Id int `form:"id" binding:"required"`
}
