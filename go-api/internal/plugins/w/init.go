package w

import "go-api/internal/pluginregistry"

func init() {
	pluginregistry.Register(pluginregistry.Plugin{
		Name:           "w",
		RegisterRoutes: RegisterRoutes,
		EnsureTable:    func() { W().EnsureTable() },
	})
}
