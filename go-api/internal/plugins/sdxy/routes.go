package sdxy

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册 SDXY 平台路由。
func RegisterRoutes(api *gin.RouterGroup) {
	sdxy := api.Group("/sdxy")
	{
		sdxy.GET("/config", SDXYConfigGet)
		sdxy.POST("/config", SDXYConfigSave)
		sdxy.GET("/price", SDXYGetPrice)
		sdxy.GET("/orders", SDXYOrderList)
		sdxy.POST("/add", SDXYAddOrder)
		sdxy.POST("/delete", SDXYDeleteOrder)
		sdxy.POST("/refund", SDXYRefundOrder)
		sdxy.POST("/pause", SDXYPauseOrder)
		sdxy.POST("/get-user-info", SDXYGetUserInfo)
		sdxy.POST("/send-code", SDXYSendCode)
		sdxy.POST("/get-user-info-by-code", SDXYGetUserInfoByCode)
		sdxy.POST("/update-run-rule", SDXYUpdateRunRule)
		sdxy.POST("/log", SDXYGetRunTask)
		sdxy.POST("/change-task-time", SDXYChangeTaskTime)
		sdxy.POST("/delay-task", SDXYDelayTask)
	}
}
