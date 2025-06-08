# 房间套餐管理使用示例

## 📋 套餐类型说明

### 1. 灵活时长套餐 (flexible)
- **特点**: 用户可以自由选择预订时长
- **适用场景**: 按小时收费的会议室
- **配置**: 设置最小和最大时长限制

### 2. 固定时长套餐 (fixed_hours)
- **特点**: 固定时长，如3小时套餐、6小时套餐
- **适用场景**: 特定时长的会议包、培训包
- **配置**: 设置具体的固定小时数

### 3. 全天套餐 (daily)
- **特点**: 24小时全天使用
- **适用场景**: 全天会议、研讨会
- **配置**: 自动设为24小时

### 4. 周套餐 (weekly)
- **特点**: 7天168小时长期使用
- **适用场景**: 长期项目、团队工作室
- **配置**: 自动设为168小时

## 🚀 API 使用示例

### 创建3小时工作套餐

```bash
POST /api/admin/rooms/packages
```

```json
{
  "room_id": 1,
  "package_name": "3小时工作套餐",
  "description": "适合短期会议和小组讨论，固定3小时",
  "package_type": "fixed_hours",
  "fixed_hours": 3,
  "min_hours": 3,
  "max_hours": 3,
  "base_price": 180.00,
  "priority": 10
}
```

### 为套餐添加定价规则

```bash
POST /api/admin/rooms/package-rules
```

```json
{
  "package_id": 1,
  "rule_name": "工作日标准价",
  "day_type": "weekday",
  "price_type": "fixed",
  "price_value": 180.00,
  "min_hours": 3,
  "max_hours": 3
}
```

```json
{
  "package_id": 1,
  "rule_name": "周末加价20%",
  "day_type": "weekend",
  "price_type": "multiply",
  "price_value": 1.2,
  "min_hours": 3,
  "max_hours": 3
}
```

### 创建灵活时长套餐

```bash
POST /api/admin/rooms/packages
```

```json
{
  "room_id": 2,
  "package_name": "灵活时长套餐",
  "description": "按需预订，1-12小时任选",
  "package_type": "flexible",
  "fixed_hours": 0,
  "min_hours": 1,
  "max_hours": 12,
  "base_price": 60.00,
  "priority": 5
}
```

### 添加时间段优惠规则

```json
{
  "package_id": 4,
  "rule_name": "工作日时段9-18点8折",
  "day_type": "weekday",
  "time_start": "09:00",
  "time_end": "18:00",
  "price_type": "multiply",
  "price_value": 0.8,
  "min_hours": 1,
  "max_hours": 12
}
```

## 💡 套餐配置最佳实践

### 1. 3小时工作套餐
```json
{
  "package_name": "3小时工作套餐",
  "package_type": "fixed_hours",
  "fixed_hours": 3,
  "base_price": 180.00,
  "rules": [
    {
      "rule_name": "工作日标准价",
      "day_type": "weekday",
      "price_type": "fixed",
      "price_value": 180.00
    },
    {
      "rule_name": "周末加价",
      "day_type": "weekend",
      "price_type": "multiply",
      "price_value": 1.2
    }
  ]
}
```

### 2. 周末6小时套餐
```json
{
  "package_name": "周末6小时套餐",
  "package_type": "fixed_hours",
  "fixed_hours": 6,
  "base_price": 320.00,
  "rules": [
    {
      "rule_name": "周末专享价",
      "day_type": "weekend",
      "price_type": "fixed",
      "price_value": 320.00
    },
    {
      "rule_name": "节假日特价",
      "day_type": "holiday",
      "price_type": "multiply",
      "price_value": 1.5
    }
  ]
}
```

### 3. 全天会议套餐
```json
{
  "package_name": "全天会议套餐",
  "package_type": "daily",
  "base_price": 800.00,
  "rules": [
    {
      "rule_name": "工作日全天价",
      "day_type": "weekday",
      "price_type": "fixed",
      "price_value": 800.00
    },
    {
      "rule_name": "周末全天价",
      "day_type": "weekend",
      "price_type": "fixed",
      "price_value": 960.00
    }
  ]
}
```

## 🎯 价格计算逻辑

### 固定时长套餐
- **规则**: 价格为总价，不按小时计算
- **示例**: 3小时套餐工作日180元，周末216元(180*1.2)

### 灵活时长套餐
- **规则**: 基础价格 × 规则调整 × 实际小时数
- **示例**: 基础60元/小时，工作日8折，预订3小时 = 60 × 0.8 × 3 = 144元

### 全天/周套餐
- **规则**: 固定总价，不论实际使用时长
- **示例**: 全天套餐800元，使用几小时都是800元

## 🔧 管理界面功能

### 套餐列表
- 显示所有套餐类型、价格、状态
- 支持按房间ID筛选
- 显示套餐类型文本说明

### 套餐编辑
- 可修改套餐类型、时长限制
- 实时验证输入参数
- 支持启用/禁用套餐

### 规则管理
- 每个套餐可配置多个定价规则
- 支持不同时间段、日期类型的差异化定价
- 规则优先级自动排序

## 📊 使用统计

通过套餐系统，您可以：
- 🎯 提供多样化的产品选择
- 💰 优化收益结构
- 📈 提高客户满意度
- 🕒 灵活应对不同时段的需求
- 🎉 推出促销活动和特殊套餐

## 🚀 实际应用场景

1. **创业公司**: 3小时敏捷会议套餐
2. **培训机构**: 全天培训套餐
3. **大型企业**: 周套餐长期项目
4. **个人用户**: 灵活时长按需预订
5. **活动公司**: 节假日特价套餐 