package service

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ── 自动商品同步状态追踪 ──

var autoSyncState struct {
	mu          sync.RWMutex
	Running     bool   `json:"running"`
	LastRunTime string `json:"last_run_time"`
	LastResult  string `json:"last_result"`
	TotalRuns   int    `json:"total_runs"`
	NextRunTime string `json:"next_run_time"`
}

// AutoSyncStatus 返回自动同步运行状态
func AutoSyncStatus() map[string]interface{} {
	autoSyncState.mu.RLock()
	defer autoSyncState.mu.RUnlock()

	cfg, _ := GetSyncConfig()
	enabled := false
	interval := 30
	if cfg != nil {
		enabled = cfg.AutoSyncEnabled
		interval = cfg.AutoSyncInterval
	}

	return map[string]interface{}{
		"enabled":       enabled,
		"interval":      interval,
		"running":       autoSyncState.Running,
		"last_run_time": autoSyncState.LastRunTime,
		"last_result":   autoSyncState.LastResult,
		"total_runs":    autoSyncState.TotalRuns,
		"next_run_time": autoSyncState.NextRunTime,
	}
}

// SetAutoSyncNextRun 设置下次运行时间（由 main.go 的 goroutine 调用）
func SetAutoSyncNextRun(t time.Time) {
	autoSyncState.mu.Lock()
	autoSyncState.NextRunTime = t.Format("2006-01-02 15:04:05")
	autoSyncState.mu.Unlock()
}

// AutoShelfCron 自动商品同步定时任务（后台 goroutine 调用）
// 从 sync_config 读取配置，逐个货源调用 SyncExecute 统一逻辑
func AutoShelfCron() {
	cfg, err := GetSyncConfig()
	if err != nil || !cfg.AutoSyncEnabled {
		return
	}
	if cfg.SupplierIDs == "" {
		return
	}

	autoSyncState.mu.Lock()
	autoSyncState.Running = true
	autoSyncState.mu.Unlock()

	totalApplied, totalFailed := 0, 0
	parts := strings.Split(cfg.SupplierIDs, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		hid, err := strconv.Atoi(p)
		if err != nil || hid <= 0 {
			continue
		}

		result, err := SyncExecute(hid)
		if err != nil {
			fmt.Printf("[AutoSync] hid=%d 同步失败: %v\n", hid, err)
			totalFailed++
			continue
		}
		totalApplied += result.Applied
		totalFailed += result.Failed
		if result.Applied > 0 || result.Failed > 0 {
			fmt.Printf("[AutoSync] hid=%d 应用%d项，失败%d项\n", hid, result.Applied, result.Failed)
		}
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf("应用%d项，失败%d项", totalApplied, totalFailed)

	autoSyncState.mu.Lock()
	autoSyncState.Running = false
	autoSyncState.LastRunTime = now
	autoSyncState.LastResult = msg
	autoSyncState.TotalRuns++
	autoSyncState.mu.Unlock()
}

// httpPostForm 发送POST表单请求
func httpPostForm(apiURL string, params map[string]string, timeoutSec int) ([]byte, error) {
	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}
	resp, err := client.PostForm(apiURL, form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// toString 安全地将 interface{} 转成字符串
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	default:
		return fmt.Sprintf("%v", val)
	}
}
