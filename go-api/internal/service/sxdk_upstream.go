package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (s *SXDKService) upstreamPost(cfg *SXDKConfig, endpoint string, data map[string]interface{}) (map[string]interface{}, error) {
	if data == nil {
		data = map[string]interface{}{}
	}
	data["admin"] = cfg.Admin
	data["token"] = cfg.Token

	jsonData, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", strings.TrimRight(cfg.BaseURL, "/")+"/"+endpoint, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("上游请求失败: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("上游返回解析失败")
	}
	return result, nil
}
