package app_service

import (
	"fmt"
	"nasa-go-admin/db"
	"nasa-go-admin/model/app_model"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WalletService struct{}

// GetUserWallet 获取用户钱包
func (s *WalletService) GetUserWallet(c *gin.Context, uid int) (interface{}, error) {
	fmt.Println("uid", uid)
	var data app_model.AppWallet

	// 开始事务
	tx := db.Dao.Begin()

	// 查询用户钱包
	err := tx.Model(&app_model.AppWallet{}).Where("user_id = ?", uid).First(&data).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 用户不存在，创建新用户钱包
			data = app_model.AppWallet{
				UserId:     uid,
				Money:      0.00,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			}
			if err := tx.Create(&data).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		} else {
			tx.Rollback()
			return nil, err
		}
	}
	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	return data, nil
}

// Recharge 用户充值
func (s *WalletService) Recharge(c *gin.Context, uid int, params app_model.AppWallet) error {
	// 开始事务
	tx := db.Dao.Begin()
	// 检查用户是否存在
	var wallet app_model.AppWallet
	err := tx.Where("user_id = ?", uid).First(&wallet).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 用户不存在，创建新用户
			wallet = app_model.AppWallet{
				UserId:     uid,
				Money:      0.00,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			}
			if err := tx.Create(&wallet).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else {
			tx.Rollback()
			return err
		}
	}

	// 更新用户余额和更新时间
	err = tx.Model(&app_model.AppWallet{}).Where("user_id = ?", uid).Updates(map[string]interface{}{
		"money":       gorm.Expr("money + ?", params.Money),
		"create_time": time.Now(),
		"update_time": time.Now(),
	}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// 获取更新后的余额
	err = tx.Where("user_id = ?", uid).First(&wallet).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// 插入充值记录
	recharge := app_model.AppRecharge{
		UserID:          uid,
		TransactionType: "recharge",
		Amount:          params.Money,
		BalanceBefore:   wallet.Money - params.Money,
		BalanceAfter:    wallet.Money,
		Description:     "用户充值",
		CreateTime:      time.Now(),
	}
	err = tx.Create(&recharge).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务并处理错误
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
