package admin

import (
	"strconv"

	"go-api/internal/dockscheduler"
	"go-api/internal/model"
	chatmodule "go-api/internal/modules/chat"
	ordermodule "go-api/internal/modules/order"
	"go-api/internal/response"
	"go-api/internal/runtimeops"

	"github.com/gin-gonic/gin"
)

func registerDashboardRoutes(admin *gin.RouterGroup) {
	admin.GET("/dashboard", AdminDashboard)
	admin.GET("/stats", AdminStats)
	admin.GET("/moneylog", AdminMoneyLog)
	admin.GET("/dock-scheduler/stats", AdminDockSchedulerStats)
	admin.GET("/dock-scheduler/logs", AdminDockSchedulerLogs)
	admin.POST("/dock-scheduler/config", AdminDockSchedulerConfig)
	admin.POST("/dock-scheduler/run", AdminDockSchedulerRunNow)
	admin.GET("/order-progress-sync/stats", AdminOrderProgressSyncStats)
	admin.GET("/order-progress-sync/logs", AdminOrderProgressSyncLogs)
	admin.POST("/order-progress-sync/config", AdminOrderProgressSyncConfig)
	admin.POST("/order-progress-sync/run", AdminOrderProgressSyncRunNow)
	admin.GET("/rank/suppliers", AdminSupplierRanking)
	admin.GET("/rank/agent-products", AdminAgentProductRanking)
	admin.GET("/chat/sessions", AdminChatSessions)
	admin.GET("/chat/messages/:list_id", AdminChatMessages)
	admin.GET("/chat/stats", AdminChatStats)
	admin.POST("/chat/cleanup", AdminChatCleanup)
}

func AdminDashboard(c *gin.Context) {
	stats, err := dashboardStats()
	if err != nil {
		response.ServerErrorf(c, err, "查询统计失败")
		return
	}
	response.Success(c, stats)
}

func AdminStats(c *gin.Context) {
	days, _ := strconv.Atoi(c.Query("days"))
	if days <= 0 {
		days = 30
	}
	stats, err := statsReport(days)
	if err != nil {
		response.ServerErrorf(c, err, "查询统计失败")
		return
	}
	response.Success(c, stats)
}

func AdminMoneyLog(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	uid := c.Query("uid")
	logType := c.Query("type")
	list, total, err := adminMoneyLogList(page, limit, uid, logType)
	if err != nil {
		response.ServerErrorf(c, err, "查询流水失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":       list,
		"pagination": gin.H{"page": page, "limit": limit, "total": total},
	})
}

func AdminDockSchedulerStats(c *gin.Context) {
	response.Success(c, dockscheduler.Snapshot())
}

func AdminDockSchedulerLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	response.Success(c, dockscheduler.RecentLogs(limit))
}

func AdminDockSchedulerConfig(c *gin.Context) {
	var body struct {
		IntervalSec int `json:"interval_sec"`
		BatchLimit  int `json:"batch_limit"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	stats, err := dockscheduler.UpdateConfig(body.IntervalSec, body.BatchLimit)
	if err != nil {
		response.ServerErrorf(c, err, err.Error())
		return
	}
	response.Success(c, stats)
}

func AdminDockSchedulerRunNow(c *gin.Context) {
	stats, err := dockscheduler.RunOnce("manual")
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, stats)
}

func AdminOrderProgressSyncStats(c *gin.Context) {
	response.Success(c, runtimeops.GetOrderProgressSyncStatus())
}

func AdminOrderProgressSyncLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	response.Success(c, runtimeops.GetOrderProgressSyncLogs(limit))
}

func AdminOrderProgressSyncConfig(c *gin.Context) {
	var body struct {
		Enabled          bool     `json:"enabled"`
		IntervalSec      int      `json:"interval_sec"`
		BatchEnabled     bool     `json:"batch_enabled"`
		BatchIntervalSec int      `json:"batch_interval_sec"`
		SupplierIDs      []int    `json:"supplier_ids"`
		ExcludedStatuses []string `json:"excluded_statuses"`
		Rules            []struct {
			Key             string `json:"key"`
			Label           string `json:"label"`
			MinAgeHours     int    `json:"min_age_hours"`
			MaxAgeHours     int    `json:"max_age_hours"`
			IntervalMinutes int    `json:"interval_minutes"`
			Enabled         bool   `json:"enabled"`
		} `json:"rules"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	rules := make([]ordermodule.AutoSyncRule, 0, len(body.Rules))
	for _, rule := range body.Rules {
		rules = append(rules, ordermodule.AutoSyncRule{
			Key:             rule.Key,
			Label:           rule.Label,
			MinAgeHours:     rule.MinAgeHours,
			MaxAgeHours:     rule.MaxAgeHours,
			IntervalMinutes: rule.IntervalMinutes,
			Enabled:         rule.Enabled,
		})
	}
	status, err := runtimeops.UpdateOrderProgressSyncConfig(runtimeops.OrderProgressSyncConfig{
		Enabled:          body.Enabled,
		IntervalSec:      body.IntervalSec,
		BatchEnabled:     body.BatchEnabled,
		BatchIntervalSec: body.BatchIntervalSec,
		SupplierIDs:      body.SupplierIDs,
		ExcludedStatuses: body.ExcludedStatuses,
		Rules:            rules,
	})
	if err != nil {
		response.ServerErrorf(c, err, err.Error())
		return
	}
	response.Success(c, status)
}

func AdminOrderProgressSyncRunNow(c *gin.Context) {
	status, err := runtimeops.RunOrderProgressSyncNow()
	if err != nil {
		response.BusinessError(c, 1001, err.Error())
		return
	}
	response.Success(c, status)
}

func AdminSupplierRanking(c *gin.Context) {
	list, err := supplierRanking()
	if err != nil {
		response.ServerErrorf(c, err, "查询货源排行失败")
		return
	}
	response.Success(c, list)
}

func AdminAgentProductRanking(c *gin.Context) {
	uid, _ := strconv.Atoi(c.Query("uid"))
	timeType := c.DefaultQuery("time", "today")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	list, err := agentProductRanking(uid, timeType, limit)
	if err != nil {
		response.ServerErrorf(c, err, "查询代理商品排行失败")
		return
	}
	response.Success(c, list)
}

func AdminChatSessions(c *gin.Context) {
	sessions, err := chatmodule.Chat().AdminSessions()
	if err != nil {
		response.ServerErrorf(c, err, err.Error())
		return
	}
	response.Success(c, sessions)
}

func AdminChatMessages(c *gin.Context) {
	listID, err := strconv.Atoi(c.Param("list_id"))
	if err != nil {
		response.BadRequest(c, "无效的会话 ID")
		return
	}

	var req model.ChatMessagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Limit = 50
	}
	rows, err := chatmodule.Chat().AdminMessages(listID, req.Limit)
	if err != nil {
		response.ServerErrorf(c, err, err.Error())
		return
	}
	response.Success(c, rows)
}

func AdminChatStats(c *gin.Context) {
	stats, err := chatmodule.Chat().ChatStats()
	if err != nil {
		response.ServerErrorf(c, err, err.Error())
		return
	}
	response.Success(c, stats)
}

func AdminChatCleanup(c *gin.Context) {
	var req struct {
		Days int `json:"days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Days < 1 {
		req.Days = 14
	}
	archived, err := chatmodule.Chat().ManualCleanup(req.Days)
	if err != nil {
		response.ServerErrorf(c, err, err.Error())
		return
	}
	trimmed, _ := chatmodule.Chat().TrimSessionMessages()
	response.Success(c, gin.H{
		"archived": archived,
		"trimmed":  trimmed,
	})
}
