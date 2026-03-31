package pluginregistry

import "github.com/gin-gonic/gin"

// Plugin 描述一个可插拔的业务模块。
type Plugin struct {
	Name           string
	RegisterRoutes func(api *gin.RouterGroup)
	EnsureTable    func()
}

var plugins []Plugin

// Register 由各插件包的 init() 调用，完成自注册。
func Register(p Plugin) {
	plugins = append(plugins, p)
}

// RegisterAllRoutes 遍历已注册插件，挂载路由。
func RegisterAllRoutes(api *gin.RouterGroup) {
	for _, p := range plugins {
		if p.RegisterRoutes != nil {
			p.RegisterRoutes(api)
		}
	}
}

// EnsureAllTables 遍历已注册插件，执行建表。
func EnsureAllTables() {
	for _, p := range plugins {
		if p.EnsureTable != nil {
			p.EnsureTable()
		}
	}
}
