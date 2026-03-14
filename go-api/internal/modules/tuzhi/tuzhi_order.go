package tuzhi

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"go-api/internal/database"
)

func (s *TuZhiService) AddOrder(uid int, form map[string]interface{}) (string, error) {
	cfg, _ := s.GetConfig()
	token, err := s.login(cfg)
	if err != nil {
		return "", err
	}

	goodsID := 0
	if gid, ok := form["goods_id"]; ok {
		switch v := gid.(type) {
		case float64:
			goodsID = int(v)
		case string:
			goodsID, _ = strconv.Atoi(v)
		}
	}
	username, _ := form["username"].(string)
	password, _ := form["password"].(string)
	workDeadline, _ := form["work_deadline"].(string)

	if workDeadline == "" {
		return "", fmt.Errorf("截至日期不能为空")
	}

	deadline, err := time.Parse("2006-01-02", workDeadline)
	if err != nil {
		return "", fmt.Errorf("截至日期格式错误")
	}
	now := time.Now()
	days := int(deadline.Sub(now).Hours()/24) + 1
	if days <= 0 {
		return "", fmt.Errorf("截至日期不能小于当前日期")
	}

	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_dakaaz WHERE user_id=? AND username=? AND goods_id=?", uid, username, goodsID).Scan(&count)
	if count > 0 {
		return "", fmt.Errorf("订单已存在")
	}

	goods, err := s.GetGoods()
	if err != nil {
		return "", err
	}
	overrides, _ := s.GetGoodsOverrides()
	var targetGoods map[string]interface{}
	for _, g := range goods {
		gidF, _ := g["id"].(float64)
		if int(gidF) == goodsID {
			targetGoods = g
			for _, ov := range overrides {
				if ov.GoodsID == goodsID && ov.Price > 0 {
					targetGoods["price"] = ov.Price
					break
				}
			}
			break
		}
	}
	if targetGoods == nil {
		return "", fmt.Errorf("商品不存在")
	}

	price, _ := targetGoods["price"].(float64)
	billingMethod := 1
	if bm, ok := targetGoods["billing_method"].(float64); ok {
		billingMethod = int(bm)
	}

	var addprice, money float64
	database.DB.QueryRow("SELECT addprice, money FROM qingka_wangke_user WHERE uid=?", uid).Scan(&addprice, &money)

	var totalMoney float64
	var billingMonths int
	if billingMethod == 2 {
		billingMonths = int(math.Ceil(float64(days) / 30))
		totalMoney = math.Round(addprice*price*100) / 100 * float64(billingMonths)
	} else {
		totalMoney = math.Round(addprice*price*100) / 100 * float64(days)
	}

	if money < totalMoney {
		return "", fmt.Errorf("余额不足，需要 %.2f 元", totalMoney)
	}

	result, err := s.upstreamRequest("POST", "/user/finance.order/add", token, form)
	if err != nil {
		return "", fmt.Errorf("上游下单失败: %v", err)
	}
	code, _ := result["code"].(float64)
	if code != 200 {
		msg, _ := result["msg"].(string)
		return "", fmt.Errorf("上游下单失败: %s", msg)
	}
	data, _ := result["data"].(map[string]interface{})
	apiID, _ := data["id"].(float64)

	nowTs := time.Now().Unix()
	workDays, _ := form["work_days"].(string)
	if workDays == "" {
		workDays = "1,2,3,4,5,6,7"
	}
	nickname, _ := form["nickname"].(string)
	school, _ := form["school"].(string)
	postname, _ := form["postname"].(string)
	address, _ := form["address"].(string)
	addressLat, _ := form["address_lat"].(string)
	addressLng, _ := form["address_lng"].(string)
	workTime, _ := form["work_time"].(string)
	offTime, _ := form["off_time"].(string)
	images, _ := form["images"].(string)
	holidayStatus := getIntFromForm(form, "holiday_status", 0)
	dailyReport := getIntFromForm(form, "daily_report", 0)
	weeklyReport := getIntFromForm(form, "weekly_report", 0)
	monthlyReport := getIntFromForm(form, "monthly_report", 0)
	weeklyReportTime := getIntFromForm(form, "weekly_report_time", 1)
	monthlyReportTime := getIntFromForm(form, "monthly_report_time", 0)
	tokenField, _ := form["token"].(string)
	uuidField, _ := form["uuid"].(string)
	userSchoolID, _ := form["user_school_id"].(string)
	randomPhone, _ := form["random_phone"].(string)
	isOffTime := getIntFromForm(form, "is_off_time", 1)
	xzPushURL, _ := form["xz_push_url"].(string)

	database.DB.Exec(`INSERT INTO qingka_wangke_dakaaz 
		(api_id, user_id, goods_id, username, password, nickname, school, postname, address, address_lat, address_lng, 
		 work_time, off_time, work_days, work_days_num, daily_report, weekly_report, monthly_report, 
		 weekly_report_time, monthly_report_time, holiday_status, token, uuid, user_school_id, random_phone, 
		 images, create_time, update_time, remark, status, is_status, work_deadline, billing_method, billing_months, 
		 is_off_time, xz_push_url, price) 
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,'',0,1,?,?,?,?,?,?)`,
		int(apiID), uid, goodsID, username, password, nickname, school, postname, address, addressLat, addressLng,
		workTime, offTime, workDays, days, dailyReport, weeklyReport, monthlyReport,
		weeklyReportTime, monthlyReportTime, holidayStatus, tokenField, uuidField, userSchoolID, randomPhone,
		images, nowTs, nowTs, workDeadline, billingMethod, billingMonths,
		isOffTime, xzPushURL, totalMoney)

	database.DB.Exec("UPDATE qingka_wangke_user SET money=money-? WHERE uid=?", totalMoney, uid)
	tuzhiLog(uid, "tuzhi-添加订单", fmt.Sprintf("商品%d %s 天数%d 扣%.2f", goodsID, username, days, totalMoney), -totalMoney)

	if billingMethod == 2 {
		var recID int
		var recPrice float64
		err := database.DB.QueryRow("SELECT id, price FROM qingka_wangke_daka_query_record WHERE user_id=? AND username=? AND is_success=1 LIMIT 1", uid, username).Scan(&recID, &recPrice)
		if err == nil && recID > 0 {
			database.DB.Exec("UPDATE qingka_wangke_user SET money=money+? WHERE uid=?", recPrice, uid)
			tuzhiLog(uid, "tuzhi-按月查询退费", fmt.Sprintf("%s 退%.2f", username, recPrice), recPrice)
			database.DB.Exec("DELETE FROM qingka_wangke_daka_query_record WHERE id=?", recID)
		}
	}

	return fmt.Sprintf("订单添加成功，扣除 %.2f 元", totalMoney), nil
}

