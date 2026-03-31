package w

import (
	"fmt"

	"go-api/internal/database"
)

func (s *WService) AdminListApps() ([]WApp, error) {
	rows, err := database.DB.Query("SELECT id, name, code, org_app_id, status, COALESCE(description,''), price, cac_type, url, COALESCE(`key`,''), COALESCE(uid,''), COALESCE(token,''), type, deleted FROM w_app WHERE deleted = 0 ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []WApp
	for rows.Next() {
		var a WApp
		if err := rows.Scan(&a.ID, &a.Name, &a.Code, &a.OrgAppID, &a.Status, &a.Desc, &a.Price, &a.CacType, &a.URL, &a.Key, &a.UID, &a.Token, &a.Type, &a.Deleted); err != nil {
			continue
		}
		list = append(list, a)
	}
	if list == nil {
		list = []WApp{}
	}
	return list, nil
}

func (s *WService) AdminSaveApp(a WApp) (int64, error) {
	if a.Name == "" || a.Code == "" {
		return 0, fmt.Errorf("项目名称和代码不能为空")
	}
	if a.ID > 0 {
		_, err := database.DB.Exec(
			"UPDATE w_app SET name=?, code=?, org_app_id=?, status=?, description=?, price=?, cac_type=?, url=?, `key`=?, uid=?, token=?, type=? WHERE id=?",
			a.Name, a.Code, a.OrgAppID, a.Status, a.Desc, a.Price, a.CacType, a.URL, a.Key, a.UID, a.Token, a.Type, a.ID,
		)
		if err != nil {
			return 0, fmt.Errorf("保存失败: %v", err)
		}
		return a.ID, nil
	}
	result, err := database.DB.Exec(
		"INSERT INTO w_app (name, code, org_app_id, status, description, price, cac_type, url, `key`, uid, token, type) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		a.Name, a.Code, a.OrgAppID, a.Status, a.Desc, a.Price, a.CacType, a.URL, a.Key, a.UID, a.Token, a.Type,
	)
	if err != nil {
		return 0, fmt.Errorf("添加失败: %v", err)
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func (s *WService) AdminDeleteApp(id int64) error {
	_, err := database.DB.Exec("UPDATE w_app SET deleted = 1 WHERE id = ?", id)
	return err
}
