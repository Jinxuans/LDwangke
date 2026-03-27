package tenant

import (
	"strconv"
	"strings"

	"go-api/internal/model"
	ordermodule "go-api/internal/modules/order"
	suppliermodule "go-api/internal/modules/supplier"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

var supplierService = suppliermodule.SharedService()

func resolveMallTenant(c *gin.Context) (*model.Tenant, error) {
	if raw := strings.TrimSpace(c.Param("tid")); raw != "" {
		tid, err := strconv.Atoi(raw)
		if err != nil || tid <= 0 {
			return nil, strconv.ErrSyntax
		}
		return lookupTenantByTID(tid)
	}

	host := strings.TrimSpace(c.GetHeader("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(c.Request.Host)
	}
	return lookupTenantByDomain(host)
}

func resolveMallTID(c *gin.Context) (int, error) {
	t, err := resolveMallTenant(c)
	if err != nil {
		return 0, err
	}
	return t.TID, nil
}

func requireMallCUser(c *gin.Context, tid int) (*model.CUser, bool) {
	token := strings.TrimSpace(c.GetHeader("X-C-Token"))
	if token == "" {
		response.Unauthorized(c, "请先登录会员账号")
		return nil, false
	}
	u, err := lookupCUserByTokenForTenant(token, tid)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return nil, false
	}
	return u, true
}

func resolveMallCUID(c *gin.Context, tid int) int {
	token := strings.TrimSpace(c.GetHeader("X-C-Token"))
	if token == "" {
		return 0
	}
	u, err := lookupCUserByTokenForTenant(token, tid)
	if err != nil {
		return 0
	}
	return u.ID
}

// GET /api/v1/mall/:tid/pay/channels
func MallPayChannels(c *gin.Context) {
	tid, err := resolveMallTID(c)
	if err != nil {
		response.BusinessError(c, 1004, "店铺不存在")
		return
	}
	channels, err := tenantService.GetMallPayChannels(tid)
	if err != nil {
		response.BusinessError(c, 1004, err.Error())
		return
	}
	response.Success(c, channels)
}

// POST /api/v1/mall/:tid/pay
func MallCreatePay(c *gin.Context) {
	tid, err := resolveMallTID(c)
	if err != nil {
		response.BusinessError(c, 1004, "店铺不存在")
		return
	}
	var req model.MallPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整信息")
		return
	}
	// Optional: associate pay order with logged-in C-end user
	cUID := 0
	if token := c.GetHeader("X-C-Token"); token != "" {
		if u, err := lookupCUserByTokenForTenant(token, tid); err == nil {
			cUID = u.ID
		}
	}
	domain := c.Request.Host
	result, err := tenantService.CreateMallPayOrder(tid, req, domain, cUID, strings.TrimSpace(c.Param("tid")) != "")
	if err != nil {
		response.BusinessError(c, 1006, err.Error())
		return
	}
	response.Success(c, result)
}

// POST /api/v1/mall/pay/notify  (易支付异步回调，不带 :tid，通过 out_trade_no 查 tid)
func MallPayNotify(c *gin.Context) {
	params := make(map[string]string)
	for k, v := range c.Request.URL.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	_ = c.Request.ParseForm()
	for k, v := range c.Request.PostForm {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	outTradeNo := params["out_trade_no"]
	if outTradeNo == "" {
		c.String(200, "fail")
		return
	}

	// 从订单号查 tid
	var tid int
	if err := lookupMallPayOrderTID(outTradeNo, &tid); err != nil || tid == 0 {
		c.String(200, "fail")
		return
	}

	if err := tenantService.ConfirmMallPayOrder(tid, params); err != nil {
		c.String(200, "fail")
		return
	}
	c.String(200, "success")
}

// C端鉴权中间件：从 Header X-C-Token 读取
func MallCUserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tid, err := resolveMallTID(c)
		if err != nil {
			response.Unauthorized(c, "店铺不存在")
			c.Abort()
			return
		}
		u, ok := requireMallCUser(c, tid)
		if !ok {
			c.Abort()
			return
		}
		c.Set("c_uid", u.ID)
		c.Set("c_tid", u.TID)
		c.Next()
	}
}

// ===== 公开接口（无需登录）=====

