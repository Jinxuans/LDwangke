package auxiliary

import (
	"database/sql"
	"fmt"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

func logAuxMoney(uid int, logType, text string, money float64) {
	now := time.Now().Format("2006-01-02 15:04:05")
	var smoney float64
	database.DB.QueryRow("SELECT COALESCE(money,0) FROM qingka_wangke_user WHERE uid = ?", uid).Scan(&smoney)
	database.DB.Exec(
		"INSERT INTO qingka_wangke_log (uid, type, text, money, smoney, addtime) VALUES (?, ?, ?, ?, ?, ?)",
		uid, logType, text, money, smoney, now,
	)
}

func cardKeyUse(uid int, content string) (int, error) {
	var ck model.CardKey
	err := database.DB.QueryRow(
		"SELECT id, money, COALESCE(status,0) FROM qingka_wangke_km WHERE content = ?", content,
	).Scan(&ck.ID, &ck.Money, &ck.Status)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("卡密不存在")
	}
	if err != nil {
		return 0, err
	}
	if ck.Status == 1 {
		return 0, fmt.Errorf("该卡密已被使用")
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = database.DB.Exec(
		"UPDATE qingka_wangke_km SET status=1, uid=?, usedtime=? WHERE id=? AND status=0",
		uid, now, ck.ID,
	)
	if err != nil {
		return 0, err
	}

	_, err = database.DB.Exec("UPDATE qingka_wangke_user SET money=money+?, zcz=zcz+? WHERE uid=?", ck.Money, ck.Money, uid)
	if err != nil {
		return 0, err
	}

	logAuxMoney(uid, "卡密充值", fmt.Sprintf("使用卡密充值%d元成功", ck.Money), float64(ck.Money))
	return ck.Money, nil
}

func listPublicActivities() ([]model.Activity, error) {
	rows, err := database.DB.Query(
		"SELECT hid, name, yaoqiu, type, num, money, addtime, endtime, status_ok, status FROM qingka_wangke_huodong WHERE status_ok='1' ORDER BY hid DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Activity
	for rows.Next() {
		var a model.Activity
		rows.Scan(&a.HID, &a.Name, &a.YaoQiu, &a.Type, &a.Num, &a.Money, &a.AddTime, &a.EndTime, &a.StatusOK, &a.Status)
		list = append(list, a)
	}
	if list == nil {
		list = []model.Activity{}
	}
	return list, nil
}

func listPublicPledgeConfigs() ([]model.PledgeConfig, error) {
	rows, err := database.DB.Query(`
		SELECT c.id, c.category_id, c.amount, c.discount_rate, c.status, c.addtime, c.days, c.cancel_fee,
		       COALESCE(f.name,'未知分类')
		FROM qingka_wangke_zhiya_config c
		LEFT JOIN qingka_wangke_fenlei f ON c.category_id = f.id
		WHERE c.status = 1
		ORDER BY c.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.PledgeConfig
	for rows.Next() {
		var p model.PledgeConfig
		rows.Scan(&p.ID, &p.CategoryID, &p.Amount, &p.DiscountRate, &p.Status, &p.AddTime, &p.Days, &p.CancelFee, &p.CategoryName)
		list = append(list, p)
	}
	if list == nil {
		list = []model.PledgeConfig{}
	}
	return list, nil
}

func createPledge(uid, configID int) error {
	var cfg model.PledgeConfig
	err := database.DB.QueryRow(
		"SELECT id, amount, discount_rate, days, status FROM qingka_wangke_zhiya_config WHERE id=?", configID,
	).Scan(&cfg.ID, &cfg.Amount, &cfg.DiscountRate, &cfg.Days, &cfg.Status)
	if err == sql.ErrNoRows {
		return fmt.Errorf("质押配置不存在")
	}
	if err != nil {
		return err
	}
	if cfg.Status != 1 {
		return fmt.Errorf("该质押配置已禁用")
	}

	var cnt int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_zhiya_records WHERE uid=? AND config_id=? AND status=1", uid, configID).Scan(&cnt)
	if cnt > 0 {
		return fmt.Errorf("您已有该分类的生效质押")
	}

	var money float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid=?", uid).Scan(&money)
	if money < cfg.Amount {
		return fmt.Errorf("余额不足，需要%.2f元", cfg.Amount)
	}

	_, err = database.DB.Exec("UPDATE qingka_wangke_user SET money=money-? WHERE uid=?", cfg.Amount, uid)
	if err != nil {
		return err
	}

	_, err = database.DB.Exec(
		"INSERT INTO qingka_wangke_zhiya_records (uid, config_id, status, addtime) VALUES (?,?,1,NOW())",
		uid, configID,
	)
	if err != nil {
		return err
	}

	logAuxMoney(uid, "质押", fmt.Sprintf("质押%.2f元，享受折扣率%.2f", cfg.Amount, cfg.DiscountRate), -cfg.Amount)
	return nil
}

func cancelPledge(uid, recordID int) error {
	var rec model.PledgeRecord
	err := database.DB.QueryRow(
		"SELECT id, uid, config_id, status FROM qingka_wangke_zhiya_records WHERE id=?", recordID,
	).Scan(&rec.ID, &rec.UID, &rec.ConfigID, &rec.Status)
	if err == sql.ErrNoRows {
		return fmt.Errorf("质押记录不存在")
	}
	if err != nil {
		return err
	}
	if rec.UID != uid {
		return fmt.Errorf("无权操作")
	}
	if rec.Status != 1 {
		return fmt.Errorf("该质押已退还")
	}

	var cfg model.PledgeConfig
	database.DB.QueryRow("SELECT amount, cancel_fee FROM qingka_wangke_zhiya_config WHERE id=?", rec.ConfigID).Scan(&cfg.Amount, &cfg.CancelFee)

	refund := cfg.Amount * (1 - cfg.CancelFee)

	_, err = database.DB.Exec("UPDATE qingka_wangke_zhiya_records SET status=0, endtime=NOW() WHERE id=?", recordID)
	if err != nil {
		return err
	}

	_, err = database.DB.Exec("UPDATE qingka_wangke_user SET money=money+? WHERE uid=?", refund, uid)
	if err != nil {
		return err
	}

	logAuxMoney(uid, "取消质押", fmt.Sprintf("取消质押，退还%.2f元（扣费比例%.0f%%）", refund, cfg.CancelFee*100), refund)
	return nil
}

func listActivePledges(uid int) ([]model.PledgeRecord, error) {
	rows, err := database.DB.Query(`
		SELECT r.id, r.config_id, r.addtime, c.amount, COALESCE(f.name,''), c.discount_rate, c.days
		FROM qingka_wangke_zhiya_records r
		LEFT JOIN qingka_wangke_zhiya_config c ON r.config_id = c.id
		LEFT JOIN qingka_wangke_fenlei f ON c.category_id = f.id
		WHERE r.uid=? AND r.status=1 ORDER BY r.id DESC`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.PledgeRecord
	for rows.Next() {
		var p model.PledgeRecord
		rows.Scan(&p.ID, &p.ConfigID, &p.AddTime, &p.Amount, &p.CatName, &p.Discount, &p.Days)
		p.UID = uid
		p.Status = 1
		list = append(list, p)
	}
	if list == nil {
		list = []model.PledgeRecord{}
	}
	return list, nil
}

func checkOrder(req model.CheckOrderRequest) ([]model.CheckOrderResult, error) {
	where := "1=1"
	var args []interface{}

	if req.OID != "" {
		where += " AND o.oid = ?"
		args = append(args, req.OID)
	}
	if req.User != "" {
		where += " AND o.user = ?"
		args = append(args, req.User)
	}
	if req.KCName != "" {
		where += " AND o.kcname LIKE ?"
		args = append(args, "%"+req.KCName+"%")
	}

	if req.OID == "" && req.User == "" {
		return nil, fmt.Errorf("请输入订单号或账号进行查询")
	}

	rows, err := database.DB.Query(`
		SELECT o.oid, COALESCE(c.name, o.ptname, ''),
		CONCAT(COALESCE(NULLIF(o.school,''),'自动识别'),' ',o.user,' ',o.pass),
		COALESCE(o.school,''), o.kcname, o.status, COALESCE(o.process,''), COALESCE(o.remarks,''), o.addtime,
		COALESCE(o.pushUid,''), COALESCE(o.pushStatus,''), COALESCE(o.pushEmail,''),
		COALESCE(o.pushEmailStatus,'0'), COALESCE(o.showdoc_push_url,''), COALESCE(o.pushShowdocStatus,'0')
		FROM qingka_wangke_order o
		LEFT JOIN qingka_wangke_class c ON o.cid = c.cid
		WHERE `+where+`
		ORDER BY o.oid DESC LIMIT 50`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.CheckOrderResult
	for rows.Next() {
		var r model.CheckOrderResult
		rows.Scan(&r.OID, &r.PtName, &r.Account, &r.School, &r.KCName, &r.Status, &r.Process, &r.Remarks, &r.AddTime,
			&r.PushUid, &r.PushStatus, &r.PushEmail, &r.PushEmailStatus, &r.ShowdocPushURL, &r.PushShowdocStatus)
		list = append(list, r)
	}
	if list == nil {
		list = []model.CheckOrderResult{}
	}
	return list, nil
}
