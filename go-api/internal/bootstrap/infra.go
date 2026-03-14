package bootstrap

import (
	"go-api/internal/app"
	"go-api/internal/cache"
	"go-api/internal/config"
	"go-api/internal/database"
	"go-api/internal/license"
	ordermodule "go-api/internal/modules/order"
	"go-api/internal/queue"
	"go-api/internal/ws"
)

// BuildApp 初始化应用级依赖并装配为 App。
func BuildApp(cfg *config.Config) *app.App {
	db := database.Connect(cfg.Database)
	rdb := cache.Connect(cfg.Redis)
	lm := license.NewManager(cfg.License)
	license.Global = lm

	dockChecker := func(oid int64) bool {
		var ds int
		err := database.DB.QueryRow("SELECT dockstatus FROM qingka_wangke_order WHERE oid = ?", oid).Scan(&ds)
		return err == nil && ds == 1
	}

	dockQueue := queue.NewDockQueue(5, 1000, func(oid int64) {
		if oid <= 0 {
			return
		}
		_, _, _ = ordermodule.NewServices().Sync.ManualDock([]int{int(oid)})
	}, dockChecker)
	queue.GlobalDockQueue = dockQueue

	return &app.App{
		Config:    cfg,
		DB:        db,
		Redis:     rdb,
		License:   lm,
		Hub:       ws.NewHub(),
		DockQueue: dockQueue,
	}
}
