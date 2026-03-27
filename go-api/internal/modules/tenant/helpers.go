package tenant

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
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
		"SELECT tid,uid,shop_name,COALESCE(shop_logo,''),COALESCE(shop_desc,''),COALESCE(domain,''),pay_config,mall_config,status,addtime FROM qingka_tenant WHERE uid=?", uid,
	).Scan(&t.TID, &t.UID, &t.ShopName, &t.ShopLogo, &t.ShopDesc, &t.Domain, &t.PayConfig, &t.MallConfig, &t.Status, &t.AddTime)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func lookupTenantByTID(tid int) (*model.Tenant, error) {
	var t model.Tenant
	err := database.DB.QueryRow(
		"SELECT tid,uid,shop_name,COALESCE(shop_logo,''),COALESCE(shop_desc,''),COALESCE(domain,''),pay_config,mall_config,status,addtime FROM qingka_tenant WHERE tid=?", tid,
	).Scan(&t.TID, &t.UID, &t.ShopName, &t.ShopLogo, &t.ShopDesc, &t.Domain, &t.PayConfig, &t.MallConfig, &t.Status, &t.AddTime)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func normalizeTenantDomain(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	raw = strings.TrimPrefix(raw, "https://")
	raw = strings.TrimPrefix(raw, "http://")
	if idx := strings.Index(raw, "/"); idx >= 0 {
		raw = raw[:idx]
	}
	host, _, err := net.SplitHostPort(raw)
	if err == nil && host != "" {
		raw = host
	}
	return strings.TrimSpace(raw)
}

func lookupTenantByDomain(domain string) (*model.Tenant, error) {
	domain = normalizeTenantDomain(domain)
	if domain == "" {
		return nil, errors.New("店铺域名未绑定")
	}

	var t model.Tenant
	err := database.DB.QueryRow(
		"SELECT tid,uid,shop_name,COALESCE(shop_logo,''),COALESCE(shop_desc,''),COALESCE(domain,''),pay_config,mall_config,status,addtime FROM qingka_tenant WHERE domain=? LIMIT 1",
		domain,
	).Scan(&t.TID, &t.UID, &t.ShopName, &t.ShopLogo, &t.ShopDesc, &t.Domain, &t.PayConfig, &t.MallConfig, &t.Status, &t.AddTime)
	if err == nil {
		return &t, nil
	}

	if strings.HasPrefix(domain, "www.") {
		return lookupTenantByDomain(strings.TrimPrefix(domain, "www."))
	}

	return nil, err
}

func lookupCUserByToken(token string) (*model.CUser, error) {
	var u model.CUser
	err := database.DB.QueryRow(
		"SELECT id,tid,COALESCE(phone,''),account,COALESCE(nickname,''),COALESCE(invite_code,''),COALESCE(referrer_id,0),COALESCE(commission_money,0), COALESCE(commission_cdmoney,0), COALESCE(commission_total,0),COALESCE(status,1),addtime FROM qingka_c_user WHERE token=?",
		token,
	).Scan(&u.ID, &u.TID, &u.Phone, &u.Account, &u.Nickname, &u.InviteCode, &u.ReferrerID, &u.CommissionMoney, &u.CommissionCD, &u.CommissionTotal, &u.Status, &u.AddTime)
	if err != nil {
		return nil, errors.New("token无效")
	}
	return &u, nil
}

func lookupCUserByTokenForTenant(token string, tid int) (*model.CUser, error) {
	u, err := lookupCUserByToken(token)
	if err != nil {
		return nil, err
	}
	if u.TID != tid {
		return nil, errors.New("账号不属于当前店铺")
	}
	return u, nil
}

func CreateTenant(uid int, req *model.TenantSaveRequest) (int64, error) {
	var exists int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_tenant WHERE uid=?", uid).Scan(&exists)
	if exists > 0 {
		return 0, errors.New("该用户已开通店铺")
	}
	res, err := database.DB.Exec(
		"INSERT INTO qingka_tenant (uid,shop_name,shop_logo,shop_desc,domain,mall_config,status,addtime) VALUES (?,?,?,?,?,?,1,?)",
		uid, req.ShopName, req.ShopLogo, req.ShopDesc, normalizeTenantDomain(req.Domain), defaultTenantMallConfigJSON(), time.Now().Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func updateTenantShop(uid int, req *model.TenantSaveRequest) error {
	_, err := database.DB.Exec(
		"UPDATE qingka_tenant SET shop_name=?,shop_logo=?,shop_desc=?,domain=? WHERE uid=?",
		req.ShopName, req.ShopLogo, req.ShopDesc, normalizeTenantDomain(req.Domain), uid,
	)
	return err
}

func updateTenantPayConfig(uid int, payConfig string) error {
	_, err := database.DB.Exec("UPDATE qingka_tenant SET pay_config=? WHERE uid=?", payConfig, uid)
	return err
}

func updateTenantMallConfig(uid int, mallConfig string) error {
	_, err := database.DB.Exec("UPDATE qingka_tenant SET mall_config=? WHERE uid=?", mallConfig, uid)
	return err
}

func defaultTenantMallConfig() model.TenantMallConfig {
	return model.TenantMallConfig{
		RegisterEnabled:  false,
		PromotionEnabled: false,
		CommissionRate:   0,
		ShowCategories:   true,
		CustomerService: model.TenantCustomerServiceConfig{
			Enabled: false,
			Type:    "wechat",
			Value:   "",
			Label:   "联系客服",
		},
	}
}

func clampCommissionRate(rate float64) float64 {
	if rate < 0 {
		return 0
	}
	if rate > 100 {
		return 100
	}
	return float64(int(rate*100+0.5)) / 100
}

func parseTenantMallConfig(raw *string) model.TenantMallConfig {
	cfg := defaultTenantMallConfig()
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return cfg
	}
	_ = json.Unmarshal([]byte(*raw), &cfg)
	cfg.CommissionRate = clampCommissionRate(cfg.CommissionRate)
	return cfg
}

func defaultTenantMallConfigJSON() string {
	buf, _ := json.Marshal(defaultTenantMallConfig())
	return string(buf)
}

func normalizeTenantMallConfig(raw string) (string, error) {
	cfg := defaultTenantMallConfig()
	if strings.TrimSpace(raw) != "" {
		if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
			return "", errors.New("商城配置格式错误")
		}
	}
	cfg.CommissionRate = clampCommissionRate(cfg.CommissionRate)
	cfg.CustomerService.Type = strings.TrimSpace(strings.ToLower(cfg.CustomerService.Type))
	switch cfg.CustomerService.Type {
	case "wechat", "qq", "phone", "link":
	default:
		cfg.CustomerService.Type = "wechat"
	}
	cfg.CustomerService.Value = strings.TrimSpace(cfg.CustomerService.Value)
	cfg.CustomerService.Label = strings.TrimSpace(cfg.CustomerService.Label)
	if cfg.CustomerService.Label == "" {
		cfg.CustomerService.Label = "联系客服"
	}
	if cfg.CustomerService.Value == "" {
		cfg.CustomerService.Enabled = false
	}
	buf, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func generateCUserInviteCode(tid int) string {
	return fmt.Sprintf("M%d%s", tid, strings.ToUpper(genToken()[:8]))
}

func uniqueCUserInviteCode(tid int) (string, error) {
	for i := 0; i < 8; i++ {
		code := generateCUserInviteCode(tid)
		var exists int
		if err := database.DB.QueryRow(
			"SELECT COUNT(*) FROM qingka_c_user WHERE tid=? AND invite_code=?",
			tid, code,
		).Scan(&exists); err != nil {
			return "", err
		}
		if exists == 0 {
			return code, nil
		}
	}
	return "", errors.New("生成推广码失败")
}

func lookupCUserByInviteCode(tid int, code string) (*model.CUser, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, errors.New("推广码为空")
	}
	var u model.CUser
	err := database.DB.QueryRow(
		"SELECT id,tid,COALESCE(phone,''),account,COALESCE(nickname,''),COALESCE(invite_code,''),COALESCE(referrer_id,0),COALESCE(commission_money,0), COALESCE(commission_cdmoney,0), COALESCE(commission_total,0),COALESCE(status,1),addtime FROM qingka_c_user WHERE tid=? AND invite_code=? LIMIT 1",
		tid, code,
	).Scan(&u.ID, &u.TID, &u.Phone, &u.Account, &u.Nickname, &u.InviteCode, &u.ReferrerID, &u.CommissionMoney, &u.CommissionCD, &u.CommissionTotal, &u.Status, &u.AddTime)
	if err != nil {
		return nil, errors.New("推广会员不存在")
	}
	return &u, nil
}

func lookupCUserReferrer(tid, id int) (string, string) {
	if tid <= 0 || id <= 0 {
		return "", ""
	}
	var account, nickname string
	_ = database.DB.QueryRow(
		"SELECT COALESCE(account,''), COALESCE(nickname,'') FROM qingka_c_user WHERE tid=? AND id=? LIMIT 1",
		tid, id,
	).Scan(&account, &nickname)
	return account, nickname
}

func listTenantProducts(tid int) ([]model.TenantProduct, error) {
	rows, err := database.DB.Query(`
		SELECT tp.id, tp.tid, tp.cid, tp.retail_price, tp.status, tp.sort,
		       COALESCE(tp.display_name,''), COALESCE(tp.cover_url,''),
		       COALESCE(NULLIF(tp.description,''), NULLIF(c.content,''), COALESCE(c.noun,''), ''),
		       COALESCE(tp.category_id,0),
		       COALESCE(mc.name,''), c.name, c.price, COALESCE(c.fenlei,'')
		FROM qingka_tenant_product tp
		LEFT JOIN qingka_wangke_class c ON c.cid = tp.cid
		LEFT JOIN qingka_tenant_mall_category mc ON mc.id = tp.category_id AND mc.tid = tp.tid
		WHERE tp.tid=? ORDER BY tp.sort ASC, tp.id ASC`, tid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.TenantProduct
	for rows.Next() {
		var p model.TenantProduct
		if err := rows.Scan(
			&p.ID,
			&p.TID,
			&p.CID,
			&p.RetailPrice,
			&p.Status,
			&p.Sort,
			&p.DisplayName,
			&p.CoverURL,
			&p.Description,
			&p.CategoryID,
			&p.CategoryName,
			&p.ClassName,
			&p.SupplyPrice,
			&p.Fenlei,
		); err != nil {
			return nil, err
		}
		if p.CategoryName == "" {
			p.CategoryName = p.Fenlei
		}
		list = append(list, p)
	}
	if list == nil {
		list = []model.TenantProduct{}
	}
	return list, nil
}

func ensureTenantMallCategory(tid, categoryID int, categoryName string) (int, string, error) {
	categoryName = strings.TrimSpace(categoryName)
	if categoryID > 0 {
		var existingName string
		err := database.DB.QueryRow(
			"SELECT COALESCE(name,'') FROM qingka_tenant_mall_category WHERE id=? AND tid=? LIMIT 1",
			categoryID, tid,
		).Scan(&existingName)
		if err == nil && strings.TrimSpace(existingName) != "" {
			return categoryID, strings.TrimSpace(existingName), nil
		}
	}
	if categoryName == "" {
		return 0, "", nil
	}

	var id int
	var existingName string
	err := database.DB.QueryRow(
		"SELECT id, COALESCE(name,'') FROM qingka_tenant_mall_category WHERE tid=? AND name=? LIMIT 1",
		tid, categoryName,
	).Scan(&id, &existingName)
	if err == nil {
		return id, strings.TrimSpace(existingName), nil
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	res, err := database.DB.Exec(
		"INSERT INTO qingka_tenant_mall_category (tid, name, sort, status, addtime) VALUES (?, ?, 10, 1, ?)",
		tid, categoryName, now,
	)
	if err != nil {
		return 0, "", err
	}
	id64, err := res.LastInsertId()
	if err != nil {
		return 0, "", err
	}
	return int(id64), categoryName, nil
}

func validateTenantProductCID(cid int) error {
	if cid <= 0 {
		return errors.New("请选择有效商品")
	}
	var name string
	var status int
	err := database.DB.QueryRow(
		"SELECT COALESCE(name,''), COALESCE(status,0) FROM qingka_wangke_class WHERE cid=?",
		cid,
	).Scan(&name, &status)
	if err != nil || name == "" {
		return errors.New("商品不存在")
	}
	if status != 1 {
		return errors.New("商品已下架，不能上架到商城")
	}
	return nil
}

func loadTenantProductSourceInfo(cid int) (name, noun, content string, status int, err error) {
	err = database.DB.QueryRow(
		"SELECT COALESCE(name,''), COALESCE(noun,''), COALESCE(content,''), COALESCE(status,0) FROM qingka_wangke_class WHERE cid=?",
		cid,
	).Scan(&name, &noun, &content, &status)
	return
}

func saveTenantProduct(tid int, req *model.TenantProductSaveRequest) error {
	if err := validateTenantProductCID(req.CID); err != nil {
		return err
	}
	if req.RetailPrice <= 0 {
		return errors.New("零售价必须大于0")
	}
	_, sourceNoun, sourceContent, _, err := loadTenantProductSourceInfo(req.CID)
	if err != nil {
		return errors.New("商品不存在")
	}
	categoryID, _, err := ensureTenantMallCategory(tid, req.CategoryID, req.CategoryName)
	if err != nil {
		return errors.New("商城分类保存失败")
	}
	description := strings.TrimSpace(req.Description)
	if description == "" {
		description = strings.TrimSpace(sourceContent)
	}
	if description == "" {
		description = strings.TrimSpace(sourceNoun)
	}
	_, err = database.DB.Exec(`
		INSERT INTO qingka_tenant_product (tid,cid,retail_price,status,sort,display_name,cover_url,description,category_id)
		VALUES (?,?,?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE retail_price=VALUES(retail_price), status=VALUES(status), sort=VALUES(sort),
		    display_name=VALUES(display_name), cover_url=VALUES(cover_url), description=VALUES(description), category_id=VALUES(category_id)`,
		tid,
		req.CID,
		req.RetailPrice,
		req.Status,
		req.Sort,
		strings.TrimSpace(req.DisplayName),
		strings.TrimSpace(req.CoverURL),
		description,
		categoryID,
	)
	return err
}

func deleteTenantProduct(tid, cid int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_tenant_product WHERE tid=? AND cid=?", tid, cid)
	return err
}

func getTenantOrderStats(tid int) (map[string]interface{}, error) {
	var total, today, pending, processing, completed int
	var totalRetail, todayRetail float64
	todayPrefix := time.Now().Format("2006-01-02") + "%"

	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE tid=?", tid).Scan(&total)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE tid=? AND addtime LIKE ?", tid, todayPrefix).Scan(&today)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE tid=? AND status IN ('待处理','进行中')", tid).Scan(&pending)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE tid=? AND status='进行中'", tid).Scan(&processing)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE tid=? AND status='已完成'", tid).Scan(&completed)
	database.DB.QueryRow("SELECT IFNULL(SUM(retail_fees),0) FROM qingka_wangke_order WHERE tid=?", tid).Scan(&totalRetail)
	database.DB.QueryRow("SELECT IFNULL(SUM(retail_fees),0) FROM qingka_wangke_order WHERE tid=? AND addtime LIKE ?", tid, todayPrefix).Scan(&todayRetail)

	var totalCost float64
	database.DB.QueryRow("SELECT IFNULL(SUM(fees),0) FROM qingka_wangke_order WHERE tid=?", tid).Scan(&totalCost)

	return map[string]interface{}{
		"total":        total,
		"today":        today,
		"pending":      pending,
		"done":         completed,
		"processing":   processing,
		"completed":    completed,
		"total_retail": totalRetail,
		"today_retail": todayRetail,
		"total_cost":   totalCost,
		"profit":       totalRetail - totalCost,
	}, nil
}

func listTenantCUsers(tid, page, limit int) ([]model.CUser, int, error) {
	offset := (page - 1) * limit
	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_c_user WHERE tid=?", tid).Scan(&total)
	rows, err := database.DB.Query(
		"SELECT id,tid,account,COALESCE(nickname,''),COALESCE(invite_code,''),COALESCE(commission_money,0),COALESCE(commission_cdmoney,0),COALESCE(commission_total,0),COALESCE(status,1),addtime FROM qingka_c_user WHERE tid=? ORDER BY id DESC LIMIT ? OFFSET ?",
		tid, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.CUser
	for rows.Next() {
		var u model.CUser
		rows.Scan(&u.ID, &u.TID, &u.Account, &u.Nickname, &u.InviteCode, &u.CommissionMoney, &u.CommissionCD, &u.CommissionTotal, &u.Status, &u.AddTime)
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
		var exists int
		if err := database.DB.QueryRow(
			"SELECT COUNT(*) FROM qingka_c_user WHERE tid=? AND account=?",
			tid, req.Account,
		).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			return errors.New("会员账号已存在")
		}
		inviteCode, err := uniqueCUserInviteCode(tid)
		if err != nil {
			return err
		}
		_, err = database.DB.Exec(
			"INSERT INTO qingka_c_user (tid,account,password,nickname,invite_code,status,addtime) VALUES (?,?,?,?,?,1,?)",
			tid, req.Account, req.Password, nickname, inviteCode, time.Now().Format("2006-01-02 15:04:05"),
		)
		return err
	}
	var exists int
	if err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM qingka_c_user WHERE tid=? AND account=? AND id<>?",
		tid, req.Account, req.ID,
	).Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		return errors.New("会员账号已存在")
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

func formatCUserMoney(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}

func listTenantMallCategories(tid int) ([]model.TenantMallCategory, error) {
	rows, err := database.DB.Query(
		"SELECT id, tid, COALESCE(name,''), COALESCE(sort,10), COALESCE(status,1), COALESCE(DATE_FORMAT(addtime,'%Y-%m-%d %H:%i:%s'),'') FROM qingka_tenant_mall_category WHERE tid=? ORDER BY sort ASC, id ASC",
		tid,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]model.TenantMallCategory, 0)
	for rows.Next() {
		var item model.TenantMallCategory
		if err := rows.Scan(&item.ID, &item.TID, &item.Name, &item.Sort, &item.Status, &item.AddTime); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

func updateTenantMallCategorySort(tid int, items []struct{ ID, Sort int }) error {
	for _, item := range items {
		if item.ID <= 0 {
			continue
		}
		if _, err := database.DB.Exec(
			"UPDATE qingka_tenant_mall_category SET sort=? WHERE tid=? AND id=?",
			item.Sort, tid, item.ID,
		); err != nil {
			return err
		}
	}
	return nil
}

func saveTenantMallCategory(tid int, req *model.TenantMallCategorySaveRequest) error {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errors.New("请输入分类名称")
	}
	status := req.Status
	if status != 0 {
		status = 1
	}
	if req.ID > 0 {
		var exists int
		if err := database.DB.QueryRow(
			"SELECT COUNT(*) FROM qingka_tenant_mall_category WHERE tid=? AND name=? AND id<>?",
			tid, name, req.ID,
		).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			return errors.New("分类名称已存在")
		}
		_, err := database.DB.Exec(
			"UPDATE qingka_tenant_mall_category SET name=?, sort=?, status=? WHERE id=? AND tid=?",
			name, req.Sort, status, req.ID, tid,
		)
		return err
	}

	var exists int
	if err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM qingka_tenant_mall_category WHERE tid=? AND name=?",
		tid, name,
	).Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		return errors.New("分类名称已存在")
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := database.DB.Exec(
		"INSERT INTO qingka_tenant_mall_category (tid, name, sort, status, addtime) VALUES (?, ?, ?, ?, ?)",
		tid, name, req.Sort, status, now,
	)
	return err
}

