package service

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

type TenantService struct{}

func NewTenantService() *TenantService { return &TenantService{} }

// ===== 商城开通 =====

// GetMallOpenPrice 读取商城开通价格，默认 99 元
func (s *TenantService) GetMallOpenPrice() float64 {
	conf, _ := NewAdminService().GetConfig()
	if v, ok := conf["mall_open_price"]; ok && v != "" {
		var price float64
		fmt.Sscanf(v, "%f", &price)
		if price > 0 {
			return price
		}
	}
	return 99
}

// OpenMall 扣余额开通商城，返回新 tid
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

// ===== 平台侧：租户管理 =====

func (s *TenantService) Create(uid int, req *model.TenantSaveRequest) (int64, error) {
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

func (s *TenantService) List(page, limit int) ([]model.Tenant, int, error) {
	offset := (page - 1) * limit
	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_tenant").Scan(&total)
	rows, err := database.DB.Query(
		"SELECT tid,uid,shop_name,COALESCE(shop_logo,''),COALESCE(shop_desc,''),COALESCE(domain,''),status,addtime FROM qingka_tenant ORDER BY tid DESC LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []model.Tenant
	for rows.Next() {
		var t model.Tenant
		rows.Scan(&t.TID, &t.UID, &t.ShopName, &t.ShopLogo, &t.ShopDesc, &t.Domain, &t.Status, &t.AddTime)
		list = append(list, t)
	}
	return list, total, nil
}

func (s *TenantService) GetByUID(uid int) (*model.Tenant, error) {
	var t model.Tenant
	err := database.DB.QueryRow(
		"SELECT tid,uid,shop_name,COALESCE(shop_logo,''),COALESCE(shop_desc,''),COALESCE(domain,''),pay_config,status,addtime FROM qingka_tenant WHERE uid=?", uid,
	).Scan(&t.TID, &t.UID, &t.ShopName, &t.ShopLogo, &t.ShopDesc, &t.Domain, &t.PayConfig, &t.Status, &t.AddTime)
	if err != nil {
		return nil, err
	}
	return &t, nil
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

func (s *TenantService) GetByDomain(domain string) (*model.Tenant, error) {
	var t model.Tenant
	err := database.DB.QueryRow(
		"SELECT tid,uid,shop_name,shop_logo,shop_desc,domain,status,addtime FROM qingka_tenant WHERE domain=? AND status=1", domain,
	).Scan(&t.TID, &t.UID, &t.ShopName, &t.ShopLogo, &t.ShopDesc, &t.Domain, &t.Status, &t.AddTime)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *TenantService) UpdateShop(uid int, req *model.TenantSaveRequest) error {
	_, err := database.DB.Exec(
		"UPDATE qingka_tenant SET shop_name=?,shop_logo=?,shop_desc=?,domain=? WHERE uid=?",
		req.ShopName, req.ShopLogo, req.ShopDesc, req.Domain, uid,
	)
	return err
}

func (s *TenantService) UpdatePayConfig(uid int, payConfig string) error {
	_, err := database.DB.Exec("UPDATE qingka_tenant SET pay_config=? WHERE uid=?", payConfig, uid)
	return err
}

func (s *TenantService) SetStatus(tid, status int) error {
	_, err := database.DB.Exec("UPDATE qingka_tenant SET status=? WHERE tid=?", status, tid)
	return err
}

// ===== B端选品 =====

func (s *TenantService) ProductList(tid int) ([]model.TenantProduct, error) {
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
	return list, nil
}

func (s *TenantService) ProductSave(tid int, req *model.TenantProductSaveRequest) error {
	_, err := database.DB.Exec(`
		INSERT INTO qingka_tenant_product (tid,cid,retail_price,status,sort)
		VALUES (?,?,?,?,?)
		ON DUPLICATE KEY UPDATE retail_price=VALUES(retail_price), status=VALUES(status), sort=VALUES(sort)`,
		tid, req.CID, req.RetailPrice, req.Status, req.Sort,
	)
	return err
}

func (s *TenantService) ProductDelete(tid, cid int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_tenant_product WHERE tid=? AND cid=?", tid, cid)
	return err
}

// ===== C端用户 =====

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

func (s *TenantService) CUserByToken(token string) (*model.CUser, error) {
	var u model.CUser
	err := database.DB.QueryRow(
		"SELECT id,tid,phone,nickname,addtime FROM qingka_c_user WHERE token=?", token,
	).Scan(&u.ID, &u.TID, &u.Phone, &u.Nickname, &u.AddTime)
	if err != nil {
		return nil, errors.New("token无效")
	}
	return &u, nil
}

func genToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// ===== C端商城商品 =====

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

// ===== C端订单查询 =====

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

// CUserOrderDetail 查询单个C端订单详情（用于状态轮询）
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

// SearchMallOrders 按账号或订单号公开查询订单（无需登录）
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

// ===== B端商城支付订单列表 =====

func (s *TenantService) TenantMallOrders(tid, page, limit int) ([]model.MallPayOrder, int, error) {
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

// ===== B端订单统计 =====

func (s *TenantService) TenantOrderStats(tid int) (map[string]interface{}, error) {
	var total, processing, completed int
	var totalRetail float64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE tid=?", tid).Scan(&total)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE tid=? AND status='进行中'", tid).Scan(&processing)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE tid=? AND status='已完成'", tid).Scan(&completed)
	database.DB.QueryRow("SELECT IFNULL(SUM(retail_fees),0) FROM qingka_wangke_order WHERE tid=?", tid).Scan(&totalRetail)

	// 供货成本
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

// ===== C端用户管理 =====

func (s *TenantService) CUserList(tid, page, limit int) ([]model.CUser, int, error) {
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
	return list, total, nil
}

func (s *TenantService) CUserSave(tid int, req *model.CUserSaveRequest) error {
	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Account
	}
	if req.ID == 0 {
		// 新增
		if req.Password == "" {
			return errors.New("新增用户需要设置密码")
		}
		_, err := database.DB.Exec(
			"INSERT INTO qingka_c_user (tid,account,password,nickname,addtime) VALUES (?,?,?,?,?)",
			tid, req.Account, req.Password, nickname, time.Now().Format("2006-01-02 15:04:05"),
		)
		return err
	}
	// 编辑
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

func (s *TenantService) CUserDelete(tid, id int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_c_user WHERE id=? AND tid=?", id, tid)
	return err
}

// ===== C端商城支付 =====

// GetTenantPayConfig 读取租户支付配置（pay_config JSON）
func (s *TenantService) GetTenantPayConfig(tid int) (map[string]string, error) {
	var raw string
	err := database.DB.QueryRow("SELECT COALESCE(pay_config,'') FROM qingka_tenant WHERE tid=?", tid).Scan(&raw)
	if err != nil {
		return nil, errors.New("店铺不存在")
	}
	result := make(map[string]string)
	if raw != "" {
		json.Unmarshal([]byte(raw), &result)
	}
	return result, nil
}

// GetMallPayChannels 获取商城可用支付渠道
func (s *TenantService) GetMallPayChannels(tid int) ([]model.PayChannel, error) {
	cfg, err := s.GetTenantPayConfig(tid)
	if err != nil {
		return nil, err
	}
	var channels []model.PayChannel
	if cfg["is_alipay"] == "1" {
		channels = append(channels, model.PayChannel{Key: "alipay", Label: "支付宝"})
	}
	if cfg["is_wxpay"] == "1" {
		channels = append(channels, model.PayChannel{Key: "wxpay", Label: "微信支付"})
	}
	if cfg["is_qqpay"] == "1" {
		channels = append(channels, model.PayChannel{Key: "qqpay", Label: "QQ支付"})
	}
	if channels == nil {
		channels = []model.PayChannel{}
	}
	return channels, nil
}

// CreateMallPayOrder 创建商城C端支付订单
func (s *TenantService) CreateMallPayOrder(tid int, req model.MallPayRequest, domain string, cUID int) (*model.MallPayCreateResponse, error) {
	// 验证店铺
	t, err := s.GetByTID(tid)
	if err != nil || t.Status != 1 {
		return nil, errors.New("店铺不存在或已关闭")
	}

	// 验证商品
	p, err := s.MallProductDetail(tid, req.CID)
	if err != nil {
		return nil, errors.New("商品不存在")
	}

	// 读取支付配置
	cfg, err := s.GetTenantPayConfig(tid)
	if err != nil {
		return nil, err
	}
	if cfg["epay_api"] == "" || cfg["epay_pid"] == "" || cfg["epay_key"] == "" {
		return nil, errors.New("店铺暂未配置支付，请联系商家")
	}

	// 生成订单号
	outTradeNo := fmt.Sprintf("%s%d", time.Now().Format("20060102150405"), 111+time.Now().UnixNano()%889)
	money := fmt.Sprintf("%.2f", p.RetailPrice)
	name := fmt.Sprintf("%s-%s", p.Name, money)
	now := time.Now().Format("2006-01-02 15:04:05")

	_, dbErr := database.DB.Exec(
		`INSERT INTO qingka_mall_pay_order (out_trade_no, tid, cid, c_uid, account, password, remark, pay_type, money, status, addtime, course_id, course_name, course_kcjs)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?, ?, ?)`,
		outTradeNo, tid, req.CID, cUID, req.Account, req.Password, req.Remark, req.PayType, p.RetailPrice, now,
		req.CourseID, req.CourseName, req.CourseKCJS,
	)
	if dbErr != nil {
		return nil, errors.New("生成订单失败")
	}

	payURL := s.buildMallEpayURL(cfg, outTradeNo, name, money, req.PayType, domain, tid)
	return &model.MallPayCreateResponse{
		OutTradeNo: outTradeNo,
		PayURL:     payURL,
		Money:      money,
	}, nil
}

// ConfirmMallPayOrder 支付回调：验签 → 更新状态 → 自动下单
func (s *TenantService) ConfirmMallPayOrder(tid int, params map[string]string) error {
	outTradeNo := params["out_trade_no"]
	tradeNo := params["trade_no"]
	tradeStatus := params["trade_status"]

	if tradeStatus != "TRADE_SUCCESS" {
		return errors.New("支付未成功")
	}

	// 读取订单
	var order model.MallPayOrder
	err := database.DB.QueryRow(
		`SELECT id, tid, cid, c_uid, account, password, remark, pay_type, money, status, course_id, course_name, course_kcjs FROM qingka_mall_pay_order WHERE out_trade_no=?`,
		outTradeNo,
	).Scan(&order.ID, &order.TID, &order.CID, &order.CUID, &order.Account, &order.Password, &order.Remark, &order.PayType, &order.Money, &order.Status, &order.CourseID, &order.CourseName, &order.CourseKCJS)
	if err != nil {
		return errors.New("订单不存在")
	}
	if order.Status != 0 {
		return nil // 已处理，幂等
	}
	if order.TID != tid {
		return errors.New("订单不属于该店铺")
	}

	// 验签
	cfg, _ := s.GetTenantPayConfig(tid)
	if !s.verifyEpaySign(params, cfg["epay_key"]) {
		return errors.New("签名验证失败")
	}

	// 更新为已支付
	database.DB.Exec(
		`UPDATE qingka_mall_pay_order SET status=1, trade_no=?, paytime=? WHERE id=?`,
		tradeNo, time.Now().Format("2006-01-02 15:04:05"), order.ID,
	)

	// 自动下单
	t, err := s.GetByTID(tid)
	if err != nil {
		return nil
	}
	orderSvc := NewOrderService()
	userInfo := fmt.Sprintf("自动识别 %s %s", order.Account, order.Password)
	item := model.OrderAddItem{UserInfo: userInfo}
	if order.CourseID != "" {
		item.Data = model.OrderAddCourse{
			ID:   order.CourseID,
			Name: order.CourseName,
			KCJS: order.CourseKCJS,
		}
	}
	addReq := model.OrderAddRequest{
		CID:  order.CID,
		Data: []model.OrderAddItem{item},
	}
	result, err := orderSvc.AddOrdersForMall(t.UID, tid, order.CUID, order.Money, addReq)
	if err == nil && result != nil && result.SuccessCount > 0 {
		firstOID := int64(0)
		if len(result.OIDs) > 0 {
			firstOID = result.OIDs[0]
		}
		database.DB.Exec(
			`UPDATE qingka_mall_pay_order SET status=2, order_id=? WHERE id=?`,
			firstOID, order.ID,
		)
	}
	return nil
}

// buildMallEpayURL 生成易支付跳转URL
func (s *TenantService) buildMallEpayURL(cfg map[string]string, outTradeNo, name, money, payType, domain string, tid int) string {
	epayAPI := strings.TrimRight(cfg["epay_api"], "/")
	pid := cfg["epay_pid"]
	key := cfg["epay_key"]
	if epayAPI == "" || pid == "" || key == "" {
		return ""
	}

	notifyURL := "https://" + domain + "/api/v1/mall/pay/notify"
	returnURL := fmt.Sprintf("https://%s/mall/%d/orders?paid=1", domain, tid)

	params := map[string]string{
		"pid":          pid,
		"type":         payType,
		"notify_url":   notifyURL,
		"return_url":   returnURL,
		"out_trade_no": outTradeNo,
		"name":         name,
		"money":        money,
		"sitename":     domain,
	}

	keys := make([]string, 0, len(params))
	for k, v := range params {
		if v != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	signStr := strings.Join(parts, "&")
	sign := fmt.Sprintf("%x", md5.Sum([]byte(signStr+key)))

	return fmt.Sprintf("%s/submit.php?%s&sign=%s&sign_type=MD5", epayAPI, signStr, sign)
}

// UserConfirmMallPay C端用户主动确认已支付，跳过验签直接触发下单
func (s *TenantService) UserConfirmMallPay(tid int, outTradeNo string) (int, int64, error) {
	var order model.MallPayOrder
	err := database.DB.QueryRow(
		`SELECT id, tid, cid, c_uid, account, password, remark, pay_type, money, status, course_id, course_name, course_kcjs FROM qingka_mall_pay_order WHERE out_trade_no=?`,
		outTradeNo,
	).Scan(&order.ID, &order.TID, &order.CID, &order.CUID, &order.Account, &order.Password, &order.Remark, &order.PayType, &order.Money, &order.Status, &order.CourseID, &order.CourseName, &order.CourseKCJS)
	if err != nil {
		return 0, 0, errors.New("订单不存在")
	}
	if order.TID != tid {
		return 0, 0, errors.New("订单不属于该店铺")
	}
	// 已下单成功，直接返回
	if order.Status == 2 {
		var orderID int64
		database.DB.QueryRow(`SELECT COALESCE(order_id,0) FROM qingka_mall_pay_order WHERE id=?`, order.ID).Scan(&orderID)
		return 2, orderID, nil
	}
	// 标记为已支付
	if order.Status == 0 {
		database.DB.Exec(
			`UPDATE qingka_mall_pay_order SET status=1, paytime=? WHERE id=?`,
			time.Now().Format("2006-01-02 15:04:05"), order.ID,
		)
		order.Status = 1
	}
	// 触发下单
	t, err := s.GetByTID(tid)
	if err != nil {
		return 1, 0, nil
	}
	orderSvc := NewOrderService()
	userInfo := fmt.Sprintf("自动识别 %s %s", order.Account, order.Password)
	item := model.OrderAddItem{UserInfo: userInfo}
	if order.CourseID != "" {
		item.Data = model.OrderAddCourse{
			ID:   order.CourseID,
			Name: order.CourseName,
			KCJS: order.CourseKCJS,
		}
	}
	addReq := model.OrderAddRequest{
		CID:  order.CID,
		Data: []model.OrderAddItem{item},
	}
	result, err := orderSvc.AddOrdersForMall(t.UID, tid, order.CUID, order.Money, addReq)
	if err != nil {
		return 1, 0, err
	}
	if result != nil && result.SuccessCount > 0 {
		firstOID := int64(0)
		if len(result.OIDs) > 0 {
			firstOID = result.OIDs[0]
		}
		database.DB.Exec(`UPDATE qingka_mall_pay_order SET status=2, order_id=? WHERE id=?`, firstOID, order.ID)
		return 2, firstOID, nil
	}
	msg := "下单失败"
	if result != nil && len(result.SkippedItems) > 0 {
		msg = strings.Join(result.SkippedItems, "; ")
	}
	return 1, 0, errors.New(msg)
}

// CheckMallPayStatus C端检测支付状态
// 返回 status: 0=未支付, 1=已支付待下单, 2=已下单成功
func (s *TenantService) CheckMallPayStatus(tid int, outTradeNo string) (int, int64, error) {
	var status int
	var orderID int64
	var orderTID int
	err := database.DB.QueryRow(
		`SELECT status, COALESCE(order_id,0), tid FROM qingka_mall_pay_order WHERE out_trade_no=?`,
		outTradeNo,
	).Scan(&status, &orderID, &orderTID)
	if err != nil {
		return 0, 0, errors.New("订单不存在")
	}
	if orderTID != tid {
		return 0, 0, errors.New("订单不属于该店铺")
	}
	return status, orderID, nil
}

// GetMallPayOrderTID 通过订单号查询所属 tid
func (s *TenantService) GetMallPayOrderTID(outTradeNo string, tid *int) error {
	return database.DB.QueryRow(
		"SELECT tid FROM qingka_mall_pay_order WHERE out_trade_no=?", outTradeNo,
	).Scan(tid)
}

// verifyEpaySign 验证易支付回调签名
func (s *TenantService) verifyEpaySign(params map[string]string, key string) bool {
	sign := params["sign"]
	if sign == "" {
		return false
	}
	keys := make([]string, 0)
	for k, v := range params {
		if k != "sign" && k != "sign_type" && v != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	signStr := strings.Join(parts, "&")
	expected := fmt.Sprintf("%x", md5.Sum([]byte(signStr+key)))
	return expected == sign
}
