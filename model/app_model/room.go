package app_model

import (
	"time"
)

// Room 房间包厢模型
type Room struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	RoomNumber  string    `json:"room_number" gorm:"column:room_number;uniqueIndex;not null;comment:房间号码"`
	RoomName    string    `json:"room_name" gorm:"column:room_name;not null;comment:房间名称"`
	RoomType    string    `json:"room_type" gorm:"column:room_type;not null;comment:房间类型(小包厢/中包厢/大包厢/豪华包厢)"`
	Capacity    int       `json:"capacity" gorm:"column:capacity;not null;comment:容纳人数"`
	HourlyRate  float64   `json:"hourly_rate" gorm:"column:hourly_rate;type:decimal(10,2);not null;comment:每小时价格"`
	Features    string    `json:"features" gorm:"column:features;type:text;comment:房间特色设施(JSON格式)"`
	Images      string    `json:"images" gorm:"column:images;type:text;comment:房间图片URLs(JSON格式)"`
	Status      int       `json:"status" gorm:"column:status;default:1;comment:房间状态(1:可用,2:使用中,3:维护中,4:停用)"`
	Floor       int       `json:"floor" gorm:"column:floor;comment:楼层"`
	Area        float64   `json:"area" gorm:"column:area;type:decimal(8,2);comment:房间面积(平方米)"`
	Description string    `json:"description" gorm:"column:description;type:text;comment:房间描述"`
	CreatedBy   int       `json:"created_by" gorm:"column:created_by;comment:创建人ID"`
	CreateTime  time.Time `json:"create_time" gorm:"column:create_time;autoCreateTime"`
	UpdateTime  time.Time `json:"update_time" gorm:"column:update_time;autoUpdateTime"`
}

// RoomBooking 房间预订模型
type RoomBooking struct {
	ID           int       `json:"id" gorm:"primaryKey;autoIncrement"`
	RoomID       int       `json:"room_id" gorm:"column:room_id;not null;comment:房间ID"`
	UserID       int       `json:"user_id" gorm:"column:user_id;not null;comment:用户ID"`
	BookingNo    string    `json:"booking_no" gorm:"column:booking_no;uniqueIndex;not null;comment:预订单号"`
	StartTime    time.Time `json:"start_time" gorm:"column:start_time;not null;comment:开始时间"`
	EndTime      time.Time `json:"end_time" gorm:"column:end_time;not null;comment:结束时间"`
	Hours        int       `json:"hours" gorm:"column:hours;not null;comment:预订小时数"`
	TotalAmount  float64   `json:"total_amount" gorm:"column:total_amount;type:decimal(10,2);not null;comment:总金额"`
	PaidAmount   float64   `json:"paid_amount" gorm:"column:paid_amount;type:decimal(10,2);default:0;comment:已支付金额"`
	Status       int       `json:"status" gorm:"column:status;default:1;comment:预订状态(1:待支付,2:已支付,3:使用中,4:已完成,5:已取消,6:已退款)"`
	PaymentID    *int      `json:"payment_id" gorm:"column:payment_id;comment:支付记录ID"`
	ContactName  string    `json:"contact_name" gorm:"column:contact_name;comment:联系人姓名"`
	ContactPhone string    `json:"contact_phone" gorm:"column:contact_phone;comment:联系人电话"`
	Remarks      string    `json:"remarks" gorm:"column:remarks;type:text;comment:备注信息"`

	// 套餐相关字段
	PackageID      *int    `json:"package_id" gorm:"column:package_id;comment:使用的套餐ID"`
	PackageName    string  `json:"package_name" gorm:"column:package_name;comment:套餐名称"`
	OriginalPrice  float64 `json:"original_price" gorm:"column:original_price;type:decimal(10,2);default:0;comment:原始价格"`
	PackagePrice   float64 `json:"package_price" gorm:"column:package_price;type:decimal(10,2);default:0;comment:套餐价格"`
	DiscountAmount float64 `json:"discount_amount" gorm:"column:discount_amount;type:decimal(10,2);default:0;comment:优惠金额"`
	PriceBreakdown string  `json:"price_breakdown" gorm:"column:price_breakdown;type:text;comment:价格明细JSON"`

	CreateTime time.Time `json:"create_time" gorm:"column:create_time;autoCreateTime"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time;autoUpdateTime"`

	// 关联查询字段
	Room    *Room        `json:"room,omitempty" gorm:"foreignKey:RoomID"`
	User    *UserApp     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Package *RoomPackage `json:"package,omitempty" gorm:"foreignKey:PackageID"`
}

