package appui

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-api/internal/database"
)

type AppuiConfig struct {
	BaseURL        string        `json:"base_url"`
	UID            string        `json:"uid"`
	Key            string        `json:"key"`
	PriceIncrement float64       `json:"price_increment"`
	Courses        []AppuiCourse `json:"courses"`
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

type appuiModuleService struct {
	client *http.Client
}

var appuiService = &appuiModuleService{client: &http.Client{Timeout: 30 * time.Second}}

func (s *appuiModuleService) GetConfig() (*AppuiConfig, error) {
	var val string
	err := database.DB.QueryRow("SELECT svalue FROM qingka_wangke_config WHERE skey = 'appui_config' LIMIT 1").Scan(&val)
	if err != nil {
		return &AppuiConfig{Courses: defaultAppuiCourses()}, nil
	}

	var cfg AppuiConfig
	_ = json.Unmarshal([]byte(val), &cfg)
	if len(cfg.Courses) == 0 {
		cfg.Courses = defaultAppuiCourses()
	}
	return &cfg, nil
}

func (s *appuiModuleService) SaveConfig(cfg *AppuiConfig) error {
	data, _ := json.Marshal(cfg)
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_config (v, k, skey, svalue) VALUES ('appui_config', '', 'appui_config', ?) ON DUPLICATE KEY UPDATE svalue = ?",
		string(data), string(data),
	)
	return err
}

func (s *appuiModuleService) GetPrice(uid int, pid string, days int) (float64, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return 0, err
	}

	rate := 1.0
	database.DB.QueryRow("SELECT rate FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&rate)
	if rate <= 0 {
		rate = 1.0
	}

	basePrice := 0.0
	for _, course := range cfg.Courses {
		if course.PID == pid {
			fmt.Sscanf(course.Price, "%f", &basePrice)
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

func (s *appuiModuleService) ListOrders(uid int, isAdmin bool, page, limit int, searchType, keyword string) ([]AppuiOrder, int, error) {
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

	total := 0
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_appui "+where, args...).Scan(&total)

	offset := (page - 1) * limit
	query := "SELECT id, uid, yid, pid, user, pass, name, address, residue_day, total_day, status, week, report, shangban_time, xiaban_time, addtime FROM qingka_wangke_appui " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	queryArgs := append(args, limit, offset)

	rows, err := database.DB.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []AppuiOrder
	for rows.Next() {
		var order AppuiOrder
		rows.Scan(&order.ID, &order.UID, &order.YID, &order.PID, &order.User, &order.Pass, &order.Name, &order.Address,
			&order.ResidueDay, &order.TotalDay, &order.Status, &order.Week, &order.Report,
			&order.ShangbanTime, &order.XiabanTime, &order.Addtime)
		orders = append(orders, order)
	}
	if orders == nil {
		orders = []AppuiOrder{}
	}
	return orders, total, nil
}

func (s *appuiModuleService) AddOrder(uid int, form map[string]interface{}) (string, error) {
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

	if pid == "" || user == "" || pass == "" || days < 1 {
		return "", fmt.Errorf("参数不完整")
	}

	price, err := s.GetPrice(uid, pid, days)
	if err != nil {
		return "", err
	}

	balance := 0.0
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&balance)
	if balance < price {
		return "", fmt.Errorf("余额不足，需要 %.2f 元，当前余额 %.2f 元", price, balance)
	}

	exists := 0
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_appui WHERE user = ? AND pid = ? AND status NOT IN ('已退款','已完成')", user, pid).Scan(&exists)
	if exists > 0 {
		return "", fmt.Errorf("该账号已有进行中的订单")
	}

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

	formData := make(map[string]string, len(upstreamData))
	for key, value := range upstreamData {
		formData[key] = fmt.Sprintf("%v", value)
	}

	resp, err := httpPostForm(cfg.BaseURL+"/appui/api.php", formData, 30)
	if err != nil {
		return "", fmt.Errorf("上游请求失败: %v", err)
	}

	var result map[string]interface{}
	_ = json.Unmarshal(resp, &result)

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

	database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", price, uid)

	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = database.DB.Exec(
		"INSERT INTO qingka_wangke_appui (uid, yid, pid, user, pass, name, address, residue_day, total_day, status, week, report, shangban_time, xiaban_time, addtime) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
		uid, yid, pid, user, pass, userName, address, days, days, "待处理", week, report, shangbanTime, xiabanTime, now,
	)
	if err != nil {
		return "", fmt.Errorf("本地保存失败: %v", err)
	}

	logContent := fmt.Sprintf("Appui打卡下单：平台%s 账号%s %d天 扣费%.2f", pid, user, days, price)
	database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'appui_add', ?, ?, ?)",
		uid, -price, logContent, now)

	return fmt.Sprintf("下单成功，扣费 %.2f 元", price), nil
}

