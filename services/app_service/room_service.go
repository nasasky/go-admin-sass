package app_service

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"nasa-go-admin/db"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/app_model"

	"gorm.io/gorm"
)

type RoomService struct{}

// ========== 房间管理服务 ==========

// CreateRoom 创建房间
func (rs *RoomService) CreateRoom(req *inout.CreateRoomReq, createdBy int) (*app_model.Room, error) {
	// 检查房间号是否已存在
	var existingRoom app_model.Room
	if err := db.Dao.Where("room_number = ?", req.RoomNumber).First(&existingRoom).Error; err == nil {
		return nil, fmt.Errorf("房间号 %s 已存在", req.RoomNumber)
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("检查房间号失败: %v", err)
	}

	room := &app_model.Room{
		RoomNumber:  req.RoomNumber,
		RoomName:    req.RoomName,
		RoomType:    req.RoomType,
		Capacity:    req.Capacity,
		HourlyRate:  req.HourlyRate,
		Features:    req.Features,
		Images:      req.Images,
		Status:      app_model.RoomStatusAvailable,
		Floor:       req.Floor,
		Area:        req.Area,
		Description: req.Description,
		CreatedBy:   createdBy,
	}

	if err := db.Dao.Create(room).Error; err != nil {
		return nil, fmt.Errorf("创建房间失败: %v", err)
	}

	log.Printf("成功创建房间: %s (ID: %d)", room.RoomName, room.ID)
	return room, nil
}

// UpdateRoom 更新房间信息
func (rs *RoomService) UpdateRoom(req *inout.UpdateRoomReq) (*app_model.Room, error) {
	var room app_model.Room
	if err := db.Dao.First(&room, req.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("房间不存在")
		}
		return nil, fmt.Errorf("查询房间失败: %v", err)
	}

	// 检查房间号是否被其他房间占用
	var existingRoom app_model.Room
	if err := db.Dao.Where("room_number = ? AND id != ?", req.RoomNumber, req.ID).First(&existingRoom).Error; err == nil {
		return nil, fmt.Errorf("房间号 %s 已被其他房间使用", req.RoomNumber)
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("检查房间号失败: %v", err)
	}

	// 更新房间信息
	updates := map[string]interface{}{
		"room_number": req.RoomNumber,
		"room_name":   req.RoomName,
		"room_type":   req.RoomType,
		"capacity":    req.Capacity,
		"hourly_rate": req.HourlyRate,
		"features":    req.Features,
		"images":      req.Images,
		"floor":       req.Floor,
		"area":        req.Area,
		"description": req.Description,
		"status":      req.Status,
	}

	if err := db.Dao.Model(&room).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新房间失败: %v", err)
	}

	log.Printf("成功更新房间: %s (ID: %d)", room.RoomName, room.ID)
	return &room, nil
}

// GetRoomList 获取房间列表
func (rs *RoomService) GetRoomList(req *inout.RoomListReq) (*inout.RoomListResp, error) {
	var rooms []app_model.Room
	var total int64

	// 构建查询条件
	query := db.Dao.Model(&app_model.Room{})

	if req.RoomType != "" {
		query = query.Where("room_type = ?", req.RoomType)
	}
	if req.Status > 0 {
		query = query.Where("status = ?", req.Status)
	}
	if req.Floor > 0 {
		query = query.Where("floor = ?", req.Floor)
	}
	if req.Keyword != "" {
		query = query.Where("room_name LIKE ? OR room_number LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.MinPrice > 0 {
		query = query.Where("hourly_rate >= ?", req.MinPrice)
	}
	if req.MaxPrice > 0 {
		query = query.Where("hourly_rate <= ?", req.MaxPrice)
	}
	if req.MinCapacity > 0 {
		query = query.Where("capacity >= ?", req.MinCapacity)
	}
	if req.MaxCapacity > 0 {
		query = query.Where("capacity <= ?", req.MaxCapacity)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("查询房间总数失败: %v", err)
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("id DESC").Find(&rooms).Error; err != nil {
		return nil, fmt.Errorf("查询房间列表失败: %v", err)
	}

	// 转换为响应格式
	list := make([]inout.RoomDetail, 0, len(rooms))
	for _, room := range rooms {
		detail := rs.convertRoomToDetail(&room)
		list = append(list, *detail)
	}

	return &inout.RoomListResp{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		List:     list,
	}, nil
}

// GetRoomDetail 获取房间详情
func (rs *RoomService) GetRoomDetail(id int) (*inout.RoomDetail, error) {
	var room app_model.Room
	if err := db.Dao.First(&room, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("房间不存在")
		}
		return nil, fmt.Errorf("查询房间失败: %v", err)
	}

	detail := rs.convertRoomToDetail(&room)

	// 查询当前预订信息
	var currentBooking app_model.RoomBooking
	now := time.Now()
	if err := db.Dao.Where("room_id = ? AND start_time <= ? AND end_time >= ? AND status IN (?)",
		id, now, now, []int{app_model.BookingStatusPaid, app_model.BookingStatusInUse}).
		First(&currentBooking).Error; err == nil {

		bookingDetail := rs.convertBookingToDetail(&currentBooking)
		detail.CurrentBooking = bookingDetail
	}

	return detail, nil
}

// UpdateRoomStatus 更新房间状态
func (rs *RoomService) UpdateRoomStatus(req *inout.UpdateRoomStatusReq) error {
	var room app_model.Room
	if err := db.Dao.First(&room, req.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("房间不存在")
		}
		return fmt.Errorf("查询房间失败: %v", err)
	}

	if err := db.Dao.Model(&room).Update("status", req.Status).Error; err != nil {
		return fmt.Errorf("更新房间状态失败: %v", err)
	}

	log.Printf("房间 %s 状态更新为: %d", room.RoomName, req.Status)
	return nil
}

// DeleteRoom 删除房间
func (rs *RoomService) DeleteRoom(id int) error {
	// 检查是否有未完成的预订
	var count int64
	if err := db.Dao.Model(&app_model.RoomBooking{}).
		Where("room_id = ? AND status IN (?)", id,
			[]int{app_model.BookingStatusPending, app_model.BookingStatusPaid, app_model.BookingStatusInUse}).
		Count(&count).Error; err != nil {
		return fmt.Errorf("检查房间预订失败: %v", err)
	}

	if count > 0 {
		return fmt.Errorf("房间有未完成的预订，无法删除")
	}

	if err := db.Dao.Delete(&app_model.Room{}, id).Error; err != nil {
		return fmt.Errorf("删除房间失败: %v", err)
	}

	log.Printf("成功删除房间 ID: %d", id)
	return nil
}

