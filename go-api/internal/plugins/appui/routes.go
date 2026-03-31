package appui

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册 AppUI 平台路由。
func RegisterRoutes(api *gin.RouterGroup) {
	appui := api.Group("/appui")
	{
		appui.GET("/config", AppuiConfigGet)
		appui.POST("/config", AppuiConfigSave)
		appui.POST("/price", AppuiGetPrice)
		appui.GET("/courses", AppuiGetCourses)
		appui.GET("/orders", AppuiOrderList)
		appui.POST("/add", AppuiAddOrder)
		appui.POST("/edit", AppuiEditOrder)
		appui.POST("/renew", AppuiRenewOrder)
		appui.POST("/delete", AppuiDeleteOrder)
	}
}
