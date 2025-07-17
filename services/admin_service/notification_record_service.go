package admin_service

import (
	"context"
	"fmt"
	"log"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/mongodb"
	"nasa-go-admin/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NotificationRecordService 推送记录服务
type NotificationRecordService struct{}

// NewNotificationRecordService 创建推送记录服务实例
func NewNotificationRecordService() *NotificationRecordService {
	return &NotificationRecordService{}
}

// SavePushRecord 保存推送记录
func (s *NotificationRecordService) SavePushRecord(record *admin_model.PushRecord) error {
	collection := mongodb.GetCollection("notification_log_db", "push_records")
	if collection == nil {
		log.Printf("⚠️ MongoDB collection 不可用，跳过推送记录保存: MessageID=%s", record.MessageID)
		return fmt.Errorf("MongoDB collection 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 设置创建和更新时间
	now := utils.GetCurrentTimeForMongo()
	if record.CreatedAt == "" {
		record.CreatedAt = now
	}
	record.UpdatedAt = now

	// 插入记录
	result, err := collection.InsertOne(ctx, record)
	if err != nil {
		log.Printf("保存推送记录失败: %v", err)
		return fmt.Errorf("保存推送记录失败: %w", err)
	}

	// 设置生成的ID
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		record.ID = oid
	}

	log.Printf("推送记录已保存: MessageID=%s, Status=%s", record.MessageID, record.Status)
	return nil
}

// SaveNotificationLog 保存通知日志
func (s *NotificationRecordService) SaveNotificationLog(logRecord *admin_model.NotificationLog) error {
	collection := mongodb.GetCollection("notification_log_db", "notification_logs")
	if collection == nil {
		log.Printf("⚠️ MongoDB collection 不可用，跳过通知日志保存: MessageID=%s", logRecord.MessageID)
		return fmt.Errorf("MongoDB collection 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 设置创建时间
	if logRecord.CreatedAt == "" {
		logRecord.CreatedAt = utils.GetCurrentTimeForMongo()
	}

	// 插入记录
	result, err := collection.InsertOne(ctx, logRecord)
	if err != nil {
		log.Printf("保存通知日志失败: %v", err)
		return fmt.Errorf("保存通知日志失败: %w", err)
	}

	// 设置生成的ID
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		logRecord.ID = oid
	}

	return nil
}

// GetPushRecordList 获取推送记录列表
func (s *NotificationRecordService) GetPushRecordList(query *admin_model.PushRecordQuery) (*admin_model.PushRecordListResponse, error) {
	collection := mongodb.GetCollection("notification_log_db", "push_records")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 构建查询条件
	filter := bson.M{}

	if query.MessageType != "" {
		filter["message_type"] = query.MessageType
	}
	if query.Target != "" {
		filter["target"] = query.Target
	}
	if query.Status != "" {
		filter["status"] = query.Status
	}
	if query.Success != nil {
		filter["success"] = *query.Success
	}
	if query.SenderID > 0 {
		filter["sender_id"] = query.SenderID
	}
	if query.Keyword != "" {
		filter["$or"] = bson.A{
			bson.M{"content": bson.M{"$regex": query.Keyword, "$options": "i"}},
			bson.M{"message_id": bson.M{"$regex": query.Keyword, "$options": "i"}},
		}
	}
	if query.StartDate != "" || query.EndDate != "" {
		timeFilter := bson.M{}
		if query.StartDate != "" {
			timeFilter["$gte"] = query.StartDate
		}
		if query.EndDate != "" {
			timeFilter["$lte"] = query.EndDate
		}
		filter["push_time"] = timeFilter
	}

	// 获取总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("获取推送记录总数失败: %w", err)
	}

	// 构建排序选项
	sortOptions := options.Find()
	if query.SortBy != "" {
		sortOrder := 1
		if query.SortOrder == "desc" {
			sortOrder = -1
		}
		sortOptions.SetSort(bson.D{{Key: query.SortBy, Value: sortOrder}})
	} else {
		// 默认按推送时间倒序
		sortOptions.SetSort(bson.D{{Key: "push_time", Value: -1}})
	}

	// 设置分页
	skip := (query.Page - 1) * query.PageSize
	sortOptions.SetSkip(int64(skip))
	sortOptions.SetLimit(int64(query.PageSize))

	// 查询记录
	cursor, err := collection.Find(ctx, filter, sortOptions)
	if err != nil {
		return nil, fmt.Errorf("查询推送记录失败: %w", err)
	}
	defer cursor.Close(ctx)

	var records []admin_model.PushRecord
	if err = cursor.All(ctx, &records); err != nil {
		return nil, fmt.Errorf("解析推送记录失败: %w", err)
	}

	// 获取统计信息
	stats, err := s.getPushRecordStats(ctx, collection, filter)
	if err != nil {
		log.Printf("获取推送记录统计失败: %v", err)
		// 不返回错误，继续返回列表数据
	}

	return &admin_model.PushRecordListResponse{
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
		Items:    records,
		Stats:    *stats,
	}, nil
}

// GetPushRecordByID 根据ID获取推送记录
func (s *NotificationRecordService) GetPushRecordByID(id string) (*admin_model.PushRecord, error) {
	collection := mongodb.GetCollection("notification_log_db", "push_records")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("无效的记录ID: %w", err)
	}

	filter := bson.M{"_id": objectID}
	var record admin_model.PushRecord
	err = collection.FindOne(ctx, filter).Decode(&record)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("推送记录不存在")
		}
		return nil, fmt.Errorf("查询推送记录失败: %w", err)
	}

	return &record, nil
}

