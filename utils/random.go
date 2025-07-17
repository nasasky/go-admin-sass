package utils

import (
	cryptoRand "crypto/rand"
	"fmt"
	mathRand "math/rand"
	"time"
)

// RandomString 生成指定长度的随机字符串
func RandomString(length int) string {
	mathRand.Seed(time.Now().UnixNano())
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[mathRand.Intn(len(letters))]
	}
	return string(b)
}

// GenerateUUID 生成一个随机的UUID
func GenerateUUID() string {
	uuid := make([]byte, 16)
	cryptoRand.Read(uuid)
	// 设置版本 (4) 和变体位
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
