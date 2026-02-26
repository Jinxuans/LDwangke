package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go-api/internal/database"
)

// PushService 推送通知服务
type PushService struct{}

func NewPushService() *PushService {
	return &PushService{}
}

// BindPushUID 绑定微信推送UID到指定订单
func (s *PushService) BindPushUID(orderID int, pushUID string) error {
	_, err := database.DB.Exec(
		"UPDATE qingka_wangke_order SET pushUid=?, pushStatus='0' WHERE oid=?",
		pushUID, orderID,
	)
	return err
}

// UnbindPushUID 解绑微信推送UID（按账号批量）
func (s *PushService) UnbindPushUIDByAccount(account string) (int64, error) {
	result, err := database.DB.Exec(
		"UPDATE qingka_wangke_order SET pushUid='', pushStatus='0' WHERE user=?",
		account,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// BatchBindPushUID 批量绑定微信推送UID
func (s *PushService) BatchBindPushUID(orderIDs []int, pushUID string) (int64, error) {
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

// BindPushEmail 绑定邮箱推送（单个订单或按账号批量）
func (s *PushService) BindPushEmail(orderID int, account string, email string) (int64, error) {
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

// BindShowDocPush 绑定ShowDoc推送（单个订单或按账号批量）
func (s *PushService) BindShowDocPush(orderID int, account string, showdocURL string) (int64, error) {
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

// WxPusherQRCode 生成WxPusher扫码二维码（从系统配置读取）
func (s *PushService) WxPusherQRCode(account string) (map[string]interface{}, error) {
	conf, _ := NewAdminService().GetConfig()
	appToken := conf["wxpusher_token"]
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

// WxPusherScanUID 查询扫码结果获取UID
func (s *PushService) WxPusherScanUID(code string) (string, error) {
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

// LogPush 记录推送日志
func (s *PushService) LogPush(orderID int, uid string, pushType string, receiver string, content string, status string) {
	var emailCol, uidCol, showdocCol interface{}
	switch pushType {
	case "email":
		emailCol = receiver
	case "wxpusher":
		uidCol = receiver
	case "showdoc":
		showdocCol = receiver
	}
	database.DB.Exec(
		`INSERT INTO qingka_wangke_push_logs (order_id, uid, type, receiver_email, receiver_uid, showdoc_url, content, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		orderID, uid, pushType, emailCol, uidCol, showdocCol, content, status,
	)
}

// PupLogin 获取Pup自动登录URL（从系统配置读取）
func (s *PushService) PupLogin(oid int) (string, error) {
	conf, _ := NewAdminService().GetConfig()
	baseURL := conf["pup_base_url"]
	plan := conf["pup_plan"]
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
