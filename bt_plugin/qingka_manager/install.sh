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

# 复制模板目录
mkdir -p "$PLUGIN_DIR/templates"
cp -f "$SRC_DIR/templates/"*.tpl "$PLUGIN_DIR/templates/" 2>/dev/null

# 创建项目目录
mkdir -p /www/wwwroot/qingka/go-api/config
mkdir -p /www/wwwroot/qingka/php-api

# 写入 client_secret（与授权站 config.toml 中 client_secret 一致）
echo -n '3f48cd7beb7c6a492b0119c40f3caf114e23a3acb3d43365939c1325b8d6a72d' > /www/wwwroot/qingka/.client_secret
chmod 600 /www/wwwroot/qingka/.client_secret

echo "✅ 青卡管理器插件安装完成"
echo "请在宝塔面板 → 软件商店 → 第三方插件 中找到「青卡管理器」"