// ========== 预订管理服务 ==========

// CreateBooking 创建预订
func (rs *RoomService) CreateBooking(req *inout.CreateBookingReq, userID int) (*app_model.RoomBooking, error) {
	// 解析开始时间
	startTime, err := time.Parse("2006-01-02 15:04:05", req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("开始时间格式错误: %v", err)
	}

	// 计算结束时间
	endTime := startTime.Add(time.Duration(req.Hours) * time.Hour)

	// 检查房间是否存在且可用
	var room app_model.Room
	if err := db.Dao.First(&room, req.RoomID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("房间不存在")
		}
		return nil, fmt.Errorf("查询房间失败: %v", err)
	}

	if !room.IsAvailable() {
		return nil, fmt.Errorf("房间当前不可用")
	}

	// 检查时间段是否可用
	if available, err := rs.CheckRoomAvailability(req.RoomID, startTime, endTime); err != nil {
		return nil, err
	} else if !available {
		return nil, fmt.Errorf("该时间段房间已被预订")
	}

	// 计算价格
	var totalAmount, originalPrice, packagePrice, discountAmount float64
	var packageName, priceBreakdown string

	originalPrice = room.HourlyRate * float64(req.Hours)

	if req.PackageID != nil {
		// 使用套餐价格
		var pkg app_model.RoomPackage
		if err := db.Dao.Preload("Rules").First(&pkg, *req.PackageID).Error; err != nil {
			return nil, fmt.Errorf("套餐不存在")
		}

		// 验证套餐是否属于该房间
		if pkg.RoomID != req.RoomID {
			return nil, fmt.Errorf("套餐不属于该房间")
		}

		// 计算套餐价格
		finalPrice, matchedRule, err := pkg.CalculatePrice(room.HourlyRate, startTime, req.Hours)
		if err != nil {
			return nil, fmt.Errorf("套餐价格计算失败: %v", err)
		}

		packagePrice = finalPrice
		totalAmount = finalPrice
		packageName = pkg.PackageName
		discountAmount = originalPrice - finalPrice

		// 构建价格明细
		breakdown := map[string]interface{}{
			"base_price":      room.HourlyRate,
			"hours":           req.Hours,
			"original_price":  originalPrice,
			"package_name":    pkg.PackageName,
			"package_type":    pkg.PackageType,
			"final_price":     finalPrice,
			"discount_amount": discountAmount,
		}
		if matchedRule != nil {
			breakdown["rule_name"] = matchedRule.RuleName
			breakdown["rule_type"] = matchedRule.PriceType
			breakdown["rule_value"] = matchedRule.PriceValue
		}

		breakdownBytes, _ := json.Marshal(breakdown)
		priceBreakdown = string(breakdownBytes)
	} else {
		// 使用基础价格
		totalAmount = originalPrice
		packagePrice = 0
	}

	// 生成预订号
	bookingNo := rs.generateBookingNo()

	booking := &app_model.RoomBooking{
		RoomID:         req.RoomID,
		UserID:         userID,
		BookingNo:      bookingNo,
		StartTime:      startTime,
		EndTime:        endTime,
		Hours:          req.Hours,
		TotalAmount:    totalAmount,
		Status:         app_model.BookingStatusPending,
		ContactName:    req.ContactName,
		ContactPhone:   req.ContactPhone,
		Remarks:        req.Remarks,
		PackageID:      req.PackageID,
		PackageName:    packageName,
		OriginalPrice:  originalPrice,
		PackagePrice:   packagePrice,
		DiscountAmount: discountAmount,
		PriceBreakdown: priceBreakdown,
	}

	if err := db.Dao.Create(booking).Error; err != nil {
		return nil, fmt.Errorf("创建预订失败: %v", err)
	}

	log.Printf("成功创建预订: %s (用户ID: %d, 房间ID: %d)", bookingNo, userID, req.RoomID)
	return booking, nil
}

// CheckRoomAvailability 检查房间可用性
func (rs *RoomService) CheckRoomAvailability(roomID int, startTime, endTime time.Time) (bool, error) {
	var count int64
	if err := db.Dao.Model(&app_model.RoomBooking{}).
		Where("room_id = ? AND status IN (?) AND ((start_time <= ? AND end_time > ?) OR (start_time < ? AND end_time >= ?))",
			roomID,
			[]int{app_model.BookingStatusPaid, app_model.BookingStatusInUse},
			startTime, startTime, endTime, endTime).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("查询房间预订失败: %v", err)
	}

	return count == 0, nil
}

// CheckAvailability 检查可用性接口（增强版，支持套餐定价）
func (rs *RoomService) CheckAvailability(req *inout.CheckAvailabilityReq) (*inout.AvailabilityResp, error) {
	startTime, err := time.Parse("2006-01-02 15:04:05", req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("开始时间格式错误: %v", err)
	}

	endTime := startTime.Add(time.Duration(req.Hours) * time.Hour)

	// 获取房间信息
	var room app_model.Room
	if err := db.Dao.First(&room, req.RoomID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &inout.AvailabilityResp{
				IsAvailable: false,
				Message:     "房间不存在",
			}, nil
		}
		return nil, fmt.Errorf("查询房间失败: %v", err)
	}

	if !room.IsAvailable() {
		return &inout.AvailabilityResp{
			IsAvailable: false,
			Message:     "房间当前不可用",
		}, nil
	}

	available, err := rs.CheckRoomAvailability(req.RoomID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// 计算价格（使用套餐规则）
	basePrice := room.HourlyRate
	duration := float64(req.Hours)

	finalPrice, _, priceErr := rs.CalculatePriceWithPackage(req.RoomID, startTime, endTime)
	if priceErr != nil {
		return nil, fmt.Errorf("计算套餐价格失败: %v", priceErr)
	}

	// 获取日期类型
	dayType := app_model.GetDayType(startTime)
	dayTypeText := rs.getDayTypeText(dayType)

	// 构建价格详情
	priceBreakdown := &inout.PriceBreakdown{
		BaseHourlyRate: basePrice,
		Hours:          req.Hours,
		SubTotal:       basePrice * duration,
		RuleType:       "package",
		RuleValue:      finalPrice - (basePrice * duration),
		Adjustment:     finalPrice - (basePrice * duration),
		FinalTotal:     finalPrice,
	}

	resp := &inout.AvailabilityResp{
		IsAvailable:    available,
		TotalAmount:    finalPrice, // 保持向后兼容
		BasePrice:      basePrice * duration,
		FinalPrice:     finalPrice,
		DayType:        dayType,
		DayTypeText:    dayTypeText,
		PriceBreakdown: priceBreakdown,
	}

	if available {
		resp.Message = "房间可预订"
	} else {
		resp.Message = "该时间段房间已被预订"
	}

	return resp, nil
}

