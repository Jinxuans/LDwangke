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

# 删除 Nginx 域名配置
DOMAIN_FILE="/www/wwwroot/qingka/domain.txt"
if [ -f "$DOMAIN_FILE" ]; then
    DOMAIN=$(cat "$DOMAIN_FILE")
    if [ -n "$DOMAIN" ]; then
        rm -f "/www/server/panel/vhost/nginx/${DOMAIN}.conf"
        nginx -t 2>/dev/null && systemctl reload nginx 2>/dev/null
        echo "✅ 已移除域名配置: $DOMAIN"
    fi
fi

# 删除插件目录
rm -rf /www/server/panel/plugin/qingka_manager
echo "✅ 已删除插件文件"

echo ""
echo "注意：项目文件保留在以下位置（如需彻底删除请手动操作）："
echo "  Go API:  /www/wwwroot/qingka/go-api/"
echo "  前端:    /var/www/admin/"
echo "  商城:    /var/www/mall/"
echo ""
echo "✅ 卸载完成"
