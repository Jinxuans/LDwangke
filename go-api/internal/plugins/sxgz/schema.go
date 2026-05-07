package sxgz

import (
	"go-api/internal/database"

	obslogger "go-api/internal/observability/logger"
)

func (s *SxgzService) EnsureTable() {
	_, err := database.DB.Exec(`CREATE TABLE IF NOT EXISTS fd_sxgz_orders (
		order_id int(11) NOT NULL AUTO_INCREMENT,
		uid int(11) NOT NULL DEFAULT 0,
		order_no varchar(64) NOT NULL,
		service_type enum('electronic','mail','both') NOT NULL DEFAULT 'electronic',
		company_id int(11) NOT NULL DEFAULT 0,
		company_name varchar(255) NOT NULL DEFAULT '',
		business_license tinyint(1) NOT NULL DEFAULT 0,
		only_business_license tinyint(1) NOT NULL DEFAULT 0,
		material_type enum('upload','mail') DEFAULT 'upload',
		uploaded_file text DEFAULT NULL,
		original_filename text DEFAULT NULL,
		file_size int(11) DEFAULT NULL,
		customer_name varchar(100) NOT NULL DEFAULT '',
		customer_email varchar(255) DEFAULT NULL,
		customer_phone varchar(20) DEFAULT NULL,
		customer_address text DEFAULT NULL,
		courier_company varchar(50) DEFAULT NULL,
		tracking_number varchar(100) DEFAULT NULL,
		return_tracking_number varchar(100) DEFAULT NULL,
		print_copies int(11) NOT NULL DEFAULT 0,
		print_options text DEFAULT NULL,
		paper_size enum('A4','A3') NOT NULL DEFAULT 'A4',
		special_requirements text DEFAULT NULL,
		base_price decimal(10,2) NOT NULL DEFAULT 0.00,
		mail_price decimal(10,2) NOT NULL DEFAULT 0.00,
		print_price decimal(10,2) NOT NULL DEFAULT 0.00,
		license_price decimal(10,2) NOT NULL DEFAULT 0.00,
		total_price decimal(10,2) NOT NULL DEFAULT 0.00,
		status enum('pending','processing','completed','delivered','cancelled','failed','refund_requested','refunded') NOT NULL DEFAULT 'pending',
		admin_notes text DEFAULT NULL,
		refund_reason text DEFAULT NULL,
		processed_files text DEFAULT NULL,
		processed_file_url text DEFAULT NULL,
		source enum('direct','agent') NOT NULL DEFAULT 'direct',
		agent_uid int(11) DEFAULT NULL,
		agent_order_id int(11) DEFAULT NULL,
		created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
		completed_at datetime DEFAULT NULL,
		PRIMARY KEY (order_id),
		UNIQUE KEY uq_order_no (order_no),
		KEY idx_uid (uid),
		KEY idx_status (status),
		KEY idx_created_at (created_at),
		KEY idx_source (source),
		KEY idx_agent_uid (agent_uid)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SXGZ orders'`)
	if err != nil {
		obslogger.L().Warn("SXGZ create orders table failed", "error", err)
	}

	_, err = database.DB.Exec(`CREATE TABLE IF NOT EXISTS fd_sxgz_company_cache (
		cid int(11) NOT NULL,
		name varchar(255) NOT NULL DEFAULT '',
		price decimal(10,2) NOT NULL DEFAULT 0.00,
		license_price decimal(10,2) NOT NULL DEFAULT 0.00,
		content text DEFAULT NULL,
		status tinyint(1) NOT NULL DEFAULT 1,
		raw_json longtext DEFAULT NULL,
		source varchar(20) NOT NULL DEFAULT 'upstream',
		updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (cid),
		KEY idx_name (name),
		KEY idx_status (status)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SXGZ company cache'`)
	if err != nil {
		obslogger.L().Warn("SXGZ create company cache table failed", "error", err)
	}

	defaultCfg := defaultSxgzConfig()
	raw, _ := defaultCfg.Marshal()
	var count int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_config WHERE v = ?", "sxgz_config").Scan(&count); err == nil && count == 0 {
		_, err = database.DB.Exec("INSERT INTO qingka_wangke_config (v, k) VALUES (?, ?)", "sxgz_config", raw)
		if err != nil {
			obslogger.L().Warn("SXGZ seed config failed", "error", err)
		}
	}
}
