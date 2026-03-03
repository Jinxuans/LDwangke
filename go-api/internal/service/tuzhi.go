package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-api/internal/database"
)

// ---------- 配置 ----------

type TuZhiConfig struct {
	Username string `json:"daka_api_username"` // 上游账号
	Password string `json:"daka_api_password"` // 上游密码
}

// TuZhiGoodsOverride 商品价格/上架覆盖
type TuZhiGoodsOverride struct {
	GoodsID int     `json:"goods_id"`
	Price   float64 `json:"price"`   // 覆盖售价，0 表示用上游原价
	Enabled int     `json:"enabled"` // 1=上架 0=下架
}

// ---------- 服务 ----------

type TuZhiService struct {
	client  *http.Client
	baseURL string
}

func NewTuZhiService() *TuZhiService {
	return &TuZhiService{
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: "http://apis.bbwace.icu",
	}
}

// EnsureTable 创建凸知打卡所需的表
func (s *TuZhiService) EnsureTable() {
	log.Println("[TuZhi] 开始检查/创建表")
	_, err := database.DB.Exec(`CREATE TABLE IF NOT EXISTS qingka_wangke_dakaaz (
		id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID',
		api_id INT(11) DEFAULT NULL COMMENT '上游订单ID',
		user_id INT(11) NOT NULL COMMENT '用户ID',
		goods_id INT(11) NOT NULL DEFAULT 0 COMMENT '所属商品ID',
		username VARCHAR(40) NOT NULL COMMENT '账号',
		password VARCHAR(30) NOT NULL COMMENT '密码',
		nickname VARCHAR(20) DEFAULT NULL COMMENT '姓名',
		school VARCHAR(100) DEFAULT NULL COMMENT '学校名称',
		postname VARCHAR(50) DEFAULT NULL COMMENT '岗位名称',
		address VARCHAR(100) DEFAULT NULL COMMENT '地址',
		address_lat VARCHAR(50) DEFAULT NULL COMMENT '纬度',
		address_lng VARCHAR(50) DEFAULT NULL COMMENT '经度',
		work_time VARCHAR(100) DEFAULT NULL COMMENT '上班打卡时间',
		off_time VARCHAR(100) DEFAULT NULL COMMENT '下班打卡时间',
		work_days VARCHAR(100) DEFAULT NULL COMMENT '打卡周期',
		work_days_num BIGINT(20) DEFAULT NULL COMMENT '打卡天数',
		work_days_ok_num BIGINT(20) NOT NULL DEFAULT 0 COMMENT '已打卡天数',
		daily_report INT(11) DEFAULT 0 COMMENT '日报',
		weekly_report INT(11) DEFAULT 0 COMMENT '周报',
		monthly_report INT(10) UNSIGNED DEFAULT 0 COMMENT '月报',
		weekly_report_time BIGINT(20) DEFAULT NULL COMMENT '周报时间',
		monthly_report_time BIGINT(20) DEFAULT NULL COMMENT '月报时间',
		holiday_status INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '法定节假日 0=不跳过 1=跳过',
		token VARCHAR(255) DEFAULT NULL,
		uuid VARCHAR(255) DEFAULT NULL,
		user_school_id VARCHAR(255) DEFAULT NULL,
		random_phone VARCHAR(255) DEFAULT NULL,
		price DECIMAL(10,2) UNSIGNED NOT NULL DEFAULT 0.00 COMMENT '扣除金额',
		images TEXT COMMENT '图片',
		create_time BIGINT(20) DEFAULT NULL COMMENT '创建时间',
		update_time BIGINT(20) DEFAULT NULL COMMENT '更新时间',
		delete_time BIGINT(20) DEFAULT NULL COMMENT '删除时间',
		remark VARCHAR(100) DEFAULT '' COMMENT '备注',
		status INT(1) UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态 0正常 1打卡中 2关闭 3已完成',
		is_status INT(1) DEFAULT 1 COMMENT '打卡状态 0失败 1正常',
		work_deadline VARCHAR(100) DEFAULT NULL COMMENT '截至打卡日期',
		billing_method TINYINT(1) UNSIGNED DEFAULT 1 COMMENT '扣费方式（1按日，2按月）',
		billing_months TINYINT(3) UNSIGNED DEFAULT 0 COMMENT '收费月数',
		is_off_time TINYINT(1) UNSIGNED DEFAULT 1 COMMENT '是否开启下班打卡 1是0否',
		xz_push_url TEXT COMMENT '息知推送地址',
		PRIMARY KEY (id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='凸知打卡订单表'`)
	if err != nil {
		log.Printf("[TuZhi] 创建 dakaaz 表失败: %v", err)
	}

	_, err = database.DB.Exec(`CREATE TABLE IF NOT EXISTS qingka_wangke_daka_query_record (
		id INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
		username VARCHAR(255) DEFAULT NULL,
		password VARCHAR(255) DEFAULT NULL,
		create_time BIGINT(20) DEFAULT NULL COMMENT '创建时间',
		user_id INT(11) DEFAULT NULL,
		is_success TINYINT(1) UNSIGNED DEFAULT 0 COMMENT '是否成功',
		price DECIMAL(10,2) UNSIGNED NOT NULL DEFAULT 0.00 COMMENT '扣除金额',
		goods_id INT(11) NOT NULL DEFAULT 0 COMMENT '所属商品ID',
		PRIMARY KEY (id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='凸知打卡查询扣费记录'`)
	if err != nil {
		log.Printf("[TuZhi] 创建 daka_query_record 表失败: %v", err)
	}
	log.Println("[TuZhi] 表检查/创建完成")
}

