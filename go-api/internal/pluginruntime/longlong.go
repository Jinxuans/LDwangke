package pluginruntime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-api/internal/config"
	"go-api/internal/database"
	ordermodule "go-api/internal/modules/order"
	suppliermodule "go-api/internal/modules/supplier"
	obslogger "go-api/internal/observability/logger"
)

const longBinPath = "/usr/bin/long"

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

type longlongCfg struct {
	LongHost      string `json:"long_host"`
	AccessKey     string `json:"access_key"`
	MysqlHost     string `json:"mysql_host"`
	MysqlPort     string `json:"mysql_port"`
	MysqlUser     string `json:"mysql_user"`
	MysqlPassword string `json:"mysql_password"`
	MysqlDatabase string `json:"mysql_database"`
	ClassTable    string `json:"class_table"`
	OrderTable    string `json:"order_table"`
	Docking       string `json:"docking"`
	Rate          string `json:"rate"`
	NamePrefix    string `json:"name_prefix"`
	Category      string `json:"category"`
	CoverPrice    bool   `json:"cover_price"`
	CoverDesc     bool   `json:"cover_desc"`
	CoverName     bool   `json:"cover_name"`
	Sort          string `json:"sort"`
	CronValue     string `json:"cron_value"`
	CronUnit      string `json:"cron_unit"`
}

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
	if cfg.Rate == "" || cfg.Rate == "0" {
		cfg.Rate = "1.5"
	}
	if cfg.CronValue == "" {
		cfg.CronValue = "30"
	}
	if cfg.CronUnit == "" {
		cfg.CronUnit = "minute"
	}
	return &cfg, nil
}

func ensureLonglongHuoyuan(cfg *longlongCfg) (int, error) {
	if docking, _ := strconv.Atoi(cfg.Docking); docking > 0 {
		var count int
		_ = database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_huoyuan WHERE hid = ?", docking).Scan(&count)
		if count > 0 {
			_, _ = database.DB.Exec(
				"UPDATE qingka_wangke_huoyuan SET url = ?, pass = ?, pt = 'longlong', name = '龙龙平台' WHERE hid = ?",
				cfg.LongHost, cfg.AccessKey, docking,
			)
			return docking, nil
		}
	}

	var hid int
	err := database.DB.QueryRow(
		"SELECT hid FROM qingka_wangke_huoyuan WHERE pass = ? AND pt = 'longlong' LIMIT 1",
		cfg.AccessKey,
	).Scan(&hid)
	if err == nil && hid > 0 {
		_, _ = database.DB.Exec("UPDATE qingka_wangke_huoyuan SET url = ? WHERE hid = ?", cfg.LongHost, hid)
		return hid, nil
	}

	result, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_huoyuan (pt, name, url, user, pass, token, ip, cookie, status, addtime, endtime) VALUES ('longlong', '龙龙平台', ?, '', ?, '', '', '', '1', NOW(), '')",
		cfg.LongHost, cfg.AccessKey,
	)
	if err != nil {
		return 0, fmt.Errorf("创建供应商记录失败: %v", err)
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

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
	default:
		return time.Duration(v) * time.Minute
	}
}

func LonglongCheckCLI() map[string]interface{} {
	result := map[string]interface{}{
		"installed": false,
		"path":      "",
		"os":        runtime.GOOS,
	}
	if runtime.GOOS == "windows" {
		result["message"] = "long CLI 仅支持 Linux 服务器，Windows 开发环境不可用"
		return result
	}
	path, err := exec.LookPath("long")
	if err != nil {
		result["message"] = "long CLI 未安装"
		return result
	}
	result["installed"] = true
	result["path"] = path
	result["message"] = "已安装"
	return result
}

