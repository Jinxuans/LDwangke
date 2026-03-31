package tutuqg

import "go-api/internal/pluginregistry"

func init() {
	pluginregistry.Register(pluginregistry.Plugin{
		Name:           "tutuqg",
		RegisterRoutes: RegisterRoutes,
		EnsureTable:    func() { TutuQG().EnsureTable() },
	})
}