// RoomUsageLog 房间使用记录
type RoomUsageLog struct {
	ID          int        `json:"id" gorm:"primaryKey;autoIncrement"`
	RoomID      int        `json:"room_id" gorm:"column:room_id;not null;comment:房间ID"`
	BookingID   int        `json:"booking_id" gorm:"column:booking_id;not null;comment:预订ID"`
	UserID      int        `json:"user_id" gorm:"column:user_id;not null;comment:用户ID"`
	CheckInAt   time.Time  `json:"check_in_at" gorm:"column:check_in_at;comment:入住时间"`
	CheckOutAt  *time.Time `json:"check_out_at" gorm:"column:check_out_at;comment:离开时间"`
	ActualHours float64    `json:"actual_hours" gorm:"column:actual_hours;type:decimal(8,2);comment:实际使用小时数"`
	ExtraFee    float64    `json:"extra_fee" gorm:"column:extra_fee;type:decimal(10,2);default:0;comment:额外费用"`
	CreateTime  time.Time  `json:"create_time" gorm:"column:create_time;autoCreateTime"`
	UpdateTime  time.Time  `json:"update_time" gorm:"column:update_time;autoUpdateTime"`
}

// 定义表名
func (Room) TableName() string {
	return "rooms"
}

func (RoomBooking) TableName() string {
	return "room_bookings"
}

func (RoomUsageLog) TableName() string {
	return "room_usage_logs"
}

// 房间状态常量
const (
	RoomStatusAvailable   = 1 // 可用
	RoomStatusOccupied    = 2 // 使用中
	RoomStatusMaintenance = 3 // 维护中
	RoomStatusDisabled    = 4 // 停用
)

// 预订状态常量
const (
	BookingStatusPending   = 1 // 待支付
	BookingStatusPaid      = 2 // 已支付
	BookingStatusInUse     = 3 // 使用中
	BookingStatusCompleted = 4 // 已完成
	BookingStatusCancelled = 5 // 已取消
	BookingStatusRefunded  = 6 // 已退款
)

// 房间类型常量
const (
	RoomTypeSmall  = "small"  // 小包厢
	RoomTypeMedium = "medium" // 中包厢
	RoomTypeLarge  = "large"  // 大包厢
	RoomTypeLuxury = "luxury" // 豪华包厢
)

// GetStatusText 获取状态文本
func (r *Room) GetStatusText() string {
	switch r.Status {
	case RoomStatusAvailable:
		return "可用"
	case RoomStatusOccupied:
		return "使用中"
	case RoomStatusMaintenance:
		return "维护中"
	case RoomStatusDisabled:
		return "停用"
	default:
		return "未知"
	}
}

// GetBookingStatusText 获取预订状态文本
func (rb *RoomBooking) GetBookingStatusText() string {
	switch rb.Status {
	case BookingStatusPending:
		return "待支付"
	case BookingStatusPaid:
		return "已支付"
	case BookingStatusInUse:
		return "使用中"
	case BookingStatusCompleted:
		return "已完成"
	case BookingStatusCancelled:
		return "已取消"
	case BookingStatusRefunded:
		return "已退款"
	default:
		return "未知"
	}
}

// IsAvailable 检查房间是否在指定时间段可用
func (r *Room) IsAvailable() bool {
	return r.Status == RoomStatusAvailable
}
