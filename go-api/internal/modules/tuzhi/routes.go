package tuzhi

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册图纸平台路由。
func RegisterRoutes(api *gin.RouterGroup) {
	tuzhi := api.Group("/tuzhi")
	{
		tuzhi.GET("/goods", TuZhiGetGoods)
		tuzhi.POST("/schools", TuZhiGetSchools)
		tuzhi.POST("/sign-info", TuZhiGetSignInfo)
		tuzhi.POST("/calculate-days", TuZhiCalculateDays)
		tuzhi.GET("/orders", TuZhiOrderList)
		tuzhi.POST("/add", TuZhiAddOrder)
		tuzhi.POST("/edit", TuZhiEditOrder)
		tuzhi.POST("/delete", TuZhiDeleteOrder)
		tuzhi.POST("/checkin-work", TuZhiCheckInWork)
		tuzhi.POST("/checkout-work", TuZhiCheckOutWork)
		tuzhi.POST("/sync", TuZhiSyncOrders)
	}
}