func (s *TuZhiService) EditOrder(uid int, isAdmin bool, form map[string]interface{}) (string, error) {
	cfg, _ := s.GetConfig()
	token, err := s.login(cfg)
	if err != nil {
		return "", err
	}

	localID := getIntFromForm(form, "id", 0)
	if localID == 0 {
		return "", fmt.Errorf("订单ID不能为空")
	}
	username, _ := form["username"].(string)
	workDeadline, _ := form["work_deadline"].(string)
	if workDeadline == "" {
		return "", fmt.Errorf("截至日期不能为空")
	}

	var apiID, origDays, origBillingMonths, billingMethod, goodsIDLocal int
	var origDeadline string
	query := "SELECT api_id, work_days_num, work_deadline, billing_months, billing_method, goods_id FROM qingka_wangke_dakaaz WHERE id=?"
	if !isAdmin {
		query += fmt.Sprintf(" AND user_id=%d", uid)
	}
	err = database.DB.QueryRow(query, localID).Scan(&apiID, &origDays, &origDeadline, &origBillingMonths, &billingMethod, &goodsIDLocal)
	if err != nil {
		return "", fmt.Errorf("订单不存在")
	}

	if workDeadline < origDeadline {
		return "", fmt.Errorf("截至日期不能小于原截至日期 %s", origDeadline)
	}

	origDl, _ := time.Parse("2006-01-02", origDeadline)
	newDl, _ := time.Parse("2006-01-02", workDeadline)
	extraDays := int(newDl.Sub(origDl).Hours()/24) + 1

	form["id"] = apiID

	goods, _ := s.GetGoods()
	overrides, _ := s.GetGoodsOverrides()
	var targetPrice float64
	for _, g := range goods {
		gidF, _ := g["id"].(float64)
		if int(gidF) == goodsIDLocal {
			targetPrice, _ = g["price"].(float64)
			for _, ov := range overrides {
				if ov.GoodsID == goodsIDLocal && ov.Price > 0 {
					targetPrice = ov.Price
					break
				}
			}
			break
		}
	}

	var addprice, money float64
	database.DB.QueryRow("SELECT addprice, money FROM qingka_wangke_user WHERE uid=?", uid).Scan(&addprice, &money)

	var extraMoney float64
	if billingMethod == 2 {
		newMonths := int(math.Ceil(float64(extraDays) / 30))
		if newMonths > origBillingMonths {
			extraMoney = math.Round(addprice*targetPrice*100) / 100 * float64(newMonths-origBillingMonths)
		}
	} else {
		if extraDays > origDays {
			extraMoney = math.Round(addprice*targetPrice*100) / 100 * float64(extraDays-origDays)
		}
	}

	if extraMoney > 0 && money < extraMoney {
		return "", fmt.Errorf("余额不足，需补费 %.2f 元", extraMoney)
	}

	result, err := s.upstreamRequest("POST", "/user/finance.order/edit", token, form)
	if err != nil {
		return "", fmt.Errorf("上游编辑失败: %v", err)
	}
	code, _ := result["code"].(float64)
	if code != 200 {
		msg, _ := result["msg"].(string)
		return "", fmt.Errorf("上游编辑失败: %s", msg)
	}

	if extraMoney > 0 {
		database.DB.Exec("UPDATE qingka_wangke_user SET money=money-? WHERE uid=?", extraMoney, uid)
		tuzhiLog(uid, "tuzhi-编辑补费", fmt.Sprintf("订单%d 补%.2f", localID, extraMoney), -extraMoney)
	}

	nowTs := time.Now().Unix()
	workDays, _ := form["work_days"].(string)
	password, _ := form["password"].(string)
	nickname, _ := form["nickname"].(string)
	school, _ := form["school"].(string)
	postname, _ := form["postname"].(string)
	address, _ := form["address"].(string)
	addressLat, _ := form["address_lat"].(string)
	addressLng, _ := form["address_lng"].(string)
	workTime, _ := form["work_time"].(string)
	offTime, _ := form["off_time"].(string)
	images, _ := form["images"].(string)
	isOffTime := getIntFromForm(form, "is_off_time", 1)
	xzPushURL, _ := form["xz_push_url"].(string)

	database.DB.Exec(`UPDATE qingka_wangke_dakaaz SET 
		username=?, password=?, nickname=?, school=?, postname=?, address=?, address_lat=?, address_lng=?,
		work_time=?, off_time=?, work_days=?, work_days_num=?, 
		daily_report=?, weekly_report=?, monthly_report=?, weekly_report_time=?, monthly_report_time=?,
		holiday_status=?, token=?, uuid=?, user_school_id=?, random_phone=?,
		images=?, update_time=?, work_deadline=?, is_off_time=?, xz_push_url=?
		WHERE id=?`,
		username, password, nickname, school, postname, address, addressLat, addressLng,
		workTime, offTime, workDays, extraDays,
		getIntFromForm(form, "daily_report", 0), getIntFromForm(form, "weekly_report", 0),
		getIntFromForm(form, "monthly_report", 0), getIntFromForm(form, "weekly_report_time", 1),
		getIntFromForm(form, "monthly_report_time", 0), getIntFromForm(form, "holiday_status", 0),
		form["token"], form["uuid"], form["user_school_id"], form["random_phone"],
		images, nowTs, workDeadline, isOffTime, xzPushURL, localID)

	msg := "订单修改成功"
	if extraMoney > 0 {
		msg = fmt.Sprintf("订单修改成功，补费 %.2f 元", extraMoney)
	}
	return msg, nil
}

