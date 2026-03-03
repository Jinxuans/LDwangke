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

// ── 至强 (simple) 平台适配器 ──
// API 文档：
//   查课: POST http://{url}/Api/Get     params: token, school, user, pass, platform
//   下单: POST http://{url}/Api/Create  params: token, platform, school, user, pass, course, courseid
//   进度: POST http://{url}/Api/Query   params: token, school, user, pass, course, courseid, cid
//   补刷: POST http://{url}/Api/Repeat  params: token, id
//   批量: POST http://{url}/Api/Thread  params: page, token, timestamp

// simpleBuildBaseURL 构建 simple 平台基础 URL
func simpleBuildBaseURL(sup *model.SupplierFull) string {
	baseURL := strings.TrimRight(sup.URL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	return baseURL
}

// simpleGetToken 获取 simple 平台的 token（优先 Token 字段，其次 Pass 字段）
func simpleGetToken(sup *model.SupplierFull) string {
	if sup.Token != "" {
		return sup.Token
	}
	return sup.Pass
}

// simpleCallQuery 至强平台查课：POST /Api/Get
// 响应: {code:1, msg:"查询成功", student:"姓名", children:[{id, real_course, label, children:[...]}]}
func simpleCallQuery(sup *model.SupplierFull, platform, school, user, pass string) (*model.SupplierQueryResult, error) {
	baseURL := simpleBuildBaseURL(sup)
	token := simpleGetToken(sup)

	formData := url.Values{}
	formData.Set("token", token)
	formData.Set("school", school)
	formData.Set("user", user)
	formData.Set("pass", pass)
	formData.Set("platform", platform)

	apiURL := baseURL + "/Api/Get"

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.PostForm(apiURL, formData)
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

	codeVal := fmt.Sprintf("%v", raw["code"])
	if codeVal != "1" {
		msg := toString(raw["msg"])
		if msg == "" {
			msg = "查课失败"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	userName := toString(raw["student"])

	// 解析 children 结构：可能有嵌套 children
	var items []model.CourseItem
	if children, ok := raw["children"].([]interface{}); ok {
		for _, child := range children {
			row, ok := child.(map[string]interface{})
			if !ok {
				continue
			}
			// 检查是否有子 children
			if subChildren, ok := row["children"].([]interface{}); ok && len(subChildren) > 0 {
				for _, sub := range subChildren {
					if m, ok := sub.(map[string]interface{}); ok {
						items = append(items, model.CourseItem{
							ID:   toString(m["id"]),
							Name: toString(m["real_course"]),
						})
					}
				}
			} else {
				// 无子节点，直接用当前行
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
	}

	return &model.SupplierQueryResult{
		Msg:      "查询成功",
		UserName: userName,
		Data:     items,
	}, nil
}

// simpleCallOrder 至强平台下单：POST /Api/Create
// 响应: {code:1, msg:"添加成功"}
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

	apiURL := baseURL + "/Api/Create"

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.PostForm(apiURL, formData)
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
	result := &model.SupplierOrderResult{
		Msg: toString(raw["msg"]),
	}

	fmt.Printf("[simpleCallOrder] codeVal=%s raw=%v\n", codeVal, raw)

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

// simpleQueryProgress 至强平台进度查询：POST /Api/Query
// 响应: {code:1, data:{id, user, pass, kcks, kcjs, ksks, ksjs, course, status, progress, process}}
func simpleQueryProgress(sup *model.SupplierFull, platform, school, user, pass, kcname, kcid string) ([]model.SupplierProgressItem, error) {
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

	apiURL := baseURL + "/Api/Query"

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.PostForm(apiURL, formData)
	if err != nil {
		return nil, fmt.Errorf("请求至强进度失败：%v", err)
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
	if codeVal != "1" {
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

	// 映射上游状态（按 PHP 逻辑）
	status := toString(data["status"])
	switch status {
	case "已暂停", "明天继续", "待处理", "考试中", "平时分", "待补单":
		status = "进行中"
	}

	remarks := toString(data["process"])
	if remarks == "" {
		remarks = "暂无详情"
	}

	progress := toString(data["progress"])

	item := model.SupplierProgressItem{
		YID:             toString(data["id"]),
		KCName:          toString(data["course"]),
		User:            toString(data["user"]),
		Status:          status,
		StatusText:      status,
		Process:         progress,
		Remarks:         remarks,
		CourseStartTime: toString(data["kcks"]),
		CourseEndTime:   toString(data["kcjs"]),
		ExamStartTime:   toString(data["ksks"]),
		ExamEndTime:     toString(data["ksjs"]),
	}

	return []model.SupplierProgressItem{item}, nil
}

// simpleResubmit 至强平台补刷：POST /Api/Repeat
// 请求: {token, id}  响应: {code:1, msg:"操作成功！"}
func simpleResubmit(sup *model.SupplierFull, yid string) (int, string, error) {
	baseURL := simpleBuildBaseURL(sup)
	token := simpleGetToken(sup)

	formData := url.Values{}
	formData.Set("token", token)
	formData.Set("id", yid)

	apiURL := baseURL + "/Api/Repeat"

	if host := extractHost(sup.URL); host != "" {
		globalRateLimiter.wait(host, 500*time.Millisecond)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.PostForm(apiURL, formData)
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
