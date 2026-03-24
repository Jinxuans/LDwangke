package dbtools

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-api/internal/database"

	"github.com/go-sql-driver/mysql"
)

type SyncRequest struct {
	Host              string `json:"host" binding:"required"`
	Port              int    `json:"port"`
	DBName            string `json:"db_name" binding:"required"`
	User              string `json:"user" binding:"required"`
	Password          string `json:"password"`
	UpdateExisting    bool   `json:"update_existing"`
	ConfirmationToken string `json:"confirmation_token,omitempty"`
}

type SyncResult struct {
	SyncTime string          `json:"sync_time"`
	Success  bool            `json:"success"`
	Details  []SyncTableInfo `json:"details"`
	Errors   []string        `json:"errors"`
	Summary  string          `json:"summary"`
}

type SyncTableInfo struct {
	Table        string `json:"table"`
	Label        string `json:"label"`
	SourceTable  string `json:"source_table,omitempty"`
	SkippedEmpty bool   `json:"skipped_empty"`
	Message      string `json:"message,omitempty"`
	LocalBefore  int    `json:"local_before,omitempty"`
	LocalAfter   int    `json:"local_after,omitempty"`
	Total        int    `json:"total"`
	Inserted     int    `json:"inserted"`
	Updated      int    `json:"updated"`
	Skipped      int    `json:"skipped"`
	Failed       int    `json:"failed"`
}

type SyncTableCheck struct {
	Table               string   `json:"table"`
	Label               string   `json:"label"`
	SourceTable         string   `json:"source_table,omitempty"`
	SourceExists        bool     `json:"source_exists"`
	LocalExists         bool     `json:"local_exists"`
	SourceCount         int      `json:"source_count"`
	LocalCount          int      `json:"local_count"`
	MissingLocalColumns []string `json:"missing_local_columns"`
	Skip                bool     `json:"skip"`
	Ready               bool     `json:"ready"`
	Message             string   `json:"message"`
}

type SyncTestResult struct {
	Connected         bool             `json:"connected"`
	Ready             bool             `json:"ready"`
	Tables            map[string]int   `json:"tables"`
	TableChecks       []SyncTableCheck `json:"table_checks"`
	Warnings          []string         `json:"warnings"`
	Summary           string           `json:"summary"`
	TestedAt          string           `json:"tested_at"`
	ConfirmationToken string           `json:"confirmation_token,omitempty"`
	Error             string           `json:"error,omitempty"`
}

type DBSyncService struct{}

type syncMatchRule struct {
	name    string
	columns []string
}

type syncColumnMeta struct {
	nullable     bool
	defaultValue sql.NullString
	dataType     string
}

type syncReferenceDef struct {
	column    string
	refTable  string
	deferSync bool
}

type syncTableDef struct {
	name         string
	label        string
	aliases      []string
	matchRules   []syncMatchRule
	mergeByMatch bool
	remapRefs    []syncReferenceDef
	columnMap    map[string]string
	copySourcePK bool
}

type syncBatchRow struct {
	sourcePK            string
	targetPK            string
	matchKey            string
	insertWithPK        bool
	insertArgsWithPK    []interface{}
	insertArgsWithoutPK []interface{}
	updateArgs          []interface{}
}

type syncErrorCollector struct {
	limit  int
	errors []string
}

type syncConfirmation struct {
	fingerprint string
	expiresAt   time.Time
}

type syncExecutionContext struct {
	idMaps           map[string]map[string]string
	localIDCache     map[string]map[string]struct{}
	sourceIDCache    map[string]map[string]struct{}
	localPKNameByTB  map[string]string
	sourcePKNameByTB map[string]string
}

type syncTargetIndex struct {
	keyToPK map[string]string
	pkToKey map[string]map[string]struct{}
}

var (
	dbSyncService = &DBSyncService{}

	syncTables = []syncTableDef{
		{
			name:         "qingka_wangke_dengji",
			label:        "等级",
			aliases:      []string{"love_learn_dengji"},
			matchRules:   []syncMatchRule{{name: "name", columns: []string{"name"}}},
			mergeByMatch: false,
		},
		{
			name:         "qingka_wangke_huoyuan",
			label:        "货源",
			aliases:      []string{"love_learn_huoyuan"},
			matchRules:   []syncMatchRule{{name: "pt-url-user", columns: []string{"pt", "url", "user"}}, {name: "name-url", columns: []string{"name", "url"}}},
			mergeByMatch: false,
		},
		{
			name:         "qingka_wangke_user",
			label:        "用户",
			aliases:      []string{"love_learn_user"},
			matchRules:   []syncMatchRule{{name: "user", columns: []string{"user"}}},
			mergeByMatch: false,
			copySourcePK: true,
			remapRefs: []syncReferenceDef{
				{column: "grade", refTable: "qingka_wangke_dengji"},
				{column: "uuid", refTable: "qingka_wangke_user", deferSync: true},
			},
			columnMap: map[string]string{
				"grade": "grade_id",
			},
		},
		{
			name:         "qingka_wangke_fenlei",
			label:        "分类",
			aliases:      []string{"love_learn_fenlei"},
			matchRules:   []syncMatchRule{{name: "name", columns: []string{"name"}}},
			mergeByMatch: false,
			remapRefs:    []syncReferenceDef{{column: "supplier_report_hid", refTable: "qingka_wangke_huoyuan"}},
		},
		{
			name:         "qingka_wangke_class",
			label:        "商品",
			aliases:      []string{"love_learn_class"},
			matchRules:   []syncMatchRule{{name: "docking-noun", columns: []string{"docking", "noun"}}, {name: "name-fenlei", columns: []string{"name", "fenlei"}}},
			mergeByMatch: false,
			remapRefs: []syncReferenceDef{
				{column: "docking", refTable: "qingka_wangke_huoyuan"},
				{column: "fenlei", refTable: "qingka_wangke_fenlei"},
			},
		},
		{
			name:         "qingka_wangke_config",
			label:        "配置",
			aliases:      []string{"love_learn_config"},
			copySourcePK: true,
		},
		{
			name:         "qingka_wangke_gonggao",
			label:        "公告",
			aliases:      []string{"qingka_wangke_notice", "qingka_wangke_homenotice", "love_learn_notice"},
			matchRules:   []syncMatchRule{{name: "title-time", columns: []string{"title", "time"}}, {name: "title-content", columns: []string{"title", "content"}}},
			mergeByMatch: false,
			remapRefs:    []syncReferenceDef{{column: "uid", refTable: "qingka_wangke_user"}},
		},
		{
			name:         "qingka_wangke_mijia",
			label:        "密价",
			aliases:      []string{"love_learn_mijia"},
			matchRules:   []syncMatchRule{{name: "uid-cid", columns: []string{"uid", "cid"}}},
			mergeByMatch: true,
			remapRefs: []syncReferenceDef{
				{column: "uid", refTable: "qingka_wangke_user"},
				{column: "cid", refTable: "qingka_wangke_class"},
			},
		},
		{
			name:         "qingka_wangke_km",
			label:        "卡密",
			aliases:      []string{"qingka_wangke_kami", "love_learn_km"},
			matchRules:   []syncMatchRule{{name: "content", columns: []string{"content"}}},
			mergeByMatch: true,
			remapRefs:    []syncReferenceDef{{column: "uid", refTable: "qingka_wangke_user"}},
		},
		{
			name:    "qingka_wangke_order",
			label:   "订单",
			aliases: []string{"love_learn_order"},
			matchRules: []syncMatchRule{
				{name: "out-trade-no", columns: []string{"out_trade_no"}},
				{name: "school-user-pass-kcid-kcname-addtime", columns: []string{"school", "user", "pass", "kcid", "kcname", "addtime"}},
				{name: "user-kcname-addtime-ptname", columns: []string{"user", "kcname", "addtime", "ptname"}},
			},
			mergeByMatch: true,
			remapRefs: []syncReferenceDef{
				{column: "uid", refTable: "qingka_wangke_user"},
				{column: "cid", refTable: "qingka_wangke_class"},
				{column: "hid", refTable: "qingka_wangke_huoyuan"},
			},
		},
		{
			name:    "qingka_wangke_pay",
			label:   "支付",
			aliases: []string{"love_learn_pay"},
			matchRules: []syncMatchRule{
				{name: "out-trade-no", columns: []string{"out_trade_no"}},
				{name: "trade-no", columns: []string{"trade_no"}},
				{name: "name-money-addtime", columns: []string{"name", "money", "addtime"}},
			},
			mergeByMatch: true,
			remapRefs:    []syncReferenceDef{{column: "uid", refTable: "qingka_wangke_user"}},
		},
	}

	syncConfirmationTTL = 10 * time.Minute

	syncConfirmationsMu sync.Mutex
	syncConfirmations   = map[string]syncConfirmation{}
)

