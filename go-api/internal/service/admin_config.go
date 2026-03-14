package service

import (
	"encoding/json"
	"strings"

	"go-api/internal/database"
)

func (s *AdminService) GetConfig() (map[string]string, error) {
	rows, err := database.DB.Query("SELECT `v`, `k` FROM qingka_wangke_config")
	if err != nil {
		return map[string]string{}, nil
	}
	defer rows.Close()

	config := make(map[string]string)
	for rows.Next() {
		var k, v string
		rows.Scan(&k, &v)
		config[k] = v
	}
	return config, nil
}

func (s *AdminService) SaveConfig(configs map[string]string) error {
	if len(configs) == 0 {
		return nil
	}

	stmt := "REPLACE INTO qingka_wangke_config (`v`, `k`) VALUES "
	placeholders := make([]string, 0, len(configs))
	params := make([]interface{}, 0, len(configs)*2)

	for k, v := range configs {
		placeholders = append(placeholders, "(?, ?)")
		params = append(params, k, v)
	}
	stmt += strings.Join(placeholders, ", ")

	_, err := database.DB.Exec(stmt, params...)
	return err
}

func (s *AdminService) GetPayData() (map[string]string, error) {
	var paydata string
	err := database.DB.QueryRow("SELECT COALESCE(paydata,'') FROM qingka_wangke_user WHERE uid = 1").Scan(&paydata)
	if err != nil {
		return map[string]string{}, nil
	}
	result := make(map[string]string)
	if paydata != "" {
		json.Unmarshal([]byte(paydata), &result)
	}
	return result, nil
}

func (s *AdminService) SavePayData(data map[string]string) error {
	existing, _ := s.GetPayData()
	for k, v := range data {
		existing[k] = v
	}
	jsonBytes, err := json.Marshal(existing)
	if err != nil {
		return err
	}
	_, err = database.DB.Exec("UPDATE qingka_wangke_user SET paydata = ? WHERE uid = 1", string(jsonBytes))
	return err
}
