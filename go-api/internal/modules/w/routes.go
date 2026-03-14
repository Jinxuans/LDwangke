package w

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册 W 平台路由。
func RegisterRoutes(api *gin.RouterGroup) {
	w := api.Group("/w")
	{
		w.GET("/apps", WGetApps)
		w.POST("/add-order", WAddOrder)
		w.GET("/orders", WGetOrders)
		w.POST("/refund", WRefundOrder)
		w.GET("/sync", WSyncOrder)
		w.GET("/resume", WResumeOrder)
		w.POST("/proxy", WProxyAction)
		w.POST("/edit-order", WEditOrder)
		w.POST("/change-status", WChangeRunStatus)
		w.POST("/remain-count", WGetRemainCount)
		w.POST("/task-data", WGetTaskData)
		w.POST("/edit-task", WEditTask)
		w.POST("/delay-task", WDelayTask)
		w.POST("/fast-delay", WFastDelayTask)
	}
}
