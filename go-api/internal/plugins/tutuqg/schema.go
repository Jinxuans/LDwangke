package tutuqg

import (
	"go-api/internal/database"
	obslogger "go-api/internal/observability/logger"
)

func (s *tutuqgService) EnsureTable() {
	_, err := database.DB.Exec(`CREATE TABLE IF NOT EXISTS tutuqg (
		oid int(11) NOT NULL AUTO_INCREMENT,
		uid int(11) NOT NULL,
		user varchar(255) NOT NULL,
		pass varchar(255) NOT NULL,
		kcname varchar(255) NOT NULL,
		days varchar(255) NOT NULL,
		ptname varchar(255) NOT NULL,
		fees varchar(255) NOT NULL,
		addtime varchar(255) NOT NULL,
		IP varchar(255) DEFAULT NULL,
		status varchar(255) DEFAULT NULL,
		remarks varchar(255) DEFAULT NULL,
		guid varchar(255) DEFAULT NULL,
		score varchar(255) NOT NULL,
		scores varchar(255) DEFAULT NULL,
		zdxf varchar(255) DEFAULT NULL,
		PRIMARY KEY (oid)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		obslogger.L().Warn("TutuQG 创建表失败", "error", err)
	}
}
