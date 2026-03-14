package tenant

import (
	"errors"
	"fmt"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

func lookupMallPayOrderTID(outTradeNo string, tid *int) error {
	return database.DB.QueryRow(
		"SELECT tid FROM qingka_mall_pay_order WHERE out_trade_no=?", outTradeNo,
	).Scan(tid)
}

func lookupTenantByUID(uid int) (*model.Tenant, error) {
	var t model.Tenant
	err := database.DB.QueryRow(
		"SELECT tid,uid,shop_name,COALESCE(shop_logo,''),COALESCE(shop_desc,''),COALESCE(domain,''),pay_config,status,addtime FROM qingka_tenant WHERE uid=?", uid,
	).Scan(&t.TID, &t.UID, &t.ShopName, &t.ShopLogo, &t.ShopDesc, &t.Domain, &t.PayConfig, &t.Status, &t.AddTime)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func lookupTenantByTID(tid int) (*model.Tenant, error) {
	var t model.Tenant
	err := database.DB.QueryRow(
		"SELECT tid,uid,shop_name,COALESCE(shop_logo,''),COALESCE(shop_desc,''),COALESCE(domain,''),pay_config,status,addtime FROM qingka_tenant WHERE tid=?", tid,
	).Scan(&t.TID, &t.UID, &t.ShopName, &t.ShopLogo, &t.ShopDesc, &t.Domain, &t.PayConfig, &t.Status, &t.AddTime)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func lookupCUserByToken(token string) (*model.CUser, error) {
	var u model.CUser
	err := database.DB.QueryRow(
		"SELECT id,tid,phone,nickname,addtime FROM qingka_c_user WHERE token=?", token,
	).Scan(&u.ID, &u.TID, &u.Phone, &u.Nickname, &u.AddTime)
	if err != nil {
		return nil, errors.New("token无效")
	}
	return &u, nil
}

func CreateTenant(uid int, req *model.TenantSaveRequest) (int64, error) {
	var exists int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_tenant WHERE uid=?", uid).Scan(&exists)
	if exists > 0 {
		return 0, errors.New("该用户已开通店铺")
	}
	res, err := database.DB.Exec(
		"INSERT INTO qingka_tenant (uid,shop_name,shop_logo,shop_desc,domain,status,addtime) VALUES (?,?,?,?,?,1,?)",
		uid, req.ShopName, req.ShopLogo, req.ShopDesc, req.Domain, time.Now().Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func updateTenantShop(uid int, req *model.TenantSaveRequest) error {
	_, err := database.DB.Exec(
		"UPDATE qingka_tenant SET shop_name=?,shop_logo=?,shop_desc=?,domain=? WHERE uid=?",
		req.ShopName, req.ShopLogo, req.ShopDesc, req.Domain, uid,
	)
	return err
}

func updateTenantPayConfig(uid int, payConfig string) error {
	_, err := database.DB.Exec("UPDATE qingka_tenant SET pay_config=? WHERE uid=?", payConfig, uid)
	return err
}

func listTenantProducts(tid int) ([]model.TenantProduct, error) {
	rows, err := database.DB.Query(`
		SELECT tp.id, tp.tid, tp.cid, tp.retail_price, tp.status, tp.sort,
		       c.name, c.price, c.fenlei
		FROM qingka_tenant_product tp
		LEFT JOIN qingka_wangke_class c ON c.cid = tp.cid
		WHERE tp.tid=? ORDER BY tp.sort ASC, tp.id ASC`, tid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.TenantProduct
	for rows.Next() {
		var p model.TenantProduct
		rows.Scan(&p.ID, &p.TID, &p.CID, &p.RetailPrice, &p.Status, &p.Sort, &p.ClassName, &p.SupplyPrice, &p.Fenlei)
		list = append(list, p)
	}
	if list == nil {
		list = []model.TenantProduct{}
	}
	return list, nil
}

func saveTenantProduct(tid int, req *model.TenantProductSaveRequest) error {
	_, err := database.DB.Exec(`
		INSERT INTO qingka_tenant_product (tid,cid,retail_price,status,sort)
		VALUES (?,?,?,?,?)
		ON DUPLICATE KEY UPDATE retail_price=VALUES(retail_price), status=VALUES(status), sort=VALUES(sort)`,
		tid, req.CID, req.RetailPrice, req.Status, req.Sort,
	)
	return err
}

func deleteTenantProduct(tid, cid int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_tenant_product WHERE tid=? AND cid=?", tid, cid)
	return err
}

func getTenantOrderStats(tid int) (map[string]interface{}, error) {
	var total, processing, completed int
	var totalRetail float64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE tid=?", tid).Scan(&total)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE tid=? AND status='进行中'", tid).Scan(&processing)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE tid=? AND status='已完成'", tid).Scan(&completed)
	database.DB.QueryRow("SELECT IFNULL(SUM(retail_fees),0) FROM qingka_wangke_order WHERE tid=?", tid).Scan(&totalRetail)

	var totalCost float64
	database.DB.QueryRow("SELECT IFNULL(SUM(fees),0) FROM qingka_wangke_order WHERE tid=?", tid).Scan(&totalCost)

	return map[string]interface{}{
		"total":        total,
		"processing":   processing,
		"completed":    completed,
		"total_retail": fmt.Sprintf("%.2f", totalRetail),
		"total_cost":   fmt.Sprintf("%.2f", totalCost),
		"profit":       fmt.Sprintf("%.2f", totalRetail-totalCost),
	}, nil
}

func listTenantCUsers(tid, page, limit int) ([]model.CUser, int, error) {
	offset := (page - 1) * limit
	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_c_user WHERE tid=?", tid).Scan(&total)
	rows, err := database.DB.Query(
		"SELECT id,tid,account,nickname,addtime FROM qingka_c_user WHERE tid=? ORDER BY id DESC LIMIT ? OFFSET ?",
		tid, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.CUser
	for rows.Next() {
		var u model.CUser
		rows.Scan(&u.ID, &u.TID, &u.Account, &u.Nickname, &u.AddTime)
		list = append(list, u)
	}
	if list == nil {
		list = []model.CUser{}
	}
	return list, total, nil
}

func saveTenantCUser(tid int, req *model.CUserSaveRequest) error {
	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Account
	}
	if req.ID == 0 {
		if req.Password == "" {
			return errors.New("新增用户需要设置密码")
		}
		_, err := database.DB.Exec(
			"INSERT INTO qingka_c_user (tid,account,password,nickname,addtime) VALUES (?,?,?,?,?)",
			tid, req.Account, req.Password, nickname, time.Now().Format("2006-01-02 15:04:05"),
		)
		return err
	}
	if req.Password != "" {
		_, err := database.DB.Exec(
			"UPDATE qingka_c_user SET account=?,password=?,nickname=? WHERE id=? AND tid=?",
			req.Account, req.Password, nickname, req.ID, tid,
		)
		return err
	}
	_, err := database.DB.Exec(
		"UPDATE qingka_c_user SET account=?,nickname=? WHERE id=? AND tid=?",
		req.Account, nickname, req.ID, tid,
	)
	return err
}

func deleteTenantCUser(tid, id int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_c_user WHERE id=? AND tid=?", id, tid)
	return err
}

func listTenantMallOrders(tid, page, limit int) ([]model.MallPayOrder, int, error) {
	offset := (page - 1) * limit
	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_mall_pay_order WHERE tid=?", tid).Scan(&total)
	rows, err := database.DB.Query(`
		SELECT id, out_trade_no, trade_no, tid, cid, c_uid, account, remark, pay_type, money, status, order_id, addtime
		FROM qingka_mall_pay_order
		WHERE tid=?
		ORDER BY id DESC LIMIT ? OFFSET ?`, tid, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.MallPayOrder
	for rows.Next() {
		var o model.MallPayOrder
		rows.Scan(&o.ID, &o.OutTradeNo, &o.TradeNo, &o.TID, &o.CID, &o.CUID, &o.Account, &o.Remark, &o.PayType, &o.Money, &o.Status, &o.OrderID, &o.AddTime)
		list = append(list, o)
	}
	if list == nil {
		list = []model.MallPayOrder{}
	}
	return list, total, nil
}
