#!/bin/bash

# Go API 诊断脚本
# 使用方法：./diagnose.sh

cd "$(dirname "$0")"

echo "========================================"
echo "  Go API 系统诊断工具"
echo "========================================"
echo ""

# 1. 检查 server 可执行文件
echo "[1/7] 检查 server 可执行文件..."
if [ -f "./server" ]; then
    echo "✓ server 文件存在"
    ls -la ./server
else
    echo "✗ server 文件不存在，需要先编译"
    echo "  编译命令：go build -o server ./cmd/server/"
fi
echo ""

# 2. 检查配置文件
echo "[2/7] 检查配置文件..."
if [ -f "./config/config.yaml" ]; then
    echo "✓ config.yaml 存在"
    echo "  JWT 配置:"
    grep -A3 "^jwt:" config/config.yaml 2>/dev/null | sed 's/^/    /'
else
    echo "✗ config.yaml 不存在"
fi
echo ""

# 3. 检查端口占用
echo "[3/7] 检查 8080 端口状态..."
if lsof -i:8080 >/dev/null 2>&1; then
    echo "⚠ 8080 端口被占用:"
    lsof -i:8080 | head -5 | sed 's/^/    /'
else
    echo "✓ 8080 端口空闲"
fi
echo ""

# 4. 检查运行中的进程
echo "[4/7] 检查运行中的进程..."
PROCS=$(pgrep -f "go-api\|server" 2>/dev/null)
if [ -n "$PROCS" ]; then
    echo "⚠ 发现相关进程:"
    ps aux | grep -E "go-api|server" | grep -v grep | head -5 | sed 's/^/    /'
else
    echo "✓ 没有运行中的相关进程"
fi
echo ""

# 5. 检查日志文件
echo "[5/7] 检查日志文件..."
if [ -f "./go-api.log" ]; then
    echo "✓ go-api.log 存在"
    echo "  最近 5 条日志:"
    tail -5 go-api.log | sed 's/^/    /'
else
    echo "✗ go-api.log 不存在"
fi
echo ""

# 6. 检查数据库连接
echo "[6/7] 检查数据库配置..."
DB_HOST=$(grep "host:" config/config.yaml | head -1 | awk '{print $2}')
DB_USER=$(grep "user:" config/config.yaml | head -1 | awk '{print $2}')
DB_NAME=$(grep "dbname:" config/config.yaml | awk '{print $2}' | tr -d '"')
echo "  数据库主机：$DB_HOST"
echo "  数据库用户：$DB_USER"
echo "  数据库名称：$DB_NAME"
echo ""

# 7. 检查 Redis 配置
echo "[7/7] 检查 Redis 配置..."
REDIS_HOST=$(grep -A5 "^redis:" config/config.yaml | grep "host:" | awk '{print $2}')
REDIS_PORT=$(grep -A5 "^redis:" config/config.yaml | grep "port:" | awk '{print $2}')
echo "  Redis 主机：$REDIS_HOST"
echo "  Redis 端口：$REDIS_PORT"
echo ""

echo "========================================"
echo "  诊断完成"
echo "========================================"
echo ""
echo "建议操作:"
echo "1. 如果 8080 端口被占用，运行：./restart.sh"
echo "2. 如果 server 文件不存在，运行：go build -o server ./cmd/server/"
echo "3. 如果服务未启动，运行：./restart.sh"
