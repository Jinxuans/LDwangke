package service

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"go-api/internal/database"
)

// ---------- 数据结构 ----------

type SDXYConfig struct {
	BaseURL    string  `json:"base_url"`     // 上游API地址
	UID        string  `json:"uid"`          // 上游用户UID
	Key        string  `json:"key"`          // 上游密钥
	PricePerKM float64 `json:"price_per_km"` // 每公里价格系数
}

type SDXYOrder struct {
	ID          int    `json:"id"`
	YID         string `json:"yid"`
	UID         int    `json:"uid"`
	User        string `json:"user"`
	Pass        string `json:"pass"`
	School      string `json:"school"`
	Distance    string `json:"distance"`
	Day         string `json:"day"`
	StartDate   string `json:"start_date"`
	StartHour   string `json:"start_hour"`
	StartMinute string `json:"start_minute"`
	EndHour     string `json:"end_hour"`
	EndMinute   string `json:"end_minute"`
	RunWeek     string `json:"run_week"`
	Status      int    `json:"status"`
	Remarks     string `json:"remarks"`
	Fees        string `json:"fees"`
	Addtime     string `json:"addtime"`
}

type SDXYService struct {
	client *http.Client
}

func NewSDXYService() *SDXYService {
	return &SDXYService{client: &http.Client{Timeout: 30 * time.Second}}
}

// ---------- 配置 ----------

func (s *SDXYService) GetConfig() (*SDXYConfig, error) {
	var val string
	err := database.DB.QueryRow("SELECT svalue FROM qingka_wangke_config WHERE skey = 'sdxy_config' LIMIT 1").Scan(&val)
	if err != nil {
		return &SDXYConfig{PricePerKM: 10}, nil
	}
	var cfg SDXYConfig
	json.Unmarshal([]byte(val), &cfg)
	return &cfg, nil
}

func (s *SDXYService) SaveConfig(cfg *SDXYConfig) error {
	data, _ := json.Marshal(cfg)
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_config (skey, svalue) VALUES ('sdxy_config', ?) ON DUPLICATE KEY UPDATE svalue = ?",
		string(data), string(data),
	)
	return err
}

// ---------- 价格 ----------

func (s *SDXYService) GetPrice(uid int) (float64, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return 0, err
	}
	var rate float64 = 1.0
	database.DB.QueryRow("SELECT rate FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&rate)
	if rate <= 0 {
		rate = 1.0
	}
	price := cfg.PricePerKM * rate
	price = math.Round(price*100) / 100
	return price, nil
}

// ---------- 订单列表 ----------

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
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_hzw_sdxy "+where, args...).Scan(&total)

	offset := (page - 1) * limit
	query := "SELECT id, yid, uid, user, pass, school, distance, day, start_date, start_hour, start_minute, end_hour, end_minute, run_week, status, remarks, fees, addtime FROM qingka_wangke_hzw_sdxy " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []SDXYOrder
	for rows.Next() {
		var o SDXYOrder
		rows.Scan(&o.ID, &o.YID, &o.UID, &o.User, &o.Pass, &o.School, &o.Distance, &o.Day,
			&o.StartDate, &o.StartHour, &o.StartMinute, &o.EndHour, &o.EndMinute,
			&o.RunWeek, &o.Status, &o.Remarks, &o.Fees, &o.Addtime)
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []SDXYOrder{}
	}
	return orders, total, nil
}

// ---------- 下单 ----------

