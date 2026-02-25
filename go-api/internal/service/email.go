package service

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strings"
	"time"

	"go-api/internal/config"
	"go-api/internal/database"
	"go-api/internal/model"
)

type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

// GetSMTPConfig 从数据库读取 SMTP 配置，没有则返回 yaml 默认值
func (s *EmailService) GetSMTPConfig() config.SMTPConfig {
	var cfg config.SMTPConfig
	err := database.DB.QueryRow(
		"SELECT host, port, user, password, from_name, encryption FROM qingka_smtp_config WHERE id = 1",
	).Scan(&cfg.Host, &cfg.Port, &cfg.User, &cfg.Password, &cfg.FromName, &cfg.Encryption)
	if err != nil || cfg.Host == "" {
		// 回退到 yaml 配置
		return config.Global.SMTP
	}
	return cfg
}

// SaveSMTPConfig 保存 SMTP 配置到数据库
func (s *EmailService) SaveSMTPConfig(cfg config.SMTPConfig) error {
	_, err := database.DB.Exec(
		"INSERT INTO qingka_smtp_config (id, host, port, user, password, from_name, encryption) VALUES (1, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE host=VALUES(host), port=VALUES(port), user=VALUES(user), password=VALUES(password), from_name=VALUES(from_name), encryption=VALUES(encryption)",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.FromName, cfg.Encryption,
	)
	return err
}

// TestSMTPConfig 测试 SMTP 配置是否可用
func (s *EmailService) TestSMTPConfig(cfg config.SMTPConfig, testTo string) error {
	if cfg.Host == "" || cfg.User == "" {
		return fmt.Errorf("SMTP 配置不完整")
	}
	from := cfg.User
	fromName := cfg.FromName
	if fromName == "" {
		fromName = "System"
	}
	msg := s.buildMessage(from, fromName, testTo, "SMTP 测试邮件", "<p>这是一封测试邮件，如果您收到此邮件说明 SMTP 配置正确。</p>")
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	switch strings.ToLower(cfg.Encryption) {
	case "ssl", "tls":
		return s.sendSSL(addr, cfg, from, testTo, msg)
	case "starttls":
		return s.sendSTARTTLS(addr, cfg, from, testTo, msg)
	default:
		return s.sendPlain(addr, cfg, from, testTo, msg)
	}
}

// SendEmail 发送单封邮件 — 优先走轮询池，池空则回退旧配置
func (s *EmailService) SendEmail(to, subject, htmlBody string) error {
	return s.SendEmailWithType(to, subject, htmlBody, "notify")
}

// SendEmailWithType 带类型的发送（用于日志分类）
func (s *EmailService) SendEmailWithType(to, subject, htmlBody, mailType string) error {
	pool := GetEmailPoolService()
	maxRetry := pool.getMaxRetry()

	// 尝试从池中选取邮箱发送（含重试）
	for attempt := 0; attempt <= maxRetry; attempt++ {
		acct := pool.Pick()
		if acct == nil {
			break // 池空，走 fallback
		}
		from := acct.User
		if acct.FromEmail != "" {
			from = acct.FromEmail
		}
		fromName := acct.Name
		if fromName == "" {
			fromName = "System"
		}
		msg := s.buildMessage(from, fromName, to, subject, htmlBody)
		addr := fmt.Sprintf("%s:%d", acct.Host, acct.Port)
		cfg := toSMTPConfig(*acct)
		var err error
		switch strings.ToLower(acct.Encryption) {
		case "ssl", "tls":
			err = s.sendSSL(addr, cfg, from, to, msg)
		case "starttls":
			err = s.sendSTARTTLS(addr, cfg, from, to, msg)
		default:
			err = s.sendPlain(addr, cfg, from, to, msg)
		}
		if err == nil {
			pool.OnSuccess(acct.ID)
			pool.LogSend(acct.ID, from, to, subject, mailType, true, "")
			return nil
		}
		log.Printf("[EmailPool] 邮箱#%d 发送失败: %v, 尝试 %d/%d", acct.ID, err, attempt+1, maxRetry+1)
		pool.OnFail(acct.ID, err.Error())
		pool.LogSend(acct.ID, from, to, subject, mailType, false, err.Error())
	}

	// Fallback: 旧的单 SMTP 配置
	cfg := s.GetSMTPConfig()
	if cfg.Host == "" || cfg.User == "" {
		return fmt.Errorf("邮箱池无可用账号且 SMTP 未配置")
	}
	from := cfg.User
	fromName := cfg.FromName
	if fromName == "" {
		fromName = "System"
	}
	msg := s.buildMessage(from, fromName, to, subject, htmlBody)
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	var err error
	switch strings.ToLower(cfg.Encryption) {
	case "ssl", "tls":
		err = s.sendSSL(addr, cfg, from, to, msg)
	case "starttls":
		err = s.sendSTARTTLS(addr, cfg, from, to, msg)
	default:
		err = s.sendPlain(addr, cfg, from, to, msg)
	}
	pool.LogSend(0, from, to, subject, mailType, err == nil, errStr(err))
	return err
}

