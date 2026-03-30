package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	obslogger "go-api/internal/observability/logger"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

// RecoveryWithRequestID 捕获 panic 并记录 request_id，方便关联请求日志与堆栈。
func RecoveryWithRequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				reqID := c.GetString("request_id")
				if reqID == "" {
					reqID = c.GetHeader("X-Request-ID")
				}
				if reqID == "" {
					reqID = "-"
				}

				err := fmt.Errorf("panic: %v", rec)
				_ = c.Error(err).SetType(gin.ErrorTypePrivate)

				obslogger.Request(c).Error("panic recovered",
					slog.String("request_id", reqID),
					slog.Any("panic", rec),
					slog.String("stack", string(debug.Stack())),
				)

				if !c.Writer.Written() {
					response.Error(c, http.StatusInternalServerError, 500, "服务器内部错误")
				} else {
					c.Abort()
				}
			}
		}()

		c.Next()
	}
}
