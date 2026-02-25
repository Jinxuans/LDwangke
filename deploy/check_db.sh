#!/bin/bash
# ========================================
#   数据库表完整性校验脚本
#   用法: bash check_db.sh
# ========================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

read -p "MySQL 用户名 [root]: " DB_USER
DB_USER=${DB_USER:-root}
read -sp "MySQL 密码: " DB_PASS
echo ""
read -p "数据库名 [7777]: " DB_NAME
DB_NAME=${DB_NAME:-7777}

# 所有项目需要的表（核心旧表 + 迁移脚本创建的表）
REQUIRED_TABLES=(
  # ── 核心旧表（PHP 遗留，迁移脚本不创建） ──
  qingka_wangke_user
  qingka_wangke_order
  qingka_wangke_class
  qingka_wangke_fenlei
  qingka_wangke_huoyuan
  qingka_wangke_config
  qingka_wangke_gonggao
  qingka_wangke_dengji
  qingka_wangke_mijia
  qingka_wangke_log
  # ── 迁移脚本创建的表 ──
  qingka_mail
  qingka_dynamic_module
  qingka_wangke_ticket
  qingka_wangke_moneylog
  qingka_chat_list
  qingka_chat_msg
  qingka_chat_msg_archive
  qingka_email_log
  qingka_smtp_config
  qingka_platform_config
  qingka_wangke_sync_config
  qingka_wangke_sync_log
  # ── 插件表（可选） ──
  qingka_wangke_flash_sdxy
  qingka_wangke_pangu_keep
  qingka_wangke_pangu_lp
  qingka_wangke_pangu_lp2
  qingka_wangke_pangu_tsn
  qingka_wangke_pangu_yyd
  qingka_wangke_pangu_sdxy
  qingka_wangke_pangu_xbd
  qingka_wangke_pangu_ydsj
  qingka_wangke_pangu_yoma
)

# 可选表（缺少不影响核心功能）
OPTIONAL_TABLES=(
  qingka_wangke_flash_sdxy
  qingka_wangke_pangu_keep
  qingka_wangke_pangu_lp
  qingka_wangke_pangu_lp2
  qingka_wangke_pangu_tsn
  qingka_wangke_pangu_yyd
  qingka_wangke_pangu_sdxy
  qingka_wangke_pangu_xbd
  qingka_wangke_pangu_ydsj
  qingka_wangke_pangu_yoma
)

is_optional() {
  for t in "${OPTIONAL_TABLES[@]}"; do
    [ "$t" = "$1" ] && return 0
  done
  return 1
}

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  数据库表完整性校验${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 检查数据库是否存在
DB_EXISTS=$(mysql -u"$DB_USER" -p"$DB_PASS" -N -e "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME='$DB_NAME'" 2>/dev/null)
if [ -z "$DB_EXISTS" ]; then
  echo -e "${RED}❌ 数据库 $DB_NAME 不存在！${NC}"
  exit 1
fi
echo -e "数据库: ${GREEN}$DB_NAME${NC} ✅"
echo ""

# 获取数据库中已有的表
EXISTING=$(mysql -u"$DB_USER" -p"$DB_PASS" -N -e "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA='$DB_NAME'" 2>/dev/null)

missing_core=0
missing_optional=0
ok=0

for table in "${REQUIRED_TABLES[@]}"; do
  if echo "$EXISTING" | grep -qw "$table"; then
    echo -e "  ✅ $table"
    ((ok++))
  else
    if is_optional "$table"; then
      echo -e "  ${YELLOW}⚠️  $table （可选，缺少不影响核心功能）${NC}"
      ((missing_optional++))
    else
      echo -e "  ${RED}❌ $table （必需！）${NC}"
      ((missing_core++))
    fi
  fi
done

echo ""
echo "────────────────────────────────"
echo -e "  总计: ${#REQUIRED_TABLES[@]} 张表"
echo -e "  已存在: ${GREEN}$ok${NC}"
[ $missing_core -gt 0 ] && echo -e "  缺少(必需): ${RED}$missing_core${NC}"
[ $missing_optional -gt 0 ] && echo -e "  缺少(可选): ${YELLOW}$missing_optional${NC}"
echo "────────────────────────────────"

if [ $missing_core -gt 0 ]; then
  echo ""
  echo -e "${RED}⚠️  有 $missing_core 张核心表缺失！${NC}"
  echo ""
  echo "修复方法："
  echo "  1. 如果是从旧系统迁移，请先导入旧数据库 SQL 备份"
  echo "  2. 然后执行迁移脚本补全新增表："
  echo "     cd /opt/deploy/go-api/migrations"
  echo "     for f in *.sql; do mysql -u$DB_USER -p $DB_NAME < \$f; done"
  echo ""
  echo "  如果是全新部署，需要先创建核心表结构（联系开发获取基础 SQL）"
  exit 1
elif [ $missing_optional -gt 0 ]; then
  echo ""
  echo -e "${YELLOW}部分可选表缺失，可执行迁移脚本补全：${NC}"
  echo "  cd /opt/deploy/go-api/migrations"
  echo "  for f in *.sql; do mysql -u$DB_USER -p $DB_NAME < \$f; done"
else
  echo ""
  echo -e "${GREEN}✅ 所有表完整，数据库就绪！${NC}"
fi
