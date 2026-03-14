package service

import (
	"errors"
	"fmt"
	"math"
	"time"

	"go-api/internal/database"
)

func AgentList(uid int, grade string, page, limit int, searchType, keywords string) ([]AgentListItem, int64, error) {
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

func AgentCreate(operatorUID int, operatorGrade string, operatorMoney, operatorAddPrice float64, req AgentCreateRequest) (string, error) {
	if req.Nickname == "" || req.User == "" || req.Pass == "" {
		return "", errors.New("所有项目不能为空")
	}
	if len(req.User) < 5 || len(req.User) > 11 {
		return "", errors.New("账号必须为QQ号")
	}

	conf := getAdminConfig()
	if getAdminConfigValue("user_htkh") == "0" {
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
		"INSERT INTO qingka_wangke_user (uuid, user, pass, name, addprice, addtime, active, grade, money) VALUES (?, ?, ?, ?, ?, ?, '1', '0', 0)",
		operatorUID, req.User, req.Pass, req.Nickname, gradeRate, now,
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

func AgentChangeStatus(operatorUID int, targetUID int, currentActive int) error {
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
		siteName := getConfiguredSiteName()
		var htmlBody string
		if newActive == 1 {
			htmlBody = templateAccountEnabled(siteName, username)
		} else {
			htmlBody = templateAccountDisabled(siteName, username)
		}
		SendEmail(email, siteName+" - 账号状态变更通知", htmlBody)
	}()

	return nil
}

func AgentResetPassword(operatorUID int, targetUID int) (string, error) {
	var targetUUID int
	err := database.DB.QueryRow("SELECT COALESCE(uuid,0) FROM qingka_wangke_user WHERE uid = ?", targetUID).Scan(&targetUUID)
	if err != nil {
		return "", errors.New("用户不存在")
	}

	if operatorUID != 1 && targetUUID != operatorUID {
		return "", errors.New("该用户不是你的下级,无法重置密码")
	}

	newPass := "123456"
	database.DB.Exec("UPDATE qingka_wangke_user SET pass = ? WHERE uid = ?", newPass, targetUID)
	wlog(targetUID, "重置密码", fmt.Sprintf("成功重置UID为%d的密码为%s", targetUID, newPass), 0)

	return fmt.Sprintf("成功重置密码为%s", newPass), nil
}

func AgentOpenSecretKey(operatorUID int, operatorMoney float64, targetUID int) error {
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
