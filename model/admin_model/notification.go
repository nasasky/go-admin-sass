package admin_model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PushRecord 推送记录
type PushRecord struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MessageID       string             `bson:"message_id" json:"message_id"`                               // 消息唯一ID
	Content         string             `bson:"content" json:"content"`                                     // 推送内容
	MessageType     string             `bson:"message_type" json:"message_type"`                           // 消息类型
	Target          string             `bson:"target" json:"target"`                                       // 推送目标
	TargetUserIDs   []int              `bson:"target_user_ids,omitempty" json:"target_user_ids,omitempty"` // 目标用户ID列表
	RecipientsCount string             `bson:"recipients_count" json:"recipients_count"`                   // 接收者数量描述
	Status          string             `bson:"status" json:"status"`                                       // 推送状态：delivered, failed
	Success         bool               `bson:"success" json:"success"`                                     // 是否成功
	Error           string             `bson:"error,omitempty" json:"error,omitempty"`                     // 错误信息
	ErrorCode       string             `bson:"error_code,omitempty" json:"error_code,omitempty"`           // 错误代码
	PushTime        string             `bson:"push_time" json:"push_time"`                                 // 推送时间
	CreatedAt       string             `bson:"created_at" json:"created_at"`                               // 创建时间
	UpdatedAt       string             `bson:"updated_at" json:"updated_at"`                               // 更新时间

	// 发送者信息
	SenderID   int    `bson:"sender_id" json:"sender_id"`     // 发送者ID
	SenderName string `bson:"sender_name" json:"sender_name"` // 发送者名称

	// 统计信息
	DeliveredCount int64 `bson:"delivered_count" json:"delivered_count"` // 实际送达数量
	FailedCount    int64 `bson:"failed_count" json:"failed_count"`       // 失败数量
	TotalCount     int64 `bson:"total_count" json:"total_count"`         // 总数量

	// 扩展信息
	Priority    int                    `bson:"priority" json:"priority"`                         // 优先级
	NeedConfirm bool                   `bson:"need_confirm" json:"need_confirm"`                 // 是否需要确认
	ExtraData   map[string]interface{} `bson:"extra_data,omitempty" json:"extra_data,omitempty"` // 扩展数据

	// 新增：是否在线
	IsOnline bool `json:"is_online"`
}

// NotificationLog 通知日志（详细记录）
type NotificationLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MessageID string             `bson:"message_id" json:"message_id"` // 关联的消息ID
	UserID    int                `bson:"user_id" json:"user_id"`       // 用户ID
	Username  string             `bson:"username" json:"username"`     // 用户名
	EventType string             `bson:"event_type" json:"event_type"` // 事件类型：sent, delivered, failed, read, confirmed
	Status    string             `bson:"status" json:"status"`         // 状态
	Timestamp string             `bson:"timestamp" json:"timestamp"`   // 时间戳
	CreatedAt string             `bson:"created_at" json:"created_at"` // 创建时间

	// 详细信息
	Error      string                 `bson:"error,omitempty" json:"error,omitempty"`             // 错误信息
	DeviceInfo map[string]interface{} `bson:"device_info,omitempty" json:"device_info,omitempty"` // 设备信息
	ClientIP   string                 `bson:"client_ip,omitempty" json:"client_ip,omitempty"`     // 客户端IP
	UserAgent  string                 `bson:"user_agent,omitempty" json:"user_agent,omitempty"`   // 用户代理
}

// PushRecordStats 推送记录统计
type PushRecordStats struct {
	TotalRecords    int64 `json:"total_records"`    // 总记录数
	SuccessRecords  int64 `json:"success_records"`  // 成功记录数
	FailedRecords   int64 `json:"failed_records"`   // 失败记录数
	TotalRecipients int64 `json:"total_recipients"` // 总接收者数
	DeliveredCount  int64 `json:"delivered_count"`  // 实际送达数
	FailedCount     int64 `json:"failed_count"`     // 失败数
}

// PushRecordQuery 推送记录查询条件
type PushRecordQuery struct {
	Page        int    `json:"page" form:"page"`
	PageSize    int    `json:"page_size" form:"page_size" binding:"max=100"`
	MessageType string `json:"message_type" form:"message_type"`
	Target      string `json:"target" form:"target"`
	Status      string `json:"status" form:"status"`
	Success     *bool  `json:"success" form:"success"`
	SenderID    int    `json:"sender_id" form:"sender_id"`
	StartDate   string `json:"start_date" form:"start_date"`
	EndDate     string `json:"end_date" form:"end_date"`
	Keyword     string `json:"keyword" form:"keyword"`
	SortBy      string `json:"sort_by" form:"sort_by"`
	SortOrder   string `json:"sort_order" form:"sort_order"`
}

// PushRecordListResponse 推送记录列表响应
type PushRecordListResponse struct {
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	Items    []PushRecord    `json:"items"`
	Stats    PushRecordStats `json:"stats"`
}

