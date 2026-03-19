package model

type Class struct {
	CID     int    `json:"cid" db:"cid"`
	Name    string `json:"name" db:"name"`
	Noun    string `json:"noun" db:"noun"`
	Price   string `json:"price" db:"price"`     // DB: varchar
	Docking string `json:"docking" db:"docking"` // DB: varchar
	Fenlei  string `json:"fenlei" db:"fenlei"`   // DB: varchar
	Status  int    `json:"status" db:"status"`
	Sort    int    `json:"sort" db:"sort"`
	Content string `json:"content,omitempty"`
}

type Category struct {
	ID                int    `json:"id" db:"id"`
	Name              string `json:"name" db:"name"`
	Sort              int    `json:"sort" db:"sort"`
	Status            string `json:"status" db:"status"`
	Time              string `json:"time" db:"time"`
	Recommend         int    `json:"recommend" db:"recommend"`
	Log               int    `json:"log" db:"log"`
	Ticket            int    `json:"ticket" db:"ticket"`
	ChangePass        int    `json:"changepass" db:"changepass"`
	AllowPause        int    `json:"allowpause" db:"allowpause"`
	SupplierReport    int    `json:"supplier_report" db:"supplier_report"`
	SupplierReportHID int    `json:"supplier_report_hid" db:"supplier_report_hid"`
}

type CategoryListRequest struct {
	Page    int    `form:"page"`
	Limit   int    `form:"limit"`
	Keyword string `form:"keyword"`
	Status  string `form:"status"`
}

type ClassListRequest struct {
	Fenlei   int    `form:"fenlei"`
	Status   int    `form:"status"`
	Search   string `form:"search"`
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	Favorite int    `form:"favorite"`
}

// ===== 查课/下单相关 =====

type CourseQueryRequest struct {
	CID      int    `json:"cid" binding:"required"`
	UserInfo string `json:"userinfo" binding:"required"`
}

type CourseItem struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	KCJS           string `json:"kcjs,omitempty"`
	StudyStartTime string `json:"studyStartTime,omitempty"`
	StudyEndTime   string `json:"studyEndTime,omitempty"`
	ExamStartTime  string `json:"examStartTime,omitempty"`
	ExamEndTime    string `json:"examEndTime,omitempty"`
	Complete       string `json:"complete,omitempty"`
}

type CourseQueryResponse struct {
	UserInfo string       `json:"userinfo"`
	UserName string       `json:"userName"`
	Msg      string       `json:"msg"`
	Data     []CourseItem `json:"data"`
}

// SupplierQueryResult 上游供应商查课响应
type SupplierQueryResult struct {
	Code     int          `json:"code"`
	Msg      string       `json:"msg"`
	UserName string       `json:"userName"`
	Data     []CourseItem `json:"data"`
}

// SupplierOrderResult 上游供应商下单响应
type SupplierOrderResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	YID  string `json:"id,omitempty"`
}

// SupplierProgressItem 上游供应商进度查询响应项 (按 PHP jdjk.php processCx 返回的 data 数组元素)
type SupplierProgressItem struct {
	YID             string `json:"id"`
	Noun            string `json:"noun"`
	KCName          string `json:"kcname"`
	User            string `json:"user"`
	Status          string `json:"status"`
	StatusText      string `json:"status_text"`
	Process         string `json:"process"`
	Remarks         string `json:"remarks"`
	CourseStartTime string `json:"courseStartTime"`
	CourseEndTime   string `json:"courseEndTime"`
	ExamStartTime   string `json:"examStartTime"`
	ExamEndTime     string `json:"examEndTime"`
}

// SupplierBatchProgressRef 为供应商级批量进度拉取提供最小订单上下文。
type SupplierBatchProgressRef struct {
	YID    string `json:"yid"`
	User   string `json:"user"`
	KCName string `json:"kcname"`
	KCID   string `json:"kcid"`
	Noun   string `json:"noun"`
}
