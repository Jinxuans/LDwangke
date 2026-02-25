# 数据库表结构清单

> 数据库名：`7777`（默认），编码：`utf8mb4`
>
> ✅ = 核心必需 | ⚡ = 迁移脚本创建 | 📦 = 插件可选

---

## 一、核心旧表（PHP 遗留，迁移脚本不创建）

### ✅ qingka_wangke_user — 用户表

| 字段 | 说明 |
|------|------|
| uid | 主键，自增 |
| uuid | 上级用户ID |
| user | 登录用户名 |
| pass | 密码（MD5） |
| name | 昵称 |
| money | 余额 |
| grade | 等级ID |
| active | 激活状态 |
| addprice | 加价金额 |
| yqm | 邀请码 |
| yqprice | 邀请奖励 |
| email | 邮箱 |
| phone | 手机号 |
| push_token | 推送Token |
| key | API Key |
| khcz | 客户充值开关 |
| sjuid | 上级ID |
| addtime | 注册时间 |
| lasttime | 最后登录时间 |

### ✅ qingka_wangke_order — 订单表

| 字段 | 说明 |
|------|------|
| oid | 主键，自增 |
| uid | 用户ID |
| cid | 课程ID |
| hid | 供应商ID |
| ptname | 平台名 |
| school | 学校 |
| name | 姓名 |
| user | 账号 |
| pass | 密码 |
| kcname | 课程名 |
| kcid | 课程ID（供应商侧） |
| status | 状态（待处理/进行中/已完成等） |
| fees | 费用 |
| process | 进度 |
| remarks | 备注 |
| dockstatus | 对接状态 |
| yid | 供应商订单ID |
| addtime | 下单时间 |

**必需索引**: `uid`, `status`, `dockstatus`, `hid`

### ✅ qingka_wangke_class — 课程/商品表

| 字段 | 说明 |
|------|------|
| cid | 主键，自增 |
| name | 课程名 |
| noun | 课程别名 |
| price | 价格（varchar） |
| docking | 对接方式（varchar） |
| fenlei | 所属分类 |
| status | 状态 |
| sort | 排序 |
| content | 说明内容 |
| yunsuan | 运算方式 |

### ✅ qingka_wangke_fenlei — 分类表

| 字段 | 说明 |
|------|------|
| id | 主键 |
| name | 分类名 |
| sort | 排序 |
| status | 状态 |
| recommend | 推荐开关 |
| log | 日志开关 |
| ticket | 工单开关 |
| changepass | 改密开关 |
| allowpause | 允许暂停 |
| supplier_report | 供应商上报开关 |
| supplier_report_hid | 上报供应商ID |

### ✅ qingka_wangke_huoyuan — 供应商表

| 字段 | 说明 |
|------|------|
| hid | 主键 |
| pt | 平台标识 |
| name | 名称 |
| url | API地址 |
| user | 用户名 |
| pass | 密码/密钥 |
| token | Token |
| money | 余额 |
| status | 状态 |

### ✅ qingka_wangke_config — 系统配置表

| 字段 | 说明 |
|------|------|
| key | 配置键（主键） |
| value | 配置值 |

**必需配置项**：

| key | 说明 |
|-----|------|
| sitename | 站点名称 |
| logo | Logo |
| hlogo | 后台Logo |
| sykg | 系统开关 |
| bz | 备注/帮助 |
| notice | 公告 |
| tcgonggao | 弹窗公告 |
| flkg | 分类开关 |
| fllx | 分类类型 |
| version | 版本号 |
| zdpay | 自动支付 |
| user_yqzc | 邀请注册开关 |
| sjqykg | 商家权益开关 |
| user_htkh | 后台客户开关 |
| dl_pkkg | 代理批控开关 |
| djfl | 等级分类 |

### ✅ qingka_wangke_gonggao — 公告表

| 字段 | 说明 |
|------|------|
| id | 主键 |
| title | 标题 |
| content | 内容 |
| time | 时间 |
| uid | 发布者UID |
| status | 状态 |
| zhiding | 置顶 |

### ✅ qingka_wangke_dengji — 等级表

| 字段 | 说明 |
|------|------|
| id | 主键 |
| sort | 排序 |
| name | 等级名 |
| rate | 折扣率 |
| money | 升级金额 |
| addkf | 加价客服 |
| gjkf | 高级客服 |
| status | 状态 |

