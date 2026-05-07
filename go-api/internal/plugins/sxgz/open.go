package sxgz

import (
	"strconv"
	"strings"

	"go-api/internal/database"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func RegisterOpenRoutes(openapi *gin.RouterGroup) {
	sxgz := openapi.Group("/sxgz")
	{
		sxgz.GET("/companies", OpenCompanies)
		sxgz.POST("/companies", OpenCompanies)
		sxgz.GET("/announcements", OpenAnnouncements)
		sxgz.POST("/announcements", OpenAnnouncements)
		sxgz.GET("/orders", OpenOrderList)
		sxgz.POST("/orders", OpenCreateOrder)
		sxgz.GET("/orders/:id", OpenOrderDetail)
		sxgz.POST("/order-file", OpenUpdateOrderFile)
		sxgz.POST("/refund", OpenApplyRefund)
	}
}

func OpenCompanies(c *gin.Context) {
	uid := c.GetInt("uid")
	search := strings.TrimSpace(c.Query("search"))
	list, err := Sxgz().GetCompanies(uid, search)
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, list)
}

func OpenAnnouncements(c *gin.Context) {
	payload := readCompatPayload(c)
	page := compatInt(payload, "page")
	pageSize := compatInt(payload, "pageSize")
	if pageSize <= 0 {
		pageSize = compatInt(payload, "limit")
	}
	noticeType := strings.TrimSpace(compatString(payload, "type"))
	if noticeType == "" {
		noticeType = "全站公告"
	}

	result, err := Sxgz().GetAnnouncements(c.Request.Context(), SxgzAnnouncementRequest{
		Page:     page,
		PageSize: pageSize,
		Type:     noticeType,
	})
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, result)
}

func OpenCreateOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	payload := readCompatPayload(c)
	req := compatOrderRequest(payload)
	if req.ServiceType == "" {
		req.ServiceType = "electronic"
	}
	if req.MaterialType == "" {
		req.MaterialType = "upload"
	}
	if req.CompanyID <= 0 || strings.TrimSpace(req.CustomerName) == "" {
		response.BusinessError(c, 0, "缺少 company_id 或 customer_name 参数")
		return
	}

	result, err := Sxgz().CreateOrder(uid, req, requestBaseURL(c))
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	_, _ = database.DB.Exec("UPDATE fd_sxgz_orders SET source = 'agent', agent_uid = ?, updated_at = NOW() WHERE order_id = ?", uid, result.OrderID)
	response.Success(c, result)
}

func OpenOrderList(c *gin.Context) {
	uid := c.GetInt("uid")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", c.DefaultQuery("size", "20")))
	status := strings.TrimSpace(c.Query("status"))
	search := strings.TrimSpace(c.Query("search"))

	result, err := Sxgz().ListOrders(uid, false, page, limit, search, status)
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, result)
}

func OpenOrderDetail(c *gin.Context) {
	uid := c.GetInt("uid")
	orderID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if orderID <= 0 {
		response.BusinessError(c, 0, "无效的订单ID")
		return
	}
	order, err := Sxgz().GetOrder(uid, orderID, false)
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, orderToMap(order))
}

func OpenUpdateOrderFile(c *gin.Context) {
	uid := c.GetInt("uid")
	payload := readCompatPayload(c)
	result, err := updateCompatOrderFile(c, uid, payload)
	if err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, result)
}

func OpenApplyRefund(c *gin.Context) {
	uid := c.GetInt("uid")
	payload := readCompatPayload(c)
	reason := strings.TrimSpace(compatString(payload, "reason"))
	if reason == "" {
		response.BusinessError(c, 0, "退款原因不能为空")
		return
	}

	order, err := findCompatOrder(uid, int64(compatInt(payload, "order_id")), compatString(payload, "order_no"))
	if err != nil {
		response.BusinessError(c, 0, "订单不存在或无权限")
		return
	}
	if err := Sxgz().ApplyRefund(uid, order.OrderID, reason); err != nil {
		response.BusinessError(c, 0, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "退款申请已提交"})
}
