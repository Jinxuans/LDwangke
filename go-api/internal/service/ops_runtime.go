package service

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"go-api/internal/cache"
	"go-api/internal/database"
	"go-api/internal/queue"
	"go-api/internal/ws"
)

type SystemInfo struct {
	GoVersion     string `json:"go_version"`
	NumCPU        int    `json:"num_cpu"`
	NumGoroutine  int    `json:"num_goroutine"`
	MemAlloc      uint64 `json:"mem_alloc"`
	MemTotalAlloc uint64 `json:"mem_total_alloc"`
	MemSys        uint64 `json:"mem_sys"`
	NumGC         uint32 `json:"num_gc"`
	LastGCPause   uint64 `json:"last_gc_pause_ns"`
	HeapObjects   uint64 `json:"heap_objects"`
	HeapInuse     uint64 `json:"heap_inuse"`
	StackInuse    uint64 `json:"stack_inuse"`
	Uptime        int64  `json:"uptime_seconds"`
	UptimeHuman   string `json:"uptime_human"`
	ServerTime    string `json:"server_time"`
	GOOS          string `json:"goos"`
	GOARCH        string `json:"goarch"`
}

func (s *OpsService) GetSystemInfo() SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(startTime)
	return SystemInfo{
		GoVersion:     runtime.Version(),
		NumCPU:        runtime.NumCPU(),
		NumGoroutine:  runtime.NumGoroutine(),
		MemAlloc:      m.Alloc,
		MemTotalAlloc: m.TotalAlloc,
		MemSys:        m.Sys,
		NumGC:         m.NumGC,
		LastGCPause:   m.PauseNs[(m.NumGC+255)%256],
		HeapObjects:   m.HeapObjects,
		HeapInuse:     m.HeapInuse,
		StackInuse:    m.StackInuse,
		Uptime:        int64(uptime.Seconds()),
		UptimeHuman:   formatDuration(uptime),
		ServerTime:    time.Now().Format("2006-01-02 15:04:05"),
		GOOS:          runtime.GOOS,
		GOARCH:        runtime.GOARCH,
	}
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%d天%d小时%d分钟", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%d小时%d分钟", hours, mins)
	}
	return fmt.Sprintf("%d分钟", mins)
}

type DBHealth struct {
	Status       string `json:"status"`
	OpenConns    int    `json:"open_conns"`
	InUse        int    `json:"in_use"`
	Idle         int    `json:"idle"`
	MaxOpenConns int    `json:"max_open_conns"`
	MaxIdleConns int    `json:"max_idle_conns"`
	PingLatency  int64  `json:"ping_latency_ms"`
	Version      string `json:"version"`
	Uptime       int    `json:"uptime_seconds"`
	Threads      int    `json:"threads"`
	Questions    int64  `json:"questions"`
	SlowQueries  int64  `json:"slow_queries"`
	TableCount   int    `json:"table_count"`
	DBSize       string `json:"db_size_mb"`
}

func (s *OpsService) GetDBHealth() DBHealth {
	h := DBHealth{Status: "unknown"}
	if database.DB == nil {
		h.Status = "disconnected"
		return h
	}

	start := time.Now()
	err := database.DB.Ping()
	h.PingLatency = time.Since(start).Milliseconds()
	if err != nil {
		h.Status = "error"
		return h
	}
	h.Status = "healthy"

	stats := database.DB.Stats()
	h.OpenConns = stats.OpenConnections
	h.InUse = stats.InUse
	h.Idle = stats.Idle
	h.MaxOpenConns = stats.MaxOpenConnections
	h.MaxIdleConns = int(stats.MaxIdleClosed)

	database.DB.QueryRow("SELECT VERSION()").Scan(&h.Version)
	database.DB.QueryRow("SHOW STATUS LIKE 'Uptime'").Scan(new(string), &h.Uptime)
	database.DB.QueryRow("SHOW STATUS LIKE 'Threads_connected'").Scan(new(string), &h.Threads)
	database.DB.QueryRow("SHOW STATUS LIKE 'Questions'").Scan(new(string), &h.Questions)
	database.DB.QueryRow("SHOW STATUS LIKE 'Slow_queries'").Scan(new(string), &h.SlowQueries)
	database.DB.QueryRow("SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE()").Scan(&h.TableCount)
	database.DB.QueryRow("SELECT ROUND(SUM(DATA_LENGTH + INDEX_LENGTH) / 1024 / 1024, 2) FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE()").Scan(&h.DBSize)

	return h
}

type RedisHealth struct {
	Status           string `json:"status"`
	PingLatency      int64  `json:"ping_latency_ms"`
	Version          string `json:"version"`
	UsedMemory       string `json:"used_memory_human"`
	UsedMemoryBytes  int64  `json:"used_memory_bytes"`
	ConnectedClients int    `json:"connected_clients"`
	TotalKeys        int64  `json:"total_keys"`
	UptimeSeconds    int64  `json:"uptime_seconds"`
	HitRate          string `json:"hit_rate"`
}

func (s *OpsService) GetRedisHealth() RedisHealth {
	h := RedisHealth{Status: "unknown"}
	if cache.RDB == nil {
		h.Status = "disconnected"
		return h
	}

	ctx := context.Background()
	start := time.Now()
	err := cache.RDB.Ping(ctx).Err()
	h.PingLatency = time.Since(start).Milliseconds()
	if err != nil {
		h.Status = "error"
		return h
	}
	h.Status = "healthy"

	info, err := cache.RDB.Info(ctx, "server", "memory", "clients", "keyspace", "stats").Result()
	if err == nil {
		h.Version = parseRedisInfo(info, "redis_version")
		h.UsedMemory = parseRedisInfo(info, "used_memory_human")
		fmt.Sscanf(parseRedisInfo(info, "used_memory"), "%d", &h.UsedMemoryBytes)
		fmt.Sscanf(parseRedisInfo(info, "connected_clients"), "%d", &h.ConnectedClients)
		fmt.Sscanf(parseRedisInfo(info, "uptime_in_seconds"), "%d", &h.UptimeSeconds)

		var hits, misses int64
		fmt.Sscanf(parseRedisInfo(info, "keyspace_hits"), "%d", &hits)
		fmt.Sscanf(parseRedisInfo(info, "keyspace_misses"), "%d", &misses)
		if hits+misses > 0 {
			rate := float64(hits) / float64(hits+misses) * 100
			h.HitRate = fmt.Sprintf("%.1f%%", rate)
		} else {
			h.HitRate = "N/A"
		}
	}

	dbSize, err := cache.RDB.DBSize(ctx).Result()
	if err == nil {
		h.TotalKeys = dbSize
	}
	return h
}

func parseRedisInfo(info, key string) string {
	lines := splitLines(info)
	prefix := key + ":"
	for _, line := range lines {
		if len(line) > len(prefix) && line[:len(prefix)] == prefix {
			return line[len(prefix):]
		}
	}
	return ""
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

type WSStatus struct {
	OnlineCount int `json:"online_count"`
}

func (s *OpsService) GetWSStatus() WSStatus {
	st := WSStatus{}
	if ws.GlobalHub != nil {
		st.OnlineCount = ws.GlobalHub.OnlineCount()
	}
	return st
}

func (s *OpsService) GetQueueStats() map[string]interface{} {
	if queue.GlobalDockQueue != nil {
		return queue.GlobalDockQueue.Stats()
	}
	return map[string]interface{}{"running": false}
}
