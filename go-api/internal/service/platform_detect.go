package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// DetectRequest 自动检测请求参数
type DetectRequest struct {
	URL    string `json:"url" binding:"required"`
	UID    string `json:"uid"`
	Key    string `json:"key"`
	Token  string `json:"token"`
	Cookie string `json:"cookie"`
}

// DetectResult 自动检测结果
type DetectResult struct {
	Success       bool              `json:"success"`
	AuthType      string            `json:"auth_type"`       // 检测到的认证方式
	SuccessCode   string            `json:"success_code"`    // 检测到的成功码
	APIStyle      string            `json:"api_style"`       // standard / rest
	BalanceOK     bool              `json:"balance_ok"`      // 余额接口是否可用
	BalanceMoney  string            `json:"balance_money"`   // 检测到的余额
	BalanceAct    string            `json:"balance_act"`     // 余额 act
	BalancePath   string            `json:"balance_path"`    // 余额路径
	BalanceField  string            `json:"balance_field"`   // 余额字段路径
	QueryOK       bool              `json:"query_ok"`        // 查课接口是否可用
	QueryAct      string            `json:"query_act"`       // 查课 act
	ClassListOK   bool              `json:"class_list_ok"`   // 课程列表接口是否可用
	CategoryOK    bool              `json:"category_ok"`     // 分类接口是否可用
	UseJSON       bool              `json:"use_json"`        // 是否使用 JSON
	ReturnsYID    bool              `json:"returns_yid"`     // 下单是否返回 YID
	Probes        []ProbeDetail     `json:"probes"`          // 所有探测详情
	SuggestedName string            `json:"suggested_name"`  // 建议平台名
	Config        map[string]string `json:"config"`          // 建议的完整配置
}

// ProbeDetail 单次探测详情
type ProbeDetail struct {
	Endpoint string `json:"endpoint"`
	Method   string `json:"method"`
	Status   string `json:"status"`  // ok / fail / timeout / error
	Code     string `json:"code"`    // 响应中的 code 值
	Msg      string `json:"msg"`     // 简要说明
	RawBody  string `json:"raw_body"` // 响应体前200字符
}

