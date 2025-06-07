# NASA Go Admin 构建和运行指南

## 📋 概述

本项目使用传统的 Go 构建方式，无需 Docker。提供了多种构建和运行选项。

## 🚀 快速开始

### 1. 构建应用

```bash
# 使用 Makefile（推荐）
make build

# 或使用构建脚本
./build.sh

# 或直接使用 go build
go build -o nasa-go-admin main.go
```

### 2. 运行应用

```bash
# 前台运行
make run
# 或
go run main.go

# 后台运行
make start
# 或
./start.sh

# 使用优化参数运行
make run-optimized
# 或
./start_optimized.sh
```

### 3. 停止应用

```bash
# 停止后台运行的应用
make stop
# 或
./stop.sh
```

## 🛠 构建选项

### 构建脚本参数

`build.sh` 脚本支持多种构建模式：

```bash
# 开发模式（默认）
./build.sh

# 测试模式（包含测试）
./build.sh -m test

# 生产模式（优化构建）
./build.sh -m prod

# 清理后构建
./build.sh --clean

# 详细输出
./build.sh -v

# 自定义输出文件名
./build.sh -o my-app

# 查看帮助
./build.sh -h
```

### Makefile 命令

```bash
make help           # 显示所有可用命令
make build          # 编译应用
make run            # 前台运行应用
make run-optimized  # 使用优化参数运行
make start          # 后台启动应用
make stop           # 停止应用
make test           # 运行测试
make clean          # 清理构建文件
make deps           # 安装依赖
make dev            # 开发模式（使用 Air 热重载）
make build-prod     # 生产环境构建
make monitoring     # 启动监控系统
make performance    # 运行性能测试
```

## 📁 项目结构

```
├── build.sh                    # 综合构建脚本
├── start.sh                    # 简单启动脚本
├── stop.sh                     # 停止脚本
├── start_optimized.sh          # 优化启动脚本
├── start_monitoring.sh         # 监控启动脚本
├── Makefile                    # Make 构建文件
├── main.go                     # 主程序入口
├── logs/                       # 日志目录
├── data/                       # 数据目录
└── tmp/                        # 临时文件目录
```

## 🔧 环境准备

### 必需依赖

- Go 1.19+
- MySQL 8.0+
- Redis 6.0+

### 可选依赖

- MongoDB 6.0+（用于日志存储）
- Air（用于热重载开发）

### 安装依赖

```bash
# 安装 Go 依赖
make deps
# 或
go mod tidy && go mod download

# 安装 Air（可选）
go install github.com/cosmtrek/air@latest
```

## 🏃‍♂️ 运行模式

### 1. 开发模式

```bash
# 使用 Air 热重载
make dev

# 或直接运行
go run main.go
```

### 2. 生产模式

```bash
# 构建生产版本
make build-prod

# 启动应用
./start.sh
```

### 3. 测试模式

```bash
# 运行测试
make test

# 测试模式构建
./build.sh -m test
```

## 📊 监控

启动监控系统（Prometheus + Grafana）：

```bash
make monitoring
# 或
./start_monitoring.sh
```

访问地址：
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin123)

## 🔍 日志管理

### 日志位置

- 访问日志: `logs/access/app.log`
- 错误日志: `logs/error/app.log`

### 查看日志

```bash
# 实时查看访问日志
tail -f logs/access/app.log

# 实时查看错误日志
tail -f logs/error/app.log
```

## ⚡ 性能测试

```bash
make performance
# 或
./scripts/performance_test_enhanced.sh
```

## 🧹 清理

```bash
# 清理构建文件
make clean

# 手动清理
rm -f nasa-go-admin *.pid
rm -rf logs/* tmp/*
```

## 🔧 故障排除

### 常见问题

1. **端口被占用**
   ```bash
   # 查找占用端口的进程
   lsof -i :8801
   
   # 杀死进程
   kill -9 <PID>
   ```

2. **权限问题**
   ```bash
   # 设置脚本执行权限
   chmod +x *.sh
   ```

3. **依赖服务未启动**
   - 确保 MySQL 在 localhost:3306 运行
   - 确保 Redis 在 localhost:6379 运行

### 健康检查

```bash
# 检查应用是否运行
curl http://localhost:8801/health

# 检查进程状态
ps aux | grep nasa-go-admin
```

## 📝 开发建议

1. 使用 `make dev` 进行开发（支持热重载）
2. 提交前运行 `make test` 确保测试通过
3. 生产部署使用 `make build-prod` 构建
4. 定期查看日志排查问题 