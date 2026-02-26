package service

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

// wlog 已在 agent.go 中定义，此处直接复用

type AuxiliaryService struct{}

func NewAuxiliaryService() *AuxiliaryService {
	return &AuxiliaryService{}
}

// ===== 卡密系统 =====

func (s *AuxiliaryService) CardKeyGenerate(money, count int) ([]string, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	var codes []string
	for i := 0; i < count; i++ {
		code := generateRandomCode(16)
		_, err := database.DB.Exec(
			"INSERT INTO qingka_wangke_km (content, money, status, addtime) VALUES (?, ?, 0, ?)",
			code, money, now,
		)
		if err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}
	return codes, nil
}

func (s *AuxiliaryService) CardKeyList(req model.CardKeyListRequest) ([]model.CardKey, int, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}
	offset := (req.Page - 1) * req.Limit

	where := "1=1"
	var args []interface{}
	if req.Status == 1 || req.Status == 0 {
		// 仅当前端显式传了 status 时才过滤（通过 query ?status=0 或 ?status=1）
	}

	var total int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_km WHERE "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	args = append(args, req.Limit, offset)
	rows, err := database.DB.Query(
		"SELECT id, content, money, COALESCE(status,0), COALESCE(uid,0), COALESCE(addtime,''), COALESCE(usedtime,'') FROM qingka_wangke_km WHERE "+where+" ORDER BY id DESC LIMIT ? OFFSET ?",
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.CardKey
	for rows.Next() {
		var ck model.CardKey
		var uid int
		rows.Scan(&ck.ID, &ck.Content, &ck.Money, &ck.Status, &uid, &ck.AddTime, &ck.UsedTime)
		if uid > 0 {
			ck.UID = &uid
		}
		list = append(list, ck)
	}
	if list == nil {
		list = []model.CardKey{}
	}
	return list, total, nil
}

func (s *AuxiliaryService) CardKeyUse(uid int, content string) (int, error) {
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

	wlog(uid, "卡密充值", fmt.Sprintf("使用卡密充值%d元成功", ck.Money), float64(ck.Money))

	return ck.Money, nil
}

func (s *AuxiliaryService) CardKeyDelete(ids []int) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	query := "DELETE FROM qingka_wangke_km WHERE status=0 AND id IN ("
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = id
	}
	query += ")"
	result, err := database.DB.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	affected, _ := result.RowsAffected()
	return int(affected), nil
}

// ===== 活动系统 =====

func (s *AuxiliaryService) ActivityList(req model.ActivityListRequest) ([]model.Activity, int, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}
	offset := (req.Page - 1) * req.Limit

	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_huodong").Scan(&total)

	rows, err := database.DB.Query(
		"SELECT hid, name, yaoqiu, type, num, money, addtime, endtime, status_ok, status FROM qingka_wangke_huodong ORDER BY hid DESC LIMIT ? OFFSET ?",
		req.Limit, offset,
	)
	if err != nil {
		return nil, 0, err
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
	return list, total, nil
}

func (s *AuxiliaryService) ActivitySave(req model.ActivitySaveRequest) error {
	if req.StatusOK == "" {
		req.StatusOK = "1"
	}
	if req.HID > 0 {
		_, err := database.DB.Exec(
			"UPDATE qingka_wangke_huodong SET name=?, yaoqiu=?, type=?, num=?, money=?, addtime=?, endtime=?, status_ok=? WHERE hid=?",
			req.Name, req.YaoQiu, req.Type, req.Num, req.Money, req.AddTime, req.EndTime, req.StatusOK, req.HID,
		)
		return err
	}
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_huodong (name, yaoqiu, type, num, money, addtime, endtime, status_ok, status) VALUES (?,?,?,?,?,?,?,?,?)",
		req.Name, req.YaoQiu, req.Type, req.Num, req.Money, req.AddTime, req.EndTime, req.StatusOK, "1",
	)
	return err
}

