package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/redis"
	"nasa-go-admin/services/admin_service"
	"strconv"
	"time"

	"nasa-go-admin/utils"

	"github.com/gin-gonic/gin"
)

var tenantsService = &admin_service.TenantsService{}

// register
func TenantsRegister(c *gin.Context) {
	var params inout.AddTenantsReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	// Check if the phone number already exists
	if tenantsService.UserExists(params.Username) {
		Resp.Err(c, 20002, "名称已存在，请修改")
		return
	}
	// Create new user
	newTanants, err := tenantsService.CreateUser(params.Username, params.Password, params.Phone, params.Type, params.RoleId)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, newTanants)

}

// update
func UpdateTenants(c *gin.Context) {
	var params inout.UpdateTenantsReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	// Check if the phone number already exists
	if tenantsService.UserExists(params.Username) {
		Resp.Err(c, 20002, "名称已存在，请修改")
		return
	}
	// Create new user
	newTanants, err := tenantsService.UpdateUser(params.Id, params.Username, params.Password, params.Phone, params.Type, params.RoleId)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, newTanants)

}

func TenantsLogin(c *gin.Context) {
	var params inout.LoginTenantsReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	// Check if the phone number already exists
	user, err := tenantsService.Login(c, params.Username, params.Password)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, user)
}

// login
func Login(c *gin.Context) {
	var params inout.LoginAdminReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, "参数错误："+err.Error())
		return
	}

	// 验证码校验和登录验证都在service层处理
	user, err := tenantsService.Login(c, params.Username, params.Password)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, user)
}

// GetRoute
func GetRoute(c *gin.Context) {

	var uid, _ = c.Get("uid")

	routerList, _ := tenantsService.GetRoutes(c, uid.(int))

	Resp.Succ(c, routerList)

}

// GetMenu
func GetMenu(c *gin.Context) {

	var uid, _ = c.Get("uid")

	menuList, _ := tenantsService.GetMenus(c, uid.(int))

	Resp.Succ(c, menuList)
}

// AddMenu
func AddMenu(c *gin.Context) {
	var params inout.AddMenuReq
	var uid, _ = c.Get("uid")
	var roleId, _ = c.Get("roleId")
	fmt.Println(roleId)
	fmt.Print(uid)
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	var parentId *int64
	if params.ParentId != 0 {
		tmp := int64(params.ParentId)
		parentId = &tmp
	} else {
		parentId = nil
	}
	menu := admin_model.PermissionMenu{
		ParentId: parentId,
		Label:    params.Title,
		Icon:     params.Icon,
		Rule:     params.Rule,
		Key:      params.Key,
		Type:     params.Type,
		Show:     params.Show,
		Sort:     params.Sort,
		Path:     params.Path,
		Title:    params.Title,
		Enable:   params.Show,
	}
	// Check if the menu already exists
	Id, err := tenantsService.AddMenu(c, menu)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, Id)
}

// GetMenuDetail
func GetMenuDetail(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		Resp.Err(c, 20001, "id不能为空")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	menu, _ := tenantsService.GetMenuDetail(c, id)

	Resp.Succ(c, menu)

}

// UpdateMenu
func UpdateMenu(c *gin.Context) {
	var params inout.UpdateMenuReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	var parentId *int64
	if params.ParentId != 0 {
		tmp := int64(params.ParentId)
		parentId = &tmp
	} else {
		parentId = nil
	}
	menu := admin_model.PermissionMenu{
		ParentId: parentId,
		Label:    params.Title,
		Icon:     params.Icon,
		Rule:     params.Rule,
		Key:      params.Key,
		Type:     params.Type,
		Show:     params.Show,
		Sort:     params.Sort,
		Path:     params.Path,
		Title:    params.Title,
		Enable:   params.Show,
	}
	// Check if the menu already exists
	id, err := tenantsService.UpdateMenu(c, params.Id, menu)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, id)
}

// DeleteMenu
func DeleteMenu(c *gin.Context) {
	var params struct {
		Ids []int `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	if len(params.Ids) == 0 {
		Resp.Err(c, 20001, "ids不能为空")
		return
	}
	err := tenantsService.DeleteMenu(c, params.Ids)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, nil)
}

// GetUserInfo
func GetUserInfo(c *gin.Context) {

	var uid, _ = c.Get("uid")

	userInfo, _ := tenantsService.GetUserInfo(c, uid.(int))

	Resp.Succ(c, userInfo)

}

// UpdateUserProfile 修改用户信息
func UpdateUserProfile(c *gin.Context) {
	var params inout.UpdateUserProfileReq
	if err := c.ShouldBindJSON(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// 获取当前用户ID
	currentUserId := c.GetInt("uid")

	// 只能修改自己的信息
	if params.Id != currentUserId {
		Resp.Err(c, 20001, "只能修改自己的信息")
		return
	}

	err := tenantsService.UpdateUserProfile(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, nil)
}

// UpdateUserPassword 修改用户密码
func UpdateUserPassword(c *gin.Context) {
	var params inout.UpdateUserPasswordReq
	if err := c.ShouldBindJSON(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// 获取当前用户ID
	currentUserId := c.GetInt("uid")

	// 只能修改自己的密码
	if params.Id != currentUserId {
		Resp.Err(c, 20001, "只能修改自己的密码")
		return
	}

	err := tenantsService.UpdateUserPassword(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, nil)
}

// 退出登录
func Logout(c *gin.Context) {

	//清除token
	tenantsService.Logout(c)
	Resp.Succ(c, nil)

}

// GetCaptcha 获取验证码
func GetCaptcha(c *gin.Context) {
	// 生成更大尺寸的验证码图片，提高清晰度
	svg, code := utils.GenerateSVG(160, 60)

	// 设置过期时间
	expireTime := time.Now().Add(5 * time.Minute).Unix()

	// 将验证码存储到Redis
	captchaData := map[string]interface{}{
		"code":   code,
		"expire": expireTime,
	}
	captchaJSON, _ := json.Marshal(captchaData)

	// 打印调试信息
	fmt.Printf("生成新验证码:\n")
	fmt.Printf("- 验证码: %v\n", code)
	fmt.Printf("- 过期时间: %v\n", expireTime)

	// 使用固定的key存储到Redis，设置5分钟过期
	captchaKey := "latest_captcha"
	err := redis.GetClient().Set(context.Background(), captchaKey, string(captchaJSON), 5*time.Minute).Err()
	if err != nil {
		fmt.Printf("Redis保存失败: %v\n", err)
		c.JSON(500, gin.H{"error": "Captcha save failed"})
		return
	}
	fmt.Printf("Redis保存成功\n")

	// 设置响应头
	c.Header("Content-Type", "image/svg+xml; charset=utf-8")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// 返回验证码图片
	c.Data(200, "image/svg+xml", svg)
}

// GetCaptchaStatus 获取验证码开关状态
func GetCaptchaStatus(c *gin.Context) {
	enabled := tenantsService.IsCaptchaEnabled()
	Resp.Succ(c, gin.H{
		"captcha_enabled": enabled,
	})
}

// UpdateCaptchaStatus 更新验证码开关状态
func UpdateCaptchaStatus(c *gin.Context) {
	var params struct {
		Enabled bool `json:"enabled" binding:"required"`
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		Resp.Err(c, 20001, "参数错误："+err.Error())
		return
	}

	err := tenantsService.UpdateCaptchaStatus(params.Enabled)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, gin.H{
		"captcha_enabled": params.Enabled,
		"message":         "验证码开关更新成功",
	})
}
