#!/bin/bash

# 登录问题诊断脚本
# 用于快速排查登录时出现的内部服务器错误

echo "=========================================="
echo "登录问题诊断工具"
echo "=========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查配置文件
echo "1. 检查配置文件..."
if [ -f "config/config.yaml" ]; then
    echo -e "${GREEN}✓${NC} 配置文件存在"
    
    # 检查数据库配置
    if grep -q "database:" config/config.yaml; then
        echo -e "${GREEN}✓${NC} 数据库配置存在"
    else
        echo -e "${RED}✗${NC} 数据库配置缺失"
    fi
    
    # 检查JWT配置
    if grep -q "jwt:" config/config.yaml; then
        echo -e "${GREEN}✓${NC} JWT配置存在"
    else
        echo -e "${RED}✗${NC} JWT配置缺失"
    fi
else
    echo -e "${RED}✗${NC} 配置文件不存在: config/config.yaml"
    echo "请从 config.example.yaml 复制并配置"
    exit 1
fi

echo ""

# 检查数据库连接
echo "2. 检查数据库连接..."

# 从配置文件读取数据库信息（简化版，实际可能需要 yq 工具）
DB_HOST=$(grep -A 5 "database:" config/config.yaml | grep "host:" | awk '{print $2}' | tr -d '"')
DB_PORT=$(grep -A 5 "database:" config/config.yaml | grep "port:" | awk '{print $2}')
DB_USER=$(grep -A 5 "database:" config/config.yaml | grep "username:" | awk '{print $2}' | tr -d '"')
DB_PASS=$(grep -A 5 "database:" config/config.yaml | grep "password:" | awk '{print $2}' | tr -d '"')
DB_NAME=$(grep -A 5 "database:" config/config.yaml | grep "database:" | awk '{print $2}' | tr -d '"')

if [ -z "$DB_HOST" ]; then
    echo -e "${YELLOW}⚠${NC} 无法从配置文件读取数据库信息，请手动检查"
else
    echo "数据库地址: $DB_HOST:$DB_PORT"
    echo "数据库名称: $DB_NAME"
    
    # 测试数据库连接
    if command -v mysql &> /dev/null; then
        if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" -e "USE $DB_NAME; SELECT 1;" &> /dev/null; then
            echo -e "${GREEN}✓${NC} 数据库连接成功"
        else
            echo -e "${RED}✗${NC} 数据库连接失败"
            echo "请检查数据库服务是否运行，以及配置是否正确"
        fi
    else
        echo -e "${YELLOW}⚠${NC} 未安装 mysql 客户端，跳过连接测试"
    fi
fi

echo ""

# 检查必要的数据库表
echo "3. 检查数据库表结构..."

if [ ! -z "$DB_HOST" ] && command -v mysql &> /dev/null; then
    # 检查 user 表
    if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" -D"$DB_NAME" -e "DESC qingka_wangke_user;" &> /dev/null; then
        echo -e "${GREEN}✓${NC} qingka_wangke_user 表存在"
        
        # 检查 pass2 字段
        if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" -D"$DB_NAME" -e "SHOW COLUMNS FROM qingka_wangke_user LIKE 'pass2';" 2>/dev/null | grep -q "pass2"; then
            echo -e "${GREEN}✓${NC} pass2 字段存在"
        else
            echo -e "${YELLOW}⚠${NC} pass2 字段不存在（可能导致登录失败）"
            echo "  修复命令: ALTER TABLE qingka_wangke_user ADD COLUMN pass2 VARCHAR(255) DEFAULT '';"
        fi
    else
        echo -e "${RED}✗${NC} qingka_wangke_user 表不存在"
    fi
    
    # 检查 config 表
    if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" -D"$DB_NAME" -e "DESC qingka_wangke_config;" &> /dev/null; then
        echo -e "${GREEN}✓${NC} qingka_wangke_config 表存在"
        
        # 检查 pass2_kg 配置
        PASS2_KG=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" -D"$DB_NAME" -se "SELECT k FROM qingka_wangke_config WHERE v='pass2_kg';" 2>/dev/null)
        if [ ! -z "$PASS2_KG" ]; then
            echo -e "${GREEN}✓${NC} pass2_kg 配置存在 (值: $PASS2_KG)"
        else
            echo -e "${YELLOW}⚠${NC} pass2_kg 配置不存在"
            echo "  修复命令: INSERT INTO qingka_wangke_config (v, k) VALUES ('pass2_kg', '0');"
        fi
    else
        echo -e "${RED}✗${NC} qingka_wangke_config 表不存在"
    fi
