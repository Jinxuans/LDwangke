package user

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

func normalizeWithdrawMethod(method string) string {
	method = strings.TrimSpace(method)
	if method == "" {
		return "manual"
	}
	return method
}

func (s *Service) CreateWithdrawRequest(uid int, req model.WithdrawCreateRequest) (int64, error) {
	amount := req.Amount
	if amount <= 0 {
		return 0, errors.New("提现金额必须大于0")
	}
	amount = float64(int(amount*100+0.5)) / 100
	if amount < 0.01 {
		return 0, errors.New("提现金额不能小于0.01元")
	}
	if strings.TrimSpace(req.AccountName) == "" {
		return 0, errors.New("请填写收款人")
	}
	if strings.TrimSpace(req.AccountNo) == "" {
		return 0, errors.New("请填写收款账号")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return 0, errors.New("系统繁忙，请稍后重试")
	}
	defer tx.Rollback()

	var mallMoney float64
	if err := tx.QueryRow("SELECT COALESCE(mall_money,0) FROM qingka_wangke_user WHERE uid=? FOR UPDATE", uid).Scan(&mallMoney); err != nil {
		return 0, errors.New("用户不存在")
	}
	if mallMoney < amount {
		return 0, fmt.Errorf("商城钱包余额不足，当前余额 %.2f 元", mallMoney)
	}

	if _, err := tx.Exec("UPDATE qingka_wangke_user SET mall_money = mall_money - ?, mall_cdmoney = mall_cdmoney + ? WHERE uid = ?", amount, amount, uid); err != nil {
		return 0, errors.New("冻结提现金额失败")
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	res, err := tx.Exec(
		`INSERT INTO qingka_withdraw_request (uid, amount, method, account_name, account_no, bank_name, note, status, addtime)
		 VALUES (?, ?, ?, ?, ?, ?, ?, 0, ?)`,
		uid, amount, normalizeWithdrawMethod(req.Method), strings.TrimSpace(req.AccountName),
		strings.TrimSpace(req.AccountNo), strings.TrimSpace(req.BankName), strings.TrimSpace(req.Note), now,
	)
	if err != nil {
		return 0, errors.New("创建提现申请失败")
	}
	requestID, _ := res.LastInsertId()

	if _, err := tx.Exec(
		"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, '商城提现申请', ?, (SELECT mall_money FROM qingka_wangke_user WHERE uid = ?), ?, ?)",
		uid, -amount, uid, fmt.Sprintf("提现申请#%d 冻结 %.2f 元", requestID, amount), now,
	); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, errors.New("提交提现申请失败")
	}
	return requestID, nil
}

