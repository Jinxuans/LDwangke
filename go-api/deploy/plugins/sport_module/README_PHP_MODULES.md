# PHP 模块部署指南

本目录包含 5 个 PHP 模块，需部署到 PHP 网站根目录下。

## 模块列表

| 模块 | 目录 | API路径 | 前端页面 | 数据表 | 定时任务 |
|------|------|---------|----------|--------|----------|
| YF打卡 | `yfdk/` | `/yfdk/daka.php` | `/index/yfdk.php` | `qingka_wangke_yfdk` | 无 |
| Appui打卡 | `appui/` | `/appui/api.php` | `/index/appui.php` | `qingka_wangke_appui` | `/appui/cron.php` |
| 小米运动 | `xm/` | `/xm/xm_apis.php` | `/index/xm.php` | `xm_project`, `xm_order` | 无 |
| 泰山打卡 | `sxdk/` | `/sxdk/api.php` | `/index/copilot.php` | `qingka_wangke_sxdk` | `/sxdk/cron.php` |
| 闪电运动 | `sdxy/` | `/sdxy/api.php` | `/index/sdxy.php` | `qingka_wangke_hzw_sdxy` | `/sdxy/cron.php` |

## 部署步骤

### 1. 导入数据库
执行 `go-api/migrations/036_php_modules.sql`，创建所有表并注册到动态模块系统。

### 2. 部署 PHP 文件

将各模块文件复制到 PHP 网站根目录：

```bash
# YF打卡
cp yfdk/daka.php       /www/wwwroot/你的站点/yfdk/daka.php
cp yfdk/index/yfdk.php /www/wwwroot/你的站点/index/yfdk.php

# Appui打卡
cp -r appui/appui/*    /www/wwwroot/你的站点/appui/
cp appui/index/appui.php /www/wwwroot/你的站点/index/appui.php

# 小米运动
cp xm/xm_apis.php      /www/wwwroot/你的站点/xm/xm_apis.php
cp -r xm/redis/         /www/wwwroot/你的站点/xm/redis/
cp xm/index/xm.php      /www/wwwroot/你的站点/index/xm.php
cp xm/index/xm_order.php /www/wwwroot/你的站点/index/xm_order.php

# 泰山打卡
cp -r sxdk/sxdk/*       /www/wwwroot/你的站点/sxdk/
cp sxdk/index/copilot.php /www/wwwroot/你的站点/index/copilot.php

# 闪电运动
cp -r sdxy/sdxy/*       /www/wwwroot/你的站点/sdxy/
cp sdxy/index/sdxy.php  /www/wwwroot/你的站点/index/sdxy.php
```

### 3. 配置 Token / 密钥

每个模块需要配置对接 Token：

- **YF打卡**: 编辑 `/yfdk/daka.php`，修改 `$token` 变量
- **Appui打卡**: 编辑 `/appui/config.php`，填写 `$docking_uid` 和 `$docking_key`
- **小米运动**: 通过后台管理 `xm_project` 表中的项目配置
- **泰山打卡**: 编辑 `/sxdk/api.php` 底部的 `$token` 和账号；编辑 `/sxdk/cron.php` 同样配置
- **闪电运动**: 编辑 `/sdxy/config.php`，填写 `$docking_uid` 和 `$docking_key`

### 4. 配置定时任务（计划任务）

以下模块需要配置定时任务来同步订单状态：

```
# Appui打卡 - 每2分钟同步一次
*/2 * * * * curl -s https://你的域名/appui/cron.php > /dev/null

# 泰山打卡 - 每2-5分钟同步一次
*/2 * * * * curl -s https://你的域名/sxdk/cron.php > /dev/null

# 闪电运动 - 每2分钟同步一次
*/2 * * * * curl -s https://你的域名/sdxy/cron.php > /dev/null
```

### 5. 前端导航

在 `/index/index.php` 中添加导航链接：

```html
<li><a class="multitabs" href="yfdk">YF打卡</a></li>
<li><a class="multitabs" href="appui">Appui打卡</a></li>
<li><a class="multitabs" href="xm">小米运动</a></li>
<li><a class="multitabs" href="copilot">泰山打卡</a></li>
<li><a class="multitabs" href="sdxy">闪电运动</a></li>
```

## 注意事项

- 所有 PHP 模块依赖 `confing/common.php`（29系统公共文件），确保该文件存在于网站根目录
- appui 和闪电运动的 `api.php`/`cron.php` 为加密代码，请勿修改
- 小米运动的 `xm_apis(toc鉴权).php` 为 TOC 鉴权版本，按需使用
- 模块已通过 `qingka_dynamic_module` 注册，可在后台"模块管理"中启用/禁用
