package platformtools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"go-api/internal/model"
)

type DetectRequest struct {
	URL    string `json:"url" binding:"required"`
	UID    string `json:"uid"`
	Key    string `json:"key"`
	Token  string `json:"token"`
	Cookie string `json:"cookie"`
}

type ProbeDetail struct {
	Endpoint string `json:"endpoint"`
	Method   string `json:"method"`
	Status   string `json:"status"`
	Code     string `json:"code"`
	Msg      string `json:"msg"`
	RawBody  string `json:"raw_body"`
}

type DetectResult struct {
	Success       bool              `json:"success"`
	AuthType      string            `json:"auth_type"`
	SuccessCode   string            `json:"success_code"`
	APIStyle      string            `json:"api_style"`
	BalanceOK     bool              `json:"balance_ok"`
	BalanceMoney  string            `json:"balance_money"`
	BalanceAct    string            `json:"balance_act"`
	BalancePath   string            `json:"balance_path"`
	BalanceField  string            `json:"balance_field"`
	QueryOK       bool              `json:"query_ok"`
	QueryAct      string            `json:"query_act"`
	ClassListOK   bool              `json:"class_list_ok"`
	CategoryOK    bool              `json:"category_ok"`
	UseJSON       bool              `json:"use_json"`
	ReturnsYID    bool              `json:"returns_yid"`
	Probes        []ProbeDetail     `json:"probes"`
	SuggestedName string            `json:"suggested_name"`
	Config        map[string]string `json:"config"`
}

type ParsedPHPConfig struct {
	PT              string   `json:"pt"`
	Name            string   `json:"name"`
	AuthType        string   `json:"auth_type"`
	APIPathStyle    string   `json:"api_path_style"`
	SuccessCodes    string   `json:"success_codes"`
	UseJSON         bool     `json:"use_json"`
	QueryAct        string   `json:"query_act"`
	QueryPath       string   `json:"query_path"`
	OrderAct        string   `json:"order_act"`
	OrderPath       string   `json:"order_path"`
	ProgressAct     string   `json:"progress_act"`
	ProgressPath    string   `json:"progress_path"`
	ProgressMethod  string   `json:"progress_method"`
	PauseAct        string   `json:"pause_act"`
	PausePath       string   `json:"pause_path"`
	ChangePassAct   string   `json:"change_pass_act"`
	ChangePassPath  string   `json:"change_pass_path"`
	ChangePassParam string   `json:"change_pass_param"`
	ChangePassID    string   `json:"change_pass_id_param"`
	LogAct          string   `json:"log_act"`
	LogPath         string   `json:"log_path"`
	LogIDParam      string   `json:"log_id_param"`
	ReturnsYID      bool     `json:"returns_yid"`
	Confidence      int      `json:"confidence"`
	Warnings        []string `json:"warnings"`
}

