package paper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-api/internal/cache"
	"go-api/internal/database"
	obslogger "go-api/internal/observability/logger"
)

const (
	zwBaseURL       = "http://www.zwgflw.top/"
	zwTokenCacheKey = "zhiwen_paper_token"
	zwTokenTTL      = 23 * time.Hour
	zwMaxRetry      = 3
)

type paperService struct {
	client *http.Client
	mu     sync.RWMutex
}

var papers = &paperService{
	client: &http.Client{Timeout: 120 * time.Second},
}

func Paper() *paperService {
	return papers
}

func (s *paperService) GetConfig() (map[string]string, error) {
	return s.getConfig()
}

func (s *paperService) SaveConfig(data map[string]string) error {
	for k, v := range data {
		if !strings.HasPrefix(k, "lunwen_api_") {
			continue
		}
		var count int
		database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_config WHERE v = ?", k).Scan(&count)
		if count > 0 {
			if _, err := database.DB.Exec("UPDATE qingka_wangke_config SET k = ? WHERE v = ?", v, k); err != nil {
				return err
			}
			continue
		}
		if _, err := database.DB.Exec("INSERT INTO qingka_wangke_config (v, k) VALUES (?, ?)", k, v); err != nil {
			return err
		}
	}
	return nil
}

func (s *paperService) getToken() (string, error) {
	ctx := context.Background()
	cached, err := cache.RDB.Get(ctx, zwTokenCacheKey).Result()
	if err == nil && cached != "" {
		return cached, nil
	}

	conf, err := s.getConfig()
	if err != nil {
		return "", fmt.Errorf("获取论文配置失败: %v", err)
	}
	username := conf["lunwen_api_username"]
	password := conf["lunwen_api_password"]
	if username == "" || password == "" {
		return "", fmt.Errorf("论文API未配置账号密码")
	}

	loginData, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	resp, err := s.client.Post(zwBaseURL+"prod-api/login", "application/json", bytes.NewReader(loginData))
	if err != nil {
		return "", fmt.Errorf("论文API登录请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("论文API登录响应解析失败: %v", err)
	}

	code, _ := result["code"].(float64)
	if int(code) != 200 {
		msg, _ := result["msg"].(string)
		return "", fmt.Errorf("论文API登录失败: %s", msg)
	}

	token, ok := result["token"].(string)
	if !ok || token == "" {
		return "", fmt.Errorf("论文API登录返回空token")
	}

	cache.RDB.Set(ctx, zwTokenCacheKey, token, zwTokenTTL)
	return token, nil
}

func (s *paperService) clearToken() {
	ctx := context.Background()
	cache.RDB.Del(ctx, zwTokenCacheKey)
}

func (s *paperService) apiGet(path string, params map[string]string) (map[string]interface{}, error) {
	return s.apiRequest("GET", path, nil, params)
}

func (s *paperService) apiPost(path string, data interface{}) (map[string]interface{}, error) {
	return s.apiRequest("POST", path, data, nil)
}

func (s *paperService) apiRequest(method, path string, data interface{}, queryParams map[string]string) (map[string]interface{}, error) {
	var lastErr error
	for attempt := 0; attempt < zwMaxRetry; attempt++ {
		result, err := s.doAPIRequest(method, path, data, queryParams)
		if err != nil {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}

		code, _ := result["code"].(float64)
		if int(code) == 401 {
			s.clearToken()
			time.Sleep(500 * time.Millisecond)
			continue
		}

		return result, nil
	}
	return nil, fmt.Errorf("论文API多次请求失败: %v", lastErr)
}

func (s *paperService) doAPIRequest(method, path string, data interface{}, queryParams map[string]string) (map[string]interface{}, error) {
	token, err := s.getToken()
	if err != nil {
		return nil, err
	}

	url := zwBaseURL + path
	if len(queryParams) > 0 {
		sep := "?"
		if strings.Contains(url, "?") {
			sep = "&"
		}
		for key, value := range queryParams {
			url += sep + key + "=" + value
			sep = "&"
		}
	}

	var req *http.Request
	if method == "GET" {
		req, err = http.NewRequest("GET", url, nil)
	} else {
		var body io.Reader
		if data != nil {
			jsonData, _ := json.Marshal(data)
			body = bytes.NewReader(jsonData)
		}
		req, err = http.NewRequest(method, url, body)
		if req != nil {
			req.Header.Set("Content-Type", "application/json")
		}
	}
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v, body: %s", err, string(respBody[:min(200, len(respBody))]))
	}
	return result, nil
}

