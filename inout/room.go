package inout

import (
	"time"
)

// ========== 房间管理相关请求 ==========

// CreateRoomReq 创建房间请求
type CreateRoomReq struct {
	RoomNumber  string  `json:"room_number" binding:"required" validate:"min=1,max=20"`
	RoomName    string  `json:"room_name" binding:"required" validate:"min=1,max=50"`
	RoomType    string  `json:"room_type" binding:"required,oneof=small medium large luxury"`
	Capacity    int     `json:"capacity" binding:"required,min=1,max=50"`
	HourlyRate  float64 `json:"hourly_rate" binding:"required,min=0"`
	Features    string  `json:"features"`
	Images      string  `json:"images"`
	Floor       int     `json:"floor" binding:"min=1,max=50"`
	Area        float64 `json:"area" binding:"min=0"`
	Description string  `json:"description"`
}

// UpdateRoomReq 更新房间请求
type UpdateRoomReq struct {
	ID          int     `json:"id" binding:"required"`
	RoomNumber  string  `json:"room_number" binding:"required"`
	RoomName    string  `json:"room_name" binding:"required"`
	RoomType    string  `json:"room_type" binding:"required,oneof=small medium large luxury"`
	Capacity    int     `json:"capacity" binding:"required,min=1,max=50"`
	HourlyRate  float64 `json:"hourly_rate" binding:"required,min=0"`
	Features    string  `json:"features"`
	Images      string  `json:"images"`
	Floor       int     `json:"floor" binding:"min=1,max=50"`
	Area        float64 `json:"area" binding:"min=0"`
	Description string  `json:"description"`
	Status      int     `json:"status" binding:"oneof=1 2 3 4"`
}

// RoomListReq 房间列表请求
type RoomListReq struct {
	Page        int     `json:"page" form:"page" binding:"min=1"`
	PageSize    int     `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	RoomType    string  `json:"room_type" form:"room_type"`
	Status      int     `json:"status" form:"status"`
	Floor       int     `json:"floor" form:"floor"`
	Keyword     string  `json:"keyword" form:"keyword"`
	MinPrice    float64 `json:"min_price" form:"min_price"`
	MaxPrice    float64 `json:"max_price" form:"max_price"`
	MinCapacity int     `json:"min_capacity" form:"min_capacity"`
	MaxCapacity int     `json:"max_capacity" form:"max_capacity"`
}

// UpdateRoomStatusReq 更新房间状态请求
type UpdateRoomStatusReq struct {
	ID     int `json:"id" binding:"required"`
	Status int `json:"status" binding:"required,oneof=1 2 3 4"`
}

// ========== 房间预订相关请求 ==========

// CreateBookingReq 创建预订请求
type CreateBookingReq struct {
	RoomID       int    `json:"room_id" binding:"required"`
	StartTime    string `json:"start_time" binding:"required"`
	Hours        int    `json:"hours" binding:"required,min=1,max=168"`
	PackageID    *int   `json:"package_id"`
	ContactName  string `json:"contact_name" binding:"required"`
	ContactPhone string `json:"contact_phone" binding:"required"`
	Remarks      string `json:"remarks"`
}

// UpdateBookingReq 更新预订请求
type UpdateBookingReq struct {
	ID           int    `json:"id" binding:"required"`
	RoomID       int    `json:"room_id" binding:"required"`
	StartTime    string `json:"start_time" binding:"required"`
	Hours        int    `json:"hours" binding:"required,min=1,max=168"`
	PackageID    *int   `json:"package_id"`
	ContactName  string `json:"contact_name" binding:"required"`
	ContactPhone string `json:"contact_phone" binding:"required"`
	Remarks      string `json:"remarks"`
}

// BookingListReq 预订列表请求
type BookingListReq struct {
	Page      int    `json:"page" form:"page" binding:"min=1"`
	PageSize  int    `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	Status    int    `json:"status" form:"status"`
	RoomID    int    `json:"room_id" form:"room_id"`
	UserID    int    `json:"user_id" form:"user_id"`
	StartDate string `json:"start_date" form:"start_date"`
	EndDate   string `json:"end_date" form:"end_date"`
	Keyword   string `json:"keyword" form:"keyword"`
}

