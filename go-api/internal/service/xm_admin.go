package service

import (
	"fmt"

	"go-api/internal/database"
)

// XMProjectAdmin 管理员视角的项目信息（含对接配置）
type XMProjectAdmin struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Query       int     `json:"query"`
	Password    int     `json:"password"`
	URL         string  `json:"url"`
	UID         string  `json:"uid"`
	Key         string  `json:"key"`
	Token       string  `json:"token"`
	Type        int     `json:"type"`
	PID         string  `json:"p_id"`
	Status      int     `json:"status"`
}

func (s *XMService) AdminListProjects() ([]XMProjectAdmin, error) {
	rows, err := database.DB.Query("SELECT id, name, COALESCE(description,''), price, `query`, password, url, uid, `key`, token, type, p_id, status FROM xm_project WHERE is_deleted = 0 ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []XMProjectAdmin
	for rows.Next() {
		var p XMProjectAdmin
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Query, &p.Password, &p.URL, &p.UID, &p.Key, &p.Token, &p.Type, &p.PID, &p.Status); err != nil {
			continue
		}
		list = append(list, p)
	}
	if list == nil {
		list = []XMProjectAdmin{}
	}
	return list, nil
}

func (s *XMService) AdminSaveProject(p XMProjectAdmin) (int, error) {
	if p.Name == "" {
		return 0, fmt.Errorf("项目名称不能为空")
	}
	if p.ID > 0 {
		_, err := database.DB.Exec(
			"UPDATE xm_project SET name=?, description=?, price=?, `query`=?, password=?, url=?, uid=?, `key`=?, token=?, type=?, p_id=?, status=? WHERE id=?",
			p.Name, p.Description, p.Price, p.Query, p.Password, p.URL, p.UID, p.Key, p.Token, p.Type, p.PID, p.Status, p.ID,
		)
		if err != nil {
			return 0, fmt.Errorf("保存失败: %v", err)
		}
		return p.ID, nil
	}
	result, err := database.DB.Exec(
		"INSERT INTO xm_project (name, description, price, `query`, password, url, uid, `key`, token, type, p_id, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		p.Name, p.Description, p.Price, p.Query, p.Password, p.URL, p.UID, p.Key, p.Token, p.Type, p.PID, p.Status,
	)
	if err != nil {
		return 0, fmt.Errorf("添加失败: %v", err)
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

func (s *XMService) AdminDeleteProject(id int) error {
	_, err := database.DB.Exec("UPDATE xm_project SET is_deleted = 1 WHERE id = ?", id)
	return err
}
