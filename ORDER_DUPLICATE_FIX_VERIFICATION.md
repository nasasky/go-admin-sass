# 重复下单问题修复验证

## 问题描述

用户反映的问题：当用户下单时库存不足，提示"无法下单"是正常的，但当设置商品库存有库存时，用户再次提交下单却提示"请勿重复下单，如有问题请联系客服"，这是不正常的，应该可以下单才对。

## 问题根因分析

### 原问题原因
1. **幂等性键时间窗口设计不合理**：使用小时级别的时间窗口(`20060102_15`)，导致一小时内的所有请求都被认为是重复请求
2. **幂等性检查时机错误**：在业务逻辑检查之前就设置了幂等性标记，导致业务失败(如库存不足)也会阻止后续正常请求
3. **没有区分业务失败和系统重复**：库存不足等业务逻辑失败也被当作重复请求处理

### 修复方案
1. **延迟幂等性设置**：只有在订单成功创建后才设置幂等性标记
2. **业务失败不设标记**：库存不足、商品下架等业务逻辑失败时不设置幂等性标记
3. **订单取消清除标记**：订单被取消时清除对应的幂等性标记

## 修复内容详情

### 1. 修改 `SecureOrderCreator.CreateOrderSecurely()` 方法

#### 修改前逻辑
```go
// 1. 幂等性检查 - 防止重复下单
idempotencyKey := fmt.Sprintf("order_create:%d:%d:%s", uid, params.GoodsId,
    time.Now().Format("20060102_15"))

isDuplicate, err := soc.idempotencyChecker.CheckAndSet(idempotencyKey, 1*time.Hour)
if err != nil {
    log.Printf("幂等性检查失败: %v", err)
} else if isDuplicate {
    return "", fmt.Errorf("请勿重复下单，如有问题请联系客服")
}

// ... 业务逻辑检查和订单创建
```

#### 修改后逻辑
```go
// 1. 生成幂等性键但先不设置，等订单成功后再设置
idempotencyKey := fmt.Sprintf("order_create:%d:%d:%s", uid, params.GoodsId,
    time.Now().Format("20060102_15"))

// 先检查是否存在重复请求，但不立即设置标记
isDuplicate, err := soc.idempotencyChecker.CheckOnly(idempotencyKey)
if err != nil {
    log.Printf("幂等性检查失败: %v", err)
} else if isDuplicate {
    return "", fmt.Errorf("请勿重复下单，如有问题请联系客服")
}

// ... 业务逻辑检查
if goods.Stock < params.Num {
    tx.Rollback()
    // 库存不足是业务逻辑问题，不设置幂等性标记，让用户补充库存后可以重新下单
    return "", fmt.Errorf("库存不足，当前库存: %d，需要: %d", goods.Stock, params.Num)
}

// ... 订单创建成功

// 订单创建成功后才设置幂等性标记
if setErr := soc.idempotencyChecker.SetIdempotencyMark(idempotencyKey, 1*time.Hour); setErr != nil {
    log.Printf("设置幂等性标记失败: %v", setErr)
    // 这个失败不影响订单创建结果
} else {
    log.Printf("✅ 已设置幂等性标记，防止重复下单: %s", idempotencyKey)
}
```

### 2. 新增 `IdempotencyChecker` 方法

#### 新增的方法
```go
// CheckOnly 仅检查幂等性，不设置标记
func (ic *IdempotencyChecker) CheckOnly(key string) (bool, error) {
    if ic.redisClient == nil {
        log.Printf("Redis不可用，跳过幂等性检查: %s", key)
        return false, nil
    }

    ctx := context.Background()
    exists, err := ic.redisClient.Exists(ctx, key).Result()
    if err != nil {
        log.Printf("幂等性检查失败: %v", err)
        return false, nil
    }

    return exists > 0, nil
}

// SetIdempotencyMark 设置幂等性标记
func (ic *IdempotencyChecker) SetIdempotencyMark(key string, expiration time.Duration) error {
    if ic.redisClient == nil {
        log.Printf("Redis不可用，跳过设置幂等性标记: %s", key)
        return nil
    }

    ctx := context.Background()
    _, err := ic.redisClient.Set(ctx, key, "1", expiration).Result()
    if err != nil {
        return fmt.Errorf("设置幂等性标记失败: %w", err)
    }
    return nil
}

// ClearIdempotencyMark 清除幂等性标记
func (ic *IdempotencyChecker) ClearIdempotencyMark(key string) error {
    if ic.redisClient == nil {
        log.Printf("Redis不可用，跳过清除幂等性标记: %s", key)
        return nil
    }

    ctx := context.Background()
    _, err := ic.redisClient.Del(ctx, key).Result()
    if err != nil {
        return fmt.Errorf("清除幂等性标记失败: %w", err)
    }
    return nil
}
```

### 3. 修改订单取消逻辑