// GetPushRecordByMessageID 根据消息ID获取推送记录
func (s *NotificationRecordService) GetPushRecordByMessageID(messageID string) (*admin_model.PushRecord, error) {
	collection := mongodb.GetCollection("notification_log_db", "push_records")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"message_id": messageID}
	var record admin_model.PushRecord
	err := collection.FindOne(ctx, filter).Decode(&record)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("推送记录不存在")
		}
		return nil, fmt.Errorf("查询推送记录失败: %w", err)
	}

	return &record, nil
}

// UpdatePushRecord 更新推送记录
func (s *NotificationRecordService) UpdatePushRecord(messageID string, updates map[string]interface{}) error {
	collection := mongodb.GetCollection("notification_log_db", "push_records")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 添加更新时间
	updates["updated_at"] = utils.GetCurrentTimeForMongo()

	filter := bson.M{"message_id": messageID}
	update := bson.M{"$set": updates}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("更新推送记录失败: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("推送记录不存在")
	}

	log.Printf("推送记录已更新: MessageID=%s", messageID)
	return nil
}

// DeletePushRecord 删除推送记录
func (s *NotificationRecordService) DeletePushRecord(id string) error {
	collection := mongodb.GetCollection("notification_log_db", "push_records")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("无效的记录ID: %w", err)
	}

	filter := bson.M{"_id": objectID}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("删除推送记录失败: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("推送记录不存在")
	}

	log.Printf("推送记录已删除: ID=%s", id)
	return nil
}

// getPushRecordStats 获取推送记录统计信息
func (s *NotificationRecordService) getPushRecordStats(ctx context.Context, collection *mongo.Collection, filter bson.M) (*admin_model.PushRecordStats, error) {
	// 成功记录数
	successFilter := bson.M{}
	for k, v := range filter {
		successFilter[k] = v
	}
	successFilter["success"] = true
	successCount, err := collection.CountDocuments(ctx, successFilter)
	if err != nil {
		return nil, err
	}

	// 失败记录数
	failedFilter := bson.M{}
	for k, v := range filter {
		failedFilter[k] = v
	}
	failedFilter["success"] = false
	failedCount, err := collection.CountDocuments(ctx, failedFilter)
	if err != nil {
		return nil, err
	}

	// 总接收者数和送达数统计
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.M{
			"_id":              nil,
			"total_recipients": bson.M{"$sum": "$total_count"},
			"delivered_count":  bson.M{"$sum": "$delivered_count"},
			"failed_count":     bson.M{"$sum": "$failed_count"},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	var totalRecipients, deliveredCount, failedCountTotal int64
	if len(results) > 0 {
		if val, ok := results[0]["total_recipients"].(int64); ok {
			totalRecipients = val
		}
		if val, ok := results[0]["delivered_count"].(int64); ok {
			deliveredCount = val
		}
		if val, ok := results[0]["failed_count"].(int64); ok {
			failedCountTotal = val
		}
	}

	return &admin_model.PushRecordStats{
		TotalRecords:    successCount + failedCount,
		SuccessRecords:  successCount,
		FailedRecords:   failedCount,
		TotalRecipients: totalRecipients,
		DeliveredCount:  deliveredCount,
		FailedCount:     failedCountTotal,
	}, nil
}

// GetNotificationLogs 获取通知日志
func (s *NotificationRecordService) GetNotificationLogs(messageID string, limit int64) ([]admin_model.NotificationLog, error) {
	collection := mongodb.GetCollection("notification_log_db", "notification_logs")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"message_id": messageID}
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetLimit(limit)

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询通知日志失败: %w", err)
	}
	defer cursor.Close(ctx)

	var logs []admin_model.NotificationLog
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("解析通知日志失败: %w", err)
	}

	return logs, nil
}

