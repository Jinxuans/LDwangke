package admin

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"go-api/internal/cache"
	"go-api/internal/database"
	"go-api/internal/dockscheduler"
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

type WSStatus struct {
	OnlineCount int `json:"online_count"`
}

type ErrorStats struct {
	TodayFailed    int   `json:"today_failed"`
	TodayException int   `json:"today_exception"`
	PendingDock    int   `json:"pending_dock"`
	StuckOrders    int   `json:"stuck_orders"`
	ErrorCounter   int64 `json:"error_counter"`
	DockFailCount  int64 `json:"dock_fail_count"`
	HTTPErrorCount int64 `json:"http_error_count"`
}

type TableSize struct {
	Name    string `json:"name"`
	Rows    int64  `json:"rows"`
	DataMB  string `json:"data_mb"`
	IndexMB string `json:"index_mb"`
	TotalMB string `json:"total_mb"`
}

type StorageInfo struct {
	UploadsSize  string `json:"uploads_size"`
	UploadsFiles int    `json:"uploads_files"`
}

type RecentErrorOrder struct {
	OID     int    `json:"oid"`
	User    string `json:"user"`
	PtName  string `json:"ptname"`
	Status  string `json:"status"`
	AddTime string `json:"addtime"`
}

type HourlyOrder struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

type OpsDashboard struct {
	System        SystemInfo             `json:"system"`
	DB            DBHealth               `json:"db"`
	Redis         RedisHealth            `json:"redis"`
	WS            WSStatus               `json:"ws"`
	DockScheduler map[string]interface{} `json:"dock_scheduler"`
	Errors        ErrorStats             `json:"errors"`
	Storage       StorageInfo            `json:"storage"`
	Tables        []TableSize            `json:"tables"`
	ErrorOrders   []RecentErrorOrder     `json:"error_orders"`
	HourlyOrders  []HourlyOrder          `json:"hourly_orders"`
}

type SupplierProbe struct {
	HID      int    `json:"hid"`
	Name     string `json:"name"`
	PT       string `json:"pt"`
	URL      string `json:"url"`
	Status   string `json:"status"`
	Latency  int64  `json:"latency_ms"`
	HTTPCode int    `json:"http_code"`
}

var adminOpsStartTime = time.Now()

func getAdminOpsDashboard() OpsDashboard {
	return OpsDashboard{
		System:        getAdminSystemInfo(),
		DB:            getAdminDBHealth(),
		Redis:         getAdminRedisHealth(),
		WS:            getAdminWSStatus(),
		DockScheduler: getAdminDockSchedulerStats(),
		Errors:        getAdminErrorStats(),
		Storage:       getAdminStorageInfo(),
		Tables:        getAdminTableSizes(),
		ErrorOrders:   getAdminRecentErrorOrders(20),
		HourlyOrders:  getAdminTodayHourlyOrders(),
	}
}

func getAdminSystemInfo() SystemInfo {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	uptime := time.Since(adminOpsStartTime)
	return SystemInfo{
		GoVersion:     runtime.Version(),
		NumCPU:        runtime.NumCPU(),
		NumGoroutine:  runtime.NumGoroutine(),
		MemAlloc:      mem.Alloc,
		MemTotalAlloc: mem.TotalAlloc,
		MemSys:        mem.Sys,
		NumGC:         mem.NumGC,
		LastGCPause:   mem.PauseNs[(mem.NumGC+255)%256],
		HeapObjects:   mem.HeapObjects,
		HeapInuse:     mem.HeapInuse,
		StackInuse:    mem.StackInuse,
		Uptime:        int64(uptime.Seconds()),
		UptimeHuman:   formatAdminOpsDuration(uptime),
		ServerTime:    time.Now().Format("2006-01-02 15:04:05"),
		GOOS:          runtime.GOOS,
		GOARCH:        runtime.GOARCH,
	}
}

func formatAdminOpsDuration(duration time.Duration) string {
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%d天%d小时%d分钟", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%d小时%d分钟", hours, minutes)
	}
	return fmt.Sprintf("%d分钟", minutes)
}

func getAdminDBHealth() DBHealth {
	health := DBHealth{Status: "unknown"}
	if database.DB == nil {
		health.Status = "disconnected"
		return health
	}

	start := time.Now()
	if err := database.DB.Ping(); err != nil {
		health.PingLatency = time.Since(start).Milliseconds()
		health.Status = "error"
		return health
	}
	health.PingLatency = time.Since(start).Milliseconds()
	health.Status = "healthy"

	stats := database.DB.Stats()
	health.OpenConns = stats.OpenConnections
	health.InUse = stats.InUse
	health.Idle = stats.Idle
	health.MaxOpenConns = stats.MaxOpenConnections
	health.MaxIdleConns = int(stats.MaxIdleClosed)

	database.DB.QueryRow("SELECT VERSION()").Scan(&health.Version)
	database.DB.QueryRow("SHOW STATUS LIKE 'Uptime'").Scan(new(string), &health.Uptime)
	database.DB.QueryRow("SHOW STATUS LIKE 'Threads_connected'").Scan(new(string), &health.Threads)
	database.DB.QueryRow("SHOW STATUS LIKE 'Questions'").Scan(new(string), &health.Questions)
	database.DB.QueryRow("SHOW STATUS LIKE 'Slow_queries'").Scan(new(string), &health.SlowQueries)
	database.DB.QueryRow("SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE()").Scan(&health.TableCount)
	database.DB.QueryRow("SELECT ROUND(SUM(DATA_LENGTH + INDEX_LENGTH) / 1024 / 1024, 2) FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE()").Scan(&health.DBSize)

	return health
}

