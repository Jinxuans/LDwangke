# 插件部署指南

> **⚠️ `d:\hzw1\149.88.74.83` 是旧系统，禁止修改。所有插件源文件统一放在 `go-api/deploy/plugins/` 下。**
>
> **⚠️ 部署到生产服务器时必须采用扁平结构（模块目录直接放在网站根目录下），因为加密 PHP 代码内部硬编码了 `../confing/common.php` 相对路径。**

## 生产服务器目录结构（扁平）

```
服务器根目录/
├── confing/common.php              # 已有（旧系统公共配置）
├── index/head.php, footer.php      # 已有（旧系统模板）
├── auth_bridge.php                 # 认证桥接（新增）
│
├── flash/                          # ===== 闪电运动 =====
│   ├── api.php                     # 后端API（加密）
│   ├── cron.php                    # 定时任务
│   ├── flash.class.php             # 业务类
│   ├── flash.config.php            # 对接配置
│   ├── static/                     # 静态资源
│   └── view/
│       ├── index.php               # 前端页面
│       └── footer.php              # 精简版 footer（去反iframe）
│
├── pangu/                          # ===== 盘古运动（9个模块共用） =====
│   ├── api.php                     # 后端API（加密）
│   ├── cron_keep.php ... cron_yyd.php  # 9个定时任务（加密）
│   ├── pangu.config.php            # 对接配置
│   ├── static/                     # 静态资源
│   └── view/
│       ├── head.php                # 精简版 head（去反iframe/反调试）
│       ├── footer.php              # 精简版 footer
│       └── pgkeep.php ... pgyyd.php    # 9个前端页面
│
├── appui/                          # ===== APPUI实习 =====
│   ├── view/index.php
│   └── service/api.php
├── baitan/                         # 摆摊打卡
│   ├── view/index.php
│   └── service/api.php
├── catka/                          # CATKA实习
│   ├── view/index.php
│   └── service/api.php
├── copilot/                        # COP实习
│   ├── view/index.php
│   └── service/api.php
├── mlsx/                           # 木兰实习（含外勤）
│   ├── view/index.php
│   └── service/api.php
│
├── paper_order/                    # 论文下单
│   ├── view/index.php
│   └── service/api.php
├── paper_dedup/                    # 论文查重
│   ├── view/index.php
│   └── service/api.php
├── paper_para_edit/                # 论文改写
│   ├── view/index.php
│   └── service/api.php
├── paper_list/                     # 论文列表
│   ├── view/index.php
│   └── service/api.php
└── shenyeai/                       # 深夜AI论文
    ├── view/index.php
    └── service/api.php
```

**为什么运动模块不能有 `service/` 子目录？**
盘古/闪电的 api.php 和 cron_*.php 是加密代码，内部硬编码了 `require('../confing/common.php')`，
因此这些文件必须直接放在根目录下一级（如 `pangu/api.php`），`../` 才能正确解析到根目录的 `confing/`。

## 开发仓库目录结构

```
go-api/deploy/
├── auth_bridge.php
├── plugin-deploy-guide.md
└── plugins/
    ├── sport_module/
    │   ├── flash/
    │   │   ├── view/index.php, footer.php
    │   │   └── service/api.php, cron.php, flash.class.php, flash.config.php, static/
    │   └── pangu/
    │       ├── view/pg*.php, head.php, footer.php
    │       └── service/api.php, cron_*.php, pangu.config.php, static/
    ├── internship_module/
    │   ├── appui/, baitan/, catka/, copilot/, mlsx/
    └── paper_module/
        ├── paper_order/, paper_dedup/, paper_para_edit/, paper_list/, shenyeai/
```

## 前置条件
1. 已执行 `013_module_view_url.sql`（加 view_url 字段）
2. 已执行 `014_pangu_plugin.sql`（建表 + 注册模块）
3. 已执行 `016_fix_module_paths.sql`（修正为扁平路径）

---

## 部署到生产服务器

### 1. 上传 PHP 文件

**认证桥接：**
```
go-api/deploy/auth_bridge.php → 服务器根目录/auth_bridge.php
```

**闪电运动（注意：service/ 下的文件要放到 flash/ 根下）：**
```
plugins/sport_module/flash/service/api.php           → 服务器根目录/flash/api.php
plugins/sport_module/flash/service/cron.php           → 服务器根目录/flash/cron.php
plugins/sport_module/flash/service/flash.class.php    → 服务器根目录/flash/flash.class.php
plugins/sport_module/flash/service/flash.config.php   → 服务器根目录/flash/flash.config.php
plugins/sport_module/flash/service/static/            → 服务器根目录/flash/static/
plugins/sport_module/flash/view/                      → 服务器根目录/flash/view/
```

**盘古运动（同理：service/ 下的文件放到 pangu/ 根下）：**
```
plugins/sport_module/pangu/service/api.php            → 服务器根目录/pangu/api.php
plugins/sport_module/pangu/service/cron_*.php          → 服务器根目录/pangu/cron_*.php
plugins/sport_module/pangu/service/pangu.config.php   → 服务器根目录/pangu/pangu.config.php
plugins/sport_module/pangu/service/static/            → 服务器根目录/pangu/static/
plugins/sport_module/pangu/view/                      → 服务器根目录/pangu/view/
```

**实习模块（直接对应复制）：**
```
plugins/internship_module/appui/   → 服务器根目录/appui/
plugins/internship_module/baitan/  → 服务器根目录/baitan/
plugins/internship_module/catka/   → 服务器根目录/catka/
plugins/internship_module/copilot/ → 服务器根目录/copilot/
plugins/internship_module/mlsx/    → 服务器根目录/mlsx/
```

