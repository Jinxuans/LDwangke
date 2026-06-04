package jiguang

import (
	"strconv"
	"strings"

	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup) {
	jg := api.Group("/jiguang")
	{
		jg.GET("/config", GetConfig)
		jg.POST("/config", SaveConfig)
		jg.GET("/products", GetProducts)
		jg.GET("/prices", GetPrices)
		jg.POST("/schools", GetSchools)
		jg.POST("/orders", CreateOrder)
		jg.GET("/orders", ListOrders)
		jg.POST("/refund/preview", RefundPreviewHandler)
		jg.POST("/refund/confirm", RefundConfirmHandler)
		jg.POST("/add-times/preview", AddTimesPreviewHandler)
		jg.POST("/add-times/confirm", AddTimesConfirmHandler)
		jg.POST("/order-logs", OrderLogsHandler)

		admin := jg.Group("/admin")
		{
			admin.POST("/sync", AdminSyncOrders)
		}
	}
}

func role(c *gin.Context) string {
	if value := c.GetString("role"); value != "" {
		return value
	}
	return ""
}

func isAdmin(c *gin.Context) bool {
	r := role(c)
	return r == "super" || r == "admin" || c.GetInt("uid") == 1
}

func requireAdmin(c *gin.Context) bool {
	if isAdmin(c) {
		return true
	}
	response.Forbidden(c, "权限不足")
	return false
}

func GetConfig(c *gin.Context) {
	cfg, err := Jiguang().loadConfig()
	if err != nil {
		response.ServerErrorf(c, err, "获取配置失败")
		return
	}
	if !isAdmin(c) {
		cfg.APIKey = ""
		cfg.UpstreamKey = ""
		cfg.UpstreamUID = 0
	}
	response.Success(c, cfg)
}

func SaveConfig(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	var cfg Config
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := Jiguang().saveConfig(normalizeConfig(cfg)); err != nil {
		response.ServerErrorf(c, err, "保存配置失败")
		return
	}
	response.SuccessMsg(c, "配置已保存")
}

func GetProducts(c *gin.Context) {
	list, err := Jiguang().Products(c.GetInt("uid"))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, list)
}

func GetPrices(c *gin.Context) {
	prices, err := Jiguang().ProductPrices(c.GetInt("uid"))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, prices)
}

func GetSchools(c *gin.Context) {
	var req struct {
		Page     int    `json:"page"`
		PageSize int    `json:"pageSize"`
		Keyword  string `json:"keyword"`
	}
	_ = c.ShouldBindJSON(&req)
	result, err := Jiguang().Schools(c.Request.Context(), req.Page, req.PageSize, req.Keyword)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func CreateOrder(c *gin.Context) {
	var req OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	result, err := Jiguang().CreateOrder(c.Request.Context(), c.GetInt("uid"), req, "local", 0)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", c.DefaultQuery("size", "20")))
	filterUID, _ := strconv.Atoi(c.Query("filter_uid"))
	list, total, err := Jiguang().ListOrders(
		c.GetInt("uid"), isAdmin(c), page, limit,
		c.Query("searchType"), c.Query("keyword"), c.Query("status"), c.Query("school"), filterUID,
	)
	if err != nil {
		response.ServerErrorf(c, err, "查询订单失败")
		return
	}
	response.SuccessPage(c, list, int64(total), page, limit)
}

func RefundPreviewHandler(c *gin.Context) {
	handleRefund(c, false)
}

func RefundConfirmHandler(c *gin.Context) {
	handleRefund(c, true)
}

func handleRefund(c *gin.Context, confirm bool) {
	var req struct {
		OrderNo string `json:"order_no"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.OrderNo) == "" {
		response.BadRequest(c, "订单号不能为空")
		return
	}
	result, err := Jiguang().RefundOrder(c.Request.Context(), c.GetInt("uid"), req.OrderNo, isAdmin(c), confirm)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func AddTimesPreviewHandler(c *gin.Context) {
	handleAddTimes(c, false)
}

func AddTimesConfirmHandler(c *gin.Context) {
	handleAddTimes(c, true)
}

func handleAddTimes(c *gin.Context, confirm bool) {
	var req struct {
		OrderNo string `json:"order_no"`
		Delta   int    `json:"delta"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.OrderNo) == "" {
		response.BadRequest(c, "参数错误")
		return
	}
	result, err := Jiguang().AddTimes(c.Request.Context(), c.GetInt("uid"), req.OrderNo, req.Delta, isAdmin(c), confirm)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func OrderLogsHandler(c *gin.Context) {
	var req struct {
		OrderNo string `json:"order_no"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.OrderNo) == "" {
		response.BadRequest(c, "订单号不能为空")
		return
	}
	result, err := Jiguang().OrderLogs(c.Request.Context(), c.GetInt("uid"), req.OrderNo, isAdmin(c))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func AdminSyncOrders(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	updated, err := Jiguang().SyncOrders(c.Request.Context())
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, gin.H{"updated": updated})
}
