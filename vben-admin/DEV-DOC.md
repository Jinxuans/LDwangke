# 网课代刷管理系统 V2 — 项目开发文档

## 1. 项目概述

基于 Vue Vben Admin + Go + PHP 重写现有网课代刷管理系统，实现前后端分离、模块化架构。

### 1.1 技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| **前端** | Vue 3 + Vite + TypeScript | Vben Admin 框架，Ant Design Vue 组件库 |
| **Go API** | Go 1.21+ / Gin | **仅负责**：网课订单、在线聊天、WebSocket 推送、认证、课程查询 |
| **PHP API** | PHP 8.x | **其余全部**：运动、实习、论文、管理后台、用户、支付 |
| **数据库** | MySQL 5.7+ | 复用现有表结构，逐步优化 |
| **缓存/队列** | Redis | 缓存热数据 + 消息队列 |
| **部署** | Nginx + Docker | 反向代理 + 容器化部署 |

### 1.2 架构图

```
                    ┌──────────────────────────────┐
                    │        Nginx 网关             │
                    │  静态文件 / 反向代理 / SSL     │
                    └──────┬────────┬───────┬──────┘
                           │        │       │
              /            │  /api/ │  /ws/ │  /php-api/
              │            │        │       │
    ┌─────────▼──┐  ┌──────▼──┐  ┌──▼──┐  ┌▼──────────┐
    │ Vben Admin │  │  Go API │  │ WS  │  │  PHP API  │
    │ Vue 3 SPA  │  │  (Gin)  │  │(Go) │  │  (模块化) │
    │ 静态文件    │  │网课+聊天 │  │推送  │  │运动/实习等│
    └────────────┘  └────┬────┘  └──┬──┘  └─────┬─────┘
                         │          │            │
                    ┌────▼──────────▼────────────▼────┐
                    │          MySQL + Redis           │
                    └────────────┬────────────────────┘
                                 │
                    ┌────────────▼────────────────────┐
                    │     Go Worker (现有队列消费)      │
                    └─────────────────────────────────┘
```

---

## 2. 目录结构

```
d:\hzw1\
├── vben-admin/                     # 前端项目 (Vben Admin)
│   ├── apps/web-antd/src/
│   │   ├── views/                  # 页面组件
│   │   │   ├── order/              # 网课订单模块
│   │   │   ├── sport/              # 校园运动模块
│   │   │   ├── intern/             # 实习打卡模块
│   │   │   ├── paper/              # 论文撰写模块
│   │   │   ├── chat/               # 在线客服模块
│   │   │   ├── admin/              # 系统管理模块
│   │   │   └── user/               # 用户中心模块
│   │   ├── api/                    # API 接口定义
│   │   ├── router/                 # 路由配置
│   │   └── stores/                 # 状态管理
│   └── DEV-DOC.md                  # 本文档
│
├── go-api/                         # Go API 服务
│   ├── cmd/server/main.go          # 入口
│   ├── internal/
│   │   ├── handler/                # 请求处理器（按模块）
│   │   │   ├── order.go
│   │   │   ├── chat.go
│   │   │   ├── class.go
│   │   │   └── auth.go
│   │   ├── middleware/             # 中间件
│   │   │   ├── auth.go             # JWT 认证
│   │   │   ├── cors.go             # 跨域
│   │   │   └── ratelimit.go        # 限流
│   │   ├── model/                  # 数据模型
│   │   ├── service/                # 业务逻辑
│   │   └── ws/                     # WebSocket
│   │       └── chat.go
│   ├── config/config.yaml          # 配置文件
│   ├── go.mod
│   └── go.sum
│
├── php-api/                        # PHP API 服务
│   ├── public/index.php            # 入口
│   ├── routes/                     # 路由定义
│   │   ├── admin.php               # 管理后台路由
│   │   ├── user.php                # 用户相关路由
│   │   ├── payment.php             # 支付相关路由
│   │   ├── sport.php               # 校园运动路由
│   │   ├── intern.php              # 实习打卡路由
│   │   └── paper.php               # 论文相关路由
│   ├── controllers/                # 控制器
│   ├── middleware/                  # 中间件
│   └── config.php                  # 配置文件
│
├── 149.88.74.83/                   # 旧系统（逐步废弃）
└── docker-compose.yml              # Docker 编排
```

