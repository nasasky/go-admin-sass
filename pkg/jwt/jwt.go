package jwt

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWT错误定义
var (
	ErrTokenExpired     = errors.New("token已过期")
	ErrTokenNotValidYet = errors.New("token尚未激活")
	ErrTokenMalformed   = errors.New("token格式错误")
	ErrTokenInvalid     = errors.New("token无效")
)

// CustomClaims JWT载荷
type CustomClaims struct {
	UID  int `json:"uid"`
	RID  int `json:"rid"`
	TYPE int `json:"type"`
	jwt.RegisteredClaims
}

// TokenType 令牌类型
type TokenType string

const (
	TokenTypeAdmin TokenType = "admin"
	TokenTypeApp   TokenType = "app"
)

// JWTManager JWT管理器
type JWTManager struct {
	signingKey []byte
	tokenType  TokenType
}

// NewJWTManager 创建JWT管理器
func NewJWTManager(tokenType TokenType) *JWTManager {
	signingKey := os.Getenv("JWT_SIGNING_KEY")
	if signingKey == "" {
		signingKey = "default-secret-key" // 开发环境默认值，生产环境必须设置
	}
	
	return &JWTManager{
		signingKey: []byte(signingKey),
		tokenType:  tokenType,
	}
}

// GenerateToken 生成token
func (j *JWTManager) GenerateToken(uid, rid, userType int, duration ...time.Duration) (string, error) {
	expiry := time.Hour // 默认1小时
	if len(duration) > 0 {
		expiry = duration[0]
	}

	claims := CustomClaims{
		UID:  uid,
		RID:  rid,
		TYPE: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    fmt.Sprintf("nasa-go-admin-%s", j.tokenType),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.signingKey)
}

// ParseToken 解析token
func (j *JWTManager) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return j.signingKey, nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, ErrTokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, ErrTokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, ErrTokenNotValidYet
			} else {
				return nil, ErrTokenInvalid
			}
		}
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// RefreshToken 刷新token
func (j *JWTManager) RefreshToken(tokenString string, duration ...time.Duration) (string, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 如果token还未过期太久，允许刷新
	expiry := time.Hour
	if len(duration) > 0 {
		expiry = duration[0]
	}

	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(expiry))
	claims.IssuedAt = jwt.NewNumericDate(time.Now())

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.signingKey)
}

// ValidateToken 验证token是否有效（不解析完整内容）
func (j *JWTManager) ValidateToken(tokenString string) bool {
	_, err := j.ParseToken(tokenString)
	return err == nil
}

// ExtractUID 从token中提取用户ID
func (j *JWTManager) ExtractUID(tokenString string) (int, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UID, nil
}

// 便捷函数
func GenerateAdminToken(uid, rid, userType int, duration ...time.Duration) (string, error) {
	manager := NewJWTManager(TokenTypeAdmin)
	return manager.GenerateToken(uid, rid, userType, duration...)
}

func GenerateAppToken(uid, rid, userType int, duration ...time.Duration) (string, error) {
	manager := NewJWTManager(TokenTypeApp)
	return manager.GenerateToken(uid, rid, userType, duration...)
}

func ParseAdminToken(tokenString string) (*CustomClaims, error) {
	manager := NewJWTManager(TokenTypeAdmin)
	return manager.ParseToken(tokenString)
}

func ParseAppToken(tokenString string) (*CustomClaims, error) {
	manager := NewJWTManager(TokenTypeApp)
	return manager.ParseToken(tokenString)
} 