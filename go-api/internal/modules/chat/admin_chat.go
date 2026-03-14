package chat

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

func (s *Service) AdminSessions() ([]model.AdminChatSession, error) {
	rows, err := database.DB.Query(`
		SELECT cl.list_id, cl.user1, cl.user2,
			COALESCE(cl.last_msg,''), COALESCE(DATE_FORMAT(cl.last_time,'%Y-%m-%d %H:%i:%s'),''),
			cl.unread1, cl.unread2,
			COALESCE(u1.user, ''), COALESCE(u2.user, '')
		FROM qingka_chat_list cl
		LEFT JOIN qingka_wangke_user u1 ON u1.uid = cl.user1
		LEFT JOIN qingka_wangke_user u2 ON u2.uid = cl.user2
		ORDER BY cl.last_time DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []model.AdminChatSession
	for rows.Next() {
		var session model.AdminChatSession
		var u1qq, u2qq sql.NullString
		if err := rows.Scan(&session.ListID, &session.User1, &session.User2, &session.LastMsg, &session.LastTime,
			&session.UnreadCount, &session.LastFromUID,
			&u1qq, &u2qq); err != nil {
			continue
		}
		if u1qq.Valid && u1qq.String != "" {
			session.User1Avatar = "https://q1.qlogo.cn/g?b=qq&nk=" + u1qq.String + "&s=100"
		}
		if u2qq.Valid && u2qq.String != "" {
			session.User2Avatar = "https://q1.qlogo.cn/g?b=qq&nk=" + u2qq.String + "&s=100"
		}
		totalUnread := session.UnreadCount + session.LastFromUID
		session.UnreadCount = totalUnread
		session.LastFromUID = 0
		database.DB.QueryRow("SELECT COALESCE(from_uid,0) FROM qingka_chat_msg WHERE list_id = ? ORDER BY msg_id DESC LIMIT 1", session.ListID).Scan(&session.LastFromUID)
		session.User1Name = fmt.Sprintf("%d", session.User1)
		session.User2Name = fmt.Sprintf("%d", session.User2)

		var la1 sql.NullString
		_ = database.DB.QueryRow("SELECT lasttime FROM qingka_wangke_user WHERE uid = ?", session.User1).Scan(&la1)
		if la1.Valid && la1.String != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", la1.String); err == nil {
				session.User1Online = time.Since(t) < 5*time.Minute
			}
		}
		var la2 sql.NullString
		_ = database.DB.QueryRow("SELECT lasttime FROM qingka_wangke_user WHERE uid = ?", session.User2).Scan(&la2)
		if la2.Valid && la2.String != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", la2.String); err == nil {
				session.User2Online = time.Since(t) < 5*time.Minute
			}
		}

		sessions = append(sessions, session)
	}

	if sessions == nil {
		sessions = []model.AdminChatSession{}
	}
	return sessions, nil
}

func (s *Service) AdminMessages(listID, limit int) ([]model.ChatMsg, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT * FROM (SELECT %s FROM qingka_chat_msg WHERE list_id = ? ORDER BY msg_id DESC LIMIT ?) AS t ORDER BY msg_id ASC", msgColumns),
		listID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var msgs []model.ChatMsg
	for rows.Next() {
		msg, err := scanMsg(rows)
		if err != nil {
			continue
		}
		msgs = append(msgs, msg)
	}
	if msgs == nil {
		msgs = []model.ChatMsg{}
	}
	return msgs, nil
}

func (s *Service) TrimSessionMessages() (int64, error) {
	rows, err := database.DB.Query("SELECT list_id FROM qingka_chat_list")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var totalTrimmed int64
	for rows.Next() {
		var listID int
		rows.Scan(&listID)

		var count int
		database.DB.QueryRow("SELECT COUNT(*) FROM qingka_chat_msg WHERE list_id = ?", listID).Scan(&count)
		if count <= 500 {
			continue
		}

		var cutoffMsgID int
		database.DB.QueryRow(
			"SELECT msg_id FROM qingka_chat_msg WHERE list_id = ? ORDER BY msg_id DESC LIMIT 1 OFFSET 499",
			listID,
		).Scan(&cutoffMsgID)

		if cutoffMsgID > 0 {
			database.DB.Exec(
				"INSERT IGNORE INTO qingka_chat_msg_archive SELECT * FROM qingka_chat_msg WHERE list_id = ? AND msg_id < ?",
				listID, cutoffMsgID,
			)
			res, _ := database.DB.Exec(
				"DELETE FROM qingka_chat_msg WHERE list_id = ? AND msg_id < ?",
				listID, cutoffMsgID,
			)
			n, _ := res.RowsAffected()
			totalTrimmed += n
		}
	}
	return totalTrimmed, nil
}

func (s *Service) ChatStats() (map[string]interface{}, error) {
	stats := map[string]interface{}{}

	var msgCount, archiveCount, sessionCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_chat_msg").Scan(&msgCount)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_chat_msg_archive").Scan(&archiveCount)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_chat_list").Scan(&sessionCount)

	var oldest string
	database.DB.QueryRow("SELECT COALESCE(MIN(addtime),'') FROM qingka_chat_msg").Scan(&oldest)

	stats["msg_count"] = msgCount
	stats["archive_count"] = archiveCount
	stats["session_count"] = sessionCount
	stats["oldest_msg"] = oldest
	return stats, nil
}

func (s *Service) ManualCleanup(days int) (int64, error) {
	if days < 1 {
		return 0, errors.New("天数必须大于0")
	}
	cutoff := time.Now().AddDate(0, 0, -days).Format("2006-01-02 15:04:05")

	res, err := database.DB.Exec(
		"INSERT IGNORE INTO qingka_chat_msg_archive SELECT * FROM qingka_chat_msg WHERE addtime < ? AND status = '已读'",
		cutoff,
	)
	if err != nil {
		return 0, err
	}
	archived, _ := res.RowsAffected()
	database.DB.Exec("DELETE FROM qingka_chat_msg WHERE addtime < ? AND status = '已读'", cutoff)
	return archived, nil
}
