# 登录内部服务器错误诊断报告

## 可能的错误原因

### 1. 数据库连接问题
**最常见原因** - 数据库连接失败或超时
- 检查数据库服务是否正常运行
- 验证数据库连接配置（host、port、用户名、密码）
- 查看数据库连接池是否耗尽

### 2. 数据库表结构问题
**字段缺失** - `pass2` 字段可能不存在
```go
// auth.go 第26行有 fallback 处理，但可能还有其他字段问题
err := database.DB.QueryRow(
    "SELECT uid, uuid, user, pass, IFNULL(pass2,''), name, money, grade, active FROM qingka_wangke_user WHERE user = ?",
    req.Username,
).Scan(&user.UID, &user.UUID, &user.User, &user.Pass, &user.Pass2, &user.Name, &user.Money, &user.Grade, &user.Active)
```

### 3. 配置表查询失败
**系统配置读取异常** - `qingka_wangke_config` 表可能不存在或数据异常
```go
// auth.go 第60行查询 pass2_kg 配置
database.DB.QueryRow("SELECT `k` FROM qingka_wangke_config WHERE `v`='pass2_kg'").Scan(&pass2Kg)
```

### 4. JWT 密钥未配置
**Token 生成失败** - JWT secret 未正确配置
```go
// auth.go 第89行生成 token
accessToken, err := s.generateToken(user, config.Global.JWT.AccessTTL)
```

### 5. Redis 连接问题
**缓存服务异常** - 设备检测功能依赖 Redis
```go
// auth.go 第267行使用 Redis
cache.RDB.SIsMember(ctx, deviceKey, fingerprint).Result()
```

## 排查步骤

### 步骤 1: 检查日志
```bash
# 查看 Go API 日志
cd 29-colnt-com/go-api
tail -f logs/app.log

# 或查看系统日志
journalctl -u go-api -f
```

### 步骤 2: 检查数据库连接
```bash
# 测试数据库连接
cd 29-colnt-com/go-api
./tools/diagnose.sh
```

### 步骤 3: 验证数据库表结构
```sql
-- 检查 user 表结构
DESC qingka_wangke_user;

-- 检查是否有 pass2 字段
SHOW COLUMNS FROM qingka_wangke_user LIKE 'pass2';

-- 检查配置表
SELECT * FROM qingka_wangke_config WHERE v = 'pass2_kg';
```

### 步骤 4: 检查配置文件
```bash
# 查看 Go API 配置
cat 29-colnt-com/go-api/config/config.yaml

# 重点检查：
# - database.host
# - database.port
# - database.username
# - database.password
# - database.database
# - jwt.secret
# - redis.addr
```

### 步骤 5: 测试登录接口
```bash
# 使用 curl 测试登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'
```

## 快速修复方案

### 方案 1: 添加缺失的 pass2 字段
```sql
ALTER TABLE qingka_wangke_user 
ADD COLUMN pass2 VARCHAR(255) DEFAULT '' COMMENT '二级密码';
```

### 方案 2: 添加配置表数据
```sql
INSERT INTO qingka_wangke_config (v, k) 
VALUES ('pass2_kg', '0') 
ON DUPLICATE KEY UPDATE k = '0';
```

### 方案 3: 增强错误处理
在 `auth.go` 的 Login 方法中添加更详细的错误日志：

```go
if err != nil {
    log.Printf("登录失败 - 用户: %s, 错误: %v", req.Username, err)
    return nil, "", fmt.Errorf("查询用户失败: %v", err)
}
```

### 方案 4: 临时禁用二级密码验证
```sql
UPDATE qingka_wangke_config 
SET k = '0' 
WHERE v = 'pass2_kg';
```

### 方案 5: 检查并修复数据库连接池
在 `config.yaml` 中调整连接池参数：
```yaml
database:
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600
```

## 监控建议

### 1. 添加健康检查接口
```go
// 在 handler 中添加
func HealthCheck(c *gin.Context) {
    // 检查数据库
    if err := database.DB.Ping(); err != nil {
        c.JSON(500, gin.H{"status": "error", "db": "disconnected"})
        return
    }
    
    // 检查 Redis
    if _, err := cache.RDB.Ping(context.Background()).Result(); err != nil {
        c.JSON(500, gin.H{"status": "error", "redis": "disconnected"})
        return
    }
    
    c.JSON(200, gin.H{"status": "ok"})
}
```

### 2. 添加详细的错误日志
在关键位置添加日志记录，便于排查问题

### 3. 设置告警
- 数据库连接失败告警
- 登录失败率超过阈值告警
- 服务响应时间过长告警

## 常见错误码对照

| 错误信息 | 可能原因 | 解决方案 |
|---------|---------|---------|
| "查询用户失败" | 数据库连接问题 | 检查数据库服务和配置 |
| "用户不存在" | 用户名错误 | 确认用户名正确 |
| "密码错误" | 密码不匹配 | 确认密码正确 |
| "账号已被禁用" | active != '1' | 检查用户状态 |
| "生成 Token 失败" | JWT 配置问题 | 检查 jwt.secret |
| "NEED_ADMIN_AUTH" | 需要二级密码 | 提供管理员二级密码 |

## 预防措施

1. **定期备份数据库**
2. **监控数据库连接数**
3. **设置合理的超时时间**
4. **使用连接池管理**
5. **添加重试机制**
6. **完善错误处理和日志**

## 联系支持

如果以上方案都无法解决问题，请提供：
1. 完整的错误日志
2. 数据库版本和配置
3. Go API 版本
4. 系统环境信息
