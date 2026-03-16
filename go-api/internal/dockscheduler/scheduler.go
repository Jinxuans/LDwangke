package dockscheduler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"go-api/internal/database"
	ordermodule "go-api/internal/modules/order"
)

const (
	defaultInterval   = 30 * time.Second
	defaultBatchSize  = 100
	startDelay        = 10 * time.Second
	maxLogEntries     = 50
	configSKey        = "pending_dock_scheduler_config"
	minIntervalSecond = 5
	maxIntervalSecond = 3600
	minBatchLimit     = 1
	maxBatchLimit     = 1000
)

type Config struct {
	IntervalSec int `json:"interval_sec"`
	BatchLimit  int `json:"batch_limit"`
}

type LogEntry struct {
	ID            int64  `json:"id"`
	Time          string `json:"time"`
	Trigger       string `json:"trigger"`
	Level         string `json:"level"`
	Message       string `json:"message"`
	Fetched       int    `json:"fetched"`
	Success       int    `json:"success"`
	Fail          int    `json:"fail"`
	PendingBefore int64  `json:"pending_before"`
	PendingAfter  int64  `json:"pending_after"`
	DurationMs    int64  `json:"duration_ms"`
}

type Stats struct {
	Running      bool   `json:"running"`
	Active       int    `json:"active"`
	Pending      int64  `json:"pending"`
	IntervalSec  int    `json:"interval_sec"`
	BatchLimit   int    `json:"batch_limit"`
	LastFetched  int    `json:"last_fetched"`
	LastSuccess  int    `json:"last_success"`
	LastFail     int    `json:"last_fail"`
	TotalSuccess int64  `json:"total_success"`
	TotalFail    int64  `json:"total_fail"`
	TotalRuns    int64  `json:"total_runs"`
	LastRunTime  string `json:"last_run_time"`
	LastTrigger  string `json:"last_trigger"`
	LastError    string `json:"last_error"`
}

type scheduler struct {
	mu sync.RWMutex

	running     bool
	active      bool
	intervalSec int
	batchLimit  int

	lastFetched  int
	lastSuccess  int
	lastFail     int
	totalSuccess int64
	totalFail    int64
	totalRuns    int64
	lastRunTime  string
	lastTrigger  string
	lastError    string

	nextLogID int64
	logs      []LogEntry
}

var global = &scheduler{
	intervalSec: int(defaultInterval / time.Second),
	batchLimit:  defaultBatchSize,
}

func Start(ctx context.Context, interval time.Duration, batchLimit int) {
	Configure(interval, batchLimit)
	if cfg, err := loadPersistedConfig(); err == nil {
		Configure(time.Duration(cfg.IntervalSec)*time.Second, cfg.BatchLimit)
	}

	global.mu.Lock()
	if global.running {
		global.mu.Unlock()
		return
	}
	global.running = true
	currentInterval := time.Duration(global.intervalSec) * time.Second
	currentBatch := global.batchLimit
	global.mu.Unlock()

	log.Printf("[PendingDock] 调度器启动，首次执行延迟 %s，轮询间隔 %s，单轮批量 %d", startDelay, currentInterval, currentBatch)
	appendLog(LogEntry{
		Time:    time.Now().Format("2006-01-02 15:04:05"),
		Trigger: "system",
		Level:   "info",
		Message: fmt.Sprintf("调度器启动，间隔 %ds，批量 %d", int(currentInterval/time.Second), currentBatch),
	})

	go func() {
		defer func() {
			global.mu.Lock()
			global.running = false
			global.active = false
			global.mu.Unlock()
		}()

		if !sleepContext(ctx, startDelay) {
			return
		}

		_, _ = RunOnce("auto")

		for {
			interval := getInterval()
			if !sleepContext(ctx, interval) {
				return
			}
			_, _ = RunOnce("auto")
		}
	}()
}

func Configure(interval time.Duration, batchLimit int) {
	cfg := normalizeConfig(Config{
		IntervalSec: int(interval / time.Second),
		BatchLimit:  batchLimit,
	})

	global.mu.Lock()
	defer global.mu.Unlock()

	global.intervalSec = cfg.IntervalSec
	global.batchLimit = cfg.BatchLimit
}