// ---------- 配置管理 ----------

func (s *TuZhiService) GetConfig() (*TuZhiConfig, error) {
	var val string
	err := database.DB.QueryRow("SELECT svalue FROM qingka_wangke_config WHERE skey = 'tuzhi_config' LIMIT 1").Scan(&val)
	if err != nil {
		return &TuZhiConfig{}, nil
	}
	var cfg TuZhiConfig
	json.Unmarshal([]byte(val), &cfg)
	return &cfg, nil
}

func (s *TuZhiService) SaveConfig(cfg *TuZhiConfig) error {
	data, _ := json.Marshal(cfg)
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_config WHERE skey = 'tuzhi_config'").Scan(&count)
	if count > 0 {
		_, err := database.DB.Exec("UPDATE qingka_wangke_config SET svalue = ? WHERE skey = 'tuzhi_config'", string(data))
		return err
	}
	_, err := database.DB.Exec("INSERT INTO qingka_wangke_config (v, k, skey, svalue) VALUES ('tuzhi_config', '', 'tuzhi_config', ?)", string(data))
	return err
}

// ---------- 商品价格覆盖管理 ----------

func (s *TuZhiService) GetGoodsOverrides() ([]TuZhiGoodsOverride, error) {
	var val string
	err := database.DB.QueryRow("SELECT svalue FROM qingka_wangke_config WHERE skey = 'tuzhi_goods_overrides' LIMIT 1").Scan(&val)
	if err != nil {
		return []TuZhiGoodsOverride{}, nil
	}
	var list []TuZhiGoodsOverride
	json.Unmarshal([]byte(val), &list)
	return list, nil
}

func (s *TuZhiService) SaveGoodsOverrides(list []TuZhiGoodsOverride) error {
	data, _ := json.Marshal(list)
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_config WHERE skey = 'tuzhi_goods_overrides'").Scan(&count)
	if count > 0 {
		_, err := database.DB.Exec("UPDATE qingka_wangke_config SET svalue = ? WHERE skey = 'tuzhi_goods_overrides'", string(data))
		return err
	}
	_, err := database.DB.Exec("INSERT INTO qingka_wangke_config (v, k, skey, svalue) VALUES ('tuzhi_goods_overrides', '', 'tuzhi_goods_overrides', ?)", string(data))
	return err
}

// ---------- 上游 API ----------

func (s *TuZhiService) login(cfg *TuZhiConfig) (string, error) {
	if cfg.Username == "" || cfg.Password == "" {
		return "", fmt.Errorf("凸知打卡未配置账号密码")
	}
	body := map[string]interface{}{
		"terminal": "2",
		"account":  cfg.Username,
		"password": cfg.Password,
	}
	jsonData, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", s.baseURL+"/user/login/account", strings.NewReader(string(jsonData)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("登录请求失败: %v", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)
	code, _ := result["code"].(float64)
	if code != 200 {
		msg, _ := result["msg"].(string)
		return "", fmt.Errorf("登录失败: %s", msg)
	}
	data, _ := result["data"].(map[string]interface{})
	token, _ := data["token"].(string)
	if token == "" {
		return "", fmt.Errorf("登录返回无token")
	}
	return token, nil
}

