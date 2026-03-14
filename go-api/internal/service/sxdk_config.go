package service

import (
	"encoding/json"
	"math"

	"go-api/internal/database"
)

var sxdkPlatformPrices = map[string]float64{
	"zxjy": 0.5, "qzt": 0.6, "xyb": 0.8, "gxy": 0.6,
	"xxy": 0.6, "xxt": 0.6, "hzj": 0.6,
}

func sxdkGetPlatformPrice(platform string, addprice float64) float64 {
	base := 10.0
	if p, ok := sxdkPlatformPrices[platform]; ok {
		base = p
	}
	return math.Round(addprice*base*100) / 100
}

func (s *SXDKService) GetConfig() (*SXDKConfig, error) {
	var val string
	err := database.DB.QueryRow("SELECT svalue FROM qingka_wangke_config WHERE skey = 'sxdk_config' LIMIT 1").Scan(&val)
	if err != nil {
		return &SXDKConfig{}, nil
	}
	var cfg SXDKConfig
	json.Unmarshal([]byte(val), &cfg)
	return &cfg, nil
}

func (s *SXDKService) SaveConfig(cfg *SXDKConfig) error {
	data, _ := json.Marshal(cfg)
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_config WHERE skey = 'sxdk_config'").Scan(&count)
	if count > 0 {
		_, err := database.DB.Exec("UPDATE qingka_wangke_config SET svalue = ? WHERE skey = 'sxdk_config'", string(data))
		return err
	}
	_, err := database.DB.Exec("INSERT INTO qingka_wangke_config (v, k, skey, svalue) VALUES ('sxdk_config', '', 'sxdk_config', ?)", string(data))
	return err
}
