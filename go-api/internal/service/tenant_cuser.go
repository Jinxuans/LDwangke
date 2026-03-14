package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"go-api/internal/database"
	"go-api/internal/model"
)

func (s *TenantService) CUserLogin(tid int, account, password string) (*model.CUser, string, error) {
	var u model.CUser
	var storedPwd string
	err := database.DB.QueryRow(
		"SELECT id,tid,account,nickname,password,addtime FROM qingka_c_user WHERE tid=? AND account=?", tid, account,
	).Scan(&u.ID, &u.TID, &u.Account, &u.Nickname, &storedPwd, &u.AddTime)
	if err != nil {
		return nil, "", errors.New("账号不存在")
	}
	if storedPwd != password {
		return nil, "", errors.New("密码错误")
	}
	token := genToken()
	database.DB.Exec("UPDATE qingka_c_user SET token=? WHERE id=?", token, u.ID)
	u.Token = token
	return &u, token, nil
}

func genToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *TenantService) CUserOrders(tid, cUID, page, limit int) ([]model.MallOrderItem, int, error) {
	offset := (page - 1) * limit
	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE tid=? AND c_uid=?", tid, cUID).Scan(&total)
	rows, err := database.DB.Query(`
		SELECT o.oid, o.cid, c.name, o.kcname, o.account, o.status, o.process, o.retail_fees, o.addtime
		FROM qingka_wangke_order o
		LEFT JOIN qingka_wangke_class c ON c.cid = o.cid
		WHERE o.tid=? AND o.c_uid=?
		ORDER BY o.oid DESC LIMIT ? OFFSET ?`, tid, cUID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.MallOrderItem
	for rows.Next() {
		var item model.MallOrderItem
		rows.Scan(&item.OID, &item.CID, &item.ClassName, &item.KCName, &item.Account, &item.Status, &item.Process, &item.RetailFees, &item.AddTime)
		list = append(list, item)
	}
	return list, total, nil
}

func (s *TenantService) CUserOrderDetail(tid, cUID int, oid int64) (*model.MallOrderItem, error) {
	var item model.MallOrderItem
	err := database.DB.QueryRow(`
		SELECT o.oid, o.cid, c.name, o.kcname, o.account, o.status, o.process, o.retail_fees, o.addtime
		FROM qingka_wangke_order o
		LEFT JOIN qingka_wangke_class c ON c.cid = o.cid
		WHERE o.oid=? AND o.tid=? AND o.c_uid=?`,
		oid, tid, cUID,
	).Scan(&item.OID, &item.CID, &item.ClassName, &item.KCName, &item.Account, &item.Status, &item.Process, &item.RetailFees, &item.AddTime)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	return &item, nil
}

func (s *TenantService) SearchMallOrders(tid int, keyword string) ([]model.MallOrderItem, error) {
	rows, err := database.DB.Query(`
		SELECT o.oid, o.cid, c.name, o.kcname, o.account, o.status, o.process, o.retail_fees, o.addtime
		FROM qingka_wangke_order o
		LEFT JOIN qingka_wangke_class c ON c.cid = o.cid
		WHERE o.tid=? AND (o.account=? OR o.oid=?)
		ORDER BY o.oid DESC LIMIT 20`, tid, keyword, keyword)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.MallOrderItem
	for rows.Next() {
		var item model.MallOrderItem
		rows.Scan(&item.OID, &item.CID, &item.ClassName, &item.KCName, &item.Account, &item.Status, &item.Process, &item.RetailFees, &item.AddTime)
		list = append(list, item)
	}
	return list, nil
}
