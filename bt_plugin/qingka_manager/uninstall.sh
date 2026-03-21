#!/bin/bash
echo "========================================="
echo "  卸载青卡管理器插件"
echo "========================================="

# 停止 Go API 服务
PID_FILE="/www/wwwroot/qingka/go-api/go-api.pid"
if [ -f "$PID_FILE" ]; then
    PID=$(cat "$PID_FILE")
    kill "$PID" 2>/dev/null
    sleep 1
    kill -9 "$PID" 2>/dev/null
    rm -f "$PID_FILE"
    echo "✅ 已停止 Go API 服务"
fi

# 停止 PHP 桥接服务
PHP_PID_FILE="/www/wwwroot/qingka/php-api/php-api.pid"
if [ -f "$PHP_PID_FILE" ]; then
    PHP_PID=$(cat "$PHP_PID_FILE")
    kill "$PHP_PID" 2>/dev/null
    sleep 1
    kill -9 "$PHP_PID" 2>/dev/null
    rm -f "$PHP_PID_FILE"
    echo "✅ 已停止 PHP 桥接服务"
fi

# 移除 systemd 服务
for SVC in qingka-api qingka-php; do
    if [ -f "/etc/systemd/system/${SVC}.service" ]; then
        systemctl stop "$SVC" 2>/dev/null
        systemctl disable "$SVC" 2>/dev/null
        rm -f "/etc/systemd/system/${SVC}.service"
        echo "✅ 已移除 systemd 服务: $SVC"
    fi
done
systemctl daemon-reload 2>/dev/null

# 移除心跳 cron 任务
crontab -l 2>/dev/null | grep -v 'qingka_heartbeat' | grep -v 'cron_heartbeat' | crontab - 2>/dev/null
echo "✅ 已移除心跳定时任务"

# 删除 Nginx 域名配置
DOMAIN_FILE="/www/wwwroot/qingka/domain.txt"
if [ -f "$DOMAIN_FILE" ]; then
    DOMAIN=$(cat "$DOMAIN_FILE")
    if [ -n "$DOMAIN" ]; then
        rm -f "/www/server/panel/vhost/nginx/${DOMAIN}.conf"
        rm -f "/www/server/panel/vhost/nginx/well-known/${DOMAIN}.conf"
        nginx -t 2>/dev/null && systemctl reload nginx 2>/dev/null
        echo "✅ 已移除域名配置: $DOMAIN"
    fi
fi

# 删除插件目录
rm -rf /www/server/panel/plugin/qingka_manager
echo "✅ 已删除插件文件"

echo ""
echo "注意：项目文件保留在以下位置（如需彻底删除请手动操作）："
echo "  Go API:   /www/wwwroot/qingka/go-api/"
echo "  PHP API:  /www/wwwroot/qingka/php-api/"
echo "  前端管理端: /www/wwwroot/qingka/admin/"
echo "  商城 H5:    /www/wwwroot/qingka/mall/"
echo ""
echo "✅ 卸载完成"
