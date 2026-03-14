package service

import (
	"fmt"
	"log"
	"math"
	"time"

	"go-api/internal/database"
)

func (s *SDXYService) ListOrders(uid int, isAdmin bool, page, limit int, searchType, keyword, statusFilter string) ([]SDXYOrder, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}

	if !isAdmin {
		where += " AND uid = ?"
		args = append(args, uid)
	}

	if keyword != "" {
		switch searchType {
		case "1":
			where += " AND id = ?"
			args = append(args, keyword)
		case "2":
			where += " AND user LIKE ?"
			args = append(args, "%"+keyword+"%")
		case "3":
			where += " AND pass LIKE ?"
			args = append(args, "%"+keyword+"%")
		case "4":
			if isAdmin {
				where += " AND uid = ?"
				args = append(args, keyword)
			}
		}
	}

	if statusFilter != "" {
		where += " AND status = ?"
		args = append(args, statusFilter)
	}

	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_flash_sdxy "+where, args...).Scan(&total)

	offset := (page - 1) * limit
	query := `SELECT id, uid, agg_order_id, sdxy_order_id, user, pass, school, num, distance,
		run_type, run_rule, pause, status, fees, created_at
		FROM qingka_wangke_flash_sdxy ` + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []SDXYOrder
	for rows.Next() {
		var o SDXYOrder
		rows.Scan(&o.ID, &o.UID, &o.AggOrderID, &o.SDXYOrderID, &o.User, &o.Pass,
			&o.School, &o.Num, &o.Distance, &o.RunType, &o.RunRule,
			&o.Pause, &o.Status, &o.Fees, &o.CreatedAt)
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []SDXYOrder{}
	}
	return orders, total, nil
}

func (s *SDXYService) AddOrder(uid int, form map[string]interface{}) (string, error) {
	phone := mapGetString(form, "phone")
	password := mapGetString(form, "password")
	distance := mapGetString(form, "dis")
	zoneId := mapGetString(form, "zone_id")
	zoneName := mapGetString(form, "zone_name")
	runType := mapGetString(form, "run_type")
	studentId := mapGetString(form, "student_id")
	runRuleId := mapGetString(form, "run_rule_id")
	taskList, _ := form["task_list"].([]interface{})

	if phone == "" {
		return "", fmt.Errorf("手机号不能为空")
	}
	if distance == "" || (zoneId == "" && zoneName == "") || runType == "" || studentId == "" || runRuleId == "" {
		log.Printf("[SDXY-AddOrder] 字段缺失: dis=%q zone_id=%q zone_name=%q run_type=%q student_id=%q run_rule_id=%q", distance, zoneId, zoneName, runType, studentId, runRuleId)
		return "", fmt.Errorf("请将信息填写完整")
	}
	taskListCount := len(taskList)
	if taskListCount <= 0 {
		return "", fmt.Errorf("请添加跑步任务")
	}

	price, err := s.GetPrice(uid)
	if err != nil {
		return "", err
	}
	money := math.Round(price*float64(taskListCount)*10000) / 10000

	var balance float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&balance)
	if balance < money {
		return "", fmt.Errorf("余额不足")
	}
	if money < 0 {
		return "", fmt.Errorf("价格异常")
	}

	params := map[string]string{
		"form[phone]":       phone,
		"form[password]":    password,
		"form[dis]":         distance,
		"form[zone_id]":     zoneId,
		"form[zone_name]":   zoneName,
		"form[run_type]":    runType,
		"form[student_id]":  studentId,
		"form[run_rule_id]": runRuleId,
	}
	for i, task := range taskList {
		if taskMap, ok := task.(map[string]interface{}); ok {
			for k, v := range taskMap {
				params[fmt.Sprintf("form[task_list][%d][%s]", i, k)] = fmt.Sprintf("%v", v)
			}
		}
	}

	result, err := s.upstreamRequest("add", params)
	if err != nil {
		log.Printf("[SDXY-AddOrder] 上游请求失败: %v", err)
		return "", err
	}
	log.Printf("[SDXY-AddOrder] 上游响应: %+v", result)

	code := mapGetFloat(result, "code")
	if code != 0 {
		msg := mapGetString(result, "msg")
		if msg == "" {
			msg = "下单失败"
		}
		return "", fmt.Errorf("%s", msg)
	}

	data, _ := result["data"].(map[string]interface{})
	aggOrderId := ""
	sdxyOrderId := ""
	school := ""
	runRule := ""
	if data != nil {
		aggOrderId = mapGetString(data, "agg_order_id")
		if subOrder, ok := data["sub_order"].(map[string]interface{}); ok {
			sdxyOrderId = mapGetString(subOrder, "sdxy_order_id")
			if student, ok := subOrder["student"].(map[string]interface{}); ok {
				if schoolObj, ok := student["school"].(map[string]interface{}); ok {
					school = mapGetString(schoolObj, "name")
				}
				if runRuleObj, ok := student["run_rule"].(map[string]interface{}); ok {
					runRule = mapGetString(runRuleObj, "label")
				}
			}
		}
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = database.DB.Exec(
		`INSERT INTO qingka_wangke_flash_sdxy
		 (uid, agg_order_id, sdxy_order_id, user, pass, school, num, distance, run_type, run_rule, pause, status, fees, created_at)
		 VALUES (?,?,?,?,?,?,?,?,?,?,1,'1',?,?)`,
		uid, aggOrderId, sdxyOrderId, phone, password, school, taskListCount, distance, runType, runRule,
		fmt.Sprintf("%.4f", money), now,
	)
	if err != nil {
		return "", fmt.Errorf("本地保存失败: %v", err)
	}

	database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ? AND money >= ?", money, uid, money)

	logContent := fmt.Sprintf("闪电闪动校园 %s %s 成功下单，扣除%.4f元！", phone, password, money)
	database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'sdxy_add', ?, ?, ?)",
		uid, -money, logContent, now)

	return fmt.Sprintf("下单成功，扣费 %.2f 元", money), nil
}

