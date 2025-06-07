package app_service

import (
	"crypto/md5"
	"fmt"
	"nasa-go-admin/db"
	"nasa-go-admin/model/app_model"
	"nasa-go-admin/pkg/jwt"
	"nasa-go-admin/redis"
	"strconv"
	"time"

	"gorm.io/gorm"
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
func (s *UserService) Login(phone, password string) (map[string]interface{}, error) {
	var user app_model.LoginUserApp
	if err := db.Dao.Where("phone = ?", phone).First(&user).Error; err != nil {
		return nil, fmt.Errorf("手机号不存在")
	}

	// Check if the password is correct
	if user.Password != fmt.Sprintf("%x", md5.Sum([]byte(password))) {
		return nil, fmt.Errorf("密码不对")
	}

	//在user这里返回多一个token值
	//user.Token = utils.GenerateTokenApp(user.ID, '0')
	//// 存储 token 到 Redis
	expiration := time.Hour * 24 // 过期时间为 24 小时
	err := redis.StoreToken(strconv.Itoa(user.ID), user.Token, expiration)
	token, err := GetToken(user.ID)
	if err != nil {
		return nil, err
	}
	user.Token = token
	err = redis.StoreUserInfo(strconv.Itoa(user.ID), map[string]interface{}{
		"username": user.Username,
		"phone":    user.Phone,
		"token":    user.Token,
	}, expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to store user info: %v", err)
	}
	responseData := map[string]interface{}{
		"user":  user,
		"token": user.Token,
	}
	return responseData, nil
}

// GetUserInfo
func (s *UserService) GetUserInfo(uid int) (*app_model.UserApp, error) {
	var user app_model.UserApp

	if err := db.Dao.Where("id = ?", uid).First(&user).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	token, err := GetToken(user.ID)
	if err != nil {
		return nil, err
	}
	user.Token = token

	return &user, nil
}

func (s *UserService) UserExists(phone string) bool {
	var existingUser app_model.UserApp
	if err := db.Dao.Where("phone = ?", phone).First(&existingUser).Error; err == nil {
		return true
	}
	return false
}

// Refresh
func (s *UserService) Refresh(uid int) (string, error) {
	token, err := GetToken(uid)
	if err != nil {
		return "", err
	}
	return token, nil
}

// 封装一个获取token的方法
func GetToken(uid int) (string, error) {
	token, err := jwt.GenerateAppToken(uid, 0, 0)
	if err != nil {
		return "", fmt.Errorf("生成令牌失败: %w", err)
	}

	// 存储 token 到 Redis
	expiration := time.Hour * 24 // 过期时间为 24 小时
	err = redis.StoreToken(strconv.Itoa(uid), token, expiration)
	if err != nil {
		return "", err
	}
	return token, nil
}

// UpdateUserInfo 更新用户信息
func (s *UserService) UpdateUserInfo(id int, username, phone, nickName, address, email string, gender int) error {
	updates := make(map[string]interface{})

	if username != "" {
		updates["username"] = username
	}
	if phone != "" {
		updates["phone"] = phone
	}
	if address != "" {
		updates["address"] = address
	}
	if email != "" {
		updates["email"] = email
	}
	if gender >= 0 {
		updates["gender"] = gender
	}

	updates["update_time"] = time.Now()

	// 更新用户基本信息
	if err := db.Dao.Table("app_user").Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新用户信息失败: %w", err)
	}

	// 更新用户详细信息（如果存在profile表）
	profileUpdates := make(map[string]interface{})
	if nickName != "" {
		profileUpdates["nickName"] = nickName
	}
	if address != "" {
		profileUpdates["address"] = address
	}
	if email != "" {
		profileUpdates["email"] = email
	}

	if len(profileUpdates) > 0 {
		// 尝试更新profile，如果不存在则创建
		var existingProfile app_model.AppProfile
		if err := db.Dao.Where("id = ?", id).First(&existingProfile).Error; err != nil {
			// Profile不存在，创建新的
			newProfile := app_model.AppProfile{
				ID:       id,
				NickName: nickName,
				Address:  address,
				Email:    email,
			}
			db.Dao.Create(&newProfile)
		} else {
			// Profile存在，更新
			db.Dao.Table("app_user").Where("id = ?", id).Updates(profileUpdates)
		}
	}

	return nil
}