// ========== 新增：管理员用户接收记录管理 ==========

// SaveAdminUserReceiveRecord 保存管理员用户接收记录
func (s *NotificationRecordService) SaveAdminUserReceiveRecord(record *admin_model.AdminUserReceiveRecord) error {
	collection := mongodb.GetCollection("notification_log_db", "admin_user_receive_records")
	if collection == nil {
		log.Printf("⚠️ MongoDB collection 不可用，跳过管理员用户接收记录保存: MessageID=%s, UserID=%d", record.MessageID, record.UserID)
		return fmt.Errorf("MongoDB collection 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 设置创建和更新时间
	now := utils.GetCurrentTimeForMongo()
	if record.CreatedAt == "" {
		record.CreatedAt = now
	}
	record.UpdatedAt = now

	// 插入记录
	result, err := collection.InsertOne(ctx, record)
	if err != nil {
		log.Printf("保存管理员用户接收记录失败: %v", err)
		return fmt.Errorf("保存管理员用户接收记录失败: %w", err)
	}

	// 设置生成的ID
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		record.ID = oid
	}

	log.Printf("管理员用户接收记录已保存: MessageID=%s, UserID=%d", record.MessageID, record.UserID)
	return nil
}

// GetAdminUserReceiveRecords 获取管理员用户接收记录列表
func (s *NotificationRecordService) GetAdminUserReceiveRecords(query *admin_model.AdminUserReceiveQuery) (*admin_model.PushRecordListResponse, error) {
	collection := mongodb.GetCollection("notification_log_db", "admin_user_receive_records")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 构建查询条件
	filter := bson.M{}

	if query.MessageID != "" {
		filter["message_id"] = query.MessageID
	}
	if query.UserID > 0 {
		filter["user_id"] = query.UserID
	}
	if query.Username != "" {
		filter["username"] = bson.M{"$regex": query.Username, "$options": "i"}
	}
	if query.IsOnline != nil {
		filter["is_online"] = *query.IsOnline
	}
	if query.IsReceived != nil {
		filter["is_received"] = *query.IsReceived
	}
	if query.IsRead != nil {
		filter["is_read"] = *query.IsRead
	}
	if query.IsConfirmed != nil {
		filter["is_confirmed"] = *query.IsConfirmed
	}
	if query.DeliveryStatus != "" {
		filter["delivery_status"] = query.DeliveryStatus
	}
	if query.PushChannel != "" {
		filter["push_channel"] = query.PushChannel
	}
	if query.StartDate != "" || query.EndDate != "" {
		timeFilter := bson.M{}
		if query.StartDate != "" {
			timeFilter["$gte"] = query.StartDate
		}
		if query.EndDate != "" {
			timeFilter["$lte"] = query.EndDate
		}
		filter["created_at"] = timeFilter
	}

	// 获取总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("获取管理员用户接收记录总数失败: %w", err)
	}

	// 构建排序选项
	sortOptions := options.Find()
	if query.SortBy != "" {
		sortOrder := 1
		if query.SortOrder == "desc" {
			sortOrder = -1
		}
		sortOptions.SetSort(bson.D{{Key: query.SortBy, Value: sortOrder}})
	} else {
		// 默认按创建时间倒序
		sortOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})
	}

	// 设置分页
	skip := (query.Page - 1) * query.PageSize
	sortOptions.SetSkip(int64(skip))
	sortOptions.SetLimit(int64(query.PageSize))

	// 查询记录
	cursor, err := collection.Find(ctx, filter, sortOptions)
	if err != nil {
		return nil, fmt.Errorf("查询管理员用户接收记录失败: %w", err)
	}
	defer cursor.Close(ctx)

	var records []admin_model.AdminUserReceiveRecord
	if err = cursor.All(ctx, &records); err != nil {
		return nil, fmt.Errorf("解析管理员用户接收记录失败: %w", err)
	}

	// 获取统计信息
	_, err = s.getAdminUserReceiveStats(ctx, collection, filter)
	if err != nil {
		log.Printf("获取管理员用户接收统计失败: %v", err)
		// 不返回错误，继续返回列表数据
	}

	// 转换为通用响应格式
	var items []admin_model.PushRecord
	for _, record := range records {
		// 动态检查用户当前在线状态
		currentOnlineStatus := s.getUserCurrentOnlineStatus(record.UserID)

		// 获取推送记录的详细信息
		pushRecord, err := s.GetPushRecordByMessageID(record.MessageID)
		if err != nil {
			log.Printf("获取推送记录失败: MessageID=%s, Error=%v", record.MessageID, err)
		}

		// 这里简化处理，实际使用时可能需要更复杂的转换
		item := admin_model.PushRecord{
			MessageID: record.MessageID,
			Status:    record.DeliveryStatus,
			Success:   record.IsReceived,
			CreatedAt: record.CreatedAt,
			UpdatedAt: record.UpdatedAt,
			IsOnline:  currentOnlineStatus,
		}

		// 如果获取到了推送记录，填充详细信息
		if pushRecord != nil {
			item.Content = pushRecord.Content
			item.MessageType = pushRecord.MessageType
			item.Target = pushRecord.Target
			item.RecipientsCount = pushRecord.RecipientsCount
			item.PushTime = pushRecord.PushTime
			item.SenderID = pushRecord.SenderID
			item.SenderName = pushRecord.SenderName
			item.Priority = pushRecord.Priority
			item.NeedConfirm = pushRecord.NeedConfirm
		}

		// 如果用户当前在线，更新状态
		if currentOnlineStatus {
			item.Status = "online"
		}

		items = append(items, item)
	}

	return &admin_model.PushRecordListResponse{
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
		Items:    items,
		Stats:    admin_model.PushRecordStats{}, // 使用新的统计数据
	}, nil
}

