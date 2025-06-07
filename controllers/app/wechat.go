package app

import (
	"github.com/gin-gonic/gin"
	"nasa-go-admin/inout"
	"nasa-go-admin/services/app_service"
)

var useWeChatService = &app_service.WeChatService{}

// WxLogin 微信小程序登录
func WxLogin(c *gin.Context) {
	// 获取请求参数
	var params inout.WxLoginParams
	if err := c.ShouldBindJSON(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	// 调用服务层方法
	userApp, err := useWeChatService.WxLogin(params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	// 返回成功响应
	Resp.Succ(c, userApp)

}
