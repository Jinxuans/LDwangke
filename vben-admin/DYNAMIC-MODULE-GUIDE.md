# 动态模块开发教程

## 架构概览

```
用户浏览器 (Vue前端)
    │
    ▼
Go后端 (代理层)  ──管理模块注册/启用/禁用
    │
    ▼
PHP后端 (业务层)  ──实际处理业务逻辑
```

系统中有三类动态模块：**运动(sport)**、**实习打卡(intern)**、**论文(paper)**。

每类模块共享同一套架构：
- **大厅页面** (`hub.vue`) — 展示该类型下所有已启用的子模块卡片
- **详情页面** (`sport.vue` / `intern.vue` / `paper.vue`) — 每个子模块的具体操作界面
- **Go代理** — 把前端请求原样转发给PHP后端
- **PHP后端** — 处理真正的业务逻辑（下单、查询等）

---

## 数据库表：`qingka_dynamic_module`

| 字段 | 说明 | 示例 |
|------|------|------|
| `id` | 自增主键 | 1 |
| `app_id` | 模块唯一标识 | `yyd`、`appui`、`paper_order` |
| `type` | 模块类型 | `sport` / `intern` / `paper` |
| `name` | 显示名称 | 云运动、APPUI打卡、论文下单 |
| `icon` | 图标 | `lucide:cloud-sun` |
| `api_base` | PHP端API路径 | `/jingyu/api.php` |
| `status` | 0=禁用 1=启用 | 1 |
| `sort` | 排序（数字小的在前） | 1 |
| `config` | JSON配置（可选） | `{"platforms":["校友邦"]}` |

---

## 请求流程（以运动模块 yyd 为例）

```
1. 用户进入 /sport/hub
2. 前端调用 GET /api/v1/modules?type=sport
3. Go返回 type=sport 且 status=1 的模块列表
4. 用户点击"云运动"卡片 → 跳转 /sport/yyd
5. 前端调用 POST /api/v1/module/yyd?act=orders
6. Go查到 yyd 的 api_base=/jingyu/api.php
7. Go转发请求到 http://127.0.0.1:9000/jingyu/api.php?appId=yyd&act=orders
8. PHP处理业务并返回结果
9. Go原样返回给前端
```

---

## 如何添加一个新模块（不需要写代码的情况）

比如要添加一个新的运动平台"阳光跑步"：

### 第1步：PHP端准备好API

确保 PHP 端能处理 `appId=ygpb&act=xxx` 的请求。

### 第2步：数据库插入记录

```sql
INSERT INTO qingka_dynamic_module 
  (app_id, type, name, icon, api_base, status, sort, config)
VALUES
  ('ygpb', 'sport', '阳光跑步', 'lucide:sun', '/jingyu/api.php', 1, 8, '{}');
```

或者通过管理后台 → 系统管理 → 模块管理 → 添加模块。

### 第3步：完成

刷新前端，运动大厅自动出现"阳光跑步"卡片，点击进入运动详情页。

---

## 如何添加一个新的模块类型（需要写代码）

比如要添加第四类模块"翻译(translate)"：

### 第1步：数据库插入模块记录

```sql
INSERT INTO qingka_dynamic_module 
  (app_id, type, name, icon, api_base, status, sort, config)
VALUES
  ('fanyi1', 'translate', '翻译服务A', 'lucide:languages', '/fanyi/api.php', 1, 301, '{}');
```

### 第2步：创建详情页

新建文件 `vben-admin/apps/web-antd/src/views/module/translate.vue`：

```vue
<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { Page } from '@vben/common-ui';
import { callModuleApi, getModuleApi } from '#/api/module';

const route = useRoute();
const appId = computed(() => route.params.appId as string);

// 你的业务逻辑...
// 调用PHP后端：callModuleApi(appId.value, 'translate_order', { text: '...' })
// GET请求：getModuleApi(appId.value, 'price', { lang: 'en' })
</script>

<template>
  <Page title="翻译服务">
    <!-- 你的UI -->
  </Page>
</template>
```

### 第3步：注册路由

编辑 `vben-admin/apps/web-antd/src/router/routes/modules/module.ts`，在 `routes` 数组里加一组：

