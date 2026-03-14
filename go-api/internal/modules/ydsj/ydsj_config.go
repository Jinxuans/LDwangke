package ydsj

import (
	"encoding/json"
	"fmt"

	"go-api/internal/database"
)

// EnsureTable 确保运动世界订单表存在
func (s *YDSJService) EnsureTable() {
	_, err := database.DB.Exec(`CREATE TABLE IF NOT EXISTS qingka_wangke_hzw_ydsj (
		id int NOT NULL AUTO_INCREMENT,
		uid int NOT NULL DEFAULT 1 COMMENT '用户UID',
		school varchar(255) NOT NULL DEFAULT '' COMMENT '学校',
		user varchar(255) NOT NULL DEFAULT '' COMMENT '用户账号',
		pass varchar(255) NOT NULL DEFAULT '' COMMENT '用户密码',
		distance varchar(255) NOT NULL DEFAULT '' COMMENT '总公里数',
		is_run int NOT NULL DEFAULT 1 COMMENT '跑步状态 0关闭 1开启',
		run_type int NOT NULL DEFAULT 0 COMMENT '跑步类型',
		start_hour varchar(255) NOT NULL DEFAULT '' COMMENT '开始小时',
		start_minute varchar(255) NOT NULL DEFAULT '' COMMENT '开始分钟',
		end_hour varchar(255) NOT NULL DEFAULT '' COMMENT '结束小时',
		end_minute varchar(255) NOT NULL DEFAULT '' COMMENT '结束分钟',
		run_week varchar(255) NOT NULL DEFAULT '' COMMENT '跑步周期',
		status int NOT NULL DEFAULT 1 COMMENT '1等待 2成功 3失败 4退款',
		remarks varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
		fees varchar(255) NOT NULL DEFAULT '' COMMENT '预扣金额',
		real_fees varchar(255) NOT NULL DEFAULT '' COMMENT '实际金额',
		addtime varchar(255) NOT NULL DEFAULT '' COMMENT '下单时间',
		yid varchar(255) NOT NULL DEFAULT '' COMMENT '上游订单ID',
		info text COMMENT '订单信息',
		tmp_info text COMMENT '操作信息',
		refund_money varchar(255) NOT NULL DEFAULT '' COMMENT '退款金额',
		PRIMARY KEY (id),
		KEY idx_uid (uid),
		KEY idx_status (status),
		KEY idx_user (user(191))
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='运动世界订单表'`)
	if err != nil {
		fmt.Printf("[YDSJ] 建表失败: %v\n", err)
	}

	patchCols := []struct{ name, ddl string }{
		{"school", "ADD COLUMN `school` varchar(255) NOT NULL DEFAULT '' COMMENT '学校' AFTER `uid`"},
		{"distance", "ADD COLUMN `distance` varchar(255) NOT NULL DEFAULT '' COMMENT '总公里数' AFTER `pass`"},
		{"start_hour", "ADD COLUMN `start_hour` varchar(255) NOT NULL DEFAULT '' COMMENT '开始小时' AFTER `run_type`"},
		{"start_minute", "ADD COLUMN `start_minute` varchar(255) NOT NULL DEFAULT '' COMMENT '开始分钟' AFTER `start_hour`"},
		{"end_hour", "ADD COLUMN `end_hour` varchar(255) NOT NULL DEFAULT '' COMMENT '结束小时' AFTER `start_minute`"},
		{"end_minute", "ADD COLUMN `end_minute` varchar(255) NOT NULL DEFAULT '' COMMENT '结束分钟' AFTER `end_hour`"},
		{"run_week", "ADD COLUMN `run_week` varchar(255) NOT NULL DEFAULT '' COMMENT '跑步周期' AFTER `end_minute`"},
	}
	for _, col := range patchCols {
		var cnt int
		database.DB.QueryRow("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_wangke_hzw_ydsj' AND COLUMN_NAME=?", col.name).Scan(&cnt)
		if cnt == 0 {
			_, e := database.DB.Exec("ALTER TABLE `qingka_wangke_hzw_ydsj` " + col.ddl)
			if e != nil {
				fmt.Printf("[YDSJ] 补列 %s 失败: %v\n", col.name, e)
			} else {
				fmt.Printf("[YDSJ] 补列 %s 成功\n", col.name)
			}
		}
	}
}

func (s *YDSJService) GetConfig() (*YDSJConfig, error) {
	var val string
	err := database.DB.QueryRow("SELECT svalue FROM qingka_wangke_config WHERE skey = 'ydsj_config' LIMIT 1").Scan(&val)
	if err != nil {
		return &YDSJConfig{PriceMultiple: 5, XbdMorningPrice: 6, XbdExercisePrice: 6.5, RealCostMultiple: 1}, nil
	}
	var cfg YDSJConfig
	json.Unmarshal([]byte(val), &cfg)
	return &cfg, nil
}

func (s *YDSJService) SaveConfig(cfg *YDSJConfig) error {
	data, _ := json.Marshal(cfg)
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_config (v, k, skey, svalue) VALUES ('ydsj_config', '', 'ydsj_config', ?) ON DUPLICATE KEY UPDATE svalue = ?",
		string(data), string(data),
	)
	return err
}
