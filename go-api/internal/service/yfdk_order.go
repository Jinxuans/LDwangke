package service

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"go-api/internal/database"
)

func (s *YFDKService) GetPrice(cid string, days int) (float64, error) {
	if cid == "" || days < 1 {
		return 0, fmt.Errorf("参数错误")
	}
	price := math.Round(float64(days)*yfdkGetPlatformPrice(cid)*100) / 100
	return price, nil
}

func (s *YFDKService) ListOrders(uid int, isAdmin bool, page, limit int, keyword, status, cid string) ([]YFDKOrder, int, error) {
	cfg, _ := s.GetConfig()
	offset := (page - 1) * limit
	var args []interface{}
	where := "1=1"

	if !isAdmin {
		where += " AND uid = ?"
		args = append(args, uid)
	}
	if keyword != "" {
		where += " AND (username LIKE ? OR password LIKE ? OR name LIKE ?)"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw)
	}
	if status != "" {
		switch status {
		case "2":
			where += " AND endtime < ?"
			args = append(args, time.Now().Format("2006-01-02"))
		case "3":
			where += " AND endtime <= ? AND endtime > ?"
			args = append(args, time.Now().AddDate(0, 0, 5).Format("2006-01-02"), time.Now().Format("2006-01-02"))
		default:
			where += " AND status = ?"
			args = append(args, status)
		}
	}
	if cid != "" {
		where += " AND cid = ?"
		args = append(args, cid)
	}

	var total int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_yfdk WHERE "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	querySQL := fmt.Sprintf("SELECT id, uid, oid, cid, username, password, COALESCE(school,''), COALESCE(name,''), COALESCE(email,''), COALESCE(offer,''), COALESCE(address,''), COALESCE(longitude,''), COALESCE(latitude,''), COALESCE(week,''), COALESCE(worktime,''), COALESCE(offwork,0), COALESCE(offtime,''), day, daily_fee, total_fee, COALESCE(day_report,1), COALESCE(week_report,0), COALESCE(week_date,7), COALESCE(month_report,0), COALESCE(month_date,25), COALESCE(skip_holidays,0), COALESCE(image,0), COALESCE(status,1), COALESCE(mark,''), endtime, COALESCE(create_time,''), COALESCE(update_time,'') FROM qingka_wangke_yfdk WHERE %s ORDER BY id DESC LIMIT ?, ?", where)
	args = append(args, offset, limit)

	rows, err := database.DB.Query(querySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []YFDKOrder
	var oidList []string
	for rows.Next() {
		var o YFDKOrder
		err := rows.Scan(&o.ID, &o.UID, &o.OID, &o.CID, &o.Username, &o.Password,
			&o.School, &o.Name, &o.Email, &o.Offer, &o.Address, &o.Longitude, &o.Latitude,
			&o.Week, &o.Worktime, &o.Offwork, &o.Offtime, &o.Day, &o.DailyFee, &o.TotalFee,
			&o.DayReport, &o.WeekReport, &o.WeekDate, &o.MonthReport, &o.MonthDate,
			&o.SkipHolidays, &o.Image, &o.Status, &o.Mark, &o.Endtime, &o.CreateTime, &o.UpdateTime)
		if err != nil {
			continue
		}
		o.Mark = "等待打卡"
		orders = append(orders, o)
		oidList = append(oidList, o.OID)
	}

	if len(oidList) > 0 && cfg != nil && cfg.BaseURL != "" {
		logResult, err := s.upstreamRequest("POST", strings.TrimRight(cfg.BaseURL, "/")+"/orders/latest-logs", map[string]interface{}{"oids": oidList}, cfg.Token)
		if err == nil {
			code, _ := logResult["code"].(float64)
			if int(code) == 200 {
				data, _ := logResult["data"].(map[string]interface{})
				logsMap, _ := data["logs"].(map[string]interface{})
				for i := range orders {
					if logEntry, ok := logsMap[orders[i].OID]; ok {
						if entry, ok := logEntry.(map[string]interface{}); ok {
							if content, ok := entry["content"].(string); ok && content != "" {
								orders[i].Mark = content
							}
						}
					}
				}
			}
		}
	}

	if orders == nil {
		orders = []YFDKOrder{}
	}
	return orders, total, nil
}

