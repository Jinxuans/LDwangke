package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-api/internal/cache"
	"go-api/internal/database"
)

// ==================== 常量 ====================

const (
	zwBaseURL       = "http://www.zwgflw.top/"
	zwTokenCacheKey = "zhiwen_paper_token"
	zwTokenTTL      = 23 * time.Hour
	zwMaxRetry      = 3
)

// ==================== 类型定义 ====================

// PaperConfig 智文论文配置
type PaperConfig struct {
	Username string            `json:"lunwen_api_username"`
	Password string            `json:"lunwen_api_password"`
	Prices   map[string]string `json:"prices"`
}

// PaperOrder 论文订单记录
type PaperOrder struct {
	ID       int     `json:"id"`
	UID      int     `json:"uid"`
	OrderID  string  `json:"order_id"`
	ShopCode string  `json:"shopcode"`
	Title    string  `json:"title"`
	Price    float64 `json:"price"`
}

// PaperService 智文论文服务
type PaperService struct {
	client *http.Client
	mu     sync.RWMutex
}

var paperServiceInstance *PaperService
var paperServiceOnce sync.Once

// NewPaperService 获取单例服务
func NewPaperService() *PaperService {
	paperServiceOnce.Do(func() {
		paperServiceInstance = &PaperService{
			client: &http.Client{Timeout: 120 * time.Second},
		}
	})
	return paperServiceInstance
}

// ==================== 建表 ====================

func (s *PaperService) EnsureTable() {
	_, err := database.DB.Exec(`CREATE TABLE IF NOT EXISTS qingka_wangke_lunwen (
		id int(11) UNSIGNED NOT NULL AUTO_INCREMENT,
		uid int(11) UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
		order_id varchar(255) DEFAULT NULL COMMENT '上游订单ID',
		shopcode varchar(100) DEFAULT NULL COMMENT '商品代码',
		title varchar(255) DEFAULT NULL COMMENT '论文标题',
		price decimal(10,2) UNSIGNED DEFAULT NULL COMMENT '扣费价格',
		PRIMARY KEY (id),
		KEY idx_uid (uid),
		KEY idx_order_id (order_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='智文论文订单表'`)
	if err != nil {
		log.Printf("[Paper] 创建 lunwen 表失败: %v", err)
	}

	// 初始化默认配置
	defaults := map[string]string{
		"lunwen_api_username":       "",
		"lunwen_api_password":       "",
		"lunwen_api_6000_price":     "30",
		"lunwen_api_8000_price":     "40",
		"lunwen_api_10000_price":    "50",
		"lunwen_api_12000_price":    "60",
		"lunwen_api_15000_price":    "75",
		"lunwen_api_rws_price":      "10",
		"lunwen_api_ktbg_price":     "10",
		"lunwen_api_jdaigchj_price": "10",
		"lunwen_api_xgdl_price":     "3",
		"lunwen_api_jcl_price":      "3",
		"lunwen_api_jdaigcl_price":  "3",
	}
	for k, v := range defaults {
		var count int
		database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_config WHERE v = ?", k).Scan(&count)
		if count == 0 {
			database.DB.Exec("INSERT INTO qingka_wangke_config (v, k) VALUES (?, ?)", k, v)
		}
	}
}

// ==================== 配置管理 ====================