---

## 3. 数据库设计

### 3.1 现有数据库表清单

数据库名：`7777`，共 **35+ 张表**，按业务模块分组：

#### 🎓 核心业务 — 网课订单

| 表名 | 说明 | 核心字段 |
|------|------|----------|
| `qingka_wangke_order` | **网课订单主表**（最重要） | oid, uid, cid, hid, ptname, school, user, pass, kcname, status, fees, addtime |
| `qingka_wangke_class` | 网课平台/课程配置 | cid, name, noun, price, docking, fenlei, status |
| `qingka_wangke_fenlei` | 课程分类 | id, name, sort, status |
| `qingka_wangke_huoyuan` | 货源/上游接口配置 | hid, name, url, user, pass, token, status |

#### 🏃 校园运动 — 多平台订单

| 表名 | 平台 | 供应商 |
|------|------|--------|
| `qingka_wangke_flash_sdxy` | 闪动校园 | 闪电 |
| `qingka_wangke_hzw_sdxy` | 闪动校园 | HZW |
| `qingka_wangke_hzw_ydsj` | 运动世界 | HZW |
| `qingka_wangke_jy_lp` | 步道乐跑 | 鲸鱼 |
| `qingka_wangke_jy_keep` | KEEP | 鲸鱼 |
| `qingka_wangke_jy_yoma` | 宥马健身 | 鲸鱼 |
| `qingka_wangke_jy_yyd` | 云运动 | 鲸鱼 |
| `qingka_wangke_ldrun` | 雷电跑步 | 雷电 |
| `qingka_wangke_aishen` | 运动 | 其他 |
| `qingka_wangke_huotui` | 运动 | 盘古 |

#### 📋 实习打卡

| 表名 | 说明 |
|------|------|
| `qingka_wangke_appui` | APPUI 打卡订单 |
| `qingka_baitan` | 摆摊打卡订单 |
| `mlsx_gslb` | 网签公司列表 |
| `mlsx_wj_wq` | 网签文件表 |

#### 📝 论文

| 表名 | 说明 |
|------|------|
| `qingka_wangke_lunwen` | 论文订单 |
| `qingka_wangke_shenyekm` | 深夜AI 卡密 |

#### 💬 聊天

| 表名 | 说明 |
|------|------|
| `qingka_chat_list` | 聊天会话列表 |
| `qingka_chat_msg` | 聊天消息记录 |

#### 👤 用户 & 财务

| 表名 | 说明 |
|------|------|
| `qingka_wangke_user` | 用户表（uid, user, pass, money, grade） |
| `qingka_wangke_pay` | 支付/充值记录 |
| `qingka_wangke_km` | 卡密充值 |
| `qingka_wangke_log` | 操作日志 |
| `qingka_wangke_mijia` | 密价设置（用户专属价格） |
| `qingka_wangke_dengji` | 用户等级配置 |
| `qingka_wangke_user_favorite` | 用户收藏 |

#### ⚙️ 系统

| 表名 | 说明 |
|------|------|
| `qingka_wangke_config` | 全局配置 (k-v) |
| `qingka_wangke_gonggao` | 公告 |
| `qingka_wangke_gongdan` | 工单 |
| `qingka_wangke_gongdan_msg` | 工单消息 |
| `qingka_wangke_huodong` | 活动 |
| `qingka_wangke_menu` | 动态菜单 |
| `qingka_wangke_zhiya_config` | 质押配置 |
| `qingka_wangke_zhiya_records` | 质押记录 |

### 3.2 核心表 ER 关系

