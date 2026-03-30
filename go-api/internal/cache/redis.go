package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go-api/internal/config"
	obslogger "go-api/internal/observability/logger"
)

var RDB *redis.Client

func Connect(cfg config.RedisConfig) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		obslogger.Fatal("Redis 连接失败", "addr", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), "error", err)
	}

	RDB = rdb
	obslogger.L().Info("Redis 连接成功", "addr", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), "db", cfg.DB)
	return rdb
}
