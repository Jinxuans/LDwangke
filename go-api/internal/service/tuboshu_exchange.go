package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"go-api/internal/database"
)

func (s *TuboshuService) handlePointsExchange(uid int, isAdmin bool, method, path string, params map[string]interface{}) (interface{}, bool, error) {
	cleanPath := strings.TrimLeft(path, "/")

	if method == "GET" && cleanPath == "points-exchange/products" {
		data, err := s.getAvailableProducts()
		return map[string]interface{}{"success": true, "data": data}, false, err
	}
	if method == "POST" && cleanPath == "points-exchange/exchange" {
		productID := getIntFromInterface(params["productId"])
		data, err := s.exchangeProduct(uid, productID)
		return map[string]interface{}{"success": true, "data": data}, false, err
	}
	if method == "GET" && cleanPath == "points-exchange/records" {
		page := getIntParam(params, "page", 1)
		pageSize := getIntParam(params, "pageSize", 10)
		data, err := s.getExchangeRecords(uid, page, pageSize)
		return map[string]interface{}{"success": true, "data": data}, false, err
	}

	if !isAdmin {
		return nil, false, fmt.Errorf("需要管理员权限")
	}

	if method == "GET" && cleanPath == "admin/points-exchange/products" {
		page := getIntParam(params, "page", 1)
		pageSize := getIntParam(params, "pageSize", 10)
		data, err := s.getAdminProducts(page, pageSize)
		return map[string]interface{}{"success": true, "data": data}, false, err
	}
	if method == "POST" && cleanPath == "admin/points-exchange/products" {
		err := s.saveProduct(params)
		return map[string]interface{}{"success": true}, false, err
	}
	if method == "DELETE" && strings.HasPrefix(cleanPath, "admin/points-exchange/products/") {
		re := regexp.MustCompile(`products/(\d+)$`)
		m := re.FindStringSubmatch(cleanPath)
		if len(m) > 1 {
			id, _ := strconv.Atoi(m[1])
			err := s.deleteProduct(id)
			return map[string]interface{}{"success": true}, false, err
		}
	}
	if method == "GET" && strings.Contains(cleanPath, "/codes") {
		re := regexp.MustCompile(`products/(\d+)/codes`)
		m := re.FindStringSubmatch(cleanPath)
		if len(m) > 1 {
			productID, _ := strconv.Atoi(m[1])
			page := getIntParam(params, "page", 1)
			pageSize := getIntParam(params, "pageSize", 20)
			data, err := s.getProductCodes(productID, page, pageSize)
			return map[string]interface{}{"success": true, "data": data}, false, err
		}
	}
	if method == "POST" && cleanPath == "admin/points-exchange/codes" {
		productID := getIntFromInterface(params["productId"])
		codes, _ := params["codes"].([]interface{})
		err := s.addCodes(productID, codes)
		return map[string]interface{}{"success": true}, false, err
	}
	if method == "DELETE" && strings.HasPrefix(cleanPath, "admin/points-exchange/codes/") {
		re := regexp.MustCompile(`codes/(\d+)$`)
		m := re.FindStringSubmatch(cleanPath)
		if len(m) > 1 {
			id, _ := strconv.Atoi(m[1])
			err := s.deleteCode(id)
			return map[string]interface{}{"success": true}, false, err
		}
	}
	if method == "GET" && cleanPath == "admin/points-exchange/records" {
		page := getIntParam(params, "page", 1)
		pageSize := getIntParam(params, "pageSize", 20)
		data, err := s.getAdminExchangeRecords(page, pageSize)
		return map[string]interface{}{"success": true, "data": data}, false, err
	}

	return nil, false, fmt.Errorf("未知的点数兑换路由")
}

