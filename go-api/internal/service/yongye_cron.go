package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"go-api/internal/database"
)

func RunYongyeCron(ctx context.Context) {
	log.Println("[Yongye] 后台同步任务启动")
	go yongyeCronRetryFailed(ctx)  // 重试失败的订单
	go yongyeCronSyncStudents(ctx) // 同步学生状态
	go yongyeCronRefund(ctx)       // 处理退款
}

// ---------- 1. 重试失败订单 ----------

func yongyeCronRetryFailed(ctx context.Context) {
	if !sleepWithContext(ctx, 60*time.Second) {
		return
	}
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[Yongye-cron-retry] panic: %v", r)
				}
			}()

			svc := Yongye()
			cfg, err := svc.GetConfig()
			if err != nil || cfg.ApiURL == "" || cfg.Token == "" {
				return
			}

			rows, err := database.DB.Query("SELECT id, pol, user, pass, school, type, zkm, ks_h, ks_m, js_h, js_m, weeks, addtime FROM yy_ydsj_dd WHERE dockstatus IN (0, 2) ORDER BY id ASC LIMIT 20")
			if err != nil {
				return
			}
			defer rows.Close()

			type retryRow struct {
				ID      int
				Pol     int
				User    string
				Pass    string
				School  string
				Type    int
				Zkm     float64
				KsH     int
				KsM     int
				JsH     int
				JsM     int
				Weeks   string
				Addtime string
			}
			var orders []retryRow
			for rows.Next() {
				var o retryRow
				rows.Scan(&o.ID, &o.Pol, &o.User, &o.Pass, &o.School, &o.Type, &o.Zkm, &o.KsH, &o.KsM, &o.JsH, &o.JsM, &o.Weeks, &o.Addtime)
				orders = append(orders, o)
			}

			for _, o := range orders {
				apiData := map[string]string{
					"isPolling": fmt.Sprintf("%d", o.Pol),
					"type":      fmt.Sprintf("%d", o.Type),
					"school":    o.School,
					"user":      o.User,
					"pass":      o.Pass,
					"zkm":       fmt.Sprintf("%.2f", o.Zkm),
					"ks_h":      fmt.Sprintf("%d", o.KsH),
					"ks_m":      fmt.Sprintf("%d", o.KsM),
					"js_h":      fmt.Sprintf("%d", o.JsH),
					"js_m":      fmt.Sprintf("%d", o.JsM),
					"weeks":     o.Weeks,
					"addtime":   o.Addtime,
				}

				respBody, err := svc.yongyeUpstreamPost(cfg, "add", apiData)
				if err != nil {
					log.Printf("[Yongye-cron-retry] 订单#%d 重试失败: %v", o.ID, err)
					if !sleepWithContext(ctx, 5*time.Second) {
						return
					}
					continue
				}

				var apiResp map[string]interface{}
				json.Unmarshal(respBody, &apiResp)

				if int(mapGetFloat(apiResp, "code")) == 1 {
					yid := fmt.Sprintf("%v", apiResp["id"])
					database.DB.Exec("UPDATE yy_ydsj_dd SET dockstatus = 1, yid = ? WHERE id = ?", yid, o.ID)
					log.Printf("[Yongye-cron-retry] 订单#%d 重试成功 yid=%s", o.ID, yid)
				} else {
					log.Printf("[Yongye-cron-retry] 订单#%d 重试失败: %s", o.ID, mapGetString(apiResp, "msg"))
				}

				if !sleepWithContext(ctx, 3*time.Second) {
					return
				}
			}
		}()

		if !sleepWithContext(ctx, 5*time.Minute) {
			return
		}
	}
}

// ---------- 2. 同步学生状态 ----------

