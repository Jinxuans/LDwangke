package service

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

// EmailPoolService 邮箱轮询池管理
type EmailPoolService struct {
	mu       sync.Mutex
	rrIndex  int // round-robin 当前索引
	lastHour int // 上次重置 hour_sent 的小时
	lastDay  int // 上次重置 today_sent 的日期(day of year)
}

var poolOnce sync.Once
var poolInstance *EmailPoolService

func EmailPool() *EmailPoolService {
	poolOnce.Do(func() {
		poolInstance = &EmailPoolService{}
	})
	return poolInstance
}

func EmailPoolList() ([]model.EmailPoolAccount, error) {
	return EmailPool().List()
}

func SaveEmailPoolAccount(req model.EmailPoolSaveRequest) error {
	return EmailPool().Save(req)
}

func DeleteEmailPoolAccount(id int) error {
	return EmailPool().Delete(id)
}

func ToggleEmailPoolStatus(id, status int) error {
	return EmailPool().ToggleStatus(id, status)
}

func ResetEmailPoolCounters() error {
	return EmailPool().ResetCounters()
}

func TestEmailPoolAccount(id int, testTo string) error {
	return EmailPool().TestAccount(id, testTo)
}

func QueryEmailPoolLogs(q model.EmailSendLogQuery) ([]model.EmailSendLog, int64, error) {
	return EmailPool().QueryLogs(q)
}

func EmailPoolStats() map[string]interface{} {
	return EmailPool().Stats()
}

// -------- CRUD --------

