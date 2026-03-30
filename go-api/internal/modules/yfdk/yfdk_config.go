package yfdk

import (
	"encoding/json"

	"go-api/internal/database"
	obslogger "go-api/internal/observability/logger"
)

func (s *YFDKService) EnsureTable() {
	obslogger.L().Info("YFDK 开始检查/创建项目表")
	_, err := database.DB.Exec(`CREATE TABLE IF NOT EXISTS qingka_wangke_yfdk_projects (
		id INT(11) NOT NULL AUTO_INCREMENT,
		cid VARCHAR(10) NOT NULL COMMENT '上游项目CID',
		name VARCHAR(100) NOT NULL COMMENT '项目名称',
		content VARCHAR(255) DEFAULT '' COMMENT '说明',
		cost_price DECIMAL(10,2) DEFAULT 0 COMMENT '成本价（上游）',
		sell_price DECIMAL(10,2) DEFAULT 0.10 COMMENT '售价',
		enabled TINYINT(1) DEFAULT 1 COMMENT '是否启用 1启用 0禁用',
		sort INT(11) DEFAULT 10 COMMENT '排序',
		create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
		update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		UNIQUE KEY uk_cid (cid)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='YF打卡项目表'`)
	if err != nil {
		obslogger.L().Warn("YFDK 创建表失败", "error", err)
	} else {
		obslogger.L().Info("YFDK 项目表检查/创建完成")
	}
}

var yfdkPlatformPrices = map[string]float64{
	"1": 0.0, "14": 0.10, "15": 0.10, "16": 0.10, "17": 0.10,
	"18": 0.10, "20": 0.10, "24": 0.30, "30": 0.10, "36": 0.10,
	"37": 0.10, "39": 0.10, "40": 0.10, "41": 0.10, "43": 0.10,
	"44": 0.10, "45": 0.10, "46": 0.10, "48": 0.10, "49": 0.10,
	"50": 0.10, "51": 0.10, "52": 0.10, "53": 0.10,
}

func yfdkGetPlatformPrice(cid string) float64 {
	if p, ok := yfdkPlatformPrices[cid]; ok {
		return p
	}
	return 0.10
}

func (s *YFDKService) GetConfig() (*YFDKConfig, error) {
	var val string
	err := database.DB.QueryRow("SELECT svalue FROM qingka_wangke_config WHERE skey = 'yfdk_config' LIMIT 1").Scan(&val)
	if err != nil {
		return &YFDKConfig{}, nil
	}
	var cfg YFDKConfig
	json.Unmarshal([]byte(val), &cfg)
	return &cfg, nil
}

func (s *YFDKService) SaveConfig(cfg *YFDKConfig) error {
	data, _ := json.Marshal(cfg)
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_config WHERE skey = 'yfdk_config'").Scan(&count)
	if count > 0 {
		_, err := database.DB.Exec("UPDATE qingka_wangke_config SET svalue = ? WHERE skey = 'yfdk_config'", string(data))
		return err
	}
	_, err := database.DB.Exec("INSERT INTO qingka_wangke_config (v, k, skey, svalue) VALUES ('yfdk_config', '', 'yfdk_config', ?)", string(data))
	return err
}
