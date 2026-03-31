package sdxy

import (
	"encoding/json"

	"go-api/internal/database"
)

func (s *SDXYService) GetConfig() (*SDXYConfig, error) {
	var val string
	err := database.DB.QueryRow("SELECT svalue FROM qingka_wangke_config WHERE skey = 'sdxy_config' LIMIT 1").Scan(&val)
	if err != nil {
		return &SDXYConfig{Price: 3.75, Endpoint: "/flash/api.php", Timeout: 30}, nil
	}
	var cfg SDXYConfig
	json.Unmarshal([]byte(val), &cfg)
	if cfg.Endpoint == "" {
		cfg.Endpoint = "/flash/api.php"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30
	}
	return &cfg, nil
}

func (s *SDXYService) SaveConfig(cfg *SDXYConfig) error {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "/flash/api.php"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30
	}
	data, _ := json.Marshal(cfg)
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_config (v, k, skey, svalue) VALUES ('sdxy_config', '', 'sdxy_config', ?) ON DUPLICATE KEY UPDATE svalue = ?",
		string(data), string(data),
	)
	return err
}
