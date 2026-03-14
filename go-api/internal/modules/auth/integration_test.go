package auth

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"

	"go-api/internal/config"
	"go-api/internal/database"
	"go-api/internal/model"

	"golang.org/x/crypto/bcrypt"
)

const authTestDriverName = "auth-module-test-driver"

var authQueryHook func(query string, args []driver.NamedValue) (driver.Rows, error)
var authExecHook func(query string, args []driver.NamedValue) error

func init() {
	sql.Register(authTestDriverName, authTestDriver{})
}

type authTestDriver struct{}

func (authTestDriver) Open(name string) (driver.Conn, error) {
	return authTestConn{}, nil
}

type authTestConn struct{}

func (authTestConn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("prepare not supported in auth test driver")
}

func (authTestConn) Close() error {
	return nil
}

func (authTestConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions not supported in auth test driver")
}

func (authTestConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if authExecHook == nil {
		return nil, errors.New("unexpected exec without hook")
	}
	if err := authExecHook(query, args); err != nil {
		return nil, err
	}
	return driver.RowsAffected(1), nil
}

func (authTestConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if authQueryHook == nil {
		return nil, errors.New("unexpected query without hook")
	}
	return authQueryHook(query, args)
}

type authTestRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *authTestRows) Columns() []string {
	return r.columns
}

func (r *authTestRows) Close() error {
	return nil
}

func (r *authTestRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

func authRows(columns []string, values ...[]driver.Value) driver.Rows {
	return &authTestRows{columns: columns, values: values}
}

func withAuthTestDB(t *testing.T, queryHook func(string, []driver.NamedValue) (driver.Rows, error), execHook func(string, []driver.NamedValue) error) {
	t.Helper()

	db, err := sql.Open(authTestDriverName, "")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	originalDB := database.DB
	originalQueryHook := authQueryHook
	originalExecHook := authExecHook
	database.DB = db
	authQueryHook = queryHook
	authExecHook = execHook

	t.Cleanup(func() {
		authQueryHook = originalQueryHook
		authExecHook = originalExecHook
		database.DB = originalDB
		_ = db.Close()
	})
}

func TestAuthServiceLoginWithBcryptDBFlow(t *testing.T) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	originalConfig := config.Global
	config.Global = &config.Config{JWT: config.JWTConfig{Secret: "test-secret", AccessTTL: 60, RefreshTTL: 120}}
	defer func() { config.Global = originalConfig }()

	execCalled := false
	withAuthTestDB(t, func(query string, args []driver.NamedValue) (driver.Rows, error) {
		switch {
		case strings.Contains(query, "FROM qingka_wangke_user WHERE user = ?"):
			return authRows(
				[]string{"uid", "uuid", "user", "pass", "pass2", "name", "money", "grade", "active"},
				[]driver.Value{1, 1001, "alice", string(hashedPass), "", "Alice", 12.5, "2", "1"},
			), nil
		case strings.Contains(query, "FROM qingka_wangke_config WHERE `v`='pass2_kg'"):
			return authRows([]string{"k"}), nil
		default:
			return nil, errors.New("unexpected query: " + query)
		}
	}, func(query string, args []driver.NamedValue) error {
		if strings.Contains(query, "UPDATE qingka_wangke_user SET lasttime = NOW(), endtime = NOW() WHERE uid = ?") {
			execCalled = len(args) == 1 && args[0].Value == int64(1)
			return nil
		}
		return errors.New("unexpected exec: " + query)
	})

	svc := NewService()
	resp, refreshToken, err := svc.Login(LoginRequest{Username: "alice", Password: "secret123"})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if resp == nil || resp.UserId != "1" || resp.Username != "alice" {
		t.Fatalf("unexpected login response: %+v", resp)
	}
	if refreshToken == "" || resp.AccessToken == "" {
		t.Fatalf("expected generated tokens, got access=%q refresh=%q", resp.AccessToken, refreshToken)
	}
	if len(resp.Roles) != 2 || resp.Roles[0] != "admin" || resp.Roles[1] != "user" {
		t.Fatalf("unexpected roles: %v", resp.Roles)
	}
	if !execCalled {
		t.Fatalf("expected lasttime update exec to run")
	}
}

func TestAuthServiceRefreshAndUserInfoWithDBFlow(t *testing.T) {
	originalConfig := config.Global
	config.Global = &config.Config{JWT: config.JWTConfig{Secret: "refresh-secret", AccessTTL: 60, RefreshTTL: 120}}
	defer func() { config.Global = originalConfig }()

	withAuthTestDB(t, func(query string, args []driver.NamedValue) (driver.Rows, error) {
		switch {
		case strings.Contains(query, "FROM qingka_wangke_user WHERE uid = ?") && strings.Contains(query, "pass, name, money, grade, active"):
			return authRows(
				[]string{"uid", "uuid", "user", "pass", "name", "money", "grade", "active"},
				[]driver.Value{7, 7001, "bob", "$2a$stub", "Bob", 66.6, "1", "1"},
			), nil
		case strings.Contains(query, "FROM qingka_wangke_user WHERE uid = ?") && strings.Contains(query, "uid, uuid, user, name, money, grade, active"):
			return authRows(
				[]string{"uid", "uuid", "user", "name", "money", "grade", "active"},
				[]driver.Value{7, 7001, "bob", "Bob", 66.6, "1", "1"},
			), nil
		default:
			return nil, errors.New("unexpected query: " + query)
		}
	}, func(query string, args []driver.NamedValue) error {
		return errors.New("unexpected exec: " + query)
	})

	svc := NewService()
	refreshToken, err := svc.generateToken(model.User{UID: 7, User: "bob", Grade: "1"}, config.Global.JWT.RefreshTTL)
	if err != nil {
		t.Fatalf("generate refresh token: %v", err)
	}

	newAccessToken, newRefreshToken, err := svc.RefreshAccessToken(refreshToken)
	if err != nil {
		t.Fatalf("refresh access token: %v", err)
	}
	if newAccessToken == "" || newRefreshToken == "" {
		t.Fatalf("expected refreshed tokens, got access=%q refresh=%q", newAccessToken, newRefreshToken)
	}

	info, err := svc.GetUserInfo(7)
	if err != nil {
		t.Fatalf("get user info: %v", err)
	}
	if info == nil || info.UserId != "7" || info.Username != "bob" || info.RealName != "Bob" {
		t.Fatalf("unexpected user info: %+v", info)
	}
}
