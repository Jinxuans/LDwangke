package service

import (
	"encoding/json"
	"log"

	"go-api/internal/database"
)

func (s *TuboshuService) EnsureTable() {
	_, err := database.DB.Exec(`CREATE TABLE IF NOT EXISTS qingka_wangke_dialogue (
		id int(11) NOT NULL AUTO_INCREMENT,
		uid int(11) NOT NULL,
		title varchar(255) NOT NULL DEFAULT '',
		state varchar(255) NOT NULL DEFAULT 'PENDING',
		download_url varchar(255) NOT NULL DEFAULT '',
		addtime varchar(255) NOT NULL DEFAULT '',
		ip varchar(255) NOT NULL DEFAULT '',
		source_id bigint(17) NOT NULL DEFAULT 0,
		dialogue_id varchar(32) NOT NULL DEFAULT '0',
		point decimal(11,2) NOT NULL DEFAULT 0.00,
		type varchar(32) NOT NULL DEFAULT '',
		PRIMARY KEY (id),
		KEY idx_uid (uid),
		KEY idx_state (state)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		log.Printf("[Tuboshu] 创建 dialogue 表失败: %v", err)
	}

	_, err = database.DB.Exec(`CREATE TABLE IF NOT EXISTS points_product (
		id INT(11) NOT NULL AUTO_INCREMENT,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		image_url VARCHAR(500),
		price DECIMAL(10,2) NOT NULL,
		status ENUM('ENABLED','DISABLED') NOT NULL DEFAULT 'ENABLED',
		sort_order INT(11) NOT NULL DEFAULT 0,
		create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		log.Printf("[Tuboshu] 创建 points_product 表失败: %v", err)
	}

	_, err = database.DB.Exec(`CREATE TABLE IF NOT EXISTS points_product_code (
		id INT(11) NOT NULL AUTO_INCREMENT,
		product_id INT(11) NOT NULL,
		code VARCHAR(500) NOT NULL,
		status ENUM('AVAILABLE','EXCHANGED') NOT NULL DEFAULT 'AVAILABLE',
		exchanged_by INT(11),
		exchanged_at DATETIME,
		create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		KEY idx_product_status (product_id, status)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		log.Printf("[Tuboshu] 创建 points_product_code 表失败: %v", err)
	}

	_, err = database.DB.Exec(`CREATE TABLE IF NOT EXISTS points_exchange_record (
		id INT(11) NOT NULL AUTO_INCREMENT,
		uid INT(11) NOT NULL,
		product_id INT(11) NOT NULL,
		product_name VARCHAR(255) NOT NULL,
		code_id INT(11) NOT NULL,
		code VARCHAR(500) NOT NULL,
		points_cost DECIMAL(10,2) NOT NULL,
		create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		KEY idx_uid (uid)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		log.Printf("[Tuboshu] 创建 points_exchange_record 表失败: %v", err)
	}

	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_config WHERE skey = 'tuboshu_config'").Scan(&count)
	if count == 0 {
		defaultCfg := TuboshuConfig{
			PriceRatio:  5,
			PriceConfig: defaultPriceConfig(),
			PageVisibility: map[string]bool{
				"ComponentStagePage": true,
				"ChatPage":           true,
				"ChartPage":          true,
				"TemplatePage":       true,
				"ReductionPage":      true,
				"AccountTable":       true,
				"TicketPage":         true,
			},
		}
		data, _ := json.Marshal(defaultCfg)
		database.DB.Exec("INSERT INTO qingka_wangke_config (v, k, skey, svalue) VALUES ('tuboshu_config', '', 'tuboshu_config', ?)", string(data))
	}
}

func defaultPriceConfig() map[string]interface{} {
	return map[string]interface{}{
		"SIMPLE_DIALOGUE": map[string]interface{}{
			"type": "id_based", "enabled": true,
			"prices": map[string]interface{}{
				"1": 44.8, "2": 12, "3": 12, "4": 12, "5": 18,
				"6": 0.6, "7": 0.6, "8": 0.6, "9": 24, "10": 12,
				"11": 12, "12": 12, "13": 3, "14": 12, "15": 12,
				"16": 12, "17": 12, "19": 0.6, "21": 0.6, "22": 0.6,
				"23": 0.6, "24": 0.6, "25": 0.6, "26": 0.6, "31": 0.6,
				"36": 4, "37": 4, "40": 0.2, "41": 0.2, "42": 0.2,
				"43": 4, "44": 4, "45": 0.1,
			},
		},
		"STAGE_DIALOGUE": map[string]interface{}{
			"type": "id_based", "enabled": true,
			"prices": map[string]interface{}{
				"1": 44.8, "2": 12, "3": 12, "4": 12, "5": 18,
				"9": 24, "10": 12, "11": 12, "12": 12, "13": 3,
				"14": 12, "15": 12, "16": 12, "17": 12,
			},
		},
		"PAPER_WRITING": map[string]interface{}{
			"type": "paper_writing",
			"config": map[string]interface{}{
				"enabled": true, "sectionBasePrice": 1.6,
				"pointBasePrice": 1.0, "v3ModelExtraPrice": 0.0,
				"reductionExtraPrice": 10.0, "enableV3Model": false,
			},
		},
		"PAPER_REDUCTION": map[string]interface{}{
			"type": "id_based", "enabled": true,
			"prices": map[string]interface{}{
				"NORMAL": 2, "ENGLISH": 4, "WEIPU": 3, "REPEAT": 5,
			},
		},
		"PAPER_OUTLINE": map[string]interface{}{
			"type": "fixed", "enabled": true, "price": 0.2,
		},
		"CHART_GENERATE": map[string]interface{}{
			"type": "fixed", "enabled": true, "price": 2.0,
		},
		"PPT_GENERATE": map[string]interface{}{
			"type": "fixed", "enabled": true, "price": 2.0,
		},
		"WORD_TEMPLATE": map[string]interface{}{
			"type": "fixed", "enabled": true, "price": 2.0,
		},
	}
}

func (s *TuboshuService) GetConfig() (*TuboshuConfig, error) {
	var val string
	err := database.DB.QueryRow("SELECT svalue FROM qingka_wangke_config WHERE skey = 'tuboshu_config' LIMIT 1").Scan(&val)
	if err != nil {
		cfg := &TuboshuConfig{PriceRatio: 5, PriceConfig: defaultPriceConfig()}
		return cfg, nil
	}
	var cfg TuboshuConfig
	json.Unmarshal([]byte(val), &cfg)
	if cfg.PriceRatio == 0 {
		cfg.PriceRatio = 5
	}
	if cfg.PriceConfig == nil {
		cfg.PriceConfig = defaultPriceConfig()
	}
	return &cfg, nil
}

func (s *TuboshuService) SaveConfig(cfg *TuboshuConfig) error {
	data, _ := json.Marshal(cfg)
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_config WHERE skey = 'tuboshu_config'").Scan(&count)
	if count > 0 {
		_, err := database.DB.Exec("UPDATE qingka_wangke_config SET svalue = ? WHERE skey = 'tuboshu_config'", string(data))
		return err
	}
	_, err := database.DB.Exec("INSERT INTO qingka_wangke_config (v, k, skey, svalue) VALUES ('tuboshu_config', '', 'tuboshu_config', ?)", string(data))
	return err
}

func (s *TuboshuService) SavePriceConfig(priceConfig map[string]interface{}) error {
	cfg, err := s.GetConfig()
	if err != nil {
		return err
	}

	for k, v := range priceConfig {
		cfg.PriceConfig[k] = v
	}

	if cfg.PageVisibility == nil {
		cfg.PageVisibility = map[string]bool{
			"ComponentStagePage": true, "ChatPage": true, "ChartPage": true,
			"TemplatePage": true, "ReductionPage": true, "AccountTable": true, "TicketPage": true,
		}
	}

	return s.SaveConfig(cfg)
}
