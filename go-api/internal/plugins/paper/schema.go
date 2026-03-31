package paper

import (
	"go-api/internal/database"
	obslogger "go-api/internal/observability/logger"
)

func (s *paperService) EnsureTable() {
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
		obslogger.L().Warn("Paper 创建 lunwen 表失败", "error", err)
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
