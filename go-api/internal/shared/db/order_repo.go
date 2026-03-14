package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

type OrderRepo struct{}

const orderColumns = "oid, uid, cid, hid, COALESCE(ptname,''), COALESCE(school,''), COALESCE(name,''), COALESCE(user,''), COALESCE(pass,''), COALESCE(kcname,''), COALESCE(kcid,''), COALESCE(status,'待处理'), COALESCE(fees,'0'), COALESCE(process,''), COALESCE(remarks,''), COALESCE(dockstatus,'0'), COALESCE(yid,''), COALESCE(addtime,''), COALESCE(pushUid,''), COALESCE(pushStatus,''), COALESCE(pushEmail,''), COALESCE(pushEmailStatus,'0'), COALESCE(showdoc_push_url,''), COALESCE(pushShowdocStatus,'0'), COALESCE((SELECT pt FROM qingka_wangke_huoyuan WHERE hid=qingka_wangke_order.hid LIMIT 1),'')"

func NewOrderRepo() *OrderRepo {
	return &OrderRepo{}
}

func (r *OrderRepo) List(uid int, grade string, req model.OrderListRequest) ([]model.Order, int64, error) {
	where := []string{"1=1"}
	args := []interface{}{}

	if grade != "2" && grade != "3" {
		where = append(where, "uid = ?")
		args = append(args, uid)
	} else if req.UID != "" {
		where = append(where, "uid = ?")
		args = append(args, req.UID)
	}

	if req.StatusText != "" {
		where = append(where, "status = ?")
		args = append(args, req.StatusText)
	}
	if req.CID != "" {
		where = append(where, "cid = ?")
		args = append(args, req.CID)
	}
	if req.Dock != "" {
		where = append(where, "dockstatus = ?")
		args = append(args, req.Dock)
	}
	if req.OID != "" {
		where = append(where, "oid = ?")
		args = append(args, req.OID)
	}
	if req.HID != "" {
		where = append(where, "hid = ?")
		args = append(args, req.HID)
	}
	if req.User != "" {
		where = append(where, "user = ?")
		args = append(args, req.User)
	}
	if req.Pass != "" {
		where = append(where, "pass = ?")
		args = append(args, req.Pass)
	}
	if req.School != "" {
		where = append(where, "school LIKE ?")
		args = append(args, "%"+req.School+"%")
	}
	if req.KCName != "" {
		where = append(where, "kcname LIKE ?")
		args = append(args, "%"+req.KCName+"%")
	}
	if req.Search != "" {
		where = append(where, "(uid LIKE ? OR ptname LIKE ? OR school LIKE ? OR user LIKE ? OR pass LIKE ? OR kcname LIKE ? OR process LIKE ? OR remarks LIKE ?)")
		s := "%" + req.Search + "%"
		args = append(args, s, s, s, s, s, s, s, s)
	}

	cutoff := time.Now().AddDate(0, 0, -100).Format("2006-01-02")
	where = append(where, "addtime >= ?")
	args = append(args, cutoff)

	whereStr := strings.Join(where, " AND ")

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM qingka_wangke_order WHERE %s", whereStr)
	if err := database.DB.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 2000 {
		req.Limit = 2000
	}

	offset := (req.Page - 1) * req.Limit
	querySQL := fmt.Sprintf("SELECT %s FROM qingka_wangke_order WHERE %s ORDER BY oid DESC LIMIT ? OFFSET ?", orderColumns, whereStr)
	queryArgs := append(args, req.Limit, offset)

	rows, err := database.DB.Query(querySQL, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			continue
		}
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []model.Order{}
	}

	return orders, total, nil
}

func (r *OrderRepo) Detail(uid int, grade string, oid int) (*model.Order, error) {
	querySQL := fmt.Sprintf("SELECT %s FROM qingka_wangke_order WHERE oid = ?", orderColumns)
	args := []interface{}{oid}

	if grade != "2" && grade != "3" {
		querySQL += " AND uid = ?"
		args = append(args, uid)
	}

	rows, err := database.DB.Query(querySQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("订单不存在")
	}

	order, err := scanOrder(rows)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepo) Stats(uid int, grade string) (*model.OrderStats, error) {
	stats := &model.OrderStats{}

	if grade == "2" || grade == "3" {
		_ = database.DB.QueryRow("SELECT COUNT(*), COALESCE(SUM(fees),0) FROM qingka_wangke_order").Scan(&stats.Total, &stats.TotalFees)
		_ = database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE status = '进行中'").Scan(&stats.Processing)
		_ = database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE status = '已完成'").Scan(&stats.Completed)
		_ = database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE status = '异常'").Scan(&stats.Failed)
		return stats, nil
	}

	_ = database.DB.QueryRow("SELECT COUNT(*), COALESCE(SUM(fees),0) FROM qingka_wangke_order WHERE uid = ?", uid).Scan(&stats.Total, &stats.TotalFees)
	_ = database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE uid = ? AND status = '进行中'", uid).Scan(&stats.Processing)
	_ = database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE uid = ? AND status = '已完成'", uid).Scan(&stats.Completed)
	_ = database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE uid = ? AND status = '异常'", uid).Scan(&stats.Failed)

	return stats, nil
}

func scanOrder(rows *sql.Rows) (model.Order, error) {
	var o model.Order
	err := rows.Scan(&o.OID, &o.UID, &o.CID, &o.HID, &o.PTName, &o.School, &o.Name, &o.User, &o.Pass, &o.KCName, &o.KCID, &o.Status, &o.Fees, &o.Process, &o.Remarks, &o.DockStatus, &o.YID, &o.AddTime, &o.PushUid, &o.PushStatus, &o.PushEmail, &o.PushEmailStatus, &o.ShowdocPushURL, &o.PushShowdocStatus, &o.SupplierPT)
	return o, err
}
