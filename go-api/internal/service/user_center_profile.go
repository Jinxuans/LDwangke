package service

import (
	"errors"
	"fmt"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

func (s *UserCenterService) Profile(uid int, grade string) (*model.UserProfile, error) {
	var p model.UserProfile
	var uuid int
	err := database.DB.QueryRow(
		"SELECT uid, COALESCE(uuid,0), user, COALESCE(name,''), COALESCE(money,0), COALESCE(addprice,1), COALESCE(`key`,''), COALESCE(yqm,''), COALESCE(yqprice,'0'), COALESCE(email,''), COALESCE(tuisongtoken,''), COALESCE(zcz,0) FROM qingka_wangke_user WHERE uid = ?",
		uid,
	).Scan(&p.UID, &uuid, &p.User, &p.Name, &p.Money, &p.AddPrice, &p.Key, &p.YQM, &p.YQPrice, &p.Email, &p.PushToken, &p.ZCZ)
	if err == nil {
		database.DB.QueryRow("SELECT COALESCE(cdmoney,0) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&p.CDMoney)
		if CrossRechargeAllowed(uid) {
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

func (s *UserCenterService) GetFavorites(uid int) ([]int, error) {
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

func (s *UserCenterService) AddFavorite(uid, cid int) error {
	var exists int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_user_favorite WHERE uid = ? AND cid = ?", uid, cid).Scan(&exists)
	if exists > 0 {
		return errors.New("已收藏过该商品")
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := database.DB.Exec("INSERT INTO qingka_wangke_user_favorite (uid, cid, addtime) VALUES (?, ?, ?)", uid, cid, now)
	return err
}

func (s *UserCenterService) RemoveFavorite(uid, cid int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_wangke_user_favorite WHERE uid = ? AND cid = ?", uid, cid)
	return err
}

func (s *UserCenterService) SetInviteCode(uid int, yqm string) error {
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

func (s *UserCenterService) SetInviteRate(uid int, yqprice float64, addprice float64) error {
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

func (s *UserCenterService) ChangeSecretKey(uid int, keyType int, money float64) (string, error) {
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

func (s *UserCenterService) LogList(uid int, grade string, page, limit int, logType, keywords string) ([]map[string]interface{}, int64, error) {
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
