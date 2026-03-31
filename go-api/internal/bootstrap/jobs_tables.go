package bootstrap

import (
	"go-api/internal/database"
	"go-api/internal/pluginregistry"
)

// InitTables 保持与旧启动流程一致的启动期表初始化。
func InitTables() {
	pluginregistry.EnsureAllTables()
	initMenuTable()
}

func initMenuTable() {
	database.DB.Exec(`CREATE TABLE IF NOT EXISTS menu_config (
		id         INT AUTO_INCREMENT PRIMARY KEY,
		menu_key   VARCHAR(100) NOT NULL UNIQUE,
		parent_key VARCHAR(100) DEFAULT '',
		title      VARCHAR(100) DEFAULT '',
		icon       VARCHAR(200) DEFAULT '',
		sort_order INT DEFAULT 0,
		visible    TINYINT DEFAULT 1,
		scope      VARCHAR(20) DEFAULT 'frontend',
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
}