func (s *YFDKService) AddOrder(uid int, form map[string]interface{}) (string, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return "", fmt.Errorf("YF打卡未配置")
	}

	cidVal, _ := form["cid"].(string)
	dayVal, _ := form["day"].(float64)
	days := int(dayVal)
	userVal, _ := form["user"].(string)
	passVal, _ := form["pass"].(string)

	if userVal == "" || passVal == "" {
		return "", fmt.Errorf("账号和密码不能为空")
	}
	if days < 1 {
		return "", fmt.Errorf("打卡天数必须大于0")
	}
	if cidVal == "" {
		return "", fmt.Errorf("请选择平台")
	}

	totalMoney := math.Round(float64(days)*yfdkGetPlatformPrice(cidVal)*100) / 100
	dailyFee := yfdkGetPlatformPrice(cidVal)

	var money float64
	err = database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&money)
	if err != nil {
		return "", fmt.Errorf("用户不存在")
	}
	if money < totalMoney {
		return "", fmt.Errorf("余额不足，当前余额：%.2f元，需要：%.2f元", money, totalMoney)
	}

	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_yfdk WHERE uid = ? AND cid = ? AND username = ?", uid, cidVal, userVal).Scan(&count)
	if count > 0 {
		return "", fmt.Errorf("该账号已存在订单，请勿重复下单")
	}

	res, err := database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ? AND money >= ?", totalMoney, uid, totalMoney)
	if err != nil {
		return "", fmt.Errorf("扣费失败")
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return "", fmt.Errorf("余额不足")
	}

	form["username"] = userVal
	form["password"] = passVal
	form["days"] = days
	form["remark"] = uid

	result, err := s.upstreamRequest("POST", strings.TrimRight(cfg.BaseURL, "/")+"/order/create", form, cfg.Token)
	if err != nil || result == nil {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", totalMoney, uid)
		yfdkLog(uid, "YF打卡-源台调用失败", fmt.Sprintf("已退还%.2f元", totalMoney), totalMoney)
		return "", fmt.Errorf("源台请求失败，已退还扣除金额")
	}

	code, _ := result["code"].(float64)
	if int(code) != 200 {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", totalMoney, uid)
		msg := "创建订单失败"
		if m, ok := result["message"].(string); ok && m != "" {
			msg = m
		}
		yfdkLog(uid, "YF打卡-创建订单失败", msg, totalMoney)
		return "", fmt.Errorf("%s，已退还扣除金额", msg)
	}

	data, _ := result["data"].(map[string]interface{})
	orderID, _ := data["order_id"].(string)
	if orderID == "" {
		if oid, ok := data["order_id"].(float64); ok {
			orderID = fmt.Sprintf("%.0f", oid)
		}
	}
	endTime, _ := data["end_time"].(string)

	weekStr := ""
	if w, ok := form["week"]; ok {
		switch wv := w.(type) {
		case []interface{}:
			parts := make([]string, len(wv))
			for i, v := range wv {
				parts[i] = fmt.Sprintf("%v", v)
			}
			weekStr = strings.Join(parts, ",")
		case string:
			weekStr = wv
		}
	}

	getString := func(key string) string {
		if v, ok := form[key]; ok {
			return fmt.Sprintf("%v", v)
		}
		return ""
	}
	getInt := func(key string, def int) int {
		if v, ok := form[key].(float64); ok {
			return int(v)
		}
		return def
	}

	_, err = database.DB.Exec(
		`INSERT INTO qingka_wangke_yfdk (uid, oid, cid, username, password, school, name, email, offer, address, longitude, latitude, week, worktime, offwork, offtime, day, daily_fee, total_fee, day_report, week_report, week_date, month_report, month_date, skip_holidays, status, endtime)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?)`,
		uid, orderID, cidVal, userVal, passVal,
		getString("school"), getString("name"), getString("email"), getString("offer"),
		getString("address"), getString("longitude"), getString("latitude"),
		weekStr, getString("worktime"), getInt("offwork", 0), getString("offtime"),
		days, dailyFee, totalMoney,
		getInt("day_report", 1), getInt("week_report", 0), getInt("week_date", 7),
		getInt("month_report", 0), getInt("month_date", 25), getInt("skip_holidays", 0),
		endTime,
	)
	if err != nil {
		log.Printf("[YFDK] 插入订单失败: %v", err)
	}

	yfdkLog(uid, "YF打卡-添加订单成功", fmt.Sprintf("订单ID：%s", orderID), -totalMoney)
	return fmt.Sprintf("下单成功，已扣除%.2f元", totalMoney), nil
}