// UpdateAdminUserReceiveRecord 更新管理员用户接收记录
func (s *NotificationRecordService) UpdateAdminUserReceiveRecord(messageID string, userID int, updates map[string]interface{}) error {
	collection := mongodb.GetCollection("notification_log_db", "admin_user_receive_records")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 添加更新时间
	updates["updated_at"] = utils.GetCurrentTimeForMongo()

	filter := bson.M{"message_id": messageID, "user_id": userID}
	update := bson.M{"$set": updates}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("更新管理员用户接收记录失败: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("管理员用户接收记录不存在")
	}

	log.Printf("管理员用户接收记录已更新: MessageID=%s, UserID=%d", messageID, userID)
	return nil
}

// MarkMessageAsReceived 标记消息为已接收
func (s *NotificationRecordService) MarkMessageAsReceived(messageID string, userID int, connectionID string) error {
	updates := map[string]interface{}{
		"is_received":   true,
		"received_at":   utils.GetCurrentTimeForMongo(),
		"connection_id": connectionID,
	}
	return s.UpdateAdminUserReceiveRecord(messageID, userID, updates)
}

// MarkMessageAsRead 标记消息为已读
func (s *NotificationRecordService) MarkMessageAsRead(messageID string, userID int) error {
	updates := map[string]interface{}{
		"is_read": true,
		"read_at": utils.GetCurrentTimeForMongo(),
	}
	return s.UpdateAdminUserReceiveRecord(messageID, userID, updates)
}