fi

echo ""

# 检查 Redis 连接
echo "4. 检查 Redis 连接..."

REDIS_ADDR=$(grep -A 3 "redis:" config/config.yaml | grep "addr:" | awk '{print $2}' | tr -d '"')

if [ ! -z "$REDIS_ADDR" ]; then
    echo "Redis 地址: $REDIS_ADDR"
    
    if command -v redis-cli &> /dev/null; then
        REDIS_HOST=$(echo $REDIS_ADDR | cut -d':' -f1)
        REDIS_PORT=$(echo $REDIS_ADDR | cut -d':' -f2)
        
        if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" ping &> /dev/null; then
            echo -e "${GREEN}✓${NC} Redis 连接成功"
        else
            echo -e "${YELLOW}⚠${NC} Redis 连接失败（不影响基本登录功能）"
        fi
    else
        echo -e "${YELLOW}⚠${NC} 未安装 redis-cli，跳过连接测试"
    fi
else
    echo -e "${YELLOW}⚠${NC} 未配置 Redis"
fi

echo ""

# 检查服务进程
echo "5. 检查服务状态..."

if pgrep -f "go-api" > /dev/null; then
    echo -e "${GREEN}✓${NC} go-api 服务正在运行"
    
    # 检查端口
    PORT=$(grep -A 2 "server:" config/config.yaml | grep "port:" | awk '{print $2}')
    if [ ! -z "$PORT" ]; then
        if netstat -tuln 2>/dev/null | grep -q ":$PORT " || ss -tuln 2>/dev/null | grep -q ":$PORT "; then
            echo -e "${GREEN}✓${NC} 服务端口 $PORT 正在监听"
        else
            echo -e "${RED}✗${NC} 服务端口 $PORT 未监听"
        fi
    fi
else
    echo -e "${RED}✗${NC} go-api 服务未运行"
    echo "启动命令: ./go-api 或 ./restart.sh"
fi

echo ""

# 检查日志文件
echo "6. 检查最近的错误日志..."

if [ -f "logs/app.log" ]; then
    echo "最近的错误日志（最后10条包含 error 的记录）:"
    echo "----------------------------------------"
    grep -i "error" logs/app.log | tail -10 || echo "未发现错误日志"
    echo "----------------------------------------"
elif [ -f "nohup.out" ]; then
    echo "最近的错误日志（nohup.out 最后10条包含 error 的记录）:"
    echo "----------------------------------------"
    grep -i "error" nohup.out | tail -10 || echo "未发现错误日志"
    echo "----------------------------------------"
else
    echo -e "${YELLOW}⚠${NC} 未找到日志文件"
fi

echo ""

# 测试登录接口
echo "7. 测试登录接口..."

PORT=$(grep -A 2 "server:" config/config.yaml | grep "port:" | awk '{print $2}')
if [ -z "$PORT" ]; then
    PORT=8080
fi

if command -v curl &> /dev/null; then
    echo "测试 API 健康检查..."
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$PORT/health 2>/dev/null)
    
    if [ "$HTTP_CODE" = "200" ]; then
        echo -e "${GREEN}✓${NC} API 服务响应正常"
    elif [ "$HTTP_CODE" = "000" ]; then
        echo -e "${RED}✗${NC} 无法连接到 API 服务"
    else
        echo -e "${YELLOW}⚠${NC} API 返回状态码: $HTTP_CODE"
    fi
else
    echo -e "${YELLOW}⚠${NC} 未安装 curl，跳过接口测试"
fi

echo ""
echo "=========================================="
echo "诊断完成"
echo "=========================================="
echo ""
echo "常见问题修复："
echo "1. 数据库连接失败 -> 检查数据库服务和配置"
echo "2. pass2 字段缺失 -> 执行 SQL: ALTER TABLE qingka_wangke_user ADD COLUMN pass2 VARCHAR(255) DEFAULT '';"
echo "3. 配置表缺失 -> 执行 SQL: INSERT INTO qingka_wangke_config (v, k) VALUES ('pass2_kg', '0');"
echo "4. JWT 配置问题 -> 检查 config.yaml 中的 jwt.secret"
echo "5. 服务未启动 -> 执行 ./restart.sh 或 ./go-api"
echo ""
echo "详细诊断报告请查看: LOGIN_ERROR_DIAGNOSIS.md"
