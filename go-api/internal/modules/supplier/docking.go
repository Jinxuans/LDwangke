package supplier

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-api/internal/model"
)

func waitSupplierHost(sup *model.SupplierFull) {
	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}
}

func (s *Service) CallSupplierOrder(sup *model.SupplierFull, cls *model.ClassFull, school, user, pass, kcid, kcname string, extraFields map[string]string) (*model.SupplierOrderResult, error) {
	if sup.PT == "yyy" {
		return yyyCallOrder(sup, user, pass, kcname, cls.Noun)
	}
	if sup.PT == "KUN" || sup.PT == "kunba" {
		return kunCallOrder(sup, cls.Noun, school, user, pass, kcname, kcid)
	}
	if sup.PT == "simple" {
		return simpleCallOrder(sup, cls.Noun, school, user, pass, kcname, kcid)
	}

	cfg := GetPlatformConfig(sup.PT)
	apiURL := buildSupplierURL(sup.URL, cfg.OrderAct)

	formData := url.Values{}
	formData.Set("uid", sup.User)
	formData.Set("key", sup.Pass)
	formData.Set("platform", cls.Noun)
	formData.Set("school", school)
	formData.Set("user", user)
	formData.Set("pass", pass)
	formData.Set("kcid", kcid)
	formData.Set("kcname", kcname)

	if cfg.ExtraParams && extraFields != nil {
		for k, v := range extraFields {
			if v != "" {
				formData.Set(k, v)
			}
		}
	}

	waitSupplierHost(sup)
	resp, err := s.client.PostForm(apiURL, formData)
	if err != nil {
		return nil, fmt.Errorf("请求上游失败：%v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败：%v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析响应失败：%s", string(body))
	}

	codeVal := fmt.Sprintf("%v", raw["code"])
	result := &model.SupplierOrderResult{
		Msg: fmt.Sprintf("%v", raw["msg"]),
	}

	if codeVal == cfg.SuccessCode {
		result.Code = 1
		if result.Msg == "" {
			result.Msg = "下单成功"
		}
		if cfg.ReturnsYID {
			if idVal, ok := raw["id"]; ok && idVal != nil {
				switch v := idVal.(type) {
				case string:
					result.YID = v
				case float64:
					result.YID = fmt.Sprintf("%.0f", v)
				case int:
					result.YID = fmt.Sprintf("%d", v)
				default:
					result.YID = fmt.Sprintf("%v", v)
				}
			}
			if result.YID == "" {
				if idsArr, ok := raw["ids"].([]interface{}); ok && len(idsArr) > 0 {
					switch v := idsArr[0].(type) {
					case string:
						result.YID = v
					case float64:
						result.YID = fmt.Sprintf("%.0f", v)
					default:
						result.YID = fmt.Sprintf("%v", v)
					}
				}
			}
			if result.YID == "" {
				if dataArr, ok := raw["data"].([]interface{}); ok && len(dataArr) > 0 {
					if str, ok := dataArr[0].(string); ok {
						result.YID = str
					} else {
						switch v := dataArr[0].(type) {
						case string:
							result.YID = v
						case float64:
							result.YID = fmt.Sprintf("%.0f", v)
						default:
							result.YID = fmt.Sprintf("%v", v)
						}
					}
				}
			}
			if result.YID == "" {
				if dataMap, ok := raw["data"].(map[string]interface{}); ok {
					if id, ok := dataMap["id"]; ok && id != nil {
						switch v := id.(type) {
						case string:
							result.YID = v
						case float64:
							result.YID = fmt.Sprintf("%.0f", v)
						default:
							result.YID = fmt.Sprintf("%v", v)
						}
					}
				}
			}
			if result.YID == "" && cfg.YIDInDataArray {
				if dataArr, ok := raw["data"].([]interface{}); ok && len(dataArr) > 0 {
					switch v := dataArr[0].(type) {
					case string:
						result.YID = v
					case float64:
						result.YID = fmt.Sprintf("%.0f", v)
					default:
						result.YID = fmt.Sprintf("%v", v)
					}
				}
			}
		}
	} else {
		result.Code = -1
		if result.Msg == "" {
			result.Msg = fmt.Sprintf("上游返回错误码：%s", codeVal)
		}
	}

	return result, nil
}

