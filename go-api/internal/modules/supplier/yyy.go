package supplier

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

var (
	yyySessions   = map[string]*yyySession{}
	yyySessionsMu sync.Mutex
)

func yyyBaseURL(rawURL string) string {
	baseURL := strings.TrimRight(rawURL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	return baseURL
}

func getYYYSession(baseURL, username, password string) *yyySession {
	yyySessionsMu.Lock()
	defer yyySessionsMu.Unlock()

	key := baseURL + "|" + username
	if sess, ok := yyySessions[key]; ok {
		sess.password = password
		return sess
	}

	jar, _ := cookiejar.New(nil)
	sess := &yyySession{
		client:   &http.Client{Jar: jar, Timeout: 60 * time.Second},
		baseURL:  strings.TrimRight(baseURL, "/"),
		username: username,
		password: password,
	}
	yyySessions[key] = sess
	return sess
}

func (s *yyySession) login() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.loggedIn && time.Since(s.lastLogin) < 30*time.Minute {
		return nil
	}

	body, _ := json.Marshal(map[string]string{
		"username": s.username,
		"password": s.password,
	})

	req, _ := http.NewRequest("POST", s.baseURL+"/api/login", bytes.NewReader(body))
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
	return nil
}

func (s *yyySession) doRequest(method, path string, body interface{}) ([]byte, error) {
	if err := s.login(); err != nil {
		return nil, err
	}

	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}

	requestOnce := func() ([]byte, error) {
		var bodyReader io.Reader
		if len(bodyBytes) > 0 {
			bodyReader = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequest(method, s.baseURL+path, bodyReader)
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
		return io.ReadAll(resp.Body)
	}

	respBody, err := requestOnce()
	if err != nil {
		return nil, err
	}

	var check struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(respBody, &check)
	if check.Code == 302 && strings.Contains(check.Message, "重新登录") {
		s.mu.Lock()
		s.loggedIn = false
		s.mu.Unlock()
		if err := s.login(); err != nil {
			return nil, err
		}
		return requestOnce()
	}

	return respBody, nil
}

func extractYYYMoney(raw map[string]interface{}) string {
	if data, ok := raw["data"].(map[string]interface{}); ok {
		if money, ok := data["money"]; ok {
			return fmt.Sprintf("%v", money)
		}
	}
	if money, ok := raw["money"]; ok {
		return fmt.Sprintf("%v", money)
	}
	return ""
}

func yyyGetClasses(sup *model.SupplierFull) ([]SupplierClassItem, error) {
	session := getYYYSession(yyyBaseURL(sup.URL), sup.User, sup.Pass)
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
				PriceUnit string  `json:"price_unit"`
				Price     float64 `json:"price"`
			} `json:"list"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("yyy解析商品列表失败: %s", string(respBody))
	}
	if result.Code != 200 {
		return nil, fmt.Errorf("yyy获取商品失败: %s", result.Message)
	}

	classes := make([]SupplierClassItem, 0, len(result.Data.List))
	for _, item := range result.Data.List {
		price := item.Price
		if price == 0 && item.PriceUnit != "" {
			parts := strings.Fields(item.PriceUnit)
			if len(parts) > 0 {
				fmt.Sscanf(parts[0], "%f", &price)
			}
		}
		classes = append(classes, SupplierClassItem{
			CID:          fmt.Sprintf("%d", item.ID),
			Name:         item.Name,
			Price:        price,
			Content:      item.Trans,
			CategoryName: sup.Name,
		})
	}
	return classes, nil
}

func yyyQueryBalance(sup *model.SupplierFull) (map[string]interface{}, error) {
	session := getYYYSession(yyyBaseURL(sup.URL), sup.User, sup.Pass)
	respBody, err := session.doRequest("POST", "/api/money", nil)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, fmt.Errorf("yyy余额解析失败: %s", string(respBody))
	}

	return map[string]interface{}{
		"code":  200,
		"money": extractYYYMoney(raw),
		"pt":    sup.PT,
		"name":  sup.Name,
		"hid":   sup.HID,
		"raw":   raw,
	}, nil
}

func yyyCallOrder(sup *model.SupplierFull, user, pass, kcname string, siteID string) (*model.SupplierOrderResult, error) {
	session := getYYYSession(yyyBaseURL(sup.URL), sup.User, sup.Pass)

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
		return nil, fmt.Errorf("yyy下单解析失败: %s", string(respBody))
	}

	orderResult := &model.SupplierOrderResult{Msg: result.Message}
	if result.Code == 200 {
		orderResult.Code = 1
		orderResult.Msg = "下单成功"
		if dataArr, ok := result.Data.([]interface{}); ok && len(dataArr) > 0 {
			orderResult.YID = fmt.Sprintf("%v", dataArr[0])
		}
	} else {
		orderResult.Code = -1
	}

	return orderResult, nil
}

func yyyQueryProgress(sup *model.SupplierFull, username string) ([]model.SupplierProgressItem, error) {
	session := getYYYSession(yyyBaseURL(sup.URL), sup.User, sup.Pass)
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
				ID     int    `json:"id"`
				OdName string `json:"odname"`
				Status string `json:"status"`
				Train  string `json:"train"`
				Code   int    `json:"code"`
				Note   string `json:"note"`
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

	items := make([]model.SupplierProgressItem, 0, len(result.Data.List))
	for _, o := range result.Data.List {
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
