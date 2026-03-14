package service

import (
	"go-api/internal/database"
	"go-api/internal/model"
)

func (s *AuxiliaryService) PledgeConfigList() ([]model.PledgeConfig, error) {
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

func (s *AuxiliaryService) PledgeConfigSave(req model.PledgeConfigSaveRequest) error {
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

func (s *AuxiliaryService) PledgeConfigDelete(id int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_wangke_zhiya_config WHERE id = ?", id)
	return err
}

func (s *AuxiliaryService) PledgeConfigToggle(id, status int) error {
	_, err := database.DB.Exec("UPDATE qingka_wangke_zhiya_config SET status = ? WHERE id = ?", status, id)
	return err
}

func (s *AuxiliaryService) pledgeRecordList(req model.PledgeListRequest) ([]model.PledgeRecord, int, error) {
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
