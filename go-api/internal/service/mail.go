package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

type MailService struct{}

func NewMailService() *MailService {
	return &MailService{}
}

const mailColumns = "m.id, m.from_uid, COALESCE(u1.name, u1.user, '系统'), m.to_uid, COALESCE(u2.name, u2.user, '用户'), m.title, COALESCE(m.content,''), COALESCE(m.file_url,''), COALESCE(m.file_name,''), m.status, DATE_FORMAT(m.addtime,'%Y-%m-%d %H:%i:%s')"

const mailJoin = "qingka_mail m LEFT JOIN qingka_wangke_user u1 ON m.from_uid = u1.uid LEFT JOIN qingka_wangke_user u2 ON m.to_uid = u2.uid"

func scanMail(rows *sql.Rows) (model.Mail, error) {
	var m model.Mail
	err := rows.Scan(&m.ID, &m.FromUID, &m.FromName, &m.ToUID, &m.ToName, &m.Title, &m.Content, &m.FileURL, &m.FileName, &m.Status, &m.AddTime)
	return m, err
}

func (s *MailService) List(uid int, grade string, req model.MailListRequest) ([]model.Mail, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	where := []string{}
	args := []interface{}{}

	if req.Type == "sent" {
		where = append(where, "m.from_uid = ?")
		args = append(args, uid)
	} else {
		// 收件箱
		where = append(where, "m.to_uid = ?")
		args = append(args, uid)
	}

	whereStr := strings.Join(where, " AND ")

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", mailJoin, whereStr)
	if err := database.DB.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.Limit
	querySQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY m.id DESC LIMIT ? OFFSET ?", mailColumns, mailJoin, whereStr)
	args = append(args, req.Limit, offset)

	rows, err := database.DB.Query(querySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var mails []model.Mail
	for rows.Next() {
		m, err := scanMail(rows)
		if err != nil {
			continue
		}
		mails = append(mails, m)
	}
	if mails == nil {
		mails = []model.Mail{}
	}
	return mails, total, nil
}

func (s *MailService) Detail(uid, mailID int) (*model.Mail, error) {
	querySQL := fmt.Sprintf("SELECT %s FROM %s WHERE m.id = ? AND (m.from_uid = ? OR m.to_uid = ?)", mailColumns, mailJoin)
	rows, err := database.DB.Query(querySQL, mailID, uid, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("邮件不存在")
	}
	m, err := scanMail(rows)
	if err != nil {
		return nil, err
	}

	// 如果是收件人查看，标记已读
	if m.ToUID == uid && m.Status == 0 {
		database.DB.Exec("UPDATE qingka_mail SET status = 1 WHERE id = ?", mailID)
		m.Status = 1
	}

	return &m, nil
}

func (s *MailService) Send(uid int, grade string, req model.MailSendRequest) (int64, error) {
	// 管理员可以发给任何人，普通用户只能发给管理员(uid=1)
	if grade != "2" && grade != "3" && req.ToUID != 1 {
		return 0, errors.New("只能发送给管理员")
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := database.DB.Exec(
		"INSERT INTO qingka_mail (from_uid, to_uid, title, content, file_url, file_name, status, addtime) VALUES (?, ?, ?, ?, ?, ?, 0, ?)",
		uid, req.ToUID, req.Title, req.Content, req.FileURL, req.FileName, now,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *MailService) UnreadCount(uid int) (int, error) {
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM qingka_mail WHERE to_uid = ? AND status = 0", uid).Scan(&count)
	return count, err
}

func (s *MailService) MarkRead(uid, mailID int) error {
	_, err := database.DB.Exec("UPDATE qingka_mail SET status = 1 WHERE id = ? AND to_uid = ?", mailID, uid)
	return err
}

func (s *MailService) Delete(uid, mailID int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_mail WHERE id = ? AND (from_uid = ? OR to_uid = ?)", mailID, uid, uid)
	return err
}
