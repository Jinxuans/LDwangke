package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
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
		reqID := generateRequestID()
		start := time.Now()

		c.Set("request_id", reqID)
		c.Header("X-Request-ID", reqID)

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		if query != "" {
			path = path + "?" + query
		}

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

		if status >= 500 {
			log.Printf("[ERROR] req_id=%s uid=%d %s %s -> %d (%s)%s",
				reqID, uidVal, method, path, status, duration, formatErrors(errors))
		} else if status >= 400 {
			log.Printf("[WARN]  req_id=%s uid=%d %s %s -> %d (%s)%s",
				reqID, uidVal, method, path, status, duration, formatErrors(errors))
		} else {
			log.Printf("[INFO]  req_id=%s uid=%d %s %s -> %d (%s)",
				reqID, uidVal, method, path, status, duration)
		}
	}
}

func formatErrors(s string) string {
	if s == "" {
		return ""
	}
	return " errors=[" + s + "]"
}
