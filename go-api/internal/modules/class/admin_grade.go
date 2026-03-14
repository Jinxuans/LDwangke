package class

import (
	"go-api/internal/database"
	"go-api/internal/model"
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

func (s *classService) GradeDelete(id int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_wangke_dengji WHERE id=?", id)
	return err
}