### ✅ qingka_wangke_mijia — 密价表

| 字段 | 说明 |
|------|------|
| mid | 主键 |
| uid | 用户ID |
| cid | 课程ID |
| mode | 模式 |
| price | 密价 |
| addtime | 添加时间 |

### ✅ qingka_wangke_log — 操作日志表

| 字段 | 说明 |
|------|------|
| uid | 用户ID |
| ... | 其余字段由旧系统定义 |

---

## 二、迁移脚本创建的表

### ⚡ qingka_mail — 站内信（003）

| 字段 | 说明 |
|------|------|
| id | 主键 |
| from_uid | 发送者UID |
| to_uid | 接收者UID |
| title | 标题 |
| content | 内容 |
| status | 0=未读 1=已读 |
| addtime | 时间 |

**索引**: `to_uid`

### ⚡ qingka_dynamic_module — 动态模块（004+005+013+015）

| 字段 | 说明 |
|------|------|
| id | 主键 |
| app_id | 模块标识 |
| type | 类型（sport/intern/paper） |
| name | 名称 |
| description | 描述 |
| price | 展示价格 |
| icon | Lucide 图标名 |
| api_base | PHP API 路径 |
| view_url | PHP 前端页面路径 |
| status | 0=禁用 1=启用 |
| sort | 排序 |
| config | JSON 配置 |

### ⚡ qingka_wangke_ticket — 工单表（005）

| 字段 | 说明 |
|------|------|
| id | 主键 |
| uid | 用户ID |
| oid | 订单ID |
| content | 内容 |
| reply | 回复 |
| status | 状态 |
| addtime | 时间 |

### ⚡ qingka_wangke_moneylog — 资金日志（006）

| 字段 | 说明 |
|------|------|
| id | 主键 |
| uid | 用户ID |
| type | 类型 |
| money | 金额 |
| balance | 变动后余额 |
| remark | 备注 |
| addtime | 时间 |

**索引**: `uid`

### ⚡ qingka_chat_list — 聊天会话（007）

| 字段 | 说明 |
|------|------|
| list_id | 主键 |
| user1 | 用户1 |
| user2 | 用户2 |
| last_msg | 最后消息 |
| last_time | 最后时间 |

**索引**: `user1`, `user2`

### ⚡ qingka_chat_msg — 聊天消息（007）

| 字段 | 说明 |
|------|------|
| msg_id | 主键 |
| list_id | 会话ID |
| from_uid | 发送者 |
| to_uid | 接收者 |
| content | 内容 |
| img | 图片URL |
| status | 状态 |
| addtime | 时间 |

**索引**: `list_id`

### ⚡ qingka_chat_msg_archive — 聊天归档（009）

与 `qingka_chat_msg` 结构相同，用于归档旧消息。

**索引**: `list_id`

### ⚡ qingka_email_log — 邮箱验证码日志（010）

| 字段 | 说明 |
|------|------|
| id | 主键 |
| uid | 用户ID |
| email | 邮箱地址 |
| purpose | 用途（register/reset等） |
| addtime | 时间 |

**索引**: `uid`, `email`

### ⚡ qingka_platform_config — 平台接口配置（017）

| 字段 | 说明 |
|------|------|
| id | 主键 |
| pt | 平台标识（唯一） |
| name | 平台名 |
| auth_type | 认证方式（token_param/basic_auth等） |
| api_path_style | API路径风格 |
| success_codes | 成功状态码 |
| use_json | 是否JSON请求 |
| need_proxy | 是否需要代理 |
| returns_yid | 是否返回远程订单ID |
| query_act / query_path | 查询课程接口 |
| order_act / order_path | 下单接口 |
| progress_act / progress_path | 查进度接口 |
| pause_act / resume_act | 暂停/恢复接口 |
| change_pass_act | 改密接口 |
| balance_act / balance_path | 查余额接口 |
| ... | 更多字段见 model/platform_config.go |

### ⚡ qingka_wangke_sync_config — 同步配置（019）

| 字段 | 说明 |
|------|------|
| id | 主键 |
| uid | 用户ID |
| hid | 供应商ID |
| ... | 同步参数 |

