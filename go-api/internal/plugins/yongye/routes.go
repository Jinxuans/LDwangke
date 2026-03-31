package yongye

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册 Yongye 平台路由。
func RegisterRoutes(api *gin.RouterGroup) {
	yongye := api.Group("/yongye")
	{
		yongye.GET("/config", YongyeConfigGet)
		yongye.POST("/config", YongyeConfigSave)
		yongye.GET("/schools", YongyeGetSchools)
		yongye.GET("/orders", YongyeOrderList)
		yongye.GET("/students", YongyeStudentList)
		yongye.POST("/add", YongyeAddOrder)
		yongye.POST("/refund", YongyeLocalRefund)
		yongye.POST("/refund-student", YongyeRefundStudent)
		yongye.POST("/update-student", YongyeUpdateStudent)
		yongye.POST("/toggle-polling", YongyeTogglePolling)
	}
}
