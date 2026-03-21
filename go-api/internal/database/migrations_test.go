package database

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
)

func TestParseMigrationStatementsHandlesDelimiterBlocks(t *testing.T) {
	raw := `
-- comment
CREATE TABLE IF NOT EXISTS demo (
  id INT PRIMARY KEY
);

DELIMITER //
DROP PROCEDURE IF EXISTS _patch_demo //
CREATE PROCEDURE _patch_demo()
BEGIN
  IF NOT EXISTS (SELECT 1) THEN
    ALTER TABLE demo ADD COLUMN name VARCHAR(32);
  END IF;
END //
DELIMITER ;

CALL _patch_demo();
DROP PROCEDURE IF EXISTS _patch_demo;
`

	statements, err := parseMigrationStatements(raw)
	if err != nil {
		t.Fatalf("parseMigrationStatements returned error: %v", err)
	}
	if len(statements) != 5 {
		t.Fatalf("expected 5 statements, got %d: %#v", len(statements), statements)
	}
	if !strings.Contains(statements[0], "CREATE TABLE IF NOT EXISTS demo") {
		t.Fatalf("unexpected first statement: %q", statements[0])
	}
	if !strings.HasPrefix(statements[1], "DROP PROCEDURE IF EXISTS _patch_demo") {
		t.Fatalf("unexpected second statement: %q", statements[1])
	}
	if !strings.Contains(statements[2], "CREATE PROCEDURE _patch_demo()") || !strings.Contains(statements[2], "ALTER TABLE demo ADD COLUMN name VARCHAR(32);") {
		t.Fatalf("unexpected procedure statement: %q", statements[2])
	}
	if statements[3] != "CALL _patch_demo()" {
		t.Fatalf("unexpected fourth statement: %q", statements[3])
	}
	if statements[4] != "DROP PROCEDURE IF EXISTS _patch_demo" {
		t.Fatalf("unexpected fifth statement: %q", statements[4])
	}
}

func TestParseMigrationStatementsSplitsTrailingStatements(t *testing.T) {
	raw := `
CALL _patch_demo();
DROP PROCEDURE IF EXISTS _patch_demo;
`

	statements, err := parseMigrationStatements(raw)
	if err != nil {
		t.Fatalf("parseMigrationStatements returned error: %v", err)
	}
	if len(statements) != 2 {
		t.Fatalf("expected 2 statements, got %d: %#v", len(statements), statements)
	}
	if statements[0] != "CALL _patch_demo()" {
		t.Fatalf("unexpected first statement: %q", statements[0])
	}
	if statements[1] != "DROP PROCEDURE IF EXISTS _patch_demo" {
		t.Fatalf("unexpected second statement: %q", statements[1])
	}
}

func TestIsIgnorableMigrationError(t *testing.T) {
	cases := []struct {
		err  error
		want bool
	}{
		{err: &mysql.MySQLError{Number: 1060, Message: "Duplicate column name 'x'"}, want: true},
		{err: &mysql.MySQLError{Number: 1062, Message: "Duplicate entry"}, want: true},
		{err: &mysql.MySQLError{Number: 1304, Message: "PROCEDURE already exists"}, want: true},
		{err: &mysql.MySQLError{Number: 1064, Message: "syntax error"}, want: false},
		{err: errors.New("plain error"), want: false},
	}

	for _, tc := range cases {
		if got := isIgnorableMigrationError(tc.err); got != tc.want {
			t.Fatalf("unexpected ignorable result for %v: got=%v want=%v", tc.err, got, tc.want)
		}
	}
}

func Test017PlatformConfigSeedColumnCount(t *testing.T) {
	path := filepath.Join("..", "..", "migrations", "core", "017_platform_config.sql")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration file failed: %v", err)
	}

	re := regexp.MustCompile("(?s)INSERT INTO `qingka_platform_config` \\((.*?)\\) VALUES\\s*(.*)\\s*ON DUPLICATE KEY UPDATE")
	matches := re.FindStringSubmatch(string(raw))
	if len(matches) != 3 {
		t.Fatalf("failed to locate platform config seed insert block")
	}

	columnCount := len(regexp.MustCompile("`([^`]+)`").FindAllStringSubmatch(matches[1], -1))
	rows := splitTopLevelRows(matches[2])
	if len(rows) == 0 {
		t.Fatalf("expected at least one seed row")
	}

	for idx, row := range rows {
		if got := len(splitTopLevelCSV(row[1 : len(row)-1])); got != columnCount {
			t.Fatalf("seed row %d column mismatch: got=%d want=%d row=%s", idx+1, got, columnCount, row)
		}
	}
}

