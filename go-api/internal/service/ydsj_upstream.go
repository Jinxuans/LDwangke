package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func ydsjIsConfigured(cfg *YDSJConfig) bool {
	return cfg.BaseURL != "" && cfg.UID != "" && cfg.Key != ""
}

func (s *YDSJService) ydsjRequest(act string, params map[string]string) ([]byte, error) {
	cfg, err := s.GetConfig()
	if err != nil || !ydsjIsConfigured(cfg) {
		return nil, fmt.Errorf("运动世界未配置上游接口")
	}
	return s.ydsjRequestWithCfg(cfg, act, params)
}

func (s *YDSJService) ydsjRequestWithCfg(cfg *YDSJConfig, act string, params map[string]string) ([]byte, error) {
	if params == nil {
		params = map[string]string{}
	}
	params["login_uid"] = cfg.UID
	params["login_key"] = cfg.Key

	apiURL := strings.TrimRight(cfg.BaseURL, "/") + "/ydsj/api.php?act=" + url.QueryEscape(act)

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("上游请求失败: %v", err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (s *YDSJService) GetSchools() ([]map[string]interface{}, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return nil, err
	}

	if ydsjIsConfigured(cfg) {
		respBody, err := s.ydsjRequestWithCfg(cfg, "get_school", nil)
		if err == nil {
			var result map[string]interface{}
			if json.Unmarshal(respBody, &result) == nil {
				if dataRaw, ok := result["data"].([]interface{}); ok && len(dataRaw) > 0 {
					var schools []map[string]interface{}
					for _, d := range dataRaw {
						if m, ok := d.(map[string]interface{}); ok {
							schools = append(schools, m)
						}
					}
					if len(schools) > 0 {
						return schools, nil
					}
				}
			}
		} else {
			log.Printf("[YDSJ] 上游学校列表请求失败: %v，回退本地", err)
		}
	}

	var wrapper struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(ydsjSchoolsJSON, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Data, nil
}
