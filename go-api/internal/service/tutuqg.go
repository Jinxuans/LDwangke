package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-api/internal/database"
)

// TutuQGConfig 图图强国配置
type TutuQGConfig struct {
	BaseURL        string  `json:"base_url"`
	Key            string  `json:"key"`
	PriceIncrement float64 `json:"price_increment"`
}

// TutuQGOrder 图图强国订单
type TutuQGOrder struct {
	OID     int     `json:"oid"`
	UID     int     `json:"uid"`
	User    string  `json:"user"`
	Pass    string  `json:"pass"`
	KCName  string  `json:"kcname"`
	Days    string  `json:"days"`
	PTName  string  `json:"ptname"`
	Fees    string  `json:"fees"`
	AddTime string  `json:"addtime"`
	IP      *string `json:"IP"`
	Status  *string `json:"status"`
	Remarks *string `json:"remarks"`
	GUID    *string `json:"guid"`
	Score   string  `json:"score"`
	Scores  *string `json:"scores"`
	ZDXF    *string `json:"zdxf"`
}

// TutuQGService 图图强国服务
type TutuQGService struct {
	client *http.Client
}

// NewTutuQGService 创建图图强国服务
func NewTutuQGService() *TutuQGService {
	return &TutuQGService{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// EnsureTable 确保 tutuqg 表存在
func (s *TutuQGService) EnsureTable() {
	_, err := database.DB.Exec(`CREATE TABLE IF NOT EXISTS tutuqg (
		oid int(11) NOT NULL AUTO_INCREMENT,
		uid int(11) NOT NULL,
		user varchar(255) NOT NULL,
		pass varchar(255) NOT NULL,
		kcname varchar(255) NOT NULL,
		days varchar(255) NOT NULL,
		ptname varchar(255) NOT NULL,
		fees varchar(255) NOT NULL,
		addtime varchar(255) NOT NULL,
		IP varchar(255) DEFAULT NULL,
		status varchar(255) DEFAULT NULL,
		remarks varchar(255) DEFAULT NULL,
		guid varchar(255) DEFAULT NULL,
		score varchar(255) NOT NULL,
		scores varchar(255) DEFAULT NULL,
		zdxf varchar(255) DEFAULT NULL,
		PRIMARY KEY (oid)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		log.Printf("[TutuQG] 创建表失败: %v", err)
	}
}

// GetConfig 获取图图强国配置
func (s *TutuQGService) GetConfig() (*TutuQGConfig, error) {
	var val string
	err := database.DB.QueryRow("SELECT svalue FROM qingka_wangke_config WHERE skey = 'tutuqg_config' LIMIT 1").Scan(&val)
	if err != nil {
		return &TutuQGConfig{}, nil
	}
	var cfg TutuQGConfig
	json.Unmarshal([]byte(val), &cfg)
	return &cfg, nil
}

// SaveConfig 保存图图强国配置
func (s *TutuQGService) SaveConfig(cfg *TutuQGConfig) error {
	data, _ := json.Marshal(cfg)
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_config WHERE skey = 'tutuqg_config'").Scan(&count)
	if count > 0 {
		_, err := database.DB.Exec("UPDATE qingka_wangke_config SET svalue = ? WHERE skey = 'tutuqg_config'", string(data))
		return err
	}
	_, err := database.DB.Exec("INSERT INTO qingka_wangke_config (v, k, skey, svalue) VALUES ('tutuqg_config', '', 'tutuqg_config', ?)", string(data))
	return err
}

// ListOrders 查询订单列表
func (s *TutuQGService) ListOrders(uid int, isAdmin bool, page, limit int, search string) ([]TutuQGOrder, int, error) {
	offset := (page - 1) * limit
	var args []interface{}
	where := "1=1"

	if !isAdmin {
		where += " AND uid = ?"
		args = append(args, uid)
	}
	if search != "" {
		where += " AND user LIKE ?"
		args = append(args, "%"+search+"%")
	}

	var total int
	countSQL := "SELECT COUNT(*) FROM tutuqg WHERE " + where
	err := database.DB.QueryRow(countSQL, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	querySQL := fmt.Sprintf("SELECT oid, uid, user, pass, kcname, days, ptname, fees, addtime, IP, status, remarks, guid, score, scores, zdxf FROM tutuqg WHERE %s ORDER BY oid DESC LIMIT ?, ?", where)
	args = append(args, offset, limit)

	rows, err := database.DB.Query(querySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []TutuQGOrder
	for rows.Next() {
		var o TutuQGOrder
		err := rows.Scan(&o.OID, &o.UID, &o.User, &o.Pass, &o.KCName, &o.Days, &o.PTName, &o.Fees, &o.AddTime, &o.IP, &o.Status, &o.Remarks, &o.GUID, &o.Score, &o.Scores, &o.ZDXF)
		if err != nil {
			continue
		}
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []TutuQGOrder{}
	}
	return orders, total, nil
}

// GetPrice 计算下单价格
func (s *TutuQGService) GetPrice(uid int, days int) (float64, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return 0, err
	}

	var addprice float64
	err = database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
	if err != nil {
		return 0, fmt.Errorf("用户不存在或单价信息缺失")
	}

	totalCost := (addprice + cfg.PriceIncrement) * 10 / 30 * float64(days)
	return totalCost, nil
}

// AddOrder 下单
func (s *TutuQGService) AddOrder(uid int, user, pass, kcname string, days int, ip string) error {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return fmt.Errorf("图图强国未配置")
	}

	// 查询用户单价和余额
	var addprice, money float64
	err = database.DB.QueryRow("SELECT addprice, money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice, &money)
	if err != nil {
		return fmt.Errorf("未找到对应用户")
	}

	addprice += cfg.PriceIncrement
	price := addprice * 10 / 30 * float64(days)

	if money < price {
		return fmt.Errorf("余额不足，无法支付订单")
	}

	// 生成 GUID
	guid := fmt.Sprintf("%d-%d-%d", uid, time.Now().UnixNano(), days)

	// 发送到上游
	upstreamData := map[string]interface{}{
		"key":    cfg.Key,
		"user":   user,
		"pass":   pass,
		"days":   days,
		"kcname": kcname,
		"guid":   guid,
	}
	jsonData, _ := json.Marshal(upstreamData)
	apiURL := strings.TrimRight(cfg.BaseURL, "/") + "/index/api/hd_ts.php"

	resp, err := s.client.Post(apiURL, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("无法连接到目标服务器")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	msg, _ := result["message"].(string)
	if msg != "下单成功" {
		if msg == "" {
			msg = "上游返回异常"
		}
		return fmt.Errorf("下单失败: %s", msg)
	}

	// 原子扣除余额，防止并发超扣
	res, err := database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ? AND money >= ?", price, uid, price)
	if err != nil {
		return fmt.Errorf("扣除余额失败")
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("余额不足，无法支付订单")
	}

	// 查询扣款后余额用于日志
	var newBalance float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&newBalance)

	// 插入订单
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	_, err = database.DB.Exec(
		"INSERT INTO tutuqg (uid, user, pass, kcname, days, ptname, fees, addtime, IP, status, remarks, guid, score) VALUES (?, ?, ?, ?, ?, '天数下单', ?, ?, ?, '待处理', '', ?, '')",
		uid, user, pass, kcname, days, fmt.Sprintf("%.2f", price), currentTime, ip, guid,
	)
	if err != nil {
		// 订单插入失败，回滚余额
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", price, uid)
		return fmt.Errorf("写入订单失败")
	}

	// 写入日志
	logText := fmt.Sprintf("代看 %d 天 %s %s %s 扣除 %.2f 元", days, user, pass, kcname, price)
	database.DB.Exec(
		"INSERT INTO qingka_wangke_log (uid, type, text, money, smoney, ip, addtime) VALUES (?, '天数下单', ?, ?, ?, ?, ?)",
		uid, logText, fmt.Sprintf("-%.2f", price), fmt.Sprintf("%.2f", newBalance), ip, currentTime,
	)

	return nil
}

// DeleteOrder 删除订单
func (s *TutuQGService) DeleteOrder(uid int, oid int, isAdmin bool) error {
	cfg, _ := s.GetConfig()

	// 查询订单
	var orderUser, guid, status string
	var orderUID int
	err := database.DB.QueryRow("SELECT uid, user, guid, COALESCE(status,'') FROM tutuqg WHERE oid = ?", oid).Scan(&orderUID, &orderUser, &guid, &status)
	if err != nil {
		return fmt.Errorf("订单不存在")
	}
	if !isAdmin && orderUID != uid {
		return fmt.Errorf("订单不属于该用户")
	}

	// 如果状态是"源台未找到此订单"，直接删除
	if status == "源台未找到此订单" {
		database.DB.Exec("DELETE FROM tutuqg WHERE oid = ?", oid)
		return nil
	}

	// 调用上游删除
	if cfg != nil && cfg.BaseURL != "" {
		apiURL := strings.TrimRight(cfg.BaseURL, "/") + "/index/api/delete_order.php"
		formData := url.Values{}
		formData.Set("key", cfg.Key)
		formData.Set("user", orderUser)
		formData.Set("guid", guid)

		resp, err := s.client.PostForm(apiURL, formData)
		if err == nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			var result map[string]interface{}
			json.Unmarshal(body, &result)
			if code, ok := result["code"].(float64); !ok || int(code) != 1 {
				msg, _ := result["msg"].(string)
				return fmt.Errorf("上游删除失败: %s", msg)
			}
		}
	}

	database.DB.Exec("DELETE FROM tutuqg WHERE oid = ?", oid)
	return nil
}

// RenewOrder 续费订单
func (s *TutuQGService) RenewOrder(uid int, oid int, days int, isAdmin bool) error {
	cfg, _ := s.GetConfig()
	if cfg == nil || cfg.BaseURL == "" {
		return fmt.Errorf("图图强国未配置")
	}

	// 查询订单
	var orderUID int
	var orderUser, guid string
	err := database.DB.QueryRow("SELECT uid, user, guid FROM tutuqg WHERE oid = ?", oid).Scan(&orderUID, &orderUser, &guid)
	if err != nil {
		return fmt.Errorf("订单不存在")
	}
	if !isAdmin && orderUID != uid {
		return fmt.Errorf("订单不属于该用户")
	}

	// 续费扣的是当前操作者的钱（即 uid）
	billingUID := uid

	// 查询单价和余额
	var addprice, money float64
	err = database.DB.QueryRow("SELECT addprice, money FROM qingka_wangke_user WHERE uid = ?", billingUID).Scan(&addprice, &money)
	if err != nil {
		return fmt.Errorf("获取单价失败")
	}

	addprice += cfg.PriceIncrement
	requiredMoney := addprice * 10 * float64(days) / 30

	if money < requiredMoney {
		return fmt.Errorf("余额不足，续费失败")
	}

	// 调用上游续费
	apiURL := strings.TrimRight(cfg.BaseURL, "/") + "/index/api/renewOrder.php"
	formData := url.Values{}
	formData.Set("key", cfg.Key)
	formData.Set("guid", guid)
	formData.Set("user", orderUser)
	formData.Set("days", fmt.Sprintf("%d", days))

	resp, err := s.client.PostForm(apiURL, formData)
	if err != nil {
		return fmt.Errorf("无法连接上游")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	if code, ok := result["code"].(float64); !ok || int(code) != 1 {
		return fmt.Errorf("续费失败: 订单时间未正常")
	}

	// 原子扣除余额，防止并发超扣
	res, err := database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ? AND money >= ?", requiredMoney, billingUID, requiredMoney)
	if err != nil {
		return fmt.Errorf("扣除余额失败")
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("余额不足，续费失败")
	}

	// 查询扣款后余额用于日志
	var newBalance float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", billingUID).Scan(&newBalance)

	// 写入日志
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	logText := fmt.Sprintf("续费账号：%s，续费天数：%d天", orderUser, days)
	database.DB.Exec(
		"INSERT INTO qingka_wangke_log (uid, type, text, money, smoney, ip, addtime) VALUES (?, '续费', ?, ?, ?, '', ?)",
		billingUID, logText, fmt.Sprintf("-%.2f", requiredMoney), fmt.Sprintf("%.2f", newBalance), currentTime,
	)

	return nil
}

// ChangePassword 修改密码
func (s *TutuQGService) ChangePassword(uid int, oid int, newPassword string, isAdmin bool) error {
	cfg, _ := s.GetConfig()
	if cfg == nil || cfg.BaseURL == "" {
		return fmt.Errorf("图图强国未配置")
	}

	var orderUID int
	var orderUser, guid string
	err := database.DB.QueryRow("SELECT uid, user, guid FROM tutuqg WHERE oid = ?", oid).Scan(&orderUID, &orderUser, &guid)
	if err != nil {
		return fmt.Errorf("订单不存在")
	}
	if !isAdmin && orderUID != uid {
		return fmt.Errorf("订单不属于该用户")
	}

	apiURL := strings.TrimRight(cfg.BaseURL, "/") + "/index/api/update_password.php"
	formData := url.Values{}
	formData.Set("user", orderUser)
	formData.Set("guid", guid)
	formData.Set("newPassword", newPassword)

	resp, err := s.client.PostForm(apiURL, formData)
	if err != nil {
		return fmt.Errorf("无法连接上游")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	if success, ok := result["success"].(bool); !ok || !success {
		msg, _ := result["message"].(string)
		return fmt.Errorf("修改失败：%s", msg)
	}

	database.DB.Exec("UPDATE tutuqg SET pass = ? WHERE oid = ?", newPassword, oid)
	return nil
}

// ChangeToken 修改 Token
func (s *TutuQGService) ChangeToken(uid int, oid int, newToken string, isAdmin bool) error {
	cfg, _ := s.GetConfig()
	if cfg == nil || cfg.BaseURL == "" {
		return fmt.Errorf("图图强国未配置")
	}

	var orderUID int
	var orderUser, guid string
	err := database.DB.QueryRow("SELECT uid, user, guid FROM tutuqg WHERE oid = ?", oid).Scan(&orderUID, &orderUser, &guid)
	if err != nil {
		return fmt.Errorf("订单不存在")
	}
	if !isAdmin && orderUID != uid {
		return fmt.Errorf("订单不属于该用户")
	}

	apiURL := strings.TrimRight(cfg.BaseURL, "/") + "/index/api/update_token.php"
	formData := url.Values{}
	formData.Set("user", orderUser)
	formData.Set("guid", guid)
	formData.Set("newToken", newToken)

	resp, err := s.client.PostForm(apiURL, formData)
	if err != nil {
		return fmt.Errorf("无法连接上游")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	if success, ok := result["success"].(bool); !ok || !success {
		msg, _ := result["message"].(string)
		return fmt.Errorf("修改失败：%s", msg)
	}

	database.DB.Exec("UPDATE tutuqg SET kcname = ? WHERE oid = ?", newToken, oid)
	return nil
}

// RefundOrder 退单退费
func (s *TutuQGService) RefundOrder(uid int, oid int, isAdmin bool) error {
	cfg, _ := s.GetConfig()
	if cfg == nil || cfg.BaseURL == "" {
		return fmt.Errorf("图图强国未配置")
	}

	var orderUID int
	var orderUser, guid, fees string
	var remarks *string
	err := database.DB.QueryRow("SELECT uid, user, fees, guid, remarks FROM tutuqg WHERE oid = ?", oid).Scan(&orderUID, &orderUser, &fees, &guid, &remarks)
	if err != nil {
		return fmt.Errorf("订单不存在")
	}
	if !isAdmin && orderUID != uid {
		return fmt.Errorf("订单不属于该用户")
	}

	remarksStr := ""
	if remarks != nil {
		remarksStr = *remarks
	}

	// 调用上游退单
	apiURL := strings.TrimRight(cfg.BaseURL, "/") + "/index/api/process_refund.php"
	formData := url.Values{}
	formData.Set("user", orderUser)
	formData.Set("guid", guid)
	formData.Set("remarks", remarksStr)

	resp, err := s.client.PostForm(apiURL, formData)
	if err != nil {
		return fmt.Errorf("无法连接上游")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	if code, ok := result["code"].(float64); !ok || int(code) != 1 {
		msg, _ := result["msg"].(string)
		return fmt.Errorf("退单失败: %s", msg)
	}

	// 计算退款金额
	newRemarks, _ := result["remarks"].(string)
	var addprice float64
	database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", orderUID).Scan(&addprice)
	addprice += cfg.PriceIncrement
	dailyPrice := addprice * 10 / 30

	var refundAmount float64
	if remarksStr == "" {
		// remarks 为空（新订单未同步），全额退款
		fmt.Sscanf(fees, "%f", &refundAmount)
	} else if remarksStr != newRemarks {
		// 日期与源台不匹配，拒绝退单，需先同步
		return fmt.Errorf("退单失败：日期与源台不匹配，请先进行同步")
	} else {
		// 按剩余天数计算退款
		expiryTime, err := time.Parse("2006-01-02", newRemarks)
		if err == nil {
			remaining := int(time.Until(expiryTime).Hours()/24) + 1
			if remaining > 0 {
				refundAmount = dailyPrice * float64(remaining)
			}
		}
	}

	// 原子更新用户余额
	database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", refundAmount, orderUID)

	// 查询更新后余额用于日志
	var newMoney float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", orderUID).Scan(&newMoney)

	// 写入日志
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	logText := fmt.Sprintf("退单退费%s 增加%.2f元！", orderUser, refundAmount)
	database.DB.Exec(
		"INSERT INTO qingka_wangke_log (uid, type, text, money, smoney, ip, addtime) VALUES (?, '退单退费', ?, ?, ?, '', ?)",
		orderUID, logText, fmt.Sprintf("%.2f", refundAmount), fmt.Sprintf("%.2f", newMoney), currentTime,
	)

	// 删除订单
	database.DB.Exec("DELETE FROM tutuqg WHERE oid = ?", oid)
	return nil
}

// SyncOrder 同步订单
func (s *TutuQGService) SyncOrder(uid int, oid int, isAdmin bool) (string, error) {
	cfg, _ := s.GetConfig()
	if cfg == nil || cfg.BaseURL == "" {
		return "", fmt.Errorf("图图强国未配置")
	}

	var orderUID int
	var orderUser, guid string
	err := database.DB.QueryRow("SELECT uid, user, COALESCE(guid,'') FROM tutuqg WHERE oid = ?", oid).Scan(&orderUID, &orderUser, &guid)
	if err != nil {
		return "", fmt.Errorf("订单不存在")
	}
	if !isAdmin && orderUID != uid {
		return "", fmt.Errorf("订单不属于该用户")
	}

	apiURL := strings.TrimRight(cfg.BaseURL, "/") + "/index/api/syncAllOrders.php"
	formData := url.Values{}
	formData.Set("key", cfg.Key)
	formData.Set("guid", guid)
	formData.Set("user", orderUser)

	resp, err := s.client.PostForm(apiURL, formData)
	if err != nil {
		return "", fmt.Errorf("无法连接上游")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var orderData map[string]interface{}
	if err := json.Unmarshal(body, &orderData); err != nil {
		return "", fmt.Errorf("上游返回异常")
	}

	status, _ := orderData["status"].(string)
	remarksVal, _ := orderData["remarks"].(string)
	score, _ := orderData["score"].(string)
	scores, _ := orderData["scores"].(string)

	if status == "" {
		status = "源台未找到此订单"
	}
	if score == "" {
		score = "待更新"
	}

	database.DB.Exec("UPDATE tutuqg SET scores = ?, score = ?, status = ?, remarks = ? WHERE oid = ?",
		scores, score, status, remarksVal, oid)

	return "订单信息已成功更新", nil
}

// ToggleAutoRenew 切换自动续费
func (s *TutuQGService) ToggleAutoRenew(uid int, oid int, isAdmin bool) error {
	var orderUID int
	var zdxf *string
	err := database.DB.QueryRow("SELECT uid, zdxf FROM tutuqg WHERE oid = ?", oid).Scan(&orderUID, &zdxf)
	if err != nil {
		return fmt.Errorf("订单不存在")
	}
	if !isAdmin && orderUID != uid {
		return fmt.Errorf("订单不属于该用户")
	}

	var newVal *string
	if zdxf != nil && *zdxf == "2" {
		newVal = nil
	} else {
		v := "2"
		newVal = &v
	}

	_, err = database.DB.Exec("UPDATE tutuqg SET zdxf = ? WHERE oid = ?", newVal, oid)
	return err
}
