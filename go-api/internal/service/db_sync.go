package service

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"go-api/internal/database"

	"github.com/go-sql-driver/mysql"
)

// ===== 数据同步工具：从其他29系统数据库同步数据到当前系统 =====

// SyncRequest 同步请求参数
type SyncRequest struct {
	Host           string `json:"host" binding:"required"`
	Port           int    `json:"port"`
	DBName         string `json:"db_name" binding:"required"`
	User           string `json:"user" binding:"required"`
	Password       string `json:"password"`
	UpdateExisting bool   `json:"update_existing"`
}

// SyncResult 同步结果
type SyncResult struct {
	SyncTime string          `json:"sync_time"`
	Success  bool            `json:"success"`
	Details  []SyncTableInfo `json:"details"`
	Errors   []string        `json:"errors"`
	Summary  string          `json:"summary"`
}

// SyncTableInfo 每张表的同步统计
type SyncTableInfo struct {
	Table    string `json:"table"`
	Label    string `json:"label"`
	Total    int    `json:"total"`
	Inserted int    `json:"inserted"`
	Updated  int    `json:"updated"`
	Skipped  int    `json:"skipped"`
	Failed   int    `json:"failed"`
}

// SyncTestResult 连接测试结果
type SyncTestResult struct {
	Connected bool           `json:"connected"`
	Tables    map[string]int `json:"tables"`
	Error     string         `json:"error,omitempty"`
}

type DBSyncService struct{}

func NewDBSyncService() *DBSyncService {
	return &DBSyncService{}
}

// connectExternal 连接外部数据库
func (s *DBSyncService) connectExternal(req SyncRequest) (*sql.DB, error) {
	if req.Port == 0 {
		req.Port = 3306
	}
	cfg := mysql.NewConfig()
	cfg.User = req.User
	cfg.Passwd = req.Password
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%s:%d", req.Host, req.Port)
	cfg.DBName = req.DBName
	cfg.ParseTime = true
	cfg.Collation = "utf8mb4_general_ci"
	cfg.Params = map[string]string{
		"charset":  "utf8mb4",
		"sql_mode": "''",
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("连接失败: %v", err)
	}
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("无法连接到数据库: %v", err)
	}
	db.Exec("SET NAMES utf8mb4")
	db.Exec("SET CHARACTER SET utf8mb4")
	db.SetMaxOpenConns(5)
	db.SetConnMaxLifetime(2 * time.Minute)
	return db, nil
}

