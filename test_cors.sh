#!/bin/bash

# CORS 跨域测试脚本
# 用于验证修复后的CORS配置是否正常工作

echo "🔍 CORS 跨域配置测试"
echo "===================="

# 配置
SERVER_URL="http://localhost:8801"
TEST_ORIGINS=(
    "http://localhost:3000"
    "http://localhost:9848" 
    "http://192.168.1.100:3000"
    "http://192.168.0.114:9848"
    "http://10.0.0.100:8080"
)

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}测试服务器: ${SERVER_URL}${NC}"
echo ""

# 检查服务器是否运行
echo "🔧 检查服务器状态..."
if ! curl -s "${SERVER_URL}/health" > /dev/null 2>&1; then
    echo -e "${RED}❌ 服务器未运行，请先启动服务器${NC}"
    echo "启动命令: ALLOWED_ORIGINS=* GIN_MODE=debug ./nasa-go-admin"
    exit 1
fi
echo -e "${GREEN}✅ 服务器正在运行${NC}"
echo ""

# 测试各个Origin
echo "🧪 测试不同Origin的CORS配置..."
echo ""

for origin in "${TEST_ORIGINS[@]}"; do
    echo -e "${BLUE}测试Origin: ${origin}${NC}"
    
    # 测试OPTIONS预检请求
    echo "  📋 测试OPTIONS预检请求..."
    response=$(curl -s -I -X OPTIONS \
        -H "Origin: ${origin}" \
        -H "Access-Control-Request-Method: POST" \
        -H "Access-Control-Request-Headers: Content-Type,Authorization" \
        "${SERVER_URL}/api/admin/login" 2>&1)
    
    if echo "$response" | grep -q "Access-Control-Allow-Origin"; then
        allow_origin=$(echo "$response" | grep "Access-Control-Allow-Origin" | cut -d' ' -f2- | tr -d '\r')
        echo -e "    ${GREEN}✅ OPTIONS预检通过 - Allow-Origin: ${allow_origin}${NC}"
    else
        echo -e "    ${RED}❌ OPTIONS预检失败${NC}"
    fi
    
    # 测试GET请求
    echo "  📋 测试GET请求..."
    response=$(curl -s -I -H "Origin: ${origin}" "${SERVER_URL}/api/health" 2>&1)
    
    if echo "$response" | grep -q "Access-Control-Allow-Origin"; then
        allow_origin=$(echo "$response" | grep "Access-Control-Allow-Origin" | cut -d' ' -f2- | tr -d '\r')
        echo -e "    ${GREEN}✅ GET请求通过 - Allow-Origin: ${allow_origin}${NC}"
    else
        echo -e "    ${RED}❌ GET请求失败${NC}"
    fi
    
    echo ""
done

# 测试通配符支持
echo "🌟 测试通配符支持..."
echo ""

# 测试随机局域网IP
random_ip="http://192.168.$(shuf -i 1-254 -n 1).$(shuf -i 1-254 -n 1):3000"
echo -e "${BLUE}测试随机局域网IP: ${random_ip}${NC}"

response=$(curl -s -I -H "Origin: ${random_ip}" "${SERVER_URL}/api/health" 2>&1)
if echo "$response" | grep -q "Access-Control-Allow-Origin"; then
    allow_origin=$(echo "$response" | grep "Access-Control-Allow-Origin" | cut -d' ' -f2- | tr -d '\r')
    echo -e "  ${GREEN}✅ 通配符匹配成功 - Allow-Origin: ${allow_origin}${NC}"
else
    echo -e "  ${RED}❌ 通配符匹配失败${NC}"
fi
echo ""

# 显示环境变量
echo "🔧 当前环境变量配置:"
echo "ALLOWED_ORIGINS: ${ALLOWED_ORIGINS:-未设置}"
echo "GIN_MODE: ${GIN_MODE:-未设置}"
echo ""

# 建议
echo "💡 配置建议:"
echo "============"
echo "开发环境快速修复:"
echo "  export ALLOWED_ORIGINS='*'"
echo "  export GIN_MODE='debug'"
echo ""
echo "生产环境安全配置:"
echo "  export ALLOWED_ORIGINS='https://yourdomain.com,https://admin.yourdomain.com'"
echo "  export GIN_MODE='release'"
echo ""

# 验证当前配置
if [ "$ALLOWED_ORIGINS" = "*" ]; then
    echo -e "${GREEN}✅ 当前配置允许所有域名访问（开发环境推荐）${NC}"
elif [ -n "$ALLOWED_ORIGINS" ]; then
    echo -e "${YELLOW}⚠️  当前配置了特定域名: $ALLOWED_ORIGINS${NC}"
else
    echo -e "${YELLOW}⚠️  使用默认配置，可能需要设置环境变量${NC}"
fi

echo ""
echo "测试完成！" 