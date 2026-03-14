package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"go-api/internal/database"
)

func RunXMCron(ctx context.Context) {
	log.Println("[XM] 后台批量同步任务启动")
	go xmCronBatchSync(ctx)
}

// ---------- 批量同步订单状态 ----------

func xmCronBatchSync(ctx context.Context) {
	if !sleepWithContext(ctx, 2*time.Minute+30*time.Second) { // 启动延迟，错开其他 cron
		return
	}
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[XM-cron-sync] panic: %v", r)
				}
			}()

			// 获取所有项目
			projRows, err := database.DB.Query("SELECT id, name, url, `key`, uid, token, type FROM xm_project WHERE is_deleted = 0")
			if err != nil {
				log.Printf("[XM-cron-sync] 查询项目失败: %v", err)
				return
			}
			defer projRows.Close()

			type projInfo struct {
				ID    int
				Name  string
				URL   string
				Key   string
				UID   string
				Token string
				Type  string
			}
			var projects []projInfo
			for projRows.Next() {
				var p projInfo
				projRows.Scan(&p.ID, &p.Name, &p.URL, &p.Key, &p.UID, &p.Token, &p.Type)
				projects = append(projects, p)
			}
			projRows.Close()

			if len(projects) == 0 {
				return
			}

			// 按 project_id 分组收集活跃订单的 y_oid
			projOrders := make(map[int][]int)
			rows, err := database.DB.Query(`SELECT y_oid, project_id FROM xm_order 
				WHERE status NOT IN ('已取消', '已退款', '退款成功', '已完成', '已删除') 
				AND y_oid IS NOT NULL AND y_oid > 0 AND is_deleted = 0`)
			if err != nil {
				log.Printf("[XM-cron-sync] 查询订单失败: %v", err)
				return
			}
			defer rows.Close()

			for rows.Next() {
				var yOid, projectID int
				rows.Scan(&yOid, &projectID)
				if yOid > 0 && projectID > 0 {
					projOrders[projectID] = append(projOrders[projectID], yOid)
				}
			}
			rows.Close()

			if len(projOrders) == 0 {
				return
			}

			// 构建项目 map
			projMap := make(map[int]projInfo)
			for _, p := range projects {
				projMap[p.ID] = p
			}

			svc := XM()

			// 每个项目逐条同步
			for projectID, yOids := range projOrders {
				proj, ok := projMap[projectID]
				if !ok {
					continue
				}

				// 构建完整 project map 供 syncOrderFromUpstream 使用
				projectRow := map[string]interface{}{
					"url":   proj.URL,
					"key":   proj.Key,
					"uid":   proj.UID,
					"token": proj.Token,
					"type":  proj.Type,
				}

				syncCount := 0
				for _, yOid := range yOids {
					_, err := xmSyncSingle(svc, projectRow, yOid)
					if err != nil {
						log.Printf("[XM-cron-sync] [%s] y_oid=%d 同步失败: %v", proj.Name, yOid, err)
					} else {
						syncCount++
					}
					if !sleepWithContext(ctx, 500*time.Millisecond) {
						return
					}
				}

				if syncCount > 0 {
					log.Printf("[XM-cron-sync] [%s] 同步完成 %d/%d 条", proj.Name, syncCount, len(yOids))
				}
			}
		}()

		if !sleepWithContext(ctx, 5*time.Minute) {
			return
		}
	}
}

// xmSyncSingle 同步单个小米运动订单
func xmSyncSingle(svc *XMService, project map[string]interface{}, yOid int) (map[string]interface{}, error) {
	pType := 0
	if t, ok := project["type"].(string); ok {
		fmt.Sscanf(t, "%d", &pType)
	} else if t, ok := project["type"].(int64); ok {
		pType = int(t)
	}

	var externalResult map[string]interface{}
	var err error

	if pType == 0 {
		pURL := fmt.Sprintf("%v", project["url"])
		key := fmt.Sprintf("%v", project["key"])
		pUID := fmt.Sprintf("%v", project["uid"])
		params := url.Values{}
		params.Set("act", "get_orders")
		params.Set("key", key)
		params.Set("uid", pUID)
		params.Set("order_id", fmt.Sprintf("%d", yOid))
		queryURL := pURL + "?" + params.Encode()
		externalResult, err = svc.httpRequest("GET", queryURL, nil, nil)
	} else {
		pURL := strings.TrimRight(fmt.Sprintf("%v", project["url"]), "/")
		token := fmt.Sprintf("%v", project["token"])
		params := url.Values{}
		params.Set("id", fmt.Sprintf("%d", yOid))
		params.Set("page", "1")
		params.Set("page_size", "10")
		queryURL := pURL + "/list?" + params.Encode()
		externalResult, err = svc.httpRequest("GET", queryURL, nil, map[string]string{"token": token})
	}

	if err != nil {
		return nil, err
	}
	code, _ := externalResult["code"].(float64)
	if int(code) != 200 {
		msg, _ := externalResult["msg"].(string)
		return nil, fmt.Errorf("上游返回 code=%d msg=%s", int(code), msg)
	}

	dataList, _ := externalResult["data"].([]interface{})
	if dataList == nil {
		return nil, fmt.Errorf("上游返回数据格式错误")
	}

	skipFields := map[string]bool{"id": true, "user_id": true, "school": true, "account": true, "password": true, "project_id": true, "type": true, "deduction": true}

	for _, item := range dataList {
		row, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var updateParts []string
		var updateArgs []interface{}

		for field, value := range row {
			if skipFields[field] {
				continue
			}
			if field == "status_name" {
				updateParts = append(updateParts, "`status` = ?")
				updateArgs = append(updateArgs, value)
			} else if field == "run_date" {
				jsonVal, _ := json.Marshal(value)
				updateParts = append(updateParts, "`run_date` = ?")
				updateArgs = append(updateArgs, string(jsonVal))
			} else if field == "is_deleted" {
				boolVal := 0
				if v, ok := value.(bool); ok && v {
					boolVal = 1
				}
				updateParts = append(updateParts, "`is_deleted` = ?")
				updateArgs = append(updateArgs, boolVal)
			} else if field == "run_km" {
				if value == nil {
					updateParts = append(updateParts, "`run_km` = NULL")
				} else {
					updateParts = append(updateParts, "`run_km` = ?")
					updateArgs = append(updateArgs, value)
				}
			} else {
				updateParts = append(updateParts, fmt.Sprintf("`%s` = ?", field))
				updateArgs = append(updateArgs, value)
			}
		}

		if len(updateParts) > 0 {
			if updatedAt, ok := row["updated_at"].(string); ok && updatedAt != "" {
				updateParts = append(updateParts, "`updated_at` = ?")
				updateArgs = append(updateArgs, updatedAt)
			} else {
				updateParts = append(updateParts, "`updated_at` = NOW()")
			}
			updateArgs = append(updateArgs, yOid)
			sql := "UPDATE xm_order SET " + strings.Join(updateParts, ", ") + " WHERE y_oid = ? LIMIT 1"
			database.DB.Exec(sql, updateArgs...)
		}
	}

	return externalResult, nil
}
