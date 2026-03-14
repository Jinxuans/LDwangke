package service

import (
	"errors"
	"fmt"
	"time"

	"go-api/internal/database"
)

func (s *ChatService) ArchiveOldMessages() (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -14).Format("2006-01-02 15:04:05")

	res, err := database.DB.Exec(
		"INSERT IGNORE INTO qingka_chat_msg_archive SELECT * FROM qingka_chat_msg WHERE addtime < ? AND status = '已读'",
		cutoff,
	)
	if err != nil {
		return 0, fmt.Errorf("归档失败: %w", err)
	}
	archived, _ := res.RowsAffected()

	database.DB.Exec("DELETE FROM qingka_chat_msg WHERE addtime < ? AND status = '已读'", cutoff)
	return archived, nil
}

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
