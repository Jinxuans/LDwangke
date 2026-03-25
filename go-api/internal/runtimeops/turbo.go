package runtimeops

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go-api/internal/cache"
	"go-api/internal/database"
	"go-api/internal/dockscheduler"
	ordermodule "go-api/internal/modules/order"
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

type OrderProgressSyncConfig struct {
	Enabled          bool                       `json:"enabled"`
	IntervalSec      int                        `json:"interval_sec"`
	BatchEnabled     bool                       `json:"batch_enabled"`
	BatchIntervalSec int                        `json:"batch_interval_sec"`
	SupplierIDs      []int                      `json:"supplier_ids"`
	ExcludedStatuses []string                   `json:"excluded_statuses"`
	Rules            []ordermodule.AutoSyncRule `json:"rules"`
}

type OrderProgressSyncStatus struct {
	Enabled          bool                       `json:"enabled"`
	Running          bool                       `json:"running"`
	IntervalSec      int                        `json:"interval_sec"`
	BatchEnabled     bool                       `json:"batch_enabled"`
	BatchRunning     bool                       `json:"batch_running"`
	BatchIntervalSec int                        `json:"batch_interval_sec"`
	SupplierIDs      []int                      `json:"supplier_ids"`
	ExcludedStatuses []string                   `json:"excluded_statuses"`
	Rules            []ordermodule.AutoSyncRule `json:"rules"`
	LastRunTime      string                     `json:"last_run_time"`
	NextRunTime      string                     `json:"next_run_time"`
	LastUpdated      int                        `json:"last_updated"`
	LastFailed       int                        `json:"last_failed"`
	TotalRuns        int64                      `json:"total_runs"`
	LastError        string                     `json:"last_error"`
	BatchLastRunTime string                     `json:"batch_last_run_time"`
	BatchNextRunTime string                     `json:"batch_next_run_time"`
	BatchLastUpdated int                        `json:"batch_last_updated"`
	BatchLastFailed  int                        `json:"batch_last_failed"`
	BatchTotalRuns   int64                      `json:"batch_total_runs"`
	BatchLastError   string                     `json:"batch_last_error"`
}

type OrderProgressSyncLogEntry struct {
	ID               int64          `json:"id"`
	Time             string         `json:"time"`
	Mode             string         `json:"mode"`
	Trigger          string         `json:"trigger"`
	IntervalSec      int            `json:"interval_sec"`
	SupplierIDs      []int          `json:"supplier_ids"`
	SupplierNames    []string       `json:"supplier_names"`
	ExcludedStatuses []string       `json:"excluded_statuses"`
	RuleHits         map[string]int `json:"rule_hits"`
	SampleErrors     []string       `json:"sample_errors"`
	Updated          int            `json:"updated"`
	Failed           int            `json:"failed"`
	DurationMs       int64          `json:"duration_ms"`
	Error            string         `json:"error"`
	Lines            []string       `json:"lines"`
}

