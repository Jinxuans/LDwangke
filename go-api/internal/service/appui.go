package service

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"go-api/internal/database"
)

// ---------- 数据结构 ----------

type AppuiConfig struct {
	BaseURL        string        `json:"base_url"`        // 上游API地址
	UID            string        `json:"uid"`             // 上游用户UID
	Key            string        `json:"key"`             // 上游密钥
	PriceIncrement float64       `json:"price_increment"` // 加价
	Courses        []AppuiCourse `json:"courses"`         // 平台/商品列表
}

type AppuiCourse struct {
	PID       string `json:"pid"`
	Name      string `json:"name"`
	Content   string `json:"content"`
	Price     string `json:"price"`
	YesSchool int    `json:"yes_school"`
}

type AppuiOrder struct {
	ID           int    `json:"id"`
	UID          int    `json:"uid"`
	YID          string `json:"yid"`
	PID          string `json:"pid"`
	User         string `json:"user"`
	Pass         string `json:"pass"`
	Name         string `json:"name"`
	Address      string `json:"address"`
	ResidueDay   int    `json:"residue_day"`
	TotalDay     int    `json:"total_day"`
	Status       string `json:"status"`
	Week         string `json:"week"`
	Report       string `json:"report"`
	ShangbanTime string `json:"shangban_time"`
	XiabanTime   string `json:"xiaban_time"`
	Addtime      string `json:"addtime"`
}

type AppuiService struct {
	client *http.Client
}

func NewAppuiService() *AppuiService {
	return &AppuiService{client: &http.Client{Timeout: 30 * time.Second}}
}

// ---------- 配置 ----------

func (s *AppuiService) GetConfig() (*AppuiConfig, error) {
	var val string
	err := database.DB.QueryRow("SELECT svalue FROM qingka_wangke_config WHERE skey = 'appui_config' LIMIT 1").Scan(&val)
	if err != nil {
		return &AppuiConfig{
			Courses: defaultAppuiCourses(),
		}, nil
	}
	var cfg AppuiConfig
	json.Unmarshal([]byte(val), &cfg)
	if len(cfg.Courses) == 0 {
		cfg.Courses = defaultAppuiCourses()
	}
	return &cfg, nil
}

func (s *AppuiService) SaveConfig(cfg *AppuiConfig) error {
	data, _ := json.Marshal(cfg)
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_config (skey, svalue) VALUES ('appui_config', ?) ON DUPLICATE KEY UPDATE svalue = ?",
		string(data), string(data),
	)
	return err
}

func defaultAppuiCourses() []AppuiCourse {
	return []AppuiCourse{
		{PID: "1", Name: "校友邦", Content: "支持上、下班打卡！周报！", Price: "0.10", YesSchool: 0},
		{PID: "2", Name: "职校家园", Content: "支持打卡，日报，周报，月报", Price: "0.10", YesSchool: 0},
		{PID: "3", Name: "慧职教", Content: "支持打卡、实习汇报！", Price: "0.10", YesSchool: 1},
		{PID: "4", Name: "黔职通", Content: "包上班打卡，下班打卡，日报、周报、月报！", Price: "0.10", YesSchool: 0},
		{PID: "5", Name: "学习通", Content: "学习通包上班打卡！包打卡，日报！周报！月报！", Price: "0.10", YesSchool: 0},
		{PID: "6", Name: "习行学生版", Content: "暂且只支持打卡！", Price: "0.10", YesSchool: 1},
		{PID: "7", Name: "工学云", Content: "包上，下班打卡，日报，周报，月报！", Price: "0.10", YesSchool: 0},
		{PID: "8", Name: "习讯云", Content: "包打卡！", Price: "0.10", YesSchool: 1},
		{PID: "9", Name: "广西职业院校公众号", Content: "包打卡", Price: "0.10", YesSchool: 0},
	}
}

// ---------- 价格 ----------

func (s *AppuiService) GetPrice(uid int, pid string, days int) (float64, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return 0, err
	}
	var rate float64 = 1.0
	database.DB.QueryRow("SELECT rate FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&rate)
	if rate <= 0 {
		rate = 1.0
	}

	var basePrice float64
	for _, c := range cfg.Courses {
		if c.PID == pid {
			fmt.Sscanf(c.Price, "%f", &basePrice)
			break
		}
	}
	if basePrice <= 0 {
		return 0, fmt.Errorf("未找到对应平台")
	}

	total := basePrice * float64(days) * rate
	total += cfg.PriceIncrement
	total = math.Round(total*100) / 100
	return total, nil
}

