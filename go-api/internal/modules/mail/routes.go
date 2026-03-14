package mail

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册站内信路由。
func RegisterRoutes(api *gin.RouterGroup) {
	mail := api.Group("/mail")
	{
		mail.GET("/list", MailList)
		mail.GET("/unread", MailUnread)
		mail.GET("/:id", MailDetail)
		mail.POST("/send", MailSend)
		mail.POST("/upload", MailUpload)
		mail.DELETE("/:id", MailDelete)
	}
}