// GetBookingList 获取预订列表
func (rs *RoomService) GetBookingList(req *inout.BookingListReq, userID *int) (*inout.BookingListResp, error) {
	var bookings []app_model.RoomBooking
	var total int64

	query := db.Dao.Model(&app_model.RoomBooking{}).Preload("Room").Preload("User")

	// 如果指定了用户ID，只查询该用户的预订
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	if req.Status > 0 {
		query = query.Where("status = ?", req.Status)
	}
	if req.RoomID > 0 {
		query = query.Where("room_id = ?", req.RoomID)
	}
	if req.UserID > 0 && userID == nil { // 管理员可以查询指定用户
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.StartDate != "" {
		query = query.Where("start_time >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		query = query.Where("start_time <= ?", req.EndDate+" 23:59:59")
	}
	if req.Keyword != "" {
		query = query.Where("booking_no LIKE ? OR contact_name LIKE ? OR contact_phone LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("查询预订总数失败: %v", err)
	}

	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("id DESC").Find(&bookings).Error; err != nil {
		return nil, fmt.Errorf("查询预订列表失败: %v", err)
	}

	list := make([]inout.BookingDetail, 0, len(bookings))
	for _, booking := range bookings {
		detail := rs.convertBookingToDetail(&booking)
		list = append(list, *detail)
	}

	return &inout.BookingListResp{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		List:     list,
	}, nil
}

// CancelBooking 取消预订
func (rs *RoomService) CancelBooking(req *inout.CancelBookingReq, userID *int) error {
	var booking app_model.RoomBooking
	query := db.Dao

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	if err := query.First(&booking, req.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("预订不存在")
		}
		return fmt.Errorf("查询预订失败: %v", err)
	}

	if booking.Status != app_model.BookingStatusPending && booking.Status != app_model.BookingStatusPaid {
		return fmt.Errorf("当前状态无法取消")
	}

	// 更新预订状态
	updates := map[string]interface{}{
		"status": app_model.BookingStatusCancelled,
	}
	if req.Reason != "" {
		updates["remarks"] = booking.Remarks + "\n取消原因: " + req.Reason
	}

	if err := db.Dao.Model(&booking).Updates(updates).Error; err != nil {
		return fmt.Errorf("取消预订失败: %v", err)
	}

	log.Printf("预订已取消: %s (用户ID: %d)", booking.BookingNo, booking.UserID)
	return nil
}

// ========== 辅助方法 ==========

// convertRoomToDetail 转换房间模型为详情响应
func (rs *RoomService) convertRoomToDetail(room *app_model.Room) *inout.RoomDetail {
	detail := &inout.RoomDetail{
		ID:           room.ID,
		RoomNumber:   room.RoomNumber,
		RoomName:     room.RoomName,
		RoomType:     room.RoomType,
		RoomTypeText: rs.getRoomTypeText(room.RoomType),
		Capacity:     room.Capacity,
		HourlyRate:   room.HourlyRate,
		Status:       room.Status,
		StatusText:   room.GetStatusText(),
		Floor:        room.Floor,
		Area:         room.Area,
		Description:  room.Description,
		CreateTime:   room.CreateTime,
		UpdateTime:   room.UpdateTime,
		IsAvailable:  room.IsAvailable(),
	}

	// 解析特色设施
	if room.Features != "" {
		var features []string
		if err := json.Unmarshal([]byte(room.Features), &features); err == nil {
			detail.Features = features
		}
	}

	// 解析图片URLs
	if room.Images != "" {
		var images []string
		if err := json.Unmarshal([]byte(room.Images), &images); err == nil {
			detail.Images = images
		}
	}

	return detail
}

// convertBookingToDetail 转换预订模型为详情响应
func (rs *RoomService) convertBookingToDetail(booking *app_model.RoomBooking) *inout.BookingDetail {
	detail := &inout.BookingDetail{
		ID:             booking.ID,
		BookingNo:      booking.BookingNo,
		RoomID:         booking.RoomID,
		UserID:         booking.UserID,
		StartTime:      booking.StartTime,
		EndTime:        booking.EndTime,
		Hours:          booking.Hours,
		TotalAmount:    booking.TotalAmount,
		PaidAmount:     booking.PaidAmount,
		Status:         booking.Status,
		StatusText:     booking.GetBookingStatusText(),
		ContactName:    booking.ContactName,
		ContactPhone:   booking.ContactPhone,
		Remarks:        booking.Remarks,
		PackageID:      booking.PackageID,
		PackageName:    booking.PackageName,
		OriginalPrice:  booking.OriginalPrice,
		PackagePrice:   booking.PackagePrice,
		DiscountAmount: booking.DiscountAmount,
		PriceBreakdown: booking.PriceBreakdown,
		CreateTime:     booking.CreateTime,
		UpdateTime:     booking.UpdateTime,
	}

	// 关联房间信息
	if booking.Room != nil {
		detail.Room = rs.convertRoomToDetail(booking.Room)
	}

	// 关联用户信息
	if booking.User != nil {
		detail.UserInfo = &inout.UserInfo{
			ID:       booking.User.ID,
			Username: booking.User.Username,
			Phone:    booking.User.Phone,
			Nickname: booking.User.Nickname,
			Avatar:   booking.User.Avatar,
		}
	}

	// 关联套餐信息
	if booking.Package != nil {
		detail.PackageInfo = rs.convertPackageToDetail(booking.Package)
	}

	return detail
}

// getRoomTypeText 获取房间类型文本
func (rs *RoomService) getRoomTypeText(roomType string) string {
	switch roomType {
	case app_model.RoomTypeSmall:
		return "小包厢"
	case app_model.RoomTypeMedium:
		return "中包厢"
	case app_model.RoomTypeLarge:
		return "大包厢"
	case app_model.RoomTypeLuxury:
		return "豪华包厢"
	default:
		return "未知"
	}
}

// generateBookingNo 生成预订号
func (rs *RoomService) generateBookingNo() string {
	timestamp := time.Now().Format("20060102150405")
	random := strconv.FormatInt(time.Now().UnixNano()%1000000, 10)
	for len(random) < 6 {
		random = "0" + random
	}
	return "BK" + timestamp + random
}

// GetRoomStatistics 获取房间统计信息
func (rs *RoomService) GetRoomStatistics() (*inout.RoomStatistics, error) {
	stats := &inout.RoomStatistics{}

	// 统计房间数量
	var totalRooms, availableRooms, occupiedRooms, maintenanceRooms, disabledRooms int64

	db.Dao.Model(&app_model.Room{}).Count(&totalRooms)
	db.Dao.Model(&app_model.Room{}).Where("status = ?", app_model.RoomStatusAvailable).Count(&availableRooms)
	db.Dao.Model(&app_model.Room{}).Where("status = ?", app_model.RoomStatusOccupied).Count(&occupiedRooms)
	db.Dao.Model(&app_model.Room{}).Where("status = ?", app_model.RoomStatusMaintenance).Count(&maintenanceRooms)
	db.Dao.Model(&app_model.Room{}).Where("status = ?", app_model.RoomStatusDisabled).Count(&disabledRooms)

	stats.TotalRooms = int(totalRooms)
	stats.AvailableRooms = int(availableRooms)
	stats.OccupiedRooms = int(occupiedRooms)
	stats.MaintenanceRooms = int(maintenanceRooms)
	stats.DisabledRooms = int(disabledRooms)

	// 计算入住率
	if totalRooms > 0 {
		stats.OccupancyRate = float64(occupiedRooms) / float64(totalRooms) * 100
	}

	// 今日预订数
	today := time.Now().Format("2006-01-02")
	var todayBookings int64
	db.Dao.Model(&app_model.RoomBooking{}).
		Where("DATE(create_time) = ?", today).
		Count(&todayBookings)
	stats.TodayBookings = int(todayBookings)

	// 今日营收
	var todayRevenue float64
	db.Dao.Model(&app_model.RoomBooking{}).
		Where("DATE(create_time) = ? AND status IN (?)", today,
			[]int{app_model.BookingStatusPaid, app_model.BookingStatusInUse, app_model.BookingStatusCompleted}).
		Select("COALESCE(SUM(total_amount), 0)").Scan(&todayRevenue)
	stats.TodayRevenue = todayRevenue

	// 本月营收
	monthStart := time.Now().Format("2006-01-01")
	var monthlyRevenue float64
	db.Dao.Model(&app_model.RoomBooking{}).
		Where("create_time >= ? AND status IN (?)", monthStart,
			[]int{app_model.BookingStatusPaid, app_model.BookingStatusInUse, app_model.BookingStatusCompleted}).
		Select("COALESCE(SUM(total_amount), 0)").Scan(&monthlyRevenue)
	stats.MonthlyRevenue = monthlyRevenue

	return stats, nil
}

// ========== 套餐功能相关方法 ==========

// CalculatePriceWithPackage 使用套餐计算价格
func (rs *RoomService) CalculatePriceWithPackage(roomID int, startTime, endTime time.Time) (float64, []string, error) {
	// 获取房间基础价格
	var room app_model.Room
	if err := db.Dao.First(&room, roomID).Error; err != nil {
		return 0, nil, fmt.Errorf("房间不存在")
	}

	basePrice := room.HourlyRate
	duration := endTime.Sub(startTime).Hours()
	baseTotalPrice := basePrice * duration

	// 查询适用的套餐规则
	now := time.Now()
	var packages []app_model.RoomPackage
	if err := db.Dao.Where("room_id = ? AND is_active = 1 AND start_date <= ? AND end_date >= ?",
		roomID, now, now).
		Order("priority DESC").
		Find(&packages).Error; err != nil {
		log.Printf("查询套餐失败: %v", err)
		return baseTotalPrice, []string{"使用基础价格"}, nil
	}

	if len(packages) == 0 {
		return baseTotalPrice, []string{"使用基础价格"}, nil
	}

	// 按优先级应用套餐规则
	finalPrice := baseTotalPrice
	appliedRules := []string{}

	for _, pkg := range packages {
		// 查询套餐规则
		var rules []app_model.RoomPackageRule
		if err := db.Dao.Where("package_id = ? AND is_active = 1", pkg.ID).
			Order("priority DESC").
			Find(&rules).Error; err != nil {
			continue
		}

		dayType := rs.getDayType(startTime)

		for _, rule := range rules {
			// 检查日期类型匹配
			if rule.DayType != "all" && rule.DayType != dayType {
				continue
			}

			// 检查时间段匹配
			currentTimeStr := startTime.Format("15:04")
			if rule.TimeStart != "" && currentTimeStr < rule.TimeStart {
				continue
			}
			if rule.TimeEnd != "" && currentTimeStr >= rule.TimeEnd {
				continue
			}

			// 应用价格规则
			originalPrice := finalPrice
			switch rule.PriceType {
			case "fixed":
				finalPrice = rule.PriceValue * duration
			case "multiply":
				finalPrice = finalPrice * rule.PriceValue
			case "add":
				finalPrice = finalPrice + rule.PriceValue
			}

			appliedRules = append(appliedRules, fmt.Sprintf("%s: %s价格调整 (%.2f -> %.2f)",
				pkg.PackageName, rule.RuleName, originalPrice, finalPrice))
		}
	}

	if len(appliedRules) == 0 {
		appliedRules = []string{"使用基础价格"}
	}

	return finalPrice, appliedRules, nil
}

// getDayType 获取日期类型
func (rs *RoomService) getDayType(date time.Time) string {
	// 检查是否是特殊日期
	var specialDate app_model.RoomSpecialDate
	dateStr := date.Format("2006-01-02")
	if err := db.Dao.Where("date = ? AND is_active = 1", dateStr).First(&specialDate).Error; err == nil {
		return specialDate.DateType
	}

	// 检查是否是周末
	weekday := date.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return "weekend"
	}

	// 默认为工作日
	return "weekday"
}

// getDayTypeText 获取日期类型文本
func (rs *RoomService) getDayTypeText(dayType string) string {
	switch dayType {
	case "weekday":
		return "工作日"
	case "weekend":
		return "周末"
	case "holiday":
		return "节假日"
	case "special":
		return "特殊日期"
	default:
		return "全部"
	}
}

// ========== 套餐管理方法 ==========

// CreatePackage 创建套餐
func (rs *RoomService) CreatePackage(req *inout.CreatePackageReq, createdBy int) (*app_model.RoomPackage, error) {
	// 验证房间是否存在
	var room app_model.Room
	if err := db.Dao.First(&room, req.RoomID).Error; err != nil {
		return nil, fmt.Errorf("房间不存在")
	}

	// 解析日期字符串到时间类型
	var startDate, endDate *time.Time
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err != nil {
			return nil, fmt.Errorf("开始日期格式无效: %v", err)
		} else {
			startDate = &t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err != nil {
			return nil, fmt.Errorf("结束日期格式无效: %v", err)
		} else {
			endDate = &t
		}
	}

	// 验证日期范围
	if startDate != nil && endDate != nil && !endDate.After(*startDate) {
		return nil, fmt.Errorf("结束日期必须晚于开始日期")
	}

	// 创建套餐
	pkg := &app_model.RoomPackage{
		RoomID:      req.RoomID,
		PackageName: req.PackageName,
		Description: req.Description,
		PackageType: req.PackageType,
		FixedHours:  req.FixedHours,
		MinHours:    req.MinHours,
		MaxHours:    req.MaxHours,
		BasePrice:   req.BasePrice,
		StartDate:   startDate,
		EndDate:     endDate,
		Priority:    req.Priority,
		IsActive:    true, // 默认启用
	}

	if err := db.Dao.Create(pkg).Error; err != nil {
		return nil, fmt.Errorf("创建套餐失败: %v", err)
	}

	log.Printf("套餐已创建: %s (ID: %d)", pkg.PackageName, pkg.ID)
	return pkg, nil
}

// UpdatePackage 更新套餐
func (rs *RoomService) UpdatePackage(req *inout.UpdatePackageReq) (*app_model.RoomPackage, error) {
	var pkg app_model.RoomPackage
	if err := db.Dao.First(&pkg, req.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("套餐不存在")
		}
		return nil, fmt.Errorf("查询套餐失败: %v", err)
	}

	// 验证房间是否存在
	if req.RoomID != pkg.RoomID {
		var room app_model.Room
		if err := db.Dao.First(&room, req.RoomID).Error; err != nil {
			return nil, fmt.Errorf("房间不存在")
		}
		pkg.RoomID = req.RoomID
	}

	// 更新其他字段
	pkg.PackageName = req.PackageName
	pkg.Description = req.Description
	pkg.PackageType = req.PackageType
	pkg.FixedHours = req.FixedHours
	pkg.MinHours = req.MinHours
	pkg.MaxHours = req.MaxHours
	pkg.BasePrice = req.BasePrice
	pkg.Priority = req.Priority
	pkg.IsActive = req.IsActive

	// 解析和更新日期
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err != nil {
			return nil, fmt.Errorf("开始日期格式无效: %v", err)
		} else {
			pkg.StartDate = &t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err != nil {
			return nil, fmt.Errorf("结束日期格式无效: %v", err)
		} else {
			pkg.EndDate = &t
		}
	}

	// 验证日期范围
	if pkg.StartDate != nil && pkg.EndDate != nil && !pkg.EndDate.After(*pkg.StartDate) {
		return nil, fmt.Errorf("结束日期必须晚于开始日期")
	}

	if err := db.Dao.Save(&pkg).Error; err != nil {
		return nil, fmt.Errorf("更新套餐失败: %v", err)
	}

	log.Printf("套餐已更新: %s (ID: %d)", pkg.PackageName, pkg.ID)
	return &pkg, nil
}

