package handler

import (
	"go-api/internal/model"
	"go-api/internal/response"
	"go-api/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

var openAPIService = service.NewOpenAPIService()

// OpenAPIGetClass 外部API: 获取课程列表
// PHP: act=getclass / act=sxgetclass
func OpenAPIGetClass(c *gin.Context) {
	uid := c.GetInt("uid")
	money := c.GetFloat64("money")

	list, err := openAPIService.GetClassList(uid, money)
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, gin.H{"code": 1, "data": list})
}

// OpenAPIQuery 外部API: 查课
// PHP: act=get / act=sxgetclass + userinfo
func OpenAPIQuery(c *gin.Context) {
	uid := c.GetInt("uid")
	money := c.GetFloat64("money")

	cidStr := c.Query("cid")
	if cidStr == "" {
		cidStr = c.PostForm("cid")
	}
	cid, _ := strconv.Atoi(cidStr)
	if cid <= 0 {
		response.BusinessError(c, 0, "缺少 cid 参数")
		return
	}

	userinfo := c.Query("userinfo")
	if userinfo == "" {
		userinfo = c.PostForm("userinfo")
	}
	if userinfo == "" {
		response.BusinessError(c, 0, "缺少 userinfo 参数")
		return
	}

	result, err := openAPIService.QueryCourse(uid, money, cid, userinfo)
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, result)
}

// OpenAPIAddOrder 外部API: 下单
// PHP: act=sxadd
func OpenAPIAddOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	money := c.GetFloat64("money")

	cidStr := c.Query("cid")
	if cidStr == "" {
		cidStr = c.PostForm("cid")
	}
	cid, _ := strconv.Atoi(cidStr)
	if cid <= 0 {
		response.BusinessError(c, 0, "缺少 cid 参数")
		return
	}

	userinfo := c.Query("userinfo")
	if userinfo == "" {
		userinfo = c.PostForm("userinfo")
	}
	if userinfo == "" {
		response.BusinessError(c, 0, "缺少 userinfo 参数")
		return
	}

	// 构造订单请求
	req := model.OrderAddRequest{
		CID: cid,
		Data: []model.OrderAddItem{
			{UserInfo: userinfo},
		},
	}

	result, err := openAPIService.AddOrder(uid, money, req)
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, result)
}

// OpenAPIOrderList 外部API: 订单列表
// PHP: act=user_orderlist
func OpenAPIOrderList(c *gin.Context) {
	uid := c.GetInt("uid")

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	status := c.Query("status")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	list, total, err := openAPIService.OrderList(uid, page, limit, status)
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// OpenAPIBalance 外部API: 查询余额
func OpenAPIBalance(c *gin.Context) {
	money := c.GetFloat64("money")
	response.Success(c, gin.H{"money": money})
}

// OpenAPIChadan 外部API: 查单（带推送字段，对应 PHP act=chadan）
func OpenAPIChadan(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		username = c.PostForm("username")
	}
	orderID := c.Query("oid")
	if orderID == "" {
		orderID = c.PostForm("oid")
	}
	if orderID == "" {
		orderID = c.Query("id")
		if orderID == "" {
			orderID = c.PostForm("id")
		}
	}

	list, err := openAPIService.Chadan(username, orderID)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, gin.H{"code": 1, "data": list})
}

// OpenAPIBindPushUID 外部API: 绑定微信推送UID（对应 PHP act=bindpushuid）
func OpenAPIBindPushUID(c *gin.Context) {
	orderIDStr := c.PostForm("orderid")
	pushUID := c.PostForm("pushuid")
	orderID, _ := strconv.Atoi(orderIDStr)

	if orderID <= 0 {
		response.BusinessError(c, 0, "参数不全")
		return
	}

	if err := openAPIService.BindPushUID(orderID, pushUID); err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, gin.H{"code": 1, "msg": "操作成功"})
}

// OpenAPIBindPushEmail 外部API: 绑定邮箱推送（对应 PHP act=bindpushemail）
func OpenAPIBindPushEmail(c *gin.Context) {
	orderIDStr := c.PostForm("orderid")
	account := c.PostForm("account")
	pushEmail := c.PostForm("pushEmail")
	orderID, _ := strconv.Atoi(orderIDStr)

	if orderID <= 0 && account == "" {
		response.BusinessError(c, 0, "参数不全")
		return
	}

	if err := openAPIService.BindPushEmail(orderID, account, pushEmail); err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, gin.H{"code": 1, "msg": "操作成功"})
}

// OpenAPIBindShowDocPush 外部API: 绑定ShowDoc推送（对应 PHP act=bindshowdocpush）
func OpenAPIBindShowDocPush(c *gin.Context) {
	orderIDStr := c.PostForm("orderid")
	account := c.PostForm("account")
	showdocURL := c.PostForm("showdoc_url")
	orderID, _ := strconv.Atoi(orderIDStr)

	if orderID <= 0 && account == "" {
		response.BusinessError(c, 0, "参数不全")
		return
	}

	if err := openAPIService.BindShowDocPush(orderID, account, showdocURL); err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, gin.H{"code": 1, "msg": "操作成功"})
}
