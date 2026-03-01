package model

type UserManage struct {
	UID       int     `json:"uid"`
	User      string  `json:"user"`
	Name      string  `json:"name"`
	Grade     string  `json:"grade"`
	AddPrice  float64 `json:"addprice"`
	GradeName string  `json:"grade_name"`
	Balance   float64 `json:"balance"`
	YQM       string  `json:"yqm"`
	AddTime   string  `json:"addtime"`
	Status    int     `json:"status"`
}

type UserListRequest struct {
	Page     int    `json:"page" form:"page"`
	Limit    int    `json:"limit" form:"limit"`
	Keywords string `json:"keywords" form:"keywords"`
}

type ClassManage struct {
	CID     int    `json:"cid"`
	Name    string `json:"name"`
	Price   string `json:"price"`
	Content string `json:"content"`
	CateID  string `json:"cateId"`
	Status  int    `json:"status"`
	HID     string `json:"hid"`
	Sort    int    `json:"sort"`
	Noun    string `json:"noun"`
	Yunsuan string `json:"yunsuan"`
}

type ClassEditRequest struct {
	CID     int    `json:"cid"`
	Name    string `json:"name"`
	Price   string `json:"price"`
	Content string `json:"content"`
	CateID  string `json:"cateId"`
	Status  int    `json:"status"`
	HID     string `json:"hid"`
	Sort    int    `json:"sort"`
	Noun    string `json:"noun"`
	Yunsuan string `json:"yunsuan"`
}

type Supplier struct {
	HID     int    `json:"hid"`
	PT      string `json:"pt"`
	Name    string `json:"name"`
	URL     string `json:"url"`
	User    string `json:"user"`
	Pass    string `json:"pass"`
	Token   string `json:"token"`
	Money   string `json:"money"`
	Status  string `json:"status"`
	AddTime string `json:"addtime"`
}

// SupplierFull 包含供应商完整信息（含密钥），仅内部使用
type SupplierFull struct {
	HID    int    `json:"hid"`
	PT     string `json:"pt"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	User   string `json:"user"`
	Pass   string `json:"pass"`
	Token  string `json:"token"`
	IP     string `json:"ip"`
	Cookie string `json:"cookie"`
	Money  string `json:"money"`
	Status string `json:"status"`
}

// ClassFull 课程完整信息（含供应商对接字段）
type ClassFull struct {
	CID     int    `json:"cid"`
	Name    string `json:"name"`
	Noun    string `json:"noun"`
	Price   string `json:"price"`   // DB: varchar
	Docking string `json:"docking"` // DB: varchar
	Fenlei  string `json:"fenlei"`  // DB: varchar
	Status  int    `json:"status"`
	Yunsuan string `json:"yunsuan"`
	Content string `json:"content"`
}

type SystemConfig struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// 公告
type Announcement struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Time       string `json:"time"`
	UID        int    `json:"uid"`
	Status     string `json:"status"`
	Zhiding    string `json:"zhiding"`
	Author     string `json:"author"`
	Visibility int    `json:"visibility" db:"visibility"` // 0: 全体可见, 1: 直属代理可见
}

type AnnouncementListRequest struct {
	Page    int    `json:"page" form:"page"`
	Limit   int    `json:"limit" form:"limit"`
	Keyword string `json:"keyword" form:"keyword"`
}

type AnnouncementSaveRequest struct {
	ID         int    `json:"id"`
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Status     string `json:"status"`
	Zhiding    string `json:"zhiding"`
	Visibility int    `json:"visibility"`
}

// ===== 等级管理 =====
type Grade struct {
	ID     int    `json:"id"`
	Sort   int    `json:"sort"`
	Name   string `json:"name"`
	Rate   string `json:"rate"`
	Money  string `json:"money"`
	AddKF  string `json:"addkf"`
	GJKF   string `json:"gjkf"`
	Status string `json:"status"`
	Time   string `json:"time"`
}

type GradeSaveRequest struct {
	ID     int    `json:"id"`
	Sort   string `json:"sort"`
	Name   string `json:"name" binding:"required"`
	Rate   string `json:"rate"`
	Money  string `json:"money"`
	AddKF  string `json:"addkf"`
	GJKF   string `json:"gjkf"`
	Status string `json:"status"`
}

// ===== 密价设置 =====
type MiJia struct {
	MID       int    `json:"mid"`
	UID       int    `json:"uid"`
	CID       int    `json:"cid"`
	Mode      string `json:"mode"`
	Price     string `json:"price"`
	AddTime   string `json:"addtime"`
	UserName  string `json:"username"`
	ClassName string `json:"classname"`
}

type MiJiaListRequest struct {
	Page    int    `json:"page" form:"page"`
	Limit   int    `json:"limit" form:"limit"`
	UID     int    `json:"uid" form:"uid"`
	CID     int    `json:"cid" form:"cid"`
	Keyword string `json:"keyword" form:"keyword"`
}

type MiJiaSaveRequest struct {
	MID   int    `json:"mid"`
	UID   int    `json:"uid" binding:"required"`
	CID   int    `json:"cid" binding:"required"`
	Mode  string `json:"mode"`
	Price string `json:"price" binding:"required"`
}

type MiJiaBatchRequest struct {
	UID    int    `json:"uid" binding:"required"`
	Fenlei int    `json:"fenlei"`
	Mode   string `json:"mode"`
	Price  string `json:"price" binding:"required"`
}
