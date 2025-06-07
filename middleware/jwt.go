package middleware

import (
	"fmt"
	"nasa-go-admin/api"
	"nasa-go-admin/pkg/jwt"
	"nasa-go-admin/utils"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	jwtLib "github.com/golang-jwt/jwt/v5" // 或您使用的JWT库
)

// 简单的令牌缓存
var (
	tokenCache = make(map[string]tokenCacheEntry)
	cacheMutex = &sync.RWMutex{}
)

type tokenCacheEntry struct {
	UserID    int
	ExpiresAt time.Time
}

// JWTAuth 中间件，检查token
func Jwt() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			api.Resp.Err(c, 10002, "请求未携带token，无权限访问")
			c.Abort()
			return
		}
		// 去掉Bearer前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
		// parseToken 解析token包含的信息
		claims, err := jwt.ParseAdminToken(token)
		if err != nil {
			if err == jwt.ErrTokenExpired {
				api.Resp.Err(c, 10002, "授权已过期")
				c.Abort()
				return
			}
			api.Resp.Err(c, 10002, err.Error())
			c.Abort()
			return
		}

		// 继续交由下一个路由处理,并将解析出的信息传递下去
		c.Set("uid", claims.UID)
		c.Set("rid", claims.RID)
		c.Set("type", claims.TYPE)
		c.Next()
	}
}

// ParseTokenGetUID 从token字符串解析出用户ID，用于WebSocket连接
func ParseTokenGetUID(tokenStr string) (int, error) {
	if tokenStr == "" {
		return 0, fmt.Errorf("token不能为空")
	}

	// 去掉可能的Bearer前缀
	if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
		tokenStr = tokenStr[7:]
	}

	// 使用与中间件相同的JWT处理工具
	j := utils.NewJWT()
	claims, err := j.ParseToken(tokenStr)
	if err != nil {
		if err == utils.TokenExpired {
			return 0, fmt.Errorf("token已过期")
		}
		return 0, err
	}

	// 返回用户ID
	return claims.UID, nil
}

// ParseTokenGetUIDWithCache 解析JWT令牌并返回用户ID，使用缓存提高性能
func ParseTokenGetUIDWithCache(tokenString string) (int, error) {
	// 检查缓存
	cacheMutex.RLock()
	entry, found := tokenCache[tokenString]
	cacheMutex.RUnlock()

	// 如果在缓存中且未过期，直接返回
	if found && time.Now().Before(entry.ExpiresAt) {
		return entry.UserID, nil
	}

	// 解析令牌
	token, err := jwtLib.Parse(tokenString, func(token *jwtLib.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwtLib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// 返回用于验证的密钥
		return []byte("your-secret-key"), nil // 使用您的实际密钥
	})

	if err != nil {
		return 0, fmt.Errorf("无效令牌: %w", err)
	}

	// 提取声明
	if claims, ok := token.Claims.(jwtLib.MapClaims); ok && token.Valid {
		if userID, ok := claims["user_id"].(float64); ok {
			// 缓存结果
			expiresAt := time.Now().Add(30 * time.Minute) // 缓存30分钟
			if exp, ok := claims["exp"].(float64); ok {
				expTime := time.Unix(int64(exp), 0)
				if expTime.Before(expiresAt) {
					expiresAt = expTime
				}
			}

			cacheMutex.Lock()
			tokenCache[tokenString] = tokenCacheEntry{
				UserID:    int(userID),
				ExpiresAt: expiresAt,
			}
			cacheMutex.Unlock()

			return int(userID), nil
		}
		return 0, fmt.Errorf("令牌中无法找到用户ID")
	}

	return 0, fmt.Errorf("无效令牌")
}

// 定期清理过期缓存项的协程
func init() {
	go func() {
		for {
			time.Sleep(15 * time.Minute)
			cleanExpiredTokens()
		}
	}()
}

// 清理过期的令牌
func cleanExpiredTokens() {
	now := time.Now()
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	for token, entry := range tokenCache {
		if now.After(entry.ExpiresAt) {
			delete(tokenCache, token)
		}
	}
}
