package supplier

import (
	"go-api/internal/database"
	"go-api/internal/model"
)

func (s *Service) ListSuppliers() ([]model.Supplier, error) {
	rows, err := database.DB.Query(
		"SELECT hid, COALESCE(pt,''), COALESCE(name,''), COALESCE(url,''), COALESCE(user,''), COALESCE(pass,''), COALESCE(token,''), COALESCE(money,'0'), COALESCE(status,'1'), COALESCE(addtime,'') FROM qingka_wangke_huoyuan ORDER BY hid ASC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Supplier
	for rows.Next() {
		var sup model.Supplier
		rows.Scan(&sup.HID, &sup.PT, &sup.Name, &sup.URL, &sup.User, &sup.Pass, &sup.Token, &sup.Money, &sup.Status, &sup.AddTime)
		list = append(list, sup)
	}
	if list == nil {
		list = []model.Supplier{}
	}
	return list, nil
}

func (s *Service) SaveSupplier(sup model.Supplier) error {
	if sup.HID > 0 {
		_, err := database.DB.Exec(
			"UPDATE qingka_wangke_huoyuan SET name=?, url=?, user=?, pass=?, token=?, pt=?, status=? WHERE hid=?",
			sup.Name, sup.URL, sup.User, sup.Pass, sup.Token, sup.PT, sup.Status, sup.HID)
		return err
	}
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_huoyuan (name, url, user, pass, token, pt, ip, cookie, money, status, addtime, endtime) VALUES (?, ?, ?, ?, ?, ?, '', '', '0', ?, NOW(), '')",
		sup.Name, sup.URL, sup.User, sup.Pass, sup.Token, sup.PT, sup.Status)
	return err
}

func (s *Service) DeleteSupplier(hid int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_wangke_huoyuan WHERE hid = ?", hid)
	return err
}