func (s *YFDKService) DeleteOrder(uid int, id int, isAdmin bool) (string, error) {
	cfg, _ := s.GetConfig()

	var query string
	if isAdmin {
		query = "SELECT id, uid, oid, cid, username, day, daily_fee, total_fee, endtime FROM qingka_wangke_yfdk WHERE id = ?"
	} else {
		query = "SELECT id, uid, oid, cid, username, day, daily_fee, total_fee, endtime FROM qingka_wangke_yfdk WHERE id = ? AND uid = ?"
	}
	var orderID, oid, cidStr, username, endtime string
	var orderUID, dayCount int
	var dailyFee, totalFee float64
	var err error
	if isAdmin {
		err = database.DB.QueryRow(query, id).Scan(&orderID, &orderUID, &oid, &cidStr, &username, &dayCount, &dailyFee, &totalFee, &endtime)
	} else {
		err = database.DB.QueryRow(query, id, uid).Scan(&orderID, &orderUID, &oid, &cidStr, &username, &dayCount, &dailyFee, &totalFee, &endtime)
	}
	if err != nil {
		return "", fmt.Errorf("订单不存在或无权删除")
	}

	refundAmount := 0.0
	refundMsg := ""
	today := time.Now().Format("2006-01-02")

	if endtime > today {
		todayTime, _ := time.Parse("2006-01-02", today)
		endTimeP, _ := time.Parse("2006-01-02", endtime)
		refundDays := int(math.Ceil(endTimeP.Sub(todayTime).Hours() / 24))

		if dailyFee <= 0 && dayCount > 0 && totalFee > 0 {
			dailyFee = math.Round(totalFee/float64(dayCount)*100) / 100
		}
		if dailyFee <= 0 {
			dailyFee = yfdkGetPlatformPrice(cidStr)
		}

		if refundDays > 0 && dailyFee > 0 {
			refundAmount = math.Round(float64(refundDays)*dailyFee*100) / 100
			if refundAmount > totalFee {
				refundAmount = totalFee
			}
			_, err := database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", refundAmount, orderUID)
			if err == nil {
				refundMsg = fmt.Sprintf("，已退还%d天费用：%.2f元（每日%.2f元）", refundDays, refundAmount, dailyFee)
				yfdkLog(orderUID, "YF打卡-订单退款", fmt.Sprintf("订单ID:%s,账号:%s,退还%d天,金额:%.2f元", oid, username, refundDays, refundAmount), refundAmount)
			}
		}
	}

	if cfg != nil && cfg.BaseURL != "" {
		s.upstreamRequest("DELETE", strings.TrimRight(cfg.BaseURL, "/")+"/order/"+oid, nil, cfg.Token)
	}

	var delErr error
	if isAdmin {
		_, delErr = database.DB.Exec("DELETE FROM qingka_wangke_yfdk WHERE id = ?", id)
	} else {
		_, delErr = database.DB.Exec("DELETE FROM qingka_wangke_yfdk WHERE id = ? AND uid = ?", id, uid)
	}
	if delErr != nil {
		if refundAmount > 0 {
			database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", refundAmount, orderUID)
		}
		return "", fmt.Errorf("本地订单删除失败")
	}

	yfdkLog(orderUID, "YF打卡-删除订单成功", fmt.Sprintf("订单ID:%s,账号:%s", oid, username), 0)
	return "订单删除成功" + refundMsg, nil
}