// ---------- 订单列表 ----------

func (s *AppuiService) ListOrders(uid int, isAdmin bool, page, limit int, searchType, keyword string) ([]AppuiOrder, int, error) {
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

	var total int
	countSQL := "SELECT COUNT(*) FROM qingka_wangke_appui " + where
	database.DB.QueryRow(countSQL, args...).Scan(&total)

	offset := (page - 1) * limit
	query := "SELECT id, uid, yid, pid, user, pass, name, address, residue_day, total_day, status, week, report, shangban_time, xiaban_time, addtime FROM qingka_wangke_appui " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []AppuiOrder
	for rows.Next() {
		var o AppuiOrder
		rows.Scan(&o.ID, &o.UID, &o.YID, &o.PID, &o.User, &o.Pass, &o.Name, &o.Address,
			&o.ResidueDay, &o.TotalDay, &o.Status, &o.Week, &o.Report,
			&o.ShangbanTime, &o.XiabanTime, &o.Addtime)
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []AppuiOrder{}
	}
	return orders, total, nil
}

// ---------- 下单 ----------

func (s *AppuiService) AddOrder(uid int, form map[string]interface{}) (string, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return "", fmt.Errorf("Appui打卡未配置上游接口")
	}

	pid := mapGetString(form, "pid")
	user := mapGetString(form, "user")
	pass := mapGetString(form, "pass")
	userName := mapGetString(form, "userName")
	address := mapGetString(form, "address")
	days := mapGetInt(form, "days")
	week := mapGetString(form, "week")
	report := mapGetString(form, "report")
	shangbanTime := mapGetString(form, "shangban_time")
	xiabanTime := mapGetString(form, "xiaban_time")
	school := mapGetString(form, "school")
	_ = school

	if pid == "" || user == "" || pass == "" || days < 1 {
		return "", fmt.Errorf("参数不完整")
	}

	// 计算价格并扣费
	price, err := s.GetPrice(uid, pid, days)
	if err != nil {
		return "", err
	}

	var balance float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&balance)
	if balance < price {
		return "", fmt.Errorf("余额不足，需要 %.2f 元，当前余额 %.2f 元", price, balance)
	}

	// 检查重复
	var exists int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_appui WHERE user = ? AND pid = ? AND status NOT IN ('已退款','已完成')", user, pid).Scan(&exists)
	if exists > 0 {
		return "", fmt.Errorf("该账号已有进行中的订单")
	}

	// 上游下单（通用代理）
	upstreamData := map[string]interface{}{
		"login_uid":     cfg.UID,
		"login_key":     cfg.Key,
		"act":           "add",
		"pid":           pid,
		"user":          user,
		"pass":          pass,
		"days":          days,
		"week":          week,
		"report":        report,
		"shangban_time": shangbanTime,
		"xiaban_time":   xiabanTime,
		"address":       address,
	}

	formData := make(map[string]string)
	for k, v := range upstreamData {
		formData[k] = fmt.Sprintf("%v", v)
	}
	resp, err := httpPostForm(cfg.BaseURL+"/appui/api.php", formData, 30)
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
		return "", fmt.Errorf("%s", msg)
	}

	yid := mapGetString(result, "yid")
	if yid == "" {
		if data, ok := result["data"].(map[string]interface{}); ok {
			yid = mapGetString(data, "id")
		}
	}

	// 扣费
	database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", price, uid)

	// 插入订单
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = database.DB.Exec(
		"INSERT INTO qingka_wangke_appui (uid, yid, pid, user, pass, name, address, residue_day, total_day, status, week, report, shangban_time, xiaban_time, addtime) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
		uid, yid, pid, user, pass, userName, address, days, days, "待处理", week, report, shangbanTime, xiabanTime, now,
	)
	if err != nil {
		return "", fmt.Errorf("本地保存失败: %v", err)
	}

	// 记录日志
	logContent := fmt.Sprintf("Appui打卡下单：平台%s 账号%s %d天 扣费%.2f", pid, user, days, price)
	database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'appui_add', ?, ?, ?)",
		uid, -price, logContent, now)

	return fmt.Sprintf("下单成功，扣费 %.2f 元", price), nil
}

// ---------- 编辑订单 ----------

