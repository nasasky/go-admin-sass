package inout

// GetSystemLogReq 获取系统日志请求
type GetSystemLogReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Keyword  string `form:"keyword"`
}
