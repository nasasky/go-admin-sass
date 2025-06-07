# 📊 MongoDB 监控系统使用指南

## 🎯 功能概述

您的项目现在已经集成了完整的MongoDB监控系统，可以：

- ✅ **自动收集**：HTTP请求、用户行为、数据库连接等指标
- ✅ **实时存储**：所有监控数据存储到MongoDB数据库
- ✅ **API查询**：通过REST API查询监控数据
- ✅ **性能分析**：分析API性能、用户活动等

## 🚀 快速开始

### 1. 启动应用
```bash
# 启动您的应用
go run main.go
```

### 2. 访问监控API
应用启动后，监控数据会自动收集并存储到MongoDB中。

## 📈 监控API接口

### 系统概览
```bash
GET /api/admin/monitoring/overview?time_range=1h

# 返回示例：
{
  "code": 200,
  "data": {
    "timeRange": "1h",
    "timestamp": "2024-01-15T10:30:00Z",
    "stats": {
      "http_requests": 156,
      "user_logins": 23,
      "user_registers": 5,
      "db_connections": 12,
      "db_max_conns": 100
    }
  },
  "message": "success"
}
```

### HTTP请求指标
```bash
GET /api/admin/monitoring/http-metrics?limit=50

# 返回示例：
{
  "code": 200,
  "data": {
    "metrics": [
      {
        "timestamp": "2024-01-15T10:29:45Z",
        "method": "GET",
        "endpoint": "/api/admin/users",
        "status_code": 200,
        "duration": 0.125,
        "client_ip": "127.0.0.1",
        "user_id": "admin"
      }
    ],
    "total": 50
  },
  "message": "success"
}
```

### 实时统计
```bash
GET /api/admin/monitoring/realtime?time_range=1h

# 返回示例：
{
  "code": 200,
  "data": {
    "timestamp": "2024-01-15T10:30:00Z",
    "time_range": "1h",
    "http_requests": 156,
    "user_logins": 23,
    "user_registers": 5,
    "db_connections": 12,
    "system_status": "running",
    "uptime": "运行中"
  },
  "message": "success"
}
```

### 健康检查
```bash
GET /api/admin/monitoring/health

# 返回示例：
{
  "code": 200,
  "data": {
    "status": "healthy",
    "timestamp": "2024-01-15 10:30:00",
    "services": {
      "database": "connected",
      "mongodb": "connected",
      "redis": "connected"
    }
  },
  "message": "success"
}
```

## 📊 监控数据类型

### 1. HTTP请求指标
自动记录每个API请求：
- 请求时间
- 请求方法和端点
- 响应状态码
- 响应时间
- 客户端IP
- 用户ID（如果已登录）

### 2. 业务指标
自动记录业务事件：
- 用户登录 (`user_login`)
- 用户注册 (`user_register`)
- 其他自定义业务事件

### 3. 数据库指标
定期记录数据库状态：
- 当前连接数
- 空闲连接数
- 最大连接数

## 🔍 MongoDB数据查询

### 直接查询MongoDB
```javascript
// 连接到MongoDB
use admin_log_db

// 查看最近的HTTP请求
db.logs.find({"method": {$exists: true}}).sort({"timestamp": -1}).limit(10)

// 查看用户登录记录
db.logs.find({"metric_type": "user_login"}).sort({"timestamp": -1}).limit(10)

// 查看数据库连接状态
db.logs.find({"connections_in_use": {$exists: true}}).sort({"timestamp": -1}).limit(5)

// 统计最近1小时的请求数
db.logs.count({
  "method": {$exists: true},
  "timestamp": {
    $gte: new Date(Date.now() - 60*60*1000)
  }
})

// 按端点统计请求数
db.logs.aggregate([
  {$match: {"method": {$exists: true}}},
  {$group: {
    "_id": "$endpoint",
    "count": {$sum: 1},
    "avg_duration": {$avg: "$duration"}
  }},
  {$sort: {"count": -1}}
])
```