```
qingka_wangke_user (uid)
    │
    ├─1:N─→ qingka_wangke_order (uid)
    │           │
    │           ├──→ qingka_wangke_class (cid)
    │           └──→ qingka_wangke_huoyuan (hid)
    │
    ├─1:N─→ qingka_wangke_pay (uid)
    ├─1:N─→ qingka_wangke_log (uid)
    ├─1:N─→ qingka_wangke_mijia (uid + cid)
    ├─1:N─→ qingka_wangke_gongdan (uid)
    ├─1:N─→ qingka_chat_list (user1 / user2)
    │           └─1:N─→ qingka_chat_msg (list_id)
    │
    └─1:N─→ 各运动订单表 (uid)

qingka_wangke_class (cid)
    │
    ├──→ qingka_wangke_fenlei (fenlei → id)
    └──→ qingka_wangke_huoyuan (docking → hid)
```

---

## 4. API 设计

### 4.1 认证方式

- **JWT Token** — 登录后返回 access_token + refresh_token
- access_token 有效期 2 小时，refresh_token 7 天
- 请求头：`Authorization: Bearer <token>`

### 4.2 Go API 路由（高频接口）

基础路径：`/api/v1`

#### 认证

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/auth/login` | 登录 |
| POST | `/auth/refresh` | 刷新 token |
| POST | `/auth/logout` | 登出 |

#### 网课订单（最高频）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/order/list` | 订单列表（分页、筛选） |
| GET | `/order/:oid` | 订单详情 |
| GET | `/order/stats` | 订单统计 |
| POST | `/order/add` | 下单 |
| POST | `/order/batch-add` | 批量下单 |
| PUT | `/order/:oid/status` | 更新订单状态 |
| PUT | `/order/:oid/password` | 修改密码 |
| PUT | `/order/:oid/pause` | 暂停/恢复订单 |
| POST | `/order/:oid/refund` | 申请退款 |
| GET | `/order/:oid/log` | 订单日志 |
| POST | `/order/export` | 导出订单 |

#### 在线聊天（HTTP 接口）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/chat/sessions` | 会话列表 |
| GET | `/chat/messages/:list_id` | 最近消息 |
| GET | `/chat/history/:list_id` | 历史消息（向上翻页，before_id） |
| GET | `/chat/new/:list_id` | 轮询新消息（after_id） |
| POST | `/chat/send` | 发送文字消息 |
| POST | `/chat/send-image` | 发送图片消息（multipart） |
| POST | `/chat/read/:list_id` | 标记已读 |
| GET | `/chat/unread` | 未读消息总数 |
| POST | `/chat/create` | 创建/获取会话 |

#### WebSocket 推送（仅通知，非聊天）

| 方法 | 路径 | 说明 |
|------|------|------|
| WS | `/ws/push` | 推送通道（订单状态变更、支付到账、系统通知等） |

#### 课程查询

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/class/list` | 课程列表 |
| GET | `/class/search` | 课程搜索 |
| GET | `/class/categories` | 分类列表 |

### 4.3 PHP API 路由（除网课/聊天外的全部接口）

基础路径：`/php-api`

#### 校园运动

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/sport/:platform/list` | 运动订单列表（platform = sdxy/ydsj/keep/lp/yoma/yyd/ldrun） |
| POST | `/sport/:platform/add` | 运动下单 |
| PUT | `/sport/:platform/:id/pause` | 暂停/恢复 |
| PUT | `/sport/:platform/:id/status` | 更新状态 |

#### 系统管理（仅管理员）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/PUT | `/admin/config` | 系统配置 |
| CRUD | `/admin/class` | 课程管理 |
| CRUD | `/admin/category` | 分类管理 |
| CRUD | `/admin/source` | 货源管理 |
| CRUD | `/admin/level` | 等级管理 |
| CRUD | `/admin/announcement` | 公告管理 |
| CRUD | `/admin/menu` | 菜单管理 |
| GET | `/admin/stats` | 数据统计 |
| POST | `/admin/ops/cleanup` | 运维清理 |

#### 用户功能

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/PUT | `/user/profile` | 个人资料 |
| GET | `/user/log` | 操作日志 |
| CRUD | `/user/ticket` | 工单 |
| POST | `/user/recharge` | 充值 |
| GET | `/user/agent` | 代理管理 |
| GET/PUT | `/user/price` | 密价设置 |