func (s *TuZhiService) upstreamRequest(method, path string, token string, data interface{}) (map[string]interface{}, error) {
	var req *http.Request
	var err error

	if method == "GET" {
		url := s.baseURL + path
		if data != nil {
			if params, ok := data.(map[string]interface{}); ok {
				parts := []string{}
				for k, v := range params {
					parts = append(parts, fmt.Sprintf("%s=%v", k, v))
				}
				url += "?" + strings.Join(parts, "&")
			}
		}
		req, err = http.NewRequest("GET", url, nil)
	} else {
		jsonData, _ := json.Marshal(data)
		req, err = http.NewRequest("POST", s.baseURL+path, strings.NewReader(string(jsonData)))
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("上游请求失败: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("上游返回解析失败: %s", string(body))
	}
	return result, nil
}

// ---------- 获取商品列表 ----------

func (s *TuZhiService) GetGoods() ([]map[string]interface{}, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return nil, err
	}
	token, err := s.login(cfg)
	if err != nil {
		return nil, err
	}
	result, err := s.upstreamRequest("GET", "/user/mall/lists", token, nil)
	if err != nil {
		return nil, err
	}
	code, _ := result["code"].(float64)
	if code != 200 {
		msg, _ := result["msg"].(string)
		return nil, fmt.Errorf("获取商品失败: %s", msg)
	}
	data, _ := result["data"].(map[string]interface{})
	lists, _ := data["lists"].([]interface{})
	var goods []map[string]interface{}
	for _, l := range lists {
		group, _ := l.(map[string]interface{})
		items, _ := group["goods"].([]interface{})
		for _, item := range items {
			if g, ok := item.(map[string]interface{}); ok {
				goods = append(goods, g)
			}
		}
	}
	return goods, nil
}

// GetGoodsForUser 获取商品列表（含价格覆盖和上下架过滤）
func (s *TuZhiService) GetGoodsForUser(addprice float64) ([]map[string]interface{}, error) {
	goods, err := s.GetGoods()
	if err != nil {
		return nil, err
	}
	overrides, _ := s.GetGoodsOverrides()
	overrideMap := map[int]TuZhiGoodsOverride{}
	for _, o := range overrides {
		overrideMap[o.GoodsID] = o
	}

	var result []map[string]interface{}
	for _, g := range goods {
		gidF, _ := g["id"].(float64)
		gid := int(gidF)
		ov, hasOv := overrideMap[gid]
		// 检查上架状态
		if hasOv && ov.Enabled == 0 {
			continue
		}
		price, _ := g["price"].(float64)
		if hasOv && ov.Price > 0 {
			price = ov.Price
		}
		billingMethod := 1
		if bm, ok := g["billing_method"].(float64); ok {
			billingMethod = int(bm)
		}
		unit := "天"
		if billingMethod == 2 {
			unit = "月"
		}
		userPrice := math.Round(addprice*price*100) / 100
		name, _ := g["name"].(string)
		g["display_name"] = fmt.Sprintf("%s %.2f/%s", name, userPrice, unit)
		g["user_price"] = userPrice
		result = append(result, g)
	}
	return result, nil
}

// ---------- 获取学校列表 ----------

func (s *TuZhiService) GetSchools(form map[string]interface{}) (interface{}, error) {
	cfg, _ := s.GetConfig()
	token, err := s.login(cfg)
	if err != nil {
		return nil, err
	}
	result, err := s.upstreamRequest("GET", "/user/finance.order/getSchool", token, form)
	if err != nil {
		return nil, err
	}
	code, _ := result["code"].(float64)
	if code != 200 {
		msg, _ := result["msg"].(string)
		return nil, fmt.Errorf("%s", msg)
	}
	return result["data"], nil
}

// ---------- 获取签到信息 ----------