```ts
// 翻译服务
{
  component: BasicLayout,
  meta: {
    icon: 'lucide:languages',
    order: 8,
    title: '翻译服务',
  },
  name: 'TranslateModule',
  path: '/translate',
  redirect: '/translate/hub',
  children: [
    {
      name: 'TranslateHub',
      path: '/translate/hub',
      component: () => import('#/views/module/hub.vue'),
      meta: { icon: 'lucide:layout-grid', title: '翻译大厅' },
      props: { moduleType: 'translate', detailRoute: 'TranslateDetail' },
    },
    {
      name: 'TranslateDetail',
      path: '/translate/:appId',
      component: () => import('#/views/module/translate.vue'),
      meta: { title: '翻译订单', hideInMenu: true },
    },
  ],
},
```

### 第4步：（可选）更新大厅标题

编辑 `hub.vue` 的 `typeLabels` 加一行：

```ts
translate: { title: '翻译大厅', emptyText: '暂无可用的翻译模块', icon: '🌐' },
```

### 第5步：完成

侧边栏自动出现"翻译服务"菜单 → 翻译大厅 → 点击卡片进入翻译详情页。

---

## 前端API说明

文件位置：`src/api/module.ts`

| 函数 | 用途 | 示例 |
|------|------|------|
| `getModulesByTypeApi(type)` | 按类型获取启用的模块列表 | `getModulesByTypeApi('sport')` |
| `callModuleApi(appId, act, data)` | POST请求代理到PHP | `callModuleApi('yyd', 'add', { user: '张三' })` |
| `getModuleApi(appId, act, params)` | GET请求代理到PHP | `getModuleApi('yyd', 'get_price', { school: 'xx大学' })` |
| `getAllModulesApi()` | 管理员获取所有模块 | 管理页面用 |
| `saveModuleApi(data)` | 管理员保存模块 | 管理页面用 |
| `deleteModuleApi(id)` | 管理员删除模块 | 管理页面用 |

### callModuleApi vs getModuleApi

- **callModuleApi** = POST请求，用于提交数据（下单、修改、删除）
- **getModuleApi** = GET请求，用于查询数据（获取价格、列表）
  - 第三个参数 `params` 会作为URL查询参数传递：
  ```ts
  // 实际请求: GET /module/appui?act=getSchoolList&pid=1
  getModuleApi('appui', 'getSchoolList', { pid: '1' })
  ```

---

## Go后端代理逻辑

Go后端不处理任何业务逻辑，只做三件事：

1. **查模块** — 根据 `app_id` 查数据库拿到 `api_base`
2. **拼URL** — `http://127.0.0.1:9000` + `api_base` + 所有query参数
3. **转发** — 把请求原样发给PHP，把PHP的响应原样返回

```
前端请求:  POST /api/v1/module/yyd?act=orders&page=1
Go拼接为:  POST http://127.0.0.1:9000/jingyu/api.php?appId=yyd&act=orders&page=1
PHP处理后返回 → Go原样返回给前端
```

---

## 管理后台操作

路径：系统管理 → 模块管理

功能：
- 查看所有模块列表（包括禁用的）
- 添加新模块
- 编辑模块信息（名称、图标、API路径、配置等）
- 点击状态标签快速启用/禁用
- 删除模块

---

## 文件清单

```
前端:
  src/api/module.ts                          # API函数
  src/router/routes/modules/module.ts        # 路由定义（三大类型）
  src/views/module/hub.vue                   # 通用大厅（所有类型共用）
  src/views/module/sport.vue                 # 运动详情页
  src/views/module/intern.vue                # 实习详情页
  src/views/module/paper.vue                 # 论文详情页
  src/views/admin/modules.vue                # 管理员模块管理

后端:
  go-api/internal/model/module.go            # 数据模型
  go-api/internal/service/module.go          # 数据库查询
  go-api/internal/handler/module.go          # HTTP处理（列表+代理+管理CRUD）
  go-api/cmd/server/main.go                  # 路由注册
  go-api/migrations/004_create_dynamic_module_table.sql  # 建表
  go-api/migrations/005_expand_dynamic_module.sql        # 加type+种子数据
```
