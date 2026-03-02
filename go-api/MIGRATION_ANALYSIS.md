# install.sql 迁移到 Go 系统的兼容性分析

## 📊 总体评估

**兼容性：⚠️ 部分兼容（需要表名映射）**

install.sql 是一个旧的 PHP 系统数据库（表前缀：`love_learn_`），而 Go 系统使用的表前缀是 `qingka_wangke_`。两个系统的核心表结构相似但不完全相同。

## 🔍 表名映射关系

### 核心表对应关系

| install.sql (旧系统) | Go 系统 (新系统) | 兼容性 | 说明 |
|---------------------|-----------------|--------|------|
| `love_learn_user` | `qingka_wangke_user` | ✅ 高度兼容 | 用户表，字段基本一致 |
| `love_learn_order` | `qingka_wangke_order` | ⚠️ 部分兼容 | 订单表，新系统字段更多 |
| `love_learn_config` | `qingka_wangke_config` | ✅ 完全兼容 | 配置表，结构一致 |
| `love_learn_log` | `qingka_wangke_log` | ✅ 完全兼容 | 日志表 |
| `love_learn_pay` | `qingka_wangke_pay` | ✅ 完全兼容 | 支付表 |
| `love_learn_dengji` | `qingka_wangke_dengji` | ✅ 完全兼容 | 等级表 |
| `love_learn_class` | `qingka_wangke_class` | ⚠️ 部分兼容 | 课程分类表 |
| `love_learn_fenlei` | `qingka_wangke_fenlei` | ✅ 完全兼容 | 分类表 |
| `love_learn_huoyuan` | `qingka_wangke_huoyuan` | ⚠️ 部分兼容 | 货源表，新系统字段略少 |
| `love_learn_notice` | `qingka_wangke_gonggao` | ✅ 完全兼容 | 公告表 |
| - | `qingka_wangke_moneylog` | ❌ 新增 | 资金日志表（新系统独有） |
| - | `qingka_ext_menu` | ❌ 新增 | 扩展菜单表（新系统独有） |
| - | `qingka_email_template` | ❌ 新增 | 邮件模板表（新系统独有） |

### 扩展功能表（旧系统独有）

以下表在旧系统中存在，但不在 Go 系统的核心表定义中：

- `love_learn_aishen` - 爱神跑步订单
- `love_learn_appui` - AppUI 打卡订单
- `love_learn_baitan` - 摆摊打卡订单
- `love_learn_copilot` - Copilot 打卡订单
- `love_learn_daycue` - DayCue 打卡订单
- `love_learn_flash_sdxy` - 闪电山东校园订单
- `love_learn_huotui` - 火腿跑步订单
- `love_learn_jxjy_yjy` - 继续教育订单
- `love_learn_jy_*` - 各种跑步平台订单（keep、乐跑、悠马、运动世界等）
- `love_learn_pangu_*` - 盘古系列订单
- `love_learn_ldrun` - 乐动跑步订单
- `love_learn_sdxy` - 山东校园订单
- `love_learn_ss_ydsj` - 运动世界订单
- `love_learn_tutu` - 兔兔订单
- `love_learn_ydrun` - 运动跑步订单
- `love_learn_ykqg` - 悠课全国订单
- `love_learn_xm_*` - XM 系列订单
- `love_learn_km` - 卡密表
- `love_learn_gongdan` - 工单表
- `love_learn_dialogue` - 对话表
- `love_learn_mijia` - 密价表
- `love_learn_member` - 会员表
- `love_learn_pledge_*` - 质押相关表
- `love_learn_store_*` - 小店相关表
- `love_learn_global_config` - 全局配置表
- `love_learn_huoyuan_config` - 货源配置表
- `love_learn_huoyuan_log` - 货源日志表

## 🔧 迁移方案

### 方案 1：表名重命名迁移（推荐）

**适用场景**：希望将旧数据完整迁移到新系统

**步骤**：

1. **备份旧数据库**
```bash
mysqldump -u root -p old_database > backup_old.sql
```

2. **创建表名映射脚本**
```sql
-- 重命名核心表
RENAME TABLE love_learn_user TO qingka_wangke_user;
RENAME TABLE love_learn_order TO qingka_wangke_order;
RENAME TABLE love_learn_config TO qingka_wangke_config;
RENAME TABLE love_learn_log TO qingka_wangke_log;
RENAME TABLE love_learn_pay TO qingka_wangke_pay;
RENAME TABLE love_learn_dengji TO qingka_wangke_dengji;
RENAME TABLE love_learn_class TO qingka_wangke_class;
RENAME TABLE love_learn_fenlei TO qingka_wangke_fenlei;
RENAME TABLE love_learn_huoyuan TO qingka_wangke_huoyuan;
RENAME TABLE love_learn_notice TO qingka_wangke_gonggao;

-- 扩展功能表保持原名（Go 系统会识别为额外表）
-- 这些表可以继续使用，不影响核心功能
```

