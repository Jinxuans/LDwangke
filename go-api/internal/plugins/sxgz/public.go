package sxgz

import (
	"strconv"
	"strings"

	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func getRole(c *gin.Context) string {
	if role := c.GetString("role"); role != "" {
		return role
	}
	if raw, ok := c.Get("role"); ok {
		if role, ok := raw.(string); ok {
			return role
		}
	}
	return ""
}

func isAdminRole(role string) bool {
	return role == "super" || role == "admin"
}

func requireAdmin(c *gin.Context) bool {
	if isAdminRole(getRole(c)) || c.GetInt("uid") == 1 {
		return true
	}
	response.Forbidden(c, "权限不足")
	return false
}

func RegisterRoutes(api *gin.RouterGroup) {
	sxgz := api.Group("/sxgz")
	{
		sxgz.GET("/config", GetConfig)
		sxgz.POST("/config", SaveConfig)
		sxgz.GET("/companies", GetCompanies)
		sxgz.POST("/companies/refresh", RefreshCompanies)
		sxgz.GET("/license-companies", GetLicenseCompanies)
		sxgz.GET("/announcements", GetAnnouncements)
		sxgz.GET("/print-options", GetPrintOptions)
		sxgz.POST("/price/quote", QuoteOrder)
		sxgz.POST("/orders", CreateOrder)
		sxgz.GET("/orders", ListOrders)
		sxgz.GET("/orders/:id", GetOrder)
		sxgz.GET("/orders/:id/files", GetOrderFiles)
		sxgz.POST("/orders/:id/files", UploadOrderFile)
		sxgz.POST("/orders/:id/refund", ApplyRefund)

		sxgzAdmin := sxgz.Group("/admin")
		{
			sxgzAdmin.GET("/orders", AdminListOrders)
			sxgzAdmin.PATCH("/orders/:id", AdminUpdateOrder)
			sxgzAdmin.POST("/orders/:id/status", AdminUpdateStatus)
			sxgzAdmin.GET("/stats", AdminStatsHandler)
			sxgzAdmin.POST("/sync", AdminSyncOrders)
		}
	}
}

func GetConfig(c *gin.Context) {
	cfg, err := Sxgz().loadConfig()
	if err != nil {
		response.ServerErrorf(c, err, "获取配置失败")
		return
	}
	// hide secret for non-admins
	if !requireConfigAdmin(c) {
		cfg.UpstreamKey = ""
	}
	response.Success(c, cfg)
}

func SaveConfig(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	var cfg SxgzConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	cfg = normalizeSxgzConfig(cfg)
	if err := Sxgz().saveConfig(cfg); err != nil {
		response.ServerErrorf(c, err, "保存配置失败")
		return
	}
	response.SuccessMsg(c, "配置已保存")
}

func GetCompanies(c *gin.Context) {
	uid := c.GetInt("uid")
	search := strings.TrimSpace(c.Query("search"))
	list, err := Sxgz().GetCompanies(uid, search)
	if err != nil {
		response.ServerErrorf(c, err, "获取公司列表失败")
		return
	}
	response.Success(c, list)
}

func RefreshCompanies(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	var refreshCfg *SxgzConfig
	if c.Request.ContentLength != 0 {
		var cfg SxgzConfig
		if err := c.ShouldBindJSON(&cfg); err == nil && hasRefreshConfigPayload(cfg) {
			normalized := normalizeSxgzConfig(cfg)
			if err := Sxgz().saveConfig(normalized); err != nil {
				response.ServerErrorf(c, err, "保存配置失败")
				return
			}
			refreshCfg = &normalized
		}
	}
	var err error
	if refreshCfg != nil {
		_, err = Sxgz().RefreshCompaniesWithConfig(c.Request.Context(), *refreshCfg)
	} else {
		_, err = Sxgz().RefreshCompanies(c.Request.Context())
	}
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	list, err := Sxgz().GetCompanies(c.GetInt("uid"), "")
	if err != nil {
		response.ServerErrorf(c, err, "获取公司列表失败")
		return
	}
	response.Success(c, list)
}

func hasRefreshConfigPayload(cfg SxgzConfig) bool {
	return strings.TrimSpace(cfg.UpstreamURL) != "" ||
		cfg.UpstreamUID > 0 ||
		strings.TrimSpace(cfg.UpstreamKey) != "" ||
		cfg.PriceMultiplier > 0
}

func GetLicenseCompanies(c *gin.Context) {
	list, err := Sxgz().GetLicenseCompanies(c.GetInt("uid"), strings.TrimSpace(c.Query("search")))
	if err != nil {
		response.ServerErrorf(c, err, "获取营业执照公司失败")
		return
	}
	response.Success(c, list)
}

func GetAnnouncements(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", c.DefaultQuery("limit", "20")))
	noticeType := strings.TrimSpace(c.DefaultQuery("type", "全站公告"))
	result, err := Sxgz().GetAnnouncements(c.Request.Context(), SxgzAnnouncementRequest{
		Page:     page,
		PageSize: pageSize,
		Type:     noticeType,
	})
	if err != nil {
		response.ServerErrorf(c, err, "获取公告失败")
		return
	}
	response.Success(c, result)
}

func GetPrintOptions(c *gin.Context) {
	cfg, err := Sxgz().loadConfig()
	if err != nil {
		response.ServerErrorf(c, err, "获取配置失败")
		return
	}
	response.Success(c, gin.H{
		"print_options":    cfg.PrintOptions,
		"delivery_options": cfg.DeliveryOptions,
		"print_pricing":    cfg.PrintPricing,
	})
}

func QuoteOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	var req OrderQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	result, err := Sxgz().Quote(uid, req)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func CreateOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	var req OrderQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	result, err := Sxgz().CreateOrder(uid, req, requestBaseURL(c))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func ListOrders(c *gin.Context) {
	uid := c.GetInt("uid")
	role := getRole(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", c.DefaultQuery("limit", "20")))
	search := strings.TrimSpace(c.Query("search"))
	status := strings.TrimSpace(c.Query("status"))
	result, err := Sxgz().ListOrders(uid, isAdminRole(role), page, size, search, status)
	if err != nil {
		response.ServerErrorf(c, err, "查询订单失败")
		return
	}
	response.Success(c, result)
}

func GetOrder(c *gin.Context) {
	uid := c.GetInt("uid")
	role := getRole(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	order, err := Sxgz().GetOrder(uid, id, isAdminRole(role))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, orderToMap(order))
}

func GetOrderFiles(c *gin.Context) {
	uid := c.GetInt("uid")
	role := getRole(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	files, err := Sxgz().ListFileRecords(uid, id, isAdminRole(role))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, files)
}

func UploadOrderFile(c *gin.Context) {
	uid := c.GetInt("uid")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件")
		return
	}
	pageCount, _ := strconv.Atoi(c.PostForm("page_count"))
	result, err := Sxgz().UploadFile(uid, id, fileHeader, requestBaseURL(c), c.PostForm("file_print_options"), pageCount)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func ApplyRefund(c *gin.Context) {
	uid := c.GetInt("uid")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := Sxgz().ApplyRefund(uid, id, req.Reason); err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, "退款申请已提交")
}

func AdminListOrders(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	uid := c.GetInt("uid")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", c.DefaultQuery("limit", "20")))
	search := strings.TrimSpace(c.Query("search"))
	status := strings.TrimSpace(c.Query("status"))
	result, err := Sxgz().ListOrders(uid, true, page, size, search, status)
	if err != nil {
		response.ServerErrorf(c, err, "查询订单失败")
		return
	}
	response.Success(c, result)
}

func AdminUpdateOrder(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	uid := c.GetInt("uid")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req struct {
		Status       string `json:"status"`
		AdminNotes   string `json:"admin_notes"`
		RefundReason string `json:"refund_reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.Status == "" {
		current, err := Sxgz().GetOrder(uid, id, true)
		if err != nil {
			response.BusinessError(c, 1001, err.Error())
			return
		}
		req.Status = current.Status
	}
	if err := Sxgz().UpdateOrderStatus(uid, id, req.Status, req.AdminNotes, req.RefundReason, true); err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, "订单已更新")
}

func AdminUpdateStatus(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	uid := c.GetInt("uid")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的订单ID")
		return
	}
	var req struct {
		Status       string `json:"status"`
		AdminNotes   string `json:"admin_notes"`
		RefundReason string `json:"refund_reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := Sxgz().UpdateOrderStatus(uid, id, req.Status, req.AdminNotes, req.RefundReason, true); err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, "状态已更新")
}

func AdminStatsHandler(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	stats, err := Sxgz().AdminStats()
	if err != nil {
		response.ServerErrorf(c, err, "获取统计失败")
		return
	}
	response.Success(c, stats)
}

func AdminSyncOrders(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	updated, err := Sxgz().SyncOrders(c.Request.Context())
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, gin.H{"updated": updated})
}

func requestBaseURL(c *gin.Context) string {
	scheme := "http"
	if strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") || c.Request.TLS != nil {
		scheme = "https"
	}
	host := c.Request.Host
	if host == "" {
		host = c.Request.URL.Host
	}
	if host == "" {
		return ""
	}
	return scheme + "://" + host
}

func requireConfigAdmin(c *gin.Context) bool {
	return isAdminRole(getRole(c)) || c.GetInt("uid") == 1
}
