package chat

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

type Service struct{}

var chatService = NewService()

func NewService() *Service {
	return &Service{}
}

func Chat() *Service {
	return chatService
}

var chatRateMap sync.Map

type chatRateEntry struct {
	mu    sync.Mutex
	times []time.Time
}

const msgColumns = "msg_id, list_id, from_uid, to_uid, COALESCE(content,''), COALESCE(img,''), COALESCE(status,'未读'), COALESCE(DATE_FORMAT(addtime,'%Y-%m-%d %H:%i:%s'),'')"

func scanMsg(rows *sql.Rows) (model.ChatMsg, error) {
	var m model.ChatMsg
	err := rows.Scan(&m.MsgID, &m.ListID, &m.FromUID, &m.ToUID, &m.Content, &m.Img, &m.Status, &m.AddTime)
	return m, err
}

func (s *Service) Sessions(uid int) ([]model.ChatSession, error) {
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
		item := model.ChatSession{
			ListID:      cl.ListID,
			LastMsg:     cl.LastMsg,
			LastTime:    cl.LastTime,
			UnreadCount: unread,
		}
		if cl.User1 == uid {
			item.TargetUID = cl.User2
		} else {
			item.TargetUID = cl.User1
		}
		item.TargetName = fmt.Sprintf("%d", item.TargetUID)
		var qqUser, lastActive sql.NullString
		_ = database.DB.QueryRow("SELECT COALESCE(user,''), lasttime FROM qingka_wangke_user WHERE uid = ?", item.TargetUID).Scan(&qqUser, &lastActive)
		if qqUser.Valid && qqUser.String != "" {
			item.Avatar = "https://q1.qlogo.cn/g?b=qq&nk=" + qqUser.String + "&s=100"
		}
		if lastActive.Valid && lastActive.String != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", lastActive.String); err == nil {
				item.Online = time.Since(t) < 5*time.Minute
			}
		}
		sessions = append(sessions, item)
	}
	if sessions == nil {
		sessions = []model.ChatSession{}
	}
	return sessions, nil
}

func (s *Service) Messages(uid, listID int, limit int) ([]model.ChatMsg, error) {
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

func (s *Service) History(uid, listID, beforeID, limit int) ([]model.ChatMsg, error) {
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

func (s *Service) NewMessages(uid, listID, afterID int) ([]model.ChatMsg, error) {
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

func (s *Service) Send(uid int, req model.ChatSendRequest) (*model.ChatMsg, error) {
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

func (s *Service) SendImage(uid int, req model.ChatSendImageRequest, imageURL string) (*model.ChatMsg, error) {
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

func (s *Service) MarkRead(uid, listID int) error {
	_, err := database.DB.Exec("UPDATE qingka_chat_msg SET status = '已读' WHERE list_id = ? AND to_uid = ?", listID, uid)
	if err == nil {
		s.resetUnread(listID, uid)
	}
	return err
}

func (s *Service) UnreadTotal(uid int) (int, error) {
	var total int
	err := database.DB.QueryRow(
		"SELECT COALESCE(SUM(CASE WHEN user1 = ? THEN unread1 WHEN user2 = ? THEN unread2 ELSE 0 END), 0) FROM qingka_chat_list WHERE user1 = ? OR user2 = ?",
		uid, uid, uid, uid,
	).Scan(&total)
	return total, err
}

func (s *Service) CreateChat(uid int, targetUID int) (int, error) {
	if targetUID == uid {
		return 0, errors.New("不能与自己创建聊天")
	}

	var listID int
	err := database.DB.QueryRow(
		"SELECT list_id FROM qingka_chat_list WHERE (user1 = ? AND user2 = ?) OR (user1 = ? AND user2 = ?) LIMIT 1",
		uid, targetUID, targetUID, uid,
	).Scan(&listID)
	if err == nil {
		return listID, nil
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := database.DB.Exec(
		"INSERT INTO qingka_chat_list (user1, user2, last_msg, last_time) VALUES (?, ?, ' ', ?)",
		uid, targetUID, now,
	)
	if err != nil {
		return 0, err
	}

	id, _ := result.LastInsertId()
	return int(id), nil
}

func (s *Service) checkRate(uid int) error {
	val, _ := chatRateMap.LoadOrStore(uid, &chatRateEntry{})
	entry := val.(*chatRateEntry)
	entry.mu.Lock()
	defer entry.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Minute)
	filtered := entry.times[:0]
	for _, t := range entry.times {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	entry.times = filtered

	if len(entry.times) >= 20 {
		return errors.New("发送过于频繁，请稍后再试")
	}
	entry.times = append(entry.times, now)
	return nil
}

func (s *Service) hasAccess(uid, listID int) bool {
	var count int
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM qingka_chat_list WHERE list_id = ? AND (user1 = ? OR user2 = ?)",
		listID, uid, uid,
	).Scan(&count)
	return count > 0
}

func (s *Service) getPeerUID(uid, listID int) (int, error) {
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

func (s *Service) incrUnread(listID, toUID int) {
	database.DB.Exec(
		"UPDATE qingka_chat_list SET unread1 = CASE WHEN user1 = ? THEN unread1 + 1 ELSE unread1 END, unread2 = CASE WHEN user2 = ? THEN unread2 + 1 ELSE unread2 END WHERE list_id = ?",
		toUID, toUID, listID,
	)
}

func (s *Service) resetUnread(listID, uid int) {
	database.DB.Exec(
		"UPDATE qingka_chat_list SET unread1 = CASE WHEN user1 = ? THEN 0 ELSE unread1 END, unread2 = CASE WHEN user2 = ? THEN 0 ELSE unread2 END WHERE list_id = ?",
		uid, uid, listID,
	)
}
