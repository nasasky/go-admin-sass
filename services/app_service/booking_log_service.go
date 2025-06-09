package app_service

import (
	"context"
	"fmt"
	"os"
	"time"

	"nasa-go-admin/model/app_model"
	"nasa-go-admin/mongodb"
	"nasa-go-admin/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BookingLogService struct{}

// LogSchedulerStart 记录调度器启动日志
func (bls *BookingLogService) LogSchedulerStart(mode string) {
	hostname, _ := os.Hostname()
	pid := os.Getpid()

	log := &app_model.BookingStatusLog{
		LogType:   app_model.LogTypeSchedulerStart,
		Message:   "订单状态自动管理调度器已启动",
		CreatedAt: utils.GetCurrentTimeForMongo(),
		ServerInfo: app_model.ServerInfo{
			Hostname: hostname,
			PID:      pid,
			Mode:     mode,
		},
	}

	if err := bls.saveLog(log); err != nil {
		fmt.Printf("保存调度器启动日志失败: %v\n", err)
	}
}

// LogBookingActivate 记录订单激活日志
func (bls *BookingLogService) LogBookingActivate(booking *app_model.RoomBooking, roomName string) {
	oldStatus := app_model.BookingStatusPaid
	newStatus := app_model.BookingStatusInUse

	log := &app_model.BookingStatusLog{
		LogType:   app_model.LogTypeBookingActivate,
		BookingID: &booking.ID,
		BookingNo: booking.BookingNo,
		RoomID:    &booking.RoomID,
		RoomName:  roomName,
		UserID:    &booking.UserID,
		OldStatus: &oldStatus,
		NewStatus: &newStatus,
		Message:   fmt.Sprintf("订单已激活: %s (房间ID: %d, 用户ID: %d)", booking.BookingNo, booking.RoomID, booking.UserID),
		CreatedAt: utils.GetCurrentTimeForMongo(),
		Details: map[string]interface{}{
			"start_time": booking.StartTime,
			"end_time":   booking.EndTime,
			"hours":      booking.Hours,
			"amount":     booking.TotalAmount,
		},
		ServerInfo: bls.getServerInfo(),
	}

	if err := bls.saveLog(log); err != nil {
		fmt.Printf("保存订单激活日志失败: %v\n", err)
	}
}

// LogBookingComplete 记录订单完成日志
func (bls *BookingLogService) LogBookingComplete(booking *app_model.RoomBooking, roomName string, actualHours float64) {
	oldStatus := app_model.BookingStatusInUse
	newStatus := app_model.BookingStatusCompleted

	log := &app_model.BookingStatusLog{
		LogType:   app_model.LogTypeBookingComplete,
		BookingID: &booking.ID,
		BookingNo: booking.BookingNo,
		RoomID:    &booking.RoomID,
		RoomName:  roomName,
		UserID:    &booking.UserID,
		OldStatus: &oldStatus,
		NewStatus: &newStatus,
		Message:   fmt.Sprintf("订单已完成: %s (房间ID: %d, 用户ID: %d)", booking.BookingNo, booking.RoomID, booking.UserID),
		CreatedAt: utils.GetCurrentTimeForMongo(),
		Details: map[string]interface{}{
			"start_time":    booking.StartTime,
			"end_time":      booking.EndTime,
			"planned_hours": booking.Hours,
			"actual_hours":  actualHours,
			"amount":        booking.TotalAmount,
		},
		ServerInfo: bls.getServerInfo(),
	}

	if err := bls.saveLog(log); err != nil {
		fmt.Printf("保存订单完成日志失败: %v\n", err)
	}
}

// LogBookingTimeout 记录订单超时取消日志
func (bls *BookingLogService) LogBookingTimeout(booking *app_model.RoomBooking) {
	oldStatus := app_model.BookingStatusPending
	newStatus := app_model.BookingStatusCancelled

	log := &app_model.BookingStatusLog{
		LogType:   app_model.LogTypeBookingTimeout,
		BookingID: &booking.ID,
		BookingNo: booking.BookingNo,
		RoomID:    &booking.RoomID,
		UserID:    &booking.UserID,
		OldStatus: &oldStatus,
		NewStatus: &newStatus,
		Message:   fmt.Sprintf("超时订单已取消: %s (用户ID: %d)", booking.BookingNo, booking.UserID),
		CreatedAt: utils.GetCurrentTimeForMongo(),
		Details: map[string]interface{}{
			"start_time":    booking.StartTime,
			"end_time":      booking.EndTime,
			"hours":         booking.Hours,
			"amount":        booking.TotalAmount,
			"created_at":    booking.CreateTime,
			"timeout_hours": 24,
		},
		ServerInfo: bls.getServerInfo(),
	}

	if err := bls.saveLog(log); err != nil {
		fmt.Printf("保存订单超时日志失败: %v\n", err)
	}
}

// LogBookingError 记录订单状态更新失败日志
func (bls *BookingLogService) LogBookingError(bookingID int, bookingNo string, operation string, err error) {
	log := &app_model.BookingStatusLog{
		LogType:   app_model.LogTypeBookingError,
		BookingID: &bookingID,
		BookingNo: bookingNo,
		Message:   fmt.Sprintf("更新订单状态失败 (ID: %d, 操作: %s)", bookingID, operation),
		ErrorMsg:  err.Error(),
		CreatedAt: utils.GetCurrentTimeForMongo(),
		Details: map[string]interface{}{
			"operation": operation,
			"error":     err.Error(),
		},
		ServerInfo: bls.getServerInfo(),
	}

	if err := bls.saveLog(log); err != nil {
		fmt.Printf("保存订单错误日志失败: %v\n", err)
	}
}

// LogRoomError 记录房间状态更新失败日志
func (bls *BookingLogService) LogRoomError(roomID int, roomName string, operation string, err error) {
	log := &app_model.BookingStatusLog{
		LogType:   app_model.LogTypeRoomError,
		RoomID:    &roomID,
		RoomName:  roomName,
		Message:   fmt.Sprintf("更新房间状态失败 (房间ID: %d, 操作: %s)", roomID, operation),
		ErrorMsg:  err.Error(),
		CreatedAt: utils.GetCurrentTimeForMongo(),
		Details: map[string]interface{}{
			"operation": operation,
			"error":     err.Error(),
		},
		ServerInfo: bls.getServerInfo(),
	}

	if err := bls.saveLog(log); err != nil {
		fmt.Printf("保存房间错误日志失败: %v\n", err)
	}
}

// LogManualStart 记录手动开始订单日志
func (bls *BookingLogService) LogManualStart(booking *app_model.RoomBooking, roomName string, adminID int) {
	oldStatus := app_model.BookingStatusPaid
	newStatus := app_model.BookingStatusInUse

	log := &app_model.BookingStatusLog{
		LogType:   app_model.LogTypeManualStart,
		BookingID: &booking.ID,
		BookingNo: booking.BookingNo,
		RoomID:    &booking.RoomID,
		RoomName:  roomName,
		UserID:    &booking.UserID,
		OldStatus: &oldStatus,
		NewStatus: &newStatus,
		Message:   fmt.Sprintf("手动开始订单: %s (管理员ID: %d)", booking.BookingNo, adminID),
		CreatedAt: utils.GetCurrentTimeForMongo(),
		Details: map[string]interface{}{
			"admin_id":   adminID,
			"start_time": booking.StartTime,
			"end_time":   booking.EndTime,
			"hours":      booking.Hours,
			"amount":     booking.TotalAmount,
		},
		ServerInfo: bls.getServerInfo(),
	}

	if err := bls.saveLog(log); err != nil {
		fmt.Printf("保存手动开始日志失败: %v\n", err)
	}
}

// LogManualEnd 记录手动结束订单日志
func (bls *BookingLogService) LogManualEnd(booking *app_model.RoomBooking, roomName string, adminID int, actualHours float64) {
	oldStatus := app_model.BookingStatusInUse
	newStatus := app_model.BookingStatusCompleted

	log := &app_model.BookingStatusLog{
		LogType:   app_model.LogTypeManualEnd,
		BookingID: &booking.ID,
		BookingNo: booking.BookingNo,
		RoomID:    &booking.RoomID,
		RoomName:  roomName,
		UserID:    &booking.UserID,
		OldStatus: &oldStatus,
		NewStatus: &newStatus,
		Message:   fmt.Sprintf("手动结束订单: %s (管理员ID: %d)", booking.BookingNo, adminID),
		CreatedAt: utils.GetCurrentTimeForMongo(),
		Details: map[string]interface{}{
			"admin_id":      adminID,
			"start_time":    booking.StartTime,
			"end_time":      booking.EndTime,
			"planned_hours": booking.Hours,
			"actual_hours":  actualHours,
			"amount":        booking.TotalAmount,
		},
		ServerInfo: bls.getServerInfo(),
	}

	if err := bls.saveLog(log); err != nil {
		fmt.Printf("保存手动结束日志失败: %v\n", err)
	}
}

// LogUsageError 记录使用记录错误日志
func (bls *BookingLogService) LogUsageError(bookingID int, bookingNo string, operation string, err error) {
	log := &app_model.BookingStatusLog{
		LogType:   app_model.LogTypeUsageLogError,
		BookingID: &bookingID,
		BookingNo: bookingNo,
		Message:   fmt.Sprintf("使用记录错误 (订单ID: %d, 操作: %s)", bookingID, operation),
		ErrorMsg:  err.Error(),
		CreatedAt: utils.GetCurrentTimeForMongo(),
		Details: map[string]interface{}{
			"operation": operation,
			"error":     err.Error(),
		},
		ServerInfo: bls.getServerInfo(),
	}

	if err := bls.saveLog(log); err != nil {
		fmt.Printf("保存使用记录错误日志失败: %v\n", err)
	}
}

// GetLogList 获取日志列表
func (bls *BookingLogService) GetLogList(req *BookingLogListReq) (*BookingLogListResp, error) {
	collection := mongodb.GetCollection("booking_log_db", "logs")
	ctx := context.Background()

	// 构建查询条件
	filter := bson.M{}

	if req.LogType != "" {
		filter["log_type"] = req.LogType
	}
	if req.BookingID > 0 {
		filter["booking_id"] = req.BookingID
	}
	if req.BookingNo != "" {
		filter["booking_no"] = bson.M{"$regex": req.BookingNo, "$options": "i"}
	}
	if req.RoomID > 0 {
		filter["room_id"] = req.RoomID
	}
	if req.UserID > 0 {
		filter["user_id"] = req.UserID
	}
	if req.StartDate != "" && req.EndDate != "" {
		startTime, _ := time.Parse("2006-01-02", req.StartDate)
		endTime, _ := time.Parse("2006-01-02", req.EndDate)
		endTime = endTime.Add(24*time.Hour - time.Nanosecond) // 当天23:59:59

		// 由于created_at现在是字符串，使用字符串比较
		startTimeStr := startTime.Format("2006-01-02 15:04:05")
		endTimeStr := endTime.Format("2006-01-02 15:04:05")

		filter["created_at"] = bson.M{
			"$gte": startTimeStr,
			"$lte": endTimeStr,
		}
	}

	// 获取总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("查询日志总数失败: %v", err)
	}

	// 分页查询
	skip := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}}) // 按创建时间倒序

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询日志列表失败: %v", err)
	}
	defer cursor.Close(ctx)

	var logs []app_model.BookingStatusLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("解析日志数据失败: %v", err)
	}

	// 转换为响应格式
	list := make([]BookingLogDetail, 0, len(logs))
	for _, logItem := range logs {
		detail := BookingLogDetail{
			ID:          logItem.ID.Hex(),
			LogType:     logItem.LogType,
			LogTypeText: app_model.GetLogTypeText(logItem.LogType),
			BookingID:   logItem.BookingID,
			BookingNo:   logItem.BookingNo,
			RoomID:      logItem.RoomID,
			RoomName:    logItem.RoomName,
			UserID:      logItem.UserID,
			OldStatus:   logItem.OldStatus,
			NewStatus:   logItem.NewStatus,
			Message:     logItem.Message,
			ErrorMsg:    logItem.ErrorMsg,
			Details:     logItem.Details,
			CreatedAt:   logItem.CreatedAt,
			ServerInfo:  logItem.ServerInfo,
		}

		if logItem.OldStatus != nil {
			text := app_model.GetStatusText(*logItem.OldStatus)
			detail.OldStatusText = &text
		}
		if logItem.NewStatus != nil {
			text := app_model.GetStatusText(*logItem.NewStatus)
			detail.NewStatusText = &text
		}

		list = append(list, detail)
	}

	return &BookingLogListResp{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		List:     list,
	}, nil
}

