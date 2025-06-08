# 房间包厢系统API文档

## 概述

房间包厢系统提供了完整的房间管理和预订功能，包括房间信息管理、预订管理、使用记录等功能。

## 基础信息

- **基础URL**: `/api/app` (用户端) / `/api/admin` (管理端)
- **认证方式**: JWT Token (预订相关接口需要，房间查看无需认证)
- **数据格式**: JSON
- **字符编码**: UTF-8

## 用户端API

### 房间管理

> **注意**: 以下房间查看相关接口无需用户登录即可访问

#### 1. 获取房间列表

**接口地址**: `GET /api/app/rooms`

**请求参数**:
```json
{
  "page": 1,           // 页码，默认1
  "page_size": 10,     // 每页数量，默认10，最大100
  "room_type": "small", // 房间类型：small/medium/large/luxury
  "status": 1,         // 房间状态：1可用,2使用中,3维护中,4停用
  "floor": 1,          // 楼层
  "keyword": "雅致",    // 关键词搜索（房间名称/号码）
  "min_price": 50.0,   // 最低价格
  "max_price": 500.0,  // 最高价格
  "min_capacity": 2,   // 最小容量
  "max_capacity": 20   // 最大容量
}
```

**响应示例**:
```json
{
  "code": 20000,
  "message": "success",
  "data": {
    "total": 4,
    "page": 1,
    "page_size": 10,
    "list": [
      {
        "id": 1,
        "room_number": "A001",
        "room_name": "雅致小包厢",
        "room_type": "small",
        "room_type_text": "小包厢",
        "capacity": 4,
        "hourly_rate": 88.00,
        "features": ["KTV设备", "茶水服务", "WIFI"],
        "images": ["https://example.com/room1.jpg"],
        "status": 1,
        "status_text": "可用",
        "floor": 1,
        "area": 20.50,
        "description": "温馨舒适的小包厢，适合小型聚会",
        "create_time": "2024-01-01T10:00:00Z",
        "update_time": "2024-01-01T10:00:00Z",
        "is_available": true,
        "current_booking": null
      }
    ]
  }
}
```

#### 2. 获取房间详情

**接口地址**: `GET /api/app/rooms/{id}`

**路径参数**:
- `id`: 房间ID

**响应示例**:
```json
{
  "code": 20000,
  "message": "success",
  "data": {
    "id": 1,
    "room_number": "A001",
    "room_name": "雅致小包厢",
    "room_type": "small",
    "room_type_text": "小包厢",
    "capacity": 4,
    "hourly_rate": 88.00,
    "features": ["KTV设备", "茶水服务", "WIFI"],
    "images": ["https://example.com/room1.jpg"],
    "status": 1,
    "status_text": "可用",
    "floor": 1,
    "area": 20.50,
    "description": "温馨舒适的小包厢，适合小型聚会",
    "create_time": "2024-01-01T10:00:00Z",
    "update_time": "2024-01-01T10:00:00Z",
    "is_available": true,
    "current_booking": {
      "id": 123,
      "booking_no": "BK20240101120000123456",
      "start_time": "2024-01-01T14:00:00Z",
      "end_time": "2024-01-01T18:00:00Z",
      "status": 3,
      "status_text": "使用中"
    }
  }
}
```

#### 3. 检查房间可用性

**接口地址**: `POST /api/app/rooms/check-availability`

**请求参数**:
```json
{
  "room_id": 1,
  "start_time": "2024-01-01 14:00:00",
  "hours": 4
}
```

**响应示例**:
```json
{
  "code": 20000,
  "message": "success",
  "data": {
    "is_available": true,
    "message": "房间可预订",
    "total_amount": 352.00
  }
}
```

#### 4. 获取房间统计信息

**接口地址**: `GET /api/app/rooms/statistics`

