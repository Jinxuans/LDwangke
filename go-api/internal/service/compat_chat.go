package service

import (
	"database/sql"
	"errors"
	"sync"
	"time"

	"go-api/internal/model"
)

type ChatService struct{}

var chatService = &ChatService{}

func AdminSessions() ([]model.AdminChatSession, error) {
	return chatService.AdminSessions()
}

func AdminMessages(listID, limit int) ([]model.ChatMsg, error) {
	return chatService.AdminMessages(listID, limit)
}

func ArchiveOldMessages() (int64, error) {
	return chatService.ArchiveOldMessages()
}

func TrimSessionMessages() (int64, error) {
	return chatService.TrimSessionMessages()
}

func ChatStats() (map[string]interface{}, error) {
	return chatService.ChatStats()
}

func ManualCleanup(days int) (int64, error) {
	return chatService.ManualCleanup(days)
}

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
