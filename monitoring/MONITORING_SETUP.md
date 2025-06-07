# 🔍 Prometheus + Grafana 监控系统使用指南

## 📋 监控系统的实际价值

### 1. **解决的核心问题**

#### 问题场景1：性能问题快速定位
```
❌ 传统方式：
- 用户反馈："网站变慢了"
- 开发者："具体哪里慢？"
- 用户："不知道，就是慢"
- 开发者只能通过日志逐个排查，耗时长

✅ 有监控后：
- 系统自动告警："API响应时间超过1秒"
- 图表显示：/api/users 接口从100ms上升到1.2秒
- 数据库查询时间：从50ms上升到800ms
- 定位问题：某个SQL查询出现性能问题
- 解决时间：从几小时缩短到几分钟
```

#### 问题场景2：提前预防系统故障
```
❌ 传统方式：
- 系统突然崩溃
- 用户无法访问
- 紧急修复，影响业务

✅ 有监控后：
- 提前告警："数据库连接数使用率85%"
- 自动扩容或优化连接池
- 避免系统崩溃
```

#### 问题场景3：业务数据分析
```
✅ 可以实时了解：
- 每分钟用户登录数量
- 哪个API被调用最多
- 系统负载峰值时间
- 缓存命中率趋势
- 帮助优化系统架构和业务策略
```

## 🚀 快速启动监控系统

### 第一步：安装依赖
```bash
# 添加 Prometheus 依赖到 go.mod
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promauto
go get github.com/prometheus/client_golang/prometheus/promhttp
```

### 第二步：启动监控服务
```bash
# 启动 Prometheus + Grafana
docker-compose -f docker-compose-monitoring.yml up -d

# 检查服务状态
docker-compose -f docker-compose-monitoring.yml ps
```

### 第三步：修改您的应用代码
```go
// 在 main.go 中添加监控支持
package main

import (
    "nasa-go-admin/pkg/monitoring"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    // ... 现有代码 ...
    
    // 添加 Prometheus 中间件
    app.Use(monitoring.PrometheusMiddleware())
    
    // 添加指标暴露端点
    app.GET("/metrics", gin.WrapH(promhttp.Handler()))
    
    // ... 现有代码 ...
}
```

### 第四步：访问监控界面
```bash
# Prometheus (数据查询)
http://localhost:9090

# Grafana (图表展示)  
http://localhost:3000
# 默认账号: admin
# 默认密码: admin123
```

## 📊 实际使用效果

### 1. **Prometheus 界面截图说明**
访问 `http://localhost:9090` 可以看到：

```
📈 实时查询示例：
- rate(http_requests_total[1m])        # 每分钟请求量
- http_request_duration_seconds{quantile="0.95"}  # 95%响应时间
- db_connections_in_use                # 数据库连接使用情况
- redis_cache_hit_rate                 # 缓存命中率
```

### 2. **Grafana 仪表板展示**
访问 `http://localhost:3000` 可以看到：

```
📊 系统概览仪表板：
┌─────────────────┬─────────────────┬─────────────────┐
│   API 请求量     │   响应时间       │   错误率         │
│   1,234 req/min │   245ms (avg)   │   0.1%          │
└─────────────────┴─────────────────┴─────────────────┘

📊 数据库性能：
┌─────────────────┬─────────────────┬─────────────────┐
│   连接池使用率   │   查询响应时间   │   慢查询数量     │
│   45%           │   123ms (avg)   │   3 queries     │
└─────────────────┴─────────────────┴─────────────────┘

📊 系统资源：
┌─────────────────┬─────────────────┬─────────────────┐
│   CPU 使用率     │   内存使用率     │   磁盘使用率     │
│   12%           │   68%           │   34%           │
└─────────────────┴─────────────────┴─────────────────┘
```

## ⚠️ 告警示例

### 自动告警效果：
```
🚨 [CRITICAL] 2024-01-15 14:30:00
标题: API错误率过高
描述: 5xx错误率超过5%，当前值: 8.2%
建议: 立即检查应用日志和数据库状态

🚨 [WARNING] 2024-01-15 14:25:00  
标题: 数据库连接数过高
描述: 数据库连接数使用过高，当前值: 85
建议: 考虑优化连接池配置或查询优化

⚠️ [INFO] 2024-01-15 14:20:00
标题: 缓存命中率下降
描述: Redis缓存命中率低于80%，当前值: 75%
建议: 检查缓存策略和数据热度
```

## 🎯 业务价值体现

### 1. **运维效率提升**
- **问题发现时间**: 从用户反馈到自动检测（时间缩短90%）
- **问题定位时间**: 从小时级到分钟级（效率提升20倍）
- **预防性维护**: 提前发现问题，避免系统故障

### 2. **业务决策支持**
```go
// 实际业务数据示例
业务高峰期: 每天 19:00-22:00 (用户活跃度最高)
热门功能: /api/goods/list (调用量占30%)
性能瓶颈: 用户权限查询 (平均200ms)
优化效果: 缓存优化后响应时间降低60%
```

### 3. **成本控制**
- **资源使用优化**: 了解实际资源需求，避免过度配置
- **扩容时机**: 基于数据决策何时扩容
- **故障成本**: 减少因故障导致的业务损失

## 🔧 实际操作步骤

### 立即可以尝试：

1. **启动监控服务**：
```bash
docker-compose -f docker-compose-monitoring.yml up -d
```

2. **修改 go.mod 添加依赖**：
```bash
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
```

3. **在应用中集成监控** (修改您的 main.go)：
```go
// 添加这几行即可
import "github.com/prometheus/client_golang/prometheus/promhttp"

// 在路由中添加
app.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

4. **访问监控界面**：
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin123)

### 5分钟内就能看到效果！

启动后，您就能实时看到：
- 每个API的调用次数和响应时间
- 系统CPU、内存使用情况  
- 数据库连接池状态
- 任何异常都会自动告警

这就是 Prometheus + Grafana 的核心价值 - **让系统状态可视化，问题可预测，运维更智能**！ 