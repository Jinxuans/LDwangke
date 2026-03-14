package user

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
	commonmodule "go-api/internal/modules/common"

	"golang.org/x/crypto/bcrypt"
)

type Service struct{}

var userService = NewService()

func NewService() *Service {
	return &Service{}
}

func User() *Service {
	return userService
}

func (s *Service) Profile(uid int, grade string) (*model.UserProfile, error) {
	var p model.UserProfile
	var uuid int
	err := database.DB.QueryRow(
		"SELECT uid, COALESCE(uuid,0), user, COALESCE(name,''), COALESCE(money,0), COALESCE(addprice,1), COALESCE(`key`,''), COALESCE(yqm,''), COALESCE(yqprice,'0'), COALESCE(email,''), COALESCE(tuisongtoken,''), COALESCE(zcz,0) FROM qingka_wangke_user WHERE uid = ?",
		uid,
	).Scan(&p.UID, &uuid, &p.User, &p.Name, &p.Money, &p.AddPrice, &p.Key, &p.YQM, &p.YQPrice, &p.Email, &p.PushToken, &p.ZCZ)
	if err == nil {
		database.DB.QueryRow("SELECT COALESCE(cdmoney,0) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&p.CDMoney)
		if commonmodule.CrossRechargeAllowed(uid) {
			p.KHCZ = 1
		}
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}

	if grade != "2" && grade != "3" && p.AddPrice < 0.1 {
		database.DB.Exec("UPDATE qingka_wangke_user SET addprice='1' WHERE uid = ?", uid)
		p.AddPrice = 1
	}

	rateFormats := []string{
		fmt.Sprintf("%.2f", p.AddPrice),
		fmt.Sprintf("%g", p.AddPrice),
		fmt.Sprintf("%.1f", p.AddPrice),
	}
	for _, rf := range rateFormats {
		database.DB.QueryRow("SELECT COALESCE(name,'') FROM qingka_wangke_dengji WHERE rate = ? AND status = '1' LIMIT 1", rf).Scan(&p.GradeName)
		if p.GradeName != "" {
			break
		}
	}
	if p.GradeName == "" {
		p.GradeName = fmt.Sprintf("费率%g", p.AddPrice)
	}

	todayStart := time.Now().Format("2006-01-02") + " 00:00:00"
	todayEnd := time.Now().Format("2006-01-02") + " 23:59:59"

	orderWhere := fmt.Sprintf("uid='%d'", uid)
	if grade == "2" || grade == "3" {
		orderWhere = "1=1"
	}

	database.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM qingka_wangke_order WHERE %s", orderWhere)).Scan(&p.OrderTotal)
	database.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM qingka_wangke_order WHERE %s AND addtime BETWEEN ? AND ?", orderWhere), todayStart, todayEnd).Scan(&p.TodayOrders)
	database.DB.QueryRow(fmt.Sprintf("SELECT COALESCE(SUM(fees),0) FROM qingka_wangke_order WHERE %s AND addtime BETWEEN ? AND ?", orderWhere), todayStart, todayEnd).Scan(&p.TodaySpend)

	if uuid > 0 {
		database.DB.QueryRow("SELECT COALESCE(user,''), COALESCE(notice,'') FROM qingka_wangke_user WHERE uid = ?", uuid).Scan(&p.SJUser, &p.SJNotice)
	}

	database.DB.QueryRow("SELECT COALESCE(`k`,'') FROM qingka_wangke_config WHERE `v` = 'notice'").Scan(&p.Notice)

	agentWhere := fmt.Sprintf("uuid='%d'", uid)
	if grade == "2" || grade == "3" {
		agentWhere = "1=1"
	}
	var stats model.AgentStats
	database.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM qingka_wangke_user WHERE %s", agentWhere)).Scan(&stats.DLZS)
	database.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM qingka_wangke_user WHERE %s AND endtime > ?", agentWhere), todayStart).Scan(&stats.DLDL)
	database.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM qingka_wangke_user WHERE %s AND addtime > ?", agentWhere), todayStart).Scan(&stats.DLZC)
	database.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM qingka_wangke_order WHERE %s AND addtime > ?", orderWhere), todayStart).Scan(&stats.JRJD)
	p.AgentStats = &stats

	return &p, nil
}

func (s *Service) ChangePassword(uid int, oldPass, newPass string) error {
	var dbPass string
	err := database.DB.QueryRow("SELECT pass FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&dbPass)
	if err != nil {
		return errors.New("用户不存在")
	}

	if oldPass != dbPass {
		return errors.New("旧密码错误")
	}
	if len(newPass) < 6 {
		return errors.New("新密码至少6位")
	}

	_, err = database.DB.Exec("UPDATE qingka_wangke_user SET pass = ? WHERE uid = ?", newPass, uid)
	if err != nil {
		return err
	}

	go func() {
		var username, email string
		database.DB.QueryRow("SELECT COALESCE(user,''), COALESCE(email,'') FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&username, &email)
		if email == "" {
			return
		}
		siteName := s.getSiteName()
		htmlBody := templatePasswordChanged(siteName, username, time.Now().Format("2006-01-02 15:04:05"))
		commonmodule.SendEmail(email, siteName+" - 密码修改通知", htmlBody)
	}()

	return nil
}

func (s *Service) ChangePass2(uid int, oldPass2, newPass2 string) error {
	var grade, dbPass2 string
	err := database.DB.QueryRow("SELECT grade, IFNULL(pass2,'') FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&grade, &dbPass2)
	if err != nil && strings.Contains(err.Error(), "pass2") {
		err = database.DB.QueryRow("SELECT grade FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&grade)
	}
	if err != nil {
		return errors.New("用户不存在")
	}
	if grade != "3" {
		return errors.New("仅管理员可设置二级密码")
	}
	if dbPass2 != "" {
		if strings.HasPrefix(dbPass2, "$2a$") || strings.HasPrefix(dbPass2, "$2b$") {
			if err := bcrypt.CompareHashAndPassword([]byte(dbPass2), []byte(oldPass2)); err != nil {
				return errors.New("旧二级密码错误")
			}
		} else if dbPass2 != oldPass2 {
			return errors.New("旧二级密码错误")
		}
	}
	if len(newPass2) < 6 {
		return errors.New("新二级密码至少6位")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPass2), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("加密失败")
	}
	_, err = database.DB.Exec("UPDATE qingka_wangke_user SET pass2 = ? WHERE uid = ?", string(hashed), uid)
	return err
}

func (s *Service) SendChangeEmailCode(uid int, newEmail string) error {
	identifier := fmt.Sprintf("%d:%s", uid, newEmail)
	if err := commonmodule.CheckVerificationRateLimit("change_email", identifier); err != nil {
		return err
	}

	code, err := commonmodule.GenerateVerificationCode("change_email", identifier, 10*time.Minute)
	if err != nil {
		return err
	}

	siteName := s.getSiteName()
	htmlBody := templateChangeEmailCode(siteName, code, 10)

	if err := commonmodule.SendEmail(newEmail, siteName+" - 邮箱变更验证码", htmlBody); err != nil {
		return fmt.Errorf("发送邮件失败: %v", err)
	}
	return nil
}

func (s *Service) ChangeEmail(uid int, newEmail, code string) error {
	identifier := fmt.Sprintf("%d:%s", uid, newEmail)
	if !commonmodule.VerifyVerificationCode("change_email", identifier, code) {
		return errors.New("验证码错误或已过期")
	}

	var oldEmail, username string
	database.DB.QueryRow("SELECT COALESCE(email,''), COALESCE(user,'') FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&oldEmail, &username)

	_, err := database.DB.Exec("UPDATE qingka_wangke_user SET email = ? WHERE uid = ?", newEmail, uid)
	if err != nil {
		return fmt.Errorf("更新邮箱失败: %v", err)
	}

	if oldEmail != "" && oldEmail != newEmail {
		go func() {
			siteName := s.getSiteName()
			htmlBody := templateEmailChanged(siteName, username, newEmail, time.Now().Format("2006-01-02 15:04:05"))
			commonmodule.SendEmail(oldEmail, siteName+" - 邮箱变更通知", htmlBody)
		}()
	}

	return nil
}

func (s *Service) GetPayChannels(uid int) ([]model.PayChannel, error) {
	payData, parentUID, err := s.getParentPayData(uid)
	if err != nil {
		return nil, err
	}

	if parentUID != 1 && !s.adminConfigEnabled("non_direct_recharge_enable") {
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

func (s *Service) CreatePayOrder(uid int, money float64, payType string, domain string) (*model.PayCreateResponse, error) {
	payData, parentUID, err := s.getParentPayData(uid)
	if err != nil {
		return nil, err
	}

	if parentUID != 1 && !s.adminConfigEnabled("non_direct_recharge_enable") {
		return nil, errors.New("请您根据上面的信息联系上家充值")
	}
	if zdpay := s.getAdminConfigValue("zdpay"); zdpay != "" {
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

func (s *Service) PayOrders(uid int, page, limit int) ([]model.PayOrder, int64, error) {
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

func (s *Service) MoneyLogList(uid int, req model.MoneyLogListRequest) ([]model.MoneyLog, int64, error) {
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

func (s *Service) CheckPayStatus(uid int, outTradeNo string) (bool, string, error) {
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
			bonus, _, _ := s.calcRechargeBonus(money)
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

		bonus, _, _ := s.calcRechargeBonus(money)
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

func (s *Service) TicketList(uid int, grade string, page, limit int) ([]model.Ticket, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	where := "t.uid = ?"
	args := []interface{}{uid}
	if grade == "2" || grade == "3" {
		where = "1=1"
		args = []interface{}{}
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_ticket t WHERE "+where, args...).Scan(&total)

	offset := (page - 1) * limit
	args2 := append(args, limit, offset)
	rows, err := database.DB.Query(
		fmt.Sprintf(`SELECT t.id, t.uid, COALESCE(t.oid,0), COALESCE(t.type,''), COALESCE(t.content,''), COALESCE(t.reply,''),
			t.status, COALESCE(DATE_FORMAT(t.addtime,'%%Y-%%m-%%d %%H:%%i:%%s'),''), COALESCE(DATE_FORMAT(t.reply_time,'%%Y-%%m-%%d %%H:%%i:%%s'),''),
			COALESCE(t.supplier_report_id,0), COALESCE(t.supplier_status,-1), COALESCE(t.supplier_answer,''),
			COALESCE(o.user,''), COALESCE(o.ptname,''), COALESCE(o.status,''), COALESCE(o.yid,'')
		FROM qingka_wangke_ticket t
		LEFT JOIN qingka_wangke_order o ON o.oid = t.oid
		WHERE %s ORDER BY t.id DESC LIMIT ? OFFSET ?`, where),
		args2...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tickets []model.Ticket
	for rows.Next() {
		var t model.Ticket
		rows.Scan(&t.ID, &t.UID, &t.OID, &t.Type, &t.Content, &t.Reply,
			&t.Status, &t.AddTime, &t.ReplyTime,
			&t.SupplierReportID, &t.SupplierStatus, &t.SupplierAnswer,
			&t.OrderUser, &t.OrderPT, &t.OrderStatus, &t.OrderYID)
		tickets = append(tickets, t)
	}
	if tickets == nil {
		tickets = []model.Ticket{}
	}
	return tickets, total, nil
}

func (s *Service) TicketCreate(uid int, req model.TicketCreateRequest) (int64, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_ticket (uid, oid, type, content, status, addtime) VALUES (?, ?, ?, ?, 1, ?)",
		uid, req.OID, req.Type, req.Content, now,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Service) TicketReply(uid int, grade string, req model.TicketReplyRequest) error {
	if grade != "2" && grade != "3" {
		return errors.New("需要管理员权限")
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := database.DB.Exec(
		"UPDATE qingka_wangke_ticket SET reply = ?, status = 2, reply_time = ? WHERE id = ?",
		req.Reply, now, req.ID,
	)
	return err
}

func (s *Service) TicketClose(uid int, grade string, ticketID int) error {
	where := "id = ? AND uid = ?"
	args := []interface{}{ticketID, uid}
	if grade == "2" || grade == "3" {
		where = "id = ?"
		args = []interface{}{ticketID}
	}
	_, err := database.DB.Exec("UPDATE qingka_wangke_ticket SET status = 3 WHERE "+where, args...)
	return err
}

func (s *Service) GetFavorites(uid int) ([]int, error) {
	rows, err := database.DB.Query("SELECT cid FROM qingka_wangke_user_favorite WHERE uid = ? ORDER BY addtime DESC", uid)
	if err != nil {
		return []int{}, err
	}
	defer rows.Close()

	var favorites []int
	for rows.Next() {
		var cid int
		rows.Scan(&cid)
		favorites = append(favorites, cid)
	}
	if favorites == nil {
		favorites = []int{}
	}
	return favorites, nil
}

func (s *Service) AddFavorite(uid, cid int) error {
	var exists int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user_favorite WHERE uid = ? AND cid = ?", uid, cid).Scan(&exists)
	if exists > 0 {
		return errors.New("已收藏过该商品")
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := database.DB.Exec("INSERT INTO qingka_wangke_user_favorite (uid, cid, addtime) VALUES (?, ?, ?)", uid, cid, now)
	return err
}

func (s *Service) RemoveFavorite(uid, cid int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_wangke_user_favorite WHERE uid = ? AND cid = ?", uid, cid)
	return err
}

func (s *Service) SetInviteCode(uid int, yqm string) error {
	if len(yqm) < 4 {
		return errors.New("邀请码最少4位")
	}
	var cnt int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user WHERE yqm = ? AND uid != ?", yqm, uid).Scan(&cnt)
	if cnt > 0 {
		return errors.New("该邀请码已被使用，请换一个")
	}
	_, err := database.DB.Exec("UPDATE qingka_wangke_user SET yqm = ? WHERE uid = ?", yqm, uid)
	return err
}

func (s *Service) SetInviteRate(uid int, yqprice float64, addprice float64) error {
	if yqprice < addprice {
		return errors.New("下级默认费率不能比你低")
	}
	if yqprice > 100 {
		return errors.New("邀请费率不能超过100")
	}
	if int(yqprice*100)%5 != 0 {
		return errors.New("邀请费率必须为0.05的倍数")
	}

	var yqm string
	database.DB.QueryRow("SELECT COALESCE(yqm,'') FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&yqm)
	if yqm == "" {
		yqm = fmt.Sprintf("%05d", time.Now().UnixNano()%100000)
		database.DB.Exec("UPDATE qingka_wangke_user SET yqm = ?, yqprice = ? WHERE uid = ?", yqm, yqprice, uid)
	} else {
		database.DB.Exec("UPDATE qingka_wangke_user SET yqprice = ? WHERE uid = ?", yqprice, uid)
	}
	return nil
}

func (s *Service) ChangeSecretKey(uid int, keyType int, money float64) (string, error) {
	newKey := fmt.Sprintf("%x", time.Now().UnixNano())
	if len(newKey) > 16 {
		newKey = newKey[:16]
	}

	if keyType == 1 {
		var currentKey string
		database.DB.QueryRow("SELECT COALESCE(`key`,'') FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&currentKey)

		if money >= 100 {
			database.DB.Exec("UPDATE qingka_wangke_user SET `key` = ? WHERE uid = ?", newKey, uid)
			return newKey, nil
		}
		if money >= 5 {
			database.DB.Exec("UPDATE qingka_wangke_user SET `key` = ?, money = money - 5 WHERE uid = ?", newKey, uid)
			now := time.Now().Format("2006-01-02 15:04:05")
			database.DB.Exec(
				"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, '扣费', -5, (SELECT money FROM qingka_wangke_user WHERE uid = ?), '开通接口扣费5元', ?)",
				uid, uid, now,
			)
			return newKey, nil
		}
		return "", errors.New("余额不足，需要5元开通费用")
	}
	if keyType == 3 {
		var currentKey string
		database.DB.QueryRow("SELECT COALESCE(`key`,'') FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&currentKey)
		if currentKey == "" || currentKey == "0" {
			return "", errors.New("请先开通API密钥")
		}
		database.DB.Exec("UPDATE qingka_wangke_user SET `key` = ? WHERE uid = ?", newKey, uid)
		return newKey, nil
	}
	return "", errors.New("未知操作类型")
}

func (s *Service) LogList(uid int, grade string, page, limit int, logType, keywords string) ([]map[string]interface{}, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	where := "1=1"
	args := []interface{}{}
	if grade != "2" && grade != "3" {
		where = "uid = ?"
		args = append(args, uid)
	}

	if keywords != "" {
		switch logType {
		case "uid":
			where += " AND uid = ?"
			args = append(args, keywords)
		case "type":
			where += " AND type = ?"
			args = append(args, keywords)
		case "text":
			where += " AND text LIKE ?"
			args = append(args, "%"+keywords+"%")
		case "money":
			where += " AND money = ?"
			args = append(args, keywords)
		case "ip":
			where += " AND ip = ?"
			args = append(args, keywords)
		default:
			where += " AND (type LIKE ? OR text LIKE ? OR uid LIKE ?)"
			k := "%" + keywords + "%"
			args = append(args, k, k, k)
		}
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_log WHERE "+where, args...).Scan(&total)

	offset := (page - 1) * limit
	args2 := append(args, limit, offset)
	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT id, uid, COALESCE(type,''), COALESCE(text,''), COALESCE(money,0), COALESCE(smoney,0), COALESCE(ip,''), COALESCE(addtime,'') FROM qingka_wangke_log WHERE %s ORDER BY id DESC LIMIT ? OFFSET ?", where),
		args2...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var id, rowUID int
		var logType2, text, ip, addtime string
		var money, smoney float64
		rows.Scan(&id, &rowUID, &logType2, &text, &money, &smoney, &ip, &addtime)
		list = append(list, map[string]interface{}{
			"id":      id,
			"uid":     rowUID,
			"type":    logType2,
			"text":    text,
			"money":   money,
			"smoney":  smoney,
			"ip":      ip,
			"addtime": addtime,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	return list, total, nil
}

func (s *Service) getRechargeBonusConfig() *commonmodule.RechargeBonusConfig {
	return commonmodule.LoadRechargeBonusConfig()
}

func (s *Service) calcRechargeBonus(money float64) (bonus float64, pct float64, isActivity bool) {
	cfg := s.getRechargeBonusConfig()
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

func (s *Service) getParentPayData(uid int) (map[string]string, int, error) {
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

func (s *Service) buildEpayURL(payData map[string]string, outTradeNo, name, money, payType, domain string) string {
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

func (s *Service) adminConfigEnabled(key string) bool {
	return s.getAdminConfigValue(key) == "1"
}

func (s *Service) getAdminConfigValue(key string) string {
	conf, _ := commonmodule.GetAdminConfigMap()
	return conf[key]
}

func (s *Service) getSiteName() string {
	siteName := s.getAdminConfigValue("sitename")
	if siteName == "" {
		siteName = "System"
	}
	return siteName
}

func templatePasswordChanged(siteName, username, changeTime string) string {
	return fmt.Sprintf(`<html><body><h2>%s</h2><p>账号 <strong>%s</strong> 的登录密码已于 <strong>%s</strong> 成功修改。</p></body></html>`, siteName, username, changeTime)
}

func templateChangeEmailCode(siteName, code string, expireMinutes int) string {
	return fmt.Sprintf(`<html><body><h2>%s</h2><p>您正在变更绑定邮箱，请使用以下验证码完成验证：</p><p style="font-size:24px;font-weight:bold;">%s</p><p>验证码 %d 分钟内有效。</p></body></html>`, siteName, code, expireMinutes)
}

func templateEmailChanged(siteName, username, newEmail, changeTime string) string {
	return fmt.Sprintf(`<html><body><h2>%s</h2><p>账号 <strong>%s</strong> 的绑定邮箱已于 <strong>%s</strong> 变更为 <strong>%s</strong>。</p></body></html>`, siteName, username, changeTime, newEmail)
}