**论文模块（直接对应复制）：**
```
plugins/paper_module/paper_order/     → 服务器根目录/paper_order/
plugins/paper_module/paper_dedup/     → 服务器根目录/paper_dedup/
plugins/paper_module/paper_para_edit/ → 服务器根目录/paper_para_edit/
plugins/paper_module/paper_list/      → 服务器根目录/paper_list/
plugins/paper_module/shenyeai/        → 服务器根目录/shenyeai/
```

### 2. 修改配置
- `flash/flash.config.php`：填写 `$docking_api`、`$docking_uid`、`$docking_key`、`$sdxy_price`
- `pangu/pangu.config.php`：填写 `$docking_api`、`$docking_uid`、`$docking_key`、各模块价格

### 3. 添加计划任务
```bash
# 闪电（1个）
*/60 * * * * cd /www/wwwroot/根目录/flash && php cron.php

# 盘古（9个）
*/60 * * * * cd /www/wwwroot/根目录/pangu && php cron_keep.php
*/60 * * * * cd /www/wwwroot/根目录/pangu && php cron_lp.php
*/60 * * * * cd /www/wwwroot/根目录/pangu && php cron_lp2.php
*/60 * * * * cd /www/wwwroot/根目录/pangu && php cron_sdxy.php
*/60 * * * * cd /www/wwwroot/根目录/pangu && php cron_tsn.php
*/60 * * * * cd /www/wwwroot/根目录/pangu && php cron_xbd.php
*/60 * * * * cd /www/wwwroot/根目录/pangu && php cron_ydsj.php
*/60 * * * * cd /www/wwwroot/根目录/pangu && php cron_yoma.php
*/60 * * * * cd /www/wwwroot/根目录/pangu && php cron_yyd.php
```

---

## 模块总览

### 运动模块

| app_id | 名称 | 渠道 | view_url | api_base |
|--------|------|------|----------|----------|
| flash_sdxy | 闪动校园 | flash | /flash/view/index.php | /flash/api.php |
| pg_keep | Keep运动 | pangu | /pangu/view/pgkeep.php | /pangu/api.php |
| pg_lp | 乐跑 | pangu | /pangu/view/pglp.php | /pangu/api.php |
| pg_lp2 | 乐跑2 | pangu | /pangu/view/pglp2.php | /pangu/api.php |
| pg_sdxy | 闪动校园(盘古) | pangu | /pangu/view/pgsdxy.php | /pangu/api.php |
| pg_tsn | 体适能 | pangu | /pangu/view/pgtsn.php | /pangu/api.php |
| pg_xbd | 小步点 | pangu | /pangu/view/pgxbd.php | /pangu/api.php |
| pg_ydsj | 运动世界 | pangu | /pangu/view/pgydsj.php | /pangu/api.php |
| pg_yoma | 悦马健身 | pangu | /pangu/view/pgyoma.php | /pangu/api.php |
| pg_yyd | 云运动 | pangu | /pangu/view/pgyyd.php | /pangu/api.php |

### 实习模块

| app_id | 名称 | 渠道 | view_url | api_base |
|--------|------|------|----------|----------|
| appui | APPUI实习 | appui | /appui/view/index.php | /appui/service/api.php |
| baitan | 摆摊打卡 | baitan | /baitan/view/index.php | /baitan/service/api.php |
| catka | CATKA实习 | catka | /catka/view/index.php | /catka/service/api.php |
| copilot | COP实习 | copilot | /copilot/view/index.php | /copilot/service/api.php |
| mlsx | 木兰实习 | mlsx | /mlsx/view/index.php | /mlsx/service/api.php |
| mlsx_wq | 木兰实习(外勤) | mlsx | /mlsx/view/index.php | /mlsx/service/api.php |

### 论文模块

| app_id | 名称 | view_url | api_base |
|--------|------|----------|----------|
| paper_order | 论文下单 | /paper_order/view/index.php | /paper_order/service/api.php |
| paper_dedup | 论文查重 | /paper_dedup/view/index.php | /paper_dedup/service/api.php |
| paper_para_edit | 论文改写 | /paper_para_edit/view/index.php | /paper_para_edit/service/api.php |
| paper_list | 论文列表 | /paper_list/view/index.php | /paper_list/service/api.php |
| shenyeai | 深夜AI论文 | /shenyeai/view/index.php | /shenyeai/service/api.php |

---

## 路径验证清单

部署后检查以下 require 路径是否正确解析：

| 文件 | require 路径 | 解析到 | 状态 |
|------|-------------|-------|------|
| `flash/cron.php` | `../confing/common.php` | `根目录/confing/common.php` | ✅ |
| `pangu/api.php` (加密) | `../confing/common.php` | `根目录/confing/common.php` | ✅ |
| `pangu/cron_*.php` (加密) | `../confing/common.php` | `根目录/confing/common.php` | ✅ |
| `pangu/view/head.php` | `../../confing/common.php` | `根目录/confing/common.php` | ✅ |
| `pangu/view/pg*.php` | `require('head.php')` | `pangu/view/head.php` | ✅ |
| `appui/service/api.php` | `__DIR__.'/../../confing/common.php'` | `根目录/confing/common.php` | ✅ |
| `auth_bridge.php` | `__DIR__.'/confing/common.php'` | `根目录/confing/common.php` | ✅ |

## 添加新插件流程

1. 在 `go-api/deploy/plugins/` 下按模块创建 `view/` + `service/` 目录
2. 部署时将模块目录扁平放到服务器根目录（如 `新模块名/`）
3. 数据库 `INSERT INTO qingka_dynamic_module` 一条记录（设置正确的 view_url 和 api_base）
4. Go / Vue 代码**零改动**
