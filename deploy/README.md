# 一键部署指南

## 使用方式

### 第一步：本地打包（Windows）

双击运行 `deploy/pack.bat`，它会自动：
1. 编译 Go 后端（Linux 版本）
2. 构建 Vue 前端
3. 打包所有文件为 `deploy/release.tar.gz`

### 第二步：上传到服务器

将 `release.tar.gz` 上传到服务器（通过宝塔文件管理或 scp）：

```bash
scp deploy/release.tar.gz root@你的服务器IP:/opt/
```

### 第三步：服务器执行安装

SSH 登录服务器后执行：

```bash
cd /opt
mkdir -p deploy && cd deploy
tar -xzf /opt/release.tar.gz
bash install.sh
```

脚本会交互式询问数据库密码等信息，然后自动完成：
- 创建数据库 + 执行迁移
- 部署 Go 后端 + 自动生成配置
- 部署 PHP API + 自动生成配置
- 部署前端静态文件
- 配置 Nginx 反向代理
- 创建 systemd 服务并启动

### 前置条件

服务器需要先安装宝塔面板，并在软件商店安装：
- Nginx
- MySQL 5.7+
- Redis 6.0+
- PHP 7.4+（如需 PHP 插件模块）

### 部署后管理

```bash
# 查看后端状态
systemctl status go-api

# 重启后端
systemctl restart go-api

# 查看日志
journalctl -u go-api -f
```

### 更新部署

重新运行 `pack.bat` 打包，上传后在服务器执行：

```bash
cd /opt/deploy
tar -xzf /opt/release.tar.gz
bash install.sh
```

脚本会自动覆盖旧文件并重启服务。