func (s *appuiModuleService) EditOrder(uid int, isAdmin bool, form map[string]interface{}) error {
	id := mapGetInt(form, "id")
	if id == 0 {
		return fmt.Errorf("订单ID不能为空")
	}

	orderUID := 0
	database.DB.QueryRow("SELECT uid FROM qingka_wangke_appui WHERE id = ?", id).Scan(&orderUID)
	if !isAdmin && orderUID != uid {
		return fmt.Errorf("无权操作")
	}

	var sets []string
	var args []interface{}

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

func (s *appuiModuleService) RenewOrder(uid int, isAdmin bool, id, days int) (string, error) {
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

	balance := 0.0
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&balance)
	if balance < price {
		return "", fmt.Errorf("余额不足，需要 %.2f 元", price)
	}

	database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", price, uid)
	database.DB.Exec("UPDATE qingka_wangke_appui SET residue_day = residue_day + ?, total_day = total_day + ? WHERE id = ?", days, days, id)

	now := time.Now().Format("2006-01-02 15:04:05")
	logContent := fmt.Sprintf("Appui打卡续费：账号%s %d天 扣费%.2f", order.User, days, price)
	database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'appui_renew', ?, ?, ?)",
		uid, -price, logContent, now)

	return fmt.Sprintf("续费成功，扣费 %.2f 元", price), nil
}

func (s *appuiModuleService) DeleteOrder(uid, id int, isAdmin bool) (string, error) {
	var order AppuiOrder
	err := database.DB.QueryRow("SELECT id, uid, user, residue_day, pid, status FROM qingka_wangke_appui WHERE id = ?", id).
		Scan(&order.ID, &order.UID, &order.User, &order.ResidueDay, &order.PID, &order.Status)
	if err != nil {
		return "", fmt.Errorf("订单不存在")
	}
	if !isAdmin && order.UID != uid {
		return "", fmt.Errorf("无权操作")
	}

	refund := 0.0
	if order.ResidueDay > 0 {
		rate := 1.0
		database.DB.QueryRow("SELECT rate FROM qingka_wangke_user WHERE uid = ?", order.UID).Scan(&rate)
		if rate <= 0 {
			rate = 1.0
		}

		basePrice := 0.0
		cfg, _ := s.GetConfig()
		for _, course := range cfg.Courses {
			if course.PID == order.PID {
				fmt.Sscanf(course.Price, "%f", &basePrice)
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

func mapGetString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func mapGetInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	if v, ok := m[key].(int); ok {
		return v
	}
	var n int
	if v, ok := m[key].(string); ok {
		fmt.Sscanf(v, "%d", &n)
	}
	return n
}

func mapGetFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	if v, ok := m[key].(int); ok {
		return float64(v)
	}
	var f float64
	if v, ok := m[key].(string); ok {
		fmt.Sscanf(v, "%f", &f)
	}
	return f
}

func httpPostForm(apiURL string, params map[string]string, timeoutSec int) ([]byte, error) {
	form := url.Values{}
	for key, value := range params {
		form.Set(key, value)
	}
	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}
	resp, err := client.PostForm(apiURL, form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