func errStr(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// toSMTPConfig 将池账号转为 config.SMTPConfig 供底层发送函数使用
func toSMTPConfig(a model.EmailPoolAccount) config.SMTPConfig {
	return config.SMTPConfig{
		Host:       a.Host,
		Port:       a.Port,
		User:       a.User,
		Password:   a.Password,
		FromName:   a.Name,
		Encryption: a.Encryption,
	}
}

// SSL/TLS 直连 (port 465)
func (s *EmailService) sendSSL(addr string, cfg config.SMTPConfig, from, to string, msg []byte) error {
	tlsConfig := &tls.Config{
		ServerName: cfg.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS 连接失败: %v", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		return fmt.Errorf("SMTP 客户端创建失败: %v", err)
	}
	defer client.Close()

	auth := smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP 认证失败: %v", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM 失败: %v", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO 失败: %v", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA 失败: %v", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("写入邮件内容失败: %v", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("关闭写入失败: %v", err)
	}

	return client.Quit()
}

// STARTTLS (port 587)
func (s *EmailService) sendSTARTTLS(addr string, cfg config.SMTPConfig, from, to string, msg []byte) error {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		return fmt.Errorf("SMTP 客户端创建失败: %v", err)
	}
	defer client.Close()

	tlsConfig := &tls.Config{
		ServerName: cfg.Host,
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("STARTTLS 失败: %v", err)
	}

	auth := smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP 认证失败: %v", err)
	}

	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return client.Quit()
}

// 明文 (port 25, 内网/自建)
func (s *EmailService) sendPlain(addr string, cfg config.SMTPConfig, from, to string, msg []byte) error {
	var auth smtp.Auth
	if cfg.Password != "" {
		auth = smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)
	}
	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}

// 构建 MIME 邮件
func (s *EmailService) buildMessage(from, fromName, to, subject, htmlBody string) []byte {
	headers := fmt.Sprintf(
		"From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\nDate: %s\r\n\r\n",
		fromName, from, to, subject, time.Now().Format(time.RFC1123Z),
	)
	return []byte(headers + htmlBody)
}

// MassSend 群发邮件（异步），返回任务ID
func (s *EmailService) MassSend(target, subject, content string) (int64, error) {
	// 解析收件人列表
	emails, err := s.ResolveRecipients(target)
	if err != nil {
		return 0, err
	}
	if len(emails) == 0 {
		return 0, fmt.Errorf("没有找到有效的收件人")
	}

	// 插入群发记录
	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := database.DB.Exec(
		"INSERT INTO qingka_email_log (target, subject, content, total, success_count, fail_count, status, addtime) VALUES (?, ?, ?, ?, 0, 0, 'sending', ?)",
		target, subject, content, len(emails), now,
	)
	if err != nil {
		return 0, fmt.Errorf("创建发送记录失败: %v", err)
	}
	logID, _ := result.LastInsertId()

	// 异步发送
	go func(id int64, recipients []string) {
		successCount := 0
		failCount := 0
		for _, email := range recipients {
			if err := s.SendEmail(email, subject, content); err != nil {
				log.Printf("[EmailMassSend] 发送到 %s 失败: %v", email, err)
				failCount++
			} else {
				successCount++
			}
			// 每封间隔200ms，避免被限速
			time.Sleep(200 * time.Millisecond)
		}
		// 更新记录
		status := "done"
		if failCount > 0 && successCount == 0 {
			status = "failed"
		} else if failCount > 0 {
			status = "partial"
		}
		database.DB.Exec(
			"UPDATE qingka_email_log SET success_count=?, fail_count=?, status=? WHERE id=?",
			successCount, failCount, status, id,
		)
		log.Printf("[EmailMassSend] 任务 %d 完成: 成功 %d, 失败 %d", id, successCount, failCount)
	}(logID, emails)

	return logID, nil
}