func (s *YFDKService) RenewOrder(uid int, id int, days int, isAdmin bool) (string, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return "", fmt.Errorf("YF打卡未配置")
	}

	var query string
	if isAdmin {
		query = "SELECT id, uid, oid, cid, day, total_fee FROM qingka_wangke_yfdk WHERE id = ?"
	} else {
		query = "SELECT id, uid, oid, cid, day, total_fee FROM qingka_wangke_yfdk WHERE id = ? AND uid = ?"
	}
	var orderUID, oldDays int
	var oid, cidStr string
	var oldTotalFee float64
	if isAdmin {
		err = database.DB.QueryRow(query, id).Scan(&id, &orderUID, &oid, &cidStr, &oldDays, &oldTotalFee)
	} else {
		err = database.DB.QueryRow(query, id, uid).Scan(&id, &orderUID, &oid, &cidStr, &oldDays, &oldTotalFee)
	}
	if err != nil {
		return "", fmt.Errorf("订单不存在或无权访问")
	}

	dailyFee := yfdkGetPlatformPrice(cidStr)
	totalMoney := math.Round(float64(days)*dailyFee*100) / 100

	var money float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", orderUID).Scan(&money)
	if money < totalMoney {
		return "", fmt.Errorf("余额不足，当前余额：%.2f元，需要：%.2f元", money, totalMoney)
	}

	res, _ := database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ? AND money >= ?", totalMoney, orderUID, totalMoney)
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return "", fmt.Errorf("余额不足")
	}

	result, err := s.upstreamRequest("POST", strings.TrimRight(cfg.BaseURL, "/")+"/order/"+oid+"/renew", map[string]interface{}{"days": days}, cfg.Token)
	if err != nil || result == nil {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", totalMoney, orderUID)
		return "", fmt.Errorf("续费请求失败，已退还扣除金额")
	}
	code, _ := result["code"].(float64)
	if int(code) != 200 {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", totalMoney, orderUID)
		msg := "续费失败"
		if m, ok := result["message"].(string); ok && m != "" {
			msg = m
		}
		return "", fmt.Errorf("%s，已退还扣除金额", msg)
	}

	data, _ := result["data"].(map[string]interface{})
	newEndTime, _ := data["new_end_time"].(string)

	newDays := oldDays + days
	newTotalFee := oldTotalFee + totalMoney
	database.DB.Exec("UPDATE qingka_wangke_yfdk SET day=?, daily_fee=?, total_fee=?, endtime=?, status=1, mark='等待打卡' WHERE id=?",
		newDays, dailyFee, newTotalFee, newEndTime, id)

	yfdkLog(orderUID, "YF打卡-续费成功", fmt.Sprintf("订单ID:%s,续费%d天,扣费%.2f元", oid, days, totalMoney), -totalMoney)
	return fmt.Sprintf("续费成功！续费%d天，已扣除%.2f元", days, totalMoney), nil
}

func (s *YFDKService) SaveOrder(uid int, isAdmin bool, form map[string]interface{}) error {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return fmt.Errorf("YF打卡未配置")
	}

	idVal, _ := form["id"].(float64)
	id := int(idVal)
	if id <= 0 {
		return fmt.Errorf("订单ID不能为空")
	}

	var oid, endtime string
	if isAdmin {
		err = database.DB.QueryRow("SELECT oid, endtime FROM qingka_wangke_yfdk WHERE id = ?", id).Scan(&oid, &endtime)
	} else {
		err = database.DB.QueryRow("SELECT oid, endtime FROM qingka_wangke_yfdk WHERE id = ? AND uid = ?", id, uid).Scan(&oid, &endtime)
	}
	if err != nil {
		return fmt.Errorf("订单不存在")
	}

	if statusVal, ok := form["status"]; ok && len(form) == 2 {
		if endtime < time.Now().Format("2006-01-02") {
			if s, ok := statusVal.(float64); ok && int(s) == 1 {
				return fmt.Errorf("订单已过期,无法开启。请先续费。")
			}
		}
	}

	apiData := map[string]interface{}{}
	apiFields := []string{"password", "email", "push_url", "address", "offer", "week", "worktime", "offtime",
		"offwork", "status", "day_report", "week_report", "month_report",
		"week_date", "month_date", "skip_holidays", "name", "school",
		"enrollment_year", "device_id", "cpdaily_info", "plan_name",
		"company", "company_address", "image", "remark"}
	for _, field := range apiFields {
		if v, ok := form[field]; ok {
			apiData[field] = v
		}
	}
	if v, ok := form["longitude"]; ok {
		apiData["long"] = v
	}
	if v, ok := form["latitude"]; ok {
		apiData["lat"] = v
	}

	result, err := s.upstreamRequest("POST", strings.TrimRight(cfg.BaseURL, "/")+"/order/"+oid+"/update", apiData, cfg.Token)
	if err != nil || result == nil {
		return fmt.Errorf("更新失败")
	}
	code, _ := result["code"].(float64)
	if int(code) != 200 {
		msg := "更新失败"
		if m, ok := result["message"].(string); ok && m != "" {
			msg = m
		}
		return fmt.Errorf("%s", msg)
	}

	updates := []string{}
	updateArgs := []interface{}{}
	localFields := map[string]string{"status": "status", "worktime": "worktime", "email": "email", "offer": "offer", "name": "name", "address": "address"}
	for formKey, dbCol := range localFields {
		if v, ok := form[formKey]; ok {
			updates = append(updates, dbCol+" = ?")
			updateArgs = append(updateArgs, v)
		}
	}
	if len(updates) > 0 {
		updates = append(updates, "update_time = NOW()")
		updateArgs = append(updateArgs, id)
		database.DB.Exec("UPDATE qingka_wangke_yfdk SET "+strings.Join(updates, ", ")+" WHERE id = ?", updateArgs...)
	}

	yfdkLog(uid, "YF打卡-更新成功", fmt.Sprintf("订单ID:%d", id), 0)
	return nil
}

