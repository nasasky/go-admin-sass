package admin

import (
	"context"
	"fmt"
	"log"
	"nasa-go-admin/db"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/redis"
	"nasa-go-admin/services/admin_service"
	"nasa-go-admin/services/public_service"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GetPushRecordList 获取推送记录列表
func GetPushRecordList(c *gin.Context) {
	var query admin_model.PushRecordQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		Resp.Err(c, 20001, "请求参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10 // 默认10条记录
	}
	if query.SortBy == "" {
		query.SortBy = "push_time"
	}
	if query.SortOrder == "" {
		query.SortOrder = "desc"
	}

	recordService := admin_service.NewNotificationRecordService()
	result, err := recordService.GetPushRecordList(&query)
	if err != nil {
		Resp.Err(c, 20001, "获取推送记录失败: "+err.Error())
		return
	}

	Resp.Succ(c, result)
}

// GetPushRecordDetail 获取推送记录详情
func GetPushRecordDetail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		Resp.Err(c, 20001, "记录ID不能为空")
		return
	}

	recordService := admin_service.NewNotificationRecordService()
	record, err := recordService.GetPushRecordByID(id)
	if err != nil {
		Resp.Err(c, 20001, "获取推送记录详情失败: "+err.Error())
		return
	}

	// 获取相关的通知日志
	logs, err := recordService.GetNotificationLogs(record.MessageID, 50)
	if err != nil {
		// 不返回错误，继续返回记录详情
		log.Printf("获取通知日志失败: %v", err)
	}

	detail := map[string]interface{}{
		"record": record,
		"logs":   logs,
	}

	Resp.Succ(c, detail)
}

// DeletePushRecord 删除推送记录
func DeletePushRecord(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		Resp.Err(c, 20001, "记录ID不能为空")
		return
	}

	recordService := admin_service.NewNotificationRecordService()
	err := recordService.DeletePushRecord(id)
	if err != nil {
		Resp.Err(c, 20001, "删除推送记录失败: "+err.Error())
		return
	}

	Resp.Succ(c, nil)
}

// GetPushRecordStats 获取推送记录统计
func GetPushRecordStats(c *gin.Context) {
	var query admin_model.PushRecordQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		Resp.Err(c, 20001, "请求参数错误: "+err.Error())
		return
	}

	recordService := admin_service.NewNotificationRecordService()
	result, err := recordService.GetPushRecordList(&query)
	if err != nil {
		Resp.Err(c, 20001, "获取推送记录统计失败: "+err.Error())
		return
	}

	Resp.Succ(c, result.Stats)
}

// ResendNotification 重新发送通知
func ResendNotification(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		Resp.Err(c, 20001, "记录ID不能为空")
		return
	}

	recordService := admin_service.NewNotificationRecordService()
	record, err := recordService.GetPushRecordByID(id)
	if err != nil {
		Resp.Err(c, 20001, "获取推送记录失败: "+err.Error())
		return
	}

	// 检查记录状态
	if record.Success {
		Resp.Err(c, 20001, "该推送记录已成功，无需重新发送")
		return
	}

	// 重新发送通知
	wsService := public_service.GetWebSocketService()
	var sendErr error

	switch record.Target {
	case "all":
		sendErr = wsService.BroadcastSystemNotice(record.Content)
	case "admin":
		msg := &public_service.NotificationMessage{
			Type:      public_service.SystemNotice,
			Content:   record.Content,
			Time:      time.Now().Format("2006-01-02 15:04:05"),
			Priority:  public_service.PriorityNormal,
			Target:    public_service.TargetAdmin,
			MessageID: generateMessageID(),
		}
		sendErr = wsService.SendNotification(msg)
	case "custom":
		msg := &public_service.NotificationMessage{
			Type:      public_service.SystemNotice,
			Content:   record.Content,
			Time:      time.Now().Format("2006-01-02 15:04:05"),
			Priority:  public_service.PriorityNormal,
			Target:    public_service.TargetCustom,
			TargetIDs: record.TargetUserIDs,
			MessageID: generateMessageID(),
		}
		sendErr = wsService.SendNotification(msg)
	default:
		Resp.Err(c, 20001, "不支持的推送目标类型")
		return
	}

	if sendErr != nil {
		Resp.Err(c, 20001, "重新发送失败: "+sendErr.Error())
		return
	}

	// 更新记录状态
	updates := map[string]interface{}{
		"success":   true,
		"status":    "delivered",
		"push_time": time.Now().Format("2006-01-02 15:04:05"),
	}

	// 清除错误信息
	updates["error"] = ""
	updates["error_code"] = ""

	err = recordService.UpdatePushRecord(record.MessageID, updates)
	if err != nil {
		log.Printf("更新推送记录状态失败: %v", err)
	}

	Resp.Succ(c, map[string]interface{}{
		"message": "重新发送成功",
		"status":  "delivered",
	})
}

