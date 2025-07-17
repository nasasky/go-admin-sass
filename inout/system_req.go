package inout

// GetSystemLogReq 获取系统日志请求
type GetSystemLogReq struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	Keyword    string `form:"keyword"`
	Username   string `form:"username"`    // 访问用户名搜索
	StartDate  string `form:"start_date"`  // 开始日期时间 (格式: 2024-01-01 00:00:00)
	EndDate    string `form:"end_date"`    // 结束日期时间 (格式: 2024-01-01 23:59:59)
	Method     string `form:"method"`      // HTTP方法过滤 (GET, POST, PUT, DELETE等)
	StatusCode int    `form:"status_code"` // 状态码过滤
	ClientIP   string `form:"client_ip"`   // 客户端IP过滤
	Path       string `form:"path"`        // 请求路径过滤
}

// GetUserLogReq 获取用户端操作日志请求
type GetUserLogReq struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	Keyword    string `form:"keyword"`
	Username   string `form:"username"`    // 用户名搜索
	UserID     int    `form:"user_id"`     // 用户ID搜索
	StartDate  string `form:"start_date"`  // 开始日期时间 (格式: 2024-01-01 00:00:00)
	EndDate    string `form:"end_date"`    // 结束日期时间 (格式: 2024-01-01 23:59:59)
	Method     string `form:"method"`      // HTTP方法过滤 (GET, POST, PUT, DELETE等)
	StatusCode int    `form:"status_code"` // 状态码过滤
	ClientIP   string `form:"client_ip"`   // 客户端IP过滤
	Path       string `form:"path"`        // 请求路径过滤
	DeviceType string `form:"device_type"` // 设备类型过滤 (ios, android, web, miniapp)
	AppVersion string `form:"app_version"` // 应用版本过滤
	Action     string `form:"action"`      // 操作类型过滤 (login, register, order, etc.)
}
