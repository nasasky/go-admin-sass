package middleware

import (
	"fmt"
	"log"
	"nasa-go-admin/redis"
	"sync"

	"github.com/gin-gonic/gin"
)

var userCache sync.Map // 内存缓存

func UserInfoMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("uid")
		if !exists {
			c.AbortWithStatusJSON(401, gin.H{"error": "用户未登录"})
			return
		}

		userID := fmt.Sprintf("%v", user) // 将 uid 转为字符串

		// 从内存缓存中获取用户信息
		if cachedUserInfo, found := userCache.Load(userID); found {
			c.Set("userInfo", cachedUserInfo)
			c.Next()
			return
		}

		// 如果缓存中没有，查询 Redis
		userInfo, err := redis.GetUserInfo(userID)
		if err != nil {
			log.Printf("Failed to get user info from Redis: %v", err)
			c.AbortWithStatusJSON(500, gin.H{"error": "获取用户信息失败"})
			return
		}

		// 将用户信息存储到内存缓存和 gin.Context
		userCache.Store(userID, userInfo)
		c.Set("userInfo", userInfo)

		c.Next()
	}
}
