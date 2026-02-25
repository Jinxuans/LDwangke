package model

// EmailPoolAccount 邮箱轮询池账号
type EmailPoolAccount struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Encryption string `json:"encryption"`
	User       string `json:"user"`
	Password   string `json:"password,omitempty"`
	FromEmail  string `json:"from_email"`
	Weight     int    `json:"weight"`
	DayLimit   int    `json:"day_limit"`
	HourLimit  int    `json:"hour_limit"`
	TodaySent  int    `json:"today_sent"`
	HourSent   int    `json:"hour_sent"`
	TotalSent  int    `json:"total_sent"`
	TotalFail  int    `json:"total_fail"`
	FailStreak int    `json:"fail_streak"`
	Status     int    `json:"status"` // 1=启用 0=禁用 2=异常
	LastUsed   string `json:"last_used"`
	LastError  string `json:"last_error"`
	AddTime    string `json:"addtime"`
}

// EmailPoolSaveRequest 新增/编辑邮箱池账号
type EmailPoolSaveRequest struct {
	ID         int    `json:"id"`
	Name       string `json:"name" binding:"required"`
	Host       string `json:"host" binding:"required"`
	Port       int    `json:"port" binding:"required"`
	Encryption string `json:"encryption"`
	User       string `json:"user" binding:"required"`
	Password   string `json:"password"`
	FromEmail  string `json:"from_email"`
	Weight     int    `json:"weight"`
	DayLimit   int    `json:"day_limit"`
	HourLimit  int    `json:"hour_limit"`
	Status     int    `json:"status"`
}

// EmailSendLog 邮件发送明细
type EmailSendLog struct {
	ID        int64  `json:"id"`
	PoolID    int    `json:"pool_id"`
	FromEmail string `json:"from_email"`
	ToEmail   string `json:"to_email"`
	Subject   string `json:"subject"`
	MailType  string `json:"mail_type"`
	Status    int    `json:"status"` // 1=成功 0=失败
	Error     string `json:"error"`
	AddTime   string `json:"addtime"`
}

// EmailSendLogQuery 日志查询
type EmailSendLogQuery struct {
	Page     int    `json:"page" form:"page"`
	Limit    int    `json:"limit" form:"limit"`
	MailType string `json:"mail_type" form:"mail_type"`
	Status   int    `json:"status" form:"status"` // -1=全部 0=失败 1=成功
	ToEmail  string `json:"to_email" form:"to_email"`
}
