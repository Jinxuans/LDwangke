#!/bin/bash

# 数据库迁移功能测试脚本
# 用于测试 db_compat.go 的 Check() 和 Fix() 功能

echo "=========================================="
echo "数据库迁移功能测试"
echo "=========================================="
echo ""

# 配置
API_BASE="http://localhost:8080"
ADMIN_TOKEN=""

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查 API 是否运行
echo "1. 检查 API 服务状态..."
if curl -s -o /dev/null -w "%{http_code}" "$API_BASE/health" | grep -q "200"; then
    echo -e "${GREEN}✓ API 服务正常运行${NC}"
else
    echo -e "${RED}✗ API 服务未运行，请先启动服务${NC}"
    exit 1
fi
echo ""

# 获取管理员 Token（需要先登录）
echo "2. 获取管理员 Token..."
echo -e "${YELLOW}请输入管理员用户名（默认: admin）:${NC}"
read -r USERNAME
USERNAME=${USERNAME:-admin}

echo -e "${YELLOW}请输入管理员密码（默认: admin123）:${NC}"
read -rs PASSWORD
PASSWORD=${PASSWORD:-admin123}
echo ""

LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/api/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

ADMIN_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$ADMIN_TOKEN" ]; then
    echo -e "${RED}✗ 登录失败，无法获取 Token${NC}"
    echo "响应: $LOGIN_RESPONSE"
    echo ""
    echo -e "${YELLOW}提示: 如果管理员账号不存在，可以直接运行 Fix 接口创建${NC}"
    exit 1
fi

echo -e "${GREEN}✓ 登录成功，Token: ${ADMIN_TOKEN:0:20}...${NC}"
echo ""

# 执行数据库结构检查
echo "3. 执行数据库结构检查 (Check)..."
CHECK_RESPONSE=$(curl -s -X GET "$API_BASE/api/admin/db-compat/check" \
  -H "Authorization: Bearer $ADMIN_TOKEN")

echo "$CHECK_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$CHECK_RESPONSE"
echo ""

# 判断是否需要修复
MISSING_TABLES=$(echo "$CHECK_RESPONSE" | grep -o '"missing_tables":\[[^]]*\]' | grep -o '\[.*\]')
MISSING_COLUMNS=$(echo "$CHECK_RESPONSE" | grep -o '"missing_columns":\[[^]]*\]' | grep -o '\[.*\]')

if [ "$MISSING_TABLES" != "[]" ] || [ "$MISSING_COLUMNS" != "[]" ]; then
    echo -e "${YELLOW}⚠ 发现缺失的表或列${NC}"
    echo ""
    
    echo "4. 是否执行自动修复 (Fix)? [y/N]"
    read -r CONFIRM
    
    if [ "$CONFIRM" = "y" ] || [ "$CONFIRM" = "Y" ]; then
        echo ""
        echo "执行数据库结构修复..."
        FIX_RESPONSE=$(curl -s -X POST "$API_BASE/api/admin/db-compat/fix" \
          -H "Authorization: Bearer $ADMIN_TOKEN")
        
        echo "$FIX_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$FIX_RESPONSE"
        echo ""
        
        # 再次检查
        echo "5. 修复后再次检查..."
        CHECK_RESPONSE2=$(curl -s -X GET "$API_BASE/api/admin/db-compat/check" \
          -H "Authorization: Bearer $ADMIN_TOKEN")
        
        echo "$CHECK_RESPONSE2" | python3 -m json.tool 2>/dev/null || echo "$CHECK_RESPONSE2"
        echo ""
        
        echo -e "${GREEN}✓ 数据库迁移功能测试完成${NC}"
    else
        echo -e "${YELLOW}已取消修复操作${NC}"
    fi
else
    echo -e "${GREEN}✓ 数据库结构完整，无需修复${NC}"
fi

echo ""
echo "=========================================="
echo "测试完成"
echo "=========================================="
