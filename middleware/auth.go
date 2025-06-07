package middleware

import (
	"strings"

	"nasa-go-admin/pkg/jwt"
	"nasa-go-admin/pkg/response"

	"github.com/gin-gonic/gin"
)

// JWTAuthConfig JWT认证配置
type JWTAuthConfig struct {
	TokenType jwt.TokenType
	SkipPaths []string // 跳过认证的路径
}

// JWTAuth JWT认证中间件
func JWTAuth(config JWTAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否需要跳过认证
		for _, path := range config.SkipPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		// 获取token
		token := getTokenFromRequest(c)
		if token == "" {
			response.Abort(c, response.AUTH_ERROR, "请求未携带token，无权限访问")
			return
		}

		// 解析token
		var claims *jwt.CustomClaims
		var err error

		jwtManager := jwt.NewJWTManager(config.TokenType)
		claims, err = jwtManager.ParseToken(token)

		if err != nil {
			var message string
			switch err {
			case jwt.ErrTokenExpired:
				message = "授权已过期"
			case jwt.ErrTokenMalformed:
				message = "token格式错误"
			case jwt.ErrTokenNotValidYet:
				message = "token尚未激活"
			default:
				message = "token无效"
			}
			response.Abort(c, response.AUTH_ERROR, message)
			return
		}

		// 将用户信息存储到上下文
		c.Set("uid", claims.UID)
		c.Set("rid", claims.RID)
		c.Set("type", claims.TYPE)
		c.Set("claims", claims)

		c.Next()
	}
}

// AdminJWTAuth 管理员JWT认证中间件
func AdminJWTAuth() gin.HandlerFunc {
	return JWTAuth(JWTAuthConfig{
		TokenType: jwt.TokenTypeAdmin,
	})
}

// AppJWTAuth 应用JWT认证中间件
func AppJWTAuth() gin.HandlerFunc {
	return JWTAuth(JWTAuthConfig{
		TokenType: jwt.TokenTypeApp,
	})
}

// getTokenFromRequest 从请求中获取token
func getTokenFromRequest(c *gin.Context) string {
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

// ExtractUIDFromToken 从token字符串解析出用户ID，用于WebSocket连接
func ExtractUIDFromToken(tokenStr string, tokenType jwt.TokenType) (int, error) {
	if tokenStr == "" {
		return 0, jwt.ErrTokenInvalid
	}

	// 去掉可能的Bearer前缀
	if len(tokenStr) > 7 && strings.ToLower(tokenStr[:7]) == "bearer " {
		tokenStr = tokenStr[7:]
	}

	jwtManager := jwt.NewJWTManager(tokenType)
	return jwtManager.ExtractUID(tokenStr)
}

// RequireRole 角色权限检查中间件
func RequireRole(allowedRoles ...int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("rid")
		if !exists {
			response.Abort(c, response.FORBIDDEN, "无法获取用户角色信息")
			return
		}

		role, ok := userRole.(int)
		if !ok {
			response.Abort(c, response.FORBIDDEN, "用户角色信息格式错误")
			return
		}

		// 检查角色是否允许
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		response.Abort(c, response.FORBIDDEN, "无权限访问此资源")
	}
}

// RequireUserType 用户类型检查中间件
func RequireUserType(allowedTypes ...int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("type")
		if !exists {
			response.Abort(c, response.FORBIDDEN, "无法获取用户类型信息")
			return
		}

		uType, ok := userType.(int)
		if !ok {
			response.Abort(c, response.FORBIDDEN, "用户类型信息格式错误")
			return
		}

		// 检查用户类型是否允许
		for _, allowedType := range allowedTypes {
			if uType == allowedType {
				c.Next()
				return
			}
		}

		response.Abort(c, response.FORBIDDEN, "用户类型无权限访问此资源")
	}
}
