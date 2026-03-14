package service

import (
	"fmt"
	"strings"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

func (s *AdminService) CategoryListAll() ([]model.Category, error) {
	rows, err := database.DB.Query("SELECT id, COALESCE(name,''), COALESCE(status,'1'), COALESCE(sort,0), COALESCE(time,''), COALESCE(recommend,0), COALESCE(log,0), COALESCE(ticket,0), COALESCE(changepass,1), COALESCE(allowpause,0), COALESCE(supplier_report,0), COALESCE(supplier_report_hid,0) FROM qingka_wangke_fenlei ORDER BY sort ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []model.Category
	for rows.Next() {
		var c model.Category
		rows.Scan(&c.ID, &c.Name, &c.Status, &c.Sort, &c.Time, &c.Recommend, &c.Log, &c.Ticket, &c.ChangePass, &c.AllowPause, &c.SupplierReport, &c.SupplierReportHID)
		cats = append(cats, c)
	}
	if cats == nil {
		cats = []model.Category{}
	}
	return cats, nil
}

func (s *AdminService) CategoryListPaged(req model.CategoryListRequest) ([]model.Category, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	where := "1=1"
	args := []interface{}{}
	if req.Keyword != "" {
		where += " AND name LIKE ?"
		args = append(args, "%"+req.Keyword+"%")
	}
	if req.Status != "" {
		where += " AND status = ?"
		args = append(args, req.Status)
	}
	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_fenlei WHERE "+where, args...).Scan(&total)

	offset := (req.Page - 1) * req.Limit
	args2 := append(args, req.Limit, offset)
	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT id, COALESCE(name,''), COALESCE(status,'1'), COALESCE(sort,0), CASE WHEN time IS NOT NULL AND time != '' AND time != '0' THEN FROM_UNIXTIME(CAST(time AS UNSIGNED), '%%Y-%%m-%%d %%H:%%i') ELSE '' END, COALESCE(recommend,0), COALESCE(log,0), COALESCE(ticket,0), COALESCE(changepass,1), COALESCE(allowpause,0), COALESCE(supplier_report,0), COALESCE(supplier_report_hid,0) FROM qingka_wangke_fenlei WHERE %s ORDER BY sort ASC, id ASC LIMIT ? OFFSET ?", where),
		args2...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var cats []model.Category
	for rows.Next() {
		var c model.Category
		rows.Scan(&c.ID, &c.Name, &c.Status, &c.Sort, &c.Time, &c.Recommend, &c.Log, &c.Ticket, &c.ChangePass, &c.AllowPause, &c.SupplierReport, &c.SupplierReportHID)
		cats = append(cats, c)
	}
	if cats == nil {
		cats = []model.Category{}
	}
	return cats, total, nil
}

func (s *AdminService) CategorySave(cat model.Category) error {
	if cat.ID > 0 {
		_, err := database.DB.Exec("UPDATE qingka_wangke_fenlei SET name=?, status=?, sort=?, recommend=?, log=?, ticket=?, changepass=?, allowpause=?, supplier_report=?, supplier_report_hid=? WHERE id=?",
			cat.Name, cat.Status, cat.Sort, cat.Recommend, cat.Log, cat.Ticket, cat.ChangePass, cat.AllowPause, cat.SupplierReport, cat.SupplierReportHID, cat.ID)
		return err
	}
	_, err := database.DB.Exec("INSERT INTO qingka_wangke_fenlei (name, status, sort, time, recommend, log, ticket, changepass, allowpause, supplier_report, supplier_report_hid) VALUES (?, ?, ?, UNIX_TIMESTAMP(), ?, ?, ?, ?, ?, ?, ?)",
		cat.Name, cat.Status, cat.Sort, cat.Recommend, cat.Log, cat.Ticket, cat.ChangePass, cat.AllowPause, cat.SupplierReport, cat.SupplierReportHID)
	return err
}

func (s *AdminService) CategoryDelete(id int) error {
	database.DB.Exec("DELETE FROM qingka_wangke_class WHERE fenlei = ?", id)
	_, err := database.DB.Exec("DELETE FROM qingka_wangke_fenlei WHERE id = ?", id)
	return err
}

func (s *AdminService) CategoryQuickModify(keyword string, categoryID int) (int64, error) {
	result, err := database.DB.Exec("UPDATE qingka_wangke_class SET fenlei = ? WHERE name LIKE ?",
		categoryID, "%"+keyword+"%")
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *AdminService) CategoryUpdateSort(items []struct{ ID, Sort int }) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, item := range items {
		_, err := tx.Exec("UPDATE qingka_wangke_fenlei SET sort = ? WHERE id = ?", item.Sort, item.ID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (s *AdminService) ClassList(cateID int, keywords string, page, limit int) ([]model.ClassManage, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	where := []string{"1=1"}
	args := []interface{}{}
	if cateID > 0 {
		where = append(where, "fenlei = ?")
		args = append(args, cateID)
	}
	if keywords != "" {
		where = append(where, "(name LIKE ? OR cid = ?)")
		args = append(args, "%"+keywords+"%", keywords)
	}

	whereStr := strings.Join(where, " AND ")

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_class WHERE "+whereStr, args...).Scan(&total)

	offset := (page - 1) * limit
	args2 := append(args, limit, offset)
	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT cid, COALESCE(name,''), COALESCE(price,'0'), COALESCE(content,''), COALESCE(fenlei,'0'), COALESCE(status,0), COALESCE(docking,'0'), COALESCE(sort,0), COALESCE(noun,''), COALESCE(yunsuan,'*') FROM qingka_wangke_class WHERE %s ORDER BY sort ASC, cid DESC LIMIT ? OFFSET ?", whereStr),
		args2...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var classes []model.ClassManage
	for rows.Next() {
		var c model.ClassManage
		rows.Scan(&c.CID, &c.Name, &c.Price, &c.Content, &c.CateID, &c.Status, &c.HID, &c.Sort, &c.Noun, &c.Yunsuan)
		classes = append(classes, c)
	}
	if classes == nil {
		classes = []model.ClassManage{}
	}
	return classes, total, nil
}

func (s *AdminService) ClassSave(req model.ClassEditRequest) error {
	if req.Yunsuan == "" {
		req.Yunsuan = "*"
	}
	if req.CID > 0 {
		_, err := database.DB.Exec(
			"UPDATE qingka_wangke_class SET name=?, price=?, content=?, fenlei=?, status=?, docking=?, queryplat=?, sort=?, noun=?, yunsuan=? WHERE cid=?",
			req.Name, req.Price, req.Content, req.CateID, req.Status, req.HID, req.HID, req.Sort, req.Noun, req.Yunsuan, req.CID)
		return err
	}
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_class (name, price, content, fenlei, status, docking, queryplat, noun, getnoun, sort, yunsuan, addtime) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		req.Name, req.Price, req.Content, req.CateID, req.Status, req.HID, req.HID, req.Noun, req.Noun, req.Sort, req.Yunsuan, time.Now().Format("2006-01-02 15:04:05"))
	return err
}

func (s *AdminService) ClassToggleStatus(cid, status int) error {
	_, err := database.DB.Exec("UPDATE qingka_wangke_class SET status = ? WHERE cid = ?", status, cid)
	return err
}

func (s *AdminService) ClassBatchDelete(cids []int) (int64, error) {
	if len(cids) == 0 {
		return 0, nil
	}
	placeholders := ""
	args := make([]interface{}, len(cids))
	for i, id := range cids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}
	result, err := database.DB.Exec("DELETE FROM qingka_wangke_class WHERE cid IN ("+placeholders+")", args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *AdminService) ClassBatchCategory(cids []int, cateID string) (int64, error) {
	if len(cids) == 0 {
		return 0, nil
	}
	placeholders := ""
	args := []interface{}{cateID}
	for i, id := range cids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, id)
	}
	result, err := database.DB.Exec("UPDATE qingka_wangke_class SET fenlei = ? WHERE cid IN ("+placeholders+")", args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *AdminService) ClassBatchPrice(cids []int, rate float64, yunsuan string) (int64, error) {
	if len(cids) == 0 {
		return 0, nil
	}
	placeholders := ""
	args := []interface{}{}
	for i, id := range cids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, id)
	}
	var sql string
	if yunsuan == "+" {
		sql = fmt.Sprintf("UPDATE qingka_wangke_class SET price = CAST(CAST(price AS DECIMAL(10,2)) + %.2f AS CHAR) WHERE cid IN ("+placeholders+")", rate)
	} else {
		sql = fmt.Sprintf("UPDATE qingka_wangke_class SET price = CAST(ROUND(CAST(price AS DECIMAL(10,2)) * %.4f, 2) AS CHAR) WHERE cid IN ("+placeholders+")", rate)
	}
	result, err := database.DB.Exec(sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *AdminService) ClassBatchReplaceKeyword(search, replace, scope, scopeID string) (int64, error) {
	if search == "" {
		return 0, fmt.Errorf("搜索关键词不能为空")
	}
	where := "1=1"
	args := []interface{}{search, replace}
	switch scope {
	case "cate":
		if scopeID != "" {
			where = "fenlei = ?"
			args = append(args, scopeID)
		}
	case "docking":
		if scopeID != "" {
			where = "docking = ?"
			args = append(args, scopeID)
		}
	}
	sql := "UPDATE qingka_wangke_class SET name = REPLACE(name, ?, ?) WHERE " + where
	result, err := database.DB.Exec(sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *AdminService) ClassBatchAddPrefix(prefix, scope, scopeID string) (int64, error) {
	if prefix == "" {
		return 0, fmt.Errorf("前缀不能为空")
	}
	where := "1=1"
	args := []interface{}{prefix}
	switch scope {
	case "cate":
		if scopeID != "" {
			where = "fenlei = ?"
			args = append(args, scopeID)
		}
	case "docking":
		if scopeID != "" {
			where = "docking = ?"
			args = append(args, scopeID)
		}
	}
	where += " AND name NOT LIKE ?"
	args = append(args, prefix+"%")
	sql := "UPDATE qingka_wangke_class SET name = CONCAT(?, name) WHERE " + where
	result, err := database.DB.Exec(sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *AdminService) SupplierList() ([]model.Supplier, error) {
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

func (s *AdminService) SupplierSave(sup model.Supplier) error {
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

func (s *AdminService) SupplierDelete(hid int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_wangke_huoyuan WHERE hid = ?", hid)
	return err
}
