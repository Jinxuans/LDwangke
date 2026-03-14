package service

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"go-api/internal/database"
)

func (s *YongyeService) ListOrders(uid int, isAdmin bool, page, limit int, keyword, statusFilter string) ([]YongyeOrder, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}

	if !isAdmin {
		where += " AND uid = ?"
		args = append(args, uid)
	}
	if keyword != "" {
		where += " AND (user LIKE ? OR id = ?)"
		args = append(args, "%"+keyword+"%", keyword)
	}
	if statusFilter != "" {
		where += " AND dockstatus = ?"
		args = append(args, statusFilter)
	}

	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM yy_ydsj_dd "+where, args...).Scan(&total)

	offset := (page - 1) * limit
	query := "SELECT id, pol, uid, user, pass, school, type, zkm, ks_h, ks_m, js_h, js_m, weeks, dockstatus, yfees, fees, COALESCE(yid,''), COALESCE(yaddtime,''), COALESCE(addtime,''), COALESCE(tktext,'') FROM yy_ydsj_dd " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []YongyeOrder
	for rows.Next() {
		var o YongyeOrder
		rows.Scan(&o.ID, &o.Pol, &o.UID, &o.User, &o.Pass, &o.School, &o.Type,
			&o.Zkm, &o.KsH, &o.KsM, &o.JsH, &o.JsM, &o.Weeks,
			&o.DockStatus, &o.Yfees, &o.Fees, &o.YID, &o.Yaddtime, &o.Addtime, &o.Tktext)
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []YongyeOrder{}
	}
	return orders, total, nil
}

