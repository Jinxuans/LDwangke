package service

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"go-api/internal/database"
)

func (s *SXDKService) GetPrice(uid int, platform string) (float64, error) {
	var addprice float64
	err := database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
	if err != nil {
		return 0, fmt.Errorf("用户不存在")
	}
	return sxdkGetPlatformPrice(platform, addprice), nil
}

func (s *SXDKService) ListOrders(uid int, isAdmin bool, page, pageSize int, searchField, searchValue string) ([]map[string]interface{}, int, error) {
	offset := (page - 1) * pageSize
	var args []interface{}
	where := "1=1"

	if !isAdmin {
		where = "uid = ?"
		args = append(args, uid)
	}
	if searchValue != "" && searchField != "" {
		where += fmt.Sprintf(" AND %s LIKE ?", searchField)
		args = append(args, "%"+searchValue+"%")
	}

	var total int
	err := database.DB.QueryRow("SELECT COUNT(id) FROM qingka_wangke_sxdk WHERE "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	querySQL := fmt.Sprintf("SELECT * FROM qingka_wangke_sxdk WHERE %s ORDER BY id DESC LIMIT ?, ?", where)
	args = append(args, offset, pageSize)

	rows, err := database.DB.Query(querySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	var orders []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		if wxpushStr, ok := row["wxpush"].(string); ok {
			wxpushData := processWxpush(wxpushStr)
			for k, v := range wxpushData {
				row[k] = v
			}
		}
		orders = append(orders, row)
	}

	if orders == nil {
		orders = []map[string]interface{}{}
	}
	return orders, total, nil
}

func (s *SXDKService) AddOrder(uid int, form map[string]interface{}) (string, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return "", fmt.Errorf("泰山打卡未配置")
	}

	var addprice, money float64
	err = database.DB.QueryRow("SELECT addprice, money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice, &money)
	if err != nil {
		return "", fmt.Errorf("用户不存在")
	}

	platform, _ := form["platform"].(string)
	phone, _ := form["phone"].(string)
	password, _ := form["password"].(string)
	endTimeStr, _ := form["end_time"].(string)
	checkWeek, _ := form["check_week"].(string)

	day := timeCalcTrueday(time.Now(), endTimeStr, checkWeek)

	var count int
	database.DB.QueryRow("SELECT COUNT(1) FROM qingka_wangke_sxdk WHERE uid=? AND phone=? AND platform=?", uid, phone, platform).Scan(&count)
	if count > 0 {
		return "", fmt.Errorf("订单已存在")
	}

	endTimeParsed, _ := time.Parse("2006-01-02", endTimeStr)
	endOfDay := endTimeParsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	if endOfDay.Before(time.Now()) {
		return "", fmt.Errorf("下单天数不符合规范")
	}

	bei := 1
	if platform == "xyb" {
		if rt, ok := form["runType"]; ok {
			if rtf, ok := rt.(float64); ok && int(rtf) == 3 {
				bei = 5
			}
		}
	}
	totalMoney := sxdkGetPlatformPrice(platform, addprice) * float64(day) * float64(bei)
	if totalMoney < 0 {
		return "", fmt.Errorf("价格异常")
	}
	if money < math.Round(totalMoney) {
		return "", fmt.Errorf("余额不足")
	}

	sxdkLog(uid, "TaiShan-准备添加订单", fmt.Sprintf("%s %s %s 下单天数：%d ,结束日期：%s", platform, phone, password, day, endTimeStr), 0)
	result, err := s.upstreamPost(cfg, "addOrder", form)
	if err != nil {
		return "", fmt.Errorf("源台请求失败")
	}
	code, _ := result["code"].(float64)
	if int(code) != 0 {
		msg, _ := result["msg"].(string)
		if msg == "" {
			msg = "源台下单失败"
		}
		return "", fmt.Errorf("%s", msg)
	}

	selectResult, ok := result["selectOrderById"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("下单失败，请联系管理员")
	}
	selectCode, _ := selectResult["code"].(float64)
	if int(selectCode) != 0 {
		return "", fmt.Errorf("下单失败，源台未找到订单，请联系管理员")
	}
	selectData, _ := selectResult["data"].([]interface{})
	if len(selectData) == 0 {
		return "", fmt.Errorf("下单失败，请联系管理员")
	}
	firstOrder, _ := selectData[0].(map[string]interface{})
	sxdkID, _ := firstOrder["id"].(float64)

	upCheckTime, _ := form["up_check_time"].(string)
	if upCheckTime == "" {
		upCheckTime, _ = form["check_time"].(string)
	}
	downCheckTime, _ := form["down_check_time"].(string)
	nameVal, _ := form["name"].(string)
	address, _ := form["address"].(string)
	dayPaper, _ := form["day_paper"].(float64)
	weekPaper, _ := form["week_paper"].(float64)
	monthPaper, _ := form["month_paper"].(float64)

	wxpushMap := map[string]interface{}{"wxpush": ""}
	if platform == "xyb" {
		if rt, ok := form["runType"]; ok {
			wxpushMap["runType"] = rt
		}
	}
	wxpushJSON, _ := json.Marshal(wxpushMap)

	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = database.DB.Exec(
		`INSERT INTO qingka_wangke_sxdk (sxdkId, uid, platform, phone, password, code, wxpush, name, address, up_check_time, down_check_time, check_week, end_time, day_paper, week_paper, month_paper, createTime)
		VALUES (?, ?, ?, ?, ?, 1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		int(sxdkID), uid, platform, phone, password, string(wxpushJSON), nameVal, address,
		upCheckTime, downCheckTime, checkWeek, endTimeStr,
		int(dayPaper), int(weekPaper), int(monthPaper), now,
	)
	if err != nil {
		return "", fmt.Errorf("本地订单写入失败: %v", err)
	}

	database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", totalMoney, uid)
	sxdkLog(uid, "TaiShan-本台添加订单成功", fmt.Sprintf("%s %s %s 下单天数：%d ,结束日期：%s 扣除%.2f元！", platform, phone, password, day, endTimeStr, totalMoney), -totalMoney)

	return fmt.Sprintf("订单添加成功，扣除%.2f元！", totalMoney), nil
}

func (s *SXDKService) DeleteOrder(uid int, id int, isAdmin bool, delReturnMoney bool) (string, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return "", fmt.Errorf("泰山打卡未配置")
	}

	var sxdkID int
	var platform, checkWeek, endTimeStr, wxpushStr string
	query := "SELECT sxdkId, platform, check_week, end_time, COALESCE(wxpush,'') FROM qingka_wangke_sxdk WHERE id = ?"
	if !isAdmin {
		query += " AND (uid = ? OR 1 = ?)"
	}
	if isAdmin {
		err = database.DB.QueryRow(query, id).Scan(&sxdkID, &platform, &checkWeek, &endTimeStr, &wxpushStr)
	} else {
		err = database.DB.QueryRow(query, id, uid, uid).Scan(&sxdkID, &platform, &checkWeek, &endTimeStr, &wxpushStr)
	}
	if err != nil {
		return "", fmt.Errorf("您无此订单")
	}

	orderData := map[string]interface{}{
		"id":         sxdkID,
		"platform":   platform,
		"check_week": checkWeek,
		"end_time":   endTimeStr,
		"wxpush":     wxpushStr,
	}
	resp, err := s.upstreamPost(cfg, "deleteOrder", orderData)
	if err != nil {
		return "", fmt.Errorf("删除失败，请联系管理员")
	}
	respCode, _ := resp["code"].(float64)
	if int(respCode) != 0 {
		msg, _ := resp["msg"].(string)
		if msg == "" {
			msg = "删除失败，请联系管理员"
		}
		return "", fmt.Errorf("%s", msg)
	}

	otherMsg := ""
	returnMoney := 0.0
	if delReturnMoney {
		day := timeCalcTrueday(time.Now(), endTimeStr, checkWeek)
		wxpushData := processWxpush(wxpushStr)
		bei := 1
		if platform == "xyb" {
			if rt, ok := wxpushData["runType"]; ok {
				if rtf, ok := rt.(float64); ok && int(rtf) == 3 {
					bei = 5
				}
			}
		}
		var addprice float64
		database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
		returnMoney = sxdkGetPlatformPrice(platform, addprice) * float64(day) * float64(bei)

		endTimeParsed, _ := time.Parse("2006-01-02", endTimeStr)
		endOfDay := endTimeParsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		if returnMoney > 0 && endOfDay.After(time.Now()) {
			otherMsg = fmt.Sprintf(",订单未到期，已退款：%.2f", returnMoney)
			database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", returnMoney, uid)
		} else {
			otherMsg = ",此订单已到期，无需退款"
			returnMoney = 0
		}
	}

	database.DB.Exec("DELETE FROM qingka_wangke_sxdk WHERE id = ?", id)
	sxdkLog(uid, "TaiShan-删除订单成功", fmt.Sprintf("订单本台id：%d%s", id, otherMsg), returnMoney)
	return "删除成功" + otherMsg, nil
}

func (s *SXDKService) EditOrder(uid int, form map[string]interface{}) (string, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return "", fmt.Errorf("泰山打卡未配置")
	}

	var addprice, money float64
	err = database.DB.QueryRow("SELECT addprice, money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice, &money)
	if err != nil {
		return "", fmt.Errorf("用户不存在")
	}

	idVal, _ := form["id"].(float64)
	id := int(idVal)
	platform, _ := form["platform"].(string)
	phone, _ := form["phone"].(string)
	password, _ := form["password"].(string)
	nameVal, _ := form["name"].(string)
	address, _ := form["address"].(string)
	upCheckTime, _ := form["up_check_time"].(string)
	downCheckTime, _ := form["down_check_time"].(string)
	checkWeek, _ := form["check_week"].(string)
	endTimeStr, _ := form["end_time"].(string)
	dayPaper, _ := form["day_paper"].(float64)
	weekPaper, _ := form["week_paper"].(float64)
	monthPaper, _ := form["month_paper"].(float64)

	var sxdkID int
	var oldEndTime, oldCheckWeek, oldWxpush string
	err = database.DB.QueryRow("SELECT sxdkId, end_time, check_week FROM qingka_wangke_sxdk WHERE uid = ? AND phone = ? AND id = ?",
		uid, phone, id).Scan(&sxdkID, &oldEndTime, &oldCheckWeek)
	if err != nil {
		return "", fmt.Errorf("您无此订单")
	}

	var day int
	oldEndParsed, _ := time.Parse("2006-01-02", oldEndTime)
	oldEndOfDay := oldEndParsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	newEndParsed, _ := time.Parse("2006-01-02", endTimeStr)
	newEndOfDay := newEndParsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	if time.Now().After(oldEndOfDay) && time.Now().Before(newEndOfDay) {
		day = timeCalcTrueday(time.Now(), endTimeStr, checkWeek)
	} else {
		if newEndOfDay.Before(oldEndOfDay) || newEndOfDay.Equal(oldEndOfDay) {
			if oldCheckWeek == checkWeek {
				day = 0
			} else {
				oldDay := timeCalcTrueday(time.Now(), oldEndTime, oldCheckWeek)
				newDay := timeCalcTrueday(time.Now(), endTimeStr, checkWeek)
				day = newDay - oldDay
				if day < 0 {
					day = 0
				}
			}
		} else {
			oldDay := timeCalcTrueday(time.Now(), oldEndTime, oldCheckWeek)
			newDay := timeCalcTrueday(time.Now(), endTimeStr, checkWeek)
			day = newDay - oldDay
			if day < 0 {
				day = 0
			}
		}
	}

	bei := 1
	if platform == "xyb" {
		if rt, ok := form["runType"]; ok {
			if rtf, ok := rt.(float64); ok && int(rtf) == 3 {
				bei = 5
			}
		}
	}
	totalMoney := sxdkGetPlatformPrice(platform, addprice) * float64(day) * float64(bei)
	if money < math.Round(totalMoney) {
		return "", fmt.Errorf("余额不足")
	}

	form["id"] = sxdkID
	result, err := s.upstreamPost(cfg, "editOrder", form)
	if err != nil {
		return "", fmt.Errorf("源台编辑失败")
	}
	resultCode, _ := result["code"].(float64)
	if int(resultCode) != 0 {
		msg, _ := result["msg"].(string)
		if msg == "" {
			msg = "源台编辑失败"
		}
		return "", fmt.Errorf("%s", msg)
	}

	if upCheckTime == "" {
		if ct, ok := form["check_time"].(string); ok {
			upCheckTime = ct
		}
	}

	wxpushData := processWxpush(oldWxpush)
	if platform == "xyb" {
		if rt, ok := form["runType"]; ok {
			wxpushData["runType"] = rt
		}
	}
	wxpushJSON, _ := json.Marshal(wxpushData)

	now := time.Now().Format("2006-01-02 15:04:05")
	database.DB.Exec(
		`UPDATE qingka_wangke_sxdk SET password=?, name=?, address=?, up_check_time=?, down_check_time=?, check_week=?, end_time=?, wxpush=?, day_paper=?, week_paper=?, month_paper=?, updateTime=? WHERE id=?`,
		password, nameVal, address, upCheckTime, downCheckTime, checkWeek, endTimeStr,
		string(wxpushJSON), int(dayPaper), int(weekPaper), int(monthPaper), now, id,
	)

	database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", totalMoney, uid)
	sxdkLog(uid, "TaiShan-编辑订单成功", fmt.Sprintf("%s %s %s 增加天数：%d ,结束日期：%s 扣除%.2f元！", platform, phone, password, day, endTimeStr, totalMoney), -totalMoney)

	resultMsg, _ := result["msg"].(string)
	return fmt.Sprintf("订单修改成功,扣费：%.2f %s", totalMoney, resultMsg), nil
}

func (s *SXDKService) SearchPhoneInfo(uid int, form map[string]interface{}) (interface{}, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return nil, fmt.Errorf("泰山打卡未配置")
	}
	result, err := s.upstreamPost(cfg, "searchPhoneInfo", form)
	if err != nil {
		return nil, err
	}
	sxdkLog(uid, "TaiShan-自动获取信息", fmt.Sprintf("%v", form["platform"]), 0)
	return result, nil
}

func (s *SXDKService) NowCheck(uid int, id int, platform string) (interface{}, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return nil, fmt.Errorf("泰山打卡未配置")
	}

	var addprice, money float64
	err = database.DB.QueryRow("SELECT addprice, money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice, &money)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	pricePer := sxdkGetPlatformPrice(platform, addprice)
	if money < math.Round(pricePer) {
		return nil, fmt.Errorf("余额不足")
	}

	var sxdkID int
	var orderPlatform string
	err = database.DB.QueryRow("SELECT sxdkId, platform FROM qingka_wangke_sxdk WHERE uid = ? AND id = ?", uid, id).Scan(&sxdkID, &orderPlatform)
	if err != nil {
		return nil, fmt.Errorf("您无此订单")
	}

	result, err := s.upstreamPost(cfg, "nowCheck", map[string]interface{}{"id": sxdkID, "platform": orderPlatform})
	if err != nil {
		return nil, fmt.Errorf("打卡失败")
	}
	resultCode, _ := result["code"].(float64)
	if int(resultCode) == 0 {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", pricePer, uid)
		sxdkLog(uid, "TaiShan-立即打卡成功", fmt.Sprintf("平台：%s 扣除%.2f元！", platform, pricePer), -pricePer)
	}
	return result, nil
}

func (s *SXDKService) ProxyAction(uid int, id int, isAdmin bool, endpoint string, extraData map[string]interface{}) (interface{}, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return nil, fmt.Errorf("泰山打卡未配置")
	}

	query := "SELECT sxdkId, phone, platform, COALESCE(wxpush,'') FROM qingka_wangke_sxdk WHERE id = ?"
	if !isAdmin {
		query += " AND (uid = ? OR 1 = ?)"
	}
	var sxdkID int
	var phone, platform, wxpush string
	if isAdmin {
		err = database.DB.QueryRow(query, id).Scan(&sxdkID, &phone, &platform, &wxpush)
	} else {
		err = database.DB.QueryRow(query, id, uid, uid).Scan(&sxdkID, &phone, &platform, &wxpush)
	}
	if err != nil {
		return nil, fmt.Errorf("您无此订单")
	}

	data := map[string]interface{}{
		"id":       sxdkID,
		"phone":    phone,
		"platform": platform,
	}
	if extraData != nil {
		for k, v := range extraData {
			data[k] = v
		}
	}

	result, err := s.upstreamPost(cfg, endpoint, data)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *SXDKService) ChangeCheckCode(uid int, id int, code int, isAdmin bool) (interface{}, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return nil, fmt.Errorf("泰山打卡未配置")
	}

	query := "SELECT sxdkId, platform, code FROM qingka_wangke_sxdk WHERE id = ?"
	if !isAdmin {
		query += " AND (uid = ? OR 1 = ?)"
	}
	var sxdkID int
	var platform string
	var oldCode int
	if isAdmin {
		err = database.DB.QueryRow(query, id).Scan(&sxdkID, &platform, &oldCode)
	} else {
		err = database.DB.QueryRow(query, id, uid, uid).Scan(&sxdkID, &platform, &oldCode)
	}
	if err != nil {
		return nil, fmt.Errorf("您无此订单")
	}

	result, err := s.upstreamPost(cfg, "setCheckCode", map[string]interface{}{
		"id": sxdkID, "platform": platform, "code": code,
	})
	if err != nil {
		return nil, err
	}
	resultCode, _ := result["code"].(float64)
	if int(resultCode) == 0 {
		database.DB.Exec("UPDATE qingka_wangke_sxdk SET code = ? WHERE id = ?", code, id)
		sxdkLog(uid, "TaiShan-改变订单状态成功", fmt.Sprintf("订单本台id：%d,修改状态为：%d", id, code), 0)
	}
	return result, nil
}

func (s *SXDKService) SyncOrders() (string, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return "", fmt.Errorf("泰山打卡未配置")
	}

	result, err := s.upstreamPost(cfg, "yunOrder", nil)
	if err != nil {
		return "", fmt.Errorf("拉取失败")
	}
	resultCode, _ := result["code"].(float64)
	if int(resultCode) != 0 {
		msg, _ := result["msg"].(string)
		return "", fmt.Errorf("拉取失败：%s", msg)
	}

	data, _ := result["data"].([]interface{})
	for _, item := range data {
		row, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		rowID, _ := row["id"].(float64)
		rowPlatform, _ := row["platform"].(string)
		rowCode, _ := row["code"].(float64)
		rowWxpush, _ := row["wxpush"].(string)
		rowEndTime, _ := row["end_time"].(string)

		var orderID int
		err := database.DB.QueryRow("SELECT id FROM qingka_wangke_sxdk WHERE sxdkId = ? AND platform = ? LIMIT 1",
			int(rowID), rowPlatform).Scan(&orderID)
		if err == nil {
			database.DB.Exec("UPDATE qingka_wangke_sxdk SET code=?, wxpush=?, end_time=? WHERE id=?",
				int(rowCode), rowWxpush, rowEndTime, orderID)
		}
	}

	return fmt.Sprintf("拉取完成！同步：%d条成功", len(data)), nil
}

func (s *SXDKService) GetUserrow() (interface{}, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return nil, fmt.Errorf("泰山打卡未配置")
	}
	result, err := s.upstreamPost(cfg, "get_userrow", nil)
	if err != nil {
		return nil, err
	}
	return result, nil
}
