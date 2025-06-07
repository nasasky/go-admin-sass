# 安全修复实施方案

## 紧急安全修复清单

### 1. 密码哈希算法升级 (最高优先级)

#### 当前问题
- 系统多处使用 MD5 哈希密码
- MD5 算法已被认为不安全，容易被彩虹表攻击

#### 修复方案

**1.1 添加 bcrypt 依赖**
```bash
go get golang.org/x/crypto/bcrypt
```

**1.2 创建安全密码工具包**
```go
// pkg/security/password.go
package security

import (
    "golang.org/x/crypto/bcrypt"
    "errors"
)

const (
    // bcrypt 推荐的最小成本
    MinCost     = 12
    DefaultCost = 14
)

// HashPassword 使用 bcrypt 哈希密码
func HashPassword(password string) (string, error) {
    if len(password) < 8 {
        return "", errors.New("密码长度不能少于8位")
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
    if len(password) < 8 {
        return errors.New("密码长度不能少于8位")
    }
    if len(password) > 32 {
        return errors.New("密码长度不能超过32位")
    }
    
    hasUpper := false
    hasLower := false
    hasDigit := false
    hasSpecial := false
    
    for _, char := range password {
        switch {
        case 'A' <= char && char <= 'Z':
            hasUpper = true
        case 'a' <= char && char <= 'z':
            hasLower = true
        case '0' <= char && char <= '9':
            hasDigit = true
        case char == '!' || char == '@' || char == '#' || char == '$' || 
             char == '%' || char == '^' || char == '&' || char == '*':
            hasSpecial = true
        }
    }
    
    if !hasUpper {
        return errors.New("密码必须包含大写字母")
    }
    if !hasLower {
        return errors.New("密码必须包含小写字母")
    }
    if !hasDigit {
        return errors.New("密码必须包含数字")
    }
    if !hasSpecial {
        return errors.New("密码必须包含特殊字符(!@#$%^&*)")
    }
    
    return nil
}
```

**1.3 数据库迁移脚本**
```sql
-- 数据库迁移：添加新的密码字段
ALTER TABLE admin_user ADD COLUMN password_bcrypt VARCHAR(255);
ALTER TABLE app_user ADD COLUMN password_bcrypt VARCHAR(255);

-- 迁移完成后删除旧字段（在所有密码都迁移后执行）
-- ALTER TABLE admin_user DROP COLUMN password;
-- ALTER TABLE app_user DROP COLUMN password;
-- ALTER TABLE admin_user CHANGE password_bcrypt password VARCHAR(255);
-- ALTER TABLE app_user CHANGE password_bcrypt password VARCHAR(255);
```

**1.4 密码迁移工具**
```go
// tools/migrate_passwords.go
package main

import (
    "crypto/md5"
    "fmt"
    "log"
    "nasa-go-admin/db"
    "nasa-go-admin/pkg/security"
    "nasa-go-admin/model/admin_model"
    "nasa-go-admin/model/app_model"
)

func main() {
    db.Init()
    
    // 迁移管理员密码
    migrateAdminPasswords()
    
    // 迁移应用用户密码
    migrateAppUserPasswords()
    
    log.Println("密码迁移完成")
}

func migrateAdminPasswords() {
    var users []admin_model.AdminUser
    db.Dao.Find(&users)
    
    for _, user := range users {
        // 如果已经迁移过，跳过
        if user.PasswordBcrypt != "" {
            continue
        }
        
        // 生成临时密码或要求用户重置密码
        tempPassword := generateTempPassword()
        hashedPassword, err := security.HashPassword(tempPassword)
        if err != nil {
            log.Printf("Failed to hash password for user %d: %v", user.ID, err)
            continue
        }
        
        db.Dao.Model(&user).Update("password_bcrypt", hashedPassword)
        log.Printf("Migrated password for admin user %d, temp password: %s", user.ID, tempPassword)
    }
}

func migrateAppUserPasswords() {
    var users []app_model.AppUser
    db.Dao.Find(&users)
    
    for _, user := range users {
        if user.PasswordBcrypt != "" {
            continue
        }
        
        tempPassword := generateTempPassword()
        hashedPassword, err := security.HashPassword(tempPassword)
        if err != nil {
            log.Printf("Failed to hash password for user %d: %v", user.ID, err)
            continue
        }
        
        db.Dao.Model(&user).Update("password_bcrypt", hashedPassword)
        log.Printf("Migrated password for app user %d, temp password: %s", user.ID, tempPassword)
    }
}

func generateTempPassword() string {
    // 生成随机临时密码
    return fmt.Sprintf("Temp%d@", time.Now().Unix()%10000)
}
```

### 2. JWT 安全加强

#### 当前问题
- JWT 密钥可能使用默认值
- Token 过期时间过长
- 缺少 Token 黑名单机制

#### 修复方案

**2.1 JWT 配置加强**
```go
// pkg/jwt/security.go
package jwt

import (
    "crypto/rand"
    "encoding/hex"
    "errors"
    "os"
    "time"
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

// LoadJWTConfig 加载JWT配置
func LoadJWTConfig() *JWTConfig {
    signingKey := os.Getenv("JWT_SIGNING_KEY")
    if signingKey == "" {
        panic("JWT_SIGNING_KEY environment variable is required")
    }
    
    return &JWTConfig{
        SigningKey:      signingKey,
        AccessTokenTTL:  time.Hour * 2,     // 访问令牌2小时
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
```

