package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"go-api/internal/config"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
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
	if err := runConfiguredMigrations(db); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	bootstrapDefaultAdmin(db)
	return db
}

func runLegacySchemaPatches(db *sql.DB) {
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
}

func bootstrapDefaultAdmin(db *sql.DB) {
	// 默认管理员改为显式开关控制，避免启动期产生隐式高权限账号。
	var adminCount int
	db.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user WHERE grade='3'").Scan(&adminCount)
	if adminCount == 0 {
		if config.Global == nil || !config.Global.CreateDefaultAdminEnabled() {
			log.Println("[AutoMigrate] 未发现管理员账号，已跳过默认管理员自动创建；如需启用请设置 bootstrap.create_default_admin=true 或 GO_API_BOOTSTRAP_CREATE_DEFAULT_ADMIN=true")
			return
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("[AutoMigrate] 生成默认管理员密码哈希失败: %v", err)
			return
		}

		_, err = db.Exec("INSERT INTO qingka_wangke_user (uuid, user, pass, name, grade, active, addprice, addtime) VALUES (1, 'admin', ?, 'Admin', '3', '1', 1, NOW())", string(hashed))
		if err != nil {
			log.Printf("[AutoMigrate] 插入管理员失败: %v", err)
		} else {
			log.Println("[AutoMigrate] 已按显式开关创建默认管理员账号 admin/admin123")
		}
	}
}

func NormalizeLegacyUserPasswords(db *sql.DB) (int, int, error) {
	rows, err := db.Query("SELECT uid, COALESCE(pass,''), COALESCE(pass2,'') FROM qingka_wangke_user")
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	var passMigrated int
	var pass2Migrated int

	for rows.Next() {
		var uid int
		var pass string
		var pass2 string
		if err := rows.Scan(&uid, &pass, &pass2); err != nil {
			log.Printf("[PasswordMigration] 读取用户密码失败: %v", err)
			continue
		}

		newPass := pass
		newPass2 := pass2
		updatePass := false
		updatePass2 := false

		if pass != "" && !isBcryptHash(pass) {
			hashed, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("[PasswordMigration] 加密 uid=%d 的登录密码失败: %v", uid, err)
			} else {
				newPass = string(hashed)
				updatePass = true
				passMigrated++
			}
		}

		if pass2 != "" && !isBcryptHash(pass2) {
			hashed, err := bcrypt.GenerateFromPassword([]byte(pass2), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("[PasswordMigration] 加密 uid=%d 的二级密码失败: %v", uid, err)
			} else {
				newPass2 = string(hashed)
				updatePass2 = true
				pass2Migrated++
			}
		}

		switch {
		case updatePass && updatePass2:
			if _, err := db.Exec("UPDATE qingka_wangke_user SET pass = ?, pass2 = ? WHERE uid = ?", newPass, newPass2, uid); err != nil {
				log.Printf("[PasswordMigration] 更新 uid=%d 的密码字段失败: %v", uid, err)
			}
		case updatePass:
			if _, err := db.Exec("UPDATE qingka_wangke_user SET pass = ? WHERE uid = ?", newPass, uid); err != nil {
				log.Printf("[PasswordMigration] 更新 uid=%d 的登录密码失败: %v", uid, err)
			}
		case updatePass2:
			if _, err := db.Exec("UPDATE qingka_wangke_user SET pass2 = ? WHERE uid = ?", newPass2, uid); err != nil {
				log.Printf("[PasswordMigration] 更新 uid=%d 的二级密码失败: %v", uid, err)
			}
		}
	}
	return passMigrated, pass2Migrated, nil
}

func isBcryptHash(value string) bool {
	return strings.HasPrefix(value, "$2a$") || strings.HasPrefix(value, "$2b$") || strings.HasPrefix(value, "$2y$")
}
