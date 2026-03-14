package push

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go-api/internal/database"
	commonmodule "go-api/internal/modules/common"
)

type Service struct{}

var pushService = NewService()

func NewService() *Service {
	return &Service{}
}

func (s *Service) UnbindPushUIDByAccount(account string) (int64, error) {
	result, err := database.DB.Exec(
		"UPDATE qingka_wangke_order SET pushUid='', pushStatus='0' WHERE user=?",
		account,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *Service) BatchBindPushUID(orderIDs []int, pushUID string) (int64, error) {
	if len(orderIDs) == 0 {
		return 0, nil
	}
	placeholders := make([]string, len(orderIDs))
	args := make([]interface{}, 0, len(orderIDs)+1)
	args = append(args, pushUID)
	for i, id := range orderIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}
	query := fmt.Sprintf(
		"UPDATE qingka_wangke_order SET pushUid=?, pushStatus='0' WHERE oid IN (%s)",
		strings.Join(placeholders, ","),
	)
	result, err := database.DB.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *Service) BindPushEmail(orderID int, account string, email string) (int64, error) {
	if orderID > 0 {
		result, err := database.DB.Exec(
			"UPDATE qingka_wangke_order SET pushEmail=?, pushEmailStatus='0' WHERE oid=?",
			email, orderID,
		)
		if err != nil {
			return 0, err
		}
		return result.RowsAffected()
	}
	if account != "" {
		result, err := database.DB.Exec(
			"UPDATE qingka_wangke_order SET pushEmail=?, pushEmailStatus='0' WHERE user=?",
			email, account,
		)
		if err != nil {
			return 0, err
		}
		return result.RowsAffected()
	}
	return 0, fmt.Errorf("需要订单ID或账号")
}

func (s *Service) BindShowDocPush(orderID int, account string, showdocURL string) (int64, error) {
	if orderID > 0 {
		result, err := database.DB.Exec(
			"UPDATE qingka_wangke_order SET showdoc_push_url=?, pushShowdocStatus='0' WHERE oid=?",
			showdocURL, orderID,
		)
		if err != nil {
			return 0, err
		}
		return result.RowsAffected()
	}
	if account != "" {
		result, err := database.DB.Exec(
			"UPDATE qingka_wangke_order SET showdoc_push_url=?, pushShowdocStatus='0' WHERE user=?",
			showdocURL, account,
		)
		if err != nil {
			return 0, err
		}
		return result.RowsAffected()
	}
	return 0, fmt.Errorf("需要订单ID或账号")
}

func (s *Service) WxPusherQRCode(account string) (map[string]interface{}, error) {
	appToken := s.getAdminConfigValue("wxpusher_token")
	if appToken == "" {
		return nil, fmt.Errorf("WxPusher未配置，请在系统设置-推送与同步中配置AppToken")
	}

	body := map[string]interface{}{
		"appToken":  appToken,
		"extra":     account,
		"validTime": 300,
	}
	bodyJSON, _ := json.Marshal(body)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(
		"https://wxpusher.zjiecode.com/api/fun/create/qrcode",
		"application/json",
		strings.NewReader(string(bodyJSON)),
	)
	if err != nil {
		return nil, fmt.Errorf("请求WxPusher失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析WxPusher响应失败")
	}

	if success, _ := result["success"].(bool); success {
		data := result["data"].(map[string]interface{})
		return map[string]interface{}{
			"code":     data["code"],
			"url":      data["url"],
			"shortUrl": data["shortUrl"],
		}, nil
	}

	msg, _ := result["msg"].(string)
	return nil, fmt.Errorf("二维码生成失败: %s", msg)
}

func (s *Service) WxPusherScanUID(code string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://wxpusher.zjiecode.com/api/fun/scan-qrcode-uid?code=" + code)
	if err != nil {
		return "", fmt.Errorf("请求WxPusher失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("解析响应失败")
	}

	codeVal, _ := result["code"].(float64)
	if int(codeVal) == 1000 {
		if uid, ok := result["data"].(string); ok && uid != "" {
			return uid, nil
		}
	}

	return "", fmt.Errorf("未扫码或已过期")
}

func (s *Service) PupLogin(oid int) (string, error) {
	baseURL := s.getAdminConfigValue("pup_base_url")
	plan := s.getAdminConfigValue("pup_plan")
	if baseURL == "" {
		return "", fmt.Errorf("Pup登录地址未配置，请在系统设置-推送与同步中配置")
	}

	var yid string
	err := database.DB.QueryRow("SELECT COALESCE(yid,'') FROM qingka_wangke_order WHERE oid=?", oid).Scan(&yid)
	if err != nil {
		return "", fmt.Errorf("订单不存在")
	}
	if yid == "" {
		return "", fmt.Errorf("订单尚未对接，无法登录")
	}

	loginURL := fmt.Sprintf("%s?oid=%s", baseURL, yid)
	if plan != "" {
		loginURL += "&plan=" + plan
	}
	return loginURL, nil
}

func (s *Service) getAdminConfigValue(key string) string {
	conf, _ := commonmodule.GetAdminConfigMap()
	return conf[key]
}
