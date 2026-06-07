package baitan

import (
	"encoding/json"
	"strconv"
	"strings"

	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup) {
	bt := api.Group("/baitan")
	{
		bt.GET("/config", GetConfig)
		bt.GET("/ui-settings", GetUISettings)
		bt.POST("/config", SaveConfig)
		bt.GET("/platforms", GetPlatforms)
		bt.GET("/price", GetPrice)
		bt.GET("/orders", ListOrdersHandler)
		bt.POST("/orders", CreateOrderHandler)
		bt.POST("/phone-info", PhoneInfoHandler)
		bt.POST("/orders/:id/edit", EditOrderHandler)
		bt.POST("/orders/:id/add-days", AddDaysHandler)
		bt.POST("/orders/:id/delete", DeleteOrderHandler)
		bt.POST("/orders/:id/source", QuerySourceOrderHandler)
		bt.POST("/orders/:id/sync", SyncOneHandler)
		bt.POST("/logs", LogsHandler)
		bt.GET("/notice", NoticeHandler)
		bt.GET("/schools", SchoolsHandler)
		bt.POST("/buka/estimate", BukaEstimateHandler)
		bt.POST("/buka", BukaSubmitHandler)
		admin := bt.Group("/admin")
		{
			admin.POST("/sync", AdminSyncHandler)
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

func GetUISettings(c *gin.Context) {
	cfg, err := Baitan().loadConfig()
	if err != nil {
		response.ServerErrorf(c, err, "获取界面配置失败")
		return
	}
	response.Success(c, gin.H{
		"user_page_url":       cfg.UserPageURL,
		"user_page_text":      cfg.UserPageText,
		"user_page_intro":     cfg.UserPageIntro,
		"notice_refresh_text": cfg.NoticeRefreshText,
	})
}

func GetConfig(c *gin.Context) {
	cfg, err := Baitan().loadConfig()
	if err != nil {
		response.ServerErrorf(c, err, "获取配置失败")
		return
	}
	if !isAdmin(c) {
		cfg.Token = ""
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
	if err := Baitan().saveConfig(normalizeConfig(cfg)); err != nil {
		response.ServerErrorf(c, err, "保存配置失败")
		return
	}
	response.SuccessMsg(c, "配置已保存")
}

func GetPlatforms(c *gin.Context) {
	list, err := Baitan().Platforms(c.GetInt("uid"))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, list)
}
func GetPrice(c *gin.Context) {
	price, err := Baitan().PlatformPrice(c.GetInt("uid"), c.Query("type"))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, gin.H{"price": price})
}

func ListOrdersHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", c.DefaultQuery("size", "20")))
	filterUID, _ := strconv.Atoi(c.Query("filter_uid"))
	list, total, err := Baitan().ListOrders(c.GetInt("uid"), isAdmin(c), page, limit, c.Query("search"), c.Query("keyword"), c.Query("status"), filterUID)
	if err != nil {
		response.ServerErrorf(c, err, "查询订单失败")
		return
	}
	response.SuccessPage(c, list, int64(total), page, limit)
}

func CreateOrderHandler(c *gin.Context) {
	req, ok := bindOrderRequest(c)
	if !ok {
		return
	}
	result, err := Baitan().CreateOrder(c.Request.Context(), c.GetInt("uid"), req, "local", 0)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func PhoneInfoHandler(c *gin.Context) {
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

func EditOrderHandler(c *gin.Context) {
	req, ok := bindOrderRequest(c)
	if !ok {
		return
	}
	req.ID, _ = strconv.Atoi(c.Param("id"))
	result, err := Baitan().EditOrder(c.Request.Context(), c.GetInt("uid"), req, isAdmin(c))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func AddDaysHandler(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Days int `json:"days"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Days <= 0 {
		req.Days, _ = strconv.Atoi(c.PostForm("days"))
	}
	result, err := Baitan().AddDays(c.Request.Context(), c.GetInt("uid"), id, req.Days, isAdmin(c))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func DeleteOrderHandler(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	result, err := Baitan().DeleteOrder(c.Request.Context(), c.GetInt("uid"), id, isAdmin(c))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}
func QuerySourceOrderHandler(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	result, err := Baitan().QuerySourceOrder(c.Request.Context(), c.GetInt("uid"), id, isAdmin(c))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}
func SyncOneHandler(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	_, err := Baitan().SyncOne(c.Request.Context(), c.GetInt("uid"), id, isAdmin(c))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.SuccessMsg(c, "同步完成")
}

func LogsHandler(c *gin.Context) {
	var req struct {
		ID int `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ID <= 0 {
		response.BadRequest(c, "订单ID不能为空")
		return
	}
	result, err := Baitan().Logs(c.Request.Context(), c.GetInt("uid"), req.ID, isAdmin(c))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func NoticeHandler(c *gin.Context) {
	result, err := Baitan().Notice(c.Request.Context())
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}
func SchoolsHandler(c *gin.Context) {
	result, err := Baitan().Schools(c.Request.Context(), c.Query("platform"), c.Query("dictKey"))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func BukaEstimateHandler(c *gin.Context) {
	var req BukaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	result, err := Baitan().BukaEstimate(req)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}
func BukaSubmitHandler(c *gin.Context) {
	var req BukaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	result, err := Baitan().SubmitBuka(c.Request.Context(), c.GetInt("uid"), req, isAdmin(c))
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, result)
}

func AdminSyncHandler(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	var req struct {
		Limit int `json:"limit"`
	}
	_ = c.ShouldBindJSON(&req)
	updated, err := Baitan().SyncOrders(c.Request.Context(), req.Limit)
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, gin.H{"updated": updated})
}

func bindOrderRequest(c *gin.Context) (OrderRequest, bool) {
	var raw map[string]any
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.BadRequest(c, "参数错误")
		return OrderRequest{}, false
	}
	data, _ := json.Marshal(raw)
	var req OrderRequest
	_ = json.Unmarshal(data, &req)
	if req.Type == "" {
		req.Type = strings.TrimSpace(asString(raw["platformType"]))
	}
	if req.SID == "" {
		req.SID = strings.TrimSpace(asString(raw["schoolId"]))
	}
	req.Raw = raw
	return req, true
}
