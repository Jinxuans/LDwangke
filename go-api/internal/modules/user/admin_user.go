package user

import (
	"fmt"
	"strconv"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

func (s *Service) UserList(req model.UserListRequest) ([]model.UserManage, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	where := "1=1"
	args := []interface{}{}
	if req.Keywords != "" {
		where += " AND (user LIKE ? OR name LIKE ? OR uid = ?)"
		kw := "%" + req.Keywords + "%"
		args = append(args, kw, kw, req.Keywords)
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user WHERE "+where, args...).Scan(&total)

	gradeNameMap := s.loadGradeNameMap()

	offset := (req.Page - 1) * req.Limit
	args2 := append(args, req.Limit, offset)
	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT uid, user, COALESCE(name,''), COALESCE(grade,''), COALESCE(addprice,1), COALESCE(money,0), COALESCE(yqm,''), COALESCE(addtime,''), COALESCE(active,1) FROM qingka_wangke_user WHERE %s ORDER BY uid DESC LIMIT ? OFFSET ?", where),
		args2...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []model.UserManage
	for rows.Next() {
		var u model.UserManage
		rows.Scan(&u.UID, &u.User, &u.Name, &u.Grade, &u.AddPrice, &u.Balance, &u.YQM, &u.AddTime, &u.Status)
		u.GradeName = gradeNameMap[fmt.Sprintf("%.2f", u.AddPrice)]
		if u.GradeName == "" {
			u.GradeName = fmt.Sprintf("费率%.2f", u.AddPrice)
		}
		users = append(users, u)
	}
	if users == nil {
		users = []model.UserManage{}
	}
	return users, total, nil
}

func (s *Service) loadGradeNameMap() map[string]string {
	m := map[string]string{}
	rows, err := database.DB.Query("SELECT COALESCE(rate,''), COALESCE(name,'') FROM qingka_wangke_dengji WHERE status = '1' ORDER BY sort ASC")
	if err != nil {
		return m
	}
	defer rows.Close()
	for rows.Next() {
		var rate, name string
		rows.Scan(&rate, &name)
		m[rate] = name
		if f, err := strconv.ParseFloat(rate, 64); err == nil {
			m[fmt.Sprintf("%.2f", f)] = name
			m[fmt.Sprintf("%g", f)] = name
		}
	}
	return m
}

func (s *Service) ResetPassword(uid int, newPass string) error {
	if newPass == "" {
		newPass = "123456"
	}
	_, err := database.DB.Exec("UPDATE qingka_wangke_user SET pass = ? WHERE uid = ?", newPass, uid)
	return err
}

func (s *Service) SetBalance(uid int, balance float64) error {
	var oldMoney float64
	database.DB.QueryRow("SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&oldMoney)

	_, err := database.DB.Exec("UPDATE qingka_wangke_user SET money = ? WHERE uid = ?", balance, uid)
	if err != nil {
		return err
	}

	diff := balance - oldMoney
	now := time.Now().Format("2006-01-02 15:04:05")
	remark := fmt.Sprintf("管理员调整余额 %.2f → %.2f", oldMoney, balance)
	database.DB.Exec(
		"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, '调整', ?, ?, ?, ?)",
		uid, diff, balance, remark, now,
	)
	return nil
}

func (s *Service) SetGrade(uid int, addprice float64) error {
	_, err := database.DB.Exec("UPDATE qingka_wangke_user SET addprice = ? WHERE uid = ?", addprice, uid)
	return err
}