在 `CancelExpiredOrder()` 方法中添加清除幂等性标记的逻辑：

```go
// 清除幂等性标记，允许用户重新下单
go func() {
    idempotencyKey := fmt.Sprintf("order_create:%d:%d:%s", order.UserId, order.GoodsId,
        order.CreateTime.Format("20060102_15"))
    if clearErr := soc.idempotencyChecker.ClearIdempotencyMark(idempotencyKey); clearErr != nil {
        log.Printf("清除幂等性标记失败: %v", clearErr)
    } else {
        log.Printf("已清除订单 %s 的幂等性标记，用户可重新下单", orderNo)
    }
}()
```

## 测试验证步骤

### 测试场景1：库存不足后补货再下单
1. **准备数据**：创建一个商品，设置库存为0
2. **第一次下单**：用户尝试下单，应该提示"库存不足"
3. **补充库存**：将商品库存设置为充足数量
4. **第二次下单**：用户再次下单，应该成功创建订单，而不是提示"请勿重复下单"

### 测试场景2：商品下架后重新上架下单
1. **准备数据**：创建一个商品，设置状态为下架
2. **第一次下单**：用户尝试下单，应该提示"商品已下架或不可购买"
3. **重新上架**：将商品状态设置为正常
4. **第二次下单**：用户再次下单，应该成功创建订单

### 测试场景3：订单取消后重新下单
1. **创建订单**：用户成功创建一个pending状态的订单
2. **取消订单**：系统或用户取消订单
3. **重新下单**：用户重新下单相同商品，应该能够成功创建新订单

### 测试场景4：真正的重复下单防护
1. **快速连续请求**：用户在短时间内连续提交多个相同的下单请求
2. **预期结果**：第一个请求成功创建订单，后续请求应该被拦截并提示"请勿重复下单"

## 验证API测试

### API接口
```
POST /app/order/create
Content-Type: application/json

{
    "goods_id": 商品ID,
    "num": 购买数量
}
```

### 测试用例

#### 用例1：库存不足场景
```bash
# 1. 设置商品库存为0
# 2. 发起下单请求
curl -X POST "http://localhost:8080/app/order/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"goods_id": 1, "num": 1}'

# 预期响应：{"code": 20001, "message": "库存不足，当前库存: 0，需要: 1"}

# 3. 设置商品库存为充足数量（如：10）
# 4. 再次发起下单请求
curl -X POST "http://localhost:8080/app/order/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"goods_id": 1, "num": 1}'

# 预期响应：{"code": 0, "data": "ORD20241201120130000100010123", "message": "success"}
```

#### 用例2：真实重复下单防护
```bash
# 1. 确保商品库存充足
# 2. 快速连续发起两个相同的下单请求

# 第一个请求
curl -X POST "http://localhost:8080/app/order/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"goods_id": 1, "num": 1}' &

# 第二个请求（几乎同时发起）
curl -X POST "http://localhost:8080/app/order/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"goods_id": 1, "num": 1}' &

# 预期结果：
# 第一个请求：{"code": 0, "data": "ORD...", "message": "success"}
# 第二个请求：{"code": 20001, "message": "请勿重复下单，如有问题请联系客服"}
```

## 预期修复效果

### 修复前问题
- ❌ 库存不足下单失败后，补货再下单仍提示"请勿重复下单"
- ❌ 商品下架后重新上架，用户无法正常下单
- ❌ 订单取消后，用户无法重新下单相同商品

### 修复后效果
- ✅ 库存不足等业务逻辑失败不会阻止后续正常下单
- ✅ 只有成功创建订单后才设置幂等性标记
- ✅ 订单取消时自动清除幂等性标记
- ✅ 保持真正重复请求的防护功能
- ✅ 改善用户体验，减少客服咨询

## 监控和日志

### 关键日志标识
- `✅ 已设置幂等性标记，防止重复下单`：订单创建成功时的标记
- `已清除订单 X 的幂等性标记，用户可重新下单`：订单取消时的清理
- `库存不足是业务逻辑问题，不设置幂等性标记`：业务失败时的说明

### Redis键监控
- 幂等性键格式：`order_create:{user_id}:{goods_id}:{hour}`
- 监控键的生命周期：只有成功订单才会设置，取消订单时会清除

## 回归测试清单

- [ ] 正常下单流程不受影响
- [ ] 真实重复下单仍被正确拦截  
- [ ] 库存不足场景修复验证
- [ ] 商品下架场景修复验证
- [ ] 订单取消后重下单验证
- [ ] 并发下单安全性验证
- [ ] 分布式锁功能正常
- [ ] 钱包支付逻辑正常
- [ ] 订单超时取消正常

---

**修复完成时间**: 2024年12月
**修复版本**: v1.2.0
**影响范围**: 用户下单体验优化，减少误报"重复下单"错误 