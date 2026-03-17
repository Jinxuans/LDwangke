# Go-API 项目优化建议

## 目录

1. [安全问题](#安全问题)
2. [性能优化](#性能优化)
3. [代码质量](#代码质量)
4. [架构改进](#架构改进)
5. [实施计划](#实施计划)

---

## 安全问题

### 1. 密码明文存储 [高优先级]

**问题描述**：
`internal/service/auth.go` 中密码使用明文比对，存在严重安全隐患。

**当前代码**：
```go
if user.Pass != req.Password {
    return nil, "", errors.New("密码错误")
}
```

**解决方案**：
使用 `golang.org/x/crypto/bcrypt` 进行密码加密：

```go
import "golang.org/x/crypto/bcrypt"

// 注册时加密密码
hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// 登录时验证
err := bcrypt.CompareHashAndPassword([]byte(user.Pass), []byte(req.Password))
if err != nil {
    return nil, "", errors.New("密码错误")
}
```

**迁移方案**：
1. 新增 `password_hash` 字段存储加密密码
2. 用户登录时，如果 `password_hash` 为空，用旧逻辑验证后自动迁移
3. 迁移完成后删除明文密码字段

---

### 2. 敏感信息暴露 [高优先级]

**问题描述**：
`config/config.yaml` 包含明文敏感信息，不应提交到版本控制。

**解决方案**：

方案一：使用环境变量
```yaml
database:
  host: ${DB_HOST:127.0.0.1}
  password: ${DB_PASSWORD}
jwt:
  secret: ${JWT_SECRET}
```

方案二：使用 `.env` 文件（配合 godotenv 库）

```go
import "github.com/joho/godotenv"

func init() {
    godotenv.Load()
}
```

**必须操作**：
- 将 `config.yaml` 加入 `.gitignore`
- 创建 `config.yaml.example` 作为模板

---

### 3. MySQL 严格模式关闭 [中优先级]

**问题描述**：
`internal/database/mysql.go` 关闭了 SQL 严格模式，可能导致数据完整性问题。

**当前代码**：
```go
mysqlCfg.Params = map[string]string{"sql_mode": "''"}
```

**解决方案**：
1. 启用严格模式
2. 修复现有数据问题
3. 确保所有 NOT NULL 字段有默认值

---

### 4. 缺少安全 Headers [中优先级]

**解决方案**：
添加安全中间件 `internal/middleware/security.go`：

```go
package middleware

import "github.com/gin-gonic/gin"

func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Next()
    }
}
```

---

### 5. JWT 安全增强 [中优先级]

**建议改进**：
- 增加 Token 黑名单（登出时失效）
- 增加 JWT ID (jti) 防止重放攻击
- 缩短 Access Token 有效期

```go
type Claims struct {
    UID   int    `json:"uid"`
    User  string `json:"user"`
    Grade string `json:"grade"`
    jti   string `json:"jti"`  // 唯一标识
    jwt.RegisteredClaims
}
```

---

## 性能优化

### 1. 分布式限流 [中优先级]

**问题描述**：
`internal/middleware/ratelimit.go` 使用内存存储，多实例部署时无法共享限流状态。

**解决方案**：
使用 Redis 实现滑动窗口限流：

```go
func (rl *RedisRateLimiter) Allow(key string) bool {
    ctx := context.Background()
    now := time.Now().UnixNano()
    windowStart := now - int64(rl.window)

    pipe := rl.rdb.Pipeline()
    pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
    pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
    pipe.ZCard(ctx, key)
    pipe.Expire(ctx, key, rl.window)

    cmds, _ := pipe.Exec(ctx)
    count := cmds[2].(*redis.IntCmd).Val()
    return count <= int64(rl.limit)
}
```

---

### 2. 数据库查询优化

**建议添加索引**：

```sql
-- 订单表
CREATE INDEX idx_order_uid ON qingka_wangke_order(uid);
CREATE INDEX idx_order_status ON qingka_wangke_order(status);
CREATE INDEX idx_order_dockstatus ON qingka_wangke_order(dockstatus);
CREATE INDEX idx_order_created ON qingka_wangke_order(created_at);

-- 用户表
CREATE INDEX idx_user_grade ON qingka_wangke_user(grade);
CREATE INDEX idx_user_active ON qingka_wangke_user(active);
```

**使用 Context 控制超时**：

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := database.DB.QueryRowContext(ctx, query, args...).Scan(...)
```

---

### 3. 缓存策略

**建议增加缓存**：

| 数据类型 | 缓存时间 | 说明 |
|---------|---------|------|
| 用户信息 | 5 分钟 | 减少登录验证查询 |
| 分类数据 | 30 分钟 | 分类变更频率低 |
| 系统配置 | 10 分钟 | 配置读取频繁 |
| 商品列表 | 1 分钟 | 高频访问 |

**示例实现**：

```go
func (s *AuthService) GetUserInfo(uid int) (*model.VbenUserInfo, error) {
    ctx := context.Background()
    cacheKey := fmt.Sprintf("user:info:%d", uid)

    // 尝试从缓存获取
    cached, err := cache.RDB.Get(ctx, cacheKey).Result()
    if err == nil {
        var info model.VbenUserInfo
        json.Unmarshal([]byte(cached), &info)
        return &info, nil
    }

    // 从数据库查询
    info, err := s.getUserInfoFromDB(uid)
    if err != nil {
        return nil, err
    }

    // 写入缓存
    data, _ := json.Marshal(info)
    cache.RDB.Set(ctx, cacheKey, data, 5*time.Minute)

    return info, nil
}
```

---

### 4. 连接池监控

```go
// 添加数据库连接池监控
func monitorDBStats() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        stats := database.DB.Stats()
        log.Printf("[DB Pool] OpenConnections: %d, InUse: %d, Idle: %d, WaitCount: %d",
            stats.OpenConnections, stats.InUse, stats.Idle, stats.WaitCount)
    }
}
```

---

## 代码质量

### 1. 优雅关闭

**添加到 `cmd/server/main.go`**：

```go
// 优雅关闭
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

go func() {
    <-quit
    log.Println("正在关闭服务...")

    // 停止队列
    queue.GlobalDockQueue.Stop()

    // 关闭数据库连接
    database.DB.Close()

    // 关闭 Redis
    cache.RDB.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := r.Shutdown(ctx); err != nil {
        log.Fatal("服务关闭错误:", err)
    }
    log.Println("服务已关闭")
}()
```

---

### 2. 健康检查端点

```go
r.GET("/health", func(c *gin.Context) {
    // 检查数据库
    if err := database.DB.Ping(); err != nil {
        c.JSON(503, gin.H{"status": "unhealthy", "error": "database"})
        return
    }

    // 检查 Redis
    if err := cache.RDB.Ping(context.Background()).Err(); err != nil {
        c.JSON(503, gin.H{"status": "unhealthy", "error": "redis"})
        return
    }

    c.JSON(200, gin.H{
        "status": "healthy",
        "time":   time.Now().Format(time.RFC3339),
    })
})
```

---

### 3. 结构化日志

**使用 zap 替代 log.Printf**：

```go
import "go.uber.org/zap"

var logger *zap.Logger

func init() {
    logger, _ = zap.NewProduction()
    defer logger.Sync()
}

// 使用
logger.Info("用户登录",
    zap.Int("uid", uid),
    zap.String("ip", c.ClientIP()),
)
```

---

### 4. 统一错误处理

```go
package errors

type AppError struct {
    Code    int
    Message string
    Err     error
}

func (e *AppError) Error() string {
    return e.Message
}

func New(code int, message string) *AppError {
    return &AppError{Code: code, Message: message}
}

// 预定义错误
var (
    ErrUserNotFound     = New(1001, "用户不存在")
    ErrInvalidPassword  = New(1002, "密码错误")
    ErrUserDisabled     = New(1003, "账号已被禁用")
)
```

---

### 5. 请求日志中间件

```go
func RequestLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path

        c.Next()

        latency := time.Since(start)
        status := c.Writer.Status()

        logger.Info("HTTP Request",
            zap.String("method", c.Request.Method),
            zap.String("path", path),
            zap.Int("status", status),
            zap.Duration("latency", latency),
            zap.String("ip", c.ClientIP()),
        )
    }
}
```

---

## 架构改进

### 1. 项目结构优化

```
go-api/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/          # 配置管理
│   ├── handler/         # HTTP 处理器
│   ├── middleware/      # 中间件
│   ├── model/           # 数据模型
│   ├── repository/      # 数据访问层 (新增)
│   ├── service/         # 业务逻辑层
│   ├── cache/           # 缓存
│   ├── queue/           # 队列
│   └── response/        # 响应格式
├── pkg/                 # 可复用的公共包 (新增)
│   ├── errors/          # 错误处理
│   ├── logger/          # 日志
│   └── validator/       # 验证器
├── migrations/
│   └── core/            # 核心数据库迁移
├── docs/                # API 文档
└── scripts/             # 脚本
```

---

### 2. 添加 Repository 层

分离数据访问逻辑：

```go
// internal/repository/user.go
type UserRepository interface {
    FindByID(ctx context.Context, uid int) (*model.User, error)
    FindByUsername(ctx context.Context, username string) (*model.User, error)
    Update(ctx context.Context, user *model.User) error
}

type userRepository struct {
    db *sql.DB
}

func (r *userRepository) FindByID(ctx context.Context, uid int) (*model.User, error) {
    // 实现查询逻辑
}
```

---

### 3. 添加 Swagger 文档

```go
// 安装 swag
// go install github.com/swaggo/swag/cmd/swag@latest

import "github.com/swaggo/gin-swagger"

// @title Go API
// @version 1.0
// @description API 服务

// @Router /api/v1/auth/login [post]
func Login(c *gin.Context) {
    // ...
}
```

---

### 4. 添加 Prometheus 指标

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
}

// 添加指标端点
r.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

---

## 实施计划

### 第一阶段：安全加固 (1-2 周)

| 任务 | 优先级 | 预估时间 |
|-----|-------|---------|
| 密码加密迁移 | 高 | 2 天 |
| 敏感配置环境变量化 | 高 | 1 天 |
| 添加安全 Headers | 中 | 0.5 天 |
| JWT 黑名单机制 | 中 | 1 天 |

### 第二阶段：性能优化 (1-2 周)

| 任务 | 优先级 | 预估时间 |
|-----|-------|---------|
| Redis 分布式限流 | 中 | 1 天 |
| 添加数据库索引 | 中 | 0.5 天 |
| 缓存策略实现 | 中 | 2 天 |
| 连接池监控 | 低 | 0.5 天 |

### 第三阶段：代码质量 (1 周)

| 任务 | 优先级 | 预估时间 |
|-----|-------|---------|
| 优雅关闭 | 中 | 0.5 天 |
| 健康检查端点 | 中 | 0.5 天 |
| 结构化日志 | 低 | 1 天 |
| 统一错误处理 | 低 | 1 天 |
| 单元测试 | 中 | 2 天 |

### 第四阶段：架构改进 (持续)

| 任务 | 优先级 | 预估时间 |
|-----|-------|---------|
| Repository 层分离 | 低 | 3 天 |
| Swagger 文档 | 低 | 1 天 |
| Prometheus 监控 | 低 | 1 天 |

---

## 检查清单

### 安全检查
- [ ] 密码使用 bcrypt 加密
- [ ] 敏感配置使用环境变量
- [ ] SQL 注入防护（使用参数化查询）
- [ ] XSS 防护（安全 Headers）
- [ ] CSRF 防护
- [ ] Rate Limiting
- [ ] JWT 安全配置

### 性能检查
- [ ] 数据库索引优化
- [ ] Redis 缓存策略
- [ ] 连接池配置
- [ ] 慢查询监控

### 代码质量检查
- [ ] 单元测试覆盖率 > 60%
- [ ] 集成测试
- [ ] 代码规范检查 (golangci-lint)
- [ ] 错误处理统一
- [ ] 日志结构化

### 运维检查
- [ ] 健康检查端点
- [ ] 优雅关闭
- [ ] 监控指标
- [ ] 日志收集
- [ ] 告警配置

---

## 参考资料

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go 安全最佳实践](https://github.com/guardrailsio/awesome-golang-security)
- [Gin 框架文档](https://gin-gonic.com/docs/)
- [bcrypt 使用指南](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