func (s *SDXYService) RefundOrder(uid int, aggOrderId string, isAdmin bool) (string, error) {
	var order SDXYOrder
	err := database.DB.QueryRow(
		"SELECT id, uid, user, pass, fees FROM qingka_wangke_flash_sdxy WHERE agg_order_id = ? LIMIT 1",
		aggOrderId,
	).Scan(&order.ID, &order.UID, &order.User, &order.Pass, &order.Fees)
	if err != nil {
		return "", fmt.Errorf("订单不存在")
	}
	if !isAdmin && order.UID != uid {
		return "", fmt.Errorf("您暂无权限")
	}

	result, err := s.upstreamRequest("refund", map[string]string{
		"agg_order_id": aggOrderId,
	})
	if err != nil {
		return "", err
	}

	code := mapGetFloat(result, "code")
	if code != 0 {
		msg := mapGetString(result, "msg")
		if msg == "" {
			msg = "退款失败"
		}
		return "", fmt.Errorf("%s", msg)
	}

	remainNum := 0
	if d, ok := result["data"].(map[string]interface{}); ok {
		remainNum = int(mapGetFloat(d, "cnt"))
	}
	price, _ := s.GetPrice(order.UID)
	refundTotal := math.Round(price*float64(remainNum)*10000) / 10000

	if refundTotal > 0 {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", refundTotal, order.UID)
	}

	database.DB.Exec("UPDATE qingka_wangke_flash_sdxy SET status = '5' WHERE agg_order_id = ? LIMIT 1", aggOrderId)

	now := time.Now().Format("2006-01-02 15:04:05")
	logContent := fmt.Sprintf("闪电闪动校园 %s %s 成功退款，增加%.4f元！", order.User, order.Pass, refundTotal)
	database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'sdxy_refund', ?, ?, ?)",
		order.UID, refundTotal, logContent, now)

	return fmt.Sprintf("退款成功，退还 %.2f 元", refundTotal), nil
}

func (s *SDXYService) PauseOrder(uid int, aggOrderId string, pause int, isAdmin bool) (string, error) {
	var orderUID int
	err := database.DB.QueryRow(
		"SELECT uid FROM qingka_wangke_flash_sdxy WHERE agg_order_id = ? LIMIT 1", aggOrderId,
	).Scan(&orderUID)
	if err != nil {
		return "", fmt.Errorf("订单不存在")
	}
	if !isAdmin && orderUID != uid {
		return "", fmt.Errorf("您暂无权限")
	}

	result, err := s.upstreamRequest("pause", map[string]string{
		"agg_order_id": aggOrderId,
		"pause":        fmt.Sprintf("%d", pause),
	})
	if err != nil {
		return "", err
	}

	code := mapGetFloat(result, "code")
	if code != 0 {
		msg := mapGetString(result, "msg")
		if msg == "" {
			msg = "操作失败"
		}
		return "", fmt.Errorf("%s", msg)
	}

	database.DB.Exec("UPDATE qingka_wangke_flash_sdxy SET pause = ? WHERE agg_order_id = ? LIMIT 1", pause, aggOrderId)
	return "操作成功", nil
}
