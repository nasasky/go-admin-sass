#!/bin/bash

echo "🚀 启动 NASA-Go-Admin 监控系统..."

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker 未运行，请先启动 Docker"
    exit 1
fi

# 创建监控目录（如果不存在）
mkdir -p monitoring/grafana/provisioning/datasources
mkdir -p monitoring/grafana/provisioning/dashboards
mkdir -p monitoring/grafana/dashboards

# 创建 Grafana 数据源配置
cat > monitoring/grafana/provisioning/datasources/prometheus.yml << EOF
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
EOF

# 创建 Grafana 仪表板配置
cat > monitoring/grafana/provisioning/dashboards/dashboard.yml << EOF
apiVersion: 1

providers:
  - name: 'default'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /var/lib/grafana/dashboards
EOF

# 启动监控服务
echo "📊 启动 Prometheus + Grafana..."
docker-compose -f docker-compose-monitoring.yml up -d

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 10

# 检查服务状态
if docker-compose -f docker-compose-monitoring.yml ps | grep -q "Up"; then
    echo "✅ 监控服务启动成功！"
    echo ""
    echo "📈 访问地址："
    echo "  🔍 Prometheus (数据查询):    http://localhost:9090"
    echo "  📊 Grafana (图表界面):       http://localhost:3000"
    echo "     - 账号: admin"
    echo "     - 密码: admin123"
    echo ""
    echo "🔧 应用监控端点："
    echo "  📊 应用指标:                http://localhost:8801/metrics"
    echo ""
    echo "💡 使用提示："
    echo "  1. 启动您的应用: go run main.go"
    echo "  2. 访问一些API接口产生数据"
    echo "  3. 在 Prometheus 中查询指标"
    echo "  4. 在 Grafana 中创建仪表板"
    echo ""
    echo "📝 常用 Prometheus 查询："
    echo "  - rate(http_requests_total[1m])              # 每分钟请求数"
    echo "  - http_request_duration_seconds{quantile=\"0.95\"}  # 95%响应时间"
    echo "  - db_connections_in_use                      # 数据库连接数"
    echo "  - user_logins_total                          # 用户登录总数"
else
    echo "❌ 监控服务启动失败，请检查日志："
    docker-compose -f docker-compose-monitoring.yml logs
fi 