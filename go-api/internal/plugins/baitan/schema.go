package baitan

import (
	"go-api/internal/database"
	obslogger "go-api/internal/observability/logger"
)

func (s *Service) EnsureTable() {
	ddls := []string{
		`CREATE TABLE IF NOT EXISTS qingka_baitan (
			id INT NOT NULL AUTO_INCREMENT,
			uid INT NOT NULL DEFAULT 0 COMMENT '本站用户ID',
			type VARCHAR(10) NOT NULL DEFAULT '' COMMENT '平台',
			platform VARCHAR(10) NOT NULL DEFAULT '' COMMENT '兼容平台字段',
			userName VARCHAR(255) NOT NULL DEFAULT '' COMMENT '账号',
			passWord VARCHAR(255) NOT NULL DEFAULT '' COMMENT '密码',
			nikeName VARCHAR(255) NOT NULL DEFAULT '' COMMENT '姓名',
			sid VARCHAR(255) NOT NULL DEFAULT '' COMMENT '学校编码',
			sxdkId VARCHAR(80) NOT NULL DEFAULT '' COMMENT '上游订单ID',
			endDate DATETIME DEFAULT NULL COMMENT '到期时间',
			status VARCHAR(30) NOT NULL DEFAULT 'active' COMMENT '状态',
			code INT NOT NULL DEFAULT 1 COMMENT '打卡状态开关',
			week VARCHAR(255) NOT NULL DEFAULT '' COMMENT '打卡周期',
			report VARCHAR(255) NOT NULL DEFAULT '' COMMENT '报告类型',
			address VARCHAR(500) NOT NULL DEFAULT '' COMMENT '打卡地址',
			lon VARCHAR(50) NOT NULL DEFAULT '' COMMENT '经度',
			lat VARCHAR(50) NOT NULL DEFAULT '' COMMENT '纬度',
			version VARCHAR(100) NOT NULL DEFAULT '' COMMENT '邀请码/版本',
			weekNum INT NOT NULL DEFAULT 6,
			monthNum INT NOT NULL DEFAULT 25,
			pre_deduct DECIMAL(10,2) NOT NULL DEFAULT 0.00,
			actual_cost DECIMAL(10,2) DEFAULT NULL,
			final_charge DECIMAL(10,2) DEFAULT NULL,
			difference DECIMAL(10,2) DEFAULT NULL,
			payment_status VARCHAR(30) NOT NULL DEFAULT 'paid',
			result_data JSON DEFAULT NULL,
			error_message TEXT DEFAULT NULL,
			source VARCHAR(30) NOT NULL DEFAULT 'local',
			agent_uid INT NOT NULL DEFAULT 0,
			createTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY idx_uid (uid),
			KEY idx_userName (userName),
			KEY idx_passWord (passWord),
			KEY idx_type (type),
			KEY idx_status (status),
			KEY idx_sxdkId (sxdkId)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='摆摊实习打卡订单表'`,
	}
	for _, ddl := range ddls {
		if _, err := database.DB.Exec(ddl); err != nil {
			obslogger.L().Warn("Baitan create table failed", "error", err)
		}
	}
	columns := []string{
		"ALTER TABLE qingka_baitan ADD COLUMN uid INT NOT NULL DEFAULT 0",
		"ALTER TABLE qingka_baitan ADD COLUMN platform VARCHAR(10) NOT NULL DEFAULT ''",
		"ALTER TABLE qingka_baitan ADD COLUMN sxdkId VARCHAR(80) NOT NULL DEFAULT ''",
		"ALTER TABLE qingka_baitan ADD COLUMN code INT NOT NULL DEFAULT 1",
		"ALTER TABLE qingka_baitan ADD COLUMN address VARCHAR(500) NOT NULL DEFAULT ''",
		"ALTER TABLE qingka_baitan ADD COLUMN lon VARCHAR(50) NOT NULL DEFAULT ''",
		"ALTER TABLE qingka_baitan ADD COLUMN lat VARCHAR(50) NOT NULL DEFAULT ''",
		"ALTER TABLE qingka_baitan ADD COLUMN version VARCHAR(100) NOT NULL DEFAULT ''",
		"ALTER TABLE qingka_baitan ADD COLUMN weekNum INT NOT NULL DEFAULT 6",
		"ALTER TABLE qingka_baitan ADD COLUMN monthNum INT NOT NULL DEFAULT 25",
		"ALTER TABLE qingka_baitan ADD COLUMN pre_deduct DECIMAL(10,2) NOT NULL DEFAULT 0.00",
		"ALTER TABLE qingka_baitan ADD COLUMN actual_cost DECIMAL(10,2) DEFAULT NULL",
		"ALTER TABLE qingka_baitan ADD COLUMN final_charge DECIMAL(10,2) DEFAULT NULL",
		"ALTER TABLE qingka_baitan ADD COLUMN difference DECIMAL(10,2) DEFAULT NULL",
		"ALTER TABLE qingka_baitan ADD COLUMN payment_status VARCHAR(30) NOT NULL DEFAULT 'paid'",
		"ALTER TABLE qingka_baitan ADD COLUMN result_data JSON DEFAULT NULL",
		"ALTER TABLE qingka_baitan ADD COLUMN error_message TEXT DEFAULT NULL",
		"ALTER TABLE qingka_baitan ADD COLUMN source VARCHAR(30) NOT NULL DEFAULT 'local'",
		"ALTER TABLE qingka_baitan ADD COLUMN agent_uid INT NOT NULL DEFAULT 0",
		"ALTER TABLE qingka_baitan ADD COLUMN updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP",
	}
	for _, ddl := range columns {
		_, _ = database.DB.Exec(ddl)
	}
	cfg := defaultConfig()
	raw, _ := cfg.Marshal()
	var count int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_config WHERE v = ?", "baitan_config").Scan(&count); err == nil && count == 0 {
		if _, err := database.DB.Exec("INSERT INTO qingka_wangke_config (v, k) VALUES (?, ?)", "baitan_config", raw); err != nil {
			obslogger.L().Warn("Baitan seed config failed", "error", err)
		}
	}
}