// getTableColumns 获取指定数据库中某表的列名列表（按原始顺序）
func (s *DBSyncService) getTableColumns(db *sql.DB, tableName string) ([]string, error) {
	rows, err := db.Query(
		"SELECT COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION",
		tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []string
	for rows.Next() {
		var col string
		rows.Scan(&col)
		cols = append(cols, col)
	}
	return cols, nil
}

// getColumnType 获取源库某列的类型（用于在目标库自动补列）
func (s *DBSyncService) getColumnType(db *sql.DB, tableName, columnName string) string {
	var colType string
	db.QueryRow(
		"SELECT COLUMN_TYPE FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?",
		tableName, columnName).Scan(&colType)
	if colType == "" {
		colType = "TEXT"
	}
	return colType
}

// ensureLocalColumns 确保本地表拥有 columns 中列出的所有列，缺失的自动补上
func (s *DBSyncService) ensureLocalColumns(tableName string, extDB *sql.DB, columns []string) {
	localCols, err := s.getTableColumns(database.DB, tableName)
	if err != nil {
		return
	}
	localSet := make(map[string]bool)
	for _, c := range localCols {
		localSet[strings.ToLower(c)] = true
	}
	for _, col := range columns {
		if localSet[strings.ToLower(col)] {
			continue
		}
		colType := s.getColumnType(extDB, tableName, col)
		alterSQL := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s NULL", tableName, col, colType)
		_, err := database.DB.Exec(alterSQL)
		if err != nil {
			log.Printf("[DBSync] 自动补列 %s.%s (%s) 失败: %v", tableName, col, colType, err)
		} else {
			log.Printf("[DBSync] 自动补列 %s.%s (%s) 成功", tableName, col, colType)
		}
	}
}

// getPrimaryKey 获取表的主键列名
func (s *DBSyncService) getPrimaryKey(db *sql.DB, tableName string) string {
	var pk string
	db.QueryRow(
		"SELECT COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_KEY = 'PRI' LIMIT 1",
		tableName).Scan(&pk)
	return pk
}

// syncTableGeneric 通用表同步：动态获取列，自动补缺失列，按主键 upsert
func (s *DBSyncService) syncTableGeneric(extDB *sql.DB, tableName string, updateExisting bool) (*SyncTableInfo, error) {
	info := &SyncTableInfo{}

	// 1. 获取源表的列
	srcCols, err := s.getTableColumns(extDB, tableName)
	if err != nil || len(srcCols) == 0 {
		return nil, fmt.Errorf("源表 %s 不存在或无列", tableName)
	}

	// 2. 获取主键
	pk := s.getPrimaryKey(extDB, tableName)
	if pk == "" {
		pk = srcCols[0] // fallback: 用第一列当主键
	}

	// 3. 确保本地表有这些列（自动补列）
	s.ensureLocalColumns(tableName, extDB, srcCols)

	// 4. 构建列名列表
	var selectParts []string
	for _, col := range srcCols {
		selectParts = append(selectParts, fmt.Sprintf("`%s`", col))
	}
	colList := strings.Join(selectParts, ", ")

	// 5. 找到主键在列中的索引
	pkIdx := -1
	for i, col := range srcCols {
		if strings.EqualFold(col, pk) {
			pkIdx = i
			break
		}
	}
	if pkIdx < 0 {
		pkIdx = 0
	}

	// 6. 构建 INSERT 和 UPDATE SQL
	nonPKCols := make([]string, 0, len(srcCols)-1)
	for _, col := range srcCols {
		if !strings.EqualFold(col, pk) {
			nonPKCols = append(nonPKCols, col)
		}
	}

	allColsQuoted := make([]string, len(srcCols))
	for i, c := range srcCols {
		allColsQuoted[i] = fmt.Sprintf("`%s`", c)
	}
	placeholders := make([]string, len(srcCols))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	insertSQL := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)",
		tableName, strings.Join(allColsQuoted, ", "), strings.Join(placeholders, ", "))

	var setParts []string
	for _, col := range nonPKCols {
		setParts = append(setParts, fmt.Sprintf("`%s`=?", col))
	}
	updateSQL := ""
	if len(setParts) > 0 {
		updateSQL = fmt.Sprintf("UPDATE `%s` SET %s WHERE `%s`=?",
			tableName, strings.Join(setParts, ", "), pk)
	}

	checkSQL := fmt.Sprintf("SELECT COUNT(*) FROM `%s` WHERE `%s`=?", tableName, pk)

	// 7. 分批读取和写入（每批 1000 条，防止大表内存溢出或连接超时）
	const batchSize = 1000
	colCount := len(srcCols)
	offset := 0

	for {
		selectSQL := fmt.Sprintf("SELECT %s FROM `%s` ORDER BY `%s` LIMIT %d OFFSET %d",
			colList, tableName, pk, batchSize, offset)

		rows, err := extDB.Query(selectSQL)
		if err != nil {
			return nil, fmt.Errorf("查询源表失败(offset=%d): %v", offset, err)
		}

		batchCount := 0
		for rows.Next() {
			batchCount++
			vals := make([]sql.NullString, colCount)
			ptrs := make([]interface{}, colCount)
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				info.Failed++
				continue
			}
			info.Total++

			pkVal := vals[pkIdx].String

			var exists int
			database.DB.QueryRow(checkSQL, pkVal).Scan(&exists)

			if exists > 0 {
				if updateExisting && updateSQL != "" {
					args := make([]interface{}, 0, len(nonPKCols)+1)
					for _, col := range nonPKCols {
						for j, sc := range srcCols {
							if strings.EqualFold(sc, col) {
								if vals[j].Valid {
									args = append(args, vals[j].String)
								} else {
									args = append(args, nil)
								}
								break
							}
						}
					}
					args = append(args, pkVal)
					if _, err := database.DB.Exec(updateSQL, args...); err != nil {
						info.Failed++
					} else {
						info.Updated++
					}
				} else {
					info.Skipped++
				}
			} else {
				args := make([]interface{}, colCount)
				for i := range vals {
					if vals[i].Valid {
						args[i] = vals[i].String
					} else {
						args[i] = nil
					}
				}
				if _, err := database.DB.Exec(insertSQL, args...); err != nil {
					info.Failed++
				} else {
					info.Inserted++
				}
			}
		}
		rows.Close()

		if batchCount < batchSize {
			break // 最后一批，没有更多数据了
		}
		offset += batchSize
		log.Printf("[DBSync] %s 已处理 %d 条...", tableName, offset)
	}

	return info, nil
}

