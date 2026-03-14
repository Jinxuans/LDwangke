package push

import "github.com/gin-gonic/gin"

// RegisterPublicRoutes 注册推送绑定相关公开路由。
func RegisterPublicRoutes(r *gin.Engine) {
	push := r.Group("/api/v1/push")
	{
		push.POST("/bind-wx", PushBindWxUID)
		push.POST("/unbind-wx", PushUnbindWxUID)
		push.POST("/bind-email", PushBindEmail)
		push.POST("/unbind-email", PushUnbindEmail)
		push.POST("/bind-showdoc", PushBindShowDoc)
		push.POST("/unbind-showdoc", PushUnbindShowDoc)
		push.POST("/wx-qrcode", PushWxQRCode)
		push.POST("/wx-scan-uid", PushWxScanUID)
		push.GET("/puplogin", PushPupLogin)
	}
}
