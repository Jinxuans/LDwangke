package appui

import (
	"strconv"

	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

// ---------- 配置 ----------

func AppuiConfigGet(c *gin.Context) {
	cfg, err := appuiService.GetConfig()
	if err != nil {
		response.ServerErrorf(c, err, "获取配置失败")
		return
	}
	response.Success(c, cfg)
}

func AppuiConfigSave(c *gin.Context) {
	var cfg AppuiConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := appuiService.SaveConfig(&cfg); err != nil {
		response.ServerErrorf(c, err, "保存失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

// ---------- 价格 ----------

func AppuiGetPrice(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		PID  string `json:"pid" binding:"required"`
		Days int    `json:"days" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	price, err := appuiService.GetPrice(uid, req.PID, req.Days)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, gin.H{"price": price})
}

// ---------- 平台列表 ----------

func AppuiGetCourses(c *gin.Context) {
	cfg, err := appuiService.GetConfig()
	if err != nil {
		response.ServerErrorf(c, err, "获取失败")
		return
	}
	response.Success(c, cfg.Courses)
}

// ---------- 订单列表 ----------

func AppuiOrderList(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	searchType := c.Query("searchType")
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	orders, total, err := appuiService.ListOrders(uid, isAdmin, page, limit, searchType, keyword)
	if err != nil {
		response.ServerErrorf(c, err, "查询失败")
		return
	}
	response.SuccessPage(c, orders, int64(total), page, limit)
}

// ---------- 下单 ----------

func AppuiAddOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		Form map[string]interface{} `json:"form"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Form == nil {
		response.BadRequest(c, "参数不完整")
		return
	}
	msg, err := appuiService.AddOrder(uid, req.Form)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// ---------- 编辑 ----------

func AppuiEditOrder(c *gin.Context) {
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
	if err := appuiService.EditOrder(uid, isAdmin, req.Form); err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, "修改成功")
}

// ---------- 续费 ----------

func AppuiRenewOrder(c *gin.Context) {
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
	msg, err := appuiService.RenewOrder(uid, isAdmin, req.ID, req.Days)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// ---------- 删除/退款 ----------

func AppuiDeleteOrder(c *gin.Context) {
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
	msg, err := appuiService.DeleteOrder(uid, req.ID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}
