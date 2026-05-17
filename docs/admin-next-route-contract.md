# `admin-next` 路由命名契约

## 背景

菜单配置接口 `/menus` 返回的 `menu_key` 会直接匹配前端路由 `name`。

相关代码：

- `vben-admin/apps/web-antd/src/api/menu-config.ts`
- `vben-admin/apps/web-antd/src/router/access.ts`
- `vben-admin/apps/web-antd/src/layouts/basic.vue`

结论：

- 以下路由名在 `admin-next` 中必须保留
- 即使调整路径、组件、布局，也不能修改这些 `name`

## 核心路由

- `Root`
- `Authentication`
- `Login`
- `CodeLogin`
- `QrCodeLogin`
- `ForgetPassword`
- `Register`
- `CheckOrder`
- `FallbackNotFound`

## 控制台与基础模块

- `Dashboard`
- `Home`
- `Agent`
- `AgentList`
- `Chat`
- `ChatIndex`
- `Dock`
- `DockIndex`
- `Order`
- `OrderAdd`
- `OrderMobileAdd`
- `OrderBatchAdd`
- `OrderList`
- `OrderQuality`
- `User`
- `UserProfile`
- `UserRecharge`
- `UserMoneyLog`
- `UserTicket`
- `UserLogs`
- `UserCheckin`
- `UserActivities`
- `UserPledge`
- `Tenant`
- `TenantOrderStats`
- `TenantShop`
- `TenantMallCategories`
- `TenantProducts`
- `TenantCUsers`
- `TenantMallOrders`
- `TenantPayConfig`
- `TenantWithdraw`
- `TenantCUserWithdraw`

## 后台管理

- `Admin`
- `AdminMenus`
- `AdminSystem`
- `AdminSettings`
- `AdminAnnouncements`
- `AdminGrades`
- `AdminCourse`
- `AdminDocking`
- `AdminSuppliers`
- `AdminCategories`
- `AdminClass`
- `AdminOps`
- `AdminChat`
- `AdminTickets`
- `AdminCheckin`
- `AdminEmailGroup`
- `AdminEmailPool`
- `AdminMail`
- `AdminEmailTemplates`
- `AdminEmailLogs`
- `AdminUpstream`
- `AdminUpstreamConfig`
- `AdminPlatformConfig`
- `AdminDashboard`
- `AdminMiJia`
- `AdminModules`
- `AdminSyncMonitor`
- `AdminDockScheduler`
- `AdminOrderProgressSync`
- `AdminStats`
- `AdminRankSuppliers`
- `AdminRankAgentProducts`
- `AdminMoneyLog`
- `AdminWithdraw`
- `AdminMallCUserWithdraw`
- `AdminTenants`
- `AdminAuxGroup`
- `AdminCardKeys`
- `AdminActivities`
- `AdminPledge`
- `AdminOpsDashboard`
- `AdminDbCompat`
- `AdminPaperConfig`

## 插件与扩展

- `Checkin`
- `YFDKIndex`
- `SXDKIndex`
- `AppuiIndex`
- `TuZhiIndex`
- `Sports`
- `SDXYIndex`
- `YDSJIndex`
- `XMIndex`
- `WRunIndex`
- `YongyeIndex`
- `Paper`
- `PaperIndex`
- `TutuQG`
- `TutuQGIndex`
- `Tuboshu`
- `TuboshuIndex`
- `ExtPageWrapper`
- `AdminExtPage`
- `FrontExtPageWrapper`
- `FrontExtPage`

## 迁移规则

- 保留 `name`
- 菜单标题、图标、层级继续由后端菜单配置覆盖
- 新项目中路由声明建议显式写出 `name`
- 任何新增路由如果需要进入菜单配置体系，必须先确认 `menu_key` 设计

## 当前已落地页面

### 核心与基础模块

- `Home`
- `DockIndex`
- `OrderAdd`
- `OrderMobileAdd`
- `OrderBatchAdd`
- `OrderList`
- `OrderQuality`
- `Chat`
- `ChatIndex`
- `AgentList`

### 用户中心与商城

- `UserProfile`
- `UserRecharge`
- `UserMoneyLog`
- `UserTicket`
- `UserLogs`
- `UserCheckin`
- `UserActivities`
- `UserPledge`
- `TenantOrderStats`
- `TenantShop`
- `TenantMallCategories`
- `TenantProducts`
- `TenantCUsers`
- `TenantMallOrders`
- `TenantPayConfig`
- `TenantWithdraw`
- `TenantCUserWithdraw`

### 后台管理

- `AdminDashboard`
- `AdminMenus`
- `AdminSettings`
- `AdminAnnouncements`
- `AdminGrades`
- `AdminDocking`
- `AdminSuppliers`
- `AdminCategories`
- `AdminClass`
- `AdminChat`
- `AdminTickets`
- `AdminCheckin`
- `AdminEmailPool`
- `AdminMail`
- `AdminEmailTemplates`
- `AdminEmailLogs`
- `AdminUpstreamConfig`
- `AdminPlatformConfig`
- `AdminMiJia`
- `AdminModules`
- `AdminSyncMonitor`
- `AdminDockScheduler`
- `AdminOrderProgressSync`
- `AdminStats`
- `AdminRankSuppliers`
- `AdminRankAgentProducts`
- `AdminMoneyLog`
- `AdminWithdraw`
- `AdminMallCUserWithdraw`
- `AdminTenants`
- `AdminCardKeys`
- `AdminActivities`
- `AdminPledge`
- `AdminOpsDashboard`
- `AdminDbCompat`
- `AdminPaperConfig`

### 插件与扩展

- `YFDKIndex`
- `SXDKIndex`
- `AppuiIndex`
- `TuZhiIndex`
- `SDXYIndex`
- `YDSJIndex`
- `XMIndex`
- `WRunIndex`
- `YongyeIndex`
- `TutuQGIndex`
- `TuboshuIndex`
- `PaperIndex`
- `AdminExtPage`
- `FrontExtPage`

## 当前重点复核

- `AdminDocking`：功能已迁入，继续复核批量上架、同步预览、失效商品检查和批量命名工具的真实链路。
- `AdminSettings`：功能已迁入，继续按真实消费链路补齐剩余配置项，并保持 `rawConfig` 透传。
- 插件页：页面已落地，后续统一做 Art 风格收敛和真实流程验收。
