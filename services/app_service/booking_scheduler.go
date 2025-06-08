package app_service

import (
	"fmt"
	"log"
	"os"
	"time"

	"nasa-go-admin/db"
	"nasa-go-admin/model/app_model"
)

type BookingScheduler struct {
	logService *BookingLogService
}

// NewBookingScheduler 创建新的订单调度器实例
func NewBookingScheduler() *BookingScheduler {
	return &BookingScheduler{
		logService: &BookingLogService{},
	}
}

// StartScheduler 启动订单状态自动管理调度器
func (bs *BookingScheduler) StartScheduler() {
	// 启动定时任务，每分钟检查一次
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				bs.ProcessBookingStatusUpdates()
			}
		}
	}()

	log.Println("订单状态自动管理调度器已启动")

	// 记录调度器启动日志到MongoDB
	mode := os.Getenv("ROUTER_MODE")
	if mode == "" {
		mode = "all"
	}
	bs.logService.LogSchedulerStart(mode)
}

// ProcessBookingStatusUpdates 处理订单状态更新
func (bs *BookingScheduler) ProcessBookingStatusUpdates() {
	now := time.Now()

	// 1. 处理已支付但到了开始时间的订单，改为使用中
	bs.activateBookings(now)

	// 2. 处理使用中但到了结束时间的订单，改为已完成
	bs.completeBookings(now)

	// 3. 处理超过24小时未支付的订单，自动取消
	bs.cancelOverdueBookings(now)
}

// activateBookings 激活到达开始时间的已支付订单
func (bs *BookingScheduler) activateBookings(now time.Time) {
	var bookings []app_model.RoomBooking

	// 查询已支付且开始时间到了的订单
	if err := db.Dao.Where("status = ? AND start_time <= ?",
		app_model.BookingStatusPaid, now).Find(&bookings).Error; err != nil {
		log.Printf("查询待激活订单失败: %v", err)
		return
	}

	for _, booking := range bookings {
		// 获取房间信息
		var room app_model.Room
		if err := db.Dao.First(&room, booking.RoomID).Error; err != nil {
			log.Printf("查询房间信息失败 (房间ID: %d): %v", booking.RoomID, err)
			continue
		}

		// 更新订单状态为使用中
		if err := db.Dao.Model(&booking).Update("status", app_model.BookingStatusInUse).Error; err != nil {
			log.Printf("更新订单状态失败 (ID: %d): %v", booking.ID, err)
			bs.logService.LogBookingError(booking.ID, booking.BookingNo, "激活订单", err)
			continue
		}

		// 更新房间状态为使用中
		if err := db.Dao.Model(&app_model.Room{}).Where("id = ?", booking.RoomID).
			Update("status", app_model.RoomStatusOccupied).Error; err != nil {
			log.Printf("更新房间状态失败 (房间ID: %d): %v", booking.RoomID, err)
			bs.logService.LogRoomError(booking.RoomID, room.RoomName, "设为使用中", err)
		}

		// 创建使用记录
		usageLog := &app_model.RoomUsageLog{
			RoomID:    booking.RoomID,
			BookingID: booking.ID,
			UserID:    booking.UserID,
			CheckInAt: now,
		}

		if err := db.Dao.Create(usageLog).Error; err != nil {
			log.Printf("创建使用记录失败 (订单ID: %d): %v", booking.ID, err)
			bs.logService.LogUsageError(booking.ID, booking.BookingNo, "创建使用记录", err)
		}

		log.Printf("订单已激活: %s (房间ID: %d, 用户ID: %d)",
			booking.BookingNo, booking.RoomID, booking.UserID)

		// 记录成功激活日志
		bs.logService.LogBookingActivate(&booking, room.RoomName)
	}
}

