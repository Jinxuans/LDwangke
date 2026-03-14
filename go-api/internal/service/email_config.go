package service

import (
	"fmt"
	"strings"

	"go-api/internal/config"
	"go-api/internal/database"
)

// GetSMTPConfig 从数据库读取 SMTP 配置，没有则返回 yaml 默认值。
func (s *EmailService) GetSMTPConfig() config.SMTPConfig {
	var cfg config.SMTPConfig
	err := database.DB.QueryRow(
		"SELECT host, port, user, password, from_name, encryption FROM qingka_smtp_config WHERE id = 1",
	).Scan(&cfg.Host, &cfg.Port, &cfg.User, &cfg.Password, &cfg.FromName, &cfg.Encryption)
	if err != nil || cfg.Host == "" {
		return config.Global.SMTP
	}
	return cfg
}

// SaveSMTPConfig 保存 SMTP 配置到数据库。
func (s *EmailService) SaveSMTPConfig(cfg config.SMTPConfig) error {
	_, err := database.DB.Exec(
		"INSERT INTO qingka_smtp_config (id, host, port, user, password, from_name, encryption) VALUES (1, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE host=VALUES(host), port=VALUES(port), user=VALUES(user), password=VALUES(password), from_name=VALUES(from_name), encryption=VALUES(encryption)",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.FromName, cfg.Encryption,
	)
	return err
}

// TestSMTPConfig 测试 SMTP 配置是否可用。
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
