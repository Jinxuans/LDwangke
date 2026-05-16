package shashou

import (
	"go-api/internal/database"
	obslogger "go-api/internal/observability/logger"
)

func (s *Service) EnsureTable() {
	ddls := []string{
		`CREATE TABLE IF NOT EXISTS ss_project (
			id INT NOT NULL AUTO_INCREMENT,
			name VARCHAR(100) NOT NULL DEFAULT '鲨兽运动世界' COMMENT '项目名称',
			type TINYINT NOT NULL DEFAULT 0 COMMENT '0=源台 1=29二开',
			remote_project_id INT NOT NULL DEFAULT 0 COMMENT '上游项目ID，0=使用本地ID',
			api_url VARCHAR(255) NOT NULL DEFAULT '',
			api_key VARCHAR(255) NOT NULL DEFAULT '',
			user_id VARCHAR(50) NOT NULL DEFAULT '',
			price_normal DECIMAL(10,2) NOT NULL DEFAULT 9.00,
			price_morning DECIMAL(10,2) NOT NULL DEFAULT 10.00,
			actual_rate DECIMAL(10,2) NOT NULL DEFAULT 1.05,
			rush_fee DECIMAL(10,2) NOT NULL DEFAULT 3.00,
			query_fee DECIMAL(10,2) NOT NULL DEFAULT 1.00,
			min_balance DECIMAL(10,2) NOT NULL DEFAULT 0.00,
			status TINYINT NOT NULL DEFAULT 1,
			auto_sync TINYINT NOT NULL DEFAULT 1,
			sync_interval INT NOT NULL DEFAULT 300,
			timeout INT NOT NULL DEFAULT 30,
			remark VARCHAR(255) NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY idx_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='鲨兽运动项目配置表'`,
		`CREATE TABLE IF NOT EXISTS ss_order (
			id INT NOT NULL AUTO_INCREMENT,
			order_no VARCHAR(100) DEFAULT NULL COMMENT '上游订单号',
			user_id INT NOT NULL COMMENT '本站用户ID',
			project_id INT NOT NULL COMMENT '项目ID',
			order_type INT NOT NULL COMMENT '1课外跑 2晨跑 3查课外 4退款 5查晨跑',
			is_rush_order TINYINT(1) DEFAULT 0,
			total_distance DECIMAL(10,2) DEFAULT 0.00,
			account_count INT DEFAULT 0,
			pre_deduct DECIMAL(10,2) DEFAULT 0.00,
			actual_cost DECIMAL(10,2) DEFAULT NULL,
			final_charge DECIMAL(10,2) DEFAULT NULL,
			difference DECIMAL(10,2) DEFAULT NULL,
			rush_order_fee DECIMAL(10,2) DEFAULT 0.00,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			payment_status VARCHAR(50) DEFAULT 'pre_deducted',
			accounts JSON DEFAULT NULL,
			query_account VARCHAR(50) DEFAULT NULL,
			refund_account VARCHAR(50) DEFAULT NULL,
			result_data JSON DEFAULT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME DEFAULT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			retry_count INT DEFAULT 0,
			last_retry_time DATETIME DEFAULT NULL,
			error_message TEXT DEFAULT NULL,
			refund_km DECIMAL(10,2) DEFAULT NULL,
			source VARCHAR(30) NOT NULL DEFAULT 'local',
			agent_uid INT NOT NULL DEFAULT 0,
			PRIMARY KEY (id),
			KEY idx_order_no (order_no),
			KEY idx_user_status (user_id,status),
			KEY idx_project (project_id),
			KEY idx_created (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='鲨兽运动订单表'`,
		`CREATE TABLE IF NOT EXISTS ss_accounts (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			order_id BIGINT UNSIGNED NOT NULL,
			order_no VARCHAR(50) NOT NULL DEFAULT '',
			user_id BIGINT UNSIGNED NOT NULL,
			project_id INT NOT NULL,
			account VARCHAR(50) NOT NULL,
			password VARCHAR(100) NOT NULL DEFAULT '',
			distance DECIMAL(8,2) NOT NULL DEFAULT 0.00,
			start_hour TINYINT UNSIGNED NOT NULL DEFAULT 0,
			start_minute TINYINT UNSIGNED NOT NULL DEFAULT 0,
			end_hour TINYINT UNSIGNED NOT NULL DEFAULT 0,
			end_minute TINYINT UNSIGNED NOT NULL DEFAULT 0,
			run_days VARCHAR(20) NOT NULL DEFAULT '',
			order_type TINYINT UNSIGNED NOT NULL,
			is_rush_order TINYINT(1) DEFAULT 0,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			error_message TEXT DEFAULT NULL,
			processed_at DATETIME DEFAULT NULL,
			query_result JSON DEFAULT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY idx_order_id (order_id),
			KEY idx_order_no (order_no),
			KEY idx_user_account (user_id,account),
			KEY idx_status (status),
			KEY idx_created (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='鲨兽运动账号订单明细表'`,
	}
	for _, ddl := range ddls {
		if _, err := database.DB.Exec(ddl); err != nil {
			obslogger.L().Warn("Shashou create table failed", "error", err)
		}
	}
	columns := []string{
		"ALTER TABLE ss_project ADD COLUMN name VARCHAR(100) NOT NULL DEFAULT '鲨兽运动世界' COMMENT '项目名称' AFTER id",
		"ALTER TABLE ss_project ADD COLUMN remote_project_id INT NOT NULL DEFAULT 0 COMMENT '上游项目ID，0=使用本地ID' AFTER type",
		"ALTER TABLE ss_project ADD COLUMN auto_sync TINYINT NOT NULL DEFAULT 1",
		"ALTER TABLE ss_project ADD COLUMN sync_interval INT NOT NULL DEFAULT 300",
		"ALTER TABLE ss_project ADD COLUMN timeout INT NOT NULL DEFAULT 30",
		"ALTER TABLE ss_project ADD COLUMN remark VARCHAR(255) NOT NULL DEFAULT ''",
		"ALTER TABLE ss_order ADD COLUMN source VARCHAR(30) NOT NULL DEFAULT 'local'",
		"ALTER TABLE ss_order ADD COLUMN agent_uid INT NOT NULL DEFAULT 0",
	}
	for _, ddl := range columns {
		_, _ = database.DB.Exec(ddl)
	}
	var count int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM ss_project").Scan(&count); err == nil && count == 0 {
		_, err = database.DB.Exec(`INSERT INTO ss_project
			(name,type,remote_project_id,api_url,api_key,user_id,price_normal,price_morning,actual_rate,rush_fee,query_fee,min_balance,status,auto_sync,sync_interval,timeout,remark)
			VALUES ('鲨兽运动世界',0,0,'https://ssyd.cc','keykeykey','8888',9.00,10.00,1.05,3.00,1.00,0.00,1,1,300,30,'默认源台配置，请上线前修改')`)
		if err != nil {
			obslogger.L().Warn("Shashou seed project failed", "error", err)
		}
	}
}
