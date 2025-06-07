package miniapp

import (
	"nasa-go-admin/inout"
	"nasa-go-admin/services/miniapp_service"

	"github.com/gin-gonic/gin"
)

// SubscribeTemplate 用户订阅消息模板
func SubscribeTemplate(c *gin.Context) {
	var req inout.SubscribeTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// 存储用户订阅信息
	err := miniapp_service.SaveUserSubscription(req.OpenID, req.TemplateID)
	if err != nil {
		Resp.Err(c, 20002, "订阅失败: "+err.Error())
		return
	}

	Resp.Succ(c, gin.H{"msg": "订阅成功"})
}

// PushMessage 推送消息（可用于管理员手动触发推送）
func PushMessage(c *gin.Context) {
	var req inout.PushMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Resp.Err(c, 20001, "参数错误: "+err.Error())
		return
	}

	// 推送消息
	err := miniapp_service.SendSubscribeMsg(req.OpenID, req.TemplateID, "")
	if err != nil {
		Resp.Err(c, 20003, "推送失败: "+err.Error())
		return
	}

	Resp.Succ(c, gin.H{"msg": "推送成功"})
}