func (s *YongyeService) ListStudents(uid int, isAdmin bool, keyword string) ([]YongyeStudent, error) {
	where := "WHERE 1=1"
	args := []interface{}{}

	if !isAdmin {
		where += " AND uid = ?"
		args = append(args, uid)
	}
	if keyword != "" {
		where += " AND user LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	rows, err := database.DB.Query("SELECT id, uid, user, pass, type, zkm, weeks, status, COALESCE(tdkm,0), COALESCE(tdmoney,0), COALESCE(stulog,''), COALESCE(last_time,'') FROM yy_ydsj_student "+where+" ORDER BY id DESC", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []YongyeStudent
	for rows.Next() {
		var s YongyeStudent
		rows.Scan(&s.ID, &s.UID, &s.User, &s.Pass, &s.Type, &s.Zkm, &s.Weeks, &s.Status, &s.Tdkm, &s.Tdmoney, &s.Stulog, &s.LastTime)
		students = append(students, s)
	}
	if students == nil {
		students = []YongyeStudent{}
	}
	return students, nil
}

func (s *YongyeService) AddOrder(uid int, form map[string]interface{}) (string, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return "", err
	}
	if cfg.ApiURL == "" || cfg.Token == "" {
		return "", fmt.Errorf("永夜运动未配置上游接口")
	}

	user := mapGetString(form, "user")
	pass := mapGetString(form, "pass")
	school := mapGetString(form, "school")
	runType := mapGetInt(form, "type")
	zkm := mapGetFloat(form, "zkm")
	ksH := mapGetInt(form, "ks_h")
	ksM := mapGetInt(form, "ks_m")
	jsH := mapGetInt(form, "js_h")
	jsM := mapGetInt(form, "js_m")
	weeks := mapGetString(form, "weeks")
	isPolling := mapGetInt(form, "isPolling")

	if user == "" || pass == "" || zkm <= 0 || weeks == "" {
		return "", fmt.Errorf("参数不完整")
	}
	if school == "" {
		school = "自动识别"
	}
	if ksH < 6 {
		ksH = 6
	}
	if ksH > 22 {
		ksH = 22
	}
	if jsH < 6 {
		jsH = 6
	}
	if jsH > 22 {
		jsH = 22
	}
	if (ksH == jsH && ksM == jsM) || jsH < ksH {
		ksH = 9
		ksM = 0
		jsH = 21
		jsM = 0
	}

	schoolPrice := 3.0
	schoolResp, err := s.yongyeUpstreamPost(cfg, "school", nil)
	if err == nil {
		var schoolResult map[string]interface{}
		if json.Unmarshal(schoolResp, &schoolResult) == nil && int(mapGetFloat(schoolResult, "code")) == 1 {
			if dataRaw, ok := schoolResult["data"].([]interface{}); ok {
				for _, d := range dataRaw {
					if item, ok := d.(map[string]interface{}); ok && mapGetString(item, "name") == school {
						if runType == 1 || ksH < 9 {
							schoolPrice = mapGetFloat(item, "cpmuch")
						} else {
							schoolPrice = mapGetFloat(item, "zcmuch")
						}
						break
					}
				}
			}
		}
	}

	if cfg.Zs > 0 {
		schoolPrice = math.Round(schoolPrice/cfg.Zs*cfg.Beis*100) / 100
	}

	var addprice float64
	database.DB.QueryRow("SELECT COALESCE(addprice, 0) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
	djfl := 1.0
	if cfg.Dj > 0 && addprice > 0 {
		djfl = math.Round(addprice/cfg.Dj*100) / 100
	}
	schoolPrice = math.Round(schoolPrice*djfl*100) / 100

	yfees := math.Round(zkm*schoolPrice*100) / 100

	var balance float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&balance)
	if balance < yfees || balance < 0 {
		return "", fmt.Errorf("余额不足，需要 %.2f 元，当前余额 %.2f 元", yfees, balance)
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	var dupID int
	database.DB.QueryRow("SELECT id FROM yy_ydsj_dd WHERE uid=? AND type=? AND user=? AND pass=? AND zkm=? AND ks_h=? AND ks_m=? AND js_h=? AND js_m=? AND weeks=? LIMIT 1",
		uid, runType, user, pass, zkm, ksH, ksM, jsH, jsM, weeks).Scan(&dupID)
	if dupID > 0 {
		return "", fmt.Errorf("重复提交，已阻止")
	}

	result, err := database.DB.Exec(
		"INSERT INTO yy_ydsj_dd (pol, uid, user, pass, school, type, zkm, ks_h, ks_m, js_h, js_m, weeks, dockstatus, yfees, addtime) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,0,?,?)",
		isPolling, uid, user, pass, school, runType, zkm, ksH, ksM, jsH, jsM, weeks, yfees, now,
	)
	if err != nil {
		return "", fmt.Errorf("订单创建失败: %v", err)
	}
	localID, _ := result.LastInsertId()

	database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", yfees, uid)

	logContent := fmt.Sprintf("永夜运动下单：账号%s %.1fKM 扣费%.2f", user, zkm, yfees)
	smoney := math.Round((balance-yfees)*100) / 100
	database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'yongye_add', ?, ?, ?)",
		uid, -yfees, logContent, now)
	database.DB.Exec("INSERT INTO qingka_wangke_log (uid, type, text, money, smoney, ip) VALUES (?, 'API永夜运动下单', ?, ?, ?, '')",
		uid, logContent, fmt.Sprintf("-%.2f", yfees), fmt.Sprintf("%.2f", smoney))

	apiData := map[string]string{
		"isPolling": fmt.Sprintf("%d", isPolling),
		"type":      fmt.Sprintf("%d", runType),
		"school":    school,
		"user":      user,
		"pass":      pass,
		"zkm":       fmt.Sprintf("%.2f", zkm),
		"ks_h":      fmt.Sprintf("%d", ksH),
		"ks_m":      fmt.Sprintf("%d", ksM),
		"js_h":      fmt.Sprintf("%d", jsH),
		"js_m":      fmt.Sprintf("%d", jsM),
		"weeks":     weeks,
		"addtime":   now,
	}
	respBody, err := s.yongyeUpstreamPost(cfg, "add", apiData)
	if err != nil {
		database.DB.Exec("UPDATE yy_ydsj_dd SET dockstatus = 2 WHERE id = ?", localID)
		log.Printf("[Yongye] 上游下单失败: %v", err)
		return fmt.Sprintf("提交成功(本地#%d)，上游请求失败将自动重试", localID), nil
	}

	var apiResp map[string]interface{}
	json.Unmarshal(respBody, &apiResp)

	if int(mapGetFloat(apiResp, "code")) == 1 {
		yid := fmt.Sprintf("%v", apiResp["id"])
		database.DB.Exec("UPDATE yy_ydsj_dd SET dockstatus = 1, yid = ? WHERE id = ?", yid, localID)
		return fmt.Sprintf("提交成功，扣费 %.2f 元", yfees), nil
	}

	database.DB.Exec("UPDATE yy_ydsj_dd SET dockstatus = 0 WHERE id = ?", localID)
	return fmt.Sprintf("提交成功(本地#%d)，上游处理中", localID), nil
}

func (s *YongyeService) RefundStudent(uid int, user string, runType int, isAdmin bool) (string, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return "", err
	}

	var stuUID int
	err = database.DB.QueryRow("SELECT uid FROM yy_ydsj_student WHERE user = ? AND uid = ? LIMIT 1", user, uid).Scan(&stuUID)
	if err != nil && !isAdmin {
		return "", fmt.Errorf("你的账号下无此学生")
	}

	respBody, err := s.yongyeUpstreamPost(cfg, "tuid", map[string]string{
		"user": user,
		"type": fmt.Sprintf("%d", runType),
	})
	if err != nil {
		return "", fmt.Errorf("上游退单请求失败: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	if int(mapGetFloat(result, "code")) == 1 {
		now := time.Now().Format("2006-01-02 15:04:05")
		database.DB.Exec("INSERT INTO qingka_wangke_log (uid, type, text, money, smoney, ip) VALUES (?, 'API永夜运动退单', ?, '0', '', '')",
			uid, fmt.Sprintf("账号：%s - 退单", user))
		_ = now
		return mapGetString(result, "msg"), nil
	}

	msg := mapGetString(result, "msg")
	if msg == "" {
		msg = "退单失败"
	}
	return "", fmt.Errorf("%s", msg)
}

