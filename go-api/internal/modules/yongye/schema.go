package yongye

import (
	"go-api/internal/database"
	obslogger "go-api/internal/observability/logger"
)

func Yongye() *yongyeService {
	return yongyeServiceInstance
}

func (s *yongyeService) EnsureTable() {
	_, err := database.DB.Exec(`CREATE TABLE IF NOT EXISTS yy_ydsj_dd (
		id INT NOT NULL AUTO_INCREMENT,
		pol TINYINT NOT NULL DEFAULT 0 COMMENT '轮询模式 0=否 1=是',
		uid INT NOT NULL DEFAULT 0,
		user VARCHAR(100) NOT NULL DEFAULT '' COMMENT '学号',
		pass VARCHAR(100) NOT NULL DEFAULT '' COMMENT '密码',
		school VARCHAR(100) NOT NULL DEFAULT '自动识别' COMMENT '学校',
		type TINYINT NOT NULL DEFAULT 0 COMMENT '跑步类型 0=正常 1=晨跑',
		zkm DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '公里数',
		ks_h INT NOT NULL DEFAULT 9 COMMENT '开始小时',
		ks_m INT NOT NULL DEFAULT 0 COMMENT '开始分钟',
		js_h INT NOT NULL DEFAULT 21 COMMENT '结束小时',
		js_m INT NOT NULL DEFAULT 0 COMMENT '结束分钟',
		weeks VARCHAR(20) NOT NULL DEFAULT '' COMMENT '跑步周天',
		dockstatus TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0=未提交 1=已提交 2=失败 3=关闭 5=轮询',
		yfees DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '预扣费用',
		fees DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '实际费用',
		yid VARCHAR(50) NOT NULL DEFAULT '' COMMENT '上游订单ID',
		yaddtime VARCHAR(50) NOT NULL DEFAULT '',
		addtime DATETIME DEFAULT NULL,
		tktext TEXT COMMENT '状态日志',
		PRIMARY KEY (id),
		KEY idx_uid (uid),
		KEY idx_user (user),
		KEY idx_dockstatus (dockstatus)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='永夜运动订单表'`)
	if err != nil {
		obslogger.L().Warn("Yongye 建表 yy_ydsj_dd 失败", "error", err)
	}

	_, err = database.DB.Exec(`CREATE TABLE IF NOT EXISTS yy_ydsj_student (
		id INT NOT NULL AUTO_INCREMENT,
		uid INT NOT NULL DEFAULT 0,
		user VARCHAR(100) NOT NULL DEFAULT '' COMMENT '学号',
		pass VARCHAR(100) NOT NULL DEFAULT '' COMMENT '密码',
		type TINYINT NOT NULL DEFAULT 0 COMMENT '跑步类型',
		zkm DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '公里数',
		weeks VARCHAR(20) NOT NULL DEFAULT '' COMMENT '跑步周天',
		status TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0=正常 1=暂停 2=完成 3=退单',
		tdkm DECIMAL(10,2) DEFAULT NULL COMMENT '退单公里',
		tdmoney DECIMAL(10,2) DEFAULT NULL COMMENT '退单金额',
		stulog TEXT COMMENT '学生日志JSON',
		last_time DATETIME DEFAULT NULL,
		PRIMARY KEY (id),
		KEY idx_uid (uid),
		KEY idx_user (user)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='永夜运动学生表'`)
	if err != nil {
		obslogger.L().Warn("Yongye 建表 yy_ydsj_student 失败", "error", err)
	}
}