// GetLogStatistics 获取日志统计信息
func (bls *BookingLogService) GetLogStatistics(startDate, endDate string) (*BookingLogStatistics, error) {
	collection := mongodb.GetCollection("booking_log_db", "logs")
	ctx := context.Background()

	// 构建时间范围过滤器
	filter := bson.M{}
	if startDate != "" && endDate != "" {
		startTime, _ := time.Parse("2006-01-02", startDate)
		endTime, _ := time.Parse("2006-01-02", endDate)
		endTime = endTime.Add(24*time.Hour - time.Nanosecond)

		filter["created_at"] = bson.M{
			"$gte": startTime,
			"$lte": endTime,
		}
	}

	// 统计各类型日志数量
	pipeline := []bson.M{
		{"$match": filter},
		{"$group": bson.M{
			"_id":   "$log_type",
			"count": bson.M{"$sum": 1},
		}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("统计日志失败: %v", err)
	}
	defer cursor.Close(ctx)

	stats := &BookingLogStatistics{
		TypeStats: make(map[string]int64),
	}

	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		stats.TypeStats[result.ID] = result.Count
		stats.TotalLogs += result.Count
	}

	// 统计错误日志数量
	errorFilter := bson.M{
		"error_msg": bson.M{"$ne": ""},
	}
	if startDate != "" && endDate != "" {
		startTime, _ := time.Parse("2006-01-02", startDate)
		endTime, _ := time.Parse("2006-01-02", endDate)
		endTime = endTime.Add(24*time.Hour - time.Nanosecond)

		errorFilter["created_at"] = bson.M{
			"$gte": startTime,
			"$lte": endTime,
		}
	}

	errorCount, err := collection.CountDocuments(ctx, errorFilter)
	if err == nil {
		stats.ErrorLogs = errorCount
	}

	return stats, nil
}

