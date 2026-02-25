package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"go-api/internal/cache"
	"go-api/internal/database"
	"go-api/internal/queue"
	"go-api/internal/ws"
)

// ===== 运维看板服务 =====

type OpsService struct{}

func NewOpsService() *OpsService {
	return &OpsService{}
}

// ---------- 启动时间记录 ----------

var startTime = time.Now()

// ---------- 错误计数器 ----------

var (
	opsErrCount   int64
	opsDockFail   int64
	opsHTTPErrors int64
)

func OpsIncrError()     { atomic.AddInt64(&opsErrCount, 1) }
func OpsIncrDockFail()  { atomic.AddInt64(&opsDockFail, 1) }
func OpsIncrHTTPError() { atomic.AddInt64(&opsHTTPErrors, 1) }

// ---------- 系统信息 ----------

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

// ---------- 数据库健康 ----------

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

	// Ping 延迟
	start := time.Now()
	err := database.DB.Ping()
	h.PingLatency = time.Since(start).Milliseconds()
	if err != nil {
		h.Status = "error"
		return h
	}
	h.Status = "healthy"

	// 连接池状态
	stats := database.DB.Stats()
	h.OpenConns = stats.OpenConnections
	h.InUse = stats.InUse
	h.Idle = stats.Idle
	h.MaxOpenConns = stats.MaxOpenConnections
	// MaxIdleConns 不在 sql.DBStats 中直接暴露，设为配置值
	h.MaxIdleConns = int(stats.MaxIdleClosed) // 近似指标

	// MySQL 版本
	database.DB.QueryRow("SELECT VERSION()").Scan(&h.Version)

	// MySQL uptime
	database.DB.QueryRow("SHOW STATUS LIKE 'Uptime'").Scan(new(string), &h.Uptime)

	// Threads
	database.DB.QueryRow("SHOW STATUS LIKE 'Threads_connected'").Scan(new(string), &h.Threads)

	// Questions
	database.DB.QueryRow("SHOW STATUS LIKE 'Questions'").Scan(new(string), &h.Questions)

	// Slow queries
	database.DB.QueryRow("SHOW STATUS LIKE 'Slow_queries'").Scan(new(string), &h.SlowQueries)

	// Table count
	database.DB.QueryRow("SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE()").Scan(&h.TableCount)

	// DB size
	database.DB.QueryRow("SELECT ROUND(SUM(DATA_LENGTH + INDEX_LENGTH) / 1024 / 1024, 2) FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE()").Scan(&h.DBSize)

	return h
}

