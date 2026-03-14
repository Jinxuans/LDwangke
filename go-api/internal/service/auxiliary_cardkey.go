package service

import (
	"crypto/rand"
	"math/big"
	"strings"

	"go-api/internal/database"
	"go-api/internal/model"
)

func (s *AuxiliaryService) CardKeyGenerate(money, count int) ([]string, error) {
	keys := make([]string, 0, count)
	for i := 0; i < count; i++ {
		code := generateRandomCode(20)
		keys = append(keys, code)
		_, err := database.DB.Exec(
			"INSERT INTO qingka_wangke_km (content, money, status, addtime) VALUES (?, ?, 0, NOW())",
			code, money,
		)
		if err != nil {
			return keys, err
		}
	}
	return keys, nil
}

func (s *AuxiliaryService) cardKeyList(req model.CardKeyListRequest) ([]model.CardKey, int, error) {
	where := "1=1"
	args := []interface{}{}
	if req.Status == 0 || req.Status == 1 {
		where += " AND status = ?"
		args = append(args, req.Status)
	}

	var total int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_km WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	page := normalizeListPage(req.Page)
	limit := normalizeListLimit(req.Limit)
	offset := (page - 1) * limit
	rows, err := database.DB.Query(
		"SELECT id, content, money, status, uid, COALESCE(DATE_FORMAT(addtime,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(usedtime,'%Y-%m-%d %H:%i:%s'),'') FROM qingka_wangke_km WHERE "+where+" ORDER BY id DESC LIMIT ?, ?",
		append(args, offset, limit)...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.CardKey
	for rows.Next() {
		var item model.CardKey
		if err := rows.Scan(&item.ID, &item.Content, &item.Money, &item.Status, &item.UID, &item.AddTime, &item.UsedTime); err == nil {
			list = append(list, item)
		}
	}
	if list == nil {
		list = []model.CardKey{}
	}
	return list, total, nil
}

func (s *AuxiliaryService) CardKeyDelete(ids []int) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	res, err := database.DB.Exec("DELETE FROM qingka_wangke_km WHERE id IN ("+strings.Join(placeholders, ",")+")", args...)
	if err != nil {
		return 0, err
	}
	affected, _ := res.RowsAffected()
	return int(affected), nil
}

func generateRandomCode(length int) string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var b strings.Builder
	for i := 0; i < length; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b.WriteByte(chars[n.Int64()])
	}
	return b.String()
}
