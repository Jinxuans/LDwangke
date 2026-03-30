package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	obslogger "go-api/internal/observability/logger"
)

func generateRequestID() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// RequestLogger 记录每个请求的追踪日志，包含 request_id、方法、路径、状态码、耗时和 uid。
// request_id 写入响应头 X-Request-ID，方便前端/运维关联日志。
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if reqID == "" {
			reqID = generateRequestID()
		}
		start := time.Now()

		c.Set("request_id", reqID)
		c.Header("X-Request-ID", reqID)

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		// 尝试从 context 取 uid（JWTAuth 已写入），未认证接口为 0
		uid, _ := c.Get("uid")
		uidVal := 0
		if uid != nil {
			if v, ok := uid.(int); ok {
				uidVal = v
			}
		}

		// 收集本次请求中通过 c.Error() 记录的错误
		errors := c.Errors.ByType(gin.ErrorTypePrivate).String()

		logger := obslogger.Request(c).With(
			slog.Int("status", status),
			slog.Int64("duration_ms", duration.Milliseconds()),
			slog.Int("uid", uidVal),
		)
		if errors != "" {
			logger = logger.With(slog.String("errors", errors))
		}

		if status >= 500 {
			logger.Error("request completed")
		} else if status >= 400 {
			logger.Warn("request completed")
		} else {
			logger.Info("request completed")
		}
	}
}
