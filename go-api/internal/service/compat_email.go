package service

import "go-api/internal/config"

type EmailService struct{}

var emailService = &EmailService{}

func GetSMTPConfig() config.SMTPConfig {
	return emailService.GetSMTPConfig()
}

func SaveSMTPConfig(cfg config.SMTPConfig) error {
	return emailService.SaveSMTPConfig(cfg)
}

func TestSMTPConfig(cfg config.SMTPConfig, testTo string) error {
	return emailService.TestSMTPConfig(cfg, testTo)
}

func SendEmail(to, subject, htmlBody string) error {
	return emailService.SendEmail(to, subject, htmlBody)
}

func SendEmailWithType(to, subject, htmlBody, mailType string) error {
	return emailService.SendEmailWithType(to, subject, htmlBody, mailType)
}

func MassSend(target, subject, content string) (int64, error) {
	return emailService.MassSend(target, subject, content)
}

func ResolveRecipients(target string) ([]string, error) {
	return emailService.ResolveRecipients(target)
}

func GetSendLogs(page, limit int) ([]map[string]interface{}, int64, error) {
	return emailService.GetSendLogs(page, limit)
}
