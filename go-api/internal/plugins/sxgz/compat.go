package sxgz

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-api/internal/database"

	"github.com/gin-gonic/gin"
)

type sxgzCompatUser struct {
	UID      int
	Money    float64
	AddPrice float64
}

func RegisterCompatRoutes(r *gin.Engine) {
	r.Any("/apitaowa.php", CompatAPI)
	r.Any("/sxgz/api.php", CompatAPI)
}

func CompatAPI(c *gin.Context) {
	payload := readCompatPayload(c)
	action := strings.TrimSpace(compatString(payload, "action"))
	if action == "" {
		action = strings.TrimSpace(compatString(payload, "act"))
	}
	user, ok := compatAuth(c, payload)
	if !ok {
		return
	}

	switch action {
	case "get_companies_for_agent":
		compatGetCompanies(c, user)
	case "get_gonggao":
		compatGetGonggao(c, user, payload)
	case "create_order":
		compatCreateOrder(c, user, payload)
	case "sync_orders":
		compatSyncOrders(c, user, payload)
	case "get_order":
		compatGetOrder(c, user, payload)
	case "update_order_file":
		compatUpdateOrderFile(c, user, payload)
	case "apply_refund":
		compatApplyRefund(c, user, payload)
	default:
		compatError(c, http.StatusBadRequest, "无效的操作")
	}
}

func compatAuth(c *gin.Context, payload map[string]any) (sxgzCompatUser, bool) {
	uid := compatInt(payload, "uid")
	key := strings.TrimSpace(compatString(payload, "key"))
	if uid <= 0 || key == "" {
		compatError(c, 0, "uid或者key为空")
		return sxgzCompatUser{}, false
	}

	var user sxgzCompatUser
	var dbKey string
	err := database.DB.QueryRow(
		"SELECT uid, COALESCE(`key`,''), COALESCE(money,0), COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid = ?",
		uid,
	).Scan(&user.UID, &dbKey, &user.Money, &user.AddPrice)
	if err != nil {
		compatError(c, -1, "用户不存在")
		return sxgzCompatUser{}, false
	}
	if dbKey == "" || dbKey == "0" {
		compatError(c, -1, "你还没有开通接口哦")
		return sxgzCompatUser{}, false
	}
	if dbKey != key {
		compatError(c, -2, "密匙错误")
		return sxgzCompatUser{}, false
	}
	if user.AddPrice <= 0 {
		user.AddPrice = 1
	}
	return user, true
}

func compatGetCompanies(c *gin.Context, user sxgzCompatUser) {
	search := strings.TrimSpace(c.Query("search"))
	list, err := Sxgz().listCompanies(search, false)
	if err != nil {
		compatError(c, -1, "获取公司列表失败: "+err.Error())
		return
	}
	if len(list) == 0 {
		if _, refreshErr := Sxgz().RefreshCompanies(c.Request.Context()); refreshErr == nil {
			list, _ = Sxgz().listCompanies(search, false)
		}
	}
	compatSuccess(c, gin.H{
		"data":       list,
		"total":      len(list),
		"from_cache": true,
		"message":    "公司列表获取成功",
		"agent_uid":  user.UID,
	})
}

