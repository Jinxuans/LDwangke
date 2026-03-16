package admin

import (
	"go-api/internal/database"
	"go-api/internal/runtimeops"
)

type TurboProfile struct {
	Name                   string `json:"name"`
	CPUCores               int    `json:"cpu_cores"`
	MemTotalMB             int    `json:"mem_total_mb"`
	GOOS                   string `json:"goos"`
	GOARCH                 string `json:"goarch"`
	DBMaxOpen              int    `json:"db_max_open"`
	DBMaxIdle              int    `json:"db_max_idle"`
	DBMaxLifetime          int    `json:"db_max_lifetime_sec"`
	DBMaxIdleTime          int    `json:"db_max_idle_time_sec"`
	RedisPoolSize          int    `json:"redis_pool_size"`
	RedisMinIdle           int    `json:"redis_min_idle"`
	DockBatchLimit         int    `json:"dock_batch_limit"`
	PendingDockIntervalSec int    `json:"pending_dock_interval_sec"`
	SyncIntervalSec        int    `json:"sync_interval_sec"`
	GOMAXPROCS             int    `json:"gomaxprocs"`
	GCPercent              int    `json:"gc_percent"`
}

type TurboStatus struct {
	Enabled   bool         `json:"enabled"`
	Profile   TurboProfile `json:"profile"`
	AppliedAt string       `json:"applied_at"`
	Baseline  TurboProfile `json:"baseline"`
}

func getAdminTurboStatus() TurboStatus {
	return mapTurboStatus(runtimeops.GetTurboStatus())
}

func applyAdminTurbo(mode string) TurboStatus {
	return mapTurboStatus(runtimeops.ApplyTurbo(mode))
}

func getAdminHZWSocketURL() string {
	var socketURL string
	database.DB.QueryRow(
		"SELECT svalue FROM qingka_wangke_config WHERE skey = 'hzw_socket_url' LIMIT 1",
	).Scan(&socketURL)
	return socketURL
}

func setAdminHZWSocketURL(socketURL string) error {
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_config (v, k, skey, svalue) VALUES ('hzw_socket_url', '', 'hzw_socket_url', ?) ON DUPLICATE KEY UPDATE svalue = ?",
		socketURL, socketURL,
	)
	return err
}

func mapTurboStatus(status runtimeops.TurboStatus) TurboStatus {
	return TurboStatus{
		Enabled:   status.Enabled,
		Profile:   mapTurboProfile(status.Profile),
		AppliedAt: status.AppliedAt,
		Baseline:  mapTurboProfile(status.Baseline),
	}
}

func mapTurboProfile(profile runtimeops.TurboProfile) TurboProfile {
	return TurboProfile{
		Name:                   profile.Name,
		CPUCores:               profile.CPUCores,
		MemTotalMB:             profile.MemTotalMB,
		GOOS:                   profile.GOOS,
		GOARCH:                 profile.GOARCH,
		DBMaxOpen:              profile.DBMaxOpen,
		DBMaxIdle:              profile.DBMaxIdle,
		DBMaxLifetime:          profile.DBMaxLifetime,
		DBMaxIdleTime:          profile.DBMaxIdleTime,
		RedisPoolSize:          profile.RedisPoolSize,
		RedisMinIdle:           profile.RedisMinIdle,
		DockBatchLimit:         profile.DockBatchLimit,
		PendingDockIntervalSec: profile.PendingDockIntervalSec,
		SyncIntervalSec:        profile.SyncIntervalSec,
		GOMAXPROCS:             profile.GOMAXPROCS,
		GCPercent:              profile.GCPercent,
	}
}
