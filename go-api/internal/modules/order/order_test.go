package order

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"go-api/internal/database"
	"go-api/internal/model"
	suppliermodule "go-api/internal/modules/supplier"
)

type stubRepository struct {
	listCalled         bool
	detailCalled       bool
	statsCalled        bool
	changeCalled       bool
	cancelCalled       bool
	refundCalled       bool
	modifyCalled       bool
	manualDockCalled   bool
	syncProgressCalled bool
	autoSyncCalled     bool
	batchSyncCalled    bool
	batchResendCalled  bool
	addCalled          bool
	addForMallCalled   bool
}

type stubSupplierGateway struct {
	getSupplierByHID       func(hid int) (*model.SupplierFull, error)
	getClassFull           func(cid int) (*model.ClassFull, error)
	callOrder              func(sup *model.SupplierFull, cls *model.ClassFull, school, user, pass, kcid, kcname string, extraFields map[string]string) (*model.SupplierOrderResult, error)
	hasBatchProgressForPT  func(pt string) bool
	hasBatchProgressConfig func(sup *model.SupplierFull) bool
	queryBatchProgress     func(sup *model.SupplierFull, refs []model.SupplierBatchProgressRef) ([]model.SupplierProgressItem, error)
	queryProgress          func(sup *model.SupplierFull, yid string, username string, orderExtra map[string]string) ([]model.SupplierProgressItem, error)
	resubmitOrder          func(sup *model.SupplierFull, yid string) (int, string, error)
}

func (s *stubSupplierGateway) GetSupplierByHID(hid int) (*model.SupplierFull, error) {
	return s.getSupplierByHID(hid)
}

func (s *stubSupplierGateway) GetClassFull(cid int) (*model.ClassFull, error) {
	return s.getClassFull(cid)
}

func (s *stubSupplierGateway) CallSupplierOrder(sup *model.SupplierFull, cls *model.ClassFull, school, user, pass, kcid, kcname string, extraFields map[string]string) (*model.SupplierOrderResult, error) {
	return s.callOrder(sup, cls, school, user, pass, kcid, kcname, extraFields)
}

func (s *stubSupplierGateway) HasBatchProgressForPT(pt string) bool {
	if s.hasBatchProgressForPT == nil {
		return false
	}
	return s.hasBatchProgressForPT(pt)
}

func (s *stubSupplierGateway) HasBatchProgressConfig(sup *model.SupplierFull) bool {
	if s.hasBatchProgressConfig == nil {
		return false
	}
	return s.hasBatchProgressConfig(sup)
}

func (s *stubSupplierGateway) QueryBatchOrderProgress(sup *model.SupplierFull, refs []model.SupplierBatchProgressRef) ([]model.SupplierProgressItem, error) {
	if s.queryBatchProgress == nil {
		return nil, errors.New("query batch progress not implemented")
	}
	return s.queryBatchProgress(sup, refs)
}

func (s *stubSupplierGateway) QueryOrderProgress(sup *model.SupplierFull, yid string, username string, orderExtra map[string]string) ([]model.SupplierProgressItem, error) {
	return s.queryProgress(sup, yid, username, orderExtra)
}

func (s *stubSupplierGateway) ResubmitOrder(sup *model.SupplierFull, yid string) (int, string, error) {
	return s.resubmitOrder(sup, yid)
}

const changeStatusTestDriverName = "order-change-status-test-driver"

var changeStatusExecHook func(query string, args []driver.NamedValue) error
var changeStatusQueryHook func(query string, args []driver.NamedValue) (driver.Rows, error)

func init() {
	sql.Register(changeStatusTestDriverName, changeStatusTestDriver{})
}

type changeStatusTestDriver struct{}

func (changeStatusTestDriver) Open(name string) (driver.Conn, error) {
	return changeStatusTestConn{}, nil
}

type changeStatusTestConn struct{}

func (changeStatusTestConn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("prepare not supported in change status test driver")
}

func (changeStatusTestConn) Close() error {
	return nil
}

func (changeStatusTestConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions not supported in change status test driver")
}

func (changeStatusTestConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if changeStatusExecHook == nil {
		return nil, errors.New("unexpected exec without hook")
	}
	if err := changeStatusExecHook(query, args); err != nil {
		return nil, err
	}
	return driver.RowsAffected(1), nil
}

func (changeStatusTestConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if changeStatusQueryHook == nil {
		return nil, errors.New("unexpected query without hook")
	}
	return changeStatusQueryHook(query, args)
}

type testRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *testRows) Columns() []string {
	return r.columns
}

func (r *testRows) Close() error {
	return nil
}

