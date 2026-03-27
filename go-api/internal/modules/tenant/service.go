package tenant

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"go-api/internal/config"
	"go-api/internal/database"
	"go-api/internal/model"
	commonmodule "go-api/internal/modules/common"
	ordermodule "go-api/internal/modules/order"
)

type Service struct{}

var tenantService = NewService()

const (
	tenantPayModeSiteBalance  = "site_balance"
	tenantPayModeMerchantEpay = "merchant_epay"
)

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
		SELECT c.cid,
		       COALESCE(NULLIF(tp.display_name,''), c.name) AS display_name,
		       COALESCE(NULLIF(c.noun,''), ''),
		       COALESCE(NULLIF(tp.description,''), NULLIF(c.content,''), COALESCE(c.noun,''), ''),
		       COALESCE(tp.cover_url,''),
		       tp.retail_price,
		       COALESCE(c.fenlei,''),
		       COALESCE(tp.category_id,0),
		       COALESCE(mc.name,''),
		       c.sort
		FROM qingka_tenant_product tp
		JOIN qingka_wangke_class c ON c.cid = tp.cid
		LEFT JOIN qingka_tenant_mall_category mc ON mc.id = tp.category_id AND mc.tid = tp.tid AND mc.status=1
		WHERE tp.tid=? AND tp.status=1 AND c.status=1
		ORDER BY COALESCE(mc.sort, 9999) ASC, tp.sort ASC, c.sort ASC`, tid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.MallProduct
	for rows.Next() {
		var p model.MallProduct
		if err := rows.Scan(&p.CID, &p.Name, &p.Noun, &p.Description, &p.CoverURL, &p.RetailPrice, &p.Fenlei, &p.CategoryID, &p.FenleiName, &p.Sort); err != nil {
			return nil, err
		}
		if p.FenleiName == "" {
			p.FenleiName = p.Fenlei
		}
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
		SELECT c.cid,
		       COALESCE(NULLIF(tp.display_name,''), c.name) AS display_name,
		       COALESCE(NULLIF(c.noun,''), ''),
		       COALESCE(NULLIF(tp.description,''), NULLIF(c.content,''), COALESCE(c.noun,''), ''),
		       COALESCE(tp.cover_url,''),
		       tp.retail_price,
		       COALESCE(c.fenlei,''),
		       COALESCE(tp.category_id,0),
		       COALESCE(mc.name,''),
		       c.sort
		FROM qingka_tenant_product tp
		JOIN qingka_wangke_class c ON c.cid = tp.cid
		LEFT JOIN qingka_tenant_mall_category mc ON mc.id = tp.category_id AND mc.tid = tp.tid AND mc.status=1
		WHERE tp.tid=? AND tp.cid=? AND tp.status=1 AND c.status=1`, tid, cid,
	).Scan(&p.CID, &p.Name, &p.Noun, &p.Description, &p.CoverURL, &p.RetailPrice, &p.Fenlei, &p.CategoryID, &p.FenleiName, &p.Sort)
	if err != nil {
		return nil, errors.New("商品不存在")
	}
	if p.FenleiName == "" {
		p.FenleiName = p.Fenlei
	}
	return &p, nil
}

func (s *Service) CUserLogin(tid int, account, password string) (*model.CUser, string, error) {
	var u model.CUser
	var storedPwd string
	err := database.DB.QueryRow(
		"SELECT id,tid,COALESCE(phone,''),account,COALESCE(nickname,''),password,COALESCE(invite_code,''),COALESCE(referrer_id,0),COALESCE(commission_money,0),COALESCE(commission_cdmoney,0),COALESCE(commission_total,0),COALESCE(status,1),addtime FROM qingka_c_user WHERE tid=? AND account=?",
		tid, account,
	).Scan(&u.ID, &u.TID, &u.Phone, &u.Account, &u.Nickname, &storedPwd, &u.InviteCode, &u.ReferrerID, &u.CommissionMoney, &u.CommissionCD, &u.CommissionTotal, &u.Status, &u.AddTime)
	if err != nil {
		return nil, "", errors.New("账号不存在")
	}
	if u.Status != 1 {
		return nil, "", errors.New("账号已禁用")
	}
	if storedPwd != password {
		return nil, "", errors.New("密码错误")
	}
	token := genToken()
	database.DB.Exec("UPDATE qingka_c_user SET token=? WHERE id=?", token, u.ID)
	u.Token = token
	return &u, token, nil
}

func (s *Service) RegisterCUser(tid int, req model.CUserRegisterRequest) (*model.CUser, string, error) {
	t, err := s.GetByTID(tid)
	if err != nil || t.Status != 1 {
		return nil, "", errors.New("店铺不存在或已关闭")
	}
	cfg := parseTenantMallConfig(t.MallConfig)
	if !cfg.RegisterEnabled {
		return nil, "", errors.New("当前商城未开放会员注册")
	}

	account := strings.TrimSpace(req.Account)
	password := strings.TrimSpace(req.Password)
	nickname := strings.TrimSpace(req.Nickname)
	phone := strings.TrimSpace(req.Phone)
	if account == "" || password == "" {
		return nil, "", errors.New("请填写账号和密码")
	}
	if nickname == "" {
		nickname = account
	}

	var exists int
	if err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM qingka_c_user WHERE tid=? AND account=?",
		tid, account,
	).Scan(&exists); err != nil {
		return nil, "", errors.New("注册失败")
	}
	if exists > 0 {
		return nil, "", errors.New("会员账号已存在")
	}

	referrerID := 0
	if cfg.PromotionEnabled {
		if promoter, err := lookupCUserByInviteCode(tid, strings.TrimSpace(req.PromoterCode)); err == nil && promoter.ID > 0 {
			referrerID = promoter.ID
		}
	}

	inviteCode, err := uniqueCUserInviteCode(tid)
	if err != nil {
		return nil, "", err
	}
	token := genToken()
	now := time.Now().Format("2006-01-02 15:04:05")
	res, err := database.DB.Exec(
		`INSERT INTO qingka_c_user (tid,phone,account,password,nickname,invite_code,referrer_id,status,token,addtime)
		 VALUES (?,?,?,?,?,?,?,?,?,?)`,
		tid, phone, account, password, nickname, inviteCode, referrerID, 1, token, now,
	)
	if err != nil {
		return nil, "", errors.New("注册失败")
	}
	id64, _ := res.LastInsertId()
	return &model.CUser{
		ID:         int(id64),
		TID:        tid,
		Phone:      phone,
		Account:    account,
		Nickname:   nickname,
		InviteCode: inviteCode,
		ReferrerID: referrerID,
		Status:     1,
		Token:      token,
		AddTime:    now,
	}, token, nil
}

