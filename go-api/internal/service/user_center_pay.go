package service

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

type RechargeBonusRule struct {
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	BonusPct float64 `json:"bonus_pct"`
}

type RechargeBonusActivity struct {
	Enabled  bool                `json:"enabled"`
	Weekdays []int               `json:"weekdays"`
	Rules    []RechargeBonusRule `json:"rules"`
	Hint     string              `json:"hint"`
}

type RechargeBonusConfig struct {
	Enabled  bool                  `json:"enabled"`
	Rules    []RechargeBonusRule   `json:"rules"`
	Activity RechargeBonusActivity `json:"activity"`
}

func (s *UserCenterService) GetRechargeBonusConfig() *RechargeBonusConfig {
	var raw string
	database.DB.QueryRow("SELECT `k` FROM qingka_wangke_config WHERE `v` = 'recharge_bonus_rules'").Scan(&raw)
	if raw == "" {
		return &RechargeBonusConfig{}
	}
	var cfg RechargeBonusConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return &RechargeBonusConfig{}
	}
	return &cfg
}

func (s *UserCenterService) CalcRechargeBonus(money float64) (bonus float64, pct float64, isActivity bool) {
	cfg := s.GetRechargeBonusConfig()
	if !cfg.Enabled || len(cfg.Rules) == 0 {
		return 0, 0, false
	}

	rules := cfg.Rules
	if cfg.Activity.Enabled && len(cfg.Activity.Weekdays) > 0 && len(cfg.Activity.Rules) > 0 {
		weekday := int(time.Now().Weekday())
		for _, w := range cfg.Activity.Weekdays {
			if weekday == w {
				isActivity = true
				rules = cfg.Activity.Rules
				break
			}
		}
	}

	for _, r := range rules {
		if money >= r.Min && money < r.Max {
			pct = r.BonusPct
			break
		}
	}
	if pct <= 0 {
		return 0, 0, isActivity
	}

	bonus = math.Round(money*pct) / 100
	return bonus, pct, isActivity
}

