package sdxy

import "go-api/internal/pluginregistry"

func init() {
	pluginregistry.Register(pluginregistry.Plugin{
		Name:           "sdxy",
		RegisterRoutes: RegisterRoutes,
	})
}