func (s *Service) GetCUserProfile(tid, cUID int) (*model.CUserProfileResponse, error) {
	var u model.CUser
	err := database.DB.QueryRow(
		`SELECT id,tid,COALESCE(phone,''),account,COALESCE(nickname,''),COALESCE(invite_code,''),COALESCE(referrer_id,0),COALESCE(commission_money,0),COALESCE(commission_cdmoney,0),COALESCE(commission_total,0),COALESCE(status,1),addtime
		 FROM qingka_c_user WHERE tid=? AND id=? LIMIT 1`,
		tid, cUID,
	).Scan(&u.ID, &u.TID, &u.Phone, &u.Account, &u.Nickname, &u.InviteCode, &u.ReferrerID, &u.CommissionMoney, &u.CommissionCD, &u.CommissionTotal, &u.Status, &u.AddTime)
	if err != nil {
		return nil, errors.New("会员不存在")
	}

	var promotionOrders int
	_ = database.DB.QueryRow(
		"SELECT COUNT(*) FROM qingka_mall_pay_order WHERE tid=? AND promoter_c_uid=?",
		tid, cUID,
	).Scan(&promotionOrders)
	referrerAccount, referrerNickname := lookupCUserReferrer(tid, u.ReferrerID)
	tenantInfo, _ := s.GetByTID(tid)
	mallCfg := defaultTenantMallConfig()
	if tenantInfo != nil {
		mallCfg = parseTenantMallConfig(tenantInfo.MallConfig)
	}

	return &model.CUserProfileResponse{
		ID:                u.ID,
		TID:               u.TID,
		Account:           u.Account,
		Nickname:          u.Nickname,
		Phone:             u.Phone,
		InviteCode:        u.InviteCode,
		ReferrerID:        u.ReferrerID,
		ReferrerAccount:   referrerAccount,
		ReferrerNickname:  referrerNickname,
		CommissionMoney:   formatCUserMoney(u.CommissionMoney),
		CommissionCDMoney: formatCUserMoney(u.CommissionCD),
		CommissionTotal:   formatCUserMoney(u.CommissionTotal),
		PromotionOrders:   promotionOrders,
		PromotionEnabled:  mallCfg.PromotionEnabled,
		RegisterEnabled:   mallCfg.RegisterEnabled,
		CommissionRate:    mallCfg.CommissionRate,
		AddTime:           u.AddTime,
		MallConfigPublic:  mallCfg,
	}, nil
}

func (s *Service) CUserPromotionOrders(tid, cUID, page, limit int) ([]model.CUserPromotionOrderItem, int, error) {
	offset := (page - 1) * limit
	var total int
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM qingka_mall_pay_order WHERE tid=? AND promoter_c_uid=?",
		tid, cUID,
	).Scan(&total)

	rows, err := database.DB.Query(`
		SELECT p.id, p.out_trade_no, COALESCE(c.name,''), COALESCE(p.course_name,''), COALESCE(p.account,''), COALESCE(p.money,0), COALESCE(p.commission_amount,0), COALESCE(p.commission_rate,0), COALESCE(p.status,0), COALESCE(p.addtime,''), COALESCE(p.paytime,'')
		FROM qingka_mall_pay_order p
		LEFT JOIN qingka_wangke_class c ON c.cid = p.cid
		WHERE p.tid=? AND p.promoter_c_uid=?
		ORDER BY p.id DESC LIMIT ? OFFSET ?`,
		tid, cUID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]model.CUserPromotionOrderItem, 0)
	for rows.Next() {
		var item model.CUserPromotionOrderItem
		var money, commissionAmount, commissionRate float64
		if err := rows.Scan(&item.ID, &item.OutTradeNo, &item.ProductName, &item.CourseName, &item.BuyerAccount, &money, &commissionAmount, &commissionRate, &item.Status, &item.AddTime, &item.PayTime); err != nil {
			return nil, 0, err
		}
		item.Money = formatCUserMoney(money)
		item.CommissionAmount = formatCUserMoney(commissionAmount)
		item.CommissionRate = formatCUserMoney(commissionRate)
		item.StatusText = formatMallPayOrderStatus(item.Status, 0)
		list = append(list, item)
	}
	return list, total, nil
}

func (s *Service) MergeGuestOrdersToCUser(tid, cUID int, refs []model.MallGuestOrderAccess) ([]string, error) {
	if tid <= 0 || cUID <= 0 || len(refs) == 0 {
		return []string{}, nil
	}

	seen := make(map[string]struct{}, len(refs))
	merged := make([]string, 0, len(refs))

	for _, ref := range refs {
		outTradeNo := strings.TrimSpace(ref.OutTradeNo)
		accessToken := strings.TrimSpace(ref.AccessToken)
		if outTradeNo == "" || accessToken == "" {
			continue
		}
		if _, ok := seen[outTradeNo]; ok {
			continue
		}
		seen[outTradeNo] = struct{}{}

		var order model.MallPayOrder
		err := database.DB.QueryRow(
			`SELECT id, out_trade_no, tid, cid, c_uid, COALESCE(school,''), account, password, remark, pay_type, money, status, COALESCE(order_id,0), COALESCE(trade_no,''), COALESCE(course_items,''), COALESCE(course_id,''), COALESCE(course_name,''), COALESCE(course_kcjs,''), COALESCE(promoter_c_uid,0), COALESCE(promoter_code,''), COALESCE(commission_rate,0), COALESCE(commission_amount,0), COALESCE(commission_status,0)
			 FROM qingka_mall_pay_order WHERE out_trade_no=?`,
			outTradeNo,
		).Scan(&order.ID, &order.OutTradeNo, &order.TID, &order.CID, &order.CUID, &order.School, &order.Account, &order.Password, &order.Remark, &order.PayType, &order.Money, &order.Status, &order.OrderID, &order.TradeNo, &order.CourseItems, &order.CourseID, &order.CourseName, &order.CourseKCJS, &order.PromoterCUID, &order.PromoterCode, &order.CommissionRate, &order.CommissionAmount, &order.CommissionStatus)
		if err != nil {
			continue
		}
		if order.TID != tid {
			continue
		}
		if err := s.validateMallPayAccess(tid, order, accessToken, 0); err != nil {
			continue
		}

		tx, err := database.DB.Begin()
		if err != nil {
			return merged, err
		}

		res, err := tx.Exec(
			"UPDATE qingka_mall_pay_order SET c_uid=?, promoter_c_uid=CASE WHEN COALESCE(promoter_c_uid,0)=? THEN 0 ELSE COALESCE(promoter_c_uid,0) END, promoter_code=CASE WHEN COALESCE(promoter_c_uid,0)=? THEN '' ELSE COALESCE(promoter_code,'') END, commission_amount=CASE WHEN COALESCE(promoter_c_uid,0)=? THEN 0 ELSE COALESCE(commission_amount,0) END, commission_status=CASE WHEN COALESCE(promoter_c_uid,0)=? AND COALESCE(commission_status,0)=0 THEN -1 ELSE COALESCE(commission_status,0) END WHERE id=? AND tid=? AND COALESCE(c_uid,0)=0",
			cUID, cUID, cUID, cUID, cUID, order.ID, tid,
		)
		if err != nil {
			tx.Rollback()
			return merged, err
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			tx.Rollback()
			continue
		}

		if strings.TrimSpace(order.OutTradeNo) != "" {
			if _, err := tx.Exec(
				"UPDATE qingka_wangke_order SET c_uid=? WHERE tid=? AND COALESCE(c_uid,0)=0 AND out_trade_no=?",
				cUID, tid, order.OutTradeNo,
			); err != nil {
				tx.Rollback()
				return merged, err
			}
		}
		if order.OrderID > 0 {
			if _, err := tx.Exec(
				"UPDATE qingka_wangke_order SET c_uid=? WHERE tid=? AND COALESCE(c_uid,0)=0 AND oid=?",
				cUID, tid, order.OrderID,
			); err != nil {
				tx.Rollback()
				return merged, err
			}
		}

		if err := tx.Commit(); err != nil {
			return merged, err
		}
		merged = append(merged, outTradeNo)
	}

	return merged, nil
}

