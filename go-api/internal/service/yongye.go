package service

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"io"
	"strings"
	"time"

	"go-api/internal/database"
)

// ---------- 数据结构 ----------

type YongyeConfig struct {
	ApiURL  string  `json:"api_url"`  // 上游API地址 如 https://yy.rgrg.cc/api
	Token   string  `json:"token"`    // 上游token
	Dj      float64 `json:"dj"`      // 等级(最高级) 如0.2
	Zs      float64 `json:"zs"`      // 赠送倍率 如1.25
	Beis    float64 `json:"beis"`    // 学校价格倍数 如1.3
	Xzdj    float64 `json:"xzdj"`   // 限制等级
	Xzmo    float64 `json:"xzmo"`   // 限制余额
	Tk      float64 `json:"tk"`     // 退款手续费率 如0.01=1%
	Content string  `json:"content"` // 下单页说明
	Tcgg    string  `json:"tcgg"`   // 弹窗公告
}

type YongyeOrder struct {
	ID         int     `json:"id"`
	Pol        int     `json:"pol"`
	UID        int     `json:"uid"`
	User       string  `json:"user"`
	Pass       string  `json:"pass"`
	School     string  `json:"school"`
	Type       int     `json:"type"`
	Zkm        float64 `json:"zkm"`
	KsH        int     `json:"ks_h"`
	KsM        int     `json:"ks_m"`
	JsH        int     `json:"js_h"`
	JsM        int     `json:"js_m"`
	Weeks      string  `json:"weeks"`
	DockStatus int     `json:"dockstatus"`
	Yfees      float64 `json:"yfees"`
	Fees       float64 `json:"fees"`
	YID        string  `json:"yid"`
	Yaddtime   string  `json:"yaddtime"`
	Addtime    string  `json:"addtime"`
	Tktext     string  `json:"tktext"`
}

type YongyeStudent struct {
	ID       int     `json:"id"`
	UID      int     `json:"uid"`
	User     string  `json:"user"`
	Pass     string  `json:"pass"`
	Type     int     `json:"type"`
	Zkm      float64 `json:"zkm"`
	Weeks    string  `json:"weeks"`
	Status   int     `json:"status"`
	Tdkm     float64 `json:"tdkm"`
	Tdmoney  float64 `json:"tdmoney"`
	Stulog   string  `json:"stulog"`
	LastTime string  `json:"last_time"`
}

type YongyeSchool struct {
	Name   string  `json:"name"`
	Cpmuch float64 `json:"cpmuch"` // 晨跑价格
	Zcmuch float64 `json:"zcmuch"` // 正常跑价格
}

type YongyeService struct {
	client *http.Client
}

func NewYongyeService() *YongyeService {
	return &YongyeService{client: &http.Client{Timeout: 15 * time.Second}}
}

// EnsureTable 确保永夜运动表存在
func (s *YongyeService) EnsureTable() {
	_, err := database.DB.Exec(`CREATE TABLE IF NOT EXISTS yy_ydsj_dd (
		id INT NOT NULL AUTO_INCREMENT,
		pol TINYINT NOT NULL DEFAULT 0 COMMENT '轮询模式 0=否 1=是',
		uid INT NOT NULL DEFAULT 0,
		user VARCHAR(100) NOT NULL DEFAULT '' COMMENT '学号',
		pass VARCHAR(100) NOT NULL DEFAULT '' COMMENT '密码',
		school VARCHAR(100) NOT NULL DEFAULT '自动识别' COMMENT '学校',
		type TINYINT NOT NULL DEFAULT 0 COMMENT '跑步类型 0=正常 1=晨跑',
		zkm DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '公里数',
		ks_h INT NOT NULL DEFAULT 9 COMMENT '开始小时',
		ks_m INT NOT NULL DEFAULT 0 COMMENT '开始分钟',
		js_h INT NOT NULL DEFAULT 21 COMMENT '结束小时',
		js_m INT NOT NULL DEFAULT 0 COMMENT '结束分钟',
		weeks VARCHAR(20) NOT NULL DEFAULT '' COMMENT '跑步周天',
		dockstatus TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0=未提交 1=已提交 2=失败 3=关闭 5=轮询',
		yfees DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '预扣费用',
		fees DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '实际费用',
		yid VARCHAR(50) NOT NULL DEFAULT '' COMMENT '上游订单ID',
		yaddtime VARCHAR(50) NOT NULL DEFAULT '',
		addtime DATETIME DEFAULT NULL,
		tktext TEXT COMMENT '状态日志',
		PRIMARY KEY (id),
		KEY idx_uid (uid),
		KEY idx_user (user),
		KEY idx_dockstatus (dockstatus)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='永夜运动订单表'`)
	if err != nil {
		fmt.Printf("[Yongye] 建表 yy_ydsj_dd 失败: %v\n", err)
	}

	_, err = database.DB.Exec(`CREATE TABLE IF NOT EXISTS yy_ydsj_student (
		id INT NOT NULL AUTO_INCREMENT,
		uid INT NOT NULL DEFAULT 0,
		user VARCHAR(100) NOT NULL DEFAULT '' COMMENT '学号',
		pass VARCHAR(100) NOT NULL DEFAULT '' COMMENT '密码',
		type TINYINT NOT NULL DEFAULT 0 COMMENT '跑步类型',
		zkm DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '公里数',
		weeks VARCHAR(20) NOT NULL DEFAULT '' COMMENT '跑步周天',
		status TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0=正常 1=暂停 2=完成 3=退单',
		tdkm DECIMAL(10,2) DEFAULT NULL COMMENT '退单公里',
		tdmoney DECIMAL(10,2) DEFAULT NULL COMMENT '退单金额',
		stulog TEXT COMMENT '学生日志JSON',
		last_time DATETIME DEFAULT NULL,
		PRIMARY KEY (id),
		KEY idx_uid (uid),
		KEY idx_user (user)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='永夜运动学生表'`)
	if err != nil {
		fmt.Printf("[Yongye] 建表 yy_ydsj_student 失败: %v\n", err)
	}
}