// MarkMessageAsConfirmed 标记消息为已确认
func (s *NotificationRecordService) MarkMessageAsConfirmed(messageID string, userID int) error {
	updates := map[string]interface{}{
		"is_confirmed": true,
		"confirmed_at": utils.GetCurrentTimeForMongo(),
	}
	return s.UpdateAdminUserReceiveRecord(messageID, userID, updates)
}

// GetMessageReceiveStatus 获取消息的接收状态
func (s *NotificationRecordService) GetMessageReceiveStatus(messageID string) (map[string]interface{}, error) {
	collection := mongodb.GetCollection("notification_log_db", "admin_user_receive_records")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"message_id": messageID}

	// 统计各种状态的用户数量
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.M{
			"_id":             nil,
			"total_users":     bson.M{"$sum": 1},
			"online_users":    bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$is_online", true}}, 1, 0}}},
			"received_users":  bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$is_received", true}}, 1, 0}}},
			"read_users":      bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$is_read", true}}, 1, 0}}},
			"confirmed_users": bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$is_confirmed", true}}, 1, 0}}},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("获取消息接收状态失败: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("解析消息接收状态失败: %w", err)
	}

	if len(results) == 0 {
		return map[string]interface{}{
			"total_users":     0,
			"online_users":    0,
			"received_users":  0,
			"read_users":      0,
			"confirmed_users": 0,
			"receive_rate":    0.0,
			"read_rate":       0.0,
			"confirm_rate":    0.0,
		}, nil
	}

	result := results[0]
	totalUsers := getInt64FromBson(result, "total_users")
	onlineUsers := getInt64FromBson(result, "online_users")
	receivedUsers := getInt64FromBson(result, "received_users")
	readUsers := getInt64FromBson(result, "read_users")
	confirmedUsers := getInt64FromBson(result, "confirmed_users")

	receiveRate := 0.0
	readRate := 0.0
	confirmRate := 0.0

	if totalUsers > 0 {
		receiveRate = float64(receivedUsers) / float64(totalUsers) * 100
		readRate = float64(readUsers) / float64(totalUsers) * 100
		confirmRate = float64(confirmedUsers) / float64(totalUsers) * 100
	}

	return map[string]interface{}{
		"total_users":     totalUsers,
		"online_users":    onlineUsers,
		"received_users":  receivedUsers,
		"read_users":      readUsers,
		"confirmed_users": confirmedUsers,
		"receive_rate":    receiveRate,
		"read_rate":       readRate,
		"confirm_rate":    confirmRate,
	}, nil
}

// ========== 用户在线状态管理 ==========

