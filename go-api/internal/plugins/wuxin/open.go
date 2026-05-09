package wuxin

import (
	"strconv"
	"strings"

	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func RegisterOpenRoutes(openapi *gin.RouterGroup) {
	wuxin := openapi.Group("/wuxin")
	{
		wuxin.GET("/orders", OpenOrderList)
		wuxin.POST("/orders", OpenCreateOrder)
		wuxin.POST("/school-info", OpenSchoolInfo)
		wuxin.POST("/refund", OpenRefund)
		wuxin.POST("/records", OpenRecords)
		wuxin.POST("/edit", OpenEdit)
		wuxin.POST("/increase", OpenIncrease)
		wuxin.POST("/reassign", OpenReassign)
	}
}

func OpenSchoolInfo(c *gin.Context) {
	var req struct {
		AuthCode string `json:"auth_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, 0, "授权码不能为空")
		return
	}
	result, err := Wuxin().SchoolInfo(c.Request.Context(), strings.TrimSpace(req.AuthCode))
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, result)
}

func OpenCreateOrder(c *gin.Context) {
	var req WuxinOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, 0, "参数错误")
		return
	}
	result, err := Wuxin().CreateOrder(c.Request.Context(), c.GetInt("uid"), req, "agent", c.GetInt("uid"))
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, result)
}

func OpenOrderList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", c.DefaultQuery("size", "20")))
	list, total, err := Wuxin().ListOrders(c.GetInt("uid"), false, page, limit, c.Query("searchType"), c.Query("keyword"), nil)
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.SuccessPage(c, list, int64(total), page, limit)
}

func OpenRefund(c *gin.Context) {
	var req struct {
		ID          int    `json:"id"`
		OrderNumber string `json:"order_number"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, 0, "参数错误")
		return
	}
	result, err := Wuxin().RefundOrder(c.Request.Context(), c.GetInt("uid"), req.ID, req.OrderNumber, false)
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, result)
}

func OpenRecords(c *gin.Context) {
	var req struct {
		ID          int    `json:"id"`
		OrderNumber string `json:"order_number"`
		Page        int    `json:"page"`
		Limit       int    `json:"limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, 0, "参数错误")
		return
	}
	result, err := Wuxin().OrderRecords(c.Request.Context(), c.GetInt("uid"), req.ID, req.OrderNumber, req.Page, req.Limit, false)
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, result)
}

func OpenEdit(c *gin.Context) {
	var req struct {
		OrderNumber string            `json:"order_number"`
		Form        WuxinOrderRequest `json:"form"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, 0, "参数错误")
		return
	}
	order, err := Wuxin().findOrder(c.GetInt("uid"), 0, req.OrderNumber, false)
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	if err := Wuxin().EditOrder(c.Request.Context(), c.GetInt("uid"), order.ID, req.Form, false); err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.SuccessMsg(c, "编辑订单成功")
}

func OpenIncrease(c *gin.Context) {
	var req struct {
		OrderNumber string `json:"order_number"`
		Quantity    int    `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, 0, "参数错误")
		return
	}
	result, err := Wuxin().IncreaseOrder(c.Request.Context(), c.GetInt("uid"), 0, req.OrderNumber, req.Quantity, false)
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, result)
}

func OpenReassign(c *gin.Context) {
	var req struct {
		OrderNumber string `json:"order_number"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, 0, "参数错误")
		return
	}
	if err := Wuxin().ReassignOrder(c.Request.Context(), c.GetInt("uid"), 0, req.OrderNumber, false); err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.SuccessMsg(c, "重新分配成功")
}
