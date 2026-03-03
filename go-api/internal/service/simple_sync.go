package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

// ── 至强 (simple) 平台批量同步（替代 PHP thread.php） ──
// 使用 /Api/Thread 接口按 timestamp 增量拉取订单状态

// simpleThreadState 全局运行状态
var simpleThreadState struct {
	mu          sync.RWMutex
	running     bool
	lastRunTime string
	lastMsg     string
	syncCount   int
}

// SimpleThreadStatus 返回当前 simple 同步状态
func SimpleThreadStatus() map[string]interface{} {
	simpleThreadState.mu.RLock()
	defer simpleThreadState.mu.RUnlock()
	return map[string]interface{}{
		"running":       simpleThreadState.running,
		"last_run_time": simpleThreadState.lastRunTime,
		"last_msg":      simpleThreadState.lastMsg,
		"sync_count":    simpleThreadState.syncCount,
	}
}

// simpleThreadOnce 执行一次 Thread 批量同步
// 遍历所有 pt="simple" 的供应商，调用 /Api/Thread 拉取进度并更新本地订单
func simpleThreadOnce() (int, error) {
	// 查找所有 simple 类型供应商
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
		rows.Scan(&s.hid, &s.url, &s.user, &s.pass, &s.token)
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
			log.Printf("[Simple-Thread] hid=%d 同步失败: %v", s.hid, err)
			continue
		}
		totalUpdated += updated
	}

	return totalUpdated, nil
}

// simpleThreadSyncSupplier 对单个 simple 供应商执行 Thread 同步
func simpleThreadSyncSupplier(sup *model.SupplierFull) (int, error) {
	baseURL := simpleBuildBaseURL(sup)
	token := simpleGetToken(sup)

	// 读取上次同步时间戳
	timestampKey := fmt.Sprintf("simple_thread_ts_%d", sup.HID)
	var lastTimestamp string
	database.DB.QueryRow("SELECT `k` FROM qingka_wangke_config WHERE `v` = ?", timestampKey).Scan(&lastTimestamp)
	if lastTimestamp == "" {
		lastTimestamp = "2000-01-01"
	}

	// 保存本次同步时间
	now := time.Now().Format("2006-01-02 15:04:05")
	if lastTimestamp == "2000-01-01" {
		database.DB.Exec("INSERT INTO qingka_wangke_config (`v`, `k`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `k` = ?",
			timestampKey, now, now)
	} else {
		database.DB.Exec("UPDATE qingka_wangke_config SET `k` = ? WHERE `v` = ?", now, timestampKey)
	}

	updated := 0
	page := 1

	for {
		formData := url.Values{}
		formData.Set("page", fmt.Sprintf("%d", page))
		formData.Set("token", token)
		formData.Set("timestamp", lastTimestamp)

		apiURL := baseURL + "/Api/Thread"

		if host := extractHost(sup.URL); host != "" {
			globalRateLimiter.wait(host, 500*time.Millisecond)
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

			// 映射上游状态（按 PHP thread.php 逻辑）
			status := toString(row["status"])
			switch status {
			case "已暂停", "明天继续", "待处理", "考试中", "平时分":
				status = "进行中"
			}

			process := toString(row["process"])
			if process == "" {
				process = "暂无详情"
			} else {
				process += " | 更新时间：" + time.Now().Format("2006-01-02 15:04:05")
			}

			progress := toString(row["progress"])
			if progress != "" && !strings.HasSuffix(progress, "%") {
				progress += "%"
			}
			if progress == "" {
				progress = "0%"
			}

			yid := toString(row["id"])
			user := toString(row["user"])
			pass := toString(row["pass"])
			course := toString(row["course"])
			cid := toString(row["cid"])
			kcks := toString(row["kcks"])
			kcjs := toString(row["kcjs"])
			ksks := toString(row["ksks"])
			ksjs := toString(row["ksjs"])

			// 按 PHP thread.php 逻辑：匹配 user + pass + kcname + noun LIKE cid
			result, err := database.DB.Exec(
				"UPDATE qingka_wangke_order SET `yid` = ?, `status` = ?, `process` = ?, `remarks` = ?, "+
					"`courseStartTime` = ?, `courseEndTime` = ?, `examStartTime` = ?, `examEndTime` = ? "+
					"WHERE `user` = ? AND `pass` = ? AND `kcname` = ? AND LOCATE(?, noun) > 0",
				yid, status, progress, process,
				kcks, kcjs, ksks, ksjs,
				user, pass, course, cid,
			)
			if err != nil {
				log.Printf("[Simple-Thread] 更新失败 user=%s course=%s: %v", user, course, err)
				continue
			}
			affected, _ := result.RowsAffected()
			if affected > 0 {
				updated += int(affected)
			}
		}

		log.Printf("[Simple-Thread] hid=%d page=%d 已同步 %d 条", sup.HID, page, len(dataArr))

		if len(dataArr) < 1000 {
			break
		}
		page++
	}

	return updated, nil
}

// StartSimpleThreadSync 启动 simple 平台后台 Thread 同步守护
func StartSimpleThreadSync() {
	log.Println("[Simple-Thread] 至强平台 Thread 同步守护启动")
	go simpleThreadLoop()
}

// simpleThreadLoop Thread 同步循环
func simpleThreadLoop() {
	// 等服务启动完毕
	time.Sleep(25 * time.Second)

	simpleThreadState.mu.Lock()
	simpleThreadState.running = true
	simpleThreadState.mu.Unlock()

	for {
		cnt, err := simpleThreadOnce()
		now := time.Now().Format("2006-01-02 15:04:05")

		simpleThreadState.mu.Lock()
		simpleThreadState.lastRunTime = now
		if err != nil {
			simpleThreadState.lastMsg = fmt.Sprintf("失败: %v", err)
			log.Printf("[Simple-Thread] %s", simpleThreadState.lastMsg)
		} else if cnt > 0 {
			simpleThreadState.lastMsg = fmt.Sprintf("更新了 %d 个订单", cnt)
			simpleThreadState.syncCount += cnt
			log.Printf("[Simple-Thread] %s", simpleThreadState.lastMsg)
		} else {
			simpleThreadState.lastMsg = "无变动"
		}
		simpleThreadState.mu.Unlock()

		// 同步间隔：2 分钟
		time.Sleep(2 * time.Minute)
	}
}
