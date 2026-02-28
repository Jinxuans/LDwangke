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

// XMProject 小米运动项目
type XMProject struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Query       int     `json:"query"`
	Password    int     `json:"password"`
}

// XMOrder 小米运动订单
type XMOrder struct {
	ID        int         `json:"id"`
	UserID    int         `json:"user_id"`
	School    string      `json:"school"`
	Account   string      `json:"account"`
	Password  string      `json:"password"`
	ProjectID int         `json:"project_id"`
	Status    string      `json:"status_name"`
	Type      interface{} `json:"type"`
	TotalKM   int         `json:"total_km"`
	IsDeleted bool        `json:"is_deleted"`
	RunKM     *float64    `json:"run_km"`
	RunDate   interface{} `json:"run_date"`
	StartDay  string      `json:"start_day"`
	StartTime string      `json:"start_time"`
	EndTime   string      `json:"end_time"`
	Deduction float64     `json:"deduction"`
	UpdatedAt string      `json:"updated_at"`
}

type XMService struct {
	client *http.Client
}

func NewXMService() *XMService {
	return &XMService{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// ---------- HTTP 工具 ----------

func (s *XMService) httpRequest(method, reqURL string, body interface{}, headers map[string]string) (map[string]interface{}, error) {
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		code := -1
		if c, ok := result["code"].(float64); ok {
			code = int(c)
		}
		msg := fmt.Sprintf("外部接口 HTTP 状态异常: %d", resp.StatusCode)
		if m, ok := result["msg"].(string); ok && m != "" {
			msg = m
		}
		return map[string]interface{}{"code": float64(code), "msg": msg, "data": result["data"]}, nil
	}

	return result, nil
}

// projectRequest 根据项目类型发起请求
func (s *XMService) projectRequest(project map[string]interface{}, act string, body interface{}, method string) (map[string]interface{}, error) {
	pURL := strings.TrimSpace(fmt.Sprintf("%v", project["url"]))
	pType := 0
	if t, ok := project["type"].(int64); ok {
		pType = int(t)
	} else if t, ok := project["type"].([]uint8); ok {
		pType = int(t[0]) - '0'
	} else if t, ok := project["type"].(float64); ok {
		pType = int(t)
	}
	token := fmt.Sprintf("%v", project["token"])
	key := fmt.Sprintf("%v", project["key"])
	pUID := fmt.Sprintf("%v", project["uid"])

	headers := map[string]string{}
	var reqURL string

	if pType == 0 {
		// type=0: key/uid auth via query params
		params := url.Values{}
		params.Set("act", act)
		params.Set("key", key)
		params.Set("uid", pUID)
		reqURL = pURL + "?" + params.Encode()
	} else {
		// type=1: token auth via header
		reqURL = strings.TrimRight(pURL, "/") + "/" + act
		headers["token"] = token
	}

	if method == "" {
		method = "POST"
	}

	return s.httpRequest(method, reqURL, body, headers)
}

// ---------- 日志 ----------

func xmLog(uid int, logType, text string, money float64) {
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

// ---------- 查询项目信息（内部） ----------

func (s *XMService) getProjectRow(projectID int) (map[string]interface{}, error) {
	rows, err := database.DB.Query("SELECT * FROM xm_project WHERE id = ? LIMIT 1", projectID)
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

// GetProjects 获取项目列表
func (s *XMService) GetProjects(uid int) ([]XMProject, error) {
	var addprice float64
	err := database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	rows, err := database.DB.Query("SELECT id, name, description, price, `query`, password FROM xm_project WHERE is_deleted = 0 AND status = 0")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []XMProject
	for rows.Next() {
		var p XMProject
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Query, &p.Password); err != nil {
			continue
		}
		p.Price = math.Round(p.Price*addprice*100) / 100
		projects = append(projects, p)
	}
	if projects == nil {
		projects = []XMProject{}
	}
	return projects, nil
}

// AddOrder 创建跑步订单
func (s *XMService) AddOrder(uid int, data map[string]interface{}) (map[string]interface{}, error) {
	var addprice, money float64
	err := database.DB.QueryRow("SELECT addprice, money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice, &money)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	projectID := 0
	if v, ok := data["project_id"].(float64); ok {
		projectID = int(v)
	}
	school, _ := data["school"].(string)
	account, _ := data["account"].(string)
	password, _ := data["password"].(string)
	totalKM := 0
	if v, ok := data["total_km"].(float64); ok {
		totalKM = int(v)
	}
	runDateArray := data["run_date"]
	startDay, _ := data["start_day"].(string)
	startTime, _ := data["start_time"].(string)
	endTime, _ := data["end_time"].(string)

	var orderType *int
	if v, ok := data["type"].(float64); ok {
		t := int(v)
		orderType = &t
	}

	if projectID == 0 || school == "" || account == "" || totalKM == 0 || startDay == "" || startTime == "" || endTime == "" {
		return nil, fmt.Errorf("缺少必填参数")
	}
	if runDateArray == nil {
		return nil, fmt.Errorf("缺少必填参数")
	}

	// 查询项目
	project, err := s.getProjectRow(projectID)
	if err != nil {
		return nil, fmt.Errorf("项目不存在")
	}

	projectPrice := 0.0
	if p, ok := project["price"].(string); ok {
		fmt.Sscanf(p, "%f", &projectPrice)
	} else if p, ok := project["price"].(float64); ok {
		projectPrice = p
	}

	danjia := math.Round(projectPrice*addprice*100) / 100
	if danjia <= 0 || addprice < 0.1 {
		return nil, fmt.Errorf("单价异常，请联系管理员")
	}
	orderMoney := math.Round(float64(totalKM)*danjia*100) / 100

	if money < orderMoney {
		return nil, fmt.Errorf("余额不足")
	}

	// 写订单
	runDateJSON, _ := json.Marshal(runDateArray)
	var typeSQL interface{}
	if orderType != nil {
		typeSQL = *orderType
	}

	result, err := database.DB.Exec(
		`INSERT INTO xm_order (y_oid, user_id, school, account, password, type, project_id, status, total_km, run_km, run_date, start_day, start_time, end_time, deduction, is_deleted, created_at, updated_at)
		VALUES (NULL, ?, ?, ?, ?, ?, ?, '已下单', ?, NULL, ?, ?, ?, ?, ?, 0, NOW(), NOW())`,
		uid, school, account, password, typeSQL, projectID, totalKM, string(runDateJSON), startDay, startTime, endTime, orderMoney,
	)
	if err != nil {
		return nil, fmt.Errorf("下单失败请联系管理员")
	}
	orderID, _ := result.LastInsertId()

	// 扣余额
	database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ? LIMIT 1", orderMoney, uid)

	projectName, _ := project["name"].(string)
	xmLog(uid, "添加任务", fmt.Sprintf("项目：%s %s %s 扣除 %.2f 元", projectName, account, password, orderMoney), -orderMoney)

	// 调用外部接口
	pID, _ := project["p_id"].(string)
	if pID == "" {
		if v, ok := project["p_id"].(int64); ok {
			pID = fmt.Sprintf("%d", v)
		}
	}

	postData := map[string]interface{}{
		"project_id": pID,
		"school":     school,
		"account":    account,
		"password":   password,
		"total_km":   totalKM,
		"run_date":   runDateArray,
		"start_day":  startDay,
		"start_time": startTime,
		"end_time":   endTime,
		"type":       orderType,
	}

	externalResult, err := s.projectRequest(project, "add_order", postData, "POST")
	if err == nil && externalResult != nil {
		code, _ := externalResult["code"].(float64)
		if int(code) == 200 {
			extData, _ := externalResult["data"].(map[string]interface{})
			if extData != nil {
				if extID, ok := extData["id"].(float64); ok {
					database.DB.Exec("UPDATE xm_order SET y_oid = ?, status = '已提交' WHERE id = ? LIMIT 1", int(extID), orderID)
				}
			}
		}
	}

	// 类型映射
	typeMapping := map[int]string{0: "计分按次", 1: "计分按公里", 2: "晨跑按次", 3: "晨跑按公里"}
	var typeStr interface{}
	if orderType != nil {
		if s, ok := typeMapping[*orderType]; ok {
			typeStr = s
		}
	}

	return map[string]interface{}{
		"id":          int(orderID),
		"user_id":     uid,
		"school":      school,
		"account":     account,
		"password":    password,
		"project_id":  projectID,
		"status_name": "已提交",
		"type":        typeStr,
		"total_km":    totalKM,
		"is_deleted":  false,
		"run_km":      nil,
		"run_date":    runDateArray,
		"start_day":   startDay,
		"start_time":  startTime,
		"end_time":    endTime,
		"deduction":   orderMoney,
		"updated_at":  time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// GetOrders 查询订单列表
func (s *XMService) GetOrders(uid int, isAdmin bool, page, pageSize int, filters map[string]string) ([]XMOrder, int, error) {
	offset := (page - 1) * pageSize
	where := "is_deleted = 0"
	var args []interface{}

	if !isAdmin {
		where += " AND user_id = ?"
		args = append(args, uid)
	}
	if v := filters["account"]; v != "" {
		where += " AND account = ?"
		args = append(args, v)
	}
	if v := filters["school"]; v != "" {
		where += " AND school = ?"
		args = append(args, v)
	}
	if v := filters["status"]; v != "" {
		where += " AND status = ?"
		args = append(args, v)
	}
	if v := filters["project"]; v != "" && v != "0" {
		where += " AND project_id = ?"
		args = append(args, v)
	}
	if v := filters["order_id"]; v != "" {
		where += " AND id = ?"
		args = append(args, v)
	}

	var total int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM xm_order WHERE "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	querySQL := fmt.Sprintf("SELECT id, user_id, school, account, password, project_id, status, COALESCE(type,-1), total_km, is_deleted, run_km, run_date, start_day, start_time, end_time, deduction, updated_at FROM xm_order WHERE %s ORDER BY id DESC LIMIT ?, ?", where)
	args = append(args, offset, pageSize)

	rows, err := database.DB.Query(querySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	typeMapping := map[int]string{0: "计分按次", 1: "计分按公里", 2: "晨跑按次", 3: "晨跑按公里"}

	var orders []XMOrder
	for rows.Next() {
		var o XMOrder
		var isDeletedInt int
		var runDateStr string
		var typeInt int
		var updatedAtTime time.Time
		err := rows.Scan(&o.ID, &o.UserID, &o.School, &o.Account, &o.Password,
			&o.ProjectID, &o.Status, &typeInt, &o.TotalKM, &isDeletedInt,
			&o.RunKM, &runDateStr, &o.StartDay, &o.StartTime, &o.EndTime,
			&o.Deduction, &updatedAtTime)
		if err != nil {
			continue
		}
		o.IsDeleted = isDeletedInt == 1
		o.UpdatedAt = updatedAtTime.Format("2006-01-02 15:04:05")
		if typeInt >= 0 {
			if s, ok := typeMapping[typeInt]; ok {
				o.Type = s
			}
		} else {
			o.Type = nil
		}
		// 解析 run_date JSON
		var runDate interface{}
		json.Unmarshal([]byte(runDateStr), &runDate)
		o.RunDate = runDate

		orders = append(orders, o)
	}
	if orders == nil {
		orders = []XMOrder{}
	}
	return orders, total, nil
}

// QueryRun 查询跑步状态（代理上游）
func (s *XMService) QueryRun(uid int, data map[string]interface{}) (interface{}, error) {
	projectID := 0
	if v, ok := data["project_id"].(float64); ok {
		projectID = int(v)
	}
	account, _ := data["account"].(string)

	if account == "" || projectID == 0 {
		return nil, fmt.Errorf("缺少必填参数")
	}

	project, err := s.getProjectRow(projectID)
	if err != nil {
		return nil, fmt.Errorf("项目不存在")
	}

	pID, _ := project["p_id"].(string)
	if pID == "" {
		if v, ok := project["p_id"].(int64); ok {
			pID = fmt.Sprintf("%d", v)
		}
	}

	postData := map[string]interface{}{
		"account":    data["account"],
		"password":   data["password"],
		"project_id": pID,
	}

	result, err := s.projectRequest(project, "query_run", postData, "POST")
	if err != nil {
		return nil, err
	}
	return result, nil
}

// RefundOrder 退款
func (s *XMService) RefundOrder(uid int, orderID int, isAdmin bool) (map[string]interface{}, error) {
	// 查询订单
	var query string
	if isAdmin {
		query = "SELECT * FROM xm_order WHERE id = ? LIMIT 1"
	} else {
		query = "SELECT * FROM xm_order WHERE id = ? AND user_id = ? LIMIT 1"
	}

	rows, err := database.DB.Query(query, func() []interface{} {
		if isAdmin {
			return []interface{}{orderID}
		}
		return []interface{}{orderID, uid}
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

	isDeleted := fmt.Sprintf("%v", order["is_deleted"])
	if isDeleted == "1" {
		return nil, fmt.Errorf("该订单已删除，无法退款")
	}
	statusStr := fmt.Sprintf("%v", order["status"])
	if statusStr == "已退款" {
		return nil, fmt.Errorf("该订单已退款，请勿重复操作")
	}

	// 更新为待退款
	database.DB.Exec("UPDATE xm_order SET status = '待退款', updated_at = NOW() WHERE id = ? LIMIT 1", orderID)

	// 获取项目信息
	projectIDStr := fmt.Sprintf("%v", order["project_id"])
	projectID := 0
	fmt.Sscanf(projectIDStr, "%d", &projectID)
	project, err := s.getProjectRow(projectID)
	if err != nil {
		return nil, fmt.Errorf("项目不存在")
	}

	orderUserID := 0
	userIDStr := fmt.Sprintf("%v", order["user_id"])
	fmt.Sscanf(userIDStr, "%d", &orderUserID)

	yOidStr := fmt.Sprintf("%v", order["y_oid"])
	var refundKM float64
	var danjia float64

	if yOidStr == "<nil>" || yOidStr == "" || yOidStr == "0" {
		// 没有外部订单号，按本地数据退款
		totalKMStr := fmt.Sprintf("%v", order["total_km"])
		fmt.Sscanf(totalKMStr, "%f", &refundKM)

		var addprice float64
		database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", orderUserID).Scan(&addprice)
		projectPrice := 0.0
		if p, ok := project["price"].(string); ok {
			fmt.Sscanf(p, "%f", &projectPrice)
		} else if p, ok := project["price"].(float64); ok {
			projectPrice = p
		}
		danjia = math.Round(projectPrice*addprice*100) / 100
		if danjia <= 0 || addprice < 0.1 {
			return nil, fmt.Errorf("单价异常，请联系管理员")
		}
	} else {
		// 有外部订单号，调用外部退款
		externalOID := 0
		fmt.Sscanf(yOidStr, "%d", &externalOID)
		if externalOID <= 0 {
			return nil, fmt.Errorf("该订单未提交到外部系统，无法退款")
		}

		var externalResult map[string]interface{}

		pType := 0
		if t, ok := project["type"].(string); ok {
			fmt.Sscanf(t, "%d", &pType)
		} else if t, ok := project["type"].(int64); ok {
			pType = int(t)
		}

		if pType == 0 {
			pURL := fmt.Sprintf("%v", project["url"])
			key := fmt.Sprintf("%v", project["key"])
			pUID := fmt.Sprintf("%v", project["uid"])
			params := url.Values{}
			params.Set("act", "refund_order")
			params.Set("key", key)
			params.Set("uid", pUID)
			params.Set("order_id", fmt.Sprintf("%d", externalOID))
			queryURL := pURL + "?" + params.Encode()
			externalResult, err = s.httpRequest("GET", queryURL, nil, nil)
		} else {
			pURL := strings.TrimRight(fmt.Sprintf("%v", project["url"]), "/")
			token := fmt.Sprintf("%v", project["token"])
			params := url.Values{}
			params.Set("order_id", fmt.Sprintf("%d", externalOID))
			queryURL := pURL + "/refund?" + params.Encode()
			externalResult, err = s.httpRequest("GET", queryURL, nil, map[string]string{"token": token})
		}

		if err != nil || externalResult == nil {
			database.DB.Exec("UPDATE xm_order SET status = '退款失败', updated_at = NOW() WHERE id = ? LIMIT 1", orderID)
			return nil, fmt.Errorf("源台退款失败")
		}
		code, _ := externalResult["code"].(float64)
		if int(code) != 200 {
			database.DB.Exec("UPDATE xm_order SET status = '退款失败', updated_at = NOW() WHERE id = ? LIMIT 1", orderID)
			msg, _ := externalResult["msg"].(string)
			if msg == "" {
				msg = "源台退款失败"
			}
			return nil, fmt.Errorf("%s", msg)
		}

		extData, _ := externalResult["data"].(map[string]interface{})
		if extData != nil {
			if rk, ok := extData["refund_km"].(float64); ok {
				refundKM = rk
			}
		}

		var addprice float64
		database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", orderUserID).Scan(&addprice)
		projectPrice := 0.0
		if p, ok := project["price"].(string); ok {
			fmt.Sscanf(p, "%f", &projectPrice)
		} else if p, ok := project["price"].(float64); ok {
			projectPrice = p
		}
		danjia = math.Round(projectPrice*addprice*100) / 100
		if danjia <= 0 || addprice < 0.1 {
			return nil, fmt.Errorf("单价异常，请联系管理员")
		}
	}

	refundMoney := math.Round(refundKM*danjia*100) / 100
	if refundMoney > 0 {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ? LIMIT 1", refundMoney, orderUserID)
	}
	database.DB.Exec("UPDATE xm_order SET status = '已退款', updated_at = NOW() WHERE id = ? LIMIT 1", orderID)
	xmLog(orderUserID, "退款", fmt.Sprintf("订单 %d 退款成功，退 %.0f km，退款金额 %.2f", orderID, refundKM, refundMoney), refundMoney)

	return map[string]interface{}{
		"refund_amount": refundMoney,
		"refund_km":     refundKM,
	}, nil
}

// DeleteOrder 删除订单
func (s *XMService) DeleteOrder(uid int, orderID int, isAdmin bool) (string, error) {
	var query string
	if isAdmin {
		query = "SELECT * FROM xm_order WHERE id = ? LIMIT 1"
	} else {
		query = "SELECT * FROM xm_order WHERE id = ? AND user_id = ? LIMIT 1"
	}

	rows, err := database.DB.Query(query, func() []interface{} {
		if isAdmin {
			return []interface{}{orderID}
		}
		return []interface{}{orderID, uid}
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

	isDeleted := fmt.Sprintf("%v", order["is_deleted"])
	if isDeleted == "1" {
		return "", fmt.Errorf("该订单已删除")
	}

	project, err := s.getProjectRow(func() int {
		v := 0
		fmt.Sscanf(fmt.Sprintf("%v", order["project_id"]), "%d", &v)
		return v
	}())
	if err != nil {
		return "", fmt.Errorf("项目不存在")
	}

	yOidStr := fmt.Sprintf("%v", order["y_oid"])
	externalOID := 0
	fmt.Sscanf(yOidStr, "%d", &externalOID)
	if externalOID <= 0 {
		return "", fmt.Errorf("该订单未提交到外部系统，无法删除")
	}

	pType := 0
	if t, ok := project["type"].(string); ok {
		fmt.Sscanf(t, "%d", &pType)
	} else if t, ok := project["type"].(int64); ok {
		pType = int(t)
	}

	var externalResult map[string]interface{}
	if pType == 0 {
		pURL := fmt.Sprintf("%v", project["url"])
		key := fmt.Sprintf("%v", project["key"])
		pUID := fmt.Sprintf("%v", project["uid"])
		params := url.Values{}
		params.Set("act", "delete_order")
		params.Set("key", key)
		params.Set("uid", pUID)
		params.Set("order_id", fmt.Sprintf("%d", externalOID))
		queryURL := pURL + "?" + params.Encode()
		externalResult, err = s.httpRequest("GET", queryURL, nil, nil)
	} else {
		pURL := strings.TrimRight(fmt.Sprintf("%v", project["url"]), "/")
		token := fmt.Sprintf("%v", project["token"])
		externalResult, err = s.httpRequest("DELETE", pURL+"/delete", map[string]interface{}{"order_id": externalOID}, map[string]string{"token": token})
	}

	if err != nil || externalResult == nil {
		return "", fmt.Errorf("外部接口删除失败")
	}
	code, _ := externalResult["code"].(float64)
	if int(code) != 200 {
		msg, _ := externalResult["msg"].(string)
		if msg == "" {
			msg = "外部接口删除失败"
		}
		return "", fmt.Errorf("%s", msg)
	}

	database.DB.Exec("UPDATE xm_order SET is_deleted = 1, status = '已删除', updated_at = NOW() WHERE id = ? LIMIT 1", orderID)
	orderUserID := 0
	fmt.Sscanf(fmt.Sprintf("%v", order["user_id"]), "%d", &orderUserID)
	xmLog(orderUserID, "删除订单", fmt.Sprintf("删除订单ID: %d (外部删除成功)", orderID), 0)

	msg, _ := externalResult["msg"].(string)
	if msg == "" {
		msg = "删除成功"
	}
	return msg, nil
}

// SyncOrder 同步订单
func (s *XMService) SyncOrder(uid int, orderID int, isAdmin bool) (string, error) {
	var query string
	if isAdmin {
		query = "SELECT * FROM xm_order WHERE id = ? LIMIT 1"
	} else {
		query = "SELECT * FROM xm_order WHERE id = ? AND user_id = ? LIMIT 1"
	}

	rows, err := database.DB.Query(query, func() []interface{} {
		if isAdmin {
			return []interface{}{orderID}
		}
		return []interface{}{orderID, uid}
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

	isDeleted := fmt.Sprintf("%v", order["is_deleted"])
	if isDeleted == "1" {
		return "", fmt.Errorf("该订单已删除，无法同步")
	}

	yOidStr := fmt.Sprintf("%v", order["y_oid"])
	yOid := 0
	fmt.Sscanf(yOidStr, "%d", &yOid)
	if yOid <= 0 {
		return "", fmt.Errorf("该订单未提交到外部系统，无法同步")
	}

	projectID := 0
	fmt.Sscanf(fmt.Sprintf("%v", order["project_id"]), "%d", &projectID)
	project, err := s.getProjectRow(projectID)
	if err != nil {
		return "", fmt.Errorf("项目不存在")
	}

	// 同步
	externalResult, err := s.syncOrderFromUpstream(yOid, project)
	if err != nil {
		return "", err
	}
	code, _ := externalResult["code"].(float64)
	if int(code) != 200 {
		msg, _ := externalResult["msg"].(string)
		if msg == "" {
			msg = "外部同步失败"
		}
		return "", fmt.Errorf("%s", msg)
	}

	msg, _ := externalResult["msg"].(string)
	if msg == "" {
		msg = "同步成功"
	}
	return msg, nil
}

// syncOrderFromUpstream 从上游同步订单数据
func (s *XMService) syncOrderFromUpstream(yOid int, project map[string]interface{}) (map[string]interface{}, error) {
	pType := 0
	if t, ok := project["type"].(string); ok {
		fmt.Sscanf(t, "%d", &pType)
	} else if t, ok := project["type"].(int64); ok {
		pType = int(t)
	}

	var externalResult map[string]interface{}
	var err error

	if pType == 0 {
		pURL := fmt.Sprintf("%v", project["url"])
		key := fmt.Sprintf("%v", project["key"])
		pUID := fmt.Sprintf("%v", project["uid"])
		params := url.Values{}
		params.Set("act", "get_orders")
		params.Set("key", key)
		params.Set("uid", pUID)
		params.Set("order_id", fmt.Sprintf("%d", yOid))
		queryURL := pURL + "?" + params.Encode()
		externalResult, err = s.httpRequest("GET", queryURL, nil, nil)
	} else {
		pURL := strings.TrimRight(fmt.Sprintf("%v", project["url"]), "/")
		token := fmt.Sprintf("%v", project["token"])
		params := url.Values{}
		params.Set("id", fmt.Sprintf("%d", yOid))
		params.Set("page", "1")
		params.Set("page_size", "10")
		queryURL := pURL + "/list?" + params.Encode()
		externalResult, err = s.httpRequest("GET", queryURL, nil, map[string]string{"token": token})
	}

	if err != nil {
		return nil, err
	}
	code, _ := externalResult["code"].(float64)
	if int(code) != 200 {
		return externalResult, nil
	}

	dataList, _ := externalResult["data"].([]interface{})
	if dataList == nil {
		return map[string]interface{}{"code": float64(-1), "msg": "外部接口返回的数据格式错误"}, nil
	}

	skipFields := map[string]bool{"id": true, "user_id": true, "school": true, "account": true, "password": true, "project_id": true, "type": true, "deduction": true}

	for _, item := range dataList {
		row, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var updateParts []string
		var updateArgs []interface{}

		for field, value := range row {
			if skipFields[field] {
				continue
			}
			if field == "status_name" {
				updateParts = append(updateParts, "`status` = ?")
				updateArgs = append(updateArgs, value)
			} else if field == "run_date" {
				jsonVal, _ := json.Marshal(value)
				updateParts = append(updateParts, "`run_date` = ?")
				updateArgs = append(updateArgs, string(jsonVal))
			} else if field == "is_deleted" {
				boolVal := 0
				if v, ok := value.(bool); ok && v {
					boolVal = 1
				}
				updateParts = append(updateParts, "`is_deleted` = ?")
				updateArgs = append(updateArgs, boolVal)
			} else if field == "run_km" {
				if value == nil {
					updateParts = append(updateParts, "`run_km` = NULL")
				} else {
					updateParts = append(updateParts, "`run_km` = ?")
					updateArgs = append(updateArgs, value)
				}
			} else {
				updateParts = append(updateParts, fmt.Sprintf("`%s` = ?", field))
				updateArgs = append(updateArgs, value)
			}
		}

		if len(updateParts) > 0 {
			if updatedAt, ok := row["updated_at"].(string); ok && updatedAt != "" {
				updateParts = append(updateParts, "`updated_at` = ?")
				updateArgs = append(updateArgs, updatedAt)
			} else {
				updateParts = append(updateParts, "`updated_at` = NOW()")
			}
			updateArgs = append(updateArgs, yOid)
			sql := "UPDATE xm_order SET " + strings.Join(updateParts, ", ") + " WHERE y_oid = ? LIMIT 1"
			database.DB.Exec(sql, updateArgs...)
		}
	}

	return map[string]interface{}{"code": float64(200), "msg": "同步成功", "data": dataList}, nil
}

// GetOrderLogs 获取订单日志
func (s *XMService) GetOrderLogs(uid int, orderID int, isAdmin bool, page, pageSize int) (interface{}, error) {
	var query string
	if isAdmin {
		query = "SELECT * FROM xm_order WHERE id = ? LIMIT 1"
	} else {
		query = "SELECT * FROM xm_order WHERE id = ? AND user_id = ? LIMIT 1"
	}

	rows, err := database.DB.Query(query, func() []interface{} {
		if isAdmin {
			return []interface{}{orderID}
		}
		return []interface{}{orderID, uid}
	}()...)
	if err != nil {
		return nil, fmt.Errorf("查询失败")
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	if !rows.Next() {
		return nil, fmt.Errorf("订单不存在或无权限查看")
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

	yOidStr := fmt.Sprintf("%v", order["y_oid"])
	if yOidStr == "<nil>" || yOidStr == "" || yOidStr == "0" {
		return nil, fmt.Errorf("该订单未提交到外部系统，无法查询日志")
	}

	projectID := 0
	fmt.Sscanf(fmt.Sprintf("%v", order["project_id"]), "%d", &projectID)
	project, err := s.getProjectRow(projectID)
	if err != nil {
		return nil, fmt.Errorf("项目不存在")
	}

	pType := 0
	if t, ok := project["type"].(string); ok {
		fmt.Sscanf(t, "%d", &pType)
	} else if t, ok := project["type"].(int64); ok {
		pType = int(t)
	}

	headers := map[string]string{}
	params := url.Values{}
	params.Set("order_id", yOidStr)
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("page_size", fmt.Sprintf("%d", pageSize))

	var queryURL string
	if pType == 0 {
		pURL := fmt.Sprintf("%v", project["url"])
		key := fmt.Sprintf("%v", project["key"])
		pUID := fmt.Sprintf("%v", project["uid"])
		params.Set("act", "get_order_logs")
		params.Set("key", key)
		params.Set("uid", pUID)
		queryURL = pURL + "?" + params.Encode()
	} else {
		token := fmt.Sprintf("%v", project["token"])
		baseURL := "https://66-dd.com/api/v1/runorderlog/log"
		headers = map[string]string{
			"token":        token,
			"Content-Type": "application/json",
			"Accept":       "application/json, text/plain, */*",
		}
		queryURL = baseURL + "?" + params.Encode()
	}

	externalResult, err := s.httpRequest("GET", queryURL, nil, headers)
	if err != nil {
		return nil, fmt.Errorf("外部接口响应格式异常")
	}
	code, _ := externalResult["code"].(float64)
	if int(code) != 200 {
		msg, _ := externalResult["msg"].(string)
		if msg == "" {
			msg = "获取日志失败"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	return externalResult, nil
}

// ========== 管理员项目管理 ==========

// XMProjectAdmin 管理员视角的项目信息（含对接配置）
type XMProjectAdmin struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Query       int     `json:"query"`
	Password    int     `json:"password"`
	URL         string  `json:"url"`
	UID         string  `json:"uid"`
	Key         string  `json:"key"`
	Token       string  `json:"token"`
	Type        int     `json:"type"`
	PID         string  `json:"p_id"`
	Status      int     `json:"status"`
}

// AdminListProjects 管理员获取所有项目（含对接配置）
func (s *XMService) AdminListProjects() ([]XMProjectAdmin, error) {
	rows, err := database.DB.Query("SELECT id, name, COALESCE(description,''), price, `query`, password, url, uid, `key`, token, type, p_id, status FROM xm_project WHERE is_deleted = 0 ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []XMProjectAdmin
	for rows.Next() {
		var p XMProjectAdmin
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Query, &p.Password, &p.URL, &p.UID, &p.Key, &p.Token, &p.Type, &p.PID, &p.Status); err != nil {
			continue
		}
		list = append(list, p)
	}
	if list == nil {
		list = []XMProjectAdmin{}
	}
	return list, nil
}

// AdminSaveProject 管理员添加或编辑项目
func (s *XMService) AdminSaveProject(p XMProjectAdmin) (int, error) {
	if p.Name == "" {
		return 0, fmt.Errorf("项目名称不能为空")
	}
	if p.ID > 0 {
		// 编辑
		_, err := database.DB.Exec(
			"UPDATE xm_project SET name=?, description=?, price=?, `query`=?, password=?, url=?, uid=?, `key`=?, token=?, type=?, p_id=?, status=? WHERE id=?",
			p.Name, p.Description, p.Price, p.Query, p.Password, p.URL, p.UID, p.Key, p.Token, p.Type, p.PID, p.Status, p.ID,
		)
		if err != nil {
			return 0, fmt.Errorf("保存失败: %v", err)
		}
		return p.ID, nil
	}
	// 新增
	result, err := database.DB.Exec(
		"INSERT INTO xm_project (name, description, price, `query`, password, url, uid, `key`, token, type, p_id, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		p.Name, p.Description, p.Price, p.Query, p.Password, p.URL, p.UID, p.Key, p.Token, p.Type, p.PID, p.Status,
	)
	if err != nil {
		return 0, fmt.Errorf("添加失败: %v", err)
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

// AdminDeleteProject 管理员删除项目（软删除）
func (s *XMService) AdminDeleteProject(id int) error {
	_, err := database.DB.Exec("UPDATE xm_project SET is_deleted = 1 WHERE id = ?", id)
	return err
}