func (s *Service) CUserPayOrders(tid, cUID, page, limit int) ([]model.MallPayOrderItem, int, error) {
	offset := (page - 1) * limit
	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_mall_pay_order WHERE tid=? AND c_uid=?", tid, cUID).Scan(&total)
	rows, err := database.DB.Query(`
		SELECT p.id, p.out_trade_no, COALESCE(p.trade_no,''), p.cid, COALESCE(c.name,''), COALESCE(p.course_name,''), COALESCE(p.school,''), COALESCE(p.account,''), COALESCE(p.remark,''), COALESCE(p.pay_type,''), COALESCE(p.status,0), COALESCE(p.money,'0.00'), COALESCE(p.order_id,0), COALESCE(p.addtime,''), COALESCE(p.paytime,'')
		FROM qingka_mall_pay_order p
		LEFT JOIN qingka_wangke_class c ON c.cid = p.cid
		WHERE p.tid=? AND p.c_uid=?
		ORDER BY p.id DESC LIMIT ? OFFSET ?`, tid, cUID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.MallPayOrderItem
	for rows.Next() {
		var item model.MallPayOrderItem
		rows.Scan(&item.ID, &item.OutTradeNo, &item.TradeNo, &item.CID, &item.ProductName, &item.CourseName, &item.School, &item.Account, &item.Remark, &item.PayType, &item.Status, &item.Money, &item.OrderID, &item.AddTime, &item.PayTime)
		item.StatusText = formatMallPayOrderStatus(item.Status, item.OrderID)
		list = append(list, item)
	}
	if list == nil {
		list = []model.MallPayOrderItem{}
	}
	return list, total, nil
}

func (s *Service) CUserOrderDetail(tid, cUID int, oid int64) (*model.MallOrderItem, error) {
	var item model.MallOrderItem
	err := database.DB.QueryRow(`
		SELECT o.oid, o.cid, COALESCE(c.name,''), COALESCE(o.kcname,''), COALESCE(o.user,''), COALESCE(o.status,''), COALESCE(o.process,''), COALESCE(o.retail_fees,'0.00'), COALESCE(o.addtime,'')
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
		SELECT o.oid, o.cid, COALESCE(c.name,''), COALESCE(o.kcname,''), COALESCE(o.user,''), COALESCE(o.status,''), COALESCE(o.process,''), COALESCE(o.retail_fees,'0.00'), COALESCE(o.addtime,''), COALESCE(o.remarks,'')
		FROM qingka_wangke_order o
		LEFT JOIN qingka_wangke_class c ON c.cid = o.cid
		WHERE o.tid=? AND o.user=? AND COALESCE(o.status,'') NOT IN ('已完成', '已取消')
		ORDER BY o.oid DESC LIMIT 50`, tid, keyword)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.MallOrderItem
	for rows.Next() {
		var item model.MallOrderItem
		rows.Scan(&item.OID, &item.CID, &item.ClassName, &item.KCName, &item.Account, &item.Status, &item.Process, &item.RetailFees, &item.AddTime, &item.Remarks)
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

	payData, err := s.getMallPayData(cfg)
	if err != nil {
		return []model.PayChannel{}, nil
	}

	var channels []model.PayChannel
	if s.mallPayTypeEnabled(cfg, payData, "alipay") {
		channels = append(channels, model.PayChannel{Key: "alipay", Label: "支付宝"})
	}
	if s.mallPayTypeEnabled(cfg, payData, "wxpay") {
		channels = append(channels, model.PayChannel{Key: "wxpay", Label: "微信支付"})
	}
	if s.mallPayTypeEnabled(cfg, payData, "qqpay") {
		channels = append(channels, model.PayChannel{Key: "qqpay", Label: "QQ支付"})
	}
	if channels == nil {
		channels = []model.PayChannel{}
	}
	return channels, nil
}

func (s *Service) resolveMallPromoter(tid, cUID int, promoterCode string, cfg model.TenantMallConfig) (*model.CUser, float64, error) {
	if !cfg.PromotionEnabled {
		return nil, 0, nil
	}
	promoterCode = strings.TrimSpace(promoterCode)
	if promoterCode == "" {
		return nil, 0, nil
	}
	promoter, err := lookupCUserByInviteCode(tid, promoterCode)
	if err != nil || promoter == nil || promoter.ID <= 0 || promoter.Status != 1 {
		return nil, 0, nil
	}
	if cUID > 0 && promoter.ID == cUID {
		return nil, 0, nil
	}
	rate := clampCommissionRate(cfg.CommissionRate)
	if rate <= 0 {
		return nil, 0, nil
	}
	return promoter, rate, nil
}

func (s *Service) CreateMallPayOrder(tid int, req model.MallPayRequest, domain string, cUID int, usePathTID bool) (*model.MallPayCreateResponse, error) {
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
	payData, err := s.getMallPayData(cfg)
	if err != nil {
		if s.getMallPayMode(cfg) == tenantPayModeMerchantEpay {
			return nil, errors.New("店铺暂未配置支付，请联系商家")
		}
		return nil, errors.New("站点暂未配置支付，请联系站长")
	}
	if !s.mallPayTypeEnabled(cfg, payData, req.PayType) {
		return nil, errors.New("支付方式不可用")
	}

	courses := normalizeMallPayCourses(req)
	courseCount := len(courses)
	if courseCount == 0 {
		courseCount = 1
	}
	totalMoney := p.RetailPrice * float64(courseCount)
	totalMoney = float64(int(totalMoney*100+0.5)) / 100
	courseItems, err := json.Marshal(courses)
	if err != nil {
		return nil, errors.New("课程信息处理失败")
	}

	mallCfg := parseTenantMallConfig(t.MallConfig)
	promoter, commissionRate, _ := s.resolveMallPromoter(tid, cUID, req.PromoterCode, mallCfg)
	commissionAmount := 0.0
	if promoter != nil && commissionRate > 0 {
		commissionAmount = float64(int(totalMoney*commissionRate+0.5)) / 100
		if commissionAmount < 0 {
			commissionAmount = 0
		}
	}

	outTradeNo := fmt.Sprintf("%s%d", time.Now().Format("20060102150405"), 111+time.Now().UnixNano()%889)
	money := fmt.Sprintf("%.2f", totalMoney)
	name := "订单号" + outTradeNo
	now := time.Now().Format("2006-01-02 15:04:05")
	accessToken := s.buildMallGuestAccessToken(tid, outTradeNo)

	firstCourse := firstMallCourse(courses)

	_, dbErr := database.DB.Exec(
		`INSERT INTO qingka_mall_pay_order (out_trade_no, tid, cid, c_uid, school, account, password, remark, pay_type, money, status, addtime, course_id, course_name, course_kcjs, course_items, promoter_c_uid, promoter_code, commission_rate, commission_amount, commission_status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)`,
		outTradeNo, tid, req.CID, cUID, strings.TrimSpace(req.School), req.Account, req.Password, req.Remark, req.PayType, totalMoney, now,
		firstCourse.ID, summarizeMallCourseNames(courses), firstCourse.KCJS, string(courseItems),
		func() int {
			if promoter == nil {
				return 0
			}
			return promoter.ID
		}(),
		func() string {
			if promoter == nil {
				return ""
			}
			return promoter.InviteCode
		}(),
		commissionRate,
		commissionAmount,
	)
	if dbErr != nil {
		return nil, errors.New("生成订单失败")
	}

	payURL := s.buildMallEpayURL(payData, outTradeNo, name, money, req.PayType, domain, tid, accessToken, usePathTID)
	return &model.MallPayCreateResponse{
		OutTradeNo:  outTradeNo,
		PayURL:      payURL,
		Money:       money,
		AccessToken: accessToken,
	}, nil
}

func (s *Service) ConfirmMallPayOrder(tid int, params map[string]string) error {
	outTradeNo := params["out_trade_no"]
	tradeStatus := params["trade_status"]

	if tradeStatus != "TRADE_SUCCESS" {
		return errors.New("支付未成功")
	}

	var order model.MallPayOrder
	err := database.DB.QueryRow(
		`SELECT id, out_trade_no, tid, cid, c_uid, COALESCE(school,''), account, password, remark, pay_type, money, status, course_id, course_name, course_kcjs, COALESCE(course_items,''), COALESCE(promoter_c_uid,0), COALESCE(promoter_code,''), COALESCE(commission_rate,0), COALESCE(commission_amount,0), COALESCE(commission_status,0) FROM qingka_mall_pay_order WHERE out_trade_no=?`,
		outTradeNo,
	).Scan(&order.ID, &order.OutTradeNo, &order.TID, &order.CID, &order.CUID, &order.School, &order.Account, &order.Password, &order.Remark, &order.PayType, &order.Money, &order.Status, &order.CourseID, &order.CourseName, &order.CourseKCJS, &order.CourseItems, &order.PromoterCUID, &order.PromoterCode, &order.CommissionRate, &order.CommissionAmount, &order.CommissionStatus)
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
	payData, err := s.getMallPayData(cfg)
	if err != nil {
		return err
	}
	if !s.verifyEpaySign(params, payData["epay_key"]) {
		return errors.New("签名验证失败")
	}

	t, err := s.GetByTID(tid)
	if err != nil {
		return nil
	}
	if err := s.markMallOrderPaid(order, t.UID, params); err != nil {
		return err
	}
	order.Status = 1
	_, _, _ = s.ensureMallOrderSubmitted(tid, &order)
	return nil
}

func (s *Service) UserConfirmMallPay(tid int, outTradeNo, accessToken string, cUID int) (int, int64, error) {
	var order model.MallPayOrder
	err := database.DB.QueryRow(
		`SELECT id, out_trade_no, tid, cid, c_uid, COALESCE(school,''), account, password, remark, pay_type, money, status, course_id, course_name, course_kcjs, COALESCE(course_items,''), COALESCE(promoter_c_uid,0), COALESCE(promoter_code,''), COALESCE(commission_rate,0), COALESCE(commission_amount,0), COALESCE(commission_status,0) FROM qingka_mall_pay_order WHERE out_trade_no=?`,
		outTradeNo,
	).Scan(&order.ID, &order.OutTradeNo, &order.TID, &order.CID, &order.CUID, &order.School, &order.Account, &order.Password, &order.Remark, &order.PayType, &order.Money, &order.Status, &order.CourseID, &order.CourseName, &order.CourseKCJS, &order.CourseItems, &order.PromoterCUID, &order.PromoterCode, &order.CommissionRate, &order.CommissionAmount, &order.CommissionStatus)
	if err != nil {
		return 0, 0, errors.New("订单不存在")
	}
	if order.TID != tid {
		return 0, 0, errors.New("订单不属于该店铺")
	}
	if err := s.validateMallPayAccess(tid, order, accessToken, cUID); err != nil {
		return 0, 0, err
	}
	if order.Status == 2 {
		var orderID int64
		database.DB.QueryRow(`SELECT COALESCE(order_id,0) FROM qingka_mall_pay_order WHERE id=?`, order.ID).Scan(&orderID)
		return 2, orderID, nil
	}
	if order.Status == 0 {
		paid, err := s.syncMallOrderPayment(tid, order)
		if err != nil {
			return 0, 0, err
		}
		if !paid {
			return 0, 0, errors.New("订单未支付，请完成支付后再提交")
		}
		database.DB.QueryRow(
			`SELECT status, COALESCE(order_id,0) FROM qingka_mall_pay_order WHERE id=?`,
			order.ID,
		).Scan(&order.Status, &order.OrderID)
	}
	return s.ensureMallOrderSubmitted(tid, &order)
}

func (s *Service) CheckMallPayStatus(tid int, outTradeNo, accessToken string, cUID int) (int, int64, error) {
	var order model.MallPayOrder
	err := database.DB.QueryRow(
		`SELECT id, out_trade_no, tid, cid, c_uid, COALESCE(school,''), account, password, remark, pay_type, money, status, COALESCE(order_id,0), course_id, course_name, course_kcjs, COALESCE(course_items,''), COALESCE(promoter_c_uid,0), COALESCE(promoter_code,''), COALESCE(commission_rate,0), COALESCE(commission_amount,0), COALESCE(commission_status,0) FROM qingka_mall_pay_order WHERE out_trade_no=?`,
		outTradeNo,
	).Scan(&order.ID, &order.OutTradeNo, &order.TID, &order.CID, &order.CUID, &order.School, &order.Account, &order.Password, &order.Remark, &order.PayType, &order.Money, &order.Status, &order.OrderID, &order.CourseID, &order.CourseName, &order.CourseKCJS, &order.CourseItems, &order.PromoterCUID, &order.PromoterCode, &order.CommissionRate, &order.CommissionAmount, &order.CommissionStatus)
	if err != nil {
		return 0, 0, errors.New("订单不存在")
	}
	if order.TID != tid {
		return 0, 0, errors.New("订单不属于该店铺")
	}
	if err := s.validateMallPayAccess(tid, order, accessToken, cUID); err != nil {
		return 0, 0, err
	}
	if order.Status == 0 {
		paid, err := s.syncMallOrderPayment(tid, order)
		if err != nil {
			return 0, 0, err
		}
		if paid {
			database.DB.QueryRow(
				`SELECT status, COALESCE(order_id,0) FROM qingka_mall_pay_order WHERE id=?`,
				order.ID,
			).Scan(&order.Status, &order.OrderID)
		}
	}
	if order.Status == 1 && order.OrderID == 0 {
		return s.ensureMallOrderSubmitted(tid, &order)
	}
	return order.Status, int64(order.OrderID), nil
}

func (s *Service) GetGuestMallOrder(tid int, outTradeNo, accessToken string) (*model.MallPayOrderItem, error) {
	var payOrder model.MallPayOrder
	err := database.DB.QueryRow(
		`SELECT id, out_trade_no, COALESCE(trade_no,''), tid, cid, c_uid, COALESCE(school,''), account, password, remark, pay_type, money, status, COALESCE(order_id,0), addtime, course_id, course_name, course_kcjs, COALESCE(course_items,''), COALESCE(promoter_c_uid,0), COALESCE(promoter_code,''), COALESCE(commission_rate,0), COALESCE(commission_amount,0), COALESCE(commission_status,0)
		FROM qingka_mall_pay_order WHERE out_trade_no=?`,
		outTradeNo,
	).Scan(&payOrder.ID, &payOrder.OutTradeNo, &payOrder.TradeNo, &payOrder.TID, &payOrder.CID, &payOrder.CUID, &payOrder.School, &payOrder.Account, &payOrder.Password, &payOrder.Remark, &payOrder.PayType, &payOrder.Money, &payOrder.Status, &payOrder.OrderID, &payOrder.AddTime, &payOrder.CourseID, &payOrder.CourseName, &payOrder.CourseKCJS, &payOrder.CourseItems, &payOrder.PromoterCUID, &payOrder.PromoterCode, &payOrder.CommissionRate, &payOrder.CommissionAmount, &payOrder.CommissionStatus)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	if payOrder.TID != tid {
		return nil, errors.New("订单不属于该店铺")
	}
	if payOrder.CUID != 0 {
		return nil, errors.New("该订单需登录后查看")
	}
	if err := s.validateMallPayAccess(tid, payOrder, accessToken, 0); err != nil {
		return nil, err
	}

	item := &model.MallPayOrderItem{
		ID:          payOrder.ID,
		OutTradeNo:  outTradeNo,
		TradeNo:     payOrder.TradeNo,
		CID:         payOrder.CID,
		ProductName: "",
		CourseName:  payOrder.CourseName,
		School:      payOrder.School,
		Account:     payOrder.Account,
		Remark:      payOrder.Remark,
		PayType:     payOrder.PayType,
		Status:      payOrder.Status,
		StatusText:  formatMallPayOrderStatus(payOrder.Status, payOrder.OrderID),
		Money:       fmt.Sprintf("%.2f", payOrder.Money),
		OrderID:     payOrder.OrderID,
		AddTime:     payOrder.AddTime,
	}
	_ = database.DB.QueryRow("SELECT COALESCE(name,'') FROM qingka_wangke_class WHERE cid=?", payOrder.CID).Scan(&item.ProductName)
	return item, nil
}

func formatMallPayOrderStatus(status, orderID int) string {
	switch status {
	case 0:
		return "待支付"
	case 1:
		if orderID < 0 {
			return "下单中"
		}
		return "已支付"
	case 2:
		return "已下单"
	default:
		return "待处理"
	}
}

func (s *Service) buildMallOrderAddRequest(order model.MallPayOrder) model.OrderAddRequest {
	school := strings.TrimSpace(order.School)
	if school == "" {
		school = "自动识别"
	}
	userInfo := fmt.Sprintf("%s %s %s", school, order.Account, order.Password)
	courses := decodeMallOrderCourses(order)
	if len(courses) == 0 {
		courses = []model.OrderAddCourse{{
			ID:   order.CourseID,
			Name: order.CourseName,
			KCJS: order.CourseKCJS,
		}}
	}

	data := make([]model.OrderAddItem, 0, len(courses))
	for _, course := range courses {
		data = append(data, model.OrderAddItem{
			UserInfo: userInfo,
			Data:     course,
		})
	}

	return model.OrderAddRequest{
		CID:  order.CID,
		Data: data,
	}
}

func (s *Service) submitMallOrder(bUID, tid int, order model.MallPayOrder) (*model.OrderAddResult, error) {
	req := s.buildMallOrderAddRequest(order)
	count := len(req.Data)
	if count == 0 {
		count = 1
	}
	unitRetailPrice := order.Money / float64(count)
	return ordermodule.NewServices().Command.AddForMall(bUID, tid, order.CUID, unitRetailPrice, order.OutTradeNo, req)
}

func (s *Service) ensureMallOrderSubmitted(tid int, order *model.MallPayOrder) (int, int64, error) {
	if order == nil {
		return 0, 0, errors.New("订单不存在")
	}
	if order.OrderID > 0 || order.Status == 2 {
		if order.OrderID == 0 {
			database.DB.QueryRow(`SELECT COALESCE(order_id,0) FROM qingka_mall_pay_order WHERE id=?`, order.ID).Scan(&order.OrderID)
		}
		return 2, int64(order.OrderID), nil
	}
	if order.Status == 0 {
		return 0, 0, nil
	}
	if order.OrderID < 0 {
		return 1, 0, nil
	}
	claimed, err := s.claimMallOrderSubmission(order.ID)
	if err != nil {
		return 1, 0, err
	}
	if !claimed {
		database.DB.QueryRow(`SELECT status, COALESCE(order_id,0) FROM qingka_mall_pay_order WHERE id=?`, order.ID).Scan(&order.Status, &order.OrderID)
		if order.OrderID > 0 || order.Status == 2 {
			return 2, int64(order.OrderID), nil
		}
		return 1, 0, nil
	}

	t, err := s.GetByTID(tid)
	if err != nil {
		database.DB.Exec(`UPDATE qingka_mall_pay_order SET order_id=0 WHERE id=? AND order_id=-1`, order.ID)
		return 1, 0, nil
	}
	result, err := s.submitMallOrder(t.UID, tid, *order)
	if err != nil {
		database.DB.Exec(`UPDATE qingka_mall_pay_order SET order_id=0 WHERE id=? AND order_id=-1`, order.ID)
		return 1, 0, err
	}
	if result != nil && result.SuccessCount > 0 {
		firstOID := int64(0)
		if len(result.OIDs) > 0 {
			firstOID = result.OIDs[0]
		}
		database.DB.Exec(`UPDATE qingka_mall_pay_order SET status=2, order_id=? WHERE id=?`, firstOID, order.ID)
		order.Status = 2
		order.OrderID = int(firstOID)
		_ = s.settleMallPromotionCommission(tid, order.ID)
		return 2, firstOID, nil
	}
	database.DB.Exec(`UPDATE qingka_mall_pay_order SET order_id=0 WHERE id=? AND order_id=-1`, order.ID)
	msg := "下单失败"
	if result != nil && len(result.SkippedItems) > 0 {
		msg = strings.Join(result.SkippedItems, "; ")
	}
	return 1, 0, errors.New(msg)
}

func (s *Service) claimMallOrderSubmission(id int) (bool, error) {
	res, err := database.DB.Exec(
		`UPDATE qingka_mall_pay_order
		 SET order_id=-1
		 WHERE id=? AND status=1 AND COALESCE(order_id,0)=0`,
		id,
	)
	if err != nil {
		return false, err
	}
	affected, _ := res.RowsAffected()
	return affected > 0, nil
}

func (s *Service) settleMallPromotionCommission(tid, payOrderID int) error {
	if tid <= 0 || payOrderID <= 0 {
		return nil
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var order model.MallPayOrder
	var payTime string
	err = tx.QueryRow(
		`SELECT id, tid, COALESCE(c_uid,0), COALESCE(account,''), COALESCE(out_trade_no,''), COALESCE(status,0), COALESCE(promoter_c_uid,0), COALESCE(promoter_code,''), COALESCE(commission_rate,0), COALESCE(commission_amount,0), COALESCE(commission_status,0), COALESCE(cid,0), COALESCE(course_name,''), COALESCE(money,0), COALESCE(paytime,''), COALESCE(addtime,'')
		 FROM qingka_mall_pay_order WHERE id=? AND tid=? LIMIT 1`,
		payOrderID, tid,
	).Scan(&order.ID, &order.TID, &order.CUID, &order.Account, &order.OutTradeNo, &order.Status, &order.PromoterCUID, &order.PromoterCode, &order.CommissionRate, &order.CommissionAmount, &order.CommissionStatus, &order.CID, &order.CourseName, &order.Money, &payTime, &order.AddTime)
	if err != nil {
		return err
	}
	_ = payTime
	if order.Status != 2 {
		return nil
	}
	if order.PromoterCUID <= 0 || order.CommissionAmount <= 0 {
		if order.CommissionStatus == 0 {
			_, _ = tx.Exec("UPDATE qingka_mall_pay_order SET commission_status=-1 WHERE id=? AND commission_status=0", order.ID)
			return tx.Commit()
		}
		return nil
	}
	if order.CUID > 0 && order.CUID == order.PromoterCUID {
		if order.CommissionStatus == 0 {
			_, _ = tx.Exec("UPDATE qingka_mall_pay_order SET commission_status=-1 WHERE id=? AND commission_status=0", order.ID)
			return tx.Commit()
		}
		return nil
	}

	res, err := tx.Exec(
		"UPDATE qingka_mall_pay_order SET commission_status=1 WHERE id=? AND tid=? AND status=2 AND COALESCE(commission_status,0)=0",
		order.ID, tid,
	)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return nil
	}

	if _, err := tx.Exec(
		"UPDATE qingka_c_user SET commission_money = commission_money + ?, commission_total = commission_total + ? WHERE id=? AND tid=?",
		order.CommissionAmount, order.CommissionAmount, order.PromoterCUID, tid,
	); err != nil {
		return err
	}

	remark := fmt.Sprintf("推广返利到账 %.2f 元", order.CommissionAmount)
	if _, err := tx.Exec(
		`INSERT INTO qingka_c_user_commission_log (tid, c_uid, pay_order_id, out_trade_no, buyer_c_uid, buyer_account, amount, rate, status, remark, addtime)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		tid, order.PromoterCUID, order.ID, order.OutTradeNo, order.CUID, order.Account, order.CommissionAmount, order.CommissionRate, remark, time.Now().Format("2006-01-02 15:04:05"),
	); err != nil {
		return err
	}

	return tx.Commit()
}

func normalizeMallPayCourses(req model.MallPayRequest) []model.OrderAddCourse {
	if len(req.Courses) > 0 {
		list := make([]model.OrderAddCourse, 0, len(req.Courses))
		for _, course := range req.Courses {
			if strings.TrimSpace(course.ID) == "" && strings.TrimSpace(course.Name) == "" {
				continue
			}
			list = append(list, model.OrderAddCourse{
				ID:   strings.TrimSpace(course.ID),
				Name: strings.TrimSpace(course.Name),
				KCJS: strings.TrimSpace(course.KCJS),
			})
		}
		return list
	}

	ids := splitMallCSV(req.CourseID)
	names := splitMallCSV(req.CourseName)
	kcjsList := splitMallCSV(req.CourseKCJS)
	maxLen := len(ids)
	if len(names) > maxLen {
		maxLen = len(names)
	}
	if len(kcjsList) > maxLen {
		maxLen = len(kcjsList)
	}
	if maxLen == 0 {
		return nil
	}

	list := make([]model.OrderAddCourse, 0, maxLen)
	for i := 0; i < maxLen; i++ {
		course := model.OrderAddCourse{}
		if i < len(ids) {
			course.ID = ids[i]
		}
		if i < len(names) {
			course.Name = names[i]
		}
		if i < len(kcjsList) {
			course.KCJS = kcjsList[i]
		}
		if course.ID == "" && course.Name == "" {
			continue
		}
		list = append(list, course)
	}
	return list
}

func splitMallCSV(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	list := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		list = append(list, part)
	}
	return list
}

func firstMallCourse(courses []model.OrderAddCourse) model.OrderAddCourse {
	if len(courses) == 0 {
		return model.OrderAddCourse{}
	}
	return courses[0]
}

func summarizeMallCourseNames(courses []model.OrderAddCourse) string {
	if len(courses) == 0 {
		return ""
	}
	if len(courses) == 1 {
		return strings.TrimSpace(courses[0].Name)
	}
	first := strings.TrimSpace(courses[0].Name)
	if first == "" {
		first = strings.TrimSpace(courses[0].ID)
	}
	return fmt.Sprintf("%s 等%d门课程", first, len(courses))
}

func decodeMallOrderCourses(order model.MallPayOrder) []model.OrderAddCourse {
	raw := strings.TrimSpace(order.CourseItems)
	if raw != "" {
		var courses []model.OrderAddCourse
		if err := json.Unmarshal([]byte(raw), &courses); err == nil {
			list := make([]model.OrderAddCourse, 0, len(courses))
			for _, course := range courses {
				if strings.TrimSpace(course.ID) == "" && strings.TrimSpace(course.Name) == "" {
					continue
				}
				list = append(list, model.OrderAddCourse{
					ID:   strings.TrimSpace(course.ID),
					Name: strings.TrimSpace(course.Name),
					KCJS: strings.TrimSpace(course.KCJS),
				})
			}
			if len(list) > 0 {
				return list
			}
		}
	}
	if strings.TrimSpace(order.CourseID) == "" && strings.TrimSpace(order.CourseName) == "" {
		return nil
	}
	return []model.OrderAddCourse{{
		ID:   strings.TrimSpace(order.CourseID),
		Name: strings.TrimSpace(order.CourseName),
		KCJS: strings.TrimSpace(order.CourseKCJS),
	}}
}

func (s *Service) buildMallEpayURL(cfg map[string]string, outTradeNo, name, money, payType, domain string, tid int, accessToken string, usePathTID bool) string {
	epayAPI := strings.TrimRight(cfg["epay_api"], "/")
	pid := cfg["epay_pid"]
	key := cfg["epay_key"]
	if epayAPI == "" || pid == "" || key == "" {
		return ""
	}

	notifyURL := "https://" + domain + "/api/v1/mall/pay/notify"
	returnURL := fmt.Sprintf("https://%s/mall/pay-result", domain)
	if usePathTID {
		returnURL = fmt.Sprintf("https://%s/mall/%d/pay-result", domain, tid)
	}
	if accessToken != "" {
		returnURL += "?access_token=" + url.QueryEscape(accessToken)
	}

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

	query := url.Values{}
	for _, k := range keys {
		query.Set(k, params[k])
	}
	query.Set("sign", sign)
	query.Set("sign_type", "MD5")
	return fmt.Sprintf("%s/submit.php?%s", epayAPI, query.Encode())
}

func (s *Service) buildExistingMallPayURL(order model.MallPayOrder, domain string, usePathTID bool) string {
	if order.TID <= 0 || strings.TrimSpace(order.OutTradeNo) == "" || order.Status != 0 {
		return ""
	}
	cfg, err := s.GetTenantPayConfig(order.TID)
	if err != nil {
		return ""
	}
	payData, err := s.getMallPayData(cfg)
	if err != nil {
		return ""
	}
	money := fmt.Sprintf("%.2f", order.Money)
	return s.buildMallEpayURL(
		payData,
		order.OutTradeNo,
		"订单号"+order.OutTradeNo,
		money,
		order.PayType,
		domain,
		order.TID,
		"",
		usePathTID,
	)
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

func (s *Service) getMallPayMode(cfg map[string]string) string {
	mode := strings.TrimSpace(cfg["pay_mode"])
	if mode == "" {
		mode = strings.TrimSpace(cfg["mode"])
	}
	switch mode {
	case tenantPayModeMerchantEpay, tenantPayModeSiteBalance:
		return mode
	}
	if strings.TrimSpace(cfg["epay_api"]) != "" || strings.TrimSpace(cfg["epay_pid"]) != "" || strings.TrimSpace(cfg["epay_key"]) != "" {
		return tenantPayModeMerchantEpay
	}
	return tenantPayModeSiteBalance
}

func (s *Service) getSitePayData() (map[string]string, error) {
	var raw string
	err := database.DB.QueryRow("SELECT COALESCE(paydata,'') FROM qingka_wangke_user WHERE uid = 1").Scan(&raw)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	if strings.TrimSpace(raw) != "" {
		if err := json.Unmarshal([]byte(raw), &result); err != nil {
			return nil, err
		}
	}
	if strings.TrimSpace(result["epay_api"]) == "" || strings.TrimSpace(result["epay_pid"]) == "" || strings.TrimSpace(result["epay_key"]) == "" {
		return nil, errors.New("站点支付配置不完整")
	}
	return result, nil
}

func (s *Service) getMallPayData(cfg map[string]string) (map[string]string, error) {
	if s.getMallPayMode(cfg) == tenantPayModeMerchantEpay {
		if strings.TrimSpace(cfg["epay_api"]) == "" || strings.TrimSpace(cfg["epay_pid"]) == "" || strings.TrimSpace(cfg["epay_key"]) == "" {
			return nil, errors.New("店铺支付配置不完整")
		}
		return cfg, nil
	}
	return s.getSitePayData()
}

func normalizeSwitchValue(v string) bool {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "1", "true", "on", "yes":
		return true
	default:
		return false
	}
}

func (s *Service) mallPayTypeEnabled(cfg, payData map[string]string, payType string) bool {
	key := ""
	switch payType {
	case "alipay":
		key = "is_alipay"
	case "wxpay":
		key = "is_wxpay"
	case "qqpay":
		key = "is_qqpay"
	default:
		return false
	}
	if s.getMallPayMode(cfg) == tenantPayModeSiteBalance {
		return normalizeSwitchValue(payData[key])
	}
	if v, ok := cfg[key]; ok && strings.TrimSpace(v) != "" {
		return normalizeSwitchValue(v)
	}
	return normalizeSwitchValue(payData[key])
}

func (s *Service) queryMallProviderOrder(order model.MallPayOrder, payData map[string]string) (map[string]string, bool, error) {
	epayAPI := strings.TrimRight(strings.TrimSpace(payData["epay_api"]), "/")
	epayPID := strings.TrimSpace(payData["epay_pid"])
	epayKey := strings.TrimSpace(payData["epay_key"])
	if epayAPI == "" || epayPID == "" || epayKey == "" {
		return nil, false, errors.New("支付接口未配置")
	}

	queryURL := fmt.Sprintf("%s/api.php?act=order&pid=%s&key=%s&out_trade_no=%s", epayAPI, epayPID, epayKey, order.OutTradeNo)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(queryURL)
	if err != nil {
		return nil, false, errors.New("查询支付状态失败")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var payResult map[string]interface{}
	_ = json.Unmarshal(body, &payResult)
	data := make(map[string]string, len(payResult))
	for k, v := range payResult {
		data[k] = strings.TrimSpace(fmt.Sprintf("%v", v))
	}
	if !s.matchMallPayQueryResult(data, epayPID, order.OutTradeNo, fmt.Sprintf("%.2f", order.Money)) {
		return nil, false, errors.New("支付结果校验失败")
	}

	status := strings.TrimSpace(data["status"])
	return data, status == "1", nil
}

func (s *Service) matchMallPayQueryResult(data map[string]string, pid, outTradeNo, money string) bool {
	if pid != "" && data["pid"] != "" && strings.TrimSpace(data["pid"]) != pid {
		return false
	}
	if outTradeNo != "" && data["out_trade_no"] != "" && strings.TrimSpace(data["out_trade_no"]) != outTradeNo {
		return false
	}
	if money != "" && data["money"] != "" && strings.TrimSpace(data["money"]) != money {
		return false
	}
	return true
}

func (s *Service) markMallOrderPaid(order model.MallPayOrder, merchantUID int, params map[string]string) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	tx, err := database.DB.Begin()
	if err != nil {
		return errors.New("系统繁忙")
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`UPDATE qingka_mall_pay_order SET status=1, trade_no=?, paytime=? WHERE id=? AND status=0`,
		strings.TrimSpace(params["trade_no"]), now, order.ID,
	)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return nil
	}

	if s.getMallPayModeFromOrder(order.TID) == tenantPayModeSiteBalance {
		if _, err := tx.Exec("UPDATE qingka_wangke_user SET mall_money = mall_money + ? WHERE uid = ?", order.Money, merchantUID); err != nil {
			return errors.New("商家余额入账失败")
		}
		remark := fmt.Sprintf("商城代收到账 %.2f 元[%s]", order.Money, order.OutTradeNo)
		if _, err := tx.Exec(
			"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, '商城代收', ?, (SELECT mall_money FROM qingka_wangke_user WHERE uid = ?), ?, ?)",
			merchantUID, order.Money, merchantUID, remark, now,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Service) getMallPayModeFromOrder(tid int) string {
	cfg, err := s.GetTenantPayConfig(tid)
	if err != nil {
		return tenantPayModeSiteBalance
	}
	return s.getMallPayMode(cfg)
}

func (s *Service) syncMallOrderPayment(tid int, order model.MallPayOrder) (bool, error) {
	cfg, err := s.GetTenantPayConfig(tid)
	if err != nil {
		return false, err
	}
	payData, err := s.getMallPayData(cfg)
	if err != nil {
		return false, err
	}
	payResult, paid, err := s.queryMallProviderOrder(order, payData)
	if err != nil || !paid {
		return paid, err
	}
	t, err := s.GetByTID(tid)
	if err != nil {
		return false, err
	}
	params := map[string]string{
		"trade_no": payResult["trade_no"],
		"type":     payResult["type"],
		"buyer":    payResult["buyer"],
		"money":    payResult["money"],
	}
	if err := s.markMallOrderPaid(order, t.UID, params); err != nil {
		return false, err
	}
	return true, nil
}

func genToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Service) buildMallGuestAccessToken(tid int, outTradeNo string) string {
	secret := "mall-guest-access"
	if config.Global != nil && config.Global.JWT.Secret != "" {
		secret = config.Global.JWT.Secret
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d:%s:%s", tid, outTradeNo, secret)))
	return hex.EncodeToString(sum[:])
}

func (s *Service) validateMallPayAccess(tid int, order model.MallPayOrder, accessToken string, cUID int) error {
	// 优先用 access_token 验证（支持已登录用户支付回跳时 session 未恢复的场景）
	if accessToken != "" {
		if accessToken == s.buildMallGuestAccessToken(tid, order.OutTradeNo) {
			return nil
		}
		return errors.New("订单凭证无效")
	}
	if order.CUID > 0 {
		if cUID == 0 {
			return errors.New("请先登录会员账号")
		}
		if order.CUID != cUID {
			return errors.New("无权访问该订单")
		}
		return nil
	}
	return errors.New("缺少订单凭证")
}

func getAdminConfigValue(key string) string {
	conf, _ := commonmodule.GetAdminConfigMap()
	return conf[key]
}