func (s *SDXYService) AddOrder(uid int, form map[string]interface{}) (string, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return "", fmt.Errorf("闪电运动未配置上游接口")
	}

	user := mapGetString(form, "user")
	pass := mapGetString(form, "pass")
	school := mapGetString(form, "school")
	distance := mapGetString(form, "distance")
	day := mapGetString(form, "day")
	startDate := mapGetString(form, "start_date")
	startHour := mapGetString(form, "start_hour")
	startMinute := mapGetString(form, "start_minute")
	endHour := mapGetString(form, "end_hour")
	endMinute := mapGetString(form, "end_minute")
	runWeek := mapGetString(form, "run_week")

	if user == "" || pass == "" || day == "" {
		return "", fmt.Errorf("参数不完整")
	}

	// 计算价格
	pricePerKM, err := s.GetPrice(uid)
	if err != nil {
		return "", err
	}
	var dist float64
	fmt.Sscanf(distance, "%f", &dist)
	var dayNum int
	fmt.Sscanf(day, "%d", &dayNum)
	totalFee := pricePerKM * dist * float64(dayNum)
	totalFee = math.Round(totalFee*100) / 100

	var balance float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&balance)
	if balance < totalFee {
		return "", fmt.Errorf("余额不足，需要 %.2f 元，当前余额 %.2f 元", totalFee, balance)
	}

	// 上游下单
	formData := map[string]string{
		"login_uid":    cfg.UID,
		"login_key":    cfg.Key,
		"act":          "add",
		"user":         user,
		"pass":         pass,
		"school":       school,
		"distance":     distance,
		"day":          day,
		"start_date":   startDate,
		"start_hour":   startHour,
		"start_minute": startMinute,
		"end_hour":     endHour,
		"end_minute":   endMinute,
		"run_week":     runWeek,
	}

	resp, err := httpPostForm(cfg.BaseURL+"/sdxy/api.php", formData, 30)
	if err != nil {
		return "", fmt.Errorf("上游请求失败: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(resp, &result)

	code := mapGetFloat(result, "code")
	if code != 0 {
		msg := mapGetString(result, "msg")
		if msg == "" {
			msg = "上游下单失败"
		}
		return "", fmt.Errorf(msg)
	}

	yid := mapGetString(result, "yid")

	// 扣费
	database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", totalFee, uid)

	// 插入订单
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = database.DB.Exec(
		"INSERT INTO qingka_wangke_hzw_sdxy (yid, uid, user, pass, school, distance, day, start_date, start_hour, start_minute, end_hour, end_minute, run_week, status, remarks, fees, addtime) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,1,?,?,?)",
		yid, uid, user, pass, school, distance, day, startDate, startHour, startMinute, endHour, endMinute, runWeek, "", fmt.Sprintf("%.2f", totalFee), now,
	)
	if err != nil {
		return "", fmt.Errorf("本地保存失败: %v", err)
	}

	// 记录日志
	logContent := fmt.Sprintf("闪电运动下单：账号%s %s天 %.1fKM/天 扣费%.2f", user, day, dist, totalFee)
	database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'sdxy_add', ?, ?, ?)",
		uid, -totalFee, logContent, now)

	return fmt.Sprintf("下单成功，扣费 %.2f 元", totalFee), nil
}

// ---------- 删除/退款 ----------

func (s *SDXYService) DeleteOrder(uid, id int, isAdmin bool) (string, error) {
	var order SDXYOrder
	err := database.DB.QueryRow("SELECT id, uid, user, fees, status FROM qingka_wangke_hzw_sdxy WHERE id = ?", id).
		Scan(&order.ID, &order.UID, &order.User, &order.Fees, &order.Status)
	if err != nil {
		return "", fmt.Errorf("订单不存在")
	}
	if !isAdmin && order.UID != uid {
		return "", fmt.Errorf("无权操作")
	}
	if order.Status == 3 {
		return "", fmt.Errorf("该订单已退款")
	}

	// 退款
	var refund float64
	fmt.Sscanf(order.Fees, "%f", &refund)
	if refund > 0 {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", refund, order.UID)
	}

	database.DB.Exec("UPDATE qingka_wangke_hzw_sdxy SET status = 3, remarks = '已退款' WHERE id = ?", id)

	now := time.Now().Format("2006-01-02 15:04:05")
	logContent := fmt.Sprintf("闪电运动退款：账号%s 退还%.2f", order.User, refund)
	database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'sdxy_refund', ?, ?, ?)",
		order.UID, refund, logContent, now)

	return fmt.Sprintf("退款成功，退还 %.2f 元", refund), nil
}
