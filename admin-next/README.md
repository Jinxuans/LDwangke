# `admin-next`

基于 Art Design Pro 初始化的新后台工作区，用于替换当前 `vben-admin/apps/web-antd`。

## 目标

- 保留现有后端接口与登录态契约
- 保留旧后台关键路由 `name`，兼容菜单配置 `menu_key`
- 逐步迁移到新的后台视觉体系
- 在迁移完成前，与旧后台并行运行

## 当前状态

- 已拉取官方 Art Design Pro 骨架
- 已执行 `clean:dev`，移除大部分演示页面
- 已开始补齐旧后台兼容层

## 开发命令

```bash
pnpm install
pnpm dev
pnpm build
```

## 环境变量

- 开发环境默认仍指向模板自带的 mock 代理
- 真正联调前，请按 `.env.local.example` 把 `VITE_API_PROXY_URL` 或 `VITE_API_URL` 指到当前后端

## 迁移文档

- 仓库级计划: `../docs/admin-next-migration-plan.md`
- 路由命名契约: `../docs/admin-next-route-contract.md`

## 当前优先级

1. 对齐登录、用户态、权限守卫
2. 对齐站点配置、菜单配置、扩展菜单接口
3. 搭建新布局壳
4. 迁移首页、订单、系统设置等首批页面
