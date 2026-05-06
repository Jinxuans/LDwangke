# `admin-next` 工作纪要

## 当前目标

- 用 `admin-next` 逐步替换 `vben-admin/apps/web-antd`
- 保持后端接口、权限模型、菜单配置、路由命名契约不变
- 保持 `menu_key === route.name`
- 最终摆脱 `@vben/*` 运行时依赖
- 页面整体持续收敛到 `art-design-pro-main` 风格

## 当前判断标准

- [ ] 像不像 Art Design Pro
- [ ] 有没有破坏现有后端与菜单契约
- [ ] 主链路是否完整可用

目标不是“能跑就行”，而是：**功能可用 + 风格收敛 + 契约不破**。

## 当前环境

- `admin-next/.env`: `VITE_ACCESS_MODE=backend`
- `admin-next/.env.development`: `VITE_API_URL=/api/v1`、`VITE_API_PROXY_URL=https://taixuex.cc`
- `admin-next/.env.production`: `VITE_API_URL=https://taixuex.cc/api/v1`

## 阶段状态

- [x] Phase 0：并行项目启动
- [x] Phase 1：运行时基础能力迁移
- [~] Phase 2：核心公共页面迁移与样式收敛
- [~] Phase 3：高频管理页面迁移
- [~] Phase 4：插件与长尾页面迁移
- [ ] Phase 5：切流与下线

## 已完成摘要

### 基础能力

- [x] `admin-next` 已创建并可独立构建
- [x] 登录、站点配置、动态菜单、通知面板已接通
- [x] 登录后首页跳转问题已修复
- [x] `/menu/all` 404 回退 `/menus` 已兼容
- [x] 多次 `pnpm build` 已通过

### 已迁入主链路页面

- [x] 控制台 / 首页
- [x] 订单新增 / 批量交单 / 移动端交单 / 订单列表 / 质量查询
- [x] 用户资料 / 充值 / 余额流水 / 操作日志 / 工单 / 聊天
- [x] 公告 / 系统设置 / 等级 / 货源 / 分类 / 课程 / 对接 / 聊天 / 工单
- [x] 商城管理：订单统计 / 店铺设置 / 商城分类 / 选品管理 / 会员管理 / 支付订单 / 支付配置 / 提现审核
- [x] 统计、运维、邮箱、辅助业务已有页面落点
- [x] YF、泰山、Appui、凸知、实习助手、闪电、运动世界、小米、鲸鱼、永夜、图图强国、土拨鼠、智文论文插件页已有页面落点

### 已完成的关键收敛

- [x] 大量普通业务页已移除展示型页头、空头卡和无意义统计块
- [x] `Home` 已从模板 demo 收敛为真实业务总览页
- [x] `dashboard/console` 已继续按 Art 卡片节奏收敛
- [x] 首页已接入真实看板接口：`/admin/dashboard`、工单、聊天、提现、调度、进度同步、公告
- [x] 全站水印已接通，并继续兼容旧版 `sykg`
- [x] `AdminSettings` 已改为旧版习惯的标签页工作区，并保持 `/admin/config` 与 `rawConfig` 保存契约
- [x] `AdminSettings` 已补齐 `user_yqzc`、`fllx`、`top_consumers_open` 等真实消费配置项
- [x] `AdminSettings` 顶部独立摘要卡已收敛为工作区内的稳定操作栏，避免切换标签时页面头部跳动
- [x] `AdminDocking` 主动作已完成本轮验收修补：筛选结果全选/清空、跨页选择去重、单个/批量/一键对接倍率与价格校验、一键对接快速新建分类
- [x] `/dashboard/console` 已按角色分流：管理员保留运营看板，普通用户切回自己的业务首页，不再共用同一套界面
- [x] `/dashboard/console` 的高消费用户头像与右上角用户菜单头像已对齐 `agent/list` 的账号头像展示逻辑，空头像时按账号回落到 QQ 头像
- [x] 侧边栏客服聊天窗已改为静默轮询，仅在会话确有变化时才拉取消息，避免刷新打断输入和阅读历史
- [x] `/chat/index` 已收敛为页面内固定视口 + 内部滚动，不再被长会话列表持续撑高
- [x] `/order/list` 已参考 `art-design-pro-main/examples/tables` 改为 `ArtSearchBar + ArtTableHeader + ArtTable`，接入列配置、列显隐、表格工具栏和内置分页，保留现有订单动作和后端契约
- [x] `admin-next` 已重新构建通过

## 当前待做

### P1

- [ ] 对照 `docs/admin-next-route-contract.md` 的已落地清单逐页验收主动作，优先排查“页面能打开但核心动作不闭环”的入口
- [~] 复核 `AdminDocking`：代码侧主动作已修补并通过构建；仍需接真实账号数据实测同步预览、失效商品检查、批量替换/前缀的后端返回
- [ ] 复核 `AdminSettings`：从旧后台和后端真实消费点补齐剩余配置项，继续保持 `rawConfig` 透传和 `/admin/config` 保存契约

### P2

- [ ] 继续排查仍有展示感、混搭感的页面，重点看插件页、邮箱页、统计页、运维页、辅助业务页
- [ ] 复核用户中心与商城长尾入口，确认是否还存在后端菜单会下发但 `admin-next` 没有对应实现的 `menu_key`

### P3

- [ ] 对插件页做统一 Art 风格收敛和真实流程验收
- [ ] 最后处理扩展页、隐藏工具页、切流检查和旧后台下线准备

## 开发前检查

- [ ] 新页面是否已有对应 `menu_key`
- [ ] 路由 `name` 是否与旧系统一致
- [ ] 页面结构是否先按 Art 工具页组织
- [ ] 是否保留当前后端契约，避免额外改后端
- [ ] 是否先完成主链路，再考虑附加功能
- [ ] 当前改动是否已经做到“能用且像 Art”

## 注意事项

- 不要误覆盖与本轮迁移无关的后端脏改动：
  - `go-api/deploy/init_db.sql`
  - `go-api/migrations/core/001_init_core_tables.sql`
- 后续继续工作时：
  - 长期规则看 `docs/admin-next-migration-plan.md`
  - 当前进度看本文件
