package app_model

import (
	"fmt"
	"time"
)

// RoomPackage 房间套餐规则
type RoomPackage struct {
	ID          int        `json:"id" gorm:"primaryKey;autoIncrement"`
	RoomID      int        `json:"room_id" gorm:"column:room_id;not null;comment:房间ID"`
	PackageName string     `json:"package_name" gorm:"column:package_name;not null;comment:套餐名称"`
	Description string     `json:"description" gorm:"column:description;type:text;comment:套餐描述"`
	PackageType string     `json:"package_type" gorm:"column:package_type;default:flexible;comment:套餐类型(flexible/fixed_hours/daily/weekly)"`
	FixedHours  int        `json:"fixed_hours" gorm:"column:fixed_hours;default:0;comment:固定时长(小时)，0表示灵活时长"`
	MinHours    int        `json:"min_hours" gorm:"column:min_hours;default:1;comment:最少预订小时数"`
	MaxHours    int        `json:"max_hours" gorm:"column:max_hours;default:24;comment:最多预订小时数"`
	BasePrice   float64    `json:"base_price" gorm:"column:base_price;type:decimal(10,2);default:0;comment:套餐基础价格"`
	IsActive    bool       `json:"is_active" gorm:"column:is_active;default:true;comment:是否启用"`
	Priority    int        `json:"priority" gorm:"column:priority;default:0;comment:优先级，数字越大优先级越高"`
	StartDate   *time.Time `json:"start_date" gorm:"column:start_date;comment:生效开始日期"`
	EndDate     *time.Time `json:"end_date" gorm:"column:end_date;comment:生效结束日期"`
	CreateTime  time.Time  `json:"create_time" gorm:"column:create_time;autoCreateTime"`
	UpdateTime  time.Time  `json:"update_time" gorm:"column:update_time;autoUpdateTime"`

	// 关联查询
	Room  *Room             `json:"room,omitempty" gorm:"foreignKey:RoomID"`
	Rules []RoomPackageRule `json:"rules,omitempty" gorm:"foreignKey:PackageID"`
}

// RoomPackageRule 套餐定价规则
type RoomPackageRule struct {
	ID         int       `json:"id" gorm:"primaryKey;autoIncrement"`
	PackageID  int       `json:"package_id" gorm:"column:package_id;not null;comment:套餐ID"`
	RuleName   string    `json:"rule_name" gorm:"column:rule_name;not null;comment:规则名称"`
	DayType    string    `json:"day_type" gorm:"column:day_type;not null;comment:日期类型(weekday/weekend/holiday/special)"`
	TimeStart  string    `json:"time_start" gorm:"column:time_start;comment:时间段开始(HH:mm)"`
	TimeEnd    string    `json:"time_end" gorm:"column:time_end;comment:时间段结束(HH:mm)"`
	PriceType  string    `json:"price_type" gorm:"column:price_type;not null;comment:价格类型(fixed/multiply/add)"`
	PriceValue float64   `json:"price_value" gorm:"column:price_value;type:decimal(10,2);not null;comment:价格值"`
	MinHours   int       `json:"min_hours" gorm:"column:min_hours;default:1;comment:最少预订小时数"`
	MaxHours   int       `json:"max_hours" gorm:"column:max_hours;default:24;comment:最多预订小时数"`
	IsActive   bool      `json:"is_active" gorm:"column:is_active;default:true;comment:是否启用"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time;autoCreateTime"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time;autoUpdateTime"`

	// 关联查询
	Package *RoomPackage `json:"package,omitempty" gorm:"foreignKey:PackageID"`
}

// RoomSpecialDate 特殊日期配置（节假日、特殊活动日等）
type RoomSpecialDate struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Date        time.Time `json:"date" gorm:"column:date;type:date;not null;comment:特殊日期"`
	DateType    string    `json:"date_type" gorm:"column:date_type;not null;comment:日期类型(holiday/festival/special)"`
	Name        string    `json:"name" gorm:"column:name;not null;comment:日期名称"`
	Description string    `json:"description" gorm:"column:description;type:text;comment:描述"`
	IsActive    bool      `json:"is_active" gorm:"column:is_active;default:true;comment:是否启用"`
	CreateTime  time.Time `json:"create_time" gorm:"column:create_time;autoCreateTime"`
	UpdateTime  time.Time `json:"update_time" gorm:"column:update_time;autoUpdateTime"`
}

// 定义表名
func (RoomPackage) TableName() string {
	return "room_packages"
}

func (RoomPackageRule) TableName() string {
	return "room_package_rules"
}

func (RoomSpecialDate) TableName() string {
	return "room_special_dates"
}

// 日期类型常量
const (
	DayTypeWeekday = "weekday" // 工作日
	DayTypeWeekend = "weekend" // 周末
	DayTypeHoliday = "holiday" // 节假日
	DayTypeSpecial = "special" // 特殊日期
)

// 价格类型常量
const (
	PriceTypeFixed    = "fixed"    // 固定价格
	PriceTypeMultiply = "multiply" // 倍数（基础价格*倍数）
	PriceTypeAdd      = "add"      // 加价（基础价格+固定金额）
)

