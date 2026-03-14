package service

import (
	"go-api/internal/database"
	"go-api/internal/model"
)

func (s *AuxiliaryService) activityList(req model.ActivityListRequest) ([]model.Activity, int, error) {
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

func (s *AuxiliaryService) ActivitySave(req model.ActivitySaveRequest) error {
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

func (s *AuxiliaryService) ActivityDelete(hid int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_wangke_huodong WHERE hid = ?", hid)
	return err
}
