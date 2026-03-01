package service

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"go-api/internal/database"
)

//go:embed ydsj_schools.json
var ydsjSchoolsJSON []byte

// ---------- 数据结构 ----------

type YDSJConfig struct {
	BaseURL          string  `json:"base_url"`           // 上游API地址 如 http://103.149.27.248:5000
	Token            string  `json:"token"`              // 上游 Bearer Token
	UID              string  `json:"uid"`                // (旧字段，保留兼容)
	Key              string  `json:"key"`                // (旧字段，保留兼容)
	PriceMultiple    float64 `json:"price_multiple"`     // 运动世界价格倍率
	XbdMorningPrice  float64 `json:"xbd_morning_price"`  // 小步点晨跑价格
	XbdExercisePrice float64 `json:"xbd_exercise_price"` // 小步点课外跑价格
	RealCostMultiple float64 `json:"real_cost_multiple"` // 实际扣费倍率
}

type YDSJOrder struct {
	ID          int    `json:"id"`
	YID         string `json:"yid"`
	UID         int    `json:"uid"`
	School      string `json:"school"`
	User        string `json:"user"`
	Pass        string `json:"pass"`
	Distance    string `json:"distance"`
	IsRun       int    `json:"is_run"`
	RunType     int    `json:"run_type"`
	StartHour   string `json:"start_hour"`
	StartMinute string `json:"start_minute"`
	EndHour     string `json:"end_hour"`
	EndMinute   string `json:"end_minute"`
	RunWeek     string `json:"run_week"`
	Status      int    `json:"status"`
	Remarks     string `json:"remarks"`
	Info        string `json:"info"`
	TmpInfo     string `json:"tmp_info"`
	Fees        string `json:"fees"`
	RealFees    string `json:"real_fees"`
	RefundMoney string `json:"refund_money"`
	Addtime     string `json:"addtime"`
}

type YDSJService struct {
	client *http.Client
}

func NewYDSJService() *YDSJService {
	return &YDSJService{client: &http.Client{Timeout: 30 * time.Second}}
}

// EnsureTable 确保运动世界订单表存在
func (s *YDSJService) EnsureTable() {
	_, err := database.DB.Exec(`CREATE TABLE IF NOT EXISTS qingka_wangke_hzw_ydsj (
		id int NOT NULL AUTO_INCREMENT,
		uid int NOT NULL DEFAULT 1 COMMENT '用户UID',
		school varchar(255) NOT NULL DEFAULT '' COMMENT '学校',
		user varchar(255) NOT NULL DEFAULT '' COMMENT '用户账号',
		pass varchar(255) NOT NULL DEFAULT '' COMMENT '用户密码',
		distance varchar(255) NOT NULL DEFAULT '' COMMENT '总公里数',
		is_run int NOT NULL DEFAULT 1 COMMENT '跑步状态 0关闭 1开启',
		run_type int NOT NULL DEFAULT 0 COMMENT '跑步类型',
		start_hour varchar(255) NOT NULL DEFAULT '' COMMENT '开始小时',
		start_minute varchar(255) NOT NULL DEFAULT '' COMMENT '开始分钟',
		end_hour varchar(255) NOT NULL DEFAULT '' COMMENT '结束小时',
		end_minute varchar(255) NOT NULL DEFAULT '' COMMENT '结束分钟',
		run_week varchar(255) NOT NULL DEFAULT '' COMMENT '跑步周期',
		status int NOT NULL DEFAULT 1 COMMENT '1等待 2成功 3失败 4退款',
		remarks varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
		fees varchar(255) NOT NULL DEFAULT '' COMMENT '预扣金额',
		real_fees varchar(255) NOT NULL DEFAULT '' COMMENT '实际金额',
		addtime varchar(255) NOT NULL DEFAULT '' COMMENT '下单时间',
		yid varchar(255) NOT NULL DEFAULT '' COMMENT '上游订单ID',
		info text COMMENT '订单信息',
		tmp_info text COMMENT '操作信息',
		refund_money varchar(255) NOT NULL DEFAULT '' COMMENT '退款金额',
		PRIMARY KEY (id),
		KEY idx_uid (uid),
		KEY idx_status (status),
		KEY idx_user (user(191))
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='运动世界订单表'`)
	if err != nil {
		fmt.Printf("[YDSJ] 建表失败: %v\n", err)
	}

	// 补列：表可能由旧版 init_db.sql 创建，缺少以下列
	patchCols := []struct{ name, ddl string }{
		{"school", "ADD COLUMN `school` varchar(255) NOT NULL DEFAULT '' COMMENT '学校' AFTER `uid`"},
		{"distance", "ADD COLUMN `distance` varchar(255) NOT NULL DEFAULT '' COMMENT '总公里数' AFTER `pass`"},
		{"start_hour", "ADD COLUMN `start_hour` varchar(255) NOT NULL DEFAULT '' COMMENT '开始小时' AFTER `run_type`"},
		{"start_minute", "ADD COLUMN `start_minute` varchar(255) NOT NULL DEFAULT '' COMMENT '开始分钟' AFTER `start_hour`"},
		{"end_hour", "ADD COLUMN `end_hour` varchar(255) NOT NULL DEFAULT '' COMMENT '结束小时' AFTER `start_minute`"},
		{"end_minute", "ADD COLUMN `end_minute` varchar(255) NOT NULL DEFAULT '' COMMENT '结束分钟' AFTER `end_hour`"},
		{"run_week", "ADD COLUMN `run_week` varchar(255) NOT NULL DEFAULT '' COMMENT '跑步周期' AFTER `end_minute`"},
	}
	for _, col := range patchCols {
		var cnt int
		database.DB.QueryRow("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_wangke_hzw_ydsj' AND COLUMN_NAME=?", col.name).Scan(&cnt)
		if cnt == 0 {
			_, e := database.DB.Exec("ALTER TABLE `qingka_wangke_hzw_ydsj` " + col.ddl)
			if e != nil {
				fmt.Printf("[YDSJ] 补列 %s 失败: %v\n", col.name, e)
			} else {
				fmt.Printf("[YDSJ] 补列 %s 成功\n", col.name)
			}
		}
	}
}

