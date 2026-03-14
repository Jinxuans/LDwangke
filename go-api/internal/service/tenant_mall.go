package service

import (
	"errors"
	"fmt"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

func (s *TenantService) GetMallOpenPrice() float64 {
	if v := getAdminConfigValue("mall_open_price"); v != "" {
		var price float64
		fmt.Sscanf(v, "%f", &price)
		if price > 0 {
			return price
		}
	}
	return 99
}

func (s *TenantService) OpenMall(uid int, shopName string) (int64, error) {
	var exists int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_tenant WHERE uid=?", uid).Scan(&exists)
	if exists > 0 {
		return 0, errors.New("已开通商城，无需重复开通")
	}

	price := s.GetMallOpenPrice()

	var money float64
	if err := database.DB.QueryRow("SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid=?", uid).Scan(&money); err != nil {
		return 0, errors.New("用户不存在")
	}
	if money < price {
		return 0, fmt.Errorf("余额不足，开通需要 %.0f 元，当前余额 %.2f 元", price, money)
	}

	res, err := database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ? AND money >= ?", price, uid, price)
	if err != nil {
		return 0, errors.New("扣费失败")
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return 0, errors.New("余额不足，请刷新后重试")
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	database.DB.Exec(
		"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, '扣费', ?, (SELECT money FROM qingka_wangke_user WHERE uid=?), ?, ?)",
		uid, -price, uid, fmt.Sprintf("开通商城扣费 %.0f 元", price), now,
	)

	if shopName == "" {
		shopName = "我的商城"
	}
	r, err := database.DB.Exec(
		"INSERT INTO qingka_tenant (uid,shop_name,status,addtime) VALUES (?,?,1,?)",
		uid, shopName, now,
	)
	if err != nil {
		return 0, errors.New("开通失败，请联系客服")
	}
	return r.LastInsertId()
}

func (s *TenantService) GetByTID(tid int) (*model.Tenant, error) {
	var t model.Tenant
	err := database.DB.QueryRow(
		"SELECT tid,uid,shop_name,COALESCE(shop_logo,''),COALESCE(shop_desc,''),COALESCE(domain,''),pay_config,status,addtime FROM qingka_tenant WHERE tid=?", tid,
	).Scan(&t.TID, &t.UID, &t.ShopName, &t.ShopLogo, &t.ShopDesc, &t.Domain, &t.PayConfig, &t.Status, &t.AddTime)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *TenantService) MallProducts(tid int) ([]model.MallProduct, error) {
	rows, err := database.DB.Query(`
		SELECT c.cid, c.name, c.noun, tp.retail_price, c.fenlei, c.sort
		FROM qingka_tenant_product tp
		JOIN qingka_wangke_class c ON c.cid = tp.cid
		WHERE tp.tid=? AND tp.status=1 AND c.status=1
		ORDER BY tp.sort ASC, c.sort ASC`, tid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.MallProduct
	for rows.Next() {
		var p model.MallProduct
		rows.Scan(&p.CID, &p.Name, &p.Noun, &p.RetailPrice, &p.Fenlei, &p.Sort)
		list = append(list, p)
	}
	return list, nil
}

func (s *TenantService) MallProductDetail(tid, cid int) (*model.MallProduct, error) {
	var p model.MallProduct
	err := database.DB.QueryRow(`
		SELECT c.cid, c.name, c.noun, tp.retail_price, c.fenlei, c.sort
		FROM qingka_tenant_product tp
		JOIN qingka_wangke_class c ON c.cid = tp.cid
		WHERE tp.tid=? AND tp.cid=? AND tp.status=1 AND c.status=1`, tid, cid,
	).Scan(&p.CID, &p.Name, &p.Noun, &p.RetailPrice, &p.Fenlei, &p.Sort)
	if err != nil {
		return nil, errors.New("商品不存在")
	}
	return &p, nil
}