func (s *Service) WithdrawRequests(uid int, req model.WithdrawListRequest) ([]model.WithdrawRequest, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	where := "uid=?"
	args := []interface{}{uid}
	if req.Status != nil && (*req.Status == -1 || *req.Status == 0 || *req.Status == 1) {
		where += " AND status=?"
		args = append(args, *req.Status)
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_withdraw_request WHERE "+where, args...).Scan(&total)

	offset := (req.Page - 1) * req.Limit
	args = append(args, req.Limit, offset)
	rows, err := database.DB.Query(
		fmt.Sprintf(`SELECT id, uid, COALESCE(amount,0), COALESCE(method,''), COALESCE(account_name,''), COALESCE(account_no,''), COALESCE(bank_name,''), COALESCE(note,''), COALESCE(status,0), COALESCE(audit_remark,''), COALESCE(audit_uid,0), COALESCE(DATE_FORMAT(addtime,'%%Y-%%m-%%d %%H:%%i:%%s'),''), COALESCE(DATE_FORMAT(audit_time,'%%Y-%%m-%%d %%H:%%i:%%s'),'')
		FROM qingka_withdraw_request WHERE %s ORDER BY id DESC LIMIT ? OFFSET ?`, where),
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.WithdrawRequest
	for rows.Next() {
		var item model.WithdrawRequest
		rows.Scan(&item.ID, &item.UID, &item.Amount, &item.Method, &item.AccountName, &item.AccountNo, &item.BankName, &item.Note, &item.Status, &item.AuditRemark, &item.AuditUID, &item.AddTime, &item.AuditTime)
		list = append(list, item)
	}
	if list == nil {
		list = []model.WithdrawRequest{}
	}
	return list, total, nil
}

func adminWithdrawRequests(req model.AdminWithdrawListRequest) ([]model.WithdrawRequest, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	where := "1=1"
	args := make([]interface{}, 0, 4)
	if strings.TrimSpace(req.UID) != "" {
		where += " AND w.uid = ?"
		args = append(args, strings.TrimSpace(req.UID))
	}
	if strings.TrimSpace(req.Status) != "" {
		where += " AND w.status = ?"
		args = append(args, strings.TrimSpace(req.Status))
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_withdraw_request w WHERE "+where, args...).Scan(&total)

	offset := (req.Page - 1) * req.Limit
	args = append(args, req.Limit, offset)
	rows, err := database.DB.Query(
		fmt.Sprintf(`SELECT w.id, w.uid, COALESCE(u.user,''), COALESCE(w.amount,0), COALESCE(w.method,''), COALESCE(w.account_name,''), COALESCE(w.account_no,''), COALESCE(w.bank_name,''), COALESCE(w.note,''), COALESCE(w.status,0), COALESCE(w.audit_remark,''), COALESCE(w.audit_uid,0), COALESCE(au.user,''), COALESCE(DATE_FORMAT(w.addtime,'%%Y-%%m-%%d %%H:%%i:%%s'),''), COALESCE(DATE_FORMAT(w.audit_time,'%%Y-%%m-%%d %%H:%%i:%%s'),'')
		FROM qingka_withdraw_request w
		LEFT JOIN qingka_wangke_user u ON u.uid = w.uid
		LEFT JOIN qingka_wangke_user au ON au.uid = w.audit_uid
		WHERE %s
		ORDER BY w.id DESC LIMIT ? OFFSET ?`, where),
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.WithdrawRequest
	for rows.Next() {
		var item model.WithdrawRequest
		rows.Scan(&item.ID, &item.UID, &item.Username, &item.Amount, &item.Method, &item.AccountName, &item.AccountNo, &item.BankName, &item.Note, &item.Status, &item.AuditRemark, &item.AuditUID, &item.AuditUser, &item.AddTime, &item.AuditTime)
		list = append(list, item)
	}
	if list == nil {
		list = []model.WithdrawRequest{}
	}
	return list, total, nil
}

func adminReviewWithdrawRequest(adminUID, id int, status int, remark string) error {
	if status != 1 && status != -1 {
		return errors.New("审核状态无效")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return errors.New("系统繁忙，请稍后重试")
	}
	defer tx.Rollback()

	var req model.WithdrawRequest
	err = tx.QueryRow(
		"SELECT id, uid, COALESCE(amount,0), COALESCE(status,0) FROM qingka_withdraw_request WHERE id=? FOR UPDATE",
		id,
	).Scan(&req.ID, &req.UID, &req.Amount, &req.Status)
	if err == sql.ErrNoRows {
		return errors.New("提现申请不存在")
	}
	if err != nil {
		return err
	}
	if req.Status != 0 {
		return errors.New("该提现申请已处理")
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	if status == 1 {
		var freezeMoney float64
		if err := tx.QueryRow("SELECT COALESCE(mall_cdmoney,0) FROM qingka_wangke_user WHERE uid=? FOR UPDATE", req.UID).Scan(&freezeMoney); err != nil {
			return errors.New("用户不存在")
		}
		if freezeMoney < req.Amount {
			return errors.New("冻结余额不足，无法审核通过")
		}
		if _, err := tx.Exec("UPDATE qingka_wangke_user SET mall_cdmoney = mall_cdmoney - ? WHERE uid = ?", req.Amount, req.UID); err != nil {
			return errors.New("扣减冻结余额失败")
		}
		if _, err := tx.Exec(
			"UPDATE qingka_withdraw_request SET status=1, audit_remark=?, audit_uid=?, audit_time=? WHERE id=?",
			strings.TrimSpace(remark), adminUID, now, id,
		); err != nil {
			return errors.New("更新提现状态失败")
		}
		if _, err := tx.Exec(
			"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, '商城提现通过', 0, (SELECT mall_money FROM qingka_wangke_user WHERE uid = ?), ?, ?)",
			req.UID, req.UID, fmt.Sprintf("提现申请#%d 审核通过 %.2f 元", id, req.Amount), now,
		); err != nil {
			return err
		}
	} else {
		var freezeMoney float64
		if err := tx.QueryRow("SELECT COALESCE(mall_cdmoney,0) FROM qingka_wangke_user WHERE uid=? FOR UPDATE", req.UID).Scan(&freezeMoney); err != nil {
			return errors.New("用户不存在")
		}
		if freezeMoney < req.Amount {
			return errors.New("冻结余额不足，无法驳回")
		}
		if _, err := tx.Exec("UPDATE qingka_wangke_user SET mall_cdmoney = mall_cdmoney - ?, mall_money = mall_money + ? WHERE uid = ?", req.Amount, req.Amount, req.UID); err != nil {
			return errors.New("退回余额失败")
		}
		if _, err := tx.Exec(
			"UPDATE qingka_withdraw_request SET status=-1, audit_remark=?, audit_uid=?, audit_time=? WHERE id=?",
			strings.TrimSpace(remark), adminUID, now, id,
		); err != nil {
			return errors.New("更新提现状态失败")
		}
		if _, err := tx.Exec(
			"INSERT INTO qingka_wangke_moneylog (uid, type, money, balance, remark, addtime) VALUES (?, '商城提现驳回', ?, (SELECT mall_money FROM qingka_wangke_user WHERE uid = ?), ?, ?)",
			req.UID, req.Amount, req.UID, fmt.Sprintf("提现申请#%d 驳回退回 %.2f 元", id, req.Amount), now,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func AdminWithdrawRequests(req model.AdminWithdrawListRequest) ([]model.WithdrawRequest, int64, error) {
	return adminWithdrawRequests(req)
}

func AdminReviewWithdrawRequest(adminUID, id int, status int, remark string) error {
	return adminReviewWithdrawRequest(adminUID, id, status, remark)
}
