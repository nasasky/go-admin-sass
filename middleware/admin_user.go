package middleware

import (
	"fmt"
	"log"
	"nasa-go-admin/redis"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var userCache sync.Map // 内存缓存

func UserInfoMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("uid")
		if !exists {
			log.Printf("[DEBUG] No user ID found in context")
			c.AbortWithStatusJSON(401, gin.H{"error": "用户未登录"})
			return
		}

		userID := fmt.Sprintf("%v", user) // 将 uid 转为字符串
		log.Printf("[DEBUG] Processing user ID: %s", userID)

		// 从内存缓存中获取用户信息
		if cachedUserInfo, found := userCache.Load(userID); found {
			log.Printf("[DEBUG] Found user info in memory cache for user %s", userID)
			c.Set("userInfo", cachedUserInfo)
			c.Next()
			return
		}

		// 如果缓存中没有，尝试多次查询 Redis
		var userInfo map[string]string
		var err error
		maxRetries := 3
		for i := 0; i < maxRetries; i++ {
			userInfo, err = redis.GetUserInfo(userID)
			if err == nil {
				break
			}
			if i < maxRetries-1 { // 不是最后一次尝试
				log.Printf("[DEBUG] Retry %d: Failed to get user info from Redis for user %s: %v", i+1, userID, err)
				time.Sleep(time.Millisecond * 100) // 短暂延迟后重试
				continue
			}
			// 最后一次尝试也失败了
			log.Printf("[ERROR] All retries failed to get user info from Redis for user %s: %v", userID, err)
			c.AbortWithStatusJSON(500, gin.H{"error": "获取用户信息失败"})
			return
		}

		// 检查用户信息是否完整
		if userInfo == nil || len(userInfo) == 0 {
			log.Printf("[ERROR] Empty user info returned from Redis for user %s", userID)
			c.AbortWithStatusJSON(500, gin.H{"error": "用户信息不完整"})
			return
		}

		// 将用户信息存储到内存缓存和 gin.Context
		userCache.Store(userID, userInfo)
		c.Set("userInfo", userInfo)
		log.Printf("[DEBUG] Successfully stored user info in memory cache for user %s", userID)

		c.Next()
	}
}
