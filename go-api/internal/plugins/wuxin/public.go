package wuxin

import (
	"strconv"
	"strings"

	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup) {
	wuxin := api.Group("/wuxin")
	{
		wuxin.GET("/config", GetConfig)
		wuxin.POST("/config", SaveConfig)
		wuxin.GET("/price", GetPrice)
		wuxin.POST("/school-info", GetSchoolInfo)
		wuxin.POST("/add", CreateOrder)
		wuxin.GET("/orders", ListOrders)
		wuxin.POST("/refund", RefundOrder)
		wuxin.POST("/records", OrderRecords)
		wuxin.POST("/order-config", OrderConfig)
		wuxin.POST("/edit", EditOrder)
		wuxin.POST("/increase", IncreaseOrder)
		wuxin.POST("/reassign", ReassignOrder)
		wuxin.POST("/edit-run-time", UnsupportedTaskAction)
		wuxin.POST("/rerun", UnsupportedTaskAction)

		admin := wuxin.Group("/admin")
		{
			admin.POST("/sync", AdminSyncOrders)
		}
	}
}

func wuxinRole(c *gin.Context) string {
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

func wuxinIsAdmin(c *gin.Context) bool {
	role := wuxinRole(c)
	return role == "super" || role == "admin" || c.GetInt("uid") == 1
}

func wuxinRequireAdmin(c *gin.Context) bool {
	if wuxinIsAdmin(c) {
		return true
	}
	response.Forbidden(c, "权限不足")
	return false
}

func GetConfig(c *gin.Context) {
	cfg, err := Wuxin().loadConfig()
	if err != nil {
		response.ServerErrorf(c, err, "获取配置失败")
		return
	}
	if !wuxinIsAdmin(c) {
		cfg.UpstreamKey = ""
		cfg.APIKey = ""
	}
	response.Success(c, cfg)
}

func SaveConfig(c *gin.Context) {
	if !wuxinRequireAdmin(c) {
		return
	}
	var cfg WuxinConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := Wuxin().saveConfig(normalizeWuxinConfig(cfg)); err != nil {
		response.ServerErrorf(c, err, "保存配置失败")
		return
	}
	response.SuccessMsg(c, "配置已保存")
}

func GetPrice(c *gin.Context) {
	result, err := Wuxin().GetPrice(c.GetInt("uid"))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func GetSchoolInfo(c *gin.Context) {
	var req struct {
		AuthCode string `json:"auth_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "授权码不能为空")
		return
	}
	result, err := Wuxin().SchoolInfo(c.Request.Context(), strings.TrimSpace(req.AuthCode))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func CreateOrder(c *gin.Context) {
	var req WuxinOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	result, err := Wuxin().CreateOrder(c.Request.Context(), c.GetInt("uid"), req, "local", 0)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", c.DefaultQuery("size", "20")))
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	var status *int
	if raw := strings.TrimSpace(c.Query("status")); raw != "" {
		n, _ := strconv.Atoi(raw)
		status = &n
	}
	list, total, err := Wuxin().ListOrders(c.GetInt("uid"), wuxinIsAdmin(c), page, limit, c.Query("searchType"), c.Query("keyword"), status)
	if err != nil {
		response.ServerErrorf(c, err, "查询订单失败")
		return
	}
	response.SuccessPage(c, list, int64(total), page, limit)
}

func RefundOrder(c *gin.Context) {
	var req struct {
		ID          int    `json:"id"`
		OrderNumber string `json:"order_number"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	result, err := Wuxin().RefundOrder(c.Request.Context(), c.GetInt("uid"), req.ID, req.OrderNumber, wuxinIsAdmin(c))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func OrderRecords(c *gin.Context) {
	var req struct {
		ID          int    `json:"id"`
		OrderNumber string `json:"order_number"`
		Page        int    `json:"page"`
		Limit       int    `json:"limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	result, err := Wuxin().OrderRecords(c.Request.Context(), c.GetInt("uid"), req.ID, req.OrderNumber, req.Page, req.Limit, wuxinIsAdmin(c))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func OrderConfig(c *gin.Context) {
	var req struct {
		ID          int    `json:"id"`
		OrderNumber string `json:"order_number"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	result, err := Wuxin().OrderConfig(c.Request.Context(), c.GetInt("uid"), req.ID, req.OrderNumber, wuxinIsAdmin(c))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func EditOrder(c *gin.Context) {
	var req struct {
		ID   int               `json:"id" binding:"required"`
		Form WuxinOrderRequest `json:"form"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := Wuxin().EditOrder(c.Request.Context(), c.GetInt("uid"), req.ID, req.Form, wuxinIsAdmin(c)); err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, "编辑订单成功")
}

func IncreaseOrder(c *gin.Context) {
	var req struct {
		ID          int    `json:"id"`
		OrderNumber string `json:"order_number"`
		Quantity    int    `json:"quantity" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	result, err := Wuxin().IncreaseOrder(c.Request.Context(), c.GetInt("uid"), req.ID, req.OrderNumber, req.Quantity, wuxinIsAdmin(c))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func ReassignOrder(c *gin.Context) {
	var req struct {
		ID          int    `json:"id"`
		OrderNumber string `json:"order_number"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := Wuxin().ReassignOrder(c.Request.Context(), c.GetInt("uid"), req.ID, req.OrderNumber, wuxinIsAdmin(c)); err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, "重新分配成功")
}

func UnsupportedTaskAction(c *gin.Context) {
	response.BusinessError(c, 1001, Wuxin().UnsupportedTaskAction().Error())
}

func AdminSyncOrders(c *gin.Context) {
	if !wuxinRequireAdmin(c) {
		return
	}
	updated, err := Wuxin().SyncOrders(c.Request.Context())
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, gin.H{"updated": updated})
}
