package model

type Order struct {
	OID        int    `json:"oid"`
	UID        int    `json:"uid"`
	CID        int    `json:"cid"`
	HID        int    `json:"hid"`
	PTName     string `json:"ptname"`
	School     string `json:"school"`
	Name       string `json:"name"`
	User       string `json:"user"`
	Pass       string `json:"pass"`
	KCName     string `json:"kcname"`
	KCID       string `json:"kcid"`
	Status     string `json:"status"`
	Fees       string `json:"fees"`
	Process    string `json:"process"`
	Remarks    string `json:"remarks"`
	DockStatus string `json:"dockstatus"`
	YID        string `json:"yid"`
	AddTime    string `json:"addtime"`
}

type OrderListRequest struct {
	Page       int    `json:"page" form:"page"`
	Limit      int    `json:"limit" form:"limit"`
	User       string `json:"user" form:"user"`
	Pass       string `json:"pass" form:"pass"`
	School     string `json:"school" form:"school"`
	OID        string `json:"oid" form:"oid"`
	CID        string `json:"cid" form:"cid"`
	KCName     string `json:"kcname" form:"kcname"`
	StatusText string `json:"status_text" form:"status_text"`
	Dock       string `json:"dock" form:"dock"`
	UID        string `json:"uid" form:"uid"`
	HID        string `json:"hid" form:"hid"`
	Search     string `json:"search" form:"search"`
}

type OrderAddRequest struct {
	CID    int            `json:"cid" binding:"required"`
	Data   []OrderAddItem `json:"data" binding:"required"`
	Remark string         `json:"remarks"`
}

type OrderAddItem struct {
	UserInfo string         `json:"userinfo"`
	UserName string         `json:"userName"`
	Data     OrderAddCourse `json:"data"`
}

type OrderAddCourse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	KCJS string `json:"kcjs"`
}

type OrderStatusRequest struct {
	Status string `json:"status"`
	OIDs   []int  `json:"oids"`
	Type   int    `json:"type"`
}

type OrderStats struct {
	Total      int     `json:"total"`
	Processing int     `json:"processing"`
	Completed  int     `json:"completed"`
	Failed     int     `json:"failed"`
	TotalFees  float64 `json:"total_fees"`
}

type OrderAddResult struct {
	SuccessCount int      `json:"success_count"`
	SkippedCount int      `json:"skipped_count"`
	TotalCost    float64  `json:"total_cost"`
	SkippedItems []string `json:"skipped_items,omitempty"`
	OIDs         []int64  `json:"oids,omitempty"`
}
