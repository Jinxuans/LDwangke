package model

// ===== 租户（B端商家）=====

type Tenant struct {
	TID       int     `json:"tid" db:"tid"`
	UID       int     `json:"uid" db:"uid"`
	ShopName  string  `json:"shop_name" db:"shop_name"`
	ShopLogo  string  `json:"shop_logo" db:"shop_logo"`
	ShopDesc  string  `json:"shop_desc" db:"shop_desc"`
	Domain    string  `json:"domain" db:"domain"`
	PayConfig *string `json:"pay_config,omitempty" db:"pay_config"`
	Status    int     `json:"status" db:"status"`
	AddTime   string  `json:"addtime" db:"addtime"`
}

type TenantSaveRequest struct {
	UID      int    `json:"uid"` // 平台侧开通时指定
	ShopName string `json:"shop_name" binding:"required"`
	ShopLogo string `json:"shop_logo"`
	ShopDesc string `json:"shop_desc"`
	Domain   string `json:"domain"`
}

type TenantPayConfigSaveRequest struct {
	PayConfig string `json:"pay_config" binding:"required"` // JSON字符串
}

// ===== B端选品 =====

type TenantProduct struct {
	ID          int     `json:"id" db:"id"`
	TID         int     `json:"tid" db:"tid"`
	CID         int     `json:"cid" db:"cid"`
	RetailPrice float64 `json:"retail_price" db:"retail_price"`
	Status      int     `json:"status" db:"status"`
	Sort        int     `json:"sort" db:"sort"`
	// JOIN 填充
	ClassName   string `json:"class_name,omitempty"`
	SupplyPrice string `json:"supply_price,omitempty"` // 供货价（平台价）
	Fenlei      string `json:"fenlei,omitempty"`
}

type TenantProductSaveRequest struct {
	CID         int     `json:"cid" binding:"required"`
	RetailPrice float64 `json:"retail_price" binding:"required"`
	Status      int     `json:"status"`
	Sort        int     `json:"sort"`
}

type TenantProductBatchSaveRequest struct {
	Items []TenantProductSaveRequest `json:"items" binding:"required"`
}

// ===== C端用户 =====

type CUser struct {
	ID       int    `json:"id" db:"id"`
	TID      int    `json:"tid" db:"tid"`
	Phone    string `json:"phone" db:"phone"`
	Account  string `json:"account,omitempty" db:"account"`
	Password string `json:"-" db:"password"`
	Nickname string `json:"nickname" db:"nickname"`
	OpenID   string `json:"openid,omitempty" db:"openid"`
	Token    string `json:"token,omitempty" db:"token"`
	AddTime  string `json:"addtime" db:"addtime"`
}

type CUserLoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CUserLoginResponse struct {
	Token    string `json:"token"`
	ID       int    `json:"id"`
	Nickname string `json:"nickname"`
	Account  string `json:"account"`
}

type CUserSaveRequest struct {
	ID       int    `json:"id"`
	Account  string `json:"account" binding:"required"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

// ===== C端商城 =====

type MallProduct struct {
	CID         int     `json:"cid"`
	Name        string  `json:"name"`
	Noun        string  `json:"noun"`
	RetailPrice float64 `json:"retail_price"`
	Fenlei      string  `json:"fenlei"`
	FenleiName  string  `json:"fenlei_name,omitempty"`
	Sort        int     `json:"sort"`
}

type MallOrderAddRequest struct {
	CID    int            `json:"cid" binding:"required"`
	Data   []OrderAddItem `json:"data" binding:"required"`
	Remark string         `json:"remarks"`
	// C端支付信息
	PayType string `json:"pay_type"` // 支付方式
}

type MallOrderItem struct {
	OID        int    `json:"oid"`
	CID        int    `json:"cid"`
	ClassName  string `json:"class_name"`
	KCName     string `json:"kcname"`
	Account    string `json:"account"`
	Status     string `json:"status"`
	Process    string `json:"process"`
	RetailFees string `json:"retail_fees"`
	AddTime    string `json:"addtime"`
}

// ===== C端商城支付 =====

type MallPayRequest struct {
	CID        int    `json:"cid" binding:"required"`
	Account    string `json:"account" binding:"required"`
	Password   string `json:"password" binding:"required"`
	Remark     string `json:"remark"`
	PayType    string `json:"pay_type" binding:"required"`
	CourseID   string `json:"course_id"`
	CourseName string `json:"course_name"`
	CourseKCJS string `json:"course_kcjs"`
}

type MallPayOrder struct {
	ID         int     `json:"id"`
	OutTradeNo string  `json:"out_trade_no"`
	TradeNo    string  `json:"trade_no"`
	TID        int     `json:"tid"`
	CID        int     `json:"cid"`
	CUID       int     `json:"c_uid"`
	Account    string  `json:"account"`
	Password   string  `json:"-"`
	Remark     string  `json:"remark"`
	PayType    string  `json:"pay_type"`
	Money      float64 `json:"money"`
	Status     int     `json:"status"`
	OrderID    int     `json:"order_id"`
	AddTime    string  `json:"addtime"`
	CourseID   string  `json:"course_id"`
	CourseName string  `json:"course_name"`
	CourseKCJS string  `json:"course_kcjs"`
}

type MallPayCreateResponse struct {
	OutTradeNo string `json:"out_trade_no"`
	PayURL     string `json:"pay_url"`
	Money      string `json:"money"`
}
