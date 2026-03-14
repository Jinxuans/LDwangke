package php

import "github.com/gin-gonic/gin"

// RegisterRoutes 注册 legacy PHP 代理与 bridge 路由。
func RegisterRoutes(r *gin.Engine, api *gin.RouterGroup) {
	r.Any("/php-api/*path", Proxy())

	phpBridge := r.Group("/internal/php-bridge")
	{
		phpBridge.POST("/money", BridgeMoneyChange)
		phpBridge.GET("/user", BridgeGetUser)
		phpBridge.POST("/order", BridgeCreateOrder)
	}

	api.GET("/php-bridge/auth-url", BridgeAuthURL)
}
