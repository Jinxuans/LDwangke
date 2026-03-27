package model

type WithdrawRequest struct {
	ID          int     `json:"id"`
	UID         int     `json:"uid"`
	Username    string  `json:"username,omitempty"`
	Amount      float64 `json:"amount"`
	Method      string  `json:"method"`
	AccountName string  `json:"account_name"`
	AccountNo   string  `json:"account_no"`
	BankName    string  `json:"bank_name"`
	Note        string  `json:"note"`
	Status      int     `json:"status"`
	AuditRemark string  `json:"audit_remark"`
	AuditUID    int     `json:"audit_uid"`
	AuditUser   string  `json:"audit_user,omitempty"`
	AddTime     string  `json:"addtime"`
	AuditTime   string  `json:"audit_time"`
}

type WithdrawCreateRequest struct {
	Amount      float64 `json:"amount" binding:"required"`
	Method      string  `json:"method"`
	AccountName string  `json:"account_name" binding:"required"`
	AccountNo   string  `json:"account_no" binding:"required"`
	BankName    string  `json:"bank_name"`
	Note        string  `json:"note"`
}

type WithdrawListRequest struct {
	Page   int  `form:"page"`
	Limit  int  `form:"limit"`
	Status *int `form:"status"`
}

type AdminWithdrawListRequest struct {
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
	Status string `form:"status"`
	UID    string `form:"uid"`
}

type WithdrawReviewRequest struct {
	Status int    `json:"status" binding:"required"`
	Remark string `json:"remark"`
}

type CUserWithdrawRequest struct {
	ID          int     `json:"id"`
	TID         int     `json:"tid"`
	CUID        int     `json:"c_uid"`
	Account     string  `json:"account,omitempty"`
	Nickname    string  `json:"nickname,omitempty"`
	Amount      float64 `json:"amount"`
	Method      string  `json:"method"`
	AccountName string  `json:"account_name"`
	AccountNo   string  `json:"account_no"`
	BankName    string  `json:"bank_name"`
	Note        string  `json:"note"`
	Status      int     `json:"status"`
	AuditRemark string  `json:"audit_remark"`
	AuditUID    int     `json:"audit_uid"`
	AuditUser   string  `json:"audit_user,omitempty"`
	AddTime     string  `json:"addtime"`
	AuditTime   string  `json:"audit_time"`
}

type AdminCUserWithdrawListRequest struct {
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
	Status string `form:"status"`
	TID    string `form:"tid"`
	CUID   string `form:"c_uid"`
}
