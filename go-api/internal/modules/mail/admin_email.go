package mail

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"go-api/internal/config"
	"go-api/internal/database"
	"go-api/internal/model"
)

type Service struct{}

type emailPoolService struct {
	mu       sync.Mutex
	rrIndex  int
	lastHour int
	lastDay  int
}

var mailService = &Service{}
var poolOnce sync.Once
var poolInstance *emailPoolService

func Mail() *Service {
	return mailService
}

func emailPool() *emailPoolService {
	poolOnce.Do(func() {
		poolInstance = &emailPoolService{}
	})
	return poolInstance
}

func (s *Service) GetSMTPConfig() config.SMTPConfig {
	var cfg config.SMTPConfig
	err := database.DB.QueryRow(
		"SELECT host, port, user, password, from_name, encryption FROM qingka_smtp_config WHERE id = 1",
	).Scan(&cfg.Host, &cfg.Port, &cfg.User, &cfg.Password, &cfg.FromName, &cfg.Encryption)
	if err != nil || cfg.Host == "" {
		return config.Global.SMTP
	}
	return cfg
}

func (s *Service) SaveSMTPConfig(cfg config.SMTPConfig) error {
	_, err := database.DB.Exec(
		"INSERT INTO qingka_smtp_config (id, host, port, user, password, from_name, encryption) VALUES (1, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE host=VALUES(host), port=VALUES(port), user=VALUES(user), password=VALUES(password), from_name=VALUES(from_name), encryption=VALUES(encryption)",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.FromName, cfg.Encryption,
	)
	return err
}

