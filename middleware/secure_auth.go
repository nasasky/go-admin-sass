package middleware

import (
	"strings"

	"nasa-go-admin/pkg/jwt"
	"nasa-go-admin/pkg/response"
	"nasa-go-admin/pkg/security"

	"github.com/gin-gonic/gin"
)

// SecureJWTAuth 安全的JWT认证中间件
func SecureJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取token
		token := getSecureTokenFromRequest(c)
		if token == "" {
			response.Abort(c, response.AUTH_ERROR, "请求未携带token，无权限访问")
			return
		}

		// 验证输入安全性
		if err := security.ValidateInput(token); err != nil {
			response.Abort(c, response.AUTH_ERROR, "token包含非法字符")
			return
		}

		// 使用安全的JWT管理器验证token
		jwtManager := jwt.NewSecureJWTManager()
		claims, err := jwtManager.ValidateToken(token)

		if err != nil {
			var message string
			switch err {
			case jwt.ErrTokenInBlacklist:
				message = "token已被撤销"
			default:
				message = err.Error()
			}
			response.Abort(c, response.AUTH_ERROR, message)
			return
		}

		// 将用户信息存储到上下文
		c.Set("uid", claims.UID)
		c.Set("rid", claims.RID)
		c.Set("type", claims.TYPE)
		c.Set("jti", claims.JTI)
		c.Set("claims", claims)

		c.Next()
	}
}

// SecureAdminJWTAuth 安全的管理员JWT认证中间件
func SecureAdminJWTAuth() gin.HandlerFunc {
	return SecureJWTAuth()
}

// SecureAppJWTAuth 安全的应用JWT认证中间件
func SecureAppJWTAuth() gin.HandlerFunc {
	return SecureJWTAuth()
}

// getSecureTokenFromRequest 从请求中获取token
func getSecureTokenFromRequest(c *gin.Context) string {
	// 1. 从Authorization header获取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// Bearer token格式
		if len(authHeader) > 7 && strings.ToLower(authHeader[:7]) == "bearer " {
			return authHeader[7:]
		}
		return authHeader
	}

	// 2. 从查询参数获取
	if token := c.Query("token"); token != "" {
		return token
	}

	// 3. 从表单参数获取
	if token := c.PostForm("token"); token != "" {
		return token
	}

	return ""
}

// RevokeTokenMiddleware 撤销token的中间件
func RevokeTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 在退出登录时撤销token
		if c.Request.URL.Path == "/auth/logout" || c.Request.URL.Path == "/api/admin/logout" {
			token := getSecureTokenFromRequest(c)
			if token != "" {
				jwtManager := jwt.NewSecureJWTManager()
				jwtManager.RevokeToken(token)
			}
		}
		c.Next()
	}
}
