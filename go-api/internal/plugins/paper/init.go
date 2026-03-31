package paper

import "go-api/internal/pluginregistry"

func init() {
	pluginregistry.Register(pluginregistry.Plugin{
		Name:           "paper",
		RegisterRoutes: RegisterRoutes,
		EnsureTable:    func() { Paper().EnsureTable() },
	})
}
