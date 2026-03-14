package service

import (
	"errors"
	"fmt"
	"math"

	"go-api/internal/database"
)

func AgentRecharge(operatorUID int, operatorGrade string, operatorMoney, operatorAddPrice float64, targetUID int, money float64) error {
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

func AgentDeduct(operatorUID int, operatorGrade string, targetUID int, money float64) error {
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

func AgentChangeGrade(operatorUID int, operatorGrade string, operatorMoney, operatorAddPrice float64, req AgentChangeGradeRequest) (string, error) {
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
	database.DB.Exec("UPDATE qingka_wangke_user SET money = ?, addprice = ? WHERE uid = ?", newMoney, newRate, req.UID)

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

func CrossRechargeAllowed(uid int) bool {
	if uid == 1 {
		return true
	}
	uidList := getAdminConfigValue("cross_recharge_uids")
	if uidList == "" {
		return false
	}
	uidStr := fmt.Sprintf("%d", uid)
	for _, s := range splitCSV(uidList) {
		if s == uidStr {
			return true
		}
	}
	return false
}

func AgentCrossRecharge(operatorUID int, targetUID int, money float64) error {
	if money <= 0 {
		return errors.New("充值金额必须大于0")
	}
	if money < 1 {
		return errors.New("最低充值1元")
	}
	if operatorUID == targetUID {
		return errors.New("不能给自己跨户充值")
	}

	if !CrossRechargeAllowed(operatorUID) {
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

	return nil
}
