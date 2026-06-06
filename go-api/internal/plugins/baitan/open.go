package baitan

import (
	"strconv"
	"strings"

	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func RegisterOpenRoutes(openapi *gin.RouterGroup) {
	bt := openapi.Group("/baitan")
	{
		bt.GET("/platforms", OpenPlatforms)
		bt.GET("/price", OpenPrice)
		bt.GET("/orders", OpenOrders)
		bt.POST("/orders", OpenCreateOrder)
		bt.POST("/phone-info", OpenPhoneInfo)
		bt.POST("/edit", OpenEditOrder)
		bt.POST("/add-days", OpenAddDays)
		bt.POST("/delete", OpenDeleteOrder)
		bt.POST("/source-order", OpenSourceOrder)
		bt.POST("/logs", OpenLogs)
		bt.GET("/notice", OpenNotice)
		bt.GET("/schools", OpenSchools)
		bt.POST("/buka/estimate", OpenBukaEstimate)
		bt.POST("/buka", OpenBukaSubmit)
	}
}

func OpenPlatforms(c *gin.Context) {
	list, err := Baitan().Platforms(c.GetInt("uid"))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, list)
}
func OpenPrice(c *gin.Context) {
	price, err := Baitan().PlatformPrice(c.GetInt("uid"), c.Query("type"))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, gin.H{"price": price})
}

func OpenOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", c.DefaultQuery("size", "20")))
	list, total, err := Baitan().ListOrders(c.GetInt("uid"), false, page, limit, c.Query("search"), c.Query("keyword"), c.Query("status"), 0)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessPage(c, list, int64(total), page, limit)
}

func OpenCreateOrder(c *gin.Context) {
	req, ok := bindOrderRequest(c)
	if !ok {
		return
	}
	result, err := Baitan().CreateOrder(c.Request.Context(), c.GetInt("uid"), req, "agent", c.GetInt("uid"))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}
func OpenPhoneInfo(c *gin.Context) {
	req, ok := bindOrderRequest(c)
	if !ok {
		return
	}
	result, err := Baitan().SearchPhoneInfo(c.Request.Context(), req)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func OpenEditOrder(c *gin.Context) {
	req, ok := bindOrderRequest(c)
	if !ok {
		return
	}
	order, err := Baitan().findOrderByAccount(c.GetInt("uid"), req.UserName, req.Type, false)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	req.ID = order.ID
	result, err := Baitan().EditOrder(c.Request.Context(), c.GetInt("uid"), req, false)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func OpenAddDays(c *gin.Context) {
	var req struct {
		UserName     string `json:"userName"`
		Type         string `json:"type"`
		PlatformType string `json:"platformType"`
		Days         int    `json:"days"`
		Remark       int    `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, 1001, "参数错误")
		return
	}
	platform := strings.TrimSpace(firstNonEmpty(req.Type, req.PlatformType))
	days := req.Days
	if days <= 0 {
		days = req.Remark
	}
	order, err := Baitan().findOrderByAccount(c.GetInt("uid"), strings.TrimSpace(req.UserName), platform, false)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	result, err := Baitan().AddDays(c.Request.Context(), c.GetInt("uid"), order.ID, days, false)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func OpenDeleteOrder(c *gin.Context) {
	var req struct {
		UserName     string `json:"userName"`
		Type         string `json:"type"`
		PlatformType string `json:"platformType"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, 1001, "参数错误")
		return
	}
	order, err := Baitan().findOrderByAccount(c.GetInt("uid"), strings.TrimSpace(req.UserName), strings.TrimSpace(firstNonEmpty(req.Type, req.PlatformType)), false)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	result, err := Baitan().DeleteOrder(c.Request.Context(), c.GetInt("uid"), order.ID, false)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func OpenSourceOrder(c *gin.Context) {
	var req struct {
		UserName     string `json:"userName"`
		Type         string `json:"type"`
		PlatformType string `json:"platformType"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, 1001, "参数错误")
		return
	}
	order, err := Baitan().findOrderByAccount(c.GetInt("uid"), strings.TrimSpace(req.UserName), strings.TrimSpace(firstNonEmpty(req.Type, req.PlatformType)), false)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	result, err := Baitan().QuerySourceOrder(c.Request.Context(), c.GetInt("uid"), order.ID, false)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func OpenLogs(c *gin.Context) {
	var req struct {
		UserName     string `json:"userName"`
		Type         string `json:"type"`
		PlatformType string `json:"platformType"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, 1001, "参数错误")
		return
	}
	order, err := Baitan().findOrderByAccount(c.GetInt("uid"), strings.TrimSpace(req.UserName), strings.TrimSpace(firstNonEmpty(req.Type, req.PlatformType)), false)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	result, err := Baitan().Logs(c.Request.Context(), c.GetInt("uid"), order.ID, false)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func OpenNotice(c *gin.Context) {
	result, err := Baitan().Notice(c.Request.Context())
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}
func OpenSchools(c *gin.Context) {
	result, err := Baitan().Schools(c.Request.Context(), c.Query("platform"), c.Query("dictKey"))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}
func OpenBukaEstimate(c *gin.Context) { BukaEstimateHandler(c) }
func OpenBukaSubmit(c *gin.Context)   { BukaSubmitHandler(c) }
