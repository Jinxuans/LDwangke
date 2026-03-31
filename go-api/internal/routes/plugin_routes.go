package routes

import (
	_ "go-api/internal/plugins" // 触发所有插件 init() 自注册
	"go-api/internal/pluginregistry"

	"github.com/gin-gonic/gin"
)

func registerPluginRoutes(api *gin.RouterGroup) {
	pluginregistry.RegisterAllRoutes(api)
}
