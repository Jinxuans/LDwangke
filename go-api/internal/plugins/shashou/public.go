package shashou

import (
	"strconv"
	"strings"

	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup) {
	ss := api.Group("/shashou")
	{
		ss.GET("/projects", GetProjects)
		ss.GET("/version-info", GetVersionInfo)
		ss.GET("/price-preview", PricePreview)
		ss.POST("/orders", CreateOrder)
		ss.GET("/orders", ListOrders)
		ss.POST("/orders/:id/sync", SyncOrder)
		ss.POST("/query", QueryAccount)
		ss.POST("/refund", RefundAccount)
		ss.GET("/accounts", ListAccounts)

		admin := ss.Group("/admin")
		{
			admin.GET("/projects", AdminListProjects)
			admin.POST("/projects", AdminSaveProject)
			admin.DELETE("/projects/:id", AdminDeleteProject)
			admin.GET("/orders", AdminListOrders)
			admin.GET("/accounts", AdminListAccounts)
			admin.POST("/orders/:id/sync", AdminSyncOrder)
			admin.POST("/sync-pending", AdminSyncPending)
		}
	}
}

func isAdmin(c *gin.Context) bool {
	role := c.GetString("role")
	return role == "super" || role == "admin" || c.GetInt("uid") == 1
}

func requireAdmin(c *gin.Context) bool {
	if isAdmin(c) {
		return true
	}
	response.Forbidden(c, "权限不足")
	return false
}

func GetVersionInfo(c *gin.Context) {
	response.Success(c, ShaShou().VersionInfo(c.Request.Context()))
}

func GetProjects(c *gin.Context) {
	list, err := ShaShou().ListProjects(isAdmin(c))
	if err != nil {
		response.ServerErrorf(c, err, "查询项目失败")
		return
	}
	response.Success(c, list)
}

func PricePreview(c *gin.Context) {
	projectID, _ := strconv.Atoi(c.Query("project_id"))
	orderType, _ := strconv.Atoi(c.DefaultQuery("order_type", "1"))
	rush := c.Query("is_rush_order") == "1" || c.Query("is_rush_order") == "true"
	result, err := ShaShou().PricePreview(c.GetInt("uid"), projectID, orderType, rush, nil)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	result, err := ShaShou().CreateOrder(c.Request.Context(), c.GetInt("uid"), req, "local", 0)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func QueryAccount(c *gin.Context) {
	var req QueryOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	result, err := ShaShou().QueryAccount(c.Request.Context(), c.GetInt("uid"), req, "local", 0)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func RefundAccount(c *gin.Context) {
	var req RefundOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	result, err := ShaShou().RefundAccount(c.Request.Context(), c.GetInt("uid"), req, isAdmin(c), "local", 0)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", c.DefaultQuery("size", "20")))
	list, total, err := ShaShou().ListOrders(c.GetInt("uid"), isAdmin(c), page, limit, c.Query("status"), c.Query("order_no"), c.Query("account"), 0)
	if err != nil {
		response.ServerErrorf(c, err, "查询订单失败")
		return
	}
	response.SuccessPage(c, list, int64(total), page, limit)
}

func ListAccounts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", c.DefaultQuery("size", "20")))
	orderType, _ := strconv.Atoi(c.Query("order_type"))
	list, total, err := ShaShou().ListAccounts(c.GetInt("uid"), isAdmin(c), page, limit, c.Query("status"), c.Query("order_no"), c.Query("account"), orderType, 0)
	if err != nil {
		response.ServerErrorf(c, err, "查询账号失败")
		return
	}
	response.SuccessPage(c, list, int64(total), page, limit)
}

func SyncOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	result, err := ShaShou().SyncOrder(c.Request.Context(), c.GetInt("uid"), id, isAdmin(c))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func AdminListProjects(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	list, err := ShaShou().ListProjects(true)
	if err != nil {
		response.ServerErrorf(c, err, "查询项目失败")
		return
	}
	response.Success(c, list)
}

func AdminSaveProject(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	var req Project
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	id, err := ShaShou().SaveProject(req)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, gin.H{"id": id})
}

func AdminDeleteProject(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	if err := ShaShou().DeleteProject(id); err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func AdminListOrders(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", c.DefaultQuery("size", "20")))
	filterUserID, _ := strconv.Atoi(c.Query("filter_user_id"))
	list, total, err := ShaShou().ListOrders(c.GetInt("uid"), true, page, limit, c.Query("status"), c.Query("order_no"), c.Query("account"), filterUserID)
	if err != nil {
		response.ServerErrorf(c, err, "查询订单失败")
		return
	}
	response.SuccessPage(c, list, int64(total), page, limit)
}

func AdminListAccounts(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", c.DefaultQuery("size", "20")))
	orderType, _ := strconv.Atoi(c.Query("order_type"))
	filterUserID, _ := strconv.Atoi(c.Query("filter_user_id"))
	list, total, err := ShaShou().ListAccounts(c.GetInt("uid"), true, page, limit, c.Query("status"), c.Query("order_no"), c.Query("account"), orderType, filterUserID)
	if err != nil {
		response.ServerErrorf(c, err, "查询账号失败")
		return
	}
	response.SuccessPage(c, list, int64(total), page, limit)
}

func AdminSyncOrder(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	result, err := ShaShou().SyncOrder(c.Request.Context(), c.GetInt("uid"), id, true)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func AdminSyncPending(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	var req struct {
		Limit int `json:"limit"`
	}
	_ = c.ShouldBindJSON(&req)
	updated, err := ShaShou().SyncPending(c.Request.Context(), req.Limit)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, gin.H{"updated": updated})
}

func cleanString(c *gin.Context, key string) string {
	return strings.TrimSpace(c.Query(key))
}