type moneyExtract struct {
	value string
	field string
}

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

	type balanceProbe struct {
		name     string
		url      string
		method   string
		authType string
		act      string
		path     string
		useJSON  bool
		buildReq func() (*http.Request, error)
	}

	var probes []balanceProbe
	if req.UID != "" && req.Key != "" {
		for _, act := range []string{"getmoney", "money"} {
			actCopy := act
			probes = append(probes, balanceProbe{
				name:   fmt.Sprintf("uid+key act=%s", actCopy),
				url:    fmt.Sprintf("%s/api.php?act=%s", baseURL, actCopy),
				method: "POST", authType: "uid_key", act: actCopy,
				buildReq: func() (*http.Request, error) {
					data := url.Values{"uid": {req.UID}, "key": {req.Key}}
					return http.NewRequest("POST", fmt.Sprintf("%s/api.php?act=%s", baseURL, actCopy), strings.NewReader(data.Encode()))
				},
			})
		}
	}
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
		probes = append(probes, balanceProbe{
			name: "token form /api/getinfo", url: baseURL + "/api/getinfo",
			method: "POST", authType: "token_only", path: "/api/getinfo",
			buildReq: func() (*http.Request, error) {
				data := url.Values{"token": {req.Key}}
				return http.NewRequest("POST", baseURL+"/api/getinfo", strings.NewReader(data.Encode()))
			},
		})
	}
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

	type probeResult struct {
		body []byte
		err  error
	}
	var wg sync.WaitGroup
	results := make([]probeResult, len(probes))
	for i, probe := range probes {
		wg.Add(1)
		go func(idx int, item balanceProbe) {
			defer wg.Done()
			httpReq, err := item.buildReq()
			if err != nil {
				results[idx] = probeResult{err: err}
				return
			}
			if httpReq.Header.Get("Content-Type") == "" {
				httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
			resp, err := client.Do(httpReq)
			if err != nil {
				results[idx] = probeResult{err: err}
				return
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			results[idx] = probeResult{body: body}
		}(i, probe)
	}
	wg.Wait()

	for i, pr := range results {
		probe := probes[i]
		detail := ProbeDetail{Endpoint: probe.name, Method: probe.method}
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
		if codeVal == "0" || codeVal == "1" || codeVal == "200" {
			detail.Status = "ok"
			detail.Msg = "成功"
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

	if result.Success && result.AuthType == "uid_key" {
		data := url.Values{"uid": {req.UID}, "key": {req.Key}}
		classDetail := probeEndpoint(client, "getclass 课程列表", fmt.Sprintf("%s/api.php?act=getclass", baseURL), data)
		result.Probes = append(result.Probes, classDetail)
		if classDetail.Status == "ok" {
			result.ClassListOK = true
		}
		cateDetail := probeEndpoint(client, "getcate 分类列表", fmt.Sprintf("%s/api.php?act=getcate", baseURL), data)
		result.Probes = append(result.Probes, cateDetail)
		if cateDetail.Status == "ok" {
			result.CategoryOK = true
		}
		getDetail := probeEndpoint(client, "get 查课接口", fmt.Sprintf("%s/api.php?act=get", baseURL), data)
		result.Probes = append(result.Probes, getDetail)
		if getDetail.Status == "ok" || (getDetail.Code != "" && getDetail.Status == "fail") {
			result.QueryOK = true
			result.QueryAct = "get"
		}
		addDetail := probeEndpoint(client, "add 下单接口", fmt.Sprintf("%s/api.php?act=add", baseURL), data)
		result.Probes = append(result.Probes, addDetail)
		if addDetail.Code != "" {
			result.Config["order_act"] = "add"
		}
		for _, act := range []string{"chadan2", "chadan", "xq", "zt", "gaimi"} {
			actDetail := probeEndpoint(client, fmt.Sprintf("%s 接口", act), fmt.Sprintf("%s/api.php?act=%s", baseURL, act), data)
			result.Probes = append(result.Probes, actDetail)
			if actDetail.Code == "" {
				continue
			}
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
		for _, path := range []string{"/api/search", "/api/chadan1", "/api/stop", "/api/update", "/api/reset", "/log/", "/api/submitWork", "/api/queryWork"} {
			detail := probeEndpointGET(client, "REST "+path, baseURL+path, req.UID, req.Key)
			result.Probes = append(result.Probes, detail)
			if detail.Code == "" && detail.Status != "ok" {
				continue
			}
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

	if result.Success && result.AuthType == "token_only" {
		token := req.Key
		tokenProbes := []struct {
			path   string
			params map[string]string
		}{
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
			if detail.Code == "" && detail.Status != "ok" {
				continue
			}
			switch tp.path {
			case "/api/get":
				result.QueryOK = true
				result.Config["query_path"] = tp.path
			case "/api/add":
				result.Config["order_path"] = tp.path
			case "/api/query":
				result.Config["progress_path"] = tp.path
				result.Config["progress_method"] = "POST"
			case "/api/refresh":
				result.Config["refresh_path"] = tp.path
			case "/api/reset":
				result.Config["resubmit_path"] = tp.path
			case "/api/stop":
				result.Config["pause_path"] = tp.path
			case "/api/update":
				result.Config["change_pass_path"] = tp.path
			case "/api/getcate":
				result.CategoryOK = true
			case "/api/getclass":
				result.ClassListOK = true
			case "/api/submitWork":
				result.Config["report_path"] = tp.path
				result.Config["report_param_style"] = "token"
			case "/api/queryWork":
				result.Config["get_report_path"] = tp.path
			}
		}
	}

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
		if u, err := url.Parse(baseURL); err == nil {
			result.SuggestedName = u.Hostname()
		}
	}

	return result
}

func BuildConfigFromDetection(result *DetectResult, pt, name string) *model.PlatformConfigSaveRequest {
	if result == nil || !result.Success {
		return nil
	}
	cfg := result.Config
	req := &model.PlatformConfigSaveRequest{
		PT:           pt,
		Name:         name,
		AuthType:     getOr(cfg, "auth_type", "uid_key"),
		APIPathStyle: getOr(cfg, "api_path_style", "standard"),
		SuccessCodes: getOr(cfg, "success_codes", "0"),
		UseJSON:      cfg["use_json"] == "true",
		ReturnsYID:   result.ReturnsYID,
	}
	req.QueryAct = getOr(cfg, "query_act", "get")
	req.QueryPath = cfg["query_path"]
	req.OrderAct = getOr(cfg, "order_act", "add")
	req.OrderPath = cfg["order_path"]
	req.ProgressAct = getOr(cfg, "progress_act", "chadan2")
	req.ProgressNoYID = getOr(cfg, "progress_no_yid", "chadan")
	req.ProgressPath = cfg["progress_path"]
	req.ProgressMethod = getOr(cfg, "progress_method", "POST")
	req.PauseAct = getOr(cfg, "pause_act", "zt")
	req.PausePath = cfg["pause_path"]
	req.PauseIDParam = getOr(cfg, "pause_id_param", "id")
	req.ChangePassAct = getOr(cfg, "change_pass_act", "gaimi")
	req.ChangePassPath = cfg["change_pass_path"]
	req.ChangePassParam = getOr(cfg, "change_pass_param", "newPwd")
	req.ChangePassIDParam = getOr(cfg, "change_pass_id_param", "id")
	req.ResubmitPath = cfg["resubmit_path"]
	req.ResubmitIDParam = getOr(cfg, "resubmit_id_param", "id")
	req.LogAct = getOr(cfg, "log_act", "xq")
	req.LogPath = cfg["log_path"]
	req.LogMethod = getOr(cfg, "log_method", "POST")
	req.LogIDParam = getOr(cfg, "log_id_param", "id")
	req.BalanceAct = getOr(cfg, "balance_act", "getmoney")
	req.BalancePath = cfg["balance_path"]
	req.BalanceMoneyField = getOr(cfg, "balance_money_field", "money")
	req.BalanceMethod = getOr(cfg, "balance_method", "POST")
	if cfg["auth_type"] == "bearer_token" {
		req.BalanceAuthType = "bearer_token"
	}
	req.ReportPath = cfg["report_path"]
	req.GetReportPath = cfg["get_report_path"]
	req.ReportParamStyle = cfg["report_param_style"]
	req.ReportAuthType = cfg["report_auth_type"]
	req.RefreshPath = cfg["refresh_path"]
	if name == "" {
		req.Name = result.SuggestedName
	}
	return req
}

func ParsePHPCode(code string) *ParsedPHPConfig {
	cfg := &ParsedPHPConfig{
		AuthType:        "uid_key",
		APIPathStyle:    "standard",
		SuccessCodes:    "0",
		QueryAct:        "get",
		OrderAct:        "add",
		ProgressAct:     "chadan2",
		ProgressMethod:  "POST",
		PauseAct:        "zt",
		ChangePassAct:   "gaimi",
		ChangePassParam: "newPwd",
		ChangePassID:    "id",
		LogAct:          "xq",
		LogIDParam:      "id",
		Warnings:        []string{},
	}

	ptRe := regexp.MustCompile(`\$type\s*==\s*["']([^"']+)["']`)
	if m := ptRe.FindStringSubmatch(code); len(m) > 1 {
		cfg.PT = m[1]
		cfg.Name = m[1]
		cfg.Confidence += 20
	} else {
		cfg.Warnings = append(cfg.Warnings, "未找到 $type == \"xxx\" 平台标识")
	}

	actRe := regexp.MustCompile(`act=([a-zA-Z0-9_]+)`)
	actMatches := actRe.FindAllStringSubmatch(code, -1)
	if len(actMatches) > 0 {
		act := actMatches[0][1]
		cfg.Confidence += 15
		switch {
		case act == "get":
			cfg.QueryAct = "get"
		case act == "add" || strings.Contains(act, "add"):
			cfg.OrderAct = act
		case strings.Contains(act, "chadan"):
			cfg.ProgressAct = act
		case act == "zt" || act == "zanting" || act == "stop":
			cfg.PauseAct = act
		case act == "gaimi" || act == "xgmm" || strings.Contains(act, "update"):
			cfg.ChangePassAct = act
		case act == "xq" || strings.Contains(act, "log"):
			cfg.LogAct = act
		default:
			if containsAny(code, "school", "user", "pass", "platform", "查课", "get") {
				cfg.QueryAct = act
			} else if containsAny(code, "kcname", "kcid", "下单") {
				cfg.OrderAct = act
			}
		}
	}

	restPathRe := regexp.MustCompile(`["'](/api/[a-zA-Z0-9_/\-]+)["']`)
	if m := restPathRe.FindStringSubmatch(code); len(m) > 1 {
		path := m[1]
		cfg.APIPathStyle = "rest"
		cfg.Confidence += 10
		pathLower := strings.ToLower(path)
		switch {
		case strings.Contains(pathLower, "query") || strings.Contains(pathLower, "course") || strings.Contains(pathLower, "login") || strings.Contains(pathLower, "get"):
			cfg.QueryPath = path
		case strings.Contains(pathLower, "add") || strings.Contains(pathLower, "create") || strings.Contains(pathLower, "order"):
			cfg.OrderPath = path
		case strings.Contains(pathLower, "search") || strings.Contains(pathLower, "chadan") || strings.Contains(pathLower, "progress"):
			cfg.ProgressPath = path
		case strings.Contains(pathLower, "freeze") || strings.Contains(pathLower, "stop") || strings.Contains(pathLower, "pause"):
			cfg.PausePath = path
		case strings.Contains(pathLower, "update") || strings.Contains(pathLower, "password") || strings.Contains(pathLower, "change"):
			cfg.ChangePassPath = path
		case strings.Contains(pathLower, "log"):
			cfg.LogPath = path
		}
	}

	if strings.Contains(code, `$a["token"]`) || strings.Contains(code, `$token`) || strings.Contains(code, `"token" =>`) {
		if !strings.Contains(code, `$a["user"]`) && !strings.Contains(code, `"uid"`) {
			cfg.AuthType = "token_field"
			cfg.Confidence += 10
		}
	}
	if strings.Contains(code, `"token" => $a["pass"]`) {
		cfg.AuthType = "token_only"
		cfg.Confidence += 10
	}

	successRe := regexp.MustCompile(`\$result\['code'\]\s*==\s*["']?(\d+)["']?`)
	successMatches := successRe.FindAllStringSubmatch(code, -1)
	if len(successMatches) > 0 {
		codes := map[string]bool{}
		for _, m := range successMatches {
			codes[m[1]] = true
		}
		codeList := make([]string, 0, len(codes))
		for c := range codes {
			codeList = append(codeList, c)
		}
		cfg.SuccessCodes = strings.Join(codeList, ",")
		cfg.Confidence += 10
	}

	multiCodeRe := regexp.MustCompile(`\$result\['code'\]\s*==\s*(\d+)\s*\|\|\s*\$result\['code'\]\s*==\s*(\d+)(?:\s*\|\|\s*\$result\['code'\]\s*==\s*(\d+))?`)
	if m := multiCodeRe.FindStringSubmatch(code); len(m) > 1 {
		codes := map[string]bool{}
		for _, c := range m[1:] {
			if c != "" {
				codes[c] = true
			}
		}
		codeList := make([]string, 0, len(codes))
		for c := range codes {
			codeList = append(codeList, c)
		}
		cfg.SuccessCodes = strings.Join(codeList, ",")
	}

	if strings.Contains(code, "json_encode") && strings.Contains(code, "CURLOPT_POSTFIELDS") {
		cfg.UseJSON = true
		cfg.Confidence += 5
	}
	if strings.Contains(code, "httpRequest") && strings.Contains(code, "true)") {
		cfg.UseJSON = true
		cfg.Confidence += 5
	}
	if strings.Contains(code, "application/json") {
		cfg.UseJSON = true
		cfg.Confidence += 5
	}

	passParamRe := regexp.MustCompile(`["'](newPwd|xgmm|newpwd|new_password|odpwd)["']\s*=>`)
	if m := passParamRe.FindStringSubmatch(code); len(m) > 1 {
		cfg.ChangePassParam = m[1]
		cfg.Confidence += 5
	}

	idParamRe := regexp.MustCompile(`["'](oid|id|ids)["']\s*=>\s*\$(?:d\["yid"\]|yid|order)`)
	if m := idParamRe.FindStringSubmatch(code); len(m) > 1 {
		cfg.ChangePassID = m[1]
		cfg.Confidence += 5
	}

	if strings.Contains(code, `$result['id']`) || strings.Contains(code, `$result['yid']`) || strings.Contains(code, `"yid"`) {
		cfg.ReturnsYID = true
		cfg.Confidence += 5
	}
	if strings.Contains(code, "getProxy") || strings.Contains(code, "CURLPROXY_SOCKS5") || strings.Contains(code, "proxy") {
		cfg.Warnings = append(cfg.Warnings, "检测到代理逻辑，需要手动启用 need_proxy")
	}
	if strings.Contains(code, "CURLOPT_CUSTOMREQUEST") && strings.Contains(code, "GET") {
		cfg.ProgressMethod = "GET"
	}
	if strings.Contains(code, "file_get_contents") {
		cfg.ProgressMethod = "GET"
	}

	pauseRe := regexp.MustCompile(`act=([a-zA-Z_]+).*(?:暂停|pause|stop|freeze)`)
	if m := pauseRe.FindStringSubmatch(code); len(m) > 1 {
		cfg.PauseAct = m[1]
	}
	if strings.Contains(code, "暂停") || strings.Contains(code, "pause") || strings.Contains(code, "freeze") {
		pauseActRe := regexp.MustCompile(`act=([a-zA-Z_]+)`)
		if m := pauseActRe.FindStringSubmatch(code); len(m) > 1 {
			cfg.PauseAct = m[1]
		}
	}

	if cfg.PT != "" {
		cfg.Confidence += 10
	}
	if cfg.Confidence > 100 {
		cfg.Confidence = 100
	}
	if cfg.Confidence < 30 {
		cfg.Warnings = append(cfg.Warnings, "解析置信度较低，建议手动核对配置")
	}
	return cfg
}

func getOr(m map[string]string, key, def string) string {
	if v, ok := m[key]; ok && v != "" {
		return v
	}
	return def
}

func tryExtractMoney(raw map[string]interface{}) moneyExtract {
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
		switch v := data.(type) {
		case float64:
			return moneyExtract{value: fmt.Sprintf("%.2f", v), field: "data"}
		case string:
			return moneyExtract{value: v, field: "data"}
		}
	}
	return moneyExtract{value: "", field: "money"}
}

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
	return parseProbeDetail(detail, body, false, 0)
}

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
	return parseProbeDetail(detail, body, false, 0)
}

func probeEndpointGET(client *http.Client, name, apiURL, uid, key string) ProbeDetail {
	detail := ProbeDetail{Endpoint: name, Method: "GET"}
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
	return parseProbeDetail(detail, body, true, resp.StatusCode)
}

func parseProbeDetail(detail ProbeDetail, body []byte, respect404 bool, statusCode int) ProbeDetail {
	bodyStr := string(body)
	if len(bodyStr) > 200 {
		detail.RawBody = bodyStr[:200]
	} else {
		detail.RawBody = bodyStr
	}
	if respect404 && statusCode == 404 {
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
	if detail.Code == "0" || detail.Code == "1" || detail.Code == "200" {
		detail.Status = "ok"
		detail.Msg = "成功"
	} else {
		detail.Status = "fail"
		if msg, ok := raw["msg"]; ok {
			detail.Msg = fmt.Sprintf("%v", msg)
		} else {
			detail.Msg = fmt.Sprintf("code=%s", detail.Code)
		}
	}
	return detail
}

func containsAny(s string, substrs ...string) bool {
	s = strings.ToLower(s)
	for _, sub := range substrs {
		if strings.Contains(s, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}
