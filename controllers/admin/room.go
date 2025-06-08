package admin

import (
	"strconv"

	"nasa-go-admin/inout"
	"nasa-go-admin/services/app_service"

	"github.com/gin-gonic/gin"
)

var adminRoomService = &app_service.RoomService{}
var bookingScheduler = app_service.NewBookingScheduler()
var bookingLogService = &app_service.BookingLogService{}

// ========== 房间管理相关接口 ==========

// CreateRoom 创建房间
func CreateRoom(c *gin.Context) {
	var req inout.CreateRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Resp.Err(c, 20001, "参数错误: "+err.Error())
		return
	}

	// 获取创建人ID
	createdBy, exists := c.Get("uid")
	if !exists {
		Resp.Err(c, 10002, "用户未登录")
		return
	}

	uid, ok := createdBy.(int)
	if !ok {
		Resp.Err(c, 10002, "用户信息错误")
		return
	}

	room, err := adminRoomService.CreateRoom(&req, uid)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, room)
}

// UpdateRoom 更新房间信息
func UpdateRoom(c *gin.Context) {
	var req inout.UpdateRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Resp.Err(c, 20001, "参数错误: "+err.Error())
		return
	}

	room, err := adminRoomService.UpdateRoom(&req)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, room)
}

// GetAdminRoomList 获取房间列表（管理后台）
func GetAdminRoomList(c *gin.Context) {
	var req inout.RoomListReq

	// 设置默认值
	req.Page = 1
	req.PageSize = 10

	if err := c.ShouldBindQuery(&req); err != nil {
		Resp.Err(c, 20001, "参数错误: "+err.Error())
		return
	}

	// 参数验证
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}

	resp, err := adminRoomService.GetRoomList(&req)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, resp)
}

// GetAdminRoomDetail 获取房间详情（管理后台）
func GetAdminRoomDetail(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		Resp.Err(c, 20001, "房间ID不能为空")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		Resp.Err(c, 20001, "房间ID格式错误")
		return
	}

	detail, err := adminRoomService.GetRoomDetail(id)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, detail)
}

// UpdateRoomStatus 更新房间状态
func UpdateRoomStatus(c *gin.Context) {
	var req inout.UpdateRoomStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Resp.Err(c, 20001, "参数错误: "+err.Error())
		return
	}

	if err := adminRoomService.UpdateRoomStatus(&req); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, gin.H{"message": "房间状态更新成功"})
}

// DeleteRoom 删除房间
func DeleteRoom(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		Resp.Err(c, 20001, "房间ID不能为空")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		Resp.Err(c, 20001, "房间ID格式错误")
		return
	}

	if err := adminRoomService.DeleteRoom(id); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, gin.H{"message": "房间删除成功"})
}

// ========== 预订管理相关接口 ==========

// GetAdminBookingList 获取预订列表（管理后台）
func GetAdminBookingList(c *gin.Context) {
	var req inout.BookingListReq

	// 设置默认值
	req.Page = 1
	req.PageSize = 10

	if err := c.ShouldBindQuery(&req); err != nil {
		Resp.Err(c, 20001, "参数错误: "+err.Error())
		return
	}

	// 参数验证
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}

	// 管理员可以查看所有预订，不限制用户ID
	resp, err := adminRoomService.GetBookingList(&req, nil)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, resp)
}

// UpdateBookingStatus 更新预订状态
func UpdateBookingStatus(c *gin.Context) {
	var req inout.UpdateBookingStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Resp.Err(c, 20001, "参数错误: "+err.Error())
		return
	}

	// 管理员可以更新任意预订状态，不限制用户ID
	if err := adminRoomService.CancelBooking(&inout.CancelBookingReq{
		ID:     req.ID,
		Reason: "管理员操作",
	}, nil); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, gin.H{"message": "预订状态更新成功"})
}

// GetRoomStatisticsAdmin 获取房间统计信息（管理后台）
func GetRoomStatisticsAdmin(c *gin.Context) {
	stats, err := adminRoomService.GetRoomStatistics()
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, stats)
}

// ========== 订单状态管理相关接口 ==========

// GetBookingStatusInfo 获取订单状态信息
func GetBookingStatusInfo(c *gin.Context) {
	var req inout.BookingStatusInfoReq
	if err := c.ShouldBindQuery(&req); err != nil {
		Resp.Err(c, 20001, "参数错误: "+err.Error())
		return
	}

	info, err := bookingScheduler.GetBookingStatusInfo(req.ID)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// 转换为响应格式
	resp := &inout.BookingStatusInfoResp{
		BookingID:       info.BookingID,
		BookingNo:       info.BookingNo,
		Status:          info.Status,
		StatusText:      info.StatusText,
		StartTime:       info.StartTime,
		EndTime:         info.EndTime,
		RoomID:          info.RoomID,
		RoomName:        info.RoomName,
		RoomStatus:      info.RoomStatus,
		RoomStatusText:  info.RoomStatusText,
		CurrentTime:     info.CurrentTime,
		CanStart:        info.CanStart,
		CanEnd:          info.CanEnd,
		ShouldAutoStart: info.ShouldAutoStart,
		ShouldAutoEnd:   info.ShouldAutoEnd,
	}

	Resp.Succ(c, resp)
}

// ManualStartBooking 手动开始订单
func ManualStartBooking(c *gin.Context) {
	var req inout.ManualStartBookingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Resp.Err(c, 20001, "参数错误: "+err.Error())
		return
	}

	// 获取管理员ID
	adminID, exists := c.Get("uid")
	if !exists {
		Resp.Err(c, 10002, "用户未登录")
		return
	}

	uid, ok := adminID.(int)
	if !ok {
		Resp.Err(c, 10002, "用户信息错误")
		return
	}

	if err := bookingScheduler.ManuallyStartBooking(req.ID, uid); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, gin.H{"message": "订单已手动开始"})
}

// ManualEndBooking 手动结束订单
func ManualEndBooking(c *gin.Context) {
	var req inout.ManualEndBookingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Resp.Err(c, 20001, "参数错误: "+err.Error())
		return
	}

	// 获取管理员ID
	adminID, exists := c.Get("uid")
	if !exists {
		Resp.Err(c, 10002, "用户未登录")
		return
	}

	uid, ok := adminID.(int)
	if !ok {
		Resp.Err(c, 10002, "用户信息错误")
		return
	}

	if err := bookingScheduler.ManuallyEndBooking(req.ID, uid); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, gin.H{"message": "订单已手动结束"})
}

// ========== 订单状态日志管理接口 ==========

// GetBookingLogList 获取订单状态日志列表
func GetBookingLogList(c *gin.Context) {
	var req app_service.BookingLogListReq

	// 设置默认值
	req.Page = 1
	req.PageSize = 10

	if err := c.ShouldBindQuery(&req); err != nil {
		Resp.Err(c, 20001, "参数错误: "+err.Error())
		return
	}

	// 参数验证
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}

	resp, err := bookingLogService.GetLogList(&req)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, resp)
}

// GetBookingLogStatistics 获取订单状态日志统计
func GetBookingLogStatistics(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	stats, err := bookingLogService.GetLogStatistics(startDate, endDate)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	Resp.Succ(c, stats)
}
