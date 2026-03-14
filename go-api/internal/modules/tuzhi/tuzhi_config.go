package tuzhi

import (
	"encoding/json"
	"log"

	"go-api/internal/database"
)

func (s *TuZhiService) EnsureTable() {
	log.Println("[TuZhi] 开始检查/创建表")
	_, err := database.DB.Exec(`CREATE TABLE IF NOT EXISTS qingka_wangke_dakaaz (
		id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID',
		api_id INT(11) DEFAULT NULL COMMENT '上游订单ID',
		user_id INT(11) NOT NULL COMMENT '用户ID',
		goods_id INT(11) NOT NULL DEFAULT 0 COMMENT '所属商品ID',
		username VARCHAR(40) NOT NULL COMMENT '账号',
		password VARCHAR(30) NOT NULL COMMENT '密码',
		nickname VARCHAR(20) DEFAULT NULL COMMENT '姓名',
		school VARCHAR(100) DEFAULT NULL COMMENT '学校名称',
		postname VARCHAR(50) DEFAULT NULL COMMENT '岗位名称',
		address VARCHAR(100) DEFAULT NULL COMMENT '地址',
		address_lat VARCHAR(50) DEFAULT NULL COMMENT '纬度',
		address_lng VARCHAR(50) DEFAULT NULL COMMENT '经度',
		work_time VARCHAR(100) DEFAULT NULL COMMENT '上班打卡时间',
		off_time VARCHAR(100) DEFAULT NULL COMMENT '下班打卡时间',
		work_days VARCHAR(100) DEFAULT NULL COMMENT '打卡周期',
		work_days_num BIGINT(20) DEFAULT NULL COMMENT '打卡天数',
		work_days_ok_num BIGINT(20) NOT NULL DEFAULT 0 COMMENT '已打卡天数',
		daily_report INT(11) DEFAULT 0 COMMENT '日报',
		weekly_report INT(11) DEFAULT 0 COMMENT '周报',
		monthly_report INT(10) UNSIGNED DEFAULT 0 COMMENT '月报',
		weekly_report_time BIGINT(20) DEFAULT NULL COMMENT '周报时间',
		monthly_report_time BIGINT(20) DEFAULT NULL COMMENT '月报时间',
		holiday_status INT(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '法定节假日 0=不跳过 1=跳过',
		token VARCHAR(255) DEFAULT NULL,
		uuid VARCHAR(255) DEFAULT NULL,
		user_school_id VARCHAR(255) DEFAULT NULL,
		random_phone VARCHAR(255) DEFAULT NULL,
		price DECIMAL(10,2) UNSIGNED NOT NULL DEFAULT 0.00 COMMENT '扣除金额',
		images TEXT COMMENT '图片',
		create_time BIGINT(20) DEFAULT NULL COMMENT '创建时间',
		update_time BIGINT(20) DEFAULT NULL COMMENT '更新时间',
		delete_time BIGINT(20) DEFAULT NULL COMMENT '删除时间',
		remark VARCHAR(100) DEFAULT '' COMMENT '备注',
		status INT(1) UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态 0正常 1打卡中 2关闭 3已完成',
		is_status INT(1) DEFAULT 1 COMMENT '打卡状态 0失败 1正常',
		work_deadline VARCHAR(100) DEFAULT NULL COMMENT '截至打卡日期',
		billing_method TINYINT(1) UNSIGNED DEFAULT 1 COMMENT '扣费方式（1按日，2按月）',
		billing_months TINYINT(3) UNSIGNED DEFAULT 0 COMMENT '收费月数',
		is_off_time TINYINT(1) UNSIGNED DEFAULT 1 COMMENT '是否开启下班打卡 1是0否',
		xz_push_url TEXT COMMENT '息知推送地址',
		PRIMARY KEY (id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='凸知打卡订单表'`)
	if err != nil {
		log.Printf("[TuZhi] 创建 dakaaz 表失败: %v", err)
	}

	_, err = database.DB.Exec(`CREATE TABLE IF NOT EXISTS qingka_wangke_daka_query_record (
		id INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
		username VARCHAR(255) DEFAULT NULL,
		password VARCHAR(255) DEFAULT NULL,
		create_time BIGINT(20) DEFAULT NULL COMMENT '创建时间',
		user_id INT(11) DEFAULT NULL,
		is_success TINYINT(1) UNSIGNED DEFAULT 0 COMMENT '是否成功',
		price DECIMAL(10,2) UNSIGNED NOT NULL DEFAULT 0.00 COMMENT '扣除金额',
		goods_id INT(11) NOT NULL DEFAULT 0 COMMENT '所属商品ID',
		PRIMARY KEY (id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='凸知打卡查询扣费记录'`)
	if err != nil {
		log.Printf("[TuZhi] 创建 daka_query_record 表失败: %v", err)
	}
	log.Println("[TuZhi] 表检查/创建完成")
}

func (s *TuZhiService) GetConfig() (*TuZhiConfig, error) {
	var val string
	err := database.DB.QueryRow("SELECT svalue FROM qingka_wangke_config WHERE skey = 'tuzhi_config' LIMIT 1").Scan(&val)
	if err != nil {
		return &TuZhiConfig{}, nil
	}
	var cfg TuZhiConfig
	json.Unmarshal([]byte(val), &cfg)
	return &cfg, nil
}

func (s *TuZhiService) SaveConfig(cfg *TuZhiConfig) error {
	data, _ := json.Marshal(cfg)
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_config WHERE skey = 'tuzhi_config'").Scan(&count)
	if count > 0 {
		_, err := database.DB.Exec("UPDATE qingka_wangke_config SET svalue = ? WHERE skey = 'tuzhi_config'", string(data))
		return err
	}
	_, err := database.DB.Exec("INSERT INTO qingka_wangke_config (v, k, skey, svalue) VALUES ('tuzhi_config', '', 'tuzhi_config', ?)", string(data))
	return err
}

func (s *TuZhiService) GetGoodsOverrides() ([]TuZhiGoodsOverride, error) {
	var val string
	err := database.DB.QueryRow("SELECT svalue FROM qingka_wangke_config WHERE skey = 'tuzhi_goods_overrides' LIMIT 1").Scan(&val)
	if err != nil {
		return []TuZhiGoodsOverride{}, nil
	}
	var list []TuZhiGoodsOverride
	json.Unmarshal([]byte(val), &list)
	return list, nil
}

func (s *TuZhiService) SaveGoodsOverrides(list []TuZhiGoodsOverride) error {
	data, _ := json.Marshal(list)
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_config WHERE skey = 'tuzhi_goods_overrides'").Scan(&count)
	if count > 0 {
		_, err := database.DB.Exec("UPDATE qingka_wangke_config SET svalue = ? WHERE skey = 'tuzhi_goods_overrides'", string(data))
		return err
	}
	_, err := database.DB.Exec("INSERT INTO qingka_wangke_config (v, k, skey, svalue) VALUES ('tuzhi_goods_overrides', '', 'tuzhi_goods_overrides', ?)", string(data))
	return err
}
