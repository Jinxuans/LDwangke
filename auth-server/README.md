# 授权站 (License Server)

基于 Rust 的独立授权管理服务，提供授权码生命周期管理、域名/机器码绑定、心跳监控、HMAC 签名验证、批量操作、试用授权、公告推送、数据导出、趋势统计。

## 技术栈

- **Axum** — HTTP 框架
- **SQLite** — 嵌入式数据库（零依赖部署）
- **DashMap** — 内存缓存（验证结果缓存5分钟）
- **HMAC-SHA256** — 请求签名验证

## 项目结构

```
30/
├── Cargo.toml          # 依赖配置
├── config.toml         # 运行时配置（部署时务必修改密钥）
├── deploy.sh           # 服务器一键部署脚本
├── src/
│   ├── main.rs         # 入口 + 路由 + 定时任务
│   ├── config.rs       # 配置加载
│   ├── model.rs        # 数据模型
│   ├── db.rs           # SQLite 操作层
│   ├── auth.rs         # HMAC 签名验证 + 授权码生成
│   ├── cache.rs        # 内存缓存（带 TTL）
│   └── handler.rs      # HTTP 处理器
└── data/
    └── license.db      # SQLite 数据库（自动创建）
```

## 部署

### 1. 服务器环境要求

- Linux (Ubuntu/Debian/CentOS)
- Rust 工具链（脚本会自动安装）

### 2. 一键部署

```bash
# 上传代码到服务器后
chmod +x deploy.sh
./deploy.sh
```

### 3. 手动部署

```bash
# 安装 Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
source $HOME/.cargo/env

# 编译
cargo build --release

# 修改配置（务必修改密钥！）
cp config.toml config.toml.bak
vim config.toml

# 运行
./target/release/license-server
```

### 4. 配置说明

编辑 `config.toml`：

```toml
[server]
host = "0.0.0.0"
port = 9800
admin_token = "你的管理后台密钥"      # ⚠️ 必须修改

[security]
client_secret = "你的客户端签名密钥"   # ⚠️ 必须修改，需与客户端一致
sign_window_secs = 300               # 签名有效窗口（秒）
verify_cache_secs = 300              # 验证缓存时间（秒）
offline_threshold_secs = 1800        # 离线判定阈值（秒）
max_default_bind = 3                 # 默认最大换绑次数

[database]
path = "data/license.db"
```

### 5. systemd 服务

部署脚本会自动创建 systemd 服务，手动管理：

```bash
systemctl start license-server
systemctl stop license-server
systemctl restart license-server
systemctl status license-server
journalctl -u license-server -f    # 查看日志
```

## API 文档

### 公开接口（客户端调用）

#### 验证授权码

```
POST /api/v1/license/verify
```

请求体：
```json
{
  "license_key": "QK-FB3C5C24...",
  "domain": "example.com",
  "machine_id": "硬件指纹",
  "version": "1.0.0",
  "timestamp": 1740000000,
  "sign": "HMAC-SHA256签名"
}
```

签名算法：
```
sign_str = "domain={domain}&license_key={key}&machine_id={mid}&timestamp={ts}&version={ver}"
sign = HMAC-SHA256(client_secret, sign_str)  → hex 小写
```

