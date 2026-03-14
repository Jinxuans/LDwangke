package service

import (
	"log"
	"strconv"
	"strings"

	"go-api/internal/database"
)

func (s *PaperService) EnsureTable() {
	_, err := database.DB.Exec(`CREATE TABLE IF NOT EXISTS qingka_wangke_lunwen (
		id int(11) UNSIGNED NOT NULL AUTO_INCREMENT,
		uid int(11) UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
		order_id varchar(255) DEFAULT NULL COMMENT '上游订单ID',
		shopcode varchar(100) DEFAULT NULL COMMENT '商品代码',
		title varchar(255) DEFAULT NULL COMMENT '论文标题',
		price decimal(10,2) UNSIGNED DEFAULT NULL COMMENT '扣费价格',
		PRIMARY KEY (id),
		KEY idx_uid (uid),
		KEY idx_order_id (order_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='智文论文订单表'`)
	if err != nil {
		log.Printf("[Paper] 创建 lunwen 表失败: %v", err)
	}

	defaults := map[string]string{
		"lunwen_api_username":       "",
		"lunwen_api_password":       "",
		"lunwen_api_6000_price":     "30",
		"lunwen_api_8000_price":     "40",
		"lunwen_api_10000_price":    "50",
		"lunwen_api_12000_price":    "60",
		"lunwen_api_15000_price":    "75",
		"lunwen_api_rws_price":      "10",
		"lunwen_api_ktbg_price":     "10",
		"lunwen_api_jdaigchj_price": "10",
		"lunwen_api_xgdl_price":     "3",
		"lunwen_api_jcl_price":      "3",
		"lunwen_api_jdaigcl_price":  "3",
	}
	for k, v := range defaults {
		var count int
		database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_config WHERE v = ?", k).Scan(&count)
		if count == 0 {
			database.DB.Exec("INSERT INTO qingka_wangke_config (v, k) VALUES (?, ?)", k, v)
		}
	}
}

func (s *PaperService) GetConfig() (map[string]string, error) {
	rows, err := database.DB.Query("SELECT v, k FROM qingka_wangke_config WHERE v LIKE 'lunwen_api_%'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conf := make(map[string]string)
	for rows.Next() {
		var key, val string
		rows.Scan(&key, &val)
		conf[key] = val
	}
	return conf, nil
}

func (s *PaperService) SaveConfig(data map[string]string) error {
	for k, v := range data {
		if !strings.HasPrefix(k, "lunwen_api_") {
			continue
		}
		var count int
		database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_config WHERE v = ?", k).Scan(&count)
		if count > 0 {
			_, err := database.DB.Exec("UPDATE qingka_wangke_config SET k = ? WHERE v = ?", v, k)
			if err != nil {
				return err
			}
		} else {
			_, err := database.DB.Exec("INSERT INTO qingka_wangke_config (v, k) VALUES (?, ?)", k, v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *PaperService) GetConfigPrice(key string) float64 {
	var val string
	err := database.DB.QueryRow("SELECT k FROM qingka_wangke_config WHERE v = ? LIMIT 1", key).Scan(&val)
	if err != nil {
		return 0
	}
	f, _ := strconv.ParseFloat(val, 64)
	return f
}

func (s *PaperService) GetUserAddPrice(uid int) float64 {
	var addprice float64
	database.DB.QueryRow("SELECT addprice FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&addprice)
	if addprice == 0 {
		addprice = 1
	}
	return addprice
}

func (s *PaperService) GetUserMoney(uid int) float64 {
	var money float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&money)
	return money
}
