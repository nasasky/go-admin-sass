# 数据库设置指南

## 快速导入数据库表

### 方法一：使用自动导入脚本（推荐）

```bash
# 1. 给脚本执行权限
chmod +x scripts/import_tables.sh

# 2. 运行导入脚本
./scripts/import_tables.sh
```

脚本会自动：
- 检查数据库连接
- 创建数据库（如果不存在）
- 导入所有必要的表
- 插入测试数据
- 验证导入结果

### 方法二：手动导入SQL文件

```bash
# 1. 登录MySQL
mysql -u root -p

# 2. 创建数据库（如果不存在）
CREATE DATABASE naive_admin CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 3. 使用数据库
USE naive_admin;

# 4. 导入SQL文件
SOURCE migrations/create_order_tables.sql;
```

## 数据库表结构

导入后将创建以下核心表：

### 1. 订单表 (app_order)
- 存储所有订单信息
- 包含订单状态、金额、用户等信息
- 已优化索引提升查询性能

### 2. 钱包交易记录表 (app_recharge)  
- 记录所有钱包相关交易
- 包含充值、支付、退款等操作
- 支持交易溯源和对账

### 3. 商品表 (app_goods)
- 商品基本信息和库存
- 支持库存管控
- 分类和状态管理

### 4. 用户钱包表 (app_user_wallet)
- 用户余额信息
- 包含乐观锁版本控制
- 支持余额冻结功能

### 5. 用户表 (app_user)
- 用户基本信息
- 支持微信登录

## 测试数据

导入完成后包含以下测试数据：

### 测试用户
- 用户ID: 1, 余额: 1000.00元
- 用户ID: 2, 余额: 500.00元

### 测试商品
- 商品ID: 1, 价格: 99.99元, 库存: 100
- 商品ID: 2, 价格: 199.99元, 库存: 50  
- 商品ID: 3, 价格: 299.99元, 库存: 30

## 验证导入成功

### 1. 启动应用
```bash
go build -o nasa-go-admin .
./nasa-go-admin
```

### 2. 测试API端点
```bash
# 健康检查
curl http://localhost:8801/health

# 订单系统健康检查
curl http://localhost:8801/api/app/order/health

# 队列日志（应该能看到超时队列）
curl http://localhost:8801/api/admin/queue/log
```

### 3. 查看队列状态
访问管理后台的队列日志页面，现在应该能正确显示 `order_timeouts` 队列中的订单。

## 常见问题

### Q: 队列日志为空？
A: 确保：
1. 数据库表已正确导入
2. Redis服务正在运行
3. 创建了测试订单后查看

### Q: 数据库连接失败？
A: 检查：
1. MySQL服务是否运行
2. 数据库连接配置是否正确
3. 用户权限是否足够

### Q: Redis连接问题？
A: 确认：
1. Redis服务状态
2. 端口配置（默认6379）
3. 密码配置（如果有）

## 性能优化

导入的SQL已包含必要的索引优化：
- 订单查询索引
- 用户相关复合索引  
- 时间范围查询索引
- 状态查询索引

定期运行以下命令优化性能：
```sql
ANALYZE TABLE app_order;
ANALYZE TABLE app_recharge;
OPTIMIZE TABLE app_order;
OPTIMIZE TABLE app_recharge;
```

## 监控建议

1. 定期检查队列积压情况
2. 监控订单创建成功率
3. 关注异常支付模式
4. 定期备份重要数据 