func (s *YFDKService) ManualClock(uid int, id int, isAdmin bool) error {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return fmt.Errorf("YF打卡未配置")
	}

	var oid string
	if isAdmin {
		err = database.DB.QueryRow("SELECT oid FROM qingka_wangke_yfdk WHERE id = ?", id).Scan(&oid)
	} else {
		err = database.DB.QueryRow("SELECT oid FROM qingka_wangke_yfdk WHERE id = ? AND uid = ?", id, uid).Scan(&oid)
	}
	if err != nil {
		return fmt.Errorf("订单不存在或无权访问")
	}

	result, err := s.upstreamRequest("POST", strings.TrimRight(cfg.BaseURL, "/")+"/order/"+oid+"/clock", nil, cfg.Token)
	if err != nil {
		return fmt.Errorf("打卡失败")
	}
	code, _ := result["code"].(float64)
	if int(code) != 200 {
		msg := "打卡失败"
		if m, ok := result["message"].(string); ok && m != "" {
			msg = m
		}
		return fmt.Errorf("%s", msg)
	}
	yfdkLog(uid, "YF打卡-手动打卡成功", fmt.Sprintf("订单ID:%s", oid), 0)
	return nil
}

func (s *YFDKService) GetOrderLogs(uid int, id int, isAdmin bool) (interface{}, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return nil, fmt.Errorf("YF打卡未配置")
	}

	var oid string
	if isAdmin {
		err = database.DB.QueryRow("SELECT oid FROM qingka_wangke_yfdk WHERE id = ?", id).Scan(&oid)
	} else {
		err = database.DB.QueryRow("SELECT oid FROM qingka_wangke_yfdk WHERE id = ? AND uid = ?", id, uid).Scan(&oid)
	}
	if err != nil {
		return nil, fmt.Errorf("订单不存在或无权访问")
	}

	result, err := s.upstreamRequest("GET", strings.TrimRight(cfg.BaseURL, "/")+"/order/"+oid+"/logs?limit=100", nil, cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("获取日志失败")
	}
	code, _ := result["code"].(float64)
	if int(code) != 200 {
		return nil, fmt.Errorf("获取日志失败")
	}
	data, _ := result["data"].(map[string]interface{})
	return data["logs"], nil
}

func (s *YFDKService) GetOrderDetail(uid int, id int, isAdmin bool) (map[string]interface{}, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return nil, fmt.Errorf("YF打卡未配置")
	}

	var oid, localUsername, localPassword, localSchool, localName string
	if isAdmin {
		err = database.DB.QueryRow("SELECT oid, username, password, school, name FROM qingka_wangke_yfdk WHERE id = ?", id).
			Scan(&oid, &localUsername, &localPassword, &localSchool, &localName)
	} else {
		err = database.DB.QueryRow("SELECT oid, username, password, school, name FROM qingka_wangke_yfdk WHERE id = ? AND uid = ?", id, uid).
			Scan(&oid, &localUsername, &localPassword, &localSchool, &localName)
	}
	if err != nil {
		return nil, fmt.Errorf("订单不存在")
	}

	result, err := s.upstreamRequest("GET", strings.TrimRight(cfg.BaseURL, "/")+"/order/"+oid, nil, cfg.Token)
	if err != nil || result == nil {
		return nil, fmt.Errorf("获取订单详情失败")
	}
	code, _ := result["code"].(float64)
	if int(code) != 200 {
		return nil, fmt.Errorf("获取订单详情失败")
	}

	data, _ := result["data"].(map[string]interface{})
	orderData, _ := data["order"].(map[string]interface{})
	orderData["username"] = localUsername
	orderData["password"] = localPassword
	orderData["local_id"] = id
	if orderData["school"] == nil || orderData["school"] == "" {
		orderData["school"] = localSchool
	}
	if orderData["name"] == nil || orderData["name"] == "" {
		orderData["name"] = localName
	}

	return orderData, nil
}