// DetectPlatform 自动检测上游平台API能力
func DetectPlatform(req DetectRequest) *DetectResult {
	result := &DetectResult{
		Probes: []ProbeDetail{},
		Config: map[string]string{},
	}

	baseURL := strings.TrimRight(req.URL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// =============================================
	// 阶段1：检测认证方式 + 余额接口（最安全的只读接口）
	// =============================================
	type balanceProbe struct {
		name      string
		url       string
		method    string
		authType  string
		act       string
		path      string
		useJSON   bool
		buildReq  func() (*http.Request, error)
	}

	var probes []balanceProbe

	// 1a. 标准 uid+key 模式
	if req.UID != "" && req.Key != "" {
		for _, act := range []string{"getmoney", "money"} {
			actCopy := act
			probes = append(probes, balanceProbe{
				name: fmt.Sprintf("uid+key act=%s", actCopy),
				url:  fmt.Sprintf("%s/api.php?act=%s", baseURL, actCopy),
				method: "POST", authType: "uid_key", act: actCopy,
				buildReq: func() (*http.Request, error) {
					data := url.Values{"uid": {req.UID}, "key": {req.Key}}
					return http.NewRequest("POST", fmt.Sprintf("%s/api.php?act=%s", baseURL, actCopy), strings.NewReader(data.Encode()))
				},
			})
		}
	}

	// 1b. token 模式 (pass字段作为token)
	if req.Key != "" {
		probes = append(probes, balanceProbe{
			name: "token /api/getinfo", url: baseURL + "/api/getinfo",
			method: "POST", authType: "token_only", path: "/api/getinfo", useJSON: true,
			buildReq: func() (*http.Request, error) {
				jsonData, _ := json.Marshal(map[string]string{"token": req.Key})
				r, _ := http.NewRequest("POST", baseURL+"/api/getinfo", strings.NewReader(string(jsonData)))
				r.Header.Set("Content-Type", "application/json")
				return r, nil
			},
		})
		// token form 模式
		probes = append(probes, balanceProbe{
			name: "token form /api/getinfo", url: baseURL + "/api/getinfo",
			method: "POST", authType: "token_only", path: "/api/getinfo",
			buildReq: func() (*http.Request, error) {
				data := url.Values{"token": {req.Key}}
				return http.NewRequest("POST", baseURL+"/api/getinfo", strings.NewReader(data.Encode()))
			},
		})
	}

	// 1c. Bearer token 模式
	if req.Token != "" {
		probes = append(probes, balanceProbe{
			name: "Bearer /api/getuserinfo/", url: baseURL + "/api/getuserinfo/",
			method: "GET", authType: "bearer_token", path: "/api/getuserinfo/",
			buildReq: func() (*http.Request, error) {
				r, _ := http.NewRequest("GET", baseURL+"/api/getuserinfo/", nil)
				r.Header.Set("Authorization", "Bearer "+req.Token)
				r.Header.Set("Content-Type", "application/json")
				if req.Cookie != "" {
					cookieStr := req.Cookie
					if !strings.Contains(cookieStr, "=") {
						cookieStr = "session_id=" + cookieStr
					}
					r.Header.Set("Cookie", cookieStr)
				}
				return r, nil
			},
		})
	}

	// 并发探测余额接口
	type probeResult struct {
		idx     int
		body    []byte
		err     error
		status  int
	}

	var wg sync.WaitGroup
	probeResults := make([]probeResult, len(probes))

	for i, p := range probes {
		wg.Add(1)
		go func(idx int, probe balanceProbe) {
			defer wg.Done()
			httpReq, err := probe.buildReq()
			if err != nil {
				probeResults[idx] = probeResult{idx: idx, err: err}
				return
			}
			if httpReq.Header.Get("Content-Type") == "" {
				httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
			resp, err := client.Do(httpReq)
			if err != nil {
				probeResults[idx] = probeResult{idx: idx, err: err}
				return
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			probeResults[idx] = probeResult{idx: idx, body: body, status: resp.StatusCode}
		}(i, p)
	}
	wg.Wait()

	// 分析余额探测结果
	for i, pr := range probeResults {
		probe := probes[i]
		detail := ProbeDetail{
			Endpoint: probe.name,
			Method:   probe.method,
		}

		if pr.err != nil {
			detail.Status = "error"
			detail.Msg = pr.err.Error()
			result.Probes = append(result.Probes, detail)
			continue
		}

		bodyStr := string(pr.body)
		if len(bodyStr) > 200 {
			detail.RawBody = bodyStr[:200]
		} else {
			detail.RawBody = bodyStr
		}

		// 尝试解析 JSON
		var raw map[string]interface{}
		if err := json.Unmarshal(pr.body, &raw); err != nil {
			detail.Status = "fail"
			detail.Msg = "非JSON响应"
			result.Probes = append(result.Probes, detail)
			continue
		}

		codeVal := ""
		if c, ok := raw["code"]; ok {
			codeVal = fmt.Sprintf("%v", c)
		}
		detail.Code = codeVal

		// 判断是否成功: code 为 0, 1, 或 200
		isSuccess := codeVal == "0" || codeVal == "1" || codeVal == "200"
		if isSuccess {
			detail.Status = "ok"
			detail.Msg = "成功"

			// 第一个成功的余额探测就采用
			if !result.BalanceOK {
				result.BalanceOK = true
				result.Success = true
				result.AuthType = probe.authType
				result.SuccessCode = codeVal
				result.BalanceAct = probe.act
				result.BalancePath = probe.path
				result.UseJSON = probe.useJSON

				if probe.path != "" {
					result.APIStyle = "rest"
				} else {
					result.APIStyle = "standard"
				}

				// 尝试提取余额值
				money := tryExtractMoney(raw)
				result.BalanceMoney = money.value
				result.BalanceField = money.field
			}
		} else {
			detail.Status = "fail"
			if msg, ok := raw["msg"]; ok {
				detail.Msg = fmt.Sprintf("code=%s, msg=%v", codeVal, msg)
			} else {
				detail.Msg = fmt.Sprintf("code=%s", codeVal)
			}
		}

		result.Probes = append(result.Probes, detail)
	}

	// =============================================
	// 阶段2：检测查课/课程列表接口（只在认证成功后探测）
	// =============================================
	if result.Success && result.AuthType == "uid_key" {
		// 探测 getclass
		classURL := fmt.Sprintf("%s/api.php?act=getclass", baseURL)
		data := url.Values{"uid": {req.UID}, "key": {req.Key}}
		classDetail := probeEndpoint(client, "getclass 课程列表", classURL, data)
		result.Probes = append(result.Probes, classDetail)
		if classDetail.Status == "ok" {
			result.ClassListOK = true
		}

		// 探测 getcate
		cateURL := fmt.Sprintf("%s/api.php?act=getcate", baseURL)
		cateDetail := probeEndpoint(client, "getcate 分类列表", cateURL, data)
		result.Probes = append(result.Probes, cateDetail)
		if cateDetail.Status == "ok" {
			result.CategoryOK = true
		}

		// 探测 get (查课 - 不传学校信息，看是否返回有效结构)
		getURL := fmt.Sprintf("%s/api.php?act=get", baseURL)
		getDetail := probeEndpoint(client, "get 查课接口", getURL, data)
		result.Probes = append(result.Probes, getDetail)
		// 即使 code 非成功（因为没传完整参数），只要返回了 JSON 就说明接口存在
		if getDetail.Status == "ok" || (getDetail.Code != "" && getDetail.Status == "fail") {
			result.QueryOK = true
			result.QueryAct = "get"
		}

		// 探测 add (下单接口存在性 - 不真正下单)
		// 我们只检查接口是否返回 JSON 错误（说明接口存在）
		addURL := fmt.Sprintf("%s/api.php?act=add", baseURL)
		addDetail := probeEndpoint(client, "add 下单接口", addURL, data)
		result.Probes = append(result.Probes, addDetail)
		// 如果返回了JSON且有msg，说明接口存在
		if addDetail.Code != "" {
			// 检查响应里是否有 id 字段的结构暗示
			result.Config["order_act"] = "add"
		}

		// 探测其他常见接口
		for _, act := range []string{"chadan2", "chadan", "xq", "zt", "gaimi"} {
			actURL := fmt.Sprintf("%s/api.php?act=%s", baseURL, act)
			actDetail := probeEndpoint(client, fmt.Sprintf("%s 接口", act), actURL, data)
			result.Probes = append(result.Probes, actDetail)
			if actDetail.Code != "" {
				// 接口存在（返回了JSON）
				switch act {
				case "chadan2":
					result.Config["progress_act"] = "chadan2"
				case "chadan":
					result.Config["progress_no_yid"] = "chadan"
				case "xq":
					result.Config["log_act"] = "xq"
				case "zt":
					result.Config["pause_act"] = "zt"
				case "gaimi":
					result.Config["change_pass_act"] = "gaimi"
				}
			}
		}

		// 探测 REST 风格接口
		for _, path := range []string{"/api/search", "/api/chadan1", "/api/stop", "/api/update", "/api/reset", "/log/", "/api/submitWork", "/api/queryWork"} {
			restURL := baseURL + path
			restDetail := probeEndpointGET(client, "REST "+path, restURL, req.UID, req.Key)
			result.Probes = append(result.Probes, restDetail)
			if restDetail.Code != "" || restDetail.Status == "ok" {
				switch {
				case strings.Contains(path, "search") || strings.Contains(path, "chadan"):
					result.Config["progress_path"] = path
					result.Config["progress_method"] = "GET"
				case strings.Contains(path, "stop"):
					result.Config["pause_path"] = path
				case strings.Contains(path, "update"):
					result.Config["change_pass_path"] = path
				case strings.Contains(path, "reset"):
					result.Config["resubmit_path"] = path
				case strings.Contains(path, "log"):
					result.Config["log_path"] = path
					result.Config["log_method"] = "GET"
				case strings.Contains(path, "submitWork"):
					result.Config["report_path"] = path
					result.Config["report_param_style"] = "token"
				case strings.Contains(path, "queryWork"):
					result.Config["get_report_path"] = path
				}
			}
		}
	}

	// =============================================
	// 阶段2b：token_only 模式探测 REST 接口
	// =============================================
	if result.Success && result.AuthType == "token_only" {
		token := req.Key
		// 探测所有 REST 风格接口（POST JSON with token）
		type tokenProbe struct {
			path   string
			params map[string]string
		}
		tokenProbes := []tokenProbe{
			{"/api/get", map[string]string{"token": token, "platform": "1", "user": "test", "pass": "test"}},
			{"/api/add", map[string]string{"token": token}},
			{"/api/query", map[string]string{"token": token, "username": "test"}},
			{"/api/refresh", map[string]string{"token": token, "id": "0", "username": "test"}},
			{"/api/reset", map[string]string{"token": token, "id": "0", "username": "test"}},
			{"/api/stop", map[string]string{"token": token, "id": "0", "username": "test"}},
			{"/api/update", map[string]string{"token": token, "id": "0", "username": "test"}},
			{"/api/getcate", map[string]string{"token": token}},
			{"/api/getclass", map[string]string{"token": token}},
			{"/api/submitWork", map[string]string{"token": token}},
			{"/api/queryWork", map[string]string{"token": token}},
		}
		for _, tp := range tokenProbes {
			detail := probeEndpointJSON(client, "REST "+tp.path, baseURL+tp.path, tp.params)
			result.Probes = append(result.Probes, detail)
			exists := detail.Code != "" || detail.Status == "ok"
			if !exists {
				continue
			}
			switch {
			case tp.path == "/api/get":
				result.QueryOK = true
				result.Config["query_path"] = tp.path
			case tp.path == "/api/add":
				result.Config["order_path"] = tp.path
			case tp.path == "/api/query":
				result.Config["progress_path"] = tp.path
				result.Config["progress_method"] = "POST"
			case tp.path == "/api/refresh":
				result.Config["refresh_path"] = tp.path
			case tp.path == "/api/reset":
				result.Config["resubmit_path"] = tp.path
			case tp.path == "/api/stop":
				result.Config["pause_path"] = tp.path
			case tp.path == "/api/update":
				result.Config["change_pass_path"] = tp.path
			case tp.path == "/api/getcate":
				result.CategoryOK = true
			case tp.path == "/api/getclass":
				result.ClassListOK = true
			case tp.path == "/api/submitWork":
				result.Config["report_path"] = tp.path
				result.Config["report_param_style"] = "token"
			case tp.path == "/api/queryWork":
				result.Config["get_report_path"] = tp.path
			}
		}
	}

	// =============================================
	// 阶段3：生成建议配置
	// =============================================
	if result.Success {
		result.Config["auth_type"] = result.AuthType
		result.Config["success_codes"] = result.SuccessCode
		result.Config["api_path_style"] = result.APIStyle
		if result.UseJSON {
			result.Config["use_json"] = "true"
		}
		if result.BalanceAct != "" {
			result.Config["balance_act"] = result.BalanceAct
		}
		if result.BalancePath != "" {
			result.Config["balance_path"] = result.BalancePath
		}
		if result.BalanceField != "" {
			result.Config["balance_money_field"] = result.BalanceField
		}
		if result.QueryAct != "" {
			result.Config["query_act"] = result.QueryAct
		}

		// 建议平台名：从URL提取域名
		if u, err := url.Parse(baseURL); err == nil {
			result.SuggestedName = u.Hostname()
		}
	}

	return result
}

// moneyExtract 余额提取结果
type moneyExtract struct {
	value string
	field string
}

// tryExtractMoney 尝试从 JSON 响应中提取余额值
func tryExtractMoney(raw map[string]interface{}) moneyExtract {
	// 尝试路径: money -> data.money -> data.remainscore -> data (如果是数字)
	if m, ok := raw["money"]; ok {
		return moneyExtract{value: fmt.Sprintf("%v", m), field: "money"}
	}
	if data, ok := raw["data"]; ok {
		if dataMap, ok := data.(map[string]interface{}); ok {
			if m, ok := dataMap["money"]; ok {
				return moneyExtract{value: fmt.Sprintf("%v", m), field: "data.money"}
			}
			if m, ok := dataMap["remainscore"]; ok {
				return moneyExtract{value: fmt.Sprintf("%v", m), field: "data.remainscore"}
			}
			if m, ok := dataMap["balance"]; ok {
				return moneyExtract{value: fmt.Sprintf("%v", m), field: "data.balance"}
			}
		}
		// data 本身可能就是余额数值
		switch v := data.(type) {
		case float64:
			return moneyExtract{value: fmt.Sprintf("%.2f", v), field: "data"}
		case string:
			return moneyExtract{value: v, field: "data"}
		}
	}
	return moneyExtract{value: "", field: "money"}
}

// probeEndpoint 探测一个标准 POST 端点
func probeEndpoint(client *http.Client, name, apiURL string, data url.Values) ProbeDetail {
	detail := ProbeDetail{Endpoint: name, Method: "POST"}

	resp, err := client.PostForm(apiURL, data)
	if err != nil {
		detail.Status = "error"
		detail.Msg = err.Error()
		return detail
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)
	if len(bodyStr) > 200 {
		detail.RawBody = bodyStr[:200]
	} else {
		detail.RawBody = bodyStr
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		detail.Status = "fail"
		detail.Msg = "非JSON响应"
		return detail
	}

	if c, ok := raw["code"]; ok {
		detail.Code = fmt.Sprintf("%v", c)
	}

	codeVal := detail.Code
	if codeVal == "0" || codeVal == "1" || codeVal == "200" {
		detail.Status = "ok"
		detail.Msg = "成功"
	} else {
		detail.Status = "fail"
		if msg, ok := raw["msg"]; ok {
			detail.Msg = fmt.Sprintf("%v", msg)
		} else {
			detail.Msg = fmt.Sprintf("code=%s", codeVal)
		}
	}

	return detail
}


// probeEndpointJSON 探测一个 JSON POST 端点
func probeEndpointJSON(client *http.Client, name, apiURL string, params map[string]string) ProbeDetail {
	detail := ProbeDetail{Endpoint: name, Method: "POST"}
	jsonData, _ := json.Marshal(params)
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonData)))
	if err != nil {
		detail.Status = "error"
		detail.Msg = err.Error()
		return detail
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		detail.Status = "error"
		detail.Msg = err.Error()
		return detail
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)
	if len(bodyStr) > 200 {
		detail.RawBody = bodyStr[:200]
	} else {
		detail.RawBody = bodyStr
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		detail.Status = "fail"
		detail.Msg = "非JSON响应"
		return detail
	}
	if c, ok := raw["code"]; ok {
		detail.Code = fmt.Sprintf("%v", c)
	}
	codeVal := detail.Code
	if codeVal == "0" || codeVal == "1" || codeVal == "200" {
		detail.Status = "ok"
		detail.Msg = "成功"
	} else {
		detail.Status = "fail"
		if msg, ok := raw["msg"]; ok {
			detail.Msg = fmt.Sprintf("%v", msg)
		} else {
			detail.Msg = fmt.Sprintf("code=%s", codeVal)
		}
	}
	return detail
}
// probeEndpointGET 探测一个 GET 端点
func probeEndpointGET(client *http.Client, name, apiURL, uid, key string) ProbeDetail {
	detail := ProbeDetail{Endpoint: name, Method: "GET"}

	// 尝试带 uid+key 参数
	u, _ := url.Parse(apiURL)
	q := u.Query()
	if uid != "" {
		q.Set("uid", uid)
	}
	if key != "" {
		q.Set("key", key)
	}
	u.RawQuery = q.Encode()

	resp, err := client.Get(u.String())
	if err != nil {
		detail.Status = "error"
		detail.Msg = err.Error()
		return detail
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)
	if len(bodyStr) > 200 {
		detail.RawBody = bodyStr[:200]
	} else {
		detail.RawBody = bodyStr
	}

	// 404 等明确的 HTTP 错误
	if resp.StatusCode == 404 {
		detail.Status = "fail"
		detail.Msg = "404 Not Found"
		return detail
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		detail.Status = "fail"
		detail.Msg = "非JSON响应"
		return detail
	}

	if c, ok := raw["code"]; ok {
		detail.Code = fmt.Sprintf("%v", c)
	}

	codeVal := detail.Code
	if codeVal == "0" || codeVal == "1" || codeVal == "200" {
		detail.Status = "ok"
		detail.Msg = "成功"
	} else {
		detail.Status = "fail"
		if msg, ok := raw["msg"]; ok {
			detail.Msg = fmt.Sprintf("%v", msg)
		} else {
			detail.Msg = fmt.Sprintf("code=%s", codeVal)
		}
	}

	return detail
}
