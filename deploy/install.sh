#!/bin/bash
set -e

# ========================================
#   一键部署脚本（服务器端执行）
#   适用于 Ubuntu + 宝塔面板
# ========================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

DEPLOY_DIR=$(cd "$(dirname "$0")" && pwd)
SITE_DIR="/www/wwwroot"
GO_DIR="$SITE_DIR/go-api"
DIST_DIR="$SITE_DIR/dist"
PHP_DIR="$SITE_DIR/php-api"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  一键部署安装脚本${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# ========== 交互式配置 ==========
read -p "MySQL 用户名 [root]: " DB_USER
DB_USER=${DB_USER:-root}

read -sp "MySQL 密码: " DB_PASS
echo ""

read -p "数据库名 [7777]: " DB_NAME
DB_NAME=${DB_NAME:-7777}

read -p "Redis 密码 (无密码直接回车): " REDIS_PASS

# 生成随机密钥
JWT_SECRET=$(openssl rand -hex 32)
BRIDGE_SECRET=$(openssl rand -hex 16)

echo ""
echo -e "${YELLOW}配置信息：${NC}"
echo "  MySQL: $DB_USER@127.0.0.1 / $DB_NAME"
echo "  JWT Secret: ${JWT_SECRET:0:16}..."
echo "  Bridge Secret: ${BRIDGE_SECRET:0:8}..."
echo ""
read -p "确认开始部署？(y/N): " CONFIRM
if [ "$CONFIRM" != "y" ] && [ "$CONFIRM" != "Y" ]; then
    echo "已取消"
    exit 0
fi

# ========== 1. 检查环境 ==========
echo ""
echo -e "${GREEN}[1/7] 检查环境...${NC}"

check_cmd() {
    if ! command -v "$1" &>/dev/null; then
        echo -e "${RED}❌ 未找到 $1，请先在宝塔面板安装${NC}"
        exit 1
    fi
}

check_cmd nginx
check_cmd mysql
check_cmd redis-cli
echo "✅ Nginx、MySQL、Redis 已安装"

# ========== 2. 创建数据库 ==========
echo ""
echo -e "${GREEN}[2/7] 初始化数据库...${NC}"

mysql -u"$DB_USER" -p"$DB_PASS" -e "CREATE DATABASE IF NOT EXISTS \`$DB_NAME\` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;" 2>/dev/null
echo "✅ 数据库 $DB_NAME 已就绪"

# 执行迁移脚本
if [ -d "$DEPLOY_DIR/go-api/migrations" ]; then
    echo "  执行迁移脚本..."
    for sql in $(ls "$DEPLOY_DIR/go-api/migrations/"*.sql 2>/dev/null | sort); do
        echo "    → $(basename $sql)"
        mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" < "$sql" 2>/dev/null || true
    done
    echo "✅ 迁移脚本执行完成"
fi

# ========== 3. 部署 Go 后端 ==========
echo ""
echo -e "${GREEN}[3/7] 部署 Go 后端...${NC}"

mkdir -p "$GO_DIR/config"
cp "$DEPLOY_DIR/go-api/go-api-linux" "$GO_DIR/"
chmod +x "$GO_DIR/go-api-linux"

# 生成配置文件
cat > "$GO_DIR/config/config.yaml" << EOF
server:
  port: 8080
  mode: release
  php_backend: "http://127.0.0.1:9000"
  php_public_url: ""
  bridge_secret: "$BRIDGE_SECRET"

database:
  host: 127.0.0.1
  port: 3306
  user: $DB_USER
  password: "$DB_PASS"
  dbname: "$DB_NAME"
  max_open_conns: 50
  max_idle_conns: 25

redis:
  host: 127.0.0.1
  port: 6379
  password: "$REDIS_PASS"
  db: 0

jwt:
  secret: "$JWT_SECRET"
  access_ttl: 7200
  refresh_ttl: 604800

cache:
  order_list_ttl: 30
  class_list_ttl: 300

smtp:
  host: "smtp.qq.com"
  port: 465
  user: ""
  password: ""
  from_name: "系统通知"
  encryption: "ssl"
EOF

echo "✅ Go 后端已部署到 $GO_DIR"

# ========== 4. 部署 PHP API ==========
echo ""
echo -e "${GREEN}[4/7] 部署 PHP API...${NC}"

if [ -d "$DEPLOY_DIR/php-api" ]; then
    mkdir -p "$PHP_DIR"
    cp -r "$DEPLOY_DIR/php-api/"* "$PHP_DIR/"

    # 更新 PHP 配置
    cat > "$PHP_DIR/config.php" << 'PHPEOF'
<?php
return [
    'database' => [
        'host'     => '127.0.0.1',
        'port'     => 3306,
PHPEOF

    # 追加动态配置
    cat >> "$PHP_DIR/config.php" << EOF
        'user'     => '$DB_USER',
        'password' => '$DB_PASS',
        'dbname'   => '$DB_NAME',
        'charset'  => 'utf8mb4',
    ],
    'redis' => [
        'host'     => '127.0.0.1',
        'port'     => 6379,
        'password' => '$REDIS_PASS',
        'db'       => 0,
    ],
    'jwt' => [
        'secret'      => '$JWT_SECRET',
        'access_ttl'  => 7200,
        'refresh_ttl' => 604800,
    ],
    'payment' => [
        'epay_api' => '',
        'epay_pid' => '',
        'epay_key' => '',
    ],
    'app' => [
        'debug'    => false,
        'timezone' => 'Asia/Shanghai',
    ],
];
EOF

    echo "✅ PHP API 已部署到 $PHP_DIR"
else
    echo "⚠️  未找到 php-api 目录，跳过"
fi

# ========== 5. 部署前端 ==========
echo ""
echo -e "${GREEN}[5/7] 部署前端...${NC}"

if [ -d "$DEPLOY_DIR/dist" ]; then
    mkdir -p "$DIST_DIR"
    cp -r "$DEPLOY_DIR/dist/"* "$DIST_DIR/"
    echo "✅ 前端已部署到 $DIST_DIR"
else
    echo -e "${RED}❌ 未找到 dist 目录，请先在本地构建前端${NC}"
fi

# ========== 6. 配置 Nginx ==========
echo ""
echo -e "${GREEN}[6/7] 配置 Nginx...${NC}"

cat > /www/server/panel/vhost/nginx/site.conf << 'NGINXEOF'
server {
    listen 80;
    server_name _;

    # Vue 前端
    location / {
        root /www/wwwroot/dist;
        try_files $uri $uri/ /index.html;
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff2?)$ {
            expires 30d;
            add_header Cache-Control "public, immutable";
        }
    }

    # Go API
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 10s;
        proxy_read_timeout 30s;
    }

    # WebSocket
    location /ws/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 86400s;
    }

    # PHP API
    location /php-api/ {
        proxy_pass http://127.0.0.1:9000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    location ~ \.php$ {
        proxy_pass http://127.0.0.1:9000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location ~ ^/(confing|assets)/ {
        proxy_pass http://127.0.0.1:9000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location ~ /\. { deny all; }
    client_max_body_size 50m;

    access_log /www/wwwlogs/site.log;
    error_log /www/wwwlogs/site.error.log;
}
NGINXEOF