func LonglongInstallCLI() (string, error) {
	if runtime.GOOS == "windows" {
		return "", fmt.Errorf("long CLI 仅支持 Linux 服务器，请在生产服务器上操作")
	}
	cfg, err := loadLonglongCfg()
	if err != nil {
		return "", fmt.Errorf("请先保存龙龙配置: %v", err)
	}

	longHost := strings.TrimRight(cfg.LongHost, "/")
	if !strings.HasPrefix(longHost, "http") {
		longHost = "http://" + longHost
	}
	downloadURL := longHost + "/long"
	obslogger.L().Info("LongLong 正在下载 CLI", "url", downloadURL)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("下载失败: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载失败: HTTP %d", resp.StatusCode)
	}

	tmpFile := "/tmp/long_install"
	out, err := os.Create(tmpFile)
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %v", err)
	}
	_, err = io.Copy(out, resp.Body)
	_ = out.Close()
	if err != nil {
		return "", fmt.Errorf("写入失败: %v", err)
	}
	if err := os.Chmod(tmpFile, 0755); err != nil {
		return "", fmt.Errorf("设置权限失败: %v", err)
	}

	cmd := exec.Command("mv", "-f", tmpFile, longBinPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		cmd2 := exec.Command("sudo", "mv", "-f", tmpFile, longBinPath)
		if output2, err2 := cmd2.CombinedOutput(); err2 != nil {
			return "", fmt.Errorf("安装失败: %s / %s", string(output), string(output2))
		}
	}

	obslogger.L().Info("LongLong CLI 安装成功", "path", longBinPath)
	return fmt.Sprintf("安装成功: %s", longBinPath), nil
}

func LonglongSyncOnce() (string, error) {
	cfg, err := loadLonglongCfg()
	if err != nil {
		return "", err
	}
	_, err = ensureLonglongHuoyuan(cfg)
	if err != nil {
		return "", err
	}

	longPath, lookErr := exec.LookPath("long")
	if lookErr != nil {
		return "", fmt.Errorf("long CLI 未安装，请先点击 '一键安装 CLI 工具' 按钮")
	}

	mysqlHost := cfg.MysqlHost
	mysqlPort := cfg.MysqlPort
	mysqlUser := cfg.MysqlUser
	mysqlPass := cfg.MysqlPassword
	mysqlDB := cfg.MysqlDatabase
	if mysqlUser == "" && config.Global != nil {
		db := config.Global.Database
		mysqlHost = db.Host
		mysqlPort = strconv.Itoa(db.Port)
		mysqlUser = db.User
		mysqlPass = db.Password
		mysqlDB = db.DBName
	}
	if mysqlHost == "" {
		mysqlHost = "127.0.0.1"
	}
	if mysqlPort == "" {
		mysqlPort = "3306"
	}
	if mysqlUser == "" {
		return "", fmt.Errorf("MySQL 配置缺失，请在配置中填写或确保 Go API 配置正确")
	}

	longHost := strings.TrimPrefix(strings.TrimPrefix(cfg.LongHost, "http://"), "https://")
	longHost = strings.TrimRight(longHost, "/")
	args := []string{
		"sync",
		"--long-host=" + longHost,
		"--access-key=" + cfg.AccessKey,
		"--mysql-host=" + mysqlHost,
		"--mysql-port=" + mysqlPort,
		"--mysql-user=" + mysqlUser,
		"--mysql-password=" + mysqlPass,
		"--mysql-database=" + mysqlDB,
	}
	if cfg.Docking != "" {
		args = append(args, "--docking="+cfg.Docking)
	}
	if cfg.Rate != "" {
		args = append(args, "--rate="+cfg.Rate)
	}
	if cfg.NamePrefix != "" {
		args = append(args, "--name-prefix="+cfg.NamePrefix)
	}
	if cfg.Category != "" {
		args = append(args, "--category="+cfg.Category)
	}
	if cfg.Sort != "" && cfg.Sort != "0" {
		args = append(args, "--sort="+cfg.Sort)
	}
	if cfg.ClassTable != "" {
		args = append(args, "--class-table="+cfg.ClassTable)
	}
	if cfg.OrderTable != "" {
		args = append(args, "--order-table="+cfg.OrderTable)
	}
	if cfg.CoverPrice {
		args = append(args, "--cover-price=true")
	}
	if cfg.CoverDesc {
		args = append(args, "--cover-desc=true")
	}
	if cfg.CoverName {
		args = append(args, "--cover-name=true")
	}

	obslogger.L().Info("LongLong Sync 执行", "path", longPath, "args", strings.Join(args, " "))
	cmd := exec.Command(longPath, args...)
	cmd.Env = append(os.Environ())
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	var resultMsg string
	if err != nil {
		if outputStr != "" {
			resultMsg = fmt.Sprintf("同步异常: %s", outputStr)
		} else {
			resultMsg = fmt.Sprintf("同步异常: %v", err)
		}
		obslogger.L().Info("LongLong Sync 结果", "message", resultMsg)
	} else if outputStr != "" {
		resultMsg = outputStr
	} else {
		resultMsg = "同步完成"
	}

	longlongState.mu.Lock()
	longlongState.lastSyncTime = time.Now().Format("2006-01-02 15:04:05")
	longlongState.lastSyncMsg = resultMsg
	longlongState.syncCount++
	longlongState.mu.Unlock()

	if err != nil {
		return "", fmt.Errorf("%s", resultMsg)
	}
	return resultMsg, nil
}