func (s *TuboshuService) getAvailableProducts() ([]map[string]interface{}, error) {
	rows, err := database.DB.Query(`SELECT id, name, description, image_url, price, sort_order
		FROM points_product WHERE status = 'ENABLED'
		AND id IN (SELECT DISTINCT product_id FROM points_product_code WHERE status = 'AVAILABLE')
		ORDER BY sort_order DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var id, sortOrder int
		var name, desc, imageURL string
		var price float64
		rows.Scan(&id, &name, &desc, &imageURL, &price, &sortOrder)

		var stock int
		database.DB.QueryRow("SELECT COUNT(*) FROM points_product_code WHERE product_id = ? AND status = 'AVAILABLE'", id).Scan(&stock)

		products = append(products, map[string]interface{}{
			"id": id, "name": name, "description": desc,
			"image_url": imageURL, "price": price, "stock": stock,
		})
	}
	if products == nil {
		products = []map[string]interface{}{}
	}
	return products, nil
}

func (s *TuboshuService) exchangeProduct(uid, productID int) (map[string]interface{}, error) {
	var name string
	var price float64
	var status string
	err := database.DB.QueryRow("SELECT name, price, status FROM points_product WHERE id = ?", productID).Scan(&name, &price, &status)
	if err != nil {
		return nil, fmt.Errorf("商品不存在")
	}
	if status != "ENABLED" {
		return nil, fmt.Errorf("商品未上架")
	}

	if err := s.checkBalance(uid, price); err != nil {
		return nil, err
	}

	var codeID int
	var code string
	err = database.DB.QueryRow("SELECT id, code FROM points_product_code WHERE product_id = ? AND status = 'AVAILABLE' LIMIT 1", productID).Scan(&codeID, &code)
	if err != nil {
		return nil, fmt.Errorf("库存不足")
	}

	res, err := database.DB.Exec("UPDATE points_product_code SET status = 'EXCHANGED', exchanged_by = ?, exchanged_at = NOW() WHERE id = ? AND status = 'AVAILABLE'", uid, codeID)
	if err != nil {
		return nil, fmt.Errorf("兑换失败")
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return nil, fmt.Errorf("兑换码已被使用")
	}

	s.deductFee(uid, price)

	database.DB.Exec("INSERT INTO points_exchange_record (uid, product_id, product_name, code_id, code, points_cost) VALUES (?, ?, ?, ?, ?, ?)",
		uid, productID, name, codeID, code, price)

	return map[string]interface{}{"code": code, "product_name": name, "cost": price}, nil
}

func (s *TuboshuService) getExchangeRecords(uid, page, pageSize int) (map[string]interface{}, error) {
	offset := (page - 1) * pageSize
	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM points_exchange_record WHERE uid = ?", uid).Scan(&total)

	rows, err := database.DB.Query("SELECT id, product_name, code, points_cost, create_time FROM points_exchange_record WHERE uid = ? ORDER BY id DESC LIMIT ?, ?", uid, offset, pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []map[string]interface{}
	for rows.Next() {
		var id int
		var productName, code, createTime string
		var cost float64
		rows.Scan(&id, &productName, &code, &cost, &createTime)
		records = append(records, map[string]interface{}{
			"id": id, "product_name": productName, "code": code, "cost": cost, "create_time": createTime,
		})
	}
	if records == nil {
		records = []map[string]interface{}{}
	}
	return map[string]interface{}{"list": records, "total": total, "page": page, "pageSize": pageSize}, nil
}

func (s *TuboshuService) getAdminProducts(page, pageSize int) (map[string]interface{}, error) {
	offset := (page - 1) * pageSize
	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM points_product").Scan(&total)

	rows, err := database.DB.Query("SELECT id, name, description, image_url, price, status, sort_order, create_time FROM points_product ORDER BY sort_order DESC, id DESC LIMIT ?, ?", offset, pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var id, sortOrder int
		var name, desc, imageURL, status, createTime string
		var price float64
		rows.Scan(&id, &name, &desc, &imageURL, &price, &status, &sortOrder, &createTime)

		var stock int
		database.DB.QueryRow("SELECT COUNT(*) FROM points_product_code WHERE product_id = ? AND status = 'AVAILABLE'", id).Scan(&stock)

		products = append(products, map[string]interface{}{
			"id": id, "name": name, "description": desc, "image_url": imageURL,
			"price": price, "status": status, "sort_order": sortOrder,
			"create_time": createTime, "stock": stock,
		})
	}
	if products == nil {
		products = []map[string]interface{}{}
	}
	return map[string]interface{}{"list": products, "total": total, "page": page, "pageSize": pageSize}, nil
}

func (s *TuboshuService) saveProduct(params map[string]interface{}) error {
	name, _ := params["name"].(string)
	desc, _ := params["description"].(string)
	imageURL, _ := params["image_url"].(string)
	price := getFloat64FromInterface(params["price"])
	status, _ := params["status"].(string)
	if status == "" {
		status = "ENABLED"
	}
	sortOrder := getIntFromInterface(params["sort_order"])
	id := getIntFromInterface(params["id"])

	if id > 0 {
		_, err := database.DB.Exec("UPDATE points_product SET name=?, description=?, image_url=?, price=?, status=?, sort_order=? WHERE id=?",
			name, desc, imageURL, price, status, sortOrder, id)
		return err
	}
	_, err := database.DB.Exec("INSERT INTO points_product (name, description, image_url, price, status, sort_order) VALUES (?, ?, ?, ?, ?, ?)",
		name, desc, imageURL, price, status, sortOrder)
	return err
}

func (s *TuboshuService) deleteProduct(id int) error {
	_, err := database.DB.Exec("DELETE FROM points_product WHERE id = ?", id)
	return err
}

func (s *TuboshuService) getProductCodes(productID, page, pageSize int) (map[string]interface{}, error) {
	offset := (page - 1) * pageSize
	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM points_product_code WHERE product_id = ?", productID).Scan(&total)

	rows, err := database.DB.Query("SELECT id, code, status, exchanged_by, exchanged_at, create_time FROM points_product_code WHERE product_id = ? ORDER BY id DESC LIMIT ?, ?", productID, offset, pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []map[string]interface{}
	for rows.Next() {
		var id int
		var code, status string
		var exchangedBy *int
		var exchangedAt, createTime *string
		rows.Scan(&id, &code, &status, &exchangedBy, &exchangedAt, &createTime)
		item := map[string]interface{}{
			"id": id, "code": code, "status": status, "create_time": createTime,
		}
		if exchangedBy != nil {
			item["exchanged_by"] = *exchangedBy
		}
		if exchangedAt != nil {
			item["exchanged_at"] = *exchangedAt
		}
		codes = append(codes, item)
	}
	if codes == nil {
		codes = []map[string]interface{}{}
	}
	return map[string]interface{}{"list": codes, "total": total, "page": page, "pageSize": pageSize}, nil
}

func (s *TuboshuService) addCodes(productID int, codes []interface{}) error {
	if productID <= 0 {
		return fmt.Errorf("商品ID无效")
	}
	for _, c := range codes {
		code := fmt.Sprintf("%v", c)
		if code != "" {
			database.DB.Exec("INSERT INTO points_product_code (product_id, code) VALUES (?, ?)", productID, code)
		}
	}
	return nil
}

func (s *TuboshuService) deleteCode(id int) error {
	_, err := database.DB.Exec("DELETE FROM points_product_code WHERE id = ? AND status = 'AVAILABLE'", id)
	return err
}

func (s *TuboshuService) getAdminExchangeRecords(page, pageSize int) (map[string]interface{}, error) {
	offset := (page - 1) * pageSize
	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM points_exchange_record").Scan(&total)

	rows, err := database.DB.Query("SELECT id, uid, product_name, code, points_cost, create_time FROM points_exchange_record ORDER BY id DESC LIMIT ?, ?", offset, pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []map[string]interface{}
	for rows.Next() {
		var id, uid int
		var productName, code, createTime string
		var cost float64
		rows.Scan(&id, &uid, &productName, &code, &cost, &createTime)
		records = append(records, map[string]interface{}{
			"id": id, "uid": uid, "product_name": productName, "code": code, "cost": cost, "create_time": createTime,
		})
	}
	if records == nil {
		records = []map[string]interface{}{}
	}
	return map[string]interface{}{"list": records, "total": total, "page": page, "pageSize": pageSize}, nil
}

func getIntParam(params map[string]interface{}, key string, defaultVal int) int {
	if v, ok := params[key]; ok {
		return getIntFromInterface(v)
	}
	return defaultVal
}

func getIntFromInterface(v interface{}) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case int64:
		return int(val)
	case string:
		n, _ := strconv.Atoi(val)
		return n
	case json.Number:
		n, _ := val.Int64()
		return int(n)
	}
	return 0
}

func getInt64FromInterface(v interface{}) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case int:
		return int64(val)
	case int64:
		return val
	case string:
		n, _ := strconv.ParseInt(val, 10, 64)
		return n
	case json.Number:
		n, _ := val.Int64()
		return n
	}
	return 0
}

func getFloat64FromInterface(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		n, _ := strconv.ParseFloat(val, 64)
		return n
	case json.Number:
		n, _ := val.Float64()
		return n
	}
	return 0
}