成功响应：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "valid": true,
    "plan": "standard",
    "expire_at": "2027-02-25 00:00:00",
    "max_users": 100,
    "max_agents": 50
  }
}
```

#### 心跳上报

```
POST /api/v1/license/heartbeat
```

请求体：
```json
{
  "license_key": "QK-FB3C5C24...",
  "machine_id": "硬件指纹",
  "version": "1.0.0",
  "timestamp": 1740000000,
  "sign": "HMAC-SHA256签名",
  "stats": { "users": 42, "orders_today": 156 }
}
```

签名算法：
```
sign_str = "license_key={key}&machine_id={mid}&timestamp={ts}&version={ver}"
sign = HMAC-SHA256(client_secret, sign_str)  → hex 小写
```

响应会包含活跃公告：
```json
{
  "code": 0,
  "data": {
    "status": "ok",
    "notices": [{"id": 1, "title": "系统维护", "content": "今晚22点维护2小时", "type": "warning"}]
  }
}
```

#### 试用申请

```
POST /api/v1/license/trial
```

请求体：
```json
{
  "machine_id": "硬件指纹",
  "domain": "example.com",
  "timestamp": 1740000000,
  "sign": "HMAC-SHA256签名"
}
```

签名算法：`HMAC-SHA256(client_secret, machine_id + timestamp)`

每个机器码只能申请一次试用，默认 7 天、plan=trial、限制功能。

### 管理接口

所有管理接口需要 JWT 认证：`Authorization: Bearer {jwt_token}`。通过登录接口获取 token。

#### 授权码管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/admin/licenses?page=1&limit=20&keyword=&status=` | 授权码列表 |
| GET | `/api/v1/admin/licenses/export?status=` | 导出 CSV |
| POST | `/api/v1/admin/license/create` | 创建授权码 |
| POST | `/api/v1/admin/license/update` | 编辑授权码 |
| POST | `/api/v1/admin/license/revoke` | 吊销授权码 |
| POST | `/api/v1/admin/license/enable` | 启用授权码 |
| POST | `/api/v1/admin/license/unbind` | 解绑机器码 |
| POST | `/api/v1/admin/license/renew` | 续期授权码 |
| DELETE | `/api/v1/admin/license/delete/{id}` | 删除授权码 |
| POST | `/api/v1/admin/license/batch_create` | 批量创建 (1-100) |
| POST | `/api/v1/admin/license/batch_revoke` | 批量吊销 |
| POST | `/api/v1/admin/license/batch_renew` | 批量续期 |
| GET | `/api/v1/admin/license/logs` | 操作日志 |
| GET | `/api/v1/admin/license/dashboard` | 统计看板 |

#### 公告管理 (仅超管)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/admin/notices?page=1&limit=20` | 公告列表 |
| POST | `/api/v1/admin/notice/create` | 创建公告 |
| POST | `/api/v1/admin/notice/update` | 编辑公告 |
| DELETE | `/api/v1/admin/notice/delete/{id}` | 删除公告 |

#### 统计报表

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/admin/stats/trend?days=30` | 趋势数据 (7-90天) |
| GET | `/api/v1/admin/stats/distribution` | 套餐分布 |

#### 用户管理 (仅超管)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/admin/users?page=1&limit=20` | 用户列表 |
| POST | `/api/v1/admin/user/create` | 创建用户 |
| POST | `/api/v1/admin/user/update` | 编辑用户 |
| DELETE | `/api/v1/admin/user/delete/{id}` | 删除用户 |
| GET | `/api/v1/admin/me` | 当前用户信息 |

#### 创建授权码

```json
{
  "domain": "example.com",
  "note": "客户A",
  "plan": "standard",
  "max_users": 100,
  "max_agents": 50,
  "expire_days": 365,
  "max_bind": 3
}
```

#### 续期

```json
{
  "id": 1,
  "expire_days": 365
}
```

## 安全设计

- **HMAC-SHA256 签名** — 客户端用 `client_secret` 对请求参数签名，防篡改
- **时间戳防重放** — 请求携带 timestamp，服务端校验 ±5分钟有效窗口
- **常量时间比较** — 签名比较使用常量时间算法，防时序攻击
- **机器码绑定** — 首次验证自动绑定，换绑有次数限制
- **JWT 认证** — 管理接口使用 JWT Token 认证，支持多用户角色
- **角色权限** — 超管(0)/授权商(1)/普通用户(2) 三级权限控制
- **操作日志** — 所有关键操作记录 IP 和详情
- **IP 速率限制** — 每 IP 每分钟 30 次请求
- **SQLite WAL 模式** — 并发安全，数据完整性保障

## 新增功能

### 批量操作
- 批量创建授权码 (1-100 个)
- 批量吊销/续期（前端多选 checkbox）
- 配额检查（授权商不能超出分配配额）

### 试用授权
- 客户端可自助申请试用码
- 每个机器码只能试用一次（7 天）
- 试用码标记 `is_trial=1`，客户端可据此展示“试用版”

### 公告推送
- 管理员发布公告，客户端通过 verify/heartbeat 自动收到
- 支持信息/警告/紧急三种类型
- 可按套餐定向推送 (`target: plan:standard`)
- 支持生效/过期时间设置

### 数据导出
- 授权码列表导出为 CSV 文件
- 支持按状态筛选导出

### 统计报表
- 30 天趋势数据（每日新增/过期/在线）
- 套餐分布统计

### 前端增强
- 授权码编辑弹窗
- 点击授权码复制到剪贴板
- 试用标记展示
- 公告管理页面（仅超管可见）

## 数据迁移

如果项目目录下存在 `licenses_db.json`（旧版授权数据），启动时会自动迁移到 SQLite 并重命名为 `.bak`。
