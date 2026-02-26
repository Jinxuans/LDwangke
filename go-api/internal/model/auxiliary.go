package model

import "time"

// ===== 卡密系统 =====

type CardKey struct {
	ID       int    `json:"id" db:"id"`
	Content  string `json:"content" db:"content"`
	Money    int    `json:"money" db:"money"`
	Status   int    `json:"status" db:"status"`
	UID      *int   `json:"uid" db:"uid"`
	AddTime  string `json:"addtime" db:"addtime"`
	UsedTime string `json:"usedtime" db:"usedtime"`
}

type CardKeyGenRequest struct {
	Money int `json:"money" binding:"required,min=1"`
	Count int `json:"count" binding:"required,min=1,max=100"`
}

type CardKeyListRequest struct {
	Page   int `form:"page" json:"page"`
	Limit  int `form:"limit" json:"limit"`
	Status int `form:"status" json:"status" binding:"-"`
}

type CardKeyUseRequest struct {
	Content string `json:"content" binding:"required"`
}

// 等级系统的 Grade / GradeSaveRequest 已在 admin.go 中定义

// ===== 活动系统 =====

type Activity struct {
	HID      int    `json:"hid" db:"hid"`
	Name     string `json:"name" db:"name"`
	YaoQiu   string `json:"yaoqiu" db:"yaoqiu"`
	Type     string `json:"type" db:"type"`
	Num      string `json:"num" db:"num"`
	Money    string `json:"money" db:"money"`
	AddTime  string `json:"addtime" db:"addtime"`
	EndTime  string `json:"endtime" db:"endtime"`
	StatusOK string `json:"status_ok" db:"status_ok"`
	Status   string `json:"status" db:"status"`
}

type ActivitySaveRequest struct {
	HID      int    `json:"hid"`
	Name     string `json:"name" binding:"required"`
	YaoQiu   string `json:"yaoqiu" binding:"required"`
	Type     string `json:"type" binding:"required"`
	Num      string `json:"num" binding:"required"`
	Money    string `json:"money" binding:"required"`
	AddTime  string `json:"addtime" binding:"required"`
	EndTime  string `json:"endtime" binding:"required"`
	StatusOK string `json:"status_ok"`
}

type ActivityListRequest struct {
	Page  int `form:"page" json:"page"`
	Limit int `form:"limit" json:"limit"`
}

type ActivityRecord struct {
	ID       int       `json:"id" db:"id"`
	HID      int       `json:"hid" db:"hid"`
	UID      int       `json:"uid" db:"uid"`
	Progress int       `json:"progress" db:"progress"`
	Status   int       `json:"status" db:"status"`
	AddTime  time.Time `json:"addtime" db:"addtime"`
}

// ===== 质押系统 =====

type PledgeConfig struct {
	ID           int     `json:"id" db:"id"`
	CategoryID   int     `json:"category_id" db:"category_id"`
	Amount       float64 `json:"amount" db:"amount"`
	DiscountRate float64 `json:"discount_rate" db:"discount_rate"`
	Status       int     `json:"status" db:"status"`
	AddTime      string  `json:"addtime" db:"addtime"`
	Days         int     `json:"days" db:"days"`
	CancelFee    float64 `json:"cancel_fee" db:"cancel_fee"`
	CategoryName string  `json:"category_name,omitempty"`
}

type PledgeConfigSaveRequest struct {
	ID           int     `json:"id"`
	CategoryID   int     `json:"category_id" binding:"required"`
	Amount       float64 `json:"amount" binding:"required"`
	DiscountRate float64 `json:"discount_rate" binding:"required"`
	Days         int     `json:"days" binding:"required,min=1"`
	CancelFee    float64 `json:"cancel_fee"`
}

type PledgeRecord struct {
	ID       int     `json:"id" db:"id"`
	UID      int     `json:"uid" db:"uid"`
	ConfigID int     `json:"config_id" db:"config_id"`
	Status   int     `json:"status" db:"status"`
	AddTime  string  `json:"addtime" db:"addtime"`
	EndTime  *string `json:"endtime" db:"endtime"`
	Amount   float64 `json:"amount,omitempty"`
	CatName  string  `json:"category_name,omitempty"`
	Discount float64 `json:"discount_rate,omitempty"`
	Days     int     `json:"days,omitempty"`
	Username string  `json:"username,omitempty"`
}

type PledgeListRequest struct {
	Page  int `form:"page" json:"page"`
	Limit int `form:"limit" json:"limit"`
	UID   int `form:"uid" json:"uid"`
}

// ===== 网签系统 =====

type MlsxCompany struct {
	ID      int    `json:"id" db:"id"`
	QYMC    string `json:"qymc" db:"qymc"`
	WQBS    string `json:"wqbs" db:"wqbs"`
	ShiJian string `json:"shijian" db:"shijian"`
}

type MlsxCompanySaveRequest struct {
	ID   int    `json:"id"`
	QYMC string `json:"qymc" binding:"required"`
	WQBS string `json:"wqbs"`
}

type MlsxFile struct {
	ID      int    `json:"id" db:"id"`
	WJID    string `json:"wjid" db:"wjid"`
	Name    string `json:"name" db:"name"`
	IP      string `json:"ip" db:"ip"`
	ShiJian string `json:"shijian" db:"shijian"`
}

// ===== 外部查单 =====

type CheckOrderRequest struct {
	User   string `form:"user" json:"user"`
	OID    string `form:"oid" json:"oid"`
	KCName string `form:"kcname" json:"kcname"`
}

type CheckOrderResult struct {
	OID     int    `json:"oid"`
	PtName  string `json:"ptname"`
	KCName  string `json:"kcname"`
	Status  string `json:"status"`
	Process string `json:"process"`
	Remarks string `json:"remarks"`
	AddTime string `json:"addtime"`
}
