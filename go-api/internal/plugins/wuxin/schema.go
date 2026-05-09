package wuxin

import (
	"go-api/internal/database"
	obslogger "go-api/internal/observability/logger"
)

func (s *WuxinService) EnsureTable() {
	_, err := database.DB.Exec(`CREATE TABLE IF NOT EXISTS wuxin_sdxy (
		id INT NOT NULL AUTO_INCREMENT COMMENT '订单ID',
		user_id INT NOT NULL DEFAULT 1 COMMENT '本站用户id',
		auth_code VARCHAR(255) NOT NULL DEFAULT '' COMMENT '授权码',
		start_date VARCHAR(255) NOT NULL DEFAULT '' COMMENT '开始跑步日期',
		residue_num INT NOT NULL DEFAULT 0 COMMENT '剩余次数',
		quantity INT NOT NULL DEFAULT 0 COMMENT '购买总数',
		completed_quantity INT NOT NULL DEFAULT 0 COMMENT '已完成数量',
		run_meter DECIMAL(11,1) NOT NULL DEFAULT 1.0 COMMENT '跑步距离',
		run_type INT NOT NULL DEFAULT 2 COMMENT '跑步类型',
		zone_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '跑区名称',
		zone_code VARCHAR(50) NOT NULL DEFAULT '' COMMENT '跑区代码',
		zone_id INT NOT NULL DEFAULT 0 COMMENT '跑区ID',
		run_time VARCHAR(255) NOT NULL DEFAULT '' COMMENT '跑步时间段',
		run_week VARCHAR(255) NOT NULL DEFAULT '' COMMENT '跑步周期',
		run_speed VARCHAR(255) NOT NULL DEFAULT '' COMMENT '跑步配速',
		status INT NOT NULL DEFAULT 0 COMMENT '本地状态',
		order_status INT NOT NULL DEFAULT 0 COMMENT '源台订单状态',
		run_status INT NOT NULL DEFAULT 1 COMMENT '跑步状态',
		mark VARCHAR(255) NOT NULL DEFAULT '' COMMENT '客户标记信息',
		remarks VARCHAR(255) NOT NULL DEFAULT '' COMMENT '源台备注',
		phone VARCHAR(20) NOT NULL DEFAULT '' COMMENT '手机号码',
		account_flag INT NOT NULL DEFAULT 0 COMMENT '授权状态',
		create_time VARCHAR(255) NOT NULL DEFAULT '' COMMENT '创建时间',
		update_time VARCHAR(255) NOT NULL DEFAULT '' COMMENT '更新时间',
		order_number VARCHAR(50) NOT NULL DEFAULT '0' COMMENT '源台订单号',
		run_plan_code VARCHAR(50) NOT NULL DEFAULT '' COMMENT '跑步计划代码',
		fence_code VARCHAR(50) NOT NULL DEFAULT '' COMMENT '围栏代码',
		schedule_config TEXT NULL COMMENT '调度配置JSON',
		next_execute_date VARCHAR(50) NOT NULL DEFAULT '' COMMENT '下次执行时间',
		fees DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '订单金额',
		source VARCHAR(30) NOT NULL DEFAULT 'local' COMMENT '订单来源',
		agent_uid INT NOT NULL DEFAULT 0 COMMENT '上级代理UID',
		PRIMARY KEY (id),
		KEY idx_user_id (user_id),
		KEY idx_order_number (order_number),
		KEY idx_auth_code (auth_code),
		KEY idx_order_status (order_status)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='无心闪动订单表'`)
	if err != nil {
		obslogger.L().Warn("Wuxin create table failed", "error", err)
	}

	columns := []string{
		"ALTER TABLE wuxin_sdxy ADD COLUMN phone VARCHAR(20) NOT NULL DEFAULT '' COMMENT '手机号码' AFTER remarks",
		"ALTER TABLE wuxin_sdxy ADD COLUMN fees DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '订单金额'",
		"ALTER TABLE wuxin_sdxy ADD COLUMN source VARCHAR(30) NOT NULL DEFAULT 'local' COMMENT '订单来源'",
		"ALTER TABLE wuxin_sdxy ADD COLUMN agent_uid INT NOT NULL DEFAULT 0 COMMENT '上级代理UID'",
	}
	for _, ddl := range columns {
		_, _ = database.DB.Exec(ddl)
	}

	cfg := defaultWuxinConfig()
	raw, _ := cfg.Marshal()
	var count int
	if err := database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_config WHERE v = ?", "wuxin_config").Scan(&count); err == nil && count == 0 {
		_, err = database.DB.Exec("INSERT INTO qingka_wangke_config (v, k) VALUES (?, ?)", "wuxin_config", raw)
		if err != nil {
			obslogger.L().Warn("Wuxin seed config failed", "error", err)
		}
	}
}
