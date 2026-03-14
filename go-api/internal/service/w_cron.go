package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"go-api/internal/database"
)

func RunWCron(ctx context.Context) {
	log.Println("[W] 后台批量同步任务启动")
	go wCronBatchSync(ctx)
	go wCronRetryWaitAdd(ctx)
}

// ---------- 1. 批量同步订单状态 (w_redis_ru + w_redis_chu 合一) ----------

func wCronBatchSync(ctx context.Context) {
	if !sleepWithContext(ctx, 2*time.Minute) { // 启动延迟
		return
	}
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[W-cron-sync] panic: %v", r)
				}
			}()

			// 获取所有项目
			appRows, err := database.DB.Query("SELECT id, name, code, org_app_id, url, `key`, uid, token, type FROM w_app WHERE deleted = 0")
			if err != nil {
				log.Printf("[W-cron-sync] 查询项目失败: %v", err)
				return
			}
			defer appRows.Close()

			type appInfo struct {
				ID       int64
				Name     string
				Code     string
				OrgAppID string
				URL      string
				Key      string
				UID      string
				Token    string
				Type     string
			}
			var apps []appInfo
			for appRows.Next() {
				var a appInfo
				appRows.Scan(&a.ID, &a.Name, &a.Code, &a.OrgAppID, &a.URL, &a.Key, &a.UID, &a.Token, &a.Type)
				apps = append(apps, a)
			}
			appRows.Close()

			if len(apps) == 0 {
				return
			}

			// 按 app_id 分组收集活跃订单的 agg_order_id
			appOrders := make(map[int64][]string)
			rows, err := database.DB.Query("SELECT agg_order_id, app_id FROM w_order WHERE status NOT IN ('END', 'REFUND') AND agg_order_id IS NOT NULL AND deleted = 0")
			if err != nil {
				log.Printf("[W-cron-sync] 查询订单失败: %v", err)
				return
			}
			defer rows.Close()

			for rows.Next() {
				var aggOrderID string
				var appID int64
				rows.Scan(&aggOrderID, &appID)
				if aggOrderID != "" && appID > 0 {
					appOrders[appID] = append(appOrders[appID], aggOrderID)
				}
			}
			rows.Close()

			if len(appOrders) == 0 {
				return
			}

			// 构建 app 信息 map
			appMap := make(map[int64]appInfo)
			for _, a := range apps {
				appMap[a.ID] = a
			}

			svc := W()

			// 每个项目批量同步
			for appID, orderIDs := range appOrders {
				app, ok := appMap[appID]
				if !ok {
					continue
				}

				// Jingyu 格式 (type=2) 走单独的同步逻辑
				if app.Type == "2" {
					appRow := map[string]interface{}{
						"url": app.URL, "code": app.Code, "key": app.Key, "uid": app.UID,
					}
					jingyuCronSync(svc, appRow, appID)
					log.Printf("[W-cron-sync] [%s] jingyu同步完成", app.Name)
					continue
				}

				// 每50个一批
				batchSize := 50
				for i := 0; i < len(orderIDs); i += batchSize {
					end := i + batchSize
					if end > len(orderIDs) {
						end = len(orderIDs)
					}
					batch := orderIDs[i:end]
					aidsStr := strings.Join(batch, ",")

					err := wSyncBatch(svc, app, aidsStr)
					if err != nil {
						log.Printf("[W-cron-sync] [%s] 批量同步失败: %v", app.Name, err)
					} else {
						log.Printf("[W-cron-sync] [%s] 批量同步 %d 条", app.Name, len(batch))
					}
					if !sleepWithContext(ctx, time.Second) {
						return
					}
				}
			}
		}()

		if !sleepWithContext(ctx, 5*time.Minute) {
			return
		}
	}
}