// 特殊日期类型常量
const (
	SpecialDateHoliday  = "holiday"  // 法定节假日
	SpecialDateFestival = "festival" // 传统节日
	SpecialDateSpecial  = "special"  // 特殊活动日
)

// 套餐类型常量
const (
	PackageTypeFlexible   = "flexible"    // 灵活时长套餐
	PackageTypeFixedHours = "fixed_hours" // 固定时长套餐(如3小时套餐)
	PackageTypeDaily      = "daily"       // 全天套餐(24小时)
	PackageTypeWeekly     = "weekly"      // 周套餐(7天)
)

// GetDayType 根据日期获取日期类型
func GetDayType(date time.Time) string {
	// 检查是否为特殊日期（这里可以查询数据库或配置）
	if IsSpecialDate(date) {
		return DayTypeSpecial
	}

	// 检查是否为周末
	weekday := date.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return DayTypeWeekend
	}

	// 默认为工作日
	return DayTypeWeekday
}

// IsSpecialDate 检查是否为特殊日期（这里是示例，实际应该查询数据库）
func IsSpecialDate(date time.Time) bool {
	// 这里可以查询 room_special_dates 表
	// 为了示例，我们硬编码一些特殊日期
	dateStr := date.Format("01-02")
	specialDates := []string{
		"01-01", // 元旦
		"02-14", // 情人节
		"05-01", // 劳动节
		"06-01", // 儿童节
		"10-01", // 国庆节
		"12-25", // 圣诞节
	}

	for _, special := range specialDates {
		if dateStr == special {
			return true
		}
	}
	return false
}

// CalculatePrice 根据套餐规则计算价格
func (rp *RoomPackage) CalculatePrice(basePrice float64, startTime time.Time, requestedHours int) (float64, *RoomPackageRule, error) {
	if !rp.IsActive {
		return basePrice, nil, nil
	}

	// 检查套餐生效时间
	if rp.StartDate != nil && startTime.Before(*rp.StartDate) {
		return basePrice, nil, nil
	}
	if rp.EndDate != nil && startTime.After(*rp.EndDate) {
		return basePrice, nil, nil
	}

	// 检查套餐类型和时长限制
	actualHours := requestedHours
	switch rp.PackageType {
	case PackageTypeFixedHours:
		// 固定时长套餐：必须使用固定小时数
		if rp.FixedHours > 0 {
			actualHours = rp.FixedHours
		}
	case PackageTypeDaily:
		// 全天套餐：固定24小时
		actualHours = 24
	case PackageTypeWeekly:
		// 周套餐：固定168小时(7天)
		actualHours = 168
	case PackageTypeFlexible:
		// 灵活套餐：检查最小最大时长限制
		if requestedHours < rp.MinHours || requestedHours > rp.MaxHours {
			return basePrice, nil, fmt.Errorf("预订时长必须在 %d-%d 小时之间", rp.MinHours, rp.MaxHours)
		}
	}

	// 获取日期类型
	dayType := GetDayType(startTime)
	timeStr := startTime.Format("15:04")

	// 找到匹配的规则，按优先级排序
	var matchedRule *RoomPackageRule
	for _, rule := range rp.Rules {
		if !rule.IsActive {
			continue
		}

		// 检查日期类型匹配
		if rule.DayType != dayType {
			continue
		}

		// 检查时间段匹配
		if rule.TimeStart != "" && rule.TimeEnd != "" {
			if timeStr < rule.TimeStart || timeStr >= rule.TimeEnd {
				continue
			}
		}

		// 检查小时数限制
		if actualHours < rule.MinHours || actualHours > rule.MaxHours {
			continue
		}

		// 找到匹配的规则
		matchedRule = &rule
		break
	}

	// 使用套餐基础价格或房间基础价格
	priceBase := basePrice
	if rp.BasePrice > 0 {
		priceBase = rp.BasePrice
	}

	// 计算最终价格
	var finalPrice float64
	if matchedRule != nil {
		switch matchedRule.PriceType {
		case PriceTypeFixed:
			finalPrice = matchedRule.PriceValue
		case PriceTypeMultiply:
			finalPrice = priceBase * matchedRule.PriceValue
		case PriceTypeAdd:
			finalPrice = priceBase + matchedRule.PriceValue
		default:
			finalPrice = priceBase
		}
	} else {
		// 没有匹配的规则，使用基础价格
		finalPrice = priceBase
	}

	// 对于固定时长套餐，价格通常是总价，而不是按小时计算
	if rp.PackageType == PackageTypeFixedHours || rp.PackageType == PackageTypeDaily || rp.PackageType == PackageTypeWeekly {
		// 固定套餐价格就是最终价格
		return finalPrice, matchedRule, nil
	} else {
		// 灵活套餐按小时计算
		return finalPrice * float64(actualHours), matchedRule, nil
	}
}

// GetActivePackage 获取房间的有效套餐（按优先级排序）
func GetActivePackageForRoom(roomID int, date time.Time) (*RoomPackage, error) {
	// 这里应该查询数据库，按优先级降序获取有效套餐
	// 为了示例，我们返回nil，实际实现时需要查询数据库
	return nil, nil
}
