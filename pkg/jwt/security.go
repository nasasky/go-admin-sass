package jwt

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"nasa-go-admin/redis"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var (
	ErrTokenInBlacklist = errors.New("token已被加入黑名单")
)

// JWTConfig JWT配置
type JWTConfig struct {
	SigningKey      string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Issuer          string
}

// SecureCustomClaims 安全的JWT载荷
type SecureCustomClaims struct {
	UID  int    `json:"uid"`
	RID  int    `json:"rid"`
	TYPE int    `json:"type"`
	JTI  string `json:"jti"` // JWT ID，用于黑名单
	jwt.RegisteredClaims
}

// LoadJWTConfig 加载JWT配置
func LoadJWTConfig() *JWTConfig {
	signingKey := os.Getenv("JWT_SIGNING_KEY")
	if signingKey == "" {
		log.Fatal("JWT_SIGNING_KEY environment variable is required")
	}

	// 验证密钥长度
	if len(signingKey) < 32 {
		log.Fatal("JWT_SIGNING_KEY must be at least 32 characters long")
	}

	return &JWTConfig{
		SigningKey:      signingKey,
		AccessTokenTTL:  time.Hour * 24,     // 访问令牌24小时
		RefreshTokenTTL: time.Hour * 24 * 7, // 刷新令牌7天
		Issuer:          "nasa-go-admin",
	}
}

// GenerateSecureKey 生成安全的JWT密钥
func GenerateSecureKey() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// SecureJWTManager 安全的JWT管理器
type SecureJWTManager struct {
	config    *JWTConfig
	blacklist *TokenBlacklist
}

// NewSecureJWTManager 创建安全的JWT管理器
func NewSecureJWTManager() *SecureJWTManager {
	config := LoadJWTConfig()
	blacklist := NewTokenBlacklist()

	return &SecureJWTManager{
		config:    config,
		blacklist: blacklist,
	}
}

// GenerateToken 生成安全的token
func (sjm *SecureJWTManager) GenerateToken(uid, rid, userType int) (string, error) {
	jti := uuid.New().String() // 生成唯一的JWT ID

	claims := SecureCustomClaims{
		UID:  uid,
		RID:  rid,
		TYPE: userType,
		JTI:  jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(sjm.config.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    sjm.config.Issuer,
			Subject:   fmt.Sprintf("user:%d", uid),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(sjm.config.SigningKey))
}

// ValidateToken 验证token（包含黑名单检查）
func (sjm *SecureJWTManager) ValidateToken(tokenString string) (*SecureCustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &SecureCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return []byte(sjm.config.SigningKey), nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, errors.New("token格式错误")
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, errors.New("token已过期")
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, errors.New("token尚未激活")
			} else {
				return nil, errors.New("token无效")
			}
		}
		return nil, err
	}

	if claims, ok := token.Claims.(*SecureCustomClaims); ok && token.Valid {
		// 检查黑名单
		isBlacklisted, err := sjm.blacklist.IsBlacklisted(claims.JTI)
		if err != nil {
			return nil, fmt.Errorf("检查黑名单失败: %v", err)
		}

		if isBlacklisted {
			return nil, ErrTokenInBlacklist
		}

		return claims, nil
	}

	return nil, errors.New("token解析失败")
}

// RevokeToken 撤销token（加入黑名单）
func (sjm *SecureJWTManager) RevokeToken(tokenString string) error {
	claims, err := sjm.ValidateToken(tokenString)
	if err != nil {
		return err
	}

	return sjm.blacklist.AddToBlacklist(claims.JTI, claims.ExpiresAt.Time)
}

// TokenBlacklist Token黑名单管理
type TokenBlacklist struct {
}

const (
	blacklistPrefix = "jwt_blacklist:"
)

// NewTokenBlacklist 创建Token黑名单管理器
func NewTokenBlacklist() *TokenBlacklist {
	return &TokenBlacklist{}
}

// AddToBlacklist 将token添加到黑名单
func (tb *TokenBlacklist) AddToBlacklist(tokenID string, expiry time.Time) error {
	key := blacklistPrefix + tokenID
	duration := time.Until(expiry)

	if duration <= 0 {
		return nil // token已过期，无需加入黑名单
	}

	redisClient := redis.GetClient()
	if redisClient == nil {
		return errors.New("Redis客户端未初始化")
	}

	return redisClient.Set(context.Background(), key, "1", duration).Err()
}

// IsBlacklisted 检查token是否在黑名单中
func (tb *TokenBlacklist) IsBlacklisted(tokenID string) (bool, error) {
	key := blacklistPrefix + tokenID
	redisClient := redis.GetClient()
	if redisClient == nil {
		return false, errors.New("Redis客户端未初始化")
	}

	result := redisClient.Exists(context.Background(), key)
	return result.Val() > 0, result.Err()
}
