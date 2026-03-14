package main

import (
	"log"

	"go-api/internal/bootstrap"
	"go-api/internal/routes"
)

func main() {
	cfg := bootstrap.LoadConfig("config/config.yaml")
	application := bootstrap.BuildApp(cfg)
	bootstrap.InitTables()

	ctx, stop := bootstrap.NotifyContext()
	defer stop()

	bootstrap.StartDockQueue(ctx, application)
	bootstrap.StartCoreJobs(ctx, application)
	bootstrap.StartChatCleanup(ctx)

	r := bootstrap.NewEngine(cfg)
	routes.RegisterAll(r, application.Hub)
	srv := bootstrap.NewHTTPServer(cfg, r)

	if err := bootstrap.Serve(ctx, srv, application); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
