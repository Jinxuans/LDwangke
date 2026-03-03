package handler

import (
	"encoding/json"
	"strconv"

	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

var wService = service.NewWService()

// WEnsureTable 启动时建表
func WEnsureTable() {
	wService.EnsureTable()
}

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

// ========== 代理/Jingyu 接口 ==========

// WProxyAction 通用代理转发
func WProxyAction(c *gin.Context) {
	var body struct {
		AppID int64                  `json:"app_id"`
		Act   string                 `json:"act"`
		Data  map[string]interface{} `json:"data"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if body.AppID <= 0 || body.Act == "" {
		response.BadRequest(c, "缺少 app_id 或 act")
		return
	}
	// 注入当前用户uid，供 get_price 等本地计算使用
	if body.Data == nil {
		body.Data = map[string]interface{}{}
	}
	body.Data["login_uid"] = c.GetInt("uid")
	resp, err := wService.ProxyAction(body.AppID, body.Act, body.Data)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	// 将上游原始响应包装成标准格式 {code:0, data: <upstream_response>}
	// 这样前端 requestClient 拦截器能正确解析
	var upstream interface{}
	if jsonErr := json.Unmarshal(resp, &upstream); jsonErr != nil {
		response.Success(c, string(resp))
		return
	}
	response.Success(c, upstream)
}

// WEditOrder 编辑订单
func WEditOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "admin" || role == "super"

	var body struct {
		OrderID int                    `json:"order_id"`
		Form    map[string]interface{} `json:"form"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if body.OrderID <= 0 {
		response.BadRequest(c, "缺少订单ID")
		return
	}

	msg, err := wService.EditOrder(uid, body.OrderID, body.Form, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// WChangeRunStatus 修改运行状态 (暂停/启动)
func WChangeRunStatus(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "admin" || role == "super"

	var body struct {
		OrderID int                    `json:"order_id"`
		Status  int                    `json:"status"`
		Form    map[string]interface{} `json:"form"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if body.OrderID <= 0 {
		response.BadRequest(c, "缺少订单ID")
		return
	}

	msg, err := wService.ChangeRunStatus(uid, body.OrderID, body.Status, body.Form, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// WGetRemainCount 获取剩余次数
func WGetRemainCount(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "admin" || role == "super"

	orderID, _ := strconv.Atoi(c.Query("order_id"))
	if orderID <= 0 {
		var body struct {
			OrderID int `json:"order_id"`
		}
		c.ShouldBindJSON(&body)
		orderID = body.OrderID
	}
	if orderID <= 0 {
		response.BadRequest(c, "缺少订单ID")
		return
	}

	resp, err := wService.GetRemainCount(uid, orderID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	c.Data(200, "application/json; charset=utf-8", resp)
}

// WGetTaskData 获取任务数据
func WGetTaskData(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "admin" || role == "super"

	orderID, _ := strconv.Atoi(c.Query("order_id"))
	if orderID <= 0 {
		var body struct {
			OrderID int `json:"order_id"`
		}
		c.ShouldBindJSON(&body)
		orderID = body.OrderID
	}
	if orderID <= 0 {
		response.BadRequest(c, "缺少订单ID")
		return
	}

	resp, err := wService.GetTaskData(uid, orderID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	c.Data(200, "application/json; charset=utf-8", resp)
}

// WEditTask 编辑任务
func WEditTask(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "admin" || role == "super"

	var body struct {
		OrderID int                    `json:"order_id"`
		Form    map[string]interface{} `json:"form"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if body.OrderID <= 0 {
		response.BadRequest(c, "缺少订单ID")
		return
	}

	msg, err := wService.EditTask(uid, body.OrderID, body.Form, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// WDelayTask 延时任务
func WDelayTask(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "admin" || role == "super"

	var body struct {
		OrderID   int    `json:"order_id"`
		RunTaskID string `json:"run_task_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if body.OrderID <= 0 {
		response.BadRequest(c, "缺少订单ID")
		return
	}

	msg, err := wService.DelayTask(uid, body.OrderID, body.RunTaskID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// WFastDelayTask 快速延时
func WFastDelayTask(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "admin" || role == "super"

	var body struct {
		OrderID int `json:"order_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if body.OrderID <= 0 {
		response.BadRequest(c, "缺少订单ID")
		return
	}

	msg, err := wService.FastDelayTask(uid, body.OrderID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}
