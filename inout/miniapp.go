package inout

// 订阅消息请求
type SubscribeTemplateReq struct {
	TemplateID string `json:"template_id" binding:"required"` // 模板ID
	OpenID     string `json:"openid" binding:"required"`      // 用户OpenID
}

// 手动推送消息请求
type PushMessageReq struct {
	OpenID     string `json:"openid" binding:"required"`      // 用户OpenID
	TemplateID string `json:"template_id" binding:"required"` // 模板ID
}

// 微信订阅消息响应
type WxMsgResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}
