package admin_service

import (
	"crypto/md5"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"nasa-go-admin/db"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/redis"
	"nasa-go-admin/utils"
	"strconv"
	"time"
)

type TenantsService struct{}

// CreateUser
func (s *TenantsService) CreateUser(username, password, phone string, roleId int) (*admin_model.TenantsUser, error) {
	var newUserApp admin_model.TenantsUser
	err := db.Dao.Transaction(func(tx *gorm.DB) error {
		newUserApp = admin_model.TenantsUser{
			Username:   username,
			Password:   fmt.Sprintf("%x", md5.Sum([]byte(password))),
			Phone:      phone,
			RoleId:     roleId,
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}
		err := tx.Create(&newUserApp).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	err = InsertProfile(&newUserApp)
	if err != nil {
		return nil, err
	}
	return &newUserApp, nil
}

func InsertProfile(a *admin_model.TenantsUser) error {
	profile := admin_model.AdminUserReq{
		Username:   a.Username,
		Password:   a.Password,
		Phone:      a.Phone,
		TenantId:   a.ID,
		CreateTime: a.CreateTime,
		UpdateTime: a.UpdateTime,
	}
	return db.Dao.Create(&profile).Error
}
func (s *TenantsService) UserExists(username string) bool {
	var existingUser admin_model.TenantsUser
	if err := db.Dao.Where("username = ?", username).First(&existingUser).Error; err == nil {
		return true
	}
	return false
}

// Login

func (s *TenantsService) LoginTenants(c *gin.Context, username, password string) (*admin_model.AdminUser, error) {
	var user admin_model.AdminUser
	hashedPassword := fmt.Sprintf("%x", md5.Sum([]byte(password)))
	err := db.Dao.Where("(username = ? OR phone = ?) AND password = ?", username, username, hashedPassword).First(&user).Error

	// 如果没有找到记录，则返回提示用户不存在
	if err != nil {
		Resp.Err(c, 20001, "用户不存在或密码错误")
		return nil, err
	}

	// 如果密码不正确，则返回提示密码错误
	if user.Password != hashedPassword {
		Resp.Err(c, 20001, "密码错误")
		return nil, fmt.Errorf("密码错误")
	}

	user.Token = utils.GenerateToken(user.ID)
	expiration := time.Hour * 24 // 过期时间为 24 小时
	err = redis.StoreToken(strconv.Itoa(user.ID), user.Token, expiration)
	if err != nil {
		return nil, err
	}

	return &user, nil

}

// Login

func (s *TenantsService) Login(c *gin.Context, username, password string) (*admin_model.AdminUser, error) {
	var user admin_model.AdminUser
	hashedPassword := fmt.Sprintf("%x", md5.Sum([]byte(password)))
	err := db.Dao.Where("(username = ? OR phone = ?) AND password = ?", username, username, hashedPassword).First(&user).Error

	// 如果没有找到记录，则返回提示用户不存在
	if err != nil {
		Resp.Err(c, 20001, "用户不存在或密码错误")
		return nil, err
	}

	// 如果密码不正确，则返回提示密码错误
	if user.Password != hashedPassword {
		Resp.Err(c, 20001, "密码错误")
		return nil, fmt.Errorf("密码错误")
	}

	user.Token = utils.GenerateToken(user.ID)
	expiration := time.Hour * 24 // 过期时间为 24 小时
	err = redis.StoreToken(strconv.Itoa(user.ID), user.Token, expiration)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
