#!/bin/bash

# 订单进度诊断和修复脚本
# 使用方法：./tools/fix-order-progress.sh [订单 ID 列表]
# 示例：./tools/fix-order-progress.sh 3 4

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CONFIG_PATH="$PROJECT_ROOT/config/config.yaml"

yaml_value() {
    local section="$1"
    local key="$2"
    awk -v section="$section" -v key="$key" '
        $0 ~ "^" section ":" { in_section = 1; next }
        in_section && /^[^[:space:]]/ { in_section = 0 }
        in_section && $1 == key ":" {
            gsub(/"/, "", $2)
            print $2
            exit
        }
    ' "$CONFIG_PATH"
}

echo "========================================"
echo "  订单进度诊断和修复工具"
echo "========================================"
echo ""

# 加载数据库配置
DB_USER="$(yaml_value database user)"
DB_PASS="$(yaml_value database password)"
DB_NAME="$(yaml_value database dbname)"
DB_HOST="$(yaml_value database host)"
DB_PORT="$(yaml_value database port)"

MYSQL_CMD="mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME"

# 检查是否指定了订单 ID
if [ $# -eq 0 ]; then
    echo "用法：$0 [订单 ID1] [订单 ID2] ..."
    echo ""
    echo "未指定订单 ID，将检查最近 10 个没有进度的订单..."
    echo ""

    # 查询最近 10 个 dockstatus=1 但 process 为空的订单
    echo "[诊断] 检查 dockstatus=1 但进度为空的订单:"
    $MYSQL_CMD -e "
        SELECT oid, uid, user, kcname, status, dockstatus, yid,
               COALESCE(process,'') as process,
               COALESCE(remarks,'') as remarks
        FROM qingka_wangke_order
        WHERE dockstatus = 1 AND (process = '' OR process IS NULL OR process = '0')
        ORDER BY oid DESC
        LIMIT 10;
    " 2>/dev/null

    echo ""
    echo "如果要同步这些订单的进度，请执行:"
    echo "  $0 <订单 ID 列表>"
    echo ""
    echo "例如：$0 3 4 5 6"
    exit 0
fi

# 检查每个订单的状态
echo "[诊断] 检查指定订单的状态:"
echo ""

for oid in "$@"; do
    echo "--- 订单 ID: $oid ---"

    # 查询订单信息
    RESULT=$($MYSQL_CMD -e "
        SELECT oid, uid, user, kcname, status, dockstatus,
               COALESCE(yid,'NULL') as yid,
               COALESCE(process,'NULL') as process,
               COALESCE(remarks,'NULL') as remarks,
               hid
        FROM qingka_wangke_order
        WHERE oid = $oid;
    " 2>/dev/null | tail -1)

    if [ -z "$RESULT" ]; then
        echo "  ✗ 订单不存在"
        continue
    fi

    echo "  订单信息：$RESULT"

    # 分析可能的问题
    DOCKSTATUS=$(echo "$RESULT" | awk '{print $6}')
    YID=$(echo "$RESULT" | awk '{print $7}')
    PROCESS=$(echo "$RESULT" | awk '{print $8}')
    HID=$(echo "$RESULT" | awk '{print $10}')

    if [ "$DOCKSTATUS" = "0" ]; then
        echo "  ⚠ 问题：dockstatus=0，订单尚未对接到上游"
        echo "     解决：需要先下单到上游供应商"
    elif [ "$DOCKSTATUS" = "1" ]; then
        if [ "$YID" = "NULL" ] || [ "$YID" = "" ]; then
            echo "  ⚠ 问题：dockstatus=1 但 yid 为空，无法查询进度"
            echo "     解决：检查上游下单是否成功"
        elif [ "$PROCESS" = "NULL" ] || [ "$PROCESS" = "" ] || [ "$PROCESS" = "0" ]; then
            echo "  ⚠ 问题：订单已对接但进度为空，需要手动同步"
            echo "     解决：使用 Go API 的同步功能"
        else
            echo "  ✓ 订单状态正常"
        fi
    elif [ "$DOCKSTATUS" = "2" ]; then
        echo "  ⚠ 问题：dockstatus=2，对接失败"
        echo "     解决：检查供应商配置，尝试重新下单"
    else
        echo "  处理状态：dockstatus=$DOCKSTATUS"
    fi

    echo ""
done

# 调用 Go API 同步进度
echo "[修复] 调用 Go API 同步订单进度..."
echo ""

# 构建 JSON 参数
OIDS_JSON="["
FIRST=true
for oid in "$@"; do
    if [ "$FIRST" = true ]; then
        OIDS_JSON="$OIDS_JSON$oid"
        FIRST=false
    else
        OIDS_JSON="$OIDS_JSON,$oid"
    fi
done
OIDS_JSON="$OIDS_JSON]"

# 调用同步接口（需要有效的管理员 token）
echo "POST /api/v1/admin/order/sync"
echo "参数：$OIDS_JSON"
echo ""

# 尝试调用 API（如果有有效的 token）
# curl -X POST http://127.0.0.1:8080/api/v1/admin/order/sync \
#   -H "Content-Type: application/json" \
#   -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
#   -d "{\"oids\": $OIDS_JSON}"

echo "请使用管理员账号登录后，在浏览器控制台执行:"
echo ""
echo "fetch('/api/v1/admin/order/sync', {"
echo "  method: 'POST',"
echo "  headers: {"
echo "    'Content-Type': 'application/json',"
echo "    'Authorization': 'Bearer ' + localStorage.getItem('accessToken')"
echo "  },"
echo "  body: JSON.stringify({oids: $OIDS_JSON})"
echo "}).then(r => r.json()).then(console.log);"
echo ""

# 或者直接修改数据库设置 dockstatus=1 强制同步
echo "========================================"
echo "  数据库修复选项（谨慎使用）"
echo "========================================"
echo ""
echo "如果订单确实已经在上游处理，可以手动设置 dockstatus=1:"
echo ""

for oid in "$@"; do
    echo "UPDATE qingka_wangke_order SET dockstatus = '1' WHERE oid = $oid;"
done

echo ""
echo "执行 SQL 后，订单将被包含在自动同步任务中（每 2 分钟）"
echo "或者立即调用上面的 API 手动同步"
echo ""
echo "========================================"
echo "  诊断完成"
echo "========================================"
