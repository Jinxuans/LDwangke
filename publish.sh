#!/bin/bash
set -e

# =========================================
#  青卡管理器 - 一键发布更新包
#  在 29.colnt.com 服务器上执行
#  用法: bash publish.sh [版本号] [更新日志]
#  示例: bash publish.sh 1.0.1 "修复订单同步问题"
# =========================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ROOT="/www/wwwroot/QK"
GO_DIR="$ROOT/go-api"
VBEN_DIR="$ROOT/vben-admin"
MALL_DIR="$ROOT/mall-h5"
PLUGIN_DIR="$ROOT/bt_plugin/qingka_manager"
PHP_DIR="$ROOT/php-api"
UPDATE_DIR="/var/www/admin/update"
TMP_DIR="/tmp/qingka_publish_$$"

VERSION="${1:-}"
CHANGELOG="${2:-无更新说明}"

if [ -z "$VERSION" ]; then
    # 自动从 version.json 递增
    if [ -f "$UPDATE_DIR/version.json" ]; then
        CURRENT=$(python3 -c "import json;print(json.load(open('$UPDATE_DIR/version.json'))['version'])" 2>/dev/null || echo "1.0.0")
        # 递增 patch 版本
        IFS='.' read -r major minor patch <<< "$CURRENT"
        patch=$((patch + 1))
        VERSION="$major.$minor.$patch"
    else
        VERSION="1.0.0"
    fi
fi

# ========== 构建函数 ==========
build_backend() {
    echo -e "${YELLOW}[Go 后端] 编译中...${NC}"
    cd "$GO_DIR"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$TMP_DIR/backend/server" ./cmd/server/
    cp -r config "$TMP_DIR/backend/" && rm -f "$TMP_DIR/backend/config/config.yaml"
    cp -r migrations "$TMP_DIR/backend/" 2>/dev/null || true
    mkdir -p "$TMP_DIR/backend/deploy"
    cp "$ROOT/deploy/init_db.sql" "$TMP_DIR/backend/deploy/"
    cp "$ROOT/deploy/reset_db.sql" "$TMP_DIR/backend/deploy/"
    cd "$TMP_DIR/backend" && tar -czf "$UPDATE_DIR/backend.tar.gz" .
    echo -e "${GREEN}✅ Go 后端编译完成${NC}"
}

build_frontend() {
    echo -e "${YELLOW}[管理前端] 构建中...${NC}"
    cd "$VBEN_DIR"
    pnpm install --frozen-lockfile 2>/dev/null || pnpm install
    pnpm build:antd
    cp -r apps/web-antd/dist/* "$TMP_DIR/frontend/"
    cd "$TMP_DIR/frontend" && tar -czf "$UPDATE_DIR/frontend.tar.gz" --exclude='favicon.ico' .
    echo -e "${GREEN}✅ 管理前端构建完成${NC}"
}

build_mall() {
    echo -e "${YELLOW}[商城H5] 构建中...${NC}"
    cd "$MALL_DIR"
    npm install 2>/dev/null
    npm run build
    cp -r dist/* "$TMP_DIR/mall/" 2>/dev/null || true
    cd "$TMP_DIR/mall" && tar -czf "$UPDATE_DIR/mall.tar.gz" --exclude='favicon.ico' .
    echo -e "${GREEN}✅ 商城前端构建完成${NC}"
}

build_php() {
    echo -e "${YELLOW}[PHP API] 打包中...${NC}"
    if [ -d "$PHP_DIR" ]; then
        cp -r "$PHP_DIR"/* "$TMP_DIR/php-api/"
        rm -f "$TMP_DIR/php-api/config.php"
        cd "$TMP_DIR/php-api" && tar -czf "$UPDATE_DIR/php-api.tar.gz" .
        echo -e "${GREEN}✅ PHP API 打包完成${NC}"
    else
        echo -e "${RED}⚠ PHP API 目录不存在，跳过${NC}"
    fi
}

build_plugin() {
    echo -e "${YELLOW}[宝塔插件] 打包中...${NC}"
    cd "$PLUGIN_DIR" && tar -czf "$UPDATE_DIR/plugin.tar.gz" .
    echo -e "${GREEN}✅ 宝塔插件打包完成${NC}"
}

show_sizes() {
    echo ""
    echo -e "${GREEN}📦 打包结果：${NC}"
    for f in backend frontend mall php-api plugin; do
        if [ -f "$UPDATE_DIR/$f.tar.gz" ]; then
            SIZE=$(du -sh "$UPDATE_DIR/$f.tar.gz" | awk '{print $1}')
            echo "  $f: $SIZE"
        fi
    done
}

# ========== 选择菜单 ==========
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}  青卡管理器 - 发布更新 v${VERSION}${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "请选择要打包的模块："
echo "  1) Go 后端"
echo "  2) 管理前端 (Vben Admin)"
echo "  3) 商城 H5"
echo "  4) PHP API"
echo "  5) 宝塔插件"
echo "  a) 全部打包"
echo ""
read -p "输入编号（多选用空格分隔，如 1 2 4）: " -a CHOICES

mkdir -p "$TMP_DIR"/{backend,frontend,mall,plugin,php-api}
mkdir -p "$UPDATE_DIR"

BUILD_ALL=false
for c in "${CHOICES[@]}"; do
    if [ "$c" = "a" ] || [ "$c" = "A" ] || [ "$c" = "all" ]; then
        BUILD_ALL=true
        break
    fi
done

if $BUILD_ALL; then
    build_backend
    build_frontend
    build_mall
    build_php
    build_plugin
else
    for c in "${CHOICES[@]}"; do
        case $c in
            1) build_backend ;;
            2) build_frontend ;;
            3) build_mall ;;
            4) build_php ;;
            5) build_plugin ;;
            *) echo -e "${RED}⚠ 未知选项: $c，跳过${NC}" ;;
        esac
    done
fi

show_sizes

# ========== 更新 version.json ==========
echo -e "${YELLOW}更新版本信息...${NC}"
TOTAL_SIZE=$(du -shc "$UPDATE_DIR"/*.tar.gz 2>/dev/null | tail -1 | awk '{print $1}')

cat > "$UPDATE_DIR/version.json" << EOF
{
    "version": "$VERSION",
    "changelog": "$CHANGELOG",
    "size": "$TOTAL_SIZE",
    "date": "$(date +%Y-%m-%d)"
}
EOF

echo -e "${GREEN}✅ 版本信息已更新${NC}"

# ========== 清理 ==========
rm -rf "$TMP_DIR"

echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}  ✅ 发布完成！v${VERSION}${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "更新源文件："
ls -lh "$UPDATE_DIR/"
echo ""
echo "客户端插件将从以下地址拉取更新："
echo "  https://29.colnt.com/update/version.json"
echo "  https://29.colnt.com/update/backend.tar.gz"
echo "  https://29.colnt.com/update/frontend.tar.gz"
echo "  https://29.colnt.com/update/mall.tar.gz"
echo "  https://29.colnt.com/update/php-api.tar.gz"
