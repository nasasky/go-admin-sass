package app

import (
	"strconv"

	"nasa-go-admin/api"
	"nasa-go-admin/inout"
	"nasa-go-admin/services/app_service"

	"github.com/gin-gonic/gin"
)

var roomService = &app_service.RoomService{}

// ========== 房间管理相关接口 ==========

// GetRoomList 获取房间列表
func GetRoomList(c *gin.Context) {
	var req inout.RoomListReq

	// 设置默认值
	req.Page = 1
	req.PageSize = 10

	if err := c.ShouldBindQuery(&req); err != nil {
		api.Resp.Err(c, 20001, "参数错误: "+err.Error())
		return
	}

	// 参数验证
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}

	resp, err := roomService.GetRoomList(&req)
	if err != nil {
		api.Resp.Err(c, 20001, err.Error())
		return
	}

	api.Resp.Succ(c, resp)
}

// GetRoomDetail 获取房间详情
func GetRoomDetail(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		api.Resp.Err(c, 20001, "房间ID不能为空")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		api.Resp.Err(c, 20001, "房间ID格式错误")
		return
	}

	detail, err := roomService.GetRoomDetail(id)
	if err != nil {
		api.Resp.Err(c, 20001, err.Error())
		return
	}

	api.Resp.Succ(c, detail)
}

// CheckRoomAvailability 检查房间可用性
func CheckRoomAvailability(c *gin.Context) {
	var req inout.CheckAvailabilityReq
	if err := c.ShouldBindJSON(&req); err != nil {
		api.Resp.Err(c, 20001, "参数错误: "+err.Error())
		return
	}

	resp, err := roomService.CheckAvailability(&req)
	if err != nil {
		api.Resp.Err(c, 20001, err.Error())
		return
	}

	api.Resp.Succ(c, resp)
}

// GetRoomStatistics 获取房间统计信息
func GetRoomStatistics(c *gin.Context) {
	stats, err := roomService.GetRoomStatistics()
	if err != nil {
		api.Resp.Err(c, 20001, err.Error())
		return
	}

	api.Resp.Succ(c, stats)
}

// ========== 预订管理相关接口 ==========

// CreateBooking 创建预订
func CreateBooking(c *gin.Context) {
	var req inout.CreateBookingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		api.Resp.Err(c, 20001, "参数错误: "+err.Error())
		return
	}

	// 获取用户ID
	userID, exists := c.Get("uid")
	if !exists {
		api.Resp.Err(c, 10002, "用户未登录")
		return
	}

	uid, ok := userID.(int)
	if !ok {
		api.Resp.Err(c, 10002, "用户信息错误")
		return
	}

	booking, err := roomService.CreateBooking(&req, uid)
	if err != nil {
		api.Resp.Err(c, 20001, err.Error())
		return
	}

	api.Resp.Succ(c, booking)
}

// GetMyBookingList 获取我的预订列表
func GetMyBookingList(c *gin.Context) {
	var req inout.BookingListReq

	// 设置默认值
	req.Page = 1
	req.PageSize = 10

	if err := c.ShouldBindQuery(&req); err != nil {
		api.Resp.Err(c, 20001, "参数错误: "+err.Error())
		return
	}

	// 参数验证
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 10
	}

	// 获取用户ID
	userID, exists := c.Get("uid")
	if !exists {
		api.Resp.Err(c, 10002, "用户未登录")
		return
	}

	uid, ok := userID.(int)
	if !ok {
		api.Resp.Err(c, 10002, "用户信息错误")
		return
	}

	resp, err := roomService.GetBookingList(&req, &uid)
	if err != nil {
		api.Resp.Err(c, 20001, err.Error())
		return
	}

	api.Resp.Succ(c, resp)
}

// GetBookingDetail 获取预订详情
func GetBookingDetail(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		api.Resp.Err(c, 20001, "预订ID不能为空")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		api.Resp.Err(c, 20001, "预订ID格式错误")
		return
	}

	// 获取用户ID
	userID, exists := c.Get("uid")
	if !exists {
		api.Resp.Err(c, 10002, "用户未登录")
		return
	}

	uid, ok := userID.(int)
	if !ok {
		api.Resp.Err(c, 10002, "用户信息错误")
		return
	}

	// 构造查询请求
	req := &inout.BookingListReq{
		Page:     1,
		PageSize: 1,
	}

	resp, err := roomService.GetBookingList(req, &uid)
	if err != nil {
		api.Resp.Err(c, 20001, err.Error())
		return
	}

	// 查找指定的预订
	var booking *inout.BookingDetail
	for _, b := range resp.List {
		if b.ID == id {
			booking = &b
			break
		}
	}

	if booking == nil {
		api.Resp.Err(c, 20001, "预订不存在或无权限访问")
		return
	}

	api.Resp.Succ(c, booking)
}

// CancelBooking 取消预订
func CancelBooking(c *gin.Context) {
	var req inout.CancelBookingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		api.Resp.Err(c, 20001, "参数错误: "+err.Error())
		return
	}

	// 获取用户ID
	userID, exists := c.Get("uid")
	if !exists {
		api.Resp.Err(c, 10002, "用户未登录")
		return
	}

	uid, ok := userID.(int)
	if !ok {
		api.Resp.Err(c, 10002, "用户信息错误")
		return
	}

	if err := roomService.CancelBooking(&req, &uid); err != nil {
		api.Resp.Err(c, 20001, err.Error())
		return
	}

	api.Resp.Succ(c, gin.H{"message": "预订已取消"})
}

// ========== 套餐相关接口 ==========

// GetRoomPackages 获取房间可用套餐
func GetRoomPackages(c *gin.Context) {
	var req inout.GetRoomPackagesReq
	if err := c.ShouldBindQuery(&req); err != nil {
		api.Resp.Err(c, 20001, "参数错误: "+err.Error())
		return
	}

	resp, err := roomService.GetRoomPackages(&req)
	if err != nil {
		api.Resp.Err(c, 20001, err.Error())
		return
	}

	api.Resp.Succ(c, resp)
}

// BookingPricePreview 预订价格预览
func BookingPricePreview(c *gin.Context) {
	var req inout.BookingPricePreviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		api.Resp.Err(c, 20001, "参数错误: "+err.Error())
		return
	}

	resp, err := roomService.BookingPricePreview(&req)
	if err != nil {
		api.Resp.Err(c, 20001, err.Error())
		return
	}

	api.Resp.Succ(c, resp)
}
