package handler

import (
	"strconv"

	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

// ---------- 小米运动 项目列表 ----------

func XMGetProjects(c *gin.Context) {
	uid := c.GetInt("uid")
	svc := service.NewXMService()
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

	svc := service.NewXMService()
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

	svc := service.NewXMService()
	orders, total, err := svc.GetOrders(uid, isAdmin, page, pageSize, filters)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.SuccessPage(c, orders, int64(total), page, pageSize)
}

// ---------- 查询跑步状态 ----------

func XMQueryRun(c *gin.Context) {
	uid := c.GetInt("uid")

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		response.BadRequest(c, "请求数据格式错误")
		return
	}

	svc := service.NewXMService()
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

	svc := service.NewXMService()
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

	svc := service.NewXMService()
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

	svc := service.NewXMService()
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

	svc := service.NewXMService()
	data, err := svc.GetOrderLogs(uid, orderID, isAdmin, page, pageSize)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

// ========== 管理员项目管理 ==========

func XMAdminListProjects(c *gin.Context) {
	svc := service.NewXMService()
	list, err := svc.AdminListProjects()
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.Success(c, list)
}

func XMAdminSaveProject(c *gin.Context) {
	var p service.XMProjectAdmin
	if err := c.ShouldBindJSON(&p); err != nil {
		response.BadRequest(c, "请求数据格式错误")
		return
	}
	svc := service.NewXMService()
	id, err := svc.AdminSaveProject(p)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, map[string]int{"id": id})
}

func XMAdminDeleteProject(c *gin.Context) {
	id, _ := strconv.Atoi(c.Query("id"))
	if id <= 0 {
		response.BadRequest(c, "缺少项目ID")
		return
	}
	svc := service.NewXMService()
	if err := svc.AdminDeleteProject(id); err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, "删除成功")
}
