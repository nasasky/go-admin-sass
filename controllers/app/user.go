package app

import (
	"github.com/gin-gonic/gin"
	"nasa-go-admin/inout"
	"nasa-go-admin/services/app_service"
)

var userService = &app_service.UserService{}

func Register(c *gin.Context) {
	var params inout.AddUserAppReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// Check if the phone number already exists
	if userService.UserExists(params.Phone) {
		Resp.Err(c, 20002, "用户已存在")
		return
	}

	// Create new user
	newUserApp, err := userService.CreateUser(params.Username, params.Password, params.Phone)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, newUserApp)
}

// Login
func Login(c *gin.Context) {
	var params inout.LoginAppReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// Check if the phone number already exists
	userApp, err := userService.Login(params.Phone, params.Password)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, userApp)
}
