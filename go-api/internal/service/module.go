package service

import (
	"fmt"

	"go-api/internal/database"
	"go-api/internal/model"
)

type ModuleService struct{}

func NewModuleService() *ModuleService {
	return &ModuleService{}
}

const moduleSelectCols = "id, app_id, COALESCE(type,'sport'), name, COALESCE(description,''), COALESCE(price,''), COALESCE(icon,''), COALESCE(api_base,''), COALESCE(view_url,''), status, sort, COALESCE(config,'{}')"

func scanModule(scanner interface{ Scan(...interface{}) error }) (model.DynamicModule, error) {
	var m model.DynamicModule
	err := scanner.Scan(&m.ID, &m.AppID, &m.Type, &m.Name, &m.Description, &m.Price, &m.Icon, &m.ApiBase, &m.ViewURL, &m.Status, &m.Sort, &m.Config)
	return m, err
}

// listModules 统一查询方法，where/args 可选
func (s *ModuleService) listModules(where string, args ...interface{}) ([]model.DynamicModule, error) {
	q := "SELECT " + moduleSelectCols + " FROM qingka_dynamic_module"
	if where != "" {
		q += " WHERE " + where
	}
	q += " ORDER BY sort ASC"

	rows, err := database.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modules []model.DynamicModule
	for rows.Next() {
		m, err := scanModule(rows)
		if err != nil {
			continue
		}
		modules = append(modules, m)
	}
	if modules == nil {
		modules = []model.DynamicModule{}
	}
	return modules, nil
}

func (s *ModuleService) ListActive() ([]model.DynamicModule, error) {
	return s.listModules("status = 1")
}

func (s *ModuleService) ListActiveByType(moduleType string) ([]model.DynamicModule, error) {
	return s.listModules("status = 1 AND type = ?", moduleType)
}

func (s *ModuleService) ListAll() ([]model.DynamicModule, error) {
	return s.listModules("")
}

func (s *ModuleService) GetByAppID(appID string) (*model.DynamicModule, error) {
	row := database.DB.QueryRow(
		"SELECT "+moduleSelectCols+" FROM qingka_dynamic_module WHERE app_id = ? AND status = 1",
		appID,
	)
	m, err := scanModule(row)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *ModuleService) Save(req model.ModuleSaveRequest) error {
	if req.AppID == "" || req.Name == "" {
		return fmt.Errorf("app_id 和 name 不能为空")
	}
	if req.Type == "" {
		req.Type = "sport"
	}
	if req.ID > 0 {
		_, err := database.DB.Exec(
			"UPDATE qingka_dynamic_module SET app_id=?, type=?, name=?, description=?, price=?, icon=?, api_base=?, view_url=?, status=?, sort=?, config=? WHERE id=?",
			req.AppID, req.Type, req.Name, req.Description, req.Price, req.Icon, req.ApiBase, req.ViewURL, req.Status, req.Sort, req.Config, req.ID,
		)
		return err
	}
	_, err := database.DB.Exec(
		"INSERT INTO qingka_dynamic_module (app_id, type, name, description, price, icon, api_base, view_url, status, sort, config) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		req.AppID, req.Type, req.Name, req.Description, req.Price, req.Icon, req.ApiBase, req.ViewURL, req.Status, req.Sort, req.Config,
	)
	return err
}

func (s *ModuleService) Delete(id int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_dynamic_module WHERE id = ?", id)
	return err
}
