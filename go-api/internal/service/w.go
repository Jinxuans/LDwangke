package service

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

// WApp 鲸鱼运动项目
type WApp struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Code     string  `json:"code"`
	OrgAppID string  `json:"org_app_id"`
	Status   int     `json:"status"`
	Desc     string  `json:"description"`
	Price    float64 `json:"price"`
	CacType  string  `json:"cac_type"`
	URL      string  `json:"url"`
	Key      string  `json:"key"`
	UID      string  `json:"uid"`
	Token    string  `json:"token"`
	Type     string  `json:"type"`
	Deleted  int     `json:"deleted"`
}

// WAppUser 用户视角项目（含用户价格）
type WAppUser struct {
	AppID    int64   `json:"app_id"`
	OrgAppID string  `json:"org_app_id"`
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	Desc     string  `json:"description"`
	CacType  string  `json:"cac_type"`
	Price    float64 `json:"price"`
}

// WOrder 鲸鱼运动订单
type WOrder struct {
	ID         int64       `json:"id"`
	AggOrderID *string     `json:"agg_order_id"`
	UserID     int64       `json:"user_id"`
	School     string      `json:"school"`
	Account    string      `json:"account"`
	Password   string      `json:"password"`
	AppID      int64       `json:"app_id"`
	AppName    string      `json:"app_name"`
	Status     string      `json:"status"`
	Num        int         `json:"num"`
	Cost       float64     `json:"cost"`
	Pause      bool        `json:"pause"`
	SubOrder   interface{} `json:"sub_order"`
	Deleted    bool        `json:"deleted"`
	Created    string      `json:"created"`
	Updated    string      `json:"updated"`
}

type WService struct {
	client *http.Client
}