func (s *paperService) apiStreamRequest(path string, data interface{}, w http.ResponseWriter) error {
	token, err := s.getToken()
	if err != nil {
		return err
	}

	url := zwBaseURL + path
	jsonData, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		body, _ := io.ReadAll(resp.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
		return nil
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming unsupported")
	}

	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
			flusher.Flush()
		}
		if err != nil {
			break
		}
	}
	return nil
}

func (s *paperService) apiFileUpload(path string, file multipart.File, fileHeader *multipart.FileHeader, extraFields map[string]string) (map[string]interface{}, error) {
	token, err := s.getToken()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", fileHeader.Filename)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}

	for key, value := range extraFields {
		writer.WriteField(key, value)
	}
	writer.Close()

	req, err := http.NewRequest("POST", zwBaseURL+path, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v", err)
	}
	return result, nil
}

func (s *paperService) getConfig() (map[string]string, error) {
	rows, err := database.DB.Query("SELECT v, k FROM qingka_wangke_config WHERE v LIKE 'lunwen_api_%'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conf := make(map[string]string)
	for rows.Next() {
		var key, val string
		rows.Scan(&key, &val)
		conf[key] = val
	}
	return conf, nil
}

func (s *paperService) getConfigPrice(key string) float64 {
	var val string
	err := database.DB.QueryRow("SELECT k FROM qingka_wangke_config WHERE v = ? LIMIT 1", key).Scan(&val)
	if err != nil {
		return 0
	}
	f, _ := strconv.ParseFloat(val, 64)
	return f
}

func (s *paperService) getUserAddPrice(uid int) float64 {
	var addprice float64
	database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
	if addprice == 0 {
		addprice = 1
	}
	return addprice
}

func (s *paperService) getUserMoney(uid int) float64 {
	var money float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&money)
	return money
}

func (s *paperService) GenerateTitles(params map[string]interface{}) (map[string]interface{}, error) {
	return s.apiPost("prod-api/system/lunwen/generate-titles", params)
}

func (s *paperService) GenerateOutline(params map[string]interface{}) (map[string]interface{}, error) {
	return s.apiPost("prod-api/system/lunwen/generate-outline", params)
}

func (s *paperService) OutlineStatus(orderID string) (map[string]interface{}, error) {
	return s.apiGet("prod-api/system/lunwen/outline-status/"+orderID, nil)
}

func (s *paperService) GetTemplateList(params map[string]string) (map[string]interface{}, error) {
	return s.apiGet("prod-api/system/template/list", params)
}

func (s *paperService) SaveTemplate(params map[string]interface{}) (map[string]interface{}, error) {
	return s.apiPost("prod-api/system/template", params)
}

func (s *paperService) PaperDownload(orderID, fileName string) (map[string]interface{}, error) {
	return s.apiGet("prod-api/system/lunwen/download/"+orderID, map[string]string{"fileName": fileName})
}

func (s *paperService) GetUpstreamList(params map[string]string) (map[string]interface{}, error) {
	return s.apiGet("prod-api/system/lunwen/list", params)
}

func (s *paperService) GenerateTask(params map[string]interface{}) (map[string]interface{}, error) {
	return s.apiPost("prod-api/system/lunwen/generate-task", params)
}

