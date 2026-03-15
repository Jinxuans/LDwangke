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

func simpleBuildBaseURL(sup *model.SupplierFull) string {
	baseURL := strings.TrimRight(sup.URL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	return baseURL
}

func simpleGetToken(sup *model.SupplierFull) string {
	if sup.Token != "" {
		return sup.Token
	}
	return sup.Pass
}

func simpleCallQuery(sup *model.SupplierFull, platform, school, user, pass string) (*model.SupplierQueryResult, error) {
	baseURL := simpleBuildBaseURL(sup)
	token := simpleGetToken(sup)

	formData := url.Values{}
	formData.Set("token", token)
	formData.Set("school", school)
	formData.Set("user", user)
	formData.Set("pass", pass)
	formData.Set("platform", platform)

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.PostForm(baseURL+"/Api/Get", formData)
	if err != nil {
		return nil, fmt.Errorf("请求至强查课失败：%v", err)
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

	if codeVal := fmt.Sprintf("%v", raw["code"]); codeVal != "1" {
		msg := toString(raw["msg"])
		if msg == "" {
			msg = "查课失败"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	userName := toString(raw["student"])
	var items []model.CourseItem
	if children, ok := raw["children"].([]interface{}); ok {
		for _, child := range children {
			row, ok := child.(map[string]interface{})
			if !ok {
				continue
			}
			if subChildren, ok := row["children"].([]interface{}); ok && len(subChildren) > 0 {
				for _, sub := range subChildren {
					if m, ok := sub.(map[string]interface{}); ok {
						items = append(items, model.CourseItem{
							ID:   toString(m["id"]),
							Name: toString(m["real_course"]),
						})
					}
				}
				continue
			}
			name := toString(row["real_course"])
			if name == "" {
				name = toString(row["label"])
			}
			items = append(items, model.CourseItem{
				ID:   toString(row["id"]),
				Name: name,
			})
		}
	}

	return &model.SupplierQueryResult{
		Msg:      "查询成功",
		UserName: userName,
		Data:     items,
	}, nil
}

func simpleCallOrder(sup *model.SupplierFull, platform, school, user, pass, kcname, kcid string) (*model.SupplierOrderResult, error) {
	baseURL := simpleBuildBaseURL(sup)
	token := simpleGetToken(sup)

	formData := url.Values{}
	formData.Set("token", token)
	formData.Set("platform", platform)
	formData.Set("school", school)
	formData.Set("user", user)
	formData.Set("pass", pass)
	formData.Set("course", kcname)
	formData.Set("courseid", kcid)

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.PostForm(baseURL+"/Api/Create", formData)
	if err != nil {
		return nil, fmt.Errorf("请求至强下单失败：%v", err)
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
	result := &model.SupplierOrderResult{Msg: toString(raw["msg"])}
	if codeVal == "1" {
		result.Code = 1
		if result.Msg == "" {
			result.Msg = "添加成功"
		}
	} else {
		result.Code = -1
		if result.Msg == "" {
			result.Msg = fmt.Sprintf("上游返回错误码：%s", codeVal)
		}
	}
	return result, nil
}

func simpleQueryProgress(sup *model.SupplierFull, platform, school, user, pass, kcname, kcid string, debugInfo progressDebugInfo) ([]model.SupplierProgressItem, error) {
	baseURL := simpleBuildBaseURL(sup)
	token := simpleGetToken(sup)

	formData := url.Values{}
	formData.Set("token", token)
	formData.Set("school", school)
	formData.Set("user", user)
	formData.Set("pass", pass)
	formData.Set("course", kcname)
	formData.Set("courseid", kcid)
	formData.Set("cid", platform)

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	debugInfo.logRequest(sup, "POST", baseURL+"/Api/Query", "application/x-www-form-urlencoded", formData.Encode())
	resp, err := client.PostForm(baseURL+"/Api/Query", formData)
	if err != nil {
		debugInfo.logRequestError(sup, err)
		return nil, fmt.Errorf("请求至强进度失败：%v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败：%v", err)
	}
	debugInfo.logResponse(sup, resp.Status, body)

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析响应失败：%s", string(body))
	}

	if codeVal := fmt.Sprintf("%v", raw["code"]); codeVal != "1" {
		msg := toString(raw["msg"])
		if msg == "" {
			msg = "进度服务器出小差了，待会试试呗~"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	data, ok := raw["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("进度数据格式异常")
	}

	status := toString(data["status"])
	switch status {
	case "已暂停", "明天继续", "待处理", "考试中", "平时分", "待补单":
		status = "进行中"
	}

	remarks := toString(data["process"])
	if remarks == "" {
		remarks = "暂无详情"
	}

	item := model.SupplierProgressItem{
		YID:             toString(data["id"]),
		KCName:          toString(data["course"]),
		User:            toString(data["user"]),
		Status:          status,
		StatusText:      status,
		Process:         toString(data["progress"]),
		Remarks:         remarks,
		CourseStartTime: toString(data["kcks"]),
		CourseEndTime:   toString(data["kcjs"]),
		ExamStartTime:   toString(data["ksks"]),
		ExamEndTime:     toString(data["ksjs"]),
	}
	return []model.SupplierProgressItem{item}, nil
}

func simpleResubmit(sup *model.SupplierFull, yid string) (int, string, error) {
	baseURL := simpleBuildBaseURL(sup)
	token := simpleGetToken(sup)

	formData := url.Values{}
	formData.Set("token", token)
	formData.Set("id", yid)

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.PostForm(baseURL+"/Api/Repeat", formData)
	if err != nil {
		return -1, "", fmt.Errorf("请求至强补刷失败：%v", err)
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
