package admin

import (
	"fmt"
	"log"
	"nasa-go-admin/inout"
	"nasa-go-admin/model/admin_model"
	"nasa-go-admin/services/admin_service"
	"nasa-go-admin/services/public_service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var systemService = &admin_service.SystemService{}

// GetSystemLog 从 MongoDB 中读取日志数据
// GetSystemLog 从 MongoDB 中读取日志数据
func GetSystemLog(c *gin.Context) {
	var req inout.GetSystemLogReq
	if err := c.ShouldBind(&req); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// 设置默认值已在服务层处理，这里可以省略

	// 调用修改后的服务方法
	result, err := systemService.GetSystemLog(req, "admin_log_db")
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// 使用统一响应方法
	Resp.Succ(c, result)
}

// GetUserLog
func GetUserLog(c *gin.Context) {
	var req inout.GetUserLogReq
	if err := c.ShouldBind(&req); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// 调用用户日志服务方法
	result, err := systemService.GetUserLog(req)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// 使用统一响应方法
	Resp.Succ(c, result)
}

func GetDictType(c *gin.Context) {
	var params inout.ListpageReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	list, err := systemService.GetDictTypeList(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, list)

}

// AddDictType
func AddDictType(c *gin.Context) {
	var params inout.AddDictTypeReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	err := systemService.AddDictType(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, nil)
}

// AddDictValue
func AddDictValue(c *gin.Context) {
	var params inout.AddDictValueReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	err := systemService.AddDictValue(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, nil)
}

// GetDictDetail 获取系统字典参数详情
func GetDictDetail(c *gin.Context) {
	var params inout.DicteReq
	if err := c.ShouldBind(&params); err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	list, err := systemService.GetDictDetailData(c, params)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, list)
}

// GetAllDictType 获取所有字典类型和字典值
func GetAllDictType(c *gin.Context) {
	list, err := systemService.GetAllDictType(c)
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}
	Resp.Succ(c, list)
}

