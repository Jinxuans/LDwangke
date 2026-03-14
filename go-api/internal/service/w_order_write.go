package service

import (
	"encoding/json"
	"fmt"
	"math"

	"go-api/internal/database"
)

func (s *WService) AddOrder(uid int, data map[string]interface{}) (map[string]interface{}, error) {
	var addprice, money float64
	err := database.DB.QueryRow("SELECT addprice, money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice, &money)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	appID := int64(0)
	if v, ok := data["app_id"].(float64); ok {
		appID = int64(v)
	}
	school, _ := data["a_school"].(string)
	account, _ := data["a_account"].(string)
	password, _ := data["a_password"].(string)
	dis := 0.0
	if v, ok := data["dis"].(float64); ok {
		dis = v
	}
	taskList, _ := data["task_list"].([]interface{})
	num := len(taskList)

	if appID == 0 || account == "" || num <= 0 || dis <= 0 {
		return nil, fmt.Errorf("缺少必填参数")
	}

	app, err := s.getAppRow(appID)
	if err != nil {
		return nil, fmt.Errorf("项目不存在或已下架")
	}

	if fmt.Sprintf("%v", app["type"]) == "2" {
		if _, hasForm := data["form"]; !hasForm {
			code := fmt.Sprintf("%v", app["code"])
			form := map[string]interface{}{
				"dis":       dis,
				"task_list": num,
			}
			switch code {
			case "bdlp":
				form["uid"] = account
				form["school_name"] = school
			case "yyd":
				form["number"] = account
				form["password"] = password
				form["school_name"] = school
			default:
				form["phone"] = account
				form["password"] = password
				form["zone_name"] = school
			}
			data["form"] = form
		}
		return s.jingyuAddOrder(uid, data, app, appID)
	}

	appPrice := 0.0
	if p, ok := app["price"].(string); ok {
		fmt.Sscanf(p, "%f", &appPrice)
	}
	cacType := fmt.Sprintf("%v", app["cac_type"])
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

	orgAppID := fmt.Sprintf("%v", app["org_app_id"])
	code := fmt.Sprintf("%v", app["code"])
	cpData := make(map[string]interface{}, len(data)+1)
	for k, v := range data {
		cpData[k] = v
	}
	cpData["app_id"] = orgAppID

	act := fmt.Sprintf("/%s/%s_order/add", code, code)
	externalResult, err := s.appRequest(app, act, cpData, "POST")
	if err == nil && externalResult != nil {
		code, _ := externalResult["code"].(float64)
		if int(code) == 0 {
			extData, _ := externalResult["data"].(map[string]interface{})
			if extData != nil {
				if aggID, ok := extData["agg_order_id"].(string); ok {
					extNum := num
					if n, ok := extData["num"].(float64); ok {
						extNum = int(n)
					}
					subJSON, _ := json.Marshal(extData["sub_order"])
					database.DB.Exec("UPDATE w_order SET agg_order_id = ?, status = 'NORMAL', num = ?, sub_order = ? WHERE id = ? LIMIT 1",
						aggID, extNum, string(subJSON), orderID)
				}
			}
		} else {
			database.DB.Exec("UPDATE w_order SET status = 'WAITADD' WHERE id = ? LIMIT 1", orderID)
		}
	} else {
		database.DB.Exec("UPDATE w_order SET status = 'WAITADD' WHERE id = ? LIMIT 1", orderID)
	}

	return map[string]interface{}{"id": int(orderID), "cost": orderMoney}, nil
}

func (s *WService) RefundOrder(uid int, wOrderID int, isAdmin bool) (map[string]interface{}, error) {
	order, err := s.loadActionOrder(uid, wOrderID, isAdmin)
	if err != nil {
		return nil, err
	}

	deletedStr := fmt.Sprintf("%v", order["deleted"])
	if deletedStr == "1" {
		return nil, fmt.Errorf("该订单已删除，无法退款")
	}
	statusStr := fmt.Sprintf("%v", order["status"])
	if statusStr == "REFUND" {
		return nil, fmt.Errorf("该订单已退款，请勿重复操作")
	}

	aggOrderID := fmt.Sprintf("%v", order["agg_order_id"])
	noAggOrderID := aggOrderID == "<nil>" || aggOrderID == ""
	if noAggOrderID && statusStr != "WAITADD" {
		return nil, fmt.Errorf("该订单未提交到外部系统，且不是待下单状态，无法退款")
	}

	res, _ := database.DB.Exec("UPDATE w_order SET status = 'WAITREFUND', updated = NOW() WHERE id = ? AND status = ? LIMIT 1", wOrderID, statusStr)
	if n, _ := res.RowsAffected(); n <= 0 {
		return nil, fmt.Errorf("订单状态已改变，请稍后再试")
	}

	appID := int64(0)
	fmt.Sscanf(fmt.Sprintf("%v", order["app_id"]), "%d", &appID)
	app, err := s.getAppRow(appID)
	if err != nil {
		return nil, fmt.Errorf("项目不存在")
	}

	if fmt.Sprintf("%v", app["type"]) == "2" {
		return s.jingyuRefundOrder(uid, wOrderID, isAdmin, order, app)
	}

	orderUserID := 0
	fmt.Sscanf(fmt.Sprintf("%v", order["user_id"]), "%d", &orderUserID)

	var refundAmount int
	var refundLogPrefix string
	if noAggOrderID {
		fmt.Sscanf(fmt.Sprintf("%v", order["num"]), "%d", &refundAmount)
		refundLogPrefix = "外部无需退款"
		database.DB.Exec("UPDATE w_order SET status = 'REFUND', updated = NOW() WHERE id = ? LIMIT 1", wOrderID)
	} else {
		code := fmt.Sprintf("%v", app["code"])
		act := fmt.Sprintf("/%s/%s_order/refund", code, code)
		postData := map[string]interface{}{"agg_order_id": aggOrderID}

		externalResult, err := s.appRequest(app, act, postData, "POST")
		if err != nil || externalResult == nil {
			database.DB.Exec("UPDATE w_order SET status = 'REFUNDFAIL', updated = NOW() WHERE id = ? LIMIT 1", wOrderID)
			return nil, fmt.Errorf("源台退款失败")
		}
		extCode, _ := externalResult["code"].(float64)
		if int(extCode) != 0 {
			database.DB.Exec("UPDATE w_order SET status = 'REFUNDFAIL', updated = NOW() WHERE id = ? LIMIT 1", wOrderID)
			msg, _ := externalResult["msg"].(string)
			if msg == "" {
				msg = "源台退款失败"
			}
			return nil, fmt.Errorf("%s", msg)
		}

		extData, _ := externalResult["data"].(map[string]interface{})
		if extData != nil {
			if cnt, ok := extData["cnt"].(float64); ok {
				refundAmount = int(cnt)
			}
			if aggOrder, ok := extData["agg_order"].(map[string]interface{}); ok {
				s.syncOrderToDB(aggOrder)
			}
		}
		refundLogPrefix = "外部退款成功"
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
	if d, ok := subData["dis"].(float64); ok && d > 0 {
		dis = d
	}

	orderCost := 0.0
	fmt.Sscanf(fmt.Sprintf("%v", order["cost"]), "%f", &orderCost)
	orderNum := 0
	fmt.Sscanf(fmt.Sprintf("%v", order["num"]), "%d", &orderNum)

	var refundMoney float64
	if cacType == "TS" {
		refundMoney = math.Round(float64(refundAmount)*danjia*100) / 100
	} else {
		refundMoney = math.Round(float64(refundAmount)*dis*danjia*100) / 100
	}
	if orderNum > 0 {
		alt := math.Round(orderCost*float64(refundAmount)/float64(orderNum)*100) / 100
		if alt < refundMoney {
			refundMoney = alt
		}
	}

	if refundMoney > 0 {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ? LIMIT 1", refundMoney, orderUserID)
	}
	wLog(orderUserID, "退款", fmt.Sprintf("订单 %d %s，退款金额 %.2f", wOrderID, refundLogPrefix, refundMoney), refundMoney)

	return map[string]interface{}{"refund_amount": refundMoney}, nil
}

func (s *WService) ResumeOrder(uid, wOrderID int, isAdmin bool) (string, error) {
	order, err := s.loadActionOrder(uid, wOrderID, isAdmin)
	if err != nil {
		return "", err
	}

	if fmt.Sprintf("%v", order["deleted"]) == "1" {
		return "", fmt.Errorf("该订单已删除，无法重新提交")
	}

	res, _ := database.DB.Exec("UPDATE w_order SET status = 'ADDING', updated = NOW() WHERE id = ? AND status = 'WAITADD' LIMIT 1", wOrderID)
	if n, _ := res.RowsAffected(); n <= 0 {
		return "", fmt.Errorf("订单状态已改变，请稍后再试")
	}

	appID := int64(0)
	fmt.Sscanf(fmt.Sprintf("%v", order["app_id"]), "%d", &appID)
	app, err := s.getAppRow(appID)
	if err != nil {
		return "", fmt.Errorf("项目不存在")
	}

	code := fmt.Sprintf("%v", app["code"])
	act := fmt.Sprintf("/%s/%s_order/add", code, code)
	subOrderStr := fmt.Sprintf("%v", order["sub_order"])

	var postData map[string]interface{}
	json.Unmarshal([]byte(subOrderStr), &postData)
	if postData == nil {
		database.DB.Exec("UPDATE w_order SET status = 'WAITADD' WHERE id = ? LIMIT 1", wOrderID)
		return "", fmt.Errorf("缺少原始请求数据，无法重新提交")
	}
	postData["app_id"] = fmt.Sprintf("%v", app["org_app_id"])

	externalResult, err := s.appRequest(app, act, postData, "POST")
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
						aggID, extNum, string(subJSON), wOrderID)
				}
			}
			msg, _ := externalResult["msg"].(string)
			if msg == "" {
				msg = "重新提交成功"
			}
			return msg, nil
		}
	}

	database.DB.Exec("UPDATE w_order SET status = 'WAITADD' WHERE id = ? LIMIT 1", wOrderID)
	msg := "源台下单失败"
	if externalResult != nil {
		if m, ok := externalResult["msg"].(string); ok && m != "" {
			msg = m
		}
	}
	return "", fmt.Errorf("本地下单成功，%s", msg)
}
