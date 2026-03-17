package database

import (
	"errors"
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
