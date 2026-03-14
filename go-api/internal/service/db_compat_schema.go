package service

import (
	"fmt"
	"strings"

	"go-api/internal/database"
)

func (s *DBCompatService) getExpectedSchema() []TableDef {
	return []TableDef{
		{
			Name: "qingka_wangke_user", PrimaryKey: "uid", AutoInc: true,
			Columns: []ColumnDef{
				{Name: "uid", Type: "INT(11)", NotNull: true},
				{Name: "uuid", Type: "INT(11)", NotNull: false, Default: "0"},
				{Name: "user", Type: "VARCHAR(255)", NotNull: true, Default: "''"},
				{Name: "pass", Type: "VARCHAR(255)", NotNull: true, Default: "''"},
				{Name: "pass2", Type: "VARCHAR(255)", NotNull: true, Default: "''", Comment: "管理员二级密码"},
				{Name: "name", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "money", Type: "DECIMAL(12,4)", NotNull: false, Default: "0"},
				{Name: "grade", Type: "VARCHAR(10)", NotNull: false, Default: "'0'"},
				{Name: "active", Type: "VARCHAR(10)", NotNull: false, Default: "'1'"},
				{Name: "addprice", Type: "DECIMAL(10,2)", NotNull: false, Default: "1"},
				{Name: "key", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "yqm", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "yqprice", Type: "VARCHAR(255)", NotNull: false, Default: "'0'"},
				{Name: "email", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "tuisongtoken", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "zcz", Type: "DECIMAL(12,4)", NotNull: false, Default: "0"},
				{Name: "cdmoney", Type: "DECIMAL(12,4)", NotNull: false, Default: "0"},
				{Name: "notice", Type: "TEXT", NotNull: false},
				{Name: "paydata", Type: "TEXT", NotNull: false},
				{Name: "addtime", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
			},
		},
		{
			Name: "qingka_wangke_order", PrimaryKey: "oid", AutoInc: true,
			Columns: []ColumnDef{
				{Name: "oid", Type: "INT(11)", NotNull: true},
				{Name: "uid", Type: "INT(11)", NotNull: false, Default: "0"},
				{Name: "cid", Type: "INT(11)", NotNull: false, Default: "0"},
				{Name: "hid", Type: "INT(11)", NotNull: false, Default: "0"},
				{Name: "ptname", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "school", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "name", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "user", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "pass", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "kcid", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "kcname", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "fees", Type: "VARCHAR(255)", NotNull: false, Default: "'0'"},
				{Name: "noun", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "addtime", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "ip", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "dockstatus", Type: "VARCHAR(255)", NotNull: false, Default: "'0'"},
				{Name: "status", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "process", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "yid", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "remarks", Type: "TEXT", NotNull: false},
				{Name: "pushUid", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "pushStatus", Type: "VARCHAR(50)", NotNull: false, Default: "''"},
				{Name: "pushEmail", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "pushEmailStatus", Type: "VARCHAR(50)", NotNull: false, Default: "'0'"},
				{Name: "showdoc_push_url", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "pushShowdocStatus", Type: "VARCHAR(50)", NotNull: false, Default: "'0'"},
				{Name: "work_state", Type: "TINYINT(4)", NotNull: false, Default: "0"},
			},
		},
		{
			Name: "qingka_wangke_config", PrimaryKey: "v", AutoInc: false,
			Columns: []ColumnDef{
				{Name: "v", Type: "VARCHAR(255)", NotNull: true, Default: "''"},
				{Name: "k", Type: "TEXT", NotNull: false},
				{Name: "skey", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "svalue", Type: "MEDIUMTEXT", NotNull: false},
			},
		},
		{
			Name: "qingka_wangke_moneylog", PrimaryKey: "id", AutoInc: true,
			Columns: []ColumnDef{
				{Name: "id", Type: "INT(11)", NotNull: true},
				{Name: "uid", Type: "INT(11)", NotNull: false, Default: "0"},
				{Name: "type", Type: "VARCHAR(50)", NotNull: false, Default: "''"},
				{Name: "money", Type: "DECIMAL(12,4)", NotNull: false, Default: "0"},
				{Name: "balance", Type: "DECIMAL(12,4)", NotNull: false, Default: "0"},
				{Name: "remark", Type: "VARCHAR(500)", NotNull: false, Default: "''"},
				{Name: "addtime", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
			},
		},
		{
			Name: "qingka_wangke_pay", PrimaryKey: "oid", AutoInc: true,
			Columns: []ColumnDef{
				{Name: "oid", Type: "INT(11)", NotNull: true},
				{Name: "uid", Type: "INT(11)", NotNull: false, Default: "0"},
				{Name: "out_trade_no", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "trade_no", Type: "VARCHAR(100)", NotNull: false, Default: "''"},
				{Name: "money", Type: "VARCHAR(255)", NotNull: false, Default: "'0'"},
				{Name: "status", Type: "INT(11)", NotNull: false, Default: "0"},
				{Name: "num", Type: "INT(11)", NotNull: false, Default: "1"},
				{Name: "name", Type: "VARCHAR(64)", NotNull: false},
				{Name: "ip", Type: "VARCHAR(20)", NotNull: false},
				{Name: "domain", Type: "VARCHAR(64)", NotNull: false},
				{Name: "addtime", Type: "DATETIME", NotNull: false},
			},
		},
		{
			Name: "qingka_wangke_dengji", PrimaryKey: "id", AutoInc: true,
			Columns: []ColumnDef{
				{Name: "id", Type: "INT(11)", NotNull: true},
				{Name: "rate", Type: "VARCHAR(50)", NotNull: false, Default: "''"},
				{Name: "name", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "status", Type: "VARCHAR(10)", NotNull: false, Default: "'1'"},
				{Name: "sort", Type: "INT(11)", NotNull: false, Default: "0"},
				{Name: "money", Type: "DECIMAL(10,2)", NotNull: false, Default: "0"},
				{Name: "addkf", Type: "VARCHAR(11)", NotNull: false, Default: "'0'"},
				{Name: "gjkf", Type: "VARCHAR(11)", NotNull: false, Default: "'0'"},
				{Name: "time", Type: "VARCHAR(11)", NotNull: false, Default: "''"},
			},
		},
		{
			Name: "qingka_wangke_class", PrimaryKey: "cid", AutoInc: true,
			Columns: []ColumnDef{
				{Name: "cid", Type: "INT(11)", NotNull: true},
				{Name: "name", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "noun", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "getnoun", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "docking", Type: "VARCHAR(255)", NotNull: false, Default: "'0'"},
				{Name: "queryplat", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "price", Type: "VARCHAR(255)", NotNull: false, Default: "'0'"},
				{Name: "yunsuan", Type: "VARCHAR(10)", NotNull: false, Default: "'*'"},
				{Name: "content", Type: "TEXT", NotNull: false},
				{Name: "fenlei", Type: "VARCHAR(50)", NotNull: false, Default: "'0'"},
				{Name: "status", Type: "INT(11)", NotNull: false, Default: "0"},
				{Name: "sort", Type: "INT(11)", NotNull: false, Default: "0"},
				{Name: "addtime", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
			},
		},
		{
			Name: "qingka_wangke_fenlei", PrimaryKey: "id", AutoInc: true,
			Columns: []ColumnDef{
				{Name: "id", Type: "INT(11)", NotNull: true},
				{Name: "sort", Type: "INT(11)", NotNull: false, Default: "0"},
				{Name: "name", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "status", Type: "VARCHAR(10)", NotNull: false, Default: "'1'"},
				{Name: "time", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "recommend", Type: "TINYINT(4)", NotNull: false, Default: "0"},
				{Name: "log", Type: "TINYINT(4)", NotNull: false, Default: "0"},
				{Name: "ticket", Type: "TINYINT(4)", NotNull: false, Default: "0"},
				{Name: "changepass", Type: "TINYINT(4)", NotNull: false, Default: "1"},
				{Name: "allowpause", Type: "TINYINT(4)", NotNull: false, Default: "0"},
				{Name: "supplier_report", Type: "TINYINT(4)", NotNull: false, Default: "0"},
				{Name: "supplier_report_hid", Type: "INT(11)", NotNull: false, Default: "0"},
			},
		},
		{
			Name: "qingka_wangke_log", PrimaryKey: "id", AutoInc: true,
			Columns: []ColumnDef{
				{Name: "id", Type: "INT(11)", NotNull: true},
				{Name: "uid", Type: "INT(11)", NotNull: false, Default: "0"},
				{Name: "type", Type: "VARCHAR(50)", NotNull: false, Default: "''"},
				{Name: "text", Type: "TEXT", NotNull: false},
				{Name: "money", Type: "VARCHAR(50)", NotNull: false, Default: "''"},
				{Name: "smoney", Type: "VARCHAR(50)", NotNull: false, Default: "''"},
				{Name: "ip", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "addtime", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
			},
		},
		{
			Name: "qingka_wangke_huoyuan", PrimaryKey: "hid", AutoInc: true,
			Columns: []ColumnDef{
				{Name: "hid", Type: "INT(11)", NotNull: true},
				{Name: "pt", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "name", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "url", Type: "VARCHAR(500)", NotNull: false, Default: "''"},
				{Name: "user", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "pass", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "token", Type: "VARCHAR(500)", NotNull: false, Default: "''"},
				{Name: "money", Type: "VARCHAR(50)", NotNull: false, Default: "'0'"},
				{Name: "status", Type: "VARCHAR(10)", NotNull: false, Default: "'1'"},
				{Name: "addtime", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
			},
		},
		{
			Name: "qingka_wangke_gonggao", PrimaryKey: "id", AutoInc: true,
			Columns: []ColumnDef{
				{Name: "id", Type: "INT(11)", NotNull: true},
				{Name: "title", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "content", Type: "TEXT", NotNull: false},
				{Name: "time", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "uid", Type: "INT(11)", NotNull: false, Default: "0"},
				{Name: "status", Type: "VARCHAR(10)", NotNull: false, Default: "'1'"},
				{Name: "zhiding", Type: "VARCHAR(10)", NotNull: false, Default: "'0'"},
				{Name: "uptime", Type: "TEXT", NotNull: false},
				{Name: "author", Type: "TEXT", NotNull: false},
				{Name: "visibility", Type: "INT", NotNull: true, Default: "0"},
			},
		},
		{
			Name: "qingka_ext_menu", PrimaryKey: "id", AutoInc: true,
			Columns: []ColumnDef{
				{Name: "id", Type: "INT(11)", NotNull: true},
				{Name: "title", Type: "VARCHAR(100)", NotNull: true, Default: "''"},
				{Name: "icon", Type: "VARCHAR(100)", NotNull: false, Default: "''"},
				{Name: "url", Type: "VARCHAR(500)", NotNull: true, Default: "''"},
				{Name: "sort_order", Type: "INT(11)", NotNull: false, Default: "0"},
				{Name: "visible", Type: "INT(11)", NotNull: false, Default: "1"},
				{Name: "scope", Type: "VARCHAR(20)", NotNull: false, Default: "'backend'"},
				{Name: "created_at", Type: "VARCHAR(50)", NotNull: false, Default: "''"},
			},
		},
		{
			Name: "qingka_email_template", PrimaryKey: "id", AutoInc: true,
			Columns: []ColumnDef{
				{Name: "id", Type: "INT(11)", NotNull: true},
				{Name: "code", Type: "VARCHAR(50)", NotNull: true, Default: "''"},
				{Name: "name", Type: "VARCHAR(100)", NotNull: true, Default: "''"},
				{Name: "subject", Type: "VARCHAR(255)", NotNull: true, Default: "''"},
				{Name: "content", Type: "TEXT", NotNull: false},
				{Name: "variables", Type: "VARCHAR(500)", NotNull: false, Default: "''"},
				{Name: "status", Type: "INT(11)", NotNull: false, Default: "1"},
				{Name: "updated_at", Type: "VARCHAR(50)", NotNull: false, Default: "''"},
				{Name: "created_at", Type: "VARCHAR(50)", NotNull: false, Default: "''"},
			},
			UniqueKeys: []string{"code"},
		},
	}
}

func (s *DBCompatService) listAllDBTables() ([]string, error) {
	rows, err := database.DB.Query(
		"SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() ORDER BY TABLE_NAME")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		tables = append(tables, name)
	}
	return tables, nil
}

func (s *DBCompatService) tableExists(tableName string) (bool, error) {
	var count int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?",
		tableName,
	).Scan(&count)
	return count > 0, err
}

func (s *DBCompatService) columnExists(tableName, columnName string) (bool, error) {
	var count int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?",
		tableName, columnName,
	).Scan(&count)
	return count > 0, err
}

func (s *DBCompatService) createTable(tbl TableDef) error {
	var cols []string
	for _, col := range tbl.Columns {
		cols = append(cols, s.columnSQL(col))
	}

	pk := fmt.Sprintf("PRIMARY KEY (`%s`)", tbl.PrimaryKey)
	cols = append(cols, pk)

	for _, uk := range tbl.UniqueKeys {
		cols = append(cols, fmt.Sprintf("UNIQUE KEY `uk_%s` (`%s`)", uk, uk))
	}

	engine := tbl.Engine
	if engine == "" {
		engine = "InnoDB"
	}
	charset := tbl.Charset
	if charset == "" {
		charset = "utf8mb4"
	}

	autoInc := ""
	if tbl.AutoInc {
		autoInc = " AUTO_INCREMENT=1"
	}

	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n  %s\n) ENGINE=%s DEFAULT CHARSET=%s%s",
		tbl.Name, strings.Join(cols, ",\n  "), engine, charset, autoInc)

	_, err := database.DB.Exec(sql)
	return err
}

func (s *DBCompatService) addColumn(tableName string, col ColumnDef) error {
	colDef := s.columnSQL(col)
	afterClause := ""
	if col.After != "" {
		afterClause = fmt.Sprintf(" AFTER `%s`", col.After)
	}
	sql := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN %s%s", tableName, colDef, afterClause)
	_, err := database.DB.Exec(sql)
	return err
}

func (s *DBCompatService) columnSQL(col ColumnDef) string {
	parts := []string{fmt.Sprintf("`%s` %s", col.Name, col.Type)}

	if col.NotNull {
		parts = append(parts, "NOT NULL")
	} else {
		parts = append(parts, "NULL")
	}

	if col.Default != "" {
		parts = append(parts, "DEFAULT "+col.Default)
	}

	if col.Comment != "" {
		parts = append(parts, fmt.Sprintf("COMMENT '%s'", col.Comment))
	}

	return strings.Join(parts, " ")
}
