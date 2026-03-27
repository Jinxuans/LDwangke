package model

// ===== 租户（B端商家）=====

type Tenant struct {
	TID        int     `json:"tid" db:"tid"`
	UID        int     `json:"uid" db:"uid"`
	ShopName   string  `json:"shop_name" db:"shop_name"`
	ShopLogo   string  `json:"shop_logo" db:"shop_logo"`
	ShopDesc   string  `json:"shop_desc" db:"shop_desc"`
	Domain     string  `json:"domain" db:"domain"`
	PayConfig  *string `json:"pay_config,omitempty" db:"pay_config"`
	MallConfig *string `json:"mall_config,omitempty" db:"mall_config"`
	Status     int     `json:"status" db:"status"`
	AddTime    string  `json:"addtime" db:"addtime"`
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

type TenantMallConfig struct {
	RegisterEnabled  bool    `json:"register_enabled"`
	PromotionEnabled bool    `json:"promotion_enabled"`
	CommissionRate   float64 `json:"commission_rate"`
	ShowCategories   bool    `json:"show_categories"`
	PopupNoticeHTML  string  `json:"popup_notice_html,omitempty"`
	CustomerService  TenantCustomerServiceConfig `json:"customer_service"`
}

type TenantCustomerServiceConfig struct {
	Enabled bool   `json:"enabled"`
	Type    string `json:"type,omitempty"`
	Value   string `json:"value,omitempty"`
	Label   string `json:"label,omitempty"`
}

type TenantMallConfigSaveRequest struct {
	MallConfig string `json:"mall_config" binding:"required"` // JSON字符串
}

type TenantMallCategory struct {
	ID      int    `json:"id" db:"id"`
	TID     int    `json:"tid" db:"tid"`
	Name    string `json:"name" db:"name"`
	Sort    int    `json:"sort" db:"sort"`
	Status  int    `json:"status" db:"status"`
	AddTime string `json:"addtime" db:"addtime"`
}

type TenantMallCategorySaveRequest struct {
	ID     int    `json:"id"`
	Name   string `json:"name" binding:"required"`
	Sort   int    `json:"sort"`
	Status int    `json:"status"`
}

// ===== B端选品 =====

type TenantProduct struct {
	ID           int     `json:"id" db:"id"`
	TID          int     `json:"tid" db:"tid"`
	CID          int     `json:"cid" db:"cid"`
	RetailPrice  float64 `json:"retail_price" db:"retail_price"`
	Status       int     `json:"status" db:"status"`
	Sort         int     `json:"sort" db:"sort"`
	DisplayName  string  `json:"display_name,omitempty" db:"display_name"`
	CoverURL     string  `json:"cover_url,omitempty" db:"cover_url"`
	Description  string  `json:"description,omitempty" db:"description"`
	CategoryID   int     `json:"category_id,omitempty" db:"category_id"`
	CategoryName string  `json:"category_name,omitempty"`
	// JOIN 填充
	ClassName   string `json:"class_name,omitempty"`
	SupplyPrice string `json:"supply_price,omitempty"` // 供货价（平台价）
	Fenlei      string `json:"fenlei,omitempty"`
}

type TenantProductSaveRequest struct {
	CID          int     `json:"cid" binding:"required"`
	RetailPrice  float64 `json:"retail_price" binding:"required"`
	Status       int     `json:"status"`
	Sort         int     `json:"sort"`
	DisplayName  string  `json:"display_name"`
	CoverURL     string  `json:"cover_url"`
	Description  string  `json:"description"`
	CategoryID   int     `json:"category_id"`
	CategoryName string  `json:"category_name"`
}

type TenantProductBatchSaveRequest struct {
	Items []TenantProductSaveRequest `json:"items" binding:"required"`
}

// ===== C端用户 =====

type CUser struct {
	ID              int     `json:"id" db:"id"`
	TID             int     `json:"tid" db:"tid"`
	Phone           string  `json:"phone" db:"phone"`
	Account         string  `json:"account,omitempty" db:"account"`
	Password        string  `json:"-" db:"password"`
	Nickname        string  `json:"nickname" db:"nickname"`
	InviteCode      string  `json:"invite_code,omitempty" db:"invite_code"`
	ReferrerID      int     `json:"referrer_id,omitempty" db:"referrer_id"`
	CommissionMoney float64 `json:"commission_money,omitempty" db:"commission_money"`
	CommissionCD    float64 `json:"commission_cdmoney,omitempty" db:"commission_cdmoney"`
	CommissionTotal float64 `json:"commission_total,omitempty" db:"commission_total"`
	Status          int     `json:"status" db:"status"`
	OpenID          string  `json:"openid,omitempty" db:"openid"`
	Token           string  `json:"token,omitempty" db:"token"`
	AddTime         string  `json:"addtime" db:"addtime"`
}

type CUserLoginRequest struct {
	Account     string                 `json:"account" binding:"required"`
	Password    string                 `json:"password" binding:"required"`
	GuestOrders []MallGuestOrderAccess `json:"guest_orders"`
}

type CUserLoginResponse struct {
	Token             string   `json:"token"`
	ID                int      `json:"id"`
	Nickname          string   `json:"nickname"`
	Account           string   `json:"account"`
	InviteCode        string   `json:"invite_code,omitempty"`
	MergedGuestOrders []string `json:"merged_guest_orders,omitempty"`
	MergedCount       int      `json:"merged_count,omitempty"`
}

type CUserRegisterRequest struct {
	Account      string `json:"account" binding:"required"`
	Password     string `json:"password" binding:"required"`
	Nickname     string `json:"nickname"`
	Phone        string `json:"phone"`
	PromoterCode string `json:"promoter_code"`
}

type CUserProfileResponse struct {
	ID                int              `json:"id"`
	TID               int              `json:"tid"`
	Account           string           `json:"account"`
	Nickname          string           `json:"nickname"`
	Phone             string           `json:"phone"`
	InviteCode        string           `json:"invite_code"`
	ReferrerID        int              `json:"referrer_id"`
	ReferrerAccount   string           `json:"referrer_account,omitempty"`
	ReferrerNickname  string           `json:"referrer_nickname,omitempty"`
	CommissionMoney   string           `json:"commission_money"`
	CommissionCDMoney string           `json:"commission_cdmoney"`
	CommissionTotal   string           `json:"commission_total"`
	PromotionOrders   int              `json:"promotion_orders"`
	PromotionEnabled  bool             `json:"promotion_enabled"`
	RegisterEnabled   bool             `json:"register_enabled"`
	CommissionRate    float64          `json:"commission_rate"`
	AddTime           string           `json:"addtime"`
	MallConfigPublic  TenantMallConfig `json:"mall_config"`
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
	Description string  `json:"description,omitempty"`
	CoverURL    string  `json:"cover_url,omitempty"`
	RetailPrice float64 `json:"retail_price"`
	Fenlei      string  `json:"fenlei"`
	CategoryID  int     `json:"category_id,omitempty"`
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
	OutTradeNo string `json:"out_trade_no,omitempty"`
	CID        int    `json:"cid"`
	ClassName  string `json:"class_name"`
	KCName     string `json:"kcname"`
	Account    string `json:"account"`
	Status     string `json:"status"`
	Process    string `json:"process"`
	RetailFees string `json:"retail_fees"`
	AddTime    string `json:"addtime"`
	Remarks    string `json:"remarks,omitempty"`
}

type MallPayOrderItem struct {
	ID          int    `json:"id"`
	OutTradeNo  string `json:"out_trade_no"`
	TradeNo     string `json:"trade_no"`
	CID         int    `json:"cid"`
	ProductName string `json:"product_name"`
	CourseName  string `json:"course_name"`
	School      string `json:"school"`
	Account     string `json:"account"`
	Remark      string `json:"remark"`
	PayType     string `json:"pay_type"`
	PayURL      string `json:"pay_url,omitempty"`
	Status      int    `json:"status"`
	StatusText  string `json:"status_text"`
	Money       string `json:"money"`
	OrderID     int    `json:"order_id"`
	AddTime     string `json:"addtime"`
	PayTime     string `json:"paytime"`
}

type CUserPromotionOrderItem struct {
	ID               int    `json:"id"`
	OutTradeNo       string `json:"out_trade_no"`
	ProductName      string `json:"product_name"`
	CourseName       string `json:"course_name"`
	BuyerAccount     string `json:"buyer_account"`
	Money            string `json:"money"`
	CommissionAmount string `json:"commission_amount"`
	CommissionRate   string `json:"commission_rate"`
	Status           int    `json:"status"`
	StatusText       string `json:"status_text"`
	AddTime          string `json:"addtime"`
	PayTime          string `json:"paytime"`
}

type MallGuestOrderAccess struct {
	OutTradeNo  string `json:"out_trade_no"`
	AccessToken string `json:"access_token"`
}

// ===== C端商城支付 =====

type MallPayRequest struct {
	CID          int              `json:"cid" binding:"required"`
	School       string           `json:"school"`
	Account      string           `json:"account" binding:"required"`
	Password     string           `json:"password" binding:"required"`
	Remark       string           `json:"remark"`
	PayType      string           `json:"pay_type" binding:"required"`
	PromoterCode string           `json:"promoter_code"`
	Courses      []OrderAddCourse `json:"courses"`
	CourseID     string           `json:"course_id"`
	CourseName   string           `json:"course_name"`
	CourseKCJS   string           `json:"course_kcjs"`
}

type MallPayOrder struct {
	ID               int     `json:"id"`
	OutTradeNo       string  `json:"out_trade_no"`
	TradeNo          string  `json:"trade_no"`
	TID              int     `json:"tid"`
	CID              int     `json:"cid"`
	CUID             int     `json:"c_uid"`
	School           string  `json:"school"`
	Account          string  `json:"account"`
	Password         string  `json:"-"`
	Remark           string  `json:"remark"`
	PayType          string  `json:"pay_type"`
	Money            float64 `json:"money"`
	Status           int     `json:"status"`
	OrderID          int     `json:"order_id"`
	AddTime          string  `json:"addtime"`
	CourseID         string  `json:"course_id"`
	CourseName       string  `json:"course_name"`
	CourseKCJS       string  `json:"course_kcjs"`
	CourseItems      string  `json:"course_items"`
	ProductName      string  `json:"product_name,omitempty"`
	OrderStatus      string  `json:"order_status,omitempty"`
	OrderProcess     string  `json:"order_process,omitempty"`
	OrderRemarks     string  `json:"order_remarks,omitempty"`
	OrderCount       int     `json:"order_count,omitempty"`
	PromoterCUID     int     `json:"promoter_c_uid,omitempty"`
	PromoterCode     string  `json:"promoter_code,omitempty"`
	CommissionRate   float64 `json:"commission_rate,omitempty"`
	CommissionAmount float64 `json:"commission_amount,omitempty"`
	CommissionStatus int     `json:"commission_status,omitempty"`
}

type MallPayCreateResponse struct {
	OutTradeNo  string `json:"out_trade_no"`
	PayURL      string `json:"pay_url"`
	Money       string `json:"money"`
	AccessToken string `json:"access_token,omitempty"`
}
