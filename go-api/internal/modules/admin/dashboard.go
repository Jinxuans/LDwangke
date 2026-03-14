package admin

import (
	"log"
	"strconv"

	"go-api/internal/model"
	chatmodule "go-api/internal/modules/chat"
	"go-api/internal/queue"
	"go-api/internal/response"

	"github.com/gin-gonic/gin"
)

func registerDashboardRoutes(admin *gin.RouterGroup) {
	admin.GET("/dashboard", AdminDashboard)
	admin.GET("/stats", AdminStats)
	admin.GET("/moneylog", AdminMoneyLog)
	admin.GET("/queue/stats", AdminQueueStats)
	admin.POST("/queue/concurrency", AdminQueueSetConcurrency)
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
		response.ServerError(c, "查询统计失败")
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
		response.ServerError(c, "查询统计失败")
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
		log.Printf("[AdminMoneyLog] 查询失败: %v", err)
		response.ServerError(c, "查询流水失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":       list,
		"pagination": gin.H{"page": page, "limit": limit, "total": total},
	})
}

func AdminQueueStats(c *gin.Context) {
	if queue.GlobalDockQueue == nil {
		response.ServerError(c, "队列未初始化")
		return
	}
	response.Success(c, queue.GlobalDockQueue.Stats())
}

func AdminQueueSetConcurrency(c *gin.Context) {
	var body struct {
		MaxWorkers int `json:"max_workers" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if queue.GlobalDockQueue == nil {
		response.ServerError(c, "队列未初始化")
		return
	}
	queue.GlobalDockQueue.SetMaxWorkers(body.MaxWorkers)
	response.Success(c, queue.GlobalDockQueue.Stats())
}

func AdminSupplierRanking(c *gin.Context) {
	list, err := supplierRanking()
	if err != nil {
		response.ServerError(c, "查询货源排行失败")
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
		response.ServerError(c, "查询代理商品排行失败")
		return
	}
	response.Success(c, list)
}

func AdminChatSessions(c *gin.Context) {
	sessions, err := chatmodule.Chat().AdminSessions()
	if err != nil {
		response.ServerError(c, err.Error())
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
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, rows)
}

func AdminChatStats(c *gin.Context) {
	stats, err := chatmodule.Chat().ChatStats()
	if err != nil {
		response.ServerError(c, err.Error())
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
		response.ServerError(c, err.Error())
		return
	}
	trimmed, _ := chatmodule.Chat().TrimSessionMessages()
	response.Success(c, gin.H{
		"archived": archived,
		"trimmed":  trimmed,
	})
}
