# NASA Go Admin Makefile
# 提供常用的开发和部署命令

.PHONY: help build run test clean deps start stop monitoring

# 默认目标
help:
	@echo "NASA Go Admin 可用命令:"
	@echo "  make build          - 编译应用"
	@echo "  make run            - 运行应用"
	@echo "  make run-optimized  - 使用优化参数运行应用"
	@echo "  make start          - 启动应用服务(后台)"
	@echo "  make stop           - 停止应用服务"
	@echo "  make test           - 运行测试"
	@echo "  make clean          - 清理构建文件"
	@echo "  make deps           - 安装依赖"
	@echo "  make monitoring     - 启动监控系统"
	@echo "  make performance    - 运行性能测试"

# 构建应用
build:
	@echo "🔨 编译应用..."
	go build -ldflags="-s -w" -o nasa-go-admin main.go

# 运行应用
run:
	@echo "🚀 启动应用..."
	go run main.go

# 使用优化参数运行
run-optimized:
	@echo "🚀 使用优化参数启动应用..."
	./start_optimized.sh

# 运行测试
test:
	@echo "🧪 运行测试..."
	go test -v ./...

# 清理构建文件
clean:
	@echo "🧹 清理构建文件..."
	rm -f nasa-go-admin
	rm -f *.pid
	rm -rf logs/*
	rm -rf tmp/*

# 安装依赖
deps:
	@echo "📦 安装依赖..."
	go mod tidy
	go mod download

# 启动应用服务
start:
	@echo "🚀 启动应用服务..."
	./nasa-go-admin &
	echo $$! > nasa-go-admin.pid
	@echo "应用已启动，PID: $$(cat nasa-go-admin.pid)"

# 停止应用服务
stop:
	@echo "🛑 停止应用服务..."
	@if [ -f nasa-go-admin.pid ]; then \
		kill $$(cat nasa-go-admin.pid) && rm -f nasa-go-admin.pid && echo "应用已停止"; \
	else \
		echo "未找到运行的应用进程"; \
	fi

# 启动监控系统
monitoring:
	@echo "📊 启动监控系统..."
	./start_monitoring.sh

# 运行性能测试
performance:
	@echo "⚡ 运行性能测试..."
	./scripts/performance_test_enhanced.sh

# 开发模式 (使用Air热重载)
dev:
	@echo "🔥 启动开发模式..."
	air -c .air.toml

# 生产构建
build-prod:
	@echo "🏭 生产环境构建..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o nasa-go-admin main.go 