## 📱 前端集成示例

### JavaScript 获取监控数据
```javascript
// 获取系统概览
async function getSystemOverview() {
  const response = await fetch('/api/admin/monitoring/overview?time_range=1h');
  const data = await response.json();
  console.log('系统概览:', data.data);
}

// 获取实时统计
async function getRealTimeStats() {
  const response = await fetch('/api/admin/monitoring/realtime');
  const data = await response.json();
  
  // 更新页面显示
  document.getElementById('http-requests').textContent = data.data.http_requests;
  document.getElementById('user-logins').textContent = data.data.user_logins;
  document.getElementById('db-connections').textContent = data.data.db_connections;
}

// 定时刷新数据
setInterval(getRealTimeStats, 30000); // 每30秒刷新一次
```

### Vue.js 组件示例
```vue
<template>
  <div class="monitoring-dashboard">
    <div class="stats-grid">
      <div class="stat-card">
        <h3>HTTP请求</h3>
        <p class="stat-value">{{ stats.http_requests }}</p>
      </div>
      <div class="stat-card">
        <h3>用户登录</h3>
        <p class="stat-value">{{ stats.user_logins }}</p>
      </div>
      <div class="stat-card">
        <h3>数据库连接</h3>
        <p class="stat-value">{{ stats.db_connections }}</p>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      stats: {
        http_requests: 0,
        user_logins: 0,
        db_connections: 0
      }
    }
  },
  mounted() {
    this.loadStats();
    setInterval(this.loadStats, 30000);
  },
  methods: {
    async loadStats() {
      try {
        const response = await fetch('/api/admin/monitoring/realtime');
        const data = await response.json();
        this.stats = data.data;
      } catch (error) {
        console.error('加载监控数据失败:', error);
      }
    }
  }
}
</script>
```

## 🛠️ 自定义监控

### 添加自定义业务指标
```go
// 在您的业务代码中
import "nasa-go-admin/pkg/monitoring"

// 记录订单创建
monitoring.SaveBusinessMetric("order_create", userID)

// 记录商品查看
monitoring.SaveBusinessMetric("product_view", productID)

// 记录支付成功
monitoring.SaveBusinessMetric("payment_success", orderID)
```

### 添加自定义HTTP指标
```go
// 在中间件中记录额外信息
func CustomMonitoringMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        
        duration := time.Since(start).Seconds()
        
        // 自定义监控逻辑
        if duration > 2.0 {
            log.Printf("慢请求警告: %s %s 耗时 %.2f秒", 
                c.Request.Method, c.FullPath(), duration)
        }
    }
}
```

## 📊 数据分析示例

### 性能分析
```javascript
// 分析最慢的API端点
db.logs.aggregate([
  {$match: {"method": {$exists: true}}},
  {$group: {
    "_id": "$endpoint",
    "avg_duration": {$avg: "$duration"},
    "max_duration": {$max: "$duration"},
    "count": {$sum: 1}
  }},
  {$sort: {"avg_duration": -1}},
  {$limit: 10}
])
```

### 用户活动分析
```javascript
// 分析用户登录时间分布
db.logs.aggregate([
  {$match: {"metric_type": "user_login"}},
  {$group: {
    "_id": {$hour: "$timestamp"},
    "count": {$sum: 1}
  }},
  {$sort: {"_id": 1}}
])
```

## 🎯 实际应用价值

### 1. **性能优化**
- 识别慢API接口
- 分析数据库连接使用情况
- 监控系统负载

### 2. **业务分析**
- 用户活跃时间分析
- 功能使用频率统计
- 用户行为路径分析

### 3. **运维监控**
- 系统健康状态监控
- 错误率统计
- 资源使用情况

### 4. **数据驱动决策**
- 基于真实数据优化产品
- 了解用户使用习惯
- 制定运营策略

现在您的监控数据都存储在MongoDB中，可以通过API或直接查询数据库来分析系统性能和用户行为！ 