package database

import (
	"bufio"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"go-api/internal/config"
	obslogger "go-api/internal/observability/logger"

	"github.com/go-sql-driver/mysql"
)

const migrationTableName = "qingka_schema_migration"
const autoMigrationBaselineVersion = 46

var migrationFilePattern = regexp.MustCompile(`^\d+[A-Za-z0-9_]*\.sql$`)

type autoBaselineCheckpoint struct {
	table           string
	baselineVersion int
}

var autoBaselineCheckpoints = []autoBaselineCheckpoint{
	{table: "qingka_dynamic_module", baselineVersion: 3},
	{table: "qingka_platform_config", baselineVersion: 16},
	{table: "qingka_wangke_sync_config", baselineVersion: 18},
	{table: "qingka_tenant", baselineVersion: 28},
	{table: "qingka_ext_menu", baselineVersion: 40},
	{table: "qingka_wangke_yfdk_projects", baselineVersion: 41},
}

func runConfiguredMigrations(db *sql.DB) error {
	if config.Global != nil && !config.Global.AutoMigrateEnabled() {
		obslogger.L().Info("Migration 已通过配置关闭自动迁移")
		return nil
	}

	if err := runSQLMigrations(db); err != nil {
		return err
	}

	runLegacySchemaPatches(db)
	return nil
}

func runSQLMigrations(db *sql.DB) error {
	migrationDir, err := resolveMigrationDir()
	if err != nil {
		obslogger.L().Warn("Migration 未找到迁移目录，跳过 SQL 自动迁移", "error", err)
		return nil
	}

	if err := ensureMigrationTable(db); err != nil {
		return fmt.Errorf("初始化迁移记录表失败: %w", err)
	}

	applied, err := loadAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("读取迁移记录失败: %w", err)
	}

	files, err := listMigrationFiles(migrationDir)
	if err != nil {
		return fmt.Errorf("列出迁移文件失败: %w", err)
	}

	if len(applied) == 0 {
		existing, err := isExistingBusinessDatabase(db)
		if err != nil {
			return fmt.Errorf("检测数据库基线状态失败: %w", err)
		}
		if existing {
			baselineVersion, missingTables, err := determineAutoBaselineVersion(db)
			if err != nil {
				return fmt.Errorf("计算自动迁移基线失败: %w", err)
			}
			if err := baselineHistoricalMigrations(db, files, baselineVersion); err != nil {
				return fmt.Errorf("写入历史迁移基线失败: %w", err)
			}
			applied, err = loadAppliedMigrations(db)
			if err != nil {
				return fmt.Errorf("刷新迁移记录失败: %w", err)
			}
			if len(missingTables) > 0 {
				obslogger.L().Info("Migration 检测到已有业务数据库但缺少关键表", "tables", strings.Join(missingTables, ", "), "baseline_version", baselineVersion)
			} else {
				obslogger.L().Info("Migration 已将历史迁移基线化", "baseline_version", baselineVersion)
			}
		}
	}

	appliedCount := 0
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("读取迁移文件 %s 失败: %w", filepath.Base(file), err)
		}
		checksum := migrationChecksum(content)
		name := filepath.Base(file)

		if oldChecksum, ok := applied[name]; ok {
			if oldChecksum != checksum {
				obslogger.L().Warn("Migration 已执行文件内容发生变化，已跳过重新执行", "name", name)
			}
			continue
		}

		statements, err := parseMigrationStatements(string(content))
		if err != nil {
			return fmt.Errorf("解析迁移文件 %s 失败: %w", name, err)
		}
		if len(statements) == 0 {
			if err := recordAppliedMigration(db, name, checksum); err != nil {
				return fmt.Errorf("记录空迁移 %s 失败: %w", name, err)
			}
			continue
		}

		obslogger.L().Info("Migration 开始执行", "name", name, "statement_count", len(statements))
		for idx, stmt := range statements {
			if _, err := db.Exec(stmt); err != nil {
				if isIgnorableMigrationError(err) {
					obslogger.L().Warn("Migration 语句已存在或已处理，忽略错误", "name", name, "statement_index", idx+1, "error", err)
					continue
				}
				return fmt.Errorf("%s 第 %d 条语句执行失败: %w", name, idx+1, err)
			}
		}

		if err := recordAppliedMigration(db, name, checksum); err != nil {
			return fmt.Errorf("记录迁移 %s 失败: %w", name, err)
		}
		appliedCount++
		obslogger.L().Info("Migration 已完成", "name", name)
	}

	if appliedCount > 0 {
		obslogger.L().Info("Migration 本次自动执行完成", "applied_count", appliedCount)
	}
	return nil
}

func resolveMigrationDir() (string, error) {
	candidates := make([]string, 0, 6)
	if config.Global != nil {
		if raw := config.Global.MigrationsDirValue(); raw != "" {
			candidates = append(candidates, raw)
		}
	}

	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(wd, "migrations", "core"),
			filepath.Join(wd, "go-api", "migrations", "core"),
		)
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "migrations", "core"),
			filepath.Join(exeDir, "go-api", "migrations", "core"),
			filepath.Join(filepath.Dir(exeDir), "migrations", "core"),
		)
	}

	seen := map[string]bool{}
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) == "" {
			continue
		}
		abs := candidate
		if !filepath.IsAbs(abs) {
			var err error
			abs, err = filepath.Abs(abs)
			if err != nil {
				continue
			}
		}
		if seen[abs] {
			continue
		}
		seen[abs] = true
		info, err := os.Stat(abs)
		if err == nil && info.IsDir() {
			return abs, nil
		}
	}

	return "", fmt.Errorf("未找到 migrations/core 目录")
}

func listMigrationFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !migrationFilePattern.MatchString(name) {
			continue
		}
		files = append(files, filepath.Join(dir, name))
	}
	sort.Strings(files)
	return files, nil
}

func ensureMigrationTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS qingka_schema_migration (
		filename VARCHAR(255) NOT NULL PRIMARY KEY,
		checksum VARCHAR(64) NOT NULL DEFAULT '',
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	return err
}

func loadAppliedMigrations(db *sql.DB) (map[string]string, error) {
	rows, err := db.Query("SELECT filename, checksum FROM qingka_schema_migration")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var filename, checksum string
		if err := rows.Scan(&filename, &checksum); err != nil {
			return nil, err
		}
		out[filename] = checksum
	}
	return out, rows.Err()
}

func isExistingBusinessDatabase(db *sql.DB) (bool, error) {
	var count int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME IN ('qingka_wangke_user', 'qingka_wangke_order', 'qingka_wangke_config')",
	).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func determineAutoBaselineVersion(db *sql.DB) (int, []string, error) {
	tableNames := make([]string, 0, len(autoBaselineCheckpoints))
	for _, checkpoint := range autoBaselineCheckpoints {
		tableNames = append(tableNames, checkpoint.table)
	}

	existingTables, err := loadExistingTables(db, tableNames)
	if err != nil {
		return 0, nil, err
	}

	return calculateAutoBaselineVersion(existingTables), collectMissingCheckpointTables(existingTables), nil
}

func loadExistingTables(db *sql.DB, tableNames []string) (map[string]bool, error) {
	if len(tableNames) == 0 {
		return map[string]bool{}, nil
	}

	placeholders := strings.TrimRight(strings.Repeat("?,", len(tableNames)), ",")
	args := make([]any, 0, len(tableNames))
	for _, tableName := range tableNames {
		args = append(args, tableName)
	}

	query := fmt.Sprintf(
		"SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME IN (%s)",
		placeholders,
	)
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	existingTables := make(map[string]bool, len(tableNames))
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		existingTables[tableName] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return existingTables, nil
}

func calculateAutoBaselineVersion(existingTables map[string]bool) int {
	baselineVersion := autoMigrationBaselineVersion
	for _, checkpoint := range autoBaselineCheckpoints {
		if existingTables[checkpoint.table] {
			continue
		}
		if checkpoint.baselineVersion < baselineVersion {
			baselineVersion = checkpoint.baselineVersion
		}
	}
	return baselineVersion
}

func collectMissingCheckpointTables(existingTables map[string]bool) []string {
	missingTables := make([]string, 0, len(autoBaselineCheckpoints))
	for _, checkpoint := range autoBaselineCheckpoints {
		if existingTables[checkpoint.table] {
			continue
		}
		missingTables = append(missingTables, checkpoint.table)
	}
	return missingTables
}

func baselineHistoricalMigrations(db *sql.DB, files []string, maxVersion int) error {
	for _, file := range files {
		name := filepath.Base(file)
		version := migrationVersion(name)
		if version <= 0 || version > maxVersion {
			continue
		}
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if err := recordAppliedMigration(db, name, migrationChecksum(content)); err != nil {
			return err
		}
	}
	return nil
}

func recordAppliedMigration(db *sql.DB, filename string, checksum string) error {
	_, err := db.Exec(
		"INSERT INTO qingka_schema_migration (filename, checksum) VALUES (?, ?) ON DUPLICATE KEY UPDATE checksum = VALUES(checksum), applied_at = CURRENT_TIMESTAMP",
		filename, checksum,
	)
	return err
}

func migrationChecksum(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func migrationVersion(name string) int {
	var version int
	for i := 0; i < len(name); i++ {
		ch := name[i]
		if ch < '0' || ch > '9' {
			break
		}
		version = version*10 + int(ch-'0')
	}
	return version
}

func parseMigrationStatements(content string) ([]string, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 1024), 10*1024*1024)

	delimiter := ";"
	var builder strings.Builder
	statements := make([]string, 0, 16)

	flush := func() {
		stmt := strings.TrimSpace(builder.String())
		builder.Reset()
		if stmt == "" {
			return
		}
		stmt = strings.TrimSpace(strings.TrimSuffix(stmt, delimiter))
		if stmt == "" || isCommentOnlyStatement(stmt) {
			return
		}
		statements = append(statements, stmt)
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		upper := strings.ToUpper(trimmed)
		if strings.HasPrefix(upper, "DELIMITER ") {
			delimiter = strings.TrimSpace(trimmed[len("DELIMITER "):])
			if delimiter == "" {
				return nil, fmt.Errorf("检测到空 DELIMITER")
			}
			continue
		}

		if builder.Len() > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(line)

		current := strings.TrimSpace(builder.String())
		if current == "" {
			continue
		}
		if strings.HasSuffix(current, delimiter) {
			flush()
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if strings.TrimSpace(builder.String()) != "" && !isCommentOnlyStatement(builder.String()) {
		statements = append(statements, strings.TrimSpace(builder.String()))
	}

	return statements, nil
}

func isCommentOnlyStatement(stmt string) bool {
	lines := strings.Split(stmt, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "/*") && strings.HasSuffix(trimmed, "*/") {
			continue
		}
		return false
	}
	return true
}

func isIgnorableMigrationError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if !errors.As(err, &mysqlErr) {
		return false
	}

	switch mysqlErr.Number {
	case 1050, 1060, 1061, 1062, 1091, 1304:
		return true
	default:
		return false
	}
}
