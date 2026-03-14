package dbtools

import (
	"database/sql"
	"testing"

	"github.com/go-sql-driver/mysql"
)

func TestSyncTableSourceCandidatesIncludesAliases(t *testing.T) {
	var gonggao syncTableDef
	for _, table := range syncTables {
		if table.name == "qingka_wangke_gonggao" {
			gonggao = table
			break
		}
	}

	if gonggao.name == "" {
		t.Fatalf("expected qingka_wangke_gonggao in sync tables")
	}

	candidates := gonggao.sourceCandidates()
	expected := []string{
		"qingka_wangke_gonggao",
		"qingka_wangke_notice",
		"love_learn_notice",
	}
	for _, name := range expected {
		if !containsString(candidates, name) {
			t.Fatalf("expected candidate %q in %v", name, candidates)
		}
	}
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func TestShouldSkipEmptySourceTable(t *testing.T) {
	if !shouldSkipEmptySourceTable(0) {
		t.Fatalf("expected empty source table to be skipped")
	}
	if shouldSkipEmptySourceTable(1) {
		t.Fatalf("expected non-empty source table to be imported")
	}
}

func TestBaseTablesDoNotMergeByBusinessKey(t *testing.T) {
	var userTable syncTableDef
	var classTable syncTableDef
	var orderTable syncTableDef
	for _, table := range syncTables {
		switch table.name {
		case "qingka_wangke_user":
			userTable = table
		case "qingka_wangke_class":
			classTable = table
		case "qingka_wangke_order":
			orderTable = table
		}
	}

	if userTable.name == "" || classTable.name == "" || orderTable.name == "" {
		t.Fatalf("expected user/class/order tables in sync config")
	}
	if userTable.mergeByMatch {
		t.Fatalf("expected user table to disable global business-key merge")
	}
	if classTable.mergeByMatch {
		t.Fatalf("expected class table to disable global business-key merge")
	}
	if !orderTable.mergeByMatch {
		t.Fatalf("expected order table to keep business-key merge enabled")
	}
}

func TestBuildFirstSyncMatchKeyUsesBusinessColumns(t *testing.T) {
	values := map[string]sql.NullString{
		"out_trade_no": {String: "OTN-001", Valid: true},
		"user":         {String: "alice", Valid: true},
		"kcname":       {String: "课程A", Valid: true},
		"addtime":      {String: "2026-03-14 20:00:00", Valid: true},
	}

	key, ok := buildFirstSyncMatchKey([]syncMatchRule{
		{name: "out-trade-no", columns: []string{"out_trade_no"}},
		{name: "user-kcname-addtime", columns: []string{"user", "kcname", "addtime"}},
	}, values)
	if !ok {
		t.Fatalf("expected business match key to be generated")
	}
	if key != "out-trade-no::OTN-001" {
		t.Fatalf("unexpected match key %q", key)
	}
}

func TestApplySyncReferenceRemapsUsesMappedIDs(t *testing.T) {
	values := map[string]sql.NullString{
		"uid": {String: "12", Valid: true},
		"cid": {String: "34", Valid: true},
	}
	warnings, err := applySyncReferenceRemaps(values, []syncReferenceDef{
		{column: "uid", refTable: "qingka_wangke_user"},
		{column: "cid", refTable: "qingka_wangke_class"},
	}, newSyncExecutionContext(), func(refTable, sourceValue string) (string, bool, string, error) {
		if refTable == "qingka_wangke_user" {
			return "1001", true, "", nil
		}
		if refTable == "qingka_wangke_class" {
			return "2002", true, "", nil
		}
		return "", false, "", sql.ErrNoRows
	})
	if err != nil {
		t.Fatalf("expected remap to succeed, got %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
	if values["uid"].String != "1001" {
		t.Fatalf("expected uid to be remapped, got %q", values["uid"].String)
	}
	if values["cid"].String != "2002" {
		t.Fatalf("expected cid to be remapped, got %q", values["cid"].String)
	}
}

func TestApplySyncReferenceRemapsFailsWhenMappingMissing(t *testing.T) {
	ctx := newSyncExecutionContext()
	values := map[string]sql.NullString{
		"uid": {String: "12", Valid: true},
	}

	_, err := applySyncReferenceRemaps(values, []syncReferenceDef{{column: "uid", refTable: "qingka_wangke_user"}}, ctx, func(refTable, sourceValue string) (string, bool, string, error) {
		return "", false, "", sql.ErrNoRows
	})
	if err == nil {
		t.Fatalf("expected remap failure when mapping is missing")
	}
}

func TestResolveSyncReferenceTargetFallsBackToSameID(t *testing.T) {
	service := &DBSyncService{}
	ctx := newSyncExecutionContext()
	ctx.localIDCache["qingka_wangke_user"] = map[string]struct{}{
		"578": {},
	}

	target, resolved, _, err := service.resolveSyncReferenceTarget(nil, ctx, "qingka_wangke_user", "578")
	if err != nil {
		t.Fatalf("expected fallback to same id, got %v", err)
	}
	if !resolved {
		t.Fatalf("expected fallback same id to count as resolved")
	}
	if target != "578" {
		t.Fatalf("expected same id fallback, got %q", target)
	}
	if remembered, ok := ctx.resolve("qingka_wangke_user", "578"); !ok || remembered != "578" {
		t.Fatalf("expected fallback to be remembered, got %q", remembered)
	}
}

func TestResolveSyncReferenceTargetUsesRememberedMapping(t *testing.T) {
	service := &DBSyncService{}
	ctx := newSyncExecutionContext()
	ctx.remember("qingka_wangke_user", "12", "1001")

	target, resolved, _, err := service.resolveSyncReferenceTarget(nil, ctx, "qingka_wangke_user", "12")
	if err != nil {
		t.Fatalf("expected remembered mapping, got %v", err)
	}
	if !resolved {
		t.Fatalf("expected remembered mapping to count as resolved")
	}
	if target != "1001" {
		t.Fatalf("expected remembered target id, got %q", target)
	}
}

func TestApplySyncReferenceRemapsKeepsRawValueWhenNoMappingFound(t *testing.T) {
	values := map[string]sql.NullString{
		"docking": {String: "1", Valid: true},
	}

	warnings, err := applySyncReferenceRemaps(values, []syncReferenceDef{{column: "docking", refTable: "qingka_wangke_huoyuan"}}, newSyncExecutionContext(), func(refTable, sourceValue string) (string, bool, string, error) {
		return sourceValue, false, "源库不存在 qingka_wangke_huoyuan=1 的关联记录", nil
	})
	if err != nil {
		t.Fatalf("expected raw value fallback, got %v", err)
	}
	if values["docking"].String != "1" {
		t.Fatalf("expected raw docking value to remain, got %q", values["docking"].String)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected one warning, got %v", warnings)
	}
	if warnings[0] != "字段 docking 保留原值 1（源库不存在 qingka_wangke_huoyuan=1 的关联记录）" {
		t.Fatalf("unexpected warning: %v", warnings)
	}
}

func TestNormalizeSyncWriteValueUsesEmptyStringForNonNullableText(t *testing.T) {
	got := normalizeSyncWriteValue(sql.NullString{}, syncColumnMeta{
		nullable: false,
		dataType: "varchar",
	})
	if got != "" {
		t.Fatalf("expected empty string fallback, got %#v", got)
	}
}

func TestNormalizeSyncWriteValueUsesColumnDefaultWhenPresent(t *testing.T) {
	got := normalizeSyncWriteValue(sql.NullString{}, syncColumnMeta{
		nullable:     false,
		defaultValue: sql.NullString{String: "0", Valid: true},
		dataType:     "int",
	})
	if got != "0" {
		t.Fatalf("expected default value fallback, got %#v", got)
	}
}

func TestIsSyncDuplicatePrimaryError(t *testing.T) {
	err := &mysql.MySQLError{
		Number:  1062,
		Message: "Duplicate entry '123663' for key 'PRIMARY'",
	}
	if !isSyncDuplicatePrimaryError(err) {
		t.Fatalf("expected duplicate primary key error to be detected")
	}
}
