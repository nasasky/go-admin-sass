# CORS 跨域配置说明

## 问题描述

当前项目在同一服务器但使用不同二级域名访问时会出现CORS跨域拦截问题。

## 解决方案

已经为项目配置了灵活的CORS中间件，支持多种域名配置方式。

## 配置方式

### 1. 环境变量配置（推荐）

在你的环境变量或`.env`文件中设置：

```bash
# 允许多个域名，用逗号分隔
ALLOWED_ORIGINS=http://localhost:3000,https://admin.yourdomain.com,https://api.yourdomain.com,https://app.yourdomain.com

# 或者使用通配符支持所有二级域名
ALLOWED_ORIGINS=*.yourdomain.com,*.yourdomain.cn

# 开发环境可以允许所有域名
ALLOWED_ORIGINS=*
```

### 2. 代码配置

在 `config/cors.go` 文件中修改默认配置：

```go
allowedOrigins = []string{
    "http://localhost:3000",
    "https://admin.yourdomain.com",
    "https://api.yourdomain.com", 
    "https://app.yourdomain.com",
    "*.yourdomain.com", // 支持所有二级域名
}
```

### 3. 不同环境的配置示例

#### 开发环境
```bash
ALLOWED_ORIGINS=*
GIN_MODE=debug
```

#### 测试环境
```bash
ALLOWED_ORIGINS=*.test.yourdomain.com,http://localhost:*
GIN_MODE=test
```

#### 生产环境
```bash
ALLOWED_ORIGINS=https://admin.yourdomain.com,https://api.yourdomain.com,https://app.yourdomain.com
GIN_MODE=release
```

## 支持的域名格式

1. **完整域名**: `https://admin.yourdomain.com`
2. **通配符子域名**: `*.yourdomain.com`
3. **端口通配符**: `http://localhost:*`
4. **全部允许**: `*` （仅建议在开发环境使用）

## 验证配置

启动服务后，可以通过以下方式验证CORS配置：

### 1. 浏览器开发者工具
检查网络请求的响应头是否包含：
- `Access-Control-Allow-Origin`
- `Access-Control-Allow-Methods`
- `Access-Control-Allow-Headers`

### 2. curl测试
```bash
curl -H "Origin: https://admin.yourdomain.com" \
     -H "Access-Control-Request-Method: POST" \
     -H "Access-Control-Request-Headers: X-Requested-With" \
     -X OPTIONS \
     http://localhost:8801/api/admin/login
```

### 3. 健康检查
访问 `http://localhost:8801/health` 检查服务状态

## 当前配置特性

✅ 支持二级域名通配符匹配  
✅ 支持环境变量动态配置  
✅ 自动区分开发/生产环境  
✅ 完整的HTTP方法支持  
✅ 安全的请求头配置  
✅ 预检请求(OPTIONS)支持  

## 常见问题

### Q: 为什么还是被CORS拦截？
A: 检查以下几点：
1. 确认环境变量 `ALLOWED_ORIGINS` 是否正确设置
2. 确认请求的域名格式是否与配置匹配
3. 确认服务是否重启以应用新配置

### Q: 通配符不生效？
A: 确保通配符格式正确：
- ✅ `*.yourdomain.com`
- ❌ `*yourdomain.com`
- ❌ `*.yourdomain.com/*`

### Q: 如何调试CORS问题？
A: 
1. 查看浏览器控制台的错误信息
2. 检查Network标签页中的预检请求(OPTIONS)
3. 查看服务器日志中的请求记录

## 安全建议

🔒 **生产环境**: 明确指定允许的域名，避免使用通配符  
🔒 **敏感接口**: 考虑额外的Origin验证  
🔒 **HTTPS**: 生产环境强制使用HTTPS  

## 示例启动命令

```bash
# 开发环境
ALLOWED_ORIGINS=* GIN_MODE=debug ./nasa-go-admin

# 生产环境  
ALLOWED_ORIGINS=https://admin.yourdomain.com,https://api.yourdomain.com GIN_MODE=release ./nasa-go-admin
``` 