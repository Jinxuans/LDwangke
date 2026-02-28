package handler

import (
	"strconv"

	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

var wService = service.NewWService()

// WGetApps 获取项目列表（用户视角）
func WGetApps(c *gin.Context) {
	uid := c.GetInt("uid")
	list, err := wService.GetApps(uid)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, list)
}

// WGetOrders 获取订单列表
func WGetOrders(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "admin" || role == "super"

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filters := map[string]string{
		"account": c.Query("account"),
		"school":  c.Query("school"),
		"status":  c.Query("status"),
		"app_id":  c.Query("app_id"),
	}

	orders, total, err := wService.GetOrders(uid, isAdmin, page, pageSize, filters)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.SuccessPage(c, orders, int64(total), page, pageSize)
}

// WAddOrder 创建订单
func WAddOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		response.BadRequest(c, "请求数据格式错误")
		return
	}
	result, err := wService.AddOrder(uid, data)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, result)
}

// WRefundOrder 退款
func WRefundOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "admin" || role == "super"

	wOrderID, _ := strconv.Atoi(c.Query("w_order_id"))
	if wOrderID <= 0 {
		// 尝试从POST body获取
		var body struct {
			WOrderID int `json:"w_order_id"`
		}
		c.ShouldBindJSON(&body)
		wOrderID = body.WOrderID
	}
	if wOrderID <= 0 {
		response.BadRequest(c, "缺少订单ID")
		return
	}

	result, err := wService.RefundOrder(uid, wOrderID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, result)
}

// WSyncOrder 同步订单
func WSyncOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "admin" || role == "super"

	wOrderID, _ := strconv.Atoi(c.Query("w_order_id"))
	if wOrderID <= 0 {
		response.BadRequest(c, "缺少订单ID")
		return
	}

	msg, err := wService.SyncOrder(uid, wOrderID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// WResumeOrder 重新提交失败订单
func WResumeOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "admin" || role == "super"

	wOrderID, _ := strconv.Atoi(c.Query("w_order_id"))
	if wOrderID <= 0 {
		response.BadRequest(c, "缺少订单ID")
		return
	}

	msg, err := wService.ResumeOrder(uid, wOrderID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// ========== 管理员接口 ==========

// WAdminListApps 管理员获取所有项目
func WAdminListApps(c *gin.Context) {
	list, err := wService.AdminListApps()
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.Success(c, list)
}

// WAdminSaveApp 管理员保存项目
func WAdminSaveApp(c *gin.Context) {
	var app service.WApp
	if err := c.ShouldBindJSON(&app); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	id, err := wService.AdminSaveApp(app)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, map[string]int64{"id": id})
}

// WAdminDeleteApp 管理员删除项目
func WAdminDeleteApp(c *gin.Context) {
	idStr := c.Query("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if id <= 0 {
		response.BadRequest(c, "缺少ID")
		return
	}
	if err := wService.AdminDeleteApp(id); err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, "删除成功")
}
