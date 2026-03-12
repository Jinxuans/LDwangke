#!/bin/bash

# Go API 服务重启脚本
# 使用方法：./tools/restart.sh

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

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

BIN_PATH="$(find_binary)"
LOG_PATH="$PROJECT_ROOT/go-api.log"

echo "=== Go API 服务重启脚本 ==="
echo "脚本目录：$SCRIPT_DIR"
echo "项目目录：$PROJECT_ROOT"
echo ""

echo "[1/4] 正在停止占用 8080 端口的进程..."
PID="$(lsof -t -i:8080 2>/dev/null)"
if [ -n "$PID" ]; then
    kill -9 "$PID" 2>/dev/null
    echo "已停止进程 PID: $PID"
    sleep 1
else
    echo "8080 端口未被占用"
fi

echo ""
echo "[2/4] 正在检查旧的 Go API 进程..."
OLD_PID="$(pgrep -f "$PROJECT_ROOT/server|$PROJECT_ROOT/go-api-linux|go-api.*server" 2>/dev/null)"
if [ -n "$OLD_PID" ]; then
    kill -9 $OLD_PID 2>/dev/null
    echo "已停止进程 PID: $OLD_PID"
else
    echo "未发现运行中的 Go API 进程"
fi

sleep 2

echo ""
echo "[3/4] 验证 8080 端口状态..."
if lsof -i:8080 >/dev/null 2>&1; then
    echo "错误：8080 端口仍然被占用，请手动检查"
    lsof -i:8080
    exit 1
else
    echo "8080 端口已释放"
fi

echo ""
echo "[4/4] 启动 Go API 服务..."
if [ -z "$BIN_PATH" ]; then
    echo "错误：未找到可执行文件 server 或 go-api-linux"
    echo "请先编译：go build -o server ./cmd/server/"
    exit 1
fi

cd "$PROJECT_ROOT"
nohup "$BIN_PATH" > "$LOG_PATH" 2>&1 &
NEW_PID=$!
echo "服务已启动，PID: $NEW_PID"

sleep 3
if ps -p "$NEW_PID" >/dev/null 2>&1; then
    echo ""
    echo "=== 服务启动成功 ==="
    echo "可执行文件：$BIN_PATH"
    echo "日志文件：$LOG_PATH"
    echo ""
    echo "最近日志:"
    tail -10 "$LOG_PATH"
else
    echo ""
    echo "=== 服务启动失败 ==="
    if [ -f "$LOG_PATH" ]; then
        tail -20 "$LOG_PATH"
    fi
    exit 1
fi

echo ""
echo "=== 重启完成 ==="
