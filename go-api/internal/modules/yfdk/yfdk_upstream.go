package yfdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go-api/internal/database"
)

func (s *YFDKService) upstreamRequest(method, url string, body interface{}, token string) (map[string]interface{}, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, _ := json.Marshal(body)
		reqBody = strings.NewReader(string(jsonData))
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("上游请求失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("上游返回解析失败")
	}
	return result, nil
}

func yfdkLog(uid int, logType, text string, money float64) {
	now := time.Now().Format("2006-01-02 15:04:05")
	var newBalance float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&newBalance)
	moneyStr := fmt.Sprintf("%.2f", money)
	if money < 0 {
		moneyStr = fmt.Sprintf("%.2f", money)
	} else if money > 0 {
		moneyStr = fmt.Sprintf("+%.2f", money)
	}
	database.DB.Exec(
		"INSERT INTO qingka_wangke_log (uid, type, text, money, smoney, ip, addtime) VALUES (?, ?, ?, ?, ?, '', ?)",
		uid, logType, text, moneyStr, fmt.Sprintf("%.2f", newBalance), now,
	)
}

func (s *YFDKService) GetProjects() (interface{}, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return nil, fmt.Errorf("YF打卡未配置")
	}
	result, err := s.upstreamRequest("GET", strings.TrimRight(cfg.BaseURL, "/")+"/projects", nil, cfg.Token)
	if err != nil {
		return nil, err
	}
	code, _ := result["code"].(float64)
	if int(code) != 200 {
		return nil, fmt.Errorf("获取项目列表失败")
	}
	data, _ := result["data"].(map[string]interface{})
	return data["projects"], nil
}

func (s *YFDKService) GetAccountInfo(cid, school, username, password, yzmCode string) (interface{}, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return nil, fmt.Errorf("YF打卡未配置")
	}
	postData := map[string]interface{}{
		"cid":      cid,
		"school":   school,
		"username": username,
		"password": password,
	}
	if yzmCode != "" {
		postData["verification_code"] = yzmCode
	}
	result, err := s.upstreamRequest("POST", strings.TrimRight(cfg.BaseURL, "/")+"/account/info", postData, cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("源台返回数据解析失败")
	}
	code, _ := result["code"].(float64)
	if int(code) != 200 {
		msg := "获取账号信息失败"
		if m, ok := result["message"].(string); ok && m != "" {
			msg = m
		} else if m, ok := result["msg"].(string); ok && m != "" {
			msg = m
		}
		return nil, fmt.Errorf("%s", msg)
	}
	data, _ := result["data"].(map[string]interface{})
	return data["account_info"], nil
}

func (s *YFDKService) GetSchools(cid string) (interface{}, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return nil, fmt.Errorf("YF打卡未配置")
	}
	url := fmt.Sprintf("%s/schools?cid=%s", strings.TrimRight(cfg.BaseURL, "/"), cid)
	result, err := s.upstreamRequest("GET", url, nil, cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("获取学校列表失败")
	}
	code, _ := result["code"].(float64)
	if int(code) != 200 {
		msg := "获取学校列表失败"
		if m, ok := result["message"].(string); ok && m != "" {
			msg = m
		}
		return nil, fmt.Errorf("%s", msg)
	}
	data, _ := result["data"].(map[string]interface{})
	return data["schools"], nil
}

func (s *YFDKService) SearchSchools(cid, keyword string) (interface{}, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return nil, fmt.Errorf("YF打卡未配置")
	}
	url := fmt.Sprintf("%s/schools/search?cid=%s&keyword=%s", strings.TrimRight(cfg.BaseURL, "/"), cid, keyword)
	result, err := s.upstreamRequest("GET", url, nil, cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("搜索学校失败")
	}
	code, _ := result["code"].(float64)
	if int(code) != 200 {
		msg := "搜索学校失败"
		if m, ok := result["message"].(string); ok && m != "" {
			msg = m
		}
		return nil, fmt.Errorf("%s", msg)
	}
	data, _ := result["data"].(map[string]interface{})
	return data["schools"], nil
}
