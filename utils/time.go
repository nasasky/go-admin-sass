package utils

import (
	"reflect"
	"time"
)

// FormatTimeFields 递归地格式化结构体中的 time.Time 字段
func FormatTimeFields(data interface{}) interface{} {
	val := reflect.ValueOf(data)
	switch val.Kind() {
	case reflect.Ptr:
		if !val.IsNil() {
			val.Elem().Set(reflect.ValueOf(FormatTimeFields(val.Elem().Interface())))
		}
	case reflect.Struct:
		newStruct := reflect.New(val.Type()).Elem()
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			fieldType := val.Type().Field(i)
			if fieldType.Type == reflect.TypeOf(time.Time{}) {
				// 格式化 time.Time 字段
				formattedTime := field.Interface().(time.Time).Format("2006-01-02 15:04:05")
				newStruct.Field(i).Set(reflect.ValueOf(formattedTime))
			} else {
				newStruct.Field(i).Set(reflect.ValueOf(FormatTimeFields(field.Interface())))
			}
		}
		return newStruct.Interface()
	case reflect.Slice:
		newSlice := reflect.MakeSlice(val.Type(), val.Len(), val.Cap())
		for i := 0; i < val.Len(); i++ {
			newSlice.Index(i).Set(reflect.ValueOf(FormatTimeFields(val.Index(i).Interface())))
		}
		return newSlice.Interface()
	case reflect.Map:
		newMap := reflect.MakeMap(val.Type())
		for _, key := range val.MapKeys() {
			newMap.SetMapIndex(key, reflect.ValueOf(FormatTimeFields(val.MapIndex(key).Interface())))
		}
		return newMap.Interface()
	default:
		return data
	}
	return data
}

// FormatTime formats the given time string to "2006-01-02 15:04:05"
func FormatTime(timeStr string) string {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