#### 支付

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/payment/create` | 创建支付订单 |
| POST | `/payment/notify` | 支付回调 |
| GET | `/payment/records` | 充值记录 |

#### 实习 & 论文

| 方法 | 路径 | 说明 |
|------|------|------|
| CRUD | `/intern/:platform/order` | 实习打卡订单 |
| CRUD | `/paper/order` | 论文订单 |

### 4.4 统一响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": { ... },
  "timestamp": 1707200000
}
```

| code | 含义 |
|------|------|
| 0 | 成功 |
| 401 | 未认证 |
| 403 | 无权限 |
| 422 | 参数错误 |
| 500 | 服务器错误 |
| 1001+ | 业务错误码 |

---

## 5. 前端页面规划

### 5.1 路由结构

```typescript
// 管理后台（admin 可见）
/admin/settings          // 网站设置
/admin/announcement      // 公告管理
/admin/level             // 等级管理
/admin/menu              // 菜单管理
/admin/class             // 网课设置
/admin/category          // 分类管理
/admin/source            // 货源管理（接口配置）
/admin/price             // 密价设置
/admin/stats             // 数据统计
/admin/ops               // 运维工具
/admin/docking           // 对接插件

// 首页
/dashboard               // 仪表盘/首页

// 网课订单
/order/add               // 查课交单
/order/quick-add         // 手机交单
/order/batch-add         // 批量交单
/order/list              // 订单汇总
/order/quality           // 质量查询

// 校园运动（子菜单按供应商分）
/sport/sdxy              // 闪动校园
/sport/ydsj              // 运动世界
/sport/keep              // KEEP
/sport/lp                // 步道乐跑
/sport/yoma              // 宥马健身
/sport/yyd               // 云运动
/sport/ldrun             // 雷电跑步
/sport/xiaomi            // 小米运动

// 实习项目
/intern/stamp            // 盖章下单
/intern/contract         // 网签下单
/intern/salary           // 工资条生成
/intern/company          // 公司列表
/intern/appui            // APPUI 打卡
/intern/catka            // CATKA 打卡
/intern/baitan           // 摆摊打卡
/intern/copilot          // COP 打卡

// 论文
/paper/order             // 论文下单
/paper/dedup             // 论文降重
/paper/edit              // 段落修改
/paper/list              // 论文管理
/paper/shenyeai          // 深夜 AI

// 用户中心
/user/profile            // 我的资料
/user/agent              // 代理管理
/user/log                // 操作日志
/user/ticket             // 工单
/user/price              // 项目价格
/user/recharge           // 充值
/user/pledge             // 质押
/user/docking            // 对接插件

// 聊天
/chat                    // 在线客服
```

### 5.2 页面组件复用策略

很多运动平台页面结构相似（表格+筛选+操作），抽象为通用组件：

```
components/
├── SportOrderTable.vue         # 运动订单通用表格
├── OrderSearchForm.vue         # 订单搜索表单
├── OrderDetailDrawer.vue       # 订单详情抽屉
├── OrderLogModal.vue           # 订单日志弹窗
├── PaginationBar.vue           # 分页组件
└── StatusTag.vue               # 状态标签
```

每个运动平台页面只需传入配置即可：

```vue
<SportOrderTable
  platform="keep"
  :columns="keepColumns"
  api="/api/v1/sport/keep/list"
/>
```

---

## 6. Go API 详细设计

### 6.1 项目初始化

```bash
cd d:\hzw1
mkdir go-api && cd go-api
go mod init go-api
go get github.com/gin-gonic/gin
go get github.com/go-sql-driver/mysql
go get github.com/redis/go-redis/v9
go get github.com/golang-jwt/jwt/v5
go get github.com/gorilla/websocket
go get gopkg.in/yaml.v3
```

### 6.2 配置文件 (config.yaml)

```yaml
server:
  port: 8080
  mode: debug  # debug / release

database:
  host: 127.0.0.1
  port: 3306
  user: root
  password: ""
  dbname: "7777"
  max_open_conns: 50
  max_idle_conns: 10

redis:
  host: 127.0.0.1
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "your-secret-key-change-in-production"
  access_ttl: 7200      # 2小时
  refresh_ttl: 604800    # 7天

cache:
  order_list_ttl: 30     # 订单列表缓存30秒
  class_list_ttl: 300    # 课程列表缓存5分钟
```