// ---------- 上游 HTTP 工具 ----------

// ydsjRequest 向上游发起带 Bearer Token 的请求
func (s *YDSJService) ydsjRequest(method, urlPath string, body interface{}) ([]byte, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" || cfg.Token == "" {
		return nil, fmt.Errorf("运动世界未配置上游接口或Token")
	}
	return s.ydsjRequestWithCfg(cfg, method, urlPath, body)
}

func (s *YDSJService) ydsjRequestWithCfg(cfg *YDSJConfig, method, urlPath string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewReader(data)
	}

	fullURL := strings.TrimRight(cfg.BaseURL, "/") + urlPath
	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("上游请求失败: %v", err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// ---------- 学校列表 ----------

func (s *YDSJService) GetSchools() ([]map[string]interface{}, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return nil, err
	}

	// 优先从上游 API 获取
	if cfg.BaseURL != "" && cfg.Token != "" {
		respBody, err := s.ydsjRequestWithCfg(cfg, "GET", "/user/getSchool", nil)
		if err == nil {
			var schools []map[string]interface{}
			if json.Unmarshal(respBody, &schools) == nil && len(schools) > 0 {
				return schools, nil
			}
		} else {
			log.Printf("[YDSJ] 上游学校列表请求失败: %v，回退本地", err)
		}
	}

	// 回退: 读取本地嵌入 JSON
	var wrapper struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(ydsjSchoolsJSON, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Data, nil
}

// ---------- 配置 ----------

func (s *YDSJService) GetConfig() (*YDSJConfig, error) {
	var val string
	err := database.DB.QueryRow("SELECT svalue FROM qingka_wangke_config WHERE skey = 'ydsj_config' LIMIT 1").Scan(&val)
	if err != nil {
		return &YDSJConfig{PriceMultiple: 5, XbdMorningPrice: 6, XbdExercisePrice: 6.5, RealCostMultiple: 1}, nil
	}
	var cfg YDSJConfig
	json.Unmarshal([]byte(val), &cfg)
	return &cfg, nil
}

func (s *YDSJService) SaveConfig(cfg *YDSJConfig) error {
	data, _ := json.Marshal(cfg)
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_config (skey, svalue) VALUES ('ydsj_config', ?) ON DUPLICATE KEY UPDATE svalue = ?",
		string(data), string(data),
	)
	return err
}

// ---------- 价格 ----------