// PostnoticeInfo 系统公告控制器
func PostnoticeInfo(c *gin.Context) {
	var req inout.SystemNoticeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Resp.Err(c, 20001, "请求参数错误: "+err.Error())
		return
	}

	// 验证内容长度
	if len(req.Content) == 0 {
		Resp.Err(c, 20001, "通知内容不能为空")
		return
	}
	if len(req.Content) > 500 {
		Resp.Err(c, 20001, "通知内容不能超过500字符")
		return
	}

	// 获取发送者信息
	senderID := c.GetInt("uid")
	senderName := "unknown"
	if userInfo, exists := c.Get("userInfo"); exists {
		if user, ok := userInfo.(map[string]string); ok {
			if name, exists := user["username"]; exists {
				senderName = name
			}
		}
	}

	wsService := public_service.GetWebSocketService()
	var err error
	var messageID string
	var targetText string
	var recipientsCount string

	// 根据推送目标创建不同的通知消息
	switch req.Target {
	case "all":
		// 广播给所有用户
		msg := &public_service.NotificationMessage{
			Type:      public_service.SystemNotice,
			Content:   req.Content,
			Time:      time.Now().Format("2006-01-02 15:04:05"),
			Priority:  public_service.PriorityNormal,
			Target:    public_service.TargetAll,
			MessageID: generateMessageID(),
		}
		err = wsService.SendNotification(msg)
		messageID = msg.MessageID
		targetText = "all"
		recipientsCount = "all_online_users"
	case "admin":
		// 只发送给管理员
		msg := &public_service.NotificationMessage{
			Type:      public_service.SystemNotice,
			Content:   req.Content,
			Time:      time.Now().Format("2006-01-02 15:04:05"),
			Priority:  public_service.PriorityNormal,
			Target:    public_service.TargetAdmin,
			MessageID: generateMessageID(),
		}
		err = wsService.SendNotification(msg)
		messageID = msg.MessageID
		targetText = "admin"
		recipientsCount = "all_admins"
	case "custom":
		// 发送给指定用户
		if len(req.UserIDs) == 0 {
			Resp.Err(c, 20001, "自定义推送需要指定用户ID列表")
			return
		}

		msg := &public_service.NotificationMessage{
			Type:      public_service.SystemNotice,
			Content:   req.Content,
			Time:      time.Now().Format("2006-01-02 15:04:05"),
			Priority:  public_service.PriorityNormal,
			Target:    public_service.TargetCustom,
			TargetIDs: req.UserIDs,
			MessageID: generateMessageID(),
		}
		err = wsService.SendNotification(msg)
		messageID = msg.MessageID
		targetText = "custom"
		recipientsCount = fmt.Sprintf("%d_users", len(req.UserIDs))
	default:
		// 默认广播给所有用户
		err = wsService.BroadcastSystemNotice(req.Content)
		targetText = "all"
		recipientsCount = "all_online_users"
	}

	// 构建推送状态响应
	pushStatus := map[string]interface{}{
		"success":          err == nil,
		"message":          req.Content,
		"push_time":        time.Now().Format("2006-01-02 15:04:05"),
		"target":           targetText,
		"message_type":     req.Type,
		"recipients_count": recipientsCount,
	}

	if messageID != "" {
		pushStatus["message_id"] = messageID
	}

	// 保存推送记录到MongoDB
	recordService := admin_service.NewNotificationRecordService()
	pushRecord := &admin_model.PushRecord{
		MessageID:       messageID,
		Content:         req.Content,
		MessageType:     req.Type,
		Target:          targetText,
		TargetUserIDs:   req.UserIDs,
		RecipientsCount: recipientsCount,
		Status:          "delivered",
		Success:         err == nil,
		PushTime:        time.Now().Format("2006-01-02 15:04:05"),
		SenderID:        senderID,
		SenderName:      senderName,
		Priority:        1, // 普通优先级
		NeedConfirm:     false,
	}

	if err != nil {
		// 推送失败
		pushStatus["error"] = err.Error()
		pushStatus["error_code"] = "PUSH_FAILED"
		pushStatus["status"] = "failed"
		pushRecord.Status = "failed"
		pushRecord.Error = err.Error()
		pushRecord.ErrorCode = "PUSH_FAILED"

		log.Printf("发送系统公告失败: %v", err)

		// 保存失败记录
		go func() {
			if saveErr := recordService.SavePushRecord(pushRecord); saveErr != nil {
				log.Printf("保存推送记录失败: %v", saveErr)
			}
		}()

		Resp.Err(c, 20001, "推送失败: "+err.Error())
		return
	}

	// 推送成功
	pushStatus["status"] = "delivered"
	log.Printf("系统公告推送成功: %s, 目标: %s", req.Content, targetText)

	// 保存成功记录
	go func() {
		if saveErr := recordService.SavePushRecord(pushRecord); saveErr != nil {
			log.Printf("保存推送记录失败: %v", saveErr)
		}
	}()

	Resp.Succ(c, pushStatus)
}

// generateMessageID 生成消息ID的辅助函数
func generateMessageID() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), uuid.New().String()[:8])
}

// 群组控制器
// func (c *GroupController) NotifyNewMember(ctx *gin.Context) {
//     groupID := 789
//     newMemberName := "张三"

//     data := map[string]interface{}{
//         "new_member_id": 123,
//         "new_member_name": newMemberName,
//         "join_time": time.Now().Format("2006-01-02 15:04:05"),
//     }

//     wsService := public_service.GetWebSocketService()
//     err := wsService.SendGroupNotification(
//         groupID,
//         fmt.Sprintf("%s加入了群组", newMemberName),
//         data,
//     )
//     if err != nil {
//         log.Printf("发送群组通知失败: %v", err)
//     }
// }

// 自定义复杂通知
// func SendCustomNotification() {
//     wsService := public_service.GetWebSocketService()