3. **运行 Go 系统的数据库修复**
```bash
# 启动 Go API
cd 29-colnt-com/go-api
go run cmd/server/main.go

# 运行修复（会自动添加缺失的列和新表）
curl -X POST http://localhost:8080/api/admin/db-compat/fix \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 方案 2：数据导出导入迁移

**适用场景**：需要清理数据或选择性迁移

**步骤**：

1. **导出核心数据**
```bash
# 导出用户数据
mysqldump -u root -p old_db love_learn_user > user_data.sql

# 导出订单数据
mysqldump -u root -p old_db love_learn_order > order_data.sql

# ... 其他核心表
```

2. **修改表名和字段**
```bash
# 使用 sed 批量替换表名
sed -i 's/love_learn_user/qingka_wangke_user/g' user_data.sql
sed -i 's/love_learn_order/qingka_wangke_order/g' order_data.sql
# ... 其他表
```

3. **导入到新系统**
```bash
mysql -u root -p new_db < user_data.sql
mysql -u root -p new_db < order_data.sql
```

### 方案 3：双系统并行运行

**适用场景**：需要逐步迁移，保持旧系统继续运行

**步骤**：

1. 在新数据库中运行 Go 系统的 Fix() 创建新表结构
2. 编写数据同步脚本，定期从旧表同步到新表
3. 逐步切换业务到新系统
4. 确认稳定后停用旧系统

## ⚠️ 注意事项

### 1. 字段差异

**用户表 (user)**
- 旧系统有：`qq_openid`, `nickname`, `faceimg`, `ck`, `xdlv`, `dd`, `dockip`, `app_token`, `vip_status`, `vip_expiry_date`, `collect_course_ids`, `store_id`, `active_pledge_json`, `pledge_power`
- 新系统有：`pass2`（二级密码）, `email`, `tuisongtoken`, `cdmoney`, `paydata`

**订单表 (order)**
- 旧系统字段更多，包含各种扩展字段
- 新系统字段相对精简，但有 `work_state`, `pushUid`, `pushStatus` 等推送相关字段

**货源表 (huoyuan)**
- 旧系统有：`ip`, `cookie`, `endtime`, `money_restrict_notice`
- 新系统字段较少

### 2. 数据类型差异

- 旧系统的 `addtime` 多为 `VARCHAR(255)`
- 新系统的 `addtime` 也是 `VARCHAR(255)`，保持一致
- 金额字段：旧系统用 `DECIMAL(10,2)`，新系统用 `DECIMAL(12,4)`，精度更高

### 3. 管理员账号

- 旧系统管理员：uid=1, user='1001011', grade=''
- 新系统管理员：uid=1, uuid=1, user='admin', grade='3'
- **迁移后需要重新设置管理员权限**

### 4. 扩展功能表

旧系统有大量扩展功能表（各种跑步平台、打卡系统等），这些表：
- ✅ 可以保留在数据库中，不影响 Go 系统运行
- ✅ Go 系统会将它们识别为"额外表"
- ⚠️ 如果需要在 Go 系统中使用这些功能，需要开发对应的 Go 模块

## 📝 迁移检查清单

- [ ] 备份旧数据库
- [ ] 确认表名映射关系
- [ ] 检查字段差异
- [ ] 测试数据导入
- [ ] 验证管理员账号
- [ ] 检查配置项迁移
- [ ] 测试核心功能
- [ ] 验证扩展功能表
- [ ] 性能测试
- [ ] 回滚方案准备

## 🎯 推荐迁移流程

```bash
# 1. 备份
mysqldump -u root -p old_db > backup_$(date +%Y%m%d).sql

# 2. 创建新数据库
mysql -u root -p -e "CREATE DATABASE new_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 3. 导入旧数据
mysql -u root -p new_db < backup_$(date +%Y%m%d).sql

# 4. 重命名核心表
mysql -u root -p new_db < rename_tables.sql

# 5. 启动 Go 系统
cd 29-colnt-com/go-api
go run cmd/server/main.go

# 6. 运行数据库修复
./test_db_compat.sh

# 7. 验证数据
# 登录系统检查用户、订单、配置等数据是否正常
```

## 💡 结论

**可以迁移，但需要注意：**

1. ✅ **核心表兼容性高**：用户、订单、配置等核心表可以通过重命名直接迁移
2. ⚠️ **需要字段补充**：新系统有一些新增字段，Fix() 会自动添加
3. ⚠️ **扩展功能需要适配**：旧系统的扩展功能表需要开发对应的 Go 模块才能使用
4. ✅ **数据安全**：迁移过程不会删除数据，只会添加缺失的结构
5. ⚠️ **管理员账号需要重置**：迁移后需要重新配置管理员权限

**建议**：先在测试环境进行完整迁移测试，确认无误后再在生产环境操作。