// ========== 新增：管理员用户接收记录管理 ==========

// GetAdminUserReceiveRecords 获取管理员用户接收记录列表
func GetAdminUserReceiveRecords(c *gin.Context) {
	var query admin_model.AdminUserReceiveQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		Resp.Err(c, 20001, "请求参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	if query.SortBy == "" {
		query.SortBy = "created_at"
	}
	if query.SortOrder == "" {
		query.SortOrder = "desc"
	}

	// 获取当前用户信息
	currentUserID := c.GetInt("uid")
	if currentUserID == 0 {
		Resp.Err(c, 20001, "用户ID无效")
		return
	}

	// 查询当前用户的user_type
	var currentUser admin_model.AdminUser
	if err := db.Dao.Where("id = ?", currentUserID).First(&currentUser).Error; err != nil {
		Resp.Err(c, 20001, "获取用户信息失败")
		return
	}

	// 如果当前用户不是管理员（user_type != 1），则只能查看自己的消息记录
	if currentUser.UserType != 1 {
		// 强制设置查询条件为用户自己的ID
		query.UserID = currentUserID
		log.Printf("非管理员用户 %d 只能查看自己的消息记录", currentUserID)
	}

	recordService := admin_service.NewNotificationRecordService()
	result, err := recordService.GetAdminUserReceiveRecords(&query)
	if err != nil {
		Resp.Err(c, 20001, "获取管理员用户接收记录失败: "+err.Error())
		return
	}

	// 将result的数据直接合并到响应中，避免过深的嵌套
	response := map[string]interface{}{
		"total":     result.Total,
		"page":      result.Page,
		"page_size": result.PageSize,
		"items":     result.Items,
		"stats":     result.Stats,
		"permission_info": map[string]interface{}{
			"current_user_id":   currentUserID,
			"current_user_type": currentUser.UserType,
			"is_admin":          currentUser.UserType == 1,
			"can_view_all":      currentUser.UserType == 1,
		},
	}

	Resp.Succ(c, response)
}

// GetMessageReceiveStatus 获取消息的接收状态统计
func GetMessageReceiveStatus(c *gin.Context) {
	messageID := c.Param("messageID")
	if messageID == "" {
		Resp.Err(c, 20001, "消息ID不能为空")
		return
	}

	recordService := admin_service.NewNotificationRecordService()
	status, err := recordService.GetMessageReceiveStatus(messageID)
	if err != nil {
		Resp.Err(c, 20001, "获取消息接收状态失败: "+err.Error())
		return
	}

	Resp.Succ(c, status)
}

// MarkMessageAsRead 标记消息为已读
func MarkMessageAsRead(c *gin.Context) {
	var req struct {
		MessageID string `json:"message_id" binding:"required"`
		UserID    int    `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Resp.Err(c, 20001, "请求参数错误: "+err.Error())
		return
	}

	recordService := admin_service.NewNotificationRecordService()
	err := recordService.MarkMessageAsRead(req.MessageID, req.UserID)
	if err != nil {
		Resp.Err(c, 20001, "标记消息已读失败: "+err.Error())
		return
	}

	Resp.Succ(c, map[string]interface{}{
		"message": "消息已标记为已读",
		"success": true,
	})
}

// MarkMessageAsConfirmed 标记消息为已确认
func MarkMessageAsConfirmed(c *gin.Context) {
	var req struct {
		MessageID string `json:"message_id" binding:"required"`
		UserID    int    `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Resp.Err(c, 20001, "请求参数错误: "+err.Error())
		return
	}

	recordService := admin_service.NewNotificationRecordService()
	err := recordService.MarkMessageAsConfirmed(req.MessageID, req.UserID)
	if err != nil {
		Resp.Err(c, 20001, "标记消息确认失败: "+err.Error())
		return
	}

	Resp.Succ(c, map[string]interface{}{
		"message": "消息已标记为已确认",
		"success": true,
	})
}

