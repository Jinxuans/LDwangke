package shashou

import (
	"encoding/json"
	"strconv"
	"strings"

	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func RegisterOpenRoutes(openapi *gin.RouterGroup) {
	ss := openapi.Group("/shashou")
	{
		ss.GET("/projects", OpenProjects)
		ss.POST("/projects", OpenProjects)
		ss.GET("/orders", OpenOrders)
		ss.POST("/orders", OpenCreateOrder)
		ss.GET("/orders/:id", OpenOrderDetail)
		ss.POST("/orders/:id/sync", OpenSyncOrder)
		ss.POST("/query", OpenQueryAccount)
		ss.POST("/refund", OpenRefundAccount)
		ss.GET("/accounts", OpenAccounts)
		ss.POST("/accounts", OpenAccounts)
	}
}

func OpenProjects(c *gin.Context) {
	list, err := ShaShou().ListProjects(false)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, list)
}

func OpenCreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, 1001, "参数错误")
		return
	}
	result, err := ShaShou().CreateOrder(c.Request.Context(), c.GetInt("uid"), req, "agent", c.GetInt("uid"))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func OpenQueryAccount(c *gin.Context) {
	var req QueryOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, 1001, "参数错误")
		return
	}
	result, err := ShaShou().QueryAccount(c.Request.Context(), c.GetInt("uid"), req, "agent", c.GetInt("uid"))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func OpenRefundAccount(c *gin.Context) {
	var req RefundOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BusinessError(c, 1001, "参数错误")
		return
	}
	result, err := ShaShou().RefundAccount(c.Request.Context(), c.GetInt("uid"), req, false, "agent", c.GetInt("uid"))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func OpenOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", c.DefaultQuery("size", "20")))
	list, total, err := ShaShou().ListOrders(c.GetInt("uid"), false, page, limit, c.Query("status"), c.Query("order_no"), c.Query("account"), 0)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessPage(c, list, int64(total), page, limit)
}

func OpenOrderDetail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	order, err := ShaShou().findOrder(c.GetInt("uid"), id, false)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, openOrderSnapshot(order))
}

func OpenSyncOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if _, err := ShaShou().SyncOrder(c.Request.Context(), c.GetInt("uid"), id, false); err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	order, err := ShaShou().findOrder(c.GetInt("uid"), id, false)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, openOrderSnapshot(order))
}

func OpenAccounts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", c.DefaultQuery("size", "20")))
	orderType, _ := strconv.Atoi(c.Query("order_type"))
	list, total, err := ShaShou().ListAccounts(c.GetInt("uid"), false, page, limit, c.Query("status"), c.Query("order_no"), c.Query("account"), orderType, 0)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessPage(c, list, int64(total), page, limit)
}

func openOrderSnapshot(order Order) gin.H {
	accounts, _, _ := ShaShou().ListAccounts(order.UserID, false, 1, 500, "", order.OrderNo, "", 0, 0)
	detail := make([]gin.H, 0, len(accounts))
	for _, acc := range accounts {
		detail = append(detail, gin.H{
			"id":            acc.ID,
			"account":       acc.Account,
			"distance":      acc.Distance,
			"status":        acc.Status,
			"error_message": acc.ErrorMessage,
			"processed_at":  acc.ProcessedAt,
		})
	}
	body := gin.H{
		"id":             order.ID,
		"order_id":       order.ID,
		"order_no":       order.OrderNo,
		"order_type":     order.OrderType,
		"status":         order.Status,
		"order_status":   order.Status,
		"payment_status": order.PaymentStatus,
		"pre_deduct":     order.PreDeduct,
		"actual_cost":    nullableFloat(order.ActualCost),
		"actual":         nullableFloat(order.ActualCost),
		"final_charge":   nullableFloat(order.FinalCharge),
		"difference":     nullableFloat(order.Difference),
		"refund_km":      nullableFloat(order.RefundKM),
		"error_message":  strings.TrimSpace(order.ErrorMessage),
		"detail": gin.H{
			"accounts":       detail,
			"total_distance": order.TotalDistance,
			"account_count":  order.AccountCount,
			"is_rush_order":  order.IsRushOrder == 1,
		},
		"amounts": gin.H{
			"pre_deduct": nullableFloat(&order.PreDeduct),
			"actual":     nullableFloat(order.ActualCost),
			"final":      nullableFloat(order.FinalCharge),
			"difference": nullableFloat(order.Difference),
		},
		"timestamps": gin.H{
			"created":   order.CreatedAt,
			"completed": order.CompletedAt,
			"updated":   order.UpdatedAt,
		},
	}
	if jsonHasValue(order.ResultData) {
		var raw map[string]any
		if err := json.Unmarshal(order.ResultData, &raw); err == nil {
			body["upstream_result"] = raw
			if id := extractUpstreamOrderID(raw); id != "" {
				body["upstream_order_id"] = id
			}
		}
	}
	return body
}

func nullableFloat(value *float64) any {
	if value == nil {
		return nil
	}
	return roundMoney(*value)
}