// ---------- Redis 健康 ----------

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

	// Ping
	start := time.Now()
	err := cache.RDB.Ping(ctx).Err()
	h.PingLatency = time.Since(start).Milliseconds()
	if err != nil {
		h.Status = "error"
		return h
	}
	h.Status = "healthy"

	// INFO
	info, err := cache.RDB.Info(ctx, "server", "memory", "clients", "keyspace", "stats").Result()
	if err == nil {
		h.Version = parseRedisInfo(info, "redis_version")
		h.UsedMemory = parseRedisInfo(info, "used_memory_human")
		fmt.Sscanf(parseRedisInfo(info, "used_memory"), "%d", &h.UsedMemoryBytes)
		fmt.Sscanf(parseRedisInfo(info, "connected_clients"), "%d", &h.ConnectedClients)
		fmt.Sscanf(parseRedisInfo(info, "uptime_in_seconds"), "%d", &h.UptimeSeconds)

		// 命中率
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

	// Key count
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

// ---------- WebSocket 状态 ----------

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

// ---------- 对接队列状态 ----------

func (s *OpsService) GetQueueStats() map[string]interface{} {
	if queue.GlobalDockQueue != nil {
		return queue.GlobalDockQueue.Stats()
	}
	return map[string]interface{}{"running": false}
}

// ---------- 供应商健康探测 ----------

type SupplierProbe struct {
	HID      int    `json:"hid"`
	Name     string `json:"name"`
	PT       string `json:"pt"`
	URL      string `json:"url"`
	Status   string `json:"status"`
	Latency  int64  `json:"latency_ms"`
	HTTPCode int    `json:"http_code"`
}

func (s *OpsService) ProbeSuppliers() []SupplierProbe {
	rows, err := database.DB.Query("SELECT hid, COALESCE(name,''), COALESCE(pt,''), COALESCE(url,'') FROM qingka_wangke_huoyuan WHERE status = 1")
	if err != nil {
		return []SupplierProbe{}
	}
	defer rows.Close()

	var suppliers []SupplierProbe
	for rows.Next() {
		var sp SupplierProbe
		rows.Scan(&sp.HID, &sp.Name, &sp.PT, &sp.URL)
		suppliers = append(suppliers, sp)
	}

	if len(suppliers) == 0 {
		return []SupplierProbe{}
	}

	// 并发探测（限制并发数10）
	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup

	client := &http.Client{Timeout: 5 * time.Second}

	for i := range suppliers {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int) {
			defer func() {
				<-sem
				wg.Done()
			}()
			sp := &suppliers[idx]
			if sp.URL == "" {
				sp.Status = "no_url"
				return
			}
			probeURL := sp.URL
			// 确保有 scheme
			if len(probeURL) > 4 && probeURL[:4] != "http" {
				probeURL = "http://" + probeURL
			}

			start := time.Now()
			resp, err := client.Get(probeURL)
			sp.Latency = time.Since(start).Milliseconds()
			if err != nil {
				sp.Status = "unreachable"
				return
			}
			resp.Body.Close()
			sp.HTTPCode = resp.StatusCode
			if resp.StatusCode >= 200 && resp.StatusCode < 400 {
				sp.Status = "healthy"
			} else {
				sp.Status = "degraded"
			}
		}(i)
	}
	wg.Wait()

	return suppliers
}

// ---------- 异常订单监控 ----------

type ErrorStats struct {
	TodayFailed    int   `json:"today_failed"`
	TodayException int   `json:"today_exception"`
	PendingDock    int   `json:"pending_dock"`
	StuckOrders    int   `json:"stuck_orders"`
	ErrorCounter   int64 `json:"error_counter"`
	DockFailCount  int64 `json:"dock_fail_count"`
	HTTPErrorCount int64 `json:"http_error_count"`
}

func (s *OpsService) GetErrorStats() ErrorStats {
	es := ErrorStats{
		ErrorCounter:   atomic.LoadInt64(&opsErrCount),
		DockFailCount:  atomic.LoadInt64(&opsDockFail),
		HTTPErrorCount: atomic.LoadInt64(&opsHTTPErrors),
	}

	// 今日失败订单
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE status = '失败' AND DATE(addtime) = CURDATE()").Scan(&es.TodayFailed)

	// 今日异常订单
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE status = '异常' AND DATE(addtime) = CURDATE()").Scan(&es.TodayException)

	// 待对接
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE dockstatus = 0").Scan(&es.PendingDock)

	// 卡单（进行中超过24小时未更新）
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE status = '进行中' AND addtime < NOW() - INTERVAL 24 HOUR").Scan(&es.StuckOrders)

	return es
}

// ---------- 表容量 Top ----------

type TableSize struct {
	Name    string `json:"name"`
	Rows    int64  `json:"rows"`
	DataMB  string `json:"data_mb"`
	IndexMB string `json:"index_mb"`
	TotalMB string `json:"total_mb"`
}

func (s *OpsService) GetTableSizes() []TableSize {
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
		var t TableSize
		rows.Scan(&t.Name, &t.Rows, &t.DataMB, &t.IndexMB, &t.TotalMB)
		result = append(result, t)
	}
	if result == nil {
		result = []TableSize{}
	}
	return result
}

// ---------- 上传目录大小 ----------

type StorageInfo struct {
	UploadsSize  string `json:"uploads_size"`
	UploadsFiles int    `json:"uploads_files"`
}