// CheckAvailabilityReq 检查房间可用性请求
type CheckAvailabilityReq struct {
	RoomID    int    `json:"room_id" binding:"required"`
	StartTime string `json:"start_time" binding:"required"`
	Hours     int    `json:"hours" binding:"required,min=1,max=24"`
}

// UpdateBookingStatusReq 更新预订状态请求
type UpdateBookingStatusReq struct {
	ID     int `json:"id" binding:"required"`
	Status int `json:"status" binding:"required,oneof=1 2 3 4 5 6"`
}

// CancelBookingReq 取消预订请求
type CancelBookingReq struct {
	ID     int    `json:"id" binding:"required"`
	Reason string `json:"reason"`
}

// ManualStartBookingReq 手动开始订单请求
type ManualStartBookingReq struct {
	ID int `json:"id" binding:"required"`
}

// ManualEndBookingReq 手动结束订单请求
type ManualEndBookingReq struct {
	ID int `json:"id" binding:"required"`
}

// BookingStatusInfoReq 获取订单状态信息请求
type BookingStatusInfoReq struct {
	ID int `json:"id" form:"id" binding:"required"`
}

// ========== 房间使用记录相关请求 ==========

// CheckInReq 入住请求
type CheckInReq struct {
	BookingID int `json:"booking_id" binding:"required"`
}

// CheckOutReq 退房请求
type CheckOutReq struct {
	BookingID int     `json:"booking_id" binding:"required"`
	ExtraFee  float64 `json:"extra_fee" binding:"min=0"`
	Remarks   string  `json:"remarks"`
}

