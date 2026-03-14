package service

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-api/internal/database"
)

func (s *TuZhiService) login(cfg *TuZhiConfig) (string, error) {
	if cfg.Username == "" || cfg.Password == "" {
		return "", fmt.Errorf("凸知打卡未配置账号密码")
	}
	body := map[string]interface{}{
		"terminal": "2",
		"account":  cfg.Username,
		"password": cfg.Password,
	}
	jsonData, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", s.baseURL+"/user/login/account", strings.NewReader(string(jsonData)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("登录请求失败: %v", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)
	code, _ := result["code"].(float64)
	if code != 200 {
		msg, _ := result["msg"].(string)
		return "", fmt.Errorf("登录失败: %s", msg)
	}
	data, _ := result["data"].(map[string]interface{})
	token, _ := data["token"].(string)
	if token == "" {
		return "", fmt.Errorf("登录返回无token")
	}
	return token, nil
}

func (s *TuZhiService) upstreamRequest(method, path string, token string, data interface{}) (map[string]interface{}, error) {
	var req *http.Request
	var err error

	if method == "GET" {
		url := s.baseURL + path
		if data != nil {
			if params, ok := data.(map[string]interface{}); ok {
				parts := []string{}
				for k, v := range params {
					parts = append(parts, fmt.Sprintf("%s=%v", k, v))
				}
				url += "?" + strings.Join(parts, "&")
			}
		}
		req, err = http.NewRequest("GET", url, nil)
	} else {
		jsonData, _ := json.Marshal(data)
		req, err = http.NewRequest("POST", s.baseURL+path, strings.NewReader(string(jsonData)))
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("上游请求失败: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("上游返回解析失败: %s", string(body))
	}
	return result, nil
}

func (s *TuZhiService) GetGoods() ([]map[string]interface{}, error) {
	cfg, err := s.GetConfig()
	if err != nil {
		return nil, err
	}
	token, err := s.login(cfg)
	if err != nil {
		return nil, err
	}
	result, err := s.upstreamRequest("GET", "/user/mall/lists", token, nil)
	if err != nil {
		return nil, err
	}
	code, _ := result["code"].(float64)
	if code != 200 {
		msg, _ := result["msg"].(string)
		return nil, fmt.Errorf("获取商品失败: %s", msg)
	}
	data, _ := result["data"].(map[string]interface{})
	lists, _ := data["lists"].([]interface{})
	var goods []map[string]interface{}
	for _, l := range lists {
		group, _ := l.(map[string]interface{})
		items, _ := group["goods"].([]interface{})
		for _, item := range items {
			if g, ok := item.(map[string]interface{}); ok {
				goods = append(goods, g)
			}
		}
	}
	return goods, nil
}

func (s *TuZhiService) GetGoodsForUser(addprice float64) ([]map[string]interface{}, error) {
	goods, err := s.GetGoods()
	if err != nil {
		return nil, err
	}
	overrides, _ := s.GetGoodsOverrides()
	overrideMap := map[int]TuZhiGoodsOverride{}
	for _, o := range overrides {
		overrideMap[o.GoodsID] = o
	}

	var result []map[string]interface{}
	for _, g := range goods {
		gidF, _ := g["id"].(float64)
		gid := int(gidF)
		ov, hasOv := overrideMap[gid]
		if hasOv && ov.Enabled == 0 {
			continue
		}
		price, _ := g["price"].(float64)
		if hasOv && ov.Price > 0 {
			price = ov.Price
		}
		billingMethod := 1
		if bm, ok := g["billing_method"].(float64); ok {
			billingMethod = int(bm)
		}
		unit := "天"
		if billingMethod == 2 {
			unit = "月"
		}
		userPrice := math.Round(addprice*price*100) / 100
		name, _ := g["name"].(string)
		g["display_name"] = fmt.Sprintf("%s %.2f/%s", name, userPrice, unit)
		g["user_price"] = userPrice
		result = append(result, g)
	}
	return result, nil
}

func (s *TuZhiService) GetSchools(form map[string]interface{}) (interface{}, error) {
	cfg, _ := s.GetConfig()
	token, err := s.login(cfg)
	if err != nil {
		return nil, err
	}
	result, err := s.upstreamRequest("GET", "/user/finance.order/getSchool", token, form)
	if err != nil {
		return nil, err
	}
	code, _ := result["code"].(float64)
	if code != 200 {
		msg, _ := result["msg"].(string)
		return nil, fmt.Errorf("%s", msg)
	}
	return result["data"], nil
}

func (s *TuZhiService) GetSignInfo(uid int, form map[string]interface{}) (interface{}, error) {
	cfg, _ := s.GetConfig()
	token, err := s.login(cfg)
	if err != nil {
		return nil, err
	}

	goodsID := 0
	if gid, ok := form["goods_id"]; ok {
		switch v := gid.(type) {
		case float64:
			goodsID = int(v)
		case string:
			goodsID, _ = strconv.Atoi(v)
		}
	}
	username, _ := form["username"].(string)

	overrides, _ := s.GetGoodsOverrides()
	goods, err := s.GetGoods()
	if err == nil {
		for _, g := range goods {
			gidF, _ := g["id"].(float64)
			if int(gidF) == goodsID {
				bm, _ := g["billing_method"].(float64)
				if int(bm) == 2 {
					var count int
					database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_dakaaz WHERE user_id=? AND username=?", uid, username).Scan(&count)
					if count == 0 {
						var recCount int
						database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_daka_query_record WHERE user_id=? AND username=? AND is_success=1", uid, username).Scan(&recCount)
						if recCount == 0 {
							price, _ := g["price"].(float64)
							for _, ov := range overrides {
								if ov.GoodsID == goodsID && ov.Price > 0 {
									price = ov.Price
									break
								}
							}
							var addprice float64
							database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid=?", uid).Scan(&addprice)
							queryPrice := math.Round(addprice*price*100) / 100

							var money float64
							database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid=?", uid).Scan(&money)
							if money < queryPrice {
								return nil, fmt.Errorf("余额不足")
							}
							now := time.Now().Unix()
							database.DB.Exec("INSERT INTO qingka_wangke_daka_query_record (username, password, create_time, user_id, is_success, price, goods_id) VALUES (?, ?, ?, ?, 0, ?, ?)",
								username, form["password"], now, uid, queryPrice, goodsID)
							database.DB.Exec("UPDATE qingka_wangke_user SET money=money-? WHERE uid=?", queryPrice, uid)
							tuzhiLog(uid, "tuzhi-按月查询扣费", fmt.Sprintf("商品%d %s 扣%.2f", goodsID, username, queryPrice), -queryPrice)
						}
					}
				}
				break
			}
		}
	}

	result, err := s.upstreamRequest("POST", "/user/finance.order/detail", token, form)
	if err != nil {
		return nil, err
	}
	code, _ := result["code"].(float64)
	if code != 200 {
		msg, _ := result["msg"].(string)
		return nil, fmt.Errorf("%s", msg)
	}
	database.DB.Exec("UPDATE qingka_wangke_daka_query_record SET is_success=1 WHERE user_id=? AND username=? LIMIT 1", uid, username)
	return result["data"], nil
}

func (s *TuZhiService) CalculateDays(form map[string]interface{}) (interface{}, error) {
	cfg, _ := s.GetConfig()
	token, err := s.login(cfg)
	if err != nil {
		return nil, err
	}
	result, err := s.upstreamRequest("POST", "/user/finance.order/calculateDays", token, form)
	if err != nil {
		return nil, err
	}
	code, _ := result["code"].(float64)
	if code != 200 {
		msg, _ := result["msg"].(string)
		return nil, fmt.Errorf("%s", msg)
	}
	return result["data"], nil
}
