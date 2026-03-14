package xm

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册 XM 平台路由。
func RegisterRoutes(api *gin.RouterGroup) {
	xm := api.Group("/xm")
	{
		xm.GET("/projects", XMGetProjects)
		xm.POST("/add-order", XMAddOrder)
		xm.POST("/add-order-km", XMAddOrderKM)
		xm.GET("/orders", XMGetOrders)
		xm.POST("/query-run", XMQueryRun)
		xm.GET("/refund", XMRefundOrder)
		xm.GET("/delete", XMDeleteOrder)
		xm.GET("/sync", XMSyncOrder)
		xm.GET("/order-logs", XMGetOrderLogs)
	}
}
