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
	if err := requireExplicitActionConfig("下单接口", cfg.OrderPath, cfg.OrderMethod, cfg.OrderParamMap); err != nil {
		return nil, err
	}
	apiURL := resolveConfiguredActionURL(sup.URL, cfg.OrderPath)

	defaultParams := defaultSupplierAuthParams(sup, cfg.AuthType)
	defaultParams["platform"] = cls.Noun
	defaultParams["school"] = school
	defaultParams["user"] = user
	defaultParams["pass"] = pass
	defaultParams["kcid"] = kcid
	defaultParams["kcname"] = kcname
	actionFields := map[string]string{
		"order.school": school,
		"order.user":   user,
		"order.pass":   pass,
		"order.kcid":   kcid,
		"order.kcname": kcname,
		"order.noun":   cls.Noun,
	}

	if cfg.ExtraParams && extraFields != nil {
		for k, v := range extraFields {
			if v != "" {
				defaultParams[k] = v
				actionFields["order."+k] = v
			}
		}
	}

	execResult, err := s.executeConfiguredAction(
		sup,
		apiURL,
		cfg.OrderMethod,
		cfg.OrderBodyType,
		cfg.OrderParamMap,
		http.MethodPost,
		"form",
		defaultParams,
		actionFields,
	)
	if err != nil {
		return nil, fmt.Errorf("请求上游失败：%v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(execResult.Body, &raw); err != nil {
		return nil, fmt.Errorf("解析响应失败：%s", string(execResult.Body))
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

// QueryOrderProgress 把“主订单查询进度”的平台差异收敛在一处。
// 现在只保留“一套进度接口配置”：
// - 运行时只读取 progress_path；
// - 不再区分“有 yid / 无 yid”两套接口。
//
// 调用方只提供：
// - sup: 本地供应商配置，决定该走哪个平台适配器；
// - yid: 上游订单号，可选，存在时会按参数映射传递；
// - username: 某些平台仍要求按账号参与查询；
// - orderExtra: 课程名、课程ID、学校等补充字段，用来兼容特殊平台的搜索接口。
func (s *Service) QueryOrderProgress(sup *model.SupplierFull, yid string, username string, orderExtra map[string]string) ([]model.SupplierProgressItem, error) {
	debugInfo := newProgressDebugInfo(orderExtra)

	// `yyy` 和 `simple` 平台历史上就有独立协议，不走通用配置拼装逻辑。
	if sup.PT == "yyy" {
		return yyyQueryProgress(sup, username, debugInfo)
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
		return simpleQueryProgress(sup, platform, school, user, pass, kcname, kcid, debugInfo)
	}

	cfg := GetPlatformConfig(sup.PT)
	actionFields := map[string]string{}
	if username != "" {
		actionFields["order.user"] = username
	}
	if yid != "" {
		actionFields["order.yid"] = yid
	}

	if orderExtra != nil {
		for k, v := range orderExtra {
			if v == "" {
				continue
			}
			actionFields["order."+k] = v
		}
	}

	// 进度查询现在始终只走同一套 endpoint 配置。
	if err := requireExplicitActionConfig("进度接口", cfg.ProgressPath, cfg.ProgressMethod, cfg.ProgressParamMap); err != nil {
		return nil, err
	}
	apiURL := resolveConfiguredActionURL(sup.URL, cfg.ProgressPath)
	params, err := buildActionParams(cfg.ProgressParamMap, sup, actionFields, nil)
	if err != nil {
		return nil, err
	}
	fallbackBodyType := "form"
	if strings.EqualFold(cfg.ProgressMethod, http.MethodGet) {
		fallbackBodyType = "query"
	}
	req, contentType, payload, err := prepareActionRequest(
		apiURL,
		normalizeActionMethod(cfg.ProgressMethod, http.MethodPost),
		normalizeActionBodyType(cfg.ProgressBodyType, fallbackBodyType, normalizeActionMethod(cfg.ProgressMethod, http.MethodPost)),
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败：%v", err)
	}
	debugInfo.logRequest(sup, req.Method, req.URL.String(), contentType, payload)
	waitSupplierHost(sup)
	resp, err := s.client.Do(req)
	if err != nil {
		debugInfo.logRequestError(sup, err)
		return nil, fmt.Errorf("请求上游失败：%v", err)
	}
	defer resp.Body.Close()

	// 上游协议非常不统一，所以这里先读成原始 body，再做宽松 JSON 解析。
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败：%v", err)
	}
	debugInfo.logResponse(sup, resp.Status, body)

	return parseConfiguredProgressResponse(body)
}

func (s *Service) PauseOrder(sup *model.SupplierFull, yid string) (int, string, error) {
	if sup.PT == "yyy" || sup.PT == "tuboshu" {
		return -1, "当前平台暂不支持暂停操作", nil
	}
	if sup.PT == "KUN" || sup.PT == "kunba" {
		return kunPauseOrder(sup, yid)
	}

	cfg := GetPlatformConfig(sup.PT)
	if err := requireExplicitActionConfig("暂停接口", cfg.PausePath, cfg.PauseMethod, cfg.PauseParamMap); err != nil {
		return -1, "", err
	}
	apiURL := resolveConfiguredActionURL(sup.URL, cfg.PausePath)
	defaultParams := defaultSupplierAuthParams(sup, cfg.AuthType)
	fallbackBodyType := "form"
	idParam := cfg.PauseIDParam
	if idParam == "" {
		idParam = "id"
	}
	defaultParams[idParam] = yid
	execResult, err := s.executeConfiguredAction(
		sup,
		apiURL,
		cfg.PauseMethod,
		cfg.PauseBodyType,
		cfg.PauseParamMap,
		http.MethodPost,
		fallbackBodyType,
		defaultParams,
		map[string]string{"order.yid": yid},
	)
	if err != nil {
		return -1, "", fmt.Errorf("请求上游失败：%v", err)
	}

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(execResult.Body, &result); err != nil {
		return -1, string(execResult.Body), nil
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
	if err := requireExplicitActionConfig("改密接口", cfg.ChangePassPath, cfg.ChangePassMethod, cfg.ChangePassParamMap); err != nil {
		return -1, "", err
	}
	apiURL := resolveConfiguredActionURL(sup.URL, cfg.ChangePassPath)
	defaultParams := defaultSupplierAuthParams(sup, cfg.AuthType)
	fallbackBodyType := "form"
	idParam := cfg.ChangePassIDParam
	if idParam == "" {
		idParam = "id"
	}
	pwdParam := cfg.ChangePassParam
	if pwdParam == "" {
		pwdParam = "newPwd"
	}
	defaultParams[idParam] = yid
	defaultParams[pwdParam] = newPwd
	execResult, err := s.executeConfiguredAction(
		sup,
		apiURL,
		cfg.ChangePassMethod,
		cfg.ChangePassBodyType,
		cfg.ChangePassParamMap,
		http.MethodPost,
		fallbackBodyType,
		defaultParams,
		map[string]string{"order.yid": yid, "action.new_password": newPwd},
	)
	if err != nil {
		return -1, "", fmt.Errorf("请求上游失败：%v", err)
	}

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(execResult.Body, &result); err != nil {
		return -1, string(execResult.Body), nil
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
	if err := requireExplicitActionConfig("补单接口", cfg.ResubmitPath, cfg.ResubmitMethod, cfg.ResubmitParamMap); err != nil {
		return -1, "", err
	}
	apiURL := resolveConfiguredActionURL(sup.URL, cfg.ResubmitPath)
	defaultParams := defaultSupplierAuthParams(sup, cfg.AuthType)
	fallbackBodyType := "form"
	idParam := cfg.ResubmitIDParam
	if idParam == "" {
		idParam = "id"
	}
	defaultParams[idParam] = yid
	execResult, err := s.executeConfiguredAction(
		sup,
		apiURL,
		cfg.ResubmitMethod,
		cfg.ResubmitBodyType,
		cfg.ResubmitParamMap,
		http.MethodPost,
		fallbackBodyType,
		defaultParams,
		map[string]string{"order.yid": yid},
	)
	if err != nil {
		return -1, "", fmt.Errorf("请求上游失败：%v", err)
	}

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(execResult.Body, &result); err != nil {
		return -1, string(execResult.Body), nil
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

	var execResult *actionExecutionResult
	var err error
	if err := requireExplicitActionConfig("日志接口", cfg.LogPath, cfg.LogMethod, cfg.LogParamMap); err != nil {
		return nil, err
	}
	if cfg.LogPath != "" {
		fields := map[string]string{}
		if yid != "" {
			fields["order.yid"] = yid
		}
		defaultParams := defaultSupplierAuthParams(sup, cfg.AuthType)
		if len(orderExtra) > 0 {
			for k, v := range orderExtra[0] {
				if v == "" {
					continue
				}
				fields["order."+k] = v
			}
			if v := orderExtra[0]["user"]; v != "" {
				defaultParams["account"] = v
			}
			if v := orderExtra[0]["pass"]; v != "" {
				defaultParams["password"] = v
			}
			if v := orderExtra[0]["kcname"]; v != "" {
				defaultParams["course"] = v
			}
			if v := orderExtra[0]["kcid"]; v != "" {
				defaultParams["courseId"] = v
			}
		}
		fallbackBodyType := "form"
		if strings.EqualFold(cfg.LogMethod, http.MethodGet) {
			fallbackBodyType = "query"
		}
		execResult, err = s.executeConfiguredActionWithClient(
			logClient,
			sup,
			resolveConfiguredActionURL(sup.URL, cfg.LogPath),
			cfg.LogMethod,
			cfg.LogBodyType,
			cfg.LogParamMap,
			http.MethodGet,
			fallbackBodyType,
			defaultParams,
			fields,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("请求上游超时或失败")
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(execResult.Body, &raw); err != nil {
		return nil, fmt.Errorf("解析响应失败：%s", string(execResult.Body))
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
	if err := requireExplicitActionConfig("提交工单接口", cfg.ReportPath, cfg.ReportMethod, cfg.ReportParamMap); err != nil {
		return 0, 0, "", err
	}
	apiURL := resolveConfiguredActionURL(sup.URL, cfg.ReportPath)
	defaultParams := map[string]string{}
	fallbackBodyType := "form"
	if cfg.ReportParamStyle == "token" {
		defaultParams["token"] = getSupplierToken(sup)
		defaultParams["type"] = ticketType
		defaultParams["id"] = yid
		defaultParams["content"] = content
		fallbackBodyType = "json"
	} else {
		defaultParams = defaultSupplierAuthParams(sup, cfg.AuthType)
		defaultParams["id"] = yid
		defaultParams["question"] = content
	}
	execResult, err := s.executeConfiguredAction(
		sup,
		apiURL,
		cfg.ReportMethod,
		cfg.ReportBodyType,
		cfg.ReportParamMap,
		http.MethodPost,
		fallbackBodyType,
		defaultParams,
		map[string]string{
			"order.yid":          yid,
			"ticket_type":        ticketType,
			"content":            content,
			"action.ticket_type": ticketType,
			"action.content":     content,
		},
	)
	if err != nil {
		return 0, 0, "", fmt.Errorf("请求上游失败：%v", err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(execResult.Body, &raw); err != nil {
		return 0, 0, "", fmt.Errorf("上游返回解析失败：%s", string(execResult.Body))
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
	if err := requireExplicitActionConfig("查询工单接口", cfg.GetReportPath, cfg.GetReportMethod, cfg.GetReportParamMap); err != nil {
		return 0, "", "", err
	}
	apiURL := resolveConfiguredActionURL(sup.URL, cfg.GetReportPath)
	defaultParams := map[string]string{}
	fallbackBodyType := "form"
	if cfg.ReportParamStyle == "token" {
		defaultParams["token"] = getSupplierToken(sup)
		defaultParams["workId"] = reportID
		fallbackBodyType = "json"
	} else {
		defaultParams = defaultSupplierAuthParams(sup, cfg.AuthType)
		defaultParams["reportId"] = reportID
	}
	execResult, err := s.executeConfiguredAction(
		sup,
		apiURL,
		cfg.GetReportMethod,
		cfg.GetReportBodyType,
		cfg.GetReportParamMap,
		http.MethodPost,
		fallbackBodyType,
		defaultParams,
		map[string]string{
			"action.report_id": reportID,
		},
	)
	if err != nil {
		return 0, "", "", fmt.Errorf("请求上游失败：%v", err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(execResult.Body, &raw); err != nil {
		return 0, "", "", fmt.Errorf("上游返回解析失败：%s", string(execResult.Body))
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
