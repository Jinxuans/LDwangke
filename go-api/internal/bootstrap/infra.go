package bootstrap

import (
	"go-api/internal/app"
	"go-api/internal/cache"
	"go-api/internal/config"
	"go-api/internal/database"
	"go-api/internal/license"
	"go-api/internal/ws"
)

// BuildApp 初始化应用级依赖并装配为 App。
func BuildApp(cfg *config.Config) *app.App {
	db := database.Connect(cfg.Database)
	rdb := cache.Connect(cfg.Redis)
	lm := license.NewManager(cfg.License)
	license.Global = lm

	return &app.App{
		Config:  cfg,
		DB:      db,
		Redis:   rdb,
		License: lm,
		Hub:     ws.NewHub(),
	}
}
