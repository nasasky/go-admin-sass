package admin_model

type FeishuChatList struct {
	Chat_id int    `json:"chat_id"`
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	OwnerId int    `json:"owner_id"`
}

type FeishuGroupListResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		HasMore bool `json:"has_more"`
		Items   []struct {
			Avatar      string `json:"avatar"`
			ChatID      string `json:"chat_id"`
			ChatStatus  string `json:"chat_status"`
			Description string `json:"description"`
			External    bool   `json:"external"`
			Name        string `json:"name"`
			OwnerID     string `json:"owner_id"`
			OwnerIDType string `json:"owner_id_type"`
			TenantKey   string `json:"tenant_key"`
		} `json:"items"`
		PageToken string `json:"page_token"`
	} `json:"data"`
}

type AddFeishuGroup struct {
	Id      int    `json:"id"`
	ChatId  string `json:"chat_id"`
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	OwnerId string `json:"owner_id"`
}

type FeishuResponse struct {
	Code              int    `json:"code"`
	Expire            int    `json:"expire"`
	Msg               string `json:"msg"`
	TenantAccessToken string `json:"tenant_access_token"`
}

// FeishuRequest 定义请求体结构
type FeishuRequest struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

type FeishuMessageRequest struct {
	ReceiveId     string `json:"receive_id"`
	ReceiveIdType string `json:"receive_id_type"`
	MsgType       string `json:"msg_type"`
	Content       string `json:"content"`
	UUID          string `json:"uuid,omitempty"` // 选填
}

func (AddFeishuGroup) TableName() string {
	return "feishu_group"
}
