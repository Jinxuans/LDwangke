package pluginruntime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
	obslogger "go-api/internal/observability/logger"
)

var simpleThreadState struct {
	mu          sync.RWMutex
	running     bool
	lastRunTime string
	lastMsg     string
	syncCount   int
}

type hostRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rateBucket
}

type rateBucket struct {
	lastTime time.Time
	mu       sync.Mutex
}

var simpleRateLimiter = &hostRateLimiter{
	limiters: make(map[string]*rateBucket),
}

func (rl *hostRateLimiter) wait(host string, interval time.Duration) {
	rl.mu.Lock()
	bucket, ok := rl.limiters[host]
	if !ok {
		bucket = &rateBucket{}
		rl.limiters[host] = bucket
	}
	rl.mu.Unlock()

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(bucket.lastTime)
	if elapsed < interval {
		time.Sleep(interval - elapsed)
	}
	bucket.lastTime = time.Now()
}

func simpleExtractHost(rawURL string) string {
	rawURL = strings.TrimRight(rawURL, "/")
	if !strings.HasPrefix(rawURL, "http") {
		rawURL = "http://" + rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return u.Host
}

func simpleBuildBaseURL(sup *model.SupplierFull) string {
	baseURL := strings.TrimRight(sup.URL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	return baseURL
}

func simpleGetToken(sup *model.SupplierFull) string {
	if sup.Token != "" {
		return sup.Token
	}
	return sup.Pass
}

func valueToString(v interface{}) string {
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

func simpleThreadOnce() (int, error) {
	rows, err := database.DB.Query(
		"SELECT hid, COALESCE(url,''), COALESCE(`user`,''), COALESCE(pass,''), COALESCE(token,'') FROM qingka_wangke_huoyuan WHERE pt = 'simple' AND status = '1'")
	if err != nil {
		return 0, fmt.Errorf("查询供应商失败: %v", err)
	}
	defer rows.Close()

	type supInfo struct {
		hid   int
		url   string
		user  string
		pass  string
		token string
	}
	var suppliers []supInfo
	for rows.Next() {
		var s supInfo
		_ = rows.Scan(&s.hid, &s.url, &s.user, &s.pass, &s.token)
		suppliers = append(suppliers, s)
	}

	if len(suppliers) == 0 {
		return 0, nil
	}

	totalUpdated := 0
	for _, s := range suppliers {
		sup := &model.SupplierFull{
			HID:   s.hid,
			PT:    "simple",
			URL:   s.url,
			User:  s.user,
			Pass:  s.pass,
			Token: s.token,
		}
		updated, err := simpleThreadSyncSupplier(sup)
		if err != nil {
			obslogger.L().Warn("Simple-Thread 同步失败", "hid", s.hid, "error", err)
			continue
		}
		totalUpdated += updated
	}

	return totalUpdated, nil
}

func simpleThreadSyncSupplier(sup *model.SupplierFull) (int, error) {
	baseURL := simpleBuildBaseURL(sup)
	token := simpleGetToken(sup)

	timestampKey := fmt.Sprintf("simple_thread_ts_%d", sup.HID)
	var lastTimestamp string
	_ = database.DB.QueryRow("SELECT `k` FROM qingka_wangke_config WHERE `v` = ?", timestampKey).Scan(&lastTimestamp)
	if lastTimestamp == "" {
		lastTimestamp = "2000-01-01"
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	if lastTimestamp == "2000-01-01" {
		_, _ = database.DB.Exec("INSERT INTO qingka_wangke_config (`v`, `k`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `k` = ?",
			timestampKey, now, now)
	} else {
		_, _ = database.DB.Exec("UPDATE qingka_wangke_config SET `k` = ? WHERE `v` = ?", now, timestampKey)
	}

	updated := 0
	page := 1

	for {
		formData := url.Values{}
		formData.Set("page", fmt.Sprintf("%d", page))
		formData.Set("token", token)
		formData.Set("timestamp", lastTimestamp)

		apiURL := baseURL + "/Api/Thread"
		if host := simpleExtractHost(sup.URL); host != "" {
			simpleRateLimiter.wait(host, 500*time.Millisecond)
		}

		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.PostForm(apiURL, formData)
		if err != nil {
			return updated, fmt.Errorf("请求Thread失败: %v", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return updated, fmt.Errorf("读取响应失败: %v", err)
		}

		var raw map[string]interface{}
		if err := json.Unmarshal(body, &raw); err != nil {
			return updated, fmt.Errorf("解析响应失败: %s", string(body))
		}

		dataArr, ok := raw["data"].([]interface{})
		if !ok || len(dataArr) == 0 {
			break
		}

		for _, item := range dataArr {
			row, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			status := valueToString(row["status"])
			switch status {
			case "已暂停", "明天继续", "待处理", "考试中", "平时分":
				status = "进行中"
			}

			process := valueToString(row["process"])
			if process == "" {
				process = "暂无详情"
			} else {
				process += " | 更新时间：" + time.Now().Format("2006-01-02 15:04:05")
			}

			progress := valueToString(row["progress"])
			if progress != "" && !strings.HasSuffix(progress, "%") {
				progress += "%"
			}
			if progress == "" {
				progress = "0%"
			}

			yid := valueToString(row["id"])
			user := valueToString(row["user"])
			pass := valueToString(row["pass"])
			course := valueToString(row["course"])
			cid := valueToString(row["cid"])
			kcks := valueToString(row["kcks"])
			kcjs := valueToString(row["kcjs"])
			ksks := valueToString(row["ksks"])
			ksjs := valueToString(row["ksjs"])

			result, err := database.DB.Exec(
				"UPDATE qingka_wangke_order SET `yid` = ?, `status` = ?, `process` = ?, `remarks` = ?, "+
					"`courseStartTime` = ?, `courseEndTime` = ?, `examStartTime` = ?, `examEndTime` = ? "+
					"WHERE `user` = ? AND `pass` = ? AND `kcname` = ? AND LOCATE(?, noun) > 0",
				yid, status, progress, process, kcks, kcjs, ksks, ksjs, user, pass, course, cid,
			)
			if err != nil {
				obslogger.L().Warn("Simple-Thread 更新失败", "user", user, "course", course, "error", err)
				continue
			}
			if affected, _ := result.RowsAffected(); affected > 0 {
				updated += int(affected)
			}
		}

		obslogger.L().Info("Simple-Thread 已同步分页数据", "hid", sup.HID, "page", page, "count", len(dataArr))
		if len(dataArr) < 1000 {
			break
		}
		page++
	}

	return updated, nil
}

func RunSimpleThreadSync(ctx context.Context) {
	obslogger.L().Info("Simple-Thread 同步守护启动")
	if !sleepWithContext(ctx, 25*time.Second) {
		return
	}

	simpleThreadState.mu.Lock()
	simpleThreadState.running = true
	simpleThreadState.mu.Unlock()
	defer func() {
		simpleThreadState.mu.Lock()
		simpleThreadState.running = false
		simpleThreadState.mu.Unlock()
	}()

	for {
		cnt, err := simpleThreadOnce()
		now := time.Now().Format("2006-01-02 15:04:05")

		simpleThreadState.mu.Lock()
		simpleThreadState.lastRunTime = now
		if err != nil {
			simpleThreadState.lastMsg = fmt.Sprintf("失败: %v", err)
			obslogger.L().Info("Simple-Thread 状态", "message", simpleThreadState.lastMsg)
		} else if cnt > 0 {
			simpleThreadState.lastMsg = fmt.Sprintf("更新了 %d 个订单", cnt)
			simpleThreadState.syncCount += cnt
			obslogger.L().Info("Simple-Thread 状态", "message", simpleThreadState.lastMsg)
		} else {
			simpleThreadState.lastMsg = "无变动"
		}
		simpleThreadState.mu.Unlock()

		if !sleepWithContext(ctx, 2*time.Minute) {
			return
		}
	}
}