func (s *UserCenterService) CheckPayStatus(uid int, outTradeNo string) (bool, string, error) {
	if outTradeNo == "" {
		return false, "", errors.New("订单号不能为空")
	}

	var oid int
	var moneyStr string
	var status int
	err := database.DB.QueryRow(
		"SELECT oid, money, status FROM qingka_wangke_pay WHERE out_trade_no = ? AND uid = ? LIMIT 1",
		outTradeNo, uid,
	).Scan(&oid, &moneyStr, &status)
	if err != nil {
		return false, "", errors.New("订单不存在")
	}

	var money float64
	fmt.Sscanf(moneyStr, "%f", &money)

	if status >= 1 {
		var logCount int
		database.DB.QueryRow(
			"SELECT COUNT(*) FROM qingka_wangke_moneylog WHERE uid = ? AND remark LIKE ?",
			uid, "%"+outTradeNo+"%",
		).Scan(&logCount)
		if logCount == 0 && money > 0 {
			now := time.Now().Format("2006-01-02 15:04:05")
			bonus, _, _ := s.CalcRechargeBonus(money)
			total := money + bonus
			database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ?, zcz = zcz + ? WHERE uid = ?", total, money, uid)
			database.DB.Exec(
				"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, '充值', ?, (SELECT money FROM qingka_wangke_user WHERE uid = ?), ?, ?)",
				uid, money, uid, fmt.Sprintf("在线充值%.2f元[%s]", money, outTradeNo), now,
			)
			if bonus > 0 {
				database.DB.Exec(
					"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, '充值赠送', ?, (SELECT money FROM qingka_wangke_user WHERE uid = ?), ?, ?)",
					uid, bonus, uid, fmt.Sprintf("充值%.2f元赠送%.2f元", money, bonus), now,
				)
			}
			return true, fmt.Sprintf("支付成功，已到账 ¥%.2f", total), nil
		}
		return true, "订单已支付", nil
	}

	payData, _, err := s.getParentPayData(uid)
	if err != nil {
		return false, "", errors.New("支付接口未配置")
	}

	epayAPI := strings.TrimRight(payData["epay_api"], "/")
	epayPID := payData["epay_pid"]
	epayKey := payData["epay_key"]

	if epayAPI == "" || epayPID == "" || epayKey == "" {
		return false, "", errors.New("支付接口未配置")
	}

	queryURL := fmt.Sprintf("%s/api.php?act=order&pid=%s&key=%s&out_trade_no=%s", epayAPI, epayPID, epayKey, outTradeNo)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(queryURL)
	if err != nil {
		return false, "", errors.New("查询支付状态失败")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var payResult map[string]interface{}
	json.Unmarshal(body, &payResult)

	var epayStatus int
	switch v := payResult["status"].(type) {
	case float64:
		epayStatus = int(v)
	case string:
		fmt.Sscanf(v, "%d", &epayStatus)
	}

	if epayStatus == 1 {
		now := time.Now().Format("2006-01-02 15:04:05")
		tx, txErr := database.DB.Begin()
		if txErr != nil {
			return false, "", errors.New("系统繁忙")
		}
		defer tx.Rollback()

		res, _ := tx.Exec("UPDATE qingka_wangke_pay SET status = 1, endtime = ? WHERE out_trade_no = ? AND status = 0", now, outTradeNo)
		affected, _ := res.RowsAffected()
		if affected == 0 {
			return true, "订单已支付", nil
		}

		bonus, _, _ := s.CalcRechargeBonus(money)
		total := money + bonus

		if _, err := tx.Exec("UPDATE qingka_wangke_user SET money = money + ?, zcz = zcz + ? WHERE uid = ?", total, money, uid); err != nil {
			return false, "", fmt.Errorf("余额更新失败，请联系客服")
		}

		tx.Exec(
			"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, '充值', ?, (SELECT money FROM qingka_wangke_user WHERE uid = ?), ?, ?)",
			uid, money, uid, fmt.Sprintf("在线充值%.2f元[%s]", money, outTradeNo), now,
		)
		if bonus > 0 {
			tx.Exec(
				"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, '充值赠送', ?, (SELECT money FROM qingka_wangke_user WHERE uid = ?), ?, ?)",
				uid, bonus, uid, fmt.Sprintf("充值%.2f元赠送%.2f元", money, bonus), now,
			)
		}

		if err := tx.Commit(); err != nil {
			return false, "", errors.New("到账失败，请联系客服")
		}
		msg := fmt.Sprintf("支付成功，已到账 ¥%.2f", money)
		if bonus > 0 {
			msg = fmt.Sprintf("支付成功，充值 ¥%.2f + 赠送 ¥%.2f = 到账 ¥%.2f", money, bonus, total)
		}
		return true, msg, nil
	}

	return false, "订单未支付，请完成支付后再刷新", nil
}

func (s *UserCenterService) getParentPayData(uid int) (map[string]string, int, error) {
	var parentUID int
	err := database.DB.QueryRow("SELECT uuid FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&parentUID)
	if err != nil {
		return nil, 0, errors.New("用户不存在")
	}

	var paydata string
	err = database.DB.QueryRow("SELECT COALESCE(paydata,'') FROM qingka_wangke_user WHERE uid = ?", parentUID).Scan(&paydata)
	if err != nil {
		return nil, parentUID, errors.New("上级用户不存在")
	}

	result := make(map[string]string)
	if paydata != "" {
		json.Unmarshal([]byte(paydata), &result)
	}
	return result, parentUID, nil
}

func (s *UserCenterService) GetPayChannels(uid int) ([]model.PayChannel, error) {
	payData, parentUID, err := s.getParentPayData(uid)
	if err != nil {
		return nil, err
	}

	if parentUID != 1 && !adminConfigEnabled("non_direct_recharge_enable") {
		return []model.PayChannel{}, nil
	}

	var channels []model.PayChannel
	if payData["is_alipay"] == "1" {
		channels = append(channels, model.PayChannel{Key: "alipay", Label: "支付宝"})
	}
	if payData["is_wxpay"] == "1" {
		channels = append(channels, model.PayChannel{Key: "wxpay", Label: "微信支付"})
	}
	if payData["is_qqpay"] == "1" {
		channels = append(channels, model.PayChannel{Key: "qqpay", Label: "QQ支付"})
	}
	if payData["is_usdt"] == "1" {
		channels = append(channels, model.PayChannel{Key: "usdt", Label: "USDT"})
	}
	if channels == nil {
		channels = []model.PayChannel{}
	}
	return channels, nil
}

func (s *UserCenterService) CreatePayOrder(uid int, money float64, payType string, domain string) (*model.PayCreateResponse, error) {
	payData, parentUID, err := s.getParentPayData(uid)
	if err != nil {
		return nil, err
	}

	if parentUID != 1 && !adminConfigEnabled("non_direct_recharge_enable") {
		return nil, errors.New("请您根据上面的信息联系上家充值")
	}
	if zdpay := getAdminConfigValue("zdpay"); zdpay != "" {
		var minPay float64
		fmt.Sscanf(zdpay, "%f", &minPay)
		if minPay > 0 && money < minPay {
			return nil, fmt.Errorf("在线充值最低%.0f元", minPay)
		}
	}

	outTradeNo := fmt.Sprintf("%s%d", time.Now().Format("20060102150405"), 111+time.Now().UnixNano()%889)
	moneyStr := fmt.Sprintf("%.2f", money)
	name := fmt.Sprintf("零食购买-%s", moneyStr)
	now := time.Now().Format("2006-01-02 15:04:05")

	result, dbErr := database.DB.Exec(
		"INSERT INTO qingka_wangke_pay (out_trade_no, trade_no, uid, num, name, money, ip, addtime, domain, status) VALUES (?, '', ?, ?, ?, ?, '', ?, ?, 0)",
		outTradeNo, uid, moneyStr, name, moneyStr, now, domain,
	)
	if dbErr != nil {
		return nil, errors.New("生成订单失败")
	}

	oid, _ := result.LastInsertId()
	payURL := s.buildEpayURL(payData, outTradeNo, name, moneyStr, payType, domain)

	return &model.PayCreateResponse{
		OID:        int(oid),
		OutTradeNo: outTradeNo,
		Money:      moneyStr,
		PayURL:     payURL,
	}, nil
}

func (s *UserCenterService) buildEpayURL(payData map[string]string, outTradeNo, name, money, payType, domain string) string {
	epayAPI := strings.TrimRight(payData["epay_api"], "/")
	pid := payData["epay_pid"]
	key := payData["epay_key"]

	if epayAPI == "" || pid == "" || key == "" {
		return ""
	}

	notifyURL := "https://" + domain + "/epay/notify_url.php"
	returnURL := "https://" + domain + "/epay/return_url.php"

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
	for k := range params {
		if params[k] != "" {
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

func (s *UserCenterService) PayOrders(uid int, page, limit int) ([]model.PayOrder, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_pay WHERE uid = ?", uid).Scan(&total)

	offset := (page - 1) * limit
	rows, err := database.DB.Query(
		"SELECT oid, out_trade_no, uid, money, status, COALESCE(DATE_FORMAT(addtime,'%Y-%m-%d %H:%i:%s'),'') FROM qingka_wangke_pay WHERE uid = ? ORDER BY oid DESC LIMIT ? OFFSET ?",
		uid, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []model.PayOrder
	for rows.Next() {
		var o model.PayOrder
		rows.Scan(&o.OID, &o.OutTradeNo, &o.UID, &o.Money, &o.Status, &o.AddTime)
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []model.PayOrder{}
	}
	return orders, total, nil
}

func (s *UserCenterService) MoneyLogList(uid int, req model.MoneyLogListRequest) ([]model.MoneyLog, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	where := "uid = ?"
	args := []interface{}{uid}
	if req.Type != "" {
		where += " AND type = ?"
		args = append(args, req.Type)
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_moneylog WHERE "+where, args...).Scan(&total)

	offset := (req.Page - 1) * req.Limit
	args2 := append(args, req.Limit, offset)
	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT id, uid, COALESCE(type,''), COALESCE(money,0), COALESCE(balance,0), COALESCE(remark,''), COALESCE(DATE_FORMAT(addtime,'%%Y-%%m-%%d %%H:%%i:%%s'),'') FROM qingka_wangke_moneylog WHERE %s ORDER BY id DESC LIMIT ? OFFSET ?", where),
		args2...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []model.MoneyLog
	for rows.Next() {
		var l model.MoneyLog
		rows.Scan(&l.ID, &l.UID, &l.Type, &l.Money, &l.Balance, &l.Remark, &l.AddTime)
		logs = append(logs, l)
	}
	if logs == nil {
		logs = []model.MoneyLog{}
	}
	return logs, total, nil
}
