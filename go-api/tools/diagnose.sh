#!/bin/bash

# Go API 诊断脚本
# 使用方法：./tools/diagnose.sh

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CONFIG_PATH="$PROJECT_ROOT/config/config.yaml"

yaml_value() {
    local section="$1"
    local key="$2"
    awk -v section="$section" -v key="$key" '
        $0 ~ "^" section ":" { in_section = 1; next }
        in_section && /^[^[:space:]]/ { in_section = 0 }
        in_section && $1 == key ":" {
            gsub(/"/, "", $2)
            print $2
            exit
        }
    ' "$CONFIG_PATH"
}

find_binary() {
    if [ -f "$PROJECT_ROOT/server" ]; then
        echo "$PROJECT_ROOT/server"
        return 0
    fi
    if [ -f "$PROJECT_ROOT/go-api-linux" ]; then
        echo "$PROJECT_ROOT/go-api-linux"
        return 0
    fi
    return 1
}

echo "========================================"
echo "  Go API 系统诊断工具"
echo "========================================"
echo "脚本目录：$SCRIPT_DIR"
echo "项目目录：$PROJECT_ROOT"
echo ""

echo "[1/7] 检查 Go API 可执行文件..."
BIN_PATH="$(find_binary)"
if [ -n "$BIN_PATH" ]; then
    echo "✓ 可执行文件存在：$BIN_PATH"
    ls -la "$BIN_PATH"
else
    echo "✗ 未找到 server 或 go-api-linux，需要先编译"
    echo "  编译命令：go build -o server ./cmd/server/"
fi
echo ""

echo "[2/7] 检查配置文件..."
if [ -f "$CONFIG_PATH" ]; then
    echo "✓ config.yaml 存在"
    echo "  JWT 配置:"
    grep -A3 "^jwt:" "$CONFIG_PATH" 2>/dev/null | sed 's/^/    /'
else
    echo "✗ config.yaml 不存在：$CONFIG_PATH"
fi
echo ""

echo "[3/7] 检查 8080 端口状态..."
if lsof -i:8080 >/dev/null 2>&1; then
    echo "⚠ 8080 端口被占用:"
    lsof -i:8080 | head -5 | sed 's/^/    /'
else
    echo "✓ 8080 端口空闲"
fi
echo ""

echo "[4/7] 检查运行中的进程..."
PROCS="$(pgrep -f "$PROJECT_ROOT/server|$PROJECT_ROOT/go-api-linux|go-api.*server" 2>/dev/null)"
if [ -n "$PROCS" ]; then
    echo "⚠ 发现相关进程:"
    ps aux | grep -E "go-api|server|go-api-linux" | grep -v grep | head -5 | sed 's/^/    /'
else
    echo "✓ 没有运行中的相关进程"
fi
echo ""

echo "[5/7] 检查日志文件..."
LOG_PATH="$PROJECT_ROOT/go-api.log"
if [ -f "$LOG_PATH" ]; then
    echo "✓ go-api.log 存在"
    echo "  最近 5 条日志:"
    tail -5 "$LOG_PATH" | sed 's/^/    /'
else
    echo "✗ go-api.log 不存在"
fi
echo ""

echo "[6/7] 检查数据库配置..."
DB_HOST="$(yaml_value database host)"
DB_USER="$(yaml_value database user)"
DB_NAME="$(yaml_value database dbname)"
echo "  数据库主机：$DB_HOST"
echo "  数据库用户：$DB_USER"
echo "  数据库名称：$DB_NAME"
echo ""

echo "[7/7] 检查 Redis 配置和公开接口..."
REDIS_HOST="$(yaml_value redis host)"
REDIS_PORT="$(yaml_value redis port)"
PORT="$(yaml_value server port)"
if [ -z "$PORT" ]; then
    PORT="8080"
fi
echo "  Redis 主机：$REDIS_HOST"
echo "  Redis 端口：$REDIS_PORT"
if command -v curl >/dev/null 2>&1; then
    HTTP_CODE="$(curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:$PORT/api/v1/site/config" 2>/dev/null)"
    echo "  公开接口 /api/v1/site/config 状态：${HTTP_CODE:-000}"
fi
echo ""

echo "========================================"
echo "  诊断完成"
echo "========================================"
echo ""
echo "建议操作:"
echo "1. 如果 8080 端口被占用，运行：./tools/restart.sh"
echo "2. 如果可执行文件不存在，运行：go build -o server ./cmd/server/"
echo "3. 如果服务未启动，运行：./tools/restart.sh"
