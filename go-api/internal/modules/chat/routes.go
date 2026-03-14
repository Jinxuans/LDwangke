package chat

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册聊天域路由。
func RegisterRoutes(api *gin.RouterGroup) {
	chat := api.Group("/chat")
	{
		chat.GET("/sessions", Sessions)
		chat.GET("/messages/:list_id", Messages)
		chat.GET("/history/:list_id", History)
		chat.GET("/new/:list_id", NewMessages)
		chat.POST("/send", Send)
		chat.POST("/send-image", SendImage)
		chat.POST("/read/:list_id", MarkRead)
		chat.GET("/unread", Unread)
		chat.POST("/create", Create)
	}
}