// GetOnlineAdminUsers 获取在线的管理员用户列表
func GetOnlineAdminUsers(c *gin.Context) {
	recordService := admin_service.NewNotificationRecordService()
	users, err := recordService.GetOnlineAdminUsers()
	if err != nil {
		Resp.Err(c, 20001, "获取在线管理员用户失败: "+err.Error())
		return
	}

	// 构建响应数据
	response := map[string]interface{}{
		"total_online": len(users),
		"users":        users,
		"timestamp":    time.Now().Format("2006-01-02 15:04:05"),
	}

	Resp.Succ(c, response)
}

// GetAdminUserReceiveStats 获取管理员用户接收统计（按时间段）
func GetAdminUserReceiveStats(c *gin.Context) {
	var query admin_model.AdminUserReceiveQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		Resp.Err(c, 20001, "请求参数错误: "+err.Error())
		return
	}

	recordService := admin_service.NewNotificationRecordService()
	result, err := recordService.GetAdminUserReceiveRecords(&query)
	if err != nil {
		Resp.Err(c, 20001, "获取管理员用户接收统计失败: "+err.Error())
		return
	}

	// 只返回统计数据
	Resp.Succ(c, result.Stats)
}

// BatchMarkAsRead 批量标记消息为已读
func BatchMarkAsRead(c *gin.Context) {
	var req struct {
		MessageIDs []string `json:"message_ids" binding:"required"`
		UserID     int      `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Resp.Err(c, 20001, "请求参数错误: "+err.Error())
		return
	}

	recordService := admin_service.NewNotificationRecordService()

	successCount := 0
	failedCount := 0
	var errors []string

	for _, messageID := range req.MessageIDs {
		err := recordService.MarkMessageAsRead(messageID, req.UserID)
		if err != nil {
			failedCount++
			errors = append(errors, fmt.Sprintf("MessageID %s: %s", messageID, err.Error()))
		} else {
			successCount++
		}
	}

	response := map[string]interface{}{
		"success_count": successCount,
		"failed_count":  failedCount,
		"total_count":   len(req.MessageIDs),
		"success_rate":  float64(successCount) / float64(len(req.MessageIDs)) * 100,
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	if failedCount > 0 && successCount == 0 {
		Resp.Err(c, 20001, "批量标记失败")
	} else {
		Resp.Succ(c, response)
	}
}

// GetUserNotificationSummary 获取用户消息摘要
func GetUserNotificationSummary(c *gin.Context) {
	userID := c.GetInt("uid")
	if userID == 0 {
		Resp.Err(c, 20001, "用户ID无效")
		return
	}

	// 这里可以添加获取用户消息摘要的逻辑
	// 包括未读消息数、最近消息等

	response := map[string]interface{}{
		"user_id":           userID,
		"unread_count":      0,
		"total_count":       0,
		"last_message_time": "",
		"online_status":     true,
	}

	Resp.Succ(c, response)
}

// GetOfflineMessages 获取用户的离线消息
func GetOfflineMessages(c *gin.Context) {
	// 获取要查询的用户ID，默认为当前登录用户
	queryUserID := c.GetInt("uid")
	if queryUserID == 0 {
		Resp.Err(c, 20001, "用户ID无效")
		return
	}

	// 如果URL参数中指定了用户ID，则使用指定的用户ID（需要权限验证）
	if urlUserID := c.Query("user_id"); urlUserID != "" {
		if parsedUserID, err := strconv.Atoi(urlUserID); err == nil {
			// 这里可以添加权限验证，确保只能查询有权限的用户
			queryUserID = parsedUserID
		}
	}

	// 获取离线消息服务
	offlineService := public_service.NewOfflineMessageService()

	// 获取离线消息
	offlineMessages, err := offlineService.GetOfflineMessages(queryUserID)
	if err != nil {
		Resp.Err(c, 20001, "获取离线消息失败: "+err.Error())
		return
	}

	// 获取离线消息数量
	messageCount, err := offlineService.GetOfflineMessageCount(queryUserID)
	if err != nil {
		log.Printf("获取离线消息数量失败: %v", err)
		messageCount = 0
	}

	// 构建响应数据
	response := map[string]interface{}{
		"user_id":          queryUserID,
		"total_count":      messageCount,
		"offline_messages": offlineMessages,
		"query_time":       time.Now().Format("2006-01-02 15:04:05"),
	}

	Resp.Succ(c, response)
}

// ClearOfflineMessages 清除用户的离线消息
func ClearOfflineMessages(c *gin.Context) {
	// 获取要清除的用户ID，默认为当前登录用户
	queryUserID := c.GetInt("uid")
	if queryUserID == 0 {
		Resp.Err(c, 20001, "用户ID无效")
		return
	}

	// 如果URL参数中指定了用户ID，则使用指定的用户ID（需要权限验证）
	if urlUserID := c.Query("user_id"); urlUserID != "" {
		if parsedUserID, err := strconv.Atoi(urlUserID); err == nil {
			// 这里可以添加权限验证，确保只能清除有权限的用户
			queryUserID = parsedUserID
		}
	}

	// 获取离线消息服务
	offlineService := public_service.NewOfflineMessageService()

	// 清除离线消息
	err := offlineService.ClearOfflineMessages(queryUserID)
	if err != nil {
		Resp.Err(c, 20001, "清除离线消息失败: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"user_id":    queryUserID,
		"message":    "离线消息已清除",
		"success":    true,
		"clear_time": time.Now().Format("2006-01-02 15:04:05"),
	}

	Resp.Succ(c, response)
}

// GetAllUsersOfflineMessages 管理员接口：获取所有用户的离线消息
func GetAllUsersOfflineMessages(c *gin.Context) {
	// 获取离线消息服务
	offlineService := public_service.NewOfflineMessageService()

	// 获取所有离线消息的key
	ctx := context.Background()
	keys, err := redis.GetClient().Keys(ctx, "offline_msg:*").Result()
	if err != nil {
		Resp.Err(c, 20001, "获取离线消息keys失败: "+err.Error())
		return
	}

	var allOfflineMessages []map[string]interface{}
	totalUsers := 0
	totalMessages := 0

	for _, key := range keys {
		// 从key中提取用户ID
		userIDStr := strings.TrimPrefix(key, "offline_msg:")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			log.Printf("解析用户ID失败: key=%s, error=%v", key, err)
			continue
		}

		// 获取该用户的离线消息
		messages, err := offlineService.GetOfflineMessages(userID)
		if err != nil {
			log.Printf("获取用户 %d 离线消息失败: %v", userID, err)
			continue
		}

		// 获取消息数量
		count, err := offlineService.GetOfflineMessageCount(userID)
		if err != nil {
			log.Printf("获取用户 %d 离线消息数量失败: %v", userID, err)
			count = 0
		}

		if count > 0 {
			// 获取用户名
			username := "unknown"
			var adminUser admin_model.AdminUser
			if err := db.Dao.Where("id = ?", userID).First(&adminUser).Error; err == nil {
				username = adminUser.Username
			}

			userOfflineData := map[string]interface{}{
				"user_id":          userID,
				"username":         username,
				"message_count":    count,
				"offline_messages": messages,
				"last_updated":     time.Now().Format("2006-01-02 15:04:05"),
			}

			allOfflineMessages = append(allOfflineMessages, userOfflineData)
			totalUsers++
			totalMessages += int(count)
		}
	}

	// 构建响应数据
	response := map[string]interface{}{
		"total_users":        totalUsers,
		"total_messages":     totalMessages,
		"users_offline_data": allOfflineMessages,
		"query_time":         time.Now().Format("2006-01-02 15:04:05"),
	}

	Resp.Succ(c, response)
}

// ClearAllUsersOfflineMessages 管理员接口：清除所有用户的离线消息
func ClearAllUsersOfflineMessages(c *gin.Context) {
	// 获取离线消息服务
	offlineService := public_service.NewOfflineMessageService()

	// 获取所有离线消息的key
	ctx := context.Background()
	keys, err := redis.GetClient().Keys(ctx, "offline_msg:*").Result()
	if err != nil {
		Resp.Err(c, 20001, "获取离线消息keys失败: "+err.Error())
		return
	}

	successCount := 0
	failedCount := 0
	var errors []string

	for _, key := range keys {
		// 从key中提取用户ID
		userIDStr := strings.TrimPrefix(key, "offline_msg:")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			log.Printf("解析用户ID失败: key=%s, error=%v", key, err)
			failedCount++
			errors = append(errors, fmt.Sprintf("用户ID解析失败: %s", key))
			continue
		}

		// 清除该用户的离线消息
		err = offlineService.ClearOfflineMessages(userID)
		if err != nil {
			failedCount++
			errors = append(errors, fmt.Sprintf("用户 %d: %s", userID, err.Error()))
		} else {
			successCount++
		}
	}

	response := map[string]interface{}{
		"success_count": successCount,
		"failed_count":  failedCount,
		"total_users":   len(keys),
		"clear_time":    time.Now().Format("2006-01-02 15:04:05"),
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	if failedCount > 0 && successCount == 0 {
		Resp.Err(c, 20001, "清除所有离线消息失败")
	} else {
		Resp.Succ(c, response)
	}
}
