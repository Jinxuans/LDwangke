package auxiliary

import (
	"crypto/rand"
	"math/big"
	"strings"

	"go-api/internal/database"
	"go-api/internal/model"
)

func (s *Service) CardKeyList(req model.CardKeyListRequest) ([]model.CardKey, int, error) {
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

func (s *Service) CardKeyGenerate(money, count int) ([]string, error) {
	keys := make([]string, 0, count)
	for i := 0; i < count; i++ {
		code := generateRandomCode(20)
		keys = append(keys, code)
		if _, err := database.DB.Exec(
			"INSERT INTO qingka_wangke_km (content, money, status, addtime) VALUES (?, ?, 0, NOW())",
			code, money,
		); err != nil {
			return keys, err
		}
	}
	return keys, nil
}

func (s *Service) CardKeyDelete(ids []int) (int, error) {
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

func (s *Service) ActivityList(req model.ActivityListRequest) ([]model.Activity, int, error) {
	var total int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_huodong").Scan(&total); err != nil {
		return nil, 0, err
	}

	page := normalizeListPage(req.Page)
	limit := normalizeListLimit(req.Limit)
	offset := (page - 1) * limit
	rows, err := database.DB.Query(
		"SELECT hid, name, COALESCE(yaoqiu,''), COALESCE(type,''), COALESCE(num,''), COALESCE(money,''), COALESCE(DATE_FORMAT(addtime,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(endtime,'%Y-%m-%d %H:%i:%s'),''), COALESCE(status_ok,''), COALESCE(status,'') FROM qingka_wangke_huodong ORDER BY hid DESC LIMIT ?, ?",
		offset, limit,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.Activity
	for rows.Next() {
		var item model.Activity
		if err := rows.Scan(&item.HID, &item.Name, &item.YaoQiu, &item.Type, &item.Num, &item.Money, &item.AddTime, &item.EndTime, &item.StatusOK, &item.Status); err == nil {
			list = append(list, item)
		}
	}
	if list == nil {
		list = []model.Activity{}
	}
	return list, total, nil
}

func (s *Service) ActivitySave(req model.ActivitySaveRequest) error {
	if req.HID > 0 {
		_, err := database.DB.Exec(
			"UPDATE qingka_wangke_huodong SET name = ?, yaoqiu = ?, type = ?, num = ?, money = ?, addtime = ?, endtime = ?, status_ok = ? WHERE hid = ?",
			req.Name, req.YaoQiu, req.Type, req.Num, req.Money, req.AddTime, req.EndTime, req.StatusOK, req.HID,
		)
		return err
	}
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_huodong (name, yaoqiu, type, num, money, addtime, endtime, status_ok) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		req.Name, req.YaoQiu, req.Type, req.Num, req.Money, req.AddTime, req.EndTime, req.StatusOK,
	)
	return err
}

func (s *Service) ActivityDelete(hid int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_wangke_huodong WHERE hid = ?", hid)
	return err
}

func (s *Service) PledgeConfigList() ([]model.PledgeConfig, error) {
	rows, err := database.DB.Query(
		`SELECT c.id, c.category_id, c.amount, c.discount_rate, c.status, COALESCE(DATE_FORMAT(c.addtime,'%Y-%m-%d %H:%i:%s'),''), c.days, c.cancel_fee,
		        COALESCE(f.name,'')
		   FROM qingka_wangke_zhiya_config c
		   LEFT JOIN qingka_wangke_fenlei f ON c.category_id = f.id
		  ORDER BY c.id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.PledgeConfig
	for rows.Next() {
		var item model.PledgeConfig
		if err := rows.Scan(&item.ID, &item.CategoryID, &item.Amount, &item.DiscountRate, &item.Status, &item.AddTime, &item.Days, &item.CancelFee, &item.CategoryName); err == nil {
			list = append(list, item)
		}
	}
	if list == nil {
		list = []model.PledgeConfig{}
	}
	return list, nil
}

func (s *Service) PledgeConfigSave(req model.PledgeConfigSaveRequest) error {
	if req.ID > 0 {
		_, err := database.DB.Exec(
			"UPDATE qingka_wangke_zhiya_config SET category_id = ?, amount = ?, discount_rate = ?, days = ?, cancel_fee = ? WHERE id = ?",
			req.CategoryID, req.Amount, req.DiscountRate, req.Days, req.CancelFee, req.ID,
		)
		return err
	}
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_zhiya_config (category_id, amount, discount_rate, status, addtime, days, cancel_fee) VALUES (?, ?, ?, 1, NOW(), ?, ?)",
		req.CategoryID, req.Amount, req.DiscountRate, req.Days, req.CancelFee,
	)
	return err
}

func (s *Service) PledgeConfigDelete(id int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_wangke_zhiya_config WHERE id = ?", id)
	return err
}

func (s *Service) PledgeConfigToggle(id, status int) error {
	_, err := database.DB.Exec("UPDATE qingka_wangke_zhiya_config SET status = ? WHERE id = ?", status, id)
	return err
}

func (s *Service) PledgeRecordList(req model.PledgeListRequest) ([]model.PledgeRecord, int, error) {
	where := "1=1"
	args := []interface{}{}
	if req.UID > 0 {
		where += " AND r.uid = ?"
		args = append(args, req.UID)
	}

	var total int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_zhiya_records r WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	page := normalizeListPage(req.Page)
	limit := normalizeListLimit(req.Limit)
	offset := (page - 1) * limit
	rows, err := database.DB.Query(
		`SELECT r.id, r.uid, r.config_id, r.status, COALESCE(DATE_FORMAT(r.addtime,'%Y-%m-%d %H:%i:%s'),''), DATE_FORMAT(r.endtime,'%Y-%m-%d %H:%i:%s'),
		        c.amount, COALESCE(f.name,''), c.discount_rate, c.days, COALESCE(u.user,'')
		   FROM qingka_wangke_zhiya_records r
		   LEFT JOIN qingka_wangke_zhiya_config c ON r.config_id = c.id
		   LEFT JOIN qingka_wangke_fenlei f ON c.category_id = f.id
		   LEFT JOIN qingka_wangke_user u ON r.uid = u.uid
		  WHERE `+where+` ORDER BY r.id DESC LIMIT ?, ?`,
		append(args, offset, limit)...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.PledgeRecord
	for rows.Next() {
		var item model.PledgeRecord
		if err := rows.Scan(&item.ID, &item.UID, &item.ConfigID, &item.Status, &item.AddTime, &item.EndTime, &item.Amount, &item.CatName, &item.Discount, &item.Days, &item.Username); err == nil {
			list = append(list, item)
		}
	}
	if list == nil {
		list = []model.PledgeRecord{}
	}
	return list, total, nil
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