// ---------- 上游 HTTP 工具 ----------

// yongyePostForm 向上游发起 POST 表单请求
func (s *YongyeService) yongyePostForm(apiURL string, params map[string]string) ([]byte, error) {
	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	resp, err := s.client.PostForm(apiURL, form)
	if err != nil {
		return nil, fmt.Errorf("上游请求失败: %v", err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// yongyeUpstreamPost 调用上游API（自动附加token）
func (s *YongyeService) yongyeUpstreamPost(cfg *YongyeConfig, endpoint string, extra map[string]string) ([]byte, error) {
	if cfg.ApiURL == "" || cfg.Token == "" {
		return nil, fmt.Errorf("永夜运动未配置上游接口")
	}
	params := map[string]string{"token": cfg.Token}
	for k, v := range extra {
		params[k] = v
	}
	apiURL := strings.TrimRight(cfg.ApiURL, "/") + "/" + strings.TrimLeft(endpoint, "/")
	return s.yongyePostForm(apiURL, params)
}

// ---------- 配置 ----------

func (s *YongyeService) GetConfig() (*YongyeConfig, error) {
	var val string
	err := database.DB.QueryRow("SELECT svalue FROM qingka_wangke_config WHERE skey = 'yongye_config' LIMIT 1").Scan(&val)
	if err != nil {
		return &YongyeConfig{Zs: 1.25, Beis: 1.3, Tk: 0.01, Xzmo: 100}, nil
	}
	var cfg YongyeConfig
	json.Unmarshal([]byte(val), &cfg)
	return &cfg, nil
}

func (s *YongyeService) SaveConfig(cfg *YongyeConfig) error {
	data, _ := json.Marshal(cfg)
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_config (skey, svalue) VALUES ('yongye_config', ?) ON DUPLICATE KEY UPDATE svalue = ?",
		string(data), string(data),
	)
	return err
}

// ---------- 学校列表 ----------

func (s *YongyeService) GetSchools(uid int) (interface{}, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return nil, err
	}

	respBody, err := s.yongyeUpstreamPost(cfg, "school", nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析学校列表失败")
	}

	code := mapGetFloat(result, "code")
	if int(code) != 1 {
		msg := mapGetString(result, "msg")
		if msg == "" {
			msg = "获取学校列表失败"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	// 获取用户等级用于价格计算
	var addprice float64
	database.DB.QueryRow("SELECT COALESCE(addprice, 0) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)

	// 计算等级折扣
	djfl := 1.0
	if cfg.Dj > 0 && addprice > 0 {
		djfl = math.Round(addprice/cfg.Dj*100) / 100
	}

	// 处理学校数据，调整价格
	if dataRaw, ok := result["data"].([]interface{}); ok {
		var schools []map[string]interface{}
		for _, d := range dataRaw {
			if school, ok := d.(map[string]interface{}); ok {
				name := mapGetString(school, "name")
				if name != "自动识别" {
					cpmuch := mapGetFloat(school, "cpmuch")
					zcmuch := mapGetFloat(school, "zcmuch")
					if cfg.Zs > 0 {
						cpmuch = math.Round(cpmuch/cfg.Zs*cfg.Beis*100) / 100
						zcmuch = math.Round(zcmuch/cfg.Zs*cfg.Beis*100) / 100
					}
					cpmuch = math.Round(cpmuch*djfl*100) / 100
					zcmuch = math.Round(zcmuch*djfl*100) / 100
					school["cpmuch"] = cpmuch
					school["zcmuch"] = zcmuch
				}
				schools = append(schools, school)
			}
		}
		result["data"] = schools
	}

	return result, nil
}

// ---------- 订单列表 ----------

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

// ---------- 学生列表 ----------

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

// ---------- 下单 ----------

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

	// 计算价格：先获取学校价格
	schoolPrice := 3.0 // 默认价格
	schoolResp, err := s.yongyeUpstreamPost(cfg, "school", nil)
	if err == nil {
		var schoolResult map[string]interface{}
		if json.Unmarshal(schoolResp, &schoolResult) == nil {
			if int(mapGetFloat(schoolResult, "code")) == 1 {
				if dataRaw, ok := schoolResult["data"].([]interface{}); ok {
					for _, d := range dataRaw {
						if item, ok := d.(map[string]interface{}); ok {
							if mapGetString(item, "name") == school {
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
		}
	}

	// 调整价格（除以赠送倍率 * 学校价格倍数）
	if cfg.Zs > 0 {
		schoolPrice = math.Round(schoolPrice/cfg.Zs*cfg.Beis*100) / 100
	}

	// 等级折扣
	var addprice float64
	database.DB.QueryRow("SELECT COALESCE(addprice, 0) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
	djfl := 1.0
	if cfg.Dj > 0 && addprice > 0 {
		djfl = math.Round(addprice/cfg.Dj*100) / 100
	}
	schoolPrice = math.Round(schoolPrice*djfl*100) / 100

	// 计算预扣款
	yfees := math.Round(zkm*schoolPrice*100) / 100

	var balance float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&balance)
	if balance < yfees || balance < 0 {
		return "", fmt.Errorf("余额不足，需要 %.2f 元，当前余额 %.2f 元", yfees, balance)
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	// 检查重复提交
	var dupID int
	database.DB.QueryRow("SELECT id FROM yy_ydsj_dd WHERE uid=? AND type=? AND user=? AND pass=? AND zkm=? AND ks_h=? AND ks_m=? AND js_h=? AND js_m=? AND weeks=? LIMIT 1",
		uid, runType, user, pass, zkm, ksH, ksM, jsH, jsM, weeks).Scan(&dupID)
	if dupID > 0 {
		return "", fmt.Errorf("重复提交，已阻止")
	}

	// 插入本地订单
	result, err := database.DB.Exec(
		"INSERT INTO yy_ydsj_dd (pol, uid, user, pass, school, type, zkm, ks_h, ks_m, js_h, js_m, weeks, dockstatus, yfees, addtime) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,0,?,?)",
		isPolling, uid, user, pass, school, runType, zkm, ksH, ksM, jsH, jsM, weeks, yfees, now,
	)
	if err != nil {
		return "", fmt.Errorf("订单创建失败: %v", err)
	}
	localID, _ := result.LastInsertId()

	// 扣费
	database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", yfees, uid)

	// 记录日志
	logContent := fmt.Sprintf("永夜运动下单：账号%s %.1fKM 扣费%.2f", user, zkm, yfees)
	smoney := math.Round((balance-yfees)*100) / 100
	database.DB.Exec("INSERT INTO qingka_wangke_moneylog (uid, type, money, mark, addtime) VALUES (?, 'yongye_add', ?, ?, ?)",
		uid, -yfees, logContent, now)
	database.DB.Exec("INSERT INTO qingka_wangke_log (uid, type, text, money, smoney, ip) VALUES (?, 'API永夜运动下单', ?, ?, ?, '')",
		uid, logContent, fmt.Sprintf("-%.2f", yfees), fmt.Sprintf("%.2f", smoney))

	// 调用上游 /add
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

// ---------- 退款（退单） ----------

func (s *YongyeService) RefundStudent(uid int, user string, runType int, isAdmin bool) (string, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return "", err
	}

	// 检查本地学生
	var stuUID int
	err = database.DB.QueryRow("SELECT uid FROM yy_ydsj_student WHERE user = ? AND uid = ? LIMIT 1", user, uid).Scan(&stuUID)
	if err != nil && !isAdmin {
		return "", fmt.Errorf("你的账号下无此学生")
	}

	// 调用上游退单
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
		// 记录日志
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

// ---------- 修改学生信息 ----------

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

	// 检查权限
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

// ---------- 轮询开关 ----------

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

	// 调用上游
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
			// 开启轮询 → 扣款
			database.DB.Exec("UPDATE yy_ydsj_dd SET dockstatus = 5, pol = 1, tktext = '开启轮询模式' WHERE id = ?", orderID)
			database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", order.Yfees, order.UID)
			database.DB.Exec("INSERT INTO qingka_wangke_log (uid, type, text, money, smoney, ip) VALUES (?, '开启轮询', ?, ?, '', '')",
				order.UID, fmt.Sprintf("订单ID：%d - 开启轮询，扣除余额", orderID), fmt.Sprintf("-%.2f", order.Yfees))
			_ = now
			return "已开启轮询", nil
		}
		// 关闭轮询 → 退款
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

// ---------- 本地退款（管理员） ----------

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
