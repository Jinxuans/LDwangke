package tenant

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
	commonmodule "go-api/internal/modules/common"
	ordermodule "go-api/internal/modules/order"
)

type Service struct{}

var tenantService = NewService()

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetMallOpenPrice() float64 {
	if v := getAdminConfigValue("mall_open_price"); v != "" {
		var price float64
		fmt.Sscanf(v, "%f", &price)
		if price > 0 {
			return price
		}
	}
	return 99
}

func (s *Service) OpenMall(uid int, shopName string) (int64, error) {
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

func (s *Service) GetByTID(tid int) (*model.Tenant, error) {
	return lookupTenantByTID(tid)
}

func (s *Service) MallProducts(tid int) ([]model.MallProduct, error) {
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
	if list == nil {
		list = []model.MallProduct{}
	}
	return list, nil
}

func (s *Service) MallProductDetail(tid, cid int) (*model.MallProduct, error) {
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

func (s *Service) CUserLogin(tid int, account, password string) (*model.CUser, string, error) {
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

func (s *Service) CUserOrders(tid, cUID, page, limit int) ([]model.MallOrderItem, int, error) {
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
	if list == nil {
		list = []model.MallOrderItem{}
	}
	return list, total, nil
}

func (s *Service) CUserOrderDetail(tid, cUID int, oid int64) (*model.MallOrderItem, error) {
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

func (s *Service) SearchMallOrders(tid int, keyword string) ([]model.MallOrderItem, error) {
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
	if list == nil {
		list = []model.MallOrderItem{}
	}
	return list, nil
}

func (s *Service) GetTenantPayConfig(tid int) (map[string]string, error) {
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

func (s *Service) GetMallPayChannels(tid int) ([]model.PayChannel, error) {
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

func (s *Service) CreateMallPayOrder(tid int, req model.MallPayRequest, domain string, cUID int) (*model.MallPayCreateResponse, error) {
	t, err := s.GetByTID(tid)
	if err != nil || t.Status != 1 {
		return nil, errors.New("店铺不存在或已关闭")
	}

	p, err := s.MallProductDetail(tid, req.CID)
	if err != nil {
		return nil, errors.New("商品不存在")
	}

	cfg, err := s.GetTenantPayConfig(tid)
	if err != nil {
		return nil, err
	}
	if cfg["epay_api"] == "" || cfg["epay_pid"] == "" || cfg["epay_key"] == "" {
		return nil, errors.New("店铺暂未配置支付，请联系商家")
	}

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

func (s *Service) ConfirmMallPayOrder(tid int, params map[string]string) error {
	outTradeNo := params["out_trade_no"]
	tradeNo := params["trade_no"]
	tradeStatus := params["trade_status"]

	if tradeStatus != "TRADE_SUCCESS" {
		return errors.New("支付未成功")
	}

	var order model.MallPayOrder
	err := database.DB.QueryRow(
		`SELECT id, tid, cid, c_uid, account, password, remark, pay_type, money, status, course_id, course_name, course_kcjs FROM qingka_mall_pay_order WHERE out_trade_no=?`,
		outTradeNo,
	).Scan(&order.ID, &order.TID, &order.CID, &order.CUID, &order.Account, &order.Password, &order.Remark, &order.PayType, &order.Money, &order.Status, &order.CourseID, &order.CourseName, &order.CourseKCJS)
	if err != nil {
		return errors.New("订单不存在")
	}
	if order.Status != 0 {
		return nil
	}
	if order.TID != tid {
		return errors.New("订单不属于该店铺")
	}

	cfg, _ := s.GetTenantPayConfig(tid)
	if !s.verifyEpaySign(params, cfg["epay_key"]) {
		return errors.New("签名验证失败")
	}

	database.DB.Exec(
		`UPDATE qingka_mall_pay_order SET status=1, trade_no=?, paytime=? WHERE id=?`,
		tradeNo, time.Now().Format("2006-01-02 15:04:05"), order.ID,
	)

	t, err := s.GetByTID(tid)
	if err != nil {
		return nil
	}
	result, err := s.submitMallOrder(t.UID, tid, order)
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

func (s *Service) UserConfirmMallPay(tid int, outTradeNo string) (int, int64, error) {
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
	if order.Status == 2 {
		var orderID int64
		database.DB.QueryRow(`SELECT COALESCE(order_id,0) FROM qingka_mall_pay_order WHERE id=?`, order.ID).Scan(&orderID)
		return 2, orderID, nil
	}
	if order.Status == 0 {
		database.DB.Exec(
			`UPDATE qingka_mall_pay_order SET status=1, paytime=? WHERE id=?`,
			time.Now().Format("2006-01-02 15:04:05"), order.ID,
		)
		order.Status = 1
	}

	t, err := s.GetByTID(tid)
	if err != nil {
		return 1, 0, nil
	}
	result, err := s.submitMallOrder(t.UID, tid, order)
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

func (s *Service) CheckMallPayStatus(tid int, outTradeNo string) (int, int64, error) {
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

func (s *Service) buildMallOrderAddRequest(order model.MallPayOrder) model.OrderAddRequest {
	userInfo := fmt.Sprintf("自动识别 %s %s", order.Account, order.Password)
	item := model.OrderAddItem{UserInfo: userInfo}
	if order.CourseID != "" {
		item.Data = model.OrderAddCourse{
			ID:   order.CourseID,
			Name: order.CourseName,
			KCJS: order.CourseKCJS,
		}
	}

	return model.OrderAddRequest{
		CID:  order.CID,
		Data: []model.OrderAddItem{item},
	}
}

func (s *Service) submitMallOrder(bUID, tid int, order model.MallPayOrder) (*model.OrderAddResult, error) {
	return ordermodule.NewServices().Command.AddForMall(bUID, tid, order.CUID, order.Money, s.buildMallOrderAddRequest(order))
}

func (s *Service) buildMallEpayURL(cfg map[string]string, outTradeNo, name, money, payType, domain string, tid int) string {
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

func (s *Service) verifyEpaySign(params map[string]string, key string) bool {
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

func genToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func getAdminConfigValue(key string) string {
	conf, _ := commonmodule.GetAdminConfigMap()
	return conf[key]
}
