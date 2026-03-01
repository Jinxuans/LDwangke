package service

import (
	"fmt"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

// ExtMenuService 扩展菜单服务
type ExtMenuService struct{}

func NewExtMenuService() *ExtMenuService {
	return &ExtMenuService{}
}

// List 获取所有扩展菜单
func (s *ExtMenuService) List() ([]model.ExtMenu, error) {
	rows, err := database.DB.Query(
		"SELECT id, title, icon, url, sort_order, visible, scope, COALESCE(created_at,'') FROM qingka_ext_menu ORDER BY sort_order, id")
	if err != nil {
		return []model.ExtMenu{}, nil
	}
	defer rows.Close()
	var list []model.ExtMenu
	for rows.Next() {
		var m model.ExtMenu
		if err := rows.Scan(&m.ID, &m.Title, &m.Icon, &m.URL, &m.SortOrder, &m.Visible, &m.Scope, &m.CreatedAt); err != nil {
			continue
		}
		list = append(list, m)
	}
	if list == nil {
		list = []model.ExtMenu{}
	}
	return list, nil
}

// ListByScope 按 scope 获取可见的扩展菜单
func (s *ExtMenuService) ListByScope(scope string) ([]model.ExtMenu, error) {
	rows, err := database.DB.Query(
		"SELECT id, title, icon, url, sort_order, visible, scope, COALESCE(created_at,'') FROM qingka_ext_menu WHERE scope=? AND visible=1 ORDER BY sort_order, id", scope)
	if err != nil {
		return []model.ExtMenu{}, nil
	}
	defer rows.Close()
	var list []model.ExtMenu
	for rows.Next() {
		var m model.ExtMenu
		if err := rows.Scan(&m.ID, &m.Title, &m.Icon, &m.URL, &m.SortOrder, &m.Visible, &m.Scope, &m.CreatedAt); err != nil {
			continue
		}
		list = append(list, m)
	}
	if list == nil {
		list = []model.ExtMenu{}
	}
	return list, nil
}

// Save 保存扩展菜单（新增或更新）
func (s *ExtMenuService) Save(req model.ExtMenuSaveRequest) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	if req.Scope == "" {
		req.Scope = "backend"
	}
	if req.ID > 0 {
		_, err := database.DB.Exec(
			"UPDATE qingka_ext_menu SET title=?, icon=?, url=?, sort_order=?, visible=?, scope=? WHERE id=?",
			req.Title, req.Icon, req.URL, req.SortOrder, req.Visible, req.Scope, req.ID)
		return err
	}
	_, err := database.DB.Exec(
		"INSERT INTO qingka_ext_menu (title, icon, url, sort_order, visible, scope, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		req.Title, req.Icon, req.URL, req.SortOrder, req.Visible, req.Scope, now)
	return err
}

// Delete 删除扩展菜单
func (s *ExtMenuService) Delete(id int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_ext_menu WHERE id=?", id)
	if err != nil {
		return fmt.Errorf("删除失败: %v", err)
	}
	return nil
}
