package w

import (
	"encoding/json"
	"strconv"

	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func WGetApps(c *gin.Context) {
	uid := c.GetInt("uid")
	list, err := W().GetApps(uid)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, list)
}

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

	orders, total, err := W().GetOrders(uid, isAdmin, page, pageSize, filters)
	if err != nil {
		response.ServerErrorf(c, err, "查询失败")
		return
	}
	response.SuccessPage(c, orders, int64(total), page, pageSize)
}

func WAddOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		response.BadRequest(c, "请求数据格式错误")
		return
	}
	result, err := W().AddOrder(uid, data)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, result)
}

func WRefundOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "admin" || role == "super"

	wOrderID, _ := strconv.Atoi(c.Query("w_order_id"))
	if wOrderID <= 0 {
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

	result, err := W().RefundOrder(uid, wOrderID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, result)
}

func WSyncOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "admin" || role == "super"

	wOrderID, _ := strconv.Atoi(c.Query("w_order_id"))
	if wOrderID <= 0 {
		response.BadRequest(c, "缺少订单ID")
		return
	}

	msg, err := W().SyncOrder(uid, wOrderID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

func WResumeOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "admin" || role == "super"

	wOrderID, _ := strconv.Atoi(c.Query("w_order_id"))
	if wOrderID <= 0 {
		response.BadRequest(c, "缺少订单ID")
		return
	}

	msg, err := W().ResumeOrder(uid, wOrderID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

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
	if body.Data == nil {
		body.Data = map[string]interface{}{}
	}
	body.Data["login_uid"] = c.GetInt("uid")
	resp, err := W().ProxyAction(body.AppID, body.Act, body.Data)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	var upstream interface{}
	if jsonErr := json.Unmarshal(resp, &upstream); jsonErr != nil {
		response.Success(c, string(resp))
		return
	}
	response.Success(c, upstream)
}

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

	msg, err := W().EditOrder(uid, body.OrderID, body.Form, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

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

	msg, err := W().ChangeRunStatus(uid, body.OrderID, body.Status, body.Form, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

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

	resp, err := W().GetRemainCount(uid, orderID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	c.Data(200, "application/json; charset=utf-8", resp)
}

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

	resp, err := W().GetTaskData(uid, orderID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	c.Data(200, "application/json; charset=utf-8", resp)
}

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

	msg, err := W().EditTask(uid, body.OrderID, body.Form, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

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

	msg, err := W().DelayTask(uid, body.OrderID, body.RunTaskID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

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

	msg, err := W().FastDelayTask(uid, body.OrderID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}
