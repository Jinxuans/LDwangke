package openapi

import (
	"go-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterCompatRoutes 注册旧版 PHP/OpenAPI 兼容入口。
func RegisterCompatRoutes(r *gin.Engine) {
	r.Any("/api.php", Compat)
	r.Any("/api/index.php", Compat)
}

// RegisterRoutes 注册 APIKey 保护的 OpenAPI 路由。
func RegisterRoutes(r *gin.Engine) {
	openapi := r.Group("/api/v1/open", middleware.APIKeyAuth())
	{
		openapi.GET("/classlist", OpenAPIGetClass)
		openapi.POST("/classlist", OpenAPIGetClass)
		openapi.GET("/query", OpenAPIQuery)
		openapi.POST("/query", OpenAPIQuery)
		openapi.GET("/order", OpenAPIAddOrder)
		openapi.POST("/order", OpenAPIAddOrder)
		openapi.GET("/orderlist", OpenAPIOrderList)
		openapi.POST("/orderlist", OpenAPIOrderList)
		openapi.GET("/balance", OpenAPIBalance)
		openapi.GET("/chadan", OpenAPIChadan)
		openapi.POST("/chadan", OpenAPIChadan)
		openapi.POST("/bindpushuid", OpenAPIBindPushUID)
		openapi.POST("/bindpushemail", OpenAPIBindPushEmail)
		openapi.POST("/bindshowdocpush", OpenAPIBindShowDocPush)
	}
}
