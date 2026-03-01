package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"go-api/internal/cache"
	"go-api/internal/database"
)

// StartYDSJCron 启动运动世界后台同步任务（替代4个PHP cron）
func StartYDSJCron() {
	log.Println("[YDSJ] 后台同步任务启动")
	go ydsjCronOrderStatus()  // cron_order.php    — 同步订单状态
	go ydsjCronOrderYID()     // cron_order_yid.php — 同步上游订单ID
	go ydsjCronOrderInfo()    // cron_order_info.php — 同步订单详情
	go ydsjCronOrderRefund()  // cron_order_refund.php — 同步退款状态
}

// ydsjUpstreamQuery 调用上游查询订单接口
func ydsjUpstreamQuery(cfg *YDSJConfig, user string, runType int) ([]map[string]interface{}, error) {
	formData := map[string]string{
		"login_uid": cfg.UID,
		"login_key": cfg.Key,
		"type":      "2",
		"keywords":  user,
		"run_type":  fmt.Sprintf("%d", runType),
	}
	resp, err := httpPostForm(cfg.BaseURL+"/ydsj/api.php?act=orders", formData, 30)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	code := int(mapGetFloat(result, "code"))
	if code != 1 {
		return nil, fmt.Errorf("上游返回 code=%d", code)
	}
	dataRaw, ok := result["data"].([]interface{})
	if !ok || len(dataRaw) == 0 {
		return nil, nil
	}
	var items []map[string]interface{}
	for _, d := range dataRaw {
		if m, ok := d.(map[string]interface{}); ok {
			items = append(items, m)
		}
	}
	return items, nil
}

func ydsjGetProjectName(runType int) string {
	if runType == 0 || runType == 1 {
		return "运动世界"
	}
	return "小步点"
}

// ---------- 1. 订单状态同步 (cron_order.php) ----------

func ydsjCronOrderStatus() {
	time.Sleep(30 * time.Second) // 启动延迟
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[YDSJ-cron-status] panic: %v", r)
				}
			}()

			svc := NewYDSJService()
			cfg, err := svc.GetConfig()
			if err != nil || cfg.BaseURL == "" {
				return
			}

			rows, err := database.DB.Query("SELECT id, uid, user, pass, fees, run_type FROM qingka_wangke_hzw_ydsj WHERE status = 1 ORDER BY id ASC")
			if err != nil {
				log.Printf("[YDSJ-cron-status] 查询失败: %v", err)
				return
			}
			defer rows.Close()

			type orderRow struct {
				ID      int
				UID     int
				User    string
				Pass    string
				Fees    string
				RunType int
			}
			var orders []orderRow
			for rows.Next() {
				var o orderRow
				rows.Scan(&o.ID, &o.UID, &o.User, &o.Pass, &o.Fees, &o.RunType)
				orders = append(orders, o)
			}

			for _, o := range orders {
				items, err := ydsjUpstreamQuery(cfg, o.User, o.RunType)
				if err != nil || len(items) == 0 {
					time.Sleep(3 * time.Second)
					continue
				}

				res := items[0]
				status := int(mapGetFloat(res, "status"))
				remarks := mapGetString(res, "remarks")
				realCost := mapGetFloat(res, "real_fees")

				// 更新状态和备注
				database.DB.Exec("UPDATE qingka_wangke_hzw_ydsj SET status = ?, remarks = ? WHERE id = ?", status, remarks, o.ID)

				now := time.Now().Format("2006-01-02 15:04:05")
				projectName := ydsjGetProjectName(o.RunType)

				var fees float64
				fmt.Sscanf(o.Fees, "%f", &fees)

				if status == 4 {
					// 上游退款 → 退还预扣金额
					if fees > 0 {
						database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", fees, o.UID)
						logContent := fmt.Sprintf("%s 退款：账号%s 源台退款，退还%.2f元", projectName, o.User, fees)
						database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'ydsj_refund', ?, ?, ?)",
							o.UID, fees, logContent, now)
					}
					log.Printf("[YDSJ-cron-status] 订单#%d 源台退款，退还%.2f", o.ID, fees)
				} else if status == 2 {
					// 处理成功 → 计算实际费用差额
					originRealCost := math.Round(realCost*cfg.RealCostMultiple*100) / 100
					difference := originRealCost - fees

					if difference > 0 {
						// 补扣
						absDiff := math.Abs(difference)
						database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", absDiff, o.UID)
						logContent := fmt.Sprintf("%s 下单成功，补扣%.2f元", projectName, absDiff)
						database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'ydsj_add', ?, ?, ?)",
							o.UID, -absDiff, logContent, now)
						log.Printf("[YDSJ-cron-status] 订单#%d 成功，补扣%.2f", o.ID, absDiff)
					} else if difference < 0 {
						// 退回
						absDiff := math.Abs(difference)
						database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", absDiff, o.UID)
						logContent := fmt.Sprintf("%s 下单成功，退回%.2f元", projectName, absDiff)
						database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'ydsj_add', ?, ?, ?)",
							o.UID, absDiff, logContent, now)
						log.Printf("[YDSJ-cron-status] 订单#%d 成功，退回%.2f", o.ID, absDiff)
					}

					database.DB.Exec("UPDATE qingka_wangke_hzw_ydsj SET real_fees = ? WHERE id = ?", fmt.Sprintf("%.2f", originRealCost), o.ID)
				}

				time.Sleep(3 * time.Second)
			}
		}()

		time.Sleep(10 * time.Minute)
	}
}