// runType: 0=运动世界晨跑, 1=运动世界课外跑, 2=小步点课外跑, 3=小步点晨跑
func (s *YDSJService) GetPrice(uid int, runType int, distance float64) (float64, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return 0, err
	}
	var rate float64 = 1.0
	database.DB.QueryRow("SELECT rate FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&rate)
	if rate <= 0 {
		rate = 1.0
	}

	var price float64
	switch runType {
	case 0, 1: // 运动世界
		price = distance * cfg.PriceMultiple * rate
	case 2: // 小步点课外跑
		price = distance * cfg.XbdExercisePrice * rate
	case 3: // 小步点晨跑
		price = distance * cfg.XbdMorningPrice * rate
	default:
		price = distance * cfg.PriceMultiple * rate
	}

	price = math.Round(price*100) / 100
	return price, nil
}

// ---------- 订单列表 ----------

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
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_hzw_ydsj "+where, args...).Scan(&total)

	offset := (page - 1) * limit
	query := "SELECT id, yid, uid, school, user, pass, distance, is_run, run_type, start_hour, start_minute, end_hour, end_minute, run_week, status, remarks, COALESCE(info,''), COALESCE(tmp_info,''), fees, COALESCE(real_fees,''), COALESCE(refund_money,''), addtime FROM qingka_wangke_hzw_ydsj " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
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

// weekNumToNames 将 "1,2,3,4,5" 转换为 ["周一","周二",...]
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

// ---------- 下单 ----------

func (s *YDSJService) AddOrder(uid int, form map[string]interface{}) (string, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" || cfg.Token == "" {
		return "", fmt.Errorf("运动世界未配置上游接口或Token")
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

	// 计算价格
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

	// 上游下单（新API格式）
	upstreamBody := map[string]interface{}{
		"xh":          user,
		"password":    pass,
		"schoolName":  school,
		"when":        weekNumToNames(runWeek),
		"km":          distance,
		"startHour":   startHour,
		"startMinute": startMinute,
		"endHour":     endHour,
		"endMinute":   endMinute,
		"bz":          "",
		"runType":     runType,
		"reservation": 0,
	}

	respBody, err := s.ydsjRequestWithCfg(cfg, "POST", "/order/submitOrder", upstreamBody)
	if err != nil {
		return "", fmt.Errorf("上游请求失败: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	success, _ := result["success"].(bool)
	msg := mapGetString(result, "msg")
	if !success {
		if msg == "" {
			msg = "上游下单失败"
		}
		return "", fmt.Errorf("%s", msg)
	}

	// 从上游响应中提取订单ID
	yid := ""
	if data, ok := result["data"].(map[string]interface{}); ok {
		if oid, ok := data["orderid"]; ok {
			yid = fmt.Sprintf("%v", oid)
		}
	}

	// 扣费
	database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", totalFee, uid)

	// 插入订单
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = database.DB.Exec(
		"INSERT INTO qingka_wangke_hzw_ydsj (yid, uid, school, user, pass, distance, is_run, run_type, start_hour, start_minute, end_hour, end_minute, run_week, status, remarks, info, tmp_info, fees, real_fees, refund_money, addtime) VALUES (?,?,?,?,?,?,1,?,?,?,?,?,?,1,'','','',?,'','',?)",
		yid, uid, school, user, pass, distance, runType, startHour, startMinute, endHour, endMinute, runWeek, fmt.Sprintf("%.2f", totalFee), now,
	)
	if err != nil {
		return "", fmt.Errorf("本地保存失败: %v", err)
	}

	// 记录日志
	logContent := fmt.Sprintf("运动世界下单：账号%s %.1fKM 扣费%.2f", user, dist, totalFee)
	database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'ydsj_add', ?, ?, ?)",
		uid, -totalFee, logContent, now)

	return fmt.Sprintf("下单成功，扣费 %.2f 元", totalFee), nil
}

// ---------- 退款 ----------

func (s *YDSJService) RefundOrder(uid, id int, isAdmin bool) (string, error) {
	var order YDSJOrder
	err := database.DB.QueryRow("SELECT id, uid, user, fees, status FROM qingka_wangke_hzw_ydsj WHERE id = ?", id).
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

// ---------- 切换跑步状态 ----------

func (s *YDSJService) ToggleRun(uid, id int, isAdmin bool) (string, error) {
	var orderUID, isRun int
	err := database.DB.QueryRow("SELECT uid, is_run FROM qingka_wangke_hzw_ydsj WHERE id = ?", id).Scan(&orderUID, &isRun)
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
