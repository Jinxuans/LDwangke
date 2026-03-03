package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"go-api/internal/config"

	"github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Connect(cfg config.DatabaseConfig) *sql.DB {
	mysqlCfg := mysql.NewConfig()
	mysqlCfg.User = cfg.User
	mysqlCfg.Passwd = cfg.Password
	mysqlCfg.Net = "tcp"
	mysqlCfg.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	mysqlCfg.DBName = cfg.DBName
	mysqlCfg.ParseTime = true
	mysqlCfg.Loc = nil // Local
	mysqlCfg.Collation = "utf8mb4_general_ci"
	// 关闭严格模式，与PHP行为一致（NOT NULL字段无默认值时自动填零值）
	mysqlCfg.Params = map[string]string{
		"sql_mode": "''",
	}

	dsn := mysqlCfg.FormatDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("数据库 Ping 失败: %v", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(3 * time.Minute)

	DB = db
	log.Println("MySQL 连接成功")
	autoMigrate(db)
	return db
}

func autoMigrate(db *sql.DB) {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_wangke_user' AND COLUMN_NAME='pass2'").Scan(&count)
	if count == 0 {
		_, err := db.Exec("ALTER TABLE `qingka_wangke_user` ADD COLUMN `pass2` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '管理员二级密码' AFTER `pass`")
		if err != nil {
			log.Printf("[AutoMigrate] 添加 pass2 列失败: %v", err)
		} else {
			log.Println("[AutoMigrate] 已自动添加 pass2 列")
		}
	}
	// 确保 config 表有 skey/svalue 列（新模块用）
	var hasSkey int
	db.QueryRow("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_wangke_config' AND COLUMN_NAME='skey'").Scan(&hasSkey)
	if hasSkey == 0 {
		// 分步添加：先加列（NULL 默认值，避免与已有行冲突），再加唯一索引
		_, err := db.Exec("ALTER TABLE `qingka_wangke_config` ADD COLUMN `skey` VARCHAR(255) DEFAULT NULL AFTER `k`, ADD COLUMN `svalue` MEDIUMTEXT AFTER `skey`")
		if err != nil {
			log.Printf("[AutoMigrate] 添加 config.skey/svalue 列失败: %v", err)
		} else {
			log.Println("[AutoMigrate] 已添加 config.skey/svalue 列")
			// 再单独添加唯一索引（NULL 值不冲突）
			_, err = db.Exec("ALTER TABLE `qingka_wangke_config` ADD UNIQUE KEY `uk_skey` (`skey`)")
			if err != nil {
				log.Printf("[AutoMigrate] 添加 config.uk_skey 索引失败(可能已存在): %v", err)
			}
		}
	} else {
		// 列已存在，确保唯一索引也存在
		var hasUK int
		db.QueryRow("SELECT COUNT(*) FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_wangke_config' AND INDEX_NAME='uk_skey'").Scan(&hasUK)
		if hasUK == 0 {
			// 先把旧的空字符串 skey 改为 NULL，避免唯一索引冲突
			db.Exec("UPDATE `qingka_wangke_config` SET `skey` = NULL WHERE `skey` = ''")
			_, err := db.Exec("ALTER TABLE `qingka_wangke_config` ADD UNIQUE KEY `uk_skey` (`skey`)")
			if err != nil {
				log.Printf("[AutoMigrate] 补建 config.uk_skey 索引失败: %v", err)
			} else {
				log.Println("[AutoMigrate] 已补建 config.uk_skey 索引")
			}
		}
	}

	// 确保 moneylog 表有 mark/remarks 列（部分模块用）
	var hasMark int
	db.QueryRow("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='qingka_wangke_moneylog' AND COLUMN_NAME='mark'").Scan(&hasMark)
	if hasMark == 0 {
		db.Exec("ALTER TABLE `qingka_wangke_moneylog` ADD COLUMN `mark` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '备注(别名)' AFTER `remark`, ADD COLUMN `remarks` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '备注(别名2)' AFTER `mark`")
		log.Println("[AutoMigrate] 已添加 moneylog.mark/remarks 列")
	}

	// 确保管理员账号存在
	var adminCount int
	db.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user WHERE grade='3'").Scan(&adminCount)
	if adminCount == 0 {
		_, err := db.Exec("INSERT INTO qingka_wangke_user (uuid, user, pass, name, grade, active, addprice, addtime) VALUES (1, 'admin', 'admin123', 'Admin', '3', '1', 1, NOW())")
		if err != nil {
			log.Printf("[AutoMigrate] 插入管理员失败: %v", err)
		} else {
			log.Println("[AutoMigrate] 已自动创建管理员账号 admin/admin123")
		}
	}
}