func TestCalculateAutoBaselineVersion(t *testing.T) {
	cases := []struct {
		name     string
		existing map[string]bool
		want     int
	}{
		{
			name: "full checkpoint schema",
			existing: map[string]bool{
				"qingka_dynamic_module":       true,
				"qingka_platform_config":      true,
				"qingka_wangke_sync_config":   true,
				"qingka_tenant":               true,
				"qingka_ext_menu":             true,
				"qingka_wangke_yfdk_projects": true,
			},
			want: 46,
		},
		{
			name: "missing dynamic module falls back before 004",
			existing: map[string]bool{
				"qingka_platform_config":      true,
				"qingka_wangke_sync_config":   true,
				"qingka_tenant":               true,
				"qingka_ext_menu":             true,
				"qingka_wangke_yfdk_projects": true,
			},
			want: 3,
		},
		{
			name: "missing platform config falls back before 017",
			existing: map[string]bool{
				"qingka_dynamic_module":       true,
				"qingka_wangke_sync_config":   true,
				"qingka_tenant":               true,
				"qingka_ext_menu":             true,
				"qingka_wangke_yfdk_projects": true,
			},
			want: 16,
		},
		{
			name: "missing sync config falls back before 019",
			existing: map[string]bool{
				"qingka_dynamic_module":       true,
				"qingka_platform_config":      true,
				"qingka_tenant":               true,
				"qingka_ext_menu":             true,
				"qingka_wangke_yfdk_projects": true,
			},
			want: 18,
		},
		{
			name: "missing tenant falls back before 029",
			existing: map[string]bool{
				"qingka_dynamic_module":       true,
				"qingka_platform_config":      true,
				"qingka_wangke_sync_config":   true,
				"qingka_ext_menu":             true,
				"qingka_wangke_yfdk_projects": true,
			},
			want: 28,
		},
		{
			name: "earliest missing checkpoint wins",
			existing: map[string]bool{
				"qingka_dynamic_module":       true,
				"qingka_tenant":               true,
				"qingka_ext_menu":             true,
				"qingka_wangke_yfdk_projects": true,
			},
			want: 16,
		},
	}

	for _, tc := range cases {
		if got := calculateAutoBaselineVersion(tc.existing); got != tc.want {
			t.Fatalf("%s: got baseline=%d want=%d", tc.name, got, tc.want)
		}
	}
}

func TestCollectMissingCheckpointTables(t *testing.T) {
	got := collectMissingCheckpointTables(map[string]bool{
		"qingka_dynamic_module":     true,
		"qingka_platform_config":    true,
		"qingka_wangke_sync_config": true,
	})

	want := []string{
		"qingka_tenant",
		"qingka_ext_menu",
		"qingka_wangke_yfdk_projects",
	}

	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("unexpected missing tables: got=%v want=%v", got, want)
	}
}

func splitTopLevelRows(valuesBlock string) []string {
	rows := make([]string, 0, 8)
	var builder strings.Builder
	depth := 0
	inString := false
	escape := false

	for _, ch := range valuesBlock {
		if ch == '\'' && !escape {
			inString = !inString
		}
		if ch == '(' && !inString {
			depth++
		}
		if depth > 0 {
			builder.WriteRune(ch)
		}
		if ch == ')' && !inString {
			depth--
			if depth == 0 {
				rows = append(rows, builder.String())
				builder.Reset()
			}
		}
		if ch == '\\' && !escape {
			escape = true
		} else {
			escape = false
		}
	}

	return rows
}

func splitTopLevelCSV(content string) []string {
	items := make([]string, 0, 16)
	var builder strings.Builder
	inString := false
	escape := false

	for _, ch := range content {
		if ch == '\'' && !escape {
			inString = !inString
		}
		if ch == ',' && !inString {
			items = append(items, strings.TrimSpace(builder.String()))
			builder.Reset()
		} else {
			builder.WriteRune(ch)
		}
		if ch == '\\' && !escape {
			escape = true
		} else {
			escape = false
		}
	}

	if tail := strings.TrimSpace(builder.String()); tail != "" {
		items = append(items, tail)
	}
	return items
}
