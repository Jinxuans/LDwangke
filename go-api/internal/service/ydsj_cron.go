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
	go ydsjCronOrderStatus() // cron_order.php    — 同步订单状态
	go ydsjCronOrderYID()    // cron_order_yid.php — 同步上游订单ID
	go ydsjCronOrderInfo()   // cron_order_info.php — 同步订单详情
	go ydsjCronOrderRefund() // cron_order_refund.php — 同步退款状态
}

// ydsjUpstreamQuery 调用上游查询订单接口（新API: POST /order/getOrderInfo）
func ydsjUpstreamQuery(cfg *YDSJConfig, user string, runType int) ([]map[string]interface{}, error) {
	svc := NewYDSJService()
	body := map[string]interface{}{
		"page":    1,
		"size":    10,
		"xh":      user,
		"runType": runType,
		"status":  "",
		"school":  "",
	}
	respBody, err := svc.ydsjRequestWithCfg(cfg, "POST", "/order/getOrderInfo", body)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
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

// ydsjMapUpstreamStatus 将上游状态字符串映射为本地数字状态
// 上游: "下单成功" → 2, "退款成功" → 4, "下单失败" → 3, 其他 → 1
func ydsjMapUpstreamStatus(statusStr string) int {
	switch statusStr {
	case "下单成功", "完成":
		return 2
	case "下单失败", "失败":
		return 3
	case "退款成功", "已退款":
		return 4
	default:
		return 1 // 进行中
	}
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
			if err != nil || cfg.BaseURL == "" || cfg.Token == "" {
				return
			}

			rows, err := database.DB.Query("SELECT id, uid, user, pass, fees, run_type, yid FROM qingka_wangke_hzw_ydsj WHERE status = 1 ORDER BY id ASC")
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
				YID     string
			}
			var orders []orderRow
			for rows.Next() {
				var o orderRow
				rows.Scan(&o.ID, &o.UID, &o.User, &o.Pass, &o.Fees, &o.RunType, &o.YID)
				orders = append(orders, o)
			}

			for _, o := range orders {
				items, err := ydsjUpstreamQuery(cfg, o.User, o.RunType)
				if err != nil || len(items) == 0 {
					time.Sleep(3 * time.Second)
					continue
				}

				// 匹配上游订单（通过 orderid 匹配 yid）
				var res map[string]interface{}
				for _, item := range items {
					oid := fmt.Sprintf("%v", item["orderid"])
					if oid == o.YID {
						res = item
						break
					}
				}
				if res == nil && len(items) > 0 {
					res = items[0] // 回退取第一条
				}

				statusStr := mapGetString(res, "status")
				status := ydsjMapUpstreamStatus(statusStr)
				remarks := mapGetString(res, "bz")
				realCost := mapGetFloat(res, "real_cost")

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
						absDiff := math.Abs(difference)
						database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", absDiff, o.UID)
						logContent := fmt.Sprintf("%s 下单成功，补扣%.2f元", projectName, absDiff)
						database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'ydsj_add', ?, ?, ?)",
							o.UID, -absDiff, logContent, now)
						log.Printf("[YDSJ-cron-status] 订单#%d 成功，补扣%.2f", o.ID, absDiff)
					} else if difference < 0 {
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
			if err != nil || cfg.BaseURL == "" || cfg.Token == "" {
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
				// 新API用 orderid 替代 yid
				oid := fmt.Sprintf("%v", items[0]["orderid"])
				if oid != "" && oid != "<nil>" {
					database.DB.Exec("UPDATE qingka_wangke_hzw_ydsj SET yid = ? WHERE id = ?", oid, o.ID)
					log.Printf("[YDSJ-cron-yid] 订单#%d 同步yid=%s", o.ID, oid)
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
			if err != nil || cfg.BaseURL == "" || cfg.Token == "" {
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

			// 通过 orderid 匹配 yid
			var matched map[string]interface{}
			for _, item := range items {
				oid := fmt.Sprintf("%v", item["orderid"])
				if oid == yid {
					matched = item
					break
				}
			}

			if matched == nil {
				cache.RDB.LPush(ctx, "ydsj_cron_ids", orderID)
				time.Sleep(60 * time.Second)
				return
			}

			// 新API响应中 status 是字符串，用于更新订单状态
			statusStr := mapGetString(matched, "status")
			status := ydsjMapUpstreamStatus(statusStr)
			database.DB.Exec("UPDATE qingka_wangke_hzw_ydsj SET status = ? WHERE id = ? AND status = 1", status, id)
			log.Printf("[YDSJ-cron-info] 订单#%d 状态已同步: %s -> %d", id, statusStr, status)
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
			if err != nil || cfg.BaseURL == "" || cfg.Token == "" {
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

				// 通过 orderid 匹配 yid
				var matched map[string]interface{}
				for _, item := range items {
					oid := fmt.Sprintf("%v", item["orderid"])
					if oid == o.YID {
						matched = item
						break
					}
				}
				if matched == nil {
					continue
				}

				statusStr := mapGetString(matched, "status")
				status := ydsjMapUpstreamStatus(statusStr)

				if status == 4 {
					// 上游退款，退还预扣金额
					var fees float64
					fmt.Sscanf(o.Fees, "%f", &fees)
					refundMoney := math.Round(fees*100) / 100

					if refundMoney > 0 {
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
			}
		}()

		time.Sleep(3 * time.Minute)
	}
}
