package xm

import (
	"strconv"

	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

// ---------- 小米运动 项目列表 ----------

func XMGetProjects(c *gin.Context) {
	uid := c.GetInt("uid")
	svc := XM()
	projects, err := svc.GetProjects(uid)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, projects)
}

// ---------- 创建订单 ----------

func XMAddOrder(c *gin.Context) {
	uid := c.GetInt("uid")

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		response.BadRequest(c, "请求数据格式错误")
		return
	}

	svc := XM()
	result, err := svc.AddOrder(uid, data)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, result)
}

// ---------- 订单列表 ----------

func XMGetOrders(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"
	if uid == 1 {
		isAdmin = true
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	filters := map[string]string{
		"account":  c.Query("account"),
		"school":   c.Query("school"),
		"status":   c.Query("status"),
		"project":  c.Query("project"),
		"order_id": c.Query("order_id"),
	}

	svc := XM()
	orders, total, err := svc.GetOrders(uid, isAdmin, page, pageSize, filters)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.SuccessPage(c, orders, int64(total), page, pageSize)
}

// ---------- 增加次数 ----------

func XMAddOrderKM(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"
	if uid == 1 {
		isAdmin = true
	}

	var body struct {
		OrderID int `json:"order_id"`
		AddKM   int `json:"add_km"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "请求数据格式错误")
		return
	}
	if body.OrderID <= 0 || body.AddKM <= 0 {
		response.BadRequest(c, "参数错误")
		return
	}

	svc := XM()
	result, err := svc.AddOrderKM(uid, body.OrderID, body.AddKM, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, result)
}

// ---------- 查询跑步状态 ----------

func XMQueryRun(c *gin.Context) {
	uid := c.GetInt("uid")

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		response.BadRequest(c, "请求数据格式错误")
		return
	}

	svc := XM()
	result, err := svc.QueryRun(uid, data)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, result)
}

// ---------- 退款 ----------

func XMRefundOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"
	if uid == 1 {
		isAdmin = true
	}

	orderID, _ := strconv.Atoi(c.Query("order_id"))
	if orderID <= 0 {
		response.BadRequest(c, "缺少订单ID")
		return
	}

	svc := XM()
	data, err := svc.RefundOrder(uid, orderID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

// ---------- 删除订单 ----------

func XMDeleteOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"
	if uid == 1 {
		isAdmin = true
	}

	orderID, _ := strconv.Atoi(c.Query("order_id"))
	if orderID <= 0 {
		response.BadRequest(c, "缺少订单ID")
		return
	}

	svc := XM()
	msg, err := svc.DeleteOrder(uid, orderID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// ---------- 同步订单 ----------

func XMSyncOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"
	if uid == 1 {
		isAdmin = true
	}

	orderID, _ := strconv.Atoi(c.Query("order_id"))
	if orderID <= 0 {
		response.BadRequest(c, "缺少订单ID")
		return
	}

	svc := XM()
	msg, err := svc.SyncOrder(uid, orderID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// ---------- 订单日志 ----------

func XMGetOrderLogs(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"
	if uid == 1 {
		isAdmin = true
	}

	orderID, _ := strconv.Atoi(c.Query("order_id"))
	if orderID <= 0 {
		response.BadRequest(c, "缺少订单ID")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	svc := XM()
	data, err := svc.GetOrderLogs(uid, orderID, isAdmin, page, pageSize)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}
