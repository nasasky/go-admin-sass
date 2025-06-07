package app

import (
	"fmt"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/app_model"
	"nasa-go-admin/services/app_service"

	"github.com/gin-gonic/gin"
)

var walletService = &app_service.WalletService{}

// GetUserWallet
func GetUserWallet(c *gin.Context) {
	uid := c.GetInt("uid")
	wallet, err := walletService.GetUserWallet(c, uid)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, wallet)
}

// Recharge
func Recharge(c *gin.Context) {
	var params inout.RechargeReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	uid := c.GetInt("uid")
	data := app_model.AppWallet{

		Money: params.Amount,
	}
	fmt.Println("data", data)
	err := walletService.Recharge(c, uid, data)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, nil)
}