### 6.3 核心代码结构

```go
// cmd/server/main.go
func main() {
    cfg := config.Load("config.yaml")
    db := database.Connect(cfg.Database)
    rdb := cache.Connect(cfg.Redis)

    r := gin.Default()
    r.Use(middleware.CORS())

    // 公开路由
    auth := r.Group("/api/v1/auth")
    {
        auth.POST("/login", handler.Login)
        auth.POST("/refresh", handler.RefreshToken)
    }

    // 需要认证的路由
    api := r.Group("/api/v1", middleware.JWTAuth())
    {
        // 订单
        order := api.Group("/order")
        {
            order.GET("/list", handler.OrderList)
            order.GET("/:oid", handler.OrderDetail)
            order.POST("/add", handler.OrderAdd)
            // ...
        }

        // 聊天
        api.GET("/chat/sessions", handler.ChatSessions)
        api.GET("/chat/messages/:list_id", handler.ChatMessages)
    }

    // WebSocket
    r.GET("/ws/chat", middleware.WSAuth(), handler.ChatWS)

    r.Run(":" + cfg.Server.Port)
}
```

### 6.4 WebSocket 推送设计（仅通知，非聊天）

WebSocket 只用于服务端向客户端推送通知，聊天走 HTTP 接口。

```go
// internal/ws/push.go
type Hub struct {
    clients    map[int]*Client    // uid -> client
    register   chan *Client
    unregister chan *Client
}

// 推送消息类型
type PushMessage struct {
    Type    string      `json:"type"`    // order_status / payment / system / chat_notify
    Title   string      `json:"title"`
    Content string      `json:"content"`
    Data    interface{} `json:"data"`
}
```

推送场景：
- **order_status** — 订单状态变更（处理完成、异常、退款等）
- **payment** — 充值到账通知
- **system** — 系统公告、维护通知
- **chat_notify** — 有新聊天消息提醒（前端收到后轮询拉取）

连接流程：
1. 客户端通过 `ws://host/ws/push?token=xxx` 建立连接
2. 服务端验证 token，注册到 Hub
3. 业务逻辑触发时，通过 Hub 推送给目标用户
4. 客户端收到推送后，按类型做 UI 提示或数据刷新

### 6.5 聊天模块设计（HTTP 接口）

聊天复用现有 `qingka_chat_list` + `qingka_chat_msg` 表，通过 HTTP 接口实现：

- **会话列表** — 查 `chat_list` 按 `last_time` 排序，批量查未读数
- **历史消息** — 查 `chat_msg` 按 `msg_id DESC` 分页
- **新消息轮询** — 前端定时请求 `after_id` 之后的新消息
- **发送消息** — 插入 `chat_msg`，更新 `chat_list.last_msg/last_time`
- **有新消息时** — 通过 WebSocket 推送 `chat_notify`，前端立即拉取

---

## 7. 前端与旧系统对照表