func longlongListenOnce() (int, error) {
	cfg, err := loadLonglongCfg()
	if err != nil {
		return 0, err
	}
	hid, err := ensureLonglongHuoyuan(cfg)
	if err != nil {
		return 0, err
	}

	rows, err := database.DB.Query(
		`SELECT oid, COALESCE(yid,''), COALESCE(user,''), COALESCE(kcname,''), COALESCE(status,'')
		 FROM qingka_wangke_order
		 WHERE hid = ? AND status NOT IN ('已完成','已退款','已退单') AND yid != ''
		 ORDER BY oid DESC LIMIT 500`, hid,
	)
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
		_ = rows.Scan(&o.oid, &o.yid, &o.user, &o.kcname, &o.status)
		orders = append(orders, o)
	}
	if len(orders) == 0 {
		return 0, nil
	}

	supService := suppliermodule.SharedService()
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
				continue
			}
			_, _ = database.DB.Exec(
				`UPDATE qingka_wangke_order SET name = ?, yid = ?, status = ?, process = ?, remarks = ?,
				 courseStartTime = ?, courseEndTime = ?, examStartTime = ?, examEndTime = ? WHERE oid = ?`,
				item.KCName, item.YID, statusText, item.Process, item.Remarks,
				item.CourseStartTime, item.CourseEndTime, item.ExamStartTime, item.ExamEndTime, o.oid,
			)
			ordermodule.NotifyOrderStatusChange(o.oid, statusText, item.Process, item.Remarks)
			updated++
		}

		time.Sleep(200 * time.Millisecond)
	}

	return updated, nil
}

func RunLonglongDaemon(ctx context.Context) {
	obslogger.L().Info("LongLong 内置同步与监听守护启动")
	go longlongSyncLoop(ctx)
	go longlongListenLoop(ctx)
}

func longlongSyncLoop(ctx context.Context) {
	if !sleepWithContext(ctx, 15*time.Second) {
		return
	}
	longlongState.mu.Lock()
	longlongState.syncRunning = true
	longlongState.mu.Unlock()
	defer func() {
		longlongState.mu.Lock()
		longlongState.syncRunning = false
		longlongState.mu.Unlock()
	}()

	for {
		cfg, err := loadLonglongCfg()
		if err != nil {
			obslogger.L().Warn("LongLong Sync 配置未就绪", "error", err)
			if !sleepWithContext(ctx, 60*time.Second) {
				return
			}
			continue
		}

		interval := parseLLCronInterval(cfg.CronValue, cfg.CronUnit)
		if interval < 5*time.Minute {
			interval = 5 * time.Minute
		}

		msg, err := LonglongSyncOnce()
		if err != nil {
			obslogger.L().Warn("LongLong Sync 失败", "error", err)
		} else {
			obslogger.L().Info("LongLong Sync 状态", "message", msg)
		}
		if !sleepWithContext(ctx, interval) {
			return
		}
	}
}

func longlongListenLoop(ctx context.Context) {
	if !sleepWithContext(ctx, 20*time.Second) {
		return
	}
	longlongState.mu.Lock()
	longlongState.listenRunning = true
	longlongState.mu.Unlock()
	defer func() {
		longlongState.mu.Lock()
		longlongState.listenRunning = false
		longlongState.mu.Unlock()
	}()

	for {
		cfg, err := loadLonglongCfg()
		if err != nil {
			if !sleepWithContext(ctx, 60*time.Second) {
				return
			}
			continue
		}

		interval := parseLLCronInterval(cfg.CronValue, cfg.CronUnit)
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
			obslogger.L().Info("LongLong Listen 状态", "message", longlongState.lastListenMsg)
		} else if cnt > 0 {
			longlongState.lastListenMsg = fmt.Sprintf("更新了 %d 个订单", cnt)
			longlongState.listenCount += cnt
			obslogger.L().Info("LongLong Listen 状态", "message", longlongState.lastListenMsg)
		} else {
			longlongState.lastListenMsg = "无变动"
		}
		longlongState.mu.Unlock()

		if !sleepWithContext(ctx, listenInterval) {
			return
		}
	}
}
