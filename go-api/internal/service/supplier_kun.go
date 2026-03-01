package service

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

// kunBuildBaseURL 构建 KUN 平台基础 URL
func kunBuildBaseURL(sup *model.SupplierFull) string {
	baseURL := strings.TrimRight(sup.URL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	return baseURL
}

// kunGetDToken 获取 KUN 平台的 dtoken（优先 Token 字段，其次 Pass 字段）
func kunGetDToken(sup *model.SupplierFull) string {
	if sup.Token != "" {
		return sup.Token
	}
	return sup.Pass
}

// kunCallQuery KUN/kunba 平台查课：GET /query/?platform=&school=&account=&password=&dtoken=
func kunCallQuery(sup *model.SupplierFull, platform, school, user, pass string) (*model.SupplierQueryResult, error) {
	baseURL := kunBuildBaseURL(sup)
	dtoken := kunGetDToken(sup)

	params := url.Values{}
	params.Set("platform", platform)
	params.Set("school", school)
	params.Set("account", user)
	params.Set("password", pass)
	params.Set("dtoken", dtoken)

	apiURL := baseURL + "/query/?" + params.Encode()

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("请求KUN查课失败：%v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败：%v", err)
	}

	// 尝试解析为标准格式 {code, data, msg}
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析响应失败：%s", string(body))
	}

	var items []model.CourseItem
	if dataArr, ok := raw["data"].([]interface{}); ok {
		for _, item := range dataArr {
			if m, ok := item.(map[string]interface{}); ok {
				items = append(items, model.CourseItem{
					ID:             toString(m["id"]),
					Name:           toString(m["name"]),
					KCJS:           toString(m["kcjs"]),
					StudyStartTime: toString(m["studyStartTime"]),
					StudyEndTime:   toString(m["studyEndTime"]),
					ExamStartTime:  toString(m["examStartTime"]),
					ExamEndTime:    toString(m["examEndTime"]),
					Complete:       toString(m["complete"]),
				})
			}
		}
	}

	msg := "查询成功"
	if m, ok := raw["msg"].(string); ok && m != "" {
		msg = m
	}

	return &model.SupplierQueryResult{
		Msg:  msg,
		Data: items,
	}, nil
}

// kunCallOrder KUN/kunba 平台下单：GET /getorder/?platform=&school=&account=&password=&course=&kcid=&dtoken=
func kunCallOrder(sup *model.SupplierFull, platform, school, user, pass, kcname, kcid string) (*model.SupplierOrderResult, error) {
	baseURL := kunBuildBaseURL(sup)
	dtoken := kunGetDToken(sup)

	params := url.Values{}
	params.Set("platform", platform)
	params.Set("school", school)
	params.Set("account", user)
	params.Set("password", pass)
	params.Set("course", kcname)
	if kcid != "" {
		params.Set("kcid", kcid)
	}
	params.Set("dtoken", dtoken)

	apiURL := baseURL + "/getorder/?" + params.Encode()

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("请求KUN下单失败：%v", err)
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

	fmt.Printf("[kunCallOrder] codeVal=%s raw=%v\n", codeVal, raw)

	if codeVal == "0" || codeVal == "200" {
		result.Code = 1
		if result.Msg == "" || result.Msg == "<nil>" {
			result.Msg = "下单成功"
		}
		// 提取 token 作为 yid（KUN 下单返回 token 用于后续操作）
		if token, ok := raw["token"].(string); ok && token != "" {
			result.YID = token
		}
		// 也尝试从 id 字段提取
		if result.YID == "" {
			if idVal, ok := raw["id"]; ok && idVal != nil {
				result.YID = fmt.Sprintf("%v", idVal)
			}
		}
		// 尝试从 data 提取
		if result.YID == "" {
			if dataArr, ok := raw["data"].([]interface{}); ok && len(dataArr) > 0 {
				result.YID = fmt.Sprintf("%v", dataArr[0])
			}
		}
		fmt.Printf("[kunCallOrder] extracted YID=%s\n", result.YID)
	} else {
		result.Code = -1
		if result.Msg == "" || result.Msg == "<nil>" {
			result.Msg = fmt.Sprintf("上游返回错误码：%s", codeVal)
		}
	}

	return result, nil
}

// kunChangePassword KUN/kunba 平台改密：GET /upPwd/?token=&pwd=&dtoken=
func kunChangePassword(sup *model.SupplierFull, yid, newPwd string) (int, string, error) {
	baseURL := kunBuildBaseURL(sup)
	dtoken := kunGetDToken(sup)

	params := url.Values{}
	params.Set("token", yid)
	params.Set("pwd", newPwd)
	params.Set("dtoken", dtoken)

	apiURL := baseURL + "/upPwd/?" + params.Encode()

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return -1, "", fmt.Errorf("请求KUN改密失败：%v", err)
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

// kunPauseOrder KUN/kunba 平台暂停：GET /uporder/?token=&state=暂停&dtoken=
func kunPauseOrder(sup *model.SupplierFull, yid string) (int, string, error) {
	baseURL := kunBuildBaseURL(sup)
	dtoken := kunGetDToken(sup)

	params := url.Values{}
	params.Set("token", yid)
	params.Set("state", "暂停")
	params.Set("dtoken", dtoken)

	apiURL := baseURL + "/uporder/?" + params.Encode()

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return -1, "", fmt.Errorf("请求KUN暂停失败：%v", err)
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
