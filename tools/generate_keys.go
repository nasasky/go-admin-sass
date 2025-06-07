package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
)

func main() {
	fmt.Println("=== NASA-Go-Admin 安全密钥生成工具 ===")
	fmt.Println()

	// 生成 JWT 签名密钥 (256 bits)
	jwtKey, err := generateSecureKey(32)
	if err != nil {
		log.Fatal("生成JWT密钥失败:", err)
	}

	// 生成会话密钥 (256 bits)
	sessionKey, err := generateSecureKey(32)
	if err != nil {
		log.Fatal("生成会话密钥失败:", err)
	}

	// 生成加密密钥 (256 bits)
	encryptionKey, err := generateSecureKey(32)
	if err != nil {
		log.Fatal("生成加密密钥失败:", err)
	}

	fmt.Println("请将以下密钥添加到您的 .env 文件中：")
	fmt.Println()
	fmt.Printf("JWT_SIGNING_KEY=%s\n", jwtKey)
	fmt.Printf("SESSION_SECRET=%s\n", sessionKey)
	fmt.Printf("ENCRYPTION_KEY=%s\n", encryptionKey)
	fmt.Println()
	fmt.Println("⚠️  重要提醒：")
	fmt.Println("1. 请妥善保管这些密钥，不要泄露给他人")
	fmt.Println("2. 生产环境中请使用不同的密钥")
	fmt.Println("3. 定期更换密钥以提高安全性")
	fmt.Println("4. 密钥长度至少32个字符")
	fmt.Println()
}

// generateSecureKey 生成指定长度的安全密钥
func generateSecureKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
