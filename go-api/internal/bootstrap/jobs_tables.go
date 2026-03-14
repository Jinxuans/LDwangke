package bootstrap

import (
	"go-api/internal/database"
	papermodule "go-api/internal/modules/paper"
	tuboshumodule "go-api/internal/modules/tuboshu"
	tutuqgmodule "go-api/internal/modules/tutuqg"
	tuzhimodule "go-api/internal/modules/tuzhi"
	wmodule "go-api/internal/modules/w"
	xmmodule "go-api/internal/modules/xm"
	ydsjmodule "go-api/internal/modules/ydsj"
	yfdkmodule "go-api/internal/modules/yfdk"
	yongyemodule "go-api/internal/modules/yongye"
)

// InitTables 保持与旧启动流程一致的启动期表初始化。
func InitTables() {
	ensurePlatformTables()
	initMenuTable()
}

// 启动期建表只在 bootstrap 用一次，直接留在装配层即可，不再额外保留顶级聚合包。
func ensurePlatformTables() {
	tutuqgmodule.TutuQG().EnsureTable()
	tuboshumodule.Tuboshu().EnsureTable()
	yongyemodule.Yongye().EnsureTable()
	ydsjmodule.YDSJ().EnsureTable()
	wmodule.W().EnsureTable()
	xmmodule.XM().EnsureTable()
	papermodule.Paper().EnsureTable()
	yfdkmodule.YFDK().EnsureTable()
	tuzhimodule.TuZhi().EnsureTable()
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
