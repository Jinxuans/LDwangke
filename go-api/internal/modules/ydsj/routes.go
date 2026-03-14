package ydsj

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册 YDSJ 平台路由。
func RegisterRoutes(api *gin.RouterGroup) {
	ydsj := api.Group("/ydsj")
	{
		ydsj.GET("/config", YDSJConfigGet)
		ydsj.POST("/config", YDSJConfigSave)
		ydsj.POST("/price", YDSJGetPrice)
		ydsj.GET("/schools", YDSJGetSchools)
		ydsj.GET("/orders", YDSJOrderList)
		ydsj.POST("/add", YDSJAddOrder)
		ydsj.POST("/refund", YDSJRefundOrder)
		ydsj.POST("/edit-remarks", YDSJEditRemarks)
		ydsj.POST("/sync-order", YDSJSyncOrder)
		ydsj.POST("/toggle-run", YDSJToggleRun)
	}
}
