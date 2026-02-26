package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// xxtCallQuery 学习通查课（支持学号和手机号登录）
// 对应 xxtgf.php 的逻辑
func xxtCallQuery(account, password, school string) (map[string]interface{}, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	isPhone, _ := regexp.MatchString(`^1[3-9]\d{9}$`, account)

	var cookie string
	var courseResult []byte

	if isPhone {
		// 手机号登录
		loginURL := "https://passport2-api.chaoxing.com/v11/loginregister?cx_xxt_passport=json"
		form := url.Values{}
		form.Set("uname", account)
		form.Set("code", password)
		form.Set("loginType", "1")
		form.Set("roleSelect", "true")

		resp, err := client.PostForm(loginURL, form)
		if err != nil {
			return nil, fmt.Errorf("登录请求失败: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var loginResp map[string]interface{}
		json.Unmarshal(body, &loginResp)

		if status, _ := loginResp["status"].(bool); !status {
			return map[string]interface{}{
				"code": -1,
				"msg":  "信息错误或者重试",
			}, nil
		}

		// 提取cookie
		cookie = extractCookies(resp)
	} else {
		// 学号登录 - 先查学校fid
		schoolEncoded := url.QueryEscape(school)
		schoolURL := fmt.Sprintf("https://passport2-api.chaoxing.com/org/searchUnis?filter=%s&product=1&type=", schoolEncoded)

		resp, err := client.Get(schoolURL)
		if err != nil {
			return nil, fmt.Errorf("查询学校失败: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var schoolResp map[string]interface{}
		json.Unmarshal(body, &schoolResp)

		froms, _ := schoolResp["froms"].([]interface{})
		if len(froms) == 0 {
			return map[string]interface{}{
				"code": -1,
				"msg":  "未找到学校信息",
			}, nil
		}
		firstSchool := froms[0].(map[string]interface{})
		fid := toString(firstSchool["schoolid"])

		// 学号登录
		loginURL := fmt.Sprintf("https://passport2-api.chaoxing.com/v6/idNumberLogin?fid=%s&idNumber=%s", fid, account)
		form := url.Values{}
		form.Set("pwd", password)
		form.Set("t", "0")

		resp2, err := client.PostForm(loginURL, form)
		if err != nil {
			return nil, fmt.Errorf("登录请求失败: %v", err)
		}
		defer resp2.Body.Close()

		body2, _ := io.ReadAll(resp2.Body)
		var loginResp map[string]interface{}
		json.Unmarshal(body2, &loginResp)

		if status, _ := loginResp["status"].(bool); !status {
			return map[string]interface{}{
				"code": -1,
				"msg":  "信息错误或者重试",
			}, nil
		}

		cookie = extractCookies(resp2)
	}

	// 获取课程列表
	req, _ := http.NewRequest("GET", "https://mooc1-api.chaoxing.com/mycourse/", nil)
	req.Header.Set("Cookie", cookie)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取课程失败: %v", err)
	}
	defer resp.Body.Close()

	courseResult, _ = io.ReadAll(resp.Body)
	var courseResp map[string]interface{}
	json.Unmarshal(courseResult, &courseResp)

	resultVal, _ := courseResp["result"].(float64)
	if int(resultVal) != 1 {
		return map[string]interface{}{
			"code": -1,
			"msg":  "信息错误或者重试",
		}, nil
	}

	channelList, _ := courseResp["channelList"].([]interface{})
	var courses []map[string]interface{}
	for _, ch := range channelList {
		channel := ch.(map[string]interface{})
		content, _ := channel["content"].(map[string]interface{})
		if content == nil {
			continue
		}
		courseObj, _ := content["course"].(map[string]interface{})
		if courseObj == nil {
			continue
		}
		dataArr, _ := courseObj["data"].([]interface{})
		if len(dataArr) == 0 {
			continue
		}
		firstCourse := dataArr[0].(map[string]interface{})
		courses = append(courses, map[string]interface{}{
			"id":       toString(firstCourse["id"]),
			"name":     toString(firstCourse["name"]),
			"imageurl": toString(firstCourse["imageurl"]),
		})
	}

	return map[string]interface{}{
		"code":     1,
		"msg":      "查询成功",
		"userName": "",
		"userinfo": school + " " + account + " " + password,
		"data":     courses,
	}, nil
}

// extractCookies 从HTTP响应中提取Set-Cookie
func extractCookies(resp *http.Response) string {
	var cookies []string
	for _, c := range resp.Cookies() {
		cookies = append(cookies, c.Name+"="+c.Value)
	}
	return strings.Join(cookies, ";")
}
