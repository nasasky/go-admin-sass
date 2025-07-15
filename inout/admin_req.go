package inout

import (
	"time"
)

type AddTenantsReq struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
	Phone    string `form:"phone" binding:"required"`
	Type     int    `form:"type" binding:"required"`
	RoleId   int    `form:"role_id" binding:"required"`
}

type UpdateTenantsReq struct {
	Id       int    `form:"id" binding:"required"`
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
	Phone    string `form:"phone" binding:"required"`
	Type     int    `form:"type" binding:"required"`
	RoleId   int    `form:"role_id" binding:"required"`
}

// UpdateUserProfileReq 修改用户信息请求
type UpdateUserProfileReq struct {
	Id       int    `json:"id" binding:"required"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
	Sex      int    `json:"sex"`
}

// UpdateUserPasswordReq 修改用户密码请求
type UpdateUserPasswordReq struct {
	Id          int    `json:"id" binding:"required"`
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type LoginAdminReq struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
	Captcha  string `form:"captcha"`
}

type LoginTenantsReq struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

type AddMenuReq struct {
	ParentId  int    `form:"parent_id"`
	Label     string `form:"label"`
	Icon      string `form:"icon"`
	Rule      string `form:"rule" `
	Key       string `form:"key" `
	Path      string `form:"path" binding:"required"`
	Type      string `form:"type" binding:"required"`
	Show      int    `form:"show"`
	Sort      int    `form:"sort"`
	Title     string `form:"title"`
	Layout    string `form:"layout"`
	KeepAlive int    `form:"keepAlive"`
}

type AddArticleReq struct {
	Title   string `form:"title" binding:"required"`
	Content string `form:"content" binding:"required"`
	Type    int    `form:"type" binding:"required"`
}

type UpdateArticleReq struct {
	Id       int    `form:"id" binding:"required"`
	Title    string `form:"title" binding:"required"`
	Content  string `form:"content" binding:"required"`
	Type     int    `form:"type" binding:"required"`
	Status   int    `form:"status"`
	Tips     string `form:"tips"`
	Isdelete int    `form:"isdelete"`
}

type GetArticleListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Seach    string `form:"seach"`
}

type MarketingItem struct {
	Id         int    `json:"id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Type       int    `json:"type"`
	UserID     int    `json:"user_id"`
	Status     int    `json:"status"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
}

type MarketingListResponse struct {
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	Items    []MarketingItem `json:"items"`
}

type GetArticleDetailReq struct {
	Id int `form:"id" binding:"required"`
}

type UpdateMenuReq struct {
	Id       int    `form:"id" binding:"required"`
	ParentId int64  `form:"parent_id"`
	Label    string `form:"label" binding:"required"`
	Icon     string `form:"icon"`
	Rule     string `form:"rule" `
	Path     string `form:"path"`
	Title    string `form:"title"`
	Key      string `form:"key" `
	Type     string `form:"type" binding:"required"`
	Show     int    `form:"show"`
	Sort     int    `form:"sort"`
}

type SettingReq struct {
	Appid  string `form:"appid" binding:"required"`
	Secret string `form:"secret" binding:"required"`
	Name   string `form:"name" binding:"required"`
	Tips   string `form:"tips"`
	Type   string `form:"type" binding:"required"`
}

type UpdateSettingReq struct {
	Id     int    `form:"id" binding:"required"`
	Appid  string `form:"appid" binding:"required"`
	Secret string `form:"secret" binding:"required"`
	Name   string `form:"name" binding:"required"`
	Tips   string `form:"tips"`
	Type   string `form:"type"`
}

type FeishuSendReq struct {
	Content   string `form:"content" binding:"required"`
	MsgType   string `form:"msg_type" binding:"required"`
	ReceiveId string `form:"receive_id" binding:"required"`
}

type AddEmployeeReq struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
	Phone    string `form:"phone" binding:"required"`
	RoleId   int    `form:"role_id" binding:"required"`
	Enable   string `form:"enable"`
	Sex      int    `form:"sex"`
	Avatar   string `form:"avatar"`
	UserType int    `form:"user_type"`
}

type UpdateEmployeeReq struct {
	Id       int    `json:"id" binding:"required"`
	Username string `json:"username" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	RoleId   int    `json:"role_id" binding:"required"`
	Enable   string `json:"enable"`
	UserType int    `json:"user_type"`
	Sex      int    `json:"sex"`
	Avatar   string `json:"avatar"`
}

type AddEmployeeGroupReq struct {
	Name  string `form:"name" binding:"required"`
	Rules string `form:"rules" binding:"required"`
}

type UpdateEmployeeGroupReq struct {
	Id    int    `form:"id" binding:"required"`
	Name  string `form:"name" binding:"required"`
	Rules string `form:"rules" binding:"required"`
}

type ListpageReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Search   string `form:"search"`
}

// GetMemberListReq 会员列表请求 - 支持更精确的搜索
type GetMemberListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Search   string `form:"search"`   // 通用搜索（兼容性）
	Name     string `form:"name"`     // 按姓名搜索
	Phone    string `form:"phone"`    // 按手机号搜索
	Username string `form:"username"` // 按用户名搜索
}

