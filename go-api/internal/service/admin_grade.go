package service

import (
	"errors"
	"fmt"
	"strings"

	"go-api/internal/database"
	"go-api/internal/model"
)

func (s *AdminService) GradeList() ([]model.Grade, error) {
	rows, err := database.DB.Query("SELECT id, COALESCE(sort,'0'), COALESCE(name,''), COALESCE(rate,'1'), COALESCE(money,'0'), COALESCE(addkf,'1'), COALESCE(gjkf,'1'), COALESCE(status,'1'), CASE WHEN time IS NOT NULL AND time != '' AND time != '0' THEN FROM_UNIXTIME(CAST(time AS UNSIGNED), '%Y-%m-%d %H:%i') ELSE '' END FROM qingka_wangke_dengji ORDER BY sort ASC, id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.Grade
	for rows.Next() {
		var g model.Grade
		rows.Scan(&g.ID, &g.Sort, &g.Name, &g.Rate, &g.Money, &g.AddKF, &g.GJKF, &g.Status, &g.Time)
		list = append(list, g)
	}
	if list == nil {
		list = []model.Grade{}
	}
	return list, nil
}

func (s *AdminService) GradeSave(req model.GradeSaveRequest) error {
	if req.Status == "" {
		req.Status = "1"
	}
	if req.Rate == "" {
		req.Rate = "1"
	}
	if req.Money == "" {
		req.Money = "0"
	}
	if req.AddKF == "" {
		req.AddKF = "1"
	}
	if req.GJKF == "" {
		req.GJKF = "1"
	}
	if req.ID > 0 {
		_, err := database.DB.Exec(
			"UPDATE qingka_wangke_dengji SET sort=?, name=?, rate=?, money=?, addkf=?, gjkf=?, status=? WHERE id=?",
			req.Sort, req.Name, req.Rate, req.Money, req.AddKF, req.GJKF, req.Status, req.ID,
		)
		return err
	}
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_dengji (sort, name, rate, money, addkf, gjkf, status, time) VALUES (?, ?, ?, ?, ?, ?, ?, UNIX_TIMESTAMP())",
		req.Sort, req.Name, req.Rate, req.Money, req.AddKF, req.GJKF, req.Status,
	)
	return err
}

func (s *AdminService) GradeDelete(id int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_wangke_dengji WHERE id=?", id)
	return err
}

func (s *AdminService) MiJiaList(req model.MiJiaListRequest) ([]model.MiJia, int64, []int, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	where := "1=1"
	args := []interface{}{}
	if req.UID > 0 {
		where += " AND m.uid = ?"
		args = append(args, req.UID)
	}
	if req.CID > 0 {
		where += " AND m.cid = ?"
		args = append(args, req.CID)
	}
	if req.Keyword != "" {
		where += " AND c.name LIKE ?"
		args = append(args, "%"+req.Keyword+"%")
	}

	var total int64
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_mijia m LEFT JOIN qingka_wangke_class c ON m.cid=c.cid WHERE "+where, countArgs...).Scan(&total)

	offset := (req.Page - 1) * req.Limit
	queryArgs := make([]interface{}, len(args))
	copy(queryArgs, args)
	queryArgs = append(queryArgs, req.Limit, offset)

	rows, err := database.DB.Query(
		"SELECT m.mid, m.uid, m.cid, COALESCE(m.mode,'0'), COALESCE(m.price,'0'), COALESCE(DATE_FORMAT(m.addtime,'%Y-%m-%d %H:%i:%s'),''), COALESCE(u.user,''), COALESCE(c.name,'') FROM qingka_wangke_mijia m LEFT JOIN qingka_wangke_user u ON m.uid=u.uid LEFT JOIN qingka_wangke_class c ON m.cid=c.cid WHERE "+where+" ORDER BY m.mid DESC LIMIT ? OFFSET ?",
		queryArgs...,
	)
	if err != nil {
		return nil, 0, nil, err
	}
	defer rows.Close()
	var list []model.MiJia
	for rows.Next() {
		var mj model.MiJia
		rows.Scan(&mj.MID, &mj.UID, &mj.CID, &mj.Mode, &mj.Price, &mj.AddTime, &mj.UserName, &mj.ClassName)
		list = append(list, mj)
	}
	if list == nil {
		list = []model.MiJia{}
	}

	var uids []int
	uidRows, uidErr := database.DB.Query("SELECT DISTINCT uid FROM qingka_wangke_mijia ORDER BY uid")
	if uidErr == nil {
		defer uidRows.Close()
		for uidRows.Next() {
			var uid int
			uidRows.Scan(&uid)
			uids = append(uids, uid)
		}
	}
	if uids == nil {
		uids = []int{}
	}

	return list, total, uids, nil
}

func (s *AdminService) MiJiaSave(req model.MiJiaSaveRequest) error {
	if req.Mode == "" {
		req.Mode = "0"
	}
	if req.MID > 0 {
		_, err := database.DB.Exec(
			"UPDATE qingka_wangke_mijia SET uid=?, cid=?, mode=?, price=? WHERE mid=?",
			req.UID, req.CID, req.Mode, req.Price, req.MID,
		)
		return err
	}
	var cnt int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_mijia WHERE uid=? AND cid=?", req.UID, req.CID).Scan(&cnt)
	if cnt > 0 {
		_, err := database.DB.Exec(
			"UPDATE qingka_wangke_mijia SET mode=?, price=? WHERE uid=? AND cid=?",
			req.Mode, req.Price, req.UID, req.CID,
		)
		return err
	}
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_mijia (uid, cid, mode, price, addtime) VALUES (?, ?, ?, ?, NOW())",
		req.UID, req.CID, req.Mode, req.Price,
	)
	return err
}

func (s *AdminService) MiJiaDelete(mids []int) error {
	if len(mids) == 0 {
		return errors.New("未指定要删除的ID")
	}
	placeholders := make([]string, len(mids))
	args := make([]interface{}, len(mids))
	for i, id := range mids {
		placeholders[i] = "?"
		args[i] = id
	}
	_, err := database.DB.Exec("DELETE FROM qingka_wangke_mijia WHERE mid IN ("+strings.Join(placeholders, ",")+")", args...)
	return err
}

func (s *AdminService) MiJiaBatch(req model.MiJiaBatchRequest) (int, error) {
	if req.Mode == "" {
		req.Mode = "0"
	}
	rows, err := database.DB.Query("SELECT cid, price FROM qingka_wangke_class WHERE fenlei = ?", req.Fenlei)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var cid int
		var origPrice string
		rows.Scan(&cid, &origPrice)

		finalPrice := req.Price
		finalMode := req.Mode
		if req.Mode == "4" {
			var op, pp float64
			fmt.Sscanf(origPrice, "%f", &op)
			fmt.Sscanf(req.Price, "%f", &pp)
			finalPrice = fmt.Sprintf("%.2f", op*pp)
			finalMode = "2"
		}

		var cnt int
		database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_mijia WHERE uid=? AND cid=?", req.UID, cid).Scan(&cnt)
		if cnt > 0 {
			database.DB.Exec("UPDATE qingka_wangke_mijia SET price=?, mode=? WHERE uid=? AND cid=?", finalPrice, finalMode, req.UID, cid)
		} else {
			database.DB.Exec("INSERT INTO qingka_wangke_mijia (uid, cid, mode, price, addtime) VALUES (?, ?, ?, ?, NOW())", req.UID, cid, finalMode, finalPrice)
		}
		count++
	}
	return count, nil
}