// UpdateAdminUserOnlineStatus 更新管理员用户在线状态
func (s *NotificationRecordService) UpdateAdminUserOnlineStatus(userID int, username string, isOnline bool, connectionID string, clientIP string, userAgent string) error {
	collection := mongodb.GetCollection("notification_log_db", "admin_user_online_status")
	if collection == nil {
		log.Printf("⚠️ MongoDB collection 不可用，跳过用户在线状态更新: UserID=%d", userID)
		return fmt.Errorf("MongoDB collection 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := utils.GetCurrentTimeForMongo()
	filter := bson.M{"user_id": userID}

	// 查找现有记录
	var existingStatus admin_model.AdminUserOnlineStatus
	err := collection.FindOne(ctx, filter).Decode(&existingStatus)

	if err == mongo.ErrNoDocuments {
		// 创建新记录
		newStatus := admin_model.AdminUserOnlineStatus{
			UserID:           userID,
			Username:         username,
			IsOnline:         isOnline,
			LastSeen:         now,
			ConnectionID:     connectionID,
			ClientIP:         clientIP,
			UserAgent:        userAgent,
			TotalOnlineCount: 1,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		if isOnline {
			newStatus.OnlineTime = now
		} else {
			newStatus.OfflineTime = now
		}

		_, err = collection.InsertOne(ctx, newStatus)
		if err != nil {
			return fmt.Errorf("创建用户在线状态失败: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("查询用户在线状态失败: %w", err)
	} else {
		// 更新现有记录
		updates := bson.M{
			"is_online":     isOnline,
			"last_seen":     now,
			"connection_id": connectionID,
			"client_ip":     clientIP,
			"user_agent":    userAgent,
			"updated_at":    now,
		}

		if isOnline && !existingStatus.IsOnline {
			// 用户上线
			updates["online_time"] = now
			updates["total_online_count"] = existingStatus.TotalOnlineCount + 1
		} else if !isOnline && existingStatus.IsOnline {
			// 用户下线
			updates["offline_time"] = now

			// 计算本次在线时长
			if existingStatus.OnlineTime != "" {
				onlineTime, _ := time.Parse("2006-01-02 15:04:05", existingStatus.OnlineTime)
				offlineTime, _ := time.Parse("2006-01-02 15:04:05", now)
				duration := offlineTime.Sub(onlineTime).Seconds()
				updates["online_duration"] = int64(duration)
				updates["total_online_time"] = existingStatus.TotalOnlineTime + int64(duration)
			}
		}

		_, err = collection.UpdateOne(ctx, filter, bson.M{"$set": updates})
		if err != nil {
			return fmt.Errorf("更新用户在线状态失败: %w", err)
		}
	}

	log.Printf("用户在线状态已更新: UserID=%d, IsOnline=%t", userID, isOnline)
	return nil
}

// GetOnlineAdminUsers 获取在线的管理员用户列表
func (s *NotificationRecordService) GetOnlineAdminUsers() ([]admin_model.AdminUserOnlineStatus, error) {
	collection := mongodb.GetCollection("notification_log_db", "admin_user_online_status")
	if collection == nil {
		log.Printf("⚠️ MongoDB collection 不可用，返回空在线用户列表")
		return []admin_model.AdminUserOnlineStatus{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 首先获取数据库中标记为在线的用户
	filter := bson.M{"is_online": true}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("查询在线管理员用户失败: %w", err)
	}
	defer cursor.Close(ctx)

	var users []admin_model.AdminUserOnlineStatus
	if err = cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("解析在线管理员用户失败: %w", err)
	}

	// 验证这些用户是否真的在Hub中有活跃连接
	// 注意：这里我们不能直接访问Hub，因为会造成循环依赖
	// 所以我们通过检查last_seen时间来判断连接是否真的活跃
	var realOnlineUsers []admin_model.AdminUserOnlineStatus
	now := time.Now()

	for _, user := range users {
		// 检查最后活跃时间，如果超过5分钟没有活动，认为已离线
		if user.LastSeen != "" {
			lastSeen, err := time.Parse("2006-01-02 15:04:05", user.LastSeen)
			if err == nil && now.Sub(lastSeen) < 5*time.Minute {
				realOnlineUsers = append(realOnlineUsers, user)
			} else {
				// 用户可能已经离线，直接更新数据库状态（避免递归调用）
				log.Printf("用户 %d 最后活跃时间过久，标记为离线", user.UserID)
				s.forceUpdateUserOfflineStatus(user.UserID, user.Username)
			}
		} else {
			// 没有最后活跃时间，也标记为离线
			log.Printf("用户 %d 没有最后活跃时间，标记为离线", user.UserID)
			s.forceUpdateUserOfflineStatus(user.UserID, user.Username)
		}
	}

	log.Printf("在线用户验证: 数据库显示=%d, 实际在线=%d", len(users), len(realOnlineUsers))
	return realOnlineUsers, nil
}

// UpdateUserReceiveRecordsOnlineStatus 更新用户所有接收记录的在线状态
func (s *NotificationRecordService) UpdateUserReceiveRecordsOnlineStatus(userID int, isOnline bool, connectionID string) error {
	collection := mongodb.GetCollection("notification_log_db", "admin_user_receive_records")
	if collection == nil {
		log.Printf("⚠️ MongoDB collection 不可用，跳过用户接收记录在线状态更新: UserID=%d", userID)
		return fmt.Errorf("MongoDB collection 不可用")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updates := bson.M{
		"is_online":     isOnline,
		"connection_id": connectionID,
		"updated_at":    utils.GetCurrentTimeForMongo(),
	}

	if isOnline {
		updates["delivery_status"] = "pending" // 用户上线，消息可以投递
	} else {
		updates["delivery_status"] = "offline" // 用户下线，消息无法投递
	}

	filter := bson.M{"user_id": userID}
	update := bson.M{"$set": updates}

	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("更新用户接收记录在线状态失败: %w", err)
	}

	log.Printf("已更新用户接收记录在线状态: UserID=%d, IsOnline=%t, UpdatedCount=%d", userID, isOnline, result.ModifiedCount)
	return nil
}

// getAdminUserReceiveStats 获取管理员用户接收统计信息
func (s *NotificationRecordService) getAdminUserReceiveStats(ctx context.Context, collection *mongo.Collection, filter bson.M) (*admin_model.AdminUserReceiveStats, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$group", Value: bson.M{
			"_id":             nil,
			"total_users":     bson.M{"$sum": 1},
			"online_users":    bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$is_online", true}}, 1, 0}}},
			"received_users":  bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$is_received", true}}, 1, 0}}},
			"read_users":      bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$is_read", true}}, 1, 0}}},
			"confirmed_users": bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$is_confirmed", true}}, 1, 0}}},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return &admin_model.AdminUserReceiveStats{}, nil
	}

	result := results[0]
	totalUsers := getInt64FromBson(result, "total_users")
	onlineUsers := getInt64FromBson(result, "online_users")
	receivedUsers := getInt64FromBson(result, "received_users")
	readUsers := getInt64FromBson(result, "read_users")
	confirmedUsers := getInt64FromBson(result, "confirmed_users")

	stats := &admin_model.AdminUserReceiveStats{
		TotalUsers:       totalUsers,
		OnlineUsers:      onlineUsers,
		OfflineUsers:     totalUsers - onlineUsers,
		ReceivedUsers:    receivedUsers,
		UnreceivedUsers:  totalUsers - receivedUsers,
		ReadUsers:        readUsers,
		UnreadUsers:      totalUsers - readUsers,
		ConfirmedUsers:   confirmedUsers,
		UnconfirmedUsers: totalUsers - confirmedUsers,
	}

	// 计算率
	if totalUsers > 0 {
		stats.ReceiveRate = float64(receivedUsers) / float64(totalUsers) * 100
		stats.ReadRate = float64(readUsers) / float64(totalUsers) * 100
		stats.ConfirmRate = float64(confirmedUsers) / float64(totalUsers) * 100
		stats.OnlineRate = float64(onlineUsers) / float64(totalUsers) * 100
	}

	return stats, nil
}

// 辅助函数：从BSON结果中获取int64值
func getInt64FromBson(result bson.M, key string) int64 {
	if val, ok := result[key]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case int32:
			return int64(v)
		case int:
			return int64(v)
		default:
			return 0
		}
	}
	return 0
}

// forceUpdateUserOfflineStatus 强制更新用户离线状态（避免递归调用）
func (s *NotificationRecordService) forceUpdateUserOfflineStatus(userID int, username string) {
	collection := mongodb.GetCollection("notification_log_db", "admin_user_online_status")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := utils.GetCurrentTimeForMongo()
	filter := bson.M{"user_id": userID}
	updates := bson.M{
		"is_online":     false,
		"last_seen":     now,
		"connection_id": "",
		"client_ip":     "",
		"user_agent":    "",
		"updated_at":    now,
		"offline_time":  now,
	}

	_, err := collection.UpdateOne(ctx, filter, bson.M{"$set": updates})
	if err != nil {
		log.Printf("强制更新用户离线状态失败: UserID=%d, Error=%v", userID, err)
	} else {
		log.Printf("强制更新用户离线状态成功: UserID=%d", userID)
	}
}

// getUserCurrentOnlineStatus 获取用户当前在线状态
func (s *NotificationRecordService) getUserCurrentOnlineStatus(userID int) bool {
	collection := mongodb.GetCollection("notification_log_db", "admin_user_online_status")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}
	var status admin_model.AdminUserOnlineStatus
	err := collection.FindOne(ctx, filter).Decode(&status)
	if err != nil {
		return false
	}

	return status.IsOnline
}