func (s *YongyeService) UpdateStudent(uid int, form map[string]interface{}, isAdmin bool) (string, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return "", err
	}

	user := mapGetString(form, "user")
	pass := mapGetString(form, "pass")
	weeks := mapGetString(form, "weeks")
	statusStr := mapGetString(form, "status")

	if user == "" {
		return "", fmt.Errorf("请传递学生账号")
	}

	if !isAdmin {
		var stuUID int
		err = database.DB.QueryRow("SELECT uid FROM yy_ydsj_student WHERE user = ? AND uid = ? LIMIT 1", user, uid).Scan(&stuUID)
		if err != nil {
			return "", fmt.Errorf("你的账号下无此学生")
		}
	}

	extra := map[string]string{"user": user}
	if pass != "" {
		extra["pass"] = pass
	}
	if weeks != "" {
		extra["weeks"] = weeks
	}
	if statusStr != "" {
		extra["status"] = statusStr
	}

	respBody, err := s.yongyeUpstreamPost(cfg, "upstu", extra)
	if err != nil {
		return "", fmt.Errorf("上游请求失败: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	if int(mapGetFloat(result, "code")) == 1 {
		return mapGetString(result, "msg"), nil
	}
	msg := mapGetString(result, "msg")
	if msg == "" {
		msg = "修改失败"
	}
	return "", fmt.Errorf("%s", msg)
}

func (s *YongyeService) TogglePolling(uid, orderID int, isAdmin bool) (string, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return "", err
	}

	var order YongyeOrder
	err = database.DB.QueryRow("SELECT id, uid, yid, pol, dockstatus, yfees FROM yy_ydsj_dd WHERE id = ?", orderID).
		Scan(&order.ID, &order.UID, &order.YID, &order.Pol, &order.DockStatus, &order.Yfees)
	if err != nil {
		return "", fmt.Errorf("订单不存在")
	}
	if !isAdmin && order.UID != uid {
		return "", fmt.Errorf("无权操作")
	}

	respBody, err := s.yongyeUpstreamPost(cfg, "polgb", map[string]string{
		"id": order.YID,
	})
	if err != nil {
		return "", fmt.Errorf("上游请求失败: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	if int(mapGetFloat(result, "code")) == 1 {
		now := time.Now().Format("2006-01-02 15:04:05")
		if order.Pol == 0 {
			database.DB.Exec("UPDATE yy_ydsj_dd SET dockstatus = 5, pol = 1, tktext = '开启轮询模式' WHERE id = ?", orderID)
			database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", order.Yfees, order.UID)
			database.DB.Exec("INSERT INTO qingka_wangke_log (uid, type, text, money, smoney, ip) VALUES (?, '开启轮询', ?, ?, '', '')",
				order.UID, fmt.Sprintf("订单ID：%d - 开启轮询，扣除余额", orderID), fmt.Sprintf("-%.2f", order.Yfees))
			_ = now
			return "已开启轮询", nil
		}
		database.DB.Exec("UPDATE yy_ydsj_dd SET dockstatus = 3, pol = 0, tktext = '关闭轮询模式' WHERE id = ?", orderID)
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", order.Yfees, order.UID)
		database.DB.Exec("INSERT INTO qingka_wangke_log (uid, type, text, money, smoney, ip) VALUES (?, '关闭轮询', ?, ?, '', '')",
			order.UID, fmt.Sprintf("订单ID：%d - 关闭轮询，返还余额", orderID), fmt.Sprintf("+%.2f", order.Yfees))
		return "已关闭轮询", nil
	}

	msg := mapGetString(result, "msg")
	if msg == "" {
		msg = "操作失败"
	}
	return "", fmt.Errorf("%s", msg)
}

func (s *YongyeService) LocalRefund(uid, orderID int, isAdmin bool) (string, error) {
	var order YongyeOrder
	err := database.DB.QueryRow("SELECT id, uid, user, yfees, dockstatus FROM yy_ydsj_dd WHERE id = ?", orderID).
		Scan(&order.ID, &order.UID, &order.User, &order.Yfees, &order.DockStatus)
	if err != nil {
		return "", fmt.Errorf("订单不存在")
	}
	if !isAdmin && order.UID != uid {
		return "", fmt.Errorf("无权操作")
	}
	if order.DockStatus == 3 {
		return "", fmt.Errorf("该订单已退款")
	}

	cfg, _ := s.GetConfig()
	tkRate := 0.0
	if cfg != nil {
		tkRate = cfg.Tk
	}

	refund := math.Round(order.Yfees*(1-tkRate)*100) / 100
	if refund > 0 {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", refund, order.UID)
	}

	database.DB.Exec("UPDATE yy_ydsj_dd SET dockstatus = 3, fees = 0, tktext = ? WHERE id = ?",
		fmt.Sprintf("退款 %.2f 元（手续费率 %.0f%%）", refund, tkRate*100), orderID)

	now := time.Now().Format("2006-01-02 15:04:05")
	logContent := fmt.Sprintf("永夜运动退款：账号%s 退还%.2f", order.User, refund)
	database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'yongye_refund', ?, ?, ?)",
		order.UID, refund, logContent, now)

	return fmt.Sprintf("退款成功，退还 %.2f 元", refund), nil
}