func compatGetGonggao(c *gin.Context, user sxgzCompatUser, payload map[string]any) {
	page := compatInt(payload, "page")
	pageSize := compatInt(payload, "pageSize")
	if pageSize <= 0 {
		pageSize = compatInt(payload, "limit")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	var parentID int
	_ = database.DB.QueryRow("SELECT COALESCE(uuid,0) FROM qingka_wangke_user WHERE uid = ?", user.UID).Scan(&parentID)

	where := "status = '1' AND (visibility = 0 OR (visibility = 1 AND uid = ?))"
	var total int
	_ = database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_gonggao WHERE "+where, parentID).Scan(&total)

	rows, err := database.DB.Query(
		"SELECT id, COALESCE(title,''), COALESCE(content,''), COALESCE(time,''), COALESCE(zhiding,'0') FROM qingka_wangke_gonggao WHERE "+where+" ORDER BY zhiding DESC, id DESC LIMIT ? OFFSET ?",
		parentID, pageSize, (page-1)*pageSize,
	)
	if err != nil {
		compatError(c, -1, "获取公告失败: "+err.Error())
		return
	}
	defer rows.Close()

	items := make([]SxgzAnnouncement, 0)
	for rows.Next() {
		var item SxgzAnnouncement
		var pinned string
		if err := rows.Scan(&item.AID, &item.Title, &item.Content, &item.PublishDate, &pinned); err != nil {
			continue
		}
		if pinned == "1" {
			item.Importance = 5
		} else {
			item.Importance = 1
		}
		items = append(items, item)
	}

	compatSuccess(c, gin.H{
		"data":     items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"hasMore":  page*pageSize < total,
		"type":     strings.TrimSpace(compatString(payload, "type")),
	})
}

func compatCreateOrder(c *gin.Context, user sxgzCompatUser, payload map[string]any) {
	req := compatOrderRequest(payload)
	if req.ServiceType == "" {
		req.ServiceType = "electronic"
	}
	if req.MaterialType == "" {
		req.MaterialType = "upload"
	}
	if req.CompanyID <= 0 || strings.TrimSpace(req.CustomerName) == "" {
		compatError(c, 400, "缺少必要字段: company_id 或 customer_name")
		return
	}

	result, err := Sxgz().CreateOrder(user.UID, req, requestBaseURL(c))
	if err != nil {
		compatError(c, 400, err.Error())
		return
	}
	_, _ = database.DB.Exec("UPDATE fd_sxgz_orders SET source = 'agent', agent_uid = ?, updated_at = NOW() WHERE order_id = ?", user.UID, result.OrderID)

	compatSuccess(c, gin.H{
		"message": "订单创建成功",
		"data": gin.H{
			"order_id":          result.OrderID,
			"order_no":          result.OrderNo,
			"upstream_order_id": result.UpstreamID,
			"total_price":       result.TotalPrice,
		},
	})
}

func compatSyncOrders(c *gin.Context, user sxgzCompatUser, payload map[string]any) {
	page := compatInt(payload, "page")
	limit := compatInt(payload, "limit")
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 500 {
		limit = 50
	}

	result, err := Sxgz().ListOrders(user.UID, false, page, limit, "", "")
	if err != nil {
		compatError(c, -1, "同步订单失败: "+err.Error())
		return
	}
	compatSuccess(c, gin.H{
		"data": result.List,
		"pagination": gin.H{
			"current_page": page,
			"total":        result.Total,
			"per_page":     limit,
		},
	})
}

func compatGetOrder(c *gin.Context, user sxgzCompatUser, payload map[string]any) {
	order, err := findCompatOrder(user.UID, int64(compatInt(payload, "order_id")), compatString(payload, "order_no"))
	if err != nil {
		compatError(c, 404, "订单不存在或无权限")
		return
	}
	compatSuccess(c, gin.H{"data": orderToMap(order)})
}

func compatUpdateOrderFile(c *gin.Context, user sxgzCompatUser, payload map[string]any) {
	result, err := updateCompatOrderFile(c, user.UID, payload)
	if err != nil {
		compatError(c, 400, err.Error())
		return
	}
	compatSuccess(c, result)
}

func updateCompatOrderFile(c *gin.Context, uid int, payload map[string]any) (gin.H, error) {
	fileURL := strings.TrimSpace(compatString(payload, "file_url"))
	if fileURL == "" {
		return nil, fmt.Errorf("缺少必要字段: file_url")
	}
	order, err := findCompatOrder(uid, int64(compatInt(payload, "order_id")), compatString(payload, "order_no"))
	if err != nil {
		return nil, fmt.Errorf("订单不存在或无权限访问")
	}
	files := normalizeCompatFileRecords(fileURL, compatString(payload, "plugin_domain"))
	rawFiles, _ := json.Marshal(files)
	originalFilename := strings.TrimSpace(compatString(payload, "original_filename"))
	if originalFilename == "" {
		originalFilename = joinFileNames(files)
	}
	fileSize := int64(compatInt(payload, "file_size"))
	if fileSize <= 0 {
		fileSize = totalFileSize(files)
	}

	fields := []string{"uploaded_file = ?", "original_filename = ?", "file_size = ?", "updated_at = NOW()"}
	args := []any{string(rawFiles), originalFilename, fileSize}
	if special := strings.TrimSpace(compatString(payload, "special_requirements")); special != "" {
		fields = append(fields, "special_requirements = ?")
		args = append(args, special)
	}
	args = append(args, order.OrderID, uid)

	_, err = database.DB.Exec("UPDATE fd_sxgz_orders SET "+strings.Join(fields, ", ")+" WHERE order_id = ? AND uid = ?", args...)
	if err != nil {
		return nil, fmt.Errorf("文件URL更新失败: %w", err)
	}

	if cfg, cfgErr := Sxgz().loadConfig(); cfgErr == nil && cfg.UpstreamEnabled() && order.AgentOrderID.Valid && order.AgentOrderID.Int64 > 0 {
		_, _ = Sxgz().callUpstreamUpdateFile(ctxWithTimeout(c.Request.Context(), 30*time.Second), cfg, order, files, requestBaseURL(c))
	}

	return gin.H{
		"message": "文件URL更新成功",
		"data": gin.H{
			"order_id":          order.OrderID,
			"order_no":          order.OrderNo,
			"file_url":          string(rawFiles),
			"original_filename": originalFilename,
			"updated_at":        time.Now().Format("2006-01-02 15:04:05"),
		},
	}, nil
}

func compatApplyRefund(c *gin.Context, user sxgzCompatUser, payload map[string]any) {
	reason := strings.TrimSpace(compatString(payload, "reason"))
	if reason == "" {
		compatError(c, 400, "退款原因不能为空")
		return
	}
	order, err := findCompatOrder(user.UID, int64(compatInt(payload, "order_id")), compatString(payload, "order_no"))
	if err != nil {
		compatError(c, 404, "订单不存在或无权限")
		return
	}
	if err := Sxgz().ApplyRefund(user.UID, order.OrderID, reason); err != nil {
		compatError(c, 400, err.Error())
		return
	}
	compatSuccess(c, gin.H{"message": "退款申请已提交，请等待管理员审核"})
}

func findCompatOrder(uid int, orderID int64, orderNo string) (*SxgzOrder, error) {
	if orderID > 0 {
		if order, err := Sxgz().GetOrder(uid, orderID, false); err == nil {
			return order, nil
		}
	}
	orderNo = strings.TrimSpace(orderNo)
	if orderNo == "" {
		return nil, sql.ErrNoRows
	}
	var localID int64
	err := database.DB.QueryRow("SELECT order_id FROM fd_sxgz_orders WHERE order_no = ? AND uid = ? LIMIT 1", orderNo, uid).Scan(&localID)
	if err != nil {
		return nil, err
	}
	return Sxgz().GetOrder(uid, localID, false)
}

func readCompatPayload(c *gin.Context) map[string]any {
	payload := map[string]any{}
	contentType := strings.ToLower(c.GetHeader("Content-Type"))
	if strings.Contains(contentType, "application/json") {
		decoder := json.NewDecoder(c.Request.Body)
		decoder.UseNumber()
		_ = decoder.Decode(&payload)
	} else {
		_ = c.Request.ParseForm()
		for key, values := range c.Request.PostForm {
			payload[key] = firstFormValue(values)
		}
	}
	for key, values := range c.Request.URL.Query() {
		payload[key] = firstFormValue(values)
	}
	return payload
}

func compatOrderRequest(payload map[string]any) OrderQuoteRequest {
	return OrderQuoteRequest{
		ServiceType:              compatString(payload, "service_type"),
		CompanyID:                compatInt(payload, "company_id"),
		CustomCompanyName:        compatString(payload, "custom_company_name"),
		CustomerName:             compatString(payload, "customer_name"),
		CustomerEmail:            compatString(payload, "customer_email"),
		CustomerPhone:            compatString(payload, "customer_phone"),
		CustomerAddress:          compatString(payload, "customer_address"),
		CourierCompany:           compatString(payload, "courier_company"),
		TrackingNumber:           compatString(payload, "tracking_number"),
		ReturnTrackingNumber:     compatString(payload, "return_tracking_number"),
		PrintCopies:              compatInt(payload, "print_copies"),
		PrintOptions:             compatStringSlice(payload["print_options"]),
		FilePrintOptions:         compatFilePrintOptions(payload["file_print_options"]),
		PaperSize:                compatString(payload, "paper_size"),
		SpecialRequirements:      compatString(payload, "special_requirements"),
		BusinessLicense:          compatBool(payload, "business_license"),
		OnlyBusinessLicense:      compatBool(payload, "only_business_license"),
		MaterialType:             compatString(payload, "material_type"),
		DeliveryOption:           compatString(payload, "delivery_option"),
		SelectedLicenseCompanies: compatIntSlice(payload["selected_license_companies"]),
	}
}

func normalizeCompatFileRecords(raw string, pluginDomain string) []SxgzFileRecord {
	records := parseFileRecords(raw)
	base := strings.TrimRight(strings.TrimSpace(pluginDomain), "/")
	if base != "" && !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "http://" + base
	}
	for i := range records {
		url := strings.TrimSpace(records[i].URL)
		if base != "" && strings.HasPrefix(url, "/") {
			url = base + url
		}
		records[i].URL = url
		if records[i].Storage == "" {
			records[i].Storage = "remote"
		}
	}
	return records
}

