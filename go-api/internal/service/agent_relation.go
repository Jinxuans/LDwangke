package service

import (
	"errors"
	"fmt"
	"time"

	"go-api/internal/database"
)

func AgentMigrateSuperior(currentUID int, targetUID int, yqm string) error {
	if targetUID <= 0 || yqm == "" {
		return errors.New("所有项目不能为空")
	}

	if !adminConfigEnabled("sjqykg") {
		return errors.New("管理员未打开迁移功能")
	}

	var targetYqm string
	err := database.DB.QueryRow("SELECT COALESCE(yqm,'') FROM qingka_wangke_user WHERE uid = ?", targetUID).Scan(&targetYqm)
	if err != nil {
		return errors.New("UID不存在，请重新输入")
	}
	if yqm != targetYqm {
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

func AgentSetInviteCode(operatorUID int, targetUID int, yqm string) error {
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