func RunOnce(trigger string) (Stats, error) {
	global.mu.Lock()
	if global.active {
		global.mu.Unlock()
		return Stats{}, errors.New("待对接订单调度正在执行")
	}
	global.active = true
	global.mu.Unlock()

	defer func() {
		global.mu.Lock()
		global.active = false
		global.mu.Unlock()
	}()

	if trigger == "" {
		trigger = "manual"
	}

	pendingBefore, _ := countPendingOrders()
	start := time.Now()

	if database.DB == nil {
		markRun(trigger, 0, 0, 0, pendingBefore, pendingBefore, time.Since(start), "数据库未初始化")
		return Snapshot(), errors.New("数据库未初始化")
	}

	limit := getBatchLimit()
	oids, err := fetchPendingOrderIDs(limit)
	if err != nil {
		markRun(trigger, 0, 0, 0, pendingBefore, pendingBefore, time.Since(start), err.Error())
		return Snapshot(), err
	}

	success, fail := 0, 0
	var runErr error
	if len(oids) > 0 {
		success, fail, runErr = ordermodule.NewServices().Sync.ManualDock(oids)
	}

	pendingAfter, _ := countPendingOrders()
	var errMsg string
	if runErr != nil {
		errMsg = runErr.Error()
	}
	markRun(trigger, len(oids), success, fail, pendingBefore, pendingAfter, time.Since(start), errMsg)

	if runErr != nil {
		return Snapshot(), runErr
	}
	return Snapshot(), nil
}

func UpdateConfig(intervalSec, batchLimit int) (Stats, error) {
	cfg := normalizeConfig(Config{IntervalSec: intervalSec, BatchLimit: batchLimit})
	Configure(time.Duration(cfg.IntervalSec)*time.Second, cfg.BatchLimit)
	if err := savePersistedConfig(cfg); err != nil {
		return Snapshot(), err
	}
	appendLog(LogEntry{
		Time:    time.Now().Format("2006-01-02 15:04:05"),
		Trigger: "config",
		Level:   "info",
		Message: fmt.Sprintf("配置已更新：间隔 %ds，批量 %d", cfg.IntervalSec, cfg.BatchLimit),
	})
	return Snapshot(), nil
}

func Snapshot() Stats {
	pending, _ := countPendingOrders()

	global.mu.RLock()
	defer global.mu.RUnlock()

	active := 0
	if global.active {
		active = 1
	}

	return Stats{
		Running:      global.running,
		Active:       active,
		Pending:      pending,
		IntervalSec:  global.intervalSec,
		BatchLimit:   global.batchLimit,
		LastFetched:  global.lastFetched,
		LastSuccess:  global.lastSuccess,
		LastFail:     global.lastFail,
		TotalSuccess: global.totalSuccess,
		TotalFail:    global.totalFail,
		TotalRuns:    global.totalRuns,
		LastRunTime:  global.lastRunTime,
		LastTrigger:  global.lastTrigger,
		LastError:    global.lastError,
	}
}

func RecentLogs(limit int) []LogEntry {
	global.mu.RLock()
	defer global.mu.RUnlock()

	if limit <= 0 || limit > len(global.logs) {
		limit = len(global.logs)
	}
	if limit == 0 {
		return []LogEntry{}
	}

	out := make([]LogEntry, limit)
	for i := 0; i < limit; i++ {
		out[i] = global.logs[len(global.logs)-1-i]
	}
	return out
}

func countPendingOrders() (int64, error) {
	if database.DB == nil {
		return 0, nil
	}
	var pending int64
	err := database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE dockstatus IN (0, 2)").Scan(&pending)
	return pending, err
}

func fetchPendingOrderIDs(limit int) ([]int, error) {
	if limit <= 0 {
		limit = defaultBatchSize
	}
	query := fmt.Sprintf(
		"SELECT oid FROM qingka_wangke_order WHERE dockstatus IN (0, 2) ORDER BY CASE WHEN dockstatus = 0 THEN 0 ELSE 1 END, oid ASC LIMIT %d",
		limit,
	)
	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var oids []int
	for rows.Next() {
		var oid int
		if err := rows.Scan(&oid); err != nil {
			continue
		}
		oids = append(oids, oid)
	}
	return oids, rows.Err()
}

