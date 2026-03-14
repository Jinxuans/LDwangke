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
	"go-api/internal/model"
)

// SendEmail 发送单封邮件，优先走邮箱池。
func (s *EmailService) SendEmail(to, subject, htmlBody string) error {
	return s.SendEmailWithType(to, subject, htmlBody, "notify")
}

// SendEmailWithType 带类型的发送，用于日志分类。
func (s *EmailService) SendEmailWithType(to, subject, htmlBody, mailType string) error {
	pool := EmailPool()
	maxRetry := pool.getMaxRetry()

	for attempt := 0; attempt <= maxRetry; attempt++ {
		acct := pool.Pick()
		if acct == nil {
			break
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

// toSMTPConfig 将池账号转为 config.SMTPConfig 供底层发送函数使用。
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

// sendSSL 走 SSL/TLS 直连。
func (s *EmailService) sendSSL(addr string, cfg config.SMTPConfig, from, to string, msg []byte) error {
	tlsConfig := &tls.Config{ServerName: cfg.Host}

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

// sendSTARTTLS 走 STARTTLS。
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

	tlsConfig := &tls.Config{ServerName: cfg.Host}
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

// sendPlain 走明文 SMTP。
func (s *EmailService) sendPlain(addr string, cfg config.SMTPConfig, from, to string, msg []byte) error {
	var auth smtp.Auth
	if cfg.Password != "" {
		auth = smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)
	}
	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}

// buildMessage 构建 MIME 邮件内容。
func (s *EmailService) buildMessage(from, fromName, to, subject, htmlBody string) []byte {
	headers := fmt.Sprintf(
		"From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\nDate: %s\r\n\r\n",
		fromName, from, to, subject, time.Now().Format(time.RFC1123Z),
	)
	return []byte(headers + htmlBody)
}
