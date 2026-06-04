package jiguang

import (
	"strconv"
	"strings"

	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func RegisterOpenRoutes(openapi *gin.RouterGroup) {
	jg := openapi.Group("/jiguang")
	{
		jg.GET("/products", OpenProducts)
		jg.GET("/prices", OpenPrices)
		jg.POST("/schools", OpenSchools)
		jg.POST("/orders", OpenCreateOrder)
		jg.GET("/orders", OpenOrders)
		jg.POST("/refund/preview", OpenRefundPreview)
		jg.POST("/refund/confirm", OpenRefundConfirm)
		jg.POST("/add-times/preview", OpenAddTimesPreview)
		jg.POST("/add-times/confirm", OpenAddTimesConfirm)
		jg.POST("/order-logs", OpenOrderLogs)
	}
}

func OpenProducts(c *gin.Context) {
	list, err := Jiguang().Products(c.GetInt("uid"))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, list)
}

func OpenPrices(c *gin.Context) {
	prices, err := Jiguang().ProductPrices(c.GetInt("uid"))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, prices)
}

func OpenSchools(c *gin.Context) {
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

func OpenCreateOrder(c *gin.Context) {
	var req OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, 1001, "参数错误")
		return
	}
	result, err := Jiguang().CreateOrder(c.Request.Context(), c.GetInt("uid"), req, "agent", c.GetInt("uid"))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func OpenOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", c.DefaultQuery("size", "20")))
	list, total, err := Jiguang().ListOrders(c.GetInt("uid"), false, page, limit, c.Query("searchType"), c.Query("keyword"), c.Query("status"), c.Query("school"), 0)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessPage(c, list, int64(total), page, limit)
}

func OpenRefundPreview(c *gin.Context) { openRefund(c, false) }
func OpenRefundConfirm(c *gin.Context) { openRefund(c, true) }

func openRefund(c *gin.Context, confirm bool) {
	var req struct {
		OrderNo string `json:"order_no"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.OrderNo) == "" {
		response.BusinessError(c, 1001, "订单号不能为空")
		return
	}
	result, err := Jiguang().RefundOrder(c.Request.Context(), c.GetInt("uid"), req.OrderNo, false, confirm)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func OpenAddTimesPreview(c *gin.Context) { openAddTimes(c, false) }
func OpenAddTimesConfirm(c *gin.Context) { openAddTimes(c, true) }

func openAddTimes(c *gin.Context, confirm bool) {
	var req struct {
		OrderNo string `json:"order_no"`
		Delta   int    `json:"delta"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.OrderNo) == "" {
		response.BusinessError(c, 1001, "参数错误")
		return
	}
	result, err := Jiguang().AddTimes(c.Request.Context(), c.GetInt("uid"), req.OrderNo, req.Delta, false, confirm)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func OpenOrderLogs(c *gin.Context) {
	var req struct {
		OrderNo string `json:"order_no"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.OrderNo) == "" {
		response.BusinessError(c, 1001, "订单号不能为空")
		return
	}
	result, err := Jiguang().OrderLogs(c.Request.Context(), c.GetInt("uid"), req.OrderNo, false)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}
