package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"go-api/internal/database"
	obslogger "go-api/internal/observability/logger"
)

func RunYDSJCron(ctx context.Context) {
	obslogger.L().Info("YDSJ 后台同步任务启动")
	go ydsjCronOrderStatus(ctx) // cron_order.php    — 同步订单状态
	go ydsjCronOrderYID(ctx)    // cron_order_yid.php — 同步上游订单ID
	go ydsjCronOrderInfo(ctx)   // cron_order_info.php — 同步订单详情
	go ydsjCronOrderRefund(ctx) // cron_order_refund.php — 同步退款状态
}

// ydsjUpstreamQuery 调用上游查询订单接口（LearnExp: POST /ydsj/api.php?act=orders）
func ydsjUpstreamQuery(cfg *YDSJConfig, user string, runType int) ([]map[string]interface{}, error) {
	svc := YDSJ()
	params := map[string]string{
		"type":     "2",
		"keywords": user,
		"run_type": fmt.Sprintf("%d", runType),
	}
	respBody, err := svc.ydsjRequestWithCfg(cfg, "orders", params)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	code := mapGetFloat(result, "code")
	if code != 1 {
		return nil, fmt.Errorf("上游返回错误: %s", mapGetString(result, "msg"))
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

func ydsjCronOrderStatus(ctx context.Context) {
	if !sleepWithContext(ctx, 30*time.Second) { // 启动延迟
		return
	}
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					obslogger.L().Error("YDSJ cron status panic", "panic", r)
				}
			}()

			svc := YDSJ()
			cfg, err := svc.GetConfig()
			if err != nil || !ydsjIsConfigured(cfg) {
				return
			}

			rows, err := database.DB.Query("SELECT id, uid, `user`, pass, fees, run_type, yid FROM qingka_wangke_hzw_ydsj WHERE status = 1 ORDER BY id ASC")
			if err != nil {
				obslogger.L().Warn("YDSJ cron status 查询失败", "error", err)
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
					if !sleepWithContext(ctx, 3*time.Second) {
						return
					}
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
					obslogger.L().Info("YDSJ cron status 源台退款", "order_id", o.ID, "refund", fees)
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
						obslogger.L().Info("YDSJ cron status 成功补扣", "order_id", o.ID, "amount", absDiff)
					} else if difference < 0 {
						absDiff := math.Abs(difference)
						database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", absDiff, o.UID)
						logContent := fmt.Sprintf("%s 下单成功，退回%.2f元", projectName, absDiff)
						database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'ydsj_add', ?, ?, ?)",
							o.UID, absDiff, logContent, now)
						obslogger.L().Info("YDSJ cron status 成功退回", "order_id", o.ID, "amount", absDiff)
					}

					database.DB.Exec("UPDATE qingka_wangke_hzw_ydsj SET real_fees = ? WHERE id = ?", fmt.Sprintf("%.2f", originRealCost), o.ID)
				}

				if !sleepWithContext(ctx, 3*time.Second) {
					return
				}
			}
		}()

		if !sleepWithContext(ctx, 10*time.Minute) {
			return
		}
	}
}

// ---------- 2. 上游订单ID同步 (cron_order_yid.php) ----------

func ydsjCronOrderYID(ctx context.Context) {
	if !sleepWithContext(ctx, 45*time.Second) { // 启动延迟（错开）
		return
	}
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					obslogger.L().Error("YDSJ cron yid panic", "panic", r)
				}
			}()

			svc := YDSJ()
			cfg, err := svc.GetConfig()
			if err != nil || !ydsjIsConfigured(cfg) {
				return
			}

			rows, err := database.DB.Query("SELECT id, `user`, run_type FROM qingka_wangke_hzw_ydsj WHERE yid = ''")
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
					obslogger.L().Info("YDSJ cron yid 同步完成", "order_id", o.ID, "yid", oid)
				}
			}
		}()

		if !sleepWithContext(ctx, 3*time.Minute) {
			return
		}
	}
}

// ---------- 3. 订单详情同步 (cron_order_info.php) ----------

func ydsjCronOrderInfo(ctx context.Context) {
	if !sleepWithContext(ctx, 60*time.Second) { // 启动延迟
		return
	}

	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					obslogger.L().Error("YDSJ cron info panic", "panic", r)
				}
			}()

			svc := YDSJ()
			cfg, err := svc.GetConfig()
			if err != nil || !ydsjIsConfigured(cfg) {
				return
			}

			rows, err := database.DB.Query("SELECT id, yid, `user`, run_type FROM qingka_wangke_hzw_ydsj WHERE status = 1 AND yid <> '' ORDER BY id ASC LIMIT 30")
			if err != nil {
				return
			}
			defer rows.Close()

			type row struct {
				ID      int
				YID     string
				User    string
				RunType int
			}
			var orders []row
			for rows.Next() {
				var o row
				rows.Scan(&o.ID, &o.YID, &o.User, &o.RunType)
				orders = append(orders, o)
			}

			for _, o := range orders {
				items, err := ydsjUpstreamQuery(cfg, o.User, o.RunType)
				if err != nil || len(items) == 0 {
					if !sleepWithContext(ctx, time.Second) {
						return
					}
					continue
				}

				// 通过 orderid 精确匹配 yid，不再依赖 Redis 列表串联。
				var matched map[string]interface{}
				for _, item := range items {
					oid := fmt.Sprintf("%v", item["orderid"])
					if oid == o.YID {
						matched = item
						break
					}
				}

				if matched == nil {
					if !sleepWithContext(ctx, time.Second) {
						return
					}
					continue
				}

				// 新API响应中 status 是字符串，用于更新订单状态
				statusStr := mapGetString(matched, "status")
				status := ydsjMapUpstreamStatus(statusStr)
				database.DB.Exec("UPDATE qingka_wangke_hzw_ydsj SET status = ? WHERE id = ? AND status = 1", status, o.ID)
				obslogger.L().Info("YDSJ cron info 状态已同步", "order_id", o.ID, "status_text", statusStr, "status", status)

				if !sleepWithContext(ctx, time.Second) {
					return
				}
			}
		}()

		if !sleepWithContext(ctx, 30*time.Second) {
			return
		}
	}
}

// ---------- 4. 退款状态同步 (cron_order_refund.php) ----------

func ydsjCronOrderRefund(ctx context.Context) {
	if !sleepWithContext(ctx, 90*time.Second) { // 启动延迟
		return
	}
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					obslogger.L().Error("YDSJ cron refund panic", "panic", r)
				}
			}()

			svc := YDSJ()
			cfg, err := svc.GetConfig()
			if err != nil || !ydsjIsConfigured(cfg) {
				return
			}

			rows, err := database.DB.Query("SELECT id, uid, yid, `user`, run_type, fees FROM qingka_wangke_hzw_ydsj WHERE status = 5")
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
						obslogger.L().Info("YDSJ cron refund 完成", "order_id", o.ID, "refund", refundMoney)
					}
				}
			}
		}()

		if !sleepWithContext(ctx, 3*time.Minute) {
			return
		}
	}
}