// AdminUserReceiveRecord 管理员用户接收记录（新增）
type AdminUserReceiveRecord struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MessageID string             `bson:"message_id" json:"message_id"` // 消息ID
	UserID    int                `bson:"user_id" json:"user_id"`       // 管理员用户ID
	Username  string             `bson:"username" json:"username"`     // 用户名
	UserRole  string             `bson:"user_role" json:"user_role"`   // 用户角色
	IsOnline  bool               `bson:"is_online" json:"is_online"`   // 推送时是否在线

	// 接收状态
	IsReceived  bool   `bson:"is_received" json:"is_received"`   // 是否接收到
	ReceivedAt  string `bson:"received_at" json:"received_at"`   // 接收时间
	IsRead      bool   `bson:"is_read" json:"is_read"`           // 是否已读
	ReadAt      string `bson:"read_at" json:"read_at"`           // 阅读时间
	IsConfirmed bool   `bson:"is_confirmed" json:"is_confirmed"` // 是否确认
	ConfirmedAt string `bson:"confirmed_at" json:"confirmed_at"` // 确认时间

	// 设备和环境信息
	DeviceType   string `bson:"device_type" json:"device_type"`     // 设备类型：desktop, mobile, tablet
	Platform     string `bson:"platform" json:"platform"`           // 平台：windows, mac, ios, android
	Browser      string `bson:"browser" json:"browser"`             // 浏览器
	ClientIP     string `bson:"client_ip" json:"client_ip"`         // 客户端IP
	UserAgent    string `bson:"user_agent" json:"user_agent"`       // 用户代理
	ConnectionID string `bson:"connection_id" json:"connection_id"` // WebSocket连接ID

	// 推送信息
	PushChannel    string `bson:"push_channel" json:"push_channel"`       // 推送渠道：websocket, offline
	DeliveryStatus string `bson:"delivery_status" json:"delivery_status"` // 投递状态：delivered, failed, pending
	RetryCount     int    `bson:"retry_count" json:"retry_count"`         // 重试次数
	ErrorMessage   string `bson:"error_message" json:"error_message"`     // 错误信息

	// 时间记录
	CreatedAt string `bson:"created_at" json:"created_at"` // 创建时间
	UpdatedAt string `bson:"updated_at" json:"updated_at"` // 更新时间
}

// AdminUserOnlineStatus 管理员用户在线状态（新增）
type AdminUserOnlineStatus struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID         int                `bson:"user_id" json:"user_id"`                 // 用户ID
	Username       string             `bson:"username" json:"username"`               // 用户名
	IsOnline       bool               `bson:"is_online" json:"is_online"`             // 是否在线
	LastSeen       string             `bson:"last_seen" json:"last_seen"`             // 最后在线时间
	OnlineTime     string             `bson:"online_time" json:"online_time"`         // 上线时间
	OfflineTime    string             `bson:"offline_time" json:"offline_time"`       // 下线时间
	OnlineDuration int64              `bson:"online_duration" json:"online_duration"` // 在线时长（秒）

	// 连接信息
	ConnectionID string                 `bson:"connection_id" json:"connection_id"` // 当前连接ID
	ClientIP     string                 `bson:"client_ip" json:"client_ip"`         // 客户端IP
	UserAgent    string                 `bson:"user_agent" json:"user_agent"`       // 用户代理
	DeviceInfo   map[string]interface{} `bson:"device_info" json:"device_info"`     // 设备信息

	// 统计信息
	TotalOnlineCount int64 `bson:"total_online_count" json:"total_online_count"` // 总上线次数
	TotalOnlineTime  int64 `bson:"total_online_time" json:"total_online_time"`   // 总在线时长

	CreatedAt string `bson:"created_at" json:"created_at"` // 创建时间
	UpdatedAt string `bson:"updated_at" json:"updated_at"` // 更新时间
}

// AdminUserReceiveQuery 管理员用户接收记录查询条件（新增）
type AdminUserReceiveQuery struct {
	Page           int    `json:"page" form:"page"`
	PageSize       int    `json:"page_size" form:"page_size" binding:"max=100"`
	MessageID      string `json:"message_id" form:"message_id"`
	UserID         int    `json:"user_id" form:"user_id"`
	Username       string `json:"username" form:"username"`
	IsOnline       *bool  `json:"is_online" form:"is_online"`
	IsReceived     *bool  `json:"is_received" form:"is_received"`
	IsRead         *bool  `json:"is_read" form:"is_read"`
	IsConfirmed    *bool  `json:"is_confirmed" form:"is_confirmed"`
	DeliveryStatus string `json:"delivery_status" form:"delivery_status"`
	PushChannel    string `json:"push_channel" form:"push_channel"`
	StartDate      string `json:"start_date" form:"start_date"`
	EndDate        string `json:"end_date" form:"end_date"`
	SortBy         string `json:"sort_by" form:"sort_by"`
	SortOrder      string `json:"sort_order" form:"sort_order"`
}

// AdminUserReceiveStats 管理员用户接收统计（新增）
type AdminUserReceiveStats struct {
	TotalUsers       int64   `json:"total_users"`       // 总用户数
	OnlineUsers      int64   `json:"online_users"`      // 在线用户数
	OfflineUsers     int64   `json:"offline_users"`     // 离线用户数
	ReceivedUsers    int64   `json:"received_users"`    // 已接收用户数
	UnreceivedUsers  int64   `json:"unreceived_users"`  // 未接收用户数
	ReadUsers        int64   `json:"read_users"`        // 已读用户数
	UnreadUsers      int64   `json:"unread_users"`      // 未读用户数
	ConfirmedUsers   int64   `json:"confirmed_users"`   // 已确认用户数
	UnconfirmedUsers int64   `json:"unconfirmed_users"` // 未确认用户数
	ReceiveRate      float64 `json:"receive_rate"`      // 接收率
	ReadRate         float64 `json:"read_rate"`         // 阅读率
	ConfirmRate      float64 `json:"confirm_rate"`      // 确认率
	OnlineRate       float64 `json:"online_rate"`       // 在线率
	AvgResponseTime  float64 `json:"avg_response_time"` // 平均响应时间（秒）
}