func yongyeCronSyncStudents(ctx context.Context) {
	if !sleepWithContext(ctx, 90*time.Second) {
		return
	}
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[Yongye-cron-sync] panic: %v", r)
				}
			}()

			svc := Yongye()
			cfg, err := svc.GetConfig()
			if err != nil || cfg.ApiURL == "" || cfg.Token == "" {
				return
			}

			// 获取所有已提交的订单中的学号
			rows, err := database.DB.Query("SELECT DISTINCT user, type, uid FROM yy_ydsj_dd WHERE dockstatus = 1")
			if err != nil {
				return
			}
			defer rows.Close()

			type stuKey struct {
				User string
				Type int
				UID  int
			}
			var keys []stuKey
			for rows.Next() {
				var k stuKey
				rows.Scan(&k.User, &k.Type, &k.UID)
				keys = append(keys, k)
			}

			now := time.Now().Format("2006-01-02 15:04:05")

			for _, k := range keys {
				// 检查本地学生表是否已存在
				var existID int
				database.DB.QueryRow("SELECT id FROM yy_ydsj_student WHERE user = ? AND type = ? LIMIT 1", k.User, k.Type).Scan(&existID)

				if existID == 0 {
					// 从订单中提取信息创建学生记录
					var pass, weeks string
					var zkm float64
					database.DB.QueryRow("SELECT pass, weeks, zkm FROM yy_ydsj_dd WHERE user = ? AND type = ? AND dockstatus = 1 ORDER BY id DESC LIMIT 1",
						k.User, k.Type).Scan(&pass, &weeks, &zkm)

					database.DB.Exec("INSERT INTO yy_ydsj_student (uid, user, pass, type, zkm, weeks, status, last_time) VALUES (?,?,?,?,?,?,0,?)",
						k.UID, k.User, pass, k.Type, zkm, weeks, now)
					log.Printf("[Yongye-cron-sync] 新增学生: %s type=%d", k.User, k.Type)
				} else {
					database.DB.Exec("UPDATE yy_ydsj_student SET last_time = ? WHERE id = ?", now, existID)
				}

				if !sleepWithContext(ctx, time.Second) {
					return
				}
			}
		}()

		if !sleepWithContext(ctx, 3*time.Minute) {
			return
		}
	}
}

// ---------- 3. 处理自动退款 ----------

func yongyeCronRefund(ctx context.Context) {
	if !sleepWithContext(ctx, 120*time.Second) {
		return
	}
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[Yongye-cron-refund] panic: %v", r)
				}
			}()

			svc := Yongye()
			cfg, err := svc.GetConfig()
			if err != nil || cfg.ApiURL == "" || cfg.Token == "" {
				return
			}

			// 查找退单状态的学生
			rows, err := database.DB.Query("SELECT id, uid, user, type, tdmoney FROM yy_ydsj_student WHERE status = 3 AND COALESCE(tdmoney, 0) > 0")
			if err != nil {
				return
			}
			defer rows.Close()

			type refundRow struct {
				ID      int
				UID     int
				User    string
				Type    int
				Tdmoney float64
			}
			var students []refundRow
			for rows.Next() {
				var s refundRow
				rows.Scan(&s.ID, &s.UID, &s.User, &s.Type, &s.Tdmoney)
				students = append(students, s)
			}

			for _, stu := range students {
				tkRate := cfg.Tk
				refund := math.Round(stu.Tdmoney*(1-tkRate)*100) / 100
				if refund <= 0 {
					continue
				}

				// 检查是否已退过款（避免重复）
				var refundedCount int
				database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_moneylog WHERE uid = ? AND type = 'yongye_refund' AND mark LIKE ?",
					stu.UID, fmt.Sprintf("%%学生退单%%账号%s%%", stu.User)).Scan(&refundedCount)
				if refundedCount > 0 {
					continue
				}

				database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", refund, stu.UID)
				now := time.Now().Format("2006-01-02 15:04:05")
				logContent := fmt.Sprintf("永夜运动学生退单：账号%s 退还%.2f", stu.User, refund)
				database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'yongye_refund', ?, ?, ?)",
					stu.UID, refund, logContent, now)
				log.Printf("[Yongye-cron-refund] 学生%s 退款%.2f给UID %d", stu.User, refund, stu.UID)
			}
		}()

		if !sleepWithContext(ctx, 5*time.Minute) {
			return
		}
	}
}
