package bootstrap

import (
	"go-api/internal/config"
	"go-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// NewEngine 初始化 HTTP 引擎与全局中间件。
func NewEngine(cfg *config.Config) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS())
	r.Use(middleware.DemoGuard())
	r.Static("/uploads", "./uploads")
	return r
}
