package app

import (
	"database/sql"

	"go-api/internal/config"
	"go-api/internal/license"
	"go-api/internal/ws"

	"github.com/redis/go-redis/v9"
)

// App 收纳应用启动后需要复用的核心依赖。
type App struct {
	Config  *config.Config
	DB      *sql.DB
	Redis   *redis.Client
	License *license.Manager
	Hub     *ws.Hub
}