nginx -t && systemctl reload nginx
echo "✅ Nginx 配置完成"

# ========== 7. 启动 Go 后端 ==========
echo ""
echo -e "${GREEN}[7/7] 启动 Go 后端服务...${NC}"

# 停止旧进程
systemctl stop go-api 2>/dev/null || true
pkill -f go-api-linux 2>/dev/null || true

# 创建 systemd 服务
cat > /etc/systemd/system/go-api.service << EOF
[Unit]
Description=Go API Server
After=network.target mysql.service redis.service

[Service]
Type=simple
User=www
WorkingDirectory=$GO_DIR
ExecStart=$GO_DIR/go-api-linux
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable go-api
systemctl start go-api

# 等待启动
sleep 2
if systemctl is-active --quiet go-api; then
    echo "✅ Go 后端已启动"
else
    echo -e "${RED}❌ Go 后端启动失败，查看日志: journalctl -u go-api -f${NC}"
fi

# ========== 完成 ==========
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  ✅ 部署完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 获取公网 IP
PUBLIC_IP=$(curl -s ifconfig.me 2>/dev/null || curl -s ip.sb 2>/dev/null || echo "你的服务器IP")

echo "访问地址："
echo "  前端页面: http://$PUBLIC_IP/"
echo "  API 测试: http://$PUBLIC_IP/api/v1/site/config"
echo ""
echo "管理命令："
echo "  查看状态: systemctl status go-api"
echo "  重启后端: systemctl restart go-api"
echo "  查看日志: journalctl -u go-api -f"
echo ""
echo "配置文件位置："
echo "  Go 配置: $GO_DIR/config/config.yaml"
echo "  PHP 配置: $PHP_DIR/config.php"
echo "  Nginx: /www/server/panel/vhost/nginx/site.conf"
echo ""
