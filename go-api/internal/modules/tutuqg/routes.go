package tutuqg

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册图图强国平台路由。
func RegisterRoutes(api *gin.RouterGroup) {
	tutuqg := api.Group("/tutuqg")
	{
		tutuqg.GET("/orders", TutuQGOrderList)
		tutuqg.POST("/price", TutuQGGetPrice)
		tutuqg.POST("/add", TutuQGAddOrder)
		tutuqg.POST("/delete", TutuQGDeleteOrder)
		tutuqg.POST("/renew", TutuQGRenewOrder)
		tutuqg.POST("/change-password", TutuQGChangePassword)
		tutuqg.POST("/change-token", TutuQGChangeToken)
		tutuqg.POST("/refund", TutuQGRefundOrder)
		tutuqg.POST("/sync", TutuQGSyncOrder)
		tutuqg.POST("/batch-sync", TutuQGBatchSync)
		tutuqg.POST("/toggle-renew", TutuQGToggleAutoRenew)
	}
}
