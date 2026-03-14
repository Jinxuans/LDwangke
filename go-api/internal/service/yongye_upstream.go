package service

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/url"
	"strings"

	"go-api/internal/database"
)

func (s *YongyeService) yongyePostForm(apiURL string, params map[string]string) ([]byte, error) {
	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	resp, err := s.client.PostForm(apiURL, form)
	if err != nil {
		return nil, fmt.Errorf("上游请求失败: %v", err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (s *YongyeService) yongyeUpstreamPost(cfg *YongyeConfig, endpoint string, extra map[string]string) ([]byte, error) {
	if cfg.ApiURL == "" || cfg.Token == "" {
		return nil, fmt.Errorf("永夜运动未配置上游接口")
	}
	params := map[string]string{"token": cfg.Token}
	for k, v := range extra {
		params[k] = v
	}
	apiURL := strings.TrimRight(cfg.ApiURL, "/") + "/" + strings.TrimLeft(endpoint, "/")
	return s.yongyePostForm(apiURL, params)
}

func (s *YongyeService) GetSchools(uid int) (interface{}, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return nil, err
	}

	respBody, err := s.yongyeUpstreamPost(cfg, "school", nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析学校列表失败")
	}

	code := mapGetFloat(result, "code")
	if int(code) != 1 {
		msg := mapGetString(result, "msg")
		if msg == "" {
			msg = "获取学校列表失败"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	var addprice float64
	database.DB.QueryRow("SELECT COALESCE(addprice, 0) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)

	djfl := 1.0
	if cfg.Dj > 0 && addprice > 0 {
		djfl = math.Round(addprice/cfg.Dj*100) / 100
	}

	if dataRaw, ok := result["data"].([]interface{}); ok {
		var schools []map[string]interface{}
		for _, d := range dataRaw {
			if school, ok := d.(map[string]interface{}); ok {
				name := mapGetString(school, "name")
				if name != "自动识别" {
					cpmuch := mapGetFloat(school, "cpmuch")
					zcmuch := mapGetFloat(school, "zcmuch")
					if cfg.Zs > 0 {
						cpmuch = math.Round(cpmuch/cfg.Zs*cfg.Beis*100) / 100
						zcmuch = math.Round(zcmuch/cfg.Zs*cfg.Beis*100) / 100
					}
					cpmuch = math.Round(cpmuch*djfl*100) / 100
					zcmuch = math.Round(zcmuch*djfl*100) / 100
					school["cpmuch"] = cpmuch
					school["zcmuch"] = zcmuch
				}
				schools = append(schools, school)
			}
		}
		result["data"] = schools
	}

	return result, nil
}
