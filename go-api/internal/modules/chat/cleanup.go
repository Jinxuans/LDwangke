package chat

import (
	"fmt"
	"time"

	"go-api/internal/database"
)

func (s *Service) ArchiveOldMessages() (int64, error) {
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