// CreatePackageRule 创建套餐规则
func (rs *RoomService) CreatePackageRule(req *inout.CreatePackageRuleReq) (*app_model.RoomPackageRule, error) {
	// 验证套餐是否存在
	var pkg app_model.RoomPackage
	if err := db.Dao.First(&pkg, req.PackageID).Error; err != nil {
		return nil, fmt.Errorf("套餐不存在")
	}

	// 验证时间范围
	if req.TimeStart != "" && req.TimeEnd != "" {
		if req.TimeStart >= req.TimeEnd {
			return nil, fmt.Errorf("结束时间必须晚于开始时间")
		}
	}

	// 验证调整类型和值
	if req.PriceType != "fixed" && req.PriceType != "multiply" && req.PriceType != "add" {
		return nil, fmt.Errorf("价格类型必须是 fixed, multiply 或 add")
	}
	if req.PriceValue <= 0 {
		return nil, fmt.Errorf("价格值必须大于0")
	}

	// 创建规则
	rule := &app_model.RoomPackageRule{
		PackageID:  req.PackageID,
		RuleName:   req.RuleName,
		DayType:    req.DayType,
		TimeStart:  req.TimeStart,
		TimeEnd:    req.TimeEnd,
		PriceType:  req.PriceType,
		PriceValue: req.PriceValue,
		MinHours:   req.MinHours,
		MaxHours:   req.MaxHours,
		IsActive:   true, // 默认激活
	}

	if err := db.Dao.Create(rule).Error; err != nil {
		return nil, fmt.Errorf("创建套餐规则失败: %v", err)
	}

	log.Printf("套餐规则已创建: %s (ID: %d)", rule.RuleName, rule.ID)
	return rule, nil
}

