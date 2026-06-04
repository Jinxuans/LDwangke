package jiguang

import (
	"go-api/internal/database"
	obslogger "go-api/internal/observability/logger"
)

func (s *Service) EnsureTable() {
	ddls := []string{
		`CREATE TABLE IF NOT EXISTS qingka_wangke_jiguang (
			id INT NOT NULL AUTO_INCREMENT COMMENT '本地订单ID',
			uid INT NOT NULL COMMENT '本站用户ID',
			order_no VARCHAR(50) NOT NULL DEFAULT '' COMMENT '上游订单号',
			upstream_id INT NOT NULL DEFAULT 0 COMMENT '上游订单内部ID',
			product_id INT NOT NULL DEFAULT 0 COMMENT '上游商品ID',
			product_name VARCHAR(100) NOT NULL DEFAULT '' COMMENT '商品名称',
			school_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '学校名称',
			student_name VARCHAR(100) NOT NULL DEFAULT '' COMMENT '学生姓名',
			student_account VARCHAR(100) NOT NULL DEFAULT '' COMMENT '学生学号',
			run_times INT NOT NULL DEFAULT 0 COMMENT '跑步总次数',
			completed_times INT NOT NULL DEFAULT 0 COMMENT '已完成次数',
			km_per_day DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '每次公里数',
			customer_message VARCHAR(500) NOT NULL DEFAULT '' COMMENT '特殊备注',
			status VARCHAR(30) NOT NULL DEFAULT 'pending' COMMENT '订单状态',
			fees DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '订单金额',
			refund_amount DECIMAL(10,2) DEFAULT NULL COMMENT '退款金额',
			notes TEXT COMMENT '上游备注',
			source VARCHAR(30) NOT NULL DEFAULT 'local',
			agent_uid INT NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL COMMENT '创建时间',
			updated_at DATETIME DEFAULT NULL COMMENT '更新时间',
			refunded_at DATETIME DEFAULT NULL COMMENT '退款时间',
			PRIMARY KEY (id),
			UNIQUE KEY idx_order_no (order_no),
			KEY idx_uid (uid),
			KEY idx_student_account (student_account),
			KEY idx_status (status),
			KEY idx_product_id (product_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='极光跑步订单表'`,
	}
	for _, ddl := range ddls {
		if _, err := database.DB.Exec(ddl); err != nil {
			obslogger.L().Warn("Jiguang create table failed", "error", err)
		}
	}
	columns := []string{
		"ALTER TABLE qingka_wangke_jiguang ADD COLUMN source VARCHAR(30) NOT NULL DEFAULT 'local'",
		"ALTER TABLE qingka_wangke_jiguang ADD COLUMN agent_uid INT NOT NULL DEFAULT 0",
	}
	for _, ddl := range columns {
		_, _ = database.DB.Exec(ddl)
	}
	cfg := defaultConfig()
	raw, _ := cfg.Marshal()
	var count int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_config WHERE v = ?", "jiguang_config").Scan(&count); err == nil && count == 0 {
		if _, err := database.DB.Exec("INSERT INTO qingka_wangke_config (v, k) VALUES (?, ?)", "jiguang_config", raw); err != nil {
			obslogger.L().Warn("Jiguang seed config failed", "error", err)
		}
	}
}
