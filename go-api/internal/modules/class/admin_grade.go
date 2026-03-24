package class

import (
	"database/sql"
	"errors"
	"go-api/internal/database"
	"go-api/internal/model"
	"strconv"
	"strings"
)

func (s *classService) GradeList() ([]model.Grade, error) {
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

func (s *classService) GradeSave(req model.GradeSaveRequest) error {
	req.Name = strings.TrimSpace(req.Name)
	req.Rate = strings.TrimSpace(req.Rate)
	req.Money = strings.TrimSpace(req.Money)
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
	rate, err := strconv.ParseFloat(req.Rate, 64)
	if err != nil || rate <= 0 {
		return errors.New("等级费率必须大于0")
	}
	req.Rate = formatGradeRate(rate)
	if req.Money != "" {
		money, err := strconv.ParseFloat(req.Money, 64)
		if err != nil || money < 0 {
			return errors.New("开通价格不能小于0")
		}
		req.Money = formatGradeRate(money)
	}
	if duplicated, err := s.GradeRateExists(rate, req.ID); err != nil {
		return err
	} else if duplicated {
		return errors.New("等级费率不能重复")
	}
	if duplicated, err := s.GradeNameExists(req.Name, req.ID); err != nil {
		return err
	} else if duplicated {
		return errors.New("等级名称不能重复")
	}
	if req.ID > 0 {
		if req.Status != "1" {
			assigned, err := s.CountGradeAssignments(req.ID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			if assigned > 0 {
				return errors.New("该等级仍有用户使用，不能禁用")
			}
		}
		tx, err := database.DB.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		_, err = tx.Exec(
			"UPDATE qingka_wangke_dengji SET sort=?, name=?, rate=?, money=?, addkf=?, gjkf=?, status=? WHERE id=?",
			req.Sort, req.Name, req.Rate, req.Money, req.AddKF, req.GJKF, req.Status, req.ID,
		)
		if err != nil {
			return err
		}
		_, err = tx.Exec(
			"UPDATE qingka_wangke_user SET addprice = ? WHERE grade_id = ?",
			req.Rate, req.ID,
		)
		if err != nil {
			return err
		}
		return tx.Commit()
	}
	_, err = database.DB.Exec(
		"INSERT INTO qingka_wangke_dengji (sort, name, rate, money, addkf, gjkf, status, time) VALUES (?, ?, ?, ?, ?, ?, ?, UNIX_TIMESTAMP())",
		req.Sort, req.Name, req.Rate, req.Money, req.AddKF, req.GJKF, req.Status,
	)
	return err
}

func (s *classService) GradeDelete(id int) error {
	assigned, err := s.CountGradeAssignments(id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if assigned > 0 {
		return errors.New("该等级仍有用户使用，不能删除")
	}
	_, err = database.DB.Exec("DELETE FROM qingka_wangke_dengji WHERE id=?", id)
	return err
}
