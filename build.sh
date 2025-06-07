#!/bin/bash

# NASA Go Admin 综合构建脚本
# 支持开发、测试、生产等不同环境的构建

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    echo "NASA Go Admin 构建脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help          显示帮助信息"
    echo "  -m, --mode MODE     构建模式 (dev|test|prod) [默认: dev]"
    echo "  -o, --output FILE   输出文件名 [默认: nasa-go-admin]"
    echo "  -v, --verbose       详细输出"
    echo "  --clean             构建前清理"
    echo ""
    echo "示例:"
    echo "  $0                  # 开发模式构建"
    echo "  $0 -m prod          # 生产模式构建"
    echo "  $0 --clean -m test  # 清理后测试模式构建"
}

# 默认参数
BUILD_MODE="dev"
OUTPUT_FILE="nasa-go-admin"
VERBOSE=false
CLEAN=false

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -m|--mode)
            BUILD_MODE="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        --clean)
            CLEAN=true
            shift
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 验证构建模式
case $BUILD_MODE in
    dev|test|prod)
        ;;
    *)
        log_error "无效的构建模式: $BUILD_MODE"
        log_error "支持的模式: dev, test, prod"
        exit 1
        ;;
esac

log_info "开始构建 NASA Go Admin..."
log_info "构建模式: $BUILD_MODE"
log_info "输出文件: $OUTPUT_FILE"

# 清理构建文件
if [ "$CLEAN" = true ]; then
    log_info "清理构建文件..."
    rm -f nasa-go-admin*
    rm -f *.pid
    rm -rf logs/* 2>/dev/null || true
    rm -rf tmp/* 2>/dev/null || true
fi

# 检查Go环境
if ! command -v go &> /dev/null; then
    log_error "Go 未安装或不在PATH中"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
log_info "Go版本: $GO_VERSION"

# 安装依赖
log_info "检查并安装依赖..."
go mod tidy
go mod download

# 设置构建参数
BUILD_FLAGS=""
LDFLAGS="-s -w"

# 获取版本信息
if command -v git &> /dev/null && [ -d .git ]; then
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "unknown")
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
    
    LDFLAGS="$LDFLAGS -X main.Version=$VERSION -X main.Commit=$COMMIT -X main.BuildTime=$BUILD_TIME"
fi

# 根据构建模式设置参数
case $BUILD_MODE in
    dev)
        log_info "开发模式构建..."
        export CGO_ENABLED=1
        BUILD_FLAGS="-race"
        ;;
    test)
        log_info "测试模式构建..."
        export CGO_ENABLED=1
        BUILD_FLAGS="-race"
        # 运行测试
        log_info "运行测试..."
        go test -v ./...
        ;;
    prod)
        log_info "生产模式构建..."
        export CGO_ENABLED=0
        export GOOS=linux
        BUILD_FLAGS="-a -installsuffix cgo"
        LDFLAGS="$LDFLAGS -extldflags '-static'"
        ;;
esac

# 构建应用
log_info "编译应用..."
BUILD_CMD="go build $BUILD_FLAGS -ldflags=\"$LDFLAGS\" -o $OUTPUT_FILE main.go"

if [ "$VERBOSE" = true ]; then
    log_info "构建命令: $BUILD_CMD"
fi

eval $BUILD_CMD

if [ $? -eq 0 ]; then
    log_info "构建成功！"
    
    # 显示文件信息
    if [ -f "$OUTPUT_FILE" ]; then
        FILE_SIZE=$(ls -lh "$OUTPUT_FILE" | awk '{print $5}')
        log_info "输出文件: $OUTPUT_FILE ($FILE_SIZE)"
        
        # 设置执行权限
        chmod +x "$OUTPUT_FILE"
        
        # 生产模式下创建压缩包
        if [ "$BUILD_MODE" = "prod" ]; then
            log_info "创建发布包..."
            tar -czf "${OUTPUT_FILE}-${VERSION:-latest}.tar.gz" "$OUTPUT_FILE"
            log_info "发布包: ${OUTPUT_FILE}-${VERSION:-latest}.tar.gz"
        fi
    fi
    
    log_info "构建完成！"
else
    log_error "构建失败！"
    exit 1
fi 