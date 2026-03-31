package tuboshu

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册 Tuboshu 平台路由。
func RegisterRoutes(api *gin.RouterGroup) {
	tbs := api.Group("/tuboshu")
	{
		tbs.GET("/config", TuboshuUserConfigGet)
		tbs.POST("/route", TuboshuRoute)
		tbs.POST("/route-formdata", TuboshuRouteFormData)
		tbs.GET("/orders", TuboshuOrderList)
	}
}
