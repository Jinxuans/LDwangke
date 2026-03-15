package supplier

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"go-api/internal/database"
)

func (s *Service) QueryBalance(hid int) (map[string]interface{}, error) {
	sup, err := s.GetSupplierByHID(hid)
	if err != nil {
		return nil, err
	}

	if sup.PT == "yyy" {
		result, err := yyyQueryBalance(sup)
		if err != nil {
			return nil, err
		}
		if money, ok := result["money"].(string); ok && money != "" && money != "<nil>" {
			database.DB.Exec("UPDATE qingka_wangke_huoyuan SET money = ? WHERE hid = ?", money, hid)
		}
		return result, nil
	}
	if sup.PT == "tuboshu" {
		apiURL := tuboshuBalanceAPIURL(sup.URL)
		req, reqErr := http.NewRequest("GET", apiURL, nil)
		if reqErr != nil {
			return nil, fmt.Errorf("构建请求失败：%v", reqErr)
		}
		req.Header.Set("Authorization", "Bearer "+sup.Token)
		req.Header.Set("Accept", "application/json")

		resp, err := s.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("请求土拨鼠余额接口失败：%v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("读取响应失败：%v", err)
		}

		var raw map[string]interface{}
		if err := json.Unmarshal(body, &raw); err != nil {
			return nil, fmt.Errorf("解析土拨鼠响应失败：%s", string(body))
		}

		money := extractTuboshuPoint(raw)
		database.DB.Exec("UPDATE qingka_wangke_huoyuan SET money = ? WHERE hid = ?", money, hid)
		return map[string]interface{}{
			"code":  200,
			"money": money,
			"pt":    sup.PT,
			"name":  sup.Name,
			"hid":   hid,
			"raw":   raw,
		}, nil
	}

	cfg := GetPlatformConfig(sup.PT)
	apiURL := resolveConfiguredActionURL(sup.URL, cfg.BalancePath)

	authType := cfg.BalanceAuthType
	if authType == "" {
		authType = cfg.AuthType
	}

	if strings.TrimSpace(cfg.BalanceParamMap) != "" {
		defaultParams := defaultSupplierAuthParams(sup, authType)
		execResult, err := s.executeConfiguredAction(
			sup,
			apiURL,
			cfg.BalanceMethod,
			cfg.BalanceBodyType,
			cfg.BalanceParamMap,
			http.MethodPost,
			"form",
			defaultParams,
			map[string]string{},
		)
		if err != nil {
			return nil, fmt.Errorf("请求上游余额接口失败：%v", err)
		}

		var raw map[string]interface{}
		if err := json.Unmarshal(execResult.Body, &raw); err != nil {
			return nil, fmt.Errorf("解析响应失败：%s", string(execResult.Body))
		}

		money := extractMoneyField(raw, cfg.BalanceMoneyField)
		result := map[string]interface{}{
			"code":  200,
			"money": money,
			"pt":    sup.PT,
			"name":  sup.Name,
			"hid":   hid,
			"raw":   raw,
		}
		if money != "" && money != "<nil>" {
			database.DB.Exec("UPDATE qingka_wangke_huoyuan SET money = ? WHERE hid = ?", money, hid)
		}
		return result, nil
	}

	var resp *http.Response
	switch authType {
	case "bearer_token":
		req, reqErr := http.NewRequest("GET", apiURL, nil)
		if reqErr != nil {
			return nil, fmt.Errorf("构建请求失败：%v", reqErr)
		}
		req.Header.Set("Authorization", "Bearer "+sup.Token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "*/*")
		if sup.Cookie != "" {
			cookieStr := sup.Cookie
			if !strings.Contains(cookieStr, "=") {
				cookieStr = "session_id=" + cookieStr
			}
			req.Header.Set("Cookie", cookieStr)
		}
		resp, err = s.client.Do(req)
	case "token_only":
		if cfg.UseJSON {
			jsonData, _ := json.Marshal(map[string]string{"token": getSupplierToken(sup)})
			req, _ := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			resp, err = s.client.Do(req)
		} else {
			form := url.Values{}
			form.Set("token", getSupplierToken(sup))
			resp, err = s.client.PostForm(apiURL, form)
		}
	default:
		form := url.Values{}
		form.Set("uid", sup.User)
		form.Set("key", sup.Pass)
		resp, err = s.client.PostForm(apiURL, form)
	}
	if err != nil {
		return nil, fmt.Errorf("请求上游余额接口失败：%v", err)
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

	money := extractMoneyField(raw, cfg.BalanceMoneyField)
	result := map[string]interface{}{
		"code":  200,
		"money": money,
		"pt":    sup.PT,
		"name":  sup.Name,
		"hid":   hid,
		"raw":   raw,
	}
	if money != "" && money != "<nil>" {
		database.DB.Exec("UPDATE qingka_wangke_huoyuan SET money = ? WHERE hid = ?", money, hid)
	}
	return result, nil
}

func tuboshuBalanceAPIURL(rawURL string) string {
	baseURL := strings.TrimRight(rawURL, "/")
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "https://" + baseURL
	}
	if !strings.HasSuffix(baseURL, "/api") {
		baseURL = baseURL + "/api"
	}
	return baseURL + "/userInfo"
}

func extractTuboshuPoint(raw map[string]interface{}) string {
	if data, ok := raw["data"].(map[string]interface{}); ok {
		if point, ok := data["point"]; ok {
			return fmt.Sprintf("%v", point)
		}
	}
	return "0"
}

func extractMoneyField(raw map[string]interface{}, fieldPath string) string {
	parts := strings.Split(fieldPath, ".")
	var current interface{} = raw

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return fmt.Sprintf("%v", current)
		}
	}
	return fmt.Sprintf("%v", current)
}
