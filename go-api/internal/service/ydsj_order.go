package service

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"go-api/internal/database"
)

// runType: 0=运动世界晨跑, 1=运动世界课外跑, 2=小步点课外跑, 3=小步点晨跑
func (s *YDSJService) GetPrice(uid int, runType int, distance float64) (float64, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return 0, err
	}
	rate := 1.0
	database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&rate)
	if rate <= 0 {
		rate = 1.0
	}

	var price float64
	switch runType {
	case 0, 1:
		price = distance * cfg.PriceMultiple * rate
	case 2:
		price = distance * cfg.XbdExercisePrice * rate
	case 3:
		price = distance * cfg.XbdMorningPrice * rate
	default:
		price = distance * cfg.PriceMultiple * rate
	}

	price = math.Round(price*100) / 100
	return price, nil
}

func (s *YDSJService) ListOrders(uid int, isAdmin bool, page, limit int, searchType, keyword, statusFilter string) ([]YDSJOrder, int, error) {
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
			where += " AND `user` LIKE ?"
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
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_hzw_ydsj "+where, args...).Scan(&total)

	offset := (page - 1) * limit
	query := "SELECT id, yid, uid, school, `user`, pass, distance, is_run, run_type, start_hour, start_minute, end_hour, end_minute, run_week, status, remarks, COALESCE(info,''), COALESCE(tmp_info,''), fees, COALESCE(real_fees,''), COALESCE(refund_money,''), addtime FROM qingka_wangke_hzw_ydsj " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Printf("[YDSJ] ListOrders 查询失败: %v | SQL: %s | args: %v", err, query, args)
		return nil, 0, err
	}
	defer rows.Close()

	var orders []YDSJOrder
	for rows.Next() {
		var o YDSJOrder
		rows.Scan(&o.ID, &o.YID, &o.UID, &o.School, &o.User, &o.Pass, &o.Distance,
			&o.IsRun, &o.RunType, &o.StartHour, &o.StartMinute, &o.EndHour, &o.EndMinute,
			&o.RunWeek, &o.Status, &o.Remarks, &o.Info, &o.TmpInfo,
			&o.Fees, &o.RealFees, &o.RefundMoney, &o.Addtime)
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []YDSJOrder{}
	}
	return orders, total, nil
}

func weekNumToNames(runWeek string) []string {
	nameMap := map[string]string{"1": "周一", "2": "周二", "3": "周三", "4": "周四", "5": "周五", "6": "周六", "7": "周日"}
	var names []string
	for _, w := range strings.Split(runWeek, ",") {
		w = strings.TrimSpace(w)
		if n, ok := nameMap[w]; ok {
			names = append(names, n)
		}
	}
	if len(names) == 0 {
		names = []string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"}
	}
	return names
}

func (s *YDSJService) AddOrder(uid int, form map[string]interface{}) (string, error) {
	cfg, err := s.GetConfig()
	if err != nil || !ydsjIsConfigured(cfg) {
		return "", fmt.Errorf("运动世界未配置上游接口")
	}

	school := mapGetString(form, "school")
	user := mapGetString(form, "user")
	pass := mapGetString(form, "pass")
	distance := mapGetString(form, "distance")
	runType := mapGetInt(form, "run_type")
	startHour := mapGetString(form, "start_hour")
	startMinute := mapGetString(form, "start_minute")
	endHour := mapGetString(form, "end_hour")
	endMinute := mapGetString(form, "end_minute")
	runWeek := mapGetString(form, "run_week")

	if user == "" || pass == "" || distance == "" {
		return "", fmt.Errorf("参数不完整")
	}

	var dist float64
	fmt.Sscanf(distance, "%f", &dist)
	totalFee, err := s.GetPrice(uid, runType, dist)
	if err != nil {
		return "", err
	}

	var balance float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&balance)
	if balance < totalFee {
		return "", fmt.Errorf("余额不足，需要 %.2f 元，当前余额 %.2f 元", totalFee, balance)
	}

	if school == "" {
		school = "自动识别"
	}

	upstreamParams := map[string]string{
		"school":       school,
		"user":         user,
		"pass":         pass,
		"distance":     distance,
		"run_type":     fmt.Sprintf("%d", runType),
		"start_hour":   startHour,
		"start_minute": startMinute,
		"end_hour":     endHour,
		"end_minute":   endMinute,
		"run_week":     runWeek,
		"remarks":      "",
	}

	respBody, err := s.ydsjRequestWithCfg(cfg, "add_order", upstreamParams)
	if err != nil {
		return "", fmt.Errorf("上游请求失败: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	code := mapGetFloat(result, "code")
	msg := mapGetString(result, "msg")
	if code != 1 {
		if msg == "" {
			msg = "上游下单失败"
		}
		return "", fmt.Errorf("%s", msg)
	}

	yid := ""
	if data, ok := result["data"].(map[string]interface{}); ok {
		if oid, ok := data["yid"]; ok {
			yid = fmt.Sprintf("%v", oid)
		}
	}

	database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", totalFee, uid)

	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = database.DB.Exec(
		"INSERT INTO qingka_wangke_hzw_ydsj (yid, uid, school, `user`, pass, distance, is_run, run_type, start_hour, start_minute, end_hour, end_minute, run_week, status, remarks, info, tmp_info, fees, real_fees, refund_money, addtime) VALUES (?,?,?,?,?,?,1,?,?,?,?,?,?,1,'','','',?,'','',?)",
		yid, uid, school, user, pass, distance, runType, startHour, startMinute, endHour, endMinute, runWeek, fmt.Sprintf("%.2f", totalFee), now,
	)
	if err != nil {
		return "", fmt.Errorf("本地保存失败: %v", err)
	}

	logContent := fmt.Sprintf("运动世界下单：账号%s %.1fKM 扣费%.2f", user, dist, totalFee)
	database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'ydsj_add', ?, ?, ?)",
		uid, -totalFee, logContent, now)

	return fmt.Sprintf("下单成功，扣费 %.2f 元", totalFee), nil
}