// saveLog 保存日志到MongoDB
func (bls *BookingLogService) saveLog(log *app_model.BookingStatusLog) error {
	collection := mongodb.GetCollection("booking_log_db", "logs")
	ctx := context.Background()

	_, err := collection.InsertOne(ctx, log)
	return err
}

// getServerInfo 获取服务器信息
func (bls *BookingLogService) getServerInfo() app_model.ServerInfo {
	hostname, _ := os.Hostname()
	pid := os.Getpid()
	mode := os.Getenv("ROUTER_MODE")
	if mode == "" {
		mode = "all"
	}

	return app_model.ServerInfo{
		Hostname: hostname,
		PID:      pid,
		Mode:     mode,
	}
}

// 请求和响应结构
type BookingLogListReq struct {
	Page      int    `json:"page" form:"page" binding:"min=1"`
	PageSize  int    `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	LogType   string `json:"log_type" form:"log_type"`
	BookingID int    `json:"booking_id" form:"booking_id"`
	BookingNo string `json:"booking_no" form:"booking_no"`
	RoomID    int    `json:"room_id" form:"room_id"`
	UserID    int    `json:"user_id" form:"user_id"`
	StartDate string `json:"start_date" form:"start_date"`
	EndDate   string `json:"end_date" form:"end_date"`
}

type BookingLogListResp struct {
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	List     []BookingLogDetail `json:"list"`
}

type BookingLogDetail struct {
	ID            string               `json:"id"`
	LogType       string               `json:"log_type"`
	LogTypeText   string               `json:"log_type_text"`
	BookingID     *int                 `json:"booking_id"`
	BookingNo     string               `json:"booking_no"`
	RoomID        *int                 `json:"room_id"`
	RoomName      string               `json:"room_name"`
	UserID        *int                 `json:"user_id"`
	OldStatus     *int                 `json:"old_status"`
	OldStatusText *string              `json:"old_status_text"`
	NewStatus     *int                 `json:"new_status"`
	NewStatusText *string              `json:"new_status_text"`
	Message       string               `json:"message"`
	ErrorMsg      string               `json:"error_msg"`
	Details       interface{}          `json:"details"`
	CreatedAt     string               `json:"created_at"`
	ServerInfo    app_model.ServerInfo `json:"server_info"`
}

type BookingLogStatistics struct {
	TotalLogs int64            `json:"total_logs"`
	ErrorLogs int64            `json:"error_logs"`
	TypeStats map[string]int64 `json:"type_stats"`
}
