package utils

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// MongoTime 自定义MongoDB时间类型，确保统一的时间格式
type MongoTime time.Time

// MarshalBSON 实现bson.Marshaler接口，控制MongoDB存储格式
func (t MongoTime) MarshalBSON() ([]byte, error) {
	// 以标准time.Time格式存储，但会被MongoDB转为ISO 8601
	return bson.Marshal(time.Time(t))
}

// UnmarshalBSON 实现bson.Unmarshaler接口
func (t *MongoTime) UnmarshalBSON(data []byte) error {
	var tm time.Time
	if err := bson.Unmarshal(data, &tm); err != nil {
		return err
	}
	*t = MongoTime(tm)
	return nil
}

// MarshalJSON 实现json.Marshaler接口，控制JSON输出格式
func (t MongoTime) MarshalJSON() ([]byte, error) {
	// 输出为自定义格式而不是ISO 8601
	stamp := time.Time(t).Format(DateTimeFormat)
	return []byte(fmt.Sprintf("\"%s\"", stamp)), nil
}

// String 返回格式化的时间字符串
func (t MongoTime) String() string {
	return time.Time(t).Format(DateTimeFormat)
}

// 时区处理相关函数

// ConvertToLocalTime 将UTC时间转换为本地时间
func ConvertToLocalTime(utcTime time.Time) time.Time {
	// 假设服务器时区为 Asia/Shanghai (UTC+8)
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// 如果加载失败，使用本地时区
		return utcTime.Local()
	}
	return utcTime.In(loc)
}

// ConvertToUTC 将本地时间转换为UTC时间
func ConvertToUTC(localTime time.Time) time.Time {
	return localTime.UTC()
}

// FormatMongoTime 格式化MongoDB中的时间字段
func FormatMongoTime(mongoTimeStr string) string {
	// 尝试解析多种时间格式
	formats := []string{
		time.RFC3339Nano,           // 2006-01-02T15:04:05.999999999Z07:00
		time.RFC3339,               // 2006-01-02T15:04:05Z07:00
		"2006-01-02T15:04:05.000Z", // MongoDB 常见格式
		"2006-01-02T15:04:05Z",     // 简化版 ISO 8601
		DateTimeFormat,             // 2006-01-02 15:04:05
	}

	for _, format := range formats {
		if t, err := time.Parse(format, mongoTimeStr); err == nil {
			// 转换为本地时间并格式化
			localTime := ConvertToLocalTime(t)
			return localTime.Format(DateTimeFormat)
		}
	}

	// 如果都解析失败，返回原字符串
	return mongoTimeStr
}

// GetCurrentTimeForMongo 获取适合MongoDB存储的当前时间字符串
func GetCurrentTimeForMongo() string {
	// 返回本地时间的字符串格式，直接存储为字符串避免ISO 8601格式
	return time.Now().Format(DateTimeFormat)
}

// GetCurrentUTCTimeForMongo 获取UTC时间字符串用于MongoDB存储
func GetCurrentUTCTimeForMongo() string {
	// 返回UTC时间的字符串格式
	return time.Now().UTC().Format(DateTimeFormat)
}

// ParseTimeFromMongo 从MongoDB读取的时间字符串解析为time.Time
func ParseTimeFromMongo(mongoTime interface{}) (time.Time, error) {
	switch v := mongoTime.(type) {
	case time.Time:
		return v, nil
	case string:
		return time.Parse(time.RFC3339, v)
	case primitive.DateTime:
		return v.Time(), nil
	default:
		return time.Time{}, fmt.Errorf("无法解析的时间类型: %T", v)
	}
}

// FormatTimeFieldsForResponse 格式化响应中的时间字段
func FormatTimeFieldsForResponse(data interface{}) interface{} {
	val := reflect.ValueOf(data)
	switch val.Kind() {
	case reflect.Ptr:
		if !val.IsNil() {
			elem := val.Elem()
			newVal := FormatTimeFieldsForResponse(elem.Interface())
			result := reflect.New(elem.Type())
			result.Elem().Set(reflect.ValueOf(newVal))
			return result.Interface()
		}
		return data
	case reflect.Struct:
		newStruct := reflect.New(val.Type()).Elem()
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			fieldType := val.Type().Field(i)

			if fieldType.Type == reflect.TypeOf(time.Time{}) {
				// 转换时间字段
				timeVal := field.Interface().(time.Time)
				localTime := ConvertToLocalTime(timeVal)
				newStruct.Field(i).Set(reflect.ValueOf(localTime.Format(DateTimeFormat)))
			} else if fieldType.Type == reflect.TypeOf(MongoTime{}) {
				// 处理自定义MongoTime类型
				mongoTime := field.Interface().(MongoTime)
				newStruct.Field(i).Set(reflect.ValueOf(mongoTime.String()))
			} else {
				newStruct.Field(i).Set(reflect.ValueOf(FormatTimeFieldsForResponse(field.Interface())))
			}
		}
		return newStruct.Interface()
	case reflect.Slice:
		newSlice := reflect.MakeSlice(val.Type(), val.Len(), val.Cap())
		for i := 0; i < val.Len(); i++ {
			newSlice.Index(i).Set(reflect.ValueOf(FormatTimeFieldsForResponse(val.Index(i).Interface())))
		}
		return newSlice.Interface()
	case reflect.Map:
		newMap := reflect.MakeMapWithSize(val.Type(), val.Len())
		for _, key := range val.MapKeys() {
			mapVal := val.MapIndex(key)

			// 特殊处理时间字段
			if key.Kind() == reflect.String {
				keyStr := key.String()
				if strings.Contains(strings.ToLower(keyStr), "time") ||
					strings.Contains(strings.ToLower(keyStr), "timestamp") {

					if timeStr, ok := mapVal.Interface().(string); ok {
						formatted := FormatMongoTime(timeStr)
						newMap.SetMapIndex(key, reflect.ValueOf(formatted))
						continue
					}
				}
			}

			newMap.SetMapIndex(key, reflect.ValueOf(FormatTimeFieldsForResponse(mapVal.Interface())))
		}
		return newMap.Interface()
	default:
		return data
	}
}

// TimeZoneInfo 时区信息
type TimeZoneInfo struct {
	Name   string `json:"name"`
	Offset string `json:"offset"`
	UTC    string `json:"utc"`
	Local  string `json:"local"`
}

// GetTimeZoneInfo 获取当前时区信息
func GetTimeZoneInfo() TimeZoneInfo {
	now := time.Now()
	utcTime := now.UTC()

	_, offset := now.Zone()
	offsetHours := offset / 3600
	offsetSign := "+"
	if offsetHours < 0 {
		offsetSign = "-"
		offsetHours = -offsetHours
	}

	return TimeZoneInfo{
		Name:   now.Location().String(),
		Offset: fmt.Sprintf("UTC%s%d", offsetSign, offsetHours),
		UTC:    utcTime.Format(DateTimeFormat),
		Local:  now.Format(DateTimeFormat),
	}
}
