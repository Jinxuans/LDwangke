package service

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"go-api/internal/cache"
	"go-api/internal/database"
	"go-api/internal/queue"
)

// ===== 狂暴模式（Turbo Mode）=====
// 自动识别服务器 CPU/内存，动态调整各子系统参数，热生效无需重启

// TurboProfile 性能档位参数
type TurboProfile struct {
	Name       string `json:"name"` // eco / normal / turbo / insane
	CPUCores   int    `json:"cpu_cores"`
	MemTotalMB int    `json:"mem_total_mb"`
	GOOS       string `json:"goos"`
	GOARCH     string `json:"goarch"`

	// DB 连接池
	DBMaxOpen     int `json:"db_max_open"`
	DBMaxIdle     int `json:"db_max_idle"`
	DBMaxLifetime int `json:"db_max_lifetime_sec"`
	DBMaxIdleTime int `json:"db_max_idle_time_sec"`

	// Redis 连接池
	RedisPoolSize int `json:"redis_pool_size"`
	RedisMinIdle  int `json:"redis_min_idle"`

	// 对接队列
	DockWorkers   int `json:"dock_workers"`
	DockQueueSize int `json:"dock_queue_size"`

	// 同步间隔
	SyncIntervalSec int `json:"sync_interval_sec"`

	// Go Runtime
	GOMAXPROCS int `json:"gomaxprocs"`
	GCPercent  int `json:"gc_percent"`
}

// TurboStatus 当前状态
type TurboStatus struct {
	Enabled   bool         `json:"enabled"`
	Profile   TurboProfile `json:"profile"`
	AppliedAt string       `json:"applied_at"`
	Baseline  TurboProfile `json:"baseline"`
}

var (
	turboMu       sync.RWMutex
	turboEnabled  int32 // atomic
	turboStatus   TurboStatus
	turboBaseline TurboProfile

	// 同步 ticker 热替换
	syncTickerMu   sync.Mutex
	syncTickerStop chan struct{}
)

func init() {
	turboBaseline = detectBaseline()
	turboStatus = TurboStatus{
		Enabled:  false,
		Profile:  turboBaseline,
		Baseline: turboBaseline,
	}
}

// ---------- 硬件检测 ----------

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
	// Windows/macOS: 用 MemStats.Sys 估算
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
		Name:            "normal",
		CPUCores:        runtime.NumCPU(),
		MemTotalMB:      getMemTotalMB(),
		GOOS:            runtime.GOOS,
		GOARCH:          runtime.GOARCH,
		DBMaxOpen:       50,
		DBMaxIdle:       25,
		DBMaxLifetime:   300,
		DBMaxIdleTime:   180,
		RedisPoolSize:   10,
		RedisMinIdle:    3,
		DockWorkers:     5,
		DockQueueSize:   1000,
		SyncIntervalSec: 120,
		GOMAXPROCS:      runtime.NumCPU(),
		GCPercent:       100,
	}
}

// ---------- 档位计算 ----------

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// CalcProfile 根据硬件自动计算性能档位参数
func CalcProfile(mode string) TurboProfile {
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
		p.DockWorkers = clampInt(cpus, 2, 5)
		p.DockQueueSize = 500
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
		p.DockWorkers = clampInt(cpus*2, 5, 15)
		p.DockQueueSize = 1000
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
		p.DockWorkers = clampInt(cpus*3, 10, 30)
		p.DockQueueSize = 2000
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
		p.DockWorkers = clampInt(cpus*5, 15, 50)
		p.DockQueueSize = 5000
		p.SyncIntervalSec = 30
		p.GOMAXPROCS = cpus
		p.GCPercent = 30

	default:
		return CalcProfile("normal")
	}

	// 内存不足降级保护
	if memMB < 1024 && (mode == "insane" || mode == "turbo") {
		log.Printf("[Turbo] 内存仅 %dMB，自动降级参数", memMB)
		p.DBMaxOpen = min(p.DBMaxOpen, 50)
		p.DBMaxIdle = min(p.DBMaxIdle, 25)
		p.RedisPoolSize = min(p.RedisPoolSize, 15)
		p.DockWorkers = min(p.DockWorkers, 10)
		p.GCPercent = max(p.GCPercent, 80)
	}

	return p
}