func deleteTenantMallCategory(tid, id int) error {
	if id <= 0 {
		return errors.New("分类不存在")
	}
	var used int
	if err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM qingka_tenant_product WHERE tid=? AND category_id=?",
		tid, id,
	).Scan(&used); err != nil {
		return err
	}
	if used > 0 {
		return errors.New("该分类下还有商品，无法删除")
	}
	_, err := database.DB.Exec("DELETE FROM qingka_tenant_mall_category WHERE tid=? AND id=?", tid, id)
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
		SELECT p.id, p.out_trade_no, p.trade_no, p.tid, p.cid, p.c_uid, COALESCE(p.school,''), p.account, p.remark, p.pay_type, p.money, p.status, p.order_id, COALESCE(DATE_FORMAT(p.addtime,'%Y-%m-%d %H:%i:%s'),''),
		       COALESCE(c.name,''), COALESCE(p.course_name,''), COALESCE(o.status,''), COALESCE(o.process,''), COALESCE(o.remarks,''), COALESCE(rc.cnt,0)
		FROM qingka_mall_pay_order p
		LEFT JOIN qingka_wangke_class c ON c.cid = p.cid
		LEFT JOIN qingka_wangke_order o ON o.oid = p.order_id
		LEFT JOIN (
		    SELECT out_trade_no COLLATE utf8mb4_unicode_ci AS out_trade_no, COUNT(*) AS cnt
		    FROM qingka_wangke_order
		    WHERE tid=? AND COALESCE(out_trade_no,'') <> ''
		    GROUP BY out_trade_no
		) rc ON rc.out_trade_no = p.out_trade_no COLLATE utf8mb4_unicode_ci
		WHERE p.tid=?
		ORDER BY p.id DESC LIMIT ? OFFSET ?`, tid, tid, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.MallPayOrder
	for rows.Next() {
		var o model.MallPayOrder
		rows.Scan(&o.ID, &o.OutTradeNo, &o.TradeNo, &o.TID, &o.CID, &o.CUID, &o.School, &o.Account, &o.Remark, &o.PayType, &o.Money, &o.Status, &o.OrderID, &o.AddTime, &o.ProductName, &o.CourseName, &o.OrderStatus, &o.OrderProcess, &o.OrderRemarks, &o.OrderCount)
		list = append(list, o)
	}
	if list == nil {
		list = []model.MallPayOrder{}
	}
	return list, total, nil
}

func listTenantMallLinkedOrders(tid, payOrderID int) ([]model.Order, error) {
	var outTradeNo string
	var fallbackOID int
	err := database.DB.QueryRow(
		"SELECT COALESCE(out_trade_no,''), COALESCE(order_id,0) FROM qingka_mall_pay_order WHERE id=? AND tid=?",
		payOrderID, tid,
	).Scan(&outTradeNo, &fallbackOID)
	if err != nil {
		return nil, err
	}

	query := "SELECT oid, uid, cid, hid, COALESCE(ptname,''), COALESCE(school,''), COALESCE(name,''), COALESCE(user,''), COALESCE(pass,''), COALESCE(kcname,''), COALESCE(kcid,''), COALESCE(status,'待处理'), COALESCE(fees,'0'), COALESCE(process,''), COALESCE(remarks,''), COALESCE(dockstatus,'0'), COALESCE(yid,''), COALESCE(addtime,''), COALESCE(pushUid,''), COALESCE(pushStatus,''), COALESCE(pushEmail,''), COALESCE(pushEmailStatus,'0'), COALESCE(showdoc_push_url,''), COALESCE(pushShowdocStatus,'0'), COALESCE((SELECT pt FROM qingka_wangke_huoyuan WHERE hid=qingka_wangke_order.hid LIMIT 1),'') FROM qingka_wangke_order WHERE tid=?"
	args := []interface{}{tid}
	if strings.TrimSpace(outTradeNo) != "" {
		query += " AND out_trade_no=? ORDER BY oid ASC"
		args = append(args, outTradeNo)
	} else if fallbackOID > 0 {
		query += " AND oid=?"
		args = append(args, fallbackOID)
	} else {
		return []model.Order{}, nil
	}

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.OID, &o.UID, &o.CID, &o.HID, &o.PTName, &o.School, &o.Name, &o.User, &o.Pass, &o.KCName, &o.KCID, &o.Status, &o.Fees, &o.Process, &o.Remarks, &o.DockStatus, &o.YID, &o.AddTime, &o.PushUid, &o.PushStatus, &o.PushEmail, &o.PushEmailStatus, &o.ShowdocPushURL, &o.PushShowdocStatus, &o.SupplierPT); err != nil {
			continue
		}
		list = append(list, o)
	}
	if list == nil {
		list = []model.Order{}
	}
	return list, nil
}
