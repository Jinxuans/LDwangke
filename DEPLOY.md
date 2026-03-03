# 部署教程

## 环境要求

- **Linux** CentOS 7+ / Ubuntu 20.04+
- **Go** 1.21+（仅编译需要，也可本地交叉编译后上传二进制）
- **MySQL** 5.7+ / 8.0
- **Redis** 6.0+
- **Nginx** 用于反向代理和前端托管
- **Node.js** 18+（仅前端构建需要）

---

## 一、后端部署（Go API）

### 1.1 本地交叉编译（推荐）

在 Windows 本地编译 Linux 二进制，免去服务器装 Go：

```powershell
cd d:\hzw1\go-api
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o server ./cmd/server/
```

生成的 `server` 文件上传到服务器。

### 1.2 服务器上编译

```bash
cd /opt/go-api
go build -o server ./cmd/server/
```

### 1.3 上传文件

将以下文件/目录上传到服务器（如 `/opt/go-api/`）：

```
server              # 编译好的二进制
config/config.yaml  # 配置文件
migrations/         # 数据库迁移脚本（如需要）
```

### 1.4 修改配置

编辑 `config/config.yaml`：

```yaml
server:
  port: 8080
  mode: release    # 生产环境改为 release

database:
  host: 127.0.0.1
  port: 3306
  user: your_db_user
  password: "your_db_password"
  dbname: "your_db_name"
  max_open_conns: 50
  max_idle_conns: 25

redis:
  host: 127.0.0.1
  port: 6379
  password: "your_redis_password"
  db: 0

jwt:
  secret: "改成一个随机的长字符串"   # 必须修改！
  access_ttl: 7200
  refresh_ttl: 604800

cache:
  order_list_ttl: 30
  class_list_ttl: 300
```

### 1.5 创建 systemd 服务

```bash
sudo vim /etc/systemd/system/go-api.service
```

写入：

```ini
[Unit]
Description=Go API Server
After=network.target mysql.service redis.service

[Service]
Type=simple
User=www
WorkingDirectory=/opt/go-api
ExecStart=/opt/go-api/server
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

启动：

```bash
sudo systemctl daemon-reload
sudo systemctl enable go-api
sudo systemctl start go-api
sudo systemctl status go-api    # 查看状态
journalctl -u go-api -f         # 查看日志
```

---

## 二、前端部署（Vben Admin）

### 2.1 构建前端

在本地或服务器上：

```bash
cd vben-admin

# 安装依赖
pnpm install

# 修改 API 地址（.env.production）
# VITE_GLOB_API_URL=https://你的域名/api/v1

# 构建
pnpm build:antd
```

构建产物在 `apps/web-antd/dist/` 目录。

### 2.2 上传到服务器

将 `dist/` 目录内容上传到服务器，如 `/var/www/admin/`。

---

## 三、Nginx 配置

```bash
sudo vim /etc/nginx/conf.d/admin.conf
```

```nginx
server {
    listen 80;
    server_name your-domain.com;

    # 前端静态文件
    root /var/www/admin;
    index index.html;

    # 前端路由（SPA history模式）
    location / {
        try_files $uri $uri/ /index.html;
    }

    # PHP兼容路由代理（下游系统通过 /api.php?act=xxx 调用）
    # 必须用 = 精确匹配，优先级高于宝塔默认的 ~ \.php$ 规则
    location = /api.php {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 60s;
        proxy_read_timeout 120s;
    }

    # API 反向代理（含 /api/index.php 兼容路由）
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 60s;
        proxy_read_timeout 120s;
    }

    # WebSocket 代理
    location /ws/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 3600s;
    }

    # 龙龙平台实时日志代理（可选）
    location /api/streamLogs {
        proxy_buffering off;
        set $args $args&key=你的access-key;
        proxy_pass http://122.51.236.86/api/streamLogs;
    }
}
```

```bash
sudo nginx -t
sudo systemctl reload nginx
```

### HTTPS（推荐）

```bash
# 安装 certbot
sudo apt install certbot python3-certbot-nginx   # Ubuntu
sudo yum install certbot python3-certbot-nginx   # CentOS

# 自动配置 SSL
sudo certbot --nginx -d your-domain.com
```

---

## 四、龙龙平台商品同步

龙龙平台不支持 API 拉取商品，需要使用官方 CLI 工具同步：

### 4.1 安装 long 工具

```bash
wget http://122.51.236.86/long -O long && chmod +x long && mv -f long /usr/bin/long
```

### 4.2 同步商品

```bash
long sync \
    --long-host=122.51.236.86 \
    --access-key=你的access-key \
    --mysql-user=数据库用户名 \
    --mysql-password=数据库密码 \
    --mysql-database=数据库名 \
    --docking=龙龙供应商的hid \
    --rate=1.5 \
    --name-prefix="" \
    --sort=10
```

参数说明：
- `--docking` — 龙龙供应商在 `qingka_wangke_huoyuan` 表中的 hid
- `--rate` — 定价倍率（售价 = 成本 × rate）
- `--name-prefix` — 商品名前缀（可选）
- `--sort` — 排序值（可选）

### 4.3 监听订单状态变动（实时进度）

```bash
nohup long listen \
    --long-host=122.51.236.86 \
    --access-key=你的access-key \
    --mysql-user=数据库用户名 \
    --mysql-password=数据库密码 \
    --mysql-database=数据库名 \
    > /var/log/long-listen.log 2>&1 &
```

> **注意**：如果同时使用 `long listen` 和 Go API 的 AutoSync 进度同步，两者会共存互补。`long listen` 是实时推送，AutoSync 是每2分钟轮询兜底。

---

## 五、数据库初始化

确保 MySQL 已导入基础表结构。如果是从旧系统迁移，直接导入旧数据库即可。

如果是全新部署，导入迁移脚本：

```bash
cd /opt/go-api
mysql -u root -p your_db_name < migrations/004_create_dynamic_module_table.sql
mysql -u root -p your_db_name < migrations/003_create_mail_table.sql
# ... 其他迁移脚本
```

---

## 六、防火墙

```bash
# 只开放必要端口
sudo firewall-cmd --permanent --add-port=80/tcp
sudo firewall-cmd --permanent --add-port=443/tcp
sudo firewall-cmd --reload

# 8080 端口不要对外开放，走 Nginx 代理
```

---

## 七、常用运维命令

```bash
# 查看 API 服务状态
sudo systemctl status go-api

# 重启 API 服务
sudo systemctl restart go-api

# 查看实时日志
journalctl -u go-api -f

# 更新后端（上传新的 server 二进制后）
sudo systemctl restart go-api

# 更新前端（上传新的 dist 后）
# 无需重启，Nginx 直接读取静态文件

# 查看龙龙监听进程
ps aux | grep long

# 停止龙龙监听
pkill -f long
```

---

## 八、部署检查清单

- [ ] MySQL 已安装并创建数据库
- [ ] Redis 已安装并启动
- [ ] `config.yaml` 已修改为生产配置
- [ ] JWT secret 已改为随机字符串
- [ ] server 模式改为 `release`
- [ ] 防火墙只开放 80/443
- [ ] Nginx 已配置反向代理
- [ ] HTTPS 已配置
- [ ] Go API 服务已启动
- [ ] 前端已构建并部署
- [ ] 龙龙 IP 白名单已添加服务器公网IP
- [ ] `long sync` 已同步商品
- [ ] `long listen` 已启动（实时进度）
- [ ] 测试下单和进度同步正常