func (s *AppuiService) EditOrder(uid int, isAdmin bool, form map[string]interface{}) error {
	id := mapGetInt(form, "id")
	if id == 0 {
		return fmt.Errorf("订单ID不能为空")
	}

	var orderUID int
	database.DB.QueryRow("SELECT uid FROM qingka_wangke_appui WHERE id = ?", id).Scan(&orderUID)
	if !isAdmin && orderUID != uid {
		return fmt.Errorf("无权操作")
	}

	sets := []string{}
	args := []interface{}{}

	if v, ok := form["pass"]; ok {
		sets = append(sets, "pass = ?")
		args = append(args, v)
	}
	if v, ok := form["address"]; ok {
		sets = append(sets, "address = ?")
		args = append(args, v)
	}
	if v, ok := form["week"]; ok {
		sets = append(sets, "week = ?")
		args = append(args, v)
	}
	if v, ok := form["report"]; ok {
		sets = append(sets, "report = ?")
		args = append(args, v)
	}
	if v, ok := form["shangban_time"]; ok {
		sets = append(sets, "shangban_time = ?")
		args = append(args, v)
	}
	if v, ok := form["xiaban_time"]; ok {
		sets = append(sets, "xiaban_time = ?")
		args = append(args, v)
	}

	if len(sets) == 0 {
		return fmt.Errorf("没有要修改的内容")
	}

	args = append(args, id)
	_, err := database.DB.Exec("UPDATE qingka_wangke_appui SET "+strings.Join(sets, ", ")+" WHERE id = ?", args...)
	return err
}

// ---------- 续费 ----------

func (s *AppuiService) RenewOrder(uid int, isAdmin bool, id, days int) (string, error) {
	var order AppuiOrder
	err := database.DB.QueryRow("SELECT id, uid, pid, user, status FROM qingka_wangke_appui WHERE id = ?", id).
		Scan(&order.ID, &order.UID, &order.PID, &order.User, &order.Status)
	if err != nil {
		return "", fmt.Errorf("订单不存在")
	}
	if !isAdmin && order.UID != uid {
		return "", fmt.Errorf("无权操作")
	}

	price, err := s.GetPrice(uid, order.PID, days)
	if err != nil {
		return "", err
	}

	var balance float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&balance)
	if balance < price {
		return "", fmt.Errorf("余额不足，需要 %.2f 元", price)
	}

	// 扣费并更新天数
	database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", price, uid)
	database.DB.Exec("UPDATE qingka_wangke_appui SET residue_day = residue_day + ?, total_day = total_day + ? WHERE id = ?", days, days, id)

	now := time.Now().Format("2006-01-02 15:04:05")
	logContent := fmt.Sprintf("Appui打卡续费：账号%s %d天 扣费%.2f", order.User, days, price)
	database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'appui_renew', ?, ?, ?)",
		uid, -price, logContent, now)

	return fmt.Sprintf("续费成功，扣费 %.2f 元", price), nil
}

// ---------- 删除/退款 ----------

func (s *AppuiService) DeleteOrder(uid, id int, isAdmin bool) (string, error) {
	var order AppuiOrder
	err := database.DB.QueryRow("SELECT id, uid, user, residue_day, pid, status FROM qingka_wangke_appui WHERE id = ?", id).
		Scan(&order.ID, &order.UID, &order.User, &order.ResidueDay, &order.PID, &order.Status)
	if err != nil {
		return "", fmt.Errorf("订单不存在")
	}
	if !isAdmin && order.UID != uid {
		return "", fmt.Errorf("无权操作")
	}

	// 计算退款（按剩余天数）
	refund := 0.0
	if order.ResidueDay > 0 {
		var rate float64 = 1.0
		database.DB.QueryRow("SELECT rate FROM qingka_wangke_user WHERE uid = ?", order.UID).Scan(&rate)
		if rate <= 0 {
			rate = 1.0
		}
		var basePrice float64
		cfg, _ := s.GetConfig()
		for _, c := range cfg.Courses {
			if c.PID == order.PID {
				fmt.Sscanf(c.Price, "%f", &basePrice)
				break
			}
		}
		refund = basePrice * float64(order.ResidueDay) * rate
		refund = math.Round(refund*100) / 100
	}

	if refund > 0 {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", refund, order.UID)
	}

	database.DB.Exec("UPDATE qingka_wangke_appui SET status = '已退款' WHERE id = ?", id)

	now := time.Now().Format("2006-01-02 15:04:05")
	logContent := fmt.Sprintf("Appui打卡退款：账号%s 退还%.2f", order.User, refund)
	database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'appui_refund', ?, ?, ?)",
		order.UID, refund, logContent, now)

	return fmt.Sprintf("退款成功，退还 %.2f 元", refund), nil
}
