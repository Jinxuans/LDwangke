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