func (s *TuZhiService) DeleteOrder(uid, localID int, isAdmin bool) (string, error) {
	cfg, _ := s.GetConfig()
	token, err := s.login(cfg)
	if err != nil {
		return "", err
	}

	query := "SELECT api_id, goods_id, work_deadline, billing_method FROM qingka_wangke_dakaaz WHERE id=?"
	if !isAdmin {
		query += fmt.Sprintf(" AND user_id=%d", uid)
	} else {
		query += fmt.Sprintf(" AND (user_id=%d OR 1=%d)", uid, uid)
	}
	var apiID, goodsIDLocal, billingMethod int
	var workDeadline string
	err = database.DB.QueryRow(query, localID).Scan(&apiID, &goodsIDLocal, &workDeadline, &billingMethod)
	if err != nil {
		return "", fmt.Errorf("订单不存在")
	}

	_, err = s.upstreamRequest("POST", "/user/finance.order/delete", token, map[string]interface{}{"id": apiID})
	if err != nil {
		return "", fmt.Errorf("上游删除失败: %v", err)
	}

	refund := 0.0
	if billingMethod == 1 {
		goods, _ := s.GetGoods()
		overrides, _ := s.GetGoodsOverrides()
		var price float64
		for _, g := range goods {
			gidF, _ := g["id"].(float64)
			if int(gidF) == goodsIDLocal {
				price, _ = g["price"].(float64)
				for _, ov := range overrides {
					if ov.GoodsID == goodsIDLocal && ov.Price > 0 {
						price = ov.Price
						break
					}
				}
				break
			}
		}
		dl, _ := time.Parse("2006-01-02", workDeadline)
		remaining := int(dl.Sub(time.Now()).Hours()/24) + 1
		if remaining > 0 {
			var addprice float64
			database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid=?", uid).Scan(&addprice)
			refund = math.Round(addprice*price*float64(remaining)*100) / 100
			database.DB.Exec("UPDATE qingka_wangke_user SET money=money+? WHERE uid=?", refund, uid)
		}
	}

	database.DB.Exec("DELETE FROM qingka_wangke_dakaaz WHERE id=?", localID)
	tuzhiLog(uid, "tuzhi-删除订单", fmt.Sprintf("订单%d 退%.2f", localID, refund), refund)
	return "删除成功", nil
}

