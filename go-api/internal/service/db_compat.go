package service

import (
	"fmt"
	"log"
	"strings"
	"time"

	"go-api/internal/database"
)

// ===== 数据库结构检测与修复工具 =====

// ColumnDef 列定义
type ColumnDef struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	NotNull bool   `json:"not_null"`
	Default string `json:"default"`
	After   string `json:"after"`
	Comment string `json:"comment"`
}

// TableDef 表定义
type TableDef struct {
	Name       string      `json:"name"`
	PrimaryKey string      `json:"primary_key"`
	AutoInc    bool        `json:"auto_increment"`
	Columns    []ColumnDef `json:"columns"`
	UniqueKeys []string    `json:"unique_keys"`
	Engine     string      `json:"engine"`
	Charset    string      `json:"charset"`
}

// CompatCheckResult 检查结果
type CompatCheckResult struct {
	CheckTime      string              `json:"check_time"`
	TotalTables    int                 `json:"total_tables"`
	MissingTables  []string            `json:"missing_tables"`
	ExistingTables []string            `json:"existing_tables"`
	ExtraTables    []string            `json:"extra_tables"`
	MissingColumns []MissingColumnInfo `json:"missing_columns"`
	Summary        string              `json:"summary"`
}

// MissingColumnInfo 缺失列信息
type MissingColumnInfo struct {
	Table  string `json:"table"`
	Column string `json:"column"`
	Type   string `json:"type"`
}

// CompatFixResult 修复结果
type CompatFixResult struct {
	FixTime       string   `json:"fix_time"`
	TablesCreated []string `json:"tables_created"`
	ColumnsAdded  []string `json:"columns_added"`
	Errors        []string `json:"errors"`
	AdminCreated  bool     `json:"admin_created"`
	Summary       string   `json:"summary"`
}

type DBCompatService struct{}

func NewDBCompatService() *DBCompatService {
	return &DBCompatService{}
}

// getExpectedSchema 返回系统核心必需表的期望结构
// 只包含系统运行必需的核心表，扩展功能表由各模块按需自动创建
func (s *DBCompatService) getExpectedSchema() []TableDef {
	return []TableDef{
		// ===== 核心业务表（PHP原有） =====
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
				{Name: "dockstatus", Type: "INT(11)", NotNull: false, Default: "0"},
				{Name: "status", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "process", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "yid", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "remarks", Type: "TEXT", NotNull: false},
			},
		},
		{
			Name: "qingka_wangke_config", PrimaryKey: "v", AutoInc: false,
			Columns: []ColumnDef{
				{Name: "v", Type: "VARCHAR(255)", NotNull: true, Default: "''"},
				{Name: "k", Type: "TEXT", NotNull: false},
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
				{Name: "money", Type: "VARCHAR(255)", NotNull: false, Default: "'0'"},
				{Name: "status", Type: "INT(11)", NotNull: false, Default: "0"},
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
			},
		},
		{
			Name: "qingka_wangke_class", PrimaryKey: "cid", AutoInc: true,
			Columns: []ColumnDef{
				{Name: "cid", Type: "INT(11)", NotNull: true},
				{Name: "name", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "noun", Type: "VARCHAR(255)", NotNull: false, Default: "''"},
				{Name: "docking", Type: "VARCHAR(255)", NotNull: false, Default: "'0'"},
				{Name: "price", Type: "VARCHAR(255)", NotNull: false, Default: "'0'"},
				{Name: "yunsuan", Type: "VARCHAR(10)", NotNull: false, Default: "'*'"},
				{Name: "content", Type: "TEXT", NotNull: false},
				{Name: "fenlei", Type: "VARCHAR(50)", NotNull: false, Default: "'0'"},
				{Name: "status", Type: "INT(11)", NotNull: false, Default: "0"},
				{Name: "sort", Type: "INT(11)", NotNull: false, Default: "0"},
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
				{Name: "recommend", Type: "INT(11)", NotNull: false, Default: "0"},
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
			},
		},
	}
}

// listAllDBTables 动态获取当前数据库的所有表名
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

// Check 检测数据库结构差异（只读，不修改）
func (s *DBCompatService) Check() (*CompatCheckResult, error) {
	schema := s.getExpectedSchema()
	result := &CompatCheckResult{
		CheckTime:   time.Now().Format("2006-01-02 15:04:05"),
		TotalTables: len(schema),
	}

	// 获取数据库里实际的所有表
	allDBTables, _ := s.listAllDBTables()
	dbTableSet := make(map[string]bool)
	for _, t := range allDBTables {
		dbTableSet[t] = true
	}

	// 检查核心表是否存在
	requiredSet := make(map[string]bool)
	for _, tbl := range schema {
		requiredSet[tbl.Name] = true
		if !dbTableSet[tbl.Name] {
			result.MissingTables = append(result.MissingTables, tbl.Name)
			for _, col := range tbl.Columns {
				result.MissingColumns = append(result.MissingColumns, MissingColumnInfo{
					Table: tbl.Name, Column: col.Name, Type: col.Type,
				})
			}
		} else {
			result.ExistingTables = append(result.ExistingTables, tbl.Name)
			for _, col := range tbl.Columns {
				colExists, _ := s.columnExists(tbl.Name, col.Name)
				if !colExists {
					result.MissingColumns = append(result.MissingColumns, MissingColumnInfo{
						Table: tbl.Name, Column: col.Name, Type: col.Type,
					})
				}
			}
		}
	}

	// 统计数据库里有但不在核心列表中的"额外表"
	for _, t := range allDBTables {
		if !requiredSet[t] {
			result.ExtraTables = append(result.ExtraTables, t)
		}
	}

	if result.MissingTables == nil {
		result.MissingTables = []string{}
	}
	if result.ExistingTables == nil {
		result.ExistingTables = []string{}
	}
	if result.ExtraTables == nil {
		result.ExtraTables = []string{}
	}
	if result.MissingColumns == nil {
		result.MissingColumns = []MissingColumnInfo{}
	}

	result.Summary = fmt.Sprintf("核心表 %d 张（已有 %d / 缺失 %d），缺失列 %d 个，数据库额外表 %d 张",
		result.TotalTables, len(result.ExistingTables), len(result.MissingTables),
		len(result.MissingColumns), len(result.ExtraTables))

	return result, nil
}

