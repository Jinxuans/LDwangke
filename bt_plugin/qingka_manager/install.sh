#!/bin/bash
echo "========================================="
echo "  安装青卡管理器插件"
echo "========================================="

PLUGIN_DIR="/www/server/panel/plugin/qingka_manager"
SRC_DIR=$(cd "$(dirname "$0")" && pwd)

# 创建插件目录
mkdir -p "$PLUGIN_DIR"

# 复制插件文件
cp -f "$SRC_DIR/info.json" "$PLUGIN_DIR/"
cp -f "$SRC_DIR/qingka_manager_main.py" "$PLUGIN_DIR/"
cp -f "$SRC_DIR/index.html" "$PLUGIN_DIR/"
cp -f "$SRC_DIR/__init__.py" "$PLUGIN_DIR/"
cp -f "$SRC_DIR/uninstall.sh" "$PLUGIN_DIR/"

# 创建项目目录
mkdir -p /www/wwwroot/qingka/go-api/config
mkdir -p /var/www/admin
mkdir -p /var/www/mall

echo "✅ 青卡管理器插件安装完成"
echo "请在宝塔面板 → 软件商店 → 第三方插件 中找到「青卡管理器」"
