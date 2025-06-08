# 用户端房间预订套餐系统使用指南

## 📋 功能概述

用户端房间预订系统现已支持套餐选择功能，用户可以：

1. **查看房间信息** - 浏览可用房间
2. **选择预约时间** - 设定预约开始时间和时长
3. **选择套餐** - 从多种套餐中选择最优惠的方案
4. **价格预览** - 实时查看价格计算明细
5. **确认支付** - 完成预订并支付

## 🔄 完整预订流程

### 1. 查看房间列表
```bash
GET /api/app/rooms
```

**响应示例：**
```json
{
  "code": 200,
  "data": {
    "total": 10,
    "list": [
      {
        "id": 1,
        "room_name": "豪华包厢A",
        "room_type": "luxury",
        "capacity": 8,
        "hourly_rate": 120.00,
        "status": 1,
        "status_text": "可用"
      }
    ]
  }
}
```

### 2. 查看房间详情
```bash
GET /api/app/rooms/1
```

### 3. 获取房间可用套餐
```bash
GET /api/app/rooms/packages?room_id=1&start_time=2024-06-08 14:00:00&hours=3
```

**响应示例：**
```json
{
  "code": 200,
  "data": {
    "room_id": 1,
    "start_time": "2024-06-08 14:00:00",
    "hours": 3,
    "day_type": "weekday",
    "day_type_text": "工作日",
    "packages": [
      {
        "package_id": 0,
        "package_name": "基础价格",
        "description": "按小时计费，无优惠",
        "package_type": "basic",
        "package_type_text": "基础价格",
        "base_price": 120.00,
        "final_price": 360.00,
        "original_price": 360.00,
        "discount_amount": 0,
        "discount_percent": 0,
        "is_recommended": false,
        "is_available": true
      },
      {
        "package_id": 1,
        "package_name": "3小时工作套餐",
        "description": "工作日3小时优惠套餐",
        "package_type": "fixed_hours",
        "package_type_text": "固定时长套餐",
        "fixed_hours": 3,
        "base_price": 180.00,
        "final_price": 180.00,
        "original_price": 360.00,
        "discount_amount": 180.00,
        "discount_percent": 50.0,
        "rule_name": "工作日优惠",
        "day_type": "weekday",
        "day_type_text": "工作日",
        "is_recommended": true,
        "is_available": true
      },
      {
        "package_id": 3,
        "package_name": "灵活时长套餐",
        "description": "1-12小时可选，工作日8.5折",
        "package_type": "flexible",
        "package_type_text": "灵活时长套餐",
        "min_hours": 1,
        "max_hours": 12,
        "base_price": 100.00,
        "final_price": 306.00,
        "original_price": 360.00,
        "discount_amount": 54.00,
        "discount_percent": 15.0,
        "rule_name": "工作日8.5折",
        "is_recommended": false,
        "is_available": true
      }
    ]
  }
}
```

### 4. 价格预览（可选）
```bash
POST /api/app/bookings/price-preview
Authorization: Bearer <token>
Content-Type: application/json

{
  "room_id": 1,
  "start_time": "2024-06-08 14:00:00",
  "hours": 3,
  "package_id": 1
}
```

**响应示例：**
```json
{
  "code": 200,
  "data": {
    "room_id": 1,
    "hours": 3,
    "base_price": 120.00,
    "package_id": 1,
    "package_name": "3小时工作套餐",
    "original_price": 360.00,
    "final_price": 180.00,
    "discount_amount": 180.00,
    "discount_percent": 50.0,
    "rule_name": "工作日优惠",
    "day_type": "weekday",
    "day_type_text": "工作日",
    "price_breakdown": {
      "base_hourly_rate": 120.00,
      "hours": 3,
      "sub_total": 360.00,
      "rule_type": "package",
      "rule_value": -180.00,
      "adjustment": 180.00,
      "final_total": 180.00
    }
  }
}
```

### 5. 创建预订
```bash
POST /api/app/bookings
Authorization: Bearer <token>
Content-Type: application/json

{
  "room_id": 1,
  "start_time": "2024-06-08 14:00:00",
  "hours": 3,
  "package_id": 1,
  "contact_name": "张三",
  "contact_phone": "13800138000",
  "remarks": "商务会议使用"
}
```

