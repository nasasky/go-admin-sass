package app_service

import (
	"log"
	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/app_model"

	"github.com/gin-gonic/gin"
)

type OrderRefundService struct{}

// Refund 申请退款
func (s *OrderRefundService) Refund(c *gin.Context, uid int, params inout.RefundReq) (interface{}, error) {

	// 查询订单是否存在
	var order app_model.AppOrder
	err := db.Dao.Where("id = ? AND user_id = ?", params.OrderId, uid).First(&order).Error
	if err != nil {
		return nil, err

	}

	// 退款订单
	refund := app_model.OrderRefund{
		UserId:  uid,
		Amount:  order.Amount,
		No:      order.No,
		GoodsId: order.GoodsId,
		Status:  "0",
		OrderId: order.Id,
	}
	err = db.Dao.Create(&refund).Error
	if err != nil {
		return nil, err
	}

	// 更新订单状态
	err = db.Dao.Model(&order).Update("status", "3").Error
	if err != nil {
		return nil, err
	}

	// 使用新的安全订单创建器进行订单取消
	secureCreator := GetGlobalSecureOrderCreator()
	if secureCreator != nil {
		err = secureCreator.CancelExpiredOrder(order.No)
		if err != nil {
			log.Printf("取消订单失败: %v", err)
		}
	}

	return nil, nil

}