// TestConnection 测试外部数据库连接，返回各表行数
func (s *DBSyncService) TestConnection(req SyncRequest) (*SyncTestResult, error) {
	extDB, err := s.connectExternal(req)
	if err != nil {
		return &SyncTestResult{Connected: false, Error: err.Error()}, nil
	}
	defer extDB.Close()

	tables := []string{
		"qingka_wangke_huoyuan",
		"qingka_wangke_user",
		"qingka_wangke_fenlei",
		"qingka_wangke_class",
		"qingka_wangke_order",
		"qingka_wangke_dengji",
		"qingka_wangke_config",
		"qingka_wangke_gonggao",
		"qingka_wangke_mijia",
		"qingka_wangke_km",
		"qingka_wangke_pay",
		"qingka_wangke_log",
	}

	counts := make(map[string]int)
	for _, tbl := range tables {
		var cnt int
		err := extDB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tbl)).Scan(&cnt)
		if err != nil {
			counts[tbl] = -1
		} else {
			counts[tbl] = cnt
		}
	}

	return &SyncTestResult{Connected: true, Tables: counts}, nil
}

// Execute 执行完整数据同步（通用动态同步，自动补列）
func (s *DBSyncService) Execute(req SyncRequest) (*SyncResult, error) {
	extDB, err := s.connectExternal(req)
	if err != nil {
		return nil, err
	}
	defer extDB.Close()

	result := &SyncResult{
		SyncTime: time.Now().Format("2006-01-02 15:04:05"),
		Success:  true,
	}

	// 同步顺序：先基础表后业务表
	syncTables := []struct {
		name  string
		label string
	}{
		{"qingka_wangke_dengji", "等级"},
		{"qingka_wangke_huoyuan", "货源"},
		{"qingka_wangke_user", "用户"},
		{"qingka_wangke_fenlei", "分类"},
		{"qingka_wangke_class", "商品"},
		{"qingka_wangke_config", "配置"},
		{"qingka_wangke_gonggao", "公告"},
		{"qingka_wangke_mijia", "密价"},
		{"qingka_wangke_km", "卡密"},
		{"qingka_wangke_order", "订单"},
		{"qingka_wangke_pay", "支付"},
	}

	totalInserted, totalUpdated := 0, 0
	for _, st := range syncTables {
		info, err := s.syncTableGeneric(extDB, st.name, req.UpdateExisting)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s同步失败: %v", st.label, err))
			result.Details = append(result.Details, SyncTableInfo{
				Table: st.name, Label: st.label, Failed: -1,
			})
			log.Printf("[DBSync] %s同步失败: %v", st.label, err)
			continue
		}
		info.Table = st.name
		info.Label = st.label
		result.Details = append(result.Details, *info)
		totalInserted += info.Inserted
		totalUpdated += info.Updated
		log.Printf("[DBSync] %s: 共%d条, 新增%d, 更新%d, 跳过%d",
			st.label, info.Total, info.Inserted, info.Updated, info.Skipped)
	}

	if result.Errors == nil {
		result.Errors = []string{}
	}

	result.Summary = fmt.Sprintf("同步完成，共新增 %d 条、更新 %d 条数据", totalInserted, totalUpdated)
	if len(result.Errors) > 0 {
		result.Success = false
		result.Summary += fmt.Sprintf("，%d 项出错", len(result.Errors))
	}

	return result, nil
}