func (s *Service) QueryOrderProgress(sup *model.SupplierFull, yid string, username string, orderExtra map[string]string) ([]model.SupplierProgressItem, error) {
	if sup.PT == "yyy" {
		return yyyQueryProgress(sup, username)
	}
	if sup.PT == "simple" {
		platform := ""
		kcname := ""
		kcid := ""
		school := ""
		user := username
		pass := ""
		if orderExtra != nil {
			platform = orderExtra["noun"]
			kcname = orderExtra["kcname"]
			kcid = orderExtra["kcid"]
			school = orderExtra["school"]
			if v := orderExtra["user"]; v != "" {
				user = v
			}
			pass = orderExtra["pass"]
		}
		return simpleQueryProgress(sup, platform, school, user, pass, kcname, kcid)
	}

	cfg := GetPlatformConfig(sup.PT)
	params := url.Values{}
	params.Set("uid", sup.User)
	params.Set("key", sup.Pass)

	var apiURL string
	if cfg.ProgressPath != "" {
		baseURL := strings.TrimRight(sup.URL, "/")
		if !strings.HasPrefix(baseURL, "http") {
			baseURL = "http://" + baseURL
		}
		apiURL = baseURL + cfg.ProgressPath

		if strings.Contains(cfg.ProgressPath, "/api/search") {
			params.Set("username", username)
			if orderExtra != nil {
				if v, ok := orderExtra["kcname"]; ok {
					params.Set("kcname", v)
				}
				if v, ok := orderExtra["noun"]; ok {
					params.Set("cid", v)
				}
			}
		} else {
			params.Set("username", username)
			if yid != "" && yid != "0" {
				params.Set("yid", yid)
			}
			if orderExtra != nil {
				if v, ok := orderExtra["kcname"]; ok {
					params.Set("kcname", v)
				}
				if v, ok := orderExtra["noun"]; ok {
					params.Set("cid", v)
				}
			}
		}
	} else if yid != "" && yid != "0" {
		apiURL = buildSupplierURL(sup.URL, cfg.ProgressAct)
		if cfg.UseUUIDParam {
			params.Set("uuid", yid)
		} else if cfg.UseIDParam {
			params.Set("id", yid)
		} else {
			params.Set("yid", yid)
		}
		if cfg.AlwaysUsername {
			params.Set("username", username)
		}
	} else {
		apiURL = buildSupplierURL(sup.URL, cfg.ProgressNoYID)
		params.Set("username", username)
	}

	var resp *http.Response
	var err error
	if cfg.ProgressMethod == "GET" {
		apiURL = apiURL + "?" + params.Encode()
		waitSupplierHost(sup)
		resp, err = s.client.Get(apiURL)
	} else {
		waitSupplierHost(sup)
		resp, err = s.client.PostForm(apiURL, params)
	}
	if err != nil {
		return nil, fmt.Errorf("请求上游失败：%v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败：%v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析响应失败：%s", string(body))
	}

	codeVal := ""
	if codeRaw, ok := raw["code"]; ok {
		switch v := codeRaw.(type) {
		case string:
			codeVal = v
		case float64:
			codeVal = fmt.Sprintf("%.0f", v)
		case int:
			codeVal = fmt.Sprintf("%d", v)
		default:
			codeVal = fmt.Sprintf("%v", v)
		}
	}
	if codeVal != "0" && codeVal != "1" {
		msg := ""
		if msgRaw, ok := raw["msg"]; ok {
			msg = fmt.Sprintf("%v", msgRaw)
		}
		if msg == "" {
			msg = "查询上游进度失败"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	var items []model.SupplierProgressItem
	if dataArr, ok := raw["data"].([]interface{}); ok {
		for _, item := range dataArr {
			if m, ok := item.(map[string]interface{}); ok {
				items = append(items, model.SupplierProgressItem{
					YID:             toString(m["id"]),
					KCName:          toString(m["kcname"]),
					User:            toString(m["user"]),
					Status:          toString(m["status"]),
					StatusText:      toString(m["status_text"]),
					Process:         toString(m["process"]),
					Remarks:         toString(m["remarks"]),
					CourseStartTime: toString(m["courseStartTime"]),
					CourseEndTime:   toString(m["courseEndTime"]),
					ExamStartTime:   toString(m["examStartTime"]),
					ExamEndTime:     toString(m["examEndTime"]),
				})
			}
		}
	}

	return items, nil
}

func (s *Service) PauseOrder(sup *model.SupplierFull, yid string) (int, string, error) {
	if sup.PT == "yyy" || sup.PT == "tuboshu" {
		return -1, "当前平台暂不支持暂停操作", nil
	}
	if sup.PT == "KUN" || sup.PT == "kunba" {
		return kunPauseOrder(sup, yid)
	}

	cfg := GetPlatformConfig(sup.PT)
	var resp *http.Response
	var err error

	if cfg.PausePath != "" {
		baseURL := strings.TrimRight(sup.URL, "/")
		if !strings.HasPrefix(baseURL, "http") {
			baseURL = "http://" + baseURL
		}
		apiURL := baseURL + cfg.PausePath
		jsonData, _ := json.Marshal(map[string]string{"id": yid, "username": sup.User})
		waitSupplierHost(sup)
		resp, err = s.client.Post(apiURL, "application/json", strings.NewReader(string(jsonData)))
	} else {
		pauseAct := cfg.PauseAct
		if pauseAct == "" {
			pauseAct = "zt"
		}
		apiURL := buildSupplierURL(sup.URL, pauseAct)
		formData := url.Values{}
		formData.Set("uid", sup.User)
		formData.Set("key", sup.Pass)
		idParam := cfg.PauseIDParam
		if idParam == "" {
			idParam = "id"
		}
		formData.Set(idParam, yid)
		waitSupplierHost(sup)
		resp, err = s.client.PostForm(apiURL, formData)
	}
	if err != nil {
		return -1, "", fmt.Errorf("请求上游失败：%v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return -1, string(body), nil
	}
	return result.Code, result.Msg, nil
}

func (s *Service) ChangePassword(sup *model.SupplierFull, yid, newPwd string) (int, string, error) {
	if sup.PT == "yyy" || sup.PT == "tuboshu" {
		return -1, "当前平台暂不支持改密操作", nil
	}
	if sup.PT == "KUN" || sup.PT == "kunba" {
		return kunChangePassword(sup, yid, newPwd)
	}

	cfg := GetPlatformConfig(sup.PT)
	var resp *http.Response
	var err error

	if cfg.ChangePassPath != "" {
		baseURL := strings.TrimRight(sup.URL, "/")
		if !strings.HasPrefix(baseURL, "http") {
			baseURL = "http://" + baseURL
		}
		apiURL := baseURL + cfg.ChangePassPath
		jsonData, _ := json.Marshal(map[string]string{"id": yid, "username": sup.User, "pass": newPwd})
		waitSupplierHost(sup)
		resp, err = s.client.Post(apiURL, "application/json", strings.NewReader(string(jsonData)))
	} else {
		changeAct := cfg.ChangePassAct
		if changeAct == "" {
			changeAct = "gaimi"
		}
		apiURL := buildSupplierURL(sup.URL, changeAct)
		formData := url.Values{}
		formData.Set("uid", sup.User)
		formData.Set("key", sup.Pass)
		idParam := cfg.ChangePassIDParam
		if idParam == "" {
			idParam = "id"
		}
		pwdParam := cfg.ChangePassParam
		if pwdParam == "" {
			pwdParam = "newPwd"
		}
		formData.Set(idParam, yid)
		formData.Set(pwdParam, newPwd)
		waitSupplierHost(sup)
		resp, err = s.client.PostForm(apiURL, formData)
	}
	if err != nil {
		return -1, "", fmt.Errorf("请求上游失败：%v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return -1, string(body), nil
	}
	return result.Code, result.Msg, nil
}

func (s *Service) ResubmitOrder(sup *model.SupplierFull, yid string) (int, string, error) {
	if sup.PT == "tuboshu" || sup.PT == "nx" {
		return -1, "当前平台暂不支持补单操作", nil
	}
	if sup.PT == "simple" {
		return simpleResubmit(sup, yid)
	}

	cfg := GetPlatformConfig(sup.PT)
	var resp *http.Response
	var err error

	if cfg.ResubmitPath != "" {
		baseURL := strings.TrimRight(sup.URL, "/")
		if !strings.HasPrefix(baseURL, "http") {
			baseURL = "http://" + baseURL
		}
		apiURL := baseURL + cfg.ResubmitPath
		jsonData, _ := json.Marshal(map[string]string{"id": yid, "username": sup.User})
		waitSupplierHost(sup)
		resp, err = s.client.Post(apiURL, "application/json", strings.NewReader(string(jsonData)))
	} else {
		apiURL := buildSupplierURL(sup.URL, "budan")
		formData := url.Values{}
		formData.Set("uid", sup.User)
		formData.Set("key", sup.Pass)
		formData.Set(cfg.ResubmitIDParam, yid)
		waitSupplierHost(sup)
		resp, err = s.client.PostForm(apiURL, formData)
	}
	if err != nil {
		return -1, "", fmt.Errorf("请求上游失败：%v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return -1, string(body), nil
	}
	return result.Code, result.Msg, nil
}

func (s *Service) ResetOrderScore(sup *model.SupplierFull, yid string, newScore int) (int, string, error) {
	if sup.PT != "pup" {
		return -1, "当前平台不支持重置分数", nil
	}
	if newScore < 70 || newScore > 100 {
		return -1, "分数范围 70-100", nil
	}

	apiURL := buildSupplierURL(sup.URL, "resetscore")
	formData := url.Values{}
	formData.Set("uid", sup.User)
	formData.Set("key", sup.Pass)
	formData.Set("oid", yid)
	formData.Set("newscore", fmt.Sprintf("%d", newScore))

	waitSupplierHost(sup)
	resp, err := s.client.PostForm(apiURL, formData)
	if err != nil {
		return -1, "", fmt.Errorf("请求上游失败：%v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return -1, string(body), nil
	}
	return result.Code, result.Msg, nil
}

func (s *Service) ResetOrderDuration(sup *model.SupplierFull, yid string, newDuration int) (int, string, error) {
	if sup.PT != "pup" {
		return -1, "当前平台不支持重置时长", nil
	}
	if newDuration < 0 || newDuration > 50 {
		return -1, "时长范围 0-50 小时", nil
	}

	apiURL := buildSupplierURL(sup.URL, "resetsc")
	formData := url.Values{}
	formData.Set("uid", sup.User)
	formData.Set("key", sup.Pass)
	formData.Set("oid", yid)
	formData.Set("newsc", fmt.Sprintf("%d", newDuration))

	waitSupplierHost(sup)
	resp, err := s.client.PostForm(apiURL, formData)
	if err != nil {
		return -1, "", fmt.Errorf("请求上游失败：%v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return -1, string(body), nil
	}
	return result.Code, result.Msg, nil
}

func (s *Service) ResetOrderPeriod(sup *model.SupplierFull, yid string, newPeriod int) (int, string, error) {
	if sup.PT != "pup" {
		return -1, "当前平台不支持重置周期", nil
	}
	if newPeriod < 1 || newPeriod > 20 {
		return -1, "周期范围 1-20 天", nil
	}

	apiURL := buildSupplierURL(sup.URL, "resetcron")
	formData := url.Values{}
	formData.Set("uid", sup.User)
	formData.Set("key", sup.Pass)
	formData.Set("oid", yid)
	formData.Set("newcron", fmt.Sprintf("%d", newPeriod))

	waitSupplierHost(sup)
	resp, err := s.client.PostForm(apiURL, formData)
	if err != nil {
		return -1, "", fmt.Errorf("请求上游失败：%v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return -1, string(body), nil
	}
	return result.Code, result.Msg, nil
}

func (s *Service) QueryOrderLogs(sup *model.SupplierFull, yid string, orderExtra ...map[string]string) ([]model.OrderLogEntry, error) {
	if sup.PT == "yyy" || sup.PT == "tuboshu" {
		return nil, fmt.Errorf("当前平台暂不支持查看日志")
	}
	if sup.PT == "longlong" {
		return queryLonglongLogs(sup, yid)
	}

	cfg := GetPlatformConfig(sup.PT)
	logClient := &http.Client{Timeout: 8 * time.Second}
	var resp *http.Response
	var err error

	if cfg.LogPath != "" {
		baseURL := strings.TrimRight(sup.URL, "/")
		if !strings.HasPrefix(baseURL, "http") {
			baseURL = "http://" + baseURL
		}
		if sup.PT == "wanzi" {
			apiURL := fmt.Sprintf("%s%s%s/logs?pageSize=50&uid=%s&key=%s", baseURL, cfg.LogPath, yid, sup.User, sup.Pass)
			waitSupplierHost(sup)
			resp, err = logClient.Get(apiURL)
		} else {
			params := url.Values{}
			if len(orderExtra) > 0 {
				if v, ok := orderExtra[0]["user"]; ok {
					params.Set("account", v)
				}
				if v, ok := orderExtra[0]["pass"]; ok {
					params.Set("password", v)
				}
				if v, ok := orderExtra[0]["kcname"]; ok {
					params.Set("course", v)
				}
				if v, ok := orderExtra[0]["kcid"]; ok {
					params.Set("courseId", v)
				}
			}
			params.Set("dtoken", sup.Token)
			apiURL := baseURL + cfg.LogPath + "?" + params.Encode()
			waitSupplierHost(sup)
			resp, err = logClient.Get(apiURL)
		}
	} else {
		logAct := cfg.LogAct
		if logAct == "" {
			logAct = "xq"
		}
		logIDParam := cfg.LogIDParam
		if logIDParam == "" {
			logIDParam = "id"
		}
		apiURL := buildSupplierURL(sup.URL, logAct)
		formData := url.Values{}
		formData.Set("uid", sup.User)
		formData.Set("key", sup.Pass)
		formData.Set(logIDParam, yid)
		waitSupplierHost(sup)
		resp, err = logClient.PostForm(apiURL, formData)
	}
	if err != nil {
		return nil, fmt.Errorf("请求上游超时或失败")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析响应失败：%s", string(body))
	}

	var code int
	if c, ok := raw["code"]; ok {
		json.Unmarshal(c, &code)
	}
	if code != 1 && code != 0 {
		var msg string
		if m, ok := raw["msg"]; ok {
			json.Unmarshal(m, &msg)
		}
		if msg == "" {
			msg = "查询日志失败"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	var logsRaw json.RawMessage
	if l, ok := raw["logs"]; ok {
		logsRaw = l
	} else if d, ok := raw["data"]; ok {
		logsRaw = d
	}
	if logsRaw == nil {
		return []model.OrderLogEntry{}, nil
	}

	var entries []model.OrderLogEntry
	if err := json.Unmarshal(logsRaw, &entries); err == nil {
		return entries, nil
	}

	var strLogs []string
	if err := json.Unmarshal(logsRaw, &strLogs); err == nil {
		for _, line := range strLogs {
			entries = append(entries, model.OrderLogEntry{Remarks: line})
		}
		return entries, nil
	}

	var arrLogs [][]interface{}
	if err := json.Unmarshal(logsRaw, &arrLogs); err == nil {
		for _, row := range arrLogs {
			entry := model.OrderLogEntry{}
			if len(row) > 8 {
				entry.Time = fmt.Sprintf("%v", row[8])
			}
			if len(row) > 4 {
				entry.Course = fmt.Sprintf("%v", row[4])
			}
			if len(row) > 9 {
				entry.Status = fmt.Sprintf("%v", row[9])
			}
			if len(row) > 10 {
				entry.Process = fmt.Sprintf("%v", row[10])
			}
			if len(row) > 11 {
				entry.Remarks = fmt.Sprintf("%v", row[11])
			}
			entries = append(entries, entry)
		}
		return entries, nil
	}

	return []model.OrderLogEntry{}, nil
}

func (s *Service) SubmitReport(sup *model.SupplierFull, yid, ticketType, content string) (int, int, string, error) {
	cfg := GetPlatformConfig(sup.PT)
	baseURL := strings.TrimRight(sup.URL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}

	var resp *http.Response
	var err error
	if cfg.ReportParamStyle == "token" {
		apiURL := baseURL + cfg.ReportPath
		jsonData, _ := json.Marshal(map[string]string{
			"token":   getSupplierToken(sup),
			"type":    ticketType,
			"id":      yid,
			"content": content,
		})
		req, _ := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonData)))
		req.Header.Set("Content-Type", "application/json")
		waitSupplierHost(sup)
		resp, err = s.client.Do(req)
	} else {
		apiURL := baseURL + cfg.ReportPath
		if cfg.ReportPath == "" {
			apiURL = fmt.Sprintf("%s/api.php?act=%s", baseURL, cfg.ReportAct)
		}
		formData := url.Values{
			"uid":      {sup.User},
			"key":      {sup.Pass},
			"id":       {yid},
			"question": {content},
		}
		waitSupplierHost(sup)
		resp, err = s.client.PostForm(apiURL, formData)
	}
	if err != nil {
		return 0, 0, "", fmt.Errorf("请求上游失败：%v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return 0, 0, "", fmt.Errorf("上游返回解析失败：%s", string(body))
	}

	code := 0
	if c, ok := raw["code"].(float64); ok {
		code = int(c)
	}
	msg, _ := raw["msg"].(string)
	workID := 0
	if wid, ok := raw["workId"].(float64); ok {
		workID = int(wid)
	} else if dataMap, ok := raw["data"].(map[string]interface{}); ok {
		if rid, ok := dataMap["reportId"].(float64); ok {
			workID = int(rid)
		}
	}
	return code, workID, msg, nil
}

func (s *Service) QueryReport(sup *model.SupplierFull, reportID string) (int, string, string, error) {
	cfg := GetPlatformConfig(sup.PT)
	baseURL := strings.TrimRight(sup.URL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}

	var resp *http.Response
	var err error
	if cfg.ReportParamStyle == "token" {
		apiURL := baseURL + cfg.GetReportPath
		jsonData, _ := json.Marshal(map[string]string{
			"token":  getSupplierToken(sup),
			"workId": reportID,
		})
		req, _ := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonData)))
		req.Header.Set("Content-Type", "application/json")
		waitSupplierHost(sup)
		resp, err = s.client.Do(req)
	} else {
		apiURL := baseURL + cfg.GetReportPath
		if cfg.GetReportPath == "" {
			apiURL = fmt.Sprintf("%s/api.php?act=%s", baseURL, cfg.GetReportAct)
		}
		formData := url.Values{
			"uid":      {sup.User},
			"key":      {sup.Pass},
			"reportId": {reportID},
		}
		waitSupplierHost(sup)
		resp, err = s.client.PostForm(apiURL, formData)
	}
	if err != nil {
		return 0, "", "", fmt.Errorf("请求上游失败：%v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return 0, "", "", fmt.Errorf("上游返回解析失败：%s", string(body))
	}

	code := 0
	if c, ok := raw["code"].(float64); ok {
		code = int(c)
	}
	answer := ""
	state := ""
	if dataMap, ok := raw["data"].(map[string]interface{}); ok {
		if a, ok := dataMap["answer"].(string); ok {
			answer = a
		}
		if st, ok := dataMap["state"].(string); ok {
			state = st
		}
		if state == "" {
			if s, ok := dataMap["status"].(float64); ok {
				state = fmt.Sprintf("%d", int(s))
			}
		}
	}
	return code, answer, state, nil
}
