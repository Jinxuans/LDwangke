package runtimeops

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

var (
	turboMu       sync.RWMutex
	turboEnabled  int32
	turboStatus   TurboStatus
	turboBaseline TurboProfile

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

func GetTurboStatus() TurboStatus {
	turboMu.RLock()
	defer turboMu.RUnlock()
	return turboStatus
}

// InitSyncTicker 启动主订单表的自动轮询定时器。
// 这里故意只做“调度”，不写任何订单查询逻辑：
// 真正的业务规则放在 order.Sync.AutoSyncAllProgress 中，避免 runtimeops 与订单域耦合过深。
func InitSyncTicker(interval time.Duration) {
	syncTickerMu.Lock()
	defer syncTickerMu.Unlock()

	syncTickerStop = make(chan struct{})
	log.Printf("[AutoSync] 主订单自动同步已启动，首次执行延迟 10s，轮询间隔 %s", interval)
	go runSyncTicker(interval, syncTickerStop)
}

// runSyncTicker 的职责很单纯：
// - 启动 10 秒后先触发一次，避免服务刚起时长时间没有任何自动同步；
// - 然后按给定 interval 循环执行；
// - 收到 stop 信号后退出，供 turbo 档位热更新时重建 ticker。
func runSyncTicker(interval time.Duration, stop chan struct{}) {
	select {
	case <-time.After(10 * time.Second):
		syncPendingOrderProgress()
	case <-stop:
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			syncPendingOrderProgress()
		case <-stop:
			return
		}
	}
}

// syncPendingOrderProgress 是调度层到订单域的桥接点。
// 调度层只关心“现在该执行一次了”，至于扫哪些订单、怎么查上游、怎么打日志，
// 都交给 order 模块内部完成。
func syncPendingOrderProgress() {
	if database.DB == nil {
		log.Printf("[AutoSync] 跳过执行：数据库未初始化")
		return
	}

	log.Printf("[AutoSync] 开始执行主订单自动同步")
	updated, failed, err := ordermodule.NewServices().Sync.AutoSyncAllProgress()
	if err != nil {
		log.Printf("[AutoSync] 查询进行中订单失败: %v", err)
		return
	}
	log.Printf("[AutoSync] 同步完成，更新 %d 个订单，失败 %d 个", updated, failed)
}

func updateSyncInterval(interval time.Duration) {
	syncTickerMu.Lock()
	defer syncTickerMu.Unlock()

	if syncTickerStop != nil {
		close(syncTickerStop)
	}
	syncTickerStop = make(chan struct{})
	log.Printf("[AutoSync] 主订单自动同步间隔已更新为 %s", interval)
	go runSyncTicker(interval, syncTickerStop)
}
