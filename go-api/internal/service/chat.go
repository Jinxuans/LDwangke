package service

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

type ChatService struct{}

func NewChatService() *ChatService {
	return &ChatService{}
}

// 聊天发送频率限制：每用户每分钟最多20条
var chatRateMap sync.Map

type chatRateEntry struct {
	mu    sync.Mutex
	times []time.Time
}

func (s *ChatService) checkRate(uid int) error {
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

const msgColumns = "msg_id, list_id, from_uid, to_uid, COALESCE(content,''), COALESCE(img,''), COALESCE(status,'未读'), COALESCE(DATE_FORMAT(addtime,'%Y-%m-%d %H:%i:%s'),'')"

func scanMsg(rows *sql.Rows) (model.ChatMsg, error) {
	var m model.ChatMsg
	err := rows.Scan(&m.MsgID, &m.ListID, &m.FromUID, &m.ToUID, &m.Content, &m.Img, &m.Status, &m.AddTime)
	return m, err
}

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
		// 查QQ头像和在线状态，名字用UID
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
			&s.UnreadCount, &s.LastFromUID, // reuse fields temporarily
			&u1qq, &u2qq); err != nil {
			continue
		}
		if u1qq.Valid && u1qq.String != "" {
			s.User1Avatar = "https://q1.qlogo.cn/g?b=qq&nk=" + u1qq.String + "&s=100"
		}
		if u2qq.Valid && u2qq.String != "" {
			s.User2Avatar = "https://q1.qlogo.cn/g?b=qq&nk=" + u2qq.String + "&s=100"
		}
		// unread_count = unread1 + unread2 (total unread across both users)
		totalUnread := s.UnreadCount + s.LastFromUID
		s.UnreadCount = totalUnread

		// 查最后发送者
		s.LastFromUID = 0
		database.DB.QueryRow("SELECT COALESCE(from_uid,0) FROM qingka_chat_msg WHERE list_id = ? ORDER BY msg_id DESC LIMIT 1", s.ListID).Scan(&s.LastFromUID)

		// 名字用UID
		s.User1Name = fmt.Sprintf("%d", s.User1)
		s.User2Name = fmt.Sprintf("%d", s.User2)

		// 在线状态（lasttime列可能不存在，静默处理）
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

	// 权限校验
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

	// 标记已读
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

	// 标记已读
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

	// 更新会话 + 未读计数
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

// ========== 定时归档/清理 ==========

// ArchiveOldMessages 将14天以上的已读消息归档
func (s *ChatService) ArchiveOldMessages() (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -14).Format("2006-01-02 15:04:05")

	// 移到归档表
	res, err := database.DB.Exec(
		"INSERT IGNORE INTO qingka_chat_msg_archive SELECT * FROM qingka_chat_msg WHERE addtime < ? AND status = '已读'",
		cutoff,
	)
	if err != nil {
		return 0, fmt.Errorf("归档失败: %w", err)
	}
	archived, _ := res.RowsAffected()

	// 从主表删除
	database.DB.Exec("DELETE FROM qingka_chat_msg WHERE addtime < ? AND status = '已读'", cutoff)

	return archived, nil
}

// TrimSessionMessages 每个会话只保留最近500条
func (s *ChatService) TrimSessionMessages() (int64, error) {
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

		// 找到第500条的msg_id作为截断点
		var cutoffMsgID int
		database.DB.QueryRow(
			"SELECT msg_id FROM qingka_chat_msg WHERE list_id = ? ORDER BY msg_id DESC LIMIT 1 OFFSET 499",
			listID,
		).Scan(&cutoffMsgID)

		if cutoffMsgID > 0 {
			// 归档超出的消息
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

// ChatStats 返回聊天数据统计
func (s *ChatService) ChatStats() (map[string]interface{}, error) {
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

// ManualCleanup 管理员手动清理指定天数前的已读消息
func (s *ChatService) ManualCleanup(days int) (int64, error) {
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

func (s *ChatService) CreateChat(uid int, targetUID int) (int, error) {
	if targetUID == uid {
		return 0, errors.New("不能与自己创建聊天")
	}

	// 检查是否已存在
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
