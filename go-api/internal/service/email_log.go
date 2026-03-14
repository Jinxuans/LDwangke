package service

import (
	"fmt"
	"log"
	"strings"
	"time"

	"go-api/internal/database"
)

// MassSend 群发邮件，异步发送并记录任务。
func (s *EmailService) MassSend(target, subject, content string) (int64, error) {
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
			if err := s.SendEmail(email, subject, content); err != nil {
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
		log.Printf("[EmailMassSend] 任务 %d 完成: 成功 %d, 失败 %d", id, successCount, failCount)
	}(logID, emails)

	return logID, nil
}

// ResolveRecipients 解析收件人。
func (s *EmailService) ResolveRecipients(target string) ([]string, error) {
	var rowsQuery string
	var args []interface{}

	if target == "all" {
		rowsQuery = "SELECT COALESCE(user,''), COALESCE(email,'') FROM qingka_wangke_user WHERE active = '1'"
	} else if target == "direct" {
		rowsQuery = "SELECT COALESCE(user,''), COALESCE(email,'') FROM qingka_wangke_user WHERE active = '1' AND uuid = 1"
	} else if target == "indirect" {
		rowsQuery = "SELECT COALESCE(user,''), COALESCE(email,'') FROM qingka_wangke_user WHERE active = '1' AND uuid != 1"
	} else if strings.HasPrefix(target, "grade:") {
		grade := strings.TrimPrefix(target, "grade:")
		rowsQuery = "SELECT COALESCE(user,''), COALESCE(email,'') FROM qingka_wangke_user WHERE active = '1' AND grade = ?"
		args = append(args, grade)
	} else if strings.HasPrefix(target, "uids:") {
		uidStr := strings.TrimPrefix(target, "uids:")
		uids := strings.Split(uidStr, ",")
		placeholders := make([]string, len(uids))
		for i, uid := range uids {
			placeholders[i] = "?"
			args = append(args, strings.TrimSpace(uid))
		}
		rowsQuery = fmt.Sprintf("SELECT COALESCE(user,''), COALESCE(email,'') FROM qingka_wangke_user WHERE uid IN (%s)", strings.Join(placeholders, ","))
	} else {
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
			emails = append(emails, addr)
			seen[addr] = true
		}
	}
	return emails, nil
}

// GetSendLogs 获取群发记录。
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
