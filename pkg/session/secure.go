package session

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
)

// InitSecureSession 初始化安全会话
func InitSecureSession(r *gin.Engine, redisAddr, redisPassword string) {
	// 生成随机密钥
	authKey := make([]byte, 32)
	encryptionKey := make([]byte, 32)
	if _, err := rand.Read(authKey); err != nil {
		log.Fatal("Failed to generate auth key:", err)
	}
	if _, err := rand.Read(encryptionKey); err != nil {
		log.Fatal("Failed to generate encryption key:", err)
	}

	// 使用Redis存储会话
	authKeyStr := hex.EncodeToString(authKey)
	store, err := redis.NewStore(10, "tcp", redisAddr, redisPassword, authKeyStr)
	if err != nil {
		log.Fatal("Failed to create session store:", err)
	}

	// 配置会话选项
	store.Options(sessions.Options{
		MaxAge:   3600,                          // 1小时过期
		Secure:   gin.Mode() == gin.ReleaseMode, // 生产环境仅HTTPS
		HttpOnly: true,                          // 防止XSS
		SameSite: http.SameSiteStrictMode,       // CSRF保护
		Path:     "/",
	})

	r.Use(sessions.Sessions("secure_session", store))
}

// SecureCaptchaMiddleware 安全验证码中间件
func SecureCaptchaMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		// 设置安全的会话选项
		session.Options(sessions.Options{
			MaxAge:   300, // 验证码5分钟过期
			Secure:   gin.Mode() == gin.ReleaseMode,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})

		c.Next()
	}
}
