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

func kunBuildBaseURL(sup *model.SupplierFull) string {
	baseURL := strings.TrimRight(sup.URL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	return baseURL
}

func kunGetDToken(sup *model.SupplierFull) string {
	if sup.Token != "" {
		return sup.Token
	}
	return sup.Pass
}

func kunCallQuery(sup *model.SupplierFull, platform, school, user, pass string) (*model.SupplierQueryResult, error) {
	baseURL := kunBuildBaseURL(sup)
	dtoken := kunGetDToken(sup)

	params := url.Values{}
	params.Set("platform", platform)
	params.Set("school", school)
	params.Set("account", user)
	params.Set("password", pass)
	params.Set("dtoken", dtoken)

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(baseURL + "/query/?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("请求KUN查课失败：%v", err)
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

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(baseURL + "/getorder/?" + params.Encode())
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

	if codeVal == "0" || codeVal == "200" {
		result.Code = 1
		if result.Msg == "" || result.Msg == "<nil>" {
			result.Msg = "下单成功"
		}
		if token, ok := raw["token"].(string); ok && token != "" {
			result.YID = token
		}
		if result.YID == "" {
			if idVal, ok := raw["id"]; ok && idVal != nil {
				result.YID = fmt.Sprintf("%v", idVal)
			}
		}
		if result.YID == "" {
			if dataArr, ok := raw["data"].([]interface{}); ok && len(dataArr) > 0 {
				result.YID = fmt.Sprintf("%v", dataArr[0])
			}
		}
	} else {
		result.Code = -1
		if result.Msg == "" || result.Msg == "<nil>" {
			result.Msg = fmt.Sprintf("上游返回错误码：%s", codeVal)
		}
	}

	return result, nil
}

func kunChangePassword(sup *model.SupplierFull, yid, newPwd string) (int, string, error) {
	baseURL := kunBuildBaseURL(sup)
	dtoken := kunGetDToken(sup)

	params := url.Values{}
	params.Set("token", yid)
	params.Set("pwd", newPwd)
	params.Set("dtoken", dtoken)

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(baseURL + "/upPwd/?" + params.Encode())
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

func kunPauseOrder(sup *model.SupplierFull, yid string) (int, string, error) {
	baseURL := kunBuildBaseURL(sup)
	dtoken := kunGetDToken(sup)

	params := url.Values{}
	params.Set("token", yid)
	params.Set("state", "暂停")
	params.Set("dtoken", dtoken)

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(baseURL + "/uporder/?" + params.Encode())
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
