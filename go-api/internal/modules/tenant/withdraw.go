package tenant

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-api/internal/database"
	"go-api/internal/model"
)

func normalizeCUserWithdrawMethod(method string) string {
	method = strings.TrimSpace(method)
	if method == "" {
		return "manual"
	}
	return method
}

func (s *Service) CreateCUserWithdrawRequest(tid, cUID int, req model.WithdrawCreateRequest) (int64, error) {
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

	var available float64
	if err := tx.QueryRow(
		"SELECT COALESCE(commission_money,0) FROM qingka_c_user WHERE id=? AND tid=? FOR UPDATE",
		cUID, tid,
	).Scan(&available); err != nil {
		return 0, errors.New("会员不存在")
	}
	if available < amount {
		return 0, fmt.Errorf("可提现佣金不足，当前余额 %.2f 元", available)
	}

	if _, err := tx.Exec(
		"UPDATE qingka_c_user SET commission_money = commission_money - ?, commission_cdmoney = commission_cdmoney + ? WHERE id=? AND tid=?",
		amount, amount, cUID, tid,
	); err != nil {
		return 0, errors.New("冻结提现金额失败")
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	res, err := tx.Exec(
		`INSERT INTO qingka_c_user_withdraw_request (tid, c_uid, amount, method, account_name, account_no, bank_name, note, status, addtime)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, ?)`,
		tid, cUID, amount, normalizeCUserWithdrawMethod(req.Method), strings.TrimSpace(req.AccountName),
		strings.TrimSpace(req.AccountNo), strings.TrimSpace(req.BankName), strings.TrimSpace(req.Note), now,
	)
	if err != nil {
		return 0, errors.New("创建提现申请失败")
	}
	requestID, _ := res.LastInsertId()

	if _, err := tx.Exec(
		`INSERT INTO qingka_c_user_commission_log (tid, c_uid, pay_order_id, out_trade_no, buyer_c_uid, buyer_account, amount, rate, status, remark, addtime)
		 VALUES (?, ?, 0, '', 0, '', ?, 0, 2, ?, ?)`,
		tid, cUID, -amount, fmt.Sprintf("佣金提现申请#%d 冻结 %.2f 元", requestID, amount), now,
	); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, errors.New("提交提现申请失败")
	}
	return requestID, nil
}