**2.2 Token 黑名单机制**
```go
// pkg/jwt/blacklist.go
package jwt

import (
    "context"
    "fmt"
    "nasa-go-admin/redis"
    "time"
)

const (
    blacklistPrefix = "jwt_blacklist:"
)

// TokenBlacklist Token黑名单管理
type TokenBlacklist struct {
    redisClient *redis.Client
}

// NewTokenBlacklist 创建Token黑名单管理器
func NewTokenBlacklist() *TokenBlacklist {
    return &TokenBlacklist{
        redisClient: redis.GetClient(),
    }
}

// AddToBlacklist 将token添加到黑名单
func (tb *TokenBlacklist) AddToBlacklist(tokenID string, expiry time.Time) error {
    key := blacklistPrefix + tokenID
    duration := time.Until(expiry)
    
    if duration <= 0 {
        return nil // token已过期，无需加入黑名单
    }
    
    return tb.redisClient.Set(context.Background(), key, "1", duration).Err()
}

// IsBlacklisted 检查token是否在黑名单中
func (tb *TokenBlacklist) IsBlacklisted(tokenID string) (bool, error) {
    key := blacklistPrefix + tokenID
    result := tb.redisClient.Exists(context.Background(), key)
    return result.Val() > 0, result.Err()
}

// 更新JWT管理器以支持黑名单
func (j *JWTManager) ValidateTokenWithBlacklist(tokenString string) (*CustomClaims, error) {
    claims, err := j.ParseToken(tokenString)
    if err != nil {
        return nil, err
    }
    
    // 检查黑名单
    blacklist := NewTokenBlacklist()
    isBlacklisted, err := blacklist.IsBlacklisted(claims.Id)
    if err != nil {
        return nil, err
    }
    
    if isBlacklisted {
        return nil, ErrTokenInBlacklist
    }
    
    return claims, nil
}
```

### 3. 输入验证和SQL注入防护

#### 修复方案

**3.1 统一输入验证中间件**
```go
// middleware/validation.go
package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/go-playground/validator/v10"
    "nasa-go-admin/pkg/response"
    "strings"
)

var validate *validator.Validate

func init() {
    validate = validator.New()
    
    // 注册自定义验证规则
    validate.RegisterValidation("noSqlInjection", validateNoSqlInjection)
    validate.RegisterValidation("strongPassword", validateStrongPassword)
}

// StrictValidationMiddleware 严格验证中间件
func StrictValidationMiddleware(obj interface{}) gin.HandlerFunc {
    return func(c *gin.Context) {
        if err := c.ShouldBindJSON(obj); err != nil {
            response.Error(c, response.INVALID_PARAMS, "请求参数格式错误")
            c.Abort()
            return
        }
        
        if err := validate.Struct(obj); err != nil {
            response.Error(c, response.INVALID_PARAMS, formatValidationError(err))
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// validateNoSqlInjection SQL注入检测
func validateNoSqlInjection(fl validator.FieldLevel) bool {
    value := fl.Field().String()
    
    // 检测常见的SQL注入模式
    sqlPatterns := []string{
        "' or '1'='1",
        "'; drop table",
        "'; delete from",
        "union select",
        "insert into",
        "update set",
        "' or 1=1",
        "admin'--",
        "admin'/*",
    }
    
    lowerValue := strings.ToLower(value)
    for _, pattern := range sqlPatterns {
        if strings.Contains(lowerValue, pattern) {
            return false
        }
    }
    
    return true
}

// validateStrongPassword 强密码验证
func validateStrongPassword(fl validator.FieldLevel) bool {
    password := fl.Field().String()
    return security.ValidatePasswordStrength(password) == nil
}

func formatValidationError(err error) string {
    if validationErrors, ok := err.(validator.ValidationErrors); ok {
        for _, e := range validationErrors {
            return formatFieldError(e)
        }
    }
    return "输入验证失败"
}

func formatFieldError(e validator.FieldError) string {
    switch e.Tag() {
    case "required":
        return fmt.Sprintf("%s是必填字段", e.Field())
    case "min":
        return fmt.Sprintf("%s最小长度为%s", e.Field(), e.Param())
    case "max":
        return fmt.Sprintf("%s最大长度为%s", e.Field(), e.Param())
    case "noSqlInjection":
        return fmt.Sprintf("%s包含不安全的字符", e.Field())
    case "strongPassword":
        return "密码强度不够，需要包含大小写字母、数字和特殊字符"
    default:
        return fmt.Sprintf("%s格式错误", e.Field())
    }
}
```

