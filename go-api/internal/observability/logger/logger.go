package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

var base = newLogger("go-api", "dev", slog.LevelDebug)

func Init(service, env string) {
	base = newLogger(service, env, levelForEnv(env))
	slog.SetDefault(base)
}

func L() *slog.Logger {
	return base
}

func Fatal(message string, args ...any) {
	base.Error(message, args...)
	os.Exit(1)
}

func Request(c *gin.Context) *slog.Logger {
	reqID := strings.TrimSpace(c.GetString("request_id"))
	if reqID == "" {
		reqID = strings.TrimSpace(c.GetHeader("X-Request-ID"))
	}
	if reqID == "" {
		reqID = "-"
	}

	uid, _ := c.Get("uid")
	return base.With(
		slog.String("request_id", reqID),
		slog.String("method", c.Request.Method),
		slog.String("path", c.Request.URL.RequestURI()),
		slog.Any("uid", uid),
	)
}

func newLogger(service, env string, level slog.Leveler) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			switch attr.Key {
			case slog.TimeKey:
				attr.Key = "ts"
			case slog.MessageKey:
				attr.Key = "message"
			}
			return attr
		},
	})

	return slog.New(handler).With(
		slog.String("service", service),
		slog.String("env", strings.TrimSpace(env)),
	)
}

func levelForEnv(env string) slog.Leveler {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "release", "prod", "production":
		return slog.LevelInfo
	default:
		return slog.LevelDebug
	}
}

func Discard() {
	base = slog.New(slog.NewJSONHandler(io.Discard, nil))
	slog.SetDefault(base)
}
