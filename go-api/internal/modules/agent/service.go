package agent

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"go-api/internal/database"
	commonmodule "go-api/internal/modules/common"

	"golang.org/x/crypto/bcrypt"
)

type agentService struct{}

var agents = &agentService{}

type AgentListItem struct {
	UUID     int     `json:"uuid"`
	Active   int     `json:"active"`
	UID      int     `json:"uid"`
	User     string  `json:"user"`
	Name     string  `json:"name"`
	Money    float64 `json:"money"`
	ZCZ      float64 `json:"zcz"`
	AddPrice float64 `json:"addprice"`
	YQM      string  `json:"yqm"`
	EndTime  string  `json:"endtime"`
	AddTime  string  `json:"addtime"`
	DD       int     `json:"dd"`
	Key      int     `json:"key"`
}

type AgentCreateRequest struct {
	Nickname string `json:"nickname"`
	User     string `json:"user"`
	Pass     string `json:"pass"`
	GradeID  int    `json:"gradeId"`
	Confirm  int    `json:"type"`
}

type AgentChangeGradeRequest struct {
	UID     int `json:"uid"`
	GradeID int `json:"gradeId"`
	Confirm int `json:"type"`
}

func getAdminConfig() map[string]string {
	conf, err := commonmodule.GetAdminConfigMap()
	if err != nil || conf == nil {
		return map[string]string{}
	}
	return conf
}