// UpdatePackageRule 更新套餐规则
func (rs *RoomService) UpdatePackageRule(req *inout.UpdatePackageRuleReq) (*app_model.RoomPackageRule, error) {
	var rule app_model.RoomPackageRule
	if err := db.Dao.First(&rule, req.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("套餐规则不存在")
		}
		return nil, fmt.Errorf("查询套餐规则失败: %v", err)
	}

	// 更新字段
	rule.PackageID = req.PackageID
	rule.RuleName = req.RuleName
	rule.DayType = req.DayType
	rule.TimeStart = req.TimeStart
	rule.TimeEnd = req.TimeEnd
	rule.PriceType = req.PriceType
	rule.PriceValue = req.PriceValue
	rule.MinHours = req.MinHours
	rule.MaxHours = req.MaxHours
	rule.IsActive = req.IsActive

	// 验证时间范围
	if rule.TimeStart != "" && rule.TimeEnd != "" {
		if rule.TimeStart >= rule.TimeEnd {
			return nil, fmt.Errorf("结束时间必须晚于开始时间")
		}
	}

	// 验证价格类型和值
	if rule.PriceType != "fixed" && rule.PriceType != "multiply" && rule.PriceType != "add" {
		return nil, fmt.Errorf("价格类型必须是 fixed, multiply 或 add")
	}
	if rule.PriceValue <= 0 {
		return nil, fmt.Errorf("价格值必须大于0")
	}

	if err := db.Dao.Save(&rule).Error; err != nil {
		return nil, fmt.Errorf("更新套餐规则失败: %v", err)
	}

	log.Printf("套餐规则已更新: %s (ID: %d)", rule.RuleName, rule.ID)
	return &rule, nil
}