var (
	turboMu       sync.RWMutex
	turboEnabled  int32
	turboStatus   TurboStatus
	turboBaseline TurboProfile

	syncTickerMu        sync.Mutex
	syncTickerStop      chan struct{}
	batchSyncTickerStop chan struct{}

	orderProgressMu     sync.RWMutex
	orderProgressConfig = OrderProgressSyncConfig{
		Enabled:          true,
		IntervalSec:      120,
		BatchEnabled:     true,
		BatchIntervalSec: 120,
		ExcludedStatuses: []string{"已完成", "已退款", "已取消", "失败"},
		Rules: []ordermodule.AutoSyncRule{
			{Key: "0_24h", Label: "0-24H", MinAgeHours: 0, MaxAgeHours: 24, IntervalMinutes: 10, Enabled: true},
			{Key: "1_3d", Label: "1-3d", MinAgeHours: 24, MaxAgeHours: 72, IntervalMinutes: 30, Enabled: true},
			{Key: "3_7d", Label: "3-7d", MinAgeHours: 72, MaxAgeHours: 168, IntervalMinutes: 120, Enabled: true},
			{Key: "7_15d", Label: "7-15d", MinAgeHours: 168, MaxAgeHours: 360, IntervalMinutes: 360, Enabled: true},
			{Key: "15_30d", Label: "15-30d", MinAgeHours: 360, MaxAgeHours: 720, IntervalMinutes: 720, Enabled: true},
			{Key: "30d_plus", Label: "30d+", MinAgeHours: 720, MaxAgeHours: 0, IntervalMinutes: 1440, Enabled: true},
		},
	}
	orderProgressStatus = OrderProgressSyncStatus{
		Enabled:          true,
		IntervalSec:      120,
		BatchEnabled:     true,
		BatchIntervalSec: 120,
		ExcludedStatuses: []string{"已完成", "已退款", "已取消", "失败"},
		Rules: []ordermodule.AutoSyncRule{
			{Key: "0_24h", Label: "0-24H", MinAgeHours: 0, MaxAgeHours: 24, IntervalMinutes: 10, Enabled: true},
			{Key: "1_3d", Label: "1-3d", MinAgeHours: 24, MaxAgeHours: 72, IntervalMinutes: 30, Enabled: true},
			{Key: "3_7d", Label: "3-7d", MinAgeHours: 72, MaxAgeHours: 168, IntervalMinutes: 120, Enabled: true},
			{Key: "7_15d", Label: "7-15d", MinAgeHours: 168, MaxAgeHours: 360, IntervalMinutes: 360, Enabled: true},
			{Key: "15_30d", Label: "15-30d", MinAgeHours: 360, MaxAgeHours: 720, IntervalMinutes: 720, Enabled: true},
			{Key: "30d_plus", Label: "30d+", MinAgeHours: 720, MaxAgeHours: 0, IntervalMinutes: 1440, Enabled: true},
		},
	}
	orderProgressLogID int64
	orderProgressLogs  []OrderProgressSyncLogEntry
)

func init() {
	turboBaseline = detectBaseline()
	turboStatus = TurboStatus{
		Enabled:  false,
		Profile:  turboBaseline,
		Baseline: turboBaseline,
	}
}

func getMemTotalMB() int {
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile("/proc/meminfo")
		if err == nil {
			var kb int
			fmt.Sscanf(string(data), "MemTotal: %d kB", &kb)
			if kb > 0 {
				return kb / 1024
			}
		}
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	sysMB := int(m.Sys / 1024 / 1024)
	estimated := sysMB * 4
	if estimated < 2048 {
		estimated = 2048
	}
	return estimated
}

