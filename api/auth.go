package api

import (
	"crypto/md5"
	"fmt"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model"
	"nasa-go-admin/pkg/jwt"
	"nasa-go-admin/pkg/security"
	"nasa-go-admin/utils"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

var Auth = &auth{}

type auth struct {
}

func (auth) Captcha(c *gin.Context) {
	svg, code := utils.GenerateSVG(80, 40)
	session := sessions.Default(c)
	session.Set("captch", code)
	session.Save()
	// 设置 Content-Type 为 "image/svg+xml"
	c.Header("Content-Type", "image/svg+xml; charset=utf-8")
	// 返回验证码
	c.Data(http.StatusOK, "image/svg+xml", svg)
}

func (auth) Login(c *gin.Context) {
	var params inout.LoginReq
	err := c.Bind(&params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// 验证输入安全性
	if err := security.ValidateInput(params.Username); err != nil {
		Resp.Err(c, 20001, "用户名包含非法字符")
		return
	}

	session := sessions.Default(c)
	if params.Captcha != session.Get("captch") {
		Resp.Err(c, 20001, "验证码不正确")
		return
	}

	var info *model.User
	// 首先查询用户信息
	db.Dao.Model(model.User{}).Where("username = ?", params.Username).Find(&info)
	if info.ID == 0 {
		Resp.Err(c, 20001, "账号或密码不正确")
		return
	}

	// 检查密码
	var passwordValid bool
	if info.PasswordBcrypt != "" {
		// 使用新的 bcrypt 密码验证
		passwordValid = security.CheckPasswordHash(params.Password, info.PasswordBcrypt)
	} else {
		// 兼容旧的 MD5 密码（临时方案）
		md5Hash := fmt.Sprintf("%x", md5.Sum([]byte(params.Password)))
		passwordValid = (info.Password == md5Hash)

		// 如果 MD5 验证成功，升级到 bcrypt
		if passwordValid {
			newHash, err := security.HashPassword(params.Password)
			if err == nil {
				db.Dao.Model(&info).Update("password_bcrypt", newHash)
			}
		}
	}

	if !passwordValid {
		Resp.Err(c, 20001, "账号或密码不正确")
		return
	}

	// 使用安全的JWT管理器
	jwtManager := jwt.NewSecureJWTManager()
	token, err := jwtManager.GenerateToken(info.ID, 0, 0)
	if err != nil {
		Resp.Err(c, 20001, "生成令牌失败")
		return
	}

	Resp.Succ(c, inout.LoginRes{
		AccessToken: token,
	})
}

func (auth) password(c *gin.Context) {
	var params inout.AuthPwReq
	err := c.Bind(&params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	uid, _ := c.Get("uid")
	var oldCun int64
	db.Dao.Model(model.User{}).Where("id=? and password=?", uid, fmt.Sprintf("%x", md5.Sum([]byte(params.OldPassword))))
	if oldCun > 0 {
		db.Dao.Model(model.User{}).
			Where("id=? ", uid).
			Update("password", fmt.Sprintf("%x", md5.Sum([]byte(params.NewPassword))))
	}
	Resp.Succ(c, true)
}
func (auth) Logout(c *gin.Context) {
	Resp.Succ(c, true)
}
