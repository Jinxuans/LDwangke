package yfdk

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册 YFDK 平台路由。
func RegisterRoutes(api *gin.RouterGroup) {
	yfdk := api.Group("/yfdk")
	{
		yfdk.GET("/config", ConfigGet)
		yfdk.POST("/config", ConfigSave)
		yfdk.POST("/price", YFDKGetPrice)
		yfdk.GET("/projects", YFDKGetProjects)
		yfdk.POST("/account-info", YFDKGetAccountInfo)
		yfdk.POST("/schools", YFDKGetSchools)
		yfdk.POST("/search-schools", YFDKSearchSchools)
		yfdk.GET("/orders", YFDKOrderList)
		yfdk.POST("/add", YFDKAddOrder)
		yfdk.POST("/delete", YFDKDeleteOrder)
		yfdk.POST("/renew", YFDKRenewOrder)
		yfdk.POST("/save", YFDKSaveOrder)
		yfdk.POST("/manual-clock", YFDKManualClock)
		yfdk.POST("/logs", YFDKGetOrderLogs)
		yfdk.POST("/detail", YFDKGetOrderDetail)
		yfdk.POST("/patch-report", YFDKPatchReport)
		yfdk.POST("/calculate-patch-cost", YFDKCalculatePatchCost)
	}
}