// ExportMemberListReq 导出会员列表请求
type ExportMemberListReq struct {
	Search    string `form:"search"`     // 通用搜索（兼容性）
	Name      string `form:"name"`       // 按姓名搜索
	Phone     string `form:"phone"`      // 按手机号搜索
	Username  string `form:"username"`   // 按用户名搜索
	StartDate string `form:"start_date"` // 开始日期 (格式: 2024-01-01)
	EndDate   string `form:"end_date"`   // 结束日期 (格式: 2024-01-01)
}
type DicteReq struct {
	Id       int    `form:"id" binding:"required"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Code     string `form:"code"`
}
type AddDictTypeReq struct {
	Type     string `form:"type" binding:"required"`
	TypeName string `form:"type_name" binding:"required"`
	TypeCode string `form:"type_code" binding:"required"`
	Remark   string `form:"remark"`
	IsLock   string `form:"is_lock"`
	DelFlag  string `form:"del_flag"`
	IsShow   string `form:"is_show"`
}
type AddDictValueReq struct {
	Id                int    `form:"id"`
	CodeName          string `form:"code_name" binding:"required"`
	Code              string `form:"code" binding:"required"`
	Alias             string `form:"alias"`
	CallbackShowStyle string `form:"callback_show_style"`
	Remark            string `form:"remark"`
	Sort              int    `form:"sort"`
	IsDefault         int    `form:"is_default"`
	IsLock            string `form:"is_lock"`
	IsShow            string `form:"is_show"`
	DelFlag           string `form:"del_flag"`
	SysDictTypeId     int    `form:"sys_dict_type_id"`
}

type GetEmployeeListResp struct {
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Items    []EmployeeItem `json:"items"`
}
type EmployeeItem struct {
	Id         int    `json:"id"`
	UserName   string `json:"username"`
	Phone      string `json:"phone"`
	RoleId     int    `json:"role_id"`
	UserType   int    `json:"user_type"`
	Enable     string `json:"enable"`
	Avatar     string `json:"avatar"`
	Sex        int    `json:"sex"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
}

type AddBannerReq struct {
	Title    string `form:"title" binding:"required"`
	ImageUrl string `form:"image_url" binding:"required"`
	LinkUrl  string `form:"link_url"`
	Sort     int    `form:"sort"`
	Status   int    `form:"status"`
}

type UpdateBannerReq struct {
	Id       int    `form:"id" binding:"required"`
	Title    string `form:"title"`
	ImageUrl string `form:"image_url"`
	LinkUrl  string `form:"link_url"`
	Sort     int    `form:"sort"`
	Status   int    `form:"status"`
}

type GetBannerListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Search   string `form:"search"`
}

