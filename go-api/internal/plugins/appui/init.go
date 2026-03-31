package appui

import "go-api/internal/pluginregistry"

func init() {
	pluginregistry.Register(pluginregistry.Plugin{
		Name:           "appui",
		RegisterRoutes: RegisterRoutes,
	})
}