func markRun(trigger string, fetched, success, fail int, pendingBefore, pendingAfter int64, duration time.Duration, errMsg string) {
	global.mu.Lock()
	defer global.mu.Unlock()

	global.lastFetched = fetched
	global.lastSuccess = success
	global.lastFail = fail
	global.totalSuccess += int64(success)
	global.totalFail += int64(fail)
	global.totalRuns++
	global.lastRunTime = time.Now().Format("2006-01-02 15:04:05")
	global.lastTrigger = trigger
	global.lastError = errMsg

	level := "info"
	message := fmt.Sprintf("本轮抓取 %d 单，成功 %d，失败 %d", fetched, success, fail)
	if errMsg != "" {
		level = "error"
		message = errMsg
	} else if fetched == 0 {
		message = "当前没有待对接订单"
	}

	global.nextLogID++
	global.logs = append(global.logs, LogEntry{
		ID:            global.nextLogID,
		Time:          global.lastRunTime,
		Trigger:       trigger,
		Level:         level,
		Message:       message,
		Fetched:       fetched,
		Success:       success,
		Fail:          fail,
		PendingBefore: pendingBefore,
		PendingAfter:  pendingAfter,
		DurationMs:    duration.Milliseconds(),
	})
	if len(global.logs) > maxLogEntries {
		global.logs = append([]LogEntry(nil), global.logs[len(global.logs)-maxLogEntries:]...)
	}
}

func getInterval() time.Duration {
	global.mu.RLock()
	defer global.mu.RUnlock()
	if global.intervalSec <= 0 {
		return defaultInterval
	}
	return time.Duration(global.intervalSec) * time.Second
}

func getBatchLimit() int {
	global.mu.RLock()
	defer global.mu.RUnlock()
	if global.batchLimit <= 0 {
		return defaultBatchSize
	}
	return global.batchLimit
}

func sleepContext(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func appendLog(entry LogEntry) {
	global.mu.Lock()
	defer global.mu.Unlock()

	global.nextLogID++
	entry.ID = global.nextLogID
	if entry.Time == "" {
		entry.Time = time.Now().Format("2006-01-02 15:04:05")
	}
	global.logs = append(global.logs, entry)
	if len(global.logs) > maxLogEntries {
		global.logs = append([]LogEntry(nil), global.logs[len(global.logs)-maxLogEntries:]...)
	}
}

func normalizeConfig(cfg Config) Config {
	if cfg.IntervalSec <= 0 {
		cfg.IntervalSec = int(defaultInterval / time.Second)
	}
	if cfg.IntervalSec < minIntervalSecond {
		cfg.IntervalSec = minIntervalSecond
	}
	if cfg.IntervalSec > maxIntervalSecond {
		cfg.IntervalSec = maxIntervalSecond
	}
	if cfg.BatchLimit <= 0 {
		cfg.BatchLimit = defaultBatchSize
	}
	if cfg.BatchLimit < minBatchLimit {
		cfg.BatchLimit = minBatchLimit
	}
	if cfg.BatchLimit > maxBatchLimit {
		cfg.BatchLimit = maxBatchLimit
	}
	return cfg
}

func loadPersistedConfig() (Config, error) {
	var raw string
	err := database.DB.QueryRow("SELECT COALESCE(svalue,'') FROM qingka_wangke_config WHERE skey = ? LIMIT 1", configSKey).Scan(&raw)
	if err != nil {
		if err == sql.ErrNoRows {
			return normalizeConfig(Config{}), nil
		}
		return normalizeConfig(Config{}), err
	}
	if raw == "" {
		return normalizeConfig(Config{}), nil
	}
	var cfg Config
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return normalizeConfig(Config{}), err
	}
	return normalizeConfig(cfg), nil
}

func savePersistedConfig(cfg Config) error {
	cfg = normalizeConfig(cfg)
	raw, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	_, err = database.DB.Exec(
		"INSERT INTO qingka_wangke_config (v, k, skey, svalue) VALUES (?, '', ?, ?) ON DUPLICATE KEY UPDATE svalue = ?",
		configSKey, configSKey, string(raw), string(raw),
	)
	return err
}
