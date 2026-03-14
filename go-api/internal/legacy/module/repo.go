package module

import (
	"go-api/internal/database"
	"go-api/internal/model"
)

const moduleSelectCols = "id, app_id, COALESCE(type,'sport'), name, COALESCE(description,''), COALESCE(price,''), COALESCE(icon,''), COALESCE(api_base,''), COALESCE(view_url,''), status, sort, COALESCE(config,'{}')"

func scanModule(scanner interface{ Scan(...interface{}) error }) (model.DynamicModule, error) {
	var m model.DynamicModule
	err := scanner.Scan(&m.ID, &m.AppID, &m.Type, &m.Name, &m.Description, &m.Price, &m.Icon, &m.ApiBase, &m.ViewURL, &m.Status, &m.Sort, &m.Config)
	return m, err
}

func getByAppID(appID string) (*model.DynamicModule, error) {
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
