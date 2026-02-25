#!/bin/bash

# Go API 服务重启脚本
# 使用方法：./restart.sh

cd "$(dirname "$0")"

echo "=== Go API 服务重启脚本 ==="
echo "当前目录：$(pwd)"
echo ""

# 1. 查找并停止占用 8080 端口的进程
echo "[1/4] 正在停止占用 8080 端口的进程..."
PID=$(lsof -t -i:8080 2>/dev/null)
if [ -n "$PID" ]; then
    kill -9 $PID 2>/dev/null
    echo "已停止进程 PID: $PID"
    sleep 1
else
    echo "8080 端口未被占用"
fi

# 2. 查找并停止旧的 server 进程
echo ""
echo "[2/4] 正在检查旧的 server 进程..."
OLD_PID=$(pgrep -f "go-api.*server" 2>/dev/null)
if [ -n "$OLD_PID" ]; then
    kill -9 $OLD_PID 2>/dev/null
    echo "已停止 server 进程 PID: $OLD_PID"
else
    echo "未发现运行中的 server 进程"
fi

sleep 2

# 3. 验证端口已释放
echo ""
echo "[3/4] 验证 8080 端口状态..."
if lsof -i:8080 >/dev/null 2>&1; then
    echo "错误：8080 端口仍然被占用，请手动检查"
    lsof -i:8080
    exit 1
else
    echo "8080 端口已释放"
fi

# 4. 启动新服务
echo ""
echo "[4/4] 启动 Go API 服务..."
if [ -f "./server" ]; then
    nohup ./server > go-api.log 2>&1 &
    NEW_PID=$!
    echo "服务已启动，PID: $NEW_PID"

    # 等待 3 秒检查启动状态
    sleep 3
    if ps -p $NEW_PID > /dev/null 2>&1; then
        echo ""
        echo "=== 服务启动成功 ==="
        echo "日志文件：$(pwd)/go-api.log"
        echo ""
        echo "最近日志:"
        tail -10 go-api.log
    else
        echo ""
        echo "=== 服务启动失败 ==="
        echo "请查看日志:"
        tail -20 go-api.log
        exit 1
    fi
else
    echo "错误：未找到 server 可执行文件"
    echo "请先编译：go build -o server ./cmd/server/"
    exit 1
fi

echo ""
echo "=== 重启完成 ==="
