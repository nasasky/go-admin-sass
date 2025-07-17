# CORS 跨域问题修复指南

## 问题描述
在局域网环境下访问项目时出现跨域(CORS)错误，前端无法正常请求后端API。

## 已修复的问题
1. ✅ 增强了CORS中间件的域名匹配功能
2. ✅ 添加了IP段通配符支持
3. ✅ 开发环境默认允许所有域名访问

## 解决方案

### 方案一：环境变量配置（推荐）

创建 `.env` 文件（或在系统环境变量中设置）：

```bash
# 开发环境 - 允许所有域名（最简单的解决方案）
ALLOWED_ORIGINS=*
GIN_MODE=debug

# 或者指定具体的局域网IP段（更安全）
# ALLOWED_ORIGINS=http://localhost:*,http://127.0.0.1:*,http://192.168.*:*,http://10.*:*

# 生产环境 - 指定具体域名
# ALLOWED_ORIGINS=https://yourdomain.com,https://admin.yourdomain.com
# GIN_MODE=release
```

### 方案二：启动时设置环境变量

```bash
# 开发环境启动
ALLOWED_ORIGINS=* GIN_MODE=debug ./nasa-go-admin

# 或指定局域网段
ALLOWED_ORIGINS="http://192.168.*:*,http://10.*:*,http://localhost:*" ./nasa-go-admin
```

### 方案三：修改配置文件

在 `config/config.yaml` 中添加更多允许的域名：

```yaml
security:
  allowed_origins:
    - "http://localhost:*"
    - "http://127.0.0.1:*" 
    - "http://192.168.*:*"  # 局域网段
    - "http://10.*:*"       # 企业内网段
```

## 支持的域名模式

现在支持以下几种域名配置模式：

1. **完全匹配**: `http://192.168.1.100:3000`
2. **全部允许**: `*`
3. **端口通配符**: `http://localhost:*`
4. **IP段通配符**: `http://192.168.*:*`
5. **子域名通配符**: `*.example.com`

## 常见局域网IP段

- `192.168.*.*` - 家庭/小型办公网络
- `10.*.*.*` - 企业内网
- `172.16.*.*` - `172.31.*.*` - 企业私有网络

## 验证修复结果

### 1. 检查响应头
在浏览器开发者工具的Network标签中，检查API响应是否包含：
```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD
Access-Control-Allow-Headers: Origin, Content-Type, Authorization, ...
```

### 2. 测试命令
```bash
# 测试OPTIONS预检请求
curl -X OPTIONS \
  -H "Origin: http://192.168.1.100:3000" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type,Authorization" \
  http://你的服务器IP:8801/api/admin/login

# 测试实际请求
curl -H "Origin: http://192.168.1.100:3000" \
  http://你的服务器IP:8801/api/health
```

### 3. 启动服务测试
```bash
# 使用环境变量启动
ALLOWED_ORIGINS=* GIN_MODE=debug ./nasa-go-admin

# 查看启动日志，确认CORS配置生效
```

## 安全建议

### 开发环境
- ✅ 使用 `ALLOWED_ORIGINS=*` 允许所有域名
- ✅ 设置 `GIN_MODE=debug`

### 生产环境
- ⚠️ 不要使用 `*` 通配符
- ✅ 明确指定允许的域名
- ✅ 使用HTTPS
- ✅ 设置 `GIN_MODE=release`

## 故障排除

### 如果仍然出现跨域问题：

1. **检查环境变量**
   ```bash
   echo $ALLOWED_ORIGINS
   echo $GIN_MODE
   ```

2. **查看服务器日志**
   检查启动时是否有CORS相关的错误信息

3. **浏览器控制台**
   查看具体的CORS错误信息，确认是哪个域名被拒绝

4. **重启服务**
   修改配置后需要重启服务才能生效

### 常见错误信息：
- `Access to fetch at 'http://...' from origin 'http://...' has been blocked by CORS policy`
- `No 'Access-Control-Allow-Origin' header is present on the requested resource`

## 快速修复步骤

1. **立即修复**（开发环境）：
   ```bash
   ALLOWED_ORIGINS=* GIN_MODE=debug ./nasa-go-admin
   ```

2. **创建 .env 文件**：
   ```bash
   echo "ALLOWED_ORIGINS=*" > .env
   echo "GIN_MODE=debug" >> .env
   ```

3. **重启服务**：
   ```bash
   ./stop.sh
   ./start.sh
   ```

现在你的项目应该能够在局域网环境下正常工作了！ 