// GET /api/v1/mall/:tid/info
func MallShopInfo(c *gin.Context) {
	t, err := resolveMallTenant(c)
	if err != nil {
		response.BusinessError(c, 1004, "店铺不存在")
		return
	}
	if t.Status != 1 {
		response.BusinessError(c, 1005, "店铺已关闭")
		return
	}
	// 不暴露支付配置
	t.PayConfig = nil
	response.Success(c, gin.H{
		"tid":         t.TID,
		"uid":         t.UID,
		"shop_name":   t.ShopName,
		"shop_logo":   t.ShopLogo,
		"shop_desc":   t.ShopDesc,
		"domain":      t.Domain,
		"status":      t.Status,
		"addtime":     t.AddTime,
		"mall_config": parseTenantMallConfig(t.MallConfig),
	})
}

// GET /api/v1/mall/:tid/products
func MallProductList(c *gin.Context) {
	tid, err := resolveMallTID(c)
	if err != nil {
		response.BusinessError(c, 1004, "店铺不存在")
		return
	}
	list, err := tenantService.MallProducts(tid)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.Success(c, list)
}

// GET /api/v1/mall/:tid/product/:cid
func MallProductDetail(c *gin.Context) {
	tid, err := resolveMallTID(c)
	if err != nil {
		response.BusinessError(c, 1004, "店铺不存在")
		return
	}
	cid, _ := strconv.Atoi(c.Param("cid"))
	p, err := tenantService.MallProductDetail(tid, cid)
	if err != nil {
		response.BusinessError(c, 1003, err.Error())
		return
	}
	response.Success(c, p)
}

// POST /api/v1/mall/:tid/login  （手机号+验证码，暂时直接登录，验证码逻辑可后续接入）
func MallCUserLogin(c *gin.Context) {
	tid, err := resolveMallTID(c)
	if err != nil {
		response.BusinessError(c, 1004, "店铺不存在")
		return
	}
	var req model.CUserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	u, token, err := tenantService.CUserLogin(tid, req.Account, req.Password)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	merged, err := tenantService.MergeGuestOrdersToCUser(tid, u.ID, req.GuestOrders)
	if err != nil {
		response.ServerError(c, "订单合并失败")
		return
	}
	response.Success(c, model.CUserLoginResponse{
		Token:             token,
		ID:                u.ID,
		Nickname:          u.Nickname,
		Account:           u.Account,
		InviteCode:        u.InviteCode,
		MergedGuestOrders: merged,
		MergedCount:       len(merged),
	})
}

func MallCUserRegister(c *gin.Context) {
	tid, err := resolveMallTID(c)
	if err != nil {
		response.BusinessError(c, 1004, "店铺不存在")
		return
	}
	var req model.CUserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	u, token, err := tenantService.RegisterCUser(tid, req)
	if err != nil {
		response.BusinessError(c, 1002, err.Error())
		return
	}
	response.Success(c, model.CUserLoginResponse{
		Token:      token,
		ID:         u.ID,
		Nickname:   u.Nickname,
		Account:    u.Account,
		InviteCode: u.InviteCode,
	})
}

// POST /api/v1/mall/:tid/query  查课（验证商品在该店铺上架后复用A端逻辑）
func MallQueryCourse(c *gin.Context) {
	tid, err := resolveMallTID(c)
	if err != nil {
		response.BusinessError(c, 1004, "店铺不存在或已关闭")
		return
	}

	// 验证店铺
	t, err := lookupTenantByTID(tid)
	if err != nil || t.Status != 1 {
		response.BusinessError(c, 1004, "店铺不存在或已关闭")
		return
	}

	// 预读 body 以获取 cid，再复用 ClassQueryCourse
	var req model.CourseQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写课程ID和下单信息")
		return
	}

	// 验证该课程在店铺中已上架
	if _, err := tenantService.MallProductDetail(tid, req.CID); err != nil {
		response.BusinessError(c, 1003, "商品不存在或未上架")
		return
	}

	// 复用供应商查课逻辑
	result, err := supplierService.QueryCourse(req.CID, req.UserInfo)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

// ===== 需要C端登录的接口 =====

