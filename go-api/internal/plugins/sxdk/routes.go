package sxdk

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册 SXDK 平台路由。
func RegisterRoutes(api *gin.RouterGroup) {
	sxdk := api.Group("/sxdk")
	{
		sxdk.GET("/config", ConfigGet)
		sxdk.POST("/config", ConfigSave)
		sxdk.POST("/price", SXDKGetPrice)
		sxdk.GET("/orders", SXDKOrderList)
		sxdk.POST("/add", SXDKAddOrder)
		sxdk.POST("/delete", SXDKDeleteOrder)
		sxdk.POST("/edit", SXDKEditOrder)
		sxdk.POST("/search-phone-info", SXDKSearchPhoneInfo)
		sxdk.POST("/get-log", SXDKGetLog)
		sxdk.POST("/now-check", SXDKNowCheck)
		sxdk.POST("/change-check-code", SXDKChangeCheckCode)
		sxdk.POST("/change-holiday-code", SXDKChangeHolidayCode)
		sxdk.POST("/get-wx-push", SXDKGetWxPush)
		sxdk.POST("/query-source-order", SXDKQuerySourceOrder)
		sxdk.POST("/sync", SXDKSyncOrders)
		sxdk.POST("/get-userrow", SXDKGetUserrow)
		sxdk.POST("/get-async-task", SXDKGetAsyncTask)
		sxdk.POST("/xxy-school-list", SXDKXxyGetSchoolList)
		sxdk.POST("/xxy-address-search", SXDKXxyAddressSearch)
		sxdk.POST("/xxt-school-list", SXDKXxtGetSchoolList)
	}
}
