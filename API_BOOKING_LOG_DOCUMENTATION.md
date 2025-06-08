# 订单状态日志管理 API 文档

## 概述

本文档描述了房间预订系统中订单状态日志管理的API接口。系统会自动将所有订单状态变更、错误信息和管理员操作记录到MongoDB数据库中，并提供查询和统计接口。

## 日志类型

### 系统自动记录的日志类型

| 日志类型 | 类型代码 | 描述 | 触发条件 |
|---------|---------|------|---------|
| 调度器启动 | `scheduler_start` | 订单状态自动管理调度器启动 | 系统启动时 |
| 订单激活 | `booking_activate` | 订单从已支付变为使用中 | 到达预约开始时间 |
| 订单完成 | `booking_complete` | 订单从使用中变为已完成 | 到达预约结束时间 |
| 订单超时取消 | `booking_timeout` | 超过24小时未支付的订单被取消 | 定时检查发现超时订单 |
| 订单状态更新失败 | `booking_error` | 订单状态更新操作失败 | 数据库操作异常 |
| 房间状态更新失败 | `room_error` | 房间状态更新操作失败 | 数据库操作异常 |
| 手动开始订单 | `manual_start` | 管理员手动开始订单 | 管理员操作 |
| 手动结束订单 | `manual_end` | 管理员手动结束订单 | 管理员操作 |
| 使用记录错误 | `usage_log_error` | 使用记录操作失败 | 数据库操作异常 |

## API 接口

### 1. 获取订单状态日志列表

**接口地址**: `GET /api/admin/bookings/logs`

**请求参数**:

| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| page | int | 是 | 页码，从1开始 | 1 |
| page_size | int | 是 | 每页数量，最大100 | 20 |
| log_type | string | 否 | 日志类型过滤 | booking_activate |
| booking_id | int | 否 | 订单ID过滤 | 123 |
| booking_no | string | 否 | 订单号过滤（模糊匹配） | BK202312 |
| room_id | int | 否 | 房间ID过滤 | 101 |
| user_id | int | 否 | 用户ID过滤 | 1001 |
| start_date | string | 否 | 开始日期 (YYYY-MM-DD) | 2023-12-01 |
| end_date | string | 否 | 结束日期 (YYYY-MM-DD) | 2023-12-31 |

**请求示例**:
```bash
GET /api/admin/bookings/logs?page=1&page_size=20&log_type=booking_activate&start_date=2023-12-01&end_date=2023-12-31
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "total": 150,
    "page": 1,
    "page_size": 20,
    "list": [
      {
        "id": "6570a1b2c8f4d5e6f7890123",
        "log_type": "booking_activate",
        "log_type_text": "订单激活",
        "booking_id": 123,
        "booking_no": "BK202312011530001",
        "room_id": 101,
        "room_name": "豪华包厢001",
        "user_id": 1001,
        "old_status": 2,
        "old_status_text": "已支付",
        "new_status": 3,
        "new_status_text": "使用中",
        "message": "订单已激活: BK202312011530001 (房间ID: 101, 用户ID: 1001)",
        "error_msg": "",
        "details": {
          "start_time": "2023-12-01T15:30:00Z",
          "end_time": "2023-12-01T18:30:00Z",
          "hours": 3,
          "amount": 450.00
        },
        "created_at": "2023-12-01T15:30:01Z",
        "server_info": {
          "hostname": "server-01",
          "pid": 12345,
          "mode": "admin"
        }
      }
    ]
  }
}
```

### 2. 获取订单状态日志统计

**接口地址**: `GET /api/admin/bookings/logs/statistics`

**请求参数**:

| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| start_date | string | 否 | 开始日期 (YYYY-MM-DD) | 2023-12-01 |
| end_date | string | 否 | 结束日期 (YYYY-MM-DD) | 2023-12-31 |

**请求示例**:
```bash
GET /api/admin/bookings/logs/statistics?start_date=2023-12-01&end_date=2023-12-31
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "total_logs": 1250,
    "error_logs": 15,
    "type_stats": {
      "scheduler_start": 5,
      "booking_activate": 320,
      "booking_complete": 315,
      "booking_timeout": 25,
      "booking_error": 8,
      "room_error": 4,
      "manual_start": 12,
      "manual_end": 8,
      "usage_log_error": 3
    }
  }
}
```

### 3. 获取订单状态信息

**接口地址**: `GET /api/admin/bookings/status-info`

**请求参数**:

| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| id | int | 是 | 订单ID | 123 |

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "booking_id": 123,
    "booking_no": "BK202312011530001",
    "status": 2,
    "status_text": "已支付",
    "start_time": "2023-12-01T15:30:00Z",
    "end_time": "2023-12-01T18:30:00Z",
    "room_id": 101,
    "room_name": "豪华包厢001",
    "room_status": 1,
    "room_status_text": "可用",
    "current_time": "2023-12-01T15:35:00Z",
    "can_start": true,
    "can_end": false,
    "should_auto_start": true,
    "should_auto_end": false
  }
}
```

### 4. 手动开始订单

**接口地址**: `POST /api/admin/bookings/manual-start`

**请求参数**:
```json
{
  "id": 123
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "message": "订单已手动开始"
  }
}
```

### 5. 手动结束订单

**接口地址**: `POST /api/admin/bookings/manual-end`

**请求参数**:
```json
{
  "id": 123
}
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "message": "订单已手动结束"
  }
}
```

## 日志数据结构

### 日志记录字段说明

| 字段名 | 类型 | 描述 | 示例 |
|--------|------|------|------|
| id | string | MongoDB文档ID | "6570a1b2c8f4d5e6f7890123" |
| log_type | string | 日志类型代码 | "booking_activate" |
| log_type_text | string | 日志类型中文描述 | "订单激活" |
| booking_id | int | 订单ID（可为空） | 123 |
| booking_no | string | 订单号 | "BK202312011530001" |
| room_id | int | 房间ID（可为空） | 101 |
| room_name | string | 房间名称 | "豪华包厢001" |
| user_id | int | 用户ID（可为空） | 1001 |
| old_status | int | 原状态（可为空） | 2 |
| old_status_text | string | 原状态文本（可为空） | "已支付" |
| new_status | int | 新状态（可为空） | 3 |
| new_status_text | string | 新状态文本（可为空） | "使用中" |
| message | string | 操作信息 | "订单已激活: BK202312011530001" |
| error_msg | string | 错误信息（仅错误日志） | "数据库连接失败" |
| details | object | 详细信息（JSON格式） | 见下方示例 |
| created_at | datetime | 创建时间 | "2023-12-01T15:30:01Z" |
| server_info | object | 服务器信息 | 见下方示例 |

### Details 字段示例

**订单激活/完成日志**:
```json
{
  "start_time": "2023-12-01T15:30:00Z",
  "end_time": "2023-12-01T18:30:00Z",
  "hours": 3,
  "amount": 450.00,
  "planned_hours": 3,
  "actual_hours": 2.8
}
```

**手动操作日志**:
```json
{
  "admin_id": 1,
  "start_time": "2023-12-01T15:30:00Z",
  "end_time": "2023-12-01T18:30:00Z",
  "hours": 3,
  "amount": 450.00,
  "actual_hours": 2.5
}
```

**错误日志**:
```json
{
  "operation": "激活订单",
  "error": "database connection timeout"
}
```

**超时取消日志**:
```json
{
  "start_time": "2023-12-01T15:30:00Z",
  "end_time": "2023-12-01T18:30:00Z",
  "hours": 3,
  "amount": 450.00,
  "created_at": "2023-11-30T15:30:00Z",
  "timeout_hours": 24
}
```

### Server Info 字段示例

```json
{
  "hostname": "server-01",
  "pid": 12345,
  "mode": "admin"
}
```

## 状态码说明

### 订单状态

| 状态码 | 状态名称 | 描述 |
|--------|---------|------|
| 1 | 待支付 | 订单已创建，等待支付 |
| 2 | 已支付 | 订单已支付，等待使用 |
| 3 | 使用中 | 订单正在使用中 |
| 4 | 已完成 | 订单已完成 |
| 5 | 已取消 | 订单已取消 |
| 6 | 已退款 | 订单已退款 |

### 房间状态

| 状态码 | 状态名称 | 描述 |
|--------|---------|------|
| 1 | 可用 | 房间可预订 |
| 2 | 使用中 | 房间正在使用 |
| 3 | 维护中 | 房间维护中 |
| 4 | 停用 | 房间已停用 |

## 使用场景

### 1. 监控订单状态变更

管理员可以通过日志查询接口监控所有订单的状态变更情况：

```bash
# 查询今天的所有订单激活记录
GET /api/admin/bookings/logs?log_type=booking_activate&start_date=2023-12-01&end_date=2023-12-01

# 查询特定订单的所有日志
GET /api/admin/bookings/logs?booking_id=123
```

### 2. 错误排查

当系统出现问题时，可以查询错误日志进行排查：

```bash
# 查询所有错误日志
GET /api/admin/bookings/logs?log_type=booking_error

# 查询房间状态更新错误
GET /api/admin/bookings/logs?log_type=room_error
```

### 3. 统计分析

通过统计接口分析系统运行情况：

```bash
# 查询本月的日志统计
GET /api/admin/bookings/logs/statistics?start_date=2023-12-01&end_date=2023-12-31
```

### 4. 审计管理员操作

查询管理员的手动操作记录：

```bash
# 查询所有手动开始订单的记录
GET /api/admin/bookings/logs?log_type=manual_start

# 查询所有手动结束订单的记录
GET /api/admin/bookings/logs?log_type=manual_end
```

## 注意事项

1. **权限控制**: 所有日志查询接口都需要管理员权限
2. **数据保留**: 建议定期清理过期的日志数据，避免数据库过大
3. **性能考虑**: 大量日志查询时建议使用分页和时间范围过滤
4. **时区处理**: 所有时间字段都使用UTC时间，前端需要转换为本地时间
5. **MongoDB配置**: 确保MongoDB配置正确，集合名称为 `booking_status_logs`

## 错误码

| 错误码 | 描述 | 解决方案 |
|--------|------|---------|
| 20001 | 参数错误 | 检查请求参数格式和必填项 |
| 10002 | 用户未登录 | 检查JWT Token是否有效 |
| 50001 | MongoDB连接失败 | 检查MongoDB服务状态 |
| 50002 | 数据查询失败 | 检查查询条件和数据库状态 | 