func (s *YFDKService) PatchReport(uid int, id int, startDate, endDate, reportType string, isAdmin bool) (string, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return "", fmt.Errorf("YF打卡未配置")
	}

	var oid, cidStr string
	if isAdmin {
		err = database.DB.QueryRow("SELECT oid, cid FROM qingka_wangke_yfdk WHERE id = ?", id).Scan(&oid, &cidStr)
	} else {
		err = database.DB.QueryRow("SELECT oid, cid FROM qingka_wangke_yfdk WHERE id = ? AND uid = ?", id, uid).Scan(&oid, &cidStr)
	}
	if err != nil {
		return "", fmt.Errorf("订单不存在或无权访问")
	}

	patchPrice := yfdkGetPlatformPrice(cidStr)
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	diffDays := int(end.Sub(start).Hours()/24) + 1

	var count int
	switch reportType {
	case "day":
		count = diffDays
	case "week":
		count = int(math.Ceil(float64(diffDays) / 7))
	case "month":
		count = (end.Year()-start.Year())*12 + int(end.Month()-start.Month()) + 1
	default:
		count = diffDays
	}
	if count < 1 {
		count = 1
	}
	totalCost := math.Round(patchPrice*float64(count)*100) / 100

	var money float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&money)
	if money < totalCost {
		return "", fmt.Errorf("余额不足，当前余额：%.2f元，需要：%.2f元", money, totalCost)
	}

	res, _ := database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ? AND money >= ?", totalCost, uid, totalCost)
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return "", fmt.Errorf("扣费失败")
	}

	result, err := s.upstreamRequest("POST", strings.TrimRight(cfg.BaseURL, "/")+"/order/"+oid+"/patch-report",
		map[string]interface{}{"start_date": startDate, "end_date": endDate, "type": reportType}, cfg.Token)
	if err != nil || result == nil {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", totalCost, uid)
		return "", fmt.Errorf("补报告失败，已退还扣除金额")
	}
	code, _ := result["code"].(float64)
	if int(code) != 200 {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", totalCost, uid)
		msg := "补报告失败"
		if m, ok := result["message"].(string); ok && m != "" {
			msg = m
		}
		return "", fmt.Errorf("%s", msg)
	}

	yfdkLog(uid, "YF打卡-补报告成功", fmt.Sprintf("订单ID:%s,类型:%s,日期:%s至%s,扣费:%.2f元,共%d次", oid, reportType, startDate, endDate, totalCost, count), -totalCost)
	return fmt.Sprintf("补报告成功，扣费%.2f元，共%d次", totalCost, count), nil
}

func (s *YFDKService) CalculatePatchCost(uid int, id int, startDate, endDate, reportType string, isAdmin bool) (interface{}, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return nil, fmt.Errorf("YF打卡未配置")
	}

	var oid string
	if isAdmin {
		err = database.DB.QueryRow("SELECT oid FROM qingka_wangke_yfdk WHERE id = ?", id).Scan(&oid)
	} else {
		err = database.DB.QueryRow("SELECT oid FROM qingka_wangke_yfdk WHERE id = ? AND uid = ?", id, uid).Scan(&oid)
	}
	if err != nil {
		return nil, fmt.Errorf("订单不存在或无权访问")
	}

	result, err := s.upstreamRequest("POST", strings.TrimRight(cfg.BaseURL, "/")+"/order/"+oid+"/patch-report-calculate",
		map[string]interface{}{"start": startDate, "end": endDate, "type": reportType}, cfg.Token)
	if err != nil || result == nil {
		return nil, fmt.Errorf("计算费用失败")
	}
	code, _ := result["code"].(float64)
	if int(code) != 200 {
		msg := "计算费用失败"
		if m, ok := result["message"].(string); ok && m != "" {
			msg = m
		}
		return nil, fmt.Errorf("%s", msg)
	}
	return result["data"], nil
}
