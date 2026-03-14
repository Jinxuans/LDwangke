package order

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册用户侧订单路由。
func RegisterRoutes(api *gin.RouterGroup) {
	order := api.Group("/order")
	{
		order.POST("/list", OrderList)
		order.GET("/list", OrderList)
		order.GET("/stats", OrderStats)
		order.GET("/:oid", OrderDetail)
		order.POST("/add", OrderAdd)
		order.POST("/status", OrderChangeStatus)
		order.POST("/cancel", OrderCancel)
		order.POST("/cancel/:oid", OrderCancel)
		order.POST("/refund", OrderRefund)
		order.GET("/pause", OrderPause)
		order.POST("/changepass", OrderChangePassword)
		order.GET("/resubmit", OrderResubmit)
		order.POST("/pup-reset", OrderPupReset)
		order.GET("/logs", OrderLogs)
	}
}

// RegisterAdminRoutes 注册后台订单管理相关路由。
func RegisterAdminRoutes(admin *gin.RouterGroup) {
	admin.POST("/order/dock", OrderManualDock)
	admin.POST("/order/sync", OrderSyncProgress)
	admin.POST("/order/batch-sync", OrderBatchSync)
	admin.POST("/order/batch-resend", OrderBatchResend)
	admin.POST("/order/remarks", OrderModifyRemarks)
	admin.GET("/order/pause", OrderPause)
	admin.POST("/order/changepass", OrderChangePassword)
	admin.GET("/order/resubmit", OrderResubmit)
	admin.POST("/order/pup-reset", OrderPupReset)
	admin.GET("/order/logs", OrderLogs)
	admin.GET("/ticket/order-counts", OrderTicketCounts)
}