func (s *paperService) GenerateProposal(params map[string]interface{}) (map[string]interface{}, error) {
	return s.apiPost("prod-api/system/lunwen/generate-proposal", params)
}

func (s *paperService) CountWords(file multipart.File, header *multipart.FileHeader) (map[string]interface{}, error) {
	return s.apiFileUpload("prod-api/system/lunwen/countWords", file, header, nil)
}

func (s *paperService) UploadCover(file multipart.File, header *multipart.FileHeader) (map[string]interface{}, error) {
	return s.apiFileUpload("prod-api/system/template/uploadCover", file, header, nil)
}

func (s *paperService) PaperOrderSubmit(uid int, params map[string]interface{}) (map[string]interface{}, error) {
	shopcode, _ := params["shopcode"].(string)
	if shopcode == "" {
		return nil, fmt.Errorf("请选择商品类型")
	}

	addprice := s.getUserAddPrice(uid)
	totalPrice := math.Round(s.getConfigPrice("lunwen_api_"+shopcode+"_price")*addprice*100) / 100

	if ktbg, ok := params["ktbg"]; ok {
		if v, _ := toFloat(ktbg); v == 1 {
			totalPrice += math.Round(s.getConfigPrice("lunwen_api_ktbg_price")*addprice*100) / 100
		}
	}
	if rws, ok := params["rws"]; ok {
		if v, _ := toFloat(rws); v == 1 {
			totalPrice += math.Round(s.getConfigPrice("lunwen_api_rws_price")*addprice*100) / 100
		}
	}
	if jc, ok := params["jiangchong"]; ok {
		if v, _ := toFloat(jc); v == 1 {
			totalPrice += math.Round(s.getConfigPrice("lunwen_api_jdaigchj_price")*addprice*100) / 100
		}
	}

	if s.getUserMoney(uid) < totalPrice {
		return nil, fmt.Errorf("余额不足")
	}

	result, err := s.apiPost("prod-api/system/lunwen/xiadan", params)
	if err != nil {
		return nil, err
	}

	code, _ := result["code"].(float64)
	if int(code) != 200 {
		msg, _ := result["msg"].(string)
		return nil, fmt.Errorf("下单失败: %s", msg)
	}

	title, _ := params["title"].(string)
	listResult, err := s.GetUpstreamList(map[string]string{
		"pageNum":  "1",
		"pageSize": "10",
		"shopname": "论文" + shopcode + "字",
		"title":    title,
	})
	if err == nil {
		listCode, _ := listResult["code"].(float64)
		if int(listCode) == 200 {
			if rows, ok := listResult["rows"].([]interface{}); ok && len(rows) > 0 {
				if row, ok := rows[0].(map[string]interface{}); ok {
					orderID := fmt.Sprintf("%v", row["id"])
					s.insertOrder(uid, orderID, shopcode, title, totalPrice)
				}
			}
		}
	}

	s.deductMoney(uid, totalPrice, "lunwen-下单成功", title)
	return map[string]interface{}{"code": 200, "msg": "下单成功"}, nil
}