| 旧文件 | 新路由 | 模块 |
|--------|--------|------|
| `index/home.php` | `/dashboard` | 首页 |
| `index/list.php` | `/order/list` | 订单汇总 |
| `index/zlcx.php` | `/order/quality` | 质量查询 |
| `index/add2.php` | `/order/add` | 查课交单 |
| `index/add4.php` | `/order/quick-add` | 手机交单 |
| `index/add5.php` | `/order/batch-add` | 批量交单 |
| `index/1.php` | `/chat` | 在线客服 |
| `index/webset.php` | `/admin/settings` | 网站设置 |
| `index/menu.php` | `/admin/menu` | 菜单管理 |
| `index/class.php` | `/admin/class` | 网课设置 |
| `index/fenlei.php` | `/admin/category` | 分类管理 |
| `index/huoyuan.php` | `/admin/source` | 货源管理 |
| `index/dengji.php` | `/admin/level` | 等级管理 |
| `index/gglist.php` | `/admin/announcement` | 公告管理 |
| `index/data.php` | `/admin/stats` | 数据统计 |
| `index/mijia.php` | `/admin/price` | 密价设置 |
| `index/info.php` | `/user/profile` | 我的资料 |
| `index/agent.php` | `/user/agent` | 代理管理 |
| `index/log.php` | `/user/log` | 操作日志 |
| `index/work.php` | `/user/ticket` | 工单 |
| `index/pay.php` | `/user/recharge` | 充值 |
| `index/pledge.php` | `/user/pledge` | 质押 |
| `index/flash_sdxy.php` | `/sport/sdxy` | 闪动校园 |
| `index/ydsj.php` | `/sport/ydsj` | 运动世界 |
| `index/pglp.php` | `/sport/lp` | 步道乐跑 |
| `index/pgkeep.php` / `keep.php` | `/sport/keep` | KEEP |
| `index/pgyoma.php` | `/sport/yoma` | 宥马健身 |
| `index/pgyyd.php` / `yyd.php` | `/sport/yyd` | 云运动 |
| `index/ldrun.php` | `/sport/ldrun` | 雷电跑步 |
| `index/mlsx.php` | `/intern/stamp` | 盖章下单 |
| `index/mlsx_wq.php` | `/intern/contract` | 网签下单 |
| `index/appui.php` | `/intern/appui` | APPUI 打卡 |
| `index/baitan.php` | `/intern/baitan` | 摆摊打卡 |
| `index/copilot.php` | `/intern/copilot` | COP 打卡 |
| `index/paper_order.php` | `/paper/order` | 论文下单 |
| `index/paper_dedup.php` | `/paper/dedup` | 论文降重 |
| `index/paper_list.php` | `/paper/list` | 论文管理 |

---

## 8. 迁移计划

### Phase 1：基础骨架（1-2 天）
- [x] Vben Admin 项目初始化
- [x] Go API 项目初始化（路由 + DB 连接 + JWT）
- [x] PHP API 项目初始化（模块化路由）
- [x] Nginx 反向代理配置

### Phase 2：认证系统（1 天）
- [x] Go: 登录/注册 API（复用现有 user 表）
- [x] 前端: 登录页面（Vben 自带，已对接 Go API）
- [x] JWT 中间件
- [x] 前端: 路由模块 + 占位页面（8 个路由模块 + 11 个页面组件）
- [x] Vite proxy → Go API + 关闭 Nitro Mock

### Phase 3：订单模块（2-3 天）
- [x] Go: 订单 CRUD API（列表/详情/统计/批量改状态/退款）
- [x] 前端: 订单列表页（替换 list.php）— 统计卡片 + 高级搜索 + 批量操作 + 表格 + 详情弹窗
- [x] 前端: 质量查询页（替换 zlcx.php）— 精简表格 + 状态/进度筛选
- [x] 前端: 交单页面（替换 add2.php）— 分类选课 + 多行查课 + 勾选下单
- [x] 前端: API 封装（order.ts + class.ts）

### Phase 4：聊天模块（1-2 天）
- [x] Go: 聊天 API 重写（匹配旧 DB 表 qingka_chat_list/msg，权限校验，创建会话）
- [x] 前端: 聊天页面（替换 chat.php）— 会话列表 + 消息气泡 + 轮询 + Enter 发送
- [x] 前端: API 封装（chat.ts）

### 站内信模块（新增）
- [x] Go: 站内信 API（收件箱/发件箱/详情/发送/附件上传/删除/未读数）
- [x] DB: `qingka_mail` 建表 SQL（migrations/003）
- [x] 前端: 站内信页面（收件箱/发件箱切换 + 未读标记 + 写信弹窗 + 附件上传下载）
- [x] 前端: API 封装（mail.ts）+ 路由注册（/user/mail）

