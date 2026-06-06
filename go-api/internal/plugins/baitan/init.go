package baitan

import "go-api/internal/pluginregistry"

func init() {
	pluginregistry.Register(pluginregistry.Plugin{
		Name:           "baitan",
		RegisterRoutes: RegisterRoutes,
		EnsureTable:    func() { Baitan().EnsureTable() },
	})
}
