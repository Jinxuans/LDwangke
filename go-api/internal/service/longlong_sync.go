package service

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"go-api/internal/database"
)

// ── 龙龙内置同步 & 监听（替代 long sync / long listen 命令行工具） ──

// longlongState 全局运行状态
var longlongState struct {
	mu            sync.RWMutex
	syncRunning   bool
	listenRunning bool
	lastSyncTime  string
	lastSyncMsg   string
	lastListenMsg string
	lastListenAt  string
	syncCount     int
	listenCount   int
}

// LonglongStatus 返回当前同步/监听状态
func LonglongStatus() map[string]interface{} {
	longlongState.mu.RLock()
	defer longlongState.mu.RUnlock()
	return map[string]interface{}{
		"sync_running":    longlongState.syncRunning,
		"listen_running":  longlongState.listenRunning,
		"last_sync_time":  longlongState.lastSyncTime,
		"last_sync_msg":   longlongState.lastSyncMsg,
		"last_listen_at":  longlongState.lastListenAt,
		"last_listen_msg": longlongState.lastListenMsg,
		"sync_count":      longlongState.syncCount,
		"listen_count":    longlongState.listenCount,
	}
}

// longlongCfg 龙龙配置（从 DB 读取）
type longlongCfg struct {
	LongHost   string  `json:"long_host"`
	AccessKey  string  `json:"access_key"`
	Docking    string  `json:"docking"`
	Rate       float64 `json:"rate"`
	NamePrefix string  `json:"name_prefix"`
	Category   string  `json:"category"`
	CoverPrice bool    `json:"cover_price"`
	CoverDesc  bool    `json:"cover_desc"`
	CoverName  bool    `json:"cover_name"`
	Sort       string  `json:"sort"`
	CronValue  string  `json:"cron_value"`
	CronUnit   string  `json:"cron_unit"`
}

func loadLonglongCfg() (*longlongCfg, error) {
	var raw string
	err := database.DB.QueryRow("SELECT `k` FROM qingka_wangke_config WHERE `v` = 'longlong_tool_config'").Scan(&raw)
	if err != nil || raw == "" {
		return nil, fmt.Errorf("龙龙配置未设置")
	}
	var cfg longlongCfg
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, fmt.Errorf("配置解析失败")
	}
	if cfg.LongHost == "" || cfg.AccessKey == "" {
		return nil, fmt.Errorf("龙龙配置不完整（缺少 host 或 key）")
	}
	if cfg.Rate <= 0 {
		cfg.Rate = 1.5
	}
	if cfg.CronValue == "" {
		cfg.CronValue = "30"
	}
	if cfg.CronUnit == "" {
		cfg.CronUnit = "minute"
	}
	return &cfg, nil
}