### Phase 5：运动/动态模块（2 天）
- [x] Go: 动态模块系统（qingka_dynamic_module 表 + 模块列表 + 代理转发到 PHP 后端）
- [x] DB: 建表 SQL（migrations/004）+ 预置 7 个运动平台（yyd/ydsj/pgyyd/pgydsj/keep/bdlp/ymty）
- [x] 前端: 运动大厅（模块卡片入口 hub.vue）
- [x] 前端: 通用运动订单页（sport.vue）— 订单列表/添加/任务详情/暂停恢复/退款
- [x] 前端: 动态路由（/sport/:appId）+ API 封装（module.ts）

### Phase 6：管理后台（2 天）
- [x] Go: 管理后台 API（仪表盘统计/用户管理/课程管理/分类管理/货源管理/系统设置）
- [x] 前端: 仪表盘（6 项统计卡片）
- [x] 前端: 用户管理（搜索/编辑余额等级/重置密码）
- [x] 前端: 课程管理（课程列表+分类管理 Tab 切换/CRUD/上下架）
- [x] 前端: 货源管理（列表/编辑/API 密钥管理）
- [x] 前端: 系统设置（基本/支付/邀请 三卡片 key-value 配置）
- [x] 前端: API 封装（admin.ts）+ 路由更新（5 个管理页面）

### Phase 7：用户中心 + 支付（2 天）
- [x] Go: 用户资料 API（资料查询/改密/订单统计）
- [x] Go: 充值 API（创建订单/充值记录）
- [x] Go: 工单 API（列表/创建/回复/关闭，管理员/用户分权）
- [x] DB: `qingka_wangke_ticket` 建表 SQL（migrations/005）
- [x] 前端: 用户资料页（统计卡片+基本信息+改密）
- [x] 前端: 充值页（快捷金额+手动输入+充值记录）
- [x] 前端: 工单页（列表/创建/详情/管理员回复/关闭）
- [x] 前端: API 封装（user-center.ts）

### Phase 8：实习 & 论文
- 范围外：运动/实习/论文模块保持 PHP 不变，不纳入 Go/Vue 迁移

### Phase 8.5：管理统计 + 系统设置联动（已完成）
- [x] Go: 货源排行 API（GET /admin/rank/suppliers）— 对齐旧 ddtj.php
- [x] Go: 代理商品排行 API（GET /admin/rank/agent-products）— 对齐旧 dl.php
- [x] 前端: 货源排行独立页面（rank-suppliers.vue）
- [x] 前端: 代理统计独立页面（rank-agent-products.vue）
- [x] 前端: 路由注册（/admin/rank-suppliers, /admin/rank-agent-products）归入"管理统计"分组
- [x] Go: 代理配置联动 — user_htkh 控制后台开户、dl_pkkg+djfl 平开控制接入 AgentCreate
- [x] Go: 上级迁移 API（POST /agent/migrate-superior）— sjqykg 配置控制
- [x] Go: 公开配置白名单新增 onlineStore_trdltz/sjqykg/user_yqzc
- [x] 前端: fontsZDY/fontsFamily 自定义字体应用（router/guard.ts）
- [x] 前端: qd_notice_open 渠道公告开关（dashboard）
- [x] 前端: xdsmopen 下单扫码状态显示（order/add.vue）
- 查课配置（settings/api_ck/api_xd/api_proportion/api_tongb/api_tongbc）仍由 PHP api.php 直接读取 DB 生效，无需 Go 迁移

### 配置项生效状态
| 配置键 | 状态 | 生效位置 |
|--------|------|----------|
| sitename/logo/hlogo | ✅ | guard.ts / basic.vue |
| sykg (水印) | ✅ | guard.ts / basic.vue |
| bz (维护模式) | ✅ | guard.ts / dashboard |
| notice / tcgonggao | ✅ | dashboard |
| qd_notice_open | ✅ | dashboard (控制 notice 显示) |
| version | ✅ | basic.vue footer |
| flkg / fllx | ✅ | order/add.vue 分类模式 |
| user_yqzc | ✅ | auth.go 注册控制 |
| zdpay | ✅ | user_center.go 最低充值 |
| user_htkh | ✅ | agent.go 后台开户控制 |
| dl_pkkg / djfl | ✅ | agent.go 平开控制 |
| sjqykg | ✅ | agent.go 上级迁移控制 |
| fontsZDY / fontsFamily | ✅ | guard.ts 自定义字体 |
| xdsmopen | ✅ | order/add.vue 扫码标识 |
| onlineStore_trdltz | ✅ | 公开配置已暴露 |
| settings/api_ck/api_xd/api_proportion | ✅ | PHP api.php 直读 DB |
| api_tongb/api_tongbc | ✅ | PHP api.php 直读 DB |