// GetPackageList 获取套餐列表
func (rs *RoomService) GetPackageList(req *inout.PackageListReq) (*inout.PackageListResp, error) {
	query := db.Dao.Model(&app_model.RoomPackage{})

	// 添加查询条件
	if req.RoomID != 0 {
		query = query.Where("room_id = ?", req.RoomID)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("查询套餐总数失败: %v", err)
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var packages []app_model.RoomPackage
	if err := query.Preload("Rules").
		Order("priority DESC, create_time DESC").
		Offset(offset).Limit(req.PageSize).
		Find(&packages).Error; err != nil {
		return nil, fmt.Errorf("查询套餐列表失败: %v", err)
	}

	// 转换数据
	list := make([]inout.PackageDetail, len(packages))
	for i, pkg := range packages {
		list[i] = *rs.convertPackageToDetail(&pkg)
	}

	return &inout.PackageListResp{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		List:     list,
	}, nil
}

// convertPackageToDetail 转换套餐模型为详情响应
func (rs *RoomService) convertPackageToDetail(pkg *app_model.RoomPackage) *inout.PackageDetail {
	detail := &inout.PackageDetail{
		ID:              pkg.ID,
		RoomID:          pkg.RoomID,
		PackageName:     pkg.PackageName,
		Description:     pkg.Description,
		PackageType:     pkg.PackageType,
		PackageTypeText: rs.getPackageTypeText(pkg.PackageType),
		FixedHours:      pkg.FixedHours,
		MinHours:        pkg.MinHours,
		MaxHours:        pkg.MaxHours,
		BasePrice:       pkg.BasePrice,
		StartDate:       pkg.StartDate,
		EndDate:         pkg.EndDate,
		Priority:        pkg.Priority,
		IsActive:        pkg.IsActive,
		CreateTime:      pkg.CreateTime,
		UpdateTime:      pkg.UpdateTime,
	}

	// 转换规则
	if len(pkg.Rules) > 0 {
		rules := make([]inout.PackageRuleDetail, len(pkg.Rules))
		for i, rule := range pkg.Rules {
			rules[i] = inout.PackageRuleDetail{
				ID:            rule.ID,
				PackageID:     rule.PackageID,
				RuleName:      rule.RuleName,
				DayType:       rule.DayType,
				DayTypeText:   rs.getDayTypeText(rule.DayType),
				TimeStart:     rule.TimeStart,
				TimeEnd:       rule.TimeEnd,
				PriceType:     rule.PriceType,
				PriceTypeText: rs.getPriceTypeText(rule.PriceType),
				PriceValue:    rule.PriceValue,
				MinHours:      rule.MinHours,
				MaxHours:      rule.MaxHours,
				IsActive:      rule.IsActive,
				CreateTime:    rule.CreateTime,
				UpdateTime:    rule.UpdateTime,
			}
		}
		detail.Rules = rules
	}

	return detail
}

// getAdjustmentText 获取调整类型文本说明
func (rs *RoomService) getAdjustmentText(adjustmentType string, adjustmentValue float64) string {
	switch adjustmentType {
	case "fixed":
		return fmt.Sprintf("固定价格 %.2f 元/小时", adjustmentValue)
	case "multiply":
		return fmt.Sprintf("价格倍数 %.2f", adjustmentValue)
	case "add":
		return fmt.Sprintf("增加 %.2f 元", adjustmentValue)
	default:
		return "未知调整类型"
	}
}

// DeletePackage 删除套餐
func (rs *RoomService) DeletePackage(id int) error {
	var pkg app_model.RoomPackage
	if err := db.Dao.First(&pkg, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("套餐不存在")
		}
		return fmt.Errorf("查询套餐失败: %v", err)
	}

	// 删除套餐及其关联的规则（通过外键约束自动删除）
	if err := db.Dao.Delete(&pkg).Error; err != nil {
		return fmt.Errorf("删除套餐失败: %v", err)
	}

	log.Printf("套餐已删除: %s (ID: %d)", pkg.PackageName, pkg.ID)
	return nil
}

