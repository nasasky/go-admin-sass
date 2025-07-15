# 会员统计折线图接口文档

## 接口概述

本接口用于获取会员数据的统计信息，支持按日、周、月三种维度生成折线图数据，包含新增会员数、累计会员数、趋势分析等信息。

## 接口信息

- **接口路径**: `/api/admin/member/stats`
- **请求方式**: `GET`
- **认证方式**: 需要管理员Token认证
- **内容类型**: `application/json`

## 请求参数

| 参数名 | 类型 | 必填 | 描述 | 示例值 |
|--------|------|------|------|--------|
| type | string | 是 | 统计类型：daily(每日), weekly(每周), monthly(每月) | `"daily"` |
| start_date | string | 否 | 开始日期，格式：YYYY-MM-DD | `"2024-01-01"` |
| end_date | string | 否 | 结束日期，格式：YYYY-MM-DD | `"2024-01-31"` |
| days | int | 否 | 查询天数(1-365)，当未提供start_date和end_date时使用 | `30` |

## 请求示例

### 1. 获取最近30天的每日会员统计
```bash
GET /api/admin/member/stats?type=daily&days=30
```

### 2. 获取指定日期范围的每日会员统计
```bash
GET /api/admin/member/stats?type=daily&start_date=2024-01-01&end_date=2024-01-31
```

### 3. 获取最近12周的每周会员统计
```bash
GET /api/admin/member/stats?type=weekly&days=84
```

### 4. 获取最近12个月的每月会员统计
```bash
GET /api/admin/member/stats?type=monthly&days=365
```

## 响应格式

### 成功响应 (200)

```json
{
  "code": 0,
  "msg": "成功",
  "data": {
    "type": "daily",
    "date_range": {
      "start_date": "2024-01-01",
      "end_date": "2024-01-31",
      "total_days": 31,
      "current_date": "2024-01-31"
    },
    "summary": {
      "total_members": 1250,
      "new_members_in_period": 186,
      "avg_daily_new_members": 6.0,
      "peak_new_members_day": "2024-01-15",
      "peak_new_members_count": 15,
      "growth_rate": 17.5
    },
    "chart_data": [
      {
        "date": "2024-01-01",
        "new_members": 8,
        "total_members": 1072,
        "day_of_week": "周一",
        "formatted_date": "01月01日"
      },
      {
        "date": "2024-01-02",
        "new_members": 5,
        "total_members": 1077,
        "day_of_week": "周二",
        "formatted_date": "01月02日"
      }
    ],
    "trend_info": {
      "trend": "up",
      "trend_percent": 12.3,
      "trend_desc": "会员增长呈上升趋势，增长12.3%",
      "compared_to_prev": "比上一周期增长了12.3%"
    },
    "updated_at": "2024-01-31 15:30:00"
  }
}
```

### 错误响应 (400/500)

```json
{
  "code": 20001,
  "msg": "参数错误: type字段必须是daily、weekly或monthly中的一个",
  "data": null
}
```

## 响应字段说明

### 根级字段
- `type`: 统计类型
- `date_range`: 日期范围信息
- `summary`: 汇总统计信息
- `chart_data`: 折线图数据点数组
- `trend_info`: 趋势分析信息
- `updated_at`: 数据更新时间

### date_range (日期范围信息)
- `start_date`: 开始日期
- `end_date`: 结束日期
- `total_days`: 总天数
- `current_date`: 当前日期

### summary (汇总统计信息)
- `total_members`: 总会员数
- `new_members_in_period`: 周期内新增会员数
- `avg_daily_new_members`: 日均新增会员数
- `peak_new_members_day`: 新增会员最多的一天
- `peak_new_members_count`: 新增会员最多的一天的数量
- `growth_rate`: 增长率(%)

### chart_data (折线图数据点)
- `date`: 日期（X轴）
- `new_members`: 新增会员数（Y轴1）
- `total_members`: 累计会员数（Y轴2）
- `day_of_week`: 星期几
- `formatted_date`: 格式化的日期显示

### trend_info (趋势信息)
- `trend`: 趋势类型：up(上升)、down(下降)、stable(稳定)
- `trend_percent`: 趋势百分比
- `trend_desc`: 趋势描述
- `compared_to_prev`: 与前一个周期对比

## 前端集成示例

### 使用 Chart.js 绘制折线图