**3.2 更新请求结构体**
```go
// inout/secure_req.go
package inout

// SecureLoginReq 安全登录请求
type SecureLoginReq struct {
    Username string `json:"username" binding:"required,min=3,max=20,noSqlInjection"`
    Password string `json:"password" binding:"required,strongPassword"`
    Captcha  string `json:"captcha" binding:"required,len=4"`
}

// SecureRegisterReq 安全注册请求
type SecureRegisterReq struct {
    Username string `json:"username" binding:"required,min=3,max=20,noSqlInjection"`
    Password string `json:"password" binding:"required,strongPassword"`
    Phone    string `json:"phone" binding:"required,len=11"`
    Captcha  string `json:"captcha" binding:"required,len=4"`
}
```

### 4. 安全中间件集成

**4.1 安全头中间件**
```go
// middleware/security.go
package middleware

import (
    "github.com/gin-gonic/gin"
    "time"
)

// SecurityHeaders 安全头中间件
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 防止MIME类型嗅探
        c.Header("X-Content-Type-Options", "nosniff")
        
        // 防止点击劫持
        c.Header("X-Frame-Options", "DENY")
        
        // XSS保护
        c.Header("X-XSS-Protection", "1; mode=block")
        
        // 推荐人策略
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        
        // 内容安全策略
        c.Header("Content-Security-Policy", "default-src 'self'")
        
        // HTTPS强制（生产环境）
        if gin.Mode() == gin.ReleaseMode {
            c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
        }
        
        c.Next()
    }
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware() gin.HandlerFunc {
    // 使用 Redis 实现分布式限流
    return func(c *gin.Context) {
        clientIP := c.ClientIP()
        
        // 检查限流
        allowed, err := checkRateLimit(clientIP)
        if err != nil {
            c.JSON(500, gin.H{"error": "服务器内部错误"})
            c.Abort()
            return
        }
        
        if !allowed {
            c.JSON(429, gin.H{"error": "请求过于频繁，请稍后再试"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}

func checkRateLimit(clientIP string) (bool, error) {
    // 实现基于Redis的滑动窗口限流
    // 每分钟最多100个请求
    key := fmt.Sprintf("rate_limit:%s", clientIP)
    now := time.Now().Unix()
    
    // 使用Redis pipeline提高性能
    pipe := redis.GetClient().Pipeline()
    pipe.ZRemRangeByScore(context.Background(), key, "0", fmt.Sprintf("%d", now-60))
    pipe.ZCard(context.Background(), key)
    pipe.ZAdd(context.Background(), key, &redis.Z{Score: float64(now), Member: now})
    pipe.Expire(context.Background(), key, time.Minute)
    
    results, err := pipe.Exec(context.Background())
    if err != nil {
        return false, err
    }
    
    count := results[1].(*redis.IntCmd).Val()
    return count < 100, nil
}
```

### 5. 会话安全

**5.1 安全会话管理**
```go
// pkg/session/secure.go
package session

import (
    "crypto/rand"
    "github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/redis"
    "github.com/gin-gonic/gin"
)

// InitSecureSession 初始化安全会话
func InitSecureSession(r *gin.Engine) {
    // 生成随机密钥
    authKey := make([]byte, 32)
    encryptionKey := make([]byte, 32)
    rand.Read(authKey)
    rand.Read(encryptionKey)
    
    // 使用Redis存储会话
    store, err := redis.NewStore(10, "tcp", "localhost:6379", "", authKey, encryptionKey)
    if err != nil {
        panic("Failed to create session store: " + err.Error())
    }
    
    // 配置会话选项
    store.Options(sessions.Options{
        MaxAge:   3600,          // 1小时过期
        Secure:   true,          // 仅HTTPS
        HttpOnly: true,          // 防止XSS
        SameSite: http.SameSiteStrictMode, // CSRF保护
    })
    
    r.Use(sessions.Sessions("secure_session", store))
}
```

## 实施计划

### 第一阶段：紧急修复（1周内）
1. **立即实施密码哈希升级**
   - 部署bcrypt工具包
   - 执行密码迁移脚本
   - 通知所有用户重设密码

2. **JWT安全加强**
   - 更新JWT密钥配置
   - 缩短Token过期时间
   - 实施Token黑名单

3. **添加基础安全中间件**
   - 部署安全头中间件
   - 实施基础限流

### 第二阶段：全面加强（2周内）
1. **完善输入验证**
   - 更新所有API接口验证
   - 添加SQL注入防护
   - 实施严格的参数校验

2. **会话安全**
   - 升级会话管理
   - 实施安全的Cookie配置

3. **监控和告警**
   - 添加安全事件监控
   - 实施异常登录告警

### 验证和测试
1. **安全测试**
   - 进行渗透测试
   - SQL注入测试
   - XSS攻击测试

2. **性能测试**
   - 验证安全措施对性能的影响
   - 调优限流参数

3. **功能测试**
   - 全面回归测试
   - 用户体验测试

## 注意事项

1. **数据备份**：在执行密码迁移前，务必备份数据库
2. **用户通知**：提前通知用户密码策略变更
3. **渐进部署**：建议先在测试环境验证，再逐步部署到生产环境
4. **监控告警**：部署后密切监控系统运行状态
5. **回滚计划**：准备快速回滚方案以应对紧急情况

这些安全修复措施将显著提升系统的安全性，建议按照优先级逐步实施，确保每个阶段都经过充分测试验证。 