func (s *Service) CUserWithdrawRequests(tid, cUID int, req model.WithdrawListRequest) ([]model.CUserWithdrawRequest, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	where := "tid=? AND c_uid=?"
	args := []interface{}{tid, cUID}
	if req.Status != nil && (*req.Status == -1 || *req.Status == 0 || *req.Status == 1) {
		where += " AND status=?"
		args = append(args, *req.Status)
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_c_user_withdraw_request WHERE "+where, args...).Scan(&total)

	offset := (req.Page - 1) * req.Limit
	args = append(args, req.Limit, offset)
	rows, err := database.DB.Query(
		fmt.Sprintf(`SELECT id, tid, c_uid, COALESCE(amount,0), COALESCE(method,''), COALESCE(account_name,''), COALESCE(account_no,''), COALESCE(bank_name,''), COALESCE(note,''), COALESCE(status,0), COALESCE(audit_remark,''), COALESCE(audit_uid,0), COALESCE(DATE_FORMAT(addtime,'%%Y-%%m-%%d %%H:%%i:%%s'),''), COALESCE(DATE_FORMAT(audit_time,'%%Y-%%m-%%d %%H:%%i:%%s'),'')
		FROM qingka_c_user_withdraw_request WHERE %s ORDER BY id DESC LIMIT ? OFFSET ?`, where),
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.CUserWithdrawRequest
	for rows.Next() {
		var item model.CUserWithdrawRequest
		rows.Scan(&item.ID, &item.TID, &item.CUID, &item.Amount, &item.Method, &item.AccountName, &item.AccountNo, &item.BankName, &item.Note, &item.Status, &item.AuditRemark, &item.AuditUID, &item.AddTime, &item.AuditTime)
		list = append(list, item)
	}
	if list == nil {
		list = []model.CUserWithdrawRequest{}
	}
	return list, total, nil
}

func adminCUserWithdrawRequests(req model.AdminCUserWithdrawListRequest) ([]model.CUserWithdrawRequest, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	where := "1=1"
	args := make([]interface{}, 0, 4)
	if strings.TrimSpace(req.TID) != "" {
		where += " AND w.tid=?"
		args = append(args, strings.TrimSpace(req.TID))
	}
	if strings.TrimSpace(req.CUID) != "" {
		where += " AND w.c_uid=?"
		args = append(args, strings.TrimSpace(req.CUID))
	}
	if strings.TrimSpace(req.Status) != "" {
		where += " AND w.status=?"
		args = append(args, strings.TrimSpace(req.Status))
	}

	var total int64
	database.DB.QueryRow("SELECT COUNT(*) FROM qingka_c_user_withdraw_request w WHERE "+where, args...).Scan(&total)

	offset := (req.Page - 1) * req.Limit
	args = append(args, req.Limit, offset)
	rows, err := database.DB.Query(
		fmt.Sprintf(`SELECT w.id, w.tid, w.c_uid, COALESCE(cu.account,''), COALESCE(cu.nickname,''), COALESCE(w.amount,0), COALESCE(w.method,''), COALESCE(w.account_name,''), COALESCE(w.account_no,''), COALESCE(w.bank_name,''), COALESCE(w.note,''), COALESCE(w.status,0), COALESCE(w.audit_remark,''), COALESCE(w.audit_uid,0), COALESCE(au.user,''), COALESCE(DATE_FORMAT(w.addtime,'%%Y-%%m-%%d %%H:%%i:%%s'),''), COALESCE(DATE_FORMAT(w.audit_time,'%%Y-%%m-%%d %%H:%%i:%%s'),'')
		FROM qingka_c_user_withdraw_request w
		LEFT JOIN qingka_c_user cu ON cu.id = w.c_uid
		LEFT JOIN qingka_wangke_user au ON au.uid = w.audit_uid
		WHERE %s ORDER BY w.id DESC LIMIT ? OFFSET ?`, where),
		args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.CUserWithdrawRequest
	for rows.Next() {
		var item model.CUserWithdrawRequest
		rows.Scan(&item.ID, &item.TID, &item.CUID, &item.Account, &item.Nickname, &item.Amount, &item.Method, &item.AccountName, &item.AccountNo, &item.BankName, &item.Note, &item.Status, &item.AuditRemark, &item.AuditUID, &item.AuditUser, &item.AddTime, &item.AuditTime)
		list = append(list, item)
	}
	if list == nil {
		list = []model.CUserWithdrawRequest{}
	}
	return list, total, nil
}

func adminReviewCUserWithdrawRequest(adminUID, id int, status int, remark string) error {
	return reviewCUserWithdrawRequest(adminUID, 0, id, status, remark)
}

func reviewCUserWithdrawRequest(reviewerUID, reviewerTID, id int, status int, remark string) error {
	if status != 1 && status != -1 {
		return errors.New("审核状态无效")
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return errors.New("系统繁忙，请稍后重试")
	}
	defer tx.Rollback()

	var req model.CUserWithdrawRequest
	err = tx.QueryRow(
		"SELECT id, tid, c_uid, COALESCE(amount,0), COALESCE(status,0) FROM qingka_c_user_withdraw_request WHERE id=? FOR UPDATE",
		id,
	).Scan(&req.ID, &req.TID, &req.CUID, &req.Amount, &req.Status)
	if err == sql.ErrNoRows {
		return errors.New("提现申请不存在")
	}
	if err != nil {
		return err
	}
	if req.Status != 0 {
		return errors.New("该提现申请已处理")
	}
	if reviewerTID > 0 && reviewerTID != req.TID {
		return errors.New("无权审核该提现申请")
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	if status == 1 {
		var freezeMoney float64
		if err := tx.QueryRow("SELECT COALESCE(commission_cdmoney,0) FROM qingka_c_user WHERE id=? AND tid=? FOR UPDATE", req.CUID, req.TID).Scan(&freezeMoney); err != nil {
			return errors.New("会员不存在")
		}
		if freezeMoney < req.Amount {
			return errors.New("冻结佣金不足，无法审核通过")
		}
		if _, err := tx.Exec("UPDATE qingka_c_user SET commission_cdmoney = commission_cdmoney - ? WHERE id=? AND tid=?", req.Amount, req.CUID, req.TID); err != nil {
			return errors.New("扣减冻结佣金失败")
		}
		if _, err := tx.Exec(
			"UPDATE qingka_c_user_withdraw_request SET status=1, audit_remark=?, audit_uid=?, audit_time=? WHERE id=?",
			strings.TrimSpace(remark), reviewerUID, now, id,
		); err != nil {
			return errors.New("更新提现状态失败")
		}
		if _, err := tx.Exec(
			`INSERT INTO qingka_c_user_commission_log (tid, c_uid, pay_order_id, out_trade_no, buyer_c_uid, buyer_account, amount, rate, status, remark, addtime)
			 VALUES (?, ?, 0, '', 0, '', ?, 0, 3, ?, ?)`,
			req.TID, req.CUID, 0, fmt.Sprintf("佣金提现申请#%d 商家已线下打款 %.2f 元", id, req.Amount), now,
		); err != nil {
			return err
		}
	} else {
		var freezeMoney float64
		if err := tx.QueryRow("SELECT COALESCE(commission_cdmoney,0) FROM qingka_c_user WHERE id=? AND tid=? FOR UPDATE", req.CUID, req.TID).Scan(&freezeMoney); err != nil {
			return errors.New("会员不存在")
		}
		if freezeMoney < req.Amount {
			return errors.New("冻结佣金不足，无法驳回")
		}
		if _, err := tx.Exec("UPDATE qingka_c_user SET commission_cdmoney = commission_cdmoney - ?, commission_money = commission_money + ? WHERE id=? AND tid=?", req.Amount, req.Amount, req.CUID, req.TID); err != nil {
			return errors.New("退回佣金失败")
		}
		if _, err := tx.Exec(
			"UPDATE qingka_c_user_withdraw_request SET status=-1, audit_remark=?, audit_uid=?, audit_time=? WHERE id=?",
			strings.TrimSpace(remark), reviewerUID, now, id,
		); err != nil {
			return errors.New("更新提现状态失败")
		}
		if _, err := tx.Exec(
			`INSERT INTO qingka_c_user_commission_log (tid, c_uid, pay_order_id, out_trade_no, buyer_c_uid, buyer_account, amount, rate, status, remark, addtime)
			 VALUES (?, ?, 0, '', 0, '', ?, 0, 4, ?, ?)`,
			req.TID, req.CUID, req.Amount, fmt.Sprintf("佣金提现申请#%d 驳回退回 %.2f 元", id, req.Amount), now,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func AdminCUserWithdrawRequests(req model.AdminCUserWithdrawListRequest) ([]model.CUserWithdrawRequest, int64, error) {
	return adminCUserWithdrawRequests(req)
}

func AdminReviewCUserWithdrawRequest(adminUID, id int, status int, remark string) error {
	return adminReviewCUserWithdrawRequest(adminUID, id, status, remark)
}

func (s *Service) TenantCUserWithdrawRequests(tid int, req model.AdminCUserWithdrawListRequest) ([]model.CUserWithdrawRequest, int64, error) {
	req.TID = fmt.Sprintf("%d", tid)
	return adminCUserWithdrawRequests(req)
}

func (s *Service) ReviewTenantCUserWithdrawRequest(reviewerUID, tid, id, status int, remark string) error {
	return reviewCUserWithdrawRequest(reviewerUID, tid, id, status, remark)
}
