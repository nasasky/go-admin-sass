package redis

import (
	"github.com/gin-gonic/gin"
)

func RedisMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("redis", rdb)
		c.Next()
	}
}
