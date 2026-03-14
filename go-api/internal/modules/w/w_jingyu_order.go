package w

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"go-api/internal/database"
)

// jingyuAddOrder 鲸鱼(jingyu)格式下单。
func (s *WService) jingyuAddOrder(uid int, data map[string]interface{}, app map[string]interface{}, appID int64) (map[string]interface{}, error) {
	var addprice, money float64
	err := database.DB.QueryRow("SELECT addprice, money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice, &money)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	appPrice := 0.0
	if p, ok := app["price"].(string); ok {
		fmt.Sscanf(p, "%f", &appPrice)
	}
	cacType := fmt.Sprintf("%v", app["cac_type"])
	code := fmt.Sprintf("%v", app["code"])

	form, _ := data["form"].(map[string]interface{})
	if form == nil {
		return nil, fmt.Errorf("缺少表单数据")
	}

	account := ""
	password := ""
	school := ""
	dis := 0.0
	num := 0

	switch code {
	case "keep":
		account = fmt.Sprintf("%v", form["phone"])
		password = fmt.Sprintf("%v", form["password"])
		school = fmt.Sprintf("%v", form["zone_name"])
		fmt.Sscanf(fmt.Sprintf("%v", form["dis"]), "%f", &dis)
		fmt.Sscanf(fmt.Sprintf("%v", form["task_list"]), "%d", &num)
		if num <= 0 {
			numF := 0.0
			fmt.Sscanf(fmt.Sprintf("%v", form["task_list"]), "%f", &numF)
			num = int(math.Ceil(numF))
		}
	case "bdlp":
		account = fmt.Sprintf("%v", form["uid"])
		school = fmt.Sprintf("%v", form["school_name"])
		fmt.Sscanf(fmt.Sprintf("%v", form["dis"]), "%f", &dis)
		fmt.Sscanf(fmt.Sprintf("%v", form["task_list"]), "%d", &num)
		if num <= 0 {
			numF := 0.0
			fmt.Sscanf(fmt.Sprintf("%v", form["task_list"]), "%f", &numF)
			num = int(math.Ceil(numF))
		}
	case "ymty":
		account = fmt.Sprintf("%v", form["phone"])
		password = fmt.Sprintf("%v", form["password"])
		school = fmt.Sprintf("%v", form["zone_name"])
		fmt.Sscanf(fmt.Sprintf("%v", form["dis"]), "%f", &dis)
		fmt.Sscanf(fmt.Sprintf("%v", form["task_list"]), "%d", &num)
		if num <= 0 {
			numF := 0.0
			fmt.Sscanf(fmt.Sprintf("%v", form["task_list"]), "%f", &numF)
			num = int(math.Ceil(numF))
		}
	case "yyd":
		account = fmt.Sprintf("%v", form["number"])
		password = fmt.Sprintf("%v", form["password"])
		school = fmt.Sprintf("%v", form["school_name"])
		fmt.Sscanf(fmt.Sprintf("%v", form["dis"]), "%f", &dis)
		fmt.Sscanf(fmt.Sprintf("%v", form["task_list"]), "%d", &num)
		if num <= 0 {
			numF := 0.0
			fmt.Sscanf(fmt.Sprintf("%v", form["task_list"]), "%f", &numF)
			num = int(math.Ceil(numF))
		}
	default:
		account = fmt.Sprintf("%v", form["phone"])
		if account == "<nil>" || account == "" {
			account = fmt.Sprintf("%v", form["number"])
		}
		if account == "<nil>" || account == "" {
			account = fmt.Sprintf("%v", form["uid"])
		}
		password = fmt.Sprintf("%v", form["password"])
		fmt.Sscanf(fmt.Sprintf("%v", form["dis"]), "%f", &dis)
		fmt.Sscanf(fmt.Sprintf("%v", form["task_list"]), "%d", &num)
		if num <= 0 {
			numF := 0.0
			fmt.Sscanf(fmt.Sprintf("%v", form["task_list"]), "%f", &numF)
			num = int(math.Ceil(numF))
		}
	}

	if account == "" || account == "<nil>" || num <= 0 {
		return nil, fmt.Errorf("缺少必填参数")
	}

	danjia := appPrice * addprice
	if danjia <= 0 || addprice < 0.1 {
		return nil, fmt.Errorf("单价异常，请联系管理员")
	}

	var orderMoney float64
	if cacType == "TS" {
		orderMoney = math.Round(float64(num)*danjia*100) / 100
	} else {
		orderMoney = math.Round(float64(num)*dis*danjia*100) / 100
	}

	if code == "ymty" {
		repair := fmt.Sprintf("%v", form["repair"])
		if repair == "1" {
			orderMoney = math.Round(orderMoney*1.5*100) / 100
		}
	}

	if money < orderMoney {
		return nil, fmt.Errorf("余额不足")
	}

	subOrderJSON, _ := json.Marshal(data)
	result, err := database.DB.Exec(
		`INSERT INTO w_order (agg_order_id, user_id, school, account, password, app_id, status, num, cost, pause, sub_order, deleted, created, updated)
		VALUES (NULL, ?, ?, ?, ?, ?, 'ADDING', ?, ?, 0, ?, 0, NOW(), NOW())`,
		uid, school, account, password, appID, num, orderMoney, string(subOrderJSON),
	)
	if err != nil {
		return nil, fmt.Errorf("下单失败请联系管理员")
	}
	orderID, _ := result.LastInsertId()

	res, err := database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ? AND money >= ? LIMIT 1", orderMoney, uid, orderMoney)
	if err != nil || func() bool { n, _ := res.RowsAffected(); return n <= 0 }() {
		database.DB.Exec("DELETE FROM w_order WHERE id = ? LIMIT 1", orderID)
		return nil, fmt.Errorf("余额不足")
	}

	appName := fmt.Sprintf("%v", app["name"])
	wLog(uid, "添加任务", fmt.Sprintf("项目：%s %s %s 扣除 %.2f 元", appName, account, password, orderMoney), -orderMoney)

	act := code + "_add"
	params := map[string]string{}
	for fk, fv := range flattenFormData(form, "form") {
		params[fk] = fv
	}

	upstreamResult, err := s.jingyuRequest(app, act, params)
	if err == nil && upstreamResult != nil {
		upCode := mapGetFloat(upstreamResult, "code")
		if int(upCode) == 1 {
			extData, _ := upstreamResult["data"].(map[string]interface{})
			if extData != nil {
				yid := fmt.Sprintf("%v", extData["id"])
				appOrderIDKey := code + "_order_id"
				appOrderID := fmt.Sprintf("%v", extData[appOrderIDKey])
				subJSON, _ := json.Marshal(extData)
				database.DB.Exec("UPDATE w_order SET agg_order_id = ?, status = 'NORMAL', sub_order = ? WHERE id = ? LIMIT 1",
					yid, string(subJSON), orderID)

				return map[string]interface{}{
					"id":          int(orderID),
					"cost":        orderMoney,
					appOrderIDKey: appOrderID,
				}, nil
			}
			database.DB.Exec("UPDATE w_order SET status = 'NORMAL' WHERE id = ? LIMIT 1", orderID)
		} else {
			database.DB.Exec("UPDATE w_order SET status = 'WAITADD' WHERE id = ? LIMIT 1", orderID)
		}
	} else {
		database.DB.Exec("UPDATE w_order SET status = 'WAITADD' WHERE id = ? LIMIT 1", orderID)
	}

	return map[string]interface{}{
		"id":   int(orderID),
		"cost": orderMoney,
	}, nil
}

// jingyuRefundOrder 鲸鱼(jingyu)格式退款。
func (s *WService) jingyuRefundOrder(uid int, wOrderID int, isAdmin bool, order map[string]interface{}, app map[string]interface{}) (map[string]interface{}, error) {
	code := fmt.Sprintf("%v", app["code"])
	yid := fmt.Sprintf("%v", order["agg_order_id"])
	orderUserID := 0
	fmt.Sscanf(fmt.Sprintf("%v", order["user_id"]), "%d", &orderUserID)

	remainResult, err := s.jingyuRequest(app, "get_remain_count", map[string]string{"id": yid})
	remainNum := 0
	if err == nil && remainResult != nil {
		if rCode := mapGetFloat(remainResult, "code"); int(rCode) == 1 {
			if rData, ok := remainResult["data"].(map[string]interface{}); ok {
				remainNum = int(mapGetFloat(rData, "refund_cnt"))
			}
		}
	}

	refundResult, err := s.jingyuRequest(app, "refund", map[string]string{"id": yid})
	if err != nil {
		database.DB.Exec("UPDATE w_order SET status = 'REFUNDFAIL', updated = NOW() WHERE id = ? LIMIT 1", wOrderID)
		return nil, fmt.Errorf("上游退款失败: %v", err)
	}
	rCode := mapGetFloat(refundResult, "code")
	if int(rCode) != 1 {
		database.DB.Exec("UPDATE w_order SET status = 'REFUNDFAIL', updated = NOW() WHERE id = ? LIMIT 1", wOrderID)
		msg := mapGetString(refundResult, "msg")
		if msg == "" {
			msg = "上游退款失败"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	appPrice := 0.0
	if p, ok := app["price"].(string); ok {
		fmt.Sscanf(p, "%f", &appPrice)
	}
	cacType := fmt.Sprintf("%v", app["cac_type"])

	var orderAddprice float64
	database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", orderUserID).Scan(&orderAddprice)
	danjia := appPrice * orderAddprice

	dis := 1.0
	subOrderStr := fmt.Sprintf("%v", order["sub_order"])
	var subData map[string]interface{}
	json.Unmarshal([]byte(subOrderStr), &subData)
	if subData != nil {
		if formData, ok := subData["form"].(map[string]interface{}); ok {
			if d, ok := formData["dis"]; ok {
				fmt.Sscanf(fmt.Sprintf("%v", d), "%f", &dis)
			}
		} else if d, ok := subData["dis"]; ok {
			fmt.Sscanf(fmt.Sprintf("%v", d), "%f", &dis)
		}
	}

	repairMulti := 1.0
	if code == "ymty" && subData != nil {
		if formData, ok := subData["form"].(map[string]interface{}); ok {
			if fmt.Sprintf("%v", formData["repair"]) == "1" {
				repairMulti = 1.5
			}
		}
	}

	var refundMoney float64
	if cacType == "TS" {
		refundMoney = math.Round(float64(remainNum)*danjia*repairMulti*100) / 100
	} else {
		refundMoney = math.Round(float64(remainNum)*dis*danjia*repairMulti*100) / 100
	}

	orderCost := 0.0
	fmt.Sscanf(fmt.Sprintf("%v", order["cost"]), "%f", &orderCost)
	if refundMoney > orderCost {
		refundMoney = orderCost
	}

	if refundMoney > 0 {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ? LIMIT 1", refundMoney, orderUserID)
	}

	database.DB.Exec("UPDATE w_order SET status = 'REFUND', updated = NOW() WHERE id = ? LIMIT 1", wOrderID)
	wLog(orderUserID, "退款", fmt.Sprintf("订单 %d 退款成功，退款金额 %.2f", wOrderID, refundMoney), refundMoney)

	return map[string]interface{}{
		"refund_amount": refundMoney,
	}, nil
}

// EditOrder 编辑订单。
func (s *WService) EditOrder(uid, orderID int, formData map[string]interface{}, isAdmin bool) (string, error) {
	order, err := s.getOrderRow(orderID)
	if err != nil {
		return "", err
	}

	orderUID := 0
	fmt.Sscanf(fmt.Sprintf("%v", order["user_id"]), "%d", &orderUID)
	if !isAdmin && orderUID != uid {
		return "", fmt.Errorf("您暂无权限")
	}

	appID := int64(0)
	fmt.Sscanf(fmt.Sprintf("%v", order["app_id"]), "%d", &appID)
	app, err := s.getAppRow(appID)
	if err != nil {
		return "", fmt.Errorf("项目不存在")
	}

	code := fmt.Sprintf("%v", app["code"])
	pType := fmt.Sprintf("%v", app["type"])
	yid := fmt.Sprintf("%v", order["agg_order_id"])

	if pType == "2" {
		act := code + "_edit"
		params := map[string]string{"id": yid}
		if formData != nil {
			for fk, fv := range flattenFormData(formData, "form") {
				params[fk] = fv
			}
		}

		result, err := s.jingyuRequest(app, act, params)
		if err != nil {
			return "", err
		}
		if int(mapGetFloat(result, "code")) != 1 {
			return "", fmt.Errorf("编辑失败")
		}

		subStr := fmt.Sprintf("%v", order["sub_order"])
		var sub map[string]interface{}
		json.Unmarshal([]byte(subStr), &sub)
		if sub == nil {
			sub = map[string]interface{}{}
		}
		for k, v := range formData {
			sub[k] = v
		}
		subJSON, _ := json.Marshal(sub)
		database.DB.Exec("UPDATE w_order SET sub_order = ?, updated = NOW() WHERE id = ? LIMIT 1", string(subJSON), orderID)

		account := fmt.Sprintf("%v", order["account"])
		wLog(uid, code+"编辑", fmt.Sprintf("账号 %s 编辑成功", account), 0)
		return "编辑成功", nil
	}

	act := fmt.Sprintf("/%s/%s_order/edit", code, code)
	postData := map[string]interface{}{"id": yid, "form": formData}
	result, err := s.appRequest(app, act, postData, "POST")
	if err != nil {
		return "", err
	}
	extCode, _ := result["code"].(float64)
	if int(extCode) != 0 {
		return "", fmt.Errorf("编辑失败")
	}
	return "编辑成功", nil
}

// ChangeRunStatus 修改运行状态 (暂停/启动)。
func (s *WService) ChangeRunStatus(uid, orderID, status int, formData map[string]interface{}, isAdmin bool) (string, error) {
	order, err := s.getOrderRow(orderID)
	if err != nil {
		return "", err
	}

	orderUID := 0
	fmt.Sscanf(fmt.Sprintf("%v", order["user_id"]), "%d", &orderUID)
	if !isAdmin && orderUID != uid {
		return "", fmt.Errorf("您暂无权限")
	}

	appID := int64(0)
	fmt.Sscanf(fmt.Sprintf("%v", order["app_id"]), "%d", &appID)
	app, err := s.getAppRow(appID)
	if err != nil {
		return "", fmt.Errorf("项目不存在")
	}

	pType := fmt.Sprintf("%v", app["type"])
	yid := fmt.Sprintf("%v", order["agg_order_id"])

	if pType == "2" {
		params := map[string]string{"id": yid}
		if formData != nil {
			for fk, fv := range flattenFormData(formData, "form") {
				params[fk] = fv
			}
		}
		result, err := s.jingyuRequest(app, "change_run_status", params)
		if err != nil {
			return "", err
		}
		if int(mapGetFloat(result, "code")) != 1 {
			return "", fmt.Errorf("修改失败")
		}
		database.DB.Exec("UPDATE w_order SET pause = ?, updated = NOW() WHERE id = ? LIMIT 1", status, orderID)
		return "修改成功", nil
	}

	return "", fmt.Errorf("当前项目类型不支持此操作")
}

// GetRemainCount 获取剩余次数。
func (s *WService) GetRemainCount(uid, orderID int, isAdmin bool) ([]byte, error) {
	order, err := s.getOrderRow(orderID)
	if err != nil {
		return nil, err
	}

	orderUID := 0
	fmt.Sscanf(fmt.Sprintf("%v", order["user_id"]), "%d", &orderUID)
	if !isAdmin && orderUID != uid {
		return nil, fmt.Errorf("您暂无权限")
	}

	appID := int64(0)
	fmt.Sscanf(fmt.Sprintf("%v", order["app_id"]), "%d", &appID)
	app, err := s.getAppRow(appID)
	if err != nil {
		return nil, fmt.Errorf("项目不存在")
	}

	yid := fmt.Sprintf("%v", order["agg_order_id"])
	return s.jingyuRequestRaw(app, "get_remain_count", map[string]string{"id": yid})
}

// GetTaskData 获取任务数据。
func (s *WService) GetTaskData(uid int, orderID int, isAdmin bool) ([]byte, error) {
	order, err := s.getOrderRow(orderID)
	if err != nil {
		return nil, err
	}

	orderUID := 0
	fmt.Sscanf(fmt.Sprintf("%v", order["user_id"]), "%d", &orderUID)
	if !isAdmin && orderUID != uid {
		return nil, fmt.Errorf("您暂无权限")
	}

	appID := int64(0)
	fmt.Sscanf(fmt.Sprintf("%v", order["app_id"]), "%d", &appID)
	app, err := s.getAppRow(appID)
	if err != nil {
		return nil, fmt.Errorf("项目不存在")
	}

	code := fmt.Sprintf("%v", app["code"])
	pType := fmt.Sprintf("%v", app["type"])

	if pType == "2" {
		subStr := fmt.Sprintf("%v", order["sub_order"])
		var sub map[string]interface{}
		json.Unmarshal([]byte(subStr), &sub)
		appOrderIDKey := code + "_order_id"
		appOrderID := ""
		if sub != nil {
			appOrderID = fmt.Sprintf("%v", sub[appOrderIDKey])
		}
		if appOrderID == "" || appOrderID == "<nil>" {
			return nil, fmt.Errorf("缺少子订单ID")
		}
		return s.jingyuRequestRaw(app, "get_task_data", map[string]string{appOrderIDKey: appOrderID})
	}

	return nil, fmt.Errorf("当前项目类型不支持此操作")
}

// EditTask 编辑任务。
func (s *WService) EditTask(uid int, orderID int, formData map[string]interface{}, isAdmin bool) (string, error) {
	order, err := s.getOrderRow(orderID)
	if err != nil {
		return "", err
	}

	orderUID := 0
	fmt.Sscanf(fmt.Sprintf("%v", order["user_id"]), "%d", &orderUID)
	if !isAdmin && orderUID != uid {
		return "", fmt.Errorf("您暂无权限")
	}

	appID := int64(0)
	fmt.Sscanf(fmt.Sprintf("%v", order["app_id"]), "%d", &appID)
	app, err := s.getAppRow(appID)
	if err != nil {
		return "", fmt.Errorf("项目不存在")
	}

	code := fmt.Sprintf("%v", app["code"])
	pType := fmt.Sprintf("%v", app["type"])

	if pType == "2" {
		subStr := fmt.Sprintf("%v", order["sub_order"])
		var sub map[string]interface{}
		json.Unmarshal([]byte(subStr), &sub)
		appOrderIDKey := code + "_order_id"
		appOrderID := ""
		if sub != nil {
			appOrderID = fmt.Sprintf("%v", sub[appOrderIDKey])
		}
		if appOrderID == "" || appOrderID == "<nil>" {
			return "", fmt.Errorf("缺少子订单ID")
		}

		params := map[string]string{appOrderIDKey: appOrderID}
		if formData != nil {
			for fk, fv := range flattenFormData(formData, "form") {
				params[fk] = fv
			}
		}

		result, err := s.jingyuRequest(app, "edit_task", params)
		if err != nil {
			return "", err
		}
		if int(mapGetFloat(result, "code")) != 1 {
			return "", fmt.Errorf("修改失败")
		}
		return "修改成功", nil
	}

	return "", fmt.Errorf("当前项目类型不支持此操作")
}

// DelayTask 延时任务。
func (s *WService) DelayTask(uid int, orderID int, runTaskID string, isAdmin bool) (string, error) {
	order, err := s.getOrderRow(orderID)
	if err != nil {
		return "", err
	}

	orderUID := 0
	fmt.Sscanf(fmt.Sprintf("%v", order["user_id"]), "%d", &orderUID)
	if !isAdmin && orderUID != uid {
		return "", fmt.Errorf("您暂无权限")
	}

	appID := int64(0)
	fmt.Sscanf(fmt.Sprintf("%v", order["app_id"]), "%d", &appID)
	app, err := s.getAppRow(appID)
	if err != nil {
		return "", fmt.Errorf("项目不存在")
	}

	code := fmt.Sprintf("%v", app["code"])
	pType := fmt.Sprintf("%v", app["type"])

	if pType == "2" {
		subStr := fmt.Sprintf("%v", order["sub_order"])
		var sub map[string]interface{}
		json.Unmarshal([]byte(subStr), &sub)
		appOrderIDKey := code + "_order_id"
		appOrderID := ""
		if sub != nil {
			appOrderID = fmt.Sprintf("%v", sub[appOrderIDKey])
		}
		if appOrderID == "" || appOrderID == "<nil>" {
			return "", fmt.Errorf("缺少子订单ID")
		}

		result, err := s.jingyuRequest(app, "delay_task", map[string]string{
			appOrderIDKey: appOrderID,
			"run_task_id": runTaskID,
		})
		if err != nil {
			return "", err
		}
		if int(mapGetFloat(result, "code")) != 1 {
			return "", fmt.Errorf("延时失败")
		}
		return "延时成功", nil
	}

	return "", fmt.Errorf("当前项目类型不支持此操作")
}

// FastDelayTask 快速延时 (所有任务)。
func (s *WService) FastDelayTask(uid, orderID int, isAdmin bool) (string, error) {
	order, err := s.getOrderRow(orderID)
	if err != nil {
		return "", err
	}

	orderUID := 0
	fmt.Sscanf(fmt.Sprintf("%v", order["user_id"]), "%d", &orderUID)
	if !isAdmin && orderUID != uid {
		return "", fmt.Errorf("您暂无权限")
	}

	appID := int64(0)
	fmt.Sscanf(fmt.Sprintf("%v", order["app_id"]), "%d", &appID)
	app, err := s.getAppRow(appID)
	if err != nil {
		return "", fmt.Errorf("项目不存在")
	}

	pType := fmt.Sprintf("%v", app["type"])
	yid := fmt.Sprintf("%v", order["agg_order_id"])

	if pType == "2" {
		result, err := s.jingyuRequest(app, "fast_delay_task", map[string]string{"id": yid})
		if err != nil {
			return "", err
		}
		if int(mapGetFloat(result, "code")) != 1 {
			return "", fmt.Errorf("延时失败")
		}
		database.DB.Exec("UPDATE w_order SET status = 'NORMAL', updated = NOW() WHERE id = ? LIMIT 1", orderID)
		return "延时成功", nil
	}

	return "", fmt.Errorf("当前项目类型不支持此操作")
}

// jingyuCronSync 鲸鱼(jingyu)格式定时同步。
func jingyuCronSync(svc *WService, app map[string]interface{}, appID int64) {
	code := fmt.Sprintf("%v", app["code"])
	page := 1
	pageSize := 100
	totalSynced := 0

	for {
		result, err := svc.jingyuRequest(app, "orders", map[string]string{
			"page":  fmt.Sprintf("%d", page),
			"limit": fmt.Sprintf("%d", pageSize),
		})
		if err != nil {
			break
		}

		rCode := mapGetFloat(result, "code")
		if int(rCode) != 1 {
			break
		}

		dataArr, ok := result["data"].([]interface{})
		if !ok || len(dataArr) == 0 {
			break
		}

		for _, item := range dataArr {
			if m, ok := item.(map[string]interface{}); ok {
				yid := fmt.Sprintf("%v", m["id"])
				statusDisplay := fmt.Sprintf("%v", m["status_display"])
				pause := fmt.Sprintf("%v", m["pause"])

				if yid != "" && yid != "<nil>" {
					var updateParts []string
					var updateArgs []interface{}

					if statusDisplay != "" && statusDisplay != "<nil>" {
						goStatus := mapJingyuStatus(statusDisplay)
						updateParts = append(updateParts, "`status` = ?")
						updateArgs = append(updateArgs, goStatus)
					}
					if pause != "" && pause != "<nil>" {
						pauseInt := 0
						fmt.Sscanf(pause, "%d", &pauseInt)
						updateParts = append(updateParts, "`pause` = ?")
						updateArgs = append(updateArgs, pauseInt)
					}

					if len(updateParts) > 0 {
						updateParts = append(updateParts, "`updated` = NOW()")
						updateArgs = append(updateArgs, yid)
						sql := "UPDATE w_order SET " + strings.Join(updateParts, ", ") + " WHERE agg_order_id = ? LIMIT 1"
						database.DB.Exec(sql, updateArgs...)
						totalSynced++
					}
				}
			}
		}

		pagination, _ := result["pagination"].(map[string]interface{})
		if pagination != nil {
			lastPage := int(mapGetFloat(pagination, "last_page"))
			if page >= lastPage {
				break
			}
		} else if len(dataArr) < pageSize {
			break
		}

		page++
		time.Sleep(500 * time.Millisecond)
	}

	if totalSynced > 0 {
		_ = code
	}
}
