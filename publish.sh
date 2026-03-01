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

ROOT="/www/wwwroot/29.colnt.com"
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

echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}  青卡管理器 - 发布更新 v${VERSION}${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""

mkdir -p "$TMP_DIR"/{backend,frontend,mall,plugin,php-api}
mkdir -p "$UPDATE_DIR"

# ========== 1. 编译 Go 后端 ==========
echo -e "${YELLOW}[1/6] 编译 Go 后端...${NC}"
cd "$GO_DIR"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$TMP_DIR/backend/server" ./cmd/server/
cp -r config "$TMP_DIR/backend/" && rm -f "$TMP_DIR/backend/config/config.yaml"
cp -r migrations "$TMP_DIR/backend/" 2>/dev/null || true
# 打入建表和初始化 SQL
mkdir -p "$TMP_DIR/backend/deploy"
cp "$ROOT/deploy/init_db.sql" "$TMP_DIR/backend/deploy/"
cp "$ROOT/deploy/reset_db.sql" "$TMP_DIR/backend/deploy/"
echo -e "${GREEN}✅ Go 后端编译完成${NC}"

# ========== 2. 构建 Vben Admin 前端 ==========
echo -e "${YELLOW}[2/6] 构建管理前端...${NC}"
cd "$VBEN_DIR"
pnpm install --frozen-lockfile 2>/dev/null || pnpm install
pnpm build:antd
cp -r apps/web-antd/dist/* "$TMP_DIR/frontend/"
echo -e "${GREEN}✅ 管理前端构建完成${NC}"

# ========== 3. 构建 Mall H5 ==========
echo -e "${YELLOW}[3/6] 构建商城前端...${NC}"
cd "$MALL_DIR"
npm install 2>/dev/null
npm run build
cp -r dist/* "$TMP_DIR/mall/" 2>/dev/null || true
echo -e "${GREEN}✅ 商城前端构建完成${NC}"

# ========== 4. 打包 PHP API ==========
echo -e "${YELLOW}[4/6] 打包 PHP API...${NC}"
if [ -d "$PHP_DIR" ]; then
    cp -r "$PHP_DIR"/* "$TMP_DIR/php-api/"
    # 移除运行时配置（由插件安装时生成）
    rm -f "$TMP_DIR/php-api/config.php"
    echo -e "${GREEN}✅ PHP API 打包完成${NC}"
else
    echo -e "${RED}⚠ PHP API 目录不存在，跳过${NC}"
fi

# ========== 5. 打包 tar.gz ==========
echo -e "${YELLOW}[5/6] 打包更新文件...${NC}"

cd "$TMP_DIR/backend" && tar -czf "$UPDATE_DIR/backend.tar.gz" .
cd "$TMP_DIR/frontend" && tar -czf "$UPDATE_DIR/frontend.tar.gz" --exclude='favicon.ico' .
cd "$TMP_DIR/mall" && tar -czf "$UPDATE_DIR/mall.tar.gz" --exclude='favicon.ico' .
cd "$TMP_DIR/php-api" && tar -czf "$UPDATE_DIR/php-api.tar.gz" . 2>/dev/null || true

# 打包插件本身
cd "$PLUGIN_DIR" && tar -czf "$UPDATE_DIR/plugin.tar.gz" .

# 计算总大小
TOTAL_SIZE=$(du -sh "$UPDATE_DIR"/*.tar.gz | awk '{sum+=$1} END{printf "%.1fM", sum}')
BACKEND_SIZE=$(du -sh "$UPDATE_DIR/backend.tar.gz" | awk '{print $1}')
FRONTEND_SIZE=$(du -sh "$UPDATE_DIR/frontend.tar.gz" | awk '{print $1}')
MALL_SIZE=$(du -sh "$UPDATE_DIR/mall.tar.gz" | awk '{print $1}')
PHP_SIZE=$(du -sh "$UPDATE_DIR/php-api.tar.gz" 2>/dev/null | awk '{print $1}')

echo -e "${GREEN}✅ 打包完成${NC}"
echo "  后端: $BACKEND_SIZE | 前端: $FRONTEND_SIZE | 商城: $MALL_SIZE | PHP: $PHP_SIZE"

# ========== 5. 更新 version.json ==========
echo -e "${YELLOW}[6/6] 更新版本信息...${NC}"

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
