# API 接口文档

> 基础地址: `http://你的域名:端口`
>
> 所有接口返回 JSON 格式，Content-Type: application/json

---

## 目录

- [一、PHP 兼容接口（下游对接专用）](#一php-兼容接口下游对接专用)
- [二、外部 API（密钥认证）](#二外部-api密钥认证)
- [三、公开接口（无需认证）](#三公开接口无需认证)
- [四、认证接口](#四认证接口)
- [五、用户中心](#五用户中心)
- [六、订单管理](#六订单管理)
- [七、课程/商品](#七课程商品)
- [八、运动模块](#八运动模块)
- [九、打卡模块](#九打卡模块)
- [十、聊天系统](#十聊天系统)
- [十一、站内信](#十一站内信)
- [十二、代理管理](#十二代理管理)
- [十三、管理后台](#十三管理后台)
- [十四、商城系统](#十四商城系统)
- [十五、PHP 桥接](#十五php-桥接)
- [十六、WebSocket](#十六websocket)

---

## 一、PHP 兼容接口（下游对接专用）

> **用途**：让下游 PHP 系统无需改代码即可对接本 Go 系统，参数和返回格式与 PHP 完全一致。
>
> **路径**：`/api.php?act=xxx` 或 `/api/index.php?act=xxx`
>
> **方法**：GET / POST 均可

### 1.1 查询余额 `act=getmoney`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| uid | int | ✅ | 用户ID |
| key | string | ✅ | API密钥 |

**返回示例**：
```json
{
  "code": 1,
  "msg": "查询成功",
  "user": "zhangsan",
  "name": "张三",
  "money": "100.00"
}
```

**错误码**：
| code | 说明 |
|------|------|
| -1 | 未开通接口 |
| -2 | 密匙错误 |
| 0 | 参数为空 |

---

### 1.2 获取商品列表 `act=class`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| uid | int | ✅ | 用户ID |
| key | string | ✅ | API密钥 |

**返回示例**：
```json
{
  "code": 1,
  "data": [
    {
      "cid": 1,
      "sort": 10,
      "name": "超星学习通",
      "content": "课程描述",
      "status": 1,
      "price": "5.00",
      "price5": "5.50",
      "jiage": "7.50"
    }
  ]
}
```

> `price` 为基础价，`jiage` 为用户最终价（含加价比例和密价）。

---

### 1.3 查课 `act=get` 或 `act=chake`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| uid | int | ✅ | 用户ID |
| key | string | ✅ | API密钥 |
| platform | int | ✅ | 商品ID（cid） |
| school | string | ✅ | 学校名称 |
| user | string | ✅ | 学号/账号 |
| pass | string | ✅ | 密码 |

**返回示例**：
```json
{
  "code": 1,
  "msg": "查课成功",
  "userinfo": "学校 账号 密码",
  "data": [
    {
      "id": "课程ID",
      "name": "课程名称"
    }
  ]
}
```

---

### 1.4 下单 `act=add` / `act=sxadd` / `act=getadd`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| uid | int | ✅ | 用户ID |
| key | string | ✅ | API密钥 |
| platform | int | ✅ | 商品ID（cid） |
| school | string | ✅ | 学校名称 |
| user | string | ✅ | 学号/账号 |
| pass | string | ✅ | 密码 |
| kcname | string | ✅ | 课程名称（多个用逗号分隔） |
| kcid | string | ❌ | 课程ID（多个用逗号分隔） |
| score | string | ❌ | 成绩要求 |
| shichang | string | ❌ | 时长要求 |

**返回示例**：
```json
{"code": 0, "msg": "提交成功", "status": 0, "message": "提交成功", "id": "12345"}
```

**错误返回**：
```json
{"code": -1, "msg": "余额不足", "status": -1, "message": "余额不足"}
```

---

### 1.5 查单 `act=chadan`

> 无需 uid/key 认证

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 二选一 | 查询账号 |
| oid | int | 二选一 | 订单ID |

**返回示例**：
```json
{
  "code": 1,
  "data": [
    {
      "id": 123,
      "ptname": "超星学习通",
      "school": "XX大学",
      "user": "账号",
      "kcname": "课程名",
      "addtime": "2025-01-01 12:00:00",
      "status": "进行中",
      "process": "50%",
      "remarks": "",
      "pushUid": "",
      "pushEmail": ""
    }
  ]
}
```

---

### 1.6 按账号查单 `act=cd`

> 无需 uid/key 认证

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | ✅ | 查询账号 |

**返回格式同 1.5**（不含推送字段）

---

### 1.7 补刷 `act=budan` 或 `act=bd`

> 无需 uid/key 认证

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | ✅ | 订单ID |

**返回示例**：
```json
{"code": 1, "msg": "补刷提交成功"}
```

---

### 1.8 同步进度 `act=up`

> 无需 uid/key 认证

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | ✅ | 订单ID |

**返回示例**：
```json
{"code": 1, "msg": "同步成功，请重新查询信息"}
```

---

### 1.9 绑定微信推送 `act=bindpushuid`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| orderid | int | ✅ | 订单ID |
| pushuid | string | ✅ | 微信推送UID |

---

### 1.10 绑定邮箱推送 `act=bindpushemail`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| orderid | int | 二选一 | 订单ID |
| account | string | 二选一 | 账号 |
| pushEmail | string | ✅ | 推送邮箱 |

---

### 1.11 绑定ShowDoc推送 `act=bindshowdocpush`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| orderid | int | 二选一 | 订单ID |
| account | string | 二选一 | 账号 |
| showdoc_url | string | ✅ | ShowDoc推送URL |

---

## 二、外部 API（密钥认证）

> **基础路径**：`/api/v1/open`
>
> **认证方式**：所有请求携带 `uid` + `key` 参数（Query 或 PostForm）
>
> **方法**：GET / POST 均可

### 2.1 获取商品列表

```
GET /api/v1/open/classlist?uid=1&key=xxx
```

**返回**：
```json
{
  "code": 1,
  "data": [
    {"cid": 1, "name": "超星学习通", "price": 7.5, "fenlei": "网课"}
  ]
}
```

### 2.2 查课

```
POST /api/v1/open/query?uid=1&key=xxx
参数: cid=1&userinfo=学校 账号 密码
```

### 2.3 下单

```
POST /api/v1/open/order?uid=1&key=xxx
参数: cid=1&userinfo=学校 账号 密码 课程ID 课程名
```

### 2.4 订单列表

```
GET /api/v1/open/orderlist?uid=1&key=xxx&page=1&limit=20&status=进行中
```

**返回**：
```json
{
  "list": [...],
  "total": 100,
  "page": 1,
  "limit": 20
}
```

### 2.5 查询余额

```
GET /api/v1/open/balance?uid=1&key=xxx
```

### 2.6 查单

```
GET /api/v1/open/chadan?username=xxx
GET /api/v1/open/chadan?oid=123
```

### 2.7 绑定推送

```
POST /api/v1/open/bindpushuid     参数: orderid, pushuid
POST /api/v1/open/bindpushemail   参数: orderid, account, pushEmail
POST /api/v1/open/bindshowdocpush 参数: orderid, account, showdoc_url
```

---

## 三、公开接口（无需认证）

### 3.1 站点配置

```
GET /api/v1/site/config
```

### 3.2 公开查单

```
GET /api/v1/query?username=xxx
POST /api/v1/query  body: {username: "xxx"}
```

### 3.3 推送相关

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/push/bind-wx` | 绑定微信推送 |
| POST | `/api/v1/push/unbind-wx` | 解绑微信推送 |
| POST | `/api/v1/push/bind-email` | 绑定邮箱推送 |
| POST | `/api/v1/push/unbind-email` | 解绑邮箱推送 |
| POST | `/api/v1/push/bind-showdoc` | 绑定ShowDoc推送 |
| POST | `/api/v1/push/unbind-showdoc` | 解绑ShowDoc推送 |
| POST | `/api/v1/push/wx-qrcode` | 获取微信二维码 |
| POST | `/api/v1/push/wx-scan-uid` | 扫码获取UID |
| GET  | `/api/v1/push/puplogin` | PUP登录 |

---

## 四、认证接口

> **基础路径**：`/api/v1/auth`

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/auth/login` | 登录（限流10次/分钟） |
| POST | `/auth/register` | 注册 |
| POST | `/auth/refresh-token` | 刷新Token |
| POST | `/auth/logout` | 登出 |
| POST | `/auth/send-code` | 发送验证码 |
| POST | `/auth/forgot-password` | 忘记密码 |
| POST | `/auth/reset-password` | 重置密码 |

**登录请求**：
```json
{"username": "admin", "password": "admin123"}
```

**登录返回**：
```json
{
  "code": 0,
  "data": {
    "accessToken": "eyJhbGciOiJ...",
    "refreshToken": "eyJhbGciOiJ..."
  }
}
```

> 后续所有需认证接口请求头携带：`Authorization: Bearer <accessToken>`

---

## 五、用户中心

> **基础路径**：`/api/v1/user`  |  **认证**：JWT Token

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/user/info` | 获取当前用户信息 |
| GET | `/user/profile` | 获取个人资料 |
| POST | `/user/change-password` | 修改密码 |
| POST | `/user/change-pass2` | 修改二级密码 |
| POST | `/user/change-email/code` | 发送换绑邮箱验证码 |
| POST | `/user/change-email` | 换绑邮箱 |
| GET | `/user/pay/channels` | 获取支付渠道 |
| POST | `/user/pay` | 创建充值订单 |
| GET | `/user/pay/orders` | 充值订单列表 |
| POST | `/user/pay/check` | 检查支付状态 |
| GET | `/user/moneylog` | 资金流水 |
| GET | `/user/tickets` | 工单列表 |
| POST | `/user/ticket/create` | 创建工单 |
| POST | `/user/ticket/reply` | 回复工单 |
| POST | `/user/ticket/close/:id` | 关闭工单 |
| GET | `/user/favorites` | 收藏列表 |
| POST | `/user/favorite/add` | 添加收藏 |
| POST | `/user/favorite/remove` | 取消收藏 |
| POST | `/user/invite-code` | 设置邀请码 |
| GET | `/user/grades` | 等级列表 |
| POST | `/user/set-grade` | 设置等级 |
| POST | `/user/invite-rate` | 设置返利比例 |
| POST | `/user/secret-key` | 重置API密钥 |
| POST | `/user/push-token` | 设置推送Token |
| GET | `/user/logs` | 操作日志 |
| POST | `/user/checkin` | 签到 |
| GET | `/user/checkin/status` | 签到状态 |
| POST | `/user/cardkey/use` | 使用卡密 |

---

## 六、订单管理

> **基础路径**：`/api/v1/order`  |  **认证**：JWT Token

| 方法 | 路径 | 说明 |
|------|------|------|
| POST/GET | `/order/list` | 订单列表 |
| GET | `/order/stats` | 订单统计 |
| GET | `/order/:oid` | 订单详情 |
| POST | `/order/add` | 下单 |
| POST | `/order/status` | 修改订单状态 |
| POST | `/order/cancel` | 取消订单 |
| POST | `/order/cancel/:oid` | 取消指定订单 |
| POST | `/order/refund` | 退款 |
| GET | `/order/pause` | 暂停/恢复 |
| POST | `/order/changepass` | 修改订单密码 |
| GET | `/order/resubmit` | 补单 |
| POST | `/order/pup-reset` | PUP重置 |
| GET | `/order/logs` | 订单日志 |

**下单请求**：
```json
{
  "cid": 1,
  "data": [
    {"userinfo": "学校 账号 密码 课程ID 课程名"}
  ]
}
```

---

## 七、课程/商品

> **基础路径**：`/api/v1/class`  |  **认证**：JWT Token

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/class/list` | 商品列表 |
| GET | `/class/search` | 搜索商品 |
| POST | `/class/search` | 查课 |
| GET | `/class/categories` | 分类列表 |
| GET | `/class/category-switches` | 分类开关 |

---

## 八、运动模块

### 8.1 运动世界 `/api/v1/ydsj`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/ydsj/config` | 获取配置 |
| POST | `/ydsj/config` | 保存配置 |
| POST | `/ydsj/price` | 查询价格 |
| GET | `/ydsj/schools` | 学校列表 |
| GET | `/ydsj/orders` | 订单列表 |
| POST | `/ydsj/add` | 下单 |
| POST | `/ydsj/refund` | 退款 |
| POST | `/ydsj/toggle-run` | 开始/停止跑步 |

### 8.2 小米运动 `/api/v1/xm`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/xm/projects` | 项目列表 |
| POST | `/xm/add-order` | 下单 |
| GET | `/xm/orders` | 订单列表 |
| POST | `/xm/query-run` | 查询跑步状态 |
| GET | `/xm/refund` | 退款 |
| GET | `/xm/delete` | 删除订单 |
| GET | `/xm/sync` | 同步订单 |
| GET | `/xm/order-logs` | 订单日志 |

### 8.3 鲸鱼运动 `/api/v1/w`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/w/apps` | 应用列表 |
| POST | `/w/add-order` | 下单 |
| GET | `/w/orders` | 订单列表 |
| POST | `/w/refund` | 退款 |
| GET | `/w/sync` | 同步订单 |
| GET | `/w/resume` | 恢复订单 |
| POST | `/w/proxy` | 代理操作 |
| POST | `/w/edit-order` | 编辑订单 |
| POST | `/w/change-status` | 修改运行状态 |
| POST | `/w/remain-count` | 剩余次数 |
| POST | `/w/task-data` | 任务数据 |
| POST | `/w/edit-task` | 编辑任务 |
| POST | `/w/delay-task` | 延期任务 |
| POST | `/w/fast-delay` | 快速延期 |

### 8.4 闪电运动(闪动校园) `/api/v1/sdxy`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/sdxy/config` | 获取配置 |
| POST | `/sdxy/config` | 保存配置 |
| GET | `/sdxy/price` | 查询价格 |
| GET | `/sdxy/orders` | 订单列表 |
| POST | `/sdxy/add` | 下单 |
| POST | `/sdxy/delete` | 删除订单 |
| POST | `/sdxy/refund` | 退款 |
| POST | `/sdxy/pause` | 暂停 |
| POST | `/sdxy/get-user-info` | 获取用户信息 |
| POST | `/sdxy/send-code` | 发送验证码 |
| POST | `/sdxy/get-user-info-by-code` | 通过验证码获取信息 |
| POST | `/sdxy/update-run-rule` | 更新跑步规则 |
| POST | `/sdxy/log` | 跑步任务日志 |
| POST | `/sdxy/change-task-time` | 修改任务时间 |
| POST | `/sdxy/delay-task` | 延期任务 |

---

## 九、打卡模块

### 9.1 YF打卡 `/api/v1/yfdk`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/yfdk/config` | 获取配置 |
| POST | `/yfdk/config` | 保存配置 |
| POST | `/yfdk/price` | 查询价格 |
| GET | `/yfdk/projects` | 项目列表 |
| POST | `/yfdk/account-info` | 获取账号信息 |
| POST | `/yfdk/schools` | 学校列表 |
| POST | `/yfdk/search-schools` | 搜索学校 |
| GET | `/yfdk/orders` | 订单列表 |
| POST | `/yfdk/add` | 下单 |
| POST | `/yfdk/delete` | 删除订单 |
| POST | `/yfdk/renew` | 续费 |
| POST | `/yfdk/save` | 保存订单 |
| POST | `/yfdk/manual-clock` | 手动打卡 |
| POST | `/yfdk/logs` | 订单日志 |
| POST | `/yfdk/detail` | 订单详情 |
| POST | `/yfdk/patch-report` | 补报告 |
| POST | `/yfdk/calculate-patch-cost` | 计算补报告费用 |

### 9.2 泰山打卡 `/api/v1/sxdk`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/sxdk/config` | 获取配置 |
| POST | `/sxdk/config` | 保存配置 |
| POST | `/sxdk/price` | 查询价格 |
| GET | `/sxdk/orders` | 订单列表 |
| POST | `/sxdk/add` | 下单 |
| POST | `/sxdk/delete` | 删除订单 |
| POST | `/sxdk/edit` | 编辑订单 |
| POST | `/sxdk/search-phone-info` | 查询手机号信息 |
| POST | `/sxdk/get-log` | 获取日志 |
| POST | `/sxdk/now-check` | 立即打卡 |
| POST | `/sxdk/change-check-code` | 修改打卡状态 |
| POST | `/sxdk/change-holiday-code` | 修改节假日状态 |
| POST | `/sxdk/get-wx-push` | 获取微信推送配置 |
| POST | `/sxdk/query-source-order` | 查询源站订单 |
| POST | `/sxdk/sync` | 同步订单 |
| POST | `/sxdk/get-userrow` | 获取用户行数据 |
| POST | `/sxdk/get-async-task` | 获取异步任务 |
| POST | `/sxdk/xxy-school-list` | 校信友学校列表 |
| POST | `/sxdk/xxy-address-search` | 校信友地址搜索 |
| POST | `/sxdk/xxt-school-list` | 学习通学校列表 |

### 9.3 Appui打卡 `/api/v1/appui`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/appui/config` | 获取配置 |
| POST | `/appui/config` | 保存配置 |
| POST | `/appui/price` | 查询价格 |
| GET | `/appui/courses` | 课程列表 |
| GET | `/appui/orders` | 订单列表 |
| POST | `/appui/add` | 下单 |
| POST | `/appui/edit` | 编辑订单 |
| POST | `/appui/renew` | 续费 |
| POST | `/appui/delete` | 删除/退款 |

### 9.4 图图强国 `/api/v1/tutuqg`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/tutuqg/orders` | 订单列表 |
| POST | `/tutuqg/price` | 查询价格 |
| POST | `/tutuqg/add` | 下单 |
| POST | `/tutuqg/delete` | 删除订单 |
| POST | `/tutuqg/renew` | 续费 |
| POST | `/tutuqg/change-password` | 修改密码 |
| POST | `/tutuqg/change-token` | 修改Token |
| POST | `/tutuqg/refund` | 退款 |
| POST | `/tutuqg/sync` | 同步订单 |
| POST | `/tutuqg/batch-sync` | 批量同步 |
| POST | `/tutuqg/toggle-renew` | 开关自动续费 |

### 9.5 土拨鼠论文 `/api/v1/tuboshu`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/tuboshu/config` | 获取用户配置 |
| POST | `/tuboshu/route` | 路由代理 |
| POST | `/tuboshu/route-formdata` | FormData路由代理 |
| GET | `/tuboshu/orders` | 订单列表 |

---

## 十、聊天系统

> **基础路径**：`/api/v1/chat`  |  **认证**：JWT Token

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/chat/sessions` | 会话列表 |
| GET | `/chat/messages/:list_id` | 消息列表 |
| GET | `/chat/history/:list_id` | 历史消息 |
| GET | `/chat/new/:list_id` | 新消息 |
| POST | `/chat/send` | 发送消息 |
| POST | `/chat/send-image` | 发送图片 |
| POST | `/chat/read/:list_id` | 标记已读 |
| GET | `/chat/unread` | 未读数 |
| POST | `/chat/create` | 创建会话 |

---

## 十一、站内信

> **基础路径**：`/api/v1/mail`  |  **认证**：JWT Token

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/mail/list` | 信件列表 |
| GET | `/mail/unread` | 未读数 |
| GET | `/mail/:id` | 信件详情 |
| POST | `/mail/send` | 发送信件 |
| POST | `/mail/upload` | 上传附件 |
| DELETE | `/mail/:id` | 删除信件 |

---

## 十二、代理管理

> **基础路径**：`/api/v1/agent`  |  **认证**：JWT Token

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/agent/list` | 代理列表 |
| POST | `/agent/create` | 创建代理 |
| POST | `/agent/recharge` | 充值 |
| POST | `/agent/deduct` | 扣费 |
| POST | `/agent/change-grade` | 修改等级 |
| POST | `/agent/change-status` | 修改状态 |
| POST | `/agent/reset-password` | 重置密码 |
| POST | `/agent/open-key` | 开通API密钥 |
| POST | `/agent/set-invite-code` | 设置邀请码 |
| POST | `/agent/migrate-superior` | 迁移上级 |
| GET | `/agent/cross-recharge-check` | 跨级充值检查 |
| POST | `/agent/cross-recharge` | 跨级充值 |

---

## 十三、管理后台

> **基础路径**：`/api/v1/admin`  |  **认证**：JWT Token + 管理员权限

### 13.1 仪表盘

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/dashboard` | 仪表盘数据 |
| GET | `/admin/stats` | 统计数据 |
| GET | `/admin/queue/stats` | 对接队列状态 |
| POST | `/admin/queue/concurrency` | 设置队列并发 |

### 13.2 用户管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/users` | 用户列表 |
| POST | `/admin/user/reset-pass` | 重置密码 |
| POST | `/admin/user/balance` | 调整余额 |
| POST | `/admin/user/grade` | 设置等级 |
| POST | `/admin/impersonate` | 模拟登录 |

### 13.3 商品管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/categories` | 分类列表 |
| GET | `/admin/categories/paged` | 分类列表（分页） |
| POST | `/admin/category/save` | 保存分类 |
| DELETE | `/admin/category/:id` | 删除分类 |
| POST | `/admin/category/quick-modify` | 快速修改分类 |
| POST | `/admin/category/batch-toggle` | 批量开关分类 |
| GET | `/admin/classes` | 商品列表 |
| POST | `/admin/class/save` | 保存商品 |
| POST | `/admin/class/toggle` | 上下架商品 |
| POST | `/admin/class/batch-delete` | 批量删除 |
| POST | `/admin/class/batch-category` | 批量改分类 |
| POST | `/admin/class/batch-price` | 批量改价格 |
| POST | `/admin/class/add` | 添加商品 |
| GET | `/admin/class/dropdown` | 商品下拉列表 |

### 13.4 货源管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/suppliers` | 货源列表 |
| POST | `/admin/supplier/save` | 保存货源 |
| POST | `/admin/supplier/delete` | 删除货源 |
| DELETE | `/admin/supplier/:hid` | 删除货源 |
| GET | `/admin/supplier/balance` | 查询货源余额 |
| GET | `/admin/supplier/import` | 导入商品 |
| GET | `/admin/supplier/sync-status` | 同步状态 |
| GET | `/admin/supplier/products` | 货源商品 |
| GET | `/admin/rank/suppliers` | 货源排行 |
| GET | `/admin/rank/agent-products` | 代理商品排行 |

### 13.5 订单管理（管理员）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/admin/order/dock` | 手动对接 |
| POST | `/admin/order/redock-pending` | 重新对接 |
| POST | `/admin/order/sync` | 同步进度 |
| POST | `/admin/order/batch-sync` | 批量同步 |
| POST | `/admin/order/batch-resend` | 批量重发 |
| POST | `/admin/order/remarks` | 修改备注 |
| GET | `/admin/platform-names` | 平台名称列表 |
| GET | `/admin/moneylog` | 资金日志 |

### 13.6 系统配置

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/config` | 获取配置 |
| POST | `/admin/config` | 保存配置 |
| GET | `/admin/paydata` | 支付配置 |
| POST | `/admin/paydata` | 保存支付配置 |
| GET | `/admin/grades` | 等级列表 |
| POST | `/admin/grade/save` | 保存等级 |
| DELETE | `/admin/grade/:id` | 删除等级 |
| GET | `/admin/mijia` | 密价列表 |
| POST | `/admin/mijia/save` | 保存密价 |
| POST | `/admin/mijia/delete` | 删除密价 |
| POST | `/admin/mijia/batch` | 批量设置密价 |

### 13.7 公告管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/announcements` | 公告列表 |
| POST | `/admin/announcement/save` | 保存公告 |
| DELETE | `/admin/announcement/:id` | 删除公告 |

### 13.8 邮件系统

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/admin/email/send` | 发送邮件 |
| GET | `/admin/email/logs` | 邮件日志 |
| GET | `/admin/email/preview` | 预览邮件 |
| GET | `/admin/smtp/config` | SMTP配置 |
| POST | `/admin/smtp/config` | 保存SMTP |
| POST | `/admin/smtp/test` | 测试SMTP |
| GET | `/admin/email-pool` | 邮箱池列表 |
| POST | `/admin/email-pool/save` | 保存邮箱池 |
| DELETE | `/admin/email-pool/:id` | 删除邮箱 |
| POST | `/admin/email-pool/toggle` | 开关邮箱 |
| POST | `/admin/email-pool/test` | 测试邮箱 |
| GET | `/admin/email-pool/stats` | 邮箱统计 |
| POST | `/admin/email-pool/reset-counters` | 重置计数 |
| GET | `/admin/email-send-logs` | 发送日志 |
| GET | `/admin/email-templates` | 邮件模板列表 |
| POST | `/admin/email-templates/save` | 保存模板 |
| GET | `/admin/email-templates/preview` | 预览模板 |
| POST | `/admin/email-templates/test` | 测试模板 |

### 13.9 工单管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/tickets` | 工单列表 |
| GET | `/admin/ticket/stats` | 工单统计 |
| POST | `/admin/ticket/reply` | 回复工单 |
| POST | `/admin/ticket/close/:id` | 关闭工单 |
| POST | `/admin/ticket/auto-close` | 自动关闭 |
| POST | `/admin/ticket/report` | 工单报告 |
| POST | `/admin/ticket/sync-report` | 同步报告 |

### 13.10 运维工具

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/ops/dashboard` | 运维看板 |
| GET | `/admin/ops/probe-suppliers` | 探测货源 |
| GET | `/admin/ops/table-sizes` | 表大小 |
| GET | `/admin/ops/turbo` | 狂暴模式状态 |
| POST | `/admin/ops/turbo` | 开关狂暴模式 |
| GET | `/admin/license/status` | 授权状态 |
| GET | `/admin/demo-mode` | 演示模式状态 |
| POST | `/admin/demo-mode` | 开关演示模式 |

### 13.11 商品同步

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/sync/config` | 同步配置 |
| POST | `/admin/sync/config` | 保存配置 |
| GET | `/admin/sync/preview` | 预览同步 |
| POST | `/admin/sync/execute` | 执行同步 |
| GET | `/admin/sync/logs` | 同步日志 |
| GET | `/admin/sync/suppliers` | 监控货源 |
| GET | `/admin/sync/auto-status` | 自动同步状态 |

### 13.12 其他管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/longlong-tool/config` | 龙龙工具配置 |
| POST | `/admin/longlong-tool/config` | 保存龙龙配置 |
| POST | `/admin/longlong-tool/sync` | 龙龙同步 |
| GET | `/admin/longlong-tool/status` | 龙龙状态 |
| GET | `/admin/platform-configs` | 平台配置列表 |
| POST | `/admin/platform-config/save` | 保存平台配置 |
| DELETE | `/admin/platform-config/:pt` | 删除平台配置 |
| POST | `/admin/platform-config/parse-php` | 解析PHP代码 |
| POST | `/admin/platform-config/detect` | 检测平台 |
| GET | `/admin/modules` | 模块列表 |
| POST | `/admin/module/save` | 保存模块 |
| DELETE | `/admin/module/:id` | 删除模块 |
| GET | `/admin/menus` | 菜单列表 |
| POST | `/admin/menus` | 保存菜单 |
| GET | `/admin/tenants` | 租户列表 |
| POST | `/admin/tenant/create` | 创建租户 |
| POST | `/admin/tenant/:tid/status` | 设置租户状态 |
| GET | `/admin/checkin/stats` | 签到统计 |
| GET | `/admin/cardkeys` | 卡密列表 |
| POST | `/admin/cardkey/generate` | 生成卡密 |
| POST | `/admin/cardkey/delete` | 删除卡密 |
| GET | `/admin/activities` | 活动列表 |
| POST | `/admin/activity/save` | 保存活动 |
| DELETE | `/admin/activity/:hid` | 删除活动 |
| GET | `/admin/pledge/configs` | 质押配置 |
| POST | `/admin/pledge/config/save` | 保存质押配置 |
| DELETE | `/admin/pledge/config/:id` | 删除质押配置 |
| POST | `/admin/pledge/config/toggle` | 开关质押 |
| GET | `/admin/pledge/records` | 质押记录 |
| GET | `/admin/db-compat/check` | 数据库兼容检查 |
| POST | `/admin/db-compat/fix` | 修复数据库 |
| POST | `/admin/db-sync/test` | 测试数据同步 |
| POST | `/admin/db-sync/execute` | 执行数据同步 |
| POST | `/admin/clone/execute` | 克隆执行 |
| POST | `/admin/clone/update-prices` | 克隆更新价格 |
| POST | `/admin/clone/auto-sync` | 自动同步 |

---

## 十四、商城系统

### 14.1 C端接口（公开）

> **基础路径**：`/api/v1/mall/:tid`（tid 为租户ID）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/mall/:tid/info` | 店铺信息 |
| POST | `/mall/:tid/login` | C端用户登录 |
| GET | `/mall/:tid/products` | 商品列表 |
| GET | `/mall/:tid/product/:cid` | 商品详情 |
| POST | `/mall/:tid/query` | 查课 |
| GET | `/mall/:tid/pay/channels` | 支付渠道 |
| POST | `/mall/:tid/pay` | 创建支付 |
| POST | `/mall/:tid/order` | 下单 |
| GET | `/mall/:tid/search` | 搜索订单 |
| GET | `/mall/:tid/orders` | 订单列表 |
| GET | `/mall/:tid/order/:oid` | 订单详情 |
| GET | `/mall/:tid/pay/check` | 检查支付 |
| POST | `/mall/:tid/pay/confirm` | 确认支付 |
| POST/GET | `/mall/pay/notify` | 支付回调（全局） |

### 14.2 B端管理（需认证）

> **基础路径**：`/api/v1/tenant`  |  **认证**：JWT Token

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/tenant/mall-open-price` | 开通价格 |
| POST | `/tenant/mall-open` | 开通商城 |
| GET | `/tenant/shop` | 店铺信息 |
| POST | `/tenant/shop` | 保存店铺 |
| POST | `/tenant/pay-config` | 支付配置 |
| GET | `/tenant/products` | 商品列表 |
| POST | `/tenant/product/save` | 保存商品 |
| DELETE | `/tenant/product/:cid` | 删除商品 |
| GET | `/tenant/order/stats` | 订单统计 |
| GET | `/tenant/mall-orders` | 订单列表 |
| GET | `/tenant/cusers` | 客户列表 |
| POST | `/tenant/cuser/save` | 保存客户 |
| DELETE | `/tenant/cuser/:id` | 删除客户 |

---

## 十五、PHP 桥接

> **基础路径**：`/internal/php-bridge`
>
> **认证方式**：`sign = md5(uid + ts + bridge_secret)`，时间戳有效期 ±300 秒

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/internal/php-bridge/money` | 余额变动通知 |
| GET | `/internal/php-bridge/user` | 获取用户信息 |
| POST | `/internal/php-bridge/order` | 创建订单 |
| GET | `/api/v1/php-bridge/auth-url` | 生成PHP认证URL（需JWT） |

### PHP 反向代理

```
ANY /php-api/* → 转发到 config.yaml 中配置的 php_backend
```

---

## 十六、WebSocket

```
GET /ws/push
```

> **认证**：Query 参数 `token=<JWT Token>`
>
> 用于实时推送订单状态变更、聊天消息等通知。

---

## 通用返回格式

### 成功

```json
{
  "code": 0,
  "message": "ok",
  "data": { ... }
}
```

### 失败

```json
{
  "code": 错误码,
  "message": "错误信息"
}
```

> **注意**：PHP 兼容接口（第一部分）的返回格式使用 `code: 1` 表示成功，与 PHP 系统保持一致。其他接口使用 `code: 0` 表示成功。

---

## 下游系统对接指南

### 方式一：PHP兼容模式（推荐，零改动）

下游系统只需将 API 地址从旧 PHP 域名改为本 Go 系统域名，所有参数和返回格式完全不变：

```
旧：POST http://旧PHP域名/api.php?act=add
新：POST http://Go域名/api.php?act=add
```

### 方式二：RESTful API

使用 `/api/v1/open/*` 路径，密钥认证，标准 RESTful 风格：

```
GET http://Go域名/api/v1/open/classlist?uid=1&key=xxx
POST http://Go域名/api/v1/open/order?uid=1&key=xxx
```
