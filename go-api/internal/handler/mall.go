package handler

import (
	"strconv"
	"strings"

	"go-api/internal/model"
	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

// GET /api/v1/mall/:tid/pay/channels
func MallPayChannels(c *gin.Context) {
	tid, _ := strconv.Atoi(c.Param("tid"))
	channels, err := tenantService.GetMallPayChannels(tid)
	if err != nil {
		response.BusinessError(c, 1004, err.Error())
		return
	}
	response.Success(c, channels)
}

// POST /api/v1/mall/:tid/pay
func MallCreatePay(c *gin.Context) {
	tid, _ := strconv.Atoi(c.Param("tid"))
	var req model.MallPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整信息")
		return
	}
	// Optional: associate pay order with logged-in C-end user
	cUID := 0
	if token := c.GetHeader("X-C-Token"); token != "" {
		if u, err := tenantService.CUserByToken(token); err == nil {
			cUID = u.ID
		}
	}
	domain := c.Request.Host
	result, err := tenantService.CreateMallPayOrder(tid, req, domain, cUID)
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
	if err := service.NewTenantService().GetMallPayOrderTID(outTradeNo, &tid); err != nil || tid == 0 {
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
		token := c.GetHeader("X-C-Token")
		if token == "" {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}
		u, err := tenantService.CUserByToken(token)
		if err != nil {
			response.Unauthorized(c, "登录已过期")
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
	tid, _ := strconv.Atoi(c.Param("tid"))
	t, err := tenantService.GetByTID(tid)
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
	response.Success(c, t)
}

// GET /api/v1/mall/:tid/products
func MallProductList(c *gin.Context) {
	tid, _ := strconv.Atoi(c.Param("tid"))
	list, err := tenantService.MallProducts(tid)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.Success(c, list)
}

// GET /api/v1/mall/:tid/product/:cid
func MallProductDetail(c *gin.Context) {
	tid, _ := strconv.Atoi(c.Param("tid"))
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
	tid, _ := strconv.Atoi(c.Param("tid"))
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
	response.Success(c, model.CUserLoginResponse{
		Token:    token,
		ID:       u.ID,
		Nickname: u.Nickname,
		Account:  u.Account,
	})
}

// POST /api/v1/mall/:tid/query  查课（验证商品在该店铺上架后复用A端逻辑）
func MallQueryCourse(c *gin.Context) {
	tid, _ := strconv.Atoi(c.Param("tid"))

	// 验证店铺
	t, err := tenantService.GetByTID(tid)
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
	tid, _ := strconv.Atoi(c.Param("tid"))
	cUID := 0

	// 验证店铺
	t, err := tenantService.GetByTID(tid)
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
	orderSvc := service.NewOrderService()
	addReq := model.OrderAddRequest{
		CID:    req.CID,
		Data:   req.Data,
		Remark: req.Remark,
	}
	result, err := orderSvc.AddOrdersForMall(t.UID, tid, cUID, p.RetailPrice, addReq)
	if err != nil {
		response.BusinessError(c, 1006, err.Error())
		return
	}
	response.Success(c, result)
}

// GET /api/v1/mall/:tid/orders
func MallOrderList(c *gin.Context) {
	tid, _ := strconv.Atoi(c.Param("tid"))
	cUID := 0
	if token := c.GetHeader("X-C-Token"); token != "" {
		if u, err := tenantService.CUserByToken(token); err == nil {
			cUID = u.ID
		}
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	list, total, err := tenantService.CUserOrders(tid, cUID, page, limit)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

// GET /api/v1/mall/:tid/search?keyword=xxx  按账号或订单号公开查询订单
func MallOrderSearch(c *gin.Context) {
	tid, _ := strconv.Atoi(c.Param("tid"))
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		response.BusinessError(c, 1001, "请输入查询关键词")
		return
	}
	t, err := tenantService.GetByTID(tid)
	if err != nil || t.Status != 1 {
		response.BusinessError(c, 1004, "店铺不存在或已关闭")
		return
	}
	list, err := tenantService.SearchMallOrders(tid, keyword)
	if err != nil {
		response.BusinessError(c, 1002, "查询失败")
		return
	}
	response.Success(c, list)
}

// GET /api/v1/mall/:tid/pay/check?out_trade_no=xxx  C端检测支付结果
func MallCheckPay(c *gin.Context) {
	tid, _ := strconv.Atoi(c.Param("tid"))
	outTradeNo := strings.TrimSpace(c.Query("out_trade_no"))
	if outTradeNo == "" {
		response.BadRequest(c, "缺少订单号")
		return
	}

	status, orderID, err := tenantService.CheckMallPayStatus(tid, outTradeNo)
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
	tid, _ := strconv.Atoi(c.Param("tid"))
	var body struct {
		OutTradeNo string `json:"out_trade_no"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.OutTradeNo == "" {
		response.BadRequest(c, "缺少订单号")
		return
	}
	status, orderID, err := tenantService.UserConfirmMallPay(tid, body.OutTradeNo)
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
	tid, _ := strconv.Atoi(c.Param("tid"))
	oid, _ := strconv.ParseInt(c.Param("oid"), 10, 64)

	cUID := 0
	if token := c.GetHeader("X-C-Token"); token != "" {
		if u, err := tenantService.CUserByToken(token); err == nil {
			cUID = u.ID
		}
	}
	if cUID == 0 {
		response.BusinessError(c, 1401, "请先登录")
		return
	}

	item, err := tenantService.CUserOrderDetail(tid, cUID, oid)
	if err != nil {
		response.BusinessError(c, 1003, "订单不存在")
		return
	}
	response.Success(c, item)
}