func (r *testRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

func singleRow(columns []string, values ...driver.Value) driver.Rows {
	return &testRows{
		columns: columns,
		values:  [][]driver.Value{values},
	}
}

func namedValueStrings(args []driver.NamedValue) []string {
	values := make([]string, len(args))
	for i, arg := range args {
		values[i] = fmt.Sprint(arg.Value)
	}
	return values
}

func (s *stubRepository) List(uid int, grade string, req model.OrderListRequest) ([]model.Order, int64, error) {
	s.listCalled = true
	return []model.Order{{OID: 1}}, 1, nil
}

func (s *stubRepository) Detail(uid int, grade string, oid int) (*model.Order, error) {
	s.detailCalled = true
	return &model.Order{OID: oid}, nil
}

func (s *stubRepository) Stats(uid int, grade string) (*model.OrderStats, error) {
	s.statsCalled = true
	return &model.OrderStats{Total: 3}, nil
}

func (s *stubRepository) AddOrders(uid int, req model.OrderAddRequest) (*model.OrderAddResult, error) {
	s.addCalled = true
	return &model.OrderAddResult{SuccessCount: 1}, nil
}

func (s *stubRepository) AddOrdersForMall(bUID, tid, cUID int, retailPrice float64, req model.OrderAddRequest) (*model.OrderAddResult, error) {
	s.addForMallCalled = true
	return &model.OrderAddResult{SuccessCount: 1}, nil
}

func (s *stubRepository) ChangeStatus(uid int, grade string, req model.OrderStatusRequest) error {
	s.changeCalled = true
	return nil
}

func (s *stubRepository) CancelOrder(uid int, grade string, oid int) error {
	s.cancelCalled = true
	return nil
}

func (s *stubRepository) RefundOrders(uid int, grade string, oids []int) error {
	s.refundCalled = true
	return nil
}

func (s *stubRepository) ModifyRemarks(oids []int, remarks string) error {
	s.modifyCalled = true
	return nil
}

func (s *stubRepository) ManualDockOrders(oids []int) (int, int, error) {
	s.manualDockCalled = true
	return 1, 0, nil
}

func (s *stubRepository) SyncOrderProgress(oids []int) (int, error) {
	s.syncProgressCalled = true
	return 1, nil
}

func (s *stubRepository) AutoSyncAllProgress(opts AutoSyncOptions) (int, int, error) {
	s.autoSyncCalled = true
	return 2, 1, nil
}

func (s *stubRepository) BatchSyncOrders(oids []int) (int, error) {
	s.batchSyncCalled = true
	return 2, nil
}

func (s *stubRepository) BatchResendOrders(oids []int) (int, int, error) {
	s.batchResendCalled = true
	return 1, 1, errors.New("partial failure")
}

func TestQueryServicesDelegateToRepository(t *testing.T) {
	repo := &stubRepository{}
	query := NewQueryService(repo)

	list, total, err := query.List(1, "3", model.OrderListRequest{})
	if err != nil || total != 1 || len(list) != 1 || !repo.listCalled {
		t.Fatalf("query list delegation failed: list=%v total=%d err=%v called=%v", list, total, err, repo.listCalled)
	}

	detail, err := query.Detail(1, "3", 9)
	if err != nil || detail == nil || detail.OID != 9 || !repo.detailCalled {
		t.Fatalf("query detail delegation failed: detail=%v err=%v called=%v", detail, err, repo.detailCalled)
	}

	stats, err := query.Stats(1, "3")
	if err != nil || stats == nil || stats.Total != 3 || !repo.statsCalled {
		t.Fatalf("query stats delegation failed: stats=%v err=%v called=%v", stats, err, repo.statsCalled)
	}
}

func TestSetOrderStatusNotifier(t *testing.T) {
	original := orderStatusNotifier
	defer func() {
		orderStatusNotifier = original
	}()

	called := false
	SetOrderStatusNotifier(func(oid int, newStatus string, newProcess string, remarks string) {
		called = oid == 7 && newStatus == "已完成" && newProcess == "100%" && remarks == "done"
	})
	orderStatusNotifier(7, "已完成", "100%", "done")
	if !called {
		t.Fatalf("custom notifier was not invoked as expected")
	}

	SetOrderStatusNotifier(nil)
	called = false
	orderStatusNotifier(8, "已取消", "", "")
	if called {
		t.Fatalf("notifier should reset to no-op when nil is provided")
	}
}

func TestCommandAndSyncServicesDelegateToRepository(t *testing.T) {
	repo := &stubRepository{}
	command := NewCommandService(repo)
	syncSvc := NewSyncService(repo)

	result, err := command.Add(1, model.OrderAddRequest{})
	if err != nil || result == nil || result.SuccessCount != 1 || !repo.addCalled {
		t.Fatalf("add delegation failed: result=%v err=%v called=%v", result, err, repo.addCalled)
	}

	if err := command.ChangeStatus(1, "3", model.OrderStatusRequest{Status: "已完成"}); err != nil || !repo.changeCalled {
		t.Fatalf("command delegation failed: err=%v called=%v", err, repo.changeCalled)
	}

	if err := command.Cancel(1, "3", 99); err != nil || !repo.cancelCalled {
		t.Fatalf("cancel delegation failed: err=%v called=%v", err, repo.cancelCalled)
	}

	if err := command.Refund(1, "3", []int{1, 2}); err != nil || !repo.refundCalled {
		t.Fatalf("refund delegation failed: err=%v called=%v", err, repo.refundCalled)
	}

	if err := command.ModifyRemarks([]int{1, 2}, "remark"); err != nil || !repo.modifyCalled {
		t.Fatalf("modify remarks delegation failed: err=%v called=%v", err, repo.modifyCalled)
	}

	synced, failed, err := syncSvc.ManualDock([]int{1})
	if synced != 1 || failed != 0 || err != nil || !repo.manualDockCalled {
		t.Fatalf("manual dock delegation failed: synced=%d failed=%d err=%v called=%v", synced, failed, err, repo.manualDockCalled)
	}

	progressCount, err := syncSvc.SyncProgress([]int{1})
	if progressCount != 1 || err != nil || !repo.syncProgressCalled {
		t.Fatalf("sync progress delegation failed: count=%d err=%v called=%v", progressCount, err, repo.syncProgressCalled)
	}

	autoUpdated, autoFailed, err := syncSvc.AutoSyncAllProgress(AutoSyncOptions{})
	if autoUpdated != 2 || autoFailed != 1 || err != nil || !repo.autoSyncCalled {
		t.Fatalf("auto sync delegation failed: updated=%d failed=%d err=%v called=%v", autoUpdated, autoFailed, err, repo.autoSyncCalled)
	}

	batchSynced, err := syncSvc.BatchSync([]int{1, 2})
	if batchSynced != 2 || err != nil || !repo.batchSyncCalled {
		t.Fatalf("batch sync delegation failed: count=%d err=%v called=%v", batchSynced, err, repo.batchSyncCalled)
	}

	success, fail, err := syncSvc.BatchResend([]int{1, 2})
	if success != 1 || fail != 1 || err == nil || !repo.batchResendCalled {
		t.Fatalf("sync delegation failed: success=%d fail=%d err=%v called=%v", success, fail, err, repo.batchResendCalled)
	}

	result, err = command.AddForMall(1, 2, 3, 9.9, model.OrderAddRequest{})
	if err != nil || result == nil || result.SuccessCount != 1 || !repo.addForMallCalled {
		t.Fatalf("mall command delegation failed: result=%v err=%v called=%v", result, err, repo.addForMallCalled)
	}
}

func TestLegacyRepositoryValidationGuards(t *testing.T) {
	repo := &legacyRepository{}

	t.Run("ChangeStatusRequiresOrders", func(t *testing.T) {
		err := repo.ChangeStatus(1, "1", model.OrderStatusRequest{Type: 1})
		if err == nil || err.Error() != "请选择订单" {
			t.Fatalf("expected empty order validation error, got %v", err)
		}
	})

	t.Run("ChangeStatusRejectsInvalidType", func(t *testing.T) {
		err := repo.ChangeStatus(1, "1", model.OrderStatusRequest{
			Type:   9,
			Status: "已完成",
			OIDs:   []int{1},
		})
		if err == nil || err.Error() != "无效操作类型" {
			t.Fatalf("expected invalid type error, got %v", err)
		}
	})

	t.Run("RefundOrdersRequiresOrders", func(t *testing.T) {
		err := repo.RefundOrders(1, "3", nil)
		if err == nil || err.Error() != "请选择订单" {
			t.Fatalf("expected empty refund selection error, got %v", err)
		}
	})

	t.Run("RefundOrdersRequiresPrivilegedGrade", func(t *testing.T) {
		err := repo.RefundOrders(1, "1", []int{1})
		if err == nil || err.Error() != "无权限" {
			t.Fatalf("expected permission error, got %v", err)
		}
	})

	t.Run("ModifyRemarksRequiresOrders", func(t *testing.T) {
		err := repo.ModifyRemarks(nil, "remark")
		if err == nil || err.Error() != "请选择订单" {
			t.Fatalf("expected empty remark selection error, got %v", err)
		}
	})

	t.Run("ManualDockRequiresOrders", func(t *testing.T) {
		success, fail, err := repo.ManualDockOrders(nil)
		if err == nil || err.Error() != "请选择订单" {
			t.Fatalf("expected empty manual dock selection error, got %v", err)
		}
		if success != 0 || fail != 0 {
			t.Fatalf("expected zero counts on validation failure, got success=%d fail=%d", success, fail)
		}
	})

	t.Run("SyncProgressRequiresOrders", func(t *testing.T) {
		updated, err := repo.SyncOrderProgress(nil)
		if err == nil || err.Error() != "请选择订单" {
			t.Fatalf("expected empty sync selection error, got %v", err)
		}
		if updated != 0 {
			t.Fatalf("expected zero updated count on validation failure, got %d", updated)
		}
	})

	t.Run("BatchSyncRequiresOrders", func(t *testing.T) {
		updated, err := repo.BatchSyncOrders(nil)
		if err == nil || err.Error() != "请选择要同步的订单" {
			t.Fatalf("expected empty batch sync selection error, got %v", err)
		}
		if updated != 0 {
			t.Fatalf("expected zero updated count on validation failure, got %d", updated)
		}
	})

	t.Run("BatchResendRequiresOrders", func(t *testing.T) {
		success, fail, err := repo.BatchResendOrders(nil)
		if err == nil || err.Error() != "请选择要补单的订单" {
			t.Fatalf("expected empty batch resend selection error, got %v", err)
		}
		if success != 0 || fail != 0 {
			t.Fatalf("expected zero counts on validation failure, got success=%d fail=%d", success, fail)
		}
	})
}

func TestLegacyRepositoryChangeStatusSideEffects(t *testing.T) {
	repo := &legacyRepository{}

	db, err := sql.Open(changeStatusTestDriverName, "")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	defer db.Close()

	originalDB := database.DB
	database.DB = db
	defer func() {
		database.DB = originalDB
		changeStatusExecHook = nil
		changeStatusQueryHook = nil
	}()

	originalNotifier := orderStatusNotifier
	defer func() {
		orderStatusNotifier = originalNotifier
	}()

	t.Run("StatusUpdateNotifiesEachOrderAndScopesToUID", func(t *testing.T) {
		execCalled := false
		changeStatusExecHook = func(query string, args []driver.NamedValue) error {
			execCalled = true
			if !strings.Contains(query, "UPDATE qingka_wangke_order SET status = ? WHERE oid IN (?,?)") {
				return fmt.Errorf("unexpected query: %s", query)
			}
			if !strings.Contains(query, "AND uid = ?") {
				return fmt.Errorf("expected uid scope in query: %s", query)
			}
			want := []string{"已完成", "10", "11", "9"}
			if got := namedValueStrings(args); !reflect.DeepEqual(got, want) {
				return fmt.Errorf("unexpected args: got=%v want=%v", got, want)
			}
			return nil
		}

		var notified []int
		SetOrderStatusNotifier(func(oid int, newStatus string, newProcess string, remarks string) {
			if newStatus != "已完成" || newProcess != "" || remarks != "" {
				t.Fatalf("unexpected notifier payload: oid=%d status=%q process=%q remarks=%q", oid, newStatus, newProcess, remarks)
			}
			notified = append(notified, oid)
		})

		err := repo.ChangeStatus(9, "1", model.OrderStatusRequest{
			Type:   1,
			Status: "已完成",
			OIDs:   []int{10, 11},
		})
		if err != nil {
			t.Fatalf("change status returned error: %v", err)
		}
		if !execCalled {
			t.Fatalf("expected exec to be called")
		}
		if want := []int{10, 11}; !reflect.DeepEqual(notified, want) {
			t.Fatalf("unexpected notified oids: got=%v want=%v", notified, want)
		}
	})

	t.Run("DockStatusUpdateSkipsNotifierForAdminGrade", func(t *testing.T) {
		execCalled := false
		changeStatusExecHook = func(query string, args []driver.NamedValue) error {
			execCalled = true
			if !strings.Contains(query, "UPDATE qingka_wangke_order SET dockstatus = ? WHERE oid IN (?)") {
				return fmt.Errorf("unexpected query: %s", query)
			}
			if strings.Contains(query, "AND uid = ?") {
				return fmt.Errorf("did not expect uid scope for admin grade query: %s", query)
			}
			want := []string{"处理中", "20"}
			if got := namedValueStrings(args); !reflect.DeepEqual(got, want) {
				return fmt.Errorf("unexpected args: got=%v want=%v", got, want)
			}
			return nil
		}

		notified := 0
		SetOrderStatusNotifier(func(oid int, newStatus string, newProcess string, remarks string) {
			notified++
		})

		err := repo.ChangeStatus(9, "3", model.OrderStatusRequest{
			Type:   2,
			Status: "处理中",
			OIDs:   []int{20},
		})
		if err != nil {
			t.Fatalf("dock status change returned error: %v", err)
		}
		if !execCalled {
			t.Fatalf("expected exec to be called")
		}
		if notified != 0 {
			t.Fatalf("did not expect notifier to run for dock status update, got %d calls", notified)
		}
	})
}

func TestLegacyRepositorySyncAndResendSkipTerminalOrders(t *testing.T) {
	repo := &legacyRepository{}

	db, err := sql.Open(changeStatusTestDriverName, "")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	defer db.Close()

	originalDB := database.DB
	database.DB = db
	defer func() {
		database.DB = originalDB
		changeStatusExecHook = nil
		changeStatusQueryHook = nil
	}()

	originalNotifier := orderStatusNotifier
	defer func() {
		orderStatusNotifier = originalNotifier
	}()

	t.Run("SyncProgressSkipsRefundedOrder", func(t *testing.T) {
		changeStatusQueryHook = func(query string, args []driver.NamedValue) (driver.Rows, error) {
			if !strings.Contains(query, "SELECT COALESCE(yid,''), COALESCE(hid,'0'), COALESCE(user,''), COALESCE(kcname,''), COALESCE(kcid,''), COALESCE(noun,''), COALESCE(status,'') FROM qingka_wangke_order WHERE oid = ?") {
				return nil, fmt.Errorf("unexpected query: %s", query)
			}
			if got := namedValueStrings(args); !reflect.DeepEqual(got, []string{"30"}) {
				return nil, fmt.Errorf("unexpected args: %v", got)
			}
			return singleRow(
				[]string{"yid", "hid", "user", "kcname", "kcid", "noun", "status"},
				"Y30", "8", "student", "course", "CID-30", "NOUN-30", "已退款",
			), nil
		}

		execCalls := 0
		changeStatusExecHook = func(query string, args []driver.NamedValue) error {
			execCalls++
			return nil
		}

		notifierCalls := 0
		SetOrderStatusNotifier(func(oid int, newStatus string, newProcess string, remarks string) {
			notifierCalls++
		})

		updated, err := repo.SyncOrderProgress([]int{30})
		if err != nil {
			t.Fatalf("sync progress returned error: %v", err)
		}
		if updated != 0 {
			t.Fatalf("expected refunded order to be skipped, got updated=%d", updated)
		}
		if execCalls != 0 {
			t.Fatalf("did not expect exec for refunded order, got %d calls", execCalls)
		}
		if notifierCalls != 0 {
			t.Fatalf("did not expect notifier for refunded order, got %d calls", notifierCalls)
		}
	})

	t.Run("BatchResendSkipsCancelledOrder", func(t *testing.T) {
		changeStatusQueryHook = func(query string, args []driver.NamedValue) (driver.Rows, error) {
			if !strings.Contains(query, "SELECT COALESCE(hid,0), COALESCE(yid,''), COALESCE(status,'') FROM qingka_wangke_order WHERE oid = ?") {
				return nil, fmt.Errorf("unexpected query: %s", query)
			}
			if got := namedValueStrings(args); !reflect.DeepEqual(got, []string{"31"}) {
				return nil, fmt.Errorf("unexpected args: %v", got)
			}
			return singleRow(
				[]string{"hid", "yid", "status"},
				int64(9), "Y31", "已取消",
			), nil
		}

		execCalls := 0
		changeStatusExecHook = func(query string, args []driver.NamedValue) error {
			execCalls++
			return nil
		}

		success, fail, err := repo.BatchResendOrders([]int{31})
		if err != nil {
			t.Fatalf("batch resend returned error: %v", err)
		}
		if success != 0 || fail != 1 {
			t.Fatalf("expected cancelled order to count as failure, got success=%d fail=%d", success, fail)
		}
		if execCalls != 0 {
			t.Fatalf("did not expect exec for cancelled order, got %d calls", execCalls)
		}
	})
}

func TestLegacyRepositoryBatchResendWritesRemarksWhenPlatformDoesNotSupportResubmit(t *testing.T) {
	repo := &legacyRepository{sup: suppliermodule.NewService()}

	db, err := sql.Open(changeStatusTestDriverName, "")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	defer db.Close()

	originalDB := database.DB
	database.DB = db
	defer func() {
		database.DB = originalDB
		changeStatusExecHook = nil
		changeStatusQueryHook = nil
	}()

	queryStep := 0
	changeStatusQueryHook = func(query string, args []driver.NamedValue) (driver.Rows, error) {
		queryStep++
		switch queryStep {
		case 1:
			if !strings.Contains(query, "SELECT COALESCE(hid,0), COALESCE(yid,''), COALESCE(status,'') FROM qingka_wangke_order WHERE oid = ?") {
				return nil, fmt.Errorf("unexpected order query: %s", query)
			}
			if got := namedValueStrings(args); !reflect.DeepEqual(got, []string{"40"}) {
				return nil, fmt.Errorf("unexpected order query args: %v", got)
			}
			return singleRow([]string{"hid", "yid", "status"}, int64(9), "Y40", "进行中"), nil
		case 2:
			if !strings.Contains(query, "SELECT hid, COALESCE(pt,''), COALESCE(name,''), COALESCE(url,''), COALESCE(user,''), COALESCE(pass,''), COALESCE(token,''), COALESCE(ip,''), COALESCE(cookie,''), COALESCE(money,'0'), COALESCE(status,'1') FROM qingka_wangke_huoyuan WHERE hid = ?") {
				return nil, fmt.Errorf("unexpected supplier query: %s", query)
			}
			if got := namedValueStrings(args); !reflect.DeepEqual(got, []string{"9"}) {
				return nil, fmt.Errorf("unexpected supplier query args: %v", got)
			}
			return singleRow(
				[]string{"hid", "pt", "name", "url", "user", "pass", "token", "ip", "cookie", "money", "status"},
				int64(9), "tuboshu", "Tuboshu", "http://example.test", "user", "pass", "", "", "", "0", "1",
			), nil
		default:
			return nil, fmt.Errorf("unexpected extra query: %s", query)
		}
	}

	execCalls := 0
	changeStatusExecHook = func(query string, args []driver.NamedValue) error {
		execCalls++
		if !strings.Contains(query, "UPDATE qingka_wangke_order SET remarks = ? WHERE oid = ?") {
			return fmt.Errorf("unexpected exec query: %s", query)
		}
		want := []string{"补单失败: 当前平台暂不支持补单操作", "40"}
		if got := namedValueStrings(args); !reflect.DeepEqual(got, want) {
			return fmt.Errorf("unexpected exec args: got=%v want=%v", got, want)
		}
		return nil
	}

	success, fail, err := repo.BatchResendOrders([]int{40})
	if err != nil {
		t.Fatalf("batch resend returned error: %v", err)
	}
	if success != 0 || fail != 1 {
		t.Fatalf("expected unsupported resubmit to count as one failure, got success=%d fail=%d", success, fail)
	}
	if execCalls != 1 {
		t.Fatalf("expected one remarks update, got %d", execCalls)
	}
}

func TestLegacyRepositoryCancelOrderOwnershipChecks(t *testing.T) {
	repo := &legacyRepository{}

	db, err := sql.Open(changeStatusTestDriverName, "")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	defer db.Close()

	originalDB := database.DB
	database.DB = db
	defer func() {
		database.DB = originalDB
		changeStatusExecHook = nil
		changeStatusQueryHook = nil
	}()

	t.Run("RejectsNonOwnerForNormalGrade", func(t *testing.T) {
		changeStatusQueryHook = func(query string, args []driver.NamedValue) (driver.Rows, error) {
			if !strings.Contains(query, "SELECT uid FROM qingka_wangke_order WHERE oid = ?") {
				return nil, fmt.Errorf("unexpected query: %s", query)
			}
			if got := namedValueStrings(args); !reflect.DeepEqual(got, []string{"50"}) {
				return nil, fmt.Errorf("unexpected query args: %v", got)
			}
			return singleRow([]string{"uid"}, int64(99)), nil
		}

		execCalls := 0
		changeStatusExecHook = func(query string, args []driver.NamedValue) error {
			execCalls++
			return nil
		}

		err := repo.CancelOrder(1, "1", 50)
		if err == nil || err.Error() != "无权限" {
			t.Fatalf("expected permission error, got %v", err)
		}
		if execCalls != 0 {
			t.Fatalf("did not expect cancel update for non-owner, got %d execs", execCalls)
		}
	})

	t.Run("AdminBypassesOwnershipCheck", func(t *testing.T) {
		changeStatusQueryHook = func(query string, args []driver.NamedValue) (driver.Rows, error) {
			t.Fatalf("did not expect ownership query for admin grade: %s", query)
			return nil, nil
		}

		execCalled := false
		changeStatusExecHook = func(query string, args []driver.NamedValue) error {
			execCalled = true
			if !strings.Contains(query, "UPDATE qingka_wangke_order SET status = '已取消', dockstatus = '4' WHERE oid = ?") {
				return fmt.Errorf("unexpected exec query: %s", query)
			}
			if got := namedValueStrings(args); !reflect.DeepEqual(got, []string{"51"}) {
				return fmt.Errorf("unexpected exec args: %v", got)
			}
			return nil
		}

		err := repo.CancelOrder(1, "3", 51)
		if err != nil {
			t.Fatalf("admin cancel returned error: %v", err)
		}
		if !execCalled {
			t.Fatal("expected cancel update to execute for admin")
		}
	})
}

func TestLegacyRepositoryAddOrderEarlyValidation(t *testing.T) {
	t.Run("AddOrdersReturnsSupplierClassError", func(t *testing.T) {
		repo := &legacyRepository{
			sup: &stubSupplierGateway{
				getClassFull: func(cid int) (*model.ClassFull, error) {
					if cid != 70 {
						t.Fatalf("unexpected cid: %d", cid)
					}
					return nil, errors.New("class lookup failed")
				},
			},
		}

		result, err := repo.AddOrders(1, model.OrderAddRequest{CID: 70})
		if err == nil || err.Error() != "class lookup failed" {
			t.Fatalf("expected class lookup error, got result=%v err=%v", result, err)
		}
		if result != nil {
			t.Fatalf("expected nil result on lookup error, got %v", result)
		}
	})

	t.Run("AddOrdersRejectsDisabledClass", func(t *testing.T) {
		repo := &legacyRepository{
			sup: &stubSupplierGateway{
				getClassFull: func(cid int) (*model.ClassFull, error) {
					return &model.ClassFull{CID: cid, Status: 0}, nil
				},
			},
		}

		result, err := repo.AddOrders(1, model.OrderAddRequest{CID: 71})
		if err == nil || err.Error() != "课程已下架" {
			t.Fatalf("expected disabled class error, got result=%v err=%v", result, err)
		}
		if result != nil {
			t.Fatalf("expected nil result for disabled class, got %v", result)
		}
	})

	t.Run("AddOrdersForMallReturnsSupplierClassError", func(t *testing.T) {
		repo := &legacyRepository{
			sup: &stubSupplierGateway{
				getClassFull: func(cid int) (*model.ClassFull, error) {
					if cid != 72 {
						t.Fatalf("unexpected cid: %d", cid)
					}
					return nil, errors.New("mall class lookup failed")
				},
			},
		}

		result, err := repo.AddOrdersForMall(1, 2, 3, 9.9, model.OrderAddRequest{CID: 72})
		if err == nil || err.Error() != "mall class lookup failed" {
			t.Fatalf("expected mall class lookup error, got result=%v err=%v", result, err)
		}
		if result != nil {
			t.Fatalf("expected nil result on mall lookup error, got %v", result)
		}
	})

	t.Run("AddOrdersForMallRejectsDisabledClass", func(t *testing.T) {
		repo := &legacyRepository{
			sup: &stubSupplierGateway{
				getClassFull: func(cid int) (*model.ClassFull, error) {
					return &model.ClassFull{CID: cid, Status: 0}, nil
				},
			},
		}

		result, err := repo.AddOrdersForMall(1, 2, 3, 9.9, model.OrderAddRequest{CID: 73})
		if err == nil || err.Error() != "课程已下架" {
			t.Fatalf("expected disabled mall class error, got result=%v err=%v", result, err)
		}
		if result != nil {
			t.Fatalf("expected nil result for disabled mall class, got %v", result)
		}
	})
}

func TestLegacyRepositoryManualDockAndSyncSuccessPaths(t *testing.T) {
	db, err := sql.Open(changeStatusTestDriverName, "")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	defer db.Close()

	originalDB := database.DB
	database.DB = db
	defer func() {
		database.DB = originalDB
		changeStatusExecHook = nil
		changeStatusQueryHook = nil
	}()

	t.Run("ManualDockMarksOrderDockedOnSuccess", func(t *testing.T) {
		repo := &legacyRepository{
			sup: &stubSupplierGateway{
				getClassFull: func(cid int) (*model.ClassFull, error) {
					if cid != 101 {
						t.Fatalf("unexpected class cid: %d", cid)
					}
					return &model.ClassFull{CID: 101, Docking: "9"}, nil
				},
				getSupplierByHID: func(hid int) (*model.SupplierFull, error) {
					if hid != 9 {
						t.Fatalf("unexpected supplier hid: %d", hid)
					}
					return &model.SupplierFull{HID: 9}, nil
				},
				callOrder: func(sup *model.SupplierFull, cls *model.ClassFull, school, user, pass, kcid, kcname string, extraFields map[string]string) (*model.SupplierOrderResult, error) {
					if school != "school" || user != "user" || pass != "pass" || kcid != "KC1" || kcname != "课程A" {
						t.Fatalf("unexpected dock params: school=%q user=%q pass=%q kcid=%q kcname=%q", school, user, pass, kcid, kcname)
					}
					return &model.SupplierOrderResult{Code: 1, YID: "Y60"}, nil
				},
			},
		}

		changeStatusQueryHook = func(query string, args []driver.NamedValue) (driver.Rows, error) {
			if !strings.Contains(query, "SELECT cid, COALESCE(hid,0), COALESCE(school,''), COALESCE(user,''), COALESCE(pass,''), COALESCE(kcid,''), COALESCE(kcname,'') FROM qingka_wangke_order WHERE oid = ?") {
				return nil, fmt.Errorf("unexpected query: %s", query)
			}
			if got := namedValueStrings(args); !reflect.DeepEqual(got, []string{"60"}) {
				return nil, fmt.Errorf("unexpected query args: %v", got)
			}
			return singleRow([]string{"cid", "hid", "school", "user", "pass", "kcid", "kcname"}, int64(101), int64(0), "school", "user", "pass", "KC1", "课程A"), nil
		}

		execCalled := false
		changeStatusExecHook = func(query string, args []driver.NamedValue) error {
			execCalled = true
			if !strings.Contains(query, "UPDATE qingka_wangke_order SET dockstatus = 1, yid = ?, hid = ?, status = '进行中' WHERE oid = ?") {
				return fmt.Errorf("unexpected exec query: %s", query)
			}
			if got := namedValueStrings(args); !reflect.DeepEqual(got, []string{"Y60", "9", "60"}) {
				return fmt.Errorf("unexpected exec args: %v", got)
			}
			return nil
		}

		success, fail, err := repo.ManualDockOrders([]int{60})
		if err != nil {
			t.Fatalf("manual dock returned error: %v", err)
		}
		if success != 1 || fail != 0 {
			t.Fatalf("unexpected dock counts: success=%d fail=%d", success, fail)
		}
		if !execCalled {
			t.Fatal("expected dock success update to execute")
		}
	})

	t.Run("ManualDockWritesRemarksWhenClassMissing", func(t *testing.T) {
		repo := &legacyRepository{
			sup: &stubSupplierGateway{
				getClassFull: func(cid int) (*model.ClassFull, error) {
					return nil, errors.New("class missing")
				},
				getSupplierByHID: func(hid int) (*model.SupplierFull, error) {
					t.Fatal("did not expect supplier lookup when class is missing")
					return nil, nil
				},
				callOrder: func(sup *model.SupplierFull, cls *model.ClassFull, school, user, pass, kcid, kcname string, extraFields map[string]string) (*model.SupplierOrderResult, error) {
					t.Fatal("did not expect supplier order call when class is missing")
					return nil, nil
				},
			},
		}

		changeStatusQueryHook = func(query string, args []driver.NamedValue) (driver.Rows, error) {
			return singleRow([]string{"cid", "hid", "school", "user", "pass", "kcid", "kcname"}, int64(102), int64(0), "school", "user", "pass", "KC2", "课程B"), nil
		}

		execCalled := false
		changeStatusExecHook = func(query string, args []driver.NamedValue) error {
			execCalled = true
			if !strings.Contains(query, "UPDATE qingka_wangke_order SET dockstatus = 2, remarks = ? WHERE oid = ?") {
				return fmt.Errorf("unexpected exec query: %s", query)
			}
			if got := namedValueStrings(args); !reflect.DeepEqual(got, []string{"课程不存在: class missing", "61"}) {
				return fmt.Errorf("unexpected exec args: %v", got)
			}
			return nil
		}

		success, fail, err := repo.ManualDockOrders([]int{61})
		if err != nil {
			t.Fatalf("manual dock returned error: %v", err)
		}
		if success != 0 || fail != 1 {
			t.Fatalf("unexpected dock counts: success=%d fail=%d", success, fail)
		}
		if !execCalled {
			t.Fatal("expected class missing remarks update to execute")
		}
	})

	t.Run("ManualDockWritesRemarksWhenSupplierMissing", func(t *testing.T) {
		repo := &legacyRepository{
			sup: &stubSupplierGateway{
				getClassFull: func(cid int) (*model.ClassFull, error) {
					return &model.ClassFull{CID: cid, Docking: "9"}, nil
				},
				getSupplierByHID: func(hid int) (*model.SupplierFull, error) {
					if hid != 9 {
						t.Fatalf("unexpected supplier hid: %d", hid)
					}
					return nil, errors.New("supplier missing")
				},
				callOrder: func(sup *model.SupplierFull, cls *model.ClassFull, school, user, pass, kcid, kcname string, extraFields map[string]string) (*model.SupplierOrderResult, error) {
					t.Fatal("did not expect supplier order call when supplier is missing")
					return nil, nil
				},
			},
		}

		changeStatusQueryHook = func(query string, args []driver.NamedValue) (driver.Rows, error) {
			return singleRow([]string{"cid", "hid", "school", "user", "pass", "kcid", "kcname"}, int64(104), int64(0), "school", "user", "pass", "KC4", "课程D"), nil
		}

		execCalled := false
		changeStatusExecHook = func(query string, args []driver.NamedValue) error {
			execCalled = true
			if !strings.Contains(query, "UPDATE qingka_wangke_order SET dockstatus = 2, remarks = ? WHERE oid = ?") {
				return fmt.Errorf("unexpected exec query: %s", query)
			}
			if got := namedValueStrings(args); !reflect.DeepEqual(got, []string{"供应商不存在: supplier missing", "66"}) {
				return fmt.Errorf("unexpected exec args: %v", got)
			}
			return nil
		}

		success, fail, err := repo.ManualDockOrders([]int{66})
		if err != nil {
			t.Fatalf("manual dock returned error: %v", err)
		}
		if success != 0 || fail != 1 {
			t.Fatalf("unexpected dock counts: success=%d fail=%d", success, fail)
		}
		if !execCalled {
			t.Fatal("expected supplier missing remarks update to execute")
		}
	})

	t.Run("ManualDockMarksUndockedWhenClassHasNoDocking", func(t *testing.T) {
		repo := &legacyRepository{
			sup: &stubSupplierGateway{
				getClassFull: func(cid int) (*model.ClassFull, error) {
					return &model.ClassFull{CID: cid, Docking: "0"}, nil
				},
				getSupplierByHID: func(hid int) (*model.SupplierFull, error) {
					t.Fatal("did not expect supplier lookup when class docking is zero")
					return nil, nil
				},
				callOrder: func(sup *model.SupplierFull, cls *model.ClassFull, school, user, pass, kcid, kcname string, extraFields map[string]string) (*model.SupplierOrderResult, error) {
					t.Fatal("did not expect supplier order call when class docking is zero")
					return nil, nil
				},
			},
		}

		changeStatusQueryHook = func(query string, args []driver.NamedValue) (driver.Rows, error) {
			return singleRow([]string{"cid", "hid", "school", "user", "pass", "kcid", "kcname"}, int64(103), int64(0), "school", "user", "pass", "KC3", "课程C"), nil
		}

		execCalled := false
		changeStatusExecHook = func(query string, args []driver.NamedValue) error {
			execCalled = true
			if !strings.Contains(query, "UPDATE qingka_wangke_order SET dockstatus = 99 WHERE oid = ?") {
				return fmt.Errorf("unexpected exec query: %s", query)
			}
			if got := namedValueStrings(args); !reflect.DeepEqual(got, []string{"64"}) {
				return fmt.Errorf("unexpected exec args: %v", got)
			}
			return nil
		}

		success, fail, err := repo.ManualDockOrders([]int{64})
		if err != nil {
			t.Fatalf("manual dock returned error: %v", err)
		}
		if success != 0 || fail != 1 {
			t.Fatalf("unexpected dock counts: success=%d fail=%d", success, fail)
		}
		if !execCalled {
			t.Fatal("expected dockstatus 99 update to execute")
		}
	})

	t.Run("ManualDockWritesRemarksWhenSupplierReturnsFailureMessage", func(t *testing.T) {
		repo := &legacyRepository{
			sup: &stubSupplierGateway{
				getClassFull: func(cid int) (*model.ClassFull, error) {
					return &model.ClassFull{CID: cid, Docking: "9"}, nil
				},
				getSupplierByHID: func(hid int) (*model.SupplierFull, error) {
					return &model.SupplierFull{HID: hid}, nil
				},
				callOrder: func(sup *model.SupplierFull, cls *model.ClassFull, school, user, pass, kcid, kcname string, extraFields map[string]string) (*model.SupplierOrderResult, error) {
					return &model.SupplierOrderResult{Code: 2, Msg: "quota exceeded"}, nil
				},
			},
		}

		changeStatusQueryHook = func(query string, args []driver.NamedValue) (driver.Rows, error) {
			return singleRow([]string{"cid", "hid", "school", "user", "pass", "kcid", "kcname"}, int64(105), int64(0), "school", "user", "pass", "KC5", "课程E"), nil
		}

		execCalled := false
		changeStatusExecHook = func(query string, args []driver.NamedValue) error {
			execCalled = true
			if !strings.Contains(query, "UPDATE qingka_wangke_order SET dockstatus = 2, remarks = ? WHERE oid = ?") {
				return fmt.Errorf("unexpected exec query: %s", query)
			}
			if got := namedValueStrings(args); !reflect.DeepEqual(got, []string{"对接失败: quota exceeded", "68"}) {
				return fmt.Errorf("unexpected exec args: %v", got)
			}
			return nil
		}

		success, fail, err := repo.ManualDockOrders([]int{68})
		if err != nil {
			t.Fatalf("manual dock returned error: %v", err)
		}
		if success != 0 || fail != 1 {
			t.Fatalf("unexpected dock counts: success=%d fail=%d", success, fail)
		}
		if !execCalled {
			t.Fatal("expected supplier failure remarks update to execute")
		}
	})

	t.Run("SyncProgressUpdatesOrderAndNotifies", func(t *testing.T) {
		repo := &legacyRepository{
			sup: &stubSupplierGateway{
				hasBatchProgressForPT: func(pt string) bool {
					return false
				},
				getSupplierByHID: func(hid int) (*model.SupplierFull, error) {
					if hid != 9 {
						t.Fatalf("unexpected supplier hid: %d", hid)
					}
					return &model.SupplierFull{HID: 9}, nil
				},
				queryProgress: func(sup *model.SupplierFull, yid string, username string, orderExtra map[string]string) ([]model.SupplierProgressItem, error) {
					if yid != "Y62" || username != "user" || orderExtra["kcname"] != "课程C" || orderExtra["noun"] != "NOUN-62" || orderExtra["kcid"] != "KC-62" {
						t.Fatalf("unexpected progress params: yid=%q username=%q extra=%v", yid, username, orderExtra)
					}
					return []model.SupplierProgressItem{{
						KCName:          "课程C",
						YID:             "Y62",
						StatusText:      "已完成",
						Process:         "100%",
						Remarks:         "done",
						User:            "user",
						CourseStartTime: "2026-03-01",
						CourseEndTime:   "2026-03-10",
						ExamStartTime:   "2026-03-11",
						ExamEndTime:     "2026-03-12",
					}}, nil
				},
			},
		}

		changeStatusQueryHook = func(query string, args []driver.NamedValue) (driver.Rows, error) {
			if !strings.Contains(query, "SELECT COALESCE(yid,''), COALESCE(hid,'0'), COALESCE(user,''), COALESCE(kcname,''), COALESCE(kcid,''), COALESCE(noun,''), COALESCE(status,'') FROM qingka_wangke_order WHERE oid = ?") {
				return nil, fmt.Errorf("unexpected query: %s", query)
			}
			return singleRow([]string{"yid", "hid", "user", "kcname", "kcid", "noun", "status"}, "Y62", "9", "user", "课程C", "KC-62", "NOUN-62", "进行中"), nil
		}

		execCalled := false
		changeStatusExecHook = func(query string, args []driver.NamedValue) error {
			execCalled = true
			if !strings.Contains(query, "UPDATE qingka_wangke_order SET name = ?, yid = ?, status = ?, process = ?, remarks = ?, courseStartTime = ?, courseEndTime = ?, examStartTime = ?, examEndTime = ? WHERE user = ? AND kcname = ? AND oid = ?") {
				return fmt.Errorf("unexpected exec query: %s", query)
			}
			want := []string{"课程C", "Y62", "已完成", "100%", "done", "2026-03-01", "2026-03-10", "2026-03-11", "2026-03-12", "user", "课程C", "62"}
			if got := namedValueStrings(args); !reflect.DeepEqual(got, want) {
				return fmt.Errorf("unexpected exec args: got=%v want=%v", got, want)
			}
			return nil
		}

		var notified []string
		originalNotifier := orderStatusNotifier
		defer func() { orderStatusNotifier = originalNotifier }()
		SetOrderStatusNotifier(func(oid int, newStatus string, newProcess string, remarks string) {
			notified = append(notified, fmt.Sprintf("%d|%s|%s|%s", oid, newStatus, newProcess, remarks))
		})

		updated, err := repo.SyncOrderProgress([]int{62})
		if err != nil {
			t.Fatalf("sync progress returned error: %v", err)
		}
		if updated != 1 {
			t.Fatalf("expected one updated order, got %d", updated)
		}
		if !execCalled {
			t.Fatal("expected sync progress update to execute")
		}
		if want := []string{"62|已完成|100%|done"}; !reflect.DeepEqual(notified, want) {
			t.Fatalf("unexpected notifier calls: got=%v want=%v", notified, want)
		}
	})

	t.Run("AutoSyncAllProgressUpdatesDockedOrders", func(t *testing.T) {
		repo := &legacyRepository{
			sup: &stubSupplierGateway{
				hasBatchProgressForPT: func(pt string) bool {
					return pt == "batch-progress"
				},
				getSupplierByHID: func(hid int) (*model.SupplierFull, error) {
					if hid != 9 {
						t.Fatalf("unexpected supplier hid: %d", hid)
					}
					return &model.SupplierFull{HID: 9}, nil
				},
				queryProgress: func(sup *model.SupplierFull, yid string, username string, orderExtra map[string]string) ([]model.SupplierProgressItem, error) {
					if yid != "Y70" || username != "student" || orderExtra["kcname"] != "课程D" || orderExtra["noun"] != "NOUN-70" || orderExtra["kcid"] != "KC-70" {
						t.Fatalf("unexpected progress params: yid=%q username=%q extra=%v", yid, username, orderExtra)
					}
					return []model.SupplierProgressItem{{
						KCName:          "课程D",
						YID:             "Y70",
						StatusText:      "进行中",
						Process:         "80%",
						Remarks:         "同步中",
						User:            "student",
						CourseStartTime: "2026-03-01",
						CourseEndTime:   "2026-03-08",
						ExamStartTime:   "2026-03-09",
						ExamEndTime:     "2026-03-10",
					}}, nil
				},
			},
		}

		changeStatusQueryHook = func(query string, args []driver.NamedValue) (driver.Rows, error) {
			if !strings.Contains(query, "FROM qingka_wangke_order") || !strings.Contains(query, "WHERE dockstatus = 1") {
				return nil, fmt.Errorf("unexpected query: %s", query)
			}
			return &testRows{
				columns: []string{"oid", "yid", "hid", "pt", "user", "kcname", "noun", "kcid", "status", "addtime", "updatetime"},
				values:  [][]driver.Value{{int64(70), "Y70", "9", "single-progress", "student", "课程D", "NOUN-70", "KC-70", "进行中", "2026-03-01 00:00:00", ""}},
			}, nil
		}

		execCalled := false
		changeStatusExecHook = func(query string, args []driver.NamedValue) error {
			execCalled = true
			if !strings.Contains(query, "UPDATE qingka_wangke_order SET name = ?, yid = ?, status = ?, process = ?, remarks = ?, courseStartTime = ?, courseEndTime = ?, examStartTime = ?, examEndTime = ?, updatetime = ? WHERE oid = ?") {
				return fmt.Errorf("unexpected exec query: %s", query)
			}
			got := namedValueStrings(args)
			wantPrefix := []string{"课程D", "Y70", "进行中", "80%", "同步中", "2026-03-01", "2026-03-08", "2026-03-09", "2026-03-10"}
			if len(got) != 11 || !reflect.DeepEqual(got[:9], wantPrefix) || got[9] == "" || got[10] != "70" {
				return fmt.Errorf("unexpected exec args: got=%v", got)
			}
			return nil
		}

		var notified []string
		originalNotifier := orderStatusNotifier
		defer func() { orderStatusNotifier = originalNotifier }()
		SetOrderStatusNotifier(func(oid int, newStatus string, newProcess string, remarks string) {
			notified = append(notified, fmt.Sprintf("%d|%s|%s|%s", oid, newStatus, newProcess, remarks))
		})

		updated, failed, err := repo.AutoSyncAllProgress(AutoSyncOptions{})
		if err != nil {
			t.Fatalf("auto sync returned error: %v", err)
		}
		if updated != 1 || failed != 0 {
			t.Fatalf("unexpected auto sync counts: updated=%d failed=%d", updated, failed)
		}
		if !execCalled {
			t.Fatal("expected auto sync update to execute")
		}
		if want := []string{"70|进行中|80%|同步中"}; !reflect.DeepEqual(notified, want) {
			t.Fatalf("unexpected notifier calls: got=%v want=%v", notified, want)
		}
	})

	t.Run("AutoSyncAllProgressUsesBatchProgressFeedWhenConfigured", func(t *testing.T) {
		repo := &legacyRepository{
			sup: &stubSupplierGateway{
				hasBatchProgressForPT: func(pt string) bool {
					return pt == "batch-progress"
				},
				getSupplierByHID: func(hid int) (*model.SupplierFull, error) {
					if hid != 9 {
						t.Fatalf("unexpected supplier hid: %d", hid)
					}
					return &model.SupplierFull{HID: 9, PT: "batch-progress"}, nil
				},
				hasBatchProgressConfig: func(sup *model.SupplierFull) bool {
					return sup.PT == "batch-progress"
				},
				queryBatchProgress: func(sup *model.SupplierFull, refs []model.SupplierBatchProgressRef) ([]model.SupplierProgressItem, error) {
					if len(refs) != 2 {
						t.Fatalf("unexpected batch refs length: %d", len(refs))
					}
					if refs[0].YID != "Y71" || refs[1].YID != "Y72" {
						t.Fatalf("unexpected batch refs: %+v", refs)
					}
					return []model.SupplierProgressItem{{
						YID:        "Y71",
						StatusText: "已完成",
						Process:    "100%",
						Remarks:    "批量完成",
					}}, nil
				},
				queryProgress: func(sup *model.SupplierFull, yid string, username string, orderExtra map[string]string) ([]model.SupplierProgressItem, error) {
					t.Fatalf("expected batch progress path, got single query for yid=%s", yid)
					return nil, nil
				},
			},
		}

		changeStatusQueryHook = func(query string, args []driver.NamedValue) (driver.Rows, error) {
			if !strings.Contains(query, "FROM qingka_wangke_order") || !strings.Contains(query, "WHERE dockstatus = 1") {
				return nil, fmt.Errorf("unexpected query: %s", query)
			}
			return &testRows{
				columns: []string{"oid", "yid", "hid", "pt", "user", "kcname", "noun", "kcid", "status", "addtime", "updatetime"},
				values: [][]driver.Value{
					{int64(71), "Y71", "9", "batch-progress", "student-a", "课程E", "NOUN-71", "KC-71", "进行中", "2026-03-01 00:00:00", ""},
					{int64(72), "Y72", "9", "batch-progress", "student-b", "课程F", "NOUN-72", "KC-72", "进行中", "2026-03-01 00:00:00", ""},
				},
			}, nil
		}

		var execQueries []string
		var execArgs [][]string
		changeStatusExecHook = func(query string, args []driver.NamedValue) error {
			execQueries = append(execQueries, query)
			execArgs = append(execArgs, namedValueStrings(args))
			return nil
		}

		var notified []string
		originalNotifier := orderStatusNotifier
		defer func() { orderStatusNotifier = originalNotifier }()
		SetOrderStatusNotifier(func(oid int, newStatus string, newProcess string, remarks string) {
			notified = append(notified, fmt.Sprintf("%d|%s|%s|%s", oid, newStatus, newProcess, remarks))
		})

		updated, failed, err := repo.AutoSyncAllProgress(AutoSyncOptions{})
		if err != nil {
			t.Fatalf("auto sync returned error: %v", err)
		}
		if updated != 1 || failed != 0 {
			t.Fatalf("unexpected auto sync counts: updated=%d failed=%d", updated, failed)
		}
		if len(execQueries) != 2 {
			t.Fatalf("expected two execs, got %d", len(execQueries))
		}
		if !strings.Contains(execQueries[0], "UPDATE qingka_wangke_order SET name = ?, yid = ?, status = ?, process = ?, remarks = ?, courseStartTime = ?, courseEndTime = ?, examStartTime = ?, examEndTime = ?, updatetime = ? WHERE oid = ?") {
			t.Fatalf("unexpected first exec query: %s", execQueries[0])
		}
		if want := []string{"课程E", "Y71", "已完成", "100%", "批量完成"}; !reflect.DeepEqual(execArgs[0][:5], want) {
			t.Fatalf("unexpected first exec args: %v", execArgs[0])
		}
		if got := execArgs[0][10]; got != "71" {
			t.Fatalf("unexpected updated oid: %s", got)
		}
		if !strings.Contains(execQueries[1], "UPDATE qingka_wangke_order SET updatetime = ? WHERE oid = ?") {
			t.Fatalf("unexpected second exec query: %s", execQueries[1])
		}
		if len(execArgs[1]) != 2 || execArgs[1][1] != "72" {
			t.Fatalf("unexpected touch args: %v", execArgs[1])
		}
		if want := []string{"71|已完成|100%|批量完成"}; !reflect.DeepEqual(notified, want) {
			t.Fatalf("unexpected notifier calls: got=%v want=%v", notified, want)
		}
	})

	t.Run("AutoSyncAllProgressBatchEmptyResultIsNotFailure", func(t *testing.T) {
		repo := &legacyRepository{
			sup: &stubSupplierGateway{
				hasBatchProgressForPT: func(pt string) bool {
					return pt == "batch-progress"
				},
				getSupplierByHID: func(hid int) (*model.SupplierFull, error) {
					if hid != 9 {
						t.Fatalf("unexpected supplier hid: %d", hid)
					}
					return &model.SupplierFull{HID: 9, PT: "batch-progress"}, nil
				},
				hasBatchProgressConfig: func(sup *model.SupplierFull) bool {
					return sup.PT == "batch-progress"
				},
				queryBatchProgress: func(sup *model.SupplierFull, refs []model.SupplierBatchProgressRef) ([]model.SupplierProgressItem, error) {
					return []model.SupplierProgressItem{}, nil
				},
				queryProgress: func(sup *model.SupplierFull, yid string, username string, orderExtra map[string]string) ([]model.SupplierProgressItem, error) {
					t.Fatal("expected batch-only path without single-order fallback")
					return nil, nil
				},
			},
		}

		changeStatusQueryHook = func(query string, args []driver.NamedValue) (driver.Rows, error) {
			if !strings.Contains(query, "FROM qingka_wangke_order") || !strings.Contains(query, "WHERE dockstatus = 1") {
				return nil, fmt.Errorf("unexpected query: %s", query)
			}
			return &testRows{
				columns: []string{"oid", "yid", "hid", "pt", "user", "kcname", "noun", "kcid", "status", "addtime", "updatetime"},
				values:  [][]driver.Value{{int64(73), "Y73", "9", "batch-progress", "student-c", "课程G", "NOUN-73", "KC-73", "进行中", "2026-03-01 00:00:00", ""}},
			}, nil
		}

		execCalled := false
		changeStatusExecHook = func(query string, args []driver.NamedValue) error {
			execCalled = true
			return nil
		}

		updated, failed, err := repo.AutoSyncAllProgress(AutoSyncOptions{OnlyBatchSuppliers: true, IgnoreRules: true})
		if err != nil {
			t.Fatalf("auto sync returned error: %v", err)
		}
		if updated != 0 || failed != 0 {
			t.Fatalf("unexpected auto sync counts: updated=%d failed=%d", updated, failed)
		}
		if execCalled {
			t.Fatal("expected no database update for empty batch result")
		}
	})

	t.Run("AutoSyncAllProgressMatchesBatchProgressByNounUserAndKCName", func(t *testing.T) {
		repo := &legacyRepository{
			sup: &stubSupplierGateway{
				getSupplierByHID: func(hid int) (*model.SupplierFull, error) {
					if hid != 9 {
						t.Fatalf("unexpected supplier hid: %d", hid)
					}
					return &model.SupplierFull{HID: 9, PT: "batch-progress"}, nil
				},
				hasBatchProgressConfig: func(sup *model.SupplierFull) bool {
					return sup.PT == "batch-progress"
				},
				queryBatchProgress: func(sup *model.SupplierFull, refs []model.SupplierBatchProgressRef) ([]model.SupplierProgressItem, error) {
					return []model.SupplierProgressItem{{
						Noun:    "NOUN-81",
						User:    "student-c",
						KCName:  "[2] 改革开放史♢山东·菏泽市",
						Status:  "平时分",
						Process: "68.6%",
						Remarks: "今日已完成",
					}}, nil
				},
			},
		}

		changeStatusQueryHook = func(query string, args []driver.NamedValue) (driver.Rows, error) {
			if !strings.Contains(query, "FROM qingka_wangke_order") || !strings.Contains(query, "WHERE dockstatus = 1") {
				return nil, fmt.Errorf("unexpected query: %s", query)
			}
			return &testRows{
				columns: []string{"oid", "yid", "hid", "pt", "user", "kcname", "noun", "kcid", "status", "addtime", "updatetime"},
				values:  [][]driver.Value{{int64(81), "", "9", "batch-progress", "student-c", "[2] 改革开放史♢山东·菏泽市", "NOUN-81", "KC-81", "进行中", "2026-03-01 00:00:00", ""}},
			}, nil
		}

		execCalled := false
		changeStatusExecHook = func(query string, args []driver.NamedValue) error {
			execCalled = true
			if !strings.Contains(query, "UPDATE qingka_wangke_order SET name = ?, yid = ?, status = ?, process = ?, remarks = ?, courseStartTime = ?, courseEndTime = ?, examStartTime = ?, examEndTime = ?, updatetime = ? WHERE oid = ?") {
				return fmt.Errorf("unexpected exec query: %s", query)
			}
			got := namedValueStrings(args)
			if len(got) != 11 || got[0] != "[2] 改革开放史♢山东·菏泽市" || got[2] != "平时分" || got[3] != "68.6%" || got[4] != "今日已完成" || got[10] != "81" {
				return fmt.Errorf("unexpected exec args: %v", got)
			}
			return nil
		}

		updated, failed, err := repo.AutoSyncAllProgress(AutoSyncOptions{})
		if err != nil {
			t.Fatalf("auto sync returned error: %v", err)
		}
		if updated != 1 || failed != 0 {
			t.Fatalf("unexpected auto sync counts: updated=%d failed=%d", updated, failed)
		}
		if !execCalled {
			t.Fatal("expected batch noun+user+kcname match to update order")
		}
	})

	t.Run("BatchResendSuccessUpdatesStatus", func(t *testing.T) {
		repo := &legacyRepository{
			sup: &stubSupplierGateway{
				getSupplierByHID: func(hid int) (*model.SupplierFull, error) {
					return &model.SupplierFull{HID: hid}, nil
				},
				resubmitOrder: func(sup *model.SupplierFull, yid string) (int, string, error) {
					if yid != "Y63" {
						t.Fatalf("unexpected yid: %s", yid)
					}
					return 1, "ok", nil
				},
			},
		}

		changeStatusQueryHook = func(query string, args []driver.NamedValue) (driver.Rows, error) {
			if !strings.Contains(query, "SELECT COALESCE(hid,0), COALESCE(yid,''), COALESCE(status,'') FROM qingka_wangke_order WHERE oid = ?") {
				return nil, fmt.Errorf("unexpected query: %s", query)
			}
			return singleRow([]string{"hid", "yid", "status"}, int64(9), "Y63", "进行中"), nil
		}

		execCalled := false
		changeStatusExecHook = func(query string, args []driver.NamedValue) error {
			execCalled = true
			if !strings.Contains(query, "UPDATE qingka_wangke_order SET status = '补刷中', dockstatus = 1, remarks = ?, bsnum = bsnum + 1 WHERE oid = ?") {
				return fmt.Errorf("unexpected exec query: %s", query)
			}
			got := namedValueStrings(args)
			if len(got) != 2 || got[1] != "63" || !strings.Contains(got[0], "补刷成功") {
				return fmt.Errorf("unexpected exec args: %v", got)
			}
			return nil
		}

		success, fail, err := repo.BatchResendOrders([]int{63})
		if err != nil {
			t.Fatalf("batch resend returned error: %v", err)
		}
		if success != 1 || fail != 0 {
			t.Fatalf("unexpected resend counts: success=%d fail=%d", success, fail)
		}
		if !execCalled {
			t.Fatal("expected batch resend success update to execute")
		}
	})

	t.Run("BatchResendWritesRemarksWhenResubmitErrors", func(t *testing.T) {
		repo := &legacyRepository{
			sup: &stubSupplierGateway{
				getSupplierByHID: func(hid int) (*model.SupplierFull, error) {
					return &model.SupplierFull{HID: hid}, nil
				},
				resubmitOrder: func(sup *model.SupplierFull, yid string) (int, string, error) {
					if yid != "Y67" {
						t.Fatalf("unexpected yid: %s", yid)
					}
					return 0, "", errors.New("platform timeout")
				},
			},
		}

		changeStatusQueryHook = func(query string, args []driver.NamedValue) (driver.Rows, error) {
			if !strings.Contains(query, "SELECT COALESCE(hid,0), COALESCE(yid,''), COALESCE(status,'') FROM qingka_wangke_order WHERE oid = ?") {
				return nil, fmt.Errorf("unexpected query: %s", query)
			}
			return singleRow([]string{"hid", "yid", "status"}, int64(9), "Y67", "进行中"), nil
		}

		execCalled := false
		changeStatusExecHook = func(query string, args []driver.NamedValue) error {
			execCalled = true
			if !strings.Contains(query, "UPDATE qingka_wangke_order SET remarks = ? WHERE oid = ?") {
				return fmt.Errorf("unexpected exec query: %s", query)
			}
			if got := namedValueStrings(args); !reflect.DeepEqual(got, []string{"补单失败: platform timeout", "67"}) {
				return fmt.Errorf("unexpected exec args: %v", got)
			}
			return nil
		}

		success, fail, err := repo.BatchResendOrders([]int{67})
		if err != nil {
			t.Fatalf("batch resend returned error: %v", err)
		}
		if success != 0 || fail != 1 {
			t.Fatalf("unexpected resend counts: success=%d fail=%d", success, fail)
		}
		if !execCalled {
			t.Fatal("expected resend failure remarks update to execute")
		}
	})

	t.Run("BatchResendWritesRemarksWhenPlatformReturnsFailureCode", func(t *testing.T) {
		repo := &legacyRepository{
			sup: &stubSupplierGateway{
				getSupplierByHID: func(hid int) (*model.SupplierFull, error) {
					return &model.SupplierFull{HID: hid}, nil
				},
				resubmitOrder: func(sup *model.SupplierFull, yid string) (int, string, error) {
					if yid != "Y69" {
						t.Fatalf("unexpected yid: %s", yid)
					}
					return 3, "upstream rejected", nil
				},
			},
		}

		changeStatusQueryHook = func(query string, args []driver.NamedValue) (driver.Rows, error) {
			if !strings.Contains(query, "SELECT COALESCE(hid,0), COALESCE(yid,''), COALESCE(status,'') FROM qingka_wangke_order WHERE oid = ?") {
				return nil, fmt.Errorf("unexpected query: %s", query)
			}
			return singleRow([]string{"hid", "yid", "status"}, int64(9), "Y69", "进行中"), nil
		}

		execCalled := false
		changeStatusExecHook = func(query string, args []driver.NamedValue) error {
			execCalled = true
			if !strings.Contains(query, "UPDATE qingka_wangke_order SET remarks = ? WHERE oid = ?") {
				return fmt.Errorf("unexpected exec query: %s", query)
			}
			if got := namedValueStrings(args); !reflect.DeepEqual(got, []string{"补单失败: upstream rejected", "69"}) {
				return fmt.Errorf("unexpected exec args: %v", got)
			}
			return nil
		}

		success, fail, err := repo.BatchResendOrders([]int{69})
		if err != nil {
			t.Fatalf("batch resend returned error: %v", err)
		}
		if success != 0 || fail != 1 {
			t.Fatalf("unexpected resend counts: success=%d fail=%d", success, fail)
		}
		if !execCalled {
			t.Fatal("expected platform failure remarks update to execute")
		}
	})

	t.Run("BatchResendFailsWhenSupplierMissing", func(t *testing.T) {
		repo := &legacyRepository{
			sup: &stubSupplierGateway{
				getSupplierByHID: func(hid int) (*model.SupplierFull, error) {
					return nil, errors.New("missing supplier")
				},
				resubmitOrder: func(sup *model.SupplierFull, yid string) (int, string, error) {
					t.Fatal("did not expect resubmit call when supplier lookup fails")
					return 0, "", nil
				},
			},
		}

		changeStatusQueryHook = func(query string, args []driver.NamedValue) (driver.Rows, error) {
			return singleRow([]string{"hid", "yid", "status"}, int64(9), "Y65", "进行中"), nil
		}

		execCalls := 0
		changeStatusExecHook = func(query string, args []driver.NamedValue) error {
			execCalls++
			return nil
		}

		success, fail, err := repo.BatchResendOrders([]int{65})
		if err != nil {
			t.Fatalf("batch resend returned error: %v", err)
		}
		if success != 0 || fail != 1 {
			t.Fatalf("unexpected resend counts: success=%d fail=%d", success, fail)
		}
		if execCalls != 0 {
			t.Fatalf("did not expect remarks/status updates when supplier lookup fails, got %d execs", execCalls)
		}
	})
}