func TestDBSyncConnection(req SyncRequest) (*SyncTestResult, error) {
	return dbSyncService.TestConnection(req)
}

func ExecuteDBSync(req SyncRequest) (*SyncResult, error) {
	return dbSyncService.Execute(req)
}

func newSyncExecutionContext() *syncExecutionContext {
	return &syncExecutionContext{
		idMaps:           make(map[string]map[string]string),
		localIDCache:     make(map[string]map[string]struct{}),
		sourceIDCache:    make(map[string]map[string]struct{}),
		localPKNameByTB:  make(map[string]string),
		sourcePKNameByTB: make(map[string]string),
	}
}

func newSyncErrorCollector(limit int) *syncErrorCollector {
	return &syncErrorCollector{limit: limit, errors: make([]string, 0, limit)}
}

func (c *syncErrorCollector) add(message string) {
	if c == nil || strings.TrimSpace(message) == "" {
		return
	}
	if len(c.errors) >= c.limit {
		return
	}
	c.errors = append(c.errors, message)
}

func (c *syncErrorCollector) summary() string {
	if c == nil || len(c.errors) == 0 {
		return ""
	}
	return "示例错误: " + strings.Join(c.errors, "；")
}

func (ctx *syncExecutionContext) remember(tableName, sourcePK, targetPK string) {
	sourcePK = strings.TrimSpace(sourcePK)
	targetPK = strings.TrimSpace(targetPK)
	if sourcePK == "" || targetPK == "" {
		return
	}
	tableMap, ok := ctx.idMaps[tableName]
	if !ok {
		tableMap = make(map[string]string)
		ctx.idMaps[tableName] = tableMap
	}
	tableMap[sourcePK] = targetPK
}

func (ctx *syncExecutionContext) resolve(tableName, sourcePK string) (string, bool) {
	tableMap, ok := ctx.idMaps[tableName]
	if !ok {
		return "", false
	}
	targetPK, ok := tableMap[strings.TrimSpace(sourcePK)]
	return targetPK, ok
}

func (ctx *syncExecutionContext) rememberLocalID(tableName, targetPK string) {
	targetPK = strings.TrimSpace(targetPK)
	if targetPK == "" {
		return
	}
	tableSet, ok := ctx.localIDCache[tableName]
	if !ok {
		return
	}
	tableSet[targetPK] = struct{}{}
}

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
		_ = db.Close()
		return nil, fmt.Errorf("无法连接到数据库: %v", err)
	}
	_, _ = db.Exec("SET NAMES utf8mb4")
	_, _ = db.Exec("SET CHARACTER SET utf8mb4")
	db.SetMaxOpenConns(5)
	db.SetConnMaxLifetime(2 * time.Minute)
	return db, nil
}

func (s *DBSyncService) getTableColumns(db *sql.DB, tableName string) ([]string, error) {
	rows, err := db.Query(
		"SELECT COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION",
		tableName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []string
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return nil, err
		}
		cols = append(cols, col)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return cols, nil
}