func getAdminRedisHealth() RedisHealth {
	health := RedisHealth{Status: "unknown"}
	if cache.RDB == nil {
		health.Status = "disconnected"
		return health
	}

	ctx := context.Background()
	start := time.Now()
	if err := cache.RDB.Ping(ctx).Err(); err != nil {
		health.PingLatency = time.Since(start).Milliseconds()
		health.Status = "error"
		return health
	}
	health.PingLatency = time.Since(start).Milliseconds()
	health.Status = "healthy"

	info, err := cache.RDB.Info(ctx, "server", "memory", "clients", "keyspace", "stats").Result()
	if err == nil {
		health.Version = parseAdminRedisInfo(info, "redis_version")
		health.UsedMemory = parseAdminRedisInfo(info, "used_memory_human")
		fmt.Sscanf(parseAdminRedisInfo(info, "used_memory"), "%d", &health.UsedMemoryBytes)
		fmt.Sscanf(parseAdminRedisInfo(info, "connected_clients"), "%d", &health.ConnectedClients)
		fmt.Sscanf(parseAdminRedisInfo(info, "uptime_in_seconds"), "%d", &health.UptimeSeconds)

		var hits, misses int64
		fmt.Sscanf(parseAdminRedisInfo(info, "keyspace_hits"), "%d", &hits)
		fmt.Sscanf(parseAdminRedisInfo(info, "keyspace_misses"), "%d", &misses)
		if hits+misses > 0 {
			health.HitRate = fmt.Sprintf("%.1f%%", float64(hits)/float64(hits+misses)*100)
		} else {
			health.HitRate = "N/A"
		}
	}

	if dbSize, err := cache.RDB.DBSize(ctx).Result(); err == nil {
		health.TotalKeys = dbSize
	}
	return health
}

func parseAdminRedisInfo(info, key string) string {
	prefix := key + ":"
	for _, line := range splitAdminRedisLines(info) {
		if len(line) > len(prefix) && line[:len(prefix)] == prefix {
			return line[len(prefix):]
		}
	}
	return ""
}

func splitAdminRedisLines(info string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(info); i++ {
		if info[i] != '\n' {
			continue
		}
		line := info[start:i]
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		lines = append(lines, line)
		start = i + 1
	}
	if start < len(info) {
		lines = append(lines, info[start:])
	}
	return lines
}

func getAdminWSStatus() WSStatus {
	status := WSStatus{}
	if ws.GlobalHub != nil {
		status.OnlineCount = ws.GlobalHub.OnlineCount()
	}
	return status
}

func getAdminDockSchedulerStats() map[string]interface{} {
	stats := dockscheduler.Snapshot()
	return map[string]interface{}{
		"running":       stats.Running,
		"active":        stats.Active,
		"pending":       stats.Pending,
		"interval_sec":  stats.IntervalSec,
		"batch_limit":   stats.BatchLimit,
		"last_fetched":  stats.LastFetched,
		"last_success":  stats.LastSuccess,
		"last_fail":     stats.LastFail,
		"total_success": stats.TotalSuccess,
		"total_fail":    stats.TotalFail,
		"total_runs":    stats.TotalRuns,
		"last_run_time": stats.LastRunTime,
		"last_trigger":  stats.LastTrigger,
		"last_error":    stats.LastError,
	}
}

func getAdminErrorStats() ErrorStats {
	stats := ErrorStats{}
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE status = '失败' AND DATE(addtime) = CURDATE()").Scan(&stats.TodayFailed)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE status = '异常' AND DATE(addtime) = CURDATE()").Scan(&stats.TodayException)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE dockstatus IN (0, 2)").Scan(&stats.PendingDock)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE status = '进行中' AND addtime < NOW() - INTERVAL 24 HOUR").Scan(&stats.StuckOrders)
	return stats
}

