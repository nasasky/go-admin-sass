package app

import (
	"nasa-go-admin/inout"
	"nasa-go-admin/services/app_service"

	"github.com/gin-gonic/gin"
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

// GetUserInfo
func GetUserInfo(c *gin.Context) {
	var uid, _ = c.Get("uid")
	user, err := userService.GetUserInfo(uid.(int))
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, user)
}

// Refresh
func Refresh(c *gin.Context) {
	var uid, _ = c.Get("uid")
	token, err := userService.Refresh(uid.(int))
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, token)
}

// UpdateUserInfo
func UpdateUserInfo(c *gin.Context) {
	var params inout.UpdateUserAppReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	var uid, _ = c.Get("uid")
	params.Id = uid.(int)

	// Update user information
	err := userService.UpdateUserInfo(params.Id, params.Username, params.Phone, params.NickName, params.Address, params.Email, params.Gender)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, nil)
}