func (s *DBSyncService) getTableColumnMeta(db *sql.DB, tableName string) (map[string]syncColumnMeta, error) {
	rows, err := db.Query(
		`SELECT COLUMN_NAME, IS_NULLABLE, COLUMN_DEFAULT, DATA_TYPE
		 FROM information_schema.COLUMNS
		 WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		 ORDER BY ORDINAL_POSITION`,
		tableName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metas := make(map[string]syncColumnMeta)
	for rows.Next() {
		var (
			columnName   string
			isNullable   string
			defaultValue sql.NullString
			dataType     string
		)
		if err := rows.Scan(&columnName, &isNullable, &defaultValue, &dataType); err != nil {
			return nil, err
		}
		metas[strings.ToLower(columnName)] = syncColumnMeta{
			nullable:     strings.EqualFold(isNullable, "YES"),
			defaultValue: defaultValue,
			dataType:     strings.ToLower(strings.TrimSpace(dataType)),
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return metas, nil
}

func (s *DBSyncService) countTableRows(db *sql.DB, tableName string) (int, error) {
	var cnt int
	err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)).Scan(&cnt)
	return cnt, err
}

func (s *DBSyncService) getPrimaryKey(db *sql.DB, tableName string) string {
	var pk string
	_ = db.QueryRow(
		"SELECT COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_KEY = 'PRI' LIMIT 1",
		tableName,
	).Scan(&pk)
	return pk
}

func (s *DBSyncService) findMissingLocalColumns(table syncTableDef, sourceCols, localCols []string) []string {
	localSet := make(map[string]struct{}, len(localCols))
	for _, col := range localCols {
		localSet[strings.ToLower(col)] = struct{}{}
	}

	missing := make([]string, 0)
	for _, col := range sourceCols {
		targetCol := table.targetColumnName(col)
		if targetCol == "" {
			continue
		}
		if _, ok := localSet[strings.ToLower(targetCol)]; ok {
			continue
		}
		missing = append(missing, targetCol)
	}
	return missing
}

func (t syncTableDef) sourceColumnIsReferenced(col string) bool {
	key := strings.ToLower(strings.TrimSpace(col))
	if key == "" {
		return false
	}
	for _, ref := range t.remapRefs {
		if strings.EqualFold(ref.column, key) {
			return true
		}
	}
	return false
}

func (t syncTableDef) sourceColumnUsedInMatchRules(col string) bool {
	key := strings.ToLower(strings.TrimSpace(col))
	if key == "" {
		return false
	}
	for _, rule := range t.matchRules {
		for _, ruleCol := range rule.columns {
			if strings.EqualFold(ruleCol, key) {
				return true
			}
		}
	}
	return false
}

func (t syncTableDef) filterSyncColumns(sourceCols []string, targetColumnMeta map[string]syncColumnMeta, sourcePK string) ([]string, []string, []string) {
	filteredSourceCols := make([]string, 0, len(sourceCols))
	filteredTargetCols := make([]string, 0, len(sourceCols))
	ignoredSourceCols := make([]string, 0)

	for _, sourceCol := range sourceCols {
		targetCol := t.targetColumnName(sourceCol)
		targetKey := strings.ToLower(strings.TrimSpace(targetCol))
		include := false

		switch {
		case strings.EqualFold(sourceCol, sourcePK):
			include = true
		case t.sourceColumnIsReferenced(sourceCol):
			include = true
		case t.sourceColumnUsedInMatchRules(sourceCol):
			include = true
		case targetKey != "":
			_, include = targetColumnMeta[targetKey]
		}

		if include {
			filteredSourceCols = append(filteredSourceCols, sourceCol)
			filteredTargetCols = append(filteredTargetCols, targetCol)
			continue
		}

		ignoredSourceCols = append(ignoredSourceCols, sourceCol)
	}

	return filteredSourceCols, filteredTargetCols, ignoredSourceCols
}

func (t syncTableDef) sourceCandidates() []string {
	candidates := make([]string, 0, len(t.aliases)+1)
	seen := make(map[string]struct{}, len(t.aliases)+1)
	for _, name := range append([]string{t.name}, t.aliases...) {
		normalized := strings.ToLower(strings.TrimSpace(name))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		candidates = append(candidates, name)
	}
	return candidates
}

func (s *DBSyncService) resolveSourceTable(extDB *sql.DB, table syncTableDef) (string, []string, error) {
	for _, candidate := range table.sourceCandidates() {
		cols, err := s.getTableColumns(extDB, candidate)
		if err != nil {
			continue
		}
		if len(cols) == 0 {
			continue
		}
		return candidate, cols, nil
	}
	return "", nil, fmt.Errorf("源库缺少 %s 的可用表名映射", table.label)
}

func shouldSkipEmptySourceTable(sourceCount int) bool {
	return sourceCount == 0
}

func collectSyncRuleColumns(rules []syncMatchRule) []string {
	seen := make(map[string]struct{})
	cols := make([]string, 0)
	for _, rule := range rules {
		for _, col := range rule.columns {
			key := strings.ToLower(strings.TrimSpace(col))
			if key == "" {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			cols = append(cols, col)
		}
	}
	return cols
}

func (t syncTableDef) hasMatchRules() bool {
	return len(t.matchRules) > 0
}

func (t syncTableDef) targetColumnName(sourceCol string) string {
	if t.columnMap == nil {
		return sourceCol
	}
	if mapped, ok := t.columnMap[strings.ToLower(strings.TrimSpace(sourceCol))]; ok && strings.TrimSpace(mapped) != "" {
		return mapped
	}
	return sourceCol
}

func findSyncColumnIndex(cols []string, target string) int {
	for i, col := range cols {
		if strings.EqualFold(col, target) {
			return i
		}
	}
	return -1
}

func buildSyncValueMap(cols []string, vals []sql.NullString) map[string]sql.NullString {
	values := make(map[string]sql.NullString, len(cols))
	for i, col := range cols {
		values[strings.ToLower(col)] = vals[i]
	}
	return values
}

func syncZeroValueForType(dataType string) interface{} {
	switch strings.ToLower(strings.TrimSpace(dataType)) {
	case "tinyint", "smallint", "mediumint", "int", "integer", "bigint", "decimal", "numeric", "float", "double", "real", "bit", "bool", "boolean":
		return "0"
	case "date":
		return "1970-01-01"
	case "datetime", "timestamp":
		return "1970-01-01 00:00:00"
	case "time":
		return "00:00:00"
	case "year":
		return "1970"
	case "json":
		return "{}"
	default:
		return ""
	}
}

func normalizeSyncWriteValue(value sql.NullString, meta syncColumnMeta) interface{} {
	if value.Valid {
		return value.String
	}
	if meta.nullable {
		return nil
	}
	if meta.defaultValue.Valid {
		return meta.defaultValue.String
	}
	return syncZeroValueForType(meta.dataType)
}

func normalizeSyncValue(col string, value sql.NullString) (string, bool) {
	if !value.Valid {
		return "", false
	}
	normalized := strings.TrimSpace(value.String)
	if normalized == "" {
		return "", false
	}

	lowerCol := strings.ToLower(strings.TrimSpace(col))
	if normalized == "0" {
		switch lowerCol {
		case "uid", "uuid", "cid", "hid", "id", "supplier_report_hid", "docking", "grade":
			return "", false
		}
	}

	return normalized, true
}

func buildSyncMatchKey(rule syncMatchRule, values map[string]sql.NullString) (string, bool) {
	parts := make([]string, 0, len(rule.columns))
	for _, col := range rule.columns {
		value, ok := normalizeSyncValue(col, values[strings.ToLower(col)])
		if !ok {
			return "", false
		}
		parts = append(parts, value)
	}
	if len(parts) == 0 {
		return "", false
	}
	return rule.name + "::" + strings.Join(parts, "\x1f"), true
}

func buildFirstSyncMatchKey(rules []syncMatchRule, values map[string]sql.NullString) (string, bool) {
	for _, rule := range rules {
		key, ok := buildSyncMatchKey(rule, values)
		if ok {
			return key, true
		}
	}
	return "", false
}

func shouldRemapSyncReference(value sql.NullString) bool {
	if !value.Valid {
		return false
	}
	normalized := strings.TrimSpace(value.String)
	return normalized != "" && normalized != "0"
}

func (s *DBSyncService) getLocalPrimaryKeyName(tableName string, ctx *syncExecutionContext) string {
	if pk, ok := ctx.localPKNameByTB[tableName]; ok {
		return pk
	}
	pk := s.getPrimaryKey(database.DB, tableName)
	if pk == "" {
		pk = "id"
	}
	ctx.localPKNameByTB[tableName] = pk
	return pk
}

func (s *DBSyncService) getLocalPrimaryKeySet(tableName string, ctx *syncExecutionContext) (map[string]struct{}, error) {
	if cached, ok := ctx.localIDCache[tableName]; ok {
		return cached, nil
	}

	pk := s.getLocalPrimaryKeyName(tableName, ctx)
	rows, err := database.DB.Query(fmt.Sprintf("SELECT `%s` FROM `%s`", pk, tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := make(map[string]struct{})
	for rows.Next() {
		var value sql.NullString
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		if !value.Valid {
			continue
		}
		normalized := strings.TrimSpace(value.String)
		if normalized == "" {
			continue
		}
		values[normalized] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	ctx.localIDCache[tableName] = values
	return values, nil
}

func (s *DBSyncService) getSourcePrimaryKeyName(extDB *sql.DB, tableName string, ctx *syncExecutionContext) string {
	if pk, ok := ctx.sourcePKNameByTB[tableName]; ok {
		return pk
	}
	pk := s.getPrimaryKey(extDB, tableName)
	if pk == "" {
		pk = "id"
	}
	ctx.sourcePKNameByTB[tableName] = pk
	return pk
}

func (s *DBSyncService) getSourcePrimaryKeySet(extDB *sql.DB, tableName string, ctx *syncExecutionContext) (map[string]struct{}, error) {
	if cached, ok := ctx.sourceIDCache[tableName]; ok {
		return cached, nil
	}

	pk := s.getSourcePrimaryKeyName(extDB, tableName, ctx)
	rows, err := extDB.Query(fmt.Sprintf("SELECT `%s` FROM `%s`", pk, tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := make(map[string]struct{})
	for rows.Next() {
		var value sql.NullString
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		if !value.Valid {
			continue
		}
		normalized := strings.TrimSpace(value.String)
		if normalized == "" {
			continue
		}
		values[normalized] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	ctx.sourceIDCache[tableName] = values
	return values, nil
}

func (s *DBSyncService) resolveSyncReferenceTarget(extDB *sql.DB, ctx *syncExecutionContext, refTable, sourceValue string) (string, bool, string, error) {
	if targetValue, ok := ctx.resolve(refTable, sourceValue); ok {
		return targetValue, true, "", nil
	}

	localIDs, err := s.getLocalPrimaryKeySet(refTable, ctx)
	if err != nil {
		return "", false, "", err
	}
	if _, ok := localIDs[sourceValue]; ok {
		ctx.remember(refTable, sourceValue, sourceValue)
		return sourceValue, true, "", nil
	}

	sourceIDs, err := s.getSourcePrimaryKeySet(extDB, refTable, ctx)
	if err != nil {
		return "", false, "", err
	}
	if _, ok := sourceIDs[sourceValue]; ok {
		return sourceValue, false, fmt.Sprintf("未找到 %s 映射，源库存在该记录但未导入或被重映射", refTable), nil
	}

	return sourceValue, false, fmt.Sprintf("源库不存在 %s=%s 的关联记录", refTable, sourceValue), nil
}

func applySyncReferenceRemaps(values map[string]sql.NullString, refs []syncReferenceDef, ctx *syncExecutionContext, resolver func(refTable, sourceValue string) (string, bool, string, error)) ([]string, error) {
	warnings := make([]string, 0)
	for _, ref := range refs {
		if ref.deferSync {
			continue
		}
		key := strings.ToLower(ref.column)
		current := values[key]
		if !shouldRemapSyncReference(current) {
			continue
		}
		sourceValue := strings.TrimSpace(current.String)
		targetValue, resolved, reason, err := resolver(ref.refTable, sourceValue)
		if err != nil {
			return nil, fmt.Errorf("字段 %s %v", ref.column, err)
		}
		if !resolved {
			if reason == "" {
				reason = fmt.Sprintf("未找到 %s 映射", ref.refTable)
			}
			warnings = append(warnings, fmt.Sprintf("字段 %s 保留原值 %s（%s）", ref.column, sourceValue, reason))
		}
		values[key] = sql.NullString{String: targetValue, Valid: true}
	}
	return warnings, nil
}

func newSyncTargetIndex() *syncTargetIndex {
	return &syncTargetIndex{
		keyToPK: make(map[string]string),
		pkToKey: make(map[string]map[string]struct{}),
	}
}

func (idx *syncTargetIndex) remember(targetPK, matchKey string) {
	targetPK = strings.TrimSpace(targetPK)
	matchKey = strings.TrimSpace(matchKey)
	if targetPK == "" || matchKey == "" {
		return
	}
	if _, exists := idx.keyToPK[matchKey]; !exists {
		idx.keyToPK[matchKey] = targetPK
	}
	keySet, ok := idx.pkToKey[targetPK]
	if !ok {
		keySet = make(map[string]struct{})
		idx.pkToKey[targetPK] = keySet
	}
	keySet[matchKey] = struct{}{}
}

func (idx *syncTargetIndex) resolve(matchKey string) string {
	return idx.keyToPK[strings.TrimSpace(matchKey)]
}

func (idx *syncTargetIndex) pkHasMatchKey(targetPK, matchKey string) bool {
	keySet, ok := idx.pkToKey[strings.TrimSpace(targetPK)]
	if !ok {
		return false
	}
	_, ok = keySet[strings.TrimSpace(matchKey)]
	return ok
}

func (s *DBSyncService) buildTargetMatchIndex(tableName, targetPK string, rules []syncMatchRule) (*syncTargetIndex, error) {
	if len(rules) == 0 {
		return newSyncTargetIndex(), nil
	}

	selectedCols := append([]string{targetPK}, collectSyncRuleColumns(rules)...)
	selectParts := make([]string, len(selectedCols))
	for i, col := range selectedCols {
		selectParts[i] = fmt.Sprintf("`%s`", col)
	}

	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT %s FROM `%s`", strings.Join(selectParts, ", "), tableName),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	index := newSyncTargetIndex()
	for rows.Next() {
		vals := make([]sql.NullString, len(selectedCols))
		ptrs := make([]interface{}, len(selectedCols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}

		targetPKValue, ok := normalizeSyncValue(targetPK, vals[0])
		if !ok {
			continue
		}
		valueMap := buildSyncValueMap(selectedCols[1:], vals[1:])
		for _, rule := range rules {
			key, ok := buildSyncMatchKey(rule, valueMap)
			if !ok {
				continue
			}
			index.remember(targetPKValue, key)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return index, nil
}

func (s *DBSyncService) buildPrecheck(extDB *sql.DB, req SyncRequest, issueToken bool) (*SyncTestResult, error) {
	result := &SyncTestResult{
		Connected:   true,
		Ready:       true,
		Tables:      make(map[string]int, len(syncTables)),
		TableChecks: make([]SyncTableCheck, 0, len(syncTables)),
		Warnings:    []string{},
		TestedAt:    time.Now().Format("2006-01-02 15:04:05"),
	}

	if req.UpdateExisting {
		result.Warnings = append(result.Warnings, "已开启覆盖更新，执行导入时会覆盖当前系统中已匹配到的记录。")
	}

	for _, table := range syncTables {
		check := SyncTableCheck{
			Table:       table.name,
			Label:       table.label,
			SourceCount: -1,
			LocalCount:  -1,
		}

		sourceTable, sourceCols, sourceErr := s.resolveSourceTable(extDB, table)
		switch {
		case sourceErr != nil:
			check.Message = sourceErr.Error()
			result.Tables[table.name] = -1
			result.Ready = false
		default:
			check.SourceTable = sourceTable
			check.SourceExists = true
			if count, err := s.countTableRows(extDB, sourceTable); err == nil {
				check.SourceCount = count
				result.Tables[table.name] = count
			} else {
				check.Message = "源库表计数失败"
				result.Tables[table.name] = -1
				result.Ready = false
			}
			if sourceTable != table.name {
				check.Message = appendSyncMessage(check.Message, fmt.Sprintf("已命中旧表 %s", sourceTable))
			}
		}

		if shouldSkipEmptySourceTable(check.SourceCount) {
			check.Skip = true
			check.Ready = true
			check.Message = appendSyncMessage(check.Message, "源表为空，将跳过导入")
			result.TableChecks = append(result.TableChecks, check)
			continue
		}

		localCols, localErr := s.getTableColumns(database.DB, table.name)
		switch {
		case localErr != nil:
			check.Message = appendSyncMessage(check.Message, "当前系统表不可访问")
			result.Ready = false
		case len(localCols) == 0:
			check.Message = appendSyncMessage(check.Message, "当前系统缺少该表")
			result.Ready = false
		default:
			check.LocalExists = true
			if count, err := s.countTableRows(database.DB, table.name); err == nil {
				check.LocalCount = count
			}
			targetColumnMeta, err := s.getTableColumnMeta(database.DB, table.name)
			if err != nil {
				check.Message = appendSyncMessage(check.Message, "读取当前系统字段信息失败")
				result.Ready = false
				break
			}
			sourcePK := s.getPrimaryKey(extDB, sourceTable)
			if sourcePK == "" && len(sourceCols) > 0 {
				sourcePK = sourceCols[0]
			}
			filteredSourceCols, _, ignoredSourceCols := table.filterSyncColumns(sourceCols, targetColumnMeta, sourcePK)
			missing := s.findMissingLocalColumns(table, filteredSourceCols, localCols)
			if len(missing) > 0 {
				check.MissingLocalColumns = missing
				check.Message = appendSyncMessage(check.Message, fmt.Sprintf("当前系统缺少 %d 个字段", len(missing)))
				result.Ready = false
			}
			if len(ignoredSourceCols) > 0 {
				check.Message = appendSyncMessage(check.Message, fmt.Sprintf("已忽略 %d 个源库冗余字段", len(ignoredSourceCols)))
			}
		}

		if check.Message == "" {
			check.Ready = true
			check.Message = "结构通过，可导入"
		}
		result.TableChecks = append(result.TableChecks, check)
	}

	if result.Ready {
		result.Summary = "预检查通过，可以开始导入。确认令牌 10 分钟内有效。"
		if issueToken {
			result.ConfirmationToken = issueSyncConfirmation(req)
		}
	} else {
		result.Summary = "预检查未通过。请先备份数据库，并补齐当前系统缺失的表或字段后再导入。"
		result.Warnings = append(result.Warnings, "导入已被阻止，因为当前系统结构与源库不一致。")
	}

	return result, nil
}

func appendSyncMessage(base, message string) string {
	if base == "" {
		return message
	}
	return base + "；" + message
}

func issueSyncConfirmation(req SyncRequest) string {
	token := newSyncConfirmationToken()
	record := syncConfirmation{
		fingerprint: fingerprintSyncRequest(req),
		expiresAt:   time.Now().Add(syncConfirmationTTL),
	}

	syncConfirmationsMu.Lock()
	defer syncConfirmationsMu.Unlock()

	pruneExpiredSyncConfirmationsLocked(time.Now())
	syncConfirmations[token] = record
	return token
}

func validateSyncConfirmation(req SyncRequest, consume bool) error {
	token := strings.TrimSpace(req.ConfirmationToken)
	if token == "" {
		return fmt.Errorf("请先完成预检查，再执行导入")
	}

	syncConfirmationsMu.Lock()
	defer syncConfirmationsMu.Unlock()

	now := time.Now()
	pruneExpiredSyncConfirmationsLocked(now)

	record, ok := syncConfirmations[token]
	if !ok {
		return fmt.Errorf("预检查确认已失效，请重新测试连接")
	}
	if now.After(record.expiresAt) {
		delete(syncConfirmations, token)
		return fmt.Errorf("预检查确认已过期，请重新测试连接")
	}
	if record.fingerprint != fingerprintSyncRequest(req) {
		return fmt.Errorf("数据库连接参数已变更，请重新测试连接")
	}

	if consume {
		delete(syncConfirmations, token)
	}
	return nil
}

func pruneExpiredSyncConfirmationsLocked(now time.Time) {
	for token, record := range syncConfirmations {
		if now.After(record.expiresAt) {
			delete(syncConfirmations, token)
		}
	}
}

func fingerprintSyncRequest(req SyncRequest) string {
	port := req.Port
	if port == 0 {
		port = 3306
	}
	payload := strings.Join([]string{
		strings.ToLower(strings.TrimSpace(req.Host)),
		fmt.Sprintf("%d", port),
		strings.ToLower(strings.TrimSpace(req.DBName)),
		strings.ToLower(strings.TrimSpace(req.User)),
		req.Password,
		fmt.Sprintf("%t", req.UpdateExisting),
	}, "|")
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func newSyncConfirmationToken() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err == nil {
		return hex.EncodeToString(buf)
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return hex.EncodeToString(sum[:16])
}

func (s *DBSyncService) getExistingKeys(tableName, pk string, keys []string) (map[string]struct{}, error) {
	if len(keys) == 0 {
		return map[string]struct{}{}, nil
	}

	placeholders := make([]string, len(keys))
	args := make([]interface{}, len(keys))
	for i, key := range keys {
		placeholders[i] = "?"
		args[i] = key
	}

	query := fmt.Sprintf(
		"SELECT `%s` FROM `%s` WHERE `%s` IN (%s)",
		pk,
		tableName,
		pk,
		strings.Join(placeholders, ", "),
	)
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	existing := make(map[string]struct{}, len(keys))
	for rows.Next() {
		var value sql.NullString
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		existing[value.String] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *DBSyncService) prepareSyncInsertStmt(tableName string, cols []string) (*sql.Stmt, error) {
	quotedCols := make([]string, len(cols))
	placeholders := make([]string, len(cols))
	for i, col := range cols {
		quotedCols[i] = fmt.Sprintf("`%s`", col)
		placeholders[i] = "?"
	}
	insertSQL := fmt.Sprintf(
		"INSERT INTO `%s` (%s) VALUES (%s)",
		tableName,
		strings.Join(quotedCols, ", "),
		strings.Join(placeholders, ", "),
	)
	return database.DB.Prepare(insertSQL)
}

func isSyncDuplicatePrimaryError(err error) bool {
	if err == nil {
		return false
	}
	mysqlErr, ok := err.(*mysql.MySQLError)
	if !ok {
		return false
	}
	return mysqlErr.Number == 1062 && strings.Contains(strings.ToLower(mysqlErr.Message), "primary")
}

func (s *DBSyncService) syncTableGeneric(extDB *sql.DB, table syncTableDef, sourceTableName string, ctx *syncExecutionContext, updateExisting bool) (*SyncTableInfo, error) {
	info := &SyncTableInfo{}
	errorSamples := newSyncErrorCollector(3)

	srcCols, err := s.getTableColumns(extDB, sourceTableName)
	if err != nil || len(srcCols) == 0 {
		return nil, fmt.Errorf("源表 %s 不存在或无列", sourceTableName)
	}

	pk := s.getPrimaryKey(extDB, sourceTableName)
	if pk == "" {
		pk = srcCols[0]
	}
	targetPK := s.getPrimaryKey(database.DB, table.name)
	if targetPK == "" {
		targetPK = pk
	}
	targetColumnMeta, err := s.getTableColumnMeta(database.DB, table.name)
	if err != nil {
		return nil, fmt.Errorf("读取目标表字段元数据失败: %v", err)
	}

	filteredSrcCols, filteredTargetCols, ignoredSourceCols := table.filterSyncColumns(srcCols, targetColumnMeta, pk)
	if len(filteredSrcCols) == 0 {
		return nil, fmt.Errorf("源表 %s 无可同步字段", sourceTableName)
	}
	if len(ignoredSourceCols) > 0 {
		info.Message = appendSyncMessage(info.Message, fmt.Sprintf("已忽略 %d 个源库冗余字段", len(ignoredSourceCols)))
	}

	srcCols = filteredSrcCols
	targetCols := filteredTargetCols

	pkIdx := 0
	for i, col := range srcCols {
		if strings.EqualFold(col, pk) {
			pkIdx = i
			break
		}
	}

	selectParts := make([]string, 0, len(srcCols))
	for _, col := range srcCols {
		selectParts = append(selectParts, fmt.Sprintf("`%s`", col))
	}
	colList := strings.Join(selectParts, ", ")

	insertColsWithPK := append([]string{}, targetCols...)
	insertIndexesWithPK := make([]int, len(srcCols))
	for i := range srcCols {
		insertIndexesWithPK[i] = i
	}
	insertColsWithoutPK := make([]string, 0, len(srcCols))
	insertIndexesWithoutPK := make([]int, 0, len(srcCols))
	for i, col := range targetCols {
		if strings.EqualFold(col, targetPK) {
			continue
		}
		insertColsWithoutPK = append(insertColsWithoutPK, col)
		insertIndexesWithoutPK = append(insertIndexesWithoutPK, i)
	}

	nonTargetPKCols := make([]string, 0, len(srcCols)-1)
	nonTargetPKIndexes := make([]int, 0, len(srcCols)-1)
	for i, col := range targetCols {
		if !strings.EqualFold(col, targetPK) {
			nonTargetPKCols = append(nonTargetPKCols, col)
			nonTargetPKIndexes = append(nonTargetPKIndexes, i)
		}
	}

	insertStmtWithPK, err := s.prepareSyncInsertStmt(table.name, insertColsWithPK)
	if err != nil {
		return nil, err
	}
	defer insertStmtWithPK.Close()

	insertStmtWithoutPK := insertStmtWithPK
	if !table.copySourcePK {
		insertStmtWithoutPK, err = s.prepareSyncInsertStmt(table.name, insertColsWithoutPK)
		if err != nil {
			return nil, err
		}
		defer insertStmtWithoutPK.Close()
	}

	updateSQL := ""
	var updateStmt *sql.Stmt
	if len(nonTargetPKCols) > 0 {
		setParts := make([]string, 0, len(nonTargetPKCols))
		for _, col := range nonTargetPKCols {
			setParts = append(setParts, fmt.Sprintf("`%s`=?", col))
		}
		updateSQL = fmt.Sprintf(
			"UPDATE `%s` SET %s WHERE `%s`=?",
			table.name,
			strings.Join(setParts, ", "),
			targetPK,
		)
		if updateExisting || table.copySourcePK {
			updateStmt, err = database.DB.Prepare(updateSQL)
			if err != nil {
				return nil, err
			}
			defer updateStmt.Close()
		}
	}

	matchIndex := newSyncTargetIndex()
	if table.hasMatchRules() {
		matchIndex, err = s.buildTargetMatchIndex(table.name, targetPK, table.matchRules)
		if err != nil {
			return nil, fmt.Errorf("建立目标表匹配索引失败: %v", err)
		}
	}

	const batchSize = 1000
	lastPK := ""
	hasCursor := false

	for {
		selectSQL := fmt.Sprintf(
			"SELECT %s FROM `%s` %s ORDER BY `%s` LIMIT %d",
			colList,
			sourceTableName,
			buildCursorClause(pk, hasCursor),
			pk,
			batchSize,
		)

		var rows *sql.Rows
		if hasCursor {
			rows, err = extDB.Query(selectSQL, lastPK)
		} else {
			rows, err = extDB.Query(selectSQL)
		}
		if err != nil {
			return nil, fmt.Errorf("查询源表失败: %v", err)
		}

		batchRows := make([]syncBatchRow, 0, batchSize)
		batchKeys := make([]string, 0, batchSize)
		for rows.Next() {
			vals := make([]sql.NullString, len(srcCols))
			ptrs := make([]interface{}, len(srcCols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				info.Failed++
				errorSamples.add(fmt.Sprintf("读取源行失败: %v", err))
				continue
			}

			if !vals[pkIdx].Valid {
				info.Failed++
				errorSamples.add("源表主键为空")
				continue
			}

			sourcePKValue := strings.TrimSpace(vals[pkIdx].String)
			valueMap := buildSyncValueMap(srcCols, vals)
			warnings, err := applySyncReferenceRemaps(valueMap, table.remapRefs, ctx, func(refTable, sourceValue string) (string, bool, string, error) {
				return s.resolveSyncReferenceTarget(extDB, ctx, refTable, sourceValue)
			})
			if err != nil {
				info.Failed++
				errorSamples.add(fmt.Sprintf("source pk=%s 字段重写失败: %v", sourcePKValue, err))
				log.Printf("[DBSync] %s(%s): source pk=%s remap failed: %v", table.label, sourceTableName, sourcePKValue, err)
				continue
			}
			for _, warning := range warnings {
				log.Printf("[DBSync] %s(%s): source pk=%s remap fallback: %s", table.label, sourceTableName, sourcePKValue, warning)
			}

			insertArgsWithPK := make([]interface{}, 0, len(insertIndexesWithPK))
			for _, idx := range insertIndexesWithPK {
				sourceColumnName := strings.ToLower(srcCols[idx])
				targetColumnName := strings.ToLower(targetCols[idx])
				insertArgsWithPK = append(insertArgsWithPK, normalizeSyncWriteValue(valueMap[sourceColumnName], targetColumnMeta[targetColumnName]))
			}

			insertArgsWithoutPK := make([]interface{}, 0, len(insertIndexesWithoutPK))
			for _, idx := range insertIndexesWithoutPK {
				sourceColumnName := strings.ToLower(srcCols[idx])
				targetColumnName := strings.ToLower(targetCols[idx])
				insertArgsWithoutPK = append(insertArgsWithoutPK, normalizeSyncWriteValue(valueMap[sourceColumnName], targetColumnMeta[targetColumnName]))
			}

			updateArgs := make([]interface{}, 0, len(nonTargetPKCols)+1)
			for _, idx := range nonTargetPKIndexes {
				sourceColumnName := strings.ToLower(srcCols[idx])
				targetColumnName := strings.ToLower(targetCols[idx])
				updateArgs = append(updateArgs, normalizeSyncWriteValue(valueMap[sourceColumnName], targetColumnMeta[targetColumnName]))
			}

			row := syncBatchRow{
				sourcePK:            sourcePKValue,
				insertArgsWithPK:    insertArgsWithPK,
				insertArgsWithoutPK: insertArgsWithoutPK,
				updateArgs:          updateArgs,
			}
			if table.copySourcePK {
				row.insertWithPK = true
			}
			if table.hasMatchRules() {
				if matchKey, ok := buildFirstSyncMatchKey(table.matchRules, valueMap); ok {
					row.matchKey = matchKey
				}
			}
			batchKeys = append(batchKeys, sourcePKValue)
			batchRows = append(batchRows, row)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("读取源表失败: %v", err)
		}
		if len(batchRows) == 0 {
			break
		}

		occupiedTargetPKs, err := s.getExistingKeys(table.name, targetPK, batchKeys)
		if err != nil {
			return nil, fmt.Errorf("读取目标表主键失败: %v", err)
		}

		for i := range batchRows {
			row := &batchRows[i]
			info.Total++
			_, samePKExists := occupiedTargetPKs[row.sourcePK]
			targetPKValue := ""
			exists := false
			if samePKExists {
				if row.matchKey == "" || !table.hasMatchRules() || matchIndex.pkHasMatchKey(row.sourcePK, row.matchKey) {
					targetPKValue = row.sourcePK
					exists = true
				}
			}
			if !exists && table.mergeByMatch && row.matchKey != "" && table.hasMatchRules() {
				matchedPK := matchIndex.resolve(row.matchKey)
				if matchedPK != "" {
					targetPKValue = matchedPK
					exists = true
				}
			}
			if !exists && !table.copySourcePK {
				row.insertWithPK = !samePKExists
			}
			if exists {
				if updateExisting && updateStmt != nil && updateSQL != "" {
					updateArgs := append(append([]interface{}{}, row.updateArgs...), targetPKValue)
					if _, err := updateStmt.Exec(updateArgs...); err != nil {
						info.Failed++
						errorSamples.add(fmt.Sprintf("source pk=%s 更新目标 pk=%s 失败: %v", row.sourcePK, targetPKValue, err))
					} else {
						info.Updated++
						ctx.remember(table.name, row.sourcePK, targetPKValue)
						occupiedTargetPKs[targetPKValue] = struct{}{}
					}
				} else {
					info.Skipped++
					ctx.remember(table.name, row.sourcePK, targetPKValue)
					ctx.rememberLocalID(table.name, targetPKValue)
					occupiedTargetPKs[targetPKValue] = struct{}{}
				}
				continue
			}

			insertStmt := insertStmtWithoutPK
			insertArgs := row.insertArgsWithoutPK
			if row.insertWithPK {
				insertStmt = insertStmtWithPK
				insertArgs = row.insertArgsWithPK
			}
			insertResult, err := insertStmt.Exec(insertArgs...)
			owasOverwrite := false
			if err != nil {
				if row.insertWithPK && isSyncDuplicatePrimaryError(err) {
					if table.copySourcePK && updateStmt != nil {
						// copySourcePK 模式：PK 冲突时用源数据覆盖本地记录
						updateArgs := append(append([]interface{}{}, row.updateArgs...), row.sourcePK)
						_, err = updateStmt.Exec(updateArgs...)
						if err == nil {
							owasOverwrite = true
							insertResult = nil
						}
					} else if !table.copySourcePK {
						insertResult, err = insertStmtWithoutPK.Exec(row.insertArgsWithoutPK...)
						if err == nil {
							row.insertWithPK = false
						}
					}
				}
			}
			if err != nil {
				info.Failed++
				errorSamples.add(fmt.Sprintf("source pk=%s 插入失败: %v", row.sourcePK, err))
			} else {
				if owasOverwrite {
					info.Updated++
				} else {
					info.Inserted++
				}
				targetPKValue = row.sourcePK
				if !row.insertWithPK && insertResult != nil {
					if insertID, idErr := insertResult.LastInsertId(); idErr == nil {
						targetPKValue = strconv.FormatInt(insertID, 10)
					}
				}
				ctx.remember(table.name, row.sourcePK, targetPKValue)
				ctx.rememberLocalID(table.name, targetPKValue)
				occupiedTargetPKs[targetPKValue] = struct{}{}
				if row.matchKey != "" {
					matchIndex.remember(targetPKValue, row.matchKey)
				}
			}
		}

		lastPK = batchRows[len(batchRows)-1].sourcePK
		hasCursor = true
		if len(batchRows) < batchSize {
			break
		}
		log.Printf("[DBSync] %s -> %s 已处理 %d 条...", sourceTableName, table.name, info.Total)
	}

	info.Message = appendSyncMessage(info.Message, errorSamples.summary())

	return info, nil
}

func buildCursorClause(pk string, hasCursor bool) string {
	if !hasCursor {
		return ""
	}
	return fmt.Sprintf("WHERE `%s` > ?", pk)
}

func (s *DBSyncService) syncDeferredUserRefs(extDB *sql.DB, sourceTableName string, ctx *syncExecutionContext) error {
	sourceUIDIdx := -1
	sourceUUIDIdx := -1

	srcCols, err := s.getTableColumns(extDB, sourceTableName)
	if err != nil {
		return err
	}
	for i, col := range srcCols {
		switch {
		case strings.EqualFold(col, "uid"):
			sourceUIDIdx = i
		case strings.EqualFold(col, "uuid"):
			sourceUUIDIdx = i
		}
	}
	if sourceUIDIdx < 0 || sourceUUIDIdx < 0 {
		return nil
	}

	rows, err := extDB.Query("SELECT `uid`, `uuid` FROM `" + sourceTableName + "`")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var oldUID sql.NullString
		var oldUUID sql.NullString
		if err := rows.Scan(&oldUID, &oldUUID); err != nil {
			return err
		}
		if !shouldRemapSyncReference(oldUID) || !shouldRemapSyncReference(oldUUID) {
			continue
		}
		targetUID, ok := ctx.resolve("qingka_wangke_user", oldUID.String)
		if !ok {
			continue
		}
		targetUUID, ok := ctx.resolve("qingka_wangke_user", oldUUID.String)
		if !ok {
			continue
		}
		if _, err := database.DB.Exec("UPDATE `qingka_wangke_user` SET `uuid`=? WHERE `uid`=?", targetUUID, targetUID); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (s *DBSyncService) TestConnection(req SyncRequest) (*SyncTestResult, error) {
	extDB, err := s.connectExternal(req)
	if err != nil {
		return &SyncTestResult{
			Connected: false,
			Ready:     false,
			Error:     err.Error(),
			Warnings:  []string{},
			Tables:    map[string]int{},
		}, nil
	}
	defer extDB.Close()

	return s.buildPrecheck(extDB, req, true)
}

func (s *DBSyncService) Execute(req SyncRequest) (*SyncResult, error) {
	if err := validateSyncConfirmation(req, false); err != nil {
		return nil, err
	}

	extDB, err := s.connectExternal(req)
	if err != nil {
		return nil, err
	}
	defer extDB.Close()

	precheck, err := s.buildPrecheck(extDB, req, false)
	if err != nil {
		return nil, err
	}
	if !precheck.Ready {
		return nil, errors.New(precheck.Summary)
	}
	if err := validateSyncConfirmation(req, true); err != nil {
		return nil, err
	}

	result := &SyncResult{
		SyncTime: time.Now().Format("2006-01-02 15:04:05"),
		Success:  true,
		Errors:   []string{},
	}
	ctx := newSyncExecutionContext()

	totalInserted, totalUpdated, totalSkipped, totalFailed, skippedEmptyTables := 0, 0, 0, 0, 0
	for _, table := range syncTables {
		sourceTable, _, err := s.resolveSourceTable(extDB, table)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s同步失败: %v", table.label, err))
			result.Details = append(result.Details, SyncTableInfo{
				Table:  table.name,
				Label:  table.label,
				Failed: -1,
			})
			log.Printf("[DBSync] %s同步失败: %v", table.label, err)
			continue
		}

		sourceCount, err := s.countTableRows(extDB, sourceTable)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s同步失败: 源库表计数失败: %v", table.label, err))
			result.Details = append(result.Details, SyncTableInfo{
				Table:       table.name,
				Label:       table.label,
				SourceTable: sourceTable,
				Failed:      -1,
			})
			log.Printf("[DBSync] %s同步失败: 源库表计数失败: %v", table.label, err)
			continue
		}
		if shouldSkipEmptySourceTable(sourceCount) {
			localBefore, _ := s.countTableRows(database.DB, table.name)
			result.Details = append(result.Details, SyncTableInfo{
				Table:        table.name,
				Label:        table.label,
				SourceTable:  sourceTable,
				SkippedEmpty: true,
				Message:      "源表为空，已跳过导入",
				LocalBefore:  localBefore,
				LocalAfter:   localBefore,
				Total:        0,
			})
			skippedEmptyTables++
			log.Printf("[DBSync] %s(%s -> %s): 源表为空，跳过导入", table.label, sourceTable, table.name)
			continue
		}

		localBefore, _ := s.countTableRows(database.DB, table.name)
		info, err := s.syncTableGeneric(extDB, table, sourceTable, ctx, req.UpdateExisting)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s同步失败: %v", table.label, err))
			result.Details = append(result.Details, SyncTableInfo{
				Table:       table.name,
				Label:       table.label,
				SourceTable: sourceTable,
				LocalBefore: localBefore,
				Failed:      -1,
			})
			log.Printf("[DBSync] %s同步失败: %v", table.label, err)
			continue
		}
		localAfter, _ := s.countTableRows(database.DB, table.name)

		info.Table = table.name
		info.Label = table.label
		info.SourceTable = sourceTable
		info.LocalBefore = localBefore
		info.LocalAfter = localAfter
		info.Message = appendSyncMessage(info.Message, fmt.Sprintf("本地记录 %d -> %d", localBefore, localAfter))
		result.Details = append(result.Details, *info)
		totalInserted += info.Inserted
		totalUpdated += info.Updated
		totalSkipped += info.Skipped
		totalFailed += info.Failed
		if info.Failed > 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("%s导入失败 %d 条: %s", table.label, info.Failed, strings.TrimSpace(info.Message)))
		}
		log.Printf(
			"[DBSync] %s(%s -> %s): 共%d条, 本地%d->%d, 新增%d, 更新%d, 跳过%d, 失败%d, 说明=%s",
			table.label,
			sourceTable,
			table.name,
			info.Total,
			info.LocalBefore,
			info.LocalAfter,
			info.Inserted,
			info.Updated,
			info.Skipped,
			info.Failed,
			strings.TrimSpace(info.Message),
		)

		if table.name == "qingka_wangke_user" {
			if err := s.syncDeferredUserRefs(extDB, sourceTable, ctx); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("%s关联修复失败: %v", table.label, err))
				log.Printf("[DBSync] %s关联修复失败: %v", table.label, err)
			}
			if err := s.syncImportedInviteGradeRefs(); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("%s邀请等级修复失败: %v", table.label, err))
				log.Printf("[DBSync] %s邀请等级修复失败: %v", table.label, err)
			}
		}
	}

	result.Summary = fmt.Sprintf(
		"导入完成，共新增 %d 条、更新 %d 条、跳过 %d 条、失败 %d 条数据",
		totalInserted,
		totalUpdated,
		totalSkipped,
		totalFailed,
	)
	if skippedEmptyTables > 0 {
		result.Summary += fmt.Sprintf("，另有 %d 张空表已跳过", skippedEmptyTables)
	}
	if len(result.Errors) > 0 {
		result.Success = false
		result.Summary += fmt.Sprintf("，%d 项出错", len(result.Errors))
	}
	return result, nil
}

func (s *DBSyncService) syncImportedInviteGradeRefs() error {
	// 步骤1：用 addprice 数值匹配等级费率，回填 grade_id（解决旧系统 grade 字段未使用的问题）
	_, err := database.DB.Exec(`
UPDATE qingka_wangke_user u
JOIN qingka_wangke_dengji g
  ON CAST(g.rate AS DECIMAL(10,2)) = CAST(u.addprice AS DECIMAL(10,2))
  AND g.status = '1'
SET u.grade_id = g.id
WHERE (u.grade_id IS NULL OR u.grade_id = 0)
  AND u.addprice IS NOT NULL
  AND CAST(u.addprice AS DECIMAL(10,2)) NOT IN (0, 1)`)
	if err != nil {
		return err
	}
	// 步骤2：用 yqprice 数值匹配等级费率，回填 invite_grade_id（CAST 解决精度不一致问题）
	_, err = database.DB.Exec(`
UPDATE qingka_wangke_user u
JOIN qingka_wangke_dengji g
  ON CAST(g.rate AS DECIMAL(10,2)) = CAST(u.yqprice AS DECIMAL(10,2))
  AND g.status = '1'
SET u.invite_grade_id = g.id
WHERE (u.invite_grade_id IS NULL OR u.invite_grade_id = 0)
  AND COALESCE(u.yqprice, '') NOT IN ('', '0')`)
	if err != nil {
		return err
	}
	// 步骤3：invite_grade_id 仍为空时，退化使用 grade_id
	_, err = database.DB.Exec(`
UPDATE qingka_wangke_user
SET invite_grade_id = grade_id
WHERE (invite_grade_id IS NULL OR invite_grade_id = 0)
  AND grade_id IS NOT NULL
  AND grade_id > 0`)
	return err
}