### ⚡ qingka_wangke_sync_log — 同步日志（019）

| 字段 | 说明 |
|------|------|
| id | 主键 |
| uid | 用户ID |
| ... | 同步结果 |

### ⚡ qingka_smtp_config — 邮箱池配置（021）

| 字段 | 说明 |
|------|------|
| id | 主键 |
| name | 名称 |
| host | SMTP 主机 |
| port | 端口 |
| encryption | 加密方式 |
| user | 用户名 |
| password | 密码 |
| from_email | 发件地址 |
| weight | 权重 |
| day_limit | 日发送上限 |
| hour_limit | 小时上限 |
| status | 1=启用 0=禁用 2=异常 |

### ⚡ qingka_email_send_log — 邮件发送日志（021）

| 字段 | 说明 |
|------|------|
| id | 主键 |
| pool_id | 邮箱池账号ID |
| from_email | 发件地址 |
| to_email | 收件地址 |
| subject | 主题 |
| mail_type | 类型 |
| status | 1=成功 0=失败 |
| error | 错误信息 |
| addtime | 时间 |

### ⚡ qingka_email_template — 邮件模板（022）

| 字段 | 说明 |
|------|------|
| id | 主键 |
| code | 模板代码（register/reset_password等） |
| name | 模板名称 |
| subject | 邮件主题（支持变量） |
| content | HTML内容（支持变量） |
| variables | 可用变量列表 |
| status | 1=启用 0=禁用 |

---

## 三、插件表（可选，由迁移脚本 012/014 创建）

| 表名 | 说明 |
|------|------|
| qingka_wangke_flash_sdxy | 闪电-SDXY |
| qingka_wangke_pangu_keep | 盘古-Keep |
| qingka_wangke_pangu_lp | 盘古-乐跑 |
| qingka_wangke_pangu_lp2 | 盘古-乐跑2 |
| qingka_wangke_pangu_tsn | 盘古-天数N |
| qingka_wangke_pangu_yyd | 盘古-悦运动 |
| qingka_wangke_pangu_sdxy | 盘古-SDXY |
| qingka_wangke_pangu_xbd | 盘古-校步多 |
| qingka_wangke_pangu_ydsj | 盘古-运动世界 |
| qingka_wangke_pangu_yoma | 盘古-YOMA |

> 插件表缺失不影响核心功能，仅影响对应运动模块。

---

## 四、迁移脚本执行顺序

```
migrations/
├── 003_create_mail_table.sql         # 站内信
├── 004_create_dynamic_module_table.sql # 动态模块
├── 004_seed.sql                       # 模块种子数据
├── 005_category_switches.sql          # 分类开关字段
├── 005_create_ticket_table.sql        # 工单
├── 005_expand_dynamic_module.sql      # 模块扩展字段
├── 006_create_moneylog_table.sql      # 资金日志
├── 007_create_chat_tables.sql         # 聊天
├── 008_seed_mock_orders.sql           # ⚠ 测试数据，生产跳过
├── 009_chat_archive.sql               # 聊天归档
├── 010_email_log.sql                  # 邮箱验证日志
├── 011_ticket_enhance.sql             # 工单增强
├── 012_flash_sdxy.sql                 # 闪电插件表
├── 013_module_view_url.sql            # 模块 view_url 字段
├── 014_pangu_plugin.sql               # 盘古插件表
├── 015_module_enhance.sql             # 模块 description/price
├── 016_fix_module_paths.sql           # 修复模块路径
├── 017_platform_config.sql            # 平台接口配置
├── 018_platform_balance.sql           # 平台余额字段
├── 019_sync_monitor.sql               # 同步监控
├── 020_category_supplier_report.sql   # 分类供应商上报
├── 021_email_pool.sql                 # 邮箱池
├── 021b_seed_config.sql               # 配置种子
├── 022_email_templates.sql            # 邮件模板
└── 022b_fix_templates.sql             # 修复模板
```

执行方法：
```bash
cd /www/wwwroot/go-api/migrations
# 跳过 008（测试数据）
for f in 003*.sql 004*.sql 005*.sql 006*.sql 007*.sql 009*.sql 01*.sql 02*.sql; do
  echo "执行: $f"
  mysql -u用户 -p 7777 < "$f"
done
```