```javascript
// 发起请求获取数据
async function fetchMemberStats(type = 'daily', days = 30) {
  try {
    const response = await fetch(`/api/admin/member/stats?type=${type}&days=${days}`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });
    
    const result = await response.json();
    
    if (result.code === 0) {
      return result.data;
    } else {
      throw new Error(result.msg);
    }
  } catch (error) {
    console.error('获取会员统计数据失败:', error);
    throw error;
  }
}

// 绘制折线图
function renderMemberChart(chartData) {
  const ctx = document.getElementById('memberChart').getContext('2d');
  
  const chart = new Chart(ctx, {
    type: 'line',
    data: {
      labels: chartData.map(item => item.formatted_date),
      datasets: [
        {
          label: '新增会员数',
          data: chartData.map(item => item.new_members),
          borderColor: 'rgb(75, 192, 192)',
          backgroundColor: 'rgba(75, 192, 192, 0.2)',
          yAxisID: 'y'
        },
        {
          label: '累计会员数',
          data: chartData.map(item => item.total_members),
          borderColor: 'rgb(255, 99, 132)',
          backgroundColor: 'rgba(255, 99, 132, 0.2)',
          yAxisID: 'y1'
        }
      ]
    },
    options: {
      responsive: true,
      plugins: {
        title: {
          display: true,
          text: '会员增长趋势图'
        },
        legend: {
          display: true,
          position: 'top'
        }
      },
      scales: {
        y: {
          type: 'linear',
          display: true,
          position: 'left',
          title: {
            display: true,
            text: '新增会员数'
          }
        },
        y1: {
          type: 'linear',
          display: true,
          position: 'right',
          title: {
            display: true,
            text: '累计会员数'
          },
          grid: {
            drawOnChartArea: false,
          },
        }
      }
    }
  });
  
  return chart;
}

// 使用示例
async function loadMemberStats() {
  try {
    const statsData = await fetchMemberStats('daily', 30);
    
    // 渲染图表
    const chart = renderMemberChart(statsData.chart_data);
    
    // 显示汇总信息
    document.getElementById('totalMembers').textContent = statsData.summary.total_members;
    document.getElementById('newMembers').textContent = statsData.summary.new_members_in_period;
    document.getElementById('avgDaily').textContent = statsData.summary.avg_daily_new_members;
    document.getElementById('growthRate').textContent = `${statsData.summary.growth_rate}%`;
    
    // 显示趋势信息
    document.getElementById('trendDesc').textContent = statsData.trend_info.trend_desc;
    
    console.log('会员统计数据加载完成');
  } catch (error) {
    console.error('加载会员统计数据失败:', error);
  }
}
```

### 使用 ECharts 绘制折线图

```javascript
function renderEChartsChart(chartData, summary) {
  const chartDom = document.getElementById('memberEChart');
  const myChart = echarts.init(chartDom);
  
  const option = {
    title: {
      text: '会员增长趋势',
      subtext: `总会员: ${summary.total_members} | 周期新增: ${summary.new_members_in_period}`
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross',
        label: {
          backgroundColor: '#6a7985'
        }
      }
    },
    legend: {
      data: ['新增会员数', '累计会员数']
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: chartData.map(item => item.formatted_date)
    },
    yAxis: [
      {
        type: 'value',
        name: '新增会员数',
        position: 'left',
        axisLabel: {
          formatter: '{value}'
        }
      },
      {
        type: 'value',
        name: '累计会员数',
        position: 'right',
        axisLabel: {
          formatter: '{value}'
        }
      }
    ],
    series: [
      {
        name: '新增会员数',
        type: 'line',
        yAxisIndex: 0,
        data: chartData.map(item => item.new_members),
        smooth: true,
        itemStyle: {
          color: '#5470c6'
        }
      },
      {
        name: '累计会员数',
        type: 'line',
        yAxisIndex: 1,
        data: chartData.map(item => item.total_members),
        smooth: true,
        itemStyle: {
          color: '#91cc75'
        }
      }
    ]
  };
  
  myChart.setOption(option);
  return myChart;
}
```

## 注意事项

1. **认证要求**: 此接口需要管理员权限，请确保请求头中包含有效的Token
2. **数据范围**: 
   - `days` 参数范围为 1-365 天
   - 如果同时提供 `start_date`、`end_date` 和 `days`，将优先使用日期范围
3. **性能考虑**: 
   - 建议避免查询过长的时间范围
   - 数据按需缓存，提高响应速度
4. **时区处理**: 所有日期时间均使用服务器本地时区
5. **默认值**:
   - daily: 默认查询最近30天
   - weekly: 默认查询最近12周(84天)
   - monthly: 默认查询最近12个月(365天)

## 错误代码

| 错误代码 | 描述 |
|----------|------|
| 20001 | 参数错误 |
| 10002 | 认证失败或权限不足 |
| 50000 | 服务器内部错误 | 