func (s *paperService) FileDedupSubmit(uid int, file multipart.File, header *multipart.FileHeader, wordCount int, aigc, jiangchong int) (map[string]interface{}, error) {
	addprice := s.getUserAddPrice(uid)
	var totalPrice float64
	if aigc == 1 {
		totalPrice += math.Round(float64(wordCount)/1000*s.getConfigPrice("lunwen_api_xgdl_price")*addprice*100) / 100
	}
	if jiangchong == 1 {
		totalPrice += math.Round(float64(wordCount)/1000*s.getConfigPrice("lunwen_api_jdaigcl_price")*addprice*100) / 100
	}

	if s.getUserMoney(uid) < totalPrice {
		return nil, fmt.Errorf("余额不足")
	}

	result, err := s.apiFileUpload("prod-api/system/lunwen/jiangchong", file, header, map[string]string{
		"wordCount":  strconv.Itoa(wordCount),
		"aigc":       strconv.Itoa(aigc),
		"jiangchong": strconv.Itoa(jiangchong),
	})
	if err != nil {
		return nil, err
	}

	code, _ := result["code"].(float64)
	if int(code) == 200 {
		listResult, err := s.GetUpstreamList(map[string]string{"pageNum": "1", "pageSize": "10", "shopname": "论文降重"})
		if err == nil {
			listCode, _ := listResult["code"].(float64)
			if int(listCode) == 200 {
				if rows, ok := listResult["rows"].([]interface{}); ok && len(rows) > 0 {
					if row, ok := rows[0].(map[string]interface{}); ok {
						orderID := fmt.Sprintf("%v", row["id"])
						rowTitle := fmt.Sprintf("%v", row["title"])
						rowShopcode := fmt.Sprintf("%v", row["shopcode"])
						s.insertOrder(uid, orderID, rowShopcode, rowTitle, totalPrice)
					}
				}
			}
		}
		s.deductMoney(uid, totalPrice, "lunwen-文件降重成功", "文件降重")
	}
	return result, nil
}

func (s *paperService) TextRewriteSubmit(uid int, content string, w http.ResponseWriter) error {
	price := math.Round(float64(len([]rune(content)))/1000*s.getConfigPrice("lunwen_api_jcl_price")*s.getUserAddPrice(uid)*100) / 100
	if s.getUserMoney(uid) < price {
		return fmt.Errorf("余额不足")
	}

	if err := s.apiStreamRequest("prod-api/system/lunwen/rewrite/stream", map[string]interface{}{"content": content}, w); err != nil {
		return err
	}

	listResult, err := s.GetUpstreamList(map[string]string{"pageNum": "1", "pageSize": "10"})
	if err == nil {
		listCode, _ := listResult["code"].(float64)
		if int(listCode) == 200 {
			if rows, ok := listResult["rows"].([]interface{}); ok && len(rows) > 0 {
				if row, ok := rows[0].(map[string]interface{}); ok {
					orderID := fmt.Sprintf("%v", row["id"])
					rowTitle := fmt.Sprintf("%v", row["title"])
					rowShopcode := fmt.Sprintf("%v", row["shopcode"])
					s.insertOrder(uid, orderID, rowShopcode, rowTitle, price)
				}
			}
		}
	}
	s.deductMoney(uid, price, "lunwen-文本降重成功", "文本降重")
	return nil
}

func (s *paperService) TextRewriteAIGCSubmit(uid int, content string, w http.ResponseWriter) error {
	price := math.Round(float64(len([]rune(content)))/1000*s.getConfigPrice("lunwen_api_jdaigcl_price")*s.getUserAddPrice(uid)*100) / 100
	if s.getUserMoney(uid) < price {
		return fmt.Errorf("余额不足")
	}

	if err := s.apiStreamRequest("prod-api/system/lunwen/rewrite-aigc/stream", map[string]interface{}{"content": content}, w); err != nil {
		return err
	}

	listResult, err := s.GetUpstreamList(map[string]string{"pageNum": "1", "pageSize": "10", "shopname": "降aigc"})
	if err == nil {
		listCode, _ := listResult["code"].(float64)
		if int(listCode) == 200 {
			if rows, ok := listResult["rows"].([]interface{}); ok && len(rows) > 0 {
				if row, ok := rows[0].(map[string]interface{}); ok {
					orderID := fmt.Sprintf("%v", row["id"])
					rowTitle := fmt.Sprintf("%v", row["title"])
					rowShopcode := fmt.Sprintf("%v", row["shopcode"])
					s.insertOrder(uid, orderID, rowShopcode, rowTitle, price)
				}
			}
		}
	}
	s.deductMoney(uid, price, "lunwen-文本降AIGC成功", "降AIGC")
	return nil
}

