package module

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册 legacy module 兼容路由。
func RegisterRoutes(api *gin.RouterGroup) {
	api.GET("/module/:app_id/frame-url", FrameURL)
	api.Any("/module/:app_id", Proxy)
}
