package service

import (
	"encoding/json"
	"log"

	"go-api/internal/database"
)

func (s *TutuQGService) EnsureTable() {
	_, err := database.DB.Exec(`CREATE TABLE IF NOT EXISTS tutuqg (
		oid int(11) NOT NULL AUTO_INCREMENT,
		uid int(11) NOT NULL,
		user varchar(255) NOT NULL,
		pass varchar(255) NOT NULL,
		kcname varchar(255) NOT NULL,
		days varchar(255) NOT NULL,
		ptname varchar(255) NOT NULL,
		fees varchar(255) NOT NULL,
		addtime varchar(255) NOT NULL,
		IP varchar(255) DEFAULT NULL,
		status varchar(255) DEFAULT NULL,
		remarks varchar(255) DEFAULT NULL,
		guid varchar(255) DEFAULT NULL,
		score varchar(255) NOT NULL,
		scores varchar(255) DEFAULT NULL,
		zdxf varchar(255) DEFAULT NULL,
		PRIMARY KEY (oid)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		log.Printf("[TutuQG] 创建表失败: %v", err)
	}
}

func (s *TutuQGService) GetConfig() (*TutuQGConfig, error) {
	var val string
	err := database.DB.QueryRow("SELECT svalue FROM qingka_wangke_config WHERE skey = 'tutuqg_config' LIMIT 1").Scan(&val)
	if err != nil {
		return &TutuQGConfig{}, nil
	}
	var cfg TutuQGConfig
	json.Unmarshal([]byte(val), &cfg)
	return &cfg, nil
}

func (s *TutuQGService) SaveConfig(cfg *TutuQGConfig) error {
	data, _ := json.Marshal(cfg)
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_config WHERE skey = 'tutuqg_config'").Scan(&count)
	if count > 0 {
		_, err := database.DB.Exec("UPDATE qingka_wangke_config SET svalue = ? WHERE skey = 'tutuqg_config'", string(data))
		return err
	}
	_, err := database.DB.Exec("INSERT INTO qingka_wangke_config (v, k, skey, svalue) VALUES ('tutuqg_config', '', 'tutuqg_config', ?)", string(data))
	return err
}
