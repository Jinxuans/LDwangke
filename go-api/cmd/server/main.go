package main

import (
	"os"

	"go-api/internal/bootstrap"
	obslogger "go-api/internal/observability/logger"
	"go-api/internal/routes"
)

func main() {
	cfg := bootstrap.LoadConfig("config/config.yaml")
	obslogger.Init("go-api", cfg.Server.Mode)
	application := bootstrap.BuildApp(cfg)
	bootstrap.InitTables()

	ctx, stop := bootstrap.NotifyContext()
	defer stop()

	bootstrap.StartPendingDockScheduler(ctx)
	bootstrap.StartCoreJobs(ctx, application)
	bootstrap.StartChatCleanup(ctx)

	r := bootstrap.NewEngine(cfg)
	routes.RegisterAll(r, application.Hub)
	srv := bootstrap.NewHTTPServer(cfg, r)

	if err := bootstrap.Serve(ctx, srv, application); err != nil {
		obslogger.L().Error("服务启动失败", "error", err)
		os.Exit(1)
	}
}