// GetPackageRuleList 获取套餐规则列表
func (rs *RoomService) GetPackageRuleList(packageID int) ([]*inout.PackageRuleDetail, error) {
	// 验证套餐是否存在
	var pkg app_model.RoomPackage
	if err := db.Dao.First(&pkg, packageID).Error; err != nil {
		return nil, fmt.Errorf("套餐不存在")
	}

	var rules []app_model.RoomPackageRule
	if err := db.Dao.Where("package_id = ?", packageID).
		Order("priority DESC, created_at DESC").
		Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("查询套餐规则列表失败: %v", err)
	}

	// 转换数据
	result := make([]*inout.PackageRuleDetail, len(rules))
	for i, rule := range rules {
		result[i] = &inout.PackageRuleDetail{
			ID:            rule.ID,
			PackageID:     rule.PackageID,
			RuleName:      rule.RuleName,
			DayType:       rule.DayType,
			DayTypeText:   rs.getDayTypeText(rule.DayType),
			TimeStart:     rule.TimeStart,
			TimeEnd:       rule.TimeEnd,
			PriceType:     rule.PriceType,
			PriceTypeText: rs.getPriceTypeText(rule.PriceType),
			PriceValue:    rule.PriceValue,
			MinHours:      rule.MinHours,
			MaxHours:      rule.MaxHours,
			IsActive:      rule.IsActive,
			CreateTime:    rule.CreateTime,
			UpdateTime:    rule.UpdateTime,
		}
	}

	return result, nil
}

// DeletePackageRule 删除套餐规则
func (rs *RoomService) DeletePackageRule(id int) error {
	var rule app_model.RoomPackageRule
	if err := db.Dao.First(&rule, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("套餐规则不存在")
		}
		return fmt.Errorf("查询套餐规则失败: %v", err)
	}

	if err := db.Dao.Delete(&rule).Error; err != nil {
		return fmt.Errorf("删除套餐规则失败: %v", err)
	}

	log.Printf("套餐规则已删除: %s (ID: %d)", rule.RuleName, rule.ID)
	return nil
}

// GetSpecialDateList 获取特殊日期列表
func (rs *RoomService) GetSpecialDateList(page, pageSize int) (*inout.SpecialDateListResp, error) {
	var dates []app_model.RoomSpecialDate
	var total int64

	query := db.Dao.Model(&app_model.RoomSpecialDate{})

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("查询特殊日期总数失败: %v", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("date DESC").
		Offset(offset).Limit(pageSize).
		Find(&dates).Error; err != nil {
		return nil, fmt.Errorf("查询特殊日期列表失败: %v", err)
	}

	// 转换数据
	list := make([]*inout.SpecialDateDetail, len(dates))
	for i, date := range dates {
		list[i] = &inout.SpecialDateDetail{
			ID:           date.ID,
			Date:         date.Date,
			DateType:     date.DateType,
			DateTypeText: rs.getSpecialDateTypeText(date.DateType),
			Name:         date.Name,
			Description:  date.Description,
			IsActive:     date.IsActive,
			CreateTime:   date.CreateTime,
			UpdateTime:   date.UpdateTime,
		}
	}

	return &inout.SpecialDateListResp{
		Total:    int(total),
		Page:     page,
		PageSize: pageSize,
		List:     list,
	}, nil
}

// CreateSpecialDate 创建特殊日期
func (rs *RoomService) CreateSpecialDate(date time.Time, dateType, name, description string) (*app_model.RoomSpecialDate, error) {
	// 检查日期是否已存在
	var existing app_model.RoomSpecialDate
	if err := db.Dao.Where("date = ?", date.Format("2006-01-02")).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("该日期已存在特殊日期配置")
	}

	specialDate := &app_model.RoomSpecialDate{
		Date:        date,
		DateType:    dateType,
		Name:        name,
		Description: description,
		IsActive:    true,
	}

	if err := db.Dao.Create(specialDate).Error; err != nil {
		return nil, fmt.Errorf("创建特殊日期失败: %v", err)
	}

	log.Printf("特殊日期已创建: %s (%s)", specialDate.Name, specialDate.Date.Format("2006-01-02"))
	return specialDate, nil
}

// DeleteSpecialDate 删除特殊日期
func (rs *RoomService) DeleteSpecialDate(id int) error {
	var specialDate app_model.RoomSpecialDate
	if err := db.Dao.First(&specialDate, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("特殊日期不存在")
		}
		return fmt.Errorf("查询特殊日期失败: %v", err)
	}

	if err := db.Dao.Delete(&specialDate).Error; err != nil {
		return fmt.Errorf("删除特殊日期失败: %v", err)
	}

	log.Printf("特殊日期已删除: %s (%s)", specialDate.Name, specialDate.Date.Format("2006-01-02"))
	return nil
}

// getPriceTypeText 获取价格类型文本
func (rs *RoomService) getPriceTypeText(priceType string) string {
	switch priceType {
	case "fixed":
		return "固定价格"
	case "multiply":
		return "倍数调整"
	case "add":
		return "加价调整"
	default:
		return "未知类型"
	}
}

// getSpecialDateTypeText 获取特殊日期类型文本
func (rs *RoomService) getSpecialDateTypeText(dateType string) string {
	switch dateType {
	case "holiday":
		return "法定节假日"
	case "festival":
		return "传统节日"
	case "special":
		return "特殊活动日"
	default:
		return "未知类型"
	}
}

// getPackageTypeText 获取套餐类型文本
func (rs *RoomService) getPackageTypeText(packageType string) string {
	switch packageType {
	case "flexible":
		return "灵活时长套餐"
	case "fixed_hours":
		return "固定时长套餐"
	case "daily":
		return "全天套餐"
	case "weekly":
		return "周套餐"
	default:
		return "未知类型"
	}
}

// ========== 用户端套餐选择相关方法 ==========

