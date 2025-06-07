package utils

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// 常用时间格式常量
const (
	DateFormat           = "2006-01-02"
	DateTimeFormat       = "2006-01-02 15:04:05"
	DateTimeMinuteFormat = "2006-01-02 15:04"
	TimeFormat           = "15:04:05"
	ISO8601Format        = "2006-01-02T15:04:05-07:00" // ISO 8601 格式
)

// 格式化时间为字符串
func FormatTime2(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	// 直接使用定义好的 DateTimeFormat 常量格式化时间
	return t.Local().Format(DateTimeFormat)
}

// 格式化时间为日期字符串
func FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(DateFormat)
}

// 解析 ISO 8601 格式时间
func ParseISO8601Time(input string) (time.Time, error) {
	return time.Parse(ISO8601Format, input)
}

// 格式化 ISO 8601 时间字符串为指定格式
func FormatISO8601ToCustom(input string) (string, error) {
	parsedTime, err := ParseISO8601Time(input)
	if err != nil {
		return "", err
	}
	return FormatTime2(parsedTime), nil
}

// JSONTime 自定义时间类型，用于在JSON序列化时格式化时间
type JSONTime time.Time

// MarshalJSON 实现json.Marshaler接口
func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := time.Time(t).Format(DateTimeFormat)
	return []byte(fmt.Sprintf("\"%s\"", stamp)), nil
}

// Value 实现driver.Valuer接口，用于数据库操作
func (t JSONTime) Value() (driver.Value, error) {
	if time.Time(t).IsZero() {
		return nil, nil
	}
	return time.Time(t), nil
}

// Scan 实现sql.Scanner接口，用于数据库操作
func (t *JSONTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = JSONTime(value)
		return nil
	}
	return fmt.Errorf("无法将 %v 转换为时间", v)
}
