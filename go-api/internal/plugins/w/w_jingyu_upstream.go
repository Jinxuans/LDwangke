package w

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strings"

	"go-api/internal/database"
)

// jingyuRequest POST form-urlencoded 到 jingyu 风格上游, 返回解析后 JSON。
func (s *WService) jingyuRequest(app map[string]interface{}, act string, params map[string]string) (map[string]interface{}, error) {
	pURL := strings.TrimSpace(fmt.Sprintf("%v", app["url"]))
	code := fmt.Sprintf("%v", app["code"])
	key := fmt.Sprintf("%v", app["key"])
	uid := fmt.Sprintf("%v", app["uid"])

	reqURL := fmt.Sprintf("%s?appId=%s&act=%s", pURL, url.QueryEscape(code), url.QueryEscape(act))

	if params == nil {
		params = map[string]string{}
	}
	params["uid"] = uid
	params["key"] = key

	resp, err := httpPostForm(reqURL, params, 60)
	if err != nil {
		return nil, fmt.Errorf("上游请求失败: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("上游响应解析失败")
	}
	return result, nil
}

// jingyuRequestRaw 同上但返回原始字节。
func (s *WService) jingyuRequestRaw(app map[string]interface{}, act string, params map[string]string) ([]byte, error) {
	pURL := strings.TrimSpace(fmt.Sprintf("%v", app["url"]))
	code := fmt.Sprintf("%v", app["code"])
	key := fmt.Sprintf("%v", app["key"])
	uid := fmt.Sprintf("%v", app["uid"])

	reqURL := fmt.Sprintf("%s?appId=%s&act=%s", pURL, url.QueryEscape(code), url.QueryEscape(act))

	if params == nil {
		params = map[string]string{}
	}
	params["uid"] = uid
	params["key"] = key

	return httpPostForm(reqURL, params, 60)
}

// GetPrice 获取本地用户价格 (base_price × addprice), 对应 PHP get_price。
func (s *WService) GetPrice(appID int64, uid int) ([]byte, error) {
	app, err := s.getAppRow(appID)
	if err != nil {
		return nil, err
	}
	appPrice := 0.0
	if p, ok := app["price"].(string); ok {
		fmt.Sscanf(p, "%f", &appPrice)
	}
	var addprice float64
	database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
	price := math.Round(appPrice*addprice*100) / 100
	resp, _ := json.Marshal(map[string]interface{}{"code": 1, "data": price})
	return resp, nil
}

// ProxyAction 通用代理：将请求转发到上游 (支持 type 0/1/2)。
func (s *WService) ProxyAction(appID int64, act string, data map[string]interface{}) ([]byte, error) {
	app, err := s.getAppRow(appID)
	if err != nil {
		return nil, err
	}

	pType := fmt.Sprintf("%v", app["type"])

	if act == "get_price" && pType == "2" {
		uid := 0
		if v, ok := data["login_uid"]; ok {
			fmt.Sscanf(fmt.Sprintf("%v", v), "%d", &uid)
		}
		return s.GetPrice(appID, uid)
	}

	if pType == "2" {
		params := map[string]string{}
		for k, v := range data {
			if k == "form" {
				if formMap, ok := v.(map[string]interface{}); ok {
					for fk, fv := range flattenFormData(formMap, "form") {
						params[fk] = fv
					}
				}
			} else {
				params[k] = fmt.Sprintf("%v", v)
			}
		}
		return s.jingyuRequestRaw(app, act, params)
	}

	result, err := s.appRequest(app, "/"+act, data, "POST")
	if err != nil {
		return nil, err
	}
	respJSON, _ := json.Marshal(result)
	return respJSON, nil
}
