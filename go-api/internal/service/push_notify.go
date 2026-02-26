package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go-api/internal/database"
)

// NotifyOrderStatusChange 订单状态/进度变更后，自动推送通知给已绑定的用户
// 在所有订单状态变更的地方调用此函数（异步goroutine）
func NotifyOrderStatusChange(oid int, newStatus string, newProcess string, remarks string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[Push] oid=%d panic: %v\n", oid, r)
			}
		}()

		var user, kcname, pushUid, pushEmail, showdocURL string
		err := database.DB.QueryRow(
			`SELECT COALESCE(user,''), COALESCE(kcname,''), COALESCE(pushUid,''), 
			 COALESCE(pushEmail,''), COALESCE(showdoc_push_url,'') 
			 FROM qingka_wangke_order WHERE oid=?`, oid,
		).Scan(&user, &kcname, &pushUid, &pushEmail, &showdocURL)
		if err != nil {
			return
		}

		// 没有任何推送绑定，直接返回
		if pushUid == "" && pushEmail == "" && showdocURL == "" {
			return
		}

		// 构建推送消息内容
		content := fmt.Sprintf("订单 #%d 状态更新\n账号: %s\n课程: %s\n状态: %s", oid, user, kcname, newStatus)
		if newProcess != "" {
			content += fmt.Sprintf("\n进度: %s", newProcess)
		}
		if remarks != "" {
			content += fmt.Sprintf("\n备注: %s", remarks)
		}

		// 1. WxPusher 推送
		if pushUid != "" {
			sendWxPush(oid, user, pushUid, content)
		}

		// 2. ShowDoc 推送
		if showdocURL != "" {
			sendShowDocPush(oid, user, showdocURL, content)
		}

		// 3. 邮箱推送（通过系统邮件池发送）
		if pushEmail != "" {
			sendEmailPush(oid, user, pushEmail, newStatus, content)
		}
	}()
}

// sendWxPush 发送WxPusher微信推送
func sendWxPush(oid int, user string, pushUid string, content string) {
	// 从系统配置读取 appToken
	conf, _ := NewAdminService().GetConfig()
	appToken := conf["wxpusher_token"]
	if appToken == "" {
		return
	}

	body := map[string]interface{}{
		"appToken":    appToken,
		"content":     content,
		"contentType": 1, // 文本
		"uids":        []string{pushUid},
	}
	bodyJSON, _ := json.Marshal(body)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(
		"https://wxpusher.zjiecode.com/api/send/message",
		"application/json",
		bytes.NewReader(bodyJSON),
	)
	if err != nil {
		logPush(oid, user, "wxpusher", pushUid, content, "失败")
		database.DB.Exec("UPDATE qingka_wangke_order SET pushStatus='失败' WHERE oid=?", oid)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	if code, _ := result["code"].(float64); int(code) == 1000 {
		logPush(oid, user, "wxpusher", pushUid, content, "成功")
		database.DB.Exec("UPDATE qingka_wangke_order SET pushStatus='成功' WHERE oid=?", oid)
	} else {
		logPush(oid, user, "wxpusher", pushUid, content, "失败")
		database.DB.Exec("UPDATE qingka_wangke_order SET pushStatus='失败' WHERE oid=?", oid)
	}
}

// sendShowDocPush 发送ShowDoc Webhook推送
func sendShowDocPush(oid int, user string, showdocURL string, content string) {
	body := map[string]interface{}{
		"title":   fmt.Sprintf("订单 #%d 状态更新", oid),
		"content": content,
	}
	bodyJSON, _ := json.Marshal(body)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(showdocURL, "application/json", bytes.NewReader(bodyJSON))
	if err != nil {
		logPush(oid, user, "showdoc", showdocURL, content, "失败")
		database.DB.Exec("UPDATE qingka_wangke_order SET pushShowdocStatus='失败' WHERE oid=?", oid)
		return
	}
	defer resp.Body.Close()

	logPush(oid, user, "showdoc", showdocURL, content, "成功")
	database.DB.Exec("UPDATE qingka_wangke_order SET pushShowdocStatus='成功' WHERE oid=?", oid)
}

// sendEmailPush 通过系统邮件服务发送邮箱推送
func sendEmailPush(oid int, user string, email string, status string, content string) {
	subject := fmt.Sprintf("订单 #%d 状态更新: %s", oid, status)
	htmlBody := fmt.Sprintf("<pre>%s</pre>", content)

	es := NewEmailService()
	err := es.SendEmailWithType(email, subject, htmlBody, "push")
	if err != nil {
		logPush(oid, user, "email", email, content, "失败")
		database.DB.Exec("UPDATE qingka_wangke_order SET pushEmailStatus='失败' WHERE oid=?", oid)
		return
	}

	logPush(oid, user, "email", email, content, "成功")
	database.DB.Exec("UPDATE qingka_wangke_order SET pushEmailStatus='成功' WHERE oid=?", oid)
}

// logPush 记录推送日志
func logPush(oid int, uid string, pushType string, receiver string, content string, status string) {
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
		oid, uid, pushType, emailCol, uidCol, showdocCol, content, status,
	)
}
