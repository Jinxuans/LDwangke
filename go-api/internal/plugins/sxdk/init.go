package sxdk

import "go-api/internal/pluginregistry"

func init() {
	pluginregistry.Register(pluginregistry.Plugin{
		Name:           "sxdk",
		RegisterRoutes: RegisterRoutes,
	})
}