func (s *AuxiliaryService) ActivityDelete(hid int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_wangke_huodong WHERE hid=?", hid)
	return err
}

func (s *AuxiliaryService) ActivityListPublic() ([]model.Activity, error) {
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

// ===== 质押系统 =====

func (s *AuxiliaryService) PledgeConfigList() ([]model.PledgeConfig, error) {
	rows, err := database.DB.Query(`
		SELECT c.id, c.category_id, c.amount, c.discount_rate, c.status, c.addtime, c.days, c.cancel_fee,
		       COALESCE(f.name,'未知分类')
		FROM qingka_wangke_zhiya_config c
		LEFT JOIN qingka_wangke_fenlei f ON c.category_id = f.id
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

func (s *AuxiliaryService) PledgeConfigSave(req model.PledgeConfigSaveRequest) error {
	if req.ID > 0 {
		_, err := database.DB.Exec(
			"UPDATE qingka_wangke_zhiya_config SET category_id=?, amount=?, discount_rate=?, days=?, cancel_fee=?, status=1 WHERE id=?",
			req.CategoryID, req.Amount, req.DiscountRate, req.Days, req.CancelFee, req.ID,
		)
		return err
	}
	_, err := database.DB.Exec(
		"INSERT INTO qingka_wangke_zhiya_config (category_id, amount, discount_rate, days, cancel_fee, status, addtime) VALUES (?,?,?,?,?,1,NOW())",
		req.CategoryID, req.Amount, req.DiscountRate, req.Days, req.CancelFee,
	)
	return err
}

func (s *AuxiliaryService) PledgeConfigDelete(id int) error {
	_, err := database.DB.Exec("DELETE FROM qingka_wangke_zhiya_config WHERE id=?", id)
	return err
}

func (s *AuxiliaryService) PledgeConfigToggle(id, status int) error {
	_, err := database.DB.Exec("UPDATE qingka_wangke_zhiya_config SET status=? WHERE id=?", status, id)
	return err
}

func (s *AuxiliaryService) PledgeCreate(uid, configID int) error {
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

	// 检查是否已有生效的同配置质押
	var cnt int
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_zhiya_records WHERE uid=? AND config_id=? AND status=1", uid, configID).Scan(&cnt)
	if cnt > 0 {
		return fmt.Errorf("您已有该分类的生效质押")
	}

	// 检查余额
	var money float64
	database.DB.QueryRow("SELECT money FROM qingka_wangke_user WHERE uid=?", uid).Scan(&money)
	if money < cfg.Amount {
		return fmt.Errorf("余额不足，需要%.2f元", cfg.Amount)
	}

	// 扣费
	_, err = database.DB.Exec("UPDATE qingka_wangke_user SET money=money-? WHERE uid=?", cfg.Amount, uid)
	if err != nil {
		return err
	}

	// 创建记录
	_, err = database.DB.Exec(
		"INSERT INTO qingka_wangke_zhiya_records (uid, config_id, status, addtime) VALUES (?,?,1,NOW())",
		uid, configID,
	)
	if err != nil {
		return err
	}

	wlog(uid, "质押", fmt.Sprintf("质押%.2f元，享受折扣率%.2f", cfg.Amount, cfg.DiscountRate), -cfg.Amount)
	return nil
}

func (s *AuxiliaryService) PledgeCancel(uid, recordID int) error {
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

	wlog(uid, "取消质押", fmt.Sprintf("取消质押，退还%.2f元（扣费比例%.0f%%）", refund, cfg.CancelFee*100), refund)
	return nil
}

func (s *AuxiliaryService) PledgeRecordList(req model.PledgeListRequest) ([]model.PledgeRecord, int, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}
	offset := (req.Page - 1) * req.Limit

	where := "1=1"
	var args []interface{}
	if req.UID > 0 {
		where += " AND r.uid=?"
		args = append(args, req.UID)
	}

	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_wangke_zhiya_records r WHERE "+where, countArgs...).Scan(&total)

	args = append(args, req.Limit, offset)
	rows, err := database.DB.Query(`
		SELECT r.id, r.uid, r.config_id, r.status, r.addtime, r.endtime,
		       c.amount, COALESCE(f.name,''), c.discount_rate, c.days, COALESCE(u.user,'')
		FROM qingka_wangke_zhiya_records r
		LEFT JOIN qingka_wangke_zhiya_config c ON r.config_id = c.id
		LEFT JOIN qingka_wangke_fenlei f ON c.category_id = f.id
		LEFT JOIN qingka_wangke_user u ON r.uid = u.uid
		WHERE `+where+` ORDER BY r.id DESC LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.PledgeRecord
	for rows.Next() {
		var p model.PledgeRecord
		var endtime sql.NullString
		rows.Scan(&p.ID, &p.UID, &p.ConfigID, &p.Status, &p.AddTime, &endtime,
			&p.Amount, &p.CatName, &p.Discount, &p.Days, &p.Username)
		if endtime.Valid {
			p.EndTime = &endtime.String
		}
		list = append(list, p)
	}
	if list == nil {
		list = []model.PledgeRecord{}
	}
	return list, total, nil
}

func (s *AuxiliaryService) UserActivePledges(uid int) ([]model.PledgeRecord, error) {
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

// ===== 网签系统 =====

func (s *AuxiliaryService) MlsxCompanyList() ([]model.MlsxCompany, error) {
	rows, err := database.DB.Query("SELECT id, qymc, COALESCE(wqbs,''), shijian FROM mlsx_gslb ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.MlsxCompany
	for rows.Next() {
		var c model.MlsxCompany
		rows.Scan(&c.ID, &c.QYMC, &c.WQBS, &c.ShiJian)
		list = append(list, c)
	}
	if list == nil {
		list = []model.MlsxCompany{}
	}
	return list, nil
}

func (s *AuxiliaryService) MlsxCompanySave(req model.MlsxCompanySaveRequest) error {
	if req.ID > 0 {
		_, err := database.DB.Exec("UPDATE mlsx_gslb SET qymc=?, wqbs=? WHERE id=?", req.QYMC, req.WQBS, req.ID)
		return err
	}
	_, err := database.DB.Exec("INSERT INTO mlsx_gslb (qymc, wqbs) VALUES (?, ?)", req.QYMC, req.WQBS)
	return err
}

func (s *AuxiliaryService) MlsxCompanyDelete(id int) error {
	_, err := database.DB.Exec("DELETE FROM mlsx_gslb WHERE id=?", id)
	return err
}

func (s *AuxiliaryService) MlsxFileList(wjid string) ([]model.MlsxFile, error) {
	rows, err := database.DB.Query(
		"SELECT id, wjid, name, COALESCE(ip,''), shijian FROM mlsx_wj_wq WHERE wjid=? ORDER BY id DESC", wjid,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.MlsxFile
	for rows.Next() {
		var f model.MlsxFile
		rows.Scan(&f.ID, &f.WJID, &f.Name, &f.IP, &f.ShiJian)
		list = append(list, f)
	}
	if list == nil {
		list = []model.MlsxFile{}
	}
	return list, nil
}

func (s *AuxiliaryService) MlsxFileSave(wjid, name, ip string) (int64, error) {
	result, err := database.DB.Exec(
		"INSERT INTO mlsx_wj_wq (wjid, name, ip) VALUES (?, ?, ?)", wjid, name, ip,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *AuxiliaryService) MlsxFileDelete(id int) error {
	_, err := database.DB.Exec("DELETE FROM mlsx_wj_wq WHERE id=?", id)
	return err
}

// ===== 外部查单 =====

func (s *AuxiliaryService) CheckOrder(req model.CheckOrderRequest) ([]model.CheckOrderResult, error) {
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

// ===== 工具函数 =====

func generateRandomCode(length int) string {
	b := make([]byte, length/2)
	rand.Read(b)
	return hex.EncodeToString(b)
}