// wSyncBatch 批量同步一组订单
func wSyncBatch(svc *WService, app struct {
	ID       int64
	Name     string
	Code     string
	OrgAppID string
	URL      string
	Key      string
	UID      string
	Token    string
	Type     string
}, aidsStr string) error {

	act := "/order/agg_order/view"
	pURL := strings.TrimSpace(app.URL)
	headers := map[string]string{}
	var reqURL string

	if app.Type == "0" {
		reqURL = fmt.Sprintf("%s?act=%s&key=%s&uid=%s&agg_order_id=%s",
			pURL, formatAct(act), app.Key, app.UID, aidsStr)
	} else {
		reqURL = fmt.Sprintf("%s%s?agg_order_id=%s&page_size=50",
			strings.TrimRight(pURL, "/"), act, aidsStr)
		headers["X-WTK"] = app.Token
	}

	result, err := svc.httpReq("GET", reqURL, nil, headers)
	if err != nil {
		return err
	}

	code, _ := result["code"].(float64)
	if int(code) != 0 {
		msg, _ := result["msg"].(string)
		return fmt.Errorf("上游返回错误: %s", msg)
	}

	extData, _ := result["data"].(map[string]interface{})
	if extData == nil {
		return fmt.Errorf("上游返回数据为空")
	}

	dataList, _ := extData["list"].([]interface{})
	if len(dataList) == 0 {
		return nil
	}

	// 逐条更新（简单可靠）
	for _, item := range dataList {
		order, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		svc.syncOrderToDB(order)
	}

	return nil
}

// ---------- 2. 自动重试 WAITADD 订单 ----------

func wCronRetryWaitAdd(ctx context.Context) {
	if !sleepWithContext(ctx, 3*time.Minute) { // 启动延迟
		return
	}
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[W-cron-retry] panic: %v", r)
				}
			}()

			rows, err := database.DB.Query(`SELECT o.id, o.user_id, o.app_id, o.sub_order 
				FROM w_order o WHERE o.status = 'WAITADD' AND o.deleted = 0 ORDER BY o.id ASC LIMIT 20`)
			if err != nil {
				return
			}
			defer rows.Close()

			type waitOrder struct {
				ID       int64
				UserID   int
				AppID    int64
				SubOrder string
			}
			var orders []waitOrder
			for rows.Next() {
				var o waitOrder
				var sub *string
				rows.Scan(&o.ID, &o.UserID, &o.AppID, &sub)
				if sub != nil {
					o.SubOrder = *sub
				}
				orders = append(orders, o)
			}
			rows.Close()

			if len(orders) == 0 {
				return
			}

			svc := W()

			for _, o := range orders {
				if o.SubOrder == "" {
					continue
				}

				app, err := svc.getAppRow(o.AppID)
				if err != nil {
					continue
				}

				// 乐观锁
				res, _ := database.DB.Exec("UPDATE w_order SET status = 'ADDING', updated = NOW() WHERE id = ? AND status = 'WAITADD' LIMIT 1", o.ID)
				if n, _ := res.RowsAffected(); n <= 0 {
					continue
				}

				code := fmt.Sprintf("%v", app["code"])
				act := fmt.Sprintf("/%s/%s_order/add", code, code)

				var postData map[string]interface{}
				json.Unmarshal([]byte(o.SubOrder), &postData)
				if postData == nil {
					database.DB.Exec("UPDATE w_order SET status = 'WAITADD' WHERE id = ? LIMIT 1", o.ID)
					continue
				}
				postData["app_id"] = fmt.Sprintf("%v", app["org_app_id"])

				externalResult, err := svc.appRequest(app, act, postData, "POST")
				if err == nil && externalResult != nil {
					extCode, _ := externalResult["code"].(float64)
					if int(extCode) == 0 {
						extData, _ := externalResult["data"].(map[string]interface{})
						if extData != nil {
							if aggID, ok := extData["agg_order_id"].(string); ok {
								extNum := 0
								if n, ok := extData["num"].(float64); ok {
									extNum = int(n)
								}
								subJSON, _ := json.Marshal(extData["sub_order"])
								database.DB.Exec("UPDATE w_order SET agg_order_id = ?, status = 'NORMAL', num = ?, sub_order = ? WHERE id = ? LIMIT 1",
									aggID, extNum, string(subJSON), o.ID)
								log.Printf("[W-cron-retry] 订单#%d 重新提交成功", o.ID)
								continue
							}
						}
					}
				}

				// 失败，恢复 WAITADD
				database.DB.Exec("UPDATE w_order SET status = 'WAITADD' WHERE id = ? LIMIT 1", o.ID)
				log.Printf("[W-cron-retry] 订单#%d 重新提交失败", o.ID)
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
