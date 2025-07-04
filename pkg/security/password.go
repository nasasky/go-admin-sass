package security

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

const (
	// bcrypt 推荐的最小成本
	MinCost     = 12
	DefaultCost = 14
)

// HashPassword 使用 bcrypt 哈希密码
func HashPassword(password string) (string, error) {
	if err := ValidatePasswordStrength(password); err != nil {
		return "", err
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash 验证密码
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePasswordStrength 验证密码强度
func ValidatePasswordStrength(password string) error {
	if len(password) < 6 {
		return errors.New("密码长度不能少于6位")
	}
	if len(password) > 32 {
		return errors.New("密码长度不能超过32位")
	}

	// 最基本的密码验证：不能为空且长度合适
	if password == "" {
		return errors.New("密码不能为空")
	}

	return nil
}

// isSpecialChar 检查是否为特殊字符
func isSpecialChar(char rune) bool {
	specialChars := "!@#$%^&*()_+-=[]{}|;':\",./<>?"
	for _, sc := range specialChars {
		if char == sc {
			return true
		}
	}
	return false
}

// ValidateInput 验证输入是否包含潜在的SQL注入
func ValidateInput(input string) error {
	// 检测常见的SQL注入模式
	sqlPatterns := []string{
		`(?i)(\s*(union|select|insert|update|delete|drop|create|alter|exec|execute)\s+)`,
		`(?i)(\s*(or|and)\s+\d+\s*=\s*\d+)`,
		`(?i)(\s*['";](\s*--|\s*/\*))`,
		`(?i)(\s*'\s*(or|and)\s*'[^']*'\s*=\s*'[^']*')`,
		`(?i)(union\s+select)`,
		`(?i)(insert\s+into)`,
		`(?i)(drop\s+table)`,
	}

	for _, pattern := range sqlPatterns {
		matched, _ := regexp.MatchString(pattern, input)
		if matched {
			return errors.New("输入包含不安全的字符")
		}
	}

	return nil
}