// Fix 自动修复数据库结构（创建缺失表、添加缺失列）
func (s *DBCompatService) Fix() (*CompatFixResult, error) {
	schema := s.getExpectedSchema()
	result := &CompatFixResult{
		FixTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	for _, tbl := range schema {
		exists, err := s.tableExists(tbl.Name)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("检查表 %s 失败: %v", tbl.Name, err))
			continue
		}

		if !exists {
			// 创建表
			if err := s.createTable(tbl); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("创建表 %s 失败: %v", tbl.Name, err))
			} else {
				result.TablesCreated = append(result.TablesCreated, tbl.Name)
				log.Printf("[DBCompat] 创建表: %s", tbl.Name)
			}
		} else {
			// 检查并添加缺失列
			for _, col := range tbl.Columns {
				colExists, _ := s.columnExists(tbl.Name, col.Name)
				if !colExists {
					if err := s.addColumn(tbl.Name, col); err != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("表 %s 添加列 %s 失败: %v", tbl.Name, col.Name, err))
					} else {
						desc := fmt.Sprintf("%s.%s (%s)", tbl.Name, col.Name, col.Type)
						result.ColumnsAdded = append(result.ColumnsAdded, desc)
						log.Printf("[DBCompat] 添加列: %s", desc)
					}
				}
			}
		}
	}

	// 确保管理员账号存在
	var adminCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user WHERE grade='3'").Scan(&adminCount)
	if adminCount == 0 {
		_, err := database.DB.Exec(
			"INSERT INTO qingka_wangke_user (uuid, user, pass, name, grade, active, addprice, addtime) VALUES (1, 'admin', 'admin123', 'Admin', '3', '1', 1, NOW())",
		)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("创建管理员失败: %v", err))
		} else {
			result.AdminCreated = true
			log.Println("[DBCompat] 已自动创建管理员账号 admin/admin123")
		}
	}

	if result.TablesCreated == nil {
		result.TablesCreated = []string{}
	}
	if result.ColumnsAdded == nil {
		result.ColumnsAdded = []string{}
	}
	if result.Errors == nil {
		result.Errors = []string{}
	}

	result.Summary = fmt.Sprintf("创建了 %d 张表，添加了 %d 个列，%d 个错误",
		len(result.TablesCreated), len(result.ColumnsAdded), len(result.Errors))
	if result.AdminCreated {
		result.Summary += "，已创建管理员账号 admin/admin123"
	}

	return result, nil
}

// tableExists 检查表是否存在
func (s *DBCompatService) tableExists(tableName string) (bool, error) {
	var count int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?",
		tableName,
	).Scan(&count)
	return count > 0, err
}

// columnExists 检查列是否存在
func (s *DBCompatService) columnExists(tableName, columnName string) (bool, error) {
	var count int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?",
		tableName, columnName,
	).Scan(&count)
	return count > 0, err
}

// createTable 根据 TableDef 生成并执行 CREATE TABLE
func (s *DBCompatService) createTable(tbl TableDef) error {
	var cols []string
	for _, col := range tbl.Columns {
		cols = append(cols, s.columnSQL(col))
	}

	// 主键
	pk := fmt.Sprintf("PRIMARY KEY (`%s`)", tbl.PrimaryKey)
	cols = append(cols, pk)

	// 唯一索引
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

// addColumn 添加缺失列
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

// columnSQL 生成单个列的 SQL 定义
func (s *DBCompatService) columnSQL(col ColumnDef) string {
	parts := []string{fmt.Sprintf("`%s` %s", col.Name, col.Type)}

	if col.NotNull {
		parts = append(parts, "NOT NULL")
	} else {
		parts = append(parts, "NULL")
	}

	// 主键列如果 auto_increment，在 createTable 中已通过 PK 定义
	// 这里处理 DEFAULT
	if col.Default != "" {
		parts = append(parts, "DEFAULT "+col.Default)
	}

	if col.Comment != "" {
		parts = append(parts, fmt.Sprintf("COMMENT '%s'", col.Comment))
	}

	return strings.Join(parts, " ")
}
