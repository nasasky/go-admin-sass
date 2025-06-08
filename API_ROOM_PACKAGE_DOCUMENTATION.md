# 房间套餐管理 API 文档

## 概述

房间套餐管理系统支持灵活的动态定价策略，可以根据不同的时间段、日期类型（工作日/周末/节假日）、特殊活动日等设置不同的价格规则。

## 核心功能

### 1. 套餐管理
- 创建、更新、删除房间套餐
- 设置套餐优先级和有效期
- 支持多套餐并存

### 2. 定价规则
- **日期类型定价**：工作日、周末、节假日、特殊日期
- **时间段定价**：白天、晚上等不同时段
- **价格类型**：固定价格、倍数调整、加价调整
- **优先级系统**：多规则按优先级应用

### 3. 特殊日期管理
- 法定节假日配置
- 传统节日配置
- 特殊活动日配置

---

## API 接口

### 基础信息

- **Base URL**: `/api/admin`
- **认证方式**: JWT Token
- **Content-Type**: `application/json`

---

## 1. 套餐管理

### 1.1 创建套餐

**接口地址**：`POST /rooms/packages`

**请求参数**：
```json
{
  "room_id": 1,
  "package_name": "工作日优惠套餐",
  "description": "工作日时段享受优惠价格",
  "priority": 10,
  "start_date": "2024-01-01",
  "end_date": "2024-12-31"
}
```

**字段说明**：
- `room_id`: 房间ID（必填）
- `package_name`: 套餐名称（必填）
- `description`: 套餐描述（可选）
- `priority`: 优先级，数字越大优先级越高（默认0）
- `start_date`: 生效开始日期（可选）
- `end_date`: 生效结束日期（可选）

**响应示例**：
```json
{
  "code": 200,
  "message": "OK",
  "data": {
    "data": {
      "id": 1,
      "room_id": 1,
      "package_name": "工作日优惠套餐",
      "description": "工作日时段享受优惠价格",
      "priority": 10,
      "is_active": true,
      "start_date": "2024-01-01T00:00:00Z",
      "end_date": "2024-12-31T00:00:00Z",
      "create_time": "2024-01-01T12:00:00Z",
      "update_time": "2024-01-01T12:00:00Z"
    },
    "message": "创建套餐成功"
  }
}
```

### 1.2 更新套餐

**接口地址**：`PUT /rooms/packages`

**请求参数**：
```json
{
  "id": 1,
  "room_id": 1,
  "package_name": "工作日优惠套餐（修改）",
  "description": "工作日时段享受更多优惠",
  "priority": 15,
  "start_date": "2024-01-01",
  "end_date": "2024-12-31",
  "is_active": true
}
```

### 1.3 获取套餐列表

**接口地址**：`GET /rooms/packages`

**查询参数**：
- `page`: 页码（默认1）
- `page_size`: 每页数量（默认10）
- `room_id`: 房间ID过滤（可选）

**响应示例**：
```json
{
  "code": 200,
  "message": "OK",
  "data": {
    "total": 10,
    "page": 1,
    "page_size": 10,
    "list": [
      {
        "id": 1,
        "room_id": 1,
        "package_name": "工作日优惠套餐",
        "description": "工作日时段享受优惠价格",
        "priority": 10,
        "is_active": true,
        "start_date": "2024-01-01T00:00:00Z",
        "end_date": "2024-12-31T00:00:00Z",
        "rules": [
          {
            "id": 1,
            "rule_name": "工作日白天优惠",
            "day_type": "weekday",
            "day_type_text": "工作日",
            "time_start": "09:00",
            "time_end": "18:00",
            "price_type": "multiply",
            "price_type_text": "倍数调整",
            "price_value": 0.8,
            "is_active": true
          }
        ]
      }
    ]
  }
}
```

### 1.4 删除套餐

**接口地址**：`DELETE /rooms/packages/{id}`

**路径参数**：
- `id`: 套餐ID

---

## 2. 套餐规则管理

### 2.1 创建套餐规则

**接口地址**：`POST /rooms/package-rules`

**请求参数**：
```json
{
  "package_id": 1,
  "rule_name": "工作日白天优惠",
  "description": "工作日白天时段8折优惠",
  "day_type": "weekday",
  "time_start": "09:00",
  "time_end": "18:00",
  "price_type": "multiply",
  "price_value": 0.8,
  "min_hours": 2,
  "max_hours": 8,
  "priority": 10
}
```

**字段说明**：
- `package_id`: 套餐ID（必填）
- `rule_name`: 规则名称（必填）
- `day_type`: 日期类型，可选值：`weekday`(工作日)、`weekend`(周末)、`holiday`(节假日)、`special`(特殊日期)
- `time_start`: 时间段开始（HH:mm格式）
- `time_end`: 时间段结束（HH:mm格式）
- `price_type`: 价格类型，可选值：`fixed`(固定价格)、`multiply`(倍数)、`add`(加价)
- `price_value`: 价格值
- `min_hours`: 最少预订小时数（默认1）
- `max_hours`: 最多预订小时数（默认24）
- `priority`: 规则优先级（默认0）

**价格类型说明**：
- `fixed`: 固定价格，`price_value`为每小时固定价格
- `multiply`: 倍数调整，`price_value`为倍数（如0.8表示8折，1.5表示1.5倍价格）
- `add`: 加价调整，`price_value`为每小时加价金额

### 2.2 更新套餐规则

**接口地址**：`PUT /rooms/package-rules`

### 2.3 获取套餐规则列表

**接口地址**：`GET /rooms/package-rules`

**查询参数**：
- `package_id`: 套餐ID（必填）

### 2.4 删除套餐规则

**接口地址**：`DELETE /rooms/package-rules/{id}`