func (s *Service) TestSMTPConfig(cfg config.SMTPConfig, testTo string) error {
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

func (s *Service) SendEmailWithType(to, subject, htmlBody, mailType string) error {
	maxRetry := s.getEmailPoolMaxRetry()
	for attempt := 0; attempt <= maxRetry; attempt++ {
		account := s.pickEmailPoolAccount()
		if account == nil {
			break
		}

		from := account.User
		if account.FromEmail != "" {
			from = account.FromEmail
		}
		fromName := account.Name
		if fromName == "" {
			fromName = "System"
		}
		msg := s.buildMessage(from, fromName, to, subject, htmlBody)
		addr := fmt.Sprintf("%s:%d", account.Host, account.Port)
		cfg := s.toSMTPConfig(*account)

		var err error
		switch strings.ToLower(account.Encryption) {
		case "ssl", "tls":
			err = s.sendSSL(addr, cfg, from, to, msg)
		case "starttls":
			err = s.sendSTARTTLS(addr, cfg, from, to, msg)
		default:
			err = s.sendPlain(addr, cfg, from, to, msg)
		}
		if err == nil {
			s.onEmailPoolSuccess(account.ID)
			s.logEmailPoolSend(account.ID, from, to, subject, mailType, true, "")
			return nil
		}

		log.Printf("[EmailPool] 邮箱#%d 发送失败: %v, 尝试 %d/%d", account.ID, err, attempt+1, maxRetry+1)
		s.onEmailPoolFail(account.ID, err.Error())
		s.logEmailPoolSend(account.ID, from, to, subject, mailType, false, err.Error())
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
	s.logEmailPoolSend(0, from, to, subject, mailType, err == nil, s.errString(err))
	return err
}

func (s *Service) SendEmail(to, subject, htmlBody string) error {
	return s.SendEmailWithType(to, subject, htmlBody, "notify")
}

func (s *Service) MassSend(target, subject, content string) (int64, error) {
	emails, err := s.ResolveRecipients(target)
	if err != nil {
		return 0, err
	}
	if len(emails) == 0 {
		return 0, fmt.Errorf("没有找到有效的收件人")
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := database.DB.Exec(
		"INSERT INTO qingka_email_log (target, subject, content, total, success_count, fail_count, status, addtime) VALUES (?, ?, ?, ?, 0, 0, 'sending', ?)",
		target, subject, content, len(emails), now,
	)
	if err != nil {
		return 0, fmt.Errorf("创建发送记录失败: %v", err)
	}
	logID, _ := result.LastInsertId()

	go func(id int64, recipients []string) {
		successCount := 0
		failCount := 0
		for _, email := range recipients {
			if err := s.SendEmailWithType(email, subject, content, "mass"); err != nil {
				log.Printf("[EmailMassSend] 发送到 %s 失败: %v", email, err)
				failCount++
			} else {
				successCount++
			}
			time.Sleep(200 * time.Millisecond)
		}

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
	}(logID, emails)

	return logID, nil
}

func (s *Service) ResolveRecipients(target string) ([]string, error) {
	var rowsQuery string
	var args []interface{}

	switch {
	case target == "all":
		rowsQuery = "SELECT COALESCE(user,''), COALESCE(email,'') FROM qingka_wangke_user WHERE active = '1'"
	case target == "direct":
		rowsQuery = "SELECT COALESCE(user,''), COALESCE(email,'') FROM qingka_wangke_user WHERE active = '1' AND uuid = 1"
	case target == "indirect":
		rowsQuery = "SELECT COALESCE(user,''), COALESCE(email,'') FROM qingka_wangke_user WHERE active = '1' AND uuid != 1"
	case strings.HasPrefix(target, "grade:"):
		rowsQuery = "SELECT COALESCE(user,''), COALESCE(email,'') FROM qingka_wangke_user WHERE active = '1' AND grade = ?"
		args = append(args, strings.TrimPrefix(target, "grade:"))
	case strings.HasPrefix(target, "uids:"):
		uids := strings.Split(strings.TrimPrefix(target, "uids:"), ",")
		placeholders := make([]string, len(uids))
		for i, uid := range uids {
			placeholders[i] = "?"
			args = append(args, strings.TrimSpace(uid))
		}
		rowsQuery = fmt.Sprintf("SELECT COALESCE(user,''), COALESCE(email,'') FROM qingka_wangke_user WHERE uid IN (%s)", strings.Join(placeholders, ","))
	default:
		return nil, fmt.Errorf("无效的收件人类型: %s", target)
	}

	rows, err := database.DB.Query(rowsQuery, args...)
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
		addr := emailField
		if addr == "" && qqUser != "" {
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
			seen[addr] = true
			emails = append(emails, addr)
		}
	}
	if emails == nil {
		emails = []string{}
	}
	return emails, nil
}

func (s *Service) GetSendLogs(page, limit int) ([]map[string]interface{}, int64, error) {
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

func (s *Service) EmailPoolList() ([]model.EmailPoolAccount, error) {
	rows, err := database.DB.Query(
		"SELECT id,name,host,port,encryption,user,password,from_email,weight,day_limit,hour_limit,today_sent,hour_sent,total_sent,total_fail,fail_streak,status,COALESCE(last_used,''),COALESCE(last_error,''),COALESCE(addtime,'') FROM qingka_email_pool ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.EmailPoolAccount
	for rows.Next() {
		var account model.EmailPoolAccount
		if err := rows.Scan(
			&account.ID, &account.Name, &account.Host, &account.Port, &account.Encryption, &account.User, &account.Password, &account.FromEmail,
			&account.Weight, &account.DayLimit, &account.HourLimit, &account.TodaySent, &account.HourSent,
			&account.TotalSent, &account.TotalFail, &account.FailStreak, &account.Status, &account.LastUsed, &account.LastError, &account.AddTime,
		); err != nil {
			continue
		}
		list = append(list, account)
	}
	if list == nil {
		list = []model.EmailPoolAccount{}
	}
	return list, nil
}

func (s *Service) SaveEmailPoolAccount(req model.EmailPoolSaveRequest) error {
	if req.Encryption == "" {
		req.Encryption = "ssl"
	}
	if req.Weight <= 0 {
		req.Weight = 1
	}
	now := time.Now().Format("2006-01-02 15:04:05")

	if req.ID > 0 {
		if req.Password == "" {
			_, err := database.DB.Exec(
				"UPDATE qingka_email_pool SET name=?,host=?,port=?,encryption=?,user=?,from_email=?,weight=?,day_limit=?,hour_limit=?,status=? WHERE id=?",
				req.Name, req.Host, req.Port, req.Encryption, req.User, req.FromEmail, req.Weight, req.DayLimit, req.HourLimit, req.Status, req.ID,
			)
			return err
		}
		_, err := database.DB.Exec(
			"UPDATE qingka_email_pool SET name=?,host=?,port=?,encryption=?,user=?,password=?,from_email=?,weight=?,day_limit=?,hour_limit=?,status=? WHERE id=?",
			req.Name, req.Host, req.Port, req.Encryption, req.User, req.Password, req.FromEmail, req.Weight, req.DayLimit, req.HourLimit, req.Status, req.ID,
		)
		return err
	}

	if req.Password == "" {
		return fmt.Errorf("新建邮箱必须填写授权码")
	}
	_, err := database.DB.Exec(
		"INSERT INTO qingka_email_pool (name,host,port,encryption,user,password,from_email,weight,day_limit,hour_limit,status,addtime) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)",
		req.Name, req.Host, req.Port, req.Encryption, req.User, req.Password, req.FromEmail, req.Weight, req.DayLimit, req.HourLimit, req.Status, now,
	)
	return err
}

func (s *Service) DeleteEmailPoolAccount(id int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_email_pool WHERE id=?", id)
	return err
}

func (s *Service) ToggleEmailPoolStatus(id, status int) error {
	_, err := database.DB.Exec("UPDATE qingka_email_pool SET status=?,fail_streak=0,last_error='' WHERE id=?", status, id)
	return err
}

func (s *Service) ResetEmailPoolCounters() error {
	_, err := database.DB.Exec("UPDATE qingka_email_pool SET today_sent=0,hour_sent=0")
	return err
}

func (s *Service) TestEmailPoolAccount(id int, testTo string) error {
	var account model.EmailPoolAccount
	err := database.DB.QueryRow(
		"SELECT id,host,port,encryption,user,password,from_email,name FROM qingka_email_pool WHERE id=?",
		id,
	).Scan(&account.ID, &account.Host, &account.Port, &account.Encryption, &account.User, &account.Password, &account.FromEmail, &account.Name)
	if err == sql.ErrNoRows {
		return fmt.Errorf("邮箱不存在")
	}
	if err != nil {
		return err
	}

	subject := "邮箱池测试"
	body := "<p>这是一封测试邮件，如果您收到说明该邮箱配置正确。</p>"
	return s.SendEmailWithType(testTo, subject, body, "pool_test")
}

func (s *Service) QueryEmailPoolLogs(q model.EmailSendLogQuery) ([]model.EmailSendLog, int64, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Limit <= 0 || q.Limit > 100 {
		q.Limit = 20
	}

	where := "1=1"
	var args []interface{}
	if q.MailType != "" {
		where += " AND mail_type=?"
		args = append(args, q.MailType)
	}
	if q.Status == 0 || q.Status == 1 {
		where += " AND status=?"
		args = append(args, q.Status)
	}
	if q.ToEmail != "" {
		where += " AND to_email LIKE ?"
		args = append(args, "%"+q.ToEmail+"%")
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_email_send_log WHERE "+where, args...).Scan(&total)

	offset := (q.Page - 1) * q.Limit
	queryArgs := append(args, q.Limit, offset)
	rows, err := database.DB.Query(
		"SELECT id,pool_id,from_email,to_email,subject,mail_type,status,COALESCE(error,''),COALESCE(addtime,'') FROM qingka_email_send_log WHERE "+where+" ORDER BY id DESC LIMIT ? OFFSET ?",
		queryArgs...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.EmailSendLog
	for rows.Next() {
		var item model.EmailSendLog
		if err := rows.Scan(&item.ID, &item.PoolID, &item.FromEmail, &item.ToEmail, &item.Subject, &item.MailType, &item.Status, &item.Error, &item.AddTime); err != nil {
			continue
		}
		list = append(list, item)
	}
	if list == nil {
		list = []model.EmailSendLog{}
	}
	return list, total, nil
}

func (s *Service) EmailPoolStats() map[string]interface{} {
	var totalAccounts, activeAccounts, errorAccounts int
	var todaySent, todayFail int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_email_pool").Scan(&totalAccounts)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_email_pool WHERE status=1").Scan(&activeAccounts)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_email_pool WHERE status=2").Scan(&errorAccounts)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_email_send_log WHERE status=1 AND DATE(addtime)=CURDATE()").Scan(&todaySent)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_email_send_log WHERE status=0 AND DATE(addtime)=CURDATE()").Scan(&todayFail)
	return map[string]interface{}{
		"total_accounts":  totalAccounts,
		"active_accounts": activeAccounts,
		"error_accounts":  errorAccounts,
		"today_sent":      todaySent,
		"today_fail":      todayFail,
	}
}

func (s *Service) pickEmailPoolAccount() *model.EmailPoolAccount {
	pool := emailPool()
	pool.mu.Lock()
	defer pool.mu.Unlock()

	s.autoResetCounters()
	accounts := s.availableEmailPoolAccounts()
	if len(accounts) == 0 {
		return nil
	}

	switch s.getEmailPoolStrategy() {
	case "random":
		return &accounts[rand.Intn(len(accounts))]
	case "weight":
		return s.pickEmailPoolByWeight(accounts)
	default:
		idx := pool.rrIndex % len(accounts)
		pool.rrIndex = idx + 1
		return &accounts[idx]
	}
}

func (s *Service) availableEmailPoolAccounts() []model.EmailPoolAccount {
	rows, err := database.DB.Query(
		"SELECT id,name,host,port,encryption,user,password,from_email,weight,day_limit,hour_limit,today_sent,hour_sent,total_sent,total_fail,fail_streak,status,COALESCE(last_used,''),COALESCE(last_error,'') FROM qingka_email_pool WHERE status=1 ORDER BY id",
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var list []model.EmailPoolAccount
	for rows.Next() {
		var account model.EmailPoolAccount
		if err := rows.Scan(
			&account.ID, &account.Name, &account.Host, &account.Port, &account.Encryption, &account.User, &account.Password, &account.FromEmail,
			&account.Weight, &account.DayLimit, &account.HourLimit, &account.TodaySent, &account.HourSent,
			&account.TotalSent, &account.TotalFail, &account.FailStreak, &account.Status, &account.LastUsed, &account.LastError,
		); err != nil {
			continue
		}
		if account.DayLimit > 0 && account.TodaySent >= account.DayLimit {
			continue
		}
		if account.HourLimit > 0 && account.HourSent >= account.HourLimit {
			continue
		}
		list = append(list, account)
	}
	return list
}

func (s *Service) pickEmailPoolByWeight(accounts []model.EmailPoolAccount) *model.EmailPoolAccount {
	total := 0
	for _, account := range accounts {
		total += account.Weight
	}
	if total == 0 {
		return &accounts[0]
	}
	pick := rand.Intn(total)
	for i := range accounts {
		pick -= accounts[i].Weight
		if pick < 0 {
			return &accounts[i]
		}
	}
	return &accounts[len(accounts)-1]
}

func (s *Service) getEmailPoolStrategy() string {
	var value string
	database.DB.QueryRow("SELECT `value` FROM qingka_wangke_config WHERE `key`='email_pool_strategy'").Scan(&value)
	if value == "" {
		return "round"
	}
	return value
}

func (s *Service) getEmailPoolFailThreshold() int {
	var value int
	database.DB.QueryRow("SELECT CAST(`value` AS SIGNED) FROM qingka_wangke_config WHERE `key`='email_pool_fail_threshold'").Scan(&value)
	if value <= 0 {
		return 5
	}
	return value
}

func (s *Service) getEmailPoolMaxRetry() int {
	var value int
	database.DB.QueryRow("SELECT CAST(`value` AS SIGNED) FROM qingka_wangke_config WHERE `key`='email_pool_max_retry'").Scan(&value)
	if value < 0 {
		return 2
	}
	return value
}

func (s *Service) autoResetCounters() {
	pool := emailPool()
	now := time.Now()
	hour := now.Hour()
	day := now.YearDay()
	if day != pool.lastDay {
		database.DB.Exec("UPDATE qingka_email_pool SET today_sent=0, hour_sent=0")
		pool.lastDay = day
		pool.lastHour = hour
		return
	}
	if hour != pool.lastHour {
		database.DB.Exec("UPDATE qingka_email_pool SET hour_sent=0")
		pool.lastHour = hour
	}
}

func (s *Service) onEmailPoolSuccess(poolID int) {
	now := time.Now().Format("2006-01-02 15:04:05")
	database.DB.Exec(
		"UPDATE qingka_email_pool SET today_sent=today_sent+1, hour_sent=hour_sent+1, total_sent=total_sent+1, fail_streak=0, last_used=? WHERE id=?",
		now, poolID,
	)
}

func (s *Service) onEmailPoolFail(poolID int, errMsg string) {
	now := time.Now().Format("2006-01-02 15:04:05")
	database.DB.Exec(
		"UPDATE qingka_email_pool SET total_fail=total_fail+1, fail_streak=fail_streak+1, last_error=?, last_used=? WHERE id=?",
		errMsg, now, poolID,
	)
	threshold := s.getEmailPoolFailThreshold()
	var streak int
	database.DB.QueryRow("SELECT fail_streak FROM qingka_email_pool WHERE id=?", poolID).Scan(&streak)
	if streak >= threshold {
		database.DB.Exec("UPDATE qingka_email_pool SET status=2 WHERE id=?", poolID)
		log.Printf("[EmailPool] 邮箱 #%d 连续失败 %d 次，已自动标记异常", poolID, streak)
	}
}

func (s *Service) logEmailPoolSend(poolID int, fromEmail, toEmail, subject, mailType string, success bool, errMsg string) {
	status := 1
	if !success {
		status = 0
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	database.DB.Exec(
		"INSERT INTO qingka_email_send_log (pool_id,from_email,to_email,subject,mail_type,status,error,addtime) VALUES (?,?,?,?,?,?,?,?)",
		poolID, fromEmail, toEmail, subject, mailType, status, errMsg, now,
	)
}

func (s *Service) toSMTPConfig(account model.EmailPoolAccount) config.SMTPConfig {
	return config.SMTPConfig{
		Host:       account.Host,
		Port:       account.Port,
		User:       account.User,
		Password:   account.Password,
		FromName:   account.Name,
		Encryption: account.Encryption,
	}
}

func (s *Service) sendSSL(addr string, cfg config.SMTPConfig, from, to string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: cfg.Host})
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

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA 失败: %v", err)
	}
	if _, err := writer.Write(msg); err != nil {
		return fmt.Errorf("写入邮件内容失败: %v", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("关闭写入失败: %v", err)
	}
	return client.Quit()
}

func (s *Service) sendSTARTTLS(addr string, cfg config.SMTPConfig, from, to string, msg []byte) error {
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

	if err := client.StartTLS(&tls.Config{ServerName: cfg.Host}); err != nil {
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

	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(msg); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func (s *Service) sendPlain(addr string, cfg config.SMTPConfig, from, to string, msg []byte) error {
	var auth smtp.Auth
	if cfg.Password != "" {
		auth = smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)
	}
	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}

func (s *Service) buildMessage(from, fromName, to, subject, htmlBody string) []byte {
	headers := fmt.Sprintf(
		"From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\nDate: %s\r\n\r\n",
		fromName, from, to, subject, time.Now().Format(time.RFC1123Z),
	)
	return []byte(headers + htmlBody)
}

func (s *Service) errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
