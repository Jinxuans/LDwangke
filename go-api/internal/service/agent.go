package service

import (
	"strings"
	"time"

	"go-api/internal/database"
)

// 写操作日志 (对应 PHP wlog)
func wlog(uid int, logType, text string, money float64) {
	now := time.Now().Format("2006-01-02 15:04:05")
	var smoney float64
	database.DB.QueryRow("SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&smoney)
	database.DB.Exec(
		"INSERT INTO qingka_wangke_log (uid, type, text, money, smoney, addtime) VALUES (?, ?, ?, ?, ?, ?)",
		uid, logType, text, money, smoney, now,
	)
}

type AgentListItem struct {
	UUID     int     `json:"uuid"`
	Active   int     `json:"active"`
	UID      int     `json:"uid"`
	User     string  `json:"user"`
	Name     string  `json:"name"`
	Money    float64 `json:"money"`
	ZCZ      float64 `json:"zcz"`
	AddPrice float64 `json:"addprice"`
	YQM      string  `json:"yqm"`
	EndTime  string  `json:"endtime"`
	AddTime  string  `json:"addtime"`
	DD       int     `json:"dd"`
	Key      int     `json:"key"`
}

type AgentCreateRequest struct {
	Nickname string `json:"nickname"`
	User     string `json:"user"`
	Pass     string `json:"pass"`
	GradeID  int    `json:"gradeId"`
	Confirm  int    `json:"type"` // 0=预览费用, 1=确认执行
}

type AgentChangeGradeRequest struct {
	UID     int `json:"uid"`
	GradeID int `json:"gradeId"`
	Confirm int `json:"type"` // 0=预览, 1=确认
}

func splitCSV(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		p := strings.TrimSpace(part)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
