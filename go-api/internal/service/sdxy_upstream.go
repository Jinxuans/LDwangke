package service

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"

	"go-api/internal/database"
)

func (s *SDXYService) upstreamRequest(act string, params map[string]string) (map[string]interface{}, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return nil, fmt.Errorf("闪电运动未配置上游接口")
	}
	if params == nil {
		params = map[string]string{}
	}
	params["login_uid"] = cfg.UID
	params["login_key"] = cfg.Key

	apiURL := cfg.BaseURL + cfg.Endpoint + "?act=" + url.QueryEscape(act)

	resp, err := httpPostForm(apiURL, params, cfg.Timeout)
	if err != nil {
		return nil, fmt.Errorf("上游请求失败: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("上游响应解析失败")
	}
	return result, nil
}

func (s *SDXYService) upstreamRaw(act string, params map[string]string) ([]byte, error) {
	cfg, err := s.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		return nil, fmt.Errorf("闪电运动未配置上游接口")
	}
	if params == nil {
		params = map[string]string{}
	}
	params["login_uid"] = cfg.UID
	params["login_key"] = cfg.Key

	apiURL := cfg.BaseURL + cfg.Endpoint + "?act=" + url.QueryEscape(act)
	return httpPostForm(apiURL, params, cfg.Timeout)
}

func (s *SDXYService) GetPrice(uid int) (float64, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return 0, err
	}
	var addprice float64 = 1.0
	database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
	if addprice <= 0 {
		addprice = 1.0
	}
	price := cfg.Price * addprice
	price = math.Round(price*10000) / 10000
	return price, nil
}