**响应示例**:
```json
{
  "code": 20000,
  "message": "success",
  "data": {
    "total_rooms": 10,
    "available_rooms": 8,
    "occupied_rooms": 2,
    "maintenance_rooms": 0,
    "disabled_rooms": 0,
    "occupancy_rate": 20.0,
    "today_bookings": 5,
    "today_revenue": 1500.00,
    "monthly_revenue": 45000.00
  }
}
```

### 预订管理

> **注意**: 以下预订相关接口需要用户登录

#### 1. 创建预订

**接口地址**: `POST /api/app/bookings`

**请求参数**:
```json
{
  "room_id": 1,
  "start_time": "2024-01-01 14:00:00",
  "hours": 4,
  "contact_name": "张三",
  "contact_phone": "13800138000",
  "remarks": "生日聚会"
}
```

**响应示例**:
```json
{
  "code": 20000,
  "message": "success",
  "data": {
    "id": 123,
    "booking_no": "BK20240101120000123456",
    "room_id": 1,
    "user_id": 1001,
    "start_time": "2024-01-01T14:00:00Z",
    "end_time": "2024-01-01T18:00:00Z",
    "hours": 4,
    "total_amount": 352.00,
    "paid_amount": 0.00,
    "status": 1,
    "status_text": "待支付",
    "contact_name": "张三",
    "contact_phone": "13800138000",
    "remarks": "生日聚会",
    "create_time": "2024-01-01T12:00:00Z"
  }
}
```

#### 2. 获取我的预订列表

**接口地址**: `GET /api/app/bookings`

**请求参数**:
```json
{
  "page": 1,
  "page_size": 10,
  "status": 1,         // 预订状态
  "start_date": "2024-01-01",
  "end_date": "2024-01-31",
  "keyword": "张三"    // 关键词搜索
}
```

**响应示例**:
```json
{
  "code": 20000,
  "message": "success",
  "data": {
    "total": 5,
    "page": 1,
    "page_size": 10,
    "list": [
      {
        "id": 123,
        "booking_no": "BK20240101120000123456",
        "room_id": 1,
        "user_id": 1001,
        "start_time": "2024-01-01T14:00:00Z",
        "end_time": "2024-01-01T18:00:00Z",
        "hours": 4,
        "total_amount": 352.00,
        "paid_amount": 352.00,
        "status": 2,
        "status_text": "已支付",
        "contact_name": "张三",
        "contact_phone": "13800138000",
        "remarks": "生日聚会",
        "create_time": "2024-01-01T12:00:00Z",
        "room": {
          "id": 1,
          "room_number": "A001",
          "room_name": "雅致小包厢",
          "room_type": "small",
          "room_type_text": "小包厢"
        }
      }
    ]
  }
}
```

#### 3. 获取预订详情

**接口地址**: `GET /api/app/bookings/{id}`

**路径参数**:
- `id`: 预订ID

**响应示例**: 同上单个预订对象

#### 4. 取消预订

**接口地址**: `POST /api/app/bookings/cancel`

**请求参数**:
```json
{
  "id": 123,
  "reason": "临时有事，无法到场"
}
```

**响应示例**:
```json
{
  "code": 20000,
  "message": "success",
  "data": {
    "message": "预订已取消"
  }
}
```

## 管理端API

### 房间管理

#### 1. 创建房间

**接口地址**: `POST /api/admin/rooms`

**请求参数**:
```json
{
  "room_number": "A003",
  "room_name": "豪华套房",
  "room_type": "luxury",
  "capacity": 15,
  "hourly_rate": 588.00,
  "features": "[\"KTV设备\", \"茶水服务\", \"WIFI\"]",
  "images": "[\"https://example.com/room.jpg\"]",
  "floor": 3,
  "area": 60.50,
  "description": "豪华套房，设施齐全"
}
```

#### 2. 更新房间信息

**接口地址**: `PUT /api/admin/rooms`