func splitCSV(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		p := strings.TrimSpace(part)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func wlog(uid int, logType, text string, money float64) {
	now := time.Now().Format("2006-01-02 15:04:05")
	var smoney float64
	database.DB.QueryRow("SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&smoney)
	database.DB.Exec(
		"INSERT INTO qingka_wangke_log (uid, type, text, money, smoney, addtime) VALUES (?, ?, ?, ?, ?, ?)",
		uid, logType, text, money, smoney, now,
	)
}

func (s *agentService) AgentList(uid int, grade string, page, limit int, searchType, keywords string) ([]AgentListItem, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	where := "WHERE uuid = ?"
	args := []interface{}{uid}
	if uid == 1 || grade == "2" || grade == "3" {
		where = "WHERE 1=1"
		args = []interface{}{}
	}

	if keywords != "" {
		switch searchType {
		case "1":
			where += " AND uid = ?"
			args = append(args, keywords)
		case "2":
			where += " AND user LIKE ?"
			args = append(args, "%"+keywords+"%")
		case "3":
			where += " AND yqm = ?"
			args = append(args, keywords)
		case "4":
			where += " AND name LIKE ?"
			args = append(args, "%"+keywords+"%")
		case "5":
			where += " AND addprice = ?"
			args = append(args, keywords)
		case "6":
			where += " AND money = ?"
			args = append(args, keywords)
		case "7":
			where += " AND endtime > ?"
			args = append(args, keywords)
		}
	}

	var total int64
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user "+where, countArgs...).Scan(&total)

	offset := (page - 1) * limit
	queryArgs := append(args, limit, offset)
	rows, err := database.DB.Query(
		fmt.Sprintf("SELECT uid, COALESCE(uuid,0), user, COALESCE(name,''), COALESCE(money,0), COALESCE(zcz,0), COALESCE(addprice,1), COALESCE(yqm,''), COALESCE(endtime,''), COALESCE(addtime,''), COALESCE(active,1), COALESCE(`key`,'') FROM qingka_wangke_user %s ORDER BY uid DESC LIMIT ? OFFSET ?", where),
		queryArgs...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []AgentListItem
	for rows.Next() {
		var item AgentListItem
		var keyStr string
		rows.Scan(&item.UID, &item.UUID, &item.User, &item.Name, &item.Money, &item.ZCZ, &item.AddPrice, &item.YQM, &item.EndTime, &item.AddTime, &item.Active, &keyStr)

		database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_order WHERE uid = ?", item.UID).Scan(&item.DD)

		if keyStr != "" && keyStr != "0" {
			item.Key = 1
		}
		item.Money = math.Round(item.Money*100) / 100
		list = append(list, item)
	}

	if list == nil {
		list = []AgentListItem{}
	}
	return list, total, nil
}

func (s *agentService) AgentCreate(operatorUID int, operatorGrade string, operatorMoney, operatorAddPrice float64, req AgentCreateRequest) (string, error) {
	if req.Nickname == "" || req.User == "" || req.Pass == "" {
		return "", errors.New("所有项目不能为空")
	}
	if len(req.User) < 5 || len(req.User) > 11 {
		return "", errors.New("账号必须为QQ号")
	}

	conf := getAdminConfig()
	if conf["user_htkh"] == "0" {
		return "", errors.New("暂停开户，具体开放时间等通知")
	}

	var cnt int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user WHERE user = ?", req.User).Scan(&cnt)
	if cnt > 0 {
		return "", errors.New("该账号已存在")
	}

	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user WHERE name = ?", req.Nickname).Scan(&cnt)
	if cnt > 0 {
		return "", errors.New("该昵称已存在")
	}

	var gradeName string
	var gradeRate float64
	var gradeMoney float64
	err := database.DB.QueryRow("SELECT COALESCE(name,''), COALESCE(rate,0), COALESCE(money,0) FROM qingka_wangke_dengji WHERE id = ? AND status = '1'", req.GradeID).Scan(&gradeName, &gradeRate, &gradeMoney)
	if err != nil {
		return "", errors.New("等级信息不存在")
	}

	if gradeRate < operatorAddPrice {
		return "", errors.New("费率不能比自己低哦")
	}

	openFee := 1.0
	if ktm := conf["user_ktmoney"]; ktm != "" {
		var f float64
		fmt.Sscanf(ktm, "%f", &f)
		if f > 0 {
			openFee = f
		}
	}
	kochu := math.Round(gradeMoney*(operatorAddPrice/gradeRate)*100) / 100
	totalCost := kochu + openFee

	dlPkkg := conf["dl_pkkg"]
	djfl := conf["djfl"]
	if operatorUID != 1 && dlPkkg != "" && dlPkkg != "0" {
		isTopLevel := djfl != "" && fmt.Sprintf("%.2f", gradeRate) == djfl
		isSameLevel := fmt.Sprintf("%.2f", gradeRate) == fmt.Sprintf("%.2f", operatorAddPrice)

		switch dlPkkg {
		case "1":
			if isTopLevel {
				return "", errors.New("禁止顶级用户平开")
			}
		case "2":
			if isTopLevel {
				totalCost = kochu + kochu + openFee
			}
		case "3":
			if isSameLevel {
				totalCost = kochu + kochu + openFee
			}
		}
	}

	if req.Confirm != 1 {
		msg := fmt.Sprintf("开通扣%.0f元开户费，并自动给下级充值%.0f元，将扣除%.2f余额", openFee, gradeMoney, kochu)
		return msg, nil
	}

	if operatorUID != 1 && operatorMoney < totalCost {
		return "", fmt.Errorf("余额不足开户，开户需扣除开户费%.0f元，及余额%.2f元", openFee, kochu)
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = database.DB.Exec(
		"INSERT INTO qingka_wangke_user (uuid, user, pass, name, grade_id, invite_grade_id, addprice, addtime, active, grade, money) VALUES (?, ?, ?, ?, ?, ?, ?, ?, '1', '0', 0)",
		operatorUID, req.User, req.Pass, req.Nickname, req.GradeID, req.GradeID, gradeRate, now,
	)
	if err != nil {
		return "", fmt.Errorf("添加失败: %v", err)
	}

	if operatorUID != 1 {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", openFee, operatorUID)
		wlog(operatorUID, "添加商户", fmt.Sprintf("添加商户%s成功!扣费%.0f元!", req.User, openFee), -openFee)
	}

	if gradeMoney > 0 {
		var newUID int
		database.DB.QueryRow("SELECT uid FROM qingka_wangke_user WHERE user = ? LIMIT 1", req.User).Scan(&newUID)
		database.DB.Exec("UPDATE qingka_wangke_user SET money = ?, zcz = zcz + ? WHERE uid = ?", gradeMoney, gradeMoney, newUID)
		if operatorUID != 1 {
			database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", kochu, operatorUID)
			wlog(operatorUID, "代理充值", fmt.Sprintf("成功给账号为[%s]的靓仔充值%.0f元,扣除%.2f元", req.User, gradeMoney, kochu), -kochu)
		}
		wlog(newUID, "上级充值", fmt.Sprintf("你上面的靓仔成功给你充值%.0f元", gradeMoney), gradeMoney)
	}

	return "添加成功", nil
}

func (s *agentService) AgentRecharge(operatorUID int, operatorGrade string, operatorMoney, operatorAddPrice float64, targetUID int, money float64) error {
	if money <= 0 {
		return errors.New("充值金额不合法")
	}
	if operatorUID != 1 && money < 5 {
		return errors.New("最低充值5元")
	}

	var targetUUID int
	var targetUser string
	var targetMoney, targetAddPrice float64
	err := database.DB.QueryRow("SELECT COALESCE(uuid,0), user, COALESCE(money,0), COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid = ?", targetUID).Scan(&targetUUID, &targetUser, &targetMoney, &targetAddPrice)
	if err != nil {
		return errors.New("目标用户不存在")
	}

	if operatorUID != 1 && targetUUID != operatorUID {
		return errors.New("该用户不是你的下级,无法充值")
	}
	if operatorUID == targetUID {
		return errors.New("自己不能给自己充值哦")
	}

	kochu := math.Round(money*(operatorAddPrice/targetAddPrice)*100000) / 100000
	if operatorUID != 1 && operatorMoney < kochu {
		return errors.New("您当前余额不足,无法充值")
	}

	if operatorUID != 1 {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", kochu, operatorUID)
	}
	database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ?, zcz = zcz + ? WHERE uid = ?", money, money, targetUID)

	var operatorName string
	database.DB.QueryRow("SELECT COALESCE(name,'') FROM qingka_wangke_user WHERE uid = ?", operatorUID).Scan(&operatorName)

	wlog(operatorUID, "代理充值", fmt.Sprintf("成功给账号为[%s]的靓仔充值%.2f元,实际扣费%.5f元", targetUser, money, kochu), -kochu)
	wlog(targetUID, "上级充值", fmt.Sprintf("%s已充值你%.2f元", operatorName, money), money)
	return nil
}

func (s *agentService) AgentDeduct(operatorUID int, operatorGrade string, targetUID int, money float64) error {
	if money <= 0 {
		return errors.New("扣除金额不合法")
	}

	var targetUUID int
	var targetUser string
	var targetMoney float64
	err := database.DB.QueryRow("SELECT COALESCE(uuid,0), user, COALESCE(money,0) FROM qingka_wangke_user WHERE uid = ?", targetUID).Scan(&targetUUID, &targetUser, &targetMoney)
	if err != nil {
		return errors.New("目标用户不存在")
	}

	if operatorUID != 1 && targetUUID != operatorUID {
		return errors.New("该用户不是你的下级,无法扣除")
	}
	if operatorUID == targetUID {
		return errors.New("不能扣除自己的余额")
	}
	if targetMoney < money {
		return errors.New("该用户余额不足")
	}

	database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", money, targetUID)
	if operatorUID != 1 {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ? WHERE uid = ?", money, operatorUID)
	}

	var operatorName string
	database.DB.QueryRow("SELECT COALESCE(name,'') FROM qingka_wangke_user WHERE uid = ?", operatorUID).Scan(&operatorName)

	wlog(operatorUID, "代理扣款", fmt.Sprintf("成功扣除账号为[%s]的靓仔%.2f元", targetUser, money), money)
	wlog(targetUID, "上级扣款", fmt.Sprintf("%s已扣除你%.2f元", operatorName, money), -money)
	return nil
}

func (s *agentService) AgentChangeGrade(operatorUID int, operatorGrade string, operatorMoney, operatorAddPrice float64, req AgentChangeGradeRequest) (string, error) {
	var targetUUID int
	var targetName, targetUser string
	var targetMoney, targetAddPrice float64
	err := database.DB.QueryRow("SELECT COALESCE(uuid,0), COALESCE(name,''), user, COALESCE(money,0), COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid = ?", req.UID).Scan(&targetUUID, &targetName, &targetUser, &targetMoney, &targetAddPrice)
	if err != nil {
		return "", errors.New("用户不存在")
	}

	if operatorUID != 1 && targetUUID != operatorUID {
		return "", errors.New("该用户不是你的下级,无法修改等级")
	}

	var newRate, cost float64
	err = database.DB.QueryRow("SELECT COALESCE(rate,0), COALESCE(money,0) FROM qingka_wangke_dengji WHERE id = ? AND status = '1'", req.GradeID).Scan(&newRate, &cost)
	if err != nil {
		return "", errors.New("等级信息不存在")
	}

	kochu := math.Round(cost*(operatorAddPrice/newRate)*100) / 100
	if req.Confirm != 1 {
		msg := fmt.Sprintf("改价手续费3元，并自动给下级[UID:%d]充值%.0f元，将扣除%.2f余额", req.UID, cost, kochu)
		return msg, nil
	}

	totalCost := kochu + 3
	if operatorUID != 1 && operatorMoney < totalCost {
		return "", fmt.Errorf("余额不足,改价需扣3元手续费,及余额%.2f元", kochu)
	}

	newMoney := math.Round(targetMoney/targetAddPrice*newRate*100) / 100
	moneyChange := newMoney - targetMoney

	if operatorUID != 1 {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money - 3 WHERE uid = ?", operatorUID)
	}
	database.DB.Exec("UPDATE qingka_wangke_user SET money = ?, grade_id = ?, addprice = ? WHERE uid = ?", newMoney, req.GradeID, newRate, req.UID)

	var operatorName string
	database.DB.QueryRow("SELECT COALESCE(name,'') FROM qingka_wangke_user WHERE uid = ?", operatorUID).Scan(&operatorName)

	wlog(operatorUID, "修改费率", fmt.Sprintf("修改代理%s,费率：%.2f,扣除手续费3元", targetName, newRate), -3)
	wlog(req.UID, "修改费率", fmt.Sprintf("%s修改你的费率为：%.2f,系统根据比例自动调整价格", operatorName, newRate), moneyChange)

	if cost > 0 {
		if operatorUID != 1 {
			database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", kochu, operatorUID)
		}
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ?, zcz = zcz + ? WHERE uid = ?", cost, cost, req.UID)
		wlog(operatorUID, "代理充值", fmt.Sprintf("成功给账号为[%s]的靓仔充值%.0f元,扣除%.2f元", targetUser, cost, kochu), -kochu)
		wlog(req.UID, "上级充值", fmt.Sprintf("%s成功给你充值%.0f元", operatorName, cost), cost)
	}

	return "改价成功", nil
}

func (s *agentService) AgentChangeStatus(operatorUID int, targetUID int, currentActive int) error {
	if operatorUID != 1 {
		return errors.New("无权限")
	}
	if operatorUID == targetUID {
		return errors.New("不能修改自己的状态")
	}

	newActive := 0
	logText := "封禁商户"
	if currentActive != 1 {
		newActive = 1
		logText = "解封商户"
	}

	database.DB.Exec("UPDATE qingka_wangke_user SET active = ? WHERE uid = ?", newActive, targetUID)
	wlog(operatorUID, logText, fmt.Sprintf("%s[UID %d]成功", logText, targetUID), 0)

	go func() {
		var username, email string
		database.DB.QueryRow("SELECT COALESCE(user,''), COALESCE(email,'') FROM qingka_wangke_user WHERE uid = ?", targetUID).Scan(&username, &email)
		if email == "" {
			return
		}

		siteName := getAdminConfig()["sitename"]
		if siteName == "" {
			siteName = "System"
		}

		var htmlBody string
		if newActive == 1 {
			htmlBody = templateAccountEnabled(siteName, username)
		} else {
			htmlBody = templateAccountDisabled(siteName, username)
		}
		commonmodule.SendEmail(email, siteName+" - 账号状态变更通知", htmlBody)
	}()

	return nil
}

func (s *agentService) AgentResetPassword(operatorUID int, targetUID int) (string, error) {
	var targetUUID int
	err := database.DB.QueryRow("SELECT COALESCE(uuid,0) FROM qingka_wangke_user WHERE uid = ?", targetUID).Scan(&targetUUID)
	if err != nil {
		return "", errors.New("用户不存在")
	}

	if operatorUID != 1 && targetUUID != operatorUID {
		return "", errors.New("该用户不是你的下级,无法重置密码")
	}

	newPass := "123456"
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("密码加密失败")
	}
	database.DB.Exec("UPDATE qingka_wangke_user SET pass = ? WHERE uid = ?", string(hashedPass), targetUID)
	wlog(targetUID, "重置密码", fmt.Sprintf("成功重置UID为%d的密码", targetUID), 0)
	return fmt.Sprintf("成功重置密码为%s", newPass), nil
}

func (s *agentService) AgentOpenSecretKey(operatorUID int, operatorMoney float64, targetUID int) error {
	var targetUUID int
	var targetKey string
	err := database.DB.QueryRow("SELECT COALESCE(uuid,0), COALESCE(`key`,'') FROM qingka_wangke_user WHERE uid = ?", targetUID).Scan(&targetUUID, &targetKey)
	if err != nil {
		return errors.New("用户不存在")
	}

	if operatorUID != 1 && targetUUID != operatorUID {
		return errors.New("该用户不是你的下级,无法开通")
	}
	if targetKey != "" && targetKey != "0" {
		return errors.New("该用户已开通密钥")
	}
	if operatorUID != 1 && operatorMoney < 5 {
		return errors.New("余额不足以开通，需要5元")
	}

	newKey := fmt.Sprintf("%x", time.Now().UnixNano())
	if len(newKey) > 16 {
		newKey = newKey[:16]
	}

	database.DB.Exec("UPDATE qingka_wangke_user SET `key` = ? WHERE uid = ?", newKey, targetUID)
	if operatorUID != 1 {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money - 5 WHERE uid = ?", operatorUID)
		wlog(operatorUID, "开通接口", fmt.Sprintf("给下级代理UID%d开通接口成功!扣费5积分", targetUID), -5)
	}
	wlog(targetUID, "开通接口", "你上级给你开通API接口成功!", 0)
	return nil
}

func (s *agentService) CrossRechargeAllowed(uid int) bool {
	if uid == 1 {
		return true
	}

	uidList := getAdminConfig()["cross_recharge_uids"]
	if uidList == "" {
		return false
	}

	uidStr := fmt.Sprintf("%d", uid)
	for _, item := range splitCSV(uidList) {
		if item == uidStr {
			return true
		}
	}
	return false
}

func (s *agentService) AgentCrossRecharge(operatorUID int, targetUID int, money float64) error {
	if money <= 0 {
		return errors.New("充值金额必须大于0")
	}
	if money < 1 {
		return errors.New("最低充值1元")
	}
	if operatorUID == targetUID {
		return errors.New("不能给自己跨户充值")
	}
	if !s.CrossRechargeAllowed(operatorUID) {
		return errors.New("您没有跨户充值权限")
	}

	var operatorMoney, operatorAddPrice float64
	var operatorName string
	err := database.DB.QueryRow("SELECT COALESCE(money,0), COALESCE(addprice,1), COALESCE(name,'') FROM qingka_wangke_user WHERE uid = ?", operatorUID).Scan(&operatorMoney, &operatorAddPrice, &operatorName)
	if err != nil {
		return errors.New("操作者信息查询失败")
	}

	var targetUser, targetName string
	var targetAddPrice float64
	err = database.DB.QueryRow("SELECT user, COALESCE(name,''), COALESCE(addprice,1) FROM qingka_wangke_user WHERE uid = ?", targetUID).Scan(&targetUser, &targetName, &targetAddPrice)
	if err != nil {
		return errors.New("目标用户不存在")
	}

	kochu := math.Round(money*(operatorAddPrice/targetAddPrice)*100000) / 100000
	if operatorUID != 1 && operatorMoney < kochu {
		return fmt.Errorf("余额不足，需扣 %.2f 元，当前余额 %.2f 元", kochu, operatorMoney)
	}

	if operatorUID != 1 {
		database.DB.Exec("UPDATE qingka_wangke_user SET money = money - ? WHERE uid = ?", kochu, operatorUID)
	}
	database.DB.Exec("UPDATE qingka_wangke_user SET money = money + ?, zcz = zcz + ? WHERE uid = ?", money, money, targetUID)

	wlog(operatorUID, "跨户充值", fmt.Sprintf("跨户充值给[%s](UID:%d) %.2f元,实际扣费%.5f元", targetName, targetUID, money, kochu), -kochu)
	wlog(targetUID, "跨户充值", fmt.Sprintf("%s(UID:%d)跨户充值 %.2f元", operatorName, operatorUID, money), money)
	_ = targetUser
	return nil
}

func (s *agentService) AgentMigrateSuperior(currentUID int, targetUID int, yqm string) error {
	if targetUID <= 0 || yqm == "" {
		return errors.New("所有项目不能为空")
	}
	if getAdminConfig()["sjqykg"] != "1" {
		return errors.New("管理员未打开迁移功能")
	}

	var targetYQM string
	err := database.DB.QueryRow("SELECT COALESCE(yqm,'') FROM qingka_wangke_user WHERE uid = ?", targetUID).Scan(&targetYQM)
	if err != nil {
		return errors.New("UID不存在，请重新输入")
	}
	if yqm != targetYQM {
		return errors.New("非该用户邀请码，请重新输入")
	}

	var currentUUID int
	database.DB.QueryRow("SELECT COALESCE(uuid,0) FROM qingka_wangke_user WHERE uid = ?", currentUID).Scan(&currentUUID)
	if currentUUID == targetUID {
		return errors.New("该用户已经是你的上级了")
	}
	if currentUID == targetUID {
		return errors.New("禁止填写自己的UID")
	}

	var oldSuperiorEndtime string
	database.DB.QueryRow("SELECT COALESCE(endtime,'') FROM qingka_wangke_user WHERE uid = ?", currentUUID).Scan(&oldSuperiorEndtime)
	if oldSuperiorEndtime != "" {
		sevenDaysAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		if oldSuperiorEndtime >= sevenDaysAgo {
			return errors.New("上级在七天内有登陆记录，禁止转移")
		}
	}

	_, err = database.DB.Exec("UPDATE qingka_wangke_user SET uuid = ? WHERE uid = ?", targetUID, currentUID)
	if err != nil {
		return errors.New("迁移失败,未知错误")
	}
	return nil
}

func (s *agentService) AgentAdminChangeSuperior(operatorUID int, operatorGrade string, targetUID int, superiorUID int) error {
	if operatorUID != 1 && operatorGrade != "2" && operatorGrade != "3" {
		return errors.New("无权限")
	}
	if targetUID <= 0 || superiorUID <= 0 {
		return errors.New("用户UID和上级UID不能为空")
	}
	if targetUID == superiorUID {
		return errors.New("上级UID不能填写自己")
	}

	var targetName string
	var currentSuperiorUID int
	var targetGrade string
	err := database.DB.QueryRow(
		"SELECT COALESCE(name,''), COALESCE(uuid,0), COALESCE(grade,'0') FROM qingka_wangke_user WHERE uid = ?",
		targetUID,
	).Scan(&targetName, &currentSuperiorUID, &targetGrade)
	if err != nil {
		return errors.New("目标用户不存在")
	}
	if targetGrade == "2" || targetGrade == "3" {
		return errors.New("管理员账号不允许调整上级")
	}
	if currentSuperiorUID == superiorUID {
		return errors.New("该用户已经在该上级名下")
	}

	var superiorName string
	err = database.DB.QueryRow(
		"SELECT COALESCE(name,'') FROM qingka_wangke_user WHERE uid = ?",
		superiorUID,
	).Scan(&superiorName)
	if err != nil {
		return errors.New("新上级不存在")
	}

	// 防止把用户迁移到自己的下级名下，避免代理树出现环。
	checkUID := superiorUID
	for depth := 0; depth < 100 && checkUID > 0; depth++ {
		if checkUID == targetUID {
			return errors.New("禁止迁移到自己的下级名下")
		}

		var nextUID int
		err = database.DB.QueryRow("SELECT COALESCE(uuid,0) FROM qingka_wangke_user WHERE uid = ?", checkUID).Scan(&nextUID)
		if err != nil || nextUID <= 0 || nextUID == checkUID {
			break
		}
		checkUID = nextUID
	}

	_, err = database.DB.Exec("UPDATE qingka_wangke_user SET uuid = ? WHERE uid = ?", superiorUID, targetUID)
	if err != nil {
		return errors.New("调整上级失败")
	}

	wlog(operatorUID, "调整上级", fmt.Sprintf("将用户[%s](UID:%d)的上级调整为[%s](UID:%d)", targetName, targetUID, superiorName, superiorUID), 0)
	wlog(targetUID, "调整上级", fmt.Sprintf("管理员已将你的上级调整为[%s](UID:%d)", superiorName, superiorUID), 0)
	if currentSuperiorUID > 0 && currentSuperiorUID != superiorUID {
		wlog(currentSuperiorUID, "下级迁移", fmt.Sprintf("用户[%s](UID:%d)已被管理员迁移至[%s](UID:%d)名下", targetName, targetUID, superiorName, superiorUID), 0)
	}
	return nil
}

func (s *agentService) AgentSetInviteCode(operatorUID int, targetUID int, yqm string) error {
	if len(yqm) < 4 {
		return errors.New("邀请码最少4位")
	}

	var targetUUID int
	err := database.DB.QueryRow("SELECT COALESCE(uuid,0) FROM qingka_wangke_user WHERE uid = ?", targetUID).Scan(&targetUUID)
	if err != nil {
		return errors.New("用户不存在")
	}
	if operatorUID != 1 && targetUUID != operatorUID {
		return errors.New("无权限")
	}

	var cnt int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user WHERE yqm = ? AND uid != ?", yqm, targetUID).Scan(&cnt)
	if cnt > 0 {
		return errors.New("该邀请码已被使用，请换一个")
	}

	database.DB.Exec("UPDATE qingka_wangke_user SET yqm = ? WHERE uid = ?", yqm, targetUID)
	wlog(operatorUID, "设置邀请码", fmt.Sprintf("给下级设置邀请码%s成功", yqm), 0)
	return nil
}

func emailLayout(siteName, title, body string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="margin:0;padding:0;background:#f4f5f7;font-family:'Segoe UI',Arial,sans-serif;">
<table width="100%%" cellpadding="0" cellspacing="0" style="background:#f4f5f7;padding:40px 0;">
<tr><td align="center">
<table width="520" cellpadding="0" cellspacing="0" style="background:#fff;border-radius:8px;box-shadow:0 2px 8px rgba(0,0,0,0.06);overflow:hidden;">
  <tr><td style="background:linear-gradient(135deg,#1890ff,#096dd9);padding:28px 32px;">
    <h1 style="margin:0;color:#fff;font-size:20px;">%s</h1>
  </td></tr>
  <tr><td style="padding:32px;">
    <h2 style="margin:0 0 16px;color:#333;font-size:18px;">%s</h2>
    %s
  </td></tr>
  <tr><td style="padding:20px 32px;background:#fafafa;border-top:1px solid #f0f0f0;">
    <p style="margin:0;color:#999;font-size:12px;">此邮件由系统自动发送，请勿回复。</p>
  </td></tr>
</table>
</td></tr>
</table>
</body>
</html>`, siteName, title, body)
}

func templateAccountDisabled(siteName, username string) string {
	body := fmt.Sprintf(`
    <p style="color:#555;line-height:1.8;">您的账号 <strong>%s</strong> 已被管理员禁用。</p>
    <p style="color:#999;font-size:13px;">如有疑问，请联系管理员。</p>`, username)
	return emailLayout(siteName, "账号禁用通知", body)
}

func templateAccountEnabled(siteName, username string) string {
	body := fmt.Sprintf(`
    <p style="color:#555;line-height:1.8;">您的账号 <strong>%s</strong> 已被管理员重新启用，您可以正常登录使用。</p>`, username)
	return emailLayout(siteName, "账号启用通知", body)
}
