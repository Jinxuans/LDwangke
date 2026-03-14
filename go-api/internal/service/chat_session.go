package service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

func (s *ChatService) Sessions(uid int) ([]model.ChatSession, error) {
	rows, err := database.DB.Query(
		"SELECT list_id, user1, user2, COALESCE(last_msg,''), COALESCE(DATE_FORMAT(last_time,'%Y-%m-%d %H:%i:%s'),''), unread1, unread2 FROM qingka_chat_list WHERE user1 = ? OR user2 = ? ORDER BY last_time DESC",
		uid, uid,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type chatListRow struct {
		model.ChatList
		Unread1 int
		Unread2 int
	}
	var lists []chatListRow
	for rows.Next() {
		var cl chatListRow
		if err := rows.Scan(&cl.ListID, &cl.User1, &cl.User2, &cl.LastMsg, &cl.LastTime, &cl.Unread1, &cl.Unread2); err != nil {
			continue
		}
		lists = append(lists, cl)
	}

	var sessions []model.ChatSession
	for _, cl := range lists {
		unread := cl.Unread1
		if cl.User2 == uid {
			unread = cl.Unread2
		}
		s := model.ChatSession{
			ListID:      cl.ListID,
			LastMsg:     cl.LastMsg,
			LastTime:    cl.LastTime,
			UnreadCount: unread,
		}
		if cl.User1 == uid {
			s.TargetUID = cl.User2
		} else {
			s.TargetUID = cl.User1
		}
		s.TargetName = fmt.Sprintf("%d", s.TargetUID)
		var qqUser, lastActive sql.NullString
		_ = database.DB.QueryRow("SELECT COALESCE(user,''), lasttime FROM qingka_wangke_user WHERE uid = ?", s.TargetUID).Scan(&qqUser, &lastActive)
		if qqUser.Valid && qqUser.String != "" {
			s.Avatar = "https://q1.qlogo.cn/g?b=qq&nk=" + qqUser.String + "&s=100"
		}
		if lastActive.Valid && lastActive.String != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", lastActive.String); err == nil {
				s.Online = time.Since(t) < 5*time.Minute
			}
		}
		sessions = append(sessions, s)
	}
	if sessions == nil {
		sessions = []model.ChatSession{}
	}
	return sessions, nil
}

func (s *ChatService) AdminSessions() ([]model.AdminChatSession, error) {
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
		var s model.AdminChatSession
		var u1qq, u2qq sql.NullString
		if err := rows.Scan(&s.ListID, &s.User1, &s.User2, &s.LastMsg, &s.LastTime,
			&s.UnreadCount, &s.LastFromUID,
			&u1qq, &u2qq); err != nil {
			continue
		}
		if u1qq.Valid && u1qq.String != "" {
			s.User1Avatar = "https://q1.qlogo.cn/g?b=qq&nk=" + u1qq.String + "&s=100"
		}
		if u2qq.Valid && u2qq.String != "" {
			s.User2Avatar = "https://q1.qlogo.cn/g?b=qq&nk=" + u2qq.String + "&s=100"
		}
		totalUnread := s.UnreadCount + s.LastFromUID
		s.UnreadCount = totalUnread
		s.LastFromUID = 0
		database.DB.QueryRow("SELECT COALESCE(from_uid,0) FROM qingka_chat_msg WHERE list_id = ? ORDER BY msg_id DESC LIMIT 1", s.ListID).Scan(&s.LastFromUID)
		s.User1Name = fmt.Sprintf("%d", s.User1)
		s.User2Name = fmt.Sprintf("%d", s.User2)

		var la1 sql.NullString
		_ = database.DB.QueryRow("SELECT lasttime FROM qingka_wangke_user WHERE uid = ?", s.User1).Scan(&la1)
		if la1.Valid && la1.String != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", la1.String); err == nil {
				s.User1Online = time.Since(t) < 5*time.Minute
			}
		}
		var la2 sql.NullString
		_ = database.DB.QueryRow("SELECT lasttime FROM qingka_wangke_user WHERE uid = ?", s.User2).Scan(&la2)
		if la2.Valid && la2.String != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", la2.String); err == nil {
				s.User2Online = time.Since(t) < 5*time.Minute
			}
		}

		sessions = append(sessions, s)
	}

	if sessions == nil {
		sessions = []model.AdminChatSession{}
	}
	return sessions, nil
}

func (s *ChatService) hasAccess(uid, listID int) bool {
	var count int
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM qingka_chat_list WHERE list_id = ? AND (user1 = ? OR user2 = ?)",
		listID, uid, uid,
	).Scan(&count)
	return count > 0
}

func (s *ChatService) getPeerUID(uid, listID int) (int, error) {
	var user1, user2 int
	err := database.DB.QueryRow(
		"SELECT user1, user2 FROM qingka_chat_list WHERE list_id = ? LIMIT 1",
		listID,
	).Scan(&user1, &user2)
	if err == sql.ErrNoRows {
		return 0, errors.New("会话不存在")
	}
	if err != nil {
		return 0, err
	}

	if user1 == uid {
		return user2, nil
	}
	if user2 == uid {
		return user1, nil
	}
	return 0, errors.New("无权访问此聊天")
}
