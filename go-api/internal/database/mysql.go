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
	return db
}