// AutoDetectMode 根据硬件自动选择最佳档位
func AutoDetectMode() string {
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

// ---------- 应用配置（热生效） ----------

// ApplyTurbo 切换到指定模式
func ApplyTurbo(mode string) TurboStatus {
	if mode == "auto" {
		mode = AutoDetectMode()
		log.Printf("[Turbo] 自动检测档位: %s", mode)
	}
	profile := CalcProfile(mode)
	return applyProfile(profile)
}

func applyProfile(p TurboProfile) TurboStatus {
	turboMu.Lock()
	defer turboMu.Unlock()

	log.Printf("[Turbo] 切换到 [%s] 模式: CPU=%d, Mem=%dMB", p.Name, p.CPUCores, p.MemTotalMB)

	// 1. DB 连接池
	if database.DB != nil {
		database.DB.SetMaxOpenConns(p.DBMaxOpen)
		database.DB.SetMaxIdleConns(p.DBMaxIdle)
		database.DB.SetConnMaxLifetime(time.Duration(p.DBMaxLifetime) * time.Second)
		database.DB.SetConnMaxIdleTime(time.Duration(p.DBMaxIdleTime) * time.Second)
		log.Printf("[Turbo] DB连接池: open=%d idle=%d lifetime=%ds", p.DBMaxOpen, p.DBMaxIdle, p.DBMaxLifetime)
	}

	// 2. Redis 连接池（记录参数，新连接会遵循）
	if cache.RDB != nil {
		log.Printf("[Turbo] Redis连接池: pool=%d minIdle=%d", p.RedisPoolSize, p.RedisMinIdle)
	}

	// 3. 对接队列 worker 数
	if queue.GlobalDockQueue != nil {
		queue.GlobalDockQueue.SetMaxWorkers(p.DockWorkers)
		log.Printf("[Turbo] 对接队列: workers=%d", p.DockWorkers)
	}

	// 4. Go Runtime
	prev := runtime.GOMAXPROCS(p.GOMAXPROCS)
	log.Printf("[Turbo] GOMAXPROCS: %d → %d", prev, p.GOMAXPROCS)

	oldGC := debug.SetGCPercent(p.GCPercent)
	log.Printf("[Turbo] GOGC: %d → %d", oldGC, p.GCPercent)

	// 5. 同步间隔（热替换 ticker）
	updateSyncInterval(time.Duration(p.SyncIntervalSec) * time.Second)
	log.Printf("[Turbo] 同步间隔: %ds", p.SyncIntervalSec)

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

// DisableTurbo 恢复到 normal
func DisableTurbo() TurboStatus {
	return ApplyTurbo("normal")
}

// GetTurboStatus 获取当前状态
func GetTurboStatus() TurboStatus {
	turboMu.RLock()
	defer turboMu.RUnlock()
	return turboStatus
}

// IsTurboEnabled 是否启用了非 normal 模式
func IsTurboEnabled() bool {
	return atomic.LoadInt32(&turboEnabled) == 1
}

// ---------- 同步 ticker 热替换 ----------

// InitSyncTicker 初始化同步定时器（main.go 启动时调用，替代硬编码 goroutine）
func InitSyncTicker(interval time.Duration) {
	syncTickerMu.Lock()
	defer syncTickerMu.Unlock()

	syncTickerStop = make(chan struct{})
	go runSyncTicker(interval, syncTickerStop)
}

func runSyncTicker(interval time.Duration, stop chan struct{}) {
	// 启动10秒后先跑一次
	select {
	case <-time.After(10 * time.Second):
		AutoSyncAllProgress()
	case <-stop:
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			AutoSyncAllProgress()
		case <-stop:
			return
		}
	}
}

func updateSyncInterval(interval time.Duration) {
	syncTickerMu.Lock()
	defer syncTickerMu.Unlock()

	if syncTickerStop != nil {
		close(syncTickerStop)
	}
	syncTickerStop = make(chan struct{})
	go runSyncTicker(interval, syncTickerStop)
}
