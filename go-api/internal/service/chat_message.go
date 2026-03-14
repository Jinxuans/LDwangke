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

func (s *ChatService) AdminMessages(listID, limit int) ([]model.ChatMsg, error) {
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
		m, err := scanMsg(rows)
		if err != nil {
			continue
		}
		msgs = append(msgs, m)
	}
	if msgs == nil {
		msgs = []model.ChatMsg{}
	}
	return msgs, nil
}

func (s *ChatService) Messages(uid, listID int, limit int) ([]model.ChatMsg, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if !s.hasAccess(uid, listID) {
		return nil, errors.New("无权访问此聊天")
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
		m, err := scanMsg(rows)
		if err != nil {
			continue
		}
		msgs = append(msgs, m)
	}
	if msgs == nil {
		msgs = []model.ChatMsg{}
	}

	database.DB.Exec("UPDATE qingka_chat_msg SET status = '已读' WHERE list_id = ? AND to_uid = ?", listID, uid)
	return msgs, nil
}

func (s *ChatService) NewMessages(uid, listID, afterID int) ([]model.ChatMsg, error) {
	if !s.hasAccess(uid, listID) {
		return nil, errors.New("无权访问此聊天")
	}

	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT %s FROM qingka_chat_msg WHERE list_id = ? AND msg_id > ? ORDER BY msg_id ASC LIMIT 100", msgColumns),
		listID, afterID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []model.ChatMsg
	for rows.Next() {
		m, err := scanMsg(rows)
		if err != nil {
			continue
		}
		msgs = append(msgs, m)
	}
	if msgs == nil {
		msgs = []model.ChatMsg{}
	}

	database.DB.Exec("UPDATE qingka_chat_msg SET status = '已读' WHERE list_id = ? AND to_uid = ? AND msg_id > ?", listID, uid, afterID)
	return msgs, nil
}

func (s *ChatService) History(uid, listID, beforeID, limit int) ([]model.ChatMsg, error) {
	if !s.hasAccess(uid, listID) {
		return nil, errors.New("无权访问此聊天")
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var (
		rows *sql.Rows
		err  error
	)
	if beforeID > 0 {
		rows, err = database.DB.Query(
			fmt.Sprintf("SELECT %s FROM qingka_chat_msg WHERE list_id = ? AND msg_id < ? ORDER BY msg_id DESC LIMIT ?", msgColumns),
			listID, beforeID, limit,
		)
	} else {
		rows, err = database.DB.Query(
			fmt.Sprintf("SELECT %s FROM qingka_chat_msg WHERE list_id = ? ORDER BY msg_id DESC LIMIT ?", msgColumns),
			listID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []model.ChatMsg
	for rows.Next() {
		m, err := scanMsg(rows)
		if err != nil {
			continue
		}
		msgs = append(msgs, m)
	}
	if len(msgs) == 0 {
		return []model.ChatMsg{}, nil
	}

	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

func (s *ChatService) Send(uid int, req model.ChatSendRequest) (*model.ChatMsg, error) {
	if err := s.checkRate(uid); err != nil {
		return nil, err
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, errors.New("消息不能为空")
	}
	if len(content) > 5000 {
		return nil, errors.New("消息过长")
	}

	peerUID, err := s.getPeerUID(uid, req.ListID)
	if err != nil {
		return nil, err
	}
	if req.ToUID != peerUID {
		return nil, errors.New("无效的接收者")
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := database.DB.Exec(
		"INSERT INTO qingka_chat_msg (list_id, from_uid, to_uid, content, status, addtime) VALUES (?, ?, ?, ?, '未读', ?)",
		req.ListID, uid, req.ToUID, content, now,
	)
	if err != nil {
		return nil, err
	}

	msgID, _ := result.LastInsertId()
	database.DB.Exec("UPDATE qingka_chat_list SET last_msg = ?, last_time = ? WHERE list_id = ?", content, now, req.ListID)
	s.incrUnread(req.ListID, req.ToUID)

	return &model.ChatMsg{
		MsgID:   int(msgID),
		ListID:  req.ListID,
		FromUID: uid,
		ToUID:   req.ToUID,
		Content: content,
		Status:  "未读",
		AddTime: now,
	}, nil
}

func (s *ChatService) SendImage(uid int, req model.ChatSendImageRequest, imageURL string) (*model.ChatMsg, error) {
	if err := s.checkRate(uid); err != nil {
		return nil, err
	}
	if imageURL == "" {
		return nil, errors.New("图片地址无效")
	}

	peerUID, err := s.getPeerUID(uid, req.ListID)
	if err != nil {
		return nil, err
	}
	if req.ToUID != peerUID {
		return nil, errors.New("无效的接收者")
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := database.DB.Exec(
		"INSERT INTO qingka_chat_msg (list_id, from_uid, to_uid, img, status, addtime) VALUES (?, ?, ?, ?, '未读', ?)",
		req.ListID, uid, req.ToUID, imageURL, now,
	)
	if err != nil {
		return nil, err
	}
	msgID, _ := result.LastInsertId()

	database.DB.Exec("UPDATE qingka_chat_list SET last_msg = ?, last_time = ? WHERE list_id = ?", "[图片]", now, req.ListID)
	s.incrUnread(req.ListID, req.ToUID)

	return &model.ChatMsg{
		MsgID:   int(msgID),
		ListID:  req.ListID,
		FromUID: uid,
		ToUID:   req.ToUID,
		Img:     imageURL,
		Status:  "未读",
		AddTime: now,
	}, nil
}

func (s *ChatService) MarkRead(uid, listID int) error {
	_, err := database.DB.Exec("UPDATE qingka_chat_msg SET status = '已读' WHERE list_id = ? AND to_uid = ?", listID, uid)
	if err == nil {
		s.resetUnread(listID, uid)
	}
	return err
}

func (s *ChatService) UnreadTotal(uid int) (int, error) {
	var total int
	err := database.DB.QueryRow(
		"SELECT COALESCE(SUM(CASE WHEN user1 = ? THEN unread1 WHEN user2 = ? THEN unread2 ELSE 0 END), 0) FROM qingka_chat_list WHERE user1 = ? OR user2 = ?",
		uid, uid, uid, uid,
	).Scan(&total)
	return total, err
}

func (s *ChatService) incrUnread(listID, toUID int) {
	database.DB.Exec(
		"UPDATE qingka_chat_list SET unread1 = CASE WHEN user1 = ? THEN unread1 + 1 ELSE unread1 END, unread2 = CASE WHEN user2 = ? THEN unread2 + 1 ELSE unread2 END WHERE list_id = ?",
		toUID, toUID, listID,
	)
}

func (s *ChatService) resetUnread(listID, uid int) {
	database.DB.Exec(
		"UPDATE qingka_chat_list SET unread1 = CASE WHEN user1 = ? THEN 0 ELSE unread1 END, unread2 = CASE WHEN user2 = ? THEN 0 ELSE unread2 END WHERE list_id = ?",
		uid, uid, listID,
	)
}
