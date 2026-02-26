#!/bin/bash
set -e

# ===== 授权站一键部署脚本 =====

APP_NAME="license-server"
APP_DIR=$(cd "$(dirname "$0")" && pwd)
BIN_PATH="$APP_DIR/target/release/$APP_NAME"
SERVICE_FILE="/etc/systemd/system/$APP_NAME.service"

echo "===== 授权站部署 ====="
echo "项目目录: $APP_DIR"

# 1. 检查/安装 Rust
if ! command -v cargo &> /dev/null; then
    echo "[1/5] 安装 Rust 工具链..."
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
    source "$HOME/.cargo/env"
else
    echo "[1/5] Rust 已安装: $(rustc --version)"
fi

# 2. 检查 gcc（rusqlite bundled 需要）
if ! command -v gcc &> /dev/null; then
    echo "[2/5] 安装 gcc..."
    if command -v apt-get &> /dev/null; then
        apt-get update -qq && apt-get install -y -qq build-essential
    elif command -v yum &> /dev/null; then
        yum install -y gcc gcc-c++ make
    elif command -v dnf &> /dev/null; then
        dnf install -y gcc gcc-c++ make
    else
        echo "错误: 无法自动安装 gcc，请手动安装"
        exit 1
    fi
else
    echo "[2/5] gcc 已安装: $(gcc --version | head -1)"
fi

# 3. 编译
echo "[3/5] 编译 Release 版本（首次编译约3-5分钟）..."
cd "$APP_DIR"
cargo build --release
echo "编译完成: $BIN_PATH"
echo "二进制大小: $(du -h "$BIN_PATH" | cut -f1)"

# 4. 检查配置
if [ ! -f "$APP_DIR/config.toml" ]; then
    echo "错误: 缺少 config.toml"
    exit 1
fi

# 提醒修改默认密钥
if grep -q "change-me-to-a-strong-secret-token" "$APP_DIR/config.toml"; then
    echo ""
    echo "⚠️  警告: config.toml 中的 admin_token 仍为默认值！"
    echo "⚠️  请修改 config.toml 中的 admin_token 和 client_secret！"
    echo ""
    read -p "是否继续部署？(y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "已取消，请先修改 config.toml"
        exit 0
    fi
fi

# 5. 创建 systemd 服务
echo "[4/5] 配置 systemd 服务..."

cat > "$SERVICE_FILE" << EOF
[Unit]
Description=License Server (授权站)
After=network.target

[Service]
Type=simple
WorkingDirectory=$APP_DIR
ExecStart=$BIN_PATH
Restart=always
RestartSec=5
Environment=RUST_LOG=info

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload

# 6. 启动服务
echo "[5/5] 启动服务..."
systemctl enable "$APP_NAME"
systemctl restart "$APP_NAME"

sleep 2

if systemctl is-active --quiet "$APP_NAME"; then
    PORT=$(grep -oP 'port\s*=\s*\K\d+' "$APP_DIR/config.toml" || echo "9800")
    echo ""
    echo "===== 部署成功 ====="
    echo "服务状态: 运行中"
    echo "监听端口: $PORT"
    echo "管理面板: http://服务器IP:$PORT/api/v1/admin/license/dashboard"
    echo ""
    echo "常用命令:"
    echo "  systemctl status $APP_NAME    # 查看状态"
    echo "  systemctl restart $APP_NAME   # 重启"
    echo "  journalctl -u $APP_NAME -f    # 查看日志"
    echo ""
    echo "测试接口:"
    echo "  curl http://127.0.0.1:$PORT/api/v1/admin/license/dashboard \\"
    echo "    -H 'Authorization: Bearer 你的admin_token'"
else
    echo "启动失败！查看日志:"
    echo "  journalctl -u $APP_NAME -n 50"
    exit 1
fi