func (s *EmailPoolService) List() ([]model.EmailPoolAccount, error) {
	rows, err := database.DB.Query(
		"SELECT id,name,host,port,encryption,user,password,from_email,weight,day_limit,hour_limit,today_sent,hour_sent,total_sent,total_fail,fail_streak,status,COALESCE(last_used,''),COALESCE(last_error,''),COALESCE(addtime,'') FROM qingka_email_pool ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.EmailPoolAccount
	for rows.Next() {
		var a model.EmailPoolAccount
		if err := rows.Scan(&a.ID, &a.Name, &a.Host, &a.Port, &a.Encryption, &a.User, &a.Password, &a.FromEmail,
			&a.Weight, &a.DayLimit, &a.HourLimit, &a.TodaySent, &a.HourSent,
			&a.TotalSent, &a.TotalFail, &a.FailStreak, &a.Status, &a.LastUsed, &a.LastError, &a.AddTime); err != nil {
			continue
		}
		list = append(list, a)
	}
	if list == nil {
		list = []model.EmailPoolAccount{}
	}
	return list, nil
}

func (s *EmailPoolService) Save(req model.EmailPoolSaveRequest) error {
	if req.Encryption == "" {
		req.Encryption = "ssl"
	}
	if req.Weight <= 0 {
		req.Weight = 1
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	if req.ID > 0 {
		// 编辑 — 如果 password 为空则不更新密码
		if req.Password == "" {
			_, err := database.DB.Exec(
				"UPDATE qingka_email_pool SET name=?,host=?,port=?,encryption=?,user=?,from_email=?,weight=?,day_limit=?,hour_limit=?,status=? WHERE id=?",
				req.Name, req.Host, req.Port, req.Encryption, req.User, req.FromEmail, req.Weight, req.DayLimit, req.HourLimit, req.Status, req.ID)
			return err
		}
		_, err := database.DB.Exec(
			"UPDATE qingka_email_pool SET name=?,host=?,port=?,encryption=?,user=?,password=?,from_email=?,weight=?,day_limit=?,hour_limit=?,status=? WHERE id=?",
			req.Name, req.Host, req.Port, req.Encryption, req.User, req.Password, req.FromEmail, req.Weight, req.DayLimit, req.HourLimit, req.Status, req.ID)
		return err
	}
	if req.Password == "" {
		return fmt.Errorf("新建邮箱必须填写授权码")
	}
	_, err := database.DB.Exec(
		"INSERT INTO qingka_email_pool (name,host,port,encryption,user,password,from_email,weight,day_limit,hour_limit,status,addtime) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)",
		req.Name, req.Host, req.Port, req.Encryption, req.User, req.Password, req.FromEmail, req.Weight, req.DayLimit, req.HourLimit, req.Status, now)
	return err
}

func (s *EmailPoolService) Delete(id int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_email_pool WHERE id=?", id)
	return err
}

func (s *EmailPoolService) ToggleStatus(id, status int) error {
	_, err := database.DB.Exec("UPDATE qingka_email_pool SET status=?,fail_streak=0,last_error='' WHERE id=?", status, id)
	return err
}

func (s *EmailPoolService) ResetCounters() error {
	_, err := database.DB.Exec("UPDATE qingka_email_pool SET today_sent=0,hour_sent=0")
	return err
}

// -------- 轮询调度 --------

// Pick 根据策略选一个可用邮箱，返回 nil 表示池空/全部不可用
func (s *EmailPoolService) Pick() *model.EmailPoolAccount {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 自动重置计数器
	s.autoResetCounters()

	accounts := s.availableAccounts()
	if len(accounts) == 0 {
		return nil
	}

	strategy := s.getStrategy()
	switch strategy {
	case "random":
		return &accounts[rand.Intn(len(accounts))]
	case "weight":
		return s.pickByWeight(accounts)
	default: // round
		idx := s.rrIndex % len(accounts)
		s.rrIndex = idx + 1
		return &accounts[idx]
	}
}

// availableAccounts 获取所有可用(启用+未超限)的账号
func (s *EmailPoolService) availableAccounts() []model.EmailPoolAccount {
	rows, err := database.DB.Query(
		"SELECT id,name,host,port,encryption,user,password,from_email,weight,day_limit,hour_limit,today_sent,hour_sent,total_sent,total_fail,fail_streak,status,COALESCE(last_used,''),COALESCE(last_error,'') FROM qingka_email_pool WHERE status=1 ORDER BY id")
	if err != nil {
		return nil
	}
	defer rows.Close()
	var list []model.EmailPoolAccount
	for rows.Next() {
		var a model.EmailPoolAccount
		if err := rows.Scan(&a.ID, &a.Name, &a.Host, &a.Port, &a.Encryption, &a.User, &a.Password, &a.FromEmail,
			&a.Weight, &a.DayLimit, &a.HourLimit, &a.TodaySent, &a.HourSent,
			&a.TotalSent, &a.TotalFail, &a.FailStreak, &a.Status, &a.LastUsed, &a.LastError); err != nil {
			continue
		}
		// 跳过超限的
		if a.DayLimit > 0 && a.TodaySent >= a.DayLimit {
			continue
		}
		if a.HourLimit > 0 && a.HourSent >= a.HourLimit {
			continue
		}
		list = append(list, a)
	}
	return list
}

func (s *EmailPoolService) pickByWeight(accounts []model.EmailPoolAccount) *model.EmailPoolAccount {
	total := 0
	for _, a := range accounts {
		total += a.Weight
	}
	if total == 0 {
		return &accounts[0]
	}
	r := rand.Intn(total)
	for i := range accounts {
		r -= accounts[i].Weight
		if r < 0 {
			return &accounts[i]
		}
	}
	return &accounts[len(accounts)-1]
}

func (s *EmailPoolService) getStrategy() string {
	var v string
	database.DB.QueryRow("SELECT `value` FROM qingka_wangke_config WHERE `key`='email_pool_strategy'").Scan(&v)
	if v == "" {
		return "round"
	}
	return v
}

func (s *EmailPoolService) getFailThreshold() int {
	var v int
	database.DB.QueryRow("SELECT CAST(`value` AS SIGNED) FROM qingka_wangke_config WHERE `key`='email_pool_fail_threshold'").Scan(&v)
	if v <= 0 {
		return 5
	}
	return v
}

func (s *EmailPoolService) getMaxRetry() int {
	var v int
	database.DB.QueryRow("SELECT CAST(`value` AS SIGNED) FROM qingka_wangke_config WHERE `key`='email_pool_max_retry'").Scan(&v)
	if v < 0 {
		return 2
	}
	return v
}

// autoResetCounters 每小时重置 hour_sent，每天重置 today_sent
func (s *EmailPoolService) autoResetCounters() {
	now := time.Now()
	h := now.Hour()
	d := now.YearDay()
	if d != s.lastDay {
		database.DB.Exec("UPDATE qingka_email_pool SET today_sent=0, hour_sent=0")
		s.lastDay = d
		s.lastHour = h
	} else if h != s.lastHour {
		database.DB.Exec("UPDATE qingka_email_pool SET hour_sent=0")
		s.lastHour = h
	}
}

// -------- 发送结果回调 --------

// OnSuccess 发送成功后更新计数
func (s *EmailPoolService) OnSuccess(poolID int) {
	now := time.Now().Format("2006-01-02 15:04:05")
	database.DB.Exec(
		"UPDATE qingka_email_pool SET today_sent=today_sent+1, hour_sent=hour_sent+1, total_sent=total_sent+1, fail_streak=0, last_used=? WHERE id=?",
		now, poolID)
}

// OnFail 发送失败后更新计数，连续失败超阈值则自动标异常
func (s *EmailPoolService) OnFail(poolID int, errMsg string) {
	now := time.Now().Format("2006-01-02 15:04:05")
	database.DB.Exec(
		"UPDATE qingka_email_pool SET total_fail=total_fail+1, fail_streak=fail_streak+1, last_error=?, last_used=? WHERE id=?",
		errMsg, now, poolID)
	// 检查是否超过连续失败阈值
	threshold := s.getFailThreshold()
	var streak int
	database.DB.QueryRow("SELECT fail_streak FROM qingka_email_pool WHERE id=?", poolID).Scan(&streak)
	if streak >= threshold {
		database.DB.Exec("UPDATE qingka_email_pool SET status=2 WHERE id=?", poolID) // 标记异常
		log.Printf("[EmailPool] 邮箱 #%d 连续失败 %d 次，已自动标记异常", poolID, streak)
	}
}

// -------- 发送日志 --------

func (s *EmailPoolService) LogSend(poolID int, fromEmail, toEmail, subject, mailType string, success bool, errMsg string) {
	status := 1
	if !success {
		status = 0
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	database.DB.Exec(
		"INSERT INTO qingka_email_send_log (pool_id,from_email,to_email,subject,mail_type,status,error,addtime) VALUES (?,?,?,?,?,?,?,?)",
		poolID, fromEmail, toEmail, subject, mailType, status, errMsg, now)
}

func (s *EmailPoolService) QueryLogs(q model.EmailSendLogQuery) ([]model.EmailSendLog, int64, error) {
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
		queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.EmailSendLog
	for rows.Next() {
		var l model.EmailSendLog
		if err := rows.Scan(&l.ID, &l.PoolID, &l.FromEmail, &l.ToEmail, &l.Subject, &l.MailType, &l.Status, &l.Error, &l.AddTime); err != nil {
			continue
		}
		list = append(list, l)
	}
	if list == nil {
		list = []model.EmailSendLog{}
	}
	return list, total, nil
}

// Stats 统计
func (s *EmailPoolService) Stats() map[string]interface{} {
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

// TestAccount 测试单个池账号
func (s *EmailPoolService) TestAccount(id int, testTo string) error {
	var a model.EmailPoolAccount
	err := database.DB.QueryRow(
		"SELECT id,host,port,encryption,user,password,from_email,name FROM qingka_email_pool WHERE id=?", id,
	).Scan(&a.ID, &a.Host, &a.Port, &a.Encryption, &a.User, &a.Password, &a.FromEmail, &a.Name)
	if err == sql.ErrNoRows {
		return fmt.Errorf("邮箱不存在")
	}
	if err != nil {
		return err
	}
	from := a.User
	if a.FromEmail != "" {
		from = a.FromEmail
	}
	fromName := a.Name
	if fromName == "" {
		fromName = "System"
	}
	msg := emailService.buildMessage(from, fromName, testTo, "邮箱池测试", "<p>这是一封测试邮件，如果您收到说明该邮箱配置正确。</p>")
	addr := fmt.Sprintf("%s:%d", a.Host, a.Port)
	cfg := toSMTPConfig(a)
	switch a.Encryption {
	case "ssl", "tls":
		return emailService.sendSSL(addr, cfg, from, testTo, msg)
	case "starttls":
		return emailService.sendSTARTTLS(addr, cfg, from, testTo, msg)
	default:
		return emailService.sendPlain(addr, cfg, from, testTo, msg)
	}
}