func NewWService() *WService {
	return &WService{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// ---------- HTTP 工具 ----------

func (s *WService) httpReq(method, reqURL string, body interface{}, headers map[string]string) (map[string]interface{}, error) {
	var reqBody io.Reader
	if body != nil {
		switch v := body.(type) {
		case string:
			reqBody = strings.NewReader(v)
		default:
			jsonData, _ := json.Marshal(body)
			reqBody = strings.NewReader(string(jsonData))
		}
	}

	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if method == "POST" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求外部接口失败: %v", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("外部接口返回格式错误: %s", string(respBody))
	}
	return result, nil
}

// ---------- 日志 ----------

func wLog(uid int, logType, text string, money float64) {
	now := time.Now().Format("2006-01-02 15:04:05")
	var newBalance float64
	database.DB.QueryRow("SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&newBalance)
	moneyStr := fmt.Sprintf("%.2f", money)
	if money > 0 {
		moneyStr = fmt.Sprintf("+%.2f", money)
	}
	database.DB.Exec(
		"INSERT INTO qingka_wangke_log (uid, type, text, money, smoney, ip, addtime) VALUES (?, ?, ?, ?, ?, '', ?)",
		uid, logType, text, moneyStr, fmt.Sprintf("%.2f", newBalance), now,
	)
}

// ---------- 查询项目（内部） ----------

func (s *WService) getAppRow(appID int64) (map[string]interface{}, error) {
	rows, err := database.DB.Query("SELECT * FROM w_app WHERE id = ? AND deleted = 0 LIMIT 1", appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, _ := rows.Columns()
	if !rows.Next() {
		return nil, fmt.Errorf("项目不存在")
	}
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}
	rows.Scan(valuePtrs...)
	result := make(map[string]interface{})
	for i, col := range columns {
		val := values[i]
		if b, ok := val.([]byte); ok {
			result[col] = string(b)
		} else {
			result[col] = val
		}
	}
	return result, nil
}

// ---------- 业务方法 ----------

// GetApps 获取项目列表（用户视角，含用户价格）
func (s *WService) GetApps(uid int) ([]WAppUser, error) {
	var addprice float64
	err := database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	rows, err := database.DB.Query("SELECT id, org_app_id, code, name, COALESCE(description,''), price, cac_type FROM w_app WHERE deleted = 0 AND status = 0")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []WAppUser
	for rows.Next() {
		var a WAppUser
		if err := rows.Scan(&a.AppID, &a.OrgAppID, &a.Code, &a.Name, &a.Desc, &a.Price, &a.CacType); err != nil {
			continue
		}
		a.Price = math.Round(a.Price*addprice*100) / 100
		list = append(list, a)
	}
	if list == nil {
		list = []WAppUser{}
	}
	return list, nil
}

// GetOrders 查询订单列表
func (s *WService) GetOrders(uid int, isAdmin bool, page, pageSize int, filters map[string]string) ([]WOrder, int, error) {
	offset := (page - 1) * pageSize
	where := "o.deleted = 0"
	var args []interface{}

	if !isAdmin {
		where += " AND o.user_id = ?"
		args = append(args, uid)
	}
	if v := filters["account"]; v != "" {
		where += " AND o.account = ?"
		args = append(args, v)
	}
	if v := filters["school"]; v != "" {
		where += " AND o.school = ?"
		args = append(args, v)
	}
	if v := filters["status"]; v != "" {
		where += " AND o.status = ?"
		args = append(args, v)
	}
	if v := filters["app_id"]; v != "" && v != "0" {
		where += " AND o.app_id = ?"
		args = append(args, v)
	}

	var total int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM w_order o WHERE "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	querySQL := fmt.Sprintf(`SELECT o.id, o.agg_order_id, o.user_id, COALESCE(o.school,''), o.account, o.password,
		o.app_id, COALESCE(a.name,''), o.status, o.num, o.cost, o.pause, o.sub_order, o.deleted, o.created, o.updated
		FROM w_order o LEFT JOIN w_app a ON o.app_id = a.id
		WHERE %s ORDER BY o.id DESC LIMIT ?, ?`, where)
	args = append(args, offset, pageSize)

	rows, err := database.DB.Query(querySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []WOrder
	for rows.Next() {
		var o WOrder
		var pauseInt, deletedInt int
		var aggOrderID *string
		var subOrderStr *string
		var createdTime, updatedTime time.Time
		err := rows.Scan(&o.ID, &aggOrderID, &o.UserID, &o.School, &o.Account, &o.Password,
			&o.AppID, &o.AppName, &o.Status, &o.Num, &o.Cost, &pauseInt, &subOrderStr, &deletedInt, &createdTime, &updatedTime)
		if err != nil {
			continue
		}
		o.AggOrderID = aggOrderID
		o.Pause = pauseInt == 1
		o.Deleted = deletedInt == 1
		o.Created = createdTime.Format("2006-01-02 15:04:05")
		o.Updated = updatedTime.Format("2006-01-02 15:04:05")
		if subOrderStr != nil {
			var sub interface{}
			json.Unmarshal([]byte(*subOrderStr), &sub)
			o.SubOrder = sub
		}
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []WOrder{}
	}
	return orders, total, nil
}

// AddOrder 创建订单
func (s *WService) AddOrder(uid int, data map[string]interface{}) (map[string]interface{}, error) {
	var addprice, money float64
	err := database.DB.QueryRow("SELECT addprice, money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice, &money)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	appID := int64(0)
	if v, ok := data["app_id"].(float64); ok {
		appID = int64(v)
	}
	school, _ := data["a_school"].(string)
	account, _ := data["a_account"].(string)
	password, _ := data["a_password"].(string)
	dis := 0.0
	if v, ok := data["dis"].(float64); ok {
		dis = v
	}
	taskList, _ := data["task_list"].([]interface{})
	num := len(taskList)

	if appID == 0 || account == "" || num <= 0 || dis <= 0 {
		return nil, fmt.Errorf("缺少必填参数")
	}

	// 查询项目
	app, err := s.getAppRow(appID)
	if err != nil {
		return nil, fmt.Errorf("项目不存在或已下架")
	}

	appPrice := 0.0
	if p, ok := app["price"].(string); ok {
		fmt.Sscanf(p, "%f", &appPrice)
	}
	cacType := fmt.Sprintf("%v", app["cac_type"])

	danjia := appPrice * addprice
	if danjia <= 0 || addprice < 0.1 {
		return nil, fmt.Errorf("单价异常，请联系管理员")
	}

	var orderMoney float64
	if cacType == "TS" {
		orderMoney = math.Round(float64(num)*danjia*100) / 100
	} else {
		orderMoney = math.Round(float64(num)*dis*danjia*100) / 100
	}

	if money < orderMoney {
		return nil, fmt.Errorf("余额不足")
	}

	// 保存原始请求数据用于重新提交
	subOrderJSON, _ := json.Marshal(data)

	// 写订单
	result, err := database.DB.Exec(
		`INSERT INTO w_order (agg_order_id, user_id, school, account, password, app_id, status, num, cost, pause, sub_order, deleted, created, updated)
		VALUES (NULL, ?, ?, ?, ?, ?, 'ADDING', ?, ?, 0, ?, 0, NOW(), NOW())`,
		uid, school, account, password, appID, num, orderMoney, string(subOrderJSON),
	)
	if err != nil {
		return nil, fmt.Errorf("下单失败请联系管理员")
	}
	orderID, _ := result.LastInsertId()

	// 扣余额
	res, err := database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ? AND money >= ? LIMIT 1", orderMoney, uid, orderMoney)
	if err != nil || func() bool { n, _ := res.RowsAffected(); return n <= 0 }() {
		// 余额不足，回滚订单
		database.DB.Exec("DELETE FROM w_order WHERE id = ? LIMIT 1", orderID)
		return nil, fmt.Errorf("余额不足")
	}

	appName := fmt.Sprintf("%v", app["name"])
	wLog(uid, "添加任务", fmt.Sprintf("项目：%s %s %s 扣除 %.2f 元", appName, account, password, orderMoney), -orderMoney)

	// 调用外部接口
	orgAppID := fmt.Sprintf("%v", app["org_app_id"])
	code := fmt.Sprintf("%v", app["code"])

	// 构造外部请求数据：替换 app_id 为上游的 org_app_id
	cpData := make(map[string]interface{})
	for k, v := range data {
		cpData[k] = v
	}
	cpData["app_id"] = orgAppID

	act := fmt.Sprintf("/%s/%s_order/add", code, code)
	externalResult, err := s.appRequest(app, act, cpData, "POST")

	if err == nil && externalResult != nil {
		code, _ := externalResult["code"].(float64)
		if int(code) == 0 {
			extData, _ := externalResult["data"].(map[string]interface{})
			if extData != nil {
				if aggID, ok := extData["agg_order_id"].(string); ok {
					extNum := num
					if n, ok := extData["num"].(float64); ok {
						extNum = int(n)
					}
					subJSON, _ := json.Marshal(extData["sub_order"])
					database.DB.Exec("UPDATE w_order SET agg_order_id = ?, status = 'NORMAL', num = ?, sub_order = ? WHERE id = ? LIMIT 1",
						aggID, extNum, string(subJSON), orderID)
				}
			}
		} else {
			// 外部下单失败
			database.DB.Exec("UPDATE w_order SET status = 'WAITADD' WHERE id = ? LIMIT 1", orderID)
		}
	} else {
		database.DB.Exec("UPDATE w_order SET status = 'WAITADD' WHERE id = ? LIMIT 1", orderID)
	}

	return map[string]interface{}{
		"id":   int(orderID),
		"cost": orderMoney,
	}, nil
}

// appRequest 根据项目类型发起请求
func (s *WService) appRequest(app map[string]interface{}, act string, body interface{}, method string) (map[string]interface{}, error) {
	pURL := strings.TrimSpace(fmt.Sprintf("%v", app["url"]))
	pType := fmt.Sprintf("%v", app["type"])
	token := fmt.Sprintf("%v", app["token"])
	key := fmt.Sprintf("%v", app["key"])
	pUID := fmt.Sprintf("%v", app["uid"])

	headers := map[string]string{}
	var reqURL string

	if pType == "0" {
		params := url.Values{}
		params.Set("act", formatAct(act))
		params.Set("key", key)
		params.Set("uid", pUID)
		reqURL = pURL + "?" + params.Encode()
	} else {
		reqURL = strings.TrimRight(pURL, "/") + act
		headers["X-WTK"] = token
	}

	if method == "" {
		method = "POST"
	}

	return s.httpReq(method, reqURL, body, headers)
}

func formatAct(orgAct string) string {
	return strings.ReplaceAll(strings.TrimLeft(orgAct, "/"), "/", "-")
}

func unformatAct(outAct string) string {
	s := strings.ReplaceAll(outAct, "-", "/")
	if len(s) > 0 && s[0] != '/' {
		s = "/" + s
	}
	return s
}

// RefundOrder 退款
func (s *WService) RefundOrder(uid int, wOrderID int, isAdmin bool) (map[string]interface{}, error) {
	var query string
	if isAdmin {
		query = "SELECT * FROM w_order WHERE id = ? LIMIT 1"
	} else {
		query = "SELECT * FROM w_order WHERE id = ? AND user_id = ? LIMIT 1"
	}
	rows, err := database.DB.Query(query, func() []interface{} {
		if isAdmin {
			return []interface{}{wOrderID}
		}
		return []interface{}{wOrderID, uid}
	}()...)
	if err != nil {
		return nil, fmt.Errorf("查询失败")
	}
	defer rows.Close()
	columns, _ := rows.Columns()
	if !rows.Next() {
		return nil, fmt.Errorf("订单不存在")
	}
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}
	rows.Scan(valuePtrs...)
	order := make(map[string]interface{})
	for i, col := range columns {
		val := values[i]
		if b, ok := val.([]byte); ok {
			order[col] = string(b)
		} else {
			order[col] = val
		}
	}
	rows.Close()

	deletedStr := fmt.Sprintf("%v", order["deleted"])
	if deletedStr == "1" {
		return nil, fmt.Errorf("该订单已删除，无法退款")
	}
	statusStr := fmt.Sprintf("%v", order["status"])
	if statusStr == "REFUND" {
		return nil, fmt.Errorf("该订单已退款，请勿重复操作")
	}

	aggOrderID := fmt.Sprintf("%v", order["agg_order_id"])
	noAggOrderID := aggOrderID == "<nil>" || aggOrderID == ""

	if noAggOrderID && statusStr != "WAITADD" {
		return nil, fmt.Errorf("该订单未提交到外部系统，且不是待下单状态，无法退款")
	}

	// 更新为待退款（乐观锁）
	res, _ := database.DB.Exec("UPDATE w_order SET status = 'WAITREFUND', updated = NOW() WHERE id = ? AND status = ? LIMIT 1", wOrderID, statusStr)
	if n, _ := res.RowsAffected(); n <= 0 {
		return nil, fmt.Errorf("订单状态已改变，请稍后再试")
	}

	appIDStr := fmt.Sprintf("%v", order["app_id"])
	appID := int64(0)
	fmt.Sscanf(appIDStr, "%d", &appID)
	app, err := s.getAppRow(appID)
	if err != nil {
		return nil, fmt.Errorf("项目不存在")
	}

	orderUserID := 0
	fmt.Sscanf(fmt.Sprintf("%v", order["user_id"]), "%d", &orderUserID)

	var refundAmount int
	var refundLogPrefix string

	if noAggOrderID {
		numStr := fmt.Sprintf("%v", order["num"])
		fmt.Sscanf(numStr, "%d", &refundAmount)
		refundLogPrefix = "外部无需退款"

		database.DB.Exec("UPDATE w_order SET status = 'REFUND', updated = NOW() WHERE id = ? LIMIT 1", wOrderID)
	} else {
		// 调用外部退款
		code := fmt.Sprintf("%v", app["code"])
		act := fmt.Sprintf("/%s/%s_order/refund", code, code)
		postData := map[string]interface{}{"agg_order_id": aggOrderID}

		externalResult, err := s.appRequest(app, act, postData, "POST")
		if err != nil || externalResult == nil {
			database.DB.Exec("UPDATE w_order SET status = 'REFUNDFAIL', updated = NOW() WHERE id = ? LIMIT 1", wOrderID)
			return nil, fmt.Errorf("源台退款失败")
		}
		extCode, _ := externalResult["code"].(float64)
		if int(extCode) != 0 {
			database.DB.Exec("UPDATE w_order SET status = 'REFUNDFAIL', updated = NOW() WHERE id = ? LIMIT 1", wOrderID)
			msg, _ := externalResult["msg"].(string)
			if msg == "" {
				msg = "源台退款失败"
			}
			return nil, fmt.Errorf("%s", msg)
		}

		extData, _ := externalResult["data"].(map[string]interface{})
		if extData != nil {
			if cnt, ok := extData["cnt"].(float64); ok {
				refundAmount = int(cnt)
			}
			// 同步订单状态
			if aggOrder, ok := extData["agg_order"].(map[string]interface{}); ok {
				s.syncOrderToDB(aggOrder)
			}
		}
		refundLogPrefix = "外部退款成功"
	}

	// 计算退款金额
	appPrice := 0.0
	if p, ok := app["price"].(string); ok {
		fmt.Sscanf(p, "%f", &appPrice)
	}
	cacType := fmt.Sprintf("%v", app["cac_type"])

	var orderAddprice float64
	database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", orderUserID).Scan(&orderAddprice)
	danjia := appPrice * orderAddprice

	// 从sub_order获取dis
	dis := 1.0
	subOrderStr := fmt.Sprintf("%v", order["sub_order"])
	var subData map[string]interface{}
	json.Unmarshal([]byte(subOrderStr), &subData)
	if d, ok := subData["dis"].(float64); ok && d > 0 {
		dis = d
	}

	orderCost := 0.0
	fmt.Sscanf(fmt.Sprintf("%v", order["cost"]), "%f", &orderCost)
	orderNum := 0
	fmt.Sscanf(fmt.Sprintf("%v", order["num"]), "%d", &orderNum)

	var refundMoney float64
	if cacType == "TS" {
		refundMoney = math.Round(float64(refundAmount)*danjia*100) / 100
	} else {
		refundMoney = math.Round(float64(refundAmount)*dis*danjia*100) / 100
	}
	// 取较小值
	if orderNum > 0 {
		alt := math.Round(orderCost*float64(refundAmount)/float64(orderNum)*100) / 100
		if alt < refundMoney {
			refundMoney = alt
		}
	}

	if refundMoney > 0 {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ? LIMIT 1", refundMoney, orderUserID)
	}

	wLog(orderUserID, "退款", fmt.Sprintf("订单 %d %s，退款金额 %.2f", wOrderID, refundLogPrefix, refundMoney), refundMoney)

	return map[string]interface{}{
		"refund_amount": refundMoney,
	}, nil
}

// SyncOrder 同步订单
func (s *WService) SyncOrder(uid int, wOrderID int, isAdmin bool) (string, error) {
	var query string
	if isAdmin {
		query = "SELECT * FROM w_order WHERE id = ? LIMIT 1"
	} else {
		query = "SELECT * FROM w_order WHERE id = ? AND user_id = ? LIMIT 1"
	}
	rows, err := database.DB.Query(query, func() []interface{} {
		if isAdmin {
			return []interface{}{wOrderID}
		}
		return []interface{}{wOrderID, uid}
	}()...)
	if err != nil {
		return "", fmt.Errorf("查询失败")
	}
	defer rows.Close()
	columns, _ := rows.Columns()
	if !rows.Next() {
		return "", fmt.Errorf("订单不存在")
	}
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}
	rows.Scan(valuePtrs...)
	order := make(map[string]interface{})
	for i, col := range columns {
		val := values[i]
		if b, ok := val.([]byte); ok {
			order[col] = string(b)
		} else {
			order[col] = val
		}
	}
	rows.Close()

	deletedStr := fmt.Sprintf("%v", order["deleted"])
	if deletedStr == "1" {
		return "", fmt.Errorf("该订单已删除，无法同步")
	}

	aggOrderID := fmt.Sprintf("%v", order["agg_order_id"])
	if aggOrderID == "<nil>" || aggOrderID == "" {
		return "", fmt.Errorf("该订单未提交到外部系统，无法同步")
	}

	appID := int64(0)
	fmt.Sscanf(fmt.Sprintf("%v", order["app_id"]), "%d", &appID)
	app, err := s.getAppRow(appID)
	if err != nil {
		return "", fmt.Errorf("项目不存在")
	}

	// 调用外部同步
	act := "/order/agg_order/view"
	pType := fmt.Sprintf("%v", app["type"])
	pURL := strings.TrimSpace(fmt.Sprintf("%v", app["url"]))
	token := fmt.Sprintf("%v", app["token"])
	key := fmt.Sprintf("%v", app["key"])
	pUID := fmt.Sprintf("%v", app["uid"])

	var reqURL string
	headers := map[string]string{}

	if pType == "0" {
		params := url.Values{}
		params.Set("act", formatAct(act))
		params.Set("key", key)
		params.Set("uid", pUID)
		params.Set("agg_order_id", aggOrderID)
		reqURL = pURL + "?" + params.Encode()
	} else {
		params := url.Values{}
		params.Set("agg_order_id", aggOrderID)
		params.Set("page_size", "1")
		reqURL = strings.TrimRight(pURL, "/") + act + "?" + params.Encode()
		headers["X-WTK"] = token
	}

	externalResult, err := s.httpReq("GET", reqURL, nil, headers)
	if err != nil {
		return "", fmt.Errorf("外部同步失败")
	}
	extCode, _ := externalResult["code"].(float64)
	if int(extCode) != 0 {
		msg, _ := externalResult["msg"].(string)
		if msg == "" {
			msg = "外部同步失败"
		}
		return "", fmt.Errorf("%s", msg)
	}

	extData, _ := externalResult["data"].(map[string]interface{})
	if extData == nil {
		return "", fmt.Errorf("外部接口返回数据格式错误")
	}
	dataList, _ := extData["list"].([]interface{})
	if len(dataList) == 0 {
		return "", fmt.Errorf("外部接口未查到该订单")
	}

	aggOrder, _ := dataList[0].(map[string]interface{})
	if aggOrder != nil {
		s.syncOrderToDB(aggOrder)
	}

	return "同步成功", nil
}

// syncOrderToDB 更新远程订单数据到本地
func (s *WService) syncOrderToDB(orderData map[string]interface{}) {
	aggOrderID, _ := orderData["agg_order_id"].(string)
	if aggOrderID == "" {
		return
	}

	var updateParts []string
	var updateArgs []interface{}

	if status, ok := orderData["status"].(string); ok {
		updateParts = append(updateParts, "`status` = ?")
		updateArgs = append(updateArgs, status)
	}
	if pause, ok := orderData["pause"].(bool); ok {
		p := 0
		if pause {
			p = 1
		}
		updateParts = append(updateParts, "`pause` = ?")
		updateArgs = append(updateArgs, p)
	}
	if subOrder := orderData["sub_order"]; subOrder != nil {
		subJSON, _ := json.Marshal(subOrder)
		updateParts = append(updateParts, "`sub_order` = ?")
		updateArgs = append(updateArgs, string(subJSON))
	}

	if len(updateParts) > 0 {
		if updated, ok := orderData["updated"].(string); ok && updated != "" {
			updateParts = append(updateParts, "`updated` = ?")
			updateArgs = append(updateArgs, updated)
		} else {
			updateParts = append(updateParts, "`updated` = NOW()")
		}
		updateArgs = append(updateArgs, aggOrderID)
		sql := "UPDATE w_order SET " + strings.Join(updateParts, ", ") + " WHERE agg_order_id = ? LIMIT 1"
		database.DB.Exec(sql, updateArgs...)
	}
}

// ResumeOrder 重新提交失败订单
func (s *WService) ResumeOrder(uid int, wOrderID int, isAdmin bool) (string, error) {
	var query string
	if isAdmin {
		query = "SELECT * FROM w_order WHERE id = ? LIMIT 1"
	} else {
		query = "SELECT * FROM w_order WHERE id = ? AND user_id = ? LIMIT 1"
	}
	rows, err := database.DB.Query(query, func() []interface{} {
		if isAdmin {
			return []interface{}{wOrderID}
		}
		return []interface{}{wOrderID, uid}
	}()...)
	if err != nil {
		return "", fmt.Errorf("查询失败")
	}
	defer rows.Close()
	columns, _ := rows.Columns()
	if !rows.Next() {
		return "", fmt.Errorf("订单不存在")
	}
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}
	rows.Scan(valuePtrs...)
	order := make(map[string]interface{})
	for i, col := range columns {
		val := values[i]
		if b, ok := val.([]byte); ok {
			order[col] = string(b)
		} else {
			order[col] = val
		}
	}
	rows.Close()

	deletedStr := fmt.Sprintf("%v", order["deleted"])
	if deletedStr == "1" {
		return "", fmt.Errorf("该订单已删除，无法重新提交")
	}

	// 乐观锁：只能从 WAITADD 状态重新提交
	res, _ := database.DB.Exec("UPDATE w_order SET status = 'ADDING', updated = NOW() WHERE id = ? AND status = 'WAITADD' LIMIT 1", wOrderID)
	if n, _ := res.RowsAffected(); n <= 0 {
		return "", fmt.Errorf("订单状态已改变，请稍后再试")
	}

	appID := int64(0)
	fmt.Sscanf(fmt.Sprintf("%v", order["app_id"]), "%d", &appID)
	app, err := s.getAppRow(appID)
	if err != nil {
		return "", fmt.Errorf("项目不存在")
	}

	code := fmt.Sprintf("%v", app["code"])
	act := fmt.Sprintf("/%s/%s_order/add", code, code)

	// sub_order 里保存了原始请求数据
	subOrderStr := fmt.Sprintf("%v", order["sub_order"])

	// 替换 app_id 为上游的 org_app_id
	var postData map[string]interface{}
	json.Unmarshal([]byte(subOrderStr), &postData)
	if postData == nil {
		database.DB.Exec("UPDATE w_order SET status = 'WAITADD' WHERE id = ? LIMIT 1", wOrderID)
		return "", fmt.Errorf("缺少原始请求数据，无法重新提交")
	}
	postData["app_id"] = fmt.Sprintf("%v", app["org_app_id"])

	externalResult, err := s.appRequest(app, act, postData, "POST")
	if err == nil && externalResult != nil {
		extCode, _ := externalResult["code"].(float64)
		if int(extCode) == 0 {
			extData, _ := externalResult["data"].(map[string]interface{})
			if extData != nil {
				if aggID, ok := extData["agg_order_id"].(string); ok {
					extNum := 0
					if n, ok := extData["num"].(float64); ok {
						extNum = int(n)
					}
					subJSON, _ := json.Marshal(extData["sub_order"])
					database.DB.Exec("UPDATE w_order SET agg_order_id = ?, status = 'NORMAL', num = ?, sub_order = ? WHERE id = ? LIMIT 1",
						aggID, extNum, string(subJSON), wOrderID)
				}
			}
			msg, _ := externalResult["msg"].(string)
			if msg == "" {
				msg = "重新提交成功"
			}
			return msg, nil
		}
	}

	database.DB.Exec("UPDATE w_order SET status = 'WAITADD' WHERE id = ? LIMIT 1", wOrderID)
	msg := "源台下单失败"
	if externalResult != nil {
		if m, ok := externalResult["msg"].(string); ok && m != "" {
			msg = m
		}
	}
	return "", fmt.Errorf("本地下单成功，%s", msg)
}

// ========== 管理员项目管理 ==========

// AdminListApps 管理员获取所有项目
func (s *WService) AdminListApps() ([]WApp, error) {
	rows, err := database.DB.Query("SELECT id, name, code, org_app_id, status, COALESCE(description,''), price, cac_type, url, COALESCE(`key`,''), COALESCE(uid,''), COALESCE(token,''), type, deleted FROM w_app WHERE deleted = 0 ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []WApp
	for rows.Next() {
		var a WApp
		if err := rows.Scan(&a.ID, &a.Name, &a.Code, &a.OrgAppID, &a.Status, &a.Desc, &a.Price, &a.CacType, &a.URL, &a.Key, &a.UID, &a.Token, &a.Type, &a.Deleted); err != nil {
			continue
		}
		list = append(list, a)
	}
	if list == nil {
		list = []WApp{}
	}
	return list, nil
}

// AdminSaveApp 管理员添加或编辑项目
func (s *WService) AdminSaveApp(a WApp) (int64, error) {
	if a.Name == "" || a.Code == "" {
		return 0, fmt.Errorf("项目名称和代码不能为空")
	}
	if a.ID > 0 {
		_, err := database.DB.Exec(
			"UPDATE w_app SET name=?, code=?, org_app_id=?, status=?, description=?, price=?, cac_type=?, url=?, `key`=?, uid=?, token=?, type=? WHERE id=?",
			a.Name, a.Code, a.OrgAppID, a.Status, a.Desc, a.Price, a.CacType, a.URL, a.Key, a.UID, a.Token, a.Type, a.ID,
		)
		if err != nil {
			return 0, fmt.Errorf("保存失败: %v", err)
		}
		return a.ID, nil
	}
	result, err := database.DB.Exec(
		"INSERT INTO w_app (name, code, org_app_id, status, description, price, cac_type, url, `key`, uid, token, type) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		a.Name, a.Code, a.OrgAppID, a.Status, a.Desc, a.Price, a.CacType, a.URL, a.Key, a.UID, a.Token, a.Type,
	)
	if err != nil {
		return 0, fmt.Errorf("添加失败: %v", err)
	}
	id, _ := result.LastInsertId()
	return id, nil
}

// AdminDeleteApp 管理员删除项目（软删除）
func (s *WService) AdminDeleteApp(id int64) error {
	_, err := database.DB.Exec("UPDATE w_app SET deleted = 1 WHERE id = ?", id)
	return err
}
