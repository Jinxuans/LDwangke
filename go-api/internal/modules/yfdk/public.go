package yfdk

import (
	"fmt"
	"strconv"

	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func YFDKGetPrice(c *gin.Context) {
	var req struct {
		CID string `json:"cid"`
		Day int    `json:"day"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.CID == "" || req.Day < 1 {
		response.BadRequest(c, "参数错误")
		return
	}
	svc := YFDK()
	price, err := svc.GetPrice(req.CID, req.Day)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, gin.H{"price": price, "msg": fmt.Sprintf("预计扣费：%.2f元", price)})
}

func YFDKGetProjects(c *gin.Context) {
	svc := YFDK()
	data, err := svc.GetProjects()
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

func YFDKGetAccountInfo(c *gin.Context) {
	var req struct {
		CID     string `json:"cid"`
		School  string `json:"school"`
		User    string `json:"user"`
		Pass    string `json:"pass"`
		YzmCode string `json:"yzm_code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.CID == "" || req.User == "" || req.Pass == "" {
		response.BadRequest(c, "CID、账号和密码不能为空")
		return
	}
	svc := YFDK()
	data, err := svc.GetAccountInfo(req.CID, req.School, req.User, req.Pass, req.YzmCode)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

func YFDKGetSchools(c *gin.Context) {
	var req struct {
		CID string `json:"cid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.CID == "" {
		response.BadRequest(c, "项目ID不能为空")
		return
	}
	svc := YFDK()
	data, err := svc.GetSchools(req.CID)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

func YFDKSearchSchools(c *gin.Context) {
	var req struct {
		CID     string `json:"cid"`
		Keyword string `json:"keyword"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.CID == "" || req.Keyword == "" {
		response.BadRequest(c, "项目ID和关键词不能为空")
		return
	}
	svc := YFDK()
	data, err := svc.SearchSchools(req.CID, req.Keyword)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

func YFDKOrderList(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	keyword := c.Query("keyword")
	status := c.Query("status")
	cid := c.Query("cid")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	svc := YFDK()
	orders, total, err := svc.ListOrders(uid, isAdmin, page, limit, keyword, status, cid)
	if err != nil {
		response.ServerErrorf(c, err, "查询失败")
		return
	}
	response.SuccessPage(c, orders, int64(total), page, limit)
}

func YFDKAddOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		Form map[string]interface{} `json:"form"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Form == nil {
		response.BadRequest(c, "参数不完整")
		return
	}

	svc := YFDK()
	msg, err := svc.AddOrder(uid, req.Form)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

func YFDKDeleteOrder(c *gin.Context) {
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

	svc := YFDK()
	msg, err := svc.DeleteOrder(uid, req.ID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

func YFDKRenewOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		ID   int `json:"id" binding:"required"`
		Days int `json:"days" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Days < 1 {
		response.BadRequest(c, "参数错误")
		return
	}

	svc := YFDK()
	msg, err := svc.RenewOrder(uid, req.ID, req.Days, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

func YFDKSaveOrder(c *gin.Context) {
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

	svc := YFDK()
	if err := svc.SaveOrder(uid, isAdmin, req.Form); err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, "订单保存成功")
}

func YFDKManualClock(c *gin.Context) {
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

	svc := YFDK()
	if err := svc.ManualClock(uid, req.ID, isAdmin); err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, "打卡任务已提交，请稍后查看日志")
}

func YFDKGetOrderLogs(c *gin.Context) {
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

	svc := YFDK()
	data, err := svc.GetOrderLogs(uid, req.ID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

func YFDKGetOrderDetail(c *gin.Context) {
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

	svc := YFDK()
	data, err := svc.GetOrderDetail(uid, req.ID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}

func YFDKPatchReport(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		ID        int    `json:"id" binding:"required"`
		StartDate string `json:"startDate" binding:"required"`
		EndDate   string `json:"endDate" binding:"required"`
		Type      string `json:"type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数缺失")
		return
	}
	if req.StartDate > req.EndDate {
		response.BadRequest(c, "开始日期不能大于结束日期")
		return
	}

	svc := YFDK()
	msg, err := svc.PatchReport(uid, req.ID, req.StartDate, req.EndDate, req.Type, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

func YFDKCalculatePatchCost(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	var req struct {
		ID        int    `json:"id" binding:"required"`
		StartDate string `json:"startDate" binding:"required"`
		EndDate   string `json:"endDate" binding:"required"`
		Type      string `json:"type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数缺失")
		return
	}

	svc := YFDK()
	data, err := svc.CalculatePatchCost(uid, req.ID, req.StartDate, req.EndDate, req.Type, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, data)
}
