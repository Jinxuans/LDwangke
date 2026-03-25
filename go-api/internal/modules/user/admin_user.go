package user

import (
	"fmt"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
	classmodule "go-api/internal/modules/class"

	"golang.org/x/crypto/bcrypt"
)

type gradeLookupMaps struct {
	byID map[int]string
}

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

	gradeNames := s.loadGradeNameMap()

	offset := (req.Page - 1) * req.Limit
	args2 := append(args, req.Limit, offset)
	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT uid, user, COALESCE(name,''), COALESCE(grade,''), COALESCE(grade_id,0), COALESCE(addprice,1), COALESCE(money,0), COALESCE(yqm,''), COALESCE(addtime,''), COALESCE(active,1) FROM qingka_wangke_user WHERE %s ORDER BY uid DESC LIMIT ? OFFSET ?", where),
		args2...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []model.UserManage
	for rows.Next() {
		var u model.UserManage
		rows.Scan(&u.UID, &u.User, &u.Name, &u.Grade, &u.GradeID, &u.AddPrice, &u.Balance, &u.YQM, &u.AddTime, &u.Status)
		if u.GradeID > 0 {
			u.GradeName = gradeNames.byID[u.GradeID]
		}
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

func (s *Service) loadGradeNameMap() gradeLookupMaps {
	maps := gradeLookupMaps{
		byID: map[int]string{},
	}
	rows, err := database.DB.Query("SELECT id, COALESCE(rate,''), COALESCE(name,'') FROM qingka_wangke_dengji WHERE status = '1' ORDER BY sort ASC")
	if err != nil {
		return maps
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var rate, name string
		rows.Scan(&id, &rate, &name)
		maps.byID[id] = name
	}
	return maps
}

func (s *Service) ResetPassword(uid int, newPass string) error {
	if newPass == "" {
		newPass = "123456"
	}
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = database.DB.Exec("UPDATE qingka_wangke_user SET pass = ? WHERE uid = ?", string(hashedPass), uid)
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

func (s *Service) SetGrade(uid int, gradeID int) error {
	record, err := classmodule.Classes().ResolveSelectedGrade(gradeID, true)
	if err != nil {
		return err
	}
	_, err = database.DB.Exec("UPDATE qingka_wangke_user SET grade_id = ?, addprice = ? WHERE uid = ?", record.ID, record.Rate, uid)
	return err
}