// GetRoomPackages 获取房间可用套餐
func (rs *RoomService) GetRoomPackages(req *inout.GetRoomPackagesReq) (*inout.GetRoomPackagesResp, error) {
	// 解析预约时间
	startTime, err := time.Parse("2006-01-02 15:04:05", req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("开始时间格式错误: %v", err)
	}

	// 检查房间是否存在
	var room app_model.Room
	if err := db.Dao.First(&room, req.RoomID).Error; err != nil {
		return nil, fmt.Errorf("房间不存在")
	}

	// 获取日期类型
	dayType := rs.getDayType(startTime)
	dayTypeText := rs.getDayTypeText(dayType)

	// 查询房间的活跃套餐
	var packages []app_model.RoomPackage
	if err := db.Dao.Where("room_id = ? AND is_active = 1", req.RoomID).
		Preload("Rules", "is_active = 1").
		Order("priority DESC, create_time DESC").
		Find(&packages).Error; err != nil {
		return nil, fmt.Errorf("查询套餐失败: %v", err)
	}

	// 构建套餐选项
	options := make([]*inout.RoomPackageOption, 0)

	// 添加基础价格选项（无套餐）
	basePrice := room.HourlyRate * float64(req.Hours)
	baseOption := &inout.RoomPackageOption{
		PackageID:       0,
		PackageName:     "基础价格",
		Description:     "按小时计费，无优惠",
		PackageType:     "basic",
		PackageTypeText: "基础价格",
		FixedHours:      0,
		MinHours:        1,
		MaxHours:        24,
		BasePrice:       room.HourlyRate,
		FinalPrice:      basePrice,
		OriginalPrice:   basePrice,
		DiscountAmount:  0,
		DiscountPercent: 0,
		RuleName:        "",
		DayType:         dayType,
		DayTypeText:     dayTypeText,
		IsRecommended:   false,
		IsAvailable:     true,
	}
	options = append(options, baseOption)

	// 处理每个套餐
	for _, pkg := range packages {
		// 检查套餐时间范围
		if pkg.StartDate != nil && startTime.Before(*pkg.StartDate) {
			continue
		}
		if pkg.EndDate != nil && startTime.After(*pkg.EndDate) {
			continue
		}

		// 计算套餐价格
		finalPrice, matchedRule, err := pkg.CalculatePrice(room.HourlyRate, startTime, req.Hours)

		var unavailableMsg string
		isAvailable := true

		if err != nil {
			isAvailable = false
			unavailableMsg = err.Error()
		}

		// 计算原价和优惠
		var originalPrice, discountAmount, discountPercent float64
		var ruleName string

		if isAvailable {
			if pkg.PackageType == "fixed_hours" || pkg.PackageType == "daily" || pkg.PackageType == "weekly" {
				// 固定套餐：原价按时长计算
				originalPrice = room.HourlyRate * float64(pkg.FixedHours)
			} else {
				// 灵活套餐：原价按请求时长计算
				originalPrice = room.HourlyRate * float64(req.Hours)
			}

			discountAmount = originalPrice - finalPrice
			if originalPrice > 0 {
				discountPercent = discountAmount / originalPrice * 100
			}

			if matchedRule != nil {
				ruleName = matchedRule.RuleName
			}
		}

		option := &inout.RoomPackageOption{
			PackageID:       pkg.ID,
			PackageName:     pkg.PackageName,
			Description:     pkg.Description,
			PackageType:     pkg.PackageType,
			PackageTypeText: rs.getPackageTypeText(pkg.PackageType),
			FixedHours:      pkg.FixedHours,
			MinHours:        pkg.MinHours,
			MaxHours:        pkg.MaxHours,
			BasePrice:       pkg.BasePrice,
			FinalPrice:      finalPrice,
			OriginalPrice:   originalPrice,
			DiscountAmount:  discountAmount,
			DiscountPercent: discountPercent,
			RuleName:        ruleName,
			DayType:         dayType,
			DayTypeText:     dayTypeText,
			IsRecommended:   discountPercent > 10, // 优惠超过10%推荐
			IsAvailable:     isAvailable,
			UnavailableMsg:  unavailableMsg,
		}

		options = append(options, option)
	}

	return &inout.GetRoomPackagesResp{
		RoomID:      req.RoomID,
		StartTime:   req.StartTime,
		Hours:       req.Hours,
		DayType:     dayType,
		DayTypeText: dayTypeText,
		Packages:    options,
	}, nil
}

// BookingPricePreview 预订价格预览
func (rs *RoomService) BookingPricePreview(req *inout.BookingPricePreviewReq) (*inout.BookingPricePreviewResp, error) {
	// 解析预约时间
	startTime, err := time.Parse("2006-01-02 15:04:05", req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("开始时间格式错误: %v", err)
	}

	// 检查房间是否存在
	var room app_model.Room
	if err := db.Dao.First(&room, req.RoomID).Error; err != nil {
		return nil, fmt.Errorf("房间不存在")
	}

	// 获取日期类型
	dayType := rs.getDayType(startTime)
	dayTypeText := rs.getDayTypeText(dayType)

	// 计算基础价格
	basePrice := room.HourlyRate
	originalPrice := basePrice * float64(req.Hours)
	finalPrice := originalPrice
	discountAmount := 0.0
	discountPercent := 0.0
	var packageName, ruleName string

	// 如果指定了套餐，计算套餐价格
	if req.PackageID != nil {
		var pkg app_model.RoomPackage
		if err := db.Dao.Preload("Rules").First(&pkg, *req.PackageID).Error; err != nil {
			return nil, fmt.Errorf("套餐不存在")
		}

		// 验证套餐是否属于该房间
		if pkg.RoomID != req.RoomID {
			return nil, fmt.Errorf("套餐不属于该房间")
		}

		// 计算套餐价格
		calculatedPrice, matchedRule, err := pkg.CalculatePrice(basePrice, startTime, req.Hours)
		if err != nil {
			return nil, fmt.Errorf("套餐价格计算失败: %v", err)
		}

		finalPrice = calculatedPrice
		packageName = pkg.PackageName
		discountAmount = originalPrice - finalPrice
		if originalPrice > 0 {
			discountPercent = discountAmount / originalPrice * 100
		}

		if matchedRule != nil {
			ruleName = matchedRule.RuleName
		}
	}

	// 构建价格明细
	priceBreakdown := &inout.PriceBreakdown{
		BaseHourlyRate: basePrice,
		Hours:          req.Hours,
		SubTotal:       originalPrice,
		RuleType:       "package",
		RuleValue:      finalPrice - originalPrice,
		Adjustment:     discountAmount,
		FinalTotal:     finalPrice,
	}

	return &inout.BookingPricePreviewResp{
		RoomID:          req.RoomID,
		Hours:           req.Hours,
		BasePrice:       basePrice,
		PackageID:       req.PackageID,
		PackageName:     packageName,
		OriginalPrice:   originalPrice,
		FinalPrice:      finalPrice,
		DiscountAmount:  discountAmount,
		DiscountPercent: discountPercent,
		RuleName:        ruleName,
		DayType:         dayType,
		DayTypeText:     dayTypeText,
		PriceBreakdown:  priceBreakdown,
	}, nil
}