**请求参数**:
```json
{
  "id": 1,
  "room_number": "A001",
  "room_name": "雅致小包厢（升级版）",
  "room_type": "small",
  "capacity": 6,
  "hourly_rate": 98.00,
  "features": "[\"KTV设备\", \"茶水服务\", \"WIFI\", \"投影设备\"]",
  "images": "[\"https://example.com/room1.jpg\"]",
  "floor": 1,
  "area": 25.50,
  "description": "升级后的小包厢",
  "status": 1
}
```

#### 3. 获取房间列表（管理端）

**接口地址**: `GET /api/admin/rooms`

参数和响应格式同用户端，但包含更多管理信息。

#### 4. 更新房间状态

**接口地址**: `PUT /api/admin/rooms/status`

**请求参数**:
```json
{
  "id": 1,
  "status": 3  // 1:可用,2:使用中,3:维护中,4:停用
}
```

#### 5. 删除房间

**接口地址**: `DELETE /api/admin/rooms/{id}`

**路径参数**:
- `id`: 房间ID

### 预订管理

#### 1. 获取预订列表（管理端）

**接口地址**: `GET /api/admin/bookings`

参数同用户端，但可查看所有用户的预订。

#### 2. 更新预订状态

**接口地址**: `PUT /api/admin/bookings/status`

**请求参数**:
```json
{
  "id": 123,
  "status": 4  // 预订状态
}
```

#### 3. 获取统计信息

**接口地址**: `GET /api/admin/rooms/statistics`

响应格式同用户端。

## 状态码说明

### 房间状态
- `1`: 可用
- `2`: 使用中
- `3`: 维护中
- `4`: 停用

### 预订状态
- `1`: 待支付
- `2`: 已支付
- `3`: 使用中
- `4`: 已完成
- `5`: 已取消
- `6`: 已退款

### 房间类型
- `small`: 小包厢
- `medium`: 中包厢
- `large`: 大包厢
- `luxury`: 豪华包厢

## 错误码说明

- `20000`: 成功
- `20001`: 业务逻辑错误
- `10002`: 用户未登录或token无效
- `10001`: 参数错误

## 使用示例

### 完整预订流程

1. **查看房间列表（无需登录）**
```bash
curl -X GET "http://localhost:8080/api/app/rooms?page=1&page_size=10"
```

2. **检查房间可用性（无需登录）**
```bash
curl -X POST "http://localhost:8080/api/app/rooms/check-availability" \
  -H "Content-Type: application/json" \
  -d '{
    "room_id": 1,
    "start_time": "2024-01-01 14:00:00",
    "hours": 4
  }'
```

3. **创建预订（需要登录）**
```bash
curl -X POST "http://localhost:8080/api/app/bookings" \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "room_id": 1,
    "start_time": "2024-01-01 14:00:00",
    "hours": 4,
    "contact_name": "张三",
    "contact_phone": "13800138000",
    "remarks": "生日聚会"
  }'
```

4. **查看我的预订（需要登录）**
```bash
curl -X GET "http://localhost:8080/api/app/bookings?page=1&page_size=10" \
  -H "Authorization: Bearer {token}"
```

## 注意事项

1. **时间格式**: 所有时间使用 `YYYY-MM-DD HH:mm:ss` 格式
2. **价格精度**: 价格保留两位小数
3. **并发控制**: 预订时会检查时间冲突，确保房间不会被重复预订
4. **权限控制**: 用户只能查看和操作自己的预订，管理员可以操作所有预订
5. **数据验证**: 所有输入数据都会进行格式和业务逻辑验证

## 数据库设计

### 表结构说明

- `rooms`: 房间信息表
- `room_bookings`: 房间预订表  
- `room_usage_logs`: 房间使用记录表

### 索引优化

- 房间查询：`(room_type, status, capacity)`
- 预订时间查询：`(room_id, start_time, end_time)`
- 用户预订查询：`(user_id, status, create_time)`

## 性能建议

1. **分页查询**: 建议每页不超过100条记录
2. **缓存策略**: 房间信息可以缓存，预订信息实时查询
3. **索引使用**: 利用复合索引提高查询性能
4. **连接池**: 合理配置数据库连接池大小 