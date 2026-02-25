package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"go-api/internal/model"
)

// yyySession 管理yyy平台的登录会话
type yyySession struct {
	client    *http.Client
	token     string
	baseURL   string
	username  string
	password  string
	loggedIn  bool
	lastLogin time.Time
	mu        sync.Mutex
}

// 全局 yyy session 缓存（按 supplier URL 区分）
var (
	yyySessions   = map[string]*yyySession{}
	yyySessionsMu sync.Mutex
)

// getYYYSession 获取或创建 yyy 平台的会话
func getYYYSession(baseURL, username, password string) *yyySession {
	yyySessionsMu.Lock()
	defer yyySessionsMu.Unlock()

	key := baseURL + "|" + username
	if s, ok := yyySessions[key]; ok {
		s.password = password // 更新密码
		return s
	}

	jar, _ := cookiejar.New(nil)
	s := &yyySession{
		client:   &http.Client{Jar: jar, Timeout: 60 * time.Second},
		baseURL:  strings.TrimRight(baseURL, "/"),
		username: username,
		password: password,
	}
	yyySessions[key] = s
	return s
}

// login 登录 yyy 平台获取 token
func (s *yyySession) login() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 30分钟内不重复登录
	if s.loggedIn && time.Since(s.lastLogin) < 30*time.Minute {
		return nil
	}

	body, _ := json.Marshal(map[string]string{
		"username": s.username,
		"password": s.password,
	})

	apiURL := s.baseURL + "/api/login"
	req, _ := http.NewRequest("POST", apiURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "application/json, text/plain, */*")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("yyy登录请求失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		Code int `json:"code"`
		Data struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("yyy登录解析失败: %s", string(respBody))
	}
	if result.Code != 200 {
		return fmt.Errorf("yyy登录失败: %s", result.Message)
	}

	s.token = result.Data.AccessToken

	// 设置 cookie（模拟浏览器行为）
	cookieVal, _ := json.Marshal(map[string]interface{}{
		"accessToken":  s.token,
		"expires":      int64(1919520000000),
		"refreshToken": result.Data.RefreshToken,
	})
	u, _ := url.Parse(s.baseURL)
	s.client.Jar.SetCookies(u, []*http.Cookie{
		{Name: "authorized-token", Value: url.QueryEscape(string(cookieVal)), Path: "/"},
		{Name: "multiple-tabs", Value: "true", Path: "/"},
	})

	s.loggedIn = true
	s.lastLogin = time.Now()
	fmt.Printf("[yyy] 登录成功: %s token=%s...\n", s.baseURL, s.token[:16])
	return nil
}