### Phase 9：收尾部署（1 天）
- [ ] Docker 配置
- [ ] Nginx 配置
- [ ] 域名切换
- [ ] 回归测试

**预计总工期：12-16 天**

---

## 9. 开发环境

### 本地环境

| 工具 | 版本 | 路径 |
|------|------|------|
| Node.js | v22.18.0 | 系统 PATH |
| pnpm | v10.29.2 | 系统 PATH |
| PHP | 8.4.16 | `d:\hzw1\php\php.exe` |
| MySQL | 9.4.0 | 系统 PATH |
| Redis | Memurai | `C:\Program Files\Memurai\` |
| Go | 1.25.7 | 系统 PATH |
| Git | 2.50.1 | 系统 PATH |

### 启动命令

```bash
# 前端（Vben Admin）
cd d:\hzw1\vben-admin
pnpm dev:antd
# → http://localhost:5666

# Go API
cd d:\hzw1\go-api
go run cmd/server/main.go
# → http://localhost:8080

# PHP API
cd d:\hzw1\php-api
d:\hzw1\php\php.exe -S localhost:9000 -t public
# → http://localhost:9000
```

### 数据库初始化

```bash
# 导入现有数据到本地 MySQL
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS 7777"
mysql -u root -p 7777 < d:\hzw1\149.88.74.83\7777_2026-02-06_13-08-22_mysql_data_wsplK.sql
```

---

## 10. 部署配置

### Nginx 配置

```nginx
server {
    listen 80;
    server_name your-domain.com;

    # 前端静态文件
    location / {
        root /www/wwwroot/dist;
        try_files $uri $uri/ /index.html;
    }

    # Go API
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # WebSocket
    location /ws/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # PHP API
    location /php-api/ {
        proxy_pass http://127.0.0.1:9000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Docker Compose

```yaml
version: '3.8'
services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/conf.d/default.conf
      - ./vben-admin/dist:/www/wwwroot/dist
    depends_on:
      - go-api
      - php-api

  go-api:
    build: ./go-api
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=mysql
      - REDIS_HOST=redis
    depends_on:
      - mysql
      - redis

  go-worker:
    build: ./go-worker
    environment:
      - DB_HOST=mysql
      - REDIS_HOST=redis
    depends_on:
      - mysql
      - redis

  php-api:
    image: php:8.3-fpm
    volumes:
      - ./php-api:/var/www
    ports:
      - "9000:9000"
    depends_on:
      - mysql
      - redis

  mysql:
    image: mysql:5.7
    ports:
      - "3306:3306"
    environment:
      MYSQL_ROOT_PASSWORD: your-password
      MYSQL_DATABASE: "7777"
    volumes:
      - mysql_data:/var/lib/mysql

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  mysql_data:
```

---

## 11. 安全注意事项

1. **密码存储** — 现有系统明文存储密码，新系统应使用 bcrypt 哈希
2. **SQL 注入** — 现有系统大量字符串拼接 SQL，新系统使用参数化查询
3. **JWT Secret** — 生产环境使用强随机密钥，存入环境变量
4. **CORS** — 仅允许指定域名跨域
5. **限流** — 登录接口限流防暴力破解
6. **文件上传** — 严格校验文件类型和大小
7. **敏感数据** — 学生账号密码加密存储，日志脱敏

---

## 12. 与旧系统并行运行

迁移期间新旧系统并行：

```
旧域名 old.example.com → 现有 PHP 系统（不动）
新域名 new.example.com → 新系统（Vben + Go + PHP API）
                          └── 连接同一个 MySQL 数据库
```

- 两个系统读写同一个数据库
- 逐步将用户引导到新系统
- 确认稳定后关闭旧系统