func detectBaseline() TurboProfile {
	return TurboProfile{
		Name:                   "normal",
		CPUCores:               runtime.NumCPU(),
		MemTotalMB:             getMemTotalMB(),
		GOOS:                   runtime.GOOS,
		GOARCH:                 runtime.GOARCH,
		DBMaxOpen:              50,
		DBMaxIdle:              25,
		DBMaxLifetime:          300,
		DBMaxIdleTime:          180,
		RedisPoolSize:          10,
		RedisMinIdle:           3,
		DockBatchLimit:         100,
		PendingDockIntervalSec: 30,
		SyncIntervalSec:        120,
		GOMAXPROCS:             runtime.NumCPU(),
		GCPercent:              100,
	}
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func calcProfile(mode string) TurboProfile {
	cpus := runtime.NumCPU()
	memMB := getMemTotalMB()

	p := TurboProfile{
		Name:       mode,
		CPUCores:   cpus,
		MemTotalMB: memMB,
		GOOS:       runtime.GOOS,
		GOARCH:     runtime.GOARCH,
	}

	switch mode {
	case "eco":
		p.DBMaxOpen = clampInt(cpus*5, 10, 30)
		p.DBMaxIdle = clampInt(cpus*2, 5, 15)
		p.DBMaxLifetime = 600
		p.DBMaxIdleTime = 300
		p.RedisPoolSize = clampInt(cpus*2, 5, 15)
		p.RedisMinIdle = 2
		p.DockBatchLimit = 30
		p.PendingDockIntervalSec = 60
		p.SyncIntervalSec = 300
		p.GOMAXPROCS = max(1, cpus-1)
		p.GCPercent = 200
	case "normal":
		p.DBMaxOpen = clampInt(cpus*10, 25, 100)
		p.DBMaxIdle = clampInt(cpus*5, 10, 50)
		p.DBMaxLifetime = 300
		p.DBMaxIdleTime = 180
		p.RedisPoolSize = clampInt(cpus*3, 10, 30)
		p.RedisMinIdle = 3
		p.DockBatchLimit = 100
		p.PendingDockIntervalSec = 30
		p.SyncIntervalSec = 120
		p.GOMAXPROCS = cpus
		p.GCPercent = 100
	case "turbo":
		p.DBMaxOpen = clampInt(cpus*20, 50, 200)
		p.DBMaxIdle = clampInt(cpus*10, 25, 100)
		p.DBMaxLifetime = 180
		p.DBMaxIdleTime = 120
		p.RedisPoolSize = clampInt(cpus*5, 20, 60)
		p.RedisMinIdle = clampInt(cpus*2, 5, 20)
		p.DockBatchLimit = 200
		p.PendingDockIntervalSec = 15
		p.SyncIntervalSec = 60
		p.GOMAXPROCS = cpus
		p.GCPercent = 50
	case "insane":
		p.DBMaxOpen = clampInt(cpus*30, 100, 500)
		p.DBMaxIdle = clampInt(cpus*15, 50, 250)
		p.DBMaxLifetime = 120
		p.DBMaxIdleTime = 60
		p.RedisPoolSize = clampInt(cpus*8, 30, 100)
		p.RedisMinIdle = clampInt(cpus*3, 10, 30)
		p.DockBatchLimit = 400
		p.PendingDockIntervalSec = 10
		p.SyncIntervalSec = 30
		p.GOMAXPROCS = cpus
		p.GCPercent = 30
	default:
		return calcProfile("normal")
	}

	if memMB < 1024 && (mode == "insane" || mode == "turbo") {
		log.Printf("[Turbo] 内存仅 %dMB，自动降级参数", memMB)
		p.DBMaxOpen = min(p.DBMaxOpen, 50)
		p.DBMaxIdle = min(p.DBMaxIdle, 25)
		p.RedisPoolSize = min(p.RedisPoolSize, 15)
		p.DockBatchLimit = min(p.DockBatchLimit, 100)
		p.GCPercent = max(p.GCPercent, 80)
	}

	return p
}

func autoDetectMode() string {
	cpus := runtime.NumCPU()
	memMB := getMemTotalMB()

	switch {
	case cpus >= 8 && memMB >= 8192:
		return "insane"
	case cpus >= 4 && memMB >= 4096:
		return "turbo"
	case cpus >= 2 && memMB >= 2048:
		return "normal"
	default:
		return "eco"
	}
}

func ApplyTurbo(mode string) TurboStatus {
	if mode == "auto" {
		mode = autoDetectMode()
		log.Printf("[Turbo] 自动检测档位: %s", mode)
	}
	return applyProfile(calcProfile(mode))
}

func applyProfile(p TurboProfile) TurboStatus {
	turboMu.Lock()
	defer turboMu.Unlock()

	log.Printf("[Turbo] 切换到 [%s] 模式: CPU=%d, Mem=%dMB", p.Name, p.CPUCores, p.MemTotalMB)

	if database.DB != nil {
		database.DB.SetMaxOpenConns(p.DBMaxOpen)
		database.DB.SetMaxIdleConns(p.DBMaxIdle)
		database.DB.SetConnMaxLifetime(time.Duration(p.DBMaxLifetime) * time.Second)
		database.DB.SetConnMaxIdleTime(time.Duration(p.DBMaxIdleTime) * time.Second)
		log.Printf("[Turbo] DB连接池: open=%d idle=%d lifetime=%ds", p.DBMaxOpen, p.DBMaxIdle, p.DBMaxLifetime)
	}

	if cache.RDB != nil {
		log.Printf("[Turbo] Redis连接池: pool=%d minIdle=%d", p.RedisPoolSize, p.RedisMinIdle)
	}

	dockscheduler.Configure(time.Duration(p.PendingDockIntervalSec)*time.Second, p.DockBatchLimit)
	log.Printf("[Turbo] 待对接调度: batch=%d interval=%ds", p.DockBatchLimit, p.PendingDockIntervalSec)

	prev := runtime.GOMAXPROCS(p.GOMAXPROCS)
	log.Printf("[Turbo] GOMAXPROCS: %d -> %d", prev, p.GOMAXPROCS)
	oldGC := debug.SetGCPercent(p.GCPercent)
	log.Printf("[Turbo] GOGC: %d -> %d", oldGC, p.GCPercent)

	isEnabled := p.Name != "normal"
	if isEnabled {
		atomic.StoreInt32(&turboEnabled, 1)
	} else {
		atomic.StoreInt32(&turboEnabled, 0)
	}

	turboStatus = TurboStatus{
		Enabled:   isEnabled,
		Profile:   p,
		AppliedAt: time.Now().Format("2006-01-02 15:04:05"),
		Baseline:  turboBaseline,
	}
	return turboStatus
}

func GetTurboStatus() TurboStatus {
	turboMu.RLock()
	status := turboStatus
	turboMu.RUnlock()

	orderProgressMu.RLock()
	status.Profile.SyncIntervalSec = orderProgressConfig.IntervalSec
	orderProgressMu.RUnlock()
	return status
}

func normalizeOrderProgressSyncConfig(cfg OrderProgressSyncConfig, fallbackInterval time.Duration) OrderProgressSyncConfig {
	if cfg.IntervalSec <= 0 {
		cfg.IntervalSec = int(fallbackInterval / time.Second)
	}
	if cfg.IntervalSec < 10 {
		cfg.IntervalSec = 10
	}
	if cfg.IntervalSec > 86400 {
		cfg.IntervalSec = 86400
	}
	if cfg.BatchIntervalSec <= 0 {
		cfg.BatchIntervalSec = cfg.IntervalSec
	}
	if cfg.BatchIntervalSec < 10 {
		cfg.BatchIntervalSec = 10
	}
	if cfg.BatchIntervalSec > 86400 {
		cfg.BatchIntervalSec = 86400
	}
	if cfg.ExcludedStatuses == nil {
		cfg.ExcludedStatuses = []string{"已完成", "已退款", "已取消", "失败"}
	}
	cleanStatuses := make([]string, 0, len(cfg.ExcludedStatuses))
	seenStatuses := map[string]bool{}
	for _, status := range cfg.ExcludedStatuses {
		status = strings.TrimSpace(status)
		if status == "" || seenStatuses[status] {
			continue
		}
		seenStatuses[status] = true
		cleanStatuses = append(cleanStatuses, status)
	}
	cfg.ExcludedStatuses = cleanStatuses

	cleanSupplierIDs := make([]int, 0, len(cfg.SupplierIDs))
	seenSupplierIDs := map[int]bool{}
	for _, hid := range cfg.SupplierIDs {
		if hid <= 0 || seenSupplierIDs[hid] {
			continue
		}
		seenSupplierIDs[hid] = true
		cleanSupplierIDs = append(cleanSupplierIDs, hid)
	}
	cfg.SupplierIDs = cleanSupplierIDs
	cfg.Rules = normalizeOrderProgressRules(cfg.Rules)
	return cfg
}

func defaultOrderProgressRules() []ordermodule.AutoSyncRule {
	return []ordermodule.AutoSyncRule{
		{Key: "0_24h", Label: "0-24H", MinAgeHours: 0, MaxAgeHours: 24, IntervalMinutes: 10, Enabled: true},
		{Key: "1_3d", Label: "1-3d", MinAgeHours: 24, MaxAgeHours: 72, IntervalMinutes: 30, Enabled: true},
		{Key: "3_7d", Label: "3-7d", MinAgeHours: 72, MaxAgeHours: 168, IntervalMinutes: 120, Enabled: true},
		{Key: "7_15d", Label: "7-15d", MinAgeHours: 168, MaxAgeHours: 360, IntervalMinutes: 360, Enabled: true},
		{Key: "15_30d", Label: "15-30d", MinAgeHours: 360, MaxAgeHours: 720, IntervalMinutes: 720, Enabled: true},
		{Key: "30d_plus", Label: "30d+", MinAgeHours: 720, MaxAgeHours: 0, IntervalMinutes: 1440, Enabled: true},
	}
}

func normalizeOrderProgressRules(rules []ordermodule.AutoSyncRule) []ordermodule.AutoSyncRule {
	defaults := defaultOrderProgressRules()
	if len(rules) == 0 {
		return defaults
	}

	customByKey := map[string]ordermodule.AutoSyncRule{}
	for _, rule := range rules {
		customByKey[rule.Key] = rule
	}

	out := make([]ordermodule.AutoSyncRule, 0, len(defaults))
	for _, rule := range defaults {
		if custom, ok := customByKey[rule.Key]; ok {
			rule.IntervalMinutes = custom.IntervalMinutes
			rule.Enabled = custom.Enabled
			if custom.IntervalMinutes <= 0 {
				rule.IntervalMinutes = 1
			}
		}
		out = append(out, rule)
	}
	return out
}

func loadOrderProgressSyncConfig(fallbackInterval time.Duration) OrderProgressSyncConfig {
	cfg := normalizeOrderProgressSyncConfig(OrderProgressSyncConfig{Enabled: true, BatchEnabled: true}, fallbackInterval)
	if database.DB == nil {
		return cfg
	}

	var raw string
	err := database.DB.QueryRow(
		"SELECT COALESCE(svalue,'') FROM qingka_wangke_config WHERE skey = 'order_progress_sync_config' LIMIT 1",
	).Scan(&raw)
	if err != nil || strings.TrimSpace(raw) == "" {
		return cfg
	}

	stored := cfg
	if err := json.Unmarshal([]byte(raw), &stored); err != nil {
		return cfg
	}
	if !strings.Contains(raw, "\"batch_enabled\"") {
		stored.BatchEnabled = stored.Enabled
	}
	if !strings.Contains(raw, "\"batch_interval_sec\"") {
		stored.BatchIntervalSec = stored.IntervalSec
	}
	return normalizeOrderProgressSyncConfig(stored, fallbackInterval)
}

func saveOrderProgressSyncConfig(cfg OrderProgressSyncConfig) error {
	if database.DB == nil {
		return fmt.Errorf("database not initialized")
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	_, err = database.DB.Exec(
		"INSERT INTO qingka_wangke_config (v, k, skey, svalue) VALUES ('order_progress_sync_config', '', 'order_progress_sync_config', ?) ON DUPLICATE KEY UPDATE svalue = ?",
		string(raw), string(raw),
	)
	return err
}

func applyOrderProgressSyncConfig(cfg OrderProgressSyncConfig) {
	orderProgressMu.Lock()
	orderProgressConfig = cfg
	orderProgressStatus.Enabled = cfg.Enabled
	orderProgressStatus.IntervalSec = cfg.IntervalSec
	orderProgressStatus.BatchEnabled = cfg.BatchEnabled
	orderProgressStatus.BatchIntervalSec = cfg.BatchIntervalSec
	orderProgressStatus.SupplierIDs = append([]int(nil), cfg.SupplierIDs...)
	orderProgressStatus.ExcludedStatuses = append([]string(nil), cfg.ExcludedStatuses...)
	orderProgressStatus.Rules = append([]ordermodule.AutoSyncRule(nil), cfg.Rules...)
	orderProgressMu.Unlock()
}

func GetOrderProgressSyncStatus() OrderProgressSyncStatus {
	orderProgressMu.RLock()
	defer orderProgressMu.RUnlock()
	return orderProgressStatus
}

func GetOrderProgressSyncLogs(limit int) []OrderProgressSyncLogEntry {
	orderProgressMu.RLock()
	defer orderProgressMu.RUnlock()

	if limit <= 0 || limit > len(orderProgressLogs) {
		limit = len(orderProgressLogs)
	}
	if limit == 0 {
		return []OrderProgressSyncLogEntry{}
	}

	out := make([]OrderProgressSyncLogEntry, limit)
	for i := 0; i < limit; i++ {
		out[i] = orderProgressLogs[len(orderProgressLogs)-1-i]
	}
	return out
}

func UpdateOrderProgressSyncConfig(cfg OrderProgressSyncConfig) (OrderProgressSyncStatus, error) {
	cfg = normalizeOrderProgressSyncConfig(cfg, 2*time.Minute)
	if err := saveOrderProgressSyncConfig(cfg); err != nil {
		return GetOrderProgressSyncStatus(), err
	}
	applyOrderProgressSyncConfig(cfg)
	appendOrderProgressLog(OrderProgressSyncLogEntry{
		Time:             time.Now().Format("2006-01-02 15:04:05"),
		Mode:             "config",
		Trigger:          "config",
		IntervalSec:      cfg.IntervalSec,
		SupplierIDs:      append([]int(nil), cfg.SupplierIDs...),
		SupplierNames:    []string{},
		ExcludedStatuses: append([]string(nil), cfg.ExcludedStatuses...),
		RuleHits:         summarizeRuleHits(cfg.Rules),
		SampleErrors:     []string{},
	})
	updateSyncTickers(cfg)
	return GetOrderProgressSyncStatus(), nil
}

func RunOrderProgressSyncNow() (OrderProgressSyncStatus, error) {
	orderProgressMu.RLock()
	enabled := orderProgressConfig.Enabled
	batchEnabled := orderProgressConfig.BatchEnabled
	running := orderProgressStatus.Running
	batchRunning := orderProgressStatus.BatchRunning
	orderProgressMu.RUnlock()

	if !enabled && !batchEnabled {
		return GetOrderProgressSyncStatus(), fmt.Errorf("主订单同步未启用")
	}
	if running || batchRunning {
		return GetOrderProgressSyncStatus(), fmt.Errorf("主订单同步正在执行")
	}

	if enabled {
		syncPendingOrderProgress("manual", "single", true)
	}
	if batchEnabled {
		syncPendingOrderProgress("manual", "batch", false)
	}
	return GetOrderProgressSyncStatus(), nil
}

// InitSyncTicker 启动主订单表的自动轮询定时器。
// 这里故意只做“调度”，不写任何订单查询逻辑：
// 真正的业务规则放在 order.Sync.AutoSyncAllProgress 中，避免 runtimeops 与订单域耦合过深。
func InitSyncTicker(interval time.Duration) {
	syncTickerMu.Lock()
	defer syncTickerMu.Unlock()

	cfg := loadOrderProgressSyncConfig(interval)
	applyOrderProgressSyncConfig(cfg)

	syncTickerStop = make(chan struct{})
	batchSyncTickerStop = make(chan struct{})
	log.Printf("[AutoSync] 主订单自动同步已启动，首次执行延迟 10s，轮询间隔 %s，规则数 %d", time.Duration(cfg.IntervalSec)*time.Second, len(cfg.Rules))
	log.Printf("[AutoSync] 主订单批量进度同步已启动，首次执行延迟 10s，轮询间隔 %s", time.Duration(cfg.BatchIntervalSec)*time.Second)
	appendOrderProgressLog(OrderProgressSyncLogEntry{
		Time:             time.Now().Format("2006-01-02 15:04:05"),
		Mode:             "single",
		Trigger:          "system",
		IntervalSec:      cfg.IntervalSec,
		SupplierIDs:      append([]int(nil), cfg.SupplierIDs...),
		SupplierNames:    []string{},
		ExcludedStatuses: append([]string(nil), cfg.ExcludedStatuses...),
		RuleHits:         summarizeRuleHits(cfg.Rules),
		SampleErrors:     []string{},
	})
	appendOrderProgressLog(OrderProgressSyncLogEntry{
		Time:             time.Now().Format("2006-01-02 15:04:05"),
		Mode:             "batch",
		Trigger:          "system",
		IntervalSec:      cfg.BatchIntervalSec,
		SupplierIDs:      append([]int(nil), cfg.SupplierIDs...),
		SupplierNames:    []string{},
		ExcludedStatuses: append([]string(nil), cfg.ExcludedStatuses...),
		RuleHits:         map[string]int{},
		SampleErrors:     []string{},
	})
	go runSyncTicker(time.Duration(cfg.IntervalSec)*time.Second, syncTickerStop, cfg.Enabled, "single")
	go runSyncTicker(time.Duration(cfg.BatchIntervalSec)*time.Second, batchSyncTickerStop, cfg.BatchEnabled, "batch")
}

// runSyncTicker 的职责很单纯：
// - 启动 10 秒后先触发一次，避免服务刚起时长时间没有任何自动同步；
// - 然后按给定 interval 循环执行；
// - 收到 stop 信号后退出，供 turbo 档位热更新时重建 ticker。
func runSyncTicker(interval time.Duration, stop chan struct{}, enabled bool, mode string) {
	if !enabled {
		setOrderProgressNextRun(mode, "")
		return
	}

	setOrderProgressNextRun(mode, time.Now().Add(10*time.Second).Format("2006-01-02 15:04:05"))
	select {
	case <-time.After(10 * time.Second):
		syncPendingOrderProgress("auto", mode, false)
	case <-stop:
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		setOrderProgressNextRun(mode, time.Now().Add(interval).Format("2006-01-02 15:04:05"))
		select {
		case <-ticker.C:
			syncPendingOrderProgress("auto", mode, false)
		case <-stop:
			return
		}
	}
}

// syncPendingOrderProgress 是调度层到订单域的桥接点。
// 调度层只关心“现在该执行一次了”，至于扫哪些订单、怎么查上游、怎么打日志，
// 都交给 order 模块内部完成。
func syncPendingOrderProgress(trigger string, mode string, ignoreRules bool) {
	if database.DB == nil {
		log.Printf("[AutoSync] 跳过执行：数据库未初始化")
		return
	}
	if trigger == "" {
		trigger = "auto"
	}
	if mode == "" {
		mode = "single"
	}

	orderProgressMu.Lock()
	switch mode {
	case "batch":
		if orderProgressStatus.BatchRunning {
			orderProgressMu.Unlock()
			log.Printf("[AutoSync] 跳过执行：批量同步已在进行中")
			return
		}
		orderProgressStatus.BatchRunning = true
		orderProgressStatus.BatchLastError = ""
	default:
		if orderProgressStatus.Running {
			orderProgressMu.Unlock()
			log.Printf("[AutoSync] 跳过执行：单条同步已在进行中")
			return
		}
		orderProgressStatus.Running = true
		orderProgressStatus.LastError = ""
	}
	orderProgressMu.Unlock()

	defer func() {
		orderProgressMu.Lock()
		switch mode {
		case "batch":
			orderProgressStatus.BatchRunning = false
		default:
			orderProgressStatus.Running = false
		}
		orderProgressMu.Unlock()
	}()

	if mode == "batch" {
		log.Printf("[AutoSync] 开始执行主订单批量进度同步")
	} else {
		log.Printf("[AutoSync] 开始执行主订单自动同步")
	}
	orderProgressMu.RLock()
	opts := ordermodule.AutoSyncOptions{SupplierHIDs: append([]int(nil), orderProgressConfig.SupplierIDs...), ExcludedStatuses: append([]string(nil), orderProgressConfig.ExcludedStatuses...)}
	intervalSec := orderProgressConfig.IntervalSec
	ruleHits := summarizeRuleHits(orderProgressConfig.Rules)
	if mode == "batch" {
		opts.OnlyBatchSuppliers = true
		opts.IgnoreRules = true
		intervalSec = orderProgressConfig.BatchIntervalSec
		ruleHits = map[string]int{}
	} else {
		opts.SkipBatchSuppliers = true
		if ignoreRules {
			opts.IgnoreRules = true
		} else {
			opts.Rules = append([]ordermodule.AutoSyncRule(nil), orderProgressConfig.Rules...)
		}
	}
	orderProgressMu.RUnlock()

	var collectedLines []string
	var collectedMu sync.Mutex
	logTime := time.Now().Format("2006-01-02 15:04:05")
	if mode == "batch" {
		collectedLines = append(collectedLines, logTime+" [AutoSync] 开始执行主订单批量进度同步")
	} else {
		collectedLines = append(collectedLines, logTime+" [AutoSync] 开始执行主订单自动同步")
	}
	opts.LogCollector = func(line string) {
		collectedMu.Lock()
		collectedLines = append(collectedLines, time.Now().Format("2006-01-02 15:04:05")+" [AutoSync] "+line)
		collectedMu.Unlock()
	}

	start := time.Now()
	updated, failed, err := ordermodule.NewServices().Sync.AutoSyncAllProgress(opts)
	report := ordermodule.GetLastAutoSyncReport()
	orderProgressMu.Lock()
	switch mode {
	case "batch":
		orderProgressStatus.BatchLastRunTime = time.Now().Format("2006-01-02 15:04:05")
		orderProgressStatus.BatchLastUpdated = updated
		orderProgressStatus.BatchLastFailed = failed
		orderProgressStatus.BatchTotalRuns++
		if err != nil {
			orderProgressStatus.BatchLastError = err.Error()
		}
	default:
		orderProgressStatus.LastRunTime = time.Now().Format("2006-01-02 15:04:05")
		orderProgressStatus.LastUpdated = updated
		orderProgressStatus.LastFailed = failed
		orderProgressStatus.TotalRuns++
		if err != nil {
			orderProgressStatus.LastError = err.Error()
		}
	}
	orderProgressMu.Unlock()

	// 收集完成行
	finishTime := time.Now().Format("2006-01-02 15:04:05")
	if err != nil {
		if mode == "batch" {
			collectedLines = append(collectedLines, finishTime+" [AutoSync] 批量进度同步失败: "+err.Error())
		} else {
			collectedLines = append(collectedLines, finishTime+" [AutoSync] 查询进行中订单失败: "+err.Error())
		}
	} else {
		if mode == "batch" {
			collectedLines = append(collectedLines, fmt.Sprintf("%s [AutoSync] 批量进度同步完成，更新 %d 个订单，失败 %d 个", finishTime, updated, failed))
		} else {
			collectedLines = append(collectedLines, fmt.Sprintf("%s [AutoSync] 同步完成，更新 %d 个订单，失败 %d 个", finishTime, updated, failed))
		}
	}

	appendOrderProgressLog(OrderProgressSyncLogEntry{
		Time:             logTime,
		Mode:             mode,
		Trigger:          trigger,
		IntervalSec:      intervalSec,
		SupplierIDs:      append([]int(nil), opts.SupplierHIDs...),
		SupplierNames:    append([]string(nil), report.SupplierNames...),
		ExcludedStatuses: append([]string(nil), opts.ExcludedStatuses...),
		RuleHits:         ruleHits,
		SampleErrors:     append([]string(nil), report.SampleErrors...),
		Updated:          updated,
		Failed:           failed,
		DurationMs:       time.Since(start).Milliseconds(),
		Error:            pickOrderProgressError(err, report.SampleErrors),
		Lines:            collectedLines,
	})
	if err != nil {
		if mode == "batch" {
			log.Printf("[AutoSync] 批量进度同步失败: %v", err)
		} else {
			log.Printf("[AutoSync] 查询进行中订单失败: %v", err)
		}
		return
	}
	if mode == "batch" {
		log.Printf("[AutoSync] 批量进度同步完成，更新 %d 个订单，失败 %d 个", updated, failed)
	} else {
		log.Printf("[AutoSync] 同步完成，更新 %d 个订单，失败 %d 个", updated, failed)
	}
}

func setOrderProgressNextRun(mode string, next string) {
	orderProgressMu.Lock()
	switch mode {
	case "batch":
		orderProgressStatus.BatchNextRunTime = next
	default:
		orderProgressStatus.NextRunTime = next
	}
	orderProgressMu.Unlock()
}

func updateSyncTickers(cfg OrderProgressSyncConfig) {
	syncTickerMu.Lock()
	defer syncTickerMu.Unlock()

	if syncTickerStop != nil {
		close(syncTickerStop)
	}
	if batchSyncTickerStop != nil {
		close(batchSyncTickerStop)
	}
	syncTickerStop = make(chan struct{})
	batchSyncTickerStop = make(chan struct{})
	if !cfg.Enabled {
		log.Printf("[AutoSync] 主订单自动同步已停用")
		setOrderProgressNextRun("single", "")
	} else {
		log.Printf("[AutoSync] 主订单自动同步间隔已更新为 %s", time.Duration(cfg.IntervalSec)*time.Second)
		go runSyncTicker(time.Duration(cfg.IntervalSec)*time.Second, syncTickerStop, true, "single")
	}
	if !cfg.BatchEnabled {
		log.Printf("[AutoSync] 主订单批量进度同步已停用")
		setOrderProgressNextRun("batch", "")
	} else {
		log.Printf("[AutoSync] 主订单批量进度同步间隔已更新为 %s", time.Duration(cfg.BatchIntervalSec)*time.Second)
		go runSyncTicker(time.Duration(cfg.BatchIntervalSec)*time.Second, batchSyncTickerStop, true, "batch")
	}
}

func appendOrderProgressLog(entry OrderProgressSyncLogEntry) {
	orderProgressMu.Lock()
	defer orderProgressMu.Unlock()

	orderProgressLogID++
	entry.ID = orderProgressLogID
	if entry.Time == "" {
		entry.Time = time.Now().Format("2006-01-02 15:04:05")
	}
	orderProgressLogs = append(orderProgressLogs, entry)
	if len(orderProgressLogs) > 50 {
		orderProgressLogs = append([]OrderProgressSyncLogEntry(nil), orderProgressLogs[len(orderProgressLogs)-50:]...)
	}
}

func errorToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func pickOrderProgressError(err error, sampleErrors []string) string {
	if err != nil {
		return err.Error()
	}
	if len(sampleErrors) > 0 {
		return sampleErrors[0]
	}
	return ""
}

func summarizeRuleHits(rules []ordermodule.AutoSyncRule) map[string]int {
	out := map[string]int{}
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		out[rule.Label] = rule.IntervalMinutes
	}
	return out
}