func getAdminTableSizes() []TableSize {
	rows, err := database.DB.Query(`
		SELECT TABLE_NAME, TABLE_ROWS,
			ROUND(DATA_LENGTH/1024/1024, 2),
			ROUND(INDEX_LENGTH/1024/1024, 2),
			ROUND((DATA_LENGTH + INDEX_LENGTH)/1024/1024, 2)
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = DATABASE()
		ORDER BY (DATA_LENGTH + INDEX_LENGTH) DESC
		LIMIT 15
	`)
	if err != nil {
		return []TableSize{}
	}
	defer rows.Close()

	var result []TableSize
	for rows.Next() {
		var table TableSize
		rows.Scan(&table.Name, &table.Rows, &table.DataMB, &table.IndexMB, &table.TotalMB)
		result = append(result, table)
	}
	if result == nil {
		result = []TableSize{}
	}
	return result
}

func getAdminStorageInfo() StorageInfo {
	info := StorageInfo{}
	var totalSize int64
	var fileCount int

	filepath.Walk("./uploads", func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !fileInfo.IsDir() {
			totalSize += fileInfo.Size()
			fileCount++
		}
		return nil
	})

	info.UploadsFiles = fileCount
	switch {
	case totalSize < 1024*1024:
		info.UploadsSize = fmt.Sprintf("%.1f KB", float64(totalSize)/1024)
	case totalSize < 1024*1024*1024:
		info.UploadsSize = fmt.Sprintf("%.1f MB", float64(totalSize)/1024/1024)
	default:
		info.UploadsSize = fmt.Sprintf("%.2f GB", float64(totalSize)/1024/1024/1024)
	}
	return info
}

func getAdminRecentErrorOrders(limit int) []RecentErrorOrder {
	if limit <= 0 {
		limit = 20
	}
	rows, err := database.DB.Query(
		"SELECT oid, COALESCE(user,''), COALESCE(ptname,''), COALESCE(status,''), COALESCE(DATE_FORMAT(addtime,'%Y-%m-%d %H:%i:%s'),'') FROM qingka_wangke_order WHERE status IN ('异常','失败') ORDER BY oid DESC LIMIT ?",
		limit,
	)
	if err != nil {
		return []RecentErrorOrder{}
	}
	defer rows.Close()

	var orders []RecentErrorOrder
	for rows.Next() {
		var order RecentErrorOrder
		rows.Scan(&order.OID, &order.User, &order.PtName, &order.Status, &order.AddTime)
		orders = append(orders, order)
	}
	if orders == nil {
		orders = []RecentErrorOrder{}
	}
	return orders
}

func getAdminTodayHourlyOrders() []HourlyOrder {
	rows, err := database.DB.Query(
		"SELECT HOUR(addtime) AS h, COUNT(*) AS cnt FROM qingka_wangke_order WHERE DATE(addtime) = CURDATE() GROUP BY HOUR(addtime) ORDER BY h ASC",
	)
	if err != nil {
		return []HourlyOrder{}
	}
	defer rows.Close()

	hourMap := make(map[int]int)
	for rows.Next() {
		var hour, count int
		rows.Scan(&hour, &count)
		hourMap[hour] = count
	}

	currentHour := time.Now().Hour()
	result := make([]HourlyOrder, 0, currentHour+1)
	for hour := 0; hour <= currentHour; hour++ {
		result = append(result, HourlyOrder{Hour: hour, Count: hourMap[hour]})
	}
	return result
}

func probeAdminSuppliers() []SupplierProbe {
	rows, err := database.DB.Query("SELECT hid, COALESCE(name,''), COALESCE(pt,''), COALESCE(url,'') FROM qingka_wangke_huoyuan WHERE status = 1")
	if err != nil {
		return []SupplierProbe{}
	}
	defer rows.Close()

	var suppliers []SupplierProbe
	for rows.Next() {
		var probe SupplierProbe
		rows.Scan(&probe.HID, &probe.Name, &probe.PT, &probe.URL)
		suppliers = append(suppliers, probe)
	}
	if len(suppliers) == 0 {
		return []SupplierProbe{}
	}

	sem := make(chan struct{}, 10)
	var waitGroup sync.WaitGroup
	client := &http.Client{Timeout: 5 * time.Second}

	for i := range suppliers {
		waitGroup.Add(1)
		sem <- struct{}{}
		go func(idx int) {
			defer func() {
				<-sem
				waitGroup.Done()
			}()

			probe := &suppliers[idx]
			if probe.URL == "" {
				probe.Status = "no_url"
				return
			}

			probeURL := probe.URL
			if len(probeURL) > 4 && probeURL[:4] != "http" {
				probeURL = "http://" + probeURL
			}

			start := time.Now()
			resp, err := client.Get(probeURL)
			probe.Latency = time.Since(start).Milliseconds()
			if err != nil {
				probe.Status = "unreachable"
				return
			}
			resp.Body.Close()

			probe.HTTPCode = resp.StatusCode
			if resp.StatusCode >= 200 && resp.StatusCode < 400 {
				probe.Status = "healthy"
				return
			}
			probe.Status = "degraded"
		}(i)
	}

	waitGroup.Wait()
	return suppliers
}