// completeBookings 完成到达结束时间的使用中订单
func (bs *BookingScheduler) completeBookings(now time.Time) {
	var bookings []app_model.RoomBooking

	// 查询使用中且结束时间到了的订单
	if err := db.Dao.Where("status = ? AND end_time <= ?",
		app_model.BookingStatusInUse, now).Find(&bookings).Error; err != nil {
		log.Printf("查询待完成订单失败: %v", err)
		return
	}

	for _, booking := range bookings {
		// 获取房间信息
		var room app_model.Room
		if err := db.Dao.First(&room, booking.RoomID).Error; err != nil {
			log.Printf("查询房间信息失败 (房间ID: %d): %v", booking.RoomID, err)
			continue
		}

		// 更新订单状态为已完成
		if err := db.Dao.Model(&booking).Update("status", app_model.BookingStatusCompleted).Error; err != nil {
			log.Printf("更新订单状态失败 (ID: %d): %v", booking.ID, err)
			bs.logService.LogBookingError(booking.ID, booking.BookingNo, "完成订单", err)
			continue
		}

		// 检查该房间是否还有其他使用中的订单
		var activeBookings int64
		if err := db.Dao.Model(&app_model.RoomBooking{}).
			Where("room_id = ? AND status = ? AND id != ?",
				booking.RoomID, app_model.BookingStatusInUse, booking.ID).
			Count(&activeBookings).Error; err != nil {
			log.Printf("检查房间活跃订单失败 (房间ID: %d): %v", booking.RoomID, err)
			continue
		}

		// 如果没有其他活跃订单，将房间状态改为可用
		if activeBookings == 0 {
			if err := db.Dao.Model(&app_model.Room{}).Where("id = ?", booking.RoomID).
				Update("status", app_model.RoomStatusAvailable).Error; err != nil {
				log.Printf("更新房间状态失败 (房间ID: %d): %v", booking.RoomID, err)
				bs.logService.LogRoomError(booking.RoomID, room.RoomName, "设为可用", err)
			}
		}

		// 更新使用记录的离开时间
		var usageLog app_model.RoomUsageLog
		var actualHours float64
		if err := db.Dao.Where("booking_id = ?", booking.ID).First(&usageLog).Error; err == nil {
			actualHours = now.Sub(usageLog.CheckInAt).Hours()
			updates := map[string]interface{}{
				"check_out_at": &now,
				"actual_hours": actualHours,
			}

			if err := db.Dao.Model(&usageLog).Updates(updates).Error; err != nil {
				log.Printf("更新使用记录失败 (订单ID: %d): %v", booking.ID, err)
				bs.logService.LogUsageError(booking.ID, booking.BookingNo, "更新使用记录", err)
			}
		} else {
			actualHours = float64(booking.Hours) // 如果没有使用记录，使用预订小时数
		}

		log.Printf("订单已完成: %s (房间ID: %d, 用户ID: %d)",
			booking.BookingNo, booking.RoomID, booking.UserID)

		// 记录成功完成日志
		bs.logService.LogBookingComplete(&booking, room.RoomName, actualHours)
	}
}

// cancelOverdueBookings 取消超过24小时未支付的订单
func (bs *BookingScheduler) cancelOverdueBookings(now time.Time) {
	cutoffTime := now.Add(-24 * time.Hour)

	var bookings []app_model.RoomBooking

	// 查询超过24小时未支付的订单
	if err := db.Dao.Where("status = ? AND create_time <= ?",
		app_model.BookingStatusPending, cutoffTime).Find(&bookings).Error; err != nil {
		log.Printf("查询超时未支付订单失败: %v", err)
		return
	}

	for _, booking := range bookings {
		// 更新订单状态为已取消
		updates := map[string]interface{}{
			"status":  app_model.BookingStatusCancelled,
			"remarks": booking.Remarks + "\n系统自动取消: 超过24小时未支付",
		}

		if err := db.Dao.Model(&booking).Updates(updates).Error; err != nil {
			log.Printf("取消超时订单失败 (ID: %d): %v", booking.ID, err)
			bs.logService.LogBookingError(booking.ID, booking.BookingNo, "超时取消", err)
			continue
		}

		log.Printf("超时订单已取消: %s (用户ID: %d)", booking.BookingNo, booking.UserID)

		// 记录超时取消日志
		bs.logService.LogBookingTimeout(&booking)
	}
}

// ManuallyStartBooking 手动开始订单（管理员操作）
func (bs *BookingScheduler) ManuallyStartBooking(bookingID int, adminID int) error {
	var booking app_model.RoomBooking
	if err := db.Dao.First(&booking, bookingID).Error; err != nil {
		return err
	}

	if booking.Status != app_model.BookingStatusPaid {
		return fmt.Errorf("订单状态不正确，无法开始使用")
	}

	now := time.Now()

	// 开始事务
	tx := db.Dao.Begin()

	// 更新订单状态
	if err := tx.Model(&booking).Update("status", app_model.BookingStatusInUse).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新房间状态
	if err := tx.Model(&app_model.Room{}).Where("id = ?", booking.RoomID).
		Update("status", app_model.RoomStatusOccupied).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 创建使用记录
	usageLog := &app_model.RoomUsageLog{
		RoomID:    booking.RoomID,
		BookingID: booking.ID,
		UserID:    booking.UserID,
		CheckInAt: now,
	}

	if err := tx.Create(usageLog).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	log.Printf("手动开始订单: %s", booking.BookingNo)

	// 获取房间信息并记录日志
	var room app_model.Room
	if err := db.Dao.First(&room, booking.RoomID).Error; err == nil {
		bs.logService.LogManualStart(&booking, room.RoomName, adminID)
	}

	return nil
}

