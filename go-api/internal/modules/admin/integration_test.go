package admin

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"go-api/internal/database"

	"github.com/gin-gonic/gin"
)

const adminTestDriverName = "admin-module-test-driver"

var adminQueryHook func(query string, args []driver.NamedValue) (driver.Rows, error)
var adminExecHook func(query string, args []driver.NamedValue) error

func init() {
	sql.Register(adminTestDriverName, adminTestDriver{})
}

type adminTestDriver struct{}

func (adminTestDriver) Open(name string) (driver.Conn, error) {
	return adminTestConn{}, nil
}

type adminTestConn struct{}

func (adminTestConn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("prepare not supported in admin test driver")
}

func (adminTestConn) Close() error {
	return nil
}

func (adminTestConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions not supported in admin test driver")
}

func (adminTestConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if adminExecHook == nil {
		return nil, errors.New("unexpected exec without hook")
	}
	if err := adminExecHook(query, args); err != nil {
		return nil, err
	}
	return driver.RowsAffected(1), nil
}

func (adminTestConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if adminQueryHook == nil {
		return nil, errors.New("unexpected query without hook")
	}
	return adminQueryHook(query, args)
}

type adminTestRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *adminTestRows) Columns() []string {
	return r.columns
}

func (r *adminTestRows) Close() error {
	return nil
}

func (r *adminTestRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

func adminRows(columns []string, values ...[]driver.Value) driver.Rows {
	return &adminTestRows{columns: columns, values: values}
}

func withAdminTestDB(t *testing.T, queryHook func(string, []driver.NamedValue) (driver.Rows, error), execHook func(string, []driver.NamedValue) error) {
	t.Helper()

	db, err := sql.Open(adminTestDriverName, "")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	originalDB := database.DB
	originalQueryHook := adminQueryHook
	originalExecHook := adminExecHook
	database.DB = db
	adminQueryHook = queryHook
	adminExecHook = execHook

	t.Cleanup(func() {
		adminQueryHook = originalQueryHook
		adminExecHook = originalExecHook
		database.DB = originalDB
		_ = db.Close()
	})
}

func performAdminGETRequest(t *testing.T, handler gin.HandlerFunc, target string) *testResponse {
	t.Helper()
	w := performAdminJSONRequest(t, handler, http.MethodGet, target, nil)
	resp := decodeAdminResponse(t, w)
	if w.Code != http.StatusOK || resp.Code != 0 {
		t.Fatalf("expected success response, got status=%d code=%d body=%s", w.Code, resp.Code, w.Body.String())
	}
	return &resp
}

func TestAdminConfigHandlersUseDBFlow(t *testing.T) {
	var replaceArgs []driver.NamedValue

	withAdminTestDB(t, func(query string, args []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(query, "SELECT `v`, `k` FROM qingka_wangke_config") {
			return adminRows(
				[]string{"v", "k"},
				[]driver.Value{"sitename", "Demo Site"},
				[]driver.Value{"checkin_enabled", "1"},
			), nil
		}
		return nil, errors.New("unexpected query: " + query)
	}, func(query string, args []driver.NamedValue) error {
		if strings.Contains(query, "REPLACE INTO qingka_wangke_config (`v`, `k`) VALUES") {
			replaceArgs = append([]driver.NamedValue(nil), args...)
			return nil
		}
		return errors.New("unexpected exec: " + query)
	})

	resp := performAdminGETRequest(t, AdminConfigGet, "/admin/config")
	var configMap map[string]string
	if err := json.Unmarshal(resp.Data, &configMap); err != nil {
		t.Fatalf("decode config map: %v", err)
	}
	if configMap["sitename"] != "Demo Site" || configMap["checkin_enabled"] != "1" {
		t.Fatalf("unexpected config payload: %v", configMap)
	}

	w := performAdminJSONRequest(t, AdminConfigSave, http.MethodPost, "/admin/config", gin.H{
		"sitename":         "Refined Site",
		"checkin_enabled":  "0",
	})
	saveResp := decodeAdminResponse(t, w)
	if w.Code != http.StatusOK || saveResp.Code != 0 {
		t.Fatalf("expected save success, got status=%d code=%d body=%s", w.Code, saveResp.Code, w.Body.String())
	}
	if len(replaceArgs) != 4 {
		t.Fatalf("expected 4 replace args, got %d", len(replaceArgs))
	}
}

func TestAdminPayDataHandlersUseDBFlow(t *testing.T) {
	var updateArgs []driver.NamedValue

	withAdminTestDB(t, func(query string, args []driver.NamedValue) (driver.Rows, error) {
		if strings.Contains(query, "SELECT COALESCE(paydata,'') FROM qingka_wangke_user WHERE uid = 1") {
			return adminRows(
				[]string{"paydata"},
				[]driver.Value{`{"alipay_appid":"old-app","wechat_mchid":"m-1"}`},
			), nil
		}
		return nil, errors.New("unexpected query: " + query)
	}, func(query string, args []driver.NamedValue) error {
		if strings.Contains(query, "UPDATE qingka_wangke_user SET paydata = ? WHERE uid = 1") {
			updateArgs = append([]driver.NamedValue(nil), args...)
			return nil
		}
		return errors.New("unexpected exec: " + query)
	})

	resp := performAdminGETRequest(t, AdminPayDataGet, "/admin/paydata")
	var payData map[string]string
	if err := json.Unmarshal(resp.Data, &payData); err != nil {
		t.Fatalf("decode paydata: %v", err)
	}
	if payData["alipay_appid"] != "old-app" || payData["wechat_mchid"] != "m-1" {
		t.Fatalf("unexpected paydata payload: %v", payData)
	}

	w := performAdminJSONRequest(t, AdminPayDataSave, http.MethodPost, "/admin/paydata", gin.H{
		"alipay_appid": "new-app",
	})
	saveResp := decodeAdminResponse(t, w)
	if w.Code != http.StatusOK || saveResp.Code != 0 {
		t.Fatalf("expected save success, got status=%d code=%d body=%s", w.Code, saveResp.Code, w.Body.String())
	}
	if len(updateArgs) != 1 {
		t.Fatalf("expected one update arg, got %d", len(updateArgs))
	}

	var merged map[string]string
	payload, ok := updateArgs[0].Value.(string)
	if !ok {
		t.Fatalf("expected string paydata payload, got %T", updateArgs[0].Value)
	}
	if err := json.Unmarshal([]byte(payload), &merged); err != nil {
		t.Fatalf("decode merged paydata payload: %v", err)
	}
	if merged["alipay_appid"] != "new-app" || merged["wechat_mchid"] != "m-1" {
		t.Fatalf("expected merged paydata payload, got %v", merged)
	}
}