func (s *OpsService) GetStorageInfo() StorageInfo {
	si := StorageInfo{}
	var totalSize int64
	var fileCount int

	filepath.Walk("./uploads", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			totalSize += info.Size()
			fileCount++
		}
		return nil
	})

	si.UploadsFiles = fileCount
	if totalSize < 1024*1024 {
		si.UploadsSize = fmt.Sprintf("%.1f KB", float64(totalSize)/1024)
	} else if totalSize < 1024*1024*1024 {
		si.UploadsSize = fmt.Sprintf("%.1f MB", float64(totalSize)/1024/1024)
	} else {
		si.UploadsSize = fmt.Sprintf("%.2f GB", float64(totalSize)/1024/1024/1024)
	}
	return si
}

// ---------- 近期异常订单 ----------

type RecentErrorOrder struct {
	OID     int    `json:"oid"`
	User    string `json:"user"`
	PtName  string `json:"ptname"`
	Status  string `json:"status"`
	AddTime string `json:"addtime"`
}

func (s *OpsService) GetRecentErrorOrders(limit int) []RecentErrorOrder {
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

	var list []RecentErrorOrder
	for rows.Next() {
		var o RecentErrorOrder
		rows.Scan(&o.OID, &o.User, &o.PtName, &o.Status, &o.AddTime)
		list = append(list, o)
	}
	if list == nil {
		list = []RecentErrorOrder{}
	}
	return list
}

// ---------- 每小时订单量（今日） ----------

type HourlyOrder struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

func (s *OpsService) GetTodayHourlyOrders() []HourlyOrder {
	rows, err := database.DB.Query(
		"SELECT HOUR(addtime) as h, COUNT(*) as cnt FROM qingka_wangke_order WHERE DATE(addtime) = CURDATE() GROUP BY HOUR(addtime) ORDER BY h ASC",
	)
	if err != nil {
		return []HourlyOrder{}
	}
	defer rows.Close()

	hourMap := make(map[int]int)
	for rows.Next() {
		var h, cnt int
		rows.Scan(&h, &cnt)
		hourMap[h] = cnt
	}

	// 填充所有小时
	now := time.Now().Hour()
	result := make([]HourlyOrder, 0, now+1)
	for i := 0; i <= now; i++ {
		result = append(result, HourlyOrder{Hour: i, Count: hourMap[i]})
	}
	return result
}

// ---------- 综合运维看板 ----------

type OpsDashboard struct {
	System       SystemInfo             `json:"system"`
	DB           DBHealth               `json:"db"`
	Redis        RedisHealth            `json:"redis"`
	WS           WSStatus               `json:"ws"`
	Queue        map[string]interface{} `json:"queue"`
	Errors       ErrorStats             `json:"errors"`
	Storage      StorageInfo            `json:"storage"`
	Tables       []TableSize            `json:"tables"`
	ErrorOrders  []RecentErrorOrder     `json:"error_orders"`
	HourlyOrders []HourlyOrder          `json:"hourly_orders"`
}

func (s *OpsService) GetDashboard() OpsDashboard {
	return OpsDashboard{
		System:       s.GetSystemInfo(),
		DB:           s.GetDBHealth(),
		Redis:        s.GetRedisHealth(),
		WS:           s.GetWSStatus(),
		Queue:        s.GetQueueStats(),
		Errors:       s.GetErrorStats(),
		Storage:      s.GetStorageInfo(),
		Tables:       s.GetTableSizes(),
		ErrorOrders:  s.GetRecentErrorOrders(20),
		HourlyOrders: s.GetTodayHourlyOrders(),
	}
}

// ---------- 供应商探测（单独接口，耗时较长） ----------

func (s *OpsService) GetDashboardWithProbes() (OpsDashboard, []SupplierProbe) {
	dash := s.GetDashboard()
	probes := s.ProbeSuppliers()
	return dash, probes
}

// suppress unused import
var _ = sql.ErrNoRows
var _ = log.Println
