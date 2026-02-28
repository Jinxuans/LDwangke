package handler

import (
	"strconv"

	"go-api/internal/response"
	"go-api/internal/service"

	"github.com/gin-gonic/gin"
)

// ---------- 配置 ----------

func SDXYConfigGet(c *gin.Context) {
	svc := service.NewSDXYService()
	cfg, err := svc.GetConfig()
	if err != nil {
		response.ServerError(c, "获取配置失败")
		return
	}
	response.Success(c, cfg)
}

func SDXYConfigSave(c *gin.Context) {
	var cfg service.SDXYConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	svc := service.NewSDXYService()
	if err := svc.SaveConfig(&cfg); err != nil {
		response.ServerError(c, "保存失败")
		return
	}
	response.SuccessMsg(c, "保存成功")
}

// ---------- 价格 ----------

func SDXYGetPrice(c *gin.Context) {
	uid := c.GetInt("uid")
	svc := service.NewSDXYService()
	price, err := svc.GetPrice(uid)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.Success(c, gin.H{"price": price})
}

// ---------- 订单列表 ----------

func SDXYOrderList(c *gin.Context) {
	uid := c.GetInt("uid")
	role, _ := c.Get("role")
	isAdmin := role == "super" || role == "admin"

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	searchType := c.Query("searchType")
	keyword := c.Query("keyword")
	statusFilter := c.Query("status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	svc := service.NewSDXYService()
	orders, total, err := svc.ListOrders(uid, isAdmin, page, limit, searchType, keyword, statusFilter)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.SuccessPage(c, orders, int64(total), page, limit)
}

// ---------- 下单 ----------

func SDXYAddOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		Form map[string]interface{} `json:"form"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Form == nil {
		response.BadRequest(c, "参数不完整")
		return
	}
	svc := service.NewSDXYService()
	msg, err := svc.AddOrder(uid, req.Form)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}

// ---------- 删除/退款 ----------

func SDXYDeleteOrder(c *gin.Context) {
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
	svc := service.NewSDXYService()
	msg, err := svc.DeleteOrder(uid, req.ID, isAdmin)
	if err != nil {
		response.BusinessError(c, -1, err.Error())
		return
	}
	response.SuccessMsg(c, msg)
}
