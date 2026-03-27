package model

type User struct {
	UID    int     `json:"uid" db:"uid"`
	UUID   int     `json:"uuid" db:"uuid"`
	User   string  `json:"user" db:"user"`
	Pass   string  `json:"-" db:"pass"`
	Pass2  string  `json:"-" db:"pass2"`
	Name   string  `json:"name" db:"name"`
	Money  float64 `json:"money" db:"money"`
	Grade  string  `json:"grade" db:"grade"`
	Active string  `json:"active" db:"active"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Pass2    string `json:"pass2"` // 管理员二次验证密码
}

type RegisterRequest struct {
	Username   string `json:"user" binding:"required"`
	Password   string `json:"pass" binding:"required"`
	Nickname   string `json:"name" binding:"required"`
	Invite     string `json:"yqm"`
	Email      string `json:"email"`
	VerifyCode string `json:"verify_code"`
}

type SendCodeRequest struct {
	Email   string `json:"email" binding:"required,email"`
	Purpose string `json:"purpose" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Code     string `json:"code" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ChangeEmailCodeRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
}

type ChangeEmailRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
	Code     string `json:"code" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	User         *User  `json:"user"`
}

// Vben Admin 期望的登录响应格式
type VbenLoginResponse struct {
	AccessToken string   `json:"accessToken"`
	UserId      string   `json:"userId"`
	Username    string   `json:"username"`
	RealName    string   `json:"realName"`
	Avatar      string   `json:"avatar"`
	Desc        string   `json:"desc"`
	HomePath    string   `json:"homePath"`
	Roles       []string `json:"roles"`
}

// Vben Admin 期望的用户信息格式
type VbenUserInfo struct {
	UserId   string   `json:"userId"`
	Username string   `json:"username"`
	RealName string   `json:"realName"`
	Avatar   string   `json:"avatar"`
	Desc     string   `json:"desc"`
	HomePath string   `json:"homePath"`
	Roles    []string `json:"roles"`
	Token    string   `json:"token,omitempty"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// 用户中心 (按 PHP info case)
type UserProfile struct {
	UID             int         `json:"uid"`
	User            string      `json:"user"`
	Name            string      `json:"name"`
	Money           float64     `json:"money"`
	CDMoney         float64     `json:"cdmoney"`
	MallMoney       float64     `json:"mall_money"`
	MallCDMoney     float64     `json:"mall_cdmoney"`
	Grade           string      `json:"grade"`
	GradeID         int         `json:"grade_id"`
	AddPrice        float64     `json:"addprice"`
	GradeName       string      `json:"grade_name"`
	InviteGradeID   int         `json:"invite_grade_id"`
	InviteGradeName string      `json:"invite_grade_name"`
	InviteAddPrice  float64     `json:"invite_addprice"`
	KHCZ            int         `json:"khcz"`
	Key             string      `json:"key"`
	YQM             string      `json:"yqm"`
	Email           string      `json:"email"`
	Phone           string      `json:"phone"`
	PushToken       string      `json:"push_token"`
	ZCZ             float64     `json:"zcz"`
	OrderTotal      int         `json:"order_total"`
	TodayOrders     int         `json:"today_orders"`
	TodaySpend      float64     `json:"today_spend"`
	Notice          string      `json:"notice"`
	SJUser          string      `json:"sjuser"`
	SJNotice        string      `json:"sjnotice"`
	AgentStats      *AgentStats `json:"dailitongji,omitempty"`
}

// 代理统计 (按 PHP dailitongji)
type AgentStats struct {
	DLZS int `json:"dlzs"` // 代理总数
	DLDL int `json:"dldl"` // 今日活跃代理
	DLZC int `json:"dlzc"` // 今日新注册代理
	JRJD int `json:"jrjd"` // 今日交单数
}

type ChangePasswordRequest struct {
	OldPass string `json:"oldpass" binding:"required"`
	NewPass string `json:"newpass" binding:"required"`
}

// 支付
type PayOrder struct {
	OID        int    `json:"oid"`
	OutTradeNo string `json:"out_trade_no"`
	UID        int    `json:"uid"`
	Money      string `json:"money"`
	Status     int    `json:"status"`
	AddTime    string `json:"addtime"`
}

type PayRequest struct {
	Money float64 `json:"money" binding:"required"`
	Type  string  `json:"type"`
}

type PayChannel struct {
	Key   string `json:"type"`
	Label string `json:"name"`
}

type PayCreateResponse struct {
	OID        int    `json:"oid"`
	OutTradeNo string `json:"out_trade_no"`
	Money      string `json:"money"`
	PayURL     string `json:"pay_url"`
}

// 工单
type Ticket struct {
	ID               int    `json:"id"`
	UID              int    `json:"uid"`
	OID              int    `json:"oid"`
	Type             string `json:"type"`
	Content          string `json:"content"`
	Reply            string `json:"reply"`
	Status           int    `json:"status"` // 1=待回复 2=已回复 3=已关闭
	AddTime          string `json:"addtime"`
	ReplyTime        string `json:"reply_time"`
	SupplierReportID int    `json:"supplier_report_id"`
	SupplierStatus   int    `json:"supplier_status"` // -1=未提交 0=待处理 1=处理完成 3=暂时搁置 4=处理中 6=已退款
	SupplierAnswer   string `json:"supplier_answer"`
	// 关联订单信息（查询时JOIN填充）
	OrderUser            string `json:"order_user,omitempty"`
	OrderPT              string `json:"order_pt,omitempty"`
	OrderStatus          string `json:"order_status,omitempty"`
	OrderYID             string `json:"order_yid,omitempty"`
	SupplierReportSwitch int    `json:"supplier_report_switch"`     // 分类上游反馈开关
	SupplierReportHID    int    `json:"supplier_report_hid_switch"` // 分类配置的反馈供应商HID
}

type TicketCreateRequest struct {
	OID     int    `json:"oid"`
	Type    string `json:"type"`
	Content string `json:"content" binding:"required"`
}

type TicketReplyRequest struct {
	ID    int    `json:"id" binding:"required"`
	Reply string `json:"reply" binding:"required"`
}

// 余额流水
type MoneyLog struct {
	ID      int     `json:"id"`
	UID     int     `json:"uid"`
	Type    string  `json:"type"`
	Money   float64 `json:"money"`
	Balance float64 `json:"balance"`
	Remark  string  `json:"remark"`
	AddTime string  `json:"addtime"`
}

type MoneyLogListRequest struct {
	Page  int    `json:"page" form:"page"`
	Limit int    `json:"limit" form:"limit"`
	Type  string `json:"type" form:"type"`
}
