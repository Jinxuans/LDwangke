package tuboshu

import "go-api/internal/pluginregistry"

func init() {
	pluginregistry.Register(pluginregistry.Plugin{
		Name:           "tuboshu",
		RegisterRoutes: RegisterRoutes,
		EnsureTable:    func() { Tuboshu().EnsureTable() },
	})
}