//     // 创建通知消息
//     msg := &public_service.NotificationMessage{
//         Type:      public_service.SystemUpgrade,
//         Content:   "系统升级提醒",
//         Data: map[string]interface{}{
//             "version": "v2.5.0",
//             "features": []string{"新增群聊功能", "性能优化"},
//             "upgrade_time": "2025-04-25 01:00:00",
//         },
//         Priority:  public_service.PriorityHigh,
//         Target:    public_service.TargetCustom,
//         TargetIDs: []int{101, 102, 103}, // VIP用户
//         ExcludeIDs: []int{105}, // 排除特定用户
//         NeedConfirm: true,      // 需要用户确认
//     }

//     // 发送通知
//     err := wsService.SendNotification(msg)
//     if err != nil {
//         log.Printf("发送自定义通知失败: %v", err)
//     }
// }

// 在定时任务中发送每日提醒
// func DailyNotificationJob() {
//     wsService := public_service.GetWebSocketService()

//     // 获取所有活跃用户
//     userIDs := getAllActiveUserIDs()

//     // 创建每日提醒消息
//     msg := &public_service.NotificationMessage{
//         Type:      public_service.SystemNotice,
//         Content:   "今日特惠活动已开始",
//         Data: map[string]interface{}{
//             "activity_id": "ACT20250421",
//             "discount": "8折",
//         },
//         Priority:  public_service.PriorityNormal,
//         Target:    public_service.TargetCustom,
//         TargetIDs: userIDs,
//     }

//     // 分批发送通知，避免一次发送过多
//     batchSize := 1000
//     for i := 0; i < len(userIDs); i += batchSize {
//         end := i + batchSize
//         if end > len(userIDs) {
//             end = len(userIDs)
//         }

//         batchMsg := *msg // 复制消息
//         batchMsg.TargetIDs = userIDs[i:end]

//         wsService.SendNotification(&batchMsg)

//         // 避免发送过快
//         time.Sleep(500 * time.Millisecond)
//     }
// }

// 用户登录拦截器
// func LoginSuccessInterceptor(userID int, username string, ctx *gin.Context) {
//     // 处理登录成功的其他逻辑...

//     // 发送登录通知
//     go func() {
//         wsService := public_service.GetWebSocketService()

//         // 通知当前用户
//         wsService.SendUserNotification(
//             userID,
//             public_service.UserLogin,
//             "欢迎回来，" + username,
//             map[string]interface{}{
//                 "last_login_time": time.Now().Format("2006-01-02 15:04:05"),
//                 "login_device": ctx.GetHeader("User-Agent"),
//             },
//         )

//         // 如果是管理员登录，还可以通知其他管理员
//         if isAdmin(userID) {
//             adminMsg := &public_service.NotificationMessage{
//                 Type:    public_service.UserLogin,
//                 Content: fmt.Sprintf("管理员 %s 已登录系统", username),
//                 Target:  public_service.TargetAdmin,
//                 ExcludeIDs: []int{userID}, // 排除自己
//             }
//             wsService.SendNotification(adminMsg)
//         }
//     }()
// }

// 用户提醒控制器
// func (c *UserController) RemindUnreadMessages(ctx *gin.Context) {
//     userID := 456
//     messageCount := 5

//     data := map[string]interface{}{
//         "count": messageCount,
//         "latest_message_id": "msg123456",
//     }

//     wsService := public_service.GetWebSocketService()
//     err := wsService.SendUserNotification(
//         userID,
//         public_service.MessageReceived,
//         fmt.Sprintf("您有%d条未读消息", messageCount),
//         data,
//     )
//     if err != nil {
//         log.Printf("发送未读消息通知失败: %v", err)
//     }
// }

// ClearSystemLog 清空系统访问日志
func ClearSystemLog(c *gin.Context) {
	// 调用服务方法清空系统日志
	result, err := systemService.ClearSystemLog("admin_log_db")
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// 使用统一响应方法
	Resp.Succ(c, result)
}

// ClearUserLog 清空用户端操作日志
func ClearUserLog(c *gin.Context) {
	// 调用服务方法清空用户日志
	result, err := systemService.ClearSystemLog("app_log_db")
	if err != nil {
		Resp.Err(c, 20001, err.Error())
		return
	}

	// 使用统一响应方法
	Resp.Succ(c, result)
}