func compatSuccess(c *gin.Context, body gin.H) {
	body["success"] = true
	body["code"] = 1
	if _, ok := body["message"]; !ok {
		body["message"] = "success"
	}
	c.JSON(http.StatusOK, body)
}

func compatError(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"code":    code,
		"message": message,
		"msg":     message,
	})
}

func compatString(payload map[string]any, key string) string {
	return asString(payload[key])
}

func compatInt(payload map[string]any, key string) int {
	return asInt(payload[key])
}

func compatBool(payload map[string]any, key string) bool {
	return asBool(payload[key])
}

func compatStringSlice(v any) []string {
	switch val := v.(type) {
	case []string:
		return val
	case []any:
		out := make([]string, 0, len(val))
		for _, item := range val {
			if s := strings.TrimSpace(asString(item)); s != "" {
				out = append(out, s)
			}
		}
		return out
	case string:
		val = strings.TrimSpace(val)
		if val == "" {
			return nil
		}
		var out []string
		if err := json.Unmarshal([]byte(val), &out); err == nil {
			return out
		}
		return splitCompatList(val)
	}
	return nil
}

func compatIntSlice(v any) []int {
	switch val := v.(type) {
	case []int:
		return val
	case []any:
		out := make([]int, 0, len(val))
		for _, item := range val {
			if n := asInt(item); n > 0 {
				out = append(out, n)
			}
		}
		return out
	case string:
		val = strings.TrimSpace(val)
		if val == "" {
			return nil
		}
		var out []int
		if err := json.Unmarshal([]byte(val), &out); err == nil {
			return out
		}
		parts := splitCompatList(val)
		out = make([]int, 0, len(parts))
		for _, part := range parts {
			n, _ := strconv.Atoi(part)
			if n > 0 {
				out = append(out, n)
			}
		}
		return out
	}
	return nil
}

func compatFilePrintOptions(v any) []SxgzFilePrintRequest {
	switch val := v.(type) {
	case []SxgzFilePrintRequest:
		return val
	case []any:
		data, _ := json.Marshal(val)
		var out []SxgzFilePrintRequest
		_ = json.Unmarshal(data, &out)
		return out
	case string:
		val = strings.TrimSpace(val)
		if val == "" {
			return nil
		}
		var out []SxgzFilePrintRequest
		if err := json.Unmarshal([]byte(val), &out); err == nil {
			return out
		}
	}
	return nil
}

func splitCompatList(raw string) []string {
	raw = strings.Trim(raw, "[] ")
	if raw == "" {
		return nil
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '，'
	})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.Trim(strings.TrimSpace(part), `"'`)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func firstFormValue(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