func (s *paperService) PaperParaEditSubmit(uid int, content, yijian string, w http.ResponseWriter) error {
	price := math.Round(float64(len([]rune(content)))/1000*s.getConfigPrice("lunwen_api_xgdl_price")*s.getUserAddPrice(uid)*100) / 100
	if s.getUserMoney(uid) < price {
		return fmt.Errorf("余额不足")
	}

	if err := s.apiStreamRequest("prod-api/system/lunwen/xiugai/stream", map[string]interface{}{"content": content, "yijian": yijian}, w); err != nil {
		return err
	}

	listResult, err := s.GetUpstreamList(map[string]string{"pageNum": "1", "pageSize": "10", "shopname": "段落修改"})
	if err == nil {
		listCode, _ := listResult["code"].(float64)
		if int(listCode) == 200 {
			if rows, ok := listResult["rows"].([]interface{}); ok && len(rows) > 0 {
				if row, ok := rows[0].(map[string]interface{}); ok {
					orderID := fmt.Sprintf("%v", row["id"])
					rowTitle := fmt.Sprintf("%v", row["title"])
					rowShopcode := fmt.Sprintf("%v", row["shopcode"])
					s.insertOrder(uid, orderID, rowShopcode, rowTitle, price)
				}
			}
		}
	}
	s.deductMoney(uid, price, "lunwen-段落修改成功", "段落修改")
	return nil
}

func (s *paperService) GenerateTaskWithFee(uid int, orderID string) (map[string]interface{}, error) {
	price := math.Round(s.getConfigPrice("lunwen_api_rws_price")*s.getUserAddPrice(uid)*100) / 100
	if s.getUserMoney(uid) < price {
		return nil, fmt.Errorf("余额不足")
	}

	result, err := s.GenerateTask(map[string]interface{}{"id": orderID})
	if err != nil {
		return nil, err
	}
	code, _ := result["code"].(float64)
	if int(code) == 200 {
		s.insertOrder(uid, orderID, "rws", "任务书生成", price)
		s.deductMoney(uid, price, "lunwen-生成任务书", "任务书")
	}
	return result, nil
}

func (s *paperService) GenerateProposalWithFee(uid int, orderID string) (map[string]interface{}, error) {
	price := math.Round(s.getConfigPrice("lunwen_api_ktbg_price")*s.getUserAddPrice(uid)*100) / 100
	if s.getUserMoney(uid) < price {
		return nil, fmt.Errorf("余额不足")
	}

	result, err := s.GenerateProposal(map[string]interface{}{"id": orderID})
	if err != nil {
		return nil, err
	}
	code, _ := result["code"].(float64)
	if int(code) == 200 {
		s.insertOrder(uid, orderID, "ktbg", "开题报告生成", price)
		s.deductMoney(uid, price, "lunwen-生成开题报告", "开题报告")
	}
	return result, nil
}