// ---------- 2. 上游订单ID同步 (cron_order_yid.php) ----------

func ydsjCronOrderYID() {
	time.Sleep(45 * time.Second) // 启动延迟（错开）
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[YDSJ-cron-yid] panic: %v", r)
				}
			}()

			svc := NewYDSJService()
			cfg, err := svc.GetConfig()
			if err != nil || cfg.BaseURL == "" {
				return
			}

			rows, err := database.DB.Query("SELECT id, user, run_type FROM qingka_wangke_hzw_ydsj WHERE yid = ''")
			if err != nil {
				return
			}
			defer rows.Close()

			type row struct {
				ID      int
				User    string
				RunType int
			}
			var orders []row
			for rows.Next() {
				var o row
				rows.Scan(&o.ID, &o.User, &o.RunType)
				orders = append(orders, o)
			}

			for _, o := range orders {
				items, err := ydsjUpstreamQuery(cfg, o.User, o.RunType)
				if err != nil || len(items) == 0 {
					continue
				}
				yid := mapGetString(items[0], "yid")
				if yid != "" {
					database.DB.Exec("UPDATE qingka_wangke_hzw_ydsj SET yid = ? WHERE id = ?", yid, o.ID)
					log.Printf("[YDSJ-cron-yid] 订单#%d 同步yid=%s", o.ID, yid)
				}
			}
		}()

		time.Sleep(3 * time.Minute)
	}
}

// ---------- 3. 订单详情同步 (cron_order_info.php) ----------

