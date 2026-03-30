package sxdk

import (
	"strconv"

	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func SXDKGetPrice(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		Platform string `json:"platform" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	svc := SXDK()
	price, err := svc.GetPrice(uid, req.Platform)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, gin.H{"price": price})
}

func SXDKOrderList(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	searchField := c.Query("searchField")
	searchValue := c.Query("searchValue")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	svc := SXDK()
	orders, total, err := svc.ListOrders(uid, isAdmin, page, pageSize, searchField, searchValue)
	if err != nil {
		response.ServerErrorf(c, err, "查询失败")
		return
	}
	response.SuccessPage(c, orders, int64(total), page, pageSize)
}

func SXDKAddOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		Form map[string]interface{} `json:"form"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Form == nil {
		response.BadRequest(c, "参数不完整")
		return
	}

	svc := SXDK()
	msg, err := svc.AddOrder(uid, req.Form)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

func SXDKDeleteOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		ID             int  `json:"id" binding:"required"`
		DelReturnMoney bool `json:"delReturnMoney"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	svc := SXDK()
	msg, err := svc.DeleteOrder(uid, req.ID, isAdmin, req.DelReturnMoney)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

func SXDKEditOrder(c *gin.Context) {
	uid := c.GetInt("uid")

	var req struct {
		Form map[string]interface{} `json:"form"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Form == nil {
		response.BadRequest(c, "参数不完整")
		return
	}

	svc := SXDK()
	msg, err := svc.EditOrder(uid, req.Form)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

func SXDKSearchPhoneInfo(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		Form map[string]interface{} `json:"form"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Form == nil {
		response.BadRequest(c, "参数不完整")
		return
	}

	svc := SXDK()
	data, err := svc.SearchPhoneInfo(uid, req.Form)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

func SXDKGetLog(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		ID int `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	svc := SXDK()
	data, err := svc.ProxyAction(uid, req.ID, isAdmin, "getLog", nil)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

func SXDKNowCheck(c *gin.Context) {
	uid := c.GetInt("uid")

	var req struct {
		ID       int    `json:"id" binding:"required"`
		Platform string `json:"platform" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	svc := SXDK()
	data, err := svc.NowCheck(uid, req.ID, req.Platform)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

func SXDKChangeCheckCode(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		ID   int `json:"id" binding:"required"`
		Code int `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	svc := SXDK()
	data, err := svc.ChangeCheckCode(uid, req.ID, req.Code, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

func SXDKChangeHolidayCode(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		ID   int                    `json:"id" binding:"required"`
		Form map[string]interface{} `json:"form"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	svc := SXDK()
	data, err := svc.ProxyAction(uid, req.ID, isAdmin, "setHolidayCode", req.Form)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

func SXDKGetWxPush(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		ID int `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	svc := SXDK()
	data, err := svc.ProxyAction(uid, req.ID, isAdmin, "getWxPush", nil)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

func SXDKQuerySourceOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		Form map[string]interface{} `json:"form"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Form == nil {
		response.BadRequest(c, "参数不完整")
		return
	}

	idVal, ok := req.Form["id"]
	if !ok {
		response.BadRequest(c, "缺少订单ID")
		return
	}
	id := 0
	switch v := idVal.(type) {
	case float64:
		id = int(v)
	case string:
		id, _ = strconv.Atoi(v)
	}

	svc := SXDK()
	data, err := svc.ProxyAction(uid, id, isAdmin, "selectOrderById", req.Form)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

func SXDKSyncOrders(c *gin.Context) {
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"
	if !isAdmin {
		uid := c.GetInt("uid")
		if uid != 1 {
			response.BusinessError(c, -1, "权限不足")
			return
		}
	}

	svc := SXDK()
	msg, err := svc.SyncOrders()
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

func SXDKGetUserrow(c *gin.Context) {
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"
	if !isAdmin {
		uid := c.GetInt("uid")
		if uid != 1 {
			response.BusinessError(c, -1, "权限不足")
			return
		}
	}

	svc := SXDK()
	data, err := svc.GetUserrow()
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

func SXDKXxyGetSchoolList(c *gin.Context) {
	svc := SXDK()
	cfg, err := svc.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		response.BusinessError(c, -1, "泰山打卡未配置")
		return
	}

	data, err := svc.ProxyAction(0, 0, true, "xxyGetSchoolList", nil)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

func SXDKXxyAddressSearch(c *gin.Context) {
	var req struct {
		Form map[string]interface{} `json:"form"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Form == nil {
		response.BadRequest(c, "参数不完整")
		return
	}

	svc := SXDK()
	cfg, err := svc.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		response.BusinessError(c, -1, "泰山打卡未配置")
		return
	}

	result, err := svc.ProxyAction(0, 0, true, "xxyAddressSearchPoi", req.Form)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, result)
}

func SXDKXxtGetSchoolList(c *gin.Context) {
	var req struct {
		Filter string `json:"filter"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	svc := SXDK()
	cfg, err := svc.GetConfig()
	if err != nil || cfg.BaseURL == "" {
		response.BusinessError(c, -1, "泰山打卡未配置")
		return
	}

	result, err := svc.ProxyAction(0, 0, true, "xxtGetSchoolList", map[string]interface{}{"filter": req.Filter})
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, result)
}

func SXDKGetAsyncTask(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		ID int `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	svc := SXDK()
	data, err := svc.ProxyAction(uid, req.ID, isAdmin, "getAsyncTask", nil)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}
