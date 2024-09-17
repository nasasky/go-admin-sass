package app_service

import (
	"crypto/md5"
	"fmt"
	"gorm.io/gorm"
	"nasa-go-admin/db"
	"nasa-go-admin/model/app_model"
	"nasa-go-admin/redis"
	"nasa-go-admin/utils"
	"strconv"
	"time"
)

type UserService struct{}

func (s *UserService) CreateUser(username, password, phone string) (*app_model.LoginUser, error) {
	var newUserApp app_model.LoginUser
	err := db.Dao.Transaction(func(tx *gorm.DB) error {
		newUserApp = app_model.LoginUser{
			Username:   username,
			Password:   fmt.Sprintf("%x", md5.Sum([]byte(password))),
			Phone:      phone,
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}
		err := tx.Create(&newUserApp).Error
		if err != nil {
			return err
		}
		tx.Create(&app_model.AppProfile{
			ID:       newUserApp.ID,
			NickName: newUserApp.Username,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}
	return &newUserApp, nil
}

// Login
func (s *UserService) Login(phone, password string) (*app_model.UserApp, error) {
	var user app_model.UserApp
	if err := db.Dao.Where("phone = ?", phone).First(&user).Error; err != nil {
		return nil, fmt.Errorf("手机号不存在")
	}

	// Check if the password is correct
	if user.Password != fmt.Sprintf("%x", md5.Sum([]byte(password))) {
		return nil, fmt.Errorf("密码不对")
	}

	//在user这里返回多一个token值
	user.Token = utils.GenerateTokenApp(user.ID)
	// 存储 token 到 Redis
	expiration := time.Hour * 24 // 过期时间为 24 小时
	err := redis.StoreToken(strconv.Itoa(user.ID), user.Token, expiration)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) UserExists(phone string) bool {
	var existingUser app_model.UserApp
	if err := db.Dao.Where("phone = ?", phone).First(&existingUser).Error; err == nil {
		return true
	}
	return false
}