**响应示例：**
```json
{
  "code": 200,
  "data": {
    "id": 123,
    "booking_no": "BK20240608140012345678",
    "room_id": 1,
    "start_time": "2024-06-08T14:00:00Z",
    "end_time": "2024-06-08T17:00:00Z",
    "hours": 3,
    "total_amount": 180.00,
    "status": 1,
    "status_text": "待支付",
    "package_id": 1,
    "package_name": "3小时工作套餐",
    "original_price": 360.00,
    "package_price": 180.00,
    "discount_amount": 180.00,
    "price_breakdown": "{\"base_price\":120,\"hours\":3,\"original_price\":360,\"package_name\":\"3小时工作套餐\",\"package_type\":\"fixed_hours\",\"final_price\":180,\"discount_amount\":180,\"rule_name\":\"工作日优惠\",\"rule_type\":\"fixed\",\"rule_value\":180}"
  }
}
```

### 6. 查看我的预订列表
```bash
GET /api/app/bookings?page=1&page_size=10
Authorization: Bearer <token>
```

## 📊 套餐类型说明

### 1. 基础价格 (basic)
- **特点**: 按小时计费，无优惠
- **适用**: 所有时间段
- **计费**: 房间基础价格 × 小时数

### 2. 固定时长套餐 (fixed_hours)
- **特点**: 固定时长，固定价格
- **示例**: 3小时工作套餐 180元
- **计费**: 套餐固定价格，不论实际使用时长

### 3. 灵活时长套餐 (flexible)
- **特点**: 可选时长范围，按实际时长计费
- **示例**: 1-12小时可选，工作日8.5折
- **计费**: 基础价格 × 实际小时数 × 折扣系数

### 4. 全天套餐 (daily)
- **特点**: 24小时使用权
- **适用**: 全天会议、活动等
- **计费**: 固定全天价格

### 5. 周套餐 (weekly)
- **特点**: 7天168小时使用权
- **适用**: 长期租用
- **计费**: 固定周价格

## 💡 使用建议

### 选择套餐的策略

1. **短时间使用（1-2小时）**
   - 优先选择灵活时长套餐
   - 对比基础价格和套餐价格

2. **固定时长使用（3-6小时）**
   - 优先选择对应的固定时长套餐
   - 通常有较大优惠

3. **全天使用（8小时以上）**
   - 选择全天套餐
   - 比按小时计费更优惠

4. **长期使用（多天）**
   - 选择周套餐
   - 享受长期优惠

### 最佳实践

1. **提前查看套餐**: 在预订前先查看可用套餐，选择最优惠的方案
2. **关注推荐标识**: 系统会标记优惠幅度大的套餐为推荐
3. **使用价格预览**: 确认最终价格后再提交预订
4. **注意时间限制**: 某些套餐可能有时间段限制（如工作日专享）

## 🔍 错误处理

### 常见错误码

- `20001`: 参数错误或业务逻辑错误
- `10002`: 用户未登录或认证失败
- `40001`: 房间不存在
- `40002`: 套餐不存在或不可用
- `40003`: 时间段已被预订

### 错误示例

```json
{
  "code": 20001,
  "message": "该时间段房间已被预订",
  "data": null
}
```

## 📱 前端集成建议

### 1. 套餐选择界面
```javascript
// 获取套餐列表
const getPackages = async (roomId, startTime, hours) => {
  const response = await fetch(`/api/app/rooms/packages?room_id=${roomId}&start_time=${startTime}&hours=${hours}`);
  const data = await response.json();
  return data.data.packages;
};

// 渲染套餐选项
const renderPackages = (packages) => {
  return packages.map(pkg => ({
    id: pkg.package_id,
    name: pkg.package_name,
    description: pkg.description,
    price: pkg.final_price,
    originalPrice: pkg.original_price,
    discount: pkg.discount_percent,
    recommended: pkg.is_recommended,
    available: pkg.is_available
  }));
};
```

### 2. 价格计算
```javascript
// 实时价格预览
const previewPrice = async (roomId, startTime, hours, packageId) => {
  const response = await fetch('/api/app/bookings/price-preview', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      room_id: roomId,
      start_time: startTime,
      hours: hours,
      package_id: packageId
    })
  });
  return response.json();
};
```

### 3. 创建预订
```javascript
// 提交预订
const createBooking = async (bookingData) => {
  const response = await fetch('/api/app/bookings', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify(bookingData)
  });
  return response.json();
};
```

## 🎯 总结

通过套餐系统，用户可以：

1. **节省费用**: 通过选择合适的套餐享受优惠
2. **灵活选择**: 多种套餐类型满足不同需求
3. **透明计费**: 详细的价格明细和预览功能
4. **简化流程**: 一站式预订体验

系统支持多种定价策略，包括时间段优惠、日期类型优惠、固定套餐价格等，为用户提供最优的预订体验。 