package order

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go-api/internal/database"
	commonmodule "go-api/internal/modules/common"
	mailmodule "go-api/internal/modules/mail"
	obslogger "go-api/internal/observability/logger"
)

// NotifyOrderStatusChange 订单状态/进度变更后，自动推送通知给已绑定的用户。
func NotifyOrderStatusChange(oid int, newStatus string, newProcess string, remarks string) {
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				obslogger.L().Error("OrderPush panic", "oid", oid, "panic", recovered)
			}
		}()

		var user, kcname, pushUID, pushEmail, showDocURL string
		err := database.DB.QueryRow(
			`SELECT COALESCE(user,''), COALESCE(kcname,''), COALESCE(pushUid,''),
			 COALESCE(pushEmail,''), COALESCE(showdoc_push_url,'')
			 FROM qingka_wangke_order WHERE oid=?`,
			oid,
		).Scan(&user, &kcname, &pushUID, &pushEmail, &showDocURL)
		if err != nil {
			return
		}
		if pushUID == "" && pushEmail == "" && showDocURL == "" {
			return
		}

		content := fmt.Sprintf("订单 #%d 状态更新\n账号: %s\n课程: %s\n状态: %s", oid, user, kcname, newStatus)
		if newProcess != "" {
			content += fmt.Sprintf("\n进度: %s", newProcess)
		}
		if remarks != "" {
			content += fmt.Sprintf("\n备注: %s", remarks)
		}

		if pushUID != "" {
			sendOrderWxPush(oid, user, pushUID, content)
		}
		if showDocURL != "" {
			sendOrderShowDocPush(oid, user, showDocURL, content)
		}
		if pushEmail != "" {
			sendOrderEmailPush(oid, user, pushEmail, newStatus, content)
		}
	}()
}

func sendOrderWxPush(oid int, user, pushUID, content string) {
	configMap, _ := commonmodule.GetAdminConfigMap()
	appToken := configMap["wxpusher_token"]
	if appToken == "" {
		return
	}

	body := map[string]interface{}{
		"appToken":    appToken,
		"content":     content,
		"contentType": 1,
		"uids":        []string{pushUID},
	}
	bodyJSON, _ := json.Marshal(body)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(
		"https://wxpusher.zjiecode.com/api/send/message",
		"application/json",
		bytes.NewReader(bodyJSON),
	)
	if err != nil {
		logOrderPush(oid, user, "wxpusher", pushUID, content, "失败")
		database.DB.Exec("UPDATE qingka_wangke_order SET pushStatus='失败' WHERE oid=?", oid)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	if code, _ := result["code"].(float64); int(code) == 1000 {
		logOrderPush(oid, user, "wxpusher", pushUID, content, "成功")
		database.DB.Exec("UPDATE qingka_wangke_order SET pushStatus='成功' WHERE oid=?", oid)
		return
	}

	logOrderPush(oid, user, "wxpusher", pushUID, content, "失败")
	database.DB.Exec("UPDATE qingka_wangke_order SET pushStatus='失败' WHERE oid=?", oid)
}

func sendOrderShowDocPush(oid int, user, showDocURL, content string) {
	body := map[string]interface{}{
		"title":   fmt.Sprintf("订单 #%d 状态更新", oid),
		"content": content,
	}
	bodyJSON, _ := json.Marshal(body)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(showDocURL, "application/json", bytes.NewReader(bodyJSON))
	if err != nil {
		logOrderPush(oid, user, "showdoc", showDocURL, content, "失败")
		database.DB.Exec("UPDATE qingka_wangke_order SET pushShowdocStatus='失败' WHERE oid=?", oid)
		return
	}
	defer resp.Body.Close()

	logOrderPush(oid, user, "showdoc", showDocURL, content, "成功")
	database.DB.Exec("UPDATE qingka_wangke_order SET pushShowdocStatus='成功' WHERE oid=?", oid)
}

func sendOrderEmailPush(oid int, user, email, status, content string) {
	subject := fmt.Sprintf("订单 #%d 状态更新: %s", oid, status)
	htmlBody := fmt.Sprintf("<pre>%s</pre>", content)

	if err := mailmodule.Mail().SendEmailWithType(email, subject, htmlBody, "push"); err != nil {
		logOrderPush(oid, user, "email", email, content, "失败")
		database.DB.Exec("UPDATE qingka_wangke_order SET pushEmailStatus='失败' WHERE oid=?", oid)
		return
	}

	logOrderPush(oid, user, "email", email, content, "成功")
	database.DB.Exec("UPDATE qingka_wangke_order SET pushEmailStatus='成功' WHERE oid=?", oid)
}

func logOrderPush(oid int, uid, pushType, receiver, content, status string) {
	var emailCol, uidCol, showDocCol interface{}
	switch pushType {
	case "email":
		emailCol = receiver
	case "wxpusher":
		uidCol = receiver
	case "showdoc":
		showDocCol = receiver
	}
	database.DB.Exec(
		`INSERT INTO qingka_wangke_push_logs (order_id, uid, type, receiver_email, receiver_uid, showdoc_url, content, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		oid, uid, pushType, emailCol, uidCol, showDocCol, content, status,
	)
}