func ydsjCronOrderInfo() {
	time.Sleep(60 * time.Second) // 启动延迟
	ctx := context.Background()

	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[YDSJ-cron-info] panic: %v", r)
				}
			}()

			if cache.RDB == nil {
				time.Sleep(time.Minute)
				return
			}

			svc := NewYDSJService()
			cfg, err := svc.GetConfig()
			if err != nil || cfg.BaseURL == "" {
				time.Sleep(time.Minute)
				return
			}

			orderID, err := cache.RDB.LPop(ctx, "ydsj_cron_ids").Result()
			if err != nil || orderID == "" {
				time.Sleep(time.Minute)
				return
			}

			var id int
			var yid, user string
			var runType int
			err = database.DB.QueryRow("SELECT id, yid, user, run_type FROM qingka_wangke_hzw_ydsj WHERE yid = ? LIMIT 1", orderID).
				Scan(&id, &yid, &user, &runType)
			if err != nil || user == "" || yid == "" {
				time.Sleep(time.Second)
				return
			}

			items, err := ydsjUpstreamQuery(cfg, user, runType)
			if err != nil || len(items) == 0 {
				cache.RDB.LPush(ctx, "ydsj_cron_ids", orderID)
				time.Sleep(60 * time.Second)
				return
			}

			// 查找匹配的yid
			var matched map[string]interface{}
			for _, item := range items {
				if mapGetString(item, "yid") == yid {
					matched = item
					break
				}
			}

			if matched == nil {
				cache.RDB.LPush(ctx, "ydsj_cron_ids", orderID)
				time.Sleep(60 * time.Second)
				return
			}

			info := mapGetString(matched, "info")
			if info != "" {
				tmpInfo := mapGetString(matched, "tmp_info")
				isRun := int(mapGetFloat(matched, "is_run"))
				database.DB.Exec("UPDATE qingka_wangke_hzw_ydsj SET is_run = ?, info = ?, tmp_info = ? WHERE id = ?",
					isRun, info, tmpInfo, id)
				log.Printf("[YDSJ-cron-info] 订单#%d 详情已同步", id)
			} else {
				cache.RDB.LPush(ctx, "ydsj_cron_ids", orderID)
				time.Sleep(60 * time.Second)
				return
			}
		}()

		time.Sleep(time.Second)
	}
}

// ---------- 4. 退款状态同步 (cron_order_refund.php) ----------

func ydsjCronOrderRefund() {
	time.Sleep(90 * time.Second) // 启动延迟
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[YDSJ-cron-refund] panic: %v", r)
				}
			}()

			svc := NewYDSJService()
			cfg, err := svc.GetConfig()
			if err != nil || cfg.BaseURL == "" {
				return
			}

			rows, err := database.DB.Query("SELECT id, uid, yid, user, run_type, fees FROM qingka_wangke_hzw_ydsj WHERE status = 5")
			if err != nil {
				return
			}
			defer rows.Close()

			type row struct {
				ID      int
				UID     int
				YID     string
				User    string
				RunType int
				Fees    string
			}
			var orders []row
			for rows.Next() {
				var o row
				rows.Scan(&o.ID, &o.UID, &o.YID, &o.User, &o.RunType, &o.Fees)
				orders = append(orders, o)
			}

			for _, o := range orders {
				items, err := ydsjUpstreamQuery(cfg, o.User, o.RunType)
				if err != nil || len(items) == 0 {
					continue
				}

				// 匹配yid
				var matched map[string]interface{}
				for _, item := range items {
					if mapGetString(item, "yid") == o.YID {
						matched = item
						break
					}
				}
				if matched == nil {
					continue
				}

				status := int(mapGetFloat(matched, "status"))
				refundMoney := mapGetFloat(matched, "refund_money") * cfg.RealCostMultiple

				if status == 4 && refundMoney > 0 {
					refundMoney = math.Round(refundMoney*100) / 100
					database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", refundMoney, o.UID)
					database.DB.Exec("UPDATE qingka_wangke_hzw_ydsj SET status = 4, refund_money = ? WHERE id = ?",
						fmt.Sprintf("%.2f", refundMoney), o.ID)

					now := time.Now().Format("2006-01-02 15:04:05")
					projectName := ydsjGetProjectName(o.RunType)
					logContent := fmt.Sprintf("%s 退款：账号%s 退还%.2f元", projectName, o.User, refundMoney)
					database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'ydsj_refund', ?, ?, ?)",
						o.UID, refundMoney, logContent, now)
					log.Printf("[YDSJ-cron-refund] 订单#%d 退款%.2f", o.ID, refundMoney)
				}
			}
		}()

		time.Sleep(3 * time.Minute)
	}
}