// GetConfig 获取论文配置
func (s *PaperService) GetConfig() (map[string]string, error) {
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

// SaveConfig 保存论文配置
func (s *PaperService) SaveConfig(data map[string]string) error {
	for k, v := range data {
		if !strings.HasPrefix(k, "lunwen_api_") {
			continue
		}
		var count int
		database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_config WHERE v = ?", k).Scan(&count)
		if count > 0 {
			_, err := database.DB.Exec("UPDATE qingka_wangke_config SET k = ? WHERE v = ?", v, k)
			if err != nil {
				return err
			}
		} else {
			_, err := database.DB.Exec("INSERT INTO qingka_wangke_config (v, k) VALUES (?, ?)", k, v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetConfigPrice 获取指定价格配置
func (s *PaperService) GetConfigPrice(key string) float64 {
	var val string
	err := database.DB.QueryRow("SELECT k FROM qingka_wangke_config WHERE v = ? LIMIT 1", key).Scan(&val)
	if err != nil {
		return 0
	}
	f, _ := strconv.ParseFloat(val, 64)
	return f
}

// GetUserAddPrice 获取用户费率
func (s *PaperService) GetUserAddPrice(uid int) float64 {
	var addprice float64
	database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
	if addprice == 0 {
		addprice = 1
	}
	return addprice
}

// GetUserMoney 获取用户余额
func (s *PaperService) GetUserMoney(uid int) float64 {
	var money float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&money)
	return money
}

// ==================== Token 管理 ====================

func (s *PaperService) getToken() (string, error) {
	ctx := context.Background()

	// 从 Redis 缓存
	cached, err := cache.RDB.Get(ctx, zwTokenCacheKey).Result()
	if err == nil && cached != "" {
		return cached, nil
	}

	// 从数据库获取账号密码并登录
	conf, err := s.GetConfig()
	if err != nil {
		return "", fmt.Errorf("获取论文配置失败: %v", err)
	}
	username := conf["lunwen_api_username"]
	password := conf["lunwen_api_password"]
	if username == "" || password == "" {
		return "", fmt.Errorf("论文API未配置账号密码")
	}

	// 登录
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

	// 缓存 token
	cache.RDB.Set(ctx, zwTokenCacheKey, token, zwTokenTTL)
	return token, nil
}

func (s *PaperService) clearToken() {
	ctx := context.Background()
	cache.RDB.Del(ctx, zwTokenCacheKey)
}

// ==================== 上游 API 请求 ====================

// apiGet 发送 GET 请求到上游
func (s *PaperService) apiGet(path string, params map[string]string) (map[string]interface{}, error) {
	return s.apiRequest("GET", path, nil, params)
}

// apiPost 发送 POST 请求到上游
func (s *PaperService) apiPost(path string, data interface{}) (map[string]interface{}, error) {
	return s.apiRequest("POST", path, data, nil)
}

// apiRequest 通用请求方法
func (s *PaperService) apiRequest(method, path string, data interface{}, queryParams map[string]string) (map[string]interface{}, error) {
	var lastErr error
	for attempt := 0; attempt < zwMaxRetry; attempt++ {
		result, err := s.doAPIRequest(method, path, data, queryParams)
		if err != nil {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// 检查 401
		code, _ := result["code"].(float64)
		if int(code) == 401 {
			log.Printf("[Paper] 收到401，重新登录")
			s.clearToken()
			time.Sleep(500 * time.Millisecond)
			continue
		}

		return result, nil
	}
	return nil, fmt.Errorf("论文API多次请求失败: %v", lastErr)
}

func (s *PaperService) doAPIRequest(method, path string, data interface{}, queryParams map[string]string) (map[string]interface{}, error) {
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
		for k, v := range queryParams {
			url += sep + k + "=" + v
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

// apiStreamRequest 流式请求（SSE）, 将上游SSE直接转发到 ResponseWriter
func (s *PaperService) apiStreamRequest(path string, data interface{}, w http.ResponseWriter) error {
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

	// 检查是否为SSE响应
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		// 非SSE，可能是错误响应
		body, _ := io.ReadAll(resp.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
		return nil
	}

	// 转发SSE
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

// apiFileUpload 文件上传到上游
func (s *PaperService) apiFileUpload(path string, file multipart.File, fileHeader *multipart.FileHeader, extraFields map[string]string) (map[string]interface{}, error) {
	token, err := s.getToken()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 写入文件
	part, err := writer.CreateFormFile("file", fileHeader.Filename)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}

	// 写入额外字段
	for k, v := range extraFields {
		writer.WriteField(k, v)
	}
	writer.Close()

	url := zwBaseURL + path
	req, err := http.NewRequest("POST", url, &buf)
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

// ==================== 业务方法 ====================

// GenerateTitles 生成论文标题
func (s *PaperService) GenerateTitles(params map[string]interface{}) (map[string]interface{}, error) {
	return s.apiPost("prod-api/system/lunwen/generate-titles", params)
}

// GenerateOutline 生成论文大纲
func (s *PaperService) GenerateOutline(params map[string]interface{}) (map[string]interface{}, error) {
	return s.apiPost("prod-api/system/lunwen/generate-outline", params)
}

// OutlineStatus 获取大纲状态
func (s *PaperService) OutlineStatus(orderID string) (map[string]interface{}, error) {
	return s.apiGet("prod-api/system/lunwen/outline-status/"+orderID, nil)
}

// GetShopList 获取商品列表
func (s *PaperService) GetShopList() (map[string]interface{}, error) {
	return s.apiGet("prod-api/wk/ShopInfo/getShopList", map[string]string{"type": "1"})
}

// GetShopPrice 获取商品价格
func (s *PaperService) GetShopPrice(shopCode string) (map[string]interface{}, error) {
	return s.apiGet("prod-api/wk/userPrice/getShopPrice", map[string]string{"ShopCode": shopCode})
}

// GetTemplateList 获取模板列表
func (s *PaperService) GetTemplateList(params map[string]string) (map[string]interface{}, error) {
	return s.apiGet("prod-api/system/template/list", params)
}

// SaveTemplate 保存模板
func (s *PaperService) SaveTemplate(params map[string]interface{}) (map[string]interface{}, error) {
	return s.apiPost("prod-api/system/template", params)
}

// PaperDownload 下载论文
func (s *PaperService) PaperDownload(orderID, fileName string) (map[string]interface{}, error) {
	return s.apiGet("prod-api/system/lunwen/download/"+orderID, map[string]string{"fileName": fileName})
}

// GetUpstreamList 获取上游论文列表
func (s *PaperService) GetUpstreamList(params map[string]string) (map[string]interface{}, error) {
	return s.apiGet("prod-api/system/lunwen/list", params)
}

// GenerateTask 生成任务书
func (s *PaperService) GenerateTask(params map[string]interface{}) (map[string]interface{}, error) {
	return s.apiPost("prod-api/system/lunwen/generate-task", params)
}

// GenerateProposal 生成开题报告
func (s *PaperService) GenerateProposal(params map[string]interface{}) (map[string]interface{}, error) {
	return s.apiPost("prod-api/system/lunwen/generate-proposal", params)
}

// CountWords 统计字数（文件上传）
func (s *PaperService) CountWords(file multipart.File, header *multipart.FileHeader) (map[string]interface{}, error) {
	return s.apiFileUpload("prod-api/system/lunwen/countWords", file, header, nil)
}

// UploadCover 上传模板文件
func (s *PaperService) UploadCover(file multipart.File, header *multipart.FileHeader) (map[string]interface{}, error) {
	return s.apiFileUpload("prod-api/system/template/uploadCover", file, header, nil)
}

// ==================== 下单逻辑 ====================

// PaperOrderSubmit 论文下单
func (s *PaperService) PaperOrderSubmit(uid int, params map[string]interface{}) (map[string]interface{}, error) {
	shopcode, _ := params["shopcode"].(string)
	if shopcode == "" {
		return nil, fmt.Errorf("请选择商品类型")
	}

	addprice := s.GetUserAddPrice(uid)

	// 计算价格
	var totalPrice float64
	basePrice := s.GetConfigPrice("lunwen_api_" + shopcode + "_price")
	totalPrice += math.Round(basePrice*addprice*100) / 100

	if ktbg, ok := params["ktbg"]; ok {
		if v, _ := toFloat(ktbg); v == 1 {
			totalPrice += math.Round(s.GetConfigPrice("lunwen_api_ktbg_price")*addprice*100) / 100
		}
	}
	if rws, ok := params["rws"]; ok {
		if v, _ := toFloat(rws); v == 1 {
			totalPrice += math.Round(s.GetConfigPrice("lunwen_api_rws_price")*addprice*100) / 100
		}
	}
	if jc, ok := params["jiangchong"]; ok {
		if v, _ := toFloat(jc); v == 1 {
			totalPrice += math.Round(s.GetConfigPrice("lunwen_api_jdaigchj_price")*addprice*100) / 100
		}
	}

	// 余额检查
	money := s.GetUserMoney(uid)
	if money < totalPrice {
		return nil, fmt.Errorf("余额不足")
	}

	// 调用上游下单
	result, err := s.apiPost("prod-api/system/lunwen/xiadan", params)
	if err != nil {
		return nil, err
	}

	code, _ := result["code"].(float64)
	if int(code) != 200 {
		msg, _ := result["msg"].(string)
		return nil, fmt.Errorf("下单失败: %s", msg)
	}

	// 查询上游订单信息
	title, _ := params["title"].(string)
	listParams := map[string]string{
		"pageNum":  "1",
		"pageSize": "10",
		"shopname": "论文" + shopcode + "字",
		"title":    title,
	}
	listResult, err := s.GetUpstreamList(listParams)
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

	// 扣费
	s.deductMoney(uid, totalPrice, "lunwen-下单成功", title)

	return map[string]interface{}{"code": 200, "msg": "下单成功"}, nil
}

// FileDedupSubmit 文件降重提交
func (s *PaperService) FileDedupSubmit(uid int, file multipart.File, header *multipart.FileHeader, wordCount int, aigc, jiangchong int) (map[string]interface{}, error) {
	addprice := s.GetUserAddPrice(uid)

	var totalPrice float64
	if aigc == 1 {
		totalPrice += math.Round(float64(wordCount)/1000*s.GetConfigPrice("lunwen_api_xgdl_price")*addprice*100) / 100
	}
	if jiangchong == 1 {
		totalPrice += math.Round(float64(wordCount)/1000*s.GetConfigPrice("lunwen_api_jdaigcl_price")*addprice*100) / 100
	}

	money := s.GetUserMoney(uid)
	if money < totalPrice {
		return nil, fmt.Errorf("余额不足")
	}

	// 上传文件降重
	extra := map[string]string{
		"wordCount":  strconv.Itoa(wordCount),
		"aigc":       strconv.Itoa(aigc),
		"jiangchong": strconv.Itoa(jiangchong),
	}
	result, err := s.apiFileUpload("prod-api/system/lunwen/jiangchong", file, header, extra)
	if err != nil {
		return nil, err
	}

	code, _ := result["code"].(float64)
	if int(code) == 200 {
		// 查询上游订单
		listParams := map[string]string{
			"pageNum":  "1",
			"pageSize": "10",
			"shopname": "论文降重",
		}
		listResult, err := s.GetUpstreamList(listParams)
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

// TextRewriteSubmit 文本降重
func (s *PaperService) TextRewriteSubmit(uid int, content string, w http.ResponseWriter) error {
	charCount := len([]rune(content))
	addprice := s.GetUserAddPrice(uid)
	price := math.Round(float64(charCount)/1000*s.GetConfigPrice("lunwen_api_jcl_price")*addprice*100) / 100

	money := s.GetUserMoney(uid)
	if money < price {
		return fmt.Errorf("余额不足")
	}

	err := s.apiStreamRequest("prod-api/system/lunwen/rewrite/stream", map[string]interface{}{
		"content": content,
	}, w)
	if err != nil {
		return err
	}

	// 扣费 + 记录
	listParams := map[string]string{"pageNum": "1", "pageSize": "10"}
	listResult, err := s.GetUpstreamList(listParams)
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

// TextRewriteAIGCSubmit 降低AIGC率
func (s *PaperService) TextRewriteAIGCSubmit(uid int, content string, w http.ResponseWriter) error {
	charCount := len([]rune(content))
	addprice := s.GetUserAddPrice(uid)
	price := math.Round(float64(charCount)/1000*s.GetConfigPrice("lunwen_api_jdaigcl_price")*addprice*100) / 100

	money := s.GetUserMoney(uid)
	if money < price {
		return fmt.Errorf("余额不足")
	}

	err := s.apiStreamRequest("prod-api/system/lunwen/rewrite-aigc/stream", map[string]interface{}{
		"content": content,
	}, w)
	if err != nil {
		return err
	}

	listParams := map[string]string{"pageNum": "1", "pageSize": "10", "shopname": "降aigc"}
	listResult, err := s.GetUpstreamList(listParams)
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

// PaperParaEditSubmit 段落修改
func (s *PaperService) PaperParaEditSubmit(uid int, content, yijian string, w http.ResponseWriter) error {
	charCount := len([]rune(content))
	addprice := s.GetUserAddPrice(uid)
	price := math.Round(float64(charCount)/1000*s.GetConfigPrice("lunwen_api_xgdl_price")*addprice*100) / 100

	money := s.GetUserMoney(uid)
	if money < price {
		return fmt.Errorf("余额不足")
	}

	err := s.apiStreamRequest("prod-api/system/lunwen/xiugai/stream", map[string]interface{}{
		"content": content,
		"yijian":  yijian,
	}, w)
	if err != nil {
		return err
	}

	listParams := map[string]string{"pageNum": "1", "pageSize": "10", "shopname": "段落修改"}
	listResult, err := s.GetUpstreamList(listParams)
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

// GenerateTaskWithFee 生成任务书（扣费）
func (s *PaperService) GenerateTaskWithFee(uid int, orderID string) (map[string]interface{}, error) {
	addprice := s.GetUserAddPrice(uid)
	price := math.Round(s.GetConfigPrice("lunwen_api_rws_price")*addprice*100) / 100

	money := s.GetUserMoney(uid)
	if money < price {
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

// GenerateProposalWithFee 生成开题报告（扣费）
func (s *PaperService) GenerateProposalWithFee(uid int, orderID string) (map[string]interface{}, error) {
	addprice := s.GetUserAddPrice(uid)
	price := math.Round(s.GetConfigPrice("lunwen_api_ktbg_price")*addprice*100) / 100

	money := s.GetUserMoney(uid)
	if money < price {
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

// GetOrderList 获取论文订单列表（合并本地+上游数据）
func (s *PaperService) GetOrderList(uid int, isAdmin bool, page, pageSize int, searchParams map[string]string) (map[string]interface{}, error) {
	offset := (page - 1) * pageSize

	// 查询本地订单
	whereClause := ""
	if !isAdmin {
		whereClause = fmt.Sprintf("WHERE uid = %d", uid)
	}

	var total int
	database.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM qingka_wangke_lunwen %s", whereClause)).Scan(&total)

	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT id, uid, order_id, shopcode, title, price FROM qingka_wangke_lunwen %s ORDER BY id DESC LIMIT ?, ?", whereClause),
		offset, pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	localOrders := make(map[string]PaperOrder)
	for rows.Next() {
		var o PaperOrder
		rows.Scan(&o.ID, &o.UID, &o.OrderID, &o.ShopCode, &o.Title, &o.Price)
		localOrders[o.OrderID] = o
	}

	if len(localOrders) == 0 {
		return map[string]interface{}{
			"code":  200,
			"msg":   "查询成功",
			"rows":  []interface{}{},
			"total": 0,
		}, nil
	}

	// 查询上游订单详情
	upstreamParams := map[string]string{
		"pageNum":  "1",
		"pageSize": "1000",
	}
	for k, v := range searchParams {
		if v != "" {
			upstreamParams[k] = v
		}
	}

	apiResult, err := s.GetUpstreamList(upstreamParams)
	if err != nil {
		return nil, err
	}

	apiCode, _ := apiResult["code"].(float64)
	if int(apiCode) != 200 {
		return map[string]interface{}{
			"code":  200,
			"msg":   "查询成功",
			"rows":  []interface{}{},
			"total": 0,
		}, nil
	}

	// 合并数据
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

	return map[string]interface{}{
		"code":  200,
		"msg":   "查询成功",
		"rows":  mergedRows,
		"total": total,
	}, nil
}

// GetPriceInfo 获取价格信息（前端展示用）
func (s *PaperService) GetPriceInfo(uid int) map[string]interface{} {
	addprice := s.GetUserAddPrice(uid)
	return map[string]interface{}{
		"price_6000":     math.Round(s.GetConfigPrice("lunwen_api_6000_price")*addprice*100) / 100,
		"price_8000":     math.Round(s.GetConfigPrice("lunwen_api_8000_price")*addprice*100) / 100,
		"price_10000":    math.Round(s.GetConfigPrice("lunwen_api_10000_price")*addprice*100) / 100,
		"price_12000":    math.Round(s.GetConfigPrice("lunwen_api_12000_price")*addprice*100) / 100,
		"price_15000":    math.Round(s.GetConfigPrice("lunwen_api_15000_price")*addprice*100) / 100,
		"price_rws":      math.Round(s.GetConfigPrice("lunwen_api_rws_price")*addprice*100) / 100,
		"price_ktbg":     math.Round(s.GetConfigPrice("lunwen_api_ktbg_price")*addprice*100) / 100,
		"price_jdaigchj": math.Round(s.GetConfigPrice("lunwen_api_jdaigchj_price")*addprice*100) / 100,
		"price_xgdl":     math.Round(s.GetConfigPrice("lunwen_api_xgdl_price")*addprice*100) / 100,
		"price_jcl":      math.Round(s.GetConfigPrice("lunwen_api_jcl_price")*addprice*100) / 100,
		"price_jdaigcl":  math.Round(s.GetConfigPrice("lunwen_api_jdaigcl_price")*addprice*100) / 100,
		"addprice":       addprice,
	}
}

// ==================== 内部辅助 ====================

func (s *PaperService) insertOrder(uid int, orderID, shopcode, title string, price float64) {
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_lunwen (uid, order_id, shopcode, title, price) VALUES (?, ?, ?, ?, ?)",
		uid, orderID, shopcode, title, price)
	if err != nil {
		log.Printf("[Paper] 插入订单记录失败: %v", err)
	}
}

func (s *PaperService) deductMoney(uid int, amount float64, logType, desc string) {
	_, err := database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ? LIMIT 1", amount, uid)
	if err != nil {
		log.Printf("[Paper] 扣费失败: %v", err)
		return
	}
	// 写日志
	database.DB.Exec(
		"INSERT INTO qingka_wangke_moneylog (uid, type, money, note, addtime) VALUES (?, ?, ?, ?, NOW())",
		uid, logType, -amount, fmt.Sprintf("%s 扣除%.2f元", desc, amount))
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