// ManuallyEndBooking 手动结束订单（管理员操作）
func (bs *BookingScheduler) ManuallyEndBooking(bookingID int, adminID int) error {
	var booking app_model.RoomBooking
	if err := db.Dao.First(&booking, bookingID).Error; err != nil {
		return err
	}

	if booking.Status != app_model.BookingStatusInUse {
		return fmt.Errorf("订单状态不正确，无法结束使用")
	}

	now := time.Now()

	// 开始事务
	tx := db.Dao.Begin()

	// 更新订单状态
	if err := tx.Model(&booking).Update("status", app_model.BookingStatusCompleted).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 检查房间是否还有其他活跃订单
	var activeBookings int64
	if err := tx.Model(&app_model.RoomBooking{}).
		Where("room_id = ? AND status = ? AND id != ?",
			booking.RoomID, app_model.BookingStatusInUse, booking.ID).
		Count(&activeBookings).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 如果没有其他活跃订单，将房间状态改为可用
	if activeBookings == 0 {
		if err := tx.Model(&app_model.Room{}).Where("id = ?", booking.RoomID).
			Update("status", app_model.RoomStatusAvailable).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 更新使用记录
	var usageLog app_model.RoomUsageLog
	if err := tx.Where("booking_id = ?", booking.ID).First(&usageLog).Error; err == nil {
		actualHours := now.Sub(usageLog.CheckInAt).Hours()
		updates := map[string]interface{}{
			"check_out_at": &now,
			"actual_hours": actualHours,
		}

		if err := tx.Model(&usageLog).Updates(updates).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()

	log.Printf("手动结束订单: %s", booking.BookingNo)

	// 获取房间信息和实际使用时间并记录日志
	var room app_model.Room
	var actualHours float64 = float64(booking.Hours) // 默认使用预订小时数

	if err := db.Dao.First(&room, booking.RoomID).Error; err == nil {
		// 尝试获取实际使用时间
		var usageLog app_model.RoomUsageLog
		if err := db.Dao.Where("booking_id = ?", booking.ID).First(&usageLog).Error; err == nil {
			actualHours = now.Sub(usageLog.CheckInAt).Hours()
		}
		bs.logService.LogManualEnd(&booking, room.RoomName, adminID, actualHours)
	}

	return nil
}

// GetBookingStatusInfo 获取订单状态信息（用于管理后台展示）
func (bs *BookingScheduler) GetBookingStatusInfo(bookingID int) (*BookingStatusInfo, error) {
	var booking app_model.RoomBooking
	if err := db.Dao.Preload("Room").First(&booking, bookingID).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	info := &BookingStatusInfo{
		BookingID:   booking.ID,
		BookingNo:   booking.BookingNo,
		Status:      booking.Status,
		StatusText:  booking.GetBookingStatusText(),
		StartTime:   booking.StartTime,
		EndTime:     booking.EndTime,
		RoomID:      booking.RoomID,
		CurrentTime: now,
	}

	if booking.Room != nil {
		info.RoomName = booking.Room.RoomName
		info.RoomStatus = booking.Room.Status
		info.RoomStatusText = booking.Room.GetStatusText()
	}

	// 判断是否可以手动操作
	if booking.Status == app_model.BookingStatusPaid {
		info.CanStart = true
		info.ShouldAutoStart = now.After(booking.StartTime) || now.Equal(booking.StartTime)
	}

	if booking.Status == app_model.BookingStatusInUse {
		info.CanEnd = true
		info.ShouldAutoEnd = now.After(booking.EndTime) || now.Equal(booking.EndTime)
	}

	return info, nil
}

// BookingStatusInfo 订单状态信息
type BookingStatusInfo struct {
	BookingID       int       `json:"booking_id"`
	BookingNo       string    `json:"booking_no"`
	Status          int       `json:"status"`
	StatusText      string    `json:"status_text"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	RoomID          int       `json:"room_id"`
	RoomName        string    `json:"room_name"`
	RoomStatus      int       `json:"room_status"`
	RoomStatusText  string    `json:"room_status_text"`
	CurrentTime     time.Time `json:"current_time"`
	CanStart        bool      `json:"can_start"`
	CanEnd          bool      `json:"can_end"`
	ShouldAutoStart bool      `json:"should_auto_start"`
	ShouldAutoEnd   bool      `json:"should_auto_end"`
}