// ensureLonglongHuoyuan 确保 huoyuan 表有龙龙供应商记录，返回 hid
func ensureLonglongHuoyuan(cfg *longlongCfg) (int, error) {
	// 1) 如果配置指定了 docking hid，直接用
	if docking, _ := strconv.Atoi(cfg.Docking); docking > 0 {
		var count int
		database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_huoyuan WHERE hid = ?", docking).Scan(&count)
		if count > 0 {
			database.DB.Exec(
				"UPDATE qingka_wangke_huoyuan SET url = ?, pass = ?, pt = 'longlong', name = '龙龙平台' WHERE hid = ?",
				cfg.LongHost, cfg.AccessKey, docking)
			return docking, nil
		}
	}

	// 2) 按 access_key 查找已有记录
	var hid int
	err := database.DB.QueryRow(
		"SELECT hid FROM qingka_wangke_huoyuan WHERE pass = ? AND pt = 'longlong' LIMIT 1",
		cfg.AccessKey).Scan(&hid)
	if err == nil && hid > 0 {
		database.DB.Exec("UPDATE qingka_wangke_huoyuan SET url = ? WHERE hid = ?", cfg.LongHost, hid)
		return hid, nil
	}

	// 3) 创建新记录
	result, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_huoyuan (pt, name, url, user, pass, token, ip, cookie, status, addtime, endtime) VALUES ('longlong', '龙龙平台', ?, '', ?, '', '', '', '1', NOW(), '')",
		cfg.LongHost, cfg.AccessKey)
	if err != nil {
		return 0, fmt.Errorf("创建供应商记录失败: %v", err)
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

// ── 产品同步 (替代 long sync) ──

// LonglongSyncOnce 手动触发一次产品同步
func LonglongSyncOnce() (string, error) {
	cfg, err := loadLonglongCfg()
	if err != nil {
		return "", err
	}

	hid, err := ensureLonglongHuoyuan(cfg)
	if err != nil {
		return "", err
	}

	// 使用已有的 ImportSupplierClasses 逻辑同步产品
	supService := NewSupplierService()

	// fd=0 表示全量模式（新增+更新），category=999999 表示全部分类
	category := cfg.Category
	if category == "" {
		category = "999999"
	}

	inserted, updated, msg, err := supService.ImportSupplierClasses(hid, cfg.Rate, category, cfg.NamePrefix, 0)
	if err != nil {
		return "", fmt.Errorf("同步失败: %v", err)
	}

	// 同步状态（自动下架已删除的课程）
	downCount, _, _ := supService.SyncSupplierStatus(hid)

	resultMsg := fmt.Sprintf("同步完成：新增%d，更新%d，下架%d (%s)", inserted, updated, downCount, msg)

	longlongState.mu.Lock()
	longlongState.lastSyncTime = time.Now().Format("2006-01-02 15:04:05")
	longlongState.lastSyncMsg = resultMsg
	longlongState.syncCount++
	longlongState.mu.Unlock()

	return resultMsg, nil
}

// ── 订单监听 (替代 long listen) ──

// longlongListenOnce 执行一次全量订单状态同步
func longlongListenOnce() (int, error) {
	cfg, err := loadLonglongCfg()
	if err != nil {
		return 0, err
	}

	hid, err := ensureLonglongHuoyuan(cfg)
	if err != nil {
		return 0, err
	}

	// 查找所有活跃的龙龙订单
	rows, err := database.DB.Query(
		`SELECT oid, COALESCE(yid,''), COALESCE(user,''), COALESCE(kcname,''), COALESCE(status,'')
		 FROM qingka_wangke_order
		 WHERE hid = ? AND status NOT IN ('已完成','已退款','已退单') AND yid != ''
		 ORDER BY oid DESC LIMIT 500`, hid)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type orderInfo struct {
		oid    int
		yid    string
		user   string
		kcname string
		status string
	}
	var orders []orderInfo
	for rows.Next() {
		var o orderInfo
		rows.Scan(&o.oid, &o.yid, &o.user, &o.kcname, &o.status)
		orders = append(orders, o)
	}

	if len(orders) == 0 {
		return 0, nil
	}

	supService := NewSupplierService()
	sup, err := supService.GetSupplierByHID(hid)
	if err != nil {
		return 0, err
	}

	updated := 0
	for _, o := range orders {
		items, err := supService.QueryOrderProgress(sup, o.yid, o.user, map[string]string{"kcname": o.kcname})
		if err != nil {
			continue
		}

		for _, item := range items {
			statusText := item.Status
			if item.StatusText != "" {
				statusText = item.StatusText
			}
			if statusText == o.status && item.Process == "" {
				continue // 无变化
			}
			database.DB.Exec(
				`UPDATE qingka_wangke_order SET name = ?, yid = ?, status = ?, process = ?, remarks = ?,
				 courseStartTime = ?, courseEndTime = ?, examStartTime = ?, examEndTime = ? WHERE oid = ?`,
				item.KCName, item.YID, statusText, item.Process, item.Remarks,
				item.CourseStartTime, item.CourseEndTime, item.ExamStartTime, item.ExamEndTime,
				o.oid)
			NotifyOrderStatusChange(o.oid, statusText, item.Process, item.Remarks)
			updated++
		}

		// 限速，防止打爆上游
		time.Sleep(200 * time.Millisecond)
	}

	return updated, nil
}

// ── 后台 Goroutine ──

// parseCronInterval 解析定时间隔
func parseLLCronInterval(value, unit string) time.Duration {
	v, _ := strconv.Atoi(value)
	if v <= 0 {
		v = 30
	}
	switch unit {
	case "second":
		return time.Duration(v) * time.Second
	case "hour":
		return time.Duration(v) * time.Hour
	default: // minute
		return time.Duration(v) * time.Minute
	}
}

// StartLonglongDaemon 启动龙龙后台同步&监听守护
func StartLonglongDaemon() {
	log.Println("[LongLong] 内置同步&监听守护启动")
	go longlongSyncLoop()
	go longlongListenLoop()
}

// longlongSyncLoop 产品同步循环（替代 long sync 定时任务）
func longlongSyncLoop() {
	// 等服务启动完毕
	time.Sleep(15 * time.Second)

	longlongState.mu.Lock()
	longlongState.syncRunning = true
	longlongState.mu.Unlock()

	for {
		cfg, err := loadLonglongCfg()
		if err != nil {
			log.Printf("[LongLong Sync] 配置未就绪: %v，1分钟后重试", err)
			time.Sleep(60 * time.Second)
			continue
		}

		// 产品同步间隔：默认用 cron 设置，但至少 5 分钟
		interval := parseLLCronInterval(cfg.CronValue, cfg.CronUnit)
		if interval < 5*time.Minute {
			interval = 5 * time.Minute
		}

		msg, err := LonglongSyncOnce()
		if err != nil {
			log.Printf("[LongLong Sync] 失败: %v", err)
		} else {
			log.Printf("[LongLong Sync] %s", msg)
		}

		time.Sleep(interval)
	}
}

// longlongListenLoop 订单监听循环（替代 long listen）
func longlongListenLoop() {
	// 等服务启动完毕
	time.Sleep(20 * time.Second)

	longlongState.mu.Lock()
	longlongState.listenRunning = true
	longlongState.mu.Unlock()

	for {
		cfg, err := loadLonglongCfg()
		if err != nil {
			time.Sleep(60 * time.Second)
			continue
		}

		// 订单监听间隔：使用 cron 设置，默认 30 秒
		interval := parseLLCronInterval(cfg.CronValue, cfg.CronUnit)
		// listen 走更短的间隔，取 cron 间隔的 1/3 但至少 10 秒
		listenInterval := interval / 3
		if listenInterval < 10*time.Second {
			listenInterval = 10 * time.Second
		}
		if listenInterval > 2*time.Minute {
			listenInterval = 2 * time.Minute
		}

		cnt, err := longlongListenOnce()
		now := time.Now().Format("2006-01-02 15:04:05")

		longlongState.mu.Lock()
		longlongState.lastListenAt = now
		if err != nil {
			longlongState.lastListenMsg = fmt.Sprintf("失败: %v", err)
			log.Printf("[LongLong Listen] %s", longlongState.lastListenMsg)
		} else if cnt > 0 {
			longlongState.lastListenMsg = fmt.Sprintf("更新了 %d 个订单", cnt)
			longlongState.listenCount += cnt
			log.Printf("[LongLong Listen] %s", longlongState.lastListenMsg)
		} else {
			longlongState.lastListenMsg = "无变动"
		}
		longlongState.mu.Unlock()

		time.Sleep(listenInterval)
	}
}