// UsageLogListReq 使用记录列表请求
type UsageLogListReq struct {
	Page      int    `json:"page" form:"page" binding:"min=1"`
	PageSize  int    `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	RoomID    int    `json:"room_id" form:"room_id"`
	UserID    int    `json:"user_id" form:"user_id"`
	StartDate string `json:"start_date" form:"start_date"`
	EndDate   string `json:"end_date" form:"end_date"`
}

// ========== 响应数据结构 ==========

// RoomListResp 房间列表响应
type RoomListResp struct {
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
	List     []RoomDetail `json:"list"`
}

// RoomDetail 房间详情
type RoomDetail struct {
	ID           int       `json:"id"`
	RoomNumber   string    `json:"room_number"`
	RoomName     string    `json:"room_name"`
	RoomType     string    `json:"room_type"`
	RoomTypeText string    `json:"room_type_text"`
	Capacity     int       `json:"capacity"`
	HourlyRate   float64   `json:"hourly_rate"`
	Features     []string  `json:"features"`
	Images       []string  `json:"images"`
	Status       int       `json:"status"`
	StatusText   string    `json:"status_text"`
	Floor        int       `json:"floor"`
	Area         float64   `json:"area"`
	Description  string    `json:"description"`
	CreateTime   time.Time `json:"create_time"`
	UpdateTime   time.Time `json:"update_time"`

	// 扩展信息
	IsAvailable    bool           `json:"is_available"`
	CurrentBooking *BookingDetail `json:"current_booking,omitempty"`
}

// BookingListResp 预订列表响应
type BookingListResp struct {
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	List     []BookingDetail `json:"list"`
}

// BookingDetail 预订详情
type BookingDetail struct {
	ID           int       `json:"id"`
	BookingNo    string    `json:"booking_no"`
	RoomID       int       `json:"room_id"`
	UserID       int       `json:"user_id"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Hours        int       `json:"hours"`
	TotalAmount  float64   `json:"total_amount"`
	PaidAmount   float64   `json:"paid_amount"`
	Status       int       `json:"status"`
	StatusText   string    `json:"status_text"`
	ContactName  string    `json:"contact_name"`
	ContactPhone string    `json:"contact_phone"`
	Remarks      string    `json:"remarks"`

	// 套餐相关信息
	PackageID      *int    `json:"package_id"`
	PackageName    string  `json:"package_name"`
	OriginalPrice  float64 `json:"original_price"`
	PackagePrice   float64 `json:"package_price"`
	DiscountAmount float64 `json:"discount_amount"`
	PriceBreakdown string  `json:"price_breakdown"`

	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`

	// 关联信息
	Room        *RoomDetail    `json:"room,omitempty"`
	UserInfo    *UserInfo      `json:"user_info,omitempty"`
	PackageInfo *PackageDetail `json:"package_info,omitempty"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// AvailabilityResp 可用性检查响应
type AvailabilityResp struct {
	IsAvailable    bool            `json:"is_available"`
	Message        string          `json:"message"`
	TotalAmount    float64         `json:"total_amount"`
	BasePrice      float64         `json:"base_price"`
	FinalPrice     float64         `json:"final_price"`
	PackageName    string          `json:"package_name,omitempty"`
	RuleName       string          `json:"rule_name,omitempty"`
	DayType        string          `json:"day_type"`
	DayTypeText    string          `json:"day_type_text"`
	PriceBreakdown *PriceBreakdown `json:"price_breakdown,omitempty"`
}

// UsageLogListResp 使用记录列表响应
type UsageLogListResp struct {
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	List     []UsageLogDetail `json:"list"`
}

// UsageLogDetail 使用记录详情
type UsageLogDetail struct {
	ID          int        `json:"id"`
	RoomID      int        `json:"room_id"`
	BookingID   int        `json:"booking_id"`
	UserID      int        `json:"user_id"`
	CheckInAt   time.Time  `json:"check_in_at"`
	CheckOutAt  *time.Time `json:"check_out_at"`
	ActualHours float64    `json:"actual_hours"`
	ExtraFee    float64    `json:"extra_fee"`
	CreateTime  time.Time  `json:"create_time"`

	// 关联信息
	Room     *RoomDetail    `json:"room,omitempty"`
	Booking  *BookingDetail `json:"booking,omitempty"`
	UserInfo *UserInfo      `json:"user_info,omitempty"`
}

// RoomStatistics 房间统计信息
type RoomStatistics struct {
	TotalRooms       int     `json:"total_rooms"`
	AvailableRooms   int     `json:"available_rooms"`
	OccupiedRooms    int     `json:"occupied_rooms"`
	MaintenanceRooms int     `json:"maintenance_rooms"`
	DisabledRooms    int     `json:"disabled_rooms"`
	OccupancyRate    float64 `json:"occupancy_rate"`
	TodayBookings    int     `json:"today_bookings"`
	TodayRevenue     float64 `json:"today_revenue"`
	MonthlyRevenue   float64 `json:"monthly_revenue"`
}

// ========== 套餐规则相关请求响应 ==========

// CreatePackageReq 创建套餐请求
type CreatePackageReq struct {
	RoomID      int     `json:"room_id" binding:"required"`
	PackageName string  `json:"package_name" binding:"required"`
	Description string  `json:"description"`
	PackageType string  `json:"package_type" binding:"required,oneof=flexible fixed_hours daily weekly"`
	FixedHours  int     `json:"fixed_hours" binding:"min=0,max=168"`
	MinHours    int     `json:"min_hours" binding:"min=1,max=24"`
	MaxHours    int     `json:"max_hours" binding:"min=1,max=168"`
	BasePrice   float64 `json:"base_price" binding:"min=0"`
	Priority    int     `json:"priority"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
}

// UpdatePackageReq 更新套餐请求
type UpdatePackageReq struct {
	ID          int     `json:"id" binding:"required"`
	RoomID      int     `json:"room_id" binding:"required"`
	PackageName string  `json:"package_name" binding:"required"`
	Description string  `json:"description"`
	PackageType string  `json:"package_type" binding:"required,oneof=flexible fixed_hours daily weekly"`
	FixedHours  int     `json:"fixed_hours" binding:"min=0,max=168"`
	MinHours    int     `json:"min_hours" binding:"min=1,max=24"`
	MaxHours    int     `json:"max_hours" binding:"min=1,max=168"`
	BasePrice   float64 `json:"base_price" binding:"min=0"`
	Priority    int     `json:"priority"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	IsActive    bool    `json:"is_active"`
}

// CreatePackageRuleReq 创建套餐规则请求
type CreatePackageRuleReq struct {
	PackageID  int     `json:"package_id" binding:"required"`
	RuleName   string  `json:"rule_name" binding:"required"`
	DayType    string  `json:"day_type" binding:"required,oneof=weekday weekend holiday special"`
	TimeStart  string  `json:"time_start"`
	TimeEnd    string  `json:"time_end"`
	PriceType  string  `json:"price_type" binding:"required,oneof=fixed multiply add"`
	PriceValue float64 `json:"price_value" binding:"required,min=0"`
	MinHours   int     `json:"min_hours" binding:"min=1,max=24"`
	MaxHours   int     `json:"max_hours" binding:"min=1,max=24"`
}

// UpdatePackageRuleReq 更新套餐规则请求
type UpdatePackageRuleReq struct {
	ID         int     `json:"id" binding:"required"`
	PackageID  int     `json:"package_id" binding:"required"`
	RuleName   string  `json:"rule_name" binding:"required"`
	DayType    string  `json:"day_type" binding:"required,oneof=weekday weekend holiday special"`
	TimeStart  string  `json:"time_start"`
	TimeEnd    string  `json:"time_end"`
	PriceType  string  `json:"price_type" binding:"required,oneof=fixed multiply add"`
	PriceValue float64 `json:"price_value" binding:"required,min=0"`
	MinHours   int     `json:"min_hours" binding:"min=1,max=24"`
	MaxHours   int     `json:"max_hours" binding:"min=1,max=24"`
	IsActive   bool    `json:"is_active"`
}

// PackageListReq 套餐列表请求
type PackageListReq struct {
	Page     int `json:"page" form:"page" binding:"min=1"`
	PageSize int `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	RoomID   int `json:"room_id" form:"room_id"`
}

// PriceCalculateResp 价格计算响应（增强版）
type PriceCalculateResp struct {
	IsAvailable    bool            `json:"is_available"`
	Message        string          `json:"message"`
	BasePrice      float64         `json:"base_price"`
	FinalPrice     float64         `json:"final_price"`
	PackageName    string          `json:"package_name,omitempty"`
	RuleName       string          `json:"rule_name,omitempty"`
	DayType        string          `json:"day_type"`
	DayTypeText    string          `json:"day_type_text"`
	PriceBreakdown *PriceBreakdown `json:"price_breakdown,omitempty"`
}

// PriceBreakdown 价格明细
type PriceBreakdown struct {
	BaseHourlyRate float64 `json:"base_hourly_rate"`
	Hours          int     `json:"hours"`
	SubTotal       float64 `json:"sub_total"`
	RuleType       string  `json:"rule_type"`
	RuleValue      float64 `json:"rule_value"`
	Adjustment     float64 `json:"adjustment"`
	FinalTotal     float64 `json:"final_total"`
}

// PackageDetail 套餐详情
type PackageDetail struct {
	ID              int                 `json:"id"`
	RoomID          int                 `json:"room_id"`
	PackageName     string              `json:"package_name"`
	Description     string              `json:"description"`
	PackageType     string              `json:"package_type"`
	PackageTypeText string              `json:"package_type_text"`
	FixedHours      int                 `json:"fixed_hours"`
	MinHours        int                 `json:"min_hours"`
	MaxHours        int                 `json:"max_hours"`
	BasePrice       float64             `json:"base_price"`
	IsActive        bool                `json:"is_active"`
	Priority        int                 `json:"priority"`
	StartDate       *time.Time          `json:"start_date"`
	EndDate         *time.Time          `json:"end_date"`
	CreateTime      time.Time           `json:"create_time"`
	UpdateTime      time.Time           `json:"update_time"`
	Rules           []PackageRuleDetail `json:"rules"`
	Room            *RoomDetail         `json:"room,omitempty"`
}

// PackageRuleDetail 套餐规则详情
type PackageRuleDetail struct {
	ID            int       `json:"id"`
	PackageID     int       `json:"package_id"`
	RuleName      string    `json:"rule_name"`
	DayType       string    `json:"day_type"`
	DayTypeText   string    `json:"day_type_text"`
	TimeStart     string    `json:"time_start"`
	TimeEnd       string    `json:"time_end"`
	PriceType     string    `json:"price_type"`
	PriceTypeText string    `json:"price_type_text"`
	PriceValue    float64   `json:"price_value"`
	MinHours      int       `json:"min_hours"`
	MaxHours      int       `json:"max_hours"`
	IsActive      bool      `json:"is_active"`
	CreateTime    time.Time `json:"create_time"`
	UpdateTime    time.Time `json:"update_time"`
}

// PackageListResp 套餐列表响应
type PackageListResp struct {
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	List     []PackageDetail `json:"list"`
}

// SpecialDateDetail 特殊日期详情
type SpecialDateDetail struct {
	ID           int       `json:"id"`
	Date         time.Time `json:"date"`
	DateType     string    `json:"date_type"`
	DateTypeText string    `json:"date_type_text"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	IsActive     bool      `json:"is_active"`
	CreateTime   time.Time `json:"create_time"`
	UpdateTime   time.Time `json:"update_time"`
}

// SpecialDateListResp 特殊日期列表响应
type SpecialDateListResp struct {
	Total    int                  `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
	List     []*SpecialDateDetail `json:"list"`
}

// GetRoomPackagesReq 获取房间套餐请求
type GetRoomPackagesReq struct {
	RoomID    int    `json:"room_id" form:"room_id" binding:"required"`
	StartTime string `json:"start_time" form:"start_time" binding:"required"`
	Hours     int    `json:"hours" form:"hours" binding:"required,min=1,max=168"`
}

// RoomPackageOption 房间套餐选项
type RoomPackageOption struct {
	PackageID       int     `json:"package_id"`
	PackageName     string  `json:"package_name"`
	Description     string  `json:"description"`
	PackageType     string  `json:"package_type"`
	PackageTypeText string  `json:"package_type_text"`
	FixedHours      int     `json:"fixed_hours"`
	MinHours        int     `json:"min_hours"`
	MaxHours        int     `json:"max_hours"`
	BasePrice       float64 `json:"base_price"`
	FinalPrice      float64 `json:"final_price"`
	OriginalPrice   float64 `json:"original_price"`
	DiscountAmount  float64 `json:"discount_amount"`
	DiscountPercent float64 `json:"discount_percent"`
	RuleName        string  `json:"rule_name"`
	DayType         string  `json:"day_type"`
	DayTypeText     string  `json:"day_type_text"`
	IsRecommended   bool    `json:"is_recommended"`
	IsAvailable     bool    `json:"is_available"`
	UnavailableMsg  string  `json:"unavailable_msg,omitempty"`
}

// GetRoomPackagesResp 获取房间套餐响应
type GetRoomPackagesResp struct {
	RoomID      int                  `json:"room_id"`
	StartTime   string               `json:"start_time"`
	Hours       int                  `json:"hours"`
	DayType     string               `json:"day_type"`
	DayTypeText string               `json:"day_type_text"`
	Packages    []*RoomPackageOption `json:"packages"`
}

// BookingPricePreviewReq 预订价格预览请求
type BookingPricePreviewReq struct {
	RoomID    int    `json:"room_id" binding:"required"`
	StartTime string `json:"start_time" binding:"required"`
	Hours     int    `json:"hours" binding:"required,min=1,max=168"`
	PackageID *int   `json:"package_id"`
}

// BookingPricePreviewResp 预订价格预览响应
type BookingPricePreviewResp struct {
	RoomID          int             `json:"room_id"`
	Hours           int             `json:"hours"`
	BasePrice       float64         `json:"base_price"`
	PackageID       *int            `json:"package_id"`
	PackageName     string          `json:"package_name,omitempty"`
	OriginalPrice   float64         `json:"original_price"`
	FinalPrice      float64         `json:"final_price"`
	DiscountAmount  float64         `json:"discount_amount"`
	DiscountPercent float64         `json:"discount_percent"`
	RuleName        string          `json:"rule_name,omitempty"`
	DayType         string          `json:"day_type"`
	DayTypeText     string          `json:"day_type_text"`
	PriceBreakdown  *PriceBreakdown `json:"price_breakdown,omitempty"`
}

// BookingStatusInfoResp 订单状态信息响应
type BookingStatusInfoResp struct {
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
