# 安全订单系统集成指南

## 📚 概述

本指南说明如何在现有的 NASA Go Admin 项目中集成新的安全订单系统，解决并发、卡单、支付异常等问题。

## 🔧 集成步骤

### 1. 系统初始化

在 `main.go` 或应用启动文件中初始化安全订单系统：

```go
package main

import (
    "log"
    "nasa-go-admin/db"
    "nasa-go-admin/redis"
    "nasa-go-admin/services/app_service"
)

func main() {
    // 初始化数据库
    db.InitDB()
    
    // 初始化Redis
    redisClient := redis.GetClient()
    
    // 🚀 初始化安全订单系统
    if err := app_service.InitGlobalOrderSystem(redisClient); err != nil {
        log.Fatalf("初始化订单系统失败: %v", err)
    }
    
    // 启动Web服务
    startWebServer()
}
```

### 2. 修改订单创建接口

修改现有的订单创建控制器：

```go
// controllers/app/order.go
package app

import (
    "nasa-go-admin/inout"
    "nasa-go-admin/services/app_service"
    "nasa-go-admin/utils"
    "github.com/gin-gonic/gin"
)

// CreateOrder 创建订单 - 使用安全系统
func CreateOrder(c *gin.Context) {
    var params inout.CreateOrderReq
    if err := c.ShouldBind(&params); err != nil {
        utils.Err(c, utils.ErrCodeInvalidParams, err)
        return
    }
    
    uid := c.GetInt("uid")
    
    // 🔒 使用安全订单系统创建订单
    orderSystem := app_service.GetGlobalOrderSystem()
    if orderSystem == nil {
        utils.Err(c, utils.ErrCodeInternalError, "订单系统未初始化")
        return
    }
    
    orderNo, err := orderSystem.CreateOrderWithSystem(c, uid, params)
    if err != nil {
        utils.Err(c, utils.ErrCodeInternalError, err)
        return
    }
    
    utils.Succ(c, gin.H{
        "order_no": orderNo,
        "message":  "订单创建成功",
    })
}
```

### 3. 添加健康检查接口

添加系统健康检查接口：

```go
// controllers/admin/system.go
package admin

import (
    "nasa-go-admin/services/app_service"
    "nasa-go-admin/utils"
    "github.com/gin-gonic/gin"
)

// OrderSystemHealth 订单系统健康检查
func OrderSystemHealth(c *gin.Context) {
    health := app_service.GetOrderSystemHealth()
    
    status := health["status"].(string)
    if status == "healthy" {
        c.JSON(200, health)
    } else if status == "degraded" {
        c.JSON(200, health) // 降级状态仍返回200
    } else {
        c.JSON(503, health) // 不健康返回503
    }
}

// OrderSystemStatus 订单系统状态查询
func OrderSystemStatus(c *gin.Context) {
    status := app_service.GetOrderSystemStatus()
    utils.Succ(c, status)
}
```

### 4. 路由配置

在路由配置中添加新的接口：

```go
// router/router.go
func InitRouter() *gin.Engine {
    r := gin.Default()
    
    // App路由组
    appGroup := r.Group("/api/app")
    {
        // 使用新的安全订单创建接口
        appGroup.POST("/order/create", middleware.AuthMiddleware(), app.CreateOrder)
        // ... 其他路由
    }
    
    // Admin路由组
    adminGroup := r.Group("/api/admin")
    {
        // 系统监控接口
        adminGroup.GET("/system/order/health", admin.OrderSystemHealth)
        adminGroup.GET("/system/order/status", admin.OrderSystemStatus)
        // ... 其他路由
    }
    
    return r
}
```

### 5. 数据库表结构更新

如果需要添加版本控制字段，请执行以下SQL：

```sql
-- 商品表添加版本字段（可选）
ALTER TABLE app_goods ADD COLUMN version BIGINT DEFAULT 0 COMMENT '版本号';

-- 钱包表添加版本字段（可选）
ALTER TABLE app_wallet ADD COLUMN version BIGINT DEFAULT 0 COMMENT '版本号';

-- 创建订单超时表（可选）
CREATE TABLE order_timeouts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_no VARCHAR(50) NOT NULL COMMENT '订单号',
    expire_at DATETIME NOT NULL COMMENT '过期时间',
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_expire_at (expire_at),
    INDEX idx_order_no (order_no)
) COMMENT='订单超时表';

-- 添加必要的索引
CREATE INDEX idx_app_order_status_create_time ON app_order(status, create_time);
CREATE INDEX idx_app_order_user_status ON app_order(user_id, status);
CREATE INDEX idx_app_goods_status_stock ON app_goods(status, stock);
```

## 🛠️ 配置说明

### Redis 配置

确保 Redis 配置正确：