func (s *paperService) GetOrderList(uid int, isAdmin bool, page, pageSize int, searchParams map[string]string) (map[string]interface{}, error) {
	offset := (page - 1) * pageSize
	whereClause := ""
	if !isAdmin {
		whereClause = fmt.Sprintf("WHERE uid = %d", uid)
	}

	var total int
	database.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM qingka_wangke_lunwen %s", whereClause)).Scan(&total)

	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT id, uid, order_id, shopcode, title, price FROM qingka_wangke_lunwen %s ORDER BY id DESC LIMIT ?, ?", whereClause),
		offset, pageSize,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type paperOrder struct {
		ID       int
		UID      int
		OrderID  string
		ShopCode string
		Title    string
		Price    float64
	}

	localOrders := make(map[string]paperOrder)
	for rows.Next() {
		var order paperOrder
		rows.Scan(&order.ID, &order.UID, &order.OrderID, &order.ShopCode, &order.Title, &order.Price)
		localOrders[order.OrderID] = order
	}

	if len(localOrders) == 0 {
		return map[string]interface{}{"code": 200, "msg": "查询成功", "rows": []interface{}{}, "total": 0}, nil
	}

	upstreamParams := map[string]string{"pageNum": "1", "pageSize": "1000"}
	for key, value := range searchParams {
		if value != "" {
			upstreamParams[key] = value
		}
	}

	apiResult, err := s.GetUpstreamList(upstreamParams)
	if err != nil {
		return nil, err
	}

	apiCode, _ := apiResult["code"].(float64)
	if int(apiCode) != 200 {
		return map[string]interface{}{"code": 200, "msg": "查询成功", "rows": []interface{}{}, "total": 0}, nil
	}

	var mergedRows []interface{}
	if apiRows, ok := apiResult["rows"].([]interface{}); ok {
		for _, item := range apiRows {
			if row, ok := item.(map[string]interface{}); ok {
				rowID := fmt.Sprintf("%v", row["id"])
				if localOrder, exists := localOrders[rowID]; exists {
					row["price"] = localOrder.Price
					mergedRows = append(mergedRows, row)
				}
			}
		}
	}

	return map[string]interface{}{"code": 200, "msg": "查询成功", "rows": mergedRows, "total": total}, nil
}

func (s *paperService) GetPriceInfo(uid int) map[string]interface{} {
	addprice := s.getUserAddPrice(uid)
	return map[string]interface{}{
		"price_6000":     math.Round(s.getConfigPrice("lunwen_api_6000_price")*addprice*100) / 100,
		"price_8000":     math.Round(s.getConfigPrice("lunwen_api_8000_price")*addprice*100) / 100,
		"price_10000":    math.Round(s.getConfigPrice("lunwen_api_10000_price")*addprice*100) / 100,
		"price_12000":    math.Round(s.getConfigPrice("lunwen_api_12000_price")*addprice*100) / 100,
		"price_15000":    math.Round(s.getConfigPrice("lunwen_api_15000_price")*addprice*100) / 100,
		"price_rws":      math.Round(s.getConfigPrice("lunwen_api_rws_price")*addprice*100) / 100,
		"price_ktbg":     math.Round(s.getConfigPrice("lunwen_api_ktbg_price")*addprice*100) / 100,
		"price_jdaigchj": math.Round(s.getConfigPrice("lunwen_api_jdaigchj_price")*addprice*100) / 100,
		"price_xgdl":     math.Round(s.getConfigPrice("lunwen_api_xgdl_price")*addprice*100) / 100,
		"price_jcl":      math.Round(s.getConfigPrice("lunwen_api_jcl_price")*addprice*100) / 100,
		"price_jdaigcl":  math.Round(s.getConfigPrice("lunwen_api_jdaigcl_price")*addprice*100) / 100,
		"addprice":       addprice,
	}
}

func (s *paperService) insertOrder(uid int, orderID, shopcode, title string, price float64) {
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_lunwen (uid, order_id, shopcode, title, price) VALUES (?, ?, ?, ?, ?)",
		uid, orderID, shopcode, title, price,
	)
	if err != nil {
		obslogger.L().Warn("Paper 插入订单记录失败", "error", err)
	}
}

func (s *paperService) deductMoney(uid int, amount float64, logType, desc string) {
	_, err := database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ? LIMIT 1", amount, uid)
	if err != nil {
		obslogger.L().Warn("Paper 扣费失败", "error", err)
		return
	}
	database.DB.Exec(
		"INSERT INTO qingka_wangke_moneylog (uid, type, money, note, addtime) VALUES (?, ?, ?, ?, NOW())",
		uid, logType, -amount, fmt.Sprintf("%s 扣除%.2f元", desc, amount),
	)
}

func toFloat(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case string:
		f, err := strconv.ParseFloat(val, 64)
		return f, err == nil
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	}
	return 0, false
}
