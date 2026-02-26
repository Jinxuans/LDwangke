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

	"golang.org/x/crypto/bcrypt"
)

type UserCenterService struct{}

func NewUserCenterService() *UserCenterService {
	return &UserCenterService{}
}

// 用户资料 (按 PHP info case)
func (s *UserCenterService) Profile(uid int, grade string) (*model.UserProfile, error) {
	var p model.UserProfile
	var uuid int
	err := database.DB.QueryRow(
		"SELECT uid, COALESCE(uuid,0), user, COALESCE(name,''), COALESCE(money,0), COALESCE(addprice,1), COALESCE(`key`,''), COALESCE(yqm,''), COALESCE(yqprice,'0'), COALESCE(email,''), COALESCE(tuisongtoken,''), COALESCE(zcz,0) FROM qingka_wangke_user WHERE uid = ?",
		uid,
	).Scan(&p.UID, &uuid, &p.User, &p.Name, &p.Money, &p.AddPrice, &p.Key, &p.YQM, &p.YQPrice, &p.Email, &p.PushToken, &p.ZCZ)
	if err == nil {
		// 可选字段：cdmoney（表中可能不存在）
		database.DB.QueryRow("SELECT COALESCE(cdmoney,0) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&p.CDMoney)
		// 跨户充值权限：基于系统配置 cross_recharge_uids
		if NewAgentService().CrossRechargeAllowed(uid) {
			p.KHCZ = 1
		}
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}

	// 费率安全校验（按 PHP: 非管理员 addprice < 0.1 则重置为1）
	if grade != "2" && grade != "3" && p.AddPrice < 0.1 {
		database.DB.Exec("UPDATE qingka_wangke_user SET addprice='1' WHERE uid = ?", uid)
		p.AddPrice = 1
	}

	// 解析等级名称 (从 dengji 表根据 rate 匹配，兼容 "1"/"1.00"/"1.0" 等格式)
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

	// 管理员看全局，普通用户看自己（按 PHP info case）
	orderWhere := fmt.Sprintf("uid='%d'", uid)
	if grade == "2" || grade == "3" {
		orderWhere = "1=1"
	}

	database.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM qingka_wangke_order WHERE %s", orderWhere)).Scan(&p.OrderTotal)
	database.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM qingka_wangke_order WHERE %s AND addtime BETWEEN ? AND ?", orderWhere), todayStart, todayEnd).Scan(&p.TodayOrders)
	database.DB.QueryRow(fmt.Sprintf("SELECT COALESCE(SUM(fees),0) FROM qingka_wangke_order WHERE %s AND addtime BETWEEN ? AND ?", orderWhere), todayStart, todayEnd).Scan(&p.TodaySpend)

	// 上级用户（按 PHP: $a = $DB->get_row("SELECT uid,user,notice FROM qingka_wangke_user WHERE uid='{$userrow['uuid']}'")）
	if uuid > 0 {
		database.DB.QueryRow(
			"SELECT COALESCE(user,''), COALESCE(notice,'') FROM qingka_wangke_user WHERE uid = ?", uuid,
		).Scan(&p.SJUser, &p.SJNotice)
	}

	// 系统公告（按 PHP: $conf['notice']）
	database.DB.QueryRow("SELECT COALESCE(`k`,'') FROM qingka_wangke_config WHERE `v` = 'notice'").Scan(&p.Notice)

	// 代理统计（按 PHP dailitongji）
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

// ===== 课程收藏 (按 PHP favorites case) =====

func (s *UserCenterService) GetFavorites(uid int) ([]int, error) {
	rows, err := database.DB.Query(
		"SELECT cid FROM qingka_wangke_user_favorite WHERE uid = ? ORDER BY addtime DESC", uid,
	)
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

func (s *UserCenterService) AddFavorite(uid, cid int) error {
	var exists int
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM qingka_wangke_user_favorite WHERE uid = ? AND cid = ?", uid, cid,
	).Scan(&exists)
	if exists > 0 {
		return errors.New("已收藏过该商品")
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_user_favorite (uid, cid, addtime) VALUES (?, ?, ?)", uid, cid, now,
	)
	return err
}

func (s *UserCenterService) RemoveFavorite(uid, cid int) error {
	_, err := database.DB.Exec(
		"DELETE FROM qingka_wangke_user_favorite WHERE uid = ? AND cid = ?", uid, cid,
	)
	return err
}

// ===== 充值赠送规则 =====

type RechargeBonusRule struct {
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	BonusPct float64 `json:"bonus_pct"` // 赠送百分比，如 5 表示 5%
}

type RechargeBonusActivity struct {
	Enabled  bool                `json:"enabled"`
	Weekdays []int               `json:"weekdays"` // 星期几是活动日，0=周日 1=周一 ... 6=周六
	Rules    []RechargeBonusRule `json:"rules"`    // 活动日独立的赠送规则（替换普通规则）
	Hint     string              `json:"hint"`     // 活动日自定义提示文案
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

	// 检查是否活动日（按星期几判断）
	rules := cfg.Rules
	if cfg.Activity.Enabled && len(cfg.Activity.Weekdays) > 0 && len(cfg.Activity.Rules) > 0 {
		weekday := int(time.Now().Weekday()) // 0=Sunday ... 6=Saturday
		for _, w := range cfg.Activity.Weekdays {
			if weekday == w {
				isActivity = true
				rules = cfg.Activity.Rules // 活动日使用独立规则
				break
			}
		}
	}

	// 找到匹配的区间
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

// ===== 支付状态检测 (按 PHP check_pay_status case) =====

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
		// 检查是否漏加余额（moneylog中无此订单记录则补加）
		var logCount int
		database.DB.QueryRow(
			"SELECT COUNT(*) FROM qingka_wangke_moneylog WHERE uid = ? AND remark LIKE ?",
			uid, "%"+outTradeNo+"%",
		).Scan(&logCount)
		if logCount == 0 && money > 0 {
			now := time.Now().Format("2006-01-02 15:04:05")
			// 计算充值赠送
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

	// 从上级 paydata 读取易支付配置（参照PHP）
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

	// 调用易支付查询接口
	queryURL := fmt.Sprintf("%s/api.php?act=order&pid=%s&key=%s&out_trade_no=%s",
		epayAPI, epayPID, epayKey, outTradeNo)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(queryURL)
	if err != nil {
		return false, "", errors.New("查询支付状态失败")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var payResult map[string]interface{}
	json.Unmarshal(body, &payResult)

	// 兼容 epay 返回 status 为 string 或 number
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

		// 计算充值赠送
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

// ===== 设置邀请码 (按 PHP szyqm / set_invite_code case) =====

func (s *UserCenterService) SetInviteCode(uid int, yqm string) error {
	if len(yqm) < 4 {
		return errors.New("邀请码最少4位")
	}
	// 检查是否已被使用
	var cnt int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user WHERE yqm = ? AND uid != ?", yqm, uid).Scan(&cnt)
	if cnt > 0 {
		return errors.New("该邀请码已被使用，请换一个")
	}
	_, err := database.DB.Exec("UPDATE qingka_wangke_user SET yqm = ? WHERE uid = ?", yqm, uid)
	return err
}

// ===== 邀请费率 (按 PHP set_invite_rate case) =====

func (s *UserCenterService) SetInviteRate(uid int, yqprice float64, addprice float64) error {
	if yqprice < addprice {
		return errors.New("下级默认费率不能比你低")
	}
	if yqprice > 100 {
		return errors.New("邀请费率不能超过100")
	}
	// 按 PHP: 费率必须为0.05的倍数
	if int(yqprice*100)%5 != 0 {
		return errors.New("邀请费率必须为0.05的倍数")
	}

	// 如果没有邀请码则自动生成 (按 PHP: if ($userrow['yqm'] == ""))
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

// ===== API密钥管理 (按 PHP change_secret_key case) =====

func (s *UserCenterService) ChangeSecretKey(uid int, keyType int, money float64) (string, error) {
	newKey := fmt.Sprintf("%x", time.Now().UnixNano())
	if len(newKey) > 16 {
		newKey = newKey[:16]
	}

	if keyType == 1 {
		// 开通密钥 (按 PHP: money>=100免费，>=5扣5元，否则不够)
		var currentKey string
		database.DB.QueryRow("SELECT COALESCE(`key`,'') FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&currentKey)

		if money >= 100 {
			database.DB.Exec("UPDATE qingka_wangke_user SET `key` = ? WHERE uid = ?", newKey, uid)
			return newKey, nil
		} else if money >= 5 {
			database.DB.Exec("UPDATE qingka_wangke_user SET `key` = ?, money = money - 5 WHERE uid = ?", newKey, uid)
			now := time.Now().Format("2006-01-02 15:04:05")
			database.DB.Exec(
				"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, '扣费', -5, (SELECT money FROM qingka_wangke_user WHERE uid = ?), '开通接口扣费5元', ?)",
				uid, uid, now,
			)
			return newKey, nil
		}
		return "", errors.New("余额不足，需要5元开通费用")
	} else if keyType == 3 {
		// 更换密钥
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

// ===== 操作日志 (按 PHP log_list case) =====

func (s *UserCenterService) LogList(uid int, grade string, page, limit int, logType, keywords string) ([]map[string]interface{}, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	// 按 PHP: 管理员查看所有，普通用户只看自己
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

// 修改密码
func (s *UserCenterService) ChangePassword(uid int, oldPass, newPass string) error {
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

	// 异步发送密码修改通知邮件
	go func() {
		var username, email string
		database.DB.QueryRow("SELECT COALESCE(user,''), COALESCE(email,'') FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&username, &email)
		if email == "" {
			return
		}
		conf, _ := NewAdminService().GetConfig()
		siteName := conf["sitename"]
		if siteName == "" {
			siteName = "System"
		}
		htmlBody := TemplatePasswordChanged(siteName, username, time.Now().Format("2006-01-02 15:04:05"))
		NewEmailService().SendEmail(email, siteName+" - 密码修改通知", htmlBody)
	}()

	return nil
}

// 设置/修改二级密码（管理员专用）
func (s *UserCenterService) ChangePass2(uid int, oldPass2, newPass2 string) error {
	var grade, dbPass2 string
	err := database.DB.QueryRow("SELECT grade, IFNULL(pass2,'') FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&grade, &dbPass2)
	if err != nil {
		return errors.New("用户不存在")
	}
	if grade != "3" {
		return errors.New("仅管理员可设置二级密码")
	}
	// 如果已有二级密码，需要验证旧密码
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

// SendChangeEmailCode 发送邮箱变更验证码到新邮箱
func (s *UserCenterService) SendChangeEmailCode(uid int, newEmail string) error {
	vs := NewVerificationService()
	identifier := fmt.Sprintf("%d:%s", uid, newEmail)
	if err := vs.RateLimitCheck("change_email", identifier); err != nil {
		return err
	}

	code, err := vs.GenerateCode("change_email", identifier, 10*time.Minute)
	if err != nil {
		return err
	}

	conf, _ := NewAdminService().GetConfig()
	siteName := conf["sitename"]
	if siteName == "" {
		siteName = "System"
	}
	htmlBody := TemplateChangeEmailCode(siteName, code, 10)

	es := NewEmailService()
	if err := es.SendEmail(newEmail, siteName+" - 邮箱变更验证码", htmlBody); err != nil {
		return fmt.Errorf("发送邮件失败: %v", err)
	}

	return nil
}

// ChangeEmail 确认变更邮箱
func (s *UserCenterService) ChangeEmail(uid int, newEmail, code string) error {
	vs := NewVerificationService()
	identifier := fmt.Sprintf("%d:%s", uid, newEmail)
	if !vs.VerifyCode("change_email", identifier, code) {
		return errors.New("验证码错误或已过期")
	}

	// 获取旧邮箱
	var oldEmail, username string
	database.DB.QueryRow("SELECT COALESCE(email,''), COALESCE(user,'') FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&oldEmail, &username)

	// 更新邮箱
	_, err := database.DB.Exec("UPDATE qingka_wangke_user SET email = ? WHERE uid = ?", newEmail, uid)
	if err != nil {
		return fmt.Errorf("更新邮箱失败: %v", err)
	}

	// 异步通知旧邮箱
	if oldEmail != "" && oldEmail != newEmail {
		go func() {
			conf, _ := NewAdminService().GetConfig()
			siteName := conf["sitename"]
			if siteName == "" {
				siteName = "System"
			}
			htmlBody := TemplateEmailChanged(siteName, username, newEmail, time.Now().Format("2006-01-02 15:04:05"))
			NewEmailService().SendEmail(oldEmail, siteName+" - 邮箱变更通知", htmlBody)
		}()
	}

	return nil
}

// 获取用户上级的 paydata JSON
func (s *UserCenterService) getParentPayData(uid int) (map[string]string, int, error) {
	// 查询用户的上级uid
	var parentUID int
	err := database.DB.QueryRow("SELECT uuid FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&parentUID)
	if err != nil {
		return nil, 0, errors.New("用户不存在")
	}

	// 查询上级的 paydata
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

// 获取可用支付渠道 (从上级 paydata 读取，参照PHP)
func (s *UserCenterService) GetPayChannels(uid int) ([]model.PayChannel, error) {
	payData, parentUID, err := s.getParentPayData(uid)
	if err != nil {
		return nil, err
	}

	// PHP: 只有上级是管理员(uid=1)才允许在线充值
	// non_direct_recharge_enable: 开启后非直系代理也可充值
	if parentUID != 1 {
		conf, _ := NewAdminService().GetConfig()
		if conf["non_direct_recharge_enable"] != "1" {
			return []model.PayChannel{}, nil
		}
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

// 创建充值订单 (参照PHP apisub.php case 'pay')
func (s *UserCenterService) CreatePayOrder(uid int, money float64, payType string, domain string) (*model.PayCreateResponse, error) {
	payData, parentUID, err := s.getParentPayData(uid)
	if err != nil {
		return nil, err
	}

	// 检查最低充值 (从系统配置读取)
	conf, _ := NewAdminService().GetConfig()

	// PHP: 只有上级是管理员(uid=1)才允许在线充值
	// non_direct_recharge_enable: 开启后非直系代理也可充值
	if parentUID != 1 {
		if conf["non_direct_recharge_enable"] != "1" {
			return nil, errors.New("请您根据上面的信息联系上家充值")
		}
	}
	if zdpay := conf["zdpay"]; zdpay != "" {
		var minPay float64
		fmt.Sscanf(zdpay, "%f", &minPay)
		if minPay > 0 && money < minPay {
			return nil, fmt.Errorf("在线充值最低%.0f元", minPay)
		}
	}

	// PHP: out_trade_no = date("YmdHis") . rand(111,999)
	outTradeNo := fmt.Sprintf("%s%d", time.Now().Format("20060102150405"), 111+time.Now().UnixNano()%889)
	moneyStr := fmt.Sprintf("%.2f", money)
	name := fmt.Sprintf("零食购买-%s", moneyStr)
	now := time.Now().Format("2006-01-02 15:04:05")

	// PHP INSERT: out_trade_no, uid, num, name, money, ip, addtime, domain, status
	result, dbErr := database.DB.Exec(
		"INSERT INTO qingka_wangke_pay (out_trade_no, trade_no, uid, num, name, money, ip, addtime, domain, status) VALUES (?, '', ?, ?, ?, ?, '', ?, ?, 0)",
		outTradeNo, uid, moneyStr, name, moneyStr, now, domain,
	)
	if dbErr != nil {
		return nil, errors.New("生成订单失败")
	}

	oid, _ := result.LastInsertId()

	// 用上级的 paydata 生成易支付跳转URL
	payURL := s.buildEpayURL(payData, outTradeNo, name, moneyStr, payType, domain)

	return &model.PayCreateResponse{
		OID:        int(oid),
		OutTradeNo: outTradeNo,
		Money:      moneyStr,
		PayURL:     payURL,
	}, nil
}

// 生成易支付跳转URL (按PHP epay签名逻辑，使用上级paydata中的epay配置)
func (s *UserCenterService) buildEpayURL(payData map[string]string, outTradeNo, name, money, payType, domain string) string {
	epayAPI := strings.TrimRight(payData["epay_api"], "/")
	pid := payData["epay_pid"]
	key := payData["epay_key"]

	if epayAPI == "" || pid == "" || key == "" {
		return ""
	}

	notifyURL := "https://" + domain + "/epay/notify_url.php"
	returnURL := "https://" + domain + "/epay/return_url.php"

	// 构造参数 (排除空值, sign, sign_type)
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

	// ksort + 拼接签名字符串
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

	// MD5签名: md5(signStr + key)
	sign := fmt.Sprintf("%x", md5.Sum([]byte(signStr+key)))

	// 构造跳转URL
	return fmt.Sprintf("%s/submit.php?%s&sign=%s&sign_type=MD5", epayAPI, signStr, sign)
}

// 充值记录
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

// ===== 余额流水 =====

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

// ===== 工单 =====

func (s *UserCenterService) TicketList(uid int, grade string, page, limit int) ([]model.Ticket, int64, error) {
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

// AdminTicketList 管理员工单列表（支持筛选）
func (s *UserCenterService) AdminTicketList(page, limit int, status, uid int, search string) ([]model.Ticket, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	where := "1=1"
	args := []interface{}{}
	if status > 0 {
		where += " AND t.status = ?"
		args = append(args, status)
	}
	if uid > 0 {
		where += " AND t.uid = ?"
		args = append(args, uid)
	}
	if search != "" {
		where += " AND (t.content LIKE ? OR t.reply LIKE ? OR CAST(t.oid AS CHAR) LIKE ?)"
		s := "%" + search + "%"
		args = append(args, s, s, s)
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_ticket t WHERE "+where, args...).Scan(&total)

	offset := (page - 1) * limit
	args2 := append(args, limit, offset)
	rows, err := database.DB.Query(
		fmt.Sprintf(`SELECT t.id, t.uid, COALESCE(t.oid,0), COALESCE(t.type,''), COALESCE(t.content,''), COALESCE(t.reply,''),
			t.status, COALESCE(DATE_FORMAT(t.addtime,'%%Y-%%m-%%d %%H:%%i:%%s'),''), COALESCE(DATE_FORMAT(t.reply_time,'%%Y-%%m-%%d %%H:%%i:%%s'),''),
			COALESCE(t.supplier_report_id,0), COALESCE(t.supplier_status,-1), COALESCE(t.supplier_answer,''),
			COALESCE(o.user,''), COALESCE(o.ptname,''), COALESCE(o.status,''), COALESCE(o.yid,''),
			COALESCE(f.supplier_report,0), COALESCE(f.supplier_report_hid,0)
		FROM qingka_wangke_ticket t
		LEFT JOIN qingka_wangke_order o ON o.oid = t.oid
		LEFT JOIN qingka_wangke_class c ON c.cid = o.cid
		LEFT JOIN qingka_wangke_fenlei f ON f.id = CAST(c.fenlei AS UNSIGNED)
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
			&t.OrderUser, &t.OrderPT, &t.OrderStatus, &t.OrderYID,
			&t.SupplierReportSwitch, &t.SupplierReportHID)
		tickets = append(tickets, t)
	}
	if tickets == nil {
		tickets = []model.Ticket{}
	}
	return tickets, total, nil
}

// TicketStats 工单统计
func (s *UserCenterService) TicketStats() (map[string]int64, error) {
	var total, pending, replied, closed, upPending int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_ticket").Scan(&total)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_ticket WHERE status = 1").Scan(&pending)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_ticket WHERE status = 2").Scan(&replied)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_ticket WHERE status = 3").Scan(&closed)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_ticket WHERE supplier_report_id > 0 AND supplier_status IN (0,4)").Scan(&upPending)
	return map[string]int64{
		"total": total, "pending": pending, "replied": replied,
		"closed": closed, "upstream_pending": upPending,
	}, nil
}

// TicketCountByOID 查询某订单关联的工单数量
func (s *UserCenterService) TicketCountByOID(oid int) int {
	var cnt int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_ticket WHERE oid = ?", oid).Scan(&cnt)
	return cnt
}

// TicketCountByOIDs 批量查询多个订单关联的工单数量
func (s *UserCenterService) TicketCountByOIDs(oids []int) map[int]int {
	result := make(map[int]int)
	if len(oids) == 0 {
		return result
	}
	placeholders := make([]string, len(oids))
	args := make([]interface{}, len(oids))
	for i, oid := range oids {
		placeholders[i] = "?"
		args[i] = oid
	}
	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT oid, COUNT(*) FROM qingka_wangke_ticket WHERE oid IN (%s) GROUP BY oid", strings.Join(placeholders, ",")),
		args...,
	)
	if err != nil {
		return result
	}
	defer rows.Close()
	for rows.Next() {
		var oid, cnt int
		rows.Scan(&oid, &cnt)
		result[oid] = cnt
	}
	return result
}

// AutoCloseExpiredTickets 自动关闭超过N天未回复的工单
func (s *UserCenterService) AutoCloseExpiredTickets(days int) (int64, error) {
	result, err := database.DB.Exec(
		"UPDATE qingka_wangke_ticket SET status = 3 WHERE status = 2 AND reply_time IS NOT NULL AND reply_time < DATE_SUB(NOW(), INTERVAL ? DAY)",
		days,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *UserCenterService) TicketCreate(uid int, req model.TicketCreateRequest) (int64, error) {
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

func (s *UserCenterService) TicketReply(uid int, grade string, req model.TicketReplyRequest) error {
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

func (s *UserCenterService) TicketClose(uid int, grade string, ticketID int) error {
	where := "id = ? AND uid = ?"
	args := []interface{}{ticketID, uid}
	if grade == "2" || grade == "3" {
		where = "id = ?"
		args = []interface{}{ticketID}
	}
	_, err := database.DB.Exec("UPDATE qingka_wangke_ticket SET status = 3 WHERE "+where, args...)
	return err
}

// UpdateTicketSupplierReport 更新工单的上游反馈信息
func (s *UserCenterService) UpdateTicketSupplierReport(ticketID, reportID, supplierStatus int, answer string) error {
	_, err := database.DB.Exec(
		"UPDATE qingka_wangke_ticket SET supplier_report_id = ?, supplier_status = ?, supplier_answer = ? WHERE id = ?",
		reportID, supplierStatus, answer, ticketID,
	)
	return err
}

// GetTicketByID 获取单个工单
func (s *UserCenterService) GetTicketByID(ticketID int) (*model.Ticket, error) {
	var t model.Ticket
	err := database.DB.QueryRow(
		`SELECT t.id, t.uid, COALESCE(t.oid,0), COALESCE(t.type,''), COALESCE(t.content,''), COALESCE(t.reply,''),
			t.status, COALESCE(DATE_FORMAT(t.addtime,'%Y-%m-%d %H:%i:%s'),''), COALESCE(DATE_FORMAT(t.reply_time,'%Y-%m-%d %H:%i:%s'),''),
			COALESCE(t.supplier_report_id,0), COALESCE(t.supplier_status,-1), COALESCE(t.supplier_answer,''),
			COALESCE(o.user,''), COALESCE(o.ptname,''), COALESCE(o.status,''), COALESCE(o.yid,'')
		FROM qingka_wangke_ticket t
		LEFT JOIN qingka_wangke_order o ON o.oid = t.oid
		WHERE t.id = ?`, ticketID,
	).Scan(&t.ID, &t.UID, &t.OID, &t.Type, &t.Content, &t.Reply,
		&t.Status, &t.AddTime, &t.ReplyTime,
		&t.SupplierReportID, &t.SupplierStatus, &t.SupplierAnswer,
		&t.OrderUser, &t.OrderPT, &t.OrderStatus, &t.OrderYID)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
