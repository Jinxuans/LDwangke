package xm

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strings"
	"time"

	"go-api/internal/database"
)

func (s *XMService) GetProjects(uid int) ([]XMProject, error) {
	var addprice float64
	err := database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	rows, err := database.DB.Query("SELECT id, name, description, COALESCE(NULLIF(local_price, 0), price, upstream_price, 0) AS price, `query`, password FROM xm_project WHERE is_deleted = 0 AND status = 0 ORDER BY sort_order ASC, id ASC")
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

	var pace *float64
	if v, ok := data["pace"].(float64); ok {
		pace = &v
	}
	var distance *float64
	if v, ok := data["distance"].(float64); ok {
		distance = &v
	}

	if projectID == 0 || school == "" || account == "" || totalKM == 0 || startDay == "" || startTime == "" || endTime == "" {
		return nil, fmt.Errorf("缺少必填参数")
	}
	if runDateArray == nil {
		return nil, fmt.Errorf("缺少必填参数")
	}

	project, err := s.getProjectRow(projectID)
	if err != nil {
		return nil, fmt.Errorf("项目不存在")
	}

	projectPrice := getXMProjectBasePrice(project)

	danjia := math.Round(projectPrice*addprice*100) / 100
	if danjia <= 0 || addprice < 0.1 {
		return nil, fmt.Errorf("单价异常，请联系管理员")
	}
	orderMoney := math.Round(float64(totalKM)*danjia*100) / 100

	if money < orderMoney {
		return nil, fmt.Errorf("余额不足")
	}

	runDateJSON, _ := json.Marshal(runDateArray)
	var typeSQL interface{}
	if orderType != nil {
		typeSQL = *orderType
	}

	result, err := database.DB.Exec(
		`INSERT INTO xm_order (y_oid, user_id, school, account, password, type, pace, distance, project_id, status, total_km, run_km, run_date, start_day, start_time, end_time, deduction, is_deleted, created_at, updated_at)
		VALUES (NULL, ?, ?, ?, ?, ?, ?, ?, ?, '已下单', ?, NULL, ?, ?, ?, ?, ?, 0, NOW(), NOW())`,
		uid, school, account, password, typeSQL, pace, distance, projectID, totalKM, string(runDateJSON), startDay, startTime, endTime, orderMoney,
	)
	if err != nil {
		return nil, fmt.Errorf("下单失败请联系管理员")
	}
	orderID, _ := result.LastInsertId()

	database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ? LIMIT 1", orderMoney, uid)

	projectName, _ := project["name"].(string)
	xmLog(uid, "添加任务", fmt.Sprintf("项目：%s %s %s 扣除 %.2f 元", projectName, account, password, orderMoney), -orderMoney)

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
	if pace != nil {
		postData["pace"] = *pace
	}
	if distance != nil {
		postData["distance"] = *distance
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
		"pace":        pace,
		"distance":    distance,
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

func (s *XMService) AddOrderKM(uid int, orderID int, addKM int, isAdmin bool) (map[string]interface{}, error) {
	if orderID <= 0 || addKM <= 0 {
		return nil, fmt.Errorf("参数错误")
	}

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
		return nil, fmt.Errorf("该订单已删除，无法增加次数")
	}

	projectID := 0
	fmt.Sscanf(fmt.Sprintf("%v", order["project_id"]), "%d", &projectID)
	project, err := s.getProjectRow(projectID)
	if err != nil {
		return nil, fmt.Errorf("项目不存在")
	}

	orderUserID := 0
	fmt.Sscanf(fmt.Sprintf("%v", order["user_id"]), "%d", &orderUserID)

	var addprice, userMoney float64
	err = database.DB.QueryRow("SELECT addprice, money FROM qingka_wangke_user WHERE uid = ?", orderUserID).Scan(&addprice, &userMoney)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
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
	money := math.Round(float64(addKM)*danjia*100) / 100

	if userMoney < money {
		return nil, fmt.Errorf("用户余额不足")
	}

	yOidStr := fmt.Sprintf("%v", order["y_oid"])
	yOid := 0
	fmt.Sscanf(yOidStr, "%d", &yOid)

	if yOid > 0 {
		conn, err := s.resolveProjectConnection(project)
		if err != nil {
			return nil, err
		}

		var externalResult map[string]interface{}
		if conn.AuthType == 0 {
			params := url.Values{}
			params.Set("act", "add_order_km")
			params.Set("key", conn.Key)
			params.Set("uid", conn.UID)
			queryURL := strings.TrimSpace(conn.BaseURL) + "?" + params.Encode()
			postData := map[string]interface{}{
				"order_id": yOid,
				"add_km":   addKM,
			}
			externalResult, err = s.httpRequest("POST", queryURL, postData, map[string]string{"Content-Type": "application/json"})
		} else {
			domain := extractDomain(strings.TrimSpace(conn.BaseURL))
			params := url.Values{}
			params.Set("order_id", fmt.Sprintf("%d", yOid))
			params.Set("add_km", fmt.Sprintf("%d", addKM))
			queryURL := domain + "/api/v1/runorder/add_total_km?" + params.Encode()
			externalResult, err = s.httpRequest("GET", queryURL, nil, buildXMTokenHeaders(conn.Token, map[string]string{"Accept": "application/json"}))
		}

		if err != nil || externalResult == nil {
			return nil, fmt.Errorf("外部接口返回格式错误")
		}
		code, _ := externalResult["code"].(float64)
		if int(code) != 200 {
			msg, _ := externalResult["msg"].(string)
			if msg == "" {
				msg = "外部接口增加次数失败"
			}
			return nil, fmt.Errorf("%s", msg)
		}
	}

	database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ? LIMIT 1", money, orderUserID)

	oldTotalKM := 0
	fmt.Sscanf(fmt.Sprintf("%v", order["total_km"]), "%d", &oldTotalKM)
	oldDeduction := 0.0
	fmt.Sscanf(fmt.Sprintf("%v", order["deduction"]), "%f", &oldDeduction)

	newTotalKM := oldTotalKM + addKM
	newDeduction := oldDeduction + money

	database.DB.Exec("UPDATE xm_order SET total_km = ?, deduction = ?, updated_at = NOW() WHERE id = ? LIMIT 1",
		newTotalKM, newDeduction, orderID)

	xmLog(orderUserID, "增加次数", fmt.Sprintf("订单 %d 增加 %d 次，扣除 %.2f 元", orderID, addKM, money), -money)

	var latestBalance float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", orderUserID).Scan(&latestBalance)

	return map[string]interface{}{
		"order_id":       orderID,
		"add_km":         addKM,
		"add_deduction":  fmt.Sprintf("%.4f", money),
		"latest_balance": latestBalance,
		"total_km":       newTotalKM,
		"deduction":      fmt.Sprintf("%.4f", newDeduction),
	}, nil
}

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

	querySQL := fmt.Sprintf("SELECT id, user_id, school, account, password, project_id, status, COALESCE(type,-1), pace, distance, total_km, is_deleted, run_km, run_date, start_day, start_time, end_time, deduction, updated_at FROM xm_order WHERE %s ORDER BY id DESC LIMIT ?, ?", where)
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
			&o.ProjectID, &o.Status, &typeInt, &o.Pace, &o.Distance, &o.TotalKM, &isDeletedInt,
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
	if school, ok := data["school"].(string); ok && strings.TrimSpace(school) != "" {
		postData["school"] = strings.TrimSpace(school)
	}

	result, err := s.projectRequest(project, "query_run", postData, "POST")
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *XMService) RefundOrder(uid int, orderID int, isAdmin bool) (map[string]interface{}, error) {
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

	database.DB.Exec("UPDATE xm_order SET status = '待退款', updated_at = NOW() WHERE id = ? LIMIT 1", orderID)

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
		totalKMStr := fmt.Sprintf("%v", order["total_km"])
		fmt.Sscanf(totalKMStr, "%f", &refundKM)

		var addprice float64
		database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", orderUserID).Scan(&addprice)
		projectPrice := getXMProjectBasePrice(project)
		danjia = math.Round(projectPrice*addprice*100) / 100
		if danjia <= 0 || addprice < 0.1 {
			return nil, fmt.Errorf("单价异常，请联系管理员")
		}
	} else {
		externalOID := 0
		fmt.Sscanf(yOidStr, "%d", &externalOID)
		if externalOID <= 0 {
			return nil, fmt.Errorf("该订单未提交到外部系统，无法退款")
		}

		var externalResult map[string]interface{}

		conn, err := s.resolveProjectConnection(project)
		if err != nil {
			database.DB.Exec("UPDATE xm_order SET status = '退款失败', updated_at = NOW() WHERE id = ? LIMIT 1", orderID)
			return nil, err
		}

		if conn.AuthType == 0 {
			params := url.Values{}
			params.Set("act", "refund_order")
			params.Set("key", conn.Key)
			params.Set("uid", conn.UID)
			params.Set("order_id", fmt.Sprintf("%d", externalOID))
			queryURL := strings.TrimSpace(conn.BaseURL) + "?" + params.Encode()
			externalResult, err = s.httpRequest("GET", queryURL, nil, nil)
		} else {
			params := url.Values{}
			params.Set("order_id", fmt.Sprintf("%d", externalOID))
			queryURL := strings.TrimRight(strings.TrimSpace(conn.BaseURL), "/") + "/refund?" + params.Encode()
			externalResult, err = s.httpRequest("GET", queryURL, nil, buildXMTokenHeaders(conn.Token, nil))
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
		projectPrice := getXMProjectBasePrice(project)
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

	conn, err := s.resolveProjectConnection(project)
	if err != nil {
		return "", err
	}

	var externalResult map[string]interface{}
	if conn.AuthType == 0 {
		params := url.Values{}
		params.Set("act", "delete_order")
		params.Set("key", conn.Key)
		params.Set("uid", conn.UID)
		params.Set("order_id", fmt.Sprintf("%d", externalOID))
		queryURL := strings.TrimSpace(conn.BaseURL) + "?" + params.Encode()
		externalResult, err = s.httpRequest("GET", queryURL, nil, nil)
	} else {
		pURL := strings.TrimRight(strings.TrimSpace(conn.BaseURL), "/")
		externalResult, err = s.httpRequest("DELETE", pURL+"/delete", map[string]interface{}{"order_id": externalOID}, buildXMTokenHeaders(conn.Token, nil))
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

// SyncOrder 同步订单（从外部系统获取最新订单状态并更新到本地）
// uid: 当前用户ID
// orderID: 订单ID
// isAdmin: 是否为管理员（管理员可以同步任意订单，普通用户只能同步自己的订单）
// 返回值：同步结果消息，错误信息
func (s *XMService) SyncOrder(uid int, orderID int, isAdmin bool) (string, error) {
	// ==================== 第一步：查询订单信息 ====================

	// 根据是否为管理员，构建不同的 SQL 查询语句
	var query string
	if isAdmin {
		// 管理员：只需根据订单ID查询
		query = "SELECT * FROM xm_order WHERE id = ? LIMIT 1"
	} else {
		// 普通用户：需要同时验证订单ID和所属用户ID，防止越权访问
		query = "SELECT * FROM xm_order WHERE id = ? AND user_id = ? LIMIT 1"
	}

	// 执行数据库查询
	rows, err := database.DB.Query(query, func() []interface{} {
		if isAdmin {
			return []interface{}{orderID} // 管理员只需传入 orderID
		}
		return []interface{}{orderID, uid} // 普通用户需要传入 orderID 和 uid
	}()...) // ... 是 Go 的展开运算符，将切片展开为多个参数
	if err != nil {
		return "", fmt.Errorf("查询失败")
	}
	defer rows.Close() // 延迟关闭 rows，防止资源泄漏

	// 获取查询结果的列名（字段名），比如 ["id", "user_id", "y_oid", ...]
	columns, _ := rows.Columns()

	// 检查是否有查询结果
	if !rows.Next() {
		return "", fmt.Errorf("订单不存在")
	}

	// ==================== 第二步：将查询结果扫描到 map 中 ====================
	// Go 的 database/sql 包不能直接扫描到 map，需要手动处理

	// values 用于存储每一列的实际值
	values := make([]interface{}, len(columns))
	// valuePtrs 存储指向 values 每个元素的指针，Scan 方法需要指针
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i] // 让每个指针指向对应的 value 元素
	}

	// 将当前行的数据扫描到 valuePtrs 中（实际是填充 values 数组）
	rows.Scan(valuePtrs...)

	// 将扫描结果转换为 map[string]interface{}，方便后续使用
	order := make(map[string]interface{})
	for i, col := range columns {
		val := values[i]
		// Go 从数据库读取的字符串是 []byte 类型，需要转换为 string
		if b, ok := val.([]byte); ok {
			order[col] = string(b)
		} else {
			order[col] = val
		}
	}
	rows.Close() // 提前关闭 rows（其实前面有 defer，这里是显式关闭）

	// ==================== 第三步：检查订单是否已删除 ====================

	isDeleted := fmt.Sprintf("%v", order["is_deleted"])
	if isDeleted == "1" {
		return "", fmt.Errorf("该订单已删除，无法同步")
	}

	// ==================== 第四步：验证订单是否有外部订单 ID ====================

	// y_oid 是外部系统的订单ID
	yOidStr := fmt.Sprintf("%v", order["y_oid"])
	yOid := 0
	fmt.Sscanf(yOidStr, "%d", &yOid) // 将字符串解析为整数

	if yOid <= 0 {
		return "", fmt.Errorf("该订单未提交到外部系统，无法同步")
	}

	// ==================== 第五步：获取项目信息 ====================

	// 从订单中获取 project_id（项目ID）
	projectID := 0
	fmt.Sscanf(fmt.Sprintf("%v", order["project_id"]), "%d", &projectID)

	// 查询项目详情
	project, err := s.getProjectRow(projectID)
	if err != nil {
		return "", fmt.Errorf("项目不存在")
	}

	// ==================== 第六步：调用上游接口同步订单 ====================

	// syncOrderFromUpstream 会向外部系统请求订单最新状态，并更新到本地数据库
	externalResult, err := s.syncOrderFromUpstream(yOid, project)
	if err != nil {
		return "", err
	}

	// ==================== 第七步：解析同步结果 ====================

	// 检查外部系统返回的 code 是否为 200
	code, _ := externalResult["code"].(float64) // JSON 数字在 Go 中默认是 float64
	if int(code) != 200 {
		msg, _ := externalResult["msg"].(string)
		if msg == "" {
			msg = "外部同步失败"
		}
		return "", fmt.Errorf("%s", msg)
	}

	// 同步成功，返回成功消息
	msg, _ := externalResult["msg"].(string)
	if msg == "" {
		msg = "同步成功"
	}
	return msg, nil
}

// syncOrderFromUpstream 从上游系统同步订单数据到本地数据库
// yOid: 外部系统的订单ID
// project: 项目信息（包含连接配置）
// 返回值：同步结果（包含 code、msg、data），错误信息
func (s *XMService) syncOrderFromUpstream(yOid int, project map[string]interface{}) (map[string]interface{}, error) {
	// ==================== 第一步：解析项目连接配置 ====================

	// resolveProjectConnection 从 project 中提取连接信息（BaseURL、Key、UID、Token、AuthType 等）
	conn, err := s.resolveProjectConnection(project)
	if err != nil {
		return nil, err
	}

	var externalResult map[string]interface{}

	// ==================== 第二步：根据认证类型构建请求 ====================

	if conn.AuthType == 0 {
		// ========== AuthType 0：普通项目，请求项目方配置的 URL ==========

		params := url.Values{}
		params.Set("act", "get_orders")                 // 动作：获取订单
		params.Set("key", conn.Key)                     // 项目密钥
		params.Set("uid", conn.UID)                     // 项目用户ID
		params.Set("order_id", fmt.Sprintf("%d", yOid)) // 外部订单ID

		// 构建完整 URL
		queryURL := strings.TrimSpace(conn.BaseURL) + "?" + params.Encode()

		// 发送 GET 请求
		externalResult, err = s.httpRequest("GET", queryURL, nil, nil)
	} else {
		// ========== AuthType 其他值：66-dd 项目，请求统一平台 API ==========

		// 处理基础 URL：去除两端空格，再去除末尾的 /
		pURL := strings.TrimRight(strings.TrimSpace(conn.BaseURL), "/")

		params := url.Values{}
		params.Set("id", fmt.Sprintf("%d", yOid)) // 外部订单ID
		params.Set("page", "1")                   // 页码
		params.Set("page_size", "10")             // 每页数量

		// 构建完整 URL：xxx/list?id=xxx&page=1&page_size=10
		queryURL := pURL + "/list?" + params.Encode()

		// 发送 GET 请求，携带 token 请求头
		externalResult, err = s.httpRequest("GET", queryURL, nil, buildXMTokenHeaders(conn.Token, nil))
	}

	// ==================== 第三步：检查 HTTP 请求是否成功 ====================

	if err != nil {
		return nil, err
	}

	// ==================== 第四步：检查业务状态码 ====================

	code, _ := externalResult["code"].(float64) // JSON 数字在 Go 中默认是 float64
	if int(code) != 200 {
		// 业务失败，直接返回外部系统的响应（不更新数据库）
		return externalResult, nil
	}

	// ==================== 第五步：解析返回的数据 ====================

	// 从响应中提取 data 字段（订单数据列表）
	dataList, _ := externalResult["data"].([]interface{})
	if dataList == nil {
		// data 为空或类型不对，返回错误
		return map[string]interface{}{"code": float64(-1), "msg": "外部接口返回的数据格式错误"}, nil
	}

	// ==================== 第六步：定义不需要更新的字段 ====================

	// 这些字段是本地核心字段，不允许被外部数据覆盖
	skipFields := map[string]bool{
		"id":         true, // 本地订单ID
		"y_oid":      true, // 外部订单ID，不允许覆盖
		"user_id":    true, // 用户ID
		"school":     true, // 学校信息
		"account":    true, // 账号
		"password":   true, // 密码
		"project_id": true, // 项目ID
		"type":       true, // 类型
		"deduction":  true, // 扣分/扣减记录
	}

	// ==================== 第七步：遍历数据列表，更新本地数据库 ====================

	for _, item := range dataList {
		// 将 item 断言为 map[string]interface{}
		row, ok := item.(map[string]interface{})
		if !ok {
			// 类型不对，跳过这条数据
			continue
		}

		// updateParts: 存储 SQL 更新片段，如 ["`status` = ?", "`run_km` = ?"]
		var updateParts []string
		// updateArgs: 存储 SQL 参数值
		var updateArgs []interface{}

		// 遍历外部返回的每个字段
		for field, value := range row {
			// 跳过核心字段，不允许更新
			if skipFields[field] {
				continue
			}

			// 针对不同字段做特殊处理
			if field == "status_name" {
				// status_name 映射到本地的 status 字段
				updateParts = append(updateParts, "`status` = ?")
				updateArgs = append(updateArgs, value)

			} else if field == "run_date" {
				// run_date 是日期/时间字段，转为 JSON 字符串存储
				jsonVal, _ := json.Marshal(value)
				updateParts = append(updateParts, "`run_date` = ?")
				updateArgs = append(updateArgs, string(jsonVal))

			} else if field == "is_deleted" {
				// is_deleted 可能是 bool 类型，需要转为 int (0/1)
				boolVal := 0
				if v, ok := value.(bool); ok && v {
					boolVal = 1
				}
				updateParts = append(updateParts, "`is_deleted` = ?")
				updateArgs = append(updateArgs, boolVal)

			} else if field == "run_km" {
				// run_km 可能是 NULL
				if value == nil {
					updateParts = append(updateParts, "`run_km` = NULL")
				} else {
					updateParts = append(updateParts, "`run_km` = ?")
					updateArgs = append(updateArgs, value)
				}

			} else {
				// 其他字段直接添加
				updateParts = append(updateParts, fmt.Sprintf("`%s` = ?", field))
				updateArgs = append(updateArgs, value)
			}
		}

		// ==================== 第八步：执行 SQL 更新 ====================

		if len(updateParts) > 0 {
			// 处理 updated_at 字段
			if updatedAt, ok := row["updated_at"].(string); ok && updatedAt != "" {
				// 外部系统有 updated_at，使用外部的时间
				updateParts = append(updateParts, "`updated_at` = ?")
				updateArgs = append(updateArgs, updatedAt)
			} else {
				// 外部系统没有 updated_at，使用当前时间
				updateParts = append(updateParts, "`updated_at` = NOW()")
			}

			// 添加 WHERE 条件参数
			updateArgs = append(updateArgs, yOid)

			// 构建完整 SQL：UPDATE xm_order SET `status` = ?, `run_km` = ?, `updated_at` = ? WHERE y_oid = ? LIMIT 1
			sql := "UPDATE xm_order SET " + strings.Join(updateParts, ", ") + " WHERE y_oid = ? LIMIT 1"

			// 执行更新（不检查错误，静默失败）
			database.DB.Exec(sql, updateArgs...)
		}
	}

	// ==================== 第九步：返回成功结果 ====================

	return map[string]interface{}{
		"code": float64(200),
		"msg":  "同步成功",
		"data": dataList, // 返回同步的数据列表
	}, nil
}

// GetOrderLogs 获取订单日志
// uid: 当前用户ID
// orderID: 订单ID
// isAdmin: 是否为管理员（管理员可以查看任意订单，普通用户只能查看自己的订单）
// page: 页码
// pageSize: 每页数量
func (s *XMService) GetOrderLogs(uid int, orderID int, isAdmin bool, page, pageSize int) (interface{}, error) {
	// ==================== 第一步：查询订单信息 ====================

	// 根据是否为管理员，构建不同的 SQL 查询语句
	var query string
	if isAdmin {
		// 管理员：只需根据订单ID查询
		query = "SELECT * FROM xm_order WHERE id = ? LIMIT 1"
	} else {
		// 普通用户：需要同时验证订单ID和所属用户ID，防止越权访问
		query = "SELECT * FROM xm_order WHERE id = ? AND user_id = ? LIMIT 1"
	}

	// 执行数据库查询
	// func() []interface{} { ... }() 是一个立即执行的匿名函数，用于动态构建查询参数
	rows, err := database.DB.Query(query, func() []interface{} {
		if isAdmin {
			return []interface{}{orderID} // 管理员只需传入 orderID
		}
		return []interface{}{orderID, uid} // 普通用户需要传入 orderID 和 uid
	}()...) // ... 是 Go 的展开运算符，将切片展开为多个参数
	if err != nil {
		return nil, fmt.Errorf("查询失败")
	}
	defer rows.Close() // 延迟关闭 rows，防止资源泄漏

	// 获取查询结果的列名（字段名），比如 ["id", "user_id", "y_oid", ...]
	columns, _ := rows.Columns()

	// 检查是否有查询结果
	if !rows.Next() {
		return nil, fmt.Errorf("订单不存在或无权限查看")
	}

	// ==================== 第二步：将查询结果扫描到 map 中 ====================
	// Go 的 database/sql 包不能直接扫描到 map，需要手动处理

	// values 用于存储每一列的实际值
	values := make([]interface{}, len(columns))
	// valuePtrs 存储指向 values 每个元素的指针，Scan 方法需要指针
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i] // 让每个指针指向对应的 value 元素
	}

	// 将当前行的数据扫描到 valuePtrs 中（实际是填充 values 数组）
	rows.Scan(valuePtrs...)

	// 将扫描结果转换为 map[string]interface{}，方便后续使用
	order := make(map[string]interface{})
	for i, col := range columns {
		val := values[i]
		// Go 从数据库读取的字符串是 []byte 类型，需要转换为 string
		if b, ok := val.([]byte); ok {
			order[col] = string(b)
		} else {
			order[col] = val
		}
	}
	rows.Close() // 提前关闭 rows（其实前面有 defer，这里是显式关闭）

	// ==================== 第三步：验证订单是否有外部订单 ID ====================

	// y_oid 是外部系统的订单ID，如果为空说明订单还没提交到外部系统
	yOidStr := fmt.Sprintf("%v", order["y_oid"]) // 将 y_oid 转为字符串
	if yOidStr == "<nil>" || yOidStr == "" || yOidStr == "0" {
		return nil, fmt.Errorf("该订单未提交到外部系统，无法查询日志")
	}

	// ==================== 第四步：获取项目信息并解析连接配置 ====================

	// 从订单中获取 project_id（项目ID）
	projectID := 0
	fmt.Sscanf(fmt.Sprintf("%v", order["project_id"]), "%d", &projectID)

	// 查询项目详情
	project, err := s.getProjectRow(projectID)
	if err != nil {
		return nil, fmt.Errorf("项目不存在")
	}

	// 【与之前版本的不同点】调用 resolveProjectConnection 解析项目连接配置
	// conn 包含：BaseURL, Key, UID, Token, AuthType 等信息
	conn, err := s.resolveProjectConnection(project)
	if err != nil {
		return nil, err
	}

	// ==================== 第五步：构建外部 API 请求 ====================

	headers := map[string]string{} // HTTP 请求头
	params := url.Values{}         // URL 查询参数（类似 PHP 的 http_build_query）
	params.Set("order_id", yOidStr)
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("page_size", fmt.Sprintf("%d", pageSize))

	var queryURL string
	if conn.AuthType == 0 {
		// ========== AuthType 0：普通项目，请求项目方配置的 URL ==========
		params.Set("act", "get_order_logs") // 动作：获取订单日志
		params.Set("key", conn.Key)         // 项目密钥
		params.Set("uid", conn.UID)         // 项目用户ID

		// 拼接完整 URL，类似：https://example.com/api?order_id=xxx&act=get_order_logs&...
		// strings.TrimSpace 去除 URL 两端可能的空白字符
		queryURL = strings.TrimSpace(conn.BaseURL) + "?" + params.Encode()
	} else {
		// ========== AuthType 其他值：66-dd 项目，请求统一平台 API ==========
		baseURL := "https://66-dd.com/api/v1/runorderlog/log"

		// 【与之前版本的不同点】使用 buildXMTokenHeaders 构建请求头
		// 这个函数会在 headers 中添加 token 字段
		headers = buildXMTokenHeaders(conn.Token, map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json, text/plain, */*",
		})
		queryURL = baseURL + "?" + params.Encode()
	}

	// ==================== 第六步：发送 HTTP 请求到外部系统 ====================

	externalResult, err := s.httpRequest("GET", queryURL, nil, headers)
	if err != nil {
		return nil, fmt.Errorf("外部接口响应格式异常")
	}

	// ==================== 第七步：解析外部系统响应 ====================

	// 检查外部系统返回的 code 是否为 200
	code, _ := externalResult["code"].(float64) // JSON 数字在 Go 中默认是 float64
	if int(code) != 200 {
		msg, _ := externalResult["msg"].(string)
		if msg == "" {
			msg = "获取日志失败"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	// 返回外部系统的完整响应数据
	return externalResult, nil
}
