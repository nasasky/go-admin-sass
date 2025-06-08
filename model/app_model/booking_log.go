package app_model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BookingStatusLog 订单状态管理日志
type BookingStatusLog struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	LogType    string             `bson:"log_type" json:"log_type"`       // 日志类型
	BookingID  *int               `bson:"booking_id" json:"booking_id"`   // 订单ID
	BookingNo  string             `bson:"booking_no" json:"booking_no"`   // 订单号
	RoomID     *int               `bson:"room_id" json:"room_id"`         // 房间ID
	RoomName   string             `bson:"room_name" json:"room_name"`     // 房间名称
	UserID     *int               `bson:"user_id" json:"user_id"`         // 用户ID
	OldStatus  *int               `bson:"old_status" json:"old_status"`   // 原状态
	NewStatus  *int               `bson:"new_status" json:"new_status"`   // 新状态
	Message    string             `bson:"message" json:"message"`         // 操作信息
	ErrorMsg   string             `bson:"error_msg" json:"error_msg"`     // 错误信息
	Details    interface{}        `bson:"details" json:"details"`         // 详细信息
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`   // 创建时间
	ServerInfo ServerInfo         `bson:"server_info" json:"server_info"` // 服务器信息
}

// ServerInfo 服务器信息
type ServerInfo struct {
	Hostname string `bson:"hostname" json:"hostname"` // 主机名
	PID      int    `bson:"pid" json:"pid"`           // 进程ID
	Mode     string `bson:"mode" json:"mode"`         // 运行模式
}

// 日志类型常量
const (
	LogTypeSchedulerStart  = "scheduler_start"  // 调度器启动
	LogTypeBookingActivate = "booking_activate" // 订单激活
	LogTypeBookingComplete = "booking_complete" // 订单完成
	LogTypeBookingTimeout  = "booking_timeout"  // 订单超时取消
	LogTypeBookingError    = "booking_error"    // 订单状态更新失败
	LogTypeRoomError       = "room_error"       // 房间状态更新失败
	LogTypeManualStart     = "manual_start"     // 手动开始订单
	LogTypeManualEnd       = "manual_end"       // 手动结束订单
	LogTypeUsageLogError   = "usage_log_error"  // 使用记录错误
)

// GetLogTypeText 获取日志类型文本
func GetLogTypeText(logType string) string {
	switch logType {
	case LogTypeSchedulerStart:
		return "调度器启动"
	case LogTypeBookingActivate:
		return "订单激活"
	case LogTypeBookingComplete:
		return "订单完成"
	case LogTypeBookingTimeout:
		return "订单超时取消"
	case LogTypeBookingError:
		return "订单状态更新失败"
	case LogTypeRoomError:
		return "房间状态更新失败"
	case LogTypeManualStart:
		return "手动开始订单"
	case LogTypeManualEnd:
		return "手动结束订单"
	case LogTypeUsageLogError:
		return "使用记录错误"
	default:
		return "未知类型"
	}
}

// GetStatusText 获取状态文本
func GetStatusText(status int) string {
	switch status {
	case BookingStatusPending:
		return "待支付"
	case BookingStatusPaid:
		return "已支付"
	case BookingStatusInUse:
		return "使用中"
	case BookingStatusCompleted:
		return "已完成"
	case BookingStatusCancelled:
		return "已取消"
	case BookingStatusRefunded:
		return "已退款"
	default:
		return "未知状态"
	}
}

// TableName MongoDB集合名称
func (BookingStatusLog) TableName() string {
	return "booking_status_logs"
}