// ResolveRecipients 解析收件人
// target: "all" | "grade:1" | "uids:1,2,3" | "direct" | "indirect"
func (s *EmailService) ResolveRecipients(target string) ([]string, error) {
	var rows_query string
	var args []interface{}

	if target == "all" {
		rows_query = "SELECT COALESCE(user,''), COALESCE(email,'') FROM qingka_wangke_user WHERE active = '1'"
	} else if target == "direct" {
		rows_query = "SELECT COALESCE(user,''), COALESCE(email,'') FROM qingka_wangke_user WHERE active = '1' AND uuid = 1"
	} else if target == "indirect" {
		rows_query = "SELECT COALESCE(user,''), COALESCE(email,'') FROM qingka_wangke_user WHERE active = '1' AND uuid != 1"
	} else if strings.HasPrefix(target, "grade:") {
		grade := strings.TrimPrefix(target, "grade:")
		rows_query = "SELECT COALESCE(user,''), COALESCE(email,'') FROM qingka_wangke_user WHERE active = '1' AND grade = ?"
		args = append(args, grade)
	} else if strings.HasPrefix(target, "uids:") {
		uidStr := strings.TrimPrefix(target, "uids:")
		uids := strings.Split(uidStr, ",")
		placeholders := make([]string, len(uids))
		for i, uid := range uids {
			placeholders[i] = "?"
			args = append(args, strings.TrimSpace(uid))
		}
		rows_query = fmt.Sprintf("SELECT COALESCE(user,''), COALESCE(email,'') FROM qingka_wangke_user WHERE uid IN (%s)", strings.Join(placeholders, ","))
	} else {
		return nil, fmt.Errorf("无效的收件人类型: %s", target)
	}

	rows, err := database.DB.Query(rows_query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []string
	seen := map[string]bool{}
	for rows.Next() {
		var qqUser, emailField string
		if err := rows.Scan(&qqUser, &emailField); err != nil {
			continue
		}
		// 优先用 email 字段，没有就用 QQ号@qq.com
		addr := emailField
		if addr == "" && qqUser != "" {
			// 只有纯数字的 QQ 号才拼接
			isQQ := true
			for _, c := range qqUser {
				if c < '0' || c > '9' {
					isQQ = false
					break
				}
			}
			if isQQ {
				addr = qqUser + "@qq.com"
			}
		}
		if addr != "" && !seen[addr] {
			emails = append(emails, addr)
			seen[addr] = true
		}
	}
	return emails, nil
}

// GetSendLogs 获取群发记录
func (s *EmailService) GetSendLogs(page, limit int) ([]map[string]interface{}, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_email_log").Scan(&total)

	offset := (page - 1) * limit
	rows, err := database.DB.Query(
		"SELECT id, target, subject, total, success_count, fail_count, status, addtime FROM qingka_email_log ORDER BY id DESC LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var id int64
		var target, subject, status, addtime string
		var totalN, successN, failN int
		if err := rows.Scan(&id, &target, &subject, &totalN, &successN, &failN, &status, &addtime); err != nil {
			continue
		}
		logs = append(logs, map[string]interface{}{
			"id":            id,
			"target":        target,
			"subject":       subject,
			"total":         totalN,
			"success_count": successN,
			"fail_count":    failN,
			"status":        status,
			"addtime":       addtime,
		})
	}
	if logs == nil {
		logs = []map[string]interface{}{}
	}
	return logs, total, nil
}