```yaml
# config.yaml
redis:
  addr: "localhost:6379"
  password: ""
  db: 0
  pool_size: 10
  dial_timeout: 5s
  read_timeout: 3s
  write_timeout: 3s
```

### 监控配置

可以通过环境变量配置监控参数：

```bash
# 环境变量
export ORDER_MONITOR_INTERVAL=60      # 监控检查间隔（秒）
export ORDER_TIMEOUT_MINUTES=15       # 订单超时时间（分钟）
export MAX_RETRY_ATTEMPTS=3           # 最大重试次数
export STOCK_ALERT_THRESHOLD=10       # 库存预警阈值
```

## 📊 监控和告警

### 访问监控接口

```bash
# 健康检查
curl http://localhost:8080/api/admin/system/order/health

# 系统状态
curl http://localhost:8080/api/admin/system/order/status
```

### 响应示例

健康检查响应：
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15 10:30:00",
  "components": {
    "database": {"status": "healthy"},
    "redis": {"status": "healthy"},
    "monitoring": {"status": "healthy", "running": true},
    "initialization": {"status": "healthy"}
  }
}
```

系统状态响应：
```json
{
  "initialized": true,
  "timestamp": "2024-01-15 10:30:00",
  "redis_enabled": true,
  "components": {
    "secure_creator": {"enabled": true},
    "monitoring": {
      "enabled": true,
      "running": true
    },
    "compensation": {"enabled": true},
    "monitoring_stats": {
      "orders": {
        "total_orders": 1250,
        "pending_orders": 23,
        "paid_orders": 1180,
        "cancelled_orders": 47
      },
      "payments": {
        "total_payments": 1180,
        "total_amount": 125600.50
      },
      "system": {
        "monitoring_running": true,
        "database_connected": true,
        "redis_connected": true,
        "redis_ping_ok": true
      }
    }
  }
}
```

## 🔍 故障排查

### 常见问题

1. **订单系统未初始化**
   ```
   错误：order system not initialized
   解决：确保在 main.go 中调用了 InitGlobalOrderSystem()
   ```

2. **Redis连接失败**
   ```
   错误：Redis ping失败
   解决：检查Redis服务是否启动，配置是否正确
   ```

3. **分布式锁获取失败**
   ```
   错误：系统繁忙，请稍后再试
   解决：通常是正常的并发控制，用户重试即可
   ```

4. **库存扣减失败**
   ```
   错误：库存扣减失败，请重试
   解决：检查商品库存是否充足，是否有并发冲突
   ```

### 日志分析

关键日志关键词：
- `🚀 开始安全创建订单` - 订单创建开始
- `✅ 订单创建完成` - 订单创建成功
- `🔒 成功获取分布式锁` - 锁获取成功
- `🚨 [ALERT]` - 告警信息
- `🚨🚨 [URGENT ALERT]` - 紧急告警

### 性能监控

关注指标：
- 订单创建延迟（目标：< 500ms）
- 订单成功率（目标：> 99.5%）
- 库存准确率（目标：100%）
- 挂起订单数量（告警：> 10个）

## 🔧 高级配置

### 自定义告警

实现自定义告警服务：

```go
// services/alert/custom_alert.go
package alert

import (
    "nasa-go-admin/services/app_service"
)

type CustomAlertService struct {
    // 添加邮件、短信、钉钉等告警渠道
}

func (c *CustomAlertService) SendAlert(title, message string) error {
    // 实现自定义告警逻辑
    return nil
}

func (c *CustomAlertService) SendUrgentAlert(title, message string) error {
    // 实现紧急告警逻辑
    return nil
}
```

### 自定义监控指标

扩展监控指标：

```go
// 在 OrderMonitoringService 中添加自定义检查
func (oms *OrderMonitoringService) checkCustomMetrics() error {
    // 添加业务特定的监控逻辑
    return nil
}
```

## 📈 性能优化建议

1. **数据库优化**
   - 确保已添加建议的索引
   - 定期优化查询性能
   - 考虑读写分离

2. **Redis优化**
   - 合理设置连接池大小
   - 监控Redis内存使用
   - 设置合适的过期时间

3. **应用层优化**
   - 使用连接池
   - 异步处理非关键操作
   - 合理设置超时时间

## 🔒 安全注意事项

1. **访问控制**
   - 监控接口需要适当的权限控制
   - 敏感信息不要暴露在日志中

2. **数据安全**
   - 定期备份重要数据
   - 加密敏感信息存储

3. **系统安全**
   - 定期更新依赖包
   - 监控异常访问模式

---

通过以上集成，你的订单系统将具备：
- ✅ 并发安全保护
- ✅ 分布式锁机制
- ✅ 自动超时处理
- ✅ 数据一致性保证
- ✅ 实时监控告警
- ✅ 异常自动恢复
- ✅ 完整的日志追踪 