func (s *TuZhiService) GetSignInfo(uid int, form map[string]interface{}) (interface{}, error) {
	cfg, _ := s.GetConfig()
	token, err := s.login(cfg)
	if err != nil {
		return nil, err
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

	// 按月计费的商品需要先扣费查询
	overrides, _ := s.GetGoodsOverrides()
	goods, err := s.GetGoods()
	if err == nil {
		for _, g := range goods {
			gidF, _ := g["id"].(float64)
			if int(gidF) == goodsID {
				bm, _ := g["billing_method"].(float64)
				if int(bm) == 2 {
					// 检查是否已有订单
					var count int
					database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_dakaaz WHERE user_id=? AND username=?", uid, username).Scan(&count)
					if count == 0 {
						// 检查是否已有查询记录
						var recCount int
						database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_daka_query_record WHERE user_id=? AND username=? AND is_success=1", uid, username).Scan(&recCount)
						if recCount == 0 {
							price, _ := g["price"].(float64)
							for _, ov := range overrides {
								if ov.GoodsID == goodsID && ov.Price > 0 {
									price = ov.Price
									break
								}
							}
							var addprice float64
							database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid=?", uid).Scan(&addprice)
							queryPrice := math.Round(addprice*price*100) / 100

							var money float64
							database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid=?", uid).Scan(&money)
							if money < queryPrice {
								return nil, fmt.Errorf("余额不足")
							}
							now := time.Now().Unix()
							database.DB.Exec("INSERT INTO qingka_wangke_daka_query_record (username, password, create_time, user_id, is_success, price, goods_id) VALUES (?, ?, ?, ?, 0, ?, ?)",
								username, form["password"], now, uid, queryPrice, goodsID)
							database.DB.Exec("UPDATE qingka_wangke_user SET money=money-? WHERE uid=?", queryPrice, uid)
							tuzhiLog(uid, "tuzhi-按月查询扣费", fmt.Sprintf("商品%d %s 扣%.2f", goodsID, username, queryPrice), -queryPrice)
						}
					}
				}
				break
			}
		}
	}

	result, err := s.upstreamRequest("POST", "/user/finance.order/detail", token, form)
	if err != nil {
		return nil, err
	}
	code, _ := result["code"].(float64)
	if code != 200 {
		msg, _ := result["msg"].(string)
		return nil, fmt.Errorf("%s", msg)
	}
	// 标记查询成功
	database.DB.Exec("UPDATE qingka_wangke_daka_query_record SET is_success=1 WHERE user_id=? AND username=? LIMIT 1", uid, username)
	return result["data"], nil
}

// ---------- 计算天数 ----------

func (s *TuZhiService) CalculateDays(form map[string]interface{}) (interface{}, error) {
	cfg, _ := s.GetConfig()
	token, err := s.login(cfg)
	if err != nil {
		return nil, err
	}
	result, err := s.upstreamRequest("POST", "/user/finance.order/calculateDays", token, form)
	if err != nil {
		return nil, err
	}
	code, _ := result["code"].(float64)
	if code != 200 {
		msg, _ := result["msg"].(string)
		return nil, fmt.Errorf("%s", msg)
	}
	return result["data"], nil
}

// ---------- 下单 ----------

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

	// 计算天数
	deadline, err := time.Parse("2006-01-02", workDeadline)
	if err != nil {
		return "", fmt.Errorf("截至日期格式错误")
	}
	now := time.Now()
	days := int(deadline.Sub(now).Hours()/24) + 1
	if days <= 0 {
		return "", fmt.Errorf("截至日期不能小于当前日期")
	}

	// 检查重复
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_dakaaz WHERE user_id=? AND username=? AND goods_id=?", uid, username, goodsID).Scan(&count)
	if count > 0 {
		return "", fmt.Errorf("订单已存在")
	}

	// 获取商品价格
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

	// 上游下单
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

	// 本地入库
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

	// 扣费
	database.DB.Exec("UPDATE qingka_wangke_user SET money=money-? WHERE uid=?", totalMoney, uid)
	tuzhiLog(uid, "tuzhi-添加订单", fmt.Sprintf("商品%d %s 天数%d 扣%.2f", goodsID, username, days, totalMoney), -totalMoney)

	// 按月计费退还查询费
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

// ---------- 编辑订单 ----------

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

	// 查询本地订单
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

	// 计算新增天数
	origDl, _ := time.Parse("2006-01-02", origDeadline)
	newDl, _ := time.Parse("2006-01-02", workDeadline)
	extraDays := int(newDl.Sub(origDl).Hours()/24) + 1

	// 设置上游ID
	form["id"] = apiID

	// 计算补费
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

	// 上游编辑
	result, err := s.upstreamRequest("POST", "/user/finance.order/edit", token, form)
	if err != nil {
		return "", fmt.Errorf("上游编辑失败: %v", err)
	}
	code, _ := result["code"].(float64)
	if code != 200 {
		msg, _ := result["msg"].(string)
		return "", fmt.Errorf("上游编辑失败: %s", msg)
	}

	// 补费
	if extraMoney > 0 {
		database.DB.Exec("UPDATE qingka_wangke_user SET money=money-? WHERE uid=?", extraMoney, uid)
		tuzhiLog(uid, "tuzhi-编辑补费", fmt.Sprintf("订单%d 补%.2f", localID, extraMoney), -extraMoney)
	}

	// 更新本地
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

// ---------- 删除订单 ----------

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

	// 上游删除
	_, err = s.upstreamRequest("POST", "/user/finance.order/delete", token, map[string]interface{}{"id": apiID})
	if err != nil {
		return "", fmt.Errorf("上游删除失败: %v", err)
	}

	// 按日退费
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

// ---------- 订单列表 ----------

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

// ---------- 立即打卡 ----------

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

// ---------- 同步订单状态 ----------

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

	// 按月查询超时退费
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

// ---------- 工具函数 ----------

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