---

## 3. 特殊日期管理

### 3.1 获取特殊日期列表

**接口地址**：`GET /rooms/special-dates`

**查询参数**：
- `page`: 页码（默认1）
- `page_size`: 每页数量（默认20）

### 3.2 创建特殊日期

**接口地址**：`POST /rooms/special-dates`

**请求参数**：
```json
{
  "date": "2024-02-14",
  "date_type": "festival",
  "name": "情人节",
  "description": "情人节特殊定价"
}
```

**字段说明**：
- `date`: 特殊日期（YYYY-MM-DD格式）
- `date_type`: 日期类型，可选值：`holiday`(法定节假日)、`festival`(传统节日)、`special`(特殊活动日)
- `name`: 日期名称
- `description`: 描述信息

### 3.3 删除特殊日期

**接口地址**：`DELETE /rooms/special-dates/{id}`

---

## 4. 价格计算

### 4.1 价格计算接口

**接口地址**：`POST /rooms/calculate-price`

**请求参数**：
```json
{
  "room_id": 1,
  "start_time": "2024-02-14 14:00:00",
  "hours": 3
}
```

**响应示例**：
```json
{
  "code": 200,
  "message": "OK",
  "data": {
    "room_id": 1,
    "start_time": "2024-02-14 14:00:00",
    "hours": 3,
    "final_price": 180.0,
    "applied_rules": [
      "情人节特价套餐: 情人节加价规则价格调整 (150.00 -> 180.00)"
    ]
  }
}
```

---

## 5. 使用示例

### 场景1：工作日优惠套餐

```bash
# 1. 创建工作日优惠套餐
curl -X POST /api/admin/rooms/packages \
-H "Authorization: Bearer YOUR_TOKEN" \
-H "Content-Type: application/json" \
-d '{
  "room_id": 1,
  "package_name": "工作日优惠套餐",
  "description": "工作日享受优惠价格",
  "priority": 10
}'

# 2. 添加工作日白天8折规则
curl -X POST /api/admin/rooms/package-rules \
-H "Authorization: Bearer YOUR_TOKEN" \
-H "Content-Type: application/json" \
-d '{
  "package_id": 1,
  "rule_name": "工作日白天8折",
  "day_type": "weekday",
  "time_start": "09:00",
  "time_end": "18:00",
  "price_type": "multiply",
  "price_value": 0.8,
  "priority": 10
}'
```

### 场景2：周末加价套餐

```bash
# 1. 创建周末加价套餐
curl -X POST /api/admin/rooms/packages \
-H "Authorization: Bearer YOUR_TOKEN" \
-H "Content-Type: application/json" \
-d '{
  "room_id": 1,
  "package_name": "周末加价套餐",
  "description": "周末时段价格上调",
  "priority": 20
}'

# 2. 添加周末1.5倍价格规则
curl -X POST /api/admin/rooms/package-rules \
-H "Authorization: Bearer YOUR_TOKEN" \
-H "Content-Type: application/json" \
-d '{
  "package_id": 2,
  "rule_name": "周末加价",
  "day_type": "weekend",
  "price_type": "multiply",
  "price_value": 1.5,
  "priority": 15
}'
```

### 场景3：节假日特价

```bash
# 1. 创建特殊日期
curl -X POST /api/admin/rooms/special-dates \
-H "Authorization: Bearer YOUR_TOKEN" \
-H "Content-Type: application/json" \
-d '{
  "date": "2024-02-14",
  "date_type": "festival",
  "name": "情人节",
  "description": "情人节特殊定价"
}'

# 2. 创建节假日套餐
curl -X POST /api/admin/rooms/packages \
-H "Authorization: Bearer YOUR_TOKEN" \
-H "Content-Type: application/json" \
-d '{
  "room_id": 1,
  "package_name": "节假日特价套餐",
  "priority": 30
}'

# 3. 添加节假日2倍价格规则
curl -X POST /api/admin/rooms/package-rules \
-H "Authorization: Bearer YOUR_TOKEN" \
-H "Content-Type: application/json" \
-d '{
  "package_id": 3,
  "rule_name": "节假日特价",
  "day_type": "special",
  "price_type": "multiply",
  "price_value": 2.0,
  "priority": 25
}'
```

---

## 6. 定价逻辑说明

### 6.1 优先级规则

1. **套餐优先级**：数字越大优先级越高
2. **规则优先级**：同套餐内规则按优先级排序
3. **日期类型优先级**：特殊日期 > 节假日 > 周末 > 工作日

### 6.2 价格计算流程

1. 获取房间基础价格
2. 查询生效的套餐（按优先级降序）
3. 在每个套餐中查找匹配的规则：
   - 检查日期类型是否匹配
   - 检查时间段是否匹配
   - 检查小时数限制
4. 按优先级应用第一个匹配的规则
5. 返回最终价格和应用的规则说明

### 6.3 示例计算

假设房间基础价格为50元/小时：

- **工作日白天（9:00-18:00）**：50 * 0.8 = 40元/小时
- **周末全天**：50 * 1.5 = 75元/小时  
- **情人节**：50 * 2.0 = 100元/小时

---

## 7. 错误码说明

| 错误码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 参数错误 |
| 401 | 未授权 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 8. 注意事项

1. **时间格式**：严格按照指定格式传递，时间使用24小时制
2. **优先级设置**：合理设置优先级避免规则冲突
3. **日期范围**：套餐的生效日期范围要合理设置
4. **规则限制**：同一套餐内避免创建冲突的规则
5. **价格合理性**：设置价格时要考虑业务合理性

---

这个套餐系统可以灵活支持各种复杂的定价策略，帮助实现精细化的房间价格管理。 