// POST /api/v1/mall/:tid/order
func MallOrderAdd(c *gin.Context) {
	tid, err := resolveMallTID(c)
	if err != nil {
		response.BusinessError(c, 1004, "店铺不存在或已关闭")
		return
	}
	u, ok := requireMallCUser(c, tid)
	if !ok {
		return
	}
	cUID := u.ID

	// 验证店铺
	t, err := lookupTenantByTID(tid)
	if err != nil || t.Status != 1 {
		response.BusinessError(c, 1004, "店铺不存在或已关闭")
		return
	}

	var req model.MallOrderAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 验证商品在该店铺上架
	p, err := tenantService.MallProductDetail(tid, req.CID)
	if err != nil {
		response.BusinessError(c, 1003, "商品不存在")
		return
	}

	// 用 B端 uid 下单（扣供货价）
	orderSvc := ordermodule.NewServices().Command
	addReq := model.OrderAddRequest{
		CID:    req.CID,
		Data:   req.Data,
		Remark: req.Remark,
	}
	result, err := orderSvc.AddForMall(t.UID, tid, cUID, p.RetailPrice, "", addReq)
	if err != nil {
		response.BusinessError(c, 1006, err.Error())
		return
	}
	response.Success(c, result)
}

// GET /api/v1/mall/:tid/orders  查询会员支付订单
func MallOrderList(c *gin.Context) {
	tid, err := resolveMallTID(c)
	if err != nil {
		response.BusinessError(c, 1004, "店铺不存在")
		return
	}
	u, ok := requireMallCUser(c, tid)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	list, total, err := tenantService.CUserPayOrders(tid, u.ID, page, limit)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	domain := c.Request.Host
	usePathTID := strings.TrimSpace(c.Param("tid")) != ""
	for i := range list {
		if list[i].Status == 0 {
			list[i].PayURL = tenantService.buildExistingMallPayURL(model.MallPayOrder{
				TID:        tid,
				OutTradeNo: list[i].OutTradeNo,
				PayType:    list[i].PayType,
				Money:      mustParseMallOrderMoney(list[i].Money),
				Status:     list[i].Status,
			}, domain, usePathTID)
		}
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

func MallCUserProfile(c *gin.Context) {
	tid, err := resolveMallTID(c)
	if err != nil {
		response.BusinessError(c, 1004, "店铺不存在")
		return
	}
	u, ok := requireMallCUser(c, tid)
	if !ok {
		return
	}
	profile, err := tenantService.GetCUserProfile(tid, u.ID)
	if err != nil {
		response.BusinessError(c, 1003, err.Error())
		return
	}
	response.Success(c, profile)
}

func MallPromotionOrders(c *gin.Context) {
	tid, err := resolveMallTID(c)
	if err != nil {
		response.BusinessError(c, 1004, "店铺不存在")
		return
	}
	u, ok := requireMallCUser(c, tid)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	list, total, err := tenantService.CUserPromotionOrders(tid, u.ID, page, limit)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

func MallCUserWithdrawCreate(c *gin.Context) {
	tid, err := resolveMallTID(c)
	if err != nil {
		response.BusinessError(c, 1004, "店铺不存在")
		return
	}
	u, ok := requireMallCUser(c, tid)
	if !ok {
		return
	}
	var req model.WithdrawCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整的提现信息")
		return
	}
	id, err := tenantService.CreateCUserWithdrawRequest(tid, u.ID, req)
	if err != nil {
		response.BusinessError(c, 1003, err.Error())
		return
	}
	response.Success(c, gin.H{"id": id})
}

func MallCUserWithdrawRequests(c *gin.Context) {
	tid, err := resolveMallTID(c)
	if err != nil {
		response.BusinessError(c, 1004, "店铺不存在")
		return
	}
	u, ok := requireMallCUser(c, tid)
	if !ok {
		return
	}
	var req model.WithdrawListRequest
	_ = c.ShouldBindQuery(&req)
	list, total, err := tenantService.CUserWithdrawRequests(tid, u.ID, req)
	if err != nil {
		response.ServerError(c, "查询提现记录失败")
		return
	}
	response.Success(c, gin.H{
		"list":       list,
		"pagination": gin.H{"page": req.Page, "limit": req.Limit, "total": total},
	})
}

// GET /api/v1/mall/:tid/search?keyword=xxx  按下单账号公开查询进行中的课程进度
func MallOrderSearch(c *gin.Context) {
	tid, err := resolveMallTID(c)
	if err != nil {
		response.BusinessError(c, 1004, "店铺不存在或已关闭")
		return
	}
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		response.BusinessError(c, 1001, "请输入下单账号")
		return
	}
	t, err := lookupTenantByTID(tid)
	if err != nil || t.Status != 1 {
		response.BusinessError(c, 1004, "店铺不存在或已关闭")
		return
	}
	list, err := tenantService.SearchMallOrders(tid, keyword)
	if err != nil {
		response.BusinessError(c, 1002, "查询失败: "+err.Error())
		return
	}
	response.Success(c, list)
}

// GET /api/v1/mall/:tid/pay/check?out_trade_no=xxx  C端检测支付结果
func MallCheckPay(c *gin.Context) {
	tid, err := resolveMallTID(c)
	if err != nil {
		response.BusinessError(c, 1004, "店铺不存在")
		return
	}
	outTradeNo := strings.TrimSpace(c.Query("out_trade_no"))
	accessToken := strings.TrimSpace(c.Query("access_token"))
	if outTradeNo == "" {
		response.BadRequest(c, "缺少订单号")
		return
	}

	status, orderID, err := tenantService.CheckMallPayStatus(tid, outTradeNo, accessToken, resolveMallCUID(c, tid))
	if err != nil {
		response.BusinessError(c, 1003, err.Error())
		return
	}
	response.Success(c, gin.H{
		"status":   status, // 0=未支付 1=已支付待下单 2=已下单
		"order_id": orderID,
	})
}

// POST /api/v1/mall/:tid/pay/confirm  C端主动确认已支付并触发下单
func MallConfirmPay(c *gin.Context) {
	tid, err := resolveMallTID(c)
	if err != nil {
		response.BusinessError(c, 1004, "店铺不存在")
		return
	}
	var body struct {
		OutTradeNo  string `json:"out_trade_no"`
		AccessToken string `json:"access_token"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.OutTradeNo == "" {
		response.BadRequest(c, "缺少订单号")
		return
	}
	status, orderID, err := tenantService.UserConfirmMallPay(tid, body.OutTradeNo, strings.TrimSpace(body.AccessToken), resolveMallCUID(c, tid))
	if err != nil {
		response.BusinessError(c, 1003, err.Error())
		return
	}
	response.Success(c, gin.H{
		"status":   status,
		"order_id": orderID,
	})
}

// GET /api/v1/mall/:tid/order/:oid  查询单个订单状态（C端轮询用）
func MallOrderDetail(c *gin.Context) {
	tid, err := resolveMallTID(c)
	if err != nil {
		response.BusinessError(c, 1004, "店铺不存在")
		return
	}
	oid, _ := strconv.ParseInt(c.Param("oid"), 10, 64)

	u, ok := requireMallCUser(c, tid)
	if !ok {
		return
	}

	item, err := tenantService.CUserOrderDetail(tid, u.ID, oid)
	if err != nil {
		response.BusinessError(c, 1003, "订单不存在")
		return
	}
	response.Success(c, item)
}

// GET /api/v1/mall/:tid/guest/order?out_trade_no=xxx&access_token=xxx
func MallGuestOrderDetail(c *gin.Context) {
	tid, err := resolveMallTID(c)
	if err != nil {
		response.BusinessError(c, 1004, "店铺不存在")
		return
	}
	outTradeNo := strings.TrimSpace(c.Query("out_trade_no"))
	accessToken := strings.TrimSpace(c.Query("access_token"))
	if outTradeNo == "" || accessToken == "" {
		response.BadRequest(c, "缺少匿名订单凭证")
		return
	}

	item, err := tenantService.GetGuestMallOrder(tid, outTradeNo, accessToken)
	if err != nil {
		response.BusinessError(c, 1003, err.Error())
		return
	}
	if item.Status == 0 {
		item.PayURL = tenantService.buildExistingMallPayURL(model.MallPayOrder{
			TID:        tid,
			OutTradeNo: item.OutTradeNo,
			PayType:    item.PayType,
			Money:      mustParseMallOrderMoney(item.Money),
			Status:     item.Status,
		}, c.Request.Host, strings.TrimSpace(c.Param("tid")) != "")
	}
	response.Success(c, item)
}

func mustParseMallOrderMoney(raw string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	return v
}