func (s *TuZhiService) ListOrders(uid int, isAdmin bool, page, limit int, keyword string) ([]map[string]interface{}, int, error) {
	where := "1=1"
	args := []interface{}{}
	if !isAdmin {
		where += " AND user_id=?"
		args = append(args, uid)
	}
	if keyword != "" {
		where += " AND (username LIKE ? OR nickname LIKE ?)"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}

	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_dakaaz WHERE "+where, args...).Scan(&total)

	offset := (page - 1) * limit
	args = append(args, offset, limit)
	rows, err := database.DB.Query("SELECT * FROM qingka_wangke_dakaaz WHERE "+where+" ORDER BY id DESC LIMIT ?, ?", args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	var list []map[string]interface{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		rows.Scan(ptrs...)
		row := map[string]interface{}{}
		for i, col := range cols {
			v := vals[i]
			if b, ok := v.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = v
			}
		}
		list = append(list, row)
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	return list, total, nil
}

func (s *TuZhiService) CheckInWork(uid, localID int, isAdmin bool) error {
	cfg, _ := s.GetConfig()
	token, err := s.login(cfg)
	if err != nil {
		return err
	}
	var apiID int
	query := "SELECT api_id FROM qingka_wangke_dakaaz WHERE id=?"
	if !isAdmin {
		query += fmt.Sprintf(" AND user_id=%d", uid)
	}
	err = database.DB.QueryRow(query, localID).Scan(&apiID)
	if err != nil {
		return fmt.Errorf("订单不存在")
	}
	result, err := s.upstreamRequest("POST", "/user/finance.order/checkInToWorkImmediately", token, map[string]interface{}{"id": apiID})
	if err != nil {
		return err
	}
	code, _ := result["code"].(float64)
	if code != 200 {
		msg, _ := result["msg"].(string)
		return fmt.Errorf("%s", msg)
	}
	return nil
}

func (s *TuZhiService) CheckOutWork(uid, localID int, isAdmin bool) error {
	cfg, _ := s.GetConfig()
	token, err := s.login(cfg)
	if err != nil {
		return err
	}
	var apiID int
	query := "SELECT api_id FROM qingka_wangke_dakaaz WHERE id=?"
	if !isAdmin {
		query += fmt.Sprintf(" AND user_id=%d", uid)
	}
	err = database.DB.QueryRow(query, localID).Scan(&apiID)
	if err != nil {
		return fmt.Errorf("订单不存在")
	}
	result, err := s.upstreamRequest("POST", "/user/finance.order/checkInImmediatelyAfterWork", token, map[string]interface{}{"id": apiID})
	if err != nil {
		return err
	}
	code, _ := result["code"].(float64)
	if code != 200 {
		msg, _ := result["msg"].(string)
		return fmt.Errorf("%s", msg)
	}
	return nil
}

func (s *TuZhiService) SyncOrders() (int, error) {
	cfg, _ := s.GetConfig()
	token, err := s.login(cfg)
	if err != nil {
		return 0, err
	}
	result, err := s.upstreamRequest("GET", "/user/finance.order/lists", token, map[string]interface{}{"page": 1, "limit": 10000})
	if err != nil {
		return 0, err
	}
	code, _ := result["code"].(float64)
	if code != 200 {
		msg, _ := result["msg"].(string)
		return 0, fmt.Errorf("%s", msg)
	}
	data, _ := result["data"].(map[string]interface{})
	lists, _ := data["lists"].([]interface{})
	synced := 0
	nowTs := time.Now().Unix()
	for _, item := range lists {
		order, _ := item.(map[string]interface{})
		orderID, _ := order["id"].(float64)
		goodsID, _ := order["goods_id"].(float64)
		status, _ := order["status"].(float64)
		isStatus, _ := order["is_status"].(float64)
		okNum, _ := order["work_days_ok_num"].(float64)
		remark, _ := order["remark"].(string)

		res, err := database.DB.Exec("UPDATE qingka_wangke_dakaaz SET status=?, is_status=?, work_days_ok_num=?, remark=?, update_time=? WHERE api_id=? AND goods_id=?",
			int(status), int(isStatus), int(okNum), remark, nowTs, int(orderID), int(goodsID))
		if err == nil {
			affected, _ := res.RowsAffected()
			synced += int(affected)
		}
	}

	oneDayAgo := time.Now().Unix() - 86400
	rows, err := database.DB.Query("SELECT id, user_id, price, username FROM qingka_wangke_daka_query_record WHERE is_success=0 AND create_time <= ?", oneDayAgo)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var recID, recUID int
			var recPrice float64
			var recUser string
			rows.Scan(&recID, &recUID, &recPrice, &recUser)
			database.DB.Exec("UPDATE qingka_wangke_user SET money=money+? WHERE uid=?", recPrice, recUID)
			tuzhiLog(recUID, "tuzhi-按月查询退费", fmt.Sprintf("%s 退%.2f", recUser, recPrice), recPrice)
			database.DB.Exec("DELETE FROM qingka_wangke_daka_query_record WHERE id=?", recID)
		}
	}

	return synced, nil
}

func tuzhiLog(uid int, logType, text string, money float64) {
	now := time.Now().Format("2006-01-02 15:04:05")
	var newBalance float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&newBalance)
	moneyStr := fmt.Sprintf("%.2f", money)
	if money > 0 {
		moneyStr = fmt.Sprintf("+%.2f", money)
	}
	database.DB.Exec(
		"INSERT INTO qingka_wangke_log (uid, type, text, money, smoney, ip, addtime) VALUES (?, ?, ?, ?, ?, '', ?)",
		uid, logType, text, moneyStr, fmt.Sprintf("%.2f", newBalance), now,
	)
}

func getIntFromForm(form map[string]interface{}, key string, def int) int {
	v, ok := form[key]
	if !ok {
		return def
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case string:
		i, err := strconv.Atoi(val)
		if err != nil {
			return def
		}
		return i
	case int:
		return val
	}
	return def
}