func (s *YDSJService) RefundOrder(uid, id int, isAdmin bool) (string, error) {
	var order YDSJOrder
	err := database.DB.QueryRow("SELECT id, uid, `user`, fees, status FROM qingka_wangke_hzw_ydsj WHERE id = ?", id).
		Scan(&order.ID, &order.UID, &order.User, &order.Fees, &order.Status)
	if err != nil {
		return "", fmt.Errorf("订单不存在")
	}
	if !isAdmin && order.UID != uid {
		return "", fmt.Errorf("无权操作")
	}
	if order.Status == 4 {
		return "", fmt.Errorf("该订单已退款")
	}

	var refund float64
	fmt.Sscanf(order.Fees, "%f", &refund)
	if refund > 0 {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", refund, order.UID)
	}

	database.DB.Exec("UPDATE qingka_wangke_hzw_ydsj SET status = 4, refund_money = ? WHERE id = ?", fmt.Sprintf("%.2f", refund), id)

	now := time.Now().Format("2006-01-02 15:04:05")
	logContent := fmt.Sprintf("运动世界退款：账号%s 退还%.2f", order.User, refund)
	database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'ydsj_refund', ?, ?, ?)",
		order.UID, refund, logContent, now)

	return fmt.Sprintf("退款成功，退还 %.2f 元", refund), nil
}

func (s *YDSJService) EditRemarks(uid, id int, remarks string, isAdmin bool) (string, error) {
	var orderUID int
	err := database.DB.QueryRow("SELECT uid FROM qingka_wangke_hzw_ydsj WHERE id = ?", id).Scan(&orderUID)
	if err != nil {
		return "", fmt.Errorf("订单不存在")
	}
	if !isAdmin && orderUID != uid {
		return "", fmt.Errorf("无权操作")
	}
	_, err = database.DB.Exec("UPDATE qingka_wangke_hzw_ydsj SET remarks = ? WHERE id = ?", remarks, id)
	if err != nil {
		return "", fmt.Errorf("修改失败")
	}
	return "备注修改成功", nil
}

func (s *YDSJService) SyncOrder(uid, id int, isAdmin bool) (map[string]interface{}, error) {
	var orderUID int
	var yid, user string
	var runType int
	err := database.DB.QueryRow("SELECT uid, yid, `user`, run_type FROM qingka_wangke_hzw_ydsj WHERE id = ?", id).
		Scan(&orderUID, &yid, &user, &runType)
	if err != nil {
		return nil, fmt.Errorf("订单不存在")
	}
	if !isAdmin && orderUID != uid {
		return nil, fmt.Errorf("无权操作")
	}
	if yid == "" {
		return nil, fmt.Errorf("该订单尚未提交到上游，无法同步")
	}

	cfg, err := s.GetConfig()
	if err != nil || !ydsjIsConfigured(cfg) {
		return nil, fmt.Errorf("运动世界未配置上游接口")
	}

	items, err := ydsjUpstreamQuery(cfg, user, runType)
	if err != nil || len(items) == 0 {
		return nil, fmt.Errorf("上游查询失败或无数据")
	}

	var matched map[string]interface{}
	for _, item := range items {
		oid := fmt.Sprintf("%v", item["orderid"])
		if oid == yid {
			matched = item
			break
		}
	}
	if matched == nil {
		return nil, fmt.Errorf("上游未找到匹配订单")
	}

	statusStr := mapGetString(matched, "status")
	newStatus := ydsjMapUpstreamStatus(statusStr)
	remarks := mapGetString(matched, "bz")

	database.DB.Exec("UPDATE qingka_wangke_hzw_ydsj SET status = ?, remarks = ? WHERE id = ?", newStatus, remarks, id)

	return map[string]interface{}{
		"id":         id,
		"status":     newStatus,
		"status_str": statusStr,
		"remarks":    remarks,
	}, nil
}

func (s *YDSJService) ToggleRun(uid, id int, isAdmin bool) (string, error) {
	var orderUID, isRun int
	err := database.DB.QueryRow("SELECT uid, is_run FROM qingka_wangke_hzw_ydsj WHERE id = ?", id).Scan(&orderUID, &isRun)
	log.Printf("[YDSJ] ToggleRun id=%d uid=%d is_run=%d err=%v", id, orderUID, isRun, err)
	if err != nil {
		return "", fmt.Errorf("订单不存在")
	}
	if !isAdmin && orderUID != uid {
		return "", fmt.Errorf("无权操作")
	}

	newVal := 0
	msg := "已暂停"
	if isRun == 0 {
		newVal = 1
		msg = "已开启"
	}
	database.DB.Exec("UPDATE qingka_wangke_hzw_ydsj SET is_run = ? WHERE id = ?", newVal, id)
	return msg, nil
}