type BannerItem struct {
	Id         int       `json:"id"`
	Title      string    `json:"title"`
	ImageUrl   string    `json:"image_url"`
	LinkUrl    string    `json:"link_url"`
	Sort       int       `json:"sort"`
	Status     int       `json:"status"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

type BannerListResponse struct {
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
	Items    []BannerItem `json:"items"`
}

type SystemInfoReq struct {
	SystemName  string `form:"system_name" binding:"required" json:"system_name"`   // 系统名称
	SystemTitle string `form:"system_title" binding:"required" json:"system_title"` // 系统标题
	IcpNumber   string `form:"icp_number" json:"icp_number"`                        // 备案号
	Copyright   string `form:"copyright" json:"copyright"`                          // 版权信息
	Status      int    `form:"status" json:"status"`                                // 状态：1-启用 0-禁用
}

type UpdateSystemInfoReq struct {
	Id          int    `form:"id" binding:"required" json:"id"`
	SystemName  string `form:"system_name" json:"system_name"`   // 系统名称
	SystemTitle string `form:"system_title" json:"system_title"` // 系统标题
	IcpNumber   string `form:"icp_number" json:"icp_number"`     // 备案号
	Copyright   string `form:"copyright" json:"copyright"`       // 版权信息
	Status      int    `form:"status" json:"status"`             // 状态：1-启用 0-禁用
}

type SystemInfoResponse struct {
	Id          int       `json:"id"`
	SystemName  string    `json:"system_name"`  // 系统名称
	SystemTitle string    `json:"system_title"` // 系统标题
	IcpNumber   string    `json:"icp_number"`   // 备案号
	Copyright   string    `json:"copyright"`    // 版权信息
	Status      int       `json:"status"`       // 状态：1-启用 0-禁用
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}

type GetSystemInfoListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Search   string `form:"search"`
	OrderBy  string `form:"order_by"`
}

type SystemInfoListResponse struct {
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
	Items    []SystemInfoResponse `json:"items"`
}

// SystemNoticeReq 系统通知请求
type SystemNoticeReq struct {
	Content string `json:"content" binding:"required"` // 通知内容
	Type    string `json:"type"`                       // 通知类型：system_notice, system_maintain, system_upgrade
	Target  string `json:"target"`                     // 推送目标：all, admin, custom
	UserIDs []int  `json:"user_ids"`                   // 当target为custom时，指定用户ID列表
}

// ========== 会员统计相关请求响应 ==========

// GetMemberStatsReq 获取会员统计数据请求
type GetMemberStatsReq struct {
	Type      string `form:"type" binding:"required,oneof=daily weekly monthly"` // 统计类型：daily-每日，weekly-每周，monthly-每月
	StartDate string `form:"start_date"`                                         // 开始日期 (格式: 2024-01-01)
	EndDate   string `form:"end_date"`                                           // 结束日期 (格式: 2024-01-01)
	Days      int    `form:"days" binding:"min=1,max=365"`                       // 查询天数，如果不提供start_date和end_date，则从当前日期往前推算
}

// MemberStatsResp 会员统计数据响应
type MemberStatsResp struct {
	Type      string                  `json:"type"`       // 统计类型
	DateRange *DateRangeInfo          `json:"date_range"` // 日期范围信息
	Summary   *MemberStatsSummary     `json:"summary"`    // 汇总统计信息
	ChartData []MemberStatsChartPoint `json:"chart_data"` // 折线图数据点
	TrendInfo *MemberTrendInfo        `json:"trend_info"` // 趋势信息
	UpdatedAt string                  `json:"updated_at"` // 数据更新时间
}

// DateRangeInfo 日期范围信息
type DateRangeInfo struct {
	StartDate   string `json:"start_date"`   // 开始日期
	EndDate     string `json:"end_date"`     // 结束日期
	TotalDays   int    `json:"total_days"`   // 总天数
	CurrentDate string `json:"current_date"` // 当前日期
}

// MemberStatsSummary 会员统计汇总信息
type MemberStatsSummary struct {
	TotalMembers        int     `json:"total_members"`          // 总会员数
	NewMembersInPeriod  int     `json:"new_members_in_period"`  // 周期内新增会员数
	AvgDailyNewMembers  float64 `json:"avg_daily_new_members"`  // 日均新增会员数
	PeakNewMembersDay   string  `json:"peak_new_members_day"`   // 新增会员最多的一天
	PeakNewMembersCount int     `json:"peak_new_members_count"` // 新增会员最多的一天的数量
	GrowthRate          float64 `json:"growth_rate"`            // 增长率(%)
}

// MemberStatsChartPoint 折线图数据点
type MemberStatsChartPoint struct {
	Date          string `json:"date"`           // 日期（X轴）
	NewMembers    int    `json:"new_members"`    // 新增会员数（Y轴1）
	TotalMembers  int    `json:"total_members"`  // 累计会员数（Y轴2）
	DayOfWeek     string `json:"day_of_week"`    // 星期几
	FormattedDate string `json:"formatted_date"` // 格式化的日期显示
}

// MemberTrendInfo 会员趋势信息
type MemberTrendInfo struct {
	Trend          string  `json:"trend"`            // 趋势：up-上升，down-下降，stable-稳定
	TrendPercent   float64 `json:"trend_percent"`    // 趋势百分比
	TrendDesc      string  `json:"trend_desc"`       // 趋势描述
	ComparedToPrev string  `json:"compared_to_prev"` // 与前一个周期对比
}

// ProcessContentReq 处理内容请求
type ProcessContentReq struct {
	Content string `form:"content" binding:"required" json:"content"` // 内容参数
}

// ProcessContentResp 处理内容响应
type ProcessContentResp struct {
	Message     string `json:"message"`      // 处理结果消息
	ProcessTime string `json:"process_time"` // 处理时间
	ContentInfo struct {
		Length      int    `json:"length"`       // 内容长度
		WordCount   int    `json:"word_count"`   // 单词数量
		CharCount   int    `json:"char_count"`   // 字符数量
		ProcessedAt string `json:"processed_at"` // 处理时间戳
	} `json:"content_info"` // 内容信息
}

// BaiduHotSearchItem 百度热搜项目
type BaiduHotSearchItem struct {
	Rank       int    `json:"rank"`        // 排名
	Title      string `json:"title"`       // 标题
	HotValue   string `json:"hot_value"`   // 热度值
	Link       string `json:"link"`        // 链接
	Tag        string `json:"tag"`         // 标签（热、新等）
	Desc       string `json:"desc"`        // 描述
	ImageUrl   string `json:"image_url"`   // 图片链接
	UpdateTime string `json:"update_time"` // 更新时间
}

// BaiduHotSearchResp 百度热搜响应
type BaiduHotSearchResp struct {
	Code       int                  `json:"code"`        // 状态码
	Message    string               `json:"message"`     // 消息
	Data       []BaiduHotSearchItem `json:"data"`        // 热搜数据
	Total      int                  `json:"total"`       // 总数
	UpdateTime string               `json:"update_time"` // 更新时间
}

// GetBaiduHotSearchReq 获取百度热搜请求
type GetBaiduHotSearchReq struct {
	Count int `form:"count" json:"count"` // 获取数量，默认20
}