// doRequest 发送带认证的请求
func (s *yyySession) doRequest(method, path string, body interface{}) ([]byte, error) {
	if err := s.login(); err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}

	apiURL := s.baseURL + path
	req, err := http.NewRequest(method, apiURL, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Referer", s.baseURL+"/")
	req.Header.Set("Origin", s.baseURL)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("yyy请求失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	// 检查是否需要重新登录
	var check struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	json.Unmarshal(respBody, &check)
	if check.Code == 302 && strings.Contains(check.Message, "重新登录") {
		s.mu.Lock()
		s.loggedIn = false
		s.mu.Unlock()
		if err := s.login(); err != nil {
			return nil, err
		}
		// 重试
		req2, _ := http.NewRequest(method, apiURL, bodyReader)
		req2.Header = req.Header.Clone()
		req2.Header.Set("Authorization", "Bearer "+s.token)
		resp2, err := s.client.Do(req2)
		if err != nil {
			return nil, err
		}
		defer resp2.Body.Close()
		respBody, _ = io.ReadAll(resp2.Body)
	}

	return respBody, nil
}

// yyyGetClasses 获取 yyy 平台商品列表
func yyyGetClasses(sup *model.SupplierFull) ([]SupplierClassItem, error) {
	baseURL := strings.TrimRight(sup.URL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	session := getYYYSession(baseURL, sup.User, sup.Pass)

	respBody, err := session.doRequest("POST", "/api/site", map[string]interface{}{"version": nil})
	if err != nil {
		return nil, err
	}

	var result struct {
		Code int `json:"code"`
		Data struct {
			List []struct {
				ID        int     `json:"id"`
				Name      string  `json:"name"`
				Trans     string  `json:"trans"`
				URL       string  `json:"url"`
				PriceUnit string  `json:"price_unit"`
				Price     float64 `json:"price"`
			} `json:"list"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("yyy解析商品列表失败: %s", string(respBody)[:200])
	}
	if result.Code != 200 {
		return nil, fmt.Errorf("yyy获取商品失败: %s", result.Message)
	}

	var classes []SupplierClassItem
	for _, item := range result.Data.List {
		// 解析 price_unit 中的价格，如 "1.5 /门"
		price := item.Price
		if price == 0 && item.PriceUnit != "" {
			parts := strings.Fields(item.PriceUnit)
			if len(parts) > 0 {
				fmt.Sscanf(parts[0], "%f", &price)
			}
		}
		classes = append(classes, SupplierClassItem{
			CID:     fmt.Sprintf("%d", item.ID),
			Name:    item.Name,
			Price:   price,
			Content: item.Trans,
		})
	}

	fmt.Printf("[yyy] 获取到 %d 个商品\n", len(classes))
	return classes, nil
}

// yyyQueryBalance 查询 yyy 平台余额（通过登录会话调用 /api/money）
func yyyQueryBalance(sup *model.SupplierFull) (map[string]interface{}, error) {
	baseURL := strings.TrimRight(sup.URL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	session := getYYYSession(baseURL, sup.User, sup.Pass)

	respBody, err := session.doRequest("POST", "/api/money", nil)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, fmt.Errorf("yyy余额解析失败: %s", string(respBody))
	}

	// yyy 返回格式: {"code":200,"data":{"money":"123.45"},"message":"success"}
	money := ""
	if data, ok := raw["data"].(map[string]interface{}); ok {
		if m, ok := data["money"]; ok {
			money = fmt.Sprintf("%v", m)
		}
	}
	// fallback: 根级 money
	if money == "" {
		if m, ok := raw["money"]; ok {
			money = fmt.Sprintf("%v", m)
		}
	}

	result := map[string]interface{}{
		"code":  200,
		"money": money,
		"pt":    sup.PT,
		"name":  sup.Name,
		"hid":   sup.HID,
		"raw":   raw,
	}
	return result, nil
}

// yyyCallOrder 向 yyy 平台下单
func yyyCallOrder(sup *model.SupplierFull, user, pass, kcname string, siteID string) (*model.SupplierOrderResult, error) {
	baseURL := strings.TrimRight(sup.URL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	session := getYYYSession(baseURL, sup.User, sup.Pass)

	// yyy下单格式: "账号 密码 课程名"
	orderData := fmt.Sprintf("%s %s %s", user, pass, kcname)

	respBody, err := session.doRequest("POST", "/api/order", map[string]interface{}{
		"lastoid":   siteID,
		"orderData": orderData,
		"orderNote": "",
		"search":    "0",
	})
	if err != nil {
		return nil, err
	}

	var result struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("yyy下单解析失败: %s", string(respBody)[:200])
	}

	orderResult := &model.SupplierOrderResult{
		Msg: result.Message,
	}

	if result.Code == 200 {
		orderResult.Code = 1
		orderResult.Msg = "下单成功"
		// 尝试从 data 中提取订单ID
		if dataArr, ok := result.Data.([]interface{}); ok && len(dataArr) > 0 {
			orderResult.YID = fmt.Sprintf("%v", dataArr[0])
		}
	} else {
		orderResult.Code = -1
	}

	return orderResult, nil
}

// yyyQueryProgress 查询 yyy 平台订单进度
func yyyQueryProgress(sup *model.SupplierFull, username string) ([]model.SupplierProgressItem, error) {
	baseURL := strings.TrimRight(sup.URL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	session := getYYYSession(baseURL, sup.User, sup.Pass)

	respBody, err := session.doRequest("POST", "/api/getorder", map[string]interface{}{
		"lastoid":   "",
		"odname":    username,
		"nickname":  "",
		"notetype":  "1",
		"note":      "",
		"statusbox": []string{},
		"page":      1,
		"pageSize":  50,
	})
	if err != nil {
		return nil, err
	}

	var result struct {
		Code int `json:"code"`
		Data struct {
			List []struct {
				ID        int    `json:"id"`
				OdName    string `json:"odname"`
				OdPwd     string `json:"odpwd"`
				Status    string `json:"status"`
				Nickname  string `json:"nickname"`
				Train     string `json:"train"`
				StudyDate string `json:"studydate"`
				Code      int    `json:"code"`
				Note      string `json:"note"`
				Charged   string `json:"charged"`
				AddDate   string `json:"adddate"`
			} `json:"list"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("yyy查询进度解析失败")
	}
	if result.Code != 200 {
		return nil, fmt.Errorf("yyy查询进度失败: %s", result.Message)
	}

	var items []model.SupplierProgressItem
	for _, o := range result.Data.List {
		// 映射 yyy 状态码
		status := o.Status
		process := ""
		switch o.Code {
		case 102:
			status = "已完成"
			process = "100%"
		case 104:
			status = "进行中"
		case 101:
			status = "待处理"
		case 103:
			status = "异常"
		}

		items = append(items, model.SupplierProgressItem{
			YID:        fmt.Sprintf("%d", o.ID),
			User:       o.OdName,
			KCName:     o.Train,
			Status:     status,
			StatusText: status,
			Process:    process,
			Remarks:    o.Note,
		})
	}

	